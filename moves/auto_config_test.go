package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/workfit/tester/assert"
	"testing"
)

//boardgame:codegen
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
				WithMoveNameSuffix("Test"),
			),
		}
	}

	manager, err := newGameManager(moveInstaller, false)

	assert.For(t).ThatActual(err).IsNil()

	if err != nil {
		return
	}

	move := manager.ExampleMoveByName("Shuffle Draw Stack - Test")

	var typedNil boardgame.Move

	assert.For(t).ThatActual(move).DoesNotEqual(typedNil)

}

func TestDealCardsDefaultConfig(t *testing.T) {
	moveInstaller := func(manager *boardgame.GameManager) []boardgame.MoveConfig {

		auto := NewAutoConfigurer(manager.Delegate())

		return []boardgame.MoveConfig{
			auto.MustConfig(new(moveDealCards), WithMoveName("Deal Cards")),
		}

	}

	manager, err := newGameManager(moveInstaller, false)

	assert.For(t).ThatActual(err).IsNil()

	if err != nil {
		return
	}

	move := manager.ExampleMoveByName("Deal Cards")

	var typedNil boardgame.Move

	assert.For(t).ThatActual(move).DoesNotEqual(typedNil)
}
