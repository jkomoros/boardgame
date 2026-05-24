package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
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

// Server is the main server object.
type Server struct {
	managers managerMap

	//map of game ID to players to seat. Protected by mu.
	playersToSeat map[string][]*playerToSeat

	//track which games have had "Game over!" system message emitted. Protected by mu.
	gameOverEmitted map[string]bool

	// mu protects playersToSeat and gameOverEmitted from concurrent access.
	// These maps are written from game goroutines (via PlayerMoveApplied /
	// ForceFixUp callbacks) and HTTP handler goroutines simultaneously.
	mu sync.Mutex

	storage *ServerStorageManager
	//We store the last error so that next time viewHandler is called we can
	//display it. Yes, this is a hack.
	lastErrorMessage string
	config           *config.Mode

	overriders []config.OptionOverrider

	upgrader websocket.Upgrader

	notifier *versionNotifier
	logger   *logrus.Logger

	// joinRateLimiter throttles /api/join and /api/join/seat per client IP.
	// 10 requests / minute is plenty for legitimate use (entering the code,
	// then claiming a seat) and tight enough to discourage brute-force
	// enumeration of the 234k 4-letter code namespace (spec §6.1).
	joinRateLimiter *rateLimiter

	// seatJoinLocks serializes seat-claim operations on a per-game basis so
	// concurrent /api/join/seat requests for the same room can't race
	// through the empty-slot lookup + SeatPlayer proposal. Spec §11 race-
	// resolution behavior. Lazy-allocated by getSeatJoinLock.
	seatJoinLocks   map[string]*sync.Mutex
	seatJoinLocksMu sync.Mutex
}

type renderer struct {
	s            *Server
	c            *gin.Context
	rendered     bool
	cookieCalled bool
	cookieValue  string
}

