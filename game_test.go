package boardgame

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
	"time"
)

type testInfiniteLoopGameDelegate struct {
	testGameDelegate
}

func (t *testInfiniteLoopGameDelegate) ProposeFixUpMove(state State) Move {
	return &testAlwaysLegalMove{}
}

func TestProposeMoveNonModifiableGame(t *testing.T) {
	game := testGame()

	game.SetUp(0)

	manager := game.Manager()

	id := game.Id()

	//At this point, the game has stored state in storage.

	refriedGame := manager.Game(id)

	if refriedGame == nil {
		t.Fatal("Couldn't get a game out refried")
	}

	move := &testMove{
		AString:           "foo",
		ScoreIncrement:    3,
		TargetPlayerIndex: 0,
		ABool:             true,
	}

	if err := <-refriedGame.ProposeMove(move); err != nil {
		t.Error("Propose move on refried game failed:", err)
	}

	//Update it from server
	refriedGame.Refresh()

	if refriedGame.Version() != 2 {
		t.Error("The proposed move didn't actually modify the underlying game in storage: ", refriedGame.Version())
	}

}

func TestGameSetUp(t *testing.T) {
	game := testGame()

	id := game.Id()

	if len(id) != gameIDLength {
		t.Error("Game didn't have an ID of correct length. Wanted", gameIDLength, "got", id)
	}

	if game.PlayerMoves() != nil {
		t.Error("Got moves back before SetUp was called")
	}

	if game.PlayerMoveByName("Test") != nil {
		t.Error("Move by name returned a move before SetUp was called")
	}

	move := &testMove{
		AString:           "foo",
		ScoreIncrement:    3,
		TargetPlayerIndex: 0,
		ABool:             true,
	}

	originalTestMove := move

	delayedError := game.ProposeMove(move)

	select {
	case err := <-delayedError:
		t.Error("We got something from a proposed move but the game hadn't even started", err)
	case <-time.After(time.Millisecond * 5):
		//Pass.
	}

	//TODO: we no longer test that SetUp calls the Component distribution logic.

	if err := game.SetUp(0); err != nil {
		t.Error("Calling SetUp on a previously errored game did not succeed", err)
	}

	if wrapper, _ := game.Manager().Storage().State(game, 0); wrapper == nil {
		t.Error("State 0 was not saved in storage when game set up")
	}

	if game.PlayerMoveByName("Test") == nil {
		t.Error("MoveByName didn't return a valid move when provided the proper name after calling setup")
	}

	if game.PlayerMoveByName("test") == nil {
		t.Error("MoveByName didn't return a valid move when provided with a lowercase name after calling SetUp.")
	}

	if originalTestMove == game.PlayerMoveByName("Test") {
		t.Error("MoveByName returned a non-copy")
	}

	p := game.CurrentState().(*testState)

	deck := game.Chest().Deck("test").Components()

	if p.Game.DrawDeck.Len() != len(deck) {
		t.Error("All of the components were not distributed in SetUp")
	}

}

