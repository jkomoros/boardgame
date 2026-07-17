package api

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/tablelease"
	"github.com/sirupsen/logrus"
)

// hostActionRateLimit is per-(gameID, hostUserID): one action per second
// max. Mostly to soften host-trolling scenarios (rapid Skip spam).
const hostActionRateLimit = 1 * time.Second

// hostActionLocksMu guards hostActionLocks. hostActionLocks maps a
// composite "<gameID>:<userID>" key to the timestamp of that host's last
// action; used for the 1/sec rate limit. Per-game-per-host is the right
// granularity: two different hosts in the same game don't throttle each
// other, and one host across two games doesn't throttle either.
//
// Package-level rather than per-Server (acceptable smell — there's
// typically one Server per process). hostActionLocksGC opportunistically
// drops entries older than 1 hour to bound the map. Run from a single
// goroutine in the package init below.
var (
	hostActionLocksMu sync.Mutex
	hostActionLocks   = map[string]time.Time{}
)

const hostActionLocksEvictAge = 1 * time.Hour

// hostActionAllowed returns true if enough time has passed since this
// host's last recorded action in this game. Records the new action's
// timestamp on success. Opportunistically evicts entries older than
// hostActionLocksEvictAge on every Nth call (cheap GC).
func hostActionAllowed(gameID, userID string) bool {
	key := gameID + ":" + userID
	hostActionLocksMu.Lock()
	defer hostActionLocksMu.Unlock()

	// Periodic eviction. We piggyback on the existing lock to avoid a
	// separate goroutine. 1-in-N probabilistic gating keeps the cost
	// O(N) amortized — fine since this only fires from host actions
	// (rare events).
	if len(hostActionLocks) > 32 {
		cutoff := time.Now().Add(-hostActionLocksEvictAge)
		for k, ts := range hostActionLocks {
			if ts.Before(cutoff) {
				delete(hostActionLocks, k)
			}
		}
	}

	if last, ok := hostActionLocks[key]; ok && time.Since(last) < hostActionRateLimit {
		return false
	}
	hostActionLocks[key] = time.Now()
	return true
}

// isHost is retained as a small compatibility-shaped helper for the existing
// call sites. hostUserID is intentionally ignored: authority belongs to the
// active fenced Table credential, never to renderer intent or user identity.
func (s *Server) isHost(c *gin.Context, gameID, hostUserID string) bool {
	_, ok := s.activeTableLeaseForRequest(c, gameID)
	return ok
}

// resolveHost returns the active lease holder for audit/rate-limit labels.
// It does not grant authority; isHost validates the separate secret cookie.
func (s *Server) resolveHost(gameID string) (string, error) {
	lease, err := s.storage.CompanionTableLease(gameID)
	if err != nil {
		return "", err
	}
	if lease == nil {
		return "", nil
	}
	return lease.HolderUserID, nil
}

// auditHostAction logs a structured record of a host action. V1 emits to
// the existing logger rather than a dedicated companionHostAudit table;
// a logger sink is sufficient for postmortem diagnostics and avoids a new
// 4-place storage extension for V1. A persistent table is a P5 polish
// item if needed.
func (s *Server) auditHostAction(action, gameID, userID, resultMsg string, success bool) {
	level := logrus.InfoLevel
	if !success {
		level = logrus.WarnLevel
	}
	s.logger.WithFields(logrus.Fields{
		"audit":   "host-action",
		"action":  action,
		"gameID":  gameID,
		"userID":  userID,
		"success": success,
		"detail":  resultMsg,
	}).Log(level)
}