type moveForm struct {
	Name                string
	HelpText            string
	Fields              []*moveFormField
	LegalForPlayer      bool   `json:",omitempty"`
	LegalForPlayerError string `json:",omitempty"`
	LegalForAnyone      bool   `json:",omitempty"`
	IsGatheringStart    bool   `json:",omitempty"`
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
	// supportsTableHandMode is true iff this game ships a
	// boardgame-render-game-<name>-table.ts AND -hand.ts pair (detected at
	// build time by boardgame-util; surfaced into the generated api/main.go
	// via Server.WithCompanionCapableGames). Used by doListManager so the
	// create-game form can show the "Use shared projector + phones" toggle
	// for supporting games (spec §5.3), and by /api/game/.../new-style
	// creation requests for server-side validation that a request for
	// companionMode is for an actually-supporting game.
	supportsTableHandMode bool
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
		managers:        make(managerMap),
		playersToSeat:   make(map[string][]*playerToSeat),
		gameOverEmitted: make(map[string]bool),
		storage:         storage,
		logger:          logger,
		// 10 requests / minute = 1 every 6 seconds steady-state, with a
		// burst capacity of 10. Idle buckets evict after 10 minutes.
		joinRateLimiter: newRateLimiter(10, 10.0/60.0, 10*time.Minute),
		seatJoinLocks:   make(map[string]*sync.Mutex),
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

// by defining the variable type, we verify we actually do implement the
// interface. Since it flows via FetchInejctedData, there's no type
// checking otherwise.
var testPlayerSeat interfaces.SeatPlayerSignaler = &playerToSeat{}

func (p *playerToSeat) SeatIndex() boardgame.PlayerIndex {
	return p.seatIndex
}

func (p *playerToSeat) Committed() {
	p.s.mu.Lock()
	defer p.s.mu.Unlock()

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

// managerSeatPlayerMoves returns the move names for the given manager that are a
// seat player move. If len(result) is 0, then the game does not have a seat
// player move.
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

// SetAuthCookie will set the auth cookie to the specified value. If called
// multiple times for a single request will only actually write headers for the
// last one.
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
		s.logger.Errorln("Couldn't find manager for", gameName)
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

// maybeReopenGame checks whether a game that was previously closed has open
// seats again (e.g., between rounds when ActivateEmptySeat fires). If so, it
// sets the game back to Open so new players can join.
func (s *Server) maybeReopenGame(record *boardgame.GameStorageRecord) {
	managerInfo := s.managers[record.Name]
	if managerInfo == nil || !managerInfo.playerHasSeat {
		return
	}

	game := managerInfo.manager.Game(record.ID)
	if game == nil {
		return
	}

	closedSeats := s.closedSeatsForGame(game)
	userIds := s.storage.UserIDsForGame(game.ID())
	agents := game.Agents()

	for i, uid := range userIds {
		if uid == "" && agents[i] == "" && !closedSeats[i] {
			// There's an open, unfilled, non-agent slot.
			eGame, err := s.storage.ExtendedGame(game.ID())
			if err == nil && !eGame.Open {
				eGame.Open = true
				if err := s.storage.UpdateExtendedGame(game.ID(), eGame); err != nil {
					s.logger.Errorln("Failed to reopen game:", err)
				} else {
					s.logger.Infoln("Reopened game", game.ID(), "because seats are available again")
				}
			}
			return
		}
	}
}

// closedSeatsForGame will return a slice of bools of equal length to the game's
// NumPlayers, where each one is set to true if the playerState has a Seat and
// the seat is marked as closed.
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

// getOrCreateDebugUser returns a synthetic user record for debug auto-seating.
// Creates the user in storage if it doesn't already exist.
func (s *Server) getOrCreateDebugUser(index int) *users.StorageRecord {
	id := fmt.Sprintf("debug-player-%d", index)
	user := s.storage.GetUserByID(id)
	if user == nil {
		now := time.Now().UnixNano()
		user = &users.StorageRecord{
			ID:          id,
			DisplayName: fmt.Sprintf("Debug Player %d", index),
			Created:     now,
			LastSeen:    now,
		}
		if err := s.storage.UpdateUser(user); err != nil {
			s.logger.Warnln("Couldn't create debug user:", err)
		}
	}
	return user
}

// autoCloseGameIfFull checks whether every seat in the game is filled (by a
// human or agent) or closed. If so it marks the game as not-open (3b) and
// removes any pending players from the seating queue (3a).
func (s *Server) autoCloseGameIfFull(game *boardgame.Game) {
	userIds := s.storage.UserIDsForGame(game.ID())
	closedSeats := s.closedSeatsForGame(game)
	agents := game.Agents()

	for i, uid := range userIds {
		if uid == "" && agents[i] == "" && !closedSeats[i] {
			// There is still an open, unfilled, non-agent slot.
			return
		}
	}

	// All seats are filled, occupied by agents, or closed.

	// 3b: Auto-close the game so no one else tries to join.
	eGame, err := s.storage.ExtendedGame(game.ID())
	if err == nil && eGame.Open {
		eGame.Open = false
		if err := s.storage.UpdateExtendedGame(game.ID(), eGame); err != nil {
			s.logger.Errorln("Failed to auto-close full game:", err)
		} else {
			s.logger.Infoln("Auto-closed game", game.ID(), "because all seats are filled or closed")
		}
	}

	// 3a: Clean up any pending players that can no longer be seated.
	s.mu.Lock()
	pending, hasPending := s.playersToSeat[game.ID()]
	if hasPending && len(pending) > 0 {
		s.logger.Warnln("Removing", len(pending), "pending player(s) from full game", game.ID())
		delete(s.playersToSeat, game.ID())
	}
	s.mu.Unlock()
}

// gameAPISetup fetches the game configured in the URL and puts it in context.
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
		effectiveViewingAsPlayer = slot

		// Debug auto-seating: when DisableAdminChecking is true (dev mode),
		// auto-fill all remaining empty non-agent slots with synthetic debug
		// users. This lets developers test multiplayer games without needing
		// multiple browser sessions. (Issue #774)
		if s.config.DisableAdminChecking {
			for _, debugSlot := range emptySlots[1:] { // skip slot 0 (already seated above)
				if game.Agents()[debugSlot] != "" {
					continue // skip agent slots
				}
				debugUser := s.getOrCreateDebugUser(int(debugSlot))
				if err := s.doSeatPlayer(game, debugSlot, debugUser); err != nil {
					s.logger.Warnln("Debug auto-seat failed for slot", debugSlot, ":", err)
				}
			}
		}

		// Re-check empty slots after seating.
		remainingEmptySlots := 0
		if !s.config.DisableAdminChecking {
			remainingEmptySlots = len(emptySlots) - 1 // only the real player was seated
		}
		s.setHasEmptySlots(c, remainingEmptySlots > 0)

		s.autoCloseGameIfFull(game)

	} else {
		s.setHasEmptySlots(c, len(emptySlots) != 0)
		if len(emptySlots) == 0 {
			s.autoCloseGameIfFull(game)
		}
	}
	s.setViewingAsPlayer(c, effectiveViewingAsPlayer)

}

