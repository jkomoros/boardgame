package boardgame

import (
	"reflect"
	"testing"
)

type testGameDelegate struct{}

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

func TestApplyMove(t *testing.T) {
	game := testGame()

	move := &testMove{
		AString:           "foo",
		ScoreIncrement:    3,
		TargetPlayerIndex: 0,
		ABool:             true,
	}

	oldGameName := game.Name

	game.Name = "WRONG NAME"

	if game.ApplyMove(move) {
		t.Error("Game allowed a move with wrong game name to be applied")
	}

	game.Name = oldGameName

	//testMove checks to make sure game.state.currentPlayerIndex is targetplayerindex

	move.TargetPlayerIndex = 1

	if game.ApplyMove(move) {
		t.Error("Game allowed a move to be applied where the wrong playe was current")
	}

	move.TargetPlayerIndex = 0

	if !game.ApplyMove(move) {
		t.Error("Game didn't allow a legal move to be made")
	}

	//Verify that the move was made

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

	if !game.ApplyMove(newMove) {
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

	if game.ApplyMove(moveAfterFinished) {
		t.Error("Game allowed a move to be applied after the game was finished")
	}
}
