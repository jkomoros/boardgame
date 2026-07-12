package moves

import (
	"strconv"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
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

/*
TestFixUpMultiProgressionAtomEquivalence is the design spec §5 / Task 6 Step 1
gate: it must run and pass BEFORE the seam allowlist (legal_plan.go's
legalUnsupportedMovesBaseType) is widened to admit moves.FixUpMulti.

The claim being verified: the "inProgression" declarative atom
(moves/catalog_framework.go's inProgressionConstructor) honors
AllowMultipleInProgression() the same way moves.Default's frozen imperative
chain does, for a moves.FixUpMulti-embedding move specifically (FixUpMulti is
the one framework type in the widened allowlist whose AllowMultipleInProgression()
override actually changes progression-matching behavior — FixUp and StartPhase
don't touch it). This holds BY CONSTRUCTION: inProgressionConstructor's
Evaluate does not reimplement progression matching, it calls
checker.legalMoveInProgression(ctx.State, ctx.Proposer) — literally the same
private method (moves/default.go, promoted from *Default through
FixUpMulti's embed chain: FixUpMulti -> FixUp -> Default) that
Default.Legal()'s frozen chain calls, and AllowMultipleInProgression is
consulted exactly once in that whole path (auto_config.go's
defaultMoveConfig.Satisfied). There is only one code path, so the two
"views" cannot diverge for FixUpMulti — but this test proves it empirically
against a real, live tape (not just by reading the source), per the Task 6
brief's "prove it, don't just assert it by construction" instruction.

Because the seam allowlist does not yet admit FixUpMulti, this test cannot
opt the move in via WithPreconditions and compare against Legal() itself
(that path is exactly what Step 2 unlocks). Instead it drives a REAL game
tape through a NOT-opted-in FixUpMulti-embedding move (so its Legal() runs
the ordinary frozen chain and ProposeMove's own legality gate is the source
of truth for what actually got applied), and at each tape position calls
BOTH the frozen chain's legalMoveInProgression directly AND a hand-built
"inProgression" predicate (via inProgressionConstructor(), the exact
constructor NewGameManager's plan assembly will use once FixUpMulti is
allowlisted) against that same state, asserting their pass/fail verdicts
agree — including across multiple repeated occurrences of the move in a row
(AllowMultipleInProgression's actual effect) and after the progression closes
out (where a further repeat becomes illegal for both).
*/

//boardgame:codegen
type moveFreezeFixUpMultiProgression struct {
	FixUpMulti
}

func (m *moveFreezeFixUpMultiProgression) Apply(state boardgame.State) error {
	return nil
}

// fixUpMultiProgressionInstaller wires an ordered progression in
// phaseNormalPlayDrawCard consisting of a FixUpMulti-embedding move (so
// AllowMultipleInProgression() is true, inherited from FixUpMulti) followed
// by a NoOp guard that closes the progression out. The FixUpMulti move is
// NOT opted in (no WithPreconditions — the seam does not admit it yet) and is
// configured WithIsFixUp(false) so it can be proposed manually, mirroring
// freezeMoveInstaller's "Freeze Start Normal Play" pattern above.
func fixUpMultiProgressionInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {
	auto := NewAutoConfigurer(manager.Delegate())

	return Combine(
		AddOrderedForPhase(phaseSetUp,
			auto.MustConfig(
				new(StartPhase),
				WithMoveName("FixUpMulti Gate Start Normal Play"),
				WithPhaseToStart(phaseNormalPlayDrawCard, phaseEnum),
				WithIsFixUp(false),
			),
		),
		AddOrderedForPhase(phaseNormalPlayDrawCard,
			auto.MustConfig(
				new(moveFreezeFixUpMultiProgression),
				WithMoveName("FixUpMulti Gate Progression"),
				WithIsFixUp(false),
			),
			auto.MustConfig(
				new(NoOp),
				WithMoveName("FixUpMulti Gate Guard"),
				// NoOp embeds FixUp (defaults IsFixUp true), and — per
				// moves/default.go's legalMoveInProgression doc comment —
				// legalMoveInProgression alone never decides an
				// AllowMultipleInProgression move is "no longer legal"; that's
				// normally the embedding move's own imperative Legal()
				// override's job. moveFreezeFixUpMultiProgression below has no
				// such override, so without this, the guard would already be
				// simultaneously legal after the FIRST progression repetition
				// and the engine's automatic post-move FixUp sweep would apply
				// it immediately, closing the progression out before this test
				// can manually drive multiple repetitions. WithIsFixUp(false)
				// makes the guard require an explicit ProposeMove, like
				// "Freeze Start Normal Play" above.
				WithIsFixUp(false),
			),
		),
	)
}

// fixUpMultiInProgressionAtomLegal builds and evaluates the "inProgression"
// predicate (moves/catalog_framework.go's inProgressionConstructor) for
// moveName against state/proposer — by hand, via the exact constructor
// NewGameManager's plan assembly consults, since the seam does not yet admit
// FixUpMulti into that assembly path. Returns whether the atom's verdict is
// LegalPass.
func fixUpMultiInProgressionAtomLegal(t *testing.T, moveName string, state boardgame.ImmutableState, proposer boardgame.PlayerIndex) bool {
	t.Helper()

	predicate, err := inProgressionConstructor().Constructor(inProgressionSpec(moveName), nil, nil)
	if !assert.For(t, "inProgression predicate construction").ThatActual(err).IsNil().Passed() {
		return false
	}
	if predicate == nil {
		t.Fatal("inProgression predicate constructor returned a nil predicate")
	}

	verdict := predicate.Evaluate(legal.Context{State: state, Proposer: proposer})
	return verdict.Outcome == legal.Pass
}

func TestFixUpMultiProgressionAtomEquivalence(t *testing.T) {
	manager, err := newGameManager(fixUpMultiProgressionInstaller)
	assert.For(t, "manager").ThatActual(err).IsNil()
	if manager == nil {
		return
	}

	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	startMove := game.MoveByName("FixUpMulti Gate Start Normal Play")
	assert.For(t, "start move").ThatActual(startMove).IsNotNil()
	startErr := <-game.ProposeMove(startMove, 0)
	assert.For(t, "start move propose").ThatActual(startErr).IsNil()

	const moveName = "FixUpMulti Gate Progression"

	// checkAgreement compares, at the game's CURRENT state, the frozen
	// chain's legalMoveInProgression verdict against the "inProgression"
	// atom's verdict for moveName, asserting they agree (both legal or both
	// illegal).
	checkAgreement := func(step string) {
		state := game.CurrentState()

		chainMove, ok := game.MoveByName(moveName).(*moveFreezeFixUpMultiProgression)
		if !assert.For(t, step, "move type").ThatActual(ok).IsTrue().Passed() {
			return
		}
		chainLegal := chainMove.legalMoveInProgression(state, 0) == nil
		atomLegal := fixUpMultiInProgressionAtomLegal(t, moveName, state, 0)

		assert.For(t, step, "chain legal").ThatActual(chainLegal).Equals(atomLegal)
	}

	// Before any occurrence: both the chain and the atom must agree the
	// first occurrence of the progression move is legal.
	checkAgreement("before any repetitions")

	// Propose the FixUpMulti move three times in a row. AllowMultipleInProgression
	// must keep every repetition legal for BOTH views, since they call the
	// exact same underlying method.
	for i := 0; i < 3; i++ {
		move := game.MoveByName(moveName)
		proposeErr := <-game.ProposeMove(move, 0)
		assert.For(t, "propose repetition", i).ThatActual(proposeErr).IsNil()
		checkAgreement("after repetition " + strconv.Itoa(i))
	}

	// Close the progression out with the guard move.
	guardMove := game.MoveByName("FixUpMulti Gate Guard")
	guardErr := <-game.ProposeMove(guardMove, 0)
	assert.For(t, "guard propose").ThatActual(guardErr).IsNil()

	// The progression has moved past the FixUpMulti move: a further
	// repetition is illegal for BOTH the chain and the atom.
	checkAgreement("after guard closes the progression")
}