// Checks to make sure the user is logged in, fails if not.
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

		s.mu.Lock()
		s.playersToSeat[gameID] = append(s.playersToSeat[gameID], player)
		s.mu.Unlock()

		//Now we have information waiting for SeatPlayer. Tell the engine to
		//check whether fixups need to be applied, becuase we know that
		//something outside of state has changed that might change whether moves
		//are valid. We don't have to worry about race conditions because Game's
		//mainLoop will make sure this isn't triggered while another move is
		//being processed.
		delayed := game.Manager().Internals().ForceFixUp(game)

		go func() {
			if err := <-delayed; err != nil {
				log.Println("Forced Fix Up failed: " + err.Error())
			}
		}()

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
		return
	}

	s.autoCloseGameIfFull(game)

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
	companionMode := s.getRequestCompanionMode(c)

	// Server-side validation that this manager actually supports
	// Table+Hand mode. A malicious client could POST companionMode=1 for
	// a game whose static dir doesn't ship -table.ts/-hand.ts; refuse
	// before generating a room code that would route to nowhere.
	if companionMode {
		mInfo := s.managers[managerID]
		if mInfo == nil || !mInfo.supportsTableHandMode {
			r.Error(errors.NewFriendly("This game does not support shared-projector mode."))
			return
		}
	}

	s.doNewGame(r, c, owner, manager, numPlayers, agents, open, visible, variant, companionMode)

}

