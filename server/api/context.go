package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/users"
	"log"
	"strconv"
)

const (
	ctxGameKey            = "ctxGame"
	ctxAdminAllowedKey    = "ctxAdminAllowed"
	ctxViewingPlayerAsKey = "ctxViewingPlayerAs"
	ctxUserKey            = "ctxUser"
)

const (
	qryAdminKey    = "admin"
	qryPlayerKey   = "player"
	qryGameIdKey   = "id"
	qryGameNameKey = "name"
	qryManagerKey  = "manager"
)

const (
	invalidPlayerIndex = boardgame.PlayerIndex(-10)
)

func (s *Server) getRequestManager(c *gin.Context) string {
	return c.PostForm(qryManagerKey)
}

func (s *Server) getRequestGameId(c *gin.Context) string {
	return c.Param(qryGameIdKey)
}

func (s *Server) getRequestGameName(c *gin.Context) string {
	return c.Param(qryGameNameKey)
}

func (s *Server) getRequestCookie(c *gin.Context) string {
	result, err := c.Cookie(cookieName)

	if err != nil {
		log.Println("Couldnt' get cookie:", err)
		return ""
	}

	return result
}

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

func (s *Server) calcViewingAsPlayerAndEmptySlots(userIds []string, user *users.StorageRecord) (player boardgame.PlayerIndex, emptySlots []boardgame.PlayerIndex) {

	result := boardgame.ObserverPlayerIndex

	for i, userId := range userIds {
		if userId == "" {
			emptySlots = append(emptySlots, boardgame.PlayerIndex(i))
		}
		if userId == user.Id {
			//We're here!
			result = boardgame.PlayerIndex(i)
		}
	}

	return result, emptySlots
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

	isAdmin := s.calcIsAdmin(adminAllowed, requestAdmin)

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

func (s *Server) calcAdminAllowed(user *users.StorageRecord) bool {
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

func (s *Server) calcIsAdmin(adminAllowed bool, requestAdmin bool) bool {
	return adminAllowed && requestAdmin
}

func (s *Server) getRequestAdmin(c *gin.Context) bool {

	result := c.Query(qryAdminKey) == "1"

	if result {
		return result
	}

	return c.PostForm(qryAdminKey) == "1"
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

func (s *Server) getMoveFromForm(c *gin.Context, game *boardgame.Game) (boardgame.Move, error) {

	move := game.PlayerMoveByName(c.PostForm("MoveType"))

	if move == nil {
		return nil, errors.New("Invalid MoveType")
	}

	//TODO: should we use gin's Binding to do this instead?

	for _, field := range formFields(move) {

		rawVal := c.PostForm(field.Name)

		switch field.Type {
		case boardgame.TypeInt:
			if rawVal == "" {
				return nil, errors.New(fmt.Sprint("An int field had no value", field.Name))
			}
			num, err := strconv.Atoi(rawVal)
			if err != nil {
				return nil, errors.New(fmt.Sprint("Couldn't set field", field.Name, err))
			}
			move.ReadSetter().SetProp(field.Name, num)
		case boardgame.TypeBool:
			if rawVal == "" {
				move.ReadSetter().SetProp(field.Name, false)
				continue
			}
			num, err := strconv.Atoi(rawVal)
			if err != nil {
				return nil, errors.New(fmt.Sprint("Couldn't set field", field.Name, err))
			}
			if num == 1 {
				move.ReadSetter().SetProp(field.Name, true)
			} else {
				move.ReadSetter().SetProp(field.Name, false)
			}
		case boardgame.TypeIllegal:
			return nil, errors.New(fmt.Sprint("Field", field.Name, "was an unknown value type"))
		}
	}

	return move, nil
}
