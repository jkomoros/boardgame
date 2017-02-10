package tictactoe

import (
	"testing"
)

func TestGame(t *testing.T) {

	game := NewGame()

	if game == nil {
		t.Error("Didn't get tictactoe game back")
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
