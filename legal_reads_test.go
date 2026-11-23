package boardgame

import (
	"strings"
	"testing"

	"github.com/workfit/tester/assert"
)

/*
legal_reads_test.go proves the enforced-reads design (legal_reads.go,
docs/superpowers/specs/2026-11-22-legality-enforced-reads-design.md): an
under-declared read can no longer make the field-independent memo serve a
verdict computed for a different move's fields — including the two shapes the
boot probe cannot see (a read conditional on runtime-only context, and a
concrete type assertion on ctx.Move).
*/

const enforcedReadsMoveName = "EnforcedReadsMove"

// enforcedReadsPlan resolves one game-registered predicate exactly as
// NewGameManager does (constructor registry, derived reads, sealing, boot
// validation), assembles a plan, and runs the boot probe. It returns the plan
// and the probe's error so tests can show what the probe does and does not
// catch, independently of runtime behavior.
func enforcedReadsPlan(t *testing.T, ctor *LegalPredicateConstructor, spec LegalSpec) (*legalPlan, error) {
	t.Helper()
	manager := newTestGameManger(t)
	exampleState := manager.ExampleState()
	exampleMove := manager.ExampleMoveByName("Test")
	registry := map[string]*LegalPredicateConstructor{ctor.Name: ctor}
	preds, err := resolveLegalSpecs([]LegalSpec{spec}, registry, manager.Chest(), exampleState, exampleMove.Reader())
	assert.For(t, "resolve").ThatActual(err).IsNil()
	plan := buildLegalPlanFromPredicates(enforcedReadsMoveName, preds, []LegalSpec{spec}, exampleMove)
	probeErr := legalProbeUndeclaredMoveReads(enforcedReadsMoveName, plan, map[string]bool{ctor.Name: true}, exampleState, exampleMove, manager.Chest())
	return plan, probeErr
}

func enforcedReadsConstructor(name string, reads []LegalRead, eval func(ctx LegalContext) LegalVerdict) *LegalPredicateConstructor {
	return &LegalPredicateConstructor{
		Name: name,
		Constructor: func(spec LegalSpec, _ *ComponentChest, _ func(LegalSpec) (*LegalPredicate, error)) (*LegalPredicate, error) {
			return &LegalPredicate{Name: name, Args: spec.Args, Reads: reads, Evaluate: eval}, nil
		},
	}
}

func testMoveWithScore(t *testing.T, score int) *testMove {
	t.Helper()
	m, ok := newTestGameManger(t).ExampleMoveByName("Test").(*testMove)
	if !ok {
		t.Fatal("example Test move is not a *testMove")
	}
	m.ScoreIncrement = score
	return m
}

// assertNoStaleVerdicts evaluates plan at ONE state version, with ONE
// proposer, for two moves differing only in their fields — through the
// memoized hot path on a real *Game — in both orders. Each memoized verdict
// must equal the verdict the same move gets from a fresh, memo-free
// evaluation. That equality is exactly "the memo can never be stale".
func assertNoStaleVerdicts(t *testing.T, plan *legalPlan, proposer PlayerIndex) []LegalVerdict {
	t.Helper()
	legalMove := testMoveWithScore(t, 1)
	illegalMove := testMoveWithScore(t, 2)

	var verdicts []LegalVerdict
	for _, order := range [][]Move{{legalMove, illegalMove}, {illegalMove, legalMove}} {
		game := newLegalMemoTestGame(t)
		state := game.CurrentState()
		for _, move := range order {
			memoized, _ := plan.evaluate(LegalContext{State: state, Move: move, ProposerPlayerIndex: proposer}, false)
			fresh, _ := plan.evaluate(LegalContext{State: state, Move: move, ProposerPlayerIndex: proposer}, true)
			assert.For(t, "memoized verdict equals fresh verdict").ThatActual(memoized.Outcome).Equals(fresh.Outcome)
			verdicts = append(verdicts, memoized)
		}
	}
	return verdicts
}

