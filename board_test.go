package boardgame

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestBoard(t *testing.T) {
	game := testGame(t)

	game.SetUp(0, nil, nil)

	gameState := game.CurrentState().ImmutableGameState().(*testGameState)

	board := gameState.MyBoard

	for i, space := range board.Spaces() {
		assert.For(t).ThatActual(space.Board()).Equals(board)
		assert.For(t).ThatActual(space.BoardIndex()).Equals(i)
		assert.For(t).ThatActual(space.Resizable()).IsFalse()
	}

}
