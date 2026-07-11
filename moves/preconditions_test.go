package moves

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/workfit/tester/assert"
)

/*
This file's TestLegalChainStringFreeze is the string-freeze test called for by
the declarative-legality-design runbook's Task 7: it captures, verbatim,
today's (pre-refactor) Legal() error strings for moves.Default's imperative
chain (legalInPhase / legalMoveInProgression / legalStackConstraints) and
moves.CurrentPlayer's proposer checks, using the SAME game fixtures the rest
of this package's tests use (gameState/playerState/phaseEnum/newGameManager,
all defined in game_test.go). It must stay green through every subsequent
refactor in this task: if a refactor changes one of these strings, the
refactor is wrong, not this test. See the design spec's "prime guarantee": the
imperative chain is frozen, byte-for-byte.
*/

//boardgame:codegen
type moveFreezeInPhase struct {
	Default
}

func (m *moveFreezeInPhase) Apply(state boardgame.State) error {
	return nil
}

//boardgame:codegen
type moveFreezeProgressionA struct {
	Default
}

func (m *moveFreezeProgressionA) Apply(state boardgame.State) error {
	return nil
}

//boardgame:codegen
type moveFreezeProgressionB struct {
	Default
}

func (m *moveFreezeProgressionB) Apply(state boardgame.State) error {
	return nil
}

//boardgame:codegen
type moveFreezeStack struct {
	Default
}

func (m *moveFreezeStack) Apply(state boardgame.State) error {
	return nil
}

//boardgame:codegen
type moveFreezeCurrentPlayer struct {
	CurrentPlayer
}

func (m *moveFreezeCurrentPlayer) Apply(state boardgame.State) error {
	return nil
}

// freezeMoveInstaller wires up a small, self-contained set of moves (reusing
// the shared gameState/playerState/phaseEnum fixture from game_test.go)
// purpose-built to exercise each branch of moves.Default's and
// moves.CurrentPlayer's imperative Legal() chain in isolation:
//   - moveFreezeInPhase: only legal in phaseDrawAgain, never entered, so its
//     Legal() always exercises legalInPhase's rejection branch.
//   - moveFreezeProgressionA/B (+ a NoOp guard): an ordered move progression
//     in phaseNormalPlayDrawCard, so proposing B before A exercises
//     legalMoveInProgression's rejection branch.
//   - moveFreezeStack: WithSourceProperty/WithDestinationProperty both set to
//     "DrawStack", so it exercises legalStackConstraints' rejection branch
//     via component.go's "source and destination are the same stack" check.
//   - moveFreezeCurrentPlayer: exercises CurrentPlayer.Legal()'s two proposer
//     error strings.
//   - "Freeze Start Normal Play": a non-fixup StartPhase move used only to
//     move the fixture from its default initial phase (phaseSetUp) into
//     phaseNormalPlayDrawCard, so the phase-gated cases above have a phase to
//     be legal in.
func freezeMoveInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {

	auto := NewAutoConfigurer(manager.Delegate())

	return Combine(
		AddOrderedForPhase(phaseSetUp,
			auto.MustConfig(
				new(StartPhase),
				WithMoveName("Freeze Start Normal Play"),
				WithPhaseToStart(phaseNormalPlayDrawCard, phaseEnum),
				WithIsFixUp(false),
			),
		),
		AddOrderedForPhase(phaseNormalPlayDrawCard,
			auto.MustConfig(
				new(moveFreezeProgressionA),
				WithMoveName("Freeze Progression A"),
			),
			auto.MustConfig(
				new(moveFreezeProgressionB),
				WithMoveName("Freeze Progression B"),
			),
			auto.MustConfig(
				new(NoOp),
				WithMoveName("Freeze Progression Guard"),
			),
		),
		Add(
			auto.MustConfig(
				new(moveFreezeInPhase),
				WithMoveName("Freeze In Phase"),
				WithLegalPhases(phaseDrawAgain),
			),
			auto.MustConfig(
				new(moveFreezeStack),
				WithMoveName("Freeze Stack"),
				WithLegalPhases(phaseNormalPlayDrawCard),
				WithSourceProperty("DrawStack"),
				WithDestinationProperty("DrawStack"),
			),
			auto.MustConfig(
				new(moveFreezeCurrentPlayer),
				WithMoveName("Freeze Current Player"),
				WithLegalPhases(phaseNormalPlayDrawCard),
			),
		),
	)
}

