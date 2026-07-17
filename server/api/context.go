package api

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/server/api/users"
)

const (
	ctxGameKey            = "ctxGame"
	ctxAdminAllowedKey    = "ctxAdminAllowed"
	ctxViewingPlayerAsKey = "ctxViewingPlayerAs"
	ctxUserKey            = "ctxUser"
	ctxHasEmptySlots      = "ctxHasEmptySlots"
)

const (
	qryAdminKey             = "admin"
	qryPlayerKey            = "player"
	qryAutoCurrentPlayerKey = "current"
	qryGameIDKey            = "id"
	qryGameNameKey          = "name"
	qryManagerKey           = "manager"
	qryNumPlayersKey        = "numplayers"
	qryAgentKey             = "agent-player-"
	qryGameVersion          = "version"
	qryOpen                 = "open"
	qryVisible              = "visible"
	qryFromVersion          = "from"
	// qryCompanionMode is the form/query param the create-game form sends
	// (value "1" or "0") to request Table+Hand companion mode. The server
	// validates the request against manager.supportsTableHandMode before
	// honoring it.
	qryCompanionMode = "companionMode"
)

const (
	invalidPlayerIndex = boardgame.PlayerIndex(-10)
)

func (s *Server) getRequestManager(c *gin.Context) string {
	return c.PostForm(qryManagerKey)
}

func (s *Server) getRequestNumPlayers(c *gin.Context) int {
	rawValue := c.PostForm(qryNumPlayersKey)

	if rawValue == "" {
		rawValue = "0"
	}

	numPlayers, err := strconv.Atoi(rawValue)

	if err != nil {
		return 0
	}

	return numPlayers

}

func (s *Server) getRequestAgents(c *gin.Context, expectedNum int) []string {
	var result []string
	for i := 0; i < expectedNum; i++ {
		result = append(result, c.PostForm(qryAgentKey+strconv.Itoa(i)))
	}
	return result
}

func (s *Server) getRequestGameVersion(c *gin.Context) int {
	rawVal := c.Param(qryGameVersion)

	result, _ := strconv.Atoi(rawVal)

	return result
}

// getRequestVariant will get the various config
func (s *Server) getRequestVariant(c *gin.Context, variants boardgame.VariantConfig) map[string]string {
	result := make(map[string]string)

	for key, info := range variants {
		if formVal := c.PostForm("variant_" + key); formVal != "" {
			//We were given a formval. Sanity check it was one of the ones
			//htat's legal for this game.
			legal := false
			for _, val := range info.Values {
				if val.Name == formVal {
					legal = true
				}
			}

			if legal {
				result[key] = formVal
			} else {
				//TODO: what's the idiomatic way to log this?
				log.Println("Illegal value provided for key " + key + ": " + formVal + " skipping...")
			}
		}
	}

	return result
}

func (s *Server) getRequestFromVersion(c *gin.Context) int {
	rawVal := c.Query(qryFromVersion)

	result, _ := strconv.Atoi(rawVal)

	return result
}

// getRequestCompanionMode returns whether the create-game form requested
// Table+Hand mode for this new game. Returns false unless the form
// supplies a non-zero integer in qryCompanionMode.
func (s *Server) getRequestCompanionMode(c *gin.Context) bool {
	val := c.PostForm(qryCompanionMode)
	if val == "" {
		val = c.Query(qryCompanionMode)
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return false
	}
	return n > 0
}

func (s *Server) getRequestOpen(c *gin.Context) bool {
	open := c.Query(qryOpen)

	if open == "" {

		open = c.PostForm(qryOpen)

		if open == "" {
			return false
		}
	}

	openInt, err := strconv.Atoi(open)

	if err != nil {
		return false
	}

	return openInt > 0
}

func (s *Server) getRequestVisible(c *gin.Context) bool {
	visible := c.Query(qryVisible)

	if visible == "" {

		visible = c.PostForm(qryVisible)

		if visible == "" {
			return false
		}
	}

	visibleInt, err := strconv.Atoi(visible)

	if err != nil {
		return false
	}

	return visibleInt > 0
}

func (s *Server) getRequestGameID(c *gin.Context) string {
	return c.Param(qryGameIDKey)
}

func (s *Server) getRequestGameName(c *gin.Context) string {
	result := c.Param(qryGameNameKey)
	if result != "" {
		return result
	}
	return c.Query(qryGameNameKey)
}

