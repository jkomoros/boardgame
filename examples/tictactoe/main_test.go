package tictactoe

import (
	"testing"
)

func TestGame(t *testing.T) {

	game := ticTacToeGame()

	if game == nil {
		t.Error("Didn't get tictactoe game back")
	}
}
