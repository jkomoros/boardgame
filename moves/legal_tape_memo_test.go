package moves

import (
	"reflect"
	"testing"

	"github.com/jkomoros/boardgame/legal"
	"github.com/workfit/tester/assert"
)

/*
legal_tape_memo_test.go exercises design spec §5's "Tape memoization"
engine win from the moves-package side: historicalMovesSincePhaseTransition
(default.go) now wraps computeHistoricalMovesSincePhaseTransition in
game.LegalTapeMemo (boardgame/legal_memo.go), retiring the TODO that used
to sit right there. It reuses preconditions_test.go's freezeMoveInstaller
fixture ("Freeze Progression A"/"Freeze Progression B", an ordered
progression in phaseNormalPlayDrawCard) — the SAME fixture
inprogression_test.go's TestInProgressionMatch/Reject use, so this file's
tests are provably exercising the identical game setup those already-green
tests do, just from a different angle (memoization, not just correctness).
*/

// TestHistoricalMovesSincePhaseTransitionMemoAgreesWithUnmemoized proves the
// "frozen chain path must produce IDENTICAL results" requirement directly:
// the memoized entry point (historicalMovesSincePhaseTransition) and the
// raw, unmemoized computation it wraps (computeHistoricalMovesSincePhaseTransition)
// return the exact same tape for the same (game, version, phase).
func TestHistoricalMovesSincePhaseTransitionMemoAgreesWithUnmemoized(t *testing.T) {
	manager, err := newGameManager(freezeMoveInstaller)
	assert.For(t, "manager").ThatActual(err).IsNil()

	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	startMove := game.MoveByName("Freeze Start Normal Play")
	assert.For(t, "start move").ThatActual(startMove).IsNotNil()
	startErr := <-game.ProposeMove(startMove, 0)
	assert.For(t, "start move propose").ThatActual(startErr).IsNil()

	state := game.CurrentState()
	_, currentPhase := currentPhaseInfo(state)

	d := &Default{}
	memoized := d.historicalMovesSincePhaseTransition(game, state.Version(), currentPhase)
	raw := computeHistoricalMovesSincePhaseTransition(game, state.Version(), currentPhase)

	if !reflect.DeepEqual(memoized, raw) {
		t.Fatalf("memoized tape = %+v, want identical to unmemoized %+v", memoized, raw)
	}
}

// TestInProgressionTapeSharedAcrossMoveTypes proves the tape memo is
// genuinely shared BY (GAME, VERSION) — not accidentally scoped or
// polluted per move type: resolving the "inProgression" predicate for TWO
// DIFFERENT move types ("Freeze Progression A" then "Freeze Progression B")
// against the exact same state consults the SAME cached (game, version,
// phase) tape entry for both, yet each move type still gets its own
// correct, independent verdict (A next-in-progression: Pass; B out of
// order: Fail) — a caching bug that accidentally keyed or corrupted the
// cache per-move-type would surface here as a wrong verdict for whichever
// move queries second.
func TestInProgressionTapeSharedAcrossMoveTypes(t *testing.T) {
	manager, err := newGameManager(freezeMoveInstaller)
	assert.For(t, "manager").ThatActual(err).IsNil()

	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	startMove := game.MoveByName("Freeze Start Normal Play")
	assert.For(t, "start move").ThatActual(startMove).IsNotNil()
	startErr := <-game.ProposeMove(startMove, 0)
	assert.For(t, "start move propose").ThatActual(startErr).IsNil()

	state := game.CurrentState()

	predA := resolveInProgressionForTest(t, "Freeze Progression A")
	predB := resolveInProgressionForTest(t, "Freeze Progression B")

	// Query B FIRST (populating/consulting the shared (game, version,
	// phase) tape memo before A ever touches it), then A, then A again — if
	// the cache were wrongly scoped per-move-type, or corrupted by
	// whichever move queries it first, one of these would diverge from
	// TestLegalChainStringFreeze's independently-pinned frozen-chain
	// results.
	vB := predB.Evaluate(legal.Context{State: state, Proposer: 0})
	if vB.Outcome != legal.Fail {
		t.Fatalf("Freeze Progression B (queried first): Outcome = %v, want Fail (%+v)", vB.Outcome, vB)
	}

	vA := predA.Evaluate(legal.Context{State: state, Proposer: 0})
	if vA.Outcome != legal.Pass {
		t.Fatalf("Freeze Progression A (queried second, after B populated the shared tape): Outcome = %v, want Pass (%+v)", vA.Outcome, vA)
	}

	vA2 := predA.Evaluate(legal.Context{State: state, Proposer: 0})
	if vA2.Outcome != legal.Pass {
		t.Fatalf("Freeze Progression A (queried a third time): Outcome = %v, want Pass (%+v)", vA2.Outcome, vA2)
	}
}
