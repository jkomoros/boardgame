package boardgame

import (
	"strings"
	"testing"

	"github.com/workfit/tester/assert"
)

// legalTestConstructor builds a fixed-verdict predicate constructor named
// name that returns verdict every time it is Evaluated, declaring reads as
// its Reads.
func legalTestConstructor(name string, verdict LegalVerdict, reads ...LegalRead) *LegalPredicateConstructor {
	return &LegalPredicateConstructor{
		Name: name,
		Constructor: func(spec LegalSpec, chest *ComponentChest, resolve func(LegalSpec) (*LegalPredicate, error)) (*LegalPredicate, error) {
			return &LegalPredicate{
				Name:  spec.Name,
				Args:  spec.Args,
				Reads: reads,
				Cost:  LegalCostTrivial,
				Evaluate: func(ctx LegalContext) LegalVerdict {
					return verdict
				},
			}, nil
		},
	}
}

func legalTestRegistry() map[string]*LegalPredicateConstructor {
	return map[string]*LegalPredicateConstructor{
		"alwaysPass":    legalTestConstructor("alwaysPass", LegalVerdict{Outcome: LegalPass}),
		"alwaysFail":    legalTestConstructor("alwaysFail", LegalVerdict{Outcome: LegalFail, Message: &LegalMessage{Template: "legal.test_fail"}}),
		"alwaysUnknown": legalTestConstructor("alwaysUnknown", LegalVerdict{Outcome: LegalUnknown, Reason: "test unknown"}),
		"zeroVerdict":   legalTestConstructor("zeroVerdict", LegalVerdict{}),
	}
}

func TestResolveLegalSpecsRegistryRoundTrip(t *testing.T) {
	registry := legalTestRegistry()

	t.Run("resolve a leaf", func(t *testing.T) {
		preds, err := resolveLegalSpecs([]LegalSpec{{Name: "alwaysPass", Args: []string{"x"}}}, registry, nil, nil, nil)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(len(preds)).Equals(1)
		assert.For(t).ThatActual(preds[0].Name).Equals("alwaysPass")
		assert.For(t).ThatActual(preds[0].Args).Equals([]string{"x"})
		v := evalLegalPredicate(preds[0], LegalContext{})
		assert.For(t).ThatActual(v.Outcome).Equals(LegalPass)
	})

	t.Run("resolve an any of two", func(t *testing.T) {
		spec := LegalSpec{
			Name: "any",
			Sub: []LegalSpec{
				{Name: "alwaysFail"},
				{Name: "alwaysPass"},
			},
		}
		preds, err := resolveLegalSpecs([]LegalSpec{spec}, registry, nil, nil, nil)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(len(preds)).Equals(1)
		assert.For(t).ThatActual(preds[0].Name).Equals("any")
		assert.For(t).ThatActual(len(preds[0].Sub)).Equals(2)
		v := evalLegalPredicate(preds[0], LegalContext{})
		assert.For(t).ThatActual(v.Outcome).Equals(LegalPass)
	})
}

func TestResolveLegalSpecsUnknownName(t *testing.T) {
	registry := legalTestRegistry()
	_, err := resolveLegalSpecs([]LegalSpec{{Name: "totallyNotRegistered"}}, registry, nil, nil, nil)
	assert.For(t).ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(strings.Contains(err.Error(), "totallyNotRegistered")).Equals(true)
}

func TestResolveLegalSpecsAnyDepthTwoRejected(t *testing.T) {
	registry := legalTestRegistry()
	spec := LegalSpec{
		Name: "any",
		Sub: []LegalSpec{
			{Name: "alwaysPass"},
			{
				Name: "any",
				Sub: []LegalSpec{
					{Name: "alwaysFail"},
					{Name: "alwaysUnknown"},
				},
			},
		},
	}
	_, err := resolveLegalSpecs([]LegalSpec{spec}, registry, nil, nil, nil)
	assert.For(t).ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(strings.Contains(err.Error(), "any")).Equals(true)
}

