package boardgame

import (
	"errors"
	"strings"
	"testing"

	"github.com/workfit/tester/assert"
)

// legalPlanTestPred builds a synthetic predicate that always returns outcome
// (a Fail carries a Message whose Template is name, so short-circuit ordering
// is observable), with the given declared reads.
func legalPlanTestPred(name string, outcome LegalOutcome, reads ...LegalRead) *LegalPredicate {
	return &LegalPredicate{
		Name:  name,
		Reads: reads,
		Evaluate: func(ctx LegalContext) LegalVerdict {
			switch outcome {
			case LegalFail:
				return LegalVerdict{Outcome: LegalFail, Message: &LegalMessage{Template: name}}
			case LegalUnknown:
				return LegalVerdict{Outcome: LegalUnknown, Reason: name}
			default:
				return LegalVerdict{Outcome: LegalPass}
			}
		},
	}
}

func moveReadOf(path string) LegalRead {
	return LegalRead{Path: LegalPropPath(path), Facet: LegalFacetValues}
}

func TestLegalPlanEvaluateShortCircuit(t *testing.T) {
	// fieldIndependent [pass, fail("boom"), fail("later")]: short-circuit
	// must stop at the FIRST fail, returning its message, never reaching
	// "later" or the fieldDependent bucket.
	plan := &legalPlan{
		fieldIndependent: []*LegalPredicate{
			legalPlanTestPred("first", LegalPass),
			legalPlanTestPred("boom", LegalFail),
			legalPlanTestPred("later", LegalFail),
		},
		fieldDependent: []*LegalPredicate{
			legalPlanTestPred("fieldDep", LegalFail),
		},
	}

	verdict, entries := plan.evaluate(LegalContext{}, false)
	assert.For(t).ThatActual(verdict.Outcome).Equals(LegalFail)
	assert.For(t).ThatActual(verdict.Message.Template).Equals("boom")
	assert.For(t).ThatActual(len(entries)).Equals(0)
}

func TestLegalPlanEvaluateAllPassIsLegal(t *testing.T) {
	plan := &legalPlan{
		fieldIndependent: []*LegalPredicate{legalPlanTestPred("a", LegalPass)},
		fieldDependent:   []*LegalPredicate{legalPlanTestPred("b", LegalPass)},
	}
	verdict, _ := plan.evaluate(LegalContext{}, false)
	assert.For(t).ThatActual(verdict.Outcome).Equals(LegalPass)
	assert.For(t).ThatActual(verdict.Error()).IsNil()
}

func TestLegalPlanEvaluateFieldOrder(t *testing.T) {
	// A field-INDEPENDENT pass followed by a field-DEPENDENT fail: the
	// field-dependent bucket is evaluated after field-independent, so the
	// overall verdict is the field-dependent failure (this is the ordering
	// that preserves today's proposer-check precedence, since the proposer
	// atom is field-dependent).
	plan := &legalPlan{
		fieldIndependent: []*LegalPredicate{legalPlanTestPred("indep", LegalPass)},
		fieldDependent:   []*LegalPredicate{legalPlanTestPred("dep", LegalFail)},
	}
	verdict, _ := plan.evaluate(LegalContext{}, false)
	assert.For(t).ThatActual(verdict.Outcome).Equals(LegalFail)
	assert.For(t).ThatActual(verdict.Message.Template).Equals("dep")
}

func TestLegalPlanEvaluateFailClosed(t *testing.T) {
	// A predicate that returns the invalid zero-value verdict must be
	// treated as Unknown (not Pass), and Unknown is non-Pass so the plan
	// short-circuits on it — legality is never silently granted.
	plan := &legalPlan{
		fieldIndependent: []*LegalPredicate{
			{Name: "zero", Evaluate: func(ctx LegalContext) LegalVerdict { return LegalVerdict{} }},
			legalPlanTestPred("shouldNotReach", LegalPass),
		},
	}
	verdict, _ := plan.evaluate(LegalContext{}, false)
	assert.For(t).ThatActual(verdict.Outcome).Equals(LegalUnknown)
	// Unknown produces a non-nil error (fail-closed at the Legal() boundary).
	assert.For(t).ThatActual(verdict.Error()).IsNotNil()
}

