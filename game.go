package boardgame

import (
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
	"strings"
)

//maxRecurseCount is the number of fixUp moves that can be considered normal--
//anything more than that and we'll panic because the delegate is likely going
//to return fixup moves forever.
const maxRecurseCount = 50

//A Game represents a specific game between a collection of Players. Create a
//new one with NewGame().
type Game struct {
	//Name is a string that defines the type of game this is. The name should
	//be unique but human readable. Good examples are "Tic Tac Toe",
	//"Blackjack".
	Name string

	//Delegate is an (optional) way to override behavior at key game states.
	Delegate GameDelegate

	//Finished is whether the came has been completed. If it is over, the
	//Winners will be set.
	Finished bool
	//Winners is the player indexes who were winners. Typically, this will be
	//one player, but it could be multiple in the case of tie, or 0 in the
	//case of a draw.
	Winners []int

	//The current version of State.
	version int
	//The schema of the game.
	schema int
	//TODO: allow setting this.

	//The storage to use. When we move Delegat to be a GameManager, storage
	//should live there instead.
	storage StorageManager

	//Moves is the set of all move types that are ever legal to apply in this
	//game. When a move will be proposed it should copy one of these moves.
	//Player moves are moves that can be applied by users. FixUp moves are
	//only ever returned by Delegate.ProposeFixUpMove().
	playerMoves       []Move
	fixUpMoves        []Move
	playerMovesByName map[string]Move
	fixUpMovesByName  map[string]Move

	//Memozied answer to CurrentState. Invalidated whenever ApplyMove is
	//called.
	cachedCurrentState State

	//Modifiable controls whether moves can be made on this game.
	modifiable bool

	//A unique ID provided to this game when it is created.
	id string

	//Proposed moves is where moves that have been proposed but have not yet been applied go.
	proposedMoves chan *proposedMoveItem

	//Initalized is set to True after SetUp is called.
	initalized bool
	chest      *ComponentChest

	//TODO: HistoricalState(index int) and HistoryLen() int

	//TODO: an array of Player objects.
}

const gameIDLength = 16

type DelayedError chan error

type proposedMoveItem struct {
	move Move
	//Ch is the channel we should either return an error on and then close, or
	//send nil and close.
	ch DelayedError
}

const randomStringChars = "ABCDEF0123456789"

//randomString returns a random string of the given length.
func randomString(length int) string {
	var result = ""

	for len(result) < length {
		result += string(randomStringChars[rand.Intn(len(randomStringChars))])
	}

	return result
}

//NewGame returns a new game. You must set a Chest and call AddMove with all
//moves, before calling SetUp. Then the game can be used.
func NewGame(name string, delegate GameDelegate, storage StorageManager) *Game {

	if storage == nil {
		return nil
	}

	if delegate == nil {
		return nil
	}

	result := &Game{
		Name:     name,
		Delegate: delegate,
		//TODO: set the size of chan based on something more reasonable.
		proposedMoves: make(chan *proposedMoveItem, 20),
		id:            randomString(gameIDLength),
		modifiable:    true,
		storage:       storage,
	}

	return result

}

//GameDelegate is called at various points in the game lifecycle. It is one of
//the primary ways that a specific game controls behavior over and beyond
//Moves and their Legal states.
type GameDelegate interface {

	//DistributeComponentToStarterStack is called during set up to establish
	//the Deck/Stack invariant that every component in the chest is placed in
	//precisely one Stack. Game will call this on each component in the Chest
	//in order. This is where the logic goes to make sure each Component goes
	//into its correct starter stack. As long as you put each component into a
	//Stack, the invariant will be met at the end of SetUp. If any errors are
	//returned SetUp fails. Unlike after the game has been SetUp, you can
	//modify payload directly.
	DistributeComponentToStarterStack(state State, c *Component) error

	//CheckGameFinished should return true if the game is finished, and who
	//the winners are. Called after every move is applied.
	CheckGameFinished(state State) (finished bool, winners []int)

	//ProposeFixUpMove is called after a move has been applied. It may return
	//a FixUp move, which will be applied before any other moves are applied.
	//If it returns nil, we may take the next move off of the queue. FixUp
	//moves are useful for things like shuffling a discard deck back into a
	//draw deck, or other moves that are necessary to get the GameState back
	//into reasonable shape.
	ProposeFixUpMove(state State) Move

	//DefaultNumPlayers returns the number of users that this game defaults to.
	//For example, for tictactoe, it will be 2. If 0 is provided to
	//game.SetUp(), we wil use this value instead.
	DefaultNumPlayers() int

	//StartingState should return a zero'd state object for this game type.
	//All future states for this particular game will be created by Copy()ing
	//this state. If you return nil, game.SetUp() will fail.
	StartingState(numPlayers int) State

	//StateFromBlob should deserialize a JSON string of this game's State. We
	//need it to be in a game-specific bit of logic because we don't know the
	//real type of the state stuct for this game. Be sure to inflate any
	//Stacks in the state, and set playerIndex for each UserState in order.
	//It's strongly recommended that you test a round-trip of state through
	//this method.
	StateFromBlob(blob []byte, schema int) (State, error)

	//SetGame is called during game.SetUp and passes a reference to the Game
	//that the delegate is part of.
	SetGame(game *Game)
}

