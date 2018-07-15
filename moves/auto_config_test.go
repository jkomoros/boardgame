package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/with"
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

	moveInstaller := func(manager *boardgame.GameManager) []boardgame.MoveConfig {

		auto := NewAutoConfigurer(manager.Delegate())

		return []boardgame.MoveConfig{
			auto.MustConfig(new(moveShuffleStack),
				with.MoveNameSuffix("Test"),
			),
		}
	}

	manager, err := newGameManager(moveInstaller)

	assert.For(t).ThatActual(err).IsNil()

	move := manager.ExampleMoveByName("Shuffle Draw Stack - Test")

	var typedNil boardgame.Move

	assert.For(t).ThatActual(move).DoesNotEqual(typedNil)

}

func TestDealCardsDefaultConfig(t *testing.T) {
	moveInstaller := func(manager *boardgame.GameManager) []boardgame.MoveConfig {

		auto := NewAutoConfigurer(manager.Delegate())

		return []boardgame.MoveConfig{
			auto.MustConfig(new(moveDealCards), with.MoveName("Deal Cards")),
		}

	}

	manager, err := newGameManager(moveInstaller)

	assert.For(t).ThatActual(err).IsNil()

	move := manager.ExampleMoveByName("Deal Cards")

	var typedNil boardgame.Move

	assert.For(t).ThatActual(move).DoesNotEqual(typedNil)
}
