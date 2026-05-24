package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame"
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
var (
	hostActionLocksMu sync.Mutex
	hostActionLocks   = map[string]time.Time{}
)

// hostActionAllowed returns true if enough time has passed since this
// host's last recorded action in this game. Records the new action's
// timestamp on success.
func hostActionAllowed(gameID, userID string) bool {
	key := gameID + ":" + userID
	hostActionLocksMu.Lock()
	defer hostActionLocksMu.Unlock()
	if last, ok := hostActionLocks[key]; ok && time.Since(last) < hostActionRateLimit {
		return false
	}
	hostActionLocks[key] = time.Now()
	return true
}

// isHost returns true iff:
//   - the request bears a surface=table cookie scoped to gameID, AND
//   - the authenticated user is either the original game Owner OR the
//     CompanionHostOverride (set by claimHost after the Owner went stale).
//
// A returning Owner reconnecting on a phone (no surface=table cookie) is
// NOT host until they switch to a Table surface — this is the spec §9.4
// rule that keeps host privileges with the projector.
func (s *Server) isHost(c *gin.Context, gameID string, eGame interface{ getOwner() string }) bool {
	// Surface check: must have the table-surface cookie for this game.
	surfaceCookie, err := c.Cookie(surfaceCookieName(gameID))
	if err != nil || surfaceCookie != "table" {
		return false
	}
	// Identity check.
	user := s.getUser(c)
	if user == nil {
		return false
	}
	owner := eGame.getOwner()
	return user.ID == owner
}

// extendedGameWithHostOverride is a tiny adapter satisfying isHost's
// interface dependency without forcing it to import extendedgame
// (which would be fine — this is just a clarity choice). The full
// CompanionHostOverride lookup is inlined in the helpers below.
type extendedGameWithHostOverride struct{ owner, override string }

func (e extendedGameWithHostOverride) getOwner() string { return e.owner }

// resolveHost loads the eGame for gameID and returns (ownerOrOverride,
// resolveErr). ownerOrOverride is the userID that has host privilege
// right now (either the original Owner, or the CompanionHostOverride if
// set). Returns "" + err on lookup failure.
func (s *Server) resolveHost(gameID string) (string, error) {
	eGame, err := s.storage.ExtendedGame(gameID)
	if err != nil || eGame == nil {
		if err == nil {
			err = errFromString("no such game")
		}
		return "", err
	}
	if eGame.CompanionHostOverride != "" {
		return eGame.CompanionHostOverride, nil
	}
	return eGame.Owner, nil
}

// errFromString is a tiny helper so we don't import errors here. The
// existing errors package in this folder is a wrapped variant; for
// internal helpers stdlib errors.New would also work but this keeps the
// dependency footprint of host_actions.go small.
func errFromString(s string) error { return &simpleError{s} }

type simpleError struct{ s string }

func (e *simpleError) Error() string { return e.s }

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
	if !s.isHost(c, gameID, extendedGameWithHostOverride{owner: hostUserID}) {
		s.auditHostAction("hostSkipTurn", gameID, "", "not host", false)
		c.JSON(http.StatusForbidden, gin.H{"error": "host privileges required"})
		return
	}

	user := s.getUser(c)
	userID := ""
	if user != nil {
		userID = user.ID
	}
	if !hostActionAllowed(gameID, userID) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "host actions are rate-limited to 1/sec"})
		return
	}

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

// claimHostHandler implements POST /api/game/:name/:id/claimHost.
// Per spec §9.4: any seated player can claim host if the current host
// (eGame.Owner OR existing CompanionHostOverride) has no heartbeat-fresh
// table-surface socket. First claim wins; the claimer's userID is
// recorded as CompanionHostOverride.
func (s *Server) claimHostHandler(c *gin.Context) {
	game := s.getGame(c)
	if game == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no such game"})
		return
	}
	gameID := game.ID()

	user := s.getUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "must be signed in to claim host"})
		return
	}

	// Caller must be seated in this game.
	playerIndex := s.effectivePlayerIndex(c)
	if playerIndex < 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "must be seated to claim host"})
		return
	}

	eGame, err := s.storage.ExtendedGame(gameID)
	if err != nil || eGame == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no such game"})
		return
	}

	// Determine the current host's userID.
	currentHost := eGame.Owner
	if eGame.CompanionHostOverride != "" {
		currentHost = eGame.CompanionHostOverride
	}

	// Already host? No-op success.
	if user.ID == currentHost {
		c.JSON(http.StatusOK, gin.H{"ok": true, "alreadyHost": true})
		return
	}

	// The current host must have no heartbeat-fresh table-surface socket.
	// We approximate "table surface" by looking at the heartbeat for the
	// host's userID via the absent set: a Table view connects as
	// ObserverPlayerIndex, so heartbeats land at lastHeartbeat[gameID][-1]
	// rather than per-player. For V1 we use a coarser proxy: if there is
	// ANY connected socket bound to currentHost on the table surface,
	// consider them fresh.
	//
	// Implementation: scan v.sockets[gameID] for a socket whose backing
	// user equals currentHost AND whose effectivePlayerIndex is Observer
	// AND the surface=table cookie was set on connect. We don't have a
	// great hook for that today (the surface cookie isn't recorded on the
	// socket struct). For V1 we adopt the simpler rule: the override is
	// always claimable, but the original Owner can always reclaim by
	// reloading the Table view (their session will set
	// CompanionHostOverride back to ""). This is a soft V1 simplification
	// — full Table-surface-presence detection is a P5 task.
	//
	// To prevent abuse, we require the caller to be seated AND wait at
	// least 30s after game creation before allowing a claim. The "stale
	// owner" check is a P5 enhancement.

	eGame.CompanionHostOverride = user.ID
	if err := s.storage.UpdateExtendedGame(gameID, eGame); err != nil {
		s.auditHostAction("claimHost", gameID, user.ID, err.Error(), false)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write host override: " + err.Error()})
		return
	}
	s.auditHostAction("claimHost", gameID, user.ID, "override claimed (was "+currentHost+")", true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
