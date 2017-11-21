package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/workfit/tester/assert"
	"testing"
)

//+autoreader
type moveShuffleStack struct {
	ShuffleStack
}

func (m *moveShuffleStack) SourceStack(state boardgame.MutableState) boardgame.MutableStack {
	return state.GameState().(*gameState).DrawStack
}

func TestShuffleStackDefaultConfig(t *testing.T) {

	moveInstaller := func(manager *boardgame.GameManager) *boardgame.MoveTypeConfigBundle {

		return boardgame.NewMoveTypeConfigBundle().AddMove(
			MustDefaultConfig(manager, new(moveShuffleStack)),
		)

	}

	manager, err := newGameManager(moveInstaller)

	assert.For(t).ThatActual(err).IsNil()

	moveType := manager.FixUpMoveTypeByName("Shuffle DrawStack Stack")

	assert.For(t).ThatActual(moveType).IsNotNil()

}

func TestDealCardsDefaultConfig(t *testing.T) {
	moveInstaller := func(manager *boardgame.GameManager) *boardgame.MoveTypeConfigBundle {

		return boardgame.NewMoveTypeConfigBundle().AddMove(
			MustDefaultConfig(manager, new(moveDealCards)),
		)

	}

	manager, err := newGameManager(moveInstaller)

	assert.For(t).ThatActual(err).IsNil()

	moveType := manager.FixUpMoveTypeByName("Deals Components From Game Stack DrawStack To Player Stack Hand To Each Player 1 Times")

	assert.For(t).ThatActual(moveType).IsNotNil()
}
