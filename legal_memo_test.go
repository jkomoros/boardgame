package boardgame

import (
	"testing"

	"github.com/workfit/tester/assert"
)

/*
legal_memo_test.go exercises design spec §5's two per-game memos
(legal_memo.go): the field-independent legality memo (consulted from
legal_plan.go's legalPlan.evaluate) and the move-tape memo (LegalTapeMemo,
consumed by moves/default.go's historicalMovesSincePhaseTransition — see
moves/legal_tape_memo_test.go for the cross-package sharing test, since that
requires package moves' Default/inProgression machinery, which core cannot
import).
*/

// legalMemoTestMove is a plain, non-fixup, always-legal move (baseMove, NOT
// baseFixUpMove — a fixup version would get auto-applied forever by
// ProposeFixUpMove during setUp, since it's always legal). Used only to
// deterministically advance a test game's version.
type legalMemoTestMove struct {
	baseMove
}

func (m *legalMemoTestMove) Reader() PropertyReader { return getDefaultReader(m) }
func (m *legalMemoTestMove) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(m)
}
func (m *legalMemoTestMove) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(m)
}
func (m *legalMemoTestMove) Legal(state ImmutableState, proposer PlayerIndex) error { return nil }
func (m *legalMemoTestMove) Apply(state State) error                                { return nil }

var legalMemoTestMoveConfig = NewMoveConfig(
	"Legal Memo Test Move",
	func() Move { return new(legalMemoTestMove) },
	nil,
)

// newLegalMemoTestGame boots a minimal manager whose only move type is
// legalMemoTestMove (always legal, does nothing on Apply, never a fixup),
// then sets up and returns a real, running *Game — giving these tests a
// genuine version sequence to advance through applyMove, without any of the
// precondition fuss a "real" move type would bring.
func newLegalMemoTestGame(t *testing.T) *Game {
	t.Helper()

	moveInstaller := func(manager *GameManager) []MoveConfig {
		return []MoveConfig{legalMemoTestMoveConfig}
	}

	manager, err := NewGameManager(&testGameDelegate{moveInstaller: moveInstaller}, newTestStorageManager())
	assert.For(t).ThatActual(err).IsNil()

	game, err := manager.newGameImpl("", "")
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(game.setUp(0, nil, nil)).IsNil()

	return game
}

// legalMemoAdvanceVersion applies legalMemoTestMove directly (bypassing the
// ProposeMove channel/mainLoop dance — game.applyMove is synchronous and
// unexported, fine to call directly within this package) to
// deterministically advance game's version by exactly one.
func legalMemoAdvanceVersion(t *testing.T, game *Game) {
	t.Helper()
	move := game.MoveByName("Legal Memo Test Move")
	assert.For(t).ThatActual(move).IsNotNil()
	assert.For(t).ThatActual(game.applyMove(move, AdminPlayerIndex, false, 0, selfInitiatorSentinel)).IsNil()
}

// countingFieldIndependentPlan returns a legalPlan with a single
// fieldIndependent predicate that increments *calls every time it's
// actually evaluated, so tests can distinguish a memo hit (calls unchanged)
// from a miss (calls incremented).
func countingFieldIndependentPlan(moveName string, calls *int) *legalPlan {
	return &legalPlan{
		moveName: moveName,
		fieldIndependent: []*LegalPredicate{
			{
				Name: "counting",
				Evaluate: func(ctx LegalContext) LegalVerdict {
					*calls++
					return LegalVerdict{Outcome: LegalPass}
				},
			},
		},
	}
}

// TestLegalPlanFieldIndependentMemoHitAcrossDoubleEvaluation is the
// concrete scenario the memo exists for (design spec §5's honest table):
// base/game_delegate.go's ProposeFixUpMove evaluates a candidate fixup
// move's Legal() to select it, then game.go's applyMove evaluates Legal()
// AGAIN on that exact move at the exact same (version, proposer=Admin) to
// actually apply it. The field-independent bucket must be computed once,
// not twice.
func TestLegalPlanFieldIndependentMemoHitAcrossDoubleEvaluation(t *testing.T) {
	game := newLegalMemoTestGame(t)
	state := game.CurrentState()

	calls := 0
	plan := countingFieldIndependentPlan("MemoTestMove", &calls)
	ctx := LegalContext{State: state, Proposer: AdminPlayerIndex}

	v1, _ := plan.evaluate(ctx, false)
	assert.For(t).ThatActual(v1.Outcome).Equals(LegalPass)
	assert.For(t).ThatActual(calls).Equals(1)

	// Same (moveName, version, proposer): HIT, no re-evaluation.
	v2, _ := plan.evaluate(ctx, false)
	assert.For(t).ThatActual(v2.Outcome).Equals(LegalPass)
	assert.For(t).ThatActual(calls).Equals(1)
}

// TestLegalPlanFieldIndependentMemoMissAcrossProposer proves the memo key
// includes proposer: evaluating the SAME plan/version against a different
// proposer is a fresh miss, not a stale hit.
func TestLegalPlanFieldIndependentMemoMissAcrossProposer(t *testing.T) {
	game := newLegalMemoTestGame(t)
	state := game.CurrentState()

	calls := 0
	plan := countingFieldIndependentPlan("MemoTestMove", &calls)

	plan.evaluate(LegalContext{State: state, Proposer: AdminPlayerIndex}, false)
	assert.For(t).ThatActual(calls).Equals(1)

	plan.evaluate(LegalContext{State: state, Proposer: PlayerIndex(0)}, false)
	assert.For(t).ThatActual(calls).Equals(2)
}

