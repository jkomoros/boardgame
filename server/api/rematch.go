package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/seatpresentation"
	"github.com/jkomoros/boardgame/server/api/tablelease"
	"github.com/jkomoros/boardgame/server/api/users"
)

type rematchFence struct {
	deviceID               string
	secretDigest           string
	previousDeviceID       string
	previousTransitionKind string
}

// rematchAuthority accepts either the game owner or the exact last Table
// capability. The latter intentionally ignores expiry: finished games stop
// renewing sockets, but the fenced, HttpOnly capability must still let an
// accountless projector start or follow its rematch hours later.
func (s *Server) rematchAuthority(c *gin.Context, gameID string, info *extendedgame.StorageRecord, user *users.StorageRecord) (owner, table bool) {
	owner = user != nil && info != nil && info.Owner != "" && user.ID == info.Owner
	if !tableSurfaceForRequest(c, gameID) {
		return owner, false
	}
	credential, err := c.Cookie(tableLeaseCookieName(gameID))
	if err != nil {
		return owner, false
	}
	lease, err := s.storage.CompanionTableLease(gameID)
	return owner, err == nil && tableLeaseCredentialMatches(lease, credential)
}

// beginRematchFence uses the existing cross-process Table CAS row as the
// durable mutex for one successor. Unlike ordinary live host actions, expiry
// does not revoke the last Table's rematch capability after game over.
func (s *Server) beginRematchFence(c *gin.Context, gameID string, ownerAuthorized bool) (*rematchFence, tableLeaseRenewal) {
	credential, _ := c.Cookie(tableLeaseCookieName(gameID))
	for attempts := 0; attempts < 5; attempts++ {
		now := time.Now()
		current, err := s.storage.CompanionTableLease(gameID)
		if err != nil || current == nil {
			return nil, tableLeaseRenewRetryable
		}
		if !ownerAuthorized && !tableLeaseCredentialMatches(current, credential) {
			return nil, tableLeaseRenewLost
		}
		if current.TransitionKind == tablelease.TransitionSolo {
			return nil, tableLeaseRenewLost
		}
		if current.TransitionKind == tablelease.TransitionHostAction {
			if current.Expires > now.UnixMilli() {
				return nil, tableLeaseRenewRetryable
			}
			// The creator died while holding the fence. Finished games have no
			// live transfer or heartbeat to recover it for us, so expire the
			// marker explicitly and retry from the durable rematch allocation.
			recovered := current.Clone()
			recovered.ClearTransfer()
			recovered.PreviousDeviceID = ""
			recovered.TransitionKind = ""
			_, _, recoverErr := s.storage.CompareAndSwapCompanionTableLease(gameID, current.Generation, recovered)
			if recoverErr != nil {
				return nil, tableLeaseRenewRetryable
			}
			continue
		}
		fence := &rematchFence{
			deviceID: current.DeviceID, secretDigest: current.SecretDigest,
			previousDeviceID:       current.PreviousDeviceID,
			previousTransitionKind: current.TransitionKind,
		}
		replacement := current.Clone()
		if replacement.PreviousDeviceID == "" {
			replacement.PreviousDeviceID = current.DeviceID
		}
		replacement.TransitionKind = tablelease.TransitionHostAction
		replacement.Expires = now.Add(tableLeaseTTL).UnixMilli()
		_, swapped, err := s.storage.CompareAndSwapCompanionTableLease(gameID, current.Generation, replacement)
		if err != nil {
			return nil, tableLeaseRenewRetryable
		}
		if swapped {
			return fence, tableLeaseRenewed
		}
	}
	return nil, tableLeaseRenewRetryable
}

func (s *Server) endRematchFence(gameID string, fence *rematchFence) bool {
	if fence == nil {
		return false
	}
	for attempts := 0; attempts < 5; attempts++ {
		current, err := s.storage.CompanionTableLease(gameID)
		if err != nil || current == nil || current.TransitionKind != tablelease.TransitionHostAction ||
			current.DeviceID != fence.deviceID || current.SecretDigest != fence.secretDigest {
			return false
		}
		replacement := current.Clone()
		replacement.PreviousDeviceID = fence.previousDeviceID
		replacement.TransitionKind = fence.previousTransitionKind
		_, swapped, err := s.storage.CompareAndSwapCompanionTableLease(gameID, current.Generation, replacement)
		if err != nil {
			return false
		}
		if swapped {
			return true
		}
	}
	return false
}

