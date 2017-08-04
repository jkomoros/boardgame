package api

import (
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/itsjamie/gin-cors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/errors"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/jkomoros/boardgame/server/api/users"
	"github.com/jkomoros/boardgame/server/config"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

type Server struct {
	managers managerMap
	storage  *ServerStorageManager
	//We store the last error so that next time viewHandler is called we can
	//display it. Yes, this is a hack.
	lastErrorMessage string
	config           *config.ConfigMode

	upgrader websocket.Upgrader

	notifier *versionNotifier
}

type Renderer struct {
	c            *gin.Context
	rendered     bool
	cookieCalled bool
	cookieValue  string
}

type MoveForm struct {
	Name     string
	HelpText string
	Fields   []*MoveFormField
}

type MoveFormFieldType int

type MoveFormField struct {
	Name         string
	Type         boardgame.PropertyType
	DefaultValue interface{}
}

type managerMap map[string]*boardgame.GameManager

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
NewDefaultStorageManager or NewServerStorageManager.

Use it like so:

	func main() {
		storage := server.NewDefaultStorageManager()
		defer storage.Close()
		server.NewServer(storage, mygame.NewManager(storage)).Start()
	}

*/
func NewServer(storage *ServerStorageManager, managers ...*boardgame.GameManager) *Server {

	result := &Server{
		managers: make(managerMap),
		storage:  storage,
	}

	storage.server = result

	for _, manager := range managers {
		name := manager.Delegate().Name()
		result.managers[name] = manager
		if manager.Storage() != storage {
			log.Println("The storage for one of the managers was not the same item passed in as major storage.")
			return nil
		}
	}

	result.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     result.checkOriginForSocket,
	}

	return result

}

func NewRenderer(c *gin.Context) *Renderer {
	return &Renderer{
		c,
		false,
		false,
		"",
	}
}

