package moves

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
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

func (m *moveStartPhaseDrawAgain) PhaseToStart(currentPhase enum.EnumKey) (enum.EnumKey, error) {
	return phaseDrawAgain, nil
}

//boardgame:codegen
type moveStartPhaseIllegal struct {
	StartPhase
}

func (m *moveStartPhaseIllegal) PhaseToStart(currentPhase enum.EnumKey) (enum.EnumKey, error) {
	//normal play is not a leaf node; should error
	return phaseNormalPlay, nil
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
	_, err := newGameManager(noStartPhaseMoveInstaller)
	assert.For(t).ThatActual(err).IsNotNil()
}

// zeroIterationRoundRobinMoveInstaller mirrors the phaseSetUp progression of
// defaultMoveInstaller, but the first move is a
// DealComponentsUntilPlayerCountReached whose TargetCount is 0. Because every
// player already has 0 or more components in their Hand, every player's
// PlayerConditionMet is already true, so the round robin's ConditionMet is
// satisfied before the move ever starts -- a zero-iteration round robin.
func zeroIterationRoundRobinMoveInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {

	auto := NewAutoConfigurer(manager.Delegate())

	return Combine(
		AddOrderedForPhase(phaseSetUp,
			auto.MustConfig(
				new(DealComponentsUntilPlayerCountReached),
				WithMoveName("Deal Cards Until Zero"),
				WithGameProperty("DrawStack"),
				WithPlayerProperty("Hand"),
				WithTargetCount(0),
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
		),
	)
}

// TestZeroIterationRoundRobinLegal is a framework-level regression test for a
// bug where a round robin move whose ConditionMet was already satisfied before
// it started (e.g. a DealComponentsUntilPlayerCountReached targeting a count
// every player already meets) would trip Apply()'s "found to be finished in
// our Apply, but it should have been marked finished before" invariant,
// failing game creation. The fix lives in RoundRobin.Apply: after
// startRoundRobin resets the shared round-robin state, an already-met
// condition finishes the round robin cleanly (zero iterations) and returns
// nil. It deliberately does NOT live in Legal(): ConditionMet reads the
// shared RoundRobinRoundCount, which is only reset inside Apply's
// startRoundRobin, so a Legal-side check reads stale state from the previous
// round robin and mis-rejects legitimate starts (a Legal-based fix regressed
// blackjack; see round_robin.go's comment). This test asserts game creation
// succeeds and the zero-iteration deal dealt nothing.
func TestZeroIterationRoundRobinLegal(t *testing.T) {
	manager, err := newGameManager(zeroIterationRoundRobinMoveInstaller)

	assert.For(t).ThatActual(err).IsNil()

	game, err := manager.NewDefaultGame()

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(game).IsNotNil()

	//The zero-iteration deal should have dealt nothing.
	_, playerStates := concreteStates(game.CurrentState())
	for i, player := range playerStates {
		assert.For(t, i).ThatActual(player.Hand.NumComponents()).Equals(0)
	}
}

func illegalPhaseMoveInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {

	auto := NewAutoConfigurer(manager.Delegate())

	return []boardgame.MoveConfig{
		auto.MustConfig(new(moveStartPhaseIllegal)),
	}
}

func TestPhaseIllegalConfig(t *testing.T) {
	_, err := newGameManager(illegalPhaseMoveInstaller)

	assert.For(t).ThatActual(err).IsNotNil()

	failedBecauseTreeEnum := strings.Contains(err.Error(), "TreeEnum")

	assert.For(t).ThatActual(failedBecauseTreeEnum).IsTrue()
}

func TestGeneral(t *testing.T) {
	manager, err := newGameManager(defaultMoveInstaller)

	assert.For(t).ThatActual(err).IsNil()

	game, err := manager.NewDefaultGame()

	assert.For(t).ThatActual(err).IsNil()

	//4 players, 2 rounds for inital cards, then 4 * 3 for other cards, then
	//NewStartPhase.
	assert.For(t).ThatActual(game.Version()).Equals(21)

	gameState, playerStates := concreteStates(game.CurrentState())

	assert.For(t).ThatActual(gameState.DrawStack.NumComponents()).Equals(52 - 20)
	assert.For(t).ThatActual(gameState.Phase.Value()).Equals(enum.EnumKey(phaseNormalPlayDrawCard))

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

	assert.For(t).ThatActual(manager.Delegate().CurrentPhase(game.CurrentState()).Value()).Equals(enum.EnumKey(phaseDrawAgain))

	//3 additional moves, but skipping the one player who already had 3 in
	//their hand, plus the one terminal no op.
	assert.For(t).ThatActual(game.Version()).Equals(27)

}

func TestConfigureFromTagsErrorPaths(t *testing.T) {
	// Test ConfigureFromTags error branches directly using a bare (uninflated)
	// playerState. The playerState implements SubState via base.SubState and
	// codegen, so it satisfies the interface even without full game setup.
	bare := new(playerState)
	loc := &behaviors.LocationBehavior{}

	// Field does not exist
	err := loc.ConfigureFromTags(reflect.StructTag(`location:"NoSuchField"`), bare)
	assert.For(t, "nonexistent field").ThatActual(err).IsNotNil()
	assert.For(t, "nonexistent msg").ThatActual(strings.Contains(err.Error(), "NoSuchField")).IsTrue()

	// Field exists but is not a SizedStack (Counter is an int)
	err = loc.ConfigureFromTags(reflect.StructTag(`location:"Counter"`), bare)
	assert.For(t, "wrong type").ThatActual(err).IsNotNil()
	assert.For(t, "wrong type msg").ThatActual(strings.Contains(err.Error(), "Counter")).IsTrue()
	assert.For(t, "wrong type msg detail").ThatActual(strings.Contains(err.Error(), "SizedStack")).IsTrue()

	// Field is a SizedStack type but nil (uninflated) — a nil interface
	// doesn't pass the type assertion, so this also reports "not a SizedStack".
	err = loc.ConfigureFromTags(reflect.StructTag(`location:"TokenLocation"`), bare)
	assert.For(t, "nil stack").ThatActual(err).IsNotNil()
	assert.For(t, "nil stack msg").ThatActual(strings.Contains(err.Error(), "SizedStack")).IsTrue()
}

func TestLocationBehaviorTagWiring(t *testing.T) {
	// This test verifies that the `location:"TokenLocation"` struct tag on
	// playerState.LocationBehavior auto-wires ConnectLocationStack during
	// state setup. NewGameManager calls ValidConfiguration internally, so
	// if it succeeds, the location tag was processed correctly.
	manager, err := newGameManager(defaultMoveInstaller)
	assert.For(t, "NewGameManager").ThatActual(err).IsNil()

	game, err := manager.NewDefaultGame()
	assert.For(t, "NewDefaultGame").ThatActual(err).IsNil()

	_, playerStates := concreteStates(game.CurrentState())

	for i, player := range playerStates {
		// Each player should have one token in their TokenLocation stack.
		assert.For(t, "TokenLocation components", i).ThatActual(player.TokenLocation.NumComponents()).Equals(1)

		// LocationBehavior should be usable (LocationIndexKey returns a valid key).
		key, found := player.LocationBehavior.LocationIndexKey()
		assert.For(t, "LocationIndexKey found", i).ThatActual(found).IsTrue()
		assert.For(t, "LocationIndexKey value", i).ThatActual(int(key)).Equals(0)
	}
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
