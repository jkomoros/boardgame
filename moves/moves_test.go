package moves

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/workfit/tester/assert"
)

//boardgame:codegen
type moveDealCards struct {
	DealCountComponents
}

func (m *moveDealCards) TargetCount(state boardgame.ImmutableState) int {
	return 2
}

func (m *moveDealCards) GameStack(gState boardgame.SubState) boardgame.Stack {
	return gState.(*gameState).DrawStack
}

func (m *moveDealCards) PlayerStack(pState boardgame.SubState) boardgame.Stack {
	return pState.(*playerState).Hand
}

//boardgame:codegen
type moveDealOtherCards struct {
	DealCountComponents
}

func (m *moveDealOtherCards) TargetCount(state boardgame.ImmutableState) int {
	return 3
}

func (m *moveDealOtherCards) GameStack(gState boardgame.SubState) boardgame.Stack {
	return gState.(*gameState).DrawStack
}

func (m *moveDealOtherCards) PlayerStack(pState boardgame.SubState) boardgame.Stack {
	return pState.(*playerState).OtherHand
}

//boardgame:codegen
type moveCurrentPlayerDraw struct {
	CurrentPlayer
}

func (m *moveCurrentPlayerDraw) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	return game.DrawStack.First().MoveToFirstSlot(p.Hand)
}

//boardgame:codegen
type moveStartPhaseDrawAgain struct {
	StartPhase
}

func (m *moveStartPhaseDrawAgain) PhaseToStart(currentPhase int) int {
	return phaseDrawAgain
}

//boardgame:codegen
type moveStartPhaseIllegal struct {
	StartPhase
}

func (m *moveStartPhaseIllegal) PhaseToStart(currentPhase int) int {
	//normal play is not a leaf node; should error
	return phaseNormalPlay
}

func defaultMoveInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {

	auto := NewAutoConfigurer(manager.Delegate())

	return Combine(
		AddOrderedForPhase(phaseSetUp,
			auto.MustConfig(
				new(moveDealCards),
				WithMoveName("Deal Components From Game Stack DrawStack To Player Stack Hand To Each Player 2 Times"),
			),
			auto.MustConfig(
				new(moveDealOtherCards),
				WithMoveName("Deal Other Cards OVERRIDE"),
			),
			auto.MustConfig(
				new(StartPhase),
				WithPhaseToStart(phaseNormalPlayDrawCard, phaseEnum),
			),
		),
		AddForPhase(phaseNormalPlay,
			auto.MustConfig(
				new(moveCurrentPlayerDraw),
				WithMoveName("Draw Card"),
			),
			auto.MustConfig(
				new(moveStartPhaseDrawAgain),
				WithMoveName("Start Phase Draw Again"),
				WithIsFixUp(false),
			),
		),
		AddOrderedForPhase(phaseDrawAgain,
			auto.MustConfig(
				new(DealComponentsUntilPlayerCountReached),
				WithMoveName("Deal Cards To Three"),
				WithGameProperty("DrawStack"),
				WithPlayerProperty("Hand"),
				WithTargetCount(3),
			),
			//Override teh AddOrderedForPhase error
			auto.MustConfig(
				new(NoOp),
			),
		),
	)
}

func noStartPhaseMoveInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {

	auto := NewAutoConfigurer(manager.Delegate())

	return AddOrderedForPhase(phaseDrawAgain,
		auto.MustConfig(
			new(DealComponentsUntilPlayerCountReached),
			WithMoveName("Deal Cards To Three"),
			WithGameProperty("DrawStack"),
			WithPlayerProperty("Hand"),
			WithTargetCount(3),
		),
	)
}

func TestAddOrderedForPhaseEndsWithStartPhase(t *testing.T) {
	_, err := newGameManager(noStartPhaseMoveInstaller, false)
	assert.For(t).ThatActual(err).IsNotNil()
}

func illegalPhaseMoveInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {

	auto := NewAutoConfigurer(manager.Delegate())

	return []boardgame.MoveConfig{
		auto.MustConfig(new(moveStartPhaseIllegal)),
	}
}

func TestPhaseIllegalConfig(t *testing.T) {
	_, err := newGameManager(illegalPhaseMoveInstaller, false)

	assert.For(t).ThatActual(err).IsNotNil()

	failedBecauseTreeEnum := strings.Contains(err.Error(), "TreeEnum")

	assert.For(t).ThatActual(failedBecauseTreeEnum).IsTrue()
}

func TestGeneral(t *testing.T) {
	manager, err := newGameManager(defaultMoveInstaller, false)

	assert.For(t).ThatActual(err).IsNil()

	game, err := manager.NewDefaultGame()

	assert.For(t).ThatActual(err).IsNil()

	//4 players, 2 rounds for inital cards, then 4 * 3 for other cards, then
	//NewStartPhase.
	assert.For(t).ThatActual(game.Version()).Equals(21)

	gameState, playerStates := concreteStates(game.CurrentState())

	assert.For(t).ThatActual(gameState.DrawStack.NumComponents()).Equals(52 - 20)
	assert.For(t).ThatActual(gameState.Phase.Value()).Equals(phaseNormalPlayDrawCard)

	for i, player := range playerStates {
		assert.For(t, i).ThatActual(player.Hand.NumComponents()).Equals(2)
		assert.For(t, i).ThatActual(player.OtherHand.NumComponents()).Equals(3)
	}

	historicalMovesCount(t,
		[]string{
			"Deal Components From Game Stack DrawStack To Player Stack Hand To Each Player 2 Times",
			"Deal Other Cards OVERRIDE",
			"Start Phase Normal Play > Draw Card",
		},
		[]int{
			8,
			12,
			1,
		}, game.MoveRecords(-1))

	move := game.MoveByName("Draw Card")

	assert.For(t).ThatActual(move).IsNotNil()

	err = <-game.ProposeMove(move, 0)

	assert.For(t).ThatActual(err).IsNil()

	move = game.MoveByName("Start Phase Draw Again")

	assert.For(t).ThatActual(move).IsNotNil()

	err = <-game.ProposeMove(move, 0)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(manager.Delegate().CurrentPhase(game.CurrentState())).Equals(phaseDrawAgain)

	//3 additional moves, but skipping the one player who already had 3 in
	//their hand, plus the one terminal no op.
	assert.For(t).ThatActual(game.Version()).Equals(27)

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
