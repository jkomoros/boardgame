package boardgame

import (
	"testing"
)

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

	newMove := &testMove{
		AString:           "foo",
		ScoreIncrement:    3,
		TargetPlayerIndex: 1,
		ABool:             true,
	}

	//newMove is valid at this point.
	game.Finished = true

	if game.ApplyMove(newMove) {
		t.Error("Game allowed a move to be made even though the game was Finished.")
	}

}
