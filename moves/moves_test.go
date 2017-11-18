package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/workfit/tester/assert"
	"testing"
)

//+autoreader
type moveDealCards struct {
	DealCountComponents
}

func (m *moveDealCards) TargetCount() int {
	return 2
}

func (m *moveDealCards) GameStack(gState boardgame.MutableSubState) boardgame.MutableStack {
	return gState.(*gameState).DrawStack
}

func (m *moveDealCards) PlayerStack(pState boardgame.MutablePlayerState) boardgame.MutableStack {
	return pState.(*playerState).Hand
}

func defaultMoveInstaller(manager *boardgame.GameManager) error {
	moves := []*boardgame.MoveTypeConfig{
		&boardgame.MoveTypeConfig{
			Name: "Deal Initial Cards",
			MoveConstructor: func() boardgame.Move {
				return new(moveDealCards)
			},
			IsFixUp: true,
		},
	}

	return manager.AddOrderedMovesForPhase(phaseSetUp, moves...)
}

func TestGeneral(t *testing.T) {
	manager, err := newGameManager(defaultMoveInstaller)

	assert.For(t).ThatActual(err).IsNil()

	game := manager.NewGame()

	err = game.SetUp(0, nil, nil)

	assert.For(t).ThatActual(err).IsNil()

	//4 players, 2 rounds
	assert.For(t).ThatActual(game.Version()).Equals(8)

	gameState, playerStates := concreteStates(game.CurrentState())

	assert.For(t).ThatActual(gameState.DrawStack.NumComponents()).Equals(52 - 8)

	for _, player := range playerStates {
		assert.For(t).ThatActual(player.Hand.NumComponents()).Equals(2)
	}

}