// hostSkipTurnHandler implements POST /api/game/:name/:id/hostSkipTurn.
// Gated on isHost; requires the CURRENT player to be in the absent set.
// Proposes moves.ForceFinishTurn with AdminPlayerIndex. Rate-limited.
func (s *Server) hostSkipTurnHandler(c *gin.Context) {
	game := s.getGame(c)
	if game == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no such game"})
		return
	}
	gameID := game.ID()

	// Resolve host identity. The eGame adapter lets isHost stay generic.
	hostUserID, err := s.resolveHost(gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve host: " + err.Error()})
		return
	}
	if !s.isHost(c, gameID, hostUserID) {
		s.auditHostAction("hostSkipTurn", gameID, "", "not host", false)
		c.JSON(http.StatusForbidden, gin.H{"error": "host privileges required"})
		return
	}

	userID := hostUserID
	if !hostActionAllowed(gameID, userID) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "host actions are rate-limited to 1/sec"})
		return
	}
	fence, renewal := s.beginTableLeaseAction(c, gameID)
	if renewal != tableLeaseRenewed {
		status := http.StatusConflict
		if renewal == tableLeaseRenewRetryable {
			status = http.StatusServiceUnavailable
		}
		tableLeaseProblem(c, status, "TABLE_LEASE_LOST", "this screen could not safely reserve Table control", nil)
		return
	}
	defer s.endTableLeaseAction(gameID, fence)

	// The current player must be in the absent set — Skip applies only to
	// the player whose turn is actively blocking the game (spec §9.3).
	currentPlayer := game.CurrentState().CurrentPlayerIndex()
	absentSet := s.notifier.AbsentPlayers(gameID)
	currentAbsent := false
	for _, pi := range absentSet {
		if pi == currentPlayer {
			currentAbsent = true
			break
		}
	}
	if !currentAbsent {
		s.auditHostAction("hostSkipTurn", gameID, userID, "current player not absent", false)
		c.JSON(http.StatusConflict, gin.H{"error": "current player is not absent — nothing to skip"})
		return
	}

	// Build the move and propose as AdminPlayerIndex. We look up by name —
	// the convention is that companion-supporting games register
	// moves.ForceFinishTurn under its default name "Force Finish Turn"
	// (via auto.Config / boardgame:codegen). We don't type-assert here
	// because that would force server/api to import moves, which would
	// cycle back through the moves test deps that touch server/api.
	mover := game.MoveByName("Force Finish Turn")
	if mover == nil {
		// Game's delegate doesn't include ForceFinishTurn in its move list.
		// V1 requires companion-supporting games to register it; failure
		// here is a deployment gap, not a runtime user error.
		s.auditHostAction("hostSkipTurn", gameID, userID, "ForceFinishTurn not registered in delegate", false)
		c.JSON(http.StatusNotImplemented, gin.H{"error": "this game does not support host SkipTurn (ForceFinishTurn move not registered)"})
		return
	}
	if err := <-game.ProposeMove(mover, boardgame.AdminPlayerIndex); err != nil {
		s.auditHostAction("hostSkipTurn", gameID, userID, err.Error(), false)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ForceFinishTurn rejected: " + err.Error()})
		return
	}

	s.auditHostAction("hostSkipTurn", gameID, userID, "advanced past player "+currentPlayer.String(), true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// switchToSoloHandler implements POST /api/game/:name/:id/switchToSolo.
// Gated on isHost. Clears CompanionRoomCode + CompanionLocked on the
// eGame, broadcasts a "mode-changed" WebSocket message, and clears the
// surface=table cookie on the response. The client-side handler
// (boardgame-game-state-manager.ts) reloads on receipt of mode-changed;
// the reload's HTTP response carries the cookie-clear and the loader
// falls back to the solo renderer.
//
// Spec §9.6: this is a one-way mode downgrade and a deliberate UX hazard
// for hidden-info games. The Table view's two-tap confirm() (P5.3) is
// the user-facing safety; the server doesn't second-guess.
func (s *Server) switchToSoloHandler(c *gin.Context) {
	game := s.getGame(c)
	if game == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no such game"})
		return
	}
	gameID := game.ID()

	hostUserID, err := s.resolveHost(gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve host: " + err.Error()})
		return
	}
	if !s.isHost(c, gameID, hostUserID) {
		s.auditHostAction("switchToSolo", gameID, "", "not host", false)
		c.JSON(http.StatusForbidden, gin.H{"error": "host privileges required"})
		return
	}

	userID := hostUserID
	if !hostActionAllowed(gameID, userID) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "host actions are rate-limited to 1/sec"})
		return
	}
	joinLock := s.getSeatJoinLock(gameID)
	joinLock.Lock()
	defer joinLock.Unlock()

	eGame, err := s.storage.ExtendedGame(gameID)
	if err != nil || eGame == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no such game"})
		return
	}
	updatedGame := *eGame
	updatedGame.CompanionRoomCode = ""
	updatedGame.CompanionLocked = false

	// Fence and invalidate every pending transfer before changing the separate
	// game metadata. Redemption and this CAS race on one durable row: exactly
	// one can win, so a successful receiver can never subsequently be switched
	// out from underneath by the stale source Table.
	credential, cookieErr := c.Cookie(tableLeaseCookieName(gameID))
	if cookieErr != nil {
		tableLeaseProblem(c, http.StatusConflict, "TABLE_LEASE_LOST", "this screen could not safely confirm Table control", nil)
		return
	}
	transitioned := false
	for attempts := 0; attempts < 5; attempts++ {
		current, leaseErr := s.storage.CompanionTableLease(gameID)
		if leaseErr != nil {
			tableLeaseProblem(c, http.StatusServiceUnavailable, "TABLE_LEASE_STORAGE", "this screen could not safely confirm Table control", nil)
			return
		}
		if !tableLeaseActive(current, time.Now()) || !tableLeaseCredentialMatches(current, credential) ||
			current.TransitionKind == tablelease.TransitionSolo || current.TransitionKind == tablelease.TransitionHostAction {
			tableLeaseProblem(c, http.StatusConflict, "TABLE_LEASE_LOST", "this screen could not safely confirm Table control", nil)
			return
		}
		replacement := current.Clone()
		replacement.ClearTransfer()
		replacement.PreviousDeviceID = current.DeviceID
		replacement.TransitionKind = tablelease.TransitionSolo
		replacement.Expires = time.Now().Add(tableLeaseTTL).UnixMilli()
		_, swapped, leaseErr := s.storage.CompareAndSwapCompanionTableLease(gameID, current.Generation, replacement)
		if leaseErr != nil {
			tableLeaseProblem(c, http.StatusServiceUnavailable, "TABLE_LEASE_STORAGE", "this screen could not safely confirm Table control", nil)
			return
		}
		if swapped {
			transitioned = true
			break
		}
	}
	if !transitioned {
		tableLeaseProblem(c, http.StatusConflict, "TABLE_LEASE_LOST", "another screen changed Table control", nil)
		return
	}
	if err := s.storage.UpdateExtendedGame(gameID, &updatedGame); err != nil {
		// Best-effort rollback keeps the old capability usable if the metadata
		// write itself failed. CAS protects a newer transition from rollback.
		if current, leaseErr := s.storage.CompanionTableLease(gameID); leaseErr == nil && current != nil && current.TransitionKind == tablelease.TransitionSolo && tableLeaseCredentialMatches(current, credential) {
			replacement := current.Clone()
			replacement.PreviousDeviceID = ""
			replacement.TransitionKind = ""
			_, _, _ = s.storage.CompareAndSwapCompanionTableLease(gameID, current.Generation, replacement)
		}
		s.auditHostAction("switchToSolo", gameID, userID, err.Error(), false)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update game: " + err.Error()})
		return
	}
	s.releaseTableLease(gameID, credential)
	clearTableLeaseCookie(c, gameID)

	// Clear the surface cookie for this game on the host's browser.
	// Max-Age=-1 (or 0 in some browsers) triggers immediate expiration.
	c.SetCookie(surfaceCookieName(gameID), "", -1, "/", "", false, false)

	// Broadcast mode-changed to all connected sockets so phones reload
	// into the solo renderer. Their surface=hand cookie remains stale
	// but the next state fetch (after reload) will have eGame.
	// CompanionRoomCode=="" → loader falls back to solo renderer; the
	// stale cookie no-ops there.
	s.notifier.broadcastModeChanged(gameID, "solo")

	s.auditHostAction("switchToSolo", gameID, userID, "switched to solo", true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// setRoomLockHandler implements POST /api/game/:name/:id/setRoomLock.
// Body: { "locked": true|false }. Flips eGame.CompanionLocked. Host-only;
// rate-limited; audited.
func (s *Server) setRoomLockHandler(c *gin.Context) {
	game := s.getGame(c)
	if game == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no such game"})
		return
	}
	gameID := game.ID()

	hostUserID, err := s.resolveHost(gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve host: " + err.Error()})
		return
	}
	if !s.isHost(c, gameID, hostUserID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "host privileges required"})
		return
	}
	var body struct {
		Locked *bool `json:"locked"`
	}
	if err := decodeStrictJSON(c, &body); err != nil || body.Locked == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body must contain exactly one boolean locked field"})
		return
	}

	userID := hostUserID
	if !hostActionAllowed(gameID, userID) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "host actions are rate-limited to 1/sec"})
		return
	}
	joinLock := s.getSeatJoinLock(gameID)
	joinLock.Lock()
	defer joinLock.Unlock()
	fence, renewal := s.beginTableLeaseAction(c, gameID)
	if renewal != tableLeaseRenewed {
		status := http.StatusConflict
		if renewal == tableLeaseRenewRetryable {
			status = http.StatusServiceUnavailable
		}
		tableLeaseProblem(c, status, "TABLE_LEASE_LOST", "this screen could not safely reserve Table control", nil)
		return
	}
	defer s.endTableLeaseAction(gameID, fence)

	eGame, err := s.storage.ExtendedGame(gameID)
	if err != nil || eGame == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no such game"})
		return
	}

	updatedGame := *eGame
	updatedGame.CompanionLocked = *body.Locked
	if err := s.storage.UpdateExtendedGame(gameID, &updatedGame); err != nil {
		s.auditHostAction("setRoomLock", gameID, userID, err.Error(), false)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update game: " + err.Error()})
		return
	}

	s.auditHostAction("setRoomLock", gameID, userID, "locked="+strconv.FormatBool(*body.Locked), true)
	c.JSON(http.StatusOK, gin.H{"ok": true, "locked": *body.Locked})
}
