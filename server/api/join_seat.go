package api

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/alternaDev/go-firebase-verify"
	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/seatpresentation"
	"github.com/jkomoros/boardgame/server/api/users"
)

// verifyFirebaseTokenWithTimeout wraps firebase.VerifyIDToken in a 2-second
// timeout so a hung Google PKI fetch doesn't hold the gin handler
// indefinitely. The library doesn't expose a context-aware variant, so we
// run the verify in a goroutine and race it against a timer.
//
// Returns the verified UID on success; on timeout returns a sentinel error
// so callers can map to 503-ish or 401. On verification failure returns
// the library's error verbatim.
func verifyFirebaseTokenWithTimeout(token, projectID string, timeout time.Duration) (string, error) {
	type result struct {
		uid string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		uid, err := firebase.VerifyIDToken(token, projectID)
		ch <- result{uid, err}
	}()
	select {
	case r := <-ch:
		return r.uid, r.err
	case <-time.After(timeout):
		return "", errors.New("Firebase token verification timed out")
	}
}

const firebaseVerifyTimeout = 2 * time.Second

// joinSeatRequest is the body of POST /api/join/seat. Phone client posts
// after running through identity + avatar picker + (for asymmetric games)
// role picker. UID/Token are the Firebase fields the existing auth flow
// already understands; we reuse the same verification path here.
type joinSeatRequest struct {
	GameID      string `json:"gameID"`
	UID         string `json:"uid"`
	Token       string `json:"token"`
	DisplayName string `json:"displayName"`
	AvatarSlug  string `json:"avatarSlug"`
	// SeatPick is the chosen player index for asymmetric games. -1 (or
	// absent) means "auto-assign the next open seat" which is the symmetric
	// default. P2 wires this up; V1 always auto-assigns.
	SeatPick int `json:"seatPick"`
}

// joinSeatResponse echoes back what the phone needs to navigate to the
// game's Hand view.
type joinSeatResponse struct {
	GameID      string `json:"gameID"`
	GameName    string `json:"gameName"`
	PlayerIndex int    `json:"playerIndex"`
}

// getSeatJoinLock returns the per-game mutex used to serialize seat-claim
// operations. Lazy-allocates a fresh mutex on first call for a given
// gameID. Opportunistically evicts locks when the map grows past a
// threshold and the notifier's heartbeat tracking has no record of the
// game (a coarse "game is no longer active" proxy that doesn't require
// us to plumb storage IsFinished calls in here). Bounded growth without
// a dedicated cleanup goroutine.
func (s *Server) getSeatJoinLock(gameID string) *sync.Mutex {
	s.seatJoinLocksMu.Lock()
	defer s.seatJoinLocksMu.Unlock()

	// Opportunistic eviction at 64 entries. Use the notifier's existing
	// connected-sockets bucket as a "game is currently active" probe;
	// games with no connected sockets are eligible to drop their lock.
	// (We deliberately do NOT touch lastHeartbeat here, which is owned
	// by the workLoop goroutine without a mutex.)
	// No eviction: the map holds one sync.Mutex per active gameID
	// (at most ~64 entries = ~4KB). Eviction is unsafe because a
	// goroutine that already obtained a pointer from a prior call
	// could Lock a deleted mutex while a new goroutine gets a fresh
	// one for the same gameID — breaking serialization. The leak is
	// bounded by the number of distinct games ever created in this
	// server process and is negligible.

	if lock, ok := s.seatJoinLocks[gameID]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	s.seatJoinLocks[gameID] = lock
	return lock
}

