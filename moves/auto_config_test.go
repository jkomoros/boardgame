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

	moveInstaller := func(manager *boardgame.GameManager) []boardgame.MoveTypeConfig {

		return []boardgame.MoveTypeConfig{
			auto.MustConfig(new(moveShuffleStack)),
		}
	}

	manager, err := newGameManager(moveInstaller)

	assert.For(t).ThatActual(err).IsNil()

	//TODO: change this when fixing #637
	moveType := manager.MoveTypeByName("Shuffle a Stack")

	var typedNil *boardgame.MoveType

	assert.For(t).ThatActual(moveType).DoesNotEqual(typedNil)

}

func TestDealCardsDefaultConfig(t *testing.T) {
	moveInstaller := func(manager *boardgame.GameManager) []boardgame.MoveTypeConfig {

		return []boardgame.MoveTypeConfig{
			auto.MustConfig(new(moveDealCards), WithMoveName("Deal Cards")),
		}

	}

	manager, err := newGameManager(moveInstaller)

	assert.For(t).ThatActual(err).IsNil()

	moveType := manager.MoveTypeByName("Deal Cards")

	var typedNil *boardgame.MoveType

	assert.For(t).ThatActual(moveType).DoesNotEqual(typedNil)
}
