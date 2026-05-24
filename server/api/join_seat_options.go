package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
)

// seatOptionsSlot is one row in the seat-options payload returned to the
// phone when it's deciding which slot to claim in an asymmetric game.
type seatOptionsSlot struct {
	PlayerIndex int    `json:"playerIndex"`
	Label       string `json:"label"`
	Filled      bool   `json:"filled"`
	// AvatarSlug is populated for filled slots so the phone can render
	// who's already in each seat. Empty for unfilled slots.
	AvatarSlug string `json:"avatarSlug,omitempty"`
	// DisplayName is populated for filled slots, same reason. Empty for
	// unfilled slots.
	DisplayName string `json:"displayName,omitempty"`
}

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

	// Auth: Firebase token in Authorization header (not query string —
	// JWTs in URLs leak via logs, browser history, and Referer headers).
	if !s.config.OfflineDevMode {
		token := c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required (pass token in Authorization: Bearer header)"})
			return
		}
		_, err := verifyFirebaseTokenWithTimeout(token, s.config.Firebase.ProjectID, firebaseVerifyTimeout)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "auth failed"})
			return
		}
	}

	combined, err := s.storage.CombinedGame(gameID)
	if err != nil || combined == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}
	if combined.Finished || combined.CompanionLocked || combined.CompanionRoomCode == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	mgrInfo := s.managers[combined.Name]
	if mgrInfo == nil || mgrInfo.manager == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Asymmetry detection: inspect the example player state on the manager.
	examplePlayerStates := mgrInfo.manager.ExampleState().ImmutablePlayerStates()
	var hasRole, hasTeam bool
	if len(examplePlayerStates) > 0 {
		if _, ok := examplePlayerStates[0].(behaviors.HasPlayerRole); ok {
			hasRole = true
		}
		if _, ok := examplePlayerStates[0].(behaviors.HasPlayerTeam); ok {
			hasTeam = true
		}
	}

	userIDs := s.storage.UserIDsForGame(gameID)
	numPlayers := combined.NumPlayers
	slots := make([]seatOptionsSlot, 0, numPlayers)
	for i := 0; i < numPlayers; i++ {
		slot := seatOptionsSlot{
			PlayerIndex: i,
			Label:       "Seat " + strconv.Itoa(i+1),
		}
		if i < len(userIDs) && userIDs[i] != "" {
			slot.Filled = true
			if pres, err := s.storage.SeatPresentation(gameID, boardgame.PlayerIndex(i)); err == nil && pres != nil {
				slot.AvatarSlug = pres.AvatarSlug
				slot.DisplayName = pres.DisplayName
			}
		}
		// Public-role / public-team label resolution is wired by sanitizing
		// the live game state for ObserverPlayerIndex and reading the
		// Role/Team property if visible. For V1 we leave that as a future
		// enhancement once a public-role game ships — most games are
		// hidden-role, so "Seat N" is the right default. The asymmetry
		// detection (above) is still useful because it lets the phone know
		// whether to show the picker at all.
		slots = append(slots, slot)
	}

	c.JSON(http.StatusOK, seatOptionsResponse{
		GameID:             gameID,
		GameName:           combined.Name,
		Slots:              slots,
		RequiresSeatPicker: hasRole || hasTeam,
	})
}
