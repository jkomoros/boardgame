package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/users"
	"strconv"
)

const (
	ctxGameKey            = "ctxGame"
	ctxAdminAllowedKey    = "ctxAdminAllowed"
	ctxViewingPlayerAsKey = "ctxViewingPlayerAs"
	ctxUserKey            = "ctxUser"
)

const (
	qryAdminKey  = "admin"
	qryPlayerKey = "player"
)

const (
	invalidPlayerIndex = boardgame.PlayerIndex(-10)
)

func (s *Server) setUser(c *gin.Context, user *users.StorageRecord) {
	c.Set(ctxUserKey, user)
}

func (s *Server) getUser(c *gin.Context) *users.StorageRecord {
	obj, ok := c.Get(ctxUserKey)

	if !ok {
		return nil
	}

	user, ok := obj.(*users.StorageRecord)

	if !ok {
		return nil
	}

	return user
}

func (s *Server) setGame(c *gin.Context, game *boardgame.Game) {
	c.Set(ctxGameKey, game)
}

func (s *Server) getGame(c *gin.Context) *boardgame.Game {
	obj, ok := c.Get(ctxGameKey)

	if !ok {
		return nil
	}

	game, ok := obj.(*boardgame.Game)

	if !ok {
		return nil
	}

	return game
}

func (s *Server) setViewingAsPlayer(c *gin.Context, playerIndex boardgame.PlayerIndex) {
	c.Set(ctxViewingPlayerAsKey, playerIndex)
}

func (s *Server) getViewingAsPlayer(c *gin.Context) boardgame.PlayerIndex {
	obj, ok := c.Get(ctxViewingPlayerAsKey)

	if !ok {
		return invalidPlayerIndex
	}

	playerIndex, ok := obj.(boardgame.PlayerIndex)

	if !ok {
		return invalidPlayerIndex
	}

	return playerIndex
}

func (s *Server) getRequestPlayerIndex(c *gin.Context) boardgame.PlayerIndex {
	player := c.Query(qryPlayerKey)

	if player == "" {
		return invalidPlayerIndex
	}

	playerIndexInt, err := strconv.Atoi(player)

	if err != nil {
		return invalidPlayerIndex
	}

	return boardgame.PlayerIndex(playerIndexInt)
}

func (s *Server) effectivePlayerIndex(c *gin.Context) boardgame.PlayerIndex {

	adminAllowed := s.getAdminAllowed(c)
	requestAdmin := s.getRequestAdmin(c)

	isAdmin := s.calcIsAdmin(c, adminAllowed, requestAdmin)

	requestPlayerIndex := s.getRequestPlayerIndex(c)
	viewingAsPlayer := s.getViewingAsPlayer(c)

	return s.calcEffectivePlayerIndex(isAdmin, requestPlayerIndex, viewingAsPlayer)
}

func (s *Server) calcEffectivePlayerIndex(isAdmin bool, requestPlayerIndex boardgame.PlayerIndex, viewingAsPlayer boardgame.PlayerIndex) boardgame.PlayerIndex {

	result := requestPlayerIndex

	if !isAdmin {
		result = viewingAsPlayer

		if result == invalidPlayerIndex {
			result = boardgame.ObserverPlayerIndex
		}
	}
	return requestPlayerIndex
}

func (s *Server) calcAdminAllowed(c *gin.Context, user *users.StorageRecord) bool {
	adminAllowed := true

	if user == nil {
		return false
	}

	if !s.config.DisableAdminChecking {

		//Are they allowed to be admin or not?

		matchedAdmin := false

		for _, userId := range s.config.AdminUserIds {
			if user.Id == userId {
				matchedAdmin = true
				break
			}
		}

		if !matchedAdmin {
			//Nope, you weren't an admin. Sorry!
			adminAllowed = false
		}

	}

	return adminAllowed

}

func (s *Server) setAdminAllowed(c *gin.Context, allowed bool) {
	c.Set(ctxAdminAllowedKey, allowed)
}

func (s *Server) calcIsAdmin(c *gin.Context, adminAllowed bool, requestAdmin bool) bool {
	return adminAllowed && requestAdmin
}

func (s *Server) getRequestAdmin(c *gin.Context) bool {
	return c.Query(qryAdminKey) == "1"
}

//returns true if the request asserts the user is an admin, and the user is
//allowed to be an admin.
func (s *Server) getAdminAllowed(c *gin.Context) bool {
	obj, ok := c.Get(ctxAdminAllowedKey)

	adminAllowed := false

	if !ok {
		return false
	}

	adminAllowed, ok = obj.(bool)

	if !ok {
		return false
	}

	return adminAllowed

}
