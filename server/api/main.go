package api

import (
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	cors "github.com/itsjamie/gin-cors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"github.com/jkomoros/boardgame/errors"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/jkomoros/boardgame/server/api/users"
	"github.com/sirupsen/logrus"
)

//Server is the main server object.
type Server struct {
	managers managerMap

	//map of game ID to players to seat
	playersToSeat map[string][]*playerToSeat

	storage *ServerStorageManager
	//We store the last error so that next time viewHandler is called we can
	//display it. Yes, this is a hack.
	lastErrorMessage string
	config           *config.Mode

	overriders []config.OptionOverrider

	upgrader websocket.Upgrader

	notifier *versionNotifier
	logger   *logrus.Logger
}

type renderer struct {
	s            *Server
	c            *gin.Context
	rendered     bool
	cookieCalled bool
	cookieValue  string
}

type moveForm struct {
	Name     string
	HelpText string
	Fields   []*moveFormField
}

type moveFormFieldType int

type moveFormField struct {
	Name         string
	Type         boardgame.PropertyType
	EnumName     string `json:",omitempty"`
	DefaultValue interface{}
}

type managerInfo struct {
	manager         *boardgame.GameManager
	seatPlayerMoves []string
	//If the game's playerState has seatPlayer. Typically the answer is is yes
	//if len(seatPlayerMoves) != 0, as moves.SeatPlayer and behaviors.Seat are
	//used in conjunction most often.
	playerHasSeat bool
}

type playerToSeat struct {
	s         *Server
	gameID    string
	seatIndex boardgame.PlayerIndex
}

type managerMap map[string]*managerInfo

/*

Overview of the types of handlers and methods

server.fooHandler take a context. They grab all of the dependencies and pass them to the doers.
server.doFoo takes a renderer and all dependencies that come from context. It may fetch additional items from e.g. storage. It renders the result.
server.getRequestFoo fetches an argument from the context's request and nothing else
server.getFoo grabs a thing that was stored in Context and nothing else
server.setFoo sets a thing into context and nothing else
server.calcFoo takes dependencies and returns a result, with no touching context.
*/

/*

NewServer returns a new server. Get it to run by calling Start(). storage
should a *ServerStorageManager, which can be created either from
NewServerStorageManager.

Use it like so:

	func main() {
		storage := server.NewServerStorageManager(bolt.NewStorageManager(".database"))
		defer storage.Close()
		server.NewServer(storage, mygame.NewManager(storage)).Start()
	}

*/
func NewServer(storage *ServerStorageManager, delegates ...boardgame.GameDelegate) *Server {

	logger := logrus.New()

	result := &Server{
		managers:      make(managerMap),
		playersToSeat: make(map[string][]*playerToSeat),
		storage:       storage,
		logger:        logger,
	}

	storage.server = result

	var managers []*boardgame.GameManager

	for _, delegate := range delegates {

		manager, err := boardgame.NewGameManager(delegate, storage)

		if err != nil {
			logger.Fatalln("Couldn't create manager: " + err.Error())
			return nil
		}

		name := manager.Delegate().Name()
		manager.SetLogger(logger)

		pState := manager.ExampleState().ImmutablePlayerStates()[0]
		playerHasSeat := false
		if _, ok := pState.(interfaces.Seater); ok {
			playerHasSeat = true
		}

		result.managers[name] = &managerInfo{
			manager:         manager,
			seatPlayerMoves: managerSeatPlayerMoves(manager),
			playerHasSeat:   playerHasSeat,
		}
		managers = append(managers, manager)
		if manager.Storage() != storage {
			logger.Fatalln("The storage for one of the managers was not the same item passed in as major storage.")
			return nil
		}

	}

	result.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     result.checkOriginForSocket,
	}

	storage.WithManagers(managers)

	return result

}

func (p *playerToSeat) PlayerIndex() boardgame.PlayerIndex {
	return p.seatIndex
}

func (p *playerToSeat) Commit() {
	slice := p.s.playersToSeat[p.gameID]
	if len(slice) == 0 {
		return
	}
	indexInSlice := -1
	for i, player := range slice {
		if player == p {
			indexInSlice = i
			break
		}
	}
	//I guess we weren't in our parent, weird.
	if indexInSlice == -1 {
		return
	}
	p.s.playersToSeat[p.gameID] = append(slice[:indexInSlice], slice[indexInSlice+1:]...)
}