func TestApplyMove(t *testing.T) {
	game := testGame()

	game.SetUp(0)

	move := &testMove{
		AString:           "foo",
		ScoreIncrement:    3,
		TargetPlayerIndex: 0,
		ABool:             true,
	}

	manager := game.Manager()

	oldMoves := manager.playerMoves
	oldMovesByName := manager.playerMovesByName

	manager.playerMoves = nil
	manager.playerMovesByName = make(map[string]Move)

	if err := <-game.ProposeMove(move); err == nil {
		t.Error("Game allowed a move that wasn't configured as part of game to be applied")
	}

	manager.playerMoves = oldMoves
	manager.playerMovesByName = oldMovesByName

	//testMove checks to make sure game.state.currentPlayerIndex is targetplayerindex

	move.TargetPlayerIndex = 1

	if err := <-game.ProposeMove(move); err == nil {
		t.Error("Game allowed a move to be applied where the wrong playe was current")
	}

	move.TargetPlayerIndex = 0

	if err := <-game.ProposeMove(move); err != nil {
		t.Error("Game didn't allow a legal move to be made")
	}

	//Verify that the move was made. Note that because our Delegate has a
	//FixUp move, this is also testing that not just the main move, but also
	//the fixup move was made.

	wrapper, err := game.Manager().Storage().State(game, game.Version())

	if err != nil {
		t.Error("Unexpected error", err)
	}

	currentJson, _ := json.Marshal(wrapper)
	golden := goldenJSON("basic_state_after_move.json", t)

	compareJSONObjects(currentJson, golden, "Basic state after test move", t)

	//Apply a move that should finish the game (any player has score > 5)
	newMove := &testMove{
		AString:           "foo",
		ScoreIncrement:    6,
		TargetPlayerIndex: 1,
		ABool:             true,
	}

	if err := <-game.ProposeMove(newMove); err != nil {
		t.Error("Game didn't allow a move to be made even though it was legal")
	}

	if wrapper, _ := game.Manager().Storage().State(game, 1); wrapper == nil {
		t.Error("We didn't get back state for state 1; game must not be persisting states to DB.")
	}

	//By the time err has resolved above, any fixup moves have been applied.

	if !game.Finished() {
		t.Error("Game didn't notice that a user had won")
	}

	if !reflect.DeepEqual(game.Winners(), []int{1}) {
		t.Error("Game thought the wrong players had won")
	}

	moveAfterFinished := &testMove{
		AString:           "foo",
		ScoreIncrement:    3,
		TargetPlayerIndex: 2,
		ABool:             true,
	}

	if err := <-game.ProposeMove(moveAfterFinished); err == nil {
		t.Error("Game allowed a move to be applied after the game was finished")
	}
}

func TestInfiniteProposeFixUp(t *testing.T) {
	//This test makes sure that if our GameDelegate is going to always return
	//moves that are legal, we'll bail at a certain point.

	manager := NewGameManager(&testInfiniteLoopGameDelegate{}, newTestGameChest(), newTestStorageManager())

	manager.AddPlayerMove(&testMove{})
	manager.AddFixUpMove(&testAlwaysLegalMove{})

	manager.SetUp()

	game := NewGame(manager)

	checkForPanic := func() (didPanic bool) {
		defer func() {
			if e := recover(); e != nil {
				didPanic = true
			}
		}()
		game.SetUp(0)
		return
	}

	if !checkForPanic() {
		t.Error("We didn't get an error when we had a badly behaved ProposeFixUpMove.")
	}

}

func TestGameState(t *testing.T) {
	game := testGame()

	//Force a deterministic ID
	game.id = "483BD9BC8D1A2F27"

	game.SetUp(0)

	if game.Name() != testGameName {
		t.Error("Game name was not correct")
	}

	blob, err := json.MarshalIndent(game, "", "  ")

	if err != nil {
		t.Error("Json marshal of game failed:", err)
	}

	goldenBlob, err := ioutil.ReadFile("test/game_blob.json")

	if err != nil {
		t.Error("Couldn't load golden file", err)
	}

	compareJSONObjects(blob, goldenBlob, "Sanity checking game json", t)

	//Getting this now helps verify that we invalidate currentState cache when
	//we apply a move.
	state := game.CurrentState()

	state0 := game.State(0)

	if !reflect.DeepEqual(state, state0) {
		t.Error("CurrentState at version 0 did not return state 0")
	}

	move := game.PlayerMoveByName("Test")

	if move == nil {
		t.Fatal("Couldn't find a move to make")
	}

	if err := <-game.ProposeMove(move); err != nil {
		t.Error("Couldn't make move")
	}

	state = game.State(-1)

	if state != nil {
		t.Error("Returned a state for a non-sensiscal version -1", state)
	}

	state = game.State(game.Version() + 1)

	if state != nil {
		t.Error("Returned a state for a too-high version", state)
	}

	currentState := game.CurrentState()
	state = game.State(game.Version())

	if !reflect.DeepEqual(currentState, state) {
		t.Error("State(game.Version()) and CurrentState() weren't equivalent", currentState, state)

	}

}