//DefaultGameDelegate is a struct that implements stubs for all of
//GameDelegate's methods. This makes it easy to override just one or two
//methods by creating your own struct that anonymously embeds this one. You
//almost certainly want to override StartingState.
type DefaultGameDelegate struct {
	Game *Game
}

func (d *DefaultGameDelegate) DistributeComponentToStarterStack(state State, c *Component) error {
	//The stub returns an error, because if this is called that means there
	//was a component in the deck. And if we didn't store it in a stack, then
	//we are in violation of the invariant.
	return errors.New("DistributeComponentToStarterStack was called, but the component was not stored in a stack")
}

func (d *DefaultGameDelegate) CheckGameFinished(state State) (finished bool, winners []int) {
	return false, nil
}

func (d *DefaultGameDelegate) StateFromBlob(blob []byte, schema int) (State, error) {
	return nil, errors.New("Default delegate does not know how to deserialize state objects")
}

func (d *DefaultGameDelegate) StartingState(numPlayers int) State {
	return nil
}

func (d *DefaultGameDelegate) DefaultNumPlayers() int {
	return 2
}

//The Default ProposeFixUpMove runs through all moves in FixUpMoves, in order,
//and returns the first one that is legal at the current state. In many cases,
//this behavior should be suficient and need not be overwritten. Be extra sure
//that your FixUpMoves have a conservative Legal function, otherwise you could
//get a panic from applying too many FixUp moves.
func (d *DefaultGameDelegate) ProposeFixUpMove(state State) Move {
	for _, move := range d.Game.FixUpMoves() {
		if err := move.Legal(state); err == nil {
			//Found it!
			return move
		}
	}
	//No moves apply now.
	return nil
}

func (d *DefaultGameDelegate) SetGame(game *Game) {
	d.Game = game
}

func (g *Game) MarshalJSON() ([]byte, error) {
	//We define our own MarshalJSON because if we didn't there'd be an infinite loop because of the redirects back up.
	result := map[string]interface{}{
		"Name":         g.Name,
		"Finished":     g.Finished,
		"Winners":      g.Winners,
		"CurrentState": g.CurrentState(),
		"Id":           g.Id(),
		"Version":      g.Version(),
	}

	return json.Marshal(result)
}

func (g *Game) Id() string {
	return g.id
}

//Version returns the version number of the highest State that is stored for
//this game. This number will increase by one every time a move (either Player
//or FixUp) is applied.
func (g *Game) Version() int {
	return g.version
}

//CurrentVersion returns the state object for the current state. Equivalent,
//semantically, to game.State(game.Version())
func (g *Game) CurrentState() State {
	if g.cachedCurrentState == nil {
		g.cachedCurrentState = g.State(g.Version())
	}
	return g.cachedCurrentState
}

//Returns the game's atate at the current version.
func (g *Game) State(version int) State {

	if version < 0 || version > g.Version() {
		return nil
	}

	return g.storage.State(g, version)

}