//managerSeatPlayerMoves returns the move names for the given manager that are a
//seat player move. If len(result) is 0, then the game does not have a seat
//player move.
func managerSeatPlayerMoves(manager *boardgame.GameManager) []string {
	var result []string
	for _, move := range manager.ExampleMoves() {
		if seatPlayer, ok := move.(interfaces.SeatPlayerMover); ok {
			//Technically it could return false even if it implements the
			//method, so check it explicitly returns true.
			if seatPlayer.IsSeatPlayerMove() {
				result = append(result, move.Info().Name())
			}
		}
	}
	return result
}

func (s *Server) newRenderer(c *gin.Context) *renderer {
	return &renderer{
		s,
		c,
		false,
		false,
		"",
	}
}

func (r *renderer) Error(f *errors.Friendly) {
	if r.rendered {
		r.s.logger.Errorln("Error called on already-rendered renderer")
	}

	if f == nil {
		f = errors.New("Nil error provided to r.Error")
	}

	r.writeCookie()

	r.c.JSON(http.StatusOK, gin.H{
		"Status":        "Failure",
		"Error":         f.Error(),
		"FriendlyError": f.FriendlyError(),
	})

	fields := logrus.Fields{}

	for key, val := range f.Fields() {
		fields[key] = val
	}

	fields["Friendly"] = f.FriendlyError()
	fields["Error"] = f.Error()
	fields["Secure"] = f.SecureError()

	r.s.logger.WithFields(fields).Errorln("Server error")

	r.rendered = true
}

func (r *renderer) Success(keys gin.H) {

	if r.rendered {
		panic("Success called on alread-rendered renderer")
	}

	r.writeCookie()

	if keys == nil {
		keys = gin.H{}
	}

	result := gin.H{}

	for key, val := range keys {
		result[key] = val
	}

	result["Status"] = "Success"

	r.c.JSON(http.StatusOK, result)

	r.rendered = true
}

func (r *renderer) writeCookie() {
	if r.rendered {
		return
	}
	if !r.cookieCalled {
		return
	}

	//TODO: might need to set the domain in production.

	if r.cookieValue == "" {
		//Unset the cookie
		r.c.SetCookie(cookieName, "", int(time.Now().Add(time.Hour*10000*-1).Unix()), "", "", false, false)
		return
	}

	r.c.SetCookie(cookieName, r.cookieValue, int(time.Now().Add(time.Hour*100).Unix()), "", "", false, false)
}

//SetAuthCookie will set the auth cookie to the specified value. If called
//multiple times for a single request will only actually write headers for the
//last one.
func (r *renderer) SetAuthCookie(value string) {

	//We don't write the cookies to the response yet because we might get
	//multiple SetAuthCookie calls in one response.

	r.cookieCalled = true
	r.cookieValue = value

}

func (s *Server) userSetup(c *gin.Context) {
	cookie := s.getRequestCookie(c)

	if cookie == "" {
		s.logger.Debugln("No cookie set")
		return
	}

	user := s.storage.GetUserByCookie(cookie)

	if user == nil {
		s.logger.Debugln("No user associated with that cookie")
		return
	}
	user.LastSeen = time.Now().UnixNano()
	s.storage.UpdateUser(user)

	s.setUser(c, user)

	s.setAdminAllowed(c, s.calcAdminAllowed(user))
}

func (s *Server) gameFromID(gameID, gameName string) *boardgame.Game {

	manager := s.managers[gameName].manager

	if manager == nil {
		s.logger.Errorln("Couldnt' find manager for", gameName)
		return nil
	}

	game := manager.Game(gameID)

	//TODO: figure out a way to return a meaningful error

	if game == nil {
		s.logger.Errorln("Couldn't find game with id", gameID)
		return nil
	}

	if game.Name() != gameName {
		s.logger.Errorln("The name of the game was not what we were expecting. Wanted", gameName, "got", game.Name())
		return nil
	}

	return game
}

//closedSeatsForGame will return a slice of bools of equal length to the game's
//NumPlayers, where each one is set to true if the playerState has a Seat and
//the seat is marked as closed.
func (s *Server) closedSeatsForGame(game *boardgame.Game) []bool {
	result := make([]bool, game.NumPlayers())
	info := s.managers[game.Manager().Delegate().Name()]
	if info == nil {
		return result
	}
	//If the game doesn't use Seat, then just return now
	if !info.playerHasSeat {
		return result
	}
	state := game.CurrentState()
	for i, p := range state.ImmutablePlayerStates() {
		if seater, ok := p.(interfaces.Seater); ok {
			if seater.SeatIsClosed() {
				result[i] = true
			}
		}
	}
	return result
}

