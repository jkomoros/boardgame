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
