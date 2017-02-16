package boardgame

import (
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
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

	//Manager is a reference to the GameManager that controls this game.
	//GameManager's methods will be called at key points in the lifecycle of
	//this game.
	Manager GameManager

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
func NewGame(name string, manager GameManager) *Game {

	if manager == nil {
		return nil
	}

	result := &Game{
		Name:    name,
		Manager: manager,
		//TODO: set the size of chan based on something more reasonable.
		proposedMoves: make(chan *proposedMoveItem, 20),
		id:            randomString(gameIDLength),
		modifiable:    true,
	}

	return result

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

	return g.Manager.Storage().State(g, version)

}

//SetUp should be called a single time after all of the member variables are
//set correctly, including Chest. SetUp must be called before ProposeMove can
//be called. Even if an error is returned, the game should be in a consistent
//state. If numPlayers is 0, we will use delegate.DefaultNumPlayers().
func (g *Game) SetUp(numPlayers int) error {

	if g.initalized {
		return errors.New("Game already initalized")
	}

	if g.Manager.Chest() == nil {
		return errors.New("No component chest set on manager")
	}

	if numPlayers == 0 {
		numPlayers = g.Manager.DefaultNumPlayers()
	}

	//We'll work on a copy of Payload, so if it fails at some point we can just drop it
	stateCopy := g.Manager.StartingState(numPlayers)

	if stateCopy == nil {
		return errors.New("Delegate didn't return a starter state.")
	}

	//Distribute all components to their starter locations

	for _, name := range g.Chest().DeckNames() {
		deck := g.Chest().Deck(name)
		for i, component := range deck.Components() {
			if err := g.Manager.DistributeComponentToStarterStack(stateCopy, component); err != nil {
				return errors.New("Distributing components failed for deck " + name + ":" + strconv.Itoa(i) + ":" + err.Error())
			}
		}
	}

	//TODO: do other set-up work, including FinishSetUp

	if g.Modifiable() {

		//Save the initial state to DB.
		g.Manager.Storage().SaveState(g, 0, g.schema, stateCopy)

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

//PlayerMoves is a thin wrapper around GameManager.PlayerMoves, but with all
//of the moves set to the right defaults for the current state of this game.
func (g *Game) PlayerMoves() []Move {

	if !g.initalized {
		return nil
	}

	result := g.Manager.PlayerMoves()
	for _, move := range result {
		move.DefaultsForState(g.CurrentState())
	}
	return result
}

//FixUpMoves is a thin wrapper around GameManager.FixUpMoves, but with all
//of the moves set to the right defaults for the current state of this game.
func (g *Game) FixUpMoves() []Move {

	if !g.initalized {
		return nil
	}

	result := g.Manager.FixUpMoves()
	for _, move := range result {
		move.DefaultsForState(g.CurrentState())
	}
	return result

}

//PlayerMoveByName is a thin wrapper around GameManager.PlayerMoveByName, but
//with it set to the right defaults for the current state of this game.
func (g *Game) PlayerMoveByName(name string) Move {
	if !g.initalized {
		return nil
	}

	result := g.Manager.PlayerMoveByName(name)

	if result == nil {
		return result
	}

	result.DefaultsForState(g.CurrentState())

	return result
}

//FixUpMoveByName is a thin wrapper around GameManager.FixUpMoveByName, but
//with it set to the right defaults for the current state of this game.
func (g *Game) FixUpMoveByName(name string) Move {

	if !g.initalized {
		return nil
	}

	result := g.Manager.FixUpMoveByName(name)

	if result == nil {
		return result
	}

	result.DefaultsForState(g.CurrentState())

	return result
}

//Chest is the ComponentChest in use for this game.
func (g *Game) Chest() *ComponentChest {
	return g.Manager.Chest()
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
	if err := g.Manager.Storage().SaveState(g, g.version+1, g.schema, newState); err != nil {
		return errors.New("Storage returned an error:" + err.Error())
	}

	//We succeeded in saving the state, whcih means that our version can be
	//incremented.
	g.version = g.version + 1
	//Expire the currentState cache; it's no longer valid.
	g.cachedCurrentState = nil

	//Check to see if that move made the game finished.

	finished, winners := g.Manager.CheckGameFinished(newState)

	if finished {
		g.Finished = true
		g.Winners = winners
		//TODO: persist to database here.
	}

	if recurseCount > maxRecurseCount {
		panic("We recursed deeply in fixup, which implies that ProposeFixUp has a move that is always legal. Quitting.")
	}

	move = g.Manager.ProposeFixUpMove(newState)

	if move != nil {
		//We apply the move immediately. This ensures that when
		//DelayedError resolves, all of the fix up moves have been
		//applied.
		g.applyMove(move, true, recurseCount+1)
	}

	return nil

}