// freezeCurrentPlayerWithTarget returns a fresh "Freeze Current Player" move
// (defaults set for game's current state) with TargetPlayerIndex overridden
// to target, for exercising CurrentPlayer.Legal()'s proposer-check branches.
func freezeCurrentPlayerWithTarget(game *boardgame.Game, target boardgame.PlayerIndex) boardgame.Move {
	move := game.MoveByName("Freeze Current Player").(*moveFreezeCurrentPlayer)
	move.TargetPlayerIndex = target
	return move
}

func TestLegalChainStringFreeze(t *testing.T) {

	manager, err := newGameManager(freezeMoveInstaller)
	assert.For(t, "manager").ThatActual(err).IsNil()

	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	// Captured BEFORE the phase transition below, while the game is still in
	// its default initial phase (phaseSetUp, per enum.Set.MustAddTree's
	// BranchDefaultValue: the first leaf descendant of the phase tree's
	// root). moveFreezeInPhase is legal only in phaseDrawAgain, so this
	// exercises moves/default.go's legalInPhase rejection.
	inPhaseMove := game.MoveByName("Freeze In Phase")
	assert.For(t, "in phase move").ThatActual(inPhaseMove).IsNotNil()
	inPhaseErr := inPhaseMove.Legal(game.CurrentState(), 0)
	assert.For(t, "wrong phase").ThatActual(inPhaseErr).IsNotNil()
	if inPhaseErr != nil {
		assert.For(t, "wrong phase string").ThatActual(inPhaseErr.Error()).Equals("Move is not legal in phase Set Up")
	}

	// Transition into phaseNormalPlayDrawCard so the remaining cases below
	// are captured against a state where the phase-gated moves are otherwise
	// legal (isolating each case to the ONE chain link it targets).
	startMove := game.MoveByName("Freeze Start Normal Play")
	assert.For(t, "start move").ThatActual(startMove).IsNotNil()
	startErr := <-game.ProposeMove(startMove, 0)
	assert.For(t, "start move propose").ThatActual(startErr).IsNil()

	state := game.CurrentState()

	tests := []struct {
		description string
		move        boardgame.Move
		proposer    boardgame.PlayerIndex
		wantErr     string // "" means Legal() must return nil.
	}{
		{
			"progression: B proposed before A (legalMoveInProgression)",
			game.MoveByName("Freeze Progression B"),
			0,
			"Move name does not match: Freeze Progression B is not Freeze Progression A",
		},
		{
			"stack constraint: source == destination (legalStackConstraints)",
			game.MoveByName("Freeze Stack"),
			0,
			"source and destination are the same stack",
		},
		{
			"current player: invalid target (CurrentPlayer.Legal, ObserverPlayerIndex)",
			freezeCurrentPlayerWithTarget(game, boardgame.ObserverPlayerIndex),
			0,
			"The specified target player is not valid",
		},
		{
			"current player: not your turn (CurrentPlayer.Legal, wrong target)",
			freezeCurrentPlayerWithTarget(game, 1),
			1,
			"it's not your turn",
		},
		{
			"passing case: current player move with correct defaults",
			game.MoveByName("Freeze Current Player"),
			0,
			"",
		},
	}

	for i, test := range tests {
		gotErr := test.move.Legal(state, test.proposer)
		if test.wantErr == "" {
			assert.For(t, i, test.description).ThatActual(gotErr).IsNil()
			continue
		}
		if !assert.For(t, i, test.description).ThatActual(gotErr).IsNotNil().Passed() {
			continue
		}
		assert.For(t, i, test.description).ThatActual(gotErr.Error()).Equals(test.wantErr)
	}

}
