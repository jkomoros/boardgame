package tictactoe

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestGame(t *testing.T) {

	game := NewGame(NewManager())

	if game == nil {
		t.Error("Didn't get tictactoe game back")
	}

	if game.Name() != gameName {
		t.Error("Game didn't have right name. wanted", gameName, "got", game.Name())
	}
}

func TestStateFromBlob(t *testing.T) {
	game := NewGame(NewManager())

	if err := <-game.ProposeMove(&MovePlaceToken{
		Slot:              1,
		TargetPlayerIndex: 0,
	}); err != nil {
		t.Fatal("Couldn't make move", err)
	}

	blob, err := json.Marshal(game.CurrentState())

	if err != nil {
		t.Fatal("Couldn't serialize state:", err)
	}

	reconstitutedState, err := game.Manager().Delegate().StateFromBlob(blob, 0)

	if err != nil {
		t.Error("StateFromBlob returned unexpected err", err)
	}

	if !reconstitutedState.(*mainState).Game.Slots.Inflated() {
		t.Error("The stack was not inflated when it came back from StateFromBlob")
	}

	if !reflect.DeepEqual(reconstitutedState, game.CurrentState()) {

		rStateBlob, _ := json.Marshal(reconstitutedState)
		oStateBlob, _ := json.Marshal(game.CurrentState())

		t.Error("Reconstituted state and original state were not the same. Got", string(rStateBlob), "wanted", string(oStateBlob))
	}

}

func TestCheckGameFinished(t *testing.T) {
	tests := []struct {
		input            []string
		expectedFinished bool
		expectedWinner   string
		description      string
	}{
		{
			[]string{X, X, X, Empty, Empty, Empty, Empty, Empty, Empty},
			true,
			X,
			"Row X winner",
		},
		{
			[]string{X, O, O, X, O, O, X, O, X},
			true,
			X,
			"Col X winner",
		},
		{
			[]string{X, O, X, O, X, O, O, X, O},
			true,
			Empty,
			"Draw",
		},
		{
			[]string{Empty, O, X, O, X, O, X, X, O},
			true,
			X,
			"Diagonal up",
		},
		{
			[]string{X, O, X, O, X, O, O, O, X},
			true,
			X,
			"Diagonal down",
		},
		{
			[]string{X, O, X, O, X, Empty, Empty, X, O},
			false,
			Empty,
			"No winner",
		},
	}

	for i, test := range tests {
		finished, winner := checkGameFinished(test.input)
		if finished != test.expectedFinished {
			t.Error("Error at test", i, test.description, "Got", finished, "wanted", test.expectedFinished)
		}
		if winner != test.expectedWinner {
			t.Error("Error at test", i, test.description, "Got", winner, "wanted", test.expectedWinner)
		}
	}
}