func TestLegalPlanEvaluateNilPlanFailsClosed(t *testing.T) {
	var plan *legalPlan
	verdict, _ := plan.evaluate(LegalContext{}, false)
	assert.For(t).ThatActual(verdict.Outcome).Equals(LegalUnknown)
}

func TestLegalPlanEvaluateFullLedger(t *testing.T) {
	// Full-ledger mode evaluates EVERYTHING (no short-circuit) and returns a
	// parallel entry per predicate; the overall verdict is still the FIRST
	// non-pass.
	plan := &legalPlan{
		fieldIndependent: []*LegalPredicate{
			legalPlanTestPred("indepPass", LegalPass),
			legalPlanTestPred("indepFail", LegalFail),
		},
		fieldDependent: []*LegalPredicate{
			legalPlanTestPred("depFail", LegalFail, moveReadOf("move.X")),
		},
	}
	verdict, entries := plan.evaluate(LegalContext{}, true)

	assert.For(t).ThatActual(verdict.Outcome).Equals(LegalFail)
	assert.For(t).ThatActual(verdict.Message.Template).Equals("indepFail")
	assert.For(t).ThatActual(len(entries)).Equals(3)

	assert.For(t).ThatActual(entries[0].Name).Equals("indepPass")
	assert.For(t).ThatActual(entries[0].FieldDependent).Equals(false)
	assert.For(t).ThatActual(entries[0].Verdict.Outcome).Equals(LegalPass)

	assert.For(t).ThatActual(entries[1].Name).Equals("indepFail")
	assert.For(t).ThatActual(entries[1].Verdict.Outcome).Equals(LegalFail)

	// The field-dependent predicate is marked FieldDependent and its Reads
	// are carried through for the server's per-viewer evaluability.
	assert.For(t).ThatActual(entries[2].Name).Equals("depFail")
	assert.For(t).ThatActual(entries[2].FieldDependent).Equals(true)
	assert.For(t).ThatActual(entries[2].Serializable).Equals(true)
	assert.For(t).ThatActual(len(entries[2].Reads)).Equals(1)
}

func TestLegalPlanBucketSplit(t *testing.T) {
	manager := newTestGameManger(t)
	// A move NOT implementing CustomLegaler → custom is nil.
	move := manager.ExampleMoveByName("Test")

	predicates := []*LegalPredicate{
		legalPlanTestPred("gameRead", LegalPass, LegalRead{Path: "game.DrawDeck", Facet: LegalFacetValues}),
		legalPlanTestPred("moveRead", LegalPass, moveReadOf("move.CardIndex")),
		legalPlanTestPred("noReads", LegalPass),
		legalPlanTestPred("bothReads", LegalPass, LegalRead{Path: "game.DrawDeck", Facet: LegalFacetValues}, moveReadOf("move.Other")),
	}
	specs := []LegalSpec{{Name: "gameRead"}, {Name: "moveRead"}, {Name: "noReads"}, {Name: "bothReads"}}

	plan := buildLegalPlanFromPredicates("Test", predicates, specs, move)

	// Field-INDEPENDENT: the two with no move.* read, in plan order.
	assert.For(t).ThatActual(len(plan.fieldIndependent)).Equals(2)
	assert.For(t).ThatActual(plan.fieldIndependent[0].Name).Equals("gameRead")
	assert.For(t).ThatActual(plan.fieldIndependent[1].Name).Equals("noReads")

	// Field-DEPENDENT: the two that read a move.* path.
	assert.For(t).ThatActual(len(plan.fieldDependent)).Equals(2)
	assert.For(t).ThatActual(plan.fieldDependent[0].Name).Equals("moveRead")
	assert.For(t).ThatActual(plan.fieldDependent[1].Name).Equals("bothReads")

	assert.For(t).ThatActual(plan.custom == nil).IsTrue()
}