//gameAPISetup fetches the game configured in the URL and puts it in context.
func (s *Server) gameAPISetup(c *gin.Context) {

	id := s.getRequestGameID(c)

	gameName := s.getRequestGameName(c)

	game := s.gameFromID(id, gameName)

	if game == nil {
		return
	}

	s.setGame(c, game)

	userIds := s.storage.UserIDsForGame(id)

	if userIds == nil {
		s.logger.Errorln("No userIds associated with game", logrus.Fields{
			"gameName": gameName,
			"gamdId":   id,
		})
	}

	closedSeats := s.closedSeatsForGame(game)

	user := s.getUser(c)

	if user == nil {
		s.logger.Warnln("No user provided")
		//The rest of the flow will handle a nil user fine
	}

	effectiveViewingAsPlayer, emptySlots := s.calcViewingAsPlayerAndEmptySlots(userIds, user, game.Agents(), closedSeats)

	if user != nil && effectiveViewingAsPlayer == boardgame.ObserverPlayerIndex && len(emptySlots) > 0 && len(emptySlots) == game.NumPlayers()-game.NumAgentPlayers() {
		//Special case: we're the first player, we likely just created it. Just join the thing!

		slot := emptySlots[0]

		if err := s.doSeatPlayer(game, slot, user); err != nil {
			s.logger.Errorln("Tried to set the user as player " + slot.String() + " but failed: " + err.Error())
			return
		}
		s.setHasEmptySlots(c, false)
		effectiveViewingAsPlayer = slot

	} else {
		s.setHasEmptySlots(c, len(emptySlots) != 0)
	}
	s.setViewingAsPlayer(c, effectiveViewingAsPlayer)

}

//Checks to make sure the user is logged in, fails if not.
func (s *Server) requireLoggedIn(c *gin.Context) {

	r := s.newRenderer(c)

	user := s.getUser(c)

	if user == nil {
		r.Error(errors.NewFriendly("Not logged in"))
		c.Abort()
		return
	}

	//All good!
}

func (s *Server) joinGameHandler(c *gin.Context) {
	r := s.newRenderer(c)

	game := s.getGame(c)

	if game == nil {
		r.Error(errors.NewFriendly("No such game"))
		return
	}

	user := s.getUser(c)

	userIds := s.storage.UserIDsForGame(game.ID())

	closedSeats := s.closedSeatsForGame(game)

	viewingAsPlayer, emptySlots := s.calcViewingAsPlayerAndEmptySlots(userIds, user, game.Agents(), closedSeats)

	s.doJoinGame(r, game, viewingAsPlayer, emptySlots, user)

}

func (s *Server) doSeatPlayer(game *boardgame.Game, slot boardgame.PlayerIndex, user *users.StorageRecord) error {
	if len(s.managers[game.Name()].seatPlayerMoves) > 0 {
		//This is a game that uses SeatPlayer move, so instead of adding the
		//player right now we should go into pending mode to inject the player.

		gameID := game.ID()

		player := &playerToSeat{
			s,
			gameID,
			slot,
		}

		s.playersToSeat[gameID] = append(s.playersToSeat[gameID], player)

		//Now we have information waiting for SeatPlayer. Tell the engine to
		//check whether fixups need to be applied, becuase we know that
		//something outside of state has changed that might change whether moves
		//are valid. We don't have to worry about race conditions because Game's
		//mainLoop will make sure this isn't triggered while another move is
		//being processed.
		game.Manager().Internals().ForceFixUp(game)

		//We deliberately fall through here and set that the player is
		//affirmatively in that game, even though they aren't seated. This is
		//because this 'pending' seating player currently has no way to retreat;
		//they'll be seated into that seat the next time a SeatPlayer move is
		//legal, and if another player comes right now, as far as they're
		//concerned, that seat is taken.

		//This could get out of sync if the server is shut down while the player
		//is pending but before theyr'e actually seated. See #221.
	}

	return s.storage.SetPlayerForGame(game.ID(), slot, user.ID)
}

