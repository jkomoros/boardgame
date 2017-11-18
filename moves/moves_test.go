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

//+autoreader
type moveDealOtherCards struct {
	DealCountComponents
}

func (m *moveDealOtherCards) TargetCount() int {
	return 3
}

func (m *moveDealOtherCards) GameStack(gState boardgame.MutableSubState) boardgame.MutableStack {
	return gState.(*gameState).DrawStack
}

func (m *moveDealOtherCards) PlayerStack(pState boardgame.MutablePlayerState) boardgame.MutableStack {
	return pState.(*playerState).OtherHand
}

//+autoreader
type moveCurrentPlayerDraw struct {
	CurrentPlayer
}

func (m *moveCurrentPlayerDraw) Apply(state boardgame.MutableState) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	return game.DrawStack.MoveComponent(boardgame.FirstComponentIndex, p.Hand, boardgame.FirstSlotIndex)
}

//+autoreader
type moveStartPhaseDrawAgain struct {
	StartPhase
}

func (m *moveStartPhaseDrawAgain) PhaseToStart(currentPhase int) int {
	return phaseDrawAgain
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
		&boardgame.MoveTypeConfig{
			Name: "Deal Other Cards",
			MoveConstructor: func() boardgame.Move {
				return new(moveDealOtherCards)
			},
			IsFixUp: true,
		},
		NewStartPhaseMoveConfig(manager, phaseNormalPlay, nil),
	}

	if err := manager.AddOrderedMovesForPhase(phaseSetUp, moves...); err != nil {
		return err
	}

	moves = []*boardgame.MoveTypeConfig{
		&boardgame.MoveTypeConfig{
			Name: "Draw Card",
			MoveConstructor: func() boardgame.Move {
				return new(moveCurrentPlayerDraw)
			},
		},
		&boardgame.MoveTypeConfig{
			Name: "Start Phase Draw Again",
			MoveConstructor: func() boardgame.Move {
				return new(moveStartPhaseDrawAgain)
			},
		},
	}

	return manager.AddMovesForPhase(phaseNormalPlay, moves...)
}

func TestGeneral(t *testing.T) {
	manager, err := newGameManager(defaultMoveInstaller)

	assert.For(t).ThatActual(err).IsNil()

	game := manager.NewGame()

	err = game.SetUp(0, nil, nil)

	assert.For(t).ThatActual(err).IsNil()

	//4 players, 2 rounds for inital cards, then 4 * 3 for other cards, then
	//NewStartPhase.
	assert.For(t).ThatActual(game.Version()).Equals(21)

	gameState, playerStates := concreteStates(game.CurrentState())

	assert.For(t).ThatActual(gameState.DrawStack.NumComponents()).Equals(52 - 20)
	assert.For(t).ThatActual(gameState.Phase.Value()).Equals(phaseNormalPlay)

	for i, player := range playerStates {
		assert.For(t, i).ThatActual(player.Hand.NumComponents()).Equals(2)
		assert.For(t, i).ThatActual(player.OtherHand.NumComponents()).Equals(3)
	}

	historicalMovesCount(t,
		[]string{
			"Deal Initial Cards",
			"Deal Other Cards",
			"Start Phase Normal Play",
		},
		[]int{
			8,
			12,
			1,
		}, game.MoveRecords(-1))

	move := game.PlayerMoveByName("Draw Card")

	assert.For(t).ThatActual(move).IsNotNil()

	err = <-game.ProposeMove(move, 0)

	assert.For(t).ThatActual(err).IsNil()

	move = game.PlayerMoveByName("Start Phase Draw Again")

	assert.For(t).ThatActual(move).IsNotNil()

	err = <-game.ProposeMove(move, 0)

	assert.For(t).ThatActual(err).IsNil()

}

func historicalMovesCount(t *testing.T, moveNames []string, counts []int, records []*boardgame.MoveStorageRecord) {
	if len(moveNames) != len(counts) {
		t.Error("MoveNames and counts did not match length")
	}

	currentMoveIndex := 0
	counterInMoveType := 0

	for i, move := range records {

		if counterInMoveType >= counts[currentMoveIndex] {
			currentMoveIndex++
			counterInMoveType = 0
			if currentMoveIndex > len(moveNames) {
				t.Error("Fell off end of configuration")
			}
		}

		if move.Name != moveNames[currentMoveIndex] {
			t.Error("Unexpected move at ", i, move.Name, moveNames[currentMoveIndex])
		}

		counterInMoveType++
	}
}
