package boardgame

import (
	"testing"
)

func TestApplyMove(t *testing.T) {
	game := testGame()

	move := &testMove{
		"foo",
		3,
		true,
	}

	//TODO: actually test this
	game.ApplyMove(move)
}