func (s *Server) doJoinGame(r *renderer, game *boardgame.Game, viewingAsPlayer boardgame.PlayerIndex, emptySlots []boardgame.PlayerIndex, user *users.StorageRecord) {

	if user == nil {
		r.Error(errors.New("no user provided"))
		return
	}

	eGame, err := s.storage.ExtendedGame(game.ID())

	if err != nil {
		r.Error(errors.New("Couldn't get extended information about game: " + err.Error()))
		return
	}

	if !eGame.Open {
		r.Error(errors.NewFriendly("the game is not open to people joining"))
		return
	}

	if viewingAsPlayer != boardgame.ObserverPlayerIndex {
		r.Error(errors.NewFriendly("The given player is already in the game."))
		return
	}

	if len(emptySlots) == 0 {
		r.Error(errors.NewFriendly("There aren't any empty slots in the game to join."))
		return
	}

	slot := emptySlots[0]

	if err := s.doSeatPlayer(game, slot, user); err != nil {
		r.Error(errors.New("Tried to set the user as player " + slot.String() + " but failed: " + err.Error()))
	}

	r.Success(nil)
}

func (s *Server) newGameHandler(c *gin.Context) {

	r := s.newRenderer(c)

	managerID := s.getRequestManager(c)

	numPlayers := s.getRequestNumPlayers(c)

	manager := s.managers[managerID].manager

	if manager == nil {
		r.Error(errors.NewFriendly("That is not a legal type of game").WithError(managerID + " is not a legal manager for this server"))
		return
	}

	variant := s.getRequestVariant(c, manager.Variants())

	if numPlayers == 0 && manager != nil {
		numPlayers = manager.Delegate().DefaultNumPlayers()
	}

	agents := s.getRequestAgents(c, numPlayers)

	owner := s.getUser(c)

	open := s.getRequestOpen(c)
	visible := s.getRequestVisible(c)

	s.doNewGame(r, owner, manager, numPlayers, agents, open, visible, variant)

}

func (s *Server) doNewGame(r *renderer, owner *users.StorageRecord, manager *boardgame.GameManager, numPlayers int, agents []string, open bool, visible bool, variant map[string]string) {

	if manager == nil {
		r.Error(errors.New("No manager provided"))
		return
	}

	if owner == nil {
		r.Error(errors.NewFriendly("You must be signed in to create a game."))
		return
	}

	game, err := manager.NewGame(numPlayers, variant, agents)

	if err != nil {
		//TODO: communicate the error state back to the client in a sane way
		if f, ok := err.(*errors.Friendly); ok {
			r.Error(f)
		} else {
			r.Error(errors.New(err.Error()))
		}
		return
	}

	eGame, err := s.storage.ExtendedGame(game.ID())

	if err != nil {
		r.Error(errors.New("Couldn't retrieve saved game: " + err.Error()))
		return
	}

	eGame.Owner = owner.ID
	eGame.Open = open
	eGame.Visible = visible

	//TODO: set Open, Visible based on query params.

	if err := s.storage.UpdateExtendedGame(game.ID(), eGame); err != nil {
		r.Error(errors.New("Couldn't save extended game metadata: " + err.Error()))
		return
	}

	r.Success(gin.H{
		"GameID":   game.ID(),
		"GameName": game.Name(),
	})
}

func (s *Server) listGamesHandler(c *gin.Context) {

	r := s.newRenderer(c)

	user := s.getUser(c)

	gameName := s.getRequestGameName(c)

	adminAllowed := s.getAdminAllowed(c)
	requestAdmin := s.getRequestAdmin(c)

	isAdmin := s.calcIsAdmin(adminAllowed, requestAdmin)

	s.doListGames(r, user, gameName, isAdmin)
}

func (s *Server) doListGames(r *renderer, user *users.StorageRecord, gameName string, isAdmin bool) {
	var userID string
	if user != nil {
		userID = user.ID
	}
	result := gin.H{
		"ParticipatingActiveGames":   s.listGamesWithUsers(100, listing.ParticipatingActive, userID, gameName),
		"ParticipatingFinishedGames": s.listGamesWithUsers(100, listing.ParticipatingFinished, userID, gameName),
		"VisibleJoinableActiveGames": s.listGamesWithUsers(100, listing.VisibleJoinableActive, userID, gameName),
		"VisibleActiveGames":         s.listGamesWithUsers(100, listing.VisibleActive, userID, gameName),
	}
	if isAdmin {
		result["AllGames"] = s.storage.ListGames(100, listing.All, "", gameName)
	}
	r.Success(result)
}

type gameStorageRecordWithUsers struct {
	*extendedgame.CombinedStorageRecord
	Players              []*playerBoardInfo
	ReadableLastActivity string
}