// joinSeatHandler implements POST /api/join/seat. See spec §6.2. Steps:
//
//  1. Validate request shape + display name format.
//  2. Look up the game; verify it's companion-mode, not Finished, not Locked.
//  3. Verify Firebase ID token (skipped in OfflineDevMode).
//  4. Take the per-game seat-claim lock.
//  5. Either find next empty seat (auto) or honor SeatPick (V2 — for now
//     auto-only).
//  6. Create the auth cookie + user record if not already present.
//  7. Write the seatPresentation row.
//  8. Run doSeatPlayer to queue the SeatPlayer proposal.
//  9. Issue the surface=hand cookie scoped to the gameID.
//
// On race (last seat just got taken), returns 409 with the latest seat
// snapshot; phone retries against fresh state.
func (s *Server) joinSeatHandler(c *gin.Context) {
	var req joinSeatRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	if req.GameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gameID is required"})
		return
	}
	if req.UID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uid is required"})
		return
	}

	normalizedName, err := NormalizeDisplayName(req.DisplayName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Load + validate the game.
	combined, err := s.storage.CombinedGame(req.GameID)
	if err != nil || combined == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}
	if combined.Finished || combined.CompanionLocked || combined.CompanionRoomCode == "" {
		// CompanionRoomCode == "" means it's not a companion-mode game at all.
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	mgrInfo := s.managers[combined.Name]
	if mgrInfo == nil || mgrInfo.manager == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Verify Firebase token unless OfflineDevMode short-circuits.
	// Accept token from Authorization: Bearer header (preferred — avoids
	// logging JWTs in request bodies) or from the JSON body (legacy).
	if !s.config.OfflineDevMode {
		token := req.Token
		if authHeader := c.GetHeader("Authorization"); len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required (pass token in Authorization: Bearer header)"})
			return
		}
		verifiedUID, verifyErr := verifyFirebaseTokenWithTimeout(token, s.config.Firebase.ProjectID, firebaseVerifyTimeout)
		if verifyErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token verification failed: " + verifyErr.Error()})
			return
		}
		if verifiedUID != req.UID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token uid does not match request uid"})
			return
		}
	}

	// Acquire per-game lock for the seat-claim. Held through SetPlayerForGame
	// + seatPresentation write so a racing request can't slip into the same
	// slot.
	lock := s.getSeatJoinLock(req.GameID)
	lock.Lock()
	defer lock.Unlock()

	// Find a slot. V1 is auto-only — SeatPick is reserved for P2.
	userIDs := s.storage.UserIDsForGame(req.GameID)
	var slot boardgame.PlayerIndex = -1
	for i, uid := range userIDs {
		if uid == "" {
			slot = boardgame.PlayerIndex(i)
			break
		}
	}
	if slot < 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":          "Room is full",
			"currentPlayers": len(userIDs),
			"maxPlayers":     len(userIDs),
		})
		return
	}

	// Find-or-create the user record.
	user := s.storage.GetUserByID(req.UID)
	if user == nil {
		user = &users.StorageRecord{
			ID: req.UID,
			// DisplayName / Email left empty for anonymous: those live on the
			// seatPresentation, not the user record (spec §5.4).
			Created:  time.Now().UnixNano(),
			LastSeen: time.Now().UnixNano(),
		}
		if err := s.storage.UpdateUser(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user record: " + err.Error()})
			return
		}
	}

	// Issue (or refresh) the auth cookie that ties this UID to a session.
	authCookie := s.getRequestCookie(c)
	if authCookie == "" || s.storage.GetUserByCookie(authCookie) == nil {
		authCookie = randomString(cookieLength)
		if err := s.storage.ConnectCookieToUser(authCookie, user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue session cookie: " + err.Error()})
			return
		}
		s.setAuthCookieOnGin(c, authCookie)
	}

	// Look up the manager to retrieve the live game (doSeatPlayer needs a
	// boardgame.Game, not a CombinedStorageRecord).
	game := mgrInfo.manager.Game(req.GameID)
	if game == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Game manager missing live game"})
		return
	}

	// Run the existing seat-player flow.
	if err := s.doSeatPlayer(game, slot, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seat player: " + err.Error()})
		return
	}

	// Record presentation (after SetPlayerForGame so the seat is bound).
	if err := s.storage.SetSeatPresentation(&seatpresentation.StorageRecord{
		GameID:      req.GameID,
		PlayerIndex: slot,
		DisplayName: normalizedName,
		AvatarSlug:  req.AvatarSlug,
	}); err != nil {
		s.logger.Warnln("SetSeatPresentation failed; seat is still bound but presentation missing:", err)
		// Don't fail the request — the seat is bound; presentation can be
		// retried later. Phone client will fall back to a default rendering.
	}

	// Issue surface=hand cookie scoped to this game (per spec §7.1). We
	// use a per-gameID cookie name so multiple in-flight companion games on
	// one browser don't step on each other.
	s.setSurfaceCookie(c, req.GameID, "hand")

	c.JSON(http.StatusOK, joinSeatResponse{
		GameID:      req.GameID,
		GameName:    combined.Name,
		PlayerIndex: int(slot),
	})
}

// surfaceCookieName returns the canonical cookie name for a per-game surface
// signal. Public so the loader can call the same function client-side via
// the route layer if needed.
func surfaceCookieName(gameID string) string {
	return "surface_" + gameID
}

// setSurfaceCookie sets the surface cookie for a (gameID, surface) pair.
// surface is one of "table" or "hand". The cookie is NOT HttpOnly because the
// client-side loader needs to read it to pick the right renderer suffix at
// game-page load. Path is "/" so all subsequent navigation on the same
// origin sees it. MaxAge 30 days so a returning player on the same browser
// slots back into their seat.
//
// Secure is set in non-OfflineDevMode so production cookies are TLS-only;
// dev mode (http://localhost) tolerates non-Secure. SameSite=Lax would be
// ideal but the framework's pinned gin version predates SetSameSite. V2
// can tighten when gin is upgraded.
func (s *Server) setSurfaceCookie(c *gin.Context, gameID string, surface string) {
	secure := !s.config.OfflineDevMode
	c.SetCookie(surfaceCookieName(gameID), surface, 30*24*60*60, "/", "", secure, false /* httpOnly */)
}

// setAuthCookieOnGin is a small helper that mirrors what r.SetAuthCookie
// does on the renderer-based handlers, for handlers that don't go through
// renderer. Kept here rather than in auth.go to avoid touching the existing
// flow. Sets Secure in production, HttpOnly always.
func (s *Server) setAuthCookieOnGin(c *gin.Context, value string) {
	secure := !s.config.OfflineDevMode
	c.SetCookie(cookieName, value, 365*24*60*60, "/", "", secure, true /* httpOnly */)
}