// TestLegalPlanBucketSplitPlayersMoveFieldPath verifies that a predicate
// reading a "players[move.<Field>].<Prop>" path (spec §3) lands in the
// field-dependent bucket, exactly like a bare "move.X" read: it depends on
// the move's own field value, so it must be re-evaluated per-move rather
// than memoized as field-independent.
func TestLegalPlanBucketSplitPlayersMoveFieldPath(t *testing.T) {
	manager := newTestGameManger(t)
	move := manager.ExampleMoveByName("Test")

	predicates := []*LegalPredicate{
		legalPlanTestPred("gameRead", LegalPass, LegalRead{Path: "game.DrawDeck", Facet: LegalFacetValues}),
		legalPlanTestPred("playersMoveFieldRead", LegalPass, moveReadOf("players[move.TargetPlayerIndex].Hand")),
	}
	specs := []LegalSpec{{Name: "gameRead"}, {Name: "playersMoveFieldRead"}}

	plan := buildLegalPlanFromPredicates("Test", predicates, specs, move)

	assert.For(t).ThatActual(len(plan.fieldIndependent)).Equals(1)
	assert.For(t).ThatActual(plan.fieldIndependent[0].Name).Equals("gameRead")

	assert.For(t).ThatActual(len(plan.fieldDependent)).Equals(1)
	assert.For(t).ThatActual(plan.fieldDependent[0].Name).Equals("playersMoveFieldRead")
}

// testCustomLegalerMove is a Move that also implements CustomLegaler, for
// exercising the escape-hatch wrapper and the custom bucket.
type testCustomLegalerMove struct {
	testAlwaysLegalMove
	customErr error
}

func (t *testCustomLegalerMove) LegalCustom(state ImmutableState, proposer PlayerIndex) error {
	return t.customErr
}

func TestLegalPlanCustomBucketAndWrapper(t *testing.T) {
	manager := newTestGameManger(t)

	t.Run("move implementing CustomLegaler gets a custom tail", func(t *testing.T) {
		move := &testCustomLegalerMove{}
		plan := buildLegalPlanFromPredicates("", nil, nil, move)
		assert.For(t).ThatActual(plan.custom).IsNotNil()
		assert.For(t).ThatActual(plan.custom.Name).Equals("custom")
		// The custom wrapper is opaque / unserializable.
		assert.For(t).ThatActual(plan.custom.Serializable()).Equals(false)
	})

	t.Run("wrapper passes when LegalCustom returns nil", func(t *testing.T) {
		move := &testCustomLegalerMove{customErr: nil}
		wrapper := newLegalCustomWrapper()
		v := evalLegalPredicate(wrapper, LegalContext{Move: move, State: manager.ExampleState()})
		assert.For(t).ThatActual(v.Outcome).Equals(LegalPass)
	})

	t.Run("wrapper reuses a structured *LegalError verdict", func(t *testing.T) {
		structured := (LegalVerdict{Outcome: LegalFail, Message: &LegalMessage{Template: "some.key"}}).Error()
		move := &testCustomLegalerMove{customErr: structured}
		wrapper := newLegalCustomWrapper()
		v := evalLegalPredicate(wrapper, LegalContext{Move: move, State: manager.ExampleState()})
		assert.For(t).ThatActual(v.Outcome).Equals(LegalFail)
		assert.For(t).ThatActual(v.Message.Template).Equals("some.key")
	})

	t.Run("wrapper wraps a plain error as a one-off template", func(t *testing.T) {
		move := &testCustomLegalerMove{customErr: errors.New("plain boom")}
		wrapper := newLegalCustomWrapper()
		v := evalLegalPredicate(wrapper, LegalContext{Move: move, State: manager.ExampleState()})
		assert.For(t).ThatActual(v.Outcome).Equals(LegalFail)
		assert.For(t).ThatActual(v.Message.Template).Equals("plain boom")
	})
}