func (s *Server) listGamesWithUsers(max int, list listing.Type, userID string, gameName string) []*gameStorageRecordWithUsers {
	games := s.storage.ListGames(max, list, userID, gameName)

	result := make([]*gameStorageRecordWithUsers, len(games))

	for i, game := range games {

		manager := s.managers[game.Name].manager

		//When SecretSalt is empty it will be omitted from the JSON output.

		//TODO: isn't it brittle that we only sanitize the critically
		//important SecretSalt here?
		game.SecretSalt = ""

		result[i] = &gameStorageRecordWithUsers{
			game,
			s.gamePlayerInfo(&game.GameStorageRecord, manager),
			humanize.Time(game.Modified),
		}
	}

	return result

}

func (s *Server) listManagerHandler(c *gin.Context) {
	r := s.newRenderer(c)
	s.doListManager(r)
}

func (s *Server) doListManager(r *renderer) {
	var managers []map[string]interface{}
	for name, mInfo := range s.managers {
		manager := mInfo.manager
		agents := make([]map[string]interface{}, len(manager.Agents()))
		for i, agent := range manager.Agents() {
			agents[i] = map[string]interface{}{
				"Name":        agent.Name(),
				"DisplayName": agent.DisplayName(),
			}
		}
		var variant []interface{}

		variants := manager.Variants()

		sortedKeys := make([]string, len(variants))

		i := 0

		for key := range variants {
			sortedKeys[i] = key
			i++
		}

		sort.Strings(sortedKeys)

		for _, key := range sortedKeys {

			info := variants[key]

			part := make(map[string]interface{})
			part["Name"] = info.Name
			part["DisplayName"] = info.DisplayName
			part["Description"] = info.Description

			//We need to sort the values so they're in a stable order. It
			//should be the default value (if there is one) then everything
			//else in sorted order.

			var defaultValueInfo map[string]string
			var valueInfo []map[string]string

			for _, val := range info.Values {
				valuePart := make(map[string]string)

				valuePart["Value"] = val.Name
				valuePart["DisplayName"] = val.DisplayName
				valuePart["Description"] = val.Description

				if info.Default == val.Name {
					defaultValueInfo = valuePart
				} else {
					valueInfo = append(valueInfo, valuePart)
				}

			}

			sort.Slice(valueInfo, func(i, j int) bool {
				return valueInfo[i]["Value"] < valueInfo[j]["Value"]
			})

			if defaultValueInfo != nil {
				valueInfo = append([]map[string]string{defaultValueInfo}, valueInfo...)
			}

			part["Values"] = valueInfo

			variant = append(variant, part)
		}

		managers = append(managers, map[string]interface{}{
			"Name":              name,
			"DisplayName":       manager.Delegate().DisplayName(),
			"Description":       manager.Delegate().Description(),
			"DefaultNumPlayers": manager.Delegate().DefaultNumPlayers(),
			"MinNumPlayers":     manager.Delegate().MinNumPlayers(),
			"MaxNumPlayers":     manager.Delegate().MaxNumPlayers(),
			"Agents":            agents,
			"Variant":           variant,
		})
	}

	sort.Slice(managers, func(i, j int) bool {
		return managers[i]["Name"].(string) < managers[j]["Name"].(string)
	})

	r.Success(gin.H{
		"Managers": managers,
	})

}

func (s *Server) gameVersionHandler(c *gin.Context) {

	game := s.getGame(c)

	playerIndex := s.effectivePlayerIndex(c)

	version := s.getRequestGameVersion(c)

	fromVersion := s.getRequestFromVersion(c)

	autoCurrentPlayer := s.effectiveAutoCurrentPlayer(c)

	r := s.newRenderer(c)

	s.doGameVersion(r, game, version, fromVersion, playerIndex, autoCurrentPlayer)

}

func (s *Server) moveBundles(game *boardgame.Game, moves []*boardgame.MoveStorageRecord, playerIndex boardgame.PlayerIndex, autoCurrentPlayer bool) []gin.H {
	var bundles []gin.H

	if len(moves) == 0 {
		moves = append(moves, nil)
	}

	for _, move := range moves {

		version := 0
		if move != nil {
			version = move.Version
		}

		//This is the state for the end of the bundle.
		state := game.State(version)

		if autoCurrentPlayer {
			newPlayerIndex := game.Manager().Delegate().CurrentPlayerIndex(state)
			if newPlayerIndex.Valid(state) {
				playerIndex = newPlayerIndex
			}
		}

		//If state is nil, JSONForPlayer will basically treat it as just "give the
		//current version" which is a reasonable fallback.
		bundle := gin.H{
			"Game":            game.JSONForPlayer(playerIndex, state),
			"Move":            move,
			"ViewingAsPlayer": playerIndex,
			"Forms":           s.generateForms(game),
		}

		bundles = append(bundles, bundle)

	}

	return bundles
}