func (s *Server) getRequestCookie(c *gin.Context) string {
	result, err := c.Cookie(cookieName)

	if err != nil {
		s.logger.Debugln("Couldn't get cookie:", err)
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

func (s *Server) setHasEmptySlots(c *gin.Context, hasEmptySlots bool) {
	c.Set(ctxHasEmptySlots, hasEmptySlots)
}

func (s *Server) getHasEmptySlots(c *gin.Context) bool {
	obj, ok := c.Get(ctxHasEmptySlots)

	if !ok {
		return false
	}

	emptySlots, ok := obj.(bool)

	if !ok {
		return false
	}

	return emptySlots
}

func (s *Server) calcViewingAsPlayerAndEmptySlots(userIds []string, user *users.StorageRecord, agents []string, closedSeats []bool) (player boardgame.PlayerIndex, emptySlots []boardgame.PlayerIndex) {

	result := boardgame.ObserverPlayerIndex

	if len(userIds) != len(agents) {
		panic("Agents and UserIds were different sizes")
	}

	if len(userIds) != len(closedSeats) {
		panic("UserIDs and Closed Seats were different sizes")
	}

	for i, userID := range userIds {
		if userID == "" && agents[i] == "" && !closedSeats[i] {
			emptySlots = append(emptySlots, boardgame.PlayerIndex(i))
		}
		if user != nil && userID == user.ID {
			//We're here!
			result = boardgame.PlayerIndex(i)
		}
	}

	return result, emptySlots
}

func (s *Server) getRequestPlayerIndex(c *gin.Context) boardgame.PlayerIndex {
	player := c.Query(qryPlayerKey)

	if player == "" {

		player = c.PostForm(qryPlayerKey)

		if player == "" {
			return invalidPlayerIndex
		}
	}

	playerIndexInt, err := strconv.Atoi(player)

	if err != nil {
		return invalidPlayerIndex
	}

	return boardgame.PlayerIndex(playerIndexInt)
}

func (s *Server) effectivePlayerIndex(c *gin.Context) boardgame.PlayerIndex {
	// A declared shared Table is always an observer, even after its authority
	// expires or moves and even when the browser carries seated/admin state.
	// Otherwise the recovery/terminal overlay could conceal a private hand that
	// remains present in the DOM or network response underneath it.
	if game := s.getGame(c); game != nil && tableSurfaceForRequest(c, game.ID()) {
		return boardgame.ObserverPlayerIndex
	}

	adminAllowed := s.getAdminAllowed(c)
	requestAdmin := s.getRequestAdmin(c)

	isAdmin := s.calcIsAdmin(adminAllowed, requestAdmin)

	requestPlayerIndex := s.getRequestPlayerIndex(c)
	viewingAsPlayer := s.getViewingAsPlayer(c)

	return s.calcEffectivePlayerIndex(isAdmin, requestPlayerIndex, viewingAsPlayer)
}

func (s *Server) effectiveAutoCurrentPlayer(c *gin.Context) bool {
	if game := s.getGame(c); game != nil && tableSurfaceForRequest(c, game.ID()) {
		return false
	}
	adminAllowed := s.getAdminAllowed(c)
	requestAdmin := s.getRequestAdmin(c)

	isAdmin := s.calcIsAdmin(adminAllowed, requestAdmin)

	if !isAdmin {
		return false
	}

	return s.getRequestAutoCurrentPlayer(c)
}

func (s *Server) calcEffectivePlayerIndex(isAdmin bool, requestPlayerIndex boardgame.PlayerIndex, viewingAsPlayer boardgame.PlayerIndex) boardgame.PlayerIndex {

	result := requestPlayerIndex

	if !isAdmin {
		result = viewingAsPlayer

		if result == invalidPlayerIndex {
			result = boardgame.ObserverPlayerIndex
		}
	}
	return result
}

func (s *Server) calcAdminAllowed(user *users.StorageRecord) bool {
	adminAllowed := true

	if user == nil {
		return false
	}

	if !s.config.DisableAdminChecking {

		//Are they allowed to be admin or not?

		matchedAdmin := false

		for _, userID := range s.config.AdminUserIds {
			if user.ID == userID {
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

func (s *Server) getRequestAutoCurrentPlayer(c *gin.Context) bool {

	result := c.Query(qryAutoCurrentPlayerKey) == "1"

	if result {
		return result
	}

	return c.PostForm(qryAutoCurrentPlayerKey) == "1"
}

// returns true if the request asserts the user is an admin, and the user is
// allowed to be an admin.
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

	move := game.MoveByName(c.PostForm("MoveType"))

	if move == nil {
		return nil, errors.New("Invalid MoveType")
	}

	if base.IsFixUp(move) {
		return nil, errors.New("players cannot make fixup moves")
	}

	//TODO: should we use gin's Binding to do this instead?

	inputFields, err := boardgame.ResolveMoveInputFields(move)
	if err != nil {
		return nil, fmt.Errorf("couldn't resolve move input contract: %w", err)
	}
	dispositionByName := make(map[string]boardgame.MoveInputDisposition, len(inputFields))
	for _, field := range inputFields {
		dispositionByName[field.Name] = field.Disposition
	}

	// Required creator input remains fail-closed when omitted. Context-owned,
	// server-defaulted, and unsupported fields preserve the value installed by
	// MoveByName/DefaultsForState when the request omits them. This is the server
	// half of the generated creator contract: a zero-input CurrentPlayer move
	// must not require a renderer to smuggle TargetPlayerIndex over the wire.
	if err := bindMoveFields(move, func(name string) (string, bool) {
		if raw, present := c.GetPostForm(name); present {
			return raw, true
		}
		return "", dispositionByName[name] == boardgame.MoveInputRequired
	}); err != nil {
		return nil, err
	}

	return move, nil
}

// bindMoveFields binds a move's form fields from get, which returns each field's
// raw string value and whether it was supplied. It is the shared arg-binding
// used by getMoveFromForm (form-encoded args from the move endpoint — its input
// contract decides whether omission is an error or preserves a default) and the batch preview
// handler (per-candidate JSON args — a field NOT supplied is left at its
// DefaultsForState value, so a candidate can vary just the fields it cares about
// and let sensible defaults stand for the rest, e.g. TargetPlayerIndex). It only
// sets the move's fields — never applies or reads legality.
func bindMoveFields(move boardgame.Move, get func(name string) (rawVal string, present bool)) error {
	for _, field := range formFields(move) {

		rawVal, present := get(field.Name)
		if !present {
			// Not supplied: keep the field's DefaultsForState value.
			continue
		}

		switch field.Type {
		case boardgame.TypeInt:
			if rawVal == "" {
				return errors.New(fmt.Sprint("An int field had no value", field.Name))
			}
			num, err := strconv.Atoi(rawVal)
			if err != nil {
				return errors.New(fmt.Sprint("Couldn't set field", field.Name, err))
			}
			if err := move.ReadSetter().SetIntProp(field.Name, num); err != nil {
				return errors.New("Couldn't set int prop " + field.Name + " " + err.Error())
			}
		case boardgame.TypePlayerIndex:
			if rawVal == "" {
				return errors.New("An int field had no value " + field.Name)
			}
			num, err := strconv.Atoi(rawVal)
			if err != nil {
				return errors.New("Couldn't set field " + field.Name + " " + err.Error())
			}
			if err := move.ReadSetter().SetPlayerIndexProp(field.Name, boardgame.PlayerIndex(num)); err != nil {
				return errors.New("Couldn't set int prop " + field.Name + " " + err.Error())
			}
		case boardgame.TypeBool:
			if rawVal == "" {
				if err := move.ReadSetter().SetBoolProp(field.Name, false); err != nil {
					return errors.New("Couldn't set bool prop with default: " + field.Name + " " + err.Error())
				}
				continue
			}
			num, err := strconv.Atoi(rawVal)
			if err != nil {
				return errors.New(fmt.Sprint("Couldn't set field", field.Name, err))
			}
			val := false
			if num == 1 {
				val = true
			}
			if err := move.ReadSetter().SetBoolProp(field.Name, val); err != nil {
				return errors.New("Couldnt set bool prop " + field.Name + ": " + err.Error())
			}
		case boardgame.TypeEnum:
			eVar, err := move.ReadSetter().EnumProp(field.Name)
			if err != nil {
				return errors.New("Invalid field name: " + err.Error())
			}
			//SetStringValue will also try converting to an int.

			if err := eVar.SetStringValue(rawVal); err != nil {
				return errors.New("Couldn't set field value: " + err.Error())
			}
		case boardgame.TypeIllegal:
			return errors.New(fmt.Sprint("Field", field.Name, "was an unknown value type"))
		}
	}

	return nil
}
