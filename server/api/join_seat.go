package api

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

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
		return "", errFirebaseVerifyTimeout
	}
}

const firebaseVerifyTimeout = 2 * time.Second

var errFirebaseVerifyTimeout = errors.New("Firebase token verification timed out")

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
	// SeatPick is the chosen player index for asymmetric games. -1 or
	// absent means "auto-assign the next open seat" (the symmetric
	// default). The handler pre-fills -1 before decoding so an absent
	// field auto-assigns rather than claiming seat 0.
	SeatPick int `json:"seatPick"`
	// AttemptID is stable across client retries. UID idempotency is the
	// authoritative guard; this identifier makes retries observable and leaves
	// room for stronger audit correlation without trusting client timing.
	AttemptID string `json:"attemptID"`
}

// joinSeatResponse echoes back what the phone needs to navigate to the
// game's Hand view.
type joinSeatResponse struct {
	GameID      string `json:"gameID"`
	GameName    string `json:"gameName"`
	PlayerIndex int    `json:"playerIndex"`
	Resumed     bool   `json:"resumed"`
}

// getSeatJoinLock returns the per-game mutex used to serialize seat-claim
// operations. Lazy-allocates a fresh mutex on first call for a given
// gameID. Locks are never evicted: eviction is unsafe because a
// goroutine that already obtained a pointer from a prior call could
// Lock a deleted mutex while a new goroutine gets a fresh one for the
// same gameID — breaking serialization. The leak is bounded by the
// number of distinct games ever joined in this server process
// (one sync.Mutex each) and is negligible.
func (s *Server) getSeatJoinLock(gameID string) *sync.Mutex {
	s.seatJoinLocksMu.Lock()
	defer s.seatJoinLocksMu.Unlock()

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
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	req := joinSeatRequest{SeatPick: -1}
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
	if !s.requireJoinTicket(c, req.GameID) {
		return
	}
	if strings.TrimSpace(req.AvatarSlug) == "" || !utf8.ValidString(req.AvatarSlug) || utf8.RuneCountInString(req.AvatarSlug) > 64 || !validCompanionAvatarSlug(req.AvatarSlug) {
		writeJoinProblem(c, http.StatusBadRequest, "INVALID_AVATAR", "Choose one of the supported avatars", nil)
		return
	}
	if len(req.AttemptID) > 128 {
		writeJoinProblem(c, http.StatusBadRequest, "INVALID_ATTEMPT", "attemptID is too long", nil)
		return
	}

	normalizedName, err := NormalizeDisplayName(req.DisplayName)
	if err != nil {
		writeJoinProblem(c, http.StatusBadRequest, "INVALID_NAME", err.Error(), nil)
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
			writeJoinProblem(c, http.StatusUnauthorized, "AUTH_REQUIRED", "Sign in again to continue", nil)
			return
		}
		verifiedUID, verifyErr := verifyFirebaseTokenWithTimeout(token, s.config.Firebase.ProjectID, firebaseVerifyTimeout)
		if verifyErr != nil {
			if errors.Is(verifyErr, errFirebaseVerifyTimeout) {
				writeJoinProblem(c, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "Sign-in verification is temporarily unavailable; retry", nil)
			} else {
				writeJoinProblem(c, http.StatusUnauthorized, "AUTH_INVALID", "Your sign-in expired; sign in again", nil)
			}
			return
		}
		if verifiedUID != req.UID {
			writeJoinProblem(c, http.StatusUnauthorized, "AUTH_INVALID", "Sign-in identity did not match this join", nil)
			return
		}
	}

	// Acquire per-game lock for the seat-claim. Held through SetPlayerForGame
	// + seatPresentation write so a racing request can't slip into the same
	// slot.
	lock := s.getSeatJoinLock(req.GameID)
	lock.Lock()
	defer lock.Unlock()

	// Room eligibility is intentionally loaded only after acquiring the
	// per-game mutation lock. This is the linearization point shared with
	// every phone claim, so two phones cannot both act on the same snapshot.
	combined, err := s.storage.CombinedGame(req.GameID)
	if err != nil || combined == nil || combined.CompanionRoomCode == "" {
		writeJoinProblem(c, http.StatusNotFound, "ROOM_NOT_FOUND", "Room not found", nil)
		return
	}
	mgrInfo := s.managers[combined.Name]
	if mgrInfo == nil || mgrInfo.manager == nil {
		writeJoinProblem(c, http.StatusNotFound, "ROOM_NOT_FOUND", "Room not found", nil)
		return
	}
	game := mgrInfo.manager.Game(req.GameID)
	if game == nil {
		writeJoinProblem(c, http.StatusNotFound, "ROOM_NOT_FOUND", "Room not found", nil)
		return
	}
	if combined.Finished {
		writeJoinProblem(c, http.StatusConflict, joinCodeRoomFinished, "This game has finished", nil)
		return
	}

	// Find a slot. SeatPick >= 0 means the phone chose a specific seat via
	// the seat picker (asymmetric games); honor it if it's still open, 409
	// with the latest availability so the phone can re-pick if not.
	// SeatPick < 0 means auto-assign the next open seat (symmetric games).
	userIDs := s.storage.UserIDsForGame(req.GameID)
	var slot boardgame.PlayerIndex = -1
	resumed := false
	if existing, ok := existingSeatForUID(userIDs, req.UID); ok {
		// Retrying after a committed-but-lost response is ordinary success.
		// The original seat wins even if this retry carries a different pick.
		slot = existing
		resumed = true
	} else if combined.CompanionLocked {
		writeJoinProblem(c, http.StatusConflict, joinCodeRoomLocked, "The host locked this room", nil)
		return
	} else if req.SeatPick >= 0 {
		if req.SeatPick >= len(userIDs) {
			writeJoinProblem(c, http.StatusBadRequest, "INVALID_SEAT", "seatPick out of range", nil)
			return
		}
		statuses := s.companionSeatStatuses(game)
		if req.SeatPick >= len(statuses) || !statuses[req.SeatPick].Available {
			writeJoinProblem(c, http.StatusConflict, joinCodeSeatTaken, "That seat is no longer available", gin.H{"slots": statuses})
			return
		}
		slot = boardgame.PlayerIndex(req.SeatPick)
	} else {
		statuses := s.companionSeatStatuses(game)
		for i, status := range statuses {
			if status.Available {
				slot = boardgame.PlayerIndex(i)
				break
			}
		}
		if slot < 0 {
			writeJoinProblem(c, http.StatusConflict, joinCodeRoomFull, "Room is full", gin.H{"slots": statuses})
			return
		}
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
	// A valid cookie for a DIFFERENT user (e.g. the phone was already
	// signed into the site) must be rebound too — otherwise the seat gets
	// bound to req.UID while the session stays on the old user, and the
	// hand view renders as an observer. req.UID is token-verified above,
	// so rebinding cannot be used to hijack someone else's session.
	authCookie := s.getRequestCookie(c)
	cookieUser := s.storage.GetUserByCookie(authCookie)
	if authCookie == "" || cookieUser == nil || cookieUser.ID != user.ID {
		authCookie = randomString(cookieLength)
		if err := s.storage.ConnectCookieToUser(authCookie, user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue session cookie: " + err.Error()})
			return
		}
		s.setAuthCookieOnGin(c, authCookie)
	}

	// Run the existing seat-player flow.
	if !resumed {
		if err := s.doSeatPlayer(game, slot, user); err != nil {
			// A different API instance may have won between our snapshot and
			// the storage uniqueness constraint. Re-read authoritative state:
			// our own UID means a successful replay; another UID means a typed
			// conflict, never an opaque 500 or a second pending SeatPlayer.
			freshIDs := s.storage.UserIDsForGame(req.GameID)
			if existing, ok := existingSeatForUID(freshIDs, req.UID); ok {
				slot = existing
				resumed = true
			} else if req.SeatPick < 0 {
				// Another instance may have won only the first open slot while
				// others remain. Continue against fresh canonical availability;
				// storage uniqueness arbitrates every attempt.
				claimed := false
				for _, status := range s.companionSeatStatuses(game) {
					if !status.Available {
						continue
					}
					candidate := boardgame.PlayerIndex(status.PlayerIndex)
					if err := s.doSeatPlayer(game, candidate, user); err == nil {
						slot = candidate
						claimed = true
						break
					}
					if existing, ok := existingSeatForUID(s.storage.UserIDsForGame(req.GameID), req.UID); ok {
						slot = existing
						resumed = true
						claimed = true
						break
					}
				}
				if !claimed {
					writeJoinProblem(c, http.StatusConflict, joinCodeRoomFull, "Room is full", gin.H{"slots": s.companionSeatStatuses(game)})
					return
				}
			} else {
				statuses := s.companionSeatStatuses(game)
				writeJoinProblem(c, http.StatusConflict, joinCodeSeatTaken, "That seat is no longer available", gin.H{"slots": statuses})
				return
			}
		}
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

	// Tell the room. Seat claims don't bump the game version, so without
	// this the Table view wouldn't learn about the new player until some
	// unrelated state change; clients handle presence-changed by
	// refetching gameInfo, which re-renders the avatar strip and roster.
	s.notifier.enqueuePresenceChange(req.GameID)

	// Start the absence clock: if this phone never actually connects, the
	// heartbeat scanner flags the seat absent 30s from now instead of
	// treating the no-show as permanently present.
	s.notifier.seedHeartbeat(req.GameID, slot)
	if !resumed {
		s.autoCloseGameIfFull(game)
	}

	c.JSON(http.StatusOK, joinSeatResponse{
		GameID:      req.GameID,
		GameName:    combined.Name,
		PlayerIndex: int(slot),
		Resumed:     resumed,
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