func TestEnforcedReadsUnderDeclaredMoveReadCannotGoStale(t *testing.T) {
	// Each predicate below means "legal iff move.ScoreIncrement == 1", reads
	// the move, and declares NO move read — the footgun. Before enforcement,
	// the first move evaluated at a version poisoned the memo for every other
	// move at that version.
	cases := []struct {
		name string
		eval func(ctx LegalContext) LegalVerdict
		// probeCatches records what the boot probe sees, which is no longer
		// what safety depends on.
		probeCatches bool
	}{
		{
			name: "unconditional ResolvePath",
			eval: func(ctx LegalContext) LegalVerdict {
				val, _, err := ctx.ResolvePath("move.ScoreIncrement")
				if err != nil {
					return LegalVerdict{Outcome: LegalUnknown, Reason: err.Error()}
				}
				if val.(int) == 1 {
					return LegalVerdict{Outcome: LegalPass}
				}
				return LegalVerdict{Outcome: LegalFail, Message: &LegalMessage{Template: "nope"}}
			},
			probeCatches: true,
		},
		{
			// The probe runs as ObserverPlayerIndex, so a read gated on a
			// real proposer never happens at boot.
			name: "read conditional on the proposer",
			eval: func(ctx LegalContext) LegalVerdict {
				if ctx.ProposerPlayerIndex < 0 {
					return LegalVerdict{Outcome: LegalPass}
				}
				if ctx.Move.Reader() == nil {
					return LegalVerdict{Outcome: LegalUnknown}
				}
				score, err := ctx.Move.Reader().IntProp("ScoreIncrement")
				if err == nil && score == 1 {
					return LegalVerdict{Outcome: LegalPass}
				}
				return LegalVerdict{Outcome: LegalFail, Message: &LegalMessage{Template: "nope"}}
			},
		},
		{
			// The probe's sentinel is not a *testMove, so the assertion just
			// fails at boot and the predicate passes without reading.
			name: "concrete type assertion",
			eval: func(ctx LegalContext) LegalVerdict {
				m, ok := ctx.Move.(*testMove)
				if !ok {
					return LegalVerdict{Outcome: LegalUnknown, Reason: "not a test move"}
				}
				if m.ScoreIncrement == 1 {
					return LegalVerdict{Outcome: LegalPass}
				}
				return LegalVerdict{Outcome: LegalFail, Message: &LegalMessage{Template: "nope"}}
			},
		},
		{
			// The unchecked form panics on the withheld (nil) move.
			name: "unchecked type assertion",
			eval: func(ctx LegalContext) LegalVerdict {
				if ctx.Move.(*testMove).ScoreIncrement == 1 {
					return LegalVerdict{Outcome: LegalPass}
				}
				return LegalVerdict{Outcome: LegalFail, Message: &LegalMessage{Template: "nope"}}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctor := enforcedReadsConstructor("test.underDeclared", []LegalRead{}, c.eval)
			plan, probeErr := enforcedReadsPlan(t, ctor, LegalSpec{Name: "test.underDeclared"})
			assert.For(t, "probe catches").ThatActual(probeErr != nil).Equals(c.probeCatches)
			assert.For(t, "classified field-independent").ThatActual(len(plan.fieldIndependent)).Equals(1)

			for _, verdict := range assertNoStaleVerdicts(t, plan, 0) {
				// The move is withheld, so no evaluation can reach a
				// field-specific answer: never a (stale) Pass or Fail.
				assert.For(t, "never passes on the move's fields").ThatActual(verdict.Outcome).Equals(LegalUnknown)
			}
		})
	}
}

func TestEnforcedReadsHonestDeclarationIsFieldExact(t *testing.T) {
	// Control: the same check with its read derived from its arg is
	// field-dependent and gives each move its own verdict.
	ctor := enforcedReadsConstructor("test.derived", nil, func(ctx LegalContext) LegalVerdict {
		val, _, err := ctx.ResolvePath(LegalPropPath("move.ScoreIncrement"))
		if err != nil {
			return LegalVerdict{Outcome: LegalUnknown, Reason: err.Error()}
		}
		if val.(int) == 1 {
			return LegalVerdict{Outcome: LegalPass}
		}
		return LegalVerdict{Outcome: LegalFail, Message: &LegalMessage{Template: "nope"}}
	})
	plan, probeErr := enforcedReadsPlan(t, ctor, LegalSpec{Name: "test.derived", Args: []string{"move.ScoreIncrement"}})
	assert.For(t, "probe").ThatActual(probeErr).IsNil()
	assert.For(t, "derived reads").ThatActual(plan.fieldDependent[0].Reads).Equals([]LegalRead{{Path: "move.ScoreIncrement", Facet: LegalFacetValues}})

	verdicts := assertNoStaleVerdicts(t, plan, 0)
	outcomes := []LegalOutcome{verdicts[0].Outcome, verdicts[1].Outcome, verdicts[2].Outcome, verdicts[3].Outcome}
	assert.For(t).ThatActual(outcomes).Equals([]LegalOutcome{LegalPass, LegalFail, LegalFail, LegalPass})
}

func TestEnforcedReadsScopeViolationForcesUnknown(t *testing.T) {
	// A predicate that ignores ResolvePath's refusal and returns Pass anyway
	// still cannot pass.
	pred := &LegalPredicate{
		Name:  "swallowsRefusal",
		Reads: []LegalRead{{Path: "game.MyIntSlice", Facet: LegalFacetValues}},
		Evaluate: func(ctx LegalContext) LegalVerdict {
			_, _, _ = ctx.ResolvePath("game.CurrentPlayer")
			return LegalVerdict{Outcome: LegalPass}
		},
	}
	manager := newTestGameManger(t)
	v := EvaluateLegalPredicate(pred, LegalContext{State: manager.ExampleState()})
	assert.For(t).ThatActual(v.Outcome).Equals(LegalUnknown)
	assert.For(t).ThatActual(strings.Contains(v.Reason, "game.CurrentPlayer")).IsTrue()
	assert.For(t).ThatActual(strings.Contains(v.Reason, "swallowsRefusal")).IsTrue()

	// Calling Evaluate directly is the unenforced escape for low-level tests.
	assert.For(t).ThatActual(pred.Evaluate(LegalContext{State: manager.ExampleState()}).Outcome).Equals(LegalPass)
}

