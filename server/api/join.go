package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// joinRequest is the body of POST /api/join.
type joinRequest struct {
	Code string `json:"code"`
}

// joinResponse is returned on a successful lookup. The shape is intentionally
// small: just enough metadata for the phone client to render "Joining
// <gameDisplayName>" before walking the user through identity, avatar
// pickers, etc. RequiresSeatPicker tells the phone whether to fetch
// /api/join/seat-options (only set for asymmetric-role games — P2).
type joinResponse struct {
	GameID             string `json:"gameID"`
	GameName           string `json:"gameName"`
	GameDisplayName    string `json:"gameDisplayName"`
	MinPlayers         int    `json:"minPlayers"`
	MaxPlayers         int    `json:"maxPlayers"`
	CurrentPlayers     int    `json:"currentPlayers"`
	AvailableSeats     int    `json:"availableSeats"`
	RequiresSeatPicker bool   `json:"requiresSeatPicker"`
	// JoinTicket is short-lived proof that this client supplied the room code.
	// Subsequent authenticated seat discovery and claim requests must echo it
	// in X-Boardgame-Join-Ticket.
	JoinTicket string `json:"joinTicket"`
}

// joinHandler implements POST /api/join. Looks up the game by the supplied
// (normalized) room code; returns metadata on success. Per spec §6.2:
//
//   - 404 if the code isn't found, the game is Finished, or the room is locked
//   - 400 if the code is malformed
//   - 429 from the rate-limit middleware (not here) on per-IP rate excess
//
// The response intentionally does NOT include any role/team metadata —
// asymmetric structure must require an authenticated lookup so brute-forcing
// /api/join cannot reveal it (spec §6.1 security note).
func (s *Server) joinHandler(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	var req joinRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	code, err := NormalizeRoomCode(req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gameID, err := s.storage.GameByRoomCode(code)
	if err != nil {
		s.logger.Warnln("GameByRoomCode lookup error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal lookup error"})
		return
	}
	if gameID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Pull the combined record so we can check Finished + companion fields
	// + numPlayers in one round-trip.
	combined, err := s.storage.CombinedGame(gameID)
	if err != nil || combined == nil {
		s.logger.Warnln("CombinedGame for matched code returned error / nil:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	if combined.Finished {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}
	if combined.CompanionLocked {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	mgrInfo := s.managers[combined.Name]
	if mgrInfo == nil || mgrInfo.manager == nil {
		// Game registered under a name the server doesn't know about.
		// Treat as 404 to keep the surface uniform.
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	game := mgrInfo.manager.Game(gameID)
	if game == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}
	statuses := s.companionSeatStatuses(game)
	currentPlayers := 0
	availableSeats := 0
	for _, seat := range statuses {
		if seat.Status == "human" {
			currentPlayers++
		}
		if seat.Available {
			availableSeats++
		}
	}

	resp := joinResponse{
		GameID:          gameID,
		GameName:        combined.Name,
		GameDisplayName: mgrInfo.manager.Delegate().DisplayName(),
		MinPlayers:      combined.NumPlayers, // for now: NumPlayers is fixed at create-time
		MaxPlayers:      combined.NumPlayers,
		CurrentPlayers:  currentPlayers,
		AvailableSeats:  availableSeats,
		// Note this only says WHETHER a picker is needed; the slot details
		// (which require auth) come from /api/join/seat-options, so a
		// brute-force scrape of /api/join can't learn asymmetric structure.
		RequiresSeatPicker: s.requiresSeatPicker(mgrInfo),
	}
	resp.JoinTicket, err = s.issueJoinTicket(gameID, time.Now())
	if err != nil {
		s.logger.Warnln("Could not issue join ticket:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal join error", "code": "INTERNAL_ERROR"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