func TestLegalPlanCustomEvaluatedLast(t *testing.T) {
	// The custom tail runs after both buckets: if an earlier predicate
	// fails, custom is never consulted (short-circuit). Prove it by making
	// LegalCustom panic — the plan must short-circuit on the earlier fail and
	// never reach it.
	manager := newTestGameManger(t)
	panicMove := &testCustomLegalerMove{}
	// A LegalCustom that would panic if reached.
	plan := &legalPlan{
		fieldIndependent: []*LegalPredicate{legalPlanTestPred("earlyFail", LegalFail)},
		custom: &LegalPredicate{
			Name:   "custom",
			opaque: true,
			Evaluate: func(ctx LegalContext) LegalVerdict {
				t.Errorf("custom must not be reached when an earlier predicate fails")
				return LegalVerdict{Outcome: LegalPass}
			},
		},
	}
	ctx := LegalContext{Move: panicMove, State: manager.ExampleState()}
	verdict, _ := plan.evaluate(ctx, false)
	assert.For(t).ThatActual(verdict.Message.Template).Equals("earlyFail")
}

func TestAssembleLegalSpecList(t *testing.T) {
	contributed := []LegalSpec{
		{Name: "inPhase"},
		{Name: "proposerIsCurrentPlayer"},
		{Name: "stackConstraints"},
	}
	authored := []LegalSpec{
		{Name: "propAtLeast"},
		{Name: "playerBool"},
	}
	suppressions := []string{"proposerIsCurrentPlayer", "stackConstraints"}

	specs := assembleLegalSpecList(contributed, authored, suppressions)

	// Contributed first (base-first, minus suppressed), then authored in
	// declaration order. Authored atoms are NEVER suppressed.
	names := make([]string, len(specs))
	for i, s := range specs {
		names[i] = s.Name
	}
	assert.For(t).ThatActual(strings.Join(names, ",")).Equals("inPhase,propAtLeast,playerBool")
}

func TestValidateLegalEmittedTemplates(t *testing.T) {
	table := map[string]string{"present.key": "body", "any.key": "b"}

	t.Run("present key passes", func(t *testing.T) {
		preds := []*LegalPredicate{{Name: "p", EmittedTemplates: []string{"present.key"}}}
		assert.For(t).ThatActual(validateLegalEmittedTemplates(preds, table)).IsNil()
	})

	t.Run("missing key errors naming predicate and key", func(t *testing.T) {
		preds := []*LegalPredicate{{Name: "badpred", EmittedTemplates: []string{"missing.key"}}}
		err := validateLegalEmittedTemplates(preds, table)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(strings.Contains(err.Error(), "badpred")).Equals(true)
		assert.For(t).ThatActual(strings.Contains(err.Error(), "missing.key")).Equals(true)
	})

	t.Run("recurses into Sub tree", func(t *testing.T) {
		preds := []*LegalPredicate{{
			Name:             "any",
			EmittedTemplates: []string{"any.key"},
			Sub: []*LegalPredicate{
				{Name: "child", EmittedTemplates: []string{"missing.key"}},
			},
		}}
		err := validateLegalEmittedTemplates(preds, table)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(strings.Contains(err.Error(), "child")).Equals(true)
	})
}