func rematchProblem(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"ok": false, "code": code, "error": message})
}

func (s *Server) rematchResponse(c *gin.Context, oldGameID string, oldTable bool, game *boardgame.Game) {
	newInfo, err := s.storage.ExtendedGame(game.ID())
	if err != nil || newInfo == nil || newInfo.CompanionRoomCode == "" {
		rematchProblem(c, http.StatusServiceUnavailable, "REMATCH_NOT_READY", "the rematch is still being prepared")
		return
	}
	if oldTable {
		oldLease, leaseErr := s.storage.CompanionTableLease(oldGameID)
		if leaseErr != nil || oldLease == nil || oldLease.DeviceID == "" {
			rematchProblem(c, http.StatusServiceUnavailable, "TABLE_LEASE_UNAVAILABLE", "could not carry shared-screen control into the rematch")
			return
		}
		secret, _, secretErr := s.tableLeaseCredentialForDevice(game.ID(), oldLease.DeviceID)
		if secretErr != nil {
			rematchProblem(c, http.StatusInternalServerError, "TABLE_LEASE_UNAVAILABLE", "could not carry shared-screen control into the rematch")
			return
		}
		s.setSurfaceCookie(c, game.ID(), "table")
		s.setTableLeaseCookie(c, game.ID(), oldLease.DeviceID, secret)
	}
	c.JSON(http.StatusOK, gin.H{
		"ok": true, "gameID": game.ID(), "gameName": game.Name(),
		"roomCode": newInfo.CompanionRoomCode,
	})
}

