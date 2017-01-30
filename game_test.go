package boardgame

import (
	"reflect"
	"testing"
)

type testGameDelegate struct{}

func (t *testGameDelegate) DistributeComponentToStarterStack(c *Component) {
	//TODO: test that this is done correctly.
}

func (t *testGameDelegate) CheckGameFinished(state StatePayload) (bool, []int) {
	p := state.(*testStatePayload)

	var winners []int

	for i, user := range p.users {
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

func (t *testGameDelegate) ProposeFixUpMove(state StatePayload) Move {
	move := &testMoveAdvanceCurentPlayer{}

	if err := move.Legal(state); err == nil {
		return move
	}

	return nil
}

func TestGameSetUp(t *testing.T) {
	game := testGame()

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

	if err := game.ApplyMove(move); err == nil {
		t.Error("Game allowed a move to be made before SetUp was called")
	}

	game.SetUp()

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

	oldGameName := game.Name

	game.Name = "WRONG NAME"

	if err := game.ApplyMove(move); err == nil {
		t.Error("Game allowed a move with wrong game name to be applied")
	}

	game.Name = oldGameName

	//testMove checks to make sure game.state.currentPlayerIndex is targetplayerindex

	move.TargetPlayerIndex = 1

	if err := game.ApplyMove(move); err == nil {
		t.Error("Game allowed a move to be applied where the wrong playe was current")
	}

	move.TargetPlayerIndex = 0

	if err := game.ApplyMove(move); err != nil {
		t.Error("Game didn't allow a legal move to be made")
	}

	//Verify that the move was made. Note that because our Delegate has a
	//FixUp move, this is also testing that not just the main move, but also
	//the fixup move was made.

	json := game.State.JSON()
	golden := goldenJSON("basic_state_after_move.json", t)

	compareJSONObjects(json, golden, "Basic state after test move", t)

	//Apply a move that should finish the game (any player has score > 5)
	newMove := &testMove{
		AString:           "foo",
		ScoreIncrement:    6,
		TargetPlayerIndex: 1,
		ABool:             true,
	}

	if err := game.ApplyMove(newMove); err != nil {
		t.Error("Game didn't allow a move to be made even though it was legal")
	}

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

	if err := game.ApplyMove(moveAfterFinished); err == nil {
		t.Error("Game allowed a move to be applied after the game was finished")
	}
}