func TestResolveLegalSpecsAnyRequiresTwoSubs(t *testing.T) {
	registry := legalTestRegistry()

	t.Run("zero subs", func(t *testing.T) {
		_, err := resolveLegalSpecs([]LegalSpec{{Name: "any"}}, registry, nil, nil, nil)
		assert.For(t).ThatActual(err).IsNotNil()
	})

	t.Run("one sub", func(t *testing.T) {
		_, err := resolveLegalSpecs([]LegalSpec{{Name: "any", Sub: []LegalSpec{{Name: "alwaysPass"}}}}, registry, nil, nil, nil)
		assert.For(t).ThatActual(err).IsNotNil()
	})

	t.Run("two subs ok", func(t *testing.T) {
		_, err := resolveLegalSpecs([]LegalSpec{{Name: "any", Sub: []LegalSpec{{Name: "alwaysPass"}, {Name: "alwaysFail"}}}}, registry, nil, nil, nil)
		assert.For(t).ThatActual(err).IsNil()
	})
}

func TestLegalAnyKleeneTruthTable(t *testing.T) {
	registry := legalTestRegistry()

	tests := []struct {
		name        string
		names       []string
		wantOutcome LegalOutcome
	}{
		{"pass pass", []string{"alwaysPass", "alwaysPass"}, LegalPass},
		{"pass fail", []string{"alwaysPass", "alwaysFail"}, LegalPass},
		{"fail pass", []string{"alwaysFail", "alwaysPass"}, LegalPass},
		{"pass unknown", []string{"alwaysPass", "alwaysUnknown"}, LegalPass},
		{"unknown pass", []string{"alwaysUnknown", "alwaysPass"}, LegalPass},
		{"fail fail", []string{"alwaysFail", "alwaysFail"}, LegalFail},
		{"fail unknown", []string{"alwaysFail", "alwaysUnknown"}, LegalUnknown},
		{"unknown fail", []string{"alwaysUnknown", "alwaysFail"}, LegalUnknown},
		{"unknown unknown", []string{"alwaysUnknown", "alwaysUnknown"}, LegalUnknown},
		{"fail fail unknown (3-way)", []string{"alwaysFail", "alwaysFail", "alwaysUnknown"}, LegalUnknown},
		{"fail fail pass (3-way)", []string{"alwaysFail", "alwaysFail", "alwaysPass"}, LegalPass},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var subs []LegalSpec
			for _, n := range tc.names {
				subs = append(subs, LegalSpec{Name: n})
			}
			preds, err := resolveLegalSpecs([]LegalSpec{{Name: "any", Sub: subs}}, registry, nil, nil, nil)
			assert.For(t).ThatActual(err).IsNil()
			v := evalLegalPredicate(preds[0], LegalContext{})
			assert.For(t).ThatActual(v.Outcome).Equals(tc.wantOutcome)
		})
	}

	t.Run("all-fail uses default template when no Message override", func(t *testing.T) {
		preds, err := resolveLegalSpecs([]LegalSpec{{Name: "any", Sub: []LegalSpec{{Name: "alwaysFail"}, {Name: "alwaysFail"}}}}, registry, nil, nil, nil)
		assert.For(t).ThatActual(err).IsNil()
		v := evalLegalPredicate(preds[0], LegalContext{})
		assert.For(t).ThatActual(v.Outcome).Equals(LegalFail)
		assert.For(t).ThatActual(v.Message).IsNotNil()
		assert.For(t).ThatActual(v.Message.Template).Equals("legal.any_failed")
	})

	t.Run("all-fail uses spec Message override when set", func(t *testing.T) {
		spec := LegalSpec{Name: "any", Sub: []LegalSpec{{Name: "alwaysFail"}, {Name: "alwaysFail"}}}.WithMessage("custom.failed")
		preds, err := resolveLegalSpecs([]LegalSpec{spec}, registry, nil, nil, nil)
		assert.For(t).ThatActual(err).IsNil()
		v := evalLegalPredicate(preds[0], LegalContext{})
		assert.For(t).ThatActual(v.Message.Template).Equals("custom.failed")
	})
}