// TestValidateLegalEmittedBindings covers footgun-batch F4's core validation
// semantics directly (the end-to-end boot path is
// moves.TestBootValidatesTemplatePlaceholdersAgainstEmittedBindings): a
// declared-metadata predicate whose resolved template body references an
// unemitted placeholder errors naming predicate, key, and placeholder; nil
// metadata (a game-registered predicate predating the field) skips
// validation entirely; a key missing from the table is
// validateLegalEmittedTemplates's error to report, not this one's; and the
// walk recurses into Sub trees.
func TestValidateLegalEmittedBindings(t *testing.T) {
	table := map[string]string{
		"has.placeholder": "you need {frobs} more",
		"all.bound":       "have {value}, need {min}",
		"no.placeholder":  "just words",
	}

	t.Run("unemitted placeholder errors naming predicate, key, and placeholder", func(t *testing.T) {
		preds := []*LegalPredicate{{
			Name:             "badpred",
			EmittedTemplates: []string{"has.placeholder"},
			EmittedBindings:  map[string][]string{"has.placeholder": {"value", "min"}},
		}}
		err := validateLegalEmittedBindings(preds, table)
		assert.For(t).ThatActual(err).IsNotNil()
		if err == nil {
			return
		}
		for _, want := range []string{"badpred", "has.placeholder", "frobs"} {
			assert.For(t, want).ThatActual(strings.Contains(err.Error(), want)).Equals(true)
		}
	})

	t.Run("fully-bound placeholders pass", func(t *testing.T) {
		preds := []*LegalPredicate{{
			Name:             "goodpred",
			EmittedTemplates: []string{"all.bound"},
			EmittedBindings:  map[string][]string{"all.bound": {"value", "min"}},
		}}
		assert.For(t).ThatActual(validateLegalEmittedBindings(preds, table)).IsNil()
	})

	t.Run("empty declared bindings pass against a placeholder-free body", func(t *testing.T) {
		preds := []*LegalPredicate{{
			Name:             "noBindings",
			EmittedTemplates: []string{"no.placeholder"},
			EmittedBindings:  map[string][]string{"no.placeholder": nil},
		}}
		assert.For(t).ThatActual(validateLegalEmittedBindings(preds, table)).IsNil()
	})

	t.Run("nil metadata skips validation entirely", func(t *testing.T) {
		// Same shape as the failing case above, but EmittedBindings is nil:
		// a game-registered predicate without metadata must keep booting.
		preds := []*LegalPredicate{{
			Name:             "legacyGamePred",
			EmittedTemplates: []string{"has.placeholder"},
		}}
		assert.For(t).ThatActual(validateLegalEmittedBindings(preds, table)).IsNil()
	})

	t.Run("key missing from table is skipped (validateLegalEmittedTemplates reports it)", func(t *testing.T) {
		preds := []*LegalPredicate{{
			Name:             "missingKeyPred",
			EmittedTemplates: []string{"not.in.table"},
			EmittedBindings:  map[string][]string{"not.in.table": nil},
		}}
		assert.For(t).ThatActual(validateLegalEmittedBindings(preds, table)).IsNil()
	})

	t.Run("EmittedBindings key absent from EmittedTemplates is a boot error", func(t *testing.T) {
		// A typo'd EmittedBindings key silently got ZERO validation before
		// this check existed: the per-key loop iterates EmittedTemplates, so
		// a bindings entry whose key doesn't appear there was never looked
		// at, and the key it was MEANT to cover was skipped as
		// metadata-free. Malformed metadata is a boot error, not a silent
		// validation gap.
		preds := []*LegalPredicate{{
			Name:             "typoPred",
			EmittedTemplates: []string{"all.bound"},
			EmittedBindings: map[string][]string{
				"all.bound": {"value", "min"},
				"all.buond": {"value", "min"},
			},
		}}
		err := validateLegalEmittedBindings(preds, table)
		assert.For(t).ThatActual(err).IsNotNil()
		if err == nil {
			return
		}
		for _, want := range []string{"typoPred", "all.buond", "EmittedTemplates"} {
			assert.For(t, want).ThatActual(strings.Contains(err.Error(), want)).Equals(true)
		}
	})

	t.Run("recurses into Sub tree", func(t *testing.T) {
		preds := []*LegalPredicate{{
			Name:             "any",
			EmittedTemplates: []string{"no.placeholder"},
			EmittedBindings:  map[string][]string{"no.placeholder": nil},
			Sub: []*LegalPredicate{
				{
					Name:             "badchild",
					EmittedTemplates: []string{"has.placeholder"},
					EmittedBindings:  map[string][]string{"has.placeholder": nil},
				},
			},
		}}
		err := validateLegalEmittedBindings(preds, table)
		assert.For(t).ThatActual(err).IsNotNil()
		if err == nil {
			return
		}
		assert.For(t).ThatActual(strings.Contains(err.Error(), "badchild")).Equals(true)
		assert.For(t).ThatActual(strings.Contains(err.Error(), "frobs")).Equals(true)
	})
}

