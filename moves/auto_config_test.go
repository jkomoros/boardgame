package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/auto"
	"github.com/workfit/tester/assert"
	"testing"
)

//+autoreader
type moveShuffleStack struct {
	ShuffleStack
}

func (m *moveShuffleStack) SourceStack(state boardgame.State) boardgame.Stack {
	return state.GameState().(*gameState).DrawStack
}

func TestShuffleStackDefaultConfig(t *testing.T) {

	moveInstaller := func(manager *boardgame.GameManager) *boardgame.MoveTypeConfigBundle {

		return boardgame.NewMoveTypeConfigBundle().AddMove(
			auto.MustConfig(new(moveShuffleStack)),
		)

	}

	manager, err := newGameManager(moveInstaller)

	assert.For(t).ThatActual(err).IsNil()

	moveType := manager.MoveTypeByName("Shuffle DrawStack Stack")

	assert.For(t).ThatActual(moveType).IsNotNil()

}

func TestDealCardsDefaultConfig(t *testing.T) {
	moveInstaller := func(manager *boardgame.GameManager) *boardgame.MoveTypeConfigBundle {

		return boardgame.NewMoveTypeConfigBundle().AddMove(
			auto.MustConfig(new(moveDealCards), WithMoveName("Deal Cards")),
		)

	}

	manager, err := newGameManager(moveInstaller)

	assert.For(t).ThatActual(err).IsNil()

	moveType := manager.MoveTypeByName("Deals Components From Game Stack DrawStack To Player Stack Hand To Each Player 1 Times")

	assert.For(t).ThatActual(moveType).IsNotNil()
}