// TestLegalPlanFieldIndependentMemoMissAcrossMoveName proves the memo key
// includes the move type name: two different opted-in move types never
// collide in the same game's memo.
func TestLegalPlanFieldIndependentMemoMissAcrossMoveName(t *testing.T) {
	game := newLegalMemoTestGame(t)
	state := game.CurrentState()

	calls := 0
	planA := countingFieldIndependentPlan("MoveA", &calls)
	planB := countingFieldIndependentPlan("MoveB", &calls)
	ctx := LegalContext{State: state, Proposer: AdminPlayerIndex}

	planA.evaluate(ctx, false)
	assert.For(t).ThatActual(calls).Equals(1)

	planB.evaluate(ctx, false)
	assert.For(t).ThatActual(calls).Equals(2)
}

// TestLegalPlanFieldIndependentMemoEvictsOnVersionAdvance proves the
// "bound memory: keep only the current head version per game" eviction
// rule (design spec §5's honest table): once the game's version advances,
// a query at the OLD version is a fresh miss again (its entry was
// discarded, not merely superseded), and the new head version starts its
// own cache.
func TestLegalPlanFieldIndependentMemoEvictsOnVersionAdvance(t *testing.T) {
	game := newLegalMemoTestGame(t)

	calls := 0
	plan := countingFieldIndependentPlan("MemoTestMove", &calls)

	v0State := game.CurrentState()
	ctxV0 := LegalContext{State: v0State, Proposer: AdminPlayerIndex}

	plan.evaluate(ctxV0, false)
	assert.For(t).ThatActual(calls).Equals(1)
	plan.evaluate(ctxV0, false)
	assert.For(t).ThatActual(calls).Equals(1) // still a hit before advancing

	legalMemoAdvanceVersion(t, game)
	assert.For(t).ThatActual(game.Version()).Equals(v0State.Version() + 1)

	ctxV1 := LegalContext{State: game.CurrentState(), Proposer: AdminPlayerIndex}
	plan.evaluate(ctxV1, false)
	assert.For(t).ThatActual(calls).Equals(2) // new version: fresh miss

	// Bound memory, checked DIRECTLY (not just inferred from a miss): the
	// v0 entry must be gone from the map itself, not merely shadowed by a
	// version guard on read — a broken eviction that updates the recorded
	// head version without actually clearing the map would leak unboundedly
	// while still LOOKING like a miss from the outside. len == 1 proves only
	// the current head version's single entry is resident.
	assert.For(t).ThatActual(len(game.legalFieldIndepMemo)).Equals(1)

	plan.evaluate(ctxV1, false)
	assert.For(t).ThatActual(calls).Equals(2) // new head version now caches too

	// Bound memory: the old (evicted) v0 entry is gone. A hypothetical
	// re-query against v0State (stale but still holds a valid version
	// number) is a miss again, not resurrected from some larger cache.
	plan.evaluate(ctxV0, false)
	assert.For(t).ThatActual(calls).Equals(3)
}

// TestLegalPlanFieldIndependentMemoSkippedWithNoGame proves the memo is
// simply inert (never consulted, never populated, no panic) whenever
// ctx.State has no backing *Game — e.g. GameManager.ExampleState(), or a
// bare LegalContext{} — which is the isolated-test / probe shape every
// other legal_plan_test.go test already exercises.
func TestLegalPlanFieldIndependentMemoSkippedWithNoGame(t *testing.T) {
	manager := newTestGameManger(t)

	calls := 0
	plan := countingFieldIndependentPlan("MemoTestMove", &calls)
	ctx := LegalContext{State: manager.ExampleState(), Proposer: AdminPlayerIndex}

	plan.evaluate(ctx, false)
	assert.For(t).ThatActual(calls).Equals(1)

	// No Game to memoize against: every evaluation is a fresh miss.
	plan.evaluate(ctx, false)
	assert.For(t).ThatActual(calls).Equals(2)
}

// TestLegalTapeMemoHitMissEviction is a direct unit test of Game.LegalTapeMemo
// (legal_memo.go): a hit for the exact same (version, phase) never calls
// compute again; a different version OR a different phase is a miss; and
// advancing to a new (version, phase) evicts whatever was cached before,
// so returning to an old key recomputes rather than resurrecting stale
// data — the same "keep only current head" bound as the field-independent
// memo above.
func TestLegalTapeMemoHitMissEviction(t *testing.T) {
	game := testDefaultGame(t, false)

	calls := 0
	compute := func() []*MoveStorageRecord {
		calls++
		return []*MoveStorageRecord{{Name: "x"}}
	}

	r1 := game.LegalTapeMemo(1, phaseSetUp, compute)
	assert.For(t).ThatActual(calls).Equals(1)
	assert.For(t).ThatActual(len(r1)).Equals(1)

	// Hit: same (version, phase).
	r2 := game.LegalTapeMemo(1, phaseSetUp, compute)
	assert.For(t).ThatActual(calls).Equals(1)
	assert.For(t).ThatActual(len(r2)).Equals(1)

	// Miss: different version.
	game.LegalTapeMemo(2, phaseSetUp, compute)
	assert.For(t).ThatActual(calls).Equals(2)

	// Miss: same version, different phase.
	game.LegalTapeMemo(2, phaseNormal, compute)
	assert.For(t).ThatActual(calls).Equals(3)

	// Eviction: going back to (1, phaseSetUp) recomputes — that entry was
	// discarded when the cache moved on, not resurrected.
	game.LegalTapeMemo(1, phaseSetUp, compute)
	assert.For(t).ThatActual(calls).Equals(4)
}