func (s *Server) doGameVersion(r *renderer, game *boardgame.Game, version, fromVersion int, playerIndex boardgame.PlayerIndex, autoCurrentPlayer bool) {
	if game == nil {
		r.Error(errors.NewFriendly("Couldn't find game"))
		return
	}

	if playerIndex == invalidPlayerIndex {
		r.Error(errors.New("Got invalid playerIndex"))
		return
	}

	moves, err := s.storage.Moves(game.ID(), fromVersion, version)

	//if there aren't any moves, that's only legal if it's the first version,
	//which happens sometimes when the player requests to view the game as a
	//different player.
	if fromVersion != 0 && version != 0 {
		if err != nil {
			r.Error(errors.New(err.Error()))
			return
		}
		if len(moves) == 0 {
			r.Error(errors.New("No moves in that range"))
			return
		}
	}

	r.Success(gin.H{
		"Bundles": s.moveBundles(game, moves, playerIndex, autoCurrentPlayer),
	})
}

//AddOverrides defines overrides that will be applied on top of the config we
//load. We return a reference to ourself to allow chaining of configurations.
func (s *Server) AddOverrides(overrides []config.OptionOverrider) *Server {
	s.overriders = append(s.overriders, overrides...)
	return s
}

func (s *Server) configureGameHandler(c *gin.Context) {
	game := s.getGame(c)

	var gameID string

	if game != nil {
		gameID = game.ID()
	}

	gameInfo, _ := s.storage.ExtendedGame(gameID)

	adminAllowed := s.getAdminAllowed(c)
	requestAdmin := s.getRequestAdmin(c)

	isAdmin := s.calcIsAdmin(adminAllowed, requestAdmin)

	user := s.getUser(c)

	open := s.getRequestOpen(c)
	visible := s.getRequestVisible(c)

	r := s.newRenderer(c)

	s.doConfigureGame(r, user, isAdmin, game, gameInfo, open, visible)

}

func (s *Server) doConfigureGame(r *renderer, user *users.StorageRecord, isAdmin bool, game *boardgame.Game, gameInfo *extendedgame.StorageRecord, open, visible bool) {

	if user == nil {
		r.Error(errors.New("No user provided"))
		return
	}

	if game == nil {
		r.Error(errors.New("Invalid game"))
		return
	}

	if gameInfo == nil {
		r.Error(errors.New("Couldn't fetch game info"))
		return
	}

	if !isAdmin && user.ID != gameInfo.Owner {
		r.Error(errors.NewFriendly("You are neither the owner nor an admin."))
		return
	}

	gameInfo.Open = open
	gameInfo.Visible = visible

	if err := s.storage.UpdateExtendedGame(game.ID(), gameInfo); err != nil {
		r.Error(errors.New("Error updating the extended game: " + err.Error()))
		return
	}

	r.Success(nil)

}

//gameInfo is the first payload when a game is loaded, including immutables
//like chest, but also the initial game state payload as a convenience.
func (s *Server) gameInfoHandler(c *gin.Context) {

	game := s.getGame(c)

	playerIndex := s.effectivePlayerIndex(c)

	hasEmptySlots := s.getHasEmptySlots(c)

	fromVersion := s.getRequestFromVersion(c)

	var gameID string

	if game != nil {
		gameID = game.ID()
	}

	//TODO: should this be done in gameAPISetup?
	gameInfo, _ := s.storage.ExtendedGame(gameID)

	user := s.getUser(c)

	r := s.newRenderer(c)

	s.doGameInfo(r, game, playerIndex, hasEmptySlots, gameInfo, user, fromVersion)

}

type playerBoardInfo struct {
	DisplayName string
	IsAgent     bool
	IsEmpty     bool
	PhotoURL    string
}