//SetUp should be called a single time after all of the member variables are
//set correctly, including Chest. SetUp must be called before ProposeMove can
//be called. Even if an error is returned, the game should be in a consistent
//state. If numPlayers is 0, we will use delegate.DefaultNumPlayers().
func (g *Game) SetUp(numPlayers int) error {

	if g.initalized {
		return errors.New("Game already initalized")
	}

	if g.chest == nil {
		return errors.New("No component chest set")
	}

	g.Delegate.SetGame(g)

	if numPlayers == 0 {
		numPlayers = g.Delegate.DefaultNumPlayers()
	}

	//We'll work on a copy of Payload, so if it fails at some point we can just drop it
	stateCopy := g.Delegate.StartingState(numPlayers)

	if stateCopy == nil {
		return errors.New("Delegate didn't return a starter state.")
	}

	//Distribute all components to their starter locations

	for _, name := range g.Chest().DeckNames() {
		deck := g.Chest().Deck(name)
		for i, component := range deck.Components() {
			if err := g.Delegate.DistributeComponentToStarterStack(stateCopy, component); err != nil {
				return errors.New("Distributing components failed for deck " + name + ":" + strconv.Itoa(i) + ":" + err.Error())
			}
		}
	}

	g.playerMovesByName = make(map[string]Move)
	for _, move := range g.playerMoves {
		g.playerMovesByName[strings.ToLower(move.Name())] = move
	}

	g.fixUpMovesByName = make(map[string]Move)
	for _, move := range g.fixUpMoves {
		g.fixUpMovesByName[strings.ToLower(move.Name())] = move
	}

	//TODO: do other set-up work, including FinishSetUp

	if g.Modifiable() {

		//Save the initial state to DB.
		g.storage.SaveState(g, 0, g.schema, stateCopy)

		go g.mainLoop()
	}

	g.initalized = true

	return nil
}

//MainLoop should be run in a goroutine. It is what takes moves off of
//proposedMoves and applies them. It is the only method that may call
//applyMove.
func (g *Game) mainLoop() {

	for item := range g.proposedMoves {
		if item == nil {
			return
		}
		item.ch <- g.applyMove(item.move, false, 0)
		close(item.ch)
	}

}

//Modifiable returns true if this instantiation of the game can be modified.
//If false, this instantiation is read-only: attributes can be read, but not
//written. In practice this means moves cannot be made.
func (g *Game) Modifiable() bool {
	return g.modifiable
}

//AddPlayerMove adds the specified move to the game as a move that Players can
//make. It may only be called during initalization.
func (g *Game) AddPlayerMove(move Move) {

	if g.initalized {
		return
	}
	g.playerMoves = append(g.playerMoves, move)
}

//AddFixUpMove adds a move that can only be legally made by GameDelegate as a
//FixUp move. It can only be called during initialization.
func (g *Game) AddFixUpMove(move Move) {
	if g.initalized {
		return
	}
	g.fixUpMoves = append(g.fixUpMoves, move)
}

//PlayerMoves returns all moves that are valid in this game to be made my
//players--all of the Moves that have been added via AddPlayerMove  during
//initalization. Returns nil until game.SetUp() has been called. Will return
//moves that are all copies, with them already set to the proper
//DefaultsForState.
func (g *Game) PlayerMoves() []Move {
	if !g.initalized {
		return nil
	}

	result := make([]Move, len(g.playerMoves))

	currentState := g.CurrentState()

	for i, move := range g.playerMoves {
		result[i] = move.Copy()
		result[i].DefaultsForState(currentState)
	}

	return result
}

//FixUpMoves returns all moves that are valid in this game to be made as fixup
//moves--all of the Moves that have been added via AddPlayerMove  during
//initalization. Returns nil until game.SetUp() has been called. Will return
//moves that are all copies, with them already set to the proper
//DefaultsForState.
func (g *Game) FixUpMoves() []Move {

	//TODO: test all of these fixup moves

	if !g.initalized {
		return nil
	}

	result := make([]Move, len(g.fixUpMoves))

	currentState := g.CurrentState()

	for i, move := range g.fixUpMoves {
		result[i] = move.Copy()
		result[i].DefaultsForState(currentState)
	}

	return result
}

//PlayerMoveByName returns the Move of that name from game.PlayerMoves(), if
//it exists. Names are considered without regard to case.  Will return a copy
//with defaults already set for current game state by move.DefaultsForState.
func (g *Game) PlayerMoveByName(name string) Move {
	if !g.initalized {
		return nil
	}
	name = strings.ToLower(name)
	move := g.playerMovesByName[name]

	if move == nil {
		return nil
	}

	result := move.Copy()
	result.DefaultsForState(g.CurrentState())
	return result
}