func (s *Server) doNewGame(r *renderer, c *gin.Context, owner *users.StorageRecord, manager *boardgame.GameManager, numPlayers int, agents []string, open bool, visible bool, variant map[string]string, companionMode bool) {

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

	// If companion-mode was requested, generate a room code and set the
	// host's surface=table cookie so their browser loads the projected
	// renderer on the redirect. Solo-mode games leave CompanionRoomCode
	// empty + don't get a surface cookie (renderer falls back to solo).
	var roomCode string
	if companionMode {
		code, codeErr := GenerateRoomCode(func(candidate string) (bool, error) {
			id, gerr := s.storage.GameByRoomCode(candidate)
			if gerr != nil {
				return false, gerr
			}
			return id != "", nil
		})
		if codeErr != nil {
			s.logger.Warnln("Failed to generate room code:", codeErr)
			r.Error(errors.NewFriendly("Couldn't generate a room code; please try again."))
			return
		}
		eGame.CompanionRoomCode = code
		roomCode = code
		// Surface cookie scoped to the gameID — see surfaceCookieName().
		// Path "/" so the loader sees it on the game page.
		s.setSurfaceCookie(c, game.ID(), "table")
	}

	if err := s.storage.UpdateExtendedGame(game.ID(), eGame); err != nil {
		r.Error(errors.New("Couldn't save extended game metadata: " + err.Error()))
		return
	}

	resp := gin.H{
		"GameID":   game.ID(),
		"GameName": game.Name(),
	}
	if roomCode != "" {
		resp["CompanionRoomCode"] = roomCode
	}
	r.Success(resp)
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
			"Name":                  name,
			"DisplayName":           manager.Delegate().DisplayName(),
			"Description":           manager.Delegate().Description(),
			"DefaultNumPlayers":     manager.Delegate().DefaultNumPlayers(),
			"MinNumPlayers":         manager.Delegate().MinNumPlayers(),
			"MaxNumPlayers":         manager.Delegate().MaxNumPlayers(),
			"Agents":                agents,
			"Variant":               variant,
			"SupportsTableHandMode": mInfo.supportsTableHandMode,
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

func (s *Server) moveBundles(game *boardgame.Game, moves []*boardgame.MoveStorageRecord, playerIndex boardgame.PlayerIndex, autoCurrentPlayer bool) ([]gin.H, error) {
	var bundles []gin.H

	if len(moves) == 0 {
		moves = append(moves, nil)
	}

	for i, move := range moves {

		version := 0
		if move != nil {
			version = move.Version
		}

		//This is the state for the end of the bundle.
		state := game.State(version)

		if autoCurrentPlayer {
			newPlayerIndex := game.Manager().Delegate().CurrentPlayerIndex(state)
			if newPlayerIndex.Valid(state) {
				// AnyPlayerIndex means "any player can act" but there's no
				// single player whose perspective to show. Fall through to
				// the observer view instead of using it as a viewing player.
				if newPlayerIndex == boardgame.AnyPlayerIndex {
					playerIndex = boardgame.ObserverPlayerIndex
				} else {
					playerIndex = newPlayerIndex
				}
			}
		}

		gameJSON, err := game.JSONForPlayer(playerIndex, state)

		if err != nil {
			return nil, errors.New("Couldn't seralize json for " + strconv.Itoa(i) + ": " + err.Error())
		}

		//If state is nil, JSONForPlayer will basically treat it as just "give the
		//current version" which is a reasonable fallback.

		// Only compute legality for the last bundle (the state the player
		// will interact with). Intermediate animation bundles use plain
		// forms without legality — zero extra cost.
		var forms []*moveForm
		if i == len(moves)-1 {
			forms = s.generateFormsWithLegality(game, state, playerIndex)
		} else {
			forms = s.generateForms(game)
		}

		bundle := gin.H{
			"Game":            gameJSON,
			"Move":            move,
			"ViewingAsPlayer": playerIndex,
			"Forms":           forms,
		}

		bundles = append(bundles, bundle)

	}

	return bundles, nil
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

	bundles, err := s.moveBundles(game, moves, playerIndex, autoCurrentPlayer)

	if err != nil {
		r.Error(errors.New("Couldn't generate move bundles: " + err.Error()))
		return
	}

	r.Success(gin.H{
		"Bundles": bundles,
	})
}

// AddOverrides defines overrides that will be applied on top of the config we
// load. We return a reference to ourself to allow chaining of configurations.
func (s *Server) AddOverrides(overrides []config.OptionOverrider) *Server {
	s.overriders = append(s.overriders, overrides...)
	return s
}

// WithCompanionCapableGames marks the named games as supporting Table+Hand
// companion mode (spec §5.3). Called by the generated api/main.go with the
// list boardgame-util computed at build time from filesystem walk. Returns
// the server for chaining. Names that don't match any registered manager
// are silently ignored (so a stale capability list doesn't crash startup).
//
// Panics at boot if a companion-capable game does not register a move named
// "Force Finish Turn", or if that move is registered as a FixUp (which
// would cause infinite recursion in the fixup pipeline). See
// moves.ForceFinishTurn for the required registration pattern.
func (s *Server) WithCompanionCapableGames(gameNames []string) *Server {
	for _, name := range gameNames {
		mInfo, ok := s.managers[name]
		if !ok {
			continue
		}
		mInfo.supportsTableHandMode = true

		move := mInfo.manager.ExampleMoveByName("Force Finish Turn")
		if move == nil {
			panic(fmt.Sprintf(
				"boardgame/server: game %q is companion-capable (has -table.ts and -hand.ts) "+
					"but does not register a move named \"Force Finish Turn\". "+
					"Add: auto.MustConfig(new(moves.ForceFinishTurn), "+
					"moves.WithMoveName(\"Force Finish Turn\"), moves.WithIsFixUp(false))",
				name,
			))
		}
		type isFixUpper interface{ IsFixUp() bool }
		fixUpper, _ := move.(isFixUpper)
		if fixUpper != nil && fixUpper.IsFixUp() {
			panic(fmt.Sprintf(
				"boardgame/server: game %q registers \"Force Finish Turn\" as a FixUp move. "+
					"This will cause infinite recursion in the fixup pipeline because "+
					"ForceFinishTurn.Legal() returns nil for AdminPlayerIndex (the same "+
					"identity used by fixup proposers). Fix: add moves.WithIsFixUp(false) "+
					"to the auto.MustConfig call",
				name,
			))
		}
	}
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

// gameInfo is the first payload when a game is loaded, including immutables
// like chest, but also the initial game state payload as a convenience.
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

// collectSeatPresentations returns the per-seat avatar+name records for a
// game, indexed by playerIndex. Iterates numPlayers; absent rows (e.g.
// for unjoined slots or solo-mode games) return nil from the storage and
// are dropped from the result. Cheap V1 implementation; a single bulk
// "list by gameID" storage method would be more efficient at scale.
func (s *Server) collectSeatPresentations(gameID string, numPlayers int) []gin.H {
	var out []gin.H
	for i := 0; i < numPlayers; i++ {
		rec, err := s.storage.SeatPresentation(gameID, boardgame.PlayerIndex(i))
		if err != nil || rec == nil {
			continue
		}
		out = append(out, gin.H{
			"playerIndex": rec.PlayerIndex,
			"displayName": rec.DisplayName,
			"avatarSlug":  rec.AvatarSlug,
		})
	}
	return out
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

	gameJSON, err := game.JSONForPlayer(playerIndex, state)

	if err != nil {
		r.Error(errors.New("Couldn't serialize json: " + err.Error()))
		return
	}

	// Companion-mode bundle: everything the Table+Hand view bases need to
	// render avatar strips, "Waiting…" badges, room code, host controls.
	// Empty/zero for solo-mode games (CompanionRoomCode=="") — harmless.
	// Bundled into one CompanionInfo object so the client-side plumbing
	// is a single prop traversal rather than five.
	seatPresentations := s.collectSeatPresentations(game.ID(), game.NumPlayers())
	companionInfo := gin.H{
		"CompanionMode":     gameInfo.CompanionRoomCode != "",
		"RoomCode":          gameInfo.CompanionRoomCode,
		"RoomLocked":        gameInfo.CompanionLocked,
		"SeatPresentations": seatPresentations,
		// Absent is the list of player indices currently flagged absent by
		// the heartbeat scan (spec §9.1). The Table view uses this to draw
		// "Waiting for Alice (m:ss)" badges and to decide whether to show
		// the host SkipTurn button on the current-player badge.
		"Absent": s.notifier.AbsentPlayers(game.ID()),
	}

	args := gin.H{
		"Chest":           game.Manager().Chest(),
		"Forms":           s.generateFormsWithLegality(game, state, playerIndex),
		"Game":            gameJSON,
		"Error":           s.lastErrorMessage,
		"Players":         s.gamePlayerInfo(game.StorageRecord(), game.Manager()),
		"ViewingAsPlayer": playerIndex,
		"HasEmptySlots":   hasEmptySlots,
		"GameOpen":        gameInfo.Open,
		"GameVisible":     gameInfo.Visible,
		"IsOwner":         isOwner,
		"CompanionInfo":   companionInfo,
		//The StateVersion is almost always the Game.Version, except in the
		//special case described above where lots of fix up moves have been
		//applied but no player moves yet. State blobs used to include their own
		//state version but now we have to ship it down to the client speically.
		"StateVersion": state.Version(),
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
		if starter, ok := move.(interfaces.GatheringStartMover); ok && starter.IsGatheringStartMove() {
			moveItem.IsGatheringStart = true
		}
		result = append(result, moveItem)
	}

	return result
}

// generateFormsWithLegality is like generateForms but also computes legality
// information for each move against the given state and player. This tells the
// client which moves are currently legal (for enabling/disabling buttons) and
// which are structurally possible (for showing/hiding buttons).
func (s *Server) generateFormsWithLegality(game *boardgame.Game, state boardgame.ImmutableState, playerIndex boardgame.PlayerIndex) []*moveForm {
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
		if starter, ok := move.(interfaces.GatheringStartMover); ok && starter.IsGatheringStartMove() {
			moveItem.IsGatheringStart = true
		}

		// Legality for viewing player
		if playerIndex != boardgame.ObserverPlayerIndex {
			if err := move.Legal(state, playerIndex); err != nil {
				moveItem.LegalForPlayerError = err.Error()
			} else {
				moveItem.LegalForPlayer = true
			}
		}

		// Structural legality (admin bypasses proposer checks)
		if err := move.Legal(state, boardgame.AdminPlayerIndex); err == nil {
			moveItem.LegalForAnyone = true
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

// genericHandler doesn't do much. We just register it so we automatically get
// CORS handlers triggered with the middelware.
func (s *Server) genericHandler(c *gin.Context) {
	r := s.newRenderer(c)
	r.Success(gin.H{
		"Message": "Nothing to see here.",
	})
}

// augmentPolicyWithDMs adds DM channels to a ChatPolicy based on the actual
// user IDs seated in the game. DM channels are named "dm/userA/userB" with
// IDs sorted lexicographically so both parties get the same channel name.
func (s *Server) augmentPolicyWithDMs(policy boardgame.ChatPolicy, game *boardgame.Game, userID string) boardgame.ChatPolicy {
	config := game.Manager().Delegate().ChatConfig()
	if !config.DMChatEnabled() {
		return policy
	}

	userIDs := s.storage.UserIDsForGame(game.ID())
	if userIDs == nil || userID == "" {
		return policy
	}

	for _, otherID := range userIDs {
		if otherID == "" || otherID == userID {
			continue
		}
		// Sort lexicographically for canonical channel name
		a, b := userID, otherID
		if a > b {
			a, b = b, a
		}
		ch := "dm/" + a + "/" + b
		policy.SendChannels = append(policy.SendChannels, ch)
		policy.ViewChannels = append(policy.ViewChannels, ch)
	}

	return policy
}

// Start is where you start the server, and it never returns until it's time to shut down.
// chatStorage returns the ChatStorageManager if the storage backend supports
// it, or nil if not. Checks the underlying storage manager that the
// ServerStorageManager wraps.
func (s *Server) chatStorage() boardgame.ChatStorageManager {
	if cs, ok := s.storage.StorageManager.(boardgame.ChatStorageManager); ok {
		return cs
	}
	return nil
}

func (s *Server) chatSendHandler(c *gin.Context) {
	r := s.newRenderer(c)
	game := s.getGame(c)
	if game == nil {
		r.Error(errors.NewFriendly("No such game"))
		return
	}

	cs := s.chatStorage()
	if cs == nil {
		r.Error(errors.NewFriendly("Chat is not available"))
		return
	}

	user := s.getUser(c)
	if user == nil {
		r.Error(errors.NewFriendly("Not logged in"))
		return
	}

	// Resolve the player index for this user
	userIds := s.storage.UserIDsForGame(game.ID())
	playerIndex := boardgame.ObserverPlayerIndex
	for i, uid := range userIds {
		if uid == user.ID {
			playerIndex = boardgame.PlayerIndex(i)
			break
		}
	}

	if playerIndex == boardgame.ObserverPlayerIndex {
		r.Error(errors.NewFriendly("Observers cannot send chat messages"))
		return
	}

	// Get the chat policy for this player, augmented with DM channels
	state := game.CurrentState()
	policy := game.Manager().Delegate().ChatPolicyForPlayer(state, playerIndex)
	policy = s.augmentPolicyWithDMs(policy, game, user.ID)

	if !policy.Enabled {
		r.Error(errors.NewFriendly("Chat is not available right now"))
		return
	}

	channel := c.PostForm("channel")
	body := c.PostForm("body")

	if channel == "" {
		channel = "all"
	}
	if body == "" {
		r.Error(errors.NewFriendly("Message cannot be empty"))
		return
	}
	if len(body) > 500 {
		r.Error(errors.NewFriendly("Message is too long (max 500 characters)"))
		return
	}

	// Validate channel is in SendChannels
	channelAllowed := false
	for _, ch := range policy.SendChannels {
		if ch == channel {
			channelAllowed = true
			break
		}
	}
	if !channelAllowed {
		r.Error(errors.NewFriendly("You cannot send messages to this channel"))
		return
	}

	// Validate pre-baked constraint
	if policy.PrebakedOnly {
		allowed := false
		for _, msg := range policy.AllowedMessages {
			if msg == body {
				allowed = true
				break
			}
		}
		if !allowed {
			r.Error(errors.NewFriendly("That message is not allowed"))
			return
		}
	}

	msg := &boardgame.ChatMessage{
		GameID:    game.ID(),
		Version:   game.Version(),
		Sender:    playerIndex,
		Channel:   channel,
		Body:      body,
		Timestamp: time.Now(),
	}

	if err := cs.SaveChatMessage(msg); err != nil {
		r.Error(errors.New("Failed to save chat message: " + err.Error()))
		return
	}

	// Notify all connected clients via WebSocket
	s.notifier.chatMessageSent(game.ID(), channel, msg.ID)

	r.Success(gin.H{
		"MessageID": msg.ID,
	})
}

func (s *Server) chatReadHandler(c *gin.Context) {
	r := s.newRenderer(c)
	game := s.getGame(c)
	if game == nil {
		r.Error(errors.NewFriendly("No such game"))
		return
	}

	cs := s.chatStorage()
	if cs == nil {
		r.Error(errors.NewFriendly("Chat is not available"))
		return
	}

	// Determine the viewer
	user := s.getUser(c)
	playerIndex := boardgame.ObserverPlayerIndex
	if user != nil {
		userIds := s.storage.UserIDsForGame(game.ID())
		for i, uid := range userIds {
			if uid == user.ID {
				playerIndex = boardgame.PlayerIndex(i)
				break
			}
		}
	}

	// Get the chat policy, augmented with DM channels
	state := game.CurrentState()
	policy := game.Manager().Delegate().ChatPolicyForPlayer(state, playerIndex)
	userID := ""
	if user != nil {
		userID = user.ID
	}
	policy = s.augmentPolicyWithDMs(policy, game, userID)

	channel := c.Query("channel")
	sinceID := c.Query("since")
	limitStr := c.DefaultQuery("limit", "50")
	limit := 50
	if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
		limit = n
	}
	if limit > 200 {
		limit = 200 // cap to prevent memory exhaustion
	}

	// If a specific channel is requested, verify view access
	if channel != "" {
		allowed := false
		for _, ch := range policy.ViewChannels {
			if ch == channel {
				allowed = true
				break
			}
		}
		if !allowed {
			r.Error(errors.NewFriendly("You cannot view this channel"))
			return
		}
	}

	// Fetch messages
	messages, err := cs.ChatMessages(game.ID(), channel, sinceID, limit)
	if err != nil {
		r.Error(errors.New("Failed to fetch chat messages: " + err.Error()))
		return
	}

	// Filter messages by ViewChannels if no specific channel was requested
	if channel == "" && policy.ViewChannels != nil {
		viewSet := make(map[string]bool)
		for _, ch := range policy.ViewChannels {
			viewSet[ch] = true
		}
		var filtered []*boardgame.ChatMessage
		for _, msg := range messages {
			if viewSet[msg.Channel] {
				filtered = append(filtered, msg)
			}
		}
		messages = filtered
	}

	// Build response with player display names for convenience
	type chatMessageResponse struct {
		ID        string `json:"id"`
		Channel   string `json:"channel"`
		Sender    int    `json:"sender"`
		Body      string `json:"body"`
		Timestamp int64  `json:"timestamp"`
	}

	var response []chatMessageResponse
	for _, msg := range messages {
		response = append(response, chatMessageResponse{
			ID:        msg.ID,
			Channel:   msg.Channel,
			Sender:    int(msg.Sender),
			Body:      msg.Body,
			Timestamp: msg.Timestamp.UnixMilli(),
		})
	}

	// Build user ID → player index map for DM channel name resolution
	userIDMap := make(map[string]int)
	if uids := s.storage.UserIDsForGame(game.ID()); uids != nil {
		for i, uid := range uids {
			if uid != "" {
				userIDMap[uid] = i
			}
		}
	}

	r.Success(gin.H{
		"Messages":     response,
		"ViewChannels": policy.ViewChannels,
		"UserIDMap":    userIDMap,
		"ChatConfig": gin.H{
			"Enabled":         policy.Enabled,
			"PrebakedOnly":    policy.PrebakedOnly,
			"AllowedMessages": policy.AllowedMessages,
		},
	})
}

func (s *Server) Start() {

	config, err := config.Get("", false)

	for _, o := range s.overriders {
		config.AddOverride(o)
	}

	if err != nil {
		s.logger.Errorln("Configuration error: " + err.Error())
		return
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
		s.logger.Fatalln("Couldn't connect to storage manager: ", err)
		return
	}

	s.notifier = newVersionNotifier(s)

	// Check if the storage backend supports chat
	if _, ok := s.storage.StorageManager.(boardgame.ChatStorageManager); ok {
		s.logger.Infoln("Chat storage available")
	} else {
		s.logger.Infoln("Chat storage not available — chat will be disabled")
	}

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

		// Companion-mode join endpoints. /api/join is unauthenticated (a phone
		// types in a room code before deciding whether to sign in). Both are
		// rate-limited per client IP — see spec §6.2 for the threat model.
		joinGroup := mainGroup.Group("join")
		joinGroup.Use(rateLimitMiddleware(s.joinRateLimiter))
		joinGroup.POST("", s.joinHandler)
		joinGroup.POST("seat", s.joinSeatHandler)
		joinGroup.GET("seat-options", s.joinSeatOptionsHandler)

		protectedMainGroup := mainGroup.Group("")
		protectedMainGroup.Use(s.requireLoggedIn)
		protectedMainGroup.POST("new/game", s.newGameHandler)

		gameAPIGroup := mainGroup.Group("game/:name/:id")
		gameAPIGroup.Use(s.gameAPISetup)
		{
			gameAPIGroup.GET("socket", s.socketHandler)
			gameAPIGroup.GET("info", s.gameInfoHandler)
			gameAPIGroup.GET("version/:version", s.gameVersionHandler)
			// Self-hosted QR code generator for the Table+Hand room code
			// (P5+). Unauthenticated by design — the QR encodes nothing
			// secret (just the join URL); rendering it doesn't expose
			// anything a phone scanning the projector wouldn't already see.
			gameAPIGroup.GET("qrcode.png", s.qrcodeHandler)

			//The statusHandler is conceptually here, but becuase we want to
			//optimize it so much we have it congfigured at the top level.

			protectedGameAPIGroup := gameAPIGroup.Group("")
			protectedGameAPIGroup.Use(s.requireLoggedIn)
			protectedGameAPIGroup.POST("move", s.moveHandler)
			protectedGameAPIGroup.POST("join", s.joinGameHandler)
			protectedGameAPIGroup.POST("configure", s.configureGameHandler)
			protectedGameAPIGroup.POST("chat", s.chatSendHandler)
			// Companion-mode host actions (spec §9). All require a logged-in
			// user; further gated by isHost() / seated checks in the handlers.
			protectedGameAPIGroup.POST("hostSkipTurn", s.hostSkipTurnHandler)
			protectedGameAPIGroup.POST("claimHost", s.claimHostHandler)
			protectedGameAPIGroup.POST("switchToSolo", s.switchToSoloHandler)
			protectedGameAPIGroup.POST("setRoomLock", s.setRoomLockHandler)
		}

		// Chat read endpoint — available to any user with game access
		gameAPIGroup.GET("chat", s.chatReadHandler)
	}

	if p := os.Getenv("PORT"); p != "" {
		router.Run(":" + p)
	} else {
		router.Run(":" + s.config.DefaultPort)
	}

}