func TestLegalProbeActive(t *testing.T) {
	g := &GameManager{}

	// Not probing: returns false, records nothing.
	assert.For(t).ThatActual(g.LegalProbeActive()).Equals(false)
	assert.For(t).ThatActual(g.legalProbeReached).Equals(false)

	// Probing: returns true and records reached.
	g.legalProbing = true
	assert.For(t).ThatActual(g.LegalProbeActive()).Equals(true)
	assert.For(t).ThatActual(g.legalProbeReached).Equals(true)
}

func TestLegalEvaluatePlanNotOptedIn(t *testing.T) {
	// A manager with no plans (or no plan for the named move) reports the
	// move as not-handled, so the frozen chain runs — the pure-sugar seam.
	g := &GameManager{}
	handled, err := g.LegalEvaluatePlan("Anything", nil, nil, 0)
	assert.For(t).ThatActual(handled).Equals(false)
	assert.For(t).ThatActual(err).IsNil()

	g.legalPlans = map[string]*legalPlan{"Other": {}}
	handled, err = g.LegalEvaluatePlan("Anything", nil, nil, 0)
	assert.For(t).ThatActual(handled).Equals(false)
	assert.For(t).ThatActual(err).IsNil()
}

// testRegistryDelegate embeds the standard test delegate and adds the optional
// legal.ConstructorConfigurer / legal.TemplateConfigurer surfaces, to exercise
// buildLegalRegistryAndTemplates's delegate overlay.
type testRegistryDelegate struct {
	testGameDelegate
}

func (d *testRegistryDelegate) ConfigurePredicateConstructors() []*LegalPredicateConstructor {
	return []*LegalPredicateConstructor{
		{Name: "gameSpecificPred"},
	}
}

func (d *testRegistryDelegate) ConfigureLegalTemplates() map[string]string {
	return map[string]string{"game.specific": "a game-specific template"}
}

func TestBuildLegalRegistryAndTemplatesOverlaysDelegate(t *testing.T) {
	// Seed the process-global defaults so we can prove the delegate overlays,
	// not replaces, them.
	RegisterDefaultLegalPredicateConstructors(&LegalPredicateConstructor{Name: "defaultPred"})
	RegisterDefaultLegalTemplates(map[string]string{"default.key": "default body"})

	g := &GameManager{delegate: &testRegistryDelegate{}}
	registry, templates, gameRegistered := g.buildLegalRegistryAndTemplates()

	assert.For(t).ThatActual(registry["defaultPred"]).IsNotNil()
	assert.For(t).ThatActual(registry["gameSpecificPred"]).IsNotNil()
	assert.For(t).ThatActual(templates["default.key"]).Equals("default body")
	assert.For(t).ThatActual(templates["game.specific"]).Equals("a game-specific template")

	// The game-registered set (footgun-batch F3's probe scope) contains
	// exactly the delegate-supplied names that are NOT universal defaults.
	assert.For(t).ThatActual(gameRegistered["gameSpecificPred"]).Equals(true)
	assert.For(t).ThatActual(gameRegistered["defaultPred"]).Equals(false)
}