//FixUpMoveByName returns the Move of that name from game.FixUpMoves(), if
//it exists. Names are considered without regard to case.  Will return a copy
//with defaults already set for current game state by move.DefaultsForState.
func (g *Game) FixUpMoveByName(name string) Move {
	if !g.initalized {
		return nil
	}
	name = strings.ToLower(name)
	move := g.fixUpMovesByName[name]

	if move == nil {
		return nil
	}

	result := move.Copy()
	result.DefaultsForState(g.CurrentState())
	return result
}

//Chest is the ComponentChest in use for this game.
func (g *Game) Chest() *ComponentChest {
	return g.chest
}

//SetChest is the way to associate the given Chest with this game.
func (g *Game) SetChest(chest *ComponentChest) {
	//We are only allowed to change the chest before the game is SetUp.
	if g.initalized {
		return
	}
	if chest != nil {
		chest.game = g
		//If Finish was not already called in Chest it must be now--we can't
		//have it changing anymore. This will be a no-op if Finish() was
		//already called.

		//TODO: test that a chest that has not yet had finish called will when
		//added to a game.
		chest.Finish()
	}
	g.chest = chest
}

//ProposedMove is the way to propose a move to the game. DelayedError will
//return an error in the future if the move was unable to be applied, or nil
//if the move was applied successfully. DelayedError will only resolve once
//any applicable FixUp moves have been applied already. Note: DelayedError
//won't return anything until after SetUp has been called.
func (g *Game) ProposeMove(move Move) DelayedError {

	errChan := make(DelayedError, 1)

	if g.Modifiable() {

		workItem := &proposedMoveItem{
			move: move,
			ch:   errChan,
		}

		g.proposedMoves <- workItem
	} else {
		errChan <- errors.New("Game is not modifiable")
	}

	return errChan

}

//Game applies the move to the state if it is currently legal. May only be
//called by mainLoop. Propose moves with game.ProposeMove instead.
func (g *Game) applyMove(move Move, isFixUp bool, recurseCount int) error {

	if !g.initalized {
		return errors.New("The game has not been initalized.")
	}

	if g.Finished {
		return errors.New("Game was already finished")
	}

	if isFixUp {
		if g.FixUpMoveByName(move.Name()) == nil {
			return errors.New("That move is not configured as a Fix Up move for this game.")
		}
	} else {

		//Verify that the Move is actually configured to be part of this game.
		if g.PlayerMoveByName(move.Name()) == nil {
			return errors.New("That move is not configured as a Player move for this game.")
		}
	}

	currentState := g.CurrentState()

	if err := move.Legal(currentState); err != nil {
		//It's not legal, reject.
		return errors.New("The move was not legal: " + err.Error())
	}

	newState := currentState.Copy()

	if err := move.Apply(newState); err != nil {
		return errors.New("The move's apply function returned an error:" + err.Error())
	}

	//TODO: test that if we fail to save state to storage everything's fine.
	if err := g.storage.SaveState(g, g.version+1, g.schema, newState); err != nil {
		return errors.New("Storage returned an error:" + err.Error())
	}

	//We succeeded in saving the state, whcih means that our version can be
	//incremented.
	g.version = g.version + 1
	//Expire the currentState cache; it's no longer valid.
	g.cachedCurrentState = nil

	//Check to see if that move made the game finished.
	if g.Delegate != nil {
		finished, winners := g.Delegate.CheckGameFinished(newState)

		if finished {
			g.Finished = true
			g.Winners = winners
			//TODO: persist to database here.
		}
	}

	if g.Delegate != nil {

		if recurseCount > maxRecurseCount {
			panic("We recursed deeply in fixup, which implies that ProposeFixUp has a move that is always legal. Quitting.")
		}

		move := g.Delegate.ProposeFixUpMove(newState)

		if move != nil {
			//We apply the move immediately. This ensures that when
			//DelayedError resolves, all of the fix up moves have been
			//applied.
			g.applyMove(move, true, recurseCount+1)
		}
	}

	return nil

}