func TestLegalPredicateSerializable(t *testing.T) {
	t.Run("nil is not serializable", func(t *testing.T) {
		var p *LegalPredicate
		assert.For(t).ThatActual(p.Serializable()).Equals(false)
	})

	t.Run("plain leaf is serializable", func(t *testing.T) {
		p := &LegalPredicate{Name: "leaf"}
		assert.For(t).ThatActual(p.Serializable()).Equals(true)
	})

	t.Run("opaque leaf is not serializable", func(t *testing.T) {
		p := &LegalPredicate{Name: "leaf", opaque: true}
		assert.For(t).ThatActual(p.Serializable()).Equals(false)
	})

	t.Run("compositor with all-serializable subs is serializable", func(t *testing.T) {
		p := &LegalPredicate{
			Name: "any",
			Sub: []*LegalPredicate{
				{Name: "a"},
				{Name: "b"},
			},
		}
		assert.For(t).ThatActual(p.Serializable()).Equals(true)
	})

	t.Run("compositor with an opaque sub is not serializable", func(t *testing.T) {
		p := &LegalPredicate{
			Name: "any",
			Sub: []*LegalPredicate{
				{Name: "a"},
				{Name: "b", opaque: true},
			},
		}
		assert.For(t).ThatActual(p.Serializable()).Equals(false)
	})
}

func TestEvalLegalPredicateZeroVerdictIsInvalid(t *testing.T) {
	p := &LegalPredicate{
		Name: "zeroVerdict",
		Evaluate: func(ctx LegalContext) LegalVerdict {
			return LegalVerdict{}
		},
	}
	v := evalLegalPredicate(p, LegalContext{})
	assert.For(t).ThatActual(v.Outcome).Equals(LegalUnknown)
	assert.For(t).ThatActual(v.Reason).Equals("predicate returned invalid verdict")
}

func TestEvalLegalPredicateNilPredicate(t *testing.T) {
	v := evalLegalPredicate(nil, LegalContext{})
	assert.For(t).ThatActual(v.Outcome).Equals(LegalUnknown)
}

func TestEvalLegalPredicateNilEvaluate(t *testing.T) {
	p := &LegalPredicate{Name: "noEval"}
	v := evalLegalPredicate(p, LegalContext{})
	assert.For(t).ThatActual(v.Outcome).Equals(LegalUnknown)
	assert.For(t).ThatActual(strings.Contains(v.Reason, "noEval")).Equals(true)
}

func TestEvalLegalPredicatePanicRecovery(t *testing.T) {
	t.Run("generic panic names the predicate", func(t *testing.T) {
		p := &LegalPredicate{
			Name: "boom",
			Evaluate: func(ctx LegalContext) LegalVerdict {
				panic("kaboom")
			},
		}
		v := evalLegalPredicate(p, LegalContext{Move: testMoveForLegalPredicateTests(t)})
		assert.For(t).ThatActual(v.Outcome).Equals(LegalUnknown)
		assert.For(t).ThatActual(strings.Contains(v.Reason, "boom")).Equals(true)
	})

	t.Run("nil-Move panic with no declared move read becomes undeclared move read", func(t *testing.T) {
		p := &LegalPredicate{
			Name: "touchesMoveWithoutDeclaring",
			// No Reads declared: this predicate is field-independent by its
			// own declaration, yet its Evaluate dereferences ctx.Move
			// anyway. With ctx.Move nil, that's a nil-pointer panic; the
			// guard should turn it into a specific, actionable Unknown
			// rather than a generic one.
			Evaluate: func(ctx LegalContext) LegalVerdict {
				_ = ctx.Move.Reader().Props()
				return LegalVerdict{Outcome: LegalPass}
			},
		}
		v := evalLegalPredicate(p, LegalContext{Move: nil})
		assert.For(t).ThatActual(v.Outcome).Equals(LegalUnknown)
		assert.For(t).ThatActual(v.Reason).Equals("undeclared move read")
	})

	t.Run("nil-Move panic WITH a declared move read is a generic panic, not the special guard", func(t *testing.T) {
		p := &LegalPredicate{
			Name:  "declaresMoveButNilAnyway",
			Reads: []LegalRead{{Path: "move.AString", Facet: LegalFacetValues}},
			Evaluate: func(ctx LegalContext) LegalVerdict {
				_ = ctx.Move.Reader().Props()
				return LegalVerdict{Outcome: LegalPass}
			},
		}
		v := evalLegalPredicate(p, LegalContext{Move: nil})
		assert.For(t).ThatActual(v.Outcome).Equals(LegalUnknown)
		if v.Reason == "undeclared move read" {
			t.Errorf("expected a generic panic Reason (predicate declared a move.* read), got the undeclared-move-read guard's Reason instead")
		}
		assert.For(t).ThatActual(strings.Contains(v.Reason, "declaresMoveButNilAnyway")).Equals(true)
	})
}

