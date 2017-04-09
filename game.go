package boardgame

import (
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
	"time"
)

//maxRecurseCount is the number of fixUp moves that can be considered normal--
//anything more than that and we'll panic because the delegate is likely going
//to return fixup moves forever.
const maxRecurseCount = 50

//A Game represents a specific game between a collection of Players. Create a
//new one with NewGame().
type Game struct {
	manager *GameManager

	finished bool

	winners []int

	//The current version of State.
	version int

	numPlayers int

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

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
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

//NewGame returns a new game. You must call SetUp before using it.
func NewGame(manager *GameManager) *Game {

	if manager == nil {
		return nil
	}

	result := &Game{
		manager: manager,
		//TODO: set the size of chan based on something more reasonable.
		//Note: this is also set similarly in manager.ModifiableGame
		proposedMoves: make(chan *proposedMoveItem, 20),
		id:            randomString(gameIDLength),
		modifiable:    true,
	}

	manager.modifiableGameCreated(result)

	return result

}

//Winners is the player indexes who were winners. Typically, this will be
//one player, but it could be multiple in the case of tie, or 0 in the
//case of a draw.
func (g *Game) Winners() []int {
	return g.winners
}

//Finished is whether the came has been completed. If it is over, the
//Winners will be set.
func (g *Game) Finished() bool {
	return g.finished
}

//Manager is a reference to the GameManager that controls this game.
//GameManager's methods will be called at key points in the lifecycle of
//this game.
func (g *Game) Manager() *GameManager {
	return g.manager
}

//NumPlayers returns the number of players for this game, based on how many
//PlayerStates are in CurrentState.
func (g *Game) NumPlayers() int {
	return g.numPlayers
}

//JSONForPlayer returns an object appropriate for being json'd via
//json.Marshal. The object is the equivalent to what MarshalJSON would output,
//only as an object, and with CurrentState equal to a sanitized state for the
//given player.
func (g *Game) JSONForPlayer(player PlayerIndex) interface{} {
	return map[string]interface{}{
		"Name":         g.Name(),
		"Finished":     g.Finished(),
		"Winners":      g.Winners(),
		"CurrentState": g.CurrentState().SanitizedForPlayer(player),
		"Id":           g.Id(),
		"NumPlayers":   g.NumPlayers(),
		"Version":      g.Version(),
	}
}

func (g *Game) MarshalJSON() ([]byte, error) {
	//We define our own MarshalJSON because if we didn't there'd be an infinite loop because of the redirects back up.
	return json.Marshal(g.JSONForPlayer(AdminPlayerIndex))
}

//StorageRecord returns a GameStorageRecord representing the aspects of this
//game that should be serialized to storage.
func (g *Game) StorageRecord() *GameStorageRecord {
	return &GameStorageRecord{
		Name:       g.Manager().Delegate().Name(),
		Version:    g.Version(),
		Winners:    g.Winners(),
		Finished:   g.Finished(),
		Id:         g.Id(),
		NumPlayers: g.NumPlayers(),
	}
}

func (g *Game) Name() string {
	return g.manager.Delegate().Name()
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

	record, err := g.manager.Storage().State(g.Id(), version)

	if err != nil {
		panic("State retrieval failed" + err.Error() + strconv.Itoa(version))
	}

	result, err := g.manager.stateFromRecord(record)

	if err != nil {
		panic("StateFromBlob failed: " + err.Error())
	}

	result.game = g

	return result

}

//SetUp should be called a single time after all of the member variables are
//set correctly, including Chest. SetUp must be called before ProposeMove can
//be called. Even if an error is returned, the game should be in a consistent
//state. If numPlayers is 0, we will use delegate.DefaultNumPlayers().
func (g *Game) SetUp(numPlayers int) error {

	if g.initalized {
		return errors.New("Game already initalized")
	}

	//TODO: we don't need this anymore because managers can't be created without chests.
	if g.manager.Chest() == nil {
		return errors.New("No component chest set on manager")
	}

	if numPlayers == 0 {
		numPlayers = g.manager.Delegate().DefaultNumPlayers()
	}

	if numPlayers < 1 {
		return errors.New("The number of players, " + strconv.Itoa(numPlayers) + " is not legal. There must be one or more players.")
	}

	if !g.manager.Delegate().LegalNumPlayers(numPlayers) {
		return errors.New("The number of players, " + strconv.Itoa(numPlayers) + " was not legal.")
	}

	g.numPlayers = numPlayers

	stateCopy := &state{
		game: g,
	}

	gameState, err := g.manager.emptyGameState(stateCopy)

	if err != nil {
		return err
	}

	stateCopy.gameState = gameState

	playerStates := make([]MutablePlayerState, numPlayers)

	for i := 0; i < numPlayers; i++ {
		playerState, err := g.manager.emptyPlayerState(stateCopy, i)

		if err != nil {
			return err
		}

		playerStates[i] = playerState
	}

	stateCopy.playerStates = playerStates

	dynamic, err := g.manager.emptyDynamicComponentValues(stateCopy)

	if err != nil {
		return errors.New("Couldn't create empty dynamic component values: " + err.Error())
	}

	stateCopy.dynamicComponentValues = dynamic

	g.manager.delegate.BeginSetUp(stateCopy)

	//Distribute all components to their starter locations

	for _, name := range g.Chest().DeckNames() {
		deck := g.Chest().Deck(name)
		for i, component := range deck.Components() {
			stack, err := g.manager.Delegate().DistributeComponentToStarterStack(stateCopy, component)
			if err != nil {
				return errors.New("Distributing components failed for deck " + name + ":" + strconv.Itoa(i) + ":" + err.Error())
			}
			if stack == nil {
				return errors.New("Distributing components failed for deck " + name + ":" + strconv.Itoa(i) + ": the delegate returned no stack.")
			}
			if stack.SlotsRemaining() < 1 {
				return errors.New("Distributing components failed for deck " + name + ":" + strconv.Itoa(i) + ": the stack the delegate returned had no more slots.")
			}
			stack.insertComponentAt(stack.effectiveIndex(NextSlotIndex), component)
		}
	}

	g.manager.delegate.FinishSetUp(stateCopy)

	if g.Modifiable() {

		//Save the initial state to DB.
		if err := g.manager.Storage().SaveGameAndCurrentState(g.StorageRecord(), stateCopy.StorageRecord()); err != nil {
			return errors.New("Storage failed: " + err.Error())
		}
	}

	g.initalized = true

	//See if any fixup moves apply

	//TODO: test that fixup moves are applied at the beginning.

	move := g.manager.Delegate().ProposeFixUpMove(stateCopy)

	if move != nil {
		//We apply the move immediately. This ensures that when
		//DelayedError resolves, all of the fix up moves have been
		//applied.
		if err := g.applyMove(move, true, 0); err != nil {
			//TODO: if we bail here, we haven't left Game in a consistent
			//state because we haven't rolled back what we did.
			return errors.New("Applying the first fix up move failed: " + err.Error())
		}
	}

	if g.Modifiable() {

		//Can't start this until now, otherwise we could have a race.
		go g.mainLoop()
	}

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

	result := g.manager.PlayerMoves()
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

	result := g.manager.FixUpMoves()
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

	result := g.manager.PlayerMoveByName(name)

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

	result := g.manager.FixUpMoveByName(name)

	if result == nil {
		return result
	}

	result.DefaultsForState(g.CurrentState())

	return result
}

//Chest is the ComponentChest in use for this game.
func (g *Game) Chest() *ComponentChest {
	return g.manager.Chest()
}

//Refresh goes and sets this game object to reflect the current state of the
//underlying game in Storage. Basically, when you call manager.Game() you get
//a snapshot of the game in storage at that moment. If you believe that the
//underlying game in storage has been modified, calling Refresh() will re-load
//the snapshot, effectively. Most useful after calling ProposeMove() on a non-
//modifiable game.
func (g *Game) Refresh() {

	freshGame := g.manager.Game(g.Id())

	g.cachedCurrentState = nil
	g.version = freshGame.Version()
	g.finished = freshGame.Finished()
	g.winners = freshGame.Winners()

}

//ProposedMove is the way to propose a move to the game. DelayedError will
//return an error in the future if the move was unable to be applied, or nil
//if the move was applied successfully. DelayedError will only resolve once
//any applicable FixUp moves have been applied already. Note: DelayedError
//won't return anything until after SetUp has been called. This is legal to
//call on a non-modifiable game--the change will be dispatched to a modifiable
//version of the game with this ID. However, note that if you call it on a
//non-modifiable game, even once DelayedError has resolved, the original game
//will still represent its old state. If you wantt to see its current state,
//calling game.Refresh() after DelayedError has resolved should contain the
//move changes you proposed, if they were accepted (and of course potentially
//more moves if other moves were applied in the meantime).
func (g *Game) ProposeMove(move Move) DelayedError {

	if !g.Modifiable() {
		return g.manager.proposeMoveOnGame(g.Id(), move)
	}

	errChan := make(DelayedError, 1)

	workItem := &proposedMoveItem{
		move: move,
		ch:   errChan,
	}

	if !g.initalized {
		//The channel isn't even ready to send one.
		errChan <- errors.New("Proposed a move before the game had been successfully set-up.")
		return errChan
	}

	g.proposedMoves <- workItem

	return errChan

}

//Game applies the move to the state if it is currently legal. May only be
//called by mainLoop. Propose moves with game.ProposeMove instead.
func (g *Game) applyMove(move Move, isFixUp bool, recurseCount int) error {

	if !g.initalized {
		return errors.New("The game has not been initalized.")
	}

	if g.finished {
		return errors.New("Game was already finished")
	}

	if isFixUp {
		if g.manager.FixUpMoveByName(move.Name()) == nil {
			return errors.New("That move is not configured as a Fix Up move for this game.")
		}
	} else {

		//Verify that the Move is actually configured to be part of this game.
		if g.manager.PlayerMoveByName(move.Name()) == nil {
			return errors.New("That move is not configured as a Player move for this game.")
		}
	}

	currentState := g.CurrentState().(*state)

	if err := move.Legal(currentState); err != nil {
		//It's not legal, reject.
		return errors.New("The move was not legal: " + err.Error())
	}

	newState := currentState.copy(false)

	if err := move.Apply(newState); err != nil {
		return errors.New("The move's apply function returned an error:" + err.Error())
	}

	//Check to see if that move made the game finished.

	finished, winners := g.manager.Delegate().CheckGameFinished(newState)

	if finished {
		g.finished = true
		g.winners = winners
		//TODO: persist to database here.
	}

	g.version = g.version + 1
	//Expire the currentState cache; it's no longer valid.
	g.cachedCurrentState = nil

	//TODO: test that if we fail to save state to storage everything's fine.
	if err := g.manager.Storage().SaveGameAndCurrentState(g.StorageRecord(), newState.StorageRecord()); err != nil {
		//TODO: we need to undo the temporary changes we made directly to ourselves (vesrion, finished, winners)
		return errors.New("Storage returned an error:" + err.Error())
	}

	//Ok, the state stuck and is now canonical--trigger the actions it was
	//supposed to do.
	newState.committed()

	if recurseCount > maxRecurseCount {
		panic("We recursed deeply in fixup, which implies that ProposeFixUp has a move that is always legal. Quitting.")
	}

	if g.finished {
		return nil
	}

	move = g.manager.Delegate().ProposeFixUpMove(newState)

	if move != nil {
		//We apply the move immediately. This ensures that when
		//DelayedError resolves, all of the fix up moves have been
		//applied.
		if err := g.applyMove(move, true, recurseCount+1); err != nil {
			//TODO: if we bail here, we haven't left Game in a consistent
			//state because we haven't rolled back what we did.
			return errors.New("Applying the fix up move failed: " + strconv.Itoa(recurseCount) + ": " + err.Error())
		}
	}

	return nil

}