func (s *Server) gamePlayerInfo(game *boardgame.GameStorageRecord, manager *boardgame.GameManager) []*playerBoardInfo {

	if manager == nil {
		return nil
	}

	result := make([]*playerBoardInfo, game.NumPlayers)

	userIds := s.storage.UserIDsForGame(game.ID)
	agentNames := game.Agents

	for i := range result {

		player := &playerBoardInfo{}

		result[i] = player

		if agentNames[i] != "" {
			agent := manager.AgentByName(agentNames[i])

			if agent != nil {
				player.DisplayName = agent.DisplayName()
			}
			player.IsAgent = true
			player.IsEmpty = false
			continue
		}

		userID := userIds[i]

		if userID == "" {
			player.IsEmpty = true
			player.IsAgent = false
			player.DisplayName = ""
			continue
		}

		user := s.storage.GetUserByID(userID)

		if user == nil {
			player.IsAgent = false
			player.IsEmpty = false
			player.DisplayName = "Unknown user"
			continue
		}

		player.IsAgent = false
		player.IsEmpty = false
		player.PhotoURL = user.PhotoURL
		player.DisplayName = user.EffectiveDisplayName()

		if player.DisplayName == "" {
			player.DisplayName = "Player " + strconv.Itoa(i)
		}

	}

	return result
}

func (s *Server) doGameInfo(r *renderer, game *boardgame.Game, playerIndex boardgame.PlayerIndex, hasEmptySlots bool, gameInfo *extendedgame.StorageRecord, user *users.StorageRecord, fromVersion int) {
	if game == nil {
		r.Error(errors.New("Couldn't find game"))
		return
	}

	if playerIndex == invalidPlayerIndex {
		r.Error(errors.New("Got invalid playerIndex"))
		return
	}

	if gameInfo == nil {
		r.Error(errors.New("Game info could not be fetched"))
		return
	}

	isOwner := false

	if user != nil {
		isOwner = gameInfo.Owner == user.ID
	}

	state := game.CurrentState()

	//If it's the first load and no player moves have been applied, fetch the
	//first version only so that the other moves can be fetched and then applied.
	if fromVersion == 0 {
		//We check fromVersion because sometimes we re-load info because login
		//state changed, and that shouldn't give the early version, but the
		//proper version.
		if state.Version() != 0 {
			if lastMove, err := game.Move(state.Version()); err == nil {
				if lastMove.Info().Initiator() == 1 {
					//We're in a special case where no player moves have been applied yet since the beginning of the game.
					//To ensure that the animation delays from the first moves (e.g. dealing out cards) actually play, load up state 0 and return that.
					state = game.State(0)
				}
			}
		}
	}

	args := gin.H{
		"Chest":           game.Manager().Chest(),
		"Forms":           s.generateForms(game),
		"Game":            game.JSONForPlayer(playerIndex, state),
		"Error":           s.lastErrorMessage,
		"Players":         s.gamePlayerInfo(game.StorageRecord(), game.Manager()),
		"ViewingAsPlayer": playerIndex,
		"HasEmptySlots":   hasEmptySlots,
		"GameOpen":        gameInfo.Open,
		"GameVisible":     gameInfo.Visible,
		"IsOwner":         isOwner,
	}

	s.lastErrorMessage = ""

	r.Success(args)

}

func (s *Server) moveHandler(c *gin.Context) {

	r := s.newRenderer(c)

	if c.Request.Method != http.MethodPost {
		r.Error(errors.New("this method only supports post"))
		return
	}

	game := s.getGame(c)

	if game == nil {
		r.Error(errors.New("Game not found"))
		return
	}

	proposer := s.effectivePlayerIndex(c)

	move, err := s.getMoveFromForm(c, game)

	if move == nil {

		//TODO: move this to doMakeMove once getMoveFromForm is refactored correctly.

		errString := "No move returned"

		if err != nil {
			errString = err.Error()
		}

		r.Error(errors.New("Couldn't get move: " + errString))
		return
	}

	s.doMakeMove(r, game, proposer, move)

}

func (s *Server) doMakeMove(r *renderer, game *boardgame.Game, proposer boardgame.PlayerIndex, move boardgame.Move) {

	if err := <-game.ProposeMove(move, proposer); err != nil {

		if f, ok := err.(*errors.Friendly); ok {
			r.Error(f)
		} else {
			r.Error(errors.New(err.Error()))
		}
		return
	}
	//TODO: it would be nice if we could show which fixup moves we made, too,
	//somehow.

	r.Success(nil)
}

