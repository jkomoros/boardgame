package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame/behaviors"
)

// seatOptionsSlot is one row in the seat-options payload returned to the
// phone when it's deciding which slot to claim in an asymmetric game.
type seatOptionsSlot = companionSeatStatus

type seatOptionsResponse struct {
	GameID             string            `json:"gameID"`
	GameName           string            `json:"gameName"`
	Slots              []seatOptionsSlot `json:"slots"`
	RequiresSeatPicker bool              `json:"requiresSeatPicker"`
}

// joinSeatOptionsHandler implements GET /api/join/seat-options?gameID=<>.
// Authenticated (Firebase token); the response is intentionally NOT
// returned from the unauthenticated /api/join so a brute-force scrape can't
// reveal a game's asymmetric structure (spec §6.1 security note).
//
// For each player slot in the game's NumPlayers, returns:
//   - PlayerIndex
//   - Label: "Seat N" for private-role games (the default); the actual
//     role enum display name for public-role games (i.e. games that
//     overrode the new default with `sanitize:"all:visible"` at the
//     embedding site). Detection: the playerState satisfies
//     behaviors.HasPlayerRole AND the Role property survives sanitization
//     to ObserverPlayerIndex.
//   - Filled + AvatarSlug + DisplayName for already-seated slots.
//
// requiresSeatPicker is true iff the game has any asymmetric behavior
// (HasPlayerRole OR HasPlayerTeam). Symmetric games can short-circuit
// the picker on the phone — they'll auto-assign in /api/join/seat.
func (s *Server) joinSeatOptionsHandler(c *gin.Context) {
	gameID := c.Query("gameID")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gameID query param is required"})
		return
	}
	if !s.requireJoinTicket(c, gameID) {
		return
	}

	// Auth: Firebase token in Authorization header (not query string —
	// JWTs in URLs leak via logs, browser history, and Referer headers).
	if !s.config.OfflineDevMode {
		token := c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		if token == "" {
			writeJoinProblem(c, http.StatusUnauthorized, "AUTH_REQUIRED", "Sign in again to continue", nil)
			return
		}
		_, err := verifyFirebaseTokenWithTimeout(
			c.Request.Context(), s.firebaseAuth, token, firebaseVerifyTimeout,
		)
		if err != nil {
			if errors.Is(err, errFirebaseVerifyTimeout) {
				writeJoinProblem(c, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "Sign-in verification is temporarily unavailable; retry", nil)
			} else {
				writeJoinProblem(c, http.StatusUnauthorized, "AUTH_INVALID", "Your sign-in expired; sign in again", nil)
			}
			return
		}
	}
	joinLock := s.getSeatJoinLock(gameID)
	joinLock.Lock()
	defer joinLock.Unlock()

	combined, err := s.storage.CombinedGame(gameID)
	if err != nil || combined == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}
	if combined.CompanionRoomCode == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}
	if combined.Finished {
		writeJoinProblem(c, http.StatusConflict, joinCodeRoomFinished, "This game has finished", nil)
		return
	}
	if combined.CompanionLocked {
		writeJoinProblem(c, http.StatusConflict, joinCodeRoomLocked, "The host locked this room", nil)
		return
	}

	mgrInfo := s.managers[combined.Name]
	if mgrInfo == nil || mgrInfo.manager == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	requiresPicker := s.requiresSeatPicker(mgrInfo)
	game := mgrInfo.manager.Game(gameID)
	if game == nil {
		writeJoinProblem(c, http.StatusNotFound, "ROOM_NOT_FOUND", "Room not found", nil)
		return
	}
	slots := s.companionSeatStatuses(game)

	c.JSON(http.StatusOK, seatOptionsResponse{
		GameID:             gameID,
		GameName:           combined.Name,
		Slots:              slots,
		RequiresSeatPicker: requiresPicker,
	})
}

// requiresSeatPicker reports whether the manager's playerState is asymmetric
// (satisfies behaviors.HasPlayerRole or behaviors.HasPlayerTeam), meaning
// phones should show the seat picker before claiming a seat. Symmetric games
// auto-assign in /api/join/seat.
func (s *Server) requiresSeatPicker(mgrInfo *managerInfo) bool {
	if mgrInfo == nil || mgrInfo.manager == nil {
		return false
	}
	examplePlayerStates := mgrInfo.manager.ExampleState().ImmutablePlayerStates()
	if len(examplePlayerStates) == 0 {
		return false
	}
	if _, ok := examplePlayerStates[0].(behaviors.HasPlayerRole); ok {
		return true
	}
	if _, ok := examplePlayerStates[0].(behaviors.HasPlayerTeam); ok {
		return true
	}
	return false
}