func (r *Renderer) Error(f *errors.Friendly) {
	if r.rendered {
		panic("Error called on already-rendered renderer")
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

	r.rendered = true
}

func (r *Renderer) Success(keys gin.H) {

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

func (r *Renderer) writeCookie() {
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
func (r *Renderer) SetAuthCookie(value string) {

	//We don't write the cookies to the response yet because we might get
	//multiple SetAuthCookie calls in one response.

	r.cookieCalled = true
	r.cookieValue = value

}

func (s *Server) userSetup(c *gin.Context) {
	cookie := s.getRequestCookie(c)

	if cookie == "" {
		log.Println("No cookie set")
		return
	}

	user := s.storage.GetUserByCookie(cookie)

	if user == nil {
		log.Println("No user associated with that cookie")
		return
	} else {
		user.LastSeen = time.Now().UnixNano()
		s.storage.UpdateUser(user)
	}

	s.setUser(c, user)

	s.setAdminAllowed(c, s.calcAdminAllowed(user))
}

func (s *Server) gameFromId(gameId, gameName string) *boardgame.Game {

	manager := s.managers[gameName]

	if manager == nil {
		log.Println("Couldnt' find manager for", gameName)
		return nil
	}

	game := manager.Game(gameId)

	//TODO: figure out a way to return a meaningful error

	if game == nil {
		log.Println("Couldn't find game with id", gameId)
		return nil
	}

	if game.Name() != gameName {
		log.Println("The name of the game was not what we were expecting. Wanted", gameName, "got", game.Name())
		return nil
	}

	return game
}

//gameAPISetup fetches the game configured in the URL and puts it in context.
func (s *Server) gameAPISetup(c *gin.Context) {

	id := s.getRequestGameId(c)

	gameName := s.getRequestGameName(c)

	game := s.gameFromId(id, gameName)

	if game == nil {
		return
	}

	s.setGame(c, game)

	userIds := s.storage.UserIdsForGame(id)

	if userIds == nil {
		log.Println("No userIds associated with game")
	}

	user := s.getUser(c)

	if user == nil {
		log.Println("No user provided")
		//The rest of the flow will handle a nil user fine
	}

	effectiveViewingAsPlayer, emptySlots := s.calcViewingAsPlayerAndEmptySlots(userIds, user, game.Agents())

	if user != nil && effectiveViewingAsPlayer == boardgame.ObserverPlayerIndex && len(emptySlots) > 0 && len(emptySlots) == game.NumPlayers()-game.NumAgentPlayers() {
		//Special case: we're the first player, we likely just created it. Just join the thing!

		slot := emptySlots[0]

		if err := s.storage.SetPlayerForGame(game.Id(), slot, user.Id); err != nil {
			log.Println("Tried to set the user as player " + slot.String() + " but failed: " + err.Error())
			return
		} else {
			s.setHasEmptySlots(c, false)
			effectiveViewingAsPlayer = slot
		}

	} else {
		s.setHasEmptySlots(c, len(emptySlots) != 0)
	}
	s.setViewingAsPlayer(c, effectiveViewingAsPlayer)

}

//Checks to make sure the user is logged in, fails if not.
func (s *Server) requireLoggedIn(c *gin.Context) {

	r := NewRenderer(c)

	user := s.getUser(c)

	if user == nil {
		r.Error(errors.NewFriendly("Not logged in"))
		c.Abort()
		return
	}

	//All good!
}

func (s *Server) joinGameHandler(c *gin.Context) {
	r := NewRenderer(c)

	game := s.getGame(c)

	if game == nil {
		r.Error(errors.NewFriendly("No such game"))
		return
	}

	user := s.getUser(c)

	userIds := s.storage.UserIdsForGame(game.Id())

	viewingAsPlayer, emptySlots := s.calcViewingAsPlayerAndEmptySlots(userIds, user, game.Agents())

	s.doJoinGame(r, game, viewingAsPlayer, emptySlots, user)

}

func (s *Server) doJoinGame(r *Renderer, game *boardgame.Game, viewingAsPlayer boardgame.PlayerIndex, emptySlots []boardgame.PlayerIndex, user *users.StorageRecord) {

	if user == nil {
		r.Error(errors.New("No user provided."))
		return
	}

	eGame, err := s.storage.ExtendedGame(game.Id())

	if err != nil {
		r.Error(errors.New("Couldn't get extended information about game: " + err.Error()))
		return
	}

	if !eGame.Open {
		r.Error(errors.NewFriendly("The game is not open to people joining."))
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

	if err := s.storage.SetPlayerForGame(game.Id(), slot, user.Id); err != nil {
		r.Error(errors.New("Tried to set the user as player " + slot.String() + " but failed: " + err.Error()))
		return
	}

	r.Success(nil)

}

func (s *Server) newGameHandler(c *gin.Context) {

	r := NewRenderer(c)

	managerId := s.getRequestManager(c)

	numPlayers := s.getRequestNumPlayers(c)

	manager := s.managers[managerId]

	if numPlayers == 0 && manager != nil {
		numPlayers = manager.Delegate().DefaultNumPlayers()
	}

	agents := s.getRequestAgents(c, numPlayers)

	owner := s.getUser(c)

	open := s.getRequestOpen(c)
	visible := s.getRequestVisible(c)

	s.doNewGame(r, owner, manager, numPlayers, agents, open, visible)

}

func (s *Server) doNewGame(r *Renderer, owner *users.StorageRecord, manager *boardgame.GameManager, numPlayers int, agents []string, open bool, visible bool) {

	if manager == nil {
		r.Error(errors.New("No manager provided"))
		return
	}

	if owner == nil {
		r.Error(errors.NewFriendly("You must be signed in to create a game."))
		return
	}

	game := manager.NewGame()

	if game == nil {
		r.Error(errors.New("No game could be created"))
		return
	}

	if err := game.SetUp(numPlayers, agents); err != nil {
		//TODO: communicate the error state back to the client in a sane way
		if f, ok := err.(*errors.Friendly); ok {
			r.Error(f)
		} else {
			r.Error(errors.New(err.Error()))
		}
		return
	}

	eGame, err := s.storage.ExtendedGame(game.Id())

	if err != nil {
		r.Error(errors.New("Couldn't retrieve saved game: " + err.Error()))
		return
	}

	eGame.Owner = owner.Id
	eGame.Open = open
	eGame.Visible = visible

	//TODO: set Open, Visible based on query params.

	if err := s.storage.UpdateExtendedGame(game.Id(), eGame); err != nil {
		r.Error(errors.New("Couldn't save extended game metadata: " + err.Error()))
		return
	}

	r.Success(gin.H{
		"GameId":   game.Id(),
		"GameName": game.Name(),
	})
}

func (s *Server) listGamesHandler(c *gin.Context) {

	r := NewRenderer(c)

	user := s.getUser(c)

	gameName := s.getRequestGameName(c)

	adminAllowed := s.getAdminAllowed(c)
	requestAdmin := s.getRequestAdmin(c)

	isAdmin := s.calcIsAdmin(adminAllowed, requestAdmin)

	s.doListGames(r, user, gameName, isAdmin)
}

func (s *Server) doListGames(r *Renderer, user *users.StorageRecord, gameName string, isAdmin bool) {
	var userId string
	if user != nil {
		userId = user.Id
	}
	result := gin.H{
		"ParticipatingActiveGames":   s.listGamesWithUsers(100, listing.ParticipatingActive, userId, gameName),
		"ParticipatingFinishedGames": s.listGamesWithUsers(100, listing.ParticipatingFinished, userId, gameName),
		"VisibleJoinableActiveGames": s.listGamesWithUsers(100, listing.VisibleJoinableActive, userId, gameName),
		"VisibleActiveGames":         s.listGamesWithUsers(100, listing.VisibleActive, userId, gameName),
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

func (s *Server) listGamesWithUsers(max int, list listing.Type, userId string, gameName string) []*gameStorageRecordWithUsers {
	games := s.storage.ListGames(max, list, userId, gameName)

	result := make([]*gameStorageRecordWithUsers, len(games))

	for i, game := range games {

		manager := s.managers[game.Name]

		//When SecretSalt is empty it will be omitted from the JSON output.

		//TODO: isn't it brittle that we only sanitize the critically
		//important SecretSalt here?
		game.SecretSalt = ""

		result[i] = &gameStorageRecordWithUsers{
			game,
			s.gamePlayerInfo(&game.GameStorageRecord, manager),
			humanize.Time(time.Unix(0, game.LastActivity)),
		}
	}

	return result

}

func (s *Server) listManagerHandler(c *gin.Context) {
	r := NewRenderer(c)
	s.doListManager(r)
}

func (s *Server) doListManager(r *Renderer) {
	type agentInfo struct {
		Name        string
		DisplayName string
	}
	type managerInfo struct {
		Name              string
		DisplayName       string
		DefaultNumPlayers int
		Agents            []agentInfo
	}
	var managers []managerInfo
	for name, manager := range s.managers {
		agents := make([]agentInfo, len(manager.Agents()))
		for i, agent := range manager.Agents() {
			agents[i] = agentInfo{
				agent.Name(),
				agent.DisplayName(),
			}
		}
		managers = append(managers, managerInfo{
			Name:              name,
			DisplayName:       manager.Delegate().DisplayName(),
			DefaultNumPlayers: manager.Delegate().DefaultNumPlayers(),
			Agents:            agents,
		})
	}

	sort.Slice(managers, func(i, j int) bool {
		return managers[i].Name < managers[j].Name
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

	r := NewRenderer(c)

	s.doGameVersion(r, game, version, fromVersion, playerIndex, autoCurrentPlayer)

}

func (s *Server) doGameVersion(r *Renderer, game *boardgame.Game, version, fromVersion int, playerIndex boardgame.PlayerIndex, autoCurrentPlayer bool) {
	if game == nil {
		r.Error(errors.NewFriendly("Couldn't find game"))
		return
	}

	if playerIndex == invalidPlayerIndex {
		r.Error(errors.New("Got invalid playerIndex"))
		return
	}

	//TODO: do something with fromVersion. Note that fromVersion might equal
	//version if playerIndex or autoCurrentPlayer changed.

	state := game.State(version)

	if autoCurrentPlayer {
		newPlayerIndex := game.Manager().Delegate().CurrentPlayerIndex(state)
		if newPlayerIndex.Valid(state) {
			playerIndex = newPlayerIndex
		}
	}

	//If state is nil, JSONForPlayer will basically treat it as just "give the
	//current version" which is a reasonable fallback.

	args := gin.H{
		"Bundles": []gin.H{
			gin.H{
				"Game":            game.JSONForPlayer(playerIndex, state),
				"ViewingAsPlayer": playerIndex,
				"Forms":           s.generateForms(game),
			},
		},
	}

	r.Success(args)
}

func (s *Server) configureGameHandler(c *gin.Context) {
	game := s.getGame(c)

	var gameId string

	if game != nil {
		gameId = game.Id()
	}

	gameInfo, _ := s.storage.ExtendedGame(gameId)

	adminAllowed := s.getAdminAllowed(c)
	requestAdmin := s.getRequestAdmin(c)

	isAdmin := s.calcIsAdmin(adminAllowed, requestAdmin)

	user := s.getUser(c)

	open := s.getRequestOpen(c)
	visible := s.getRequestVisible(c)

	r := NewRenderer(c)

	s.doConfigureGame(r, user, isAdmin, game, gameInfo, open, visible)

}

func (s *Server) doConfigureGame(r *Renderer, user *users.StorageRecord, isAdmin bool, game *boardgame.Game, gameInfo *extendedgame.StorageRecord, open, visible bool) {

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

	if !isAdmin && user.Id != gameInfo.Owner {
		r.Error(errors.NewFriendly("You are neither the owner nor an admin."))
		return
	}

	gameInfo.Open = open
	gameInfo.Visible = visible

	if err := s.storage.UpdateExtendedGame(game.Id(), gameInfo); err != nil {
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

	var gameId string

	if game != nil {
		gameId = game.Id()
	}

	//TODO: should this be done in gameAPISetup?
	gameInfo, _ := s.storage.ExtendedGame(gameId)

	user := s.getUser(c)

	r := NewRenderer(c)

	s.doGameInfo(r, game, playerIndex, hasEmptySlots, gameInfo, user)

}

type playerBoardInfo struct {
	DisplayName string
	IsAgent     bool
	IsEmpty     bool
	PhotoUrl    string
}

func (s *Server) gamePlayerInfo(game *boardgame.GameStorageRecord, manager *boardgame.GameManager) []*playerBoardInfo {

	if manager == nil {
		return nil
	}

	result := make([]*playerBoardInfo, game.NumPlayers)

	userIds := s.storage.UserIdsForGame(game.Id)
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

		userId := userIds[i]

		if userId == "" {
			player.IsEmpty = true
			player.IsAgent = false
			player.DisplayName = ""
			continue
		}

		user := s.storage.GetUserById(userId)

		if user == nil {
			player.IsAgent = false
			player.IsEmpty = false
			player.DisplayName = "Unknown user"
			continue
		}

		player.IsAgent = false
		player.IsEmpty = false
		player.PhotoUrl = user.PhotoUrl
		player.DisplayName = user.EffectiveDisplayName()

		if player.DisplayName == "" {
			player.DisplayName = "Player " + strconv.Itoa(i)
		}

	}

	return result
}

func (s *Server) doGameInfo(r *Renderer, game *boardgame.Game, playerIndex boardgame.PlayerIndex, hasEmptySlots bool, gameInfo *extendedgame.StorageRecord, user *users.StorageRecord) {
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
		isOwner = gameInfo.Owner == user.Id
	}

	args := gin.H{
		"Chest":           s.renderChest(game),
		"Forms":           s.generateForms(game),
		"Game":            game.JSONForPlayer(playerIndex, nil),
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

	r := NewRenderer(c)

	if c.Request.Method != http.MethodPost {
		r.Error(errors.New("This method only supports post."))
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

func (s *Server) doMakeMove(r *Renderer, game *boardgame.Game, proposer boardgame.PlayerIndex, move boardgame.Move) {

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

func (s *Server) generateForms(game *boardgame.Game) []*MoveForm {

	var result []*MoveForm

	for _, moveType := range game.Manager().PlayerMoveTypes() {

		moveItem := &MoveForm{
			Name:     moveType.Name(),
			HelpText: moveType.HelpText(),
			Fields:   formFields(moveType.NewMove(game.CurrentState())),
		}
		result = append(result, moveItem)
	}

	return result
}

func formFields(move boardgame.Move) []*MoveFormField {

	var result []*MoveFormField

	for fieldName, fieldType := range move.ReadSetter().Props() {

		val, _ := move.ReadSetter().Prop(fieldName)

		result = append(result, &MoveFormField{
			Name:         fieldName,
			Type:         fieldType,
			DefaultValue: val,
		})

	}

	return result
}

func (s *Server) renderChest(game *boardgame.Game) map[string][]interface{} {
	//Substantially copied from cli.renderChest().

	deck := make(map[string][]interface{})

	for _, name := range game.Chest().DeckNames() {

		components := game.Chest().Deck(name).Components()

		values := make([]interface{}, len(components))

		for i, component := range components {
			values[i] = struct {
				Index  int
				Values interface{}
			}{
				i,
				component.Values,
			}
		}

		deck[name] = values
	}

	return deck
}

//genericHandler doesn't do much. We just register it so we automatically get
//CORS handlers triggered with the middelware.
func (s *Server) genericHandler(c *gin.Context) {
	r := NewRenderer(c)
	r.Success(gin.H{
		"Message": "Nothing to see here.",
	})
}

//Start is where you start the server, and it never returns until it's time to shut down.
func (s *Server) Start() {

	config, err := config.Get()

	if err != nil {
		log.Println("Configuration error: " + err.Error())
	}

	log.Println("Environment Variables")
	//Dbug print out the current environment
	for _, config := range os.Environ() {
		log.Println("Environ:", config)
	}

	if v := os.Getenv("GIN_MODE"); v == "release" {
		log.Println("Using release mode config")
		s.config = config.Prod
	} else {
		log.Println("Using dev mode config")
		s.config = config.Dev
	}

	name := s.storage.Name()

	storageConfig := s.config.StorageConfig[name]

	log.Println("Connecting to storage", name, "with config '"+storageConfig+"'")

	if err := s.storage.Connect(storageConfig); err != nil {
		log.Println("Couldnt' connect to storage manager: ", err)
		return
	}

	s.notifier = newVersionNotifier()

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
