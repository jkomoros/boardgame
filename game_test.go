package boardgame

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

type testInfiniteLoopGameDelegate struct {
	testGameDelegate
}

type testGameDelegate struct {
	DefaultGameDelegate
}

func (t *testGameDelegate) DistributeComponentToStarterStack(state State, c *Component) error {
	p := state.(*testState)
	return p.Game.DrawDeck.InsertFront(c)
}

func (t *testGameDelegate) CheckGameFinished(state State) (bool, []int) {
	p := state.(*testState)

	var winners []int

	for i, user := range p.Users {
		if user.Score >= 5 {
			//This user won!
			winners = append(winners, i)

			//Keep going through to see if anyone else won at the same time
		}
	}

	if len(winners) > 0 {
		return true, winners
	}

	return false, nil
}

func (t *testInfiniteLoopGameDelegate) ProposeFixUpMove(state State) Move {
	return &testAlwaysLegalMove{}
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

	chest := game.Chest()

	game.SetChest(nil)

	if err := game.SetUp(); err == nil {
		t.Error("We were able to call game.SetUp without a Chest")
	}

	game.SetChest(chest)

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

	delegate := game.Delegate

	game.Delegate = nil

	if err := game.SetUp(); err == nil {
		t.Error("game.SetUp didn't error when we had components in a deck but only the default delegate, which errors when a component has DistributeComponentToStarterStack is called.")
	}

	if game.Delegate == nil {
		t.Error("Calling game.SetUp with no delegate did not provide us with one.")
	}

	game.Delegate = delegate

	if err := game.SetUp(); err != nil {
		t.Error("Calling SetUp on a previously errored game did not succeed", err)
	}

	if game.Delegate.(*testGameDelegate).Game != game {
		t.Error("After calling SetUp succesfully SetGame was not called.")
	}

	moves := game.PlayerMoves()

	if reflect.DeepEqual(game.playerMoves, moves) {
		t.Error("Got non-copy moves out of game after SetUp was called.")
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

	p := game.StateWrapper.State.(*testState)

	deck := game.Chest().Deck("test").Components()

	if p.Game.DrawDeck.Len() != len(deck) {
		t.Error("All of the components were not distributed in SetUp")
	}

	newChest := NewComponentChest(testGameName)

	game.SetChest(newChest)

	if game.Chest() == newChest {
		t.Error("We were able to change the chest after game.SetUp was called.")
	}

}

func TestApplyMove(t *testing.T) {
	game := testGame()

	game.SetUp()

	move := &testMove{
		AString:           "foo",
		ScoreIncrement:    3,
		TargetPlayerIndex: 0,
		ABool:             true,
	}

	oldMoves := game.playerMoves
	oldMovesByName := game.playerMovesByName

	game.playerMoves = nil
	game.playerMovesByName = make(map[string]Move)

	if err := <-game.ProposeMove(move); err == nil {
		t.Error("Game allowed a move that wasn't configured as part of game to be applied")
	}

	game.playerMoves = oldMoves
	game.playerMovesByName = oldMovesByName

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

	currentJson, _ := json.Marshal(game.StateWrapper)
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

	//By the time err has resolved above, any fixup moves have been applied.

	if !game.Finished {
		t.Error("Game didn't notice that a user had won")
	}

	if !reflect.DeepEqual(game.Winners, []int{1}) {
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

	game := testGame()

	game.Delegate = &testInfiniteLoopGameDelegate{}

	game.AddFixUpMove(&testAlwaysLegalMove{})

	game.SetUp()

	move := game.PlayerMoveByName("Test")

	if move == nil {
		t.Fatal("Couldn't find Test move")
	}
	checkForPanic := func() (didPanic bool) {
		defer func() {
			if e := recover(); e != nil {
				didPanic = true
			}
		}()
		game.applyMove(move, false, maxRecurseCount-5)
		return
	}

	if !checkForPanic() {
		t.Error("We didn't get an error when we had a badly behaved ProposeFixUpMove.")
	}

}
