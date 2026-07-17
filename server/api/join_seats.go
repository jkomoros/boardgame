package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame"
)

const (
	joinCodeTicketRequired = "JOIN_TICKET_REQUIRED"
	joinCodeExpired        = "JOIN_EXPIRED"
	joinCodeRoomLocked     = "ROOM_LOCKED"
	joinCodeRoomFinished   = "ROOM_FINISHED"
	joinCodeRoomFull       = "ROOM_FULL"
	joinCodeSeatTaken      = "SEAT_TAKEN"
)

var companionAvatarSlugs = map[string]struct{}{
	"🦊": {}, "🐻": {}, "🦁": {}, "🐯": {}, "🐸": {}, "🐙": {},
	"🦄": {}, "🐳": {}, "🦉": {}, "🐧": {}, "🐲": {}, "🦋": {},
	"🐺": {}, "🦅": {}, "🦈": {}, "🐬": {}, "🦎": {}, "🐢": {},
	"🦩": {}, "🐝": {}, "🐞": {}, "🦇": {}, "🐠": {}, "🦜": {},
}

func validCompanionAvatarSlug(slug string) bool {
	_, ok := companionAvatarSlugs[slug]
	return ok
}

type companionSeatStatus struct {
	PlayerIndex int    `json:"playerIndex"`
	Label       string `json:"label"`
	Status      string `json:"status"`
	Filled      bool   `json:"filled"`
	Available   bool   `json:"available"`
	AvatarSlug  string `json:"avatarSlug,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
}

func writeJoinProblem(c *gin.Context, status int, code, message string, extra gin.H) {
	body := gin.H{"error": message, "code": code}
	for key, value := range extra {
		body[key] = value
	}
	c.JSON(status, body)
}

func (s *Server) requireJoinTicket(c *gin.Context, gameID string) bool {
	ticket := c.GetHeader(joinTicketHeader)
	if ticket == "" {
		writeJoinProblem(c, http.StatusUnauthorized, joinCodeTicketRequired, "Scan or enter the room code again", nil)
		return false
	}
	if err := s.verifyJoinTicket(ticket, gameID, timeNow()); err != nil {
		writeJoinProblem(c, http.StatusUnauthorized, joinCodeExpired, "Your invitation expired; scan or enter the room code again", nil)
		return false
	}
	return true
}

// timeNow is a variable to keep join-ticket boundary tests deterministic.
var timeNow = func() time.Time { return time.Now() }

func (s *Server) companionSeatStatuses(game *boardgame.Game) []companionSeatStatus {
	userIDs := s.storage.UserIDsForGame(game.ID())
	agents := game.Agents()
	closed := s.closedSeatsForGame(game)
	statuses := make([]companionSeatStatus, game.NumPlayers())
	for i := range statuses {
		status := companionSeatStatus{
			PlayerIndex: i,
			Label:       "Seat " + strconv.Itoa(i+1),
			Status:      "open",
			Available:   true,
		}
		var uid string
		if i < len(userIDs) {
			uid = userIDs[i]
		}
		switch {
		case uid != "":
			status.Status = "human"
			status.Filled = true
			status.Available = false
			if pres, err := s.storage.SeatPresentation(game.ID(), boardgame.PlayerIndex(i)); err == nil && pres != nil {
				status.AvatarSlug = pres.AvatarSlug
				status.DisplayName = pres.DisplayName
			}
		case i < len(agents) && agents[i] != "":
			status.Status = "agent"
			status.Filled = true
			status.Available = false
		case i < len(closed) && closed[i]:
			status.Status = "closed"
			status.Available = false
		}
		statuses[i] = status
	}
	return statuses
}

func existingSeatForUID(userIDs []string, uid string) (boardgame.PlayerIndex, bool) {
	for i, existing := range userIDs {
		if existing == uid {
			return boardgame.PlayerIndex(i), true
		}
	}
	return boardgame.AdminPlayerIndex, false
}