func (s *Server) generateForms(game *boardgame.Game) []*moveForm {

	var result []*moveForm

	for _, move := range game.Moves() {

		if base.IsFixUp(move) {
			continue
		}

		moveItem := &moveForm{
			Name:     move.Info().Name(),
			HelpText: move.HelpText(),
			Fields:   formFields(move),
		}
		result = append(result, moveItem)
	}

	return result
}

func formFields(move boardgame.Move) []*moveFormField {

	var result []*moveFormField

	for fieldName, fieldType := range move.ReadSetter().Props() {

		val, _ := move.ReadSetter().Prop(fieldName)

		info := &moveFormField{
			Name:         fieldName,
			Type:         fieldType,
			DefaultValue: val,
		}

		if fieldType == boardgame.TypeEnum {
			enumVal, _ := move.ReadSetter().EnumProp(fieldName)
			if enumVal != nil {
				info.EnumName = enumVal.Enum().Name()
			}
		}

		result = append(result, info)

	}

	return result
}

//genericHandler doesn't do much. We just register it so we automatically get
//CORS handlers triggered with the middelware.
func (s *Server) genericHandler(c *gin.Context) {
	r := s.newRenderer(c)
	r.Success(gin.H{
		"Message": "Nothing to see here.",
	})
}

//Start is where you start the server, and it never returns until it's time to shut down.
func (s *Server) Start() {

	config, err := config.Get("", false)

	for _, o := range s.overriders {
		config.AddOverride(o)
	}

	if err != nil {
		s.logger.Errorln("Configuration error: " + err.Error())
		return
	}

	s.logger.Infoln("Environment Variables")
	//Dbug print out the current environment
	for _, config := range os.Environ() {
		s.logger.Infoln("Environ:", config)
	}

	if v := os.Getenv("GIN_MODE"); v == "release" {
		s.logger.Infoln("Using release mode config")
		s.config = config.Prod
	} else {
		s.logger.Infoln("Using dev mode config")
		s.config = config.Dev
		s.logger.SetLevel(logrus.DebugLevel)
	}

	if s.config.Firebase == nil {
		s.logger.Errorln("No firebase config provided in active mode. Required for auth.")
		return
	}

	s.logger.Infoln("Derived config: " + s.config.String())

	name := s.storage.Name()

	storageConfig := s.config.Storage[name]

	s.logger.Infoln("Connecting to storage", name, "with config '"+storageConfig+"'")

	if err := s.storage.Connect(storageConfig); err != nil {
		s.logger.Fatalln("Couldnt' connect to storage manager: ", err)
		return
	}

	s.notifier = newVersionNotifier(s)

	router := gin.New()

	router.Use(gin.Recovery(), gin.LoggerWithWriter(os.Stdout, "/_ah/health"))

	router.NoRoute(s.genericHandler)
	router.Use(cors.Middleware(cors.Config{
		Origins:        s.config.AllowedOrigins,
		RequestHeaders: "content-type, Origin",
		ExposedHeaders: "content-type",
		Methods:        "GET, POST",
		Credentials:    true,
	}))

	//We have everything prefixed by /api just in case at some point we do
	//want to host both static and api on the same logical server.
	mainGroup := router.Group("/api")
	mainGroup.Use(s.userSetup)

	{
		mainGroup.GET("list/game", s.listGamesHandler)
		mainGroup.GET("list/manager", s.listManagerHandler)

		mainGroup.POST("auth", s.authCookieHandler)

		protectedMainGroup := mainGroup.Group("")
		protectedMainGroup.Use(s.requireLoggedIn)
		protectedMainGroup.POST("new/game", s.newGameHandler)

		gameAPIGroup := mainGroup.Group("game/:name/:id")
		gameAPIGroup.Use(s.gameAPISetup)
		{
			gameAPIGroup.GET("socket", s.socketHandler)
			gameAPIGroup.GET("info", s.gameInfoHandler)
			gameAPIGroup.GET("version/:version", s.gameVersionHandler)

			//The statusHandler is conceptually here, but becuase we want to
			//optimize it so much we have it congfigured at the top level.

			protectedGameAPIGroup := gameAPIGroup.Group("")
			protectedGameAPIGroup.Use(s.requireLoggedIn)
			protectedGameAPIGroup.POST("move", s.moveHandler)
			protectedGameAPIGroup.POST("join", s.joinGameHandler)
			protectedGameAPIGroup.POST("configure", s.configureGameHandler)
		}
	}

	if p := os.Getenv("PORT"); p != "" {
		router.Run(":" + p)
	} else {
		router.Run(":" + s.config.DefaultPort)
	}

}