func TestEnforcedReadsWithholdsMoveFromFieldIndependentPredicates(t *testing.T) {
	var sawMove bool
	pred := &LegalPredicate{
		Name:  "stateOnly",
		Reads: []LegalRead{{Path: "game.CurrentPlayer", Facet: LegalFacetValues}},
		Evaluate: func(ctx LegalContext) LegalVerdict {
			sawMove = ctx.Move != nil
			return LegalVerdict{Outcome: LegalPass}
		},
	}
	EvaluateLegalPredicate(pred, LegalContext{Move: testMoveWithScore(t, 1)})
	assert.For(t, "field-independent predicate sees no move").ThatActual(sawMove).IsFalse()

	pred.Reads = append(pred.Reads, LegalRead{Path: "move.ScoreIncrement", Facet: LegalFacetValues})
	EvaluateLegalPredicate(pred, LegalContext{Move: testMoveWithScore(t, 1)})
	assert.For(t, "field-dependent predicate sees the move").ThatActual(sawMove).IsTrue()
}

func TestEnforcedReadsBoardSelectorIsFieldDependent(t *testing.T) {
	// game.Board[move.Field] resolves through the move: before enforcement
	// it was misclassified field-independent and memoized without the field.
	pred := &LegalPredicate{
		Name:     "boardSlot",
		Reads:    []LegalRead{{Path: "game.MyBoard[move.ScoreIncrement]", Facet: LegalFacetOccupancy}},
		Evaluate: func(ctx LegalContext) LegalVerdict { return LegalVerdict{Outcome: LegalPass} },
	}
	plan := buildLegalPlanFromPredicates("", []*LegalPredicate{pred}, nil, nil)
	assert.For(t).ThatActual(len(plan.fieldDependent)).Equals(1)
	assert.For(t).ThatActual(len(plan.fieldIndependent)).Equals(0)
}

func TestLegalReadsFromArgs(t *testing.T) {
	got := legalReadsFromArgs([]string{"move.SpaceIndex", ">=", "3", "Normal", "game.Board[move.Slot]", "players[move.Target].Score", "move.SpaceIndex"})
	assert.For(t).ThatActual(got).Equals([]LegalRead{
		{Path: "move.SpaceIndex", Facet: LegalFacetValues},
		{Path: "game.Board[move.Slot]", Facet: LegalFacetValues},
		{Path: "players[move.Target].Score", Facet: LegalFacetValues},
	})
	assert.For(t, "no path args is decided-empty, not nil").ThatActual(legalReadsFromArgs([]string{"3"}) != nil).IsTrue()

	// An explicit empty Reads is a declaration, not a request to derive.
	pred := &LegalPredicate{Name: "explicit", Args: []string{"move.SpaceIndex"}, Reads: []LegalRead{}}
	sealLegalPredicate(pred, pred.Args)
	assert.For(t).ThatActual(len(pred.Reads)).Equals(0)
}

func TestBootProbeReportsUndeclaredStateRead(t *testing.T) {
	ctor := enforcedReadsConstructor("test.undeclaredState", []LegalRead{{Path: "game.MyIntSlice", Facet: LegalFacetValues}},
		func(ctx LegalContext) LegalVerdict {
			if _, _, err := ctx.ResolvePath("game.CurrentPlayer"); err != nil {
				return LegalVerdict{Outcome: LegalUnknown, Reason: err.Error()}
			}
			return LegalVerdict{Outcome: LegalPass}
		})
	_, probeErr := enforcedReadsPlan(t, ctor, LegalSpec{Name: "test.undeclaredState"})
	assert.For(t).ThatActual(probeErr).IsNotNil()
	if probeErr != nil {
		for _, want := range []string{enforcedReadsMoveName, "test.undeclaredState", "game.CurrentPlayer"} {
			assert.For(t, want).ThatActual(strings.Contains(probeErr.Error(), want)).IsTrue()
		}
	}

	// Field-dependent game predicates are probed with the real example move
	// for out-of-scope resolves too.
	dependent := enforcedReadsConstructor("test.undeclaredSecondField", []LegalRead{{Path: "move.ScoreIncrement", Facet: LegalFacetValues}},
		func(ctx LegalContext) LegalVerdict {
			if _, _, err := ctx.ResolvePath("move.AString"); err != nil {
				return LegalVerdict{Outcome: LegalUnknown, Reason: err.Error()}
			}
			return LegalVerdict{Outcome: LegalPass}
		})
	_, probeErr = enforcedReadsPlan(t, dependent, LegalSpec{Name: "test.undeclaredSecondField"})
	assert.For(t).ThatActual(probeErr).IsNotNil()
	if probeErr != nil {
		assert.For(t).ThatActual(strings.Contains(probeErr.Error(), "move.AString")).IsTrue()
	}
}