// rematchHandler creates exactly one successor for a finished companion game,
// preserving configuration, agents, human seat bindings, and presentations.
// Repeated calls resume partial setup and return the same successor.
func (s *Server) rematchHandler(c *gin.Context) {
	oldGame := s.getGame(c)
	if oldGame == nil {
		rematchProblem(c, http.StatusNotFound, "GAME_NOT_FOUND", "no such game")
		return
	}
	oldCombined, err := s.storage.CombinedGame(oldGame.ID())
	if err != nil || oldCombined == nil || oldCombined.CompanionRoomCode == "" {
		rematchProblem(c, http.StatusConflict, "NOT_COMPANION_GAME", "only companion games can keep the room together")
		return
	}
	if !oldCombined.Finished {
		rematchProblem(c, http.StatusConflict, "GAME_NOT_FINISHED", "finish this game before starting a rematch")
		return
	}
	ownerAuthorized, tableAuthorized := s.rematchAuthority(c, oldGame.ID(), &oldCombined.StorageRecord, s.getUser(c))
	if !ownerAuthorized && !tableAuthorized {
		rematchProblem(c, http.StatusForbidden, "REMATCH_FORBIDDEN", "only the game owner or shared Table can start the rematch")
		return
	}

	if oldCombined.RematchReady && oldCombined.RematchGameID != "" {
		if existing := oldGame.Manager().Game(oldCombined.RematchGameID); existing != nil {
			s.rematchResponse(c, oldGame.ID(), tableAuthorized, existing)
			return
		}
		rematchProblem(c, http.StatusServiceUnavailable, "REMATCH_NOT_READY", "the rematch is temporarily unavailable")
		return
	}

	fence, renewal := s.beginRematchFence(c, oldGame.ID(), ownerAuthorized)
	if renewal != tableLeaseRenewed {
		status := http.StatusConflict
		if renewal == tableLeaseRenewRetryable {
			status = http.StatusServiceUnavailable
		}
		rematchProblem(c, status, "REMATCH_BUSY", "another screen is preparing the rematch; retry shortly")
		return
	}
	defer s.endRematchFence(oldGame.ID(), fence)

	// Re-read behind the cross-process fence. A racing successful request wins.
	oldCombined, err = s.storage.CombinedGame(oldGame.ID())
	if err != nil || oldCombined == nil {
		rematchProblem(c, http.StatusInternalServerError, "REMATCH_FAILED", "could not reload the finished game")
		return
	}
	if oldCombined.RematchGameID == "" {
		var targetID string
		for attempts := 0; attempts < 8; attempts++ {
			candidate, idErr := randomHex(8)
			if idErr != nil {
				break
			}
			if oldGame.Manager().Game(candidate) == nil {
				targetID = candidate
				break
			}
		}
		secretSalt, saltErr := randomHex(8)
		if targetID == "" || saltErr != nil {
			rematchProblem(c, http.StatusInternalServerError, "REMATCH_FAILED", "could not allocate the rematch")
			return
		}
		oldCombined.RematchGameID = targetID
		oldCombined.RematchReady = false
		if err := s.storage.UpdateExtendedGame(oldGame.ID(), &oldCombined.StorageRecord); err != nil {
			rematchProblem(c, http.StatusInternalServerError, "REMATCH_FAILED", "could not reserve the rematch")
			return
		}
		// SecretSalt is not part of extended metadata. Create immediately while
		// it is in hand; a failure before persistence leaves no published target.
		record := &boardgame.GameStorageRecord{
			ID: targetID, SecretSalt: secretSalt, NumPlayers: oldCombined.NumPlayers,
			Variant: oldCombined.Variant, Agents: oldCombined.Agents,
		}
		if _, err := oldGame.Manager().Internals().RecreateGame(record); err != nil {
			rematchProblem(c, http.StatusInternalServerError, "REMATCH_FAILED", "could not create the rematch: "+err.Error())
			return
		}
	}

	newGame := oldGame.Manager().Game(oldCombined.RematchGameID)
	if newGame == nil {
		// The reserved target may outlive a process crash before creation. Its
		// salt is unavailable, so clear only the unpublished allocation and let
		// the next retry reserve a fresh target rather than exposing a dead link.
		oldCombined.RematchGameID = ""
		oldCombined.RematchReady = false
		_ = s.storage.UpdateExtendedGame(oldGame.ID(), &oldCombined.StorageRecord)
		rematchProblem(c, http.StatusServiceUnavailable, "REMATCH_NOT_READY", "the rematch allocation was recovered; retry")
		return
	}

	newInfo, err := s.storage.ExtendedGame(newGame.ID())
	if err != nil || newInfo == nil {
		rematchProblem(c, http.StatusInternalServerError, "REMATCH_FAILED", "could not configure the rematch")
		return
	}
	newInfo.Owner = oldCombined.Owner
	// Keep the successor out of lists and ordinary join paths until its Table
	// capability and every existing seat are installed. RematchReady is the
	// client publication barrier; these fields are the server admission barrier.
	newInfo.Open = false
	newInfo.Visible = false
	newInfo.CompanionLocked = false
	if newInfo.CompanionRoomCode == "" {
		code, codeErr := GenerateRoomCode(func(candidate string) (bool, error) {
			id, lookupErr := s.storage.GameByRoomCode(candidate)
			return id != "", lookupErr
		})
		if codeErr != nil {
			rematchProblem(c, http.StatusServiceUnavailable, "REMATCH_FAILED", "could not allocate a new room code")
			return
		}
		newInfo.CompanionRoomCode = code
	}
	if err := s.storage.UpdateExtendedGame(newGame.ID(), newInfo); err != nil {
		rematchProblem(c, http.StatusInternalServerError, "REMATCH_FAILED", "could not save rematch metadata")
		return
	}

	oldLease, err := s.storage.CompanionTableLease(oldGame.ID())
	if err != nil || oldLease == nil || oldLease.DeviceID == "" {
		rematchProblem(c, http.StatusInternalServerError, "REMATCH_FAILED", "could not preserve shared-screen identity")
		return
	}
	secret, digest, err := s.tableLeaseCredentialForDevice(newGame.ID(), oldLease.DeviceID)
	if err != nil {
		rematchProblem(c, http.StatusInternalServerError, "REMATCH_FAILED", "could not preserve shared-screen credentials")
		return
	}
	newLease, err := s.storage.CompanionTableLease(newGame.ID())
	if err != nil {
		rematchProblem(c, http.StatusInternalServerError, "REMATCH_FAILED", "could not load rematch shared-screen state")
		return
	}
	if newLease == nil || newLease.Generation == 0 {
		_, swapped, leaseErr := s.storage.CompareAndSwapCompanionTableLease(newGame.ID(), 0, &tablelease.StorageRecord{
			DeviceID: oldLease.DeviceID, SecretDigest: digest,
			HolderUserID: oldLease.HolderUserID, Expires: time.Now().Add(tableLeaseTTL).UnixMilli(),
		})
		if leaseErr != nil || !swapped {
			rematchProblem(c, http.StatusServiceUnavailable, "REMATCH_FAILED", "could not reserve the rematch shared screen")
			return
		}
	} else if newLease.DeviceID != oldLease.DeviceID || newLease.SecretDigest != digest {
		rematchProblem(c, http.StatusConflict, "REMATCH_FAILED", "the rematch shared screen was claimed inconsistently")
		return
	}

	oldUserIDs := s.storage.UserIDsForGame(oldGame.ID())
	newUserIDs := s.storage.UserIDsForGame(newGame.ID())
	for i, userID := range oldUserIDs {
		if userID == "" {
			continue
		}
		if i >= len(newUserIDs) {
			rematchProblem(c, http.StatusInternalServerError, "REMATCH_FAILED", "rematch seat count changed")
			return
		}
		if newUserIDs[i] == "" {
			user := s.storage.GetUserByID(userID)
			if user == nil {
				rematchProblem(c, http.StatusInternalServerError, "REMATCH_FAILED", "a rematch player identity is missing")
				return
			}
			if err := s.doSeatPlayer(newGame, boardgame.PlayerIndex(i), user); err != nil {
				rematchProblem(c, http.StatusServiceUnavailable, "REMATCH_FAILED", "could not preserve a player seat")
				return
			}
		} else if newUserIDs[i] != userID {
			rematchProblem(c, http.StatusConflict, "REMATCH_FAILED", "a rematch seat was claimed inconsistently")
			return
		} else if i >= len(newGame.CurrentState().ImmutablePlayerStates()) {
			rematchProblem(c, http.StatusInternalServerError, "REMATCH_FAILED", "rematch seat count changed")
			return
		} else if seat, ok := newGame.CurrentState().ImmutablePlayerStates()[i].(interfaces.Seater); ok && !seat.SeatIsFilled() {
			if err := s.forceSeatPlayer(newGame, boardgame.PlayerIndex(i)); err != nil {
				rematchProblem(c, http.StatusServiceUnavailable, "REMATCH_FAILED", "could not recover a pending player seat")
				return
			}
		}
		presentation, presentationErr := s.storage.SeatPresentation(oldGame.ID(), boardgame.PlayerIndex(i))
		if presentationErr != nil {
			rematchProblem(c, http.StatusInternalServerError, "REMATCH_FAILED", "could not copy player presentation")
			return
		}
		if presentation != nil {
			if err := s.storage.SetSeatPresentation(&seatpresentation.StorageRecord{
				GameID: newGame.ID(), PlayerIndex: boardgame.PlayerIndex(i),
				DisplayName: presentation.DisplayName, AvatarSlug: presentation.AvatarSlug,
			}); err != nil {
				rematchProblem(c, http.StatusInternalServerError, "REMATCH_FAILED", "could not copy player presentation")
				return
			}
		}
	}

	// Restore the finished room's policy only after the successor is complete.
	// A full predecessor is normally closed; a partially filled open room stays
	// open so its new room code can still admit players.
	newInfo.Open = oldCombined.Open
	newInfo.Visible = oldCombined.Visible
	if err := s.storage.UpdateExtendedGame(newGame.ID(), newInfo); err != nil {
		rematchProblem(c, http.StatusServiceUnavailable, "REMATCH_FAILED", "the rematch is ready but its room policy could not be restored")
		return
	}

	oldCombined.RematchReady = true
	if err := s.storage.UpdateExtendedGame(oldGame.ID(), &oldCombined.StorageRecord); err != nil {
		rematchProblem(c, http.StatusServiceUnavailable, "REMATCH_FAILED", "the rematch is ready but could not be published")
		return
	}
	// Rematch publication is metadata-only and does not advance the finished
	// game's version. Nudge every still-open Hand immediately; their ordinary
	// authoritative info refresh then discovers the durable successor. The
	// client poll remains the recovery path for a dropped/local-only broadcast.
	s.notifier.enqueuePresenceChange(oldGame.ID())
	if tableAuthorized {
		s.setSurfaceCookie(c, newGame.ID(), "table")
		s.setTableLeaseCookie(c, newGame.ID(), oldLease.DeviceID, secret)
	}
	s.rematchResponse(c, oldGame.ID(), tableAuthorized, newGame)
}