// testMoveForLegalPredicateTests returns a non-nil Move for tests that only
// need ctx.Move to be non-nil (not a specific move type or state).
func testMoveForLegalPredicateTests(t *testing.T) Move {
	manager := newTestGameManger(t)
	return manager.ExampleMoveByName("Test")
}

func TestResolveLegalSpecsValidatesReadsAtBoot(t *testing.T) {
	manager := newTestGameManger(t)
	exampleState := manager.ExampleState()
	moveReader := manager.ExampleMoveByName("Test").Reader()

	registry := map[string]*LegalPredicateConstructor{
		"goodGameRead": legalTestConstructor("goodGameRead", LegalVerdict{Outcome: LegalPass}, LegalRead{Path: "game.DrawDeck", Facet: LegalFacetValues}),
		"badGameRead":  legalTestConstructor("badGameRead", LegalVerdict{Outcome: LegalPass}, LegalRead{Path: "game.NotARealProp", Facet: LegalFacetValues}),
		"goodMoveRead": legalTestConstructor("goodMoveRead", LegalVerdict{Outcome: LegalPass}, LegalRead{Path: "move.AString", Facet: LegalFacetValues}),
		"badMoveRead":  legalTestConstructor("badMoveRead", LegalVerdict{Outcome: LegalPass}, LegalRead{Path: "move.NotARealProp", Facet: LegalFacetValues}),
	}

	t.Run("good game read passes with exampleState provided", func(t *testing.T) {
		_, err := resolveLegalSpecs([]LegalSpec{{Name: "goodGameRead"}}, registry, nil, exampleState, moveReader)
		assert.For(t).ThatActual(err).IsNil()
	})

	t.Run("bad game read fails boot validation naming the path", func(t *testing.T) {
		_, err := resolveLegalSpecs([]LegalSpec{{Name: "badGameRead"}}, registry, nil, exampleState, moveReader)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(strings.Contains(err.Error(), "NotARealProp")).Equals(true)
	})

	t.Run("good move read passes with moveReader provided", func(t *testing.T) {
		_, err := resolveLegalSpecs([]LegalSpec{{Name: "goodMoveRead"}}, registry, nil, exampleState, moveReader)
		assert.For(t).ThatActual(err).IsNil()
	})

	t.Run("bad move read fails boot validation naming the path", func(t *testing.T) {
		_, err := resolveLegalSpecs([]LegalSpec{{Name: "badMoveRead"}}, registry, nil, exampleState, moveReader)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(strings.Contains(err.Error(), "NotARealProp")).Equals(true)
	})

	t.Run("nil exampleState skips game-read validation instead of erroring", func(t *testing.T) {
		_, err := resolveLegalSpecs([]LegalSpec{{Name: "badGameRead"}}, registry, nil, nil, moveReader)
		assert.For(t).ThatActual(err).IsNil()
	})

	t.Run("nil moveReader skips move-read validation instead of erroring", func(t *testing.T) {
		_, err := resolveLegalSpecs([]LegalSpec{{Name: "badMoveRead"}}, registry, nil, exampleState, nil)
		assert.For(t).ThatActual(err).IsNil()
	})
}

func TestUnionLegalReadsDedupes(t *testing.T) {
	subs := []*LegalPredicate{
		{Reads: []LegalRead{{Path: "game.A", Facet: LegalFacetValues}, {Path: "game.B", Facet: LegalFacetCount}}},
		{Reads: []LegalRead{{Path: "game.A", Facet: LegalFacetValues}, {Path: "game.C", Facet: LegalFacetOrder}}},
	}
	got := unionLegalReads(subs)
	assert.For(t).ThatActual(len(got)).Equals(3)
}

func TestMaxLegalCost(t *testing.T) {
	subs := []*LegalPredicate{
		{Cost: LegalCostCheap},
		{Cost: LegalCostExpensive},
		{Cost: LegalCostTrivial},
	}
	assert.For(t).ThatActual(maxLegalCost(subs)).Equals(LegalCostExpensive)
}
