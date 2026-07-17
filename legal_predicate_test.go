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
	// A weak assertion here (just "contains any") would pass even if the
	// error didn't actually identify the offending nested spec. Require the
	// stronger "nested" phrasing that names what went wrong.
	assert.For(t).ThatActual(strings.Contains(err.Error(), "nested")).Equals(true)
	assert.For(t).ThatActual(strings.Contains(err.Error(), "any")).Equals(true)
}

// TestResolveLegalSpecsAnyNestingBypass covers Finding 1: the depth-1 "any"
// rule must be enforced at resolution, not by literally inspecting each
// sub-spec's Name — a registered constructor is handed the resolve closure
// and could otherwise plant a depth-2 "any" by resolving one internally and
// returning it as its own predicate, entirely undetected by a check that
// only looks at spec.Sub[i].Name.
func TestResolveLegalSpecsAnyNestingBypass(t *testing.T) {
	t.Run("constructor-mediated nested any is rejected", func(t *testing.T) {
		registry := legalTestRegistry()
		// wrapsAnAny's spec.Name is NOT "any" — a literal-Name check on
		// spec.Sub would never see this coming — but its Constructor calls
		// resolve on an {Name: "any", ...} spec internally and returns that
		// as its own predicate, planting a nested "any" via the back door.
		registry["wrapsAnAny"] = &LegalPredicateConstructor{
			Name: "wrapsAnAny",
			Constructor: func(spec LegalSpec, chest *ComponentChest, resolve func(LegalSpec) (*LegalPredicate, error)) (*LegalPredicate, error) {
				return resolve(LegalSpec{
					Name: "any",
					Sub: []LegalSpec{
						{Name: "alwaysPass"},
						{Name: "alwaysFail"},
					},
				})
			},
		}

		outer := LegalSpec{
			Name: "any",
			Sub: []LegalSpec{
				{Name: "wrapsAnAny"},
				{Name: "alwaysPass"},
			},
		}
		_, err := resolveLegalSpecs([]LegalSpec{outer}, registry, nil, nil, nil)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(strings.Contains(err.Error(), "nested")).Equals(true)
		assert.For(t).ThatActual(strings.Contains(err.Error(), "wrapsAnAny")).Equals(true)
	})

	t.Run("hand-built nested any predicate is caught by the post-walk", func(t *testing.T) {
		registry := legalTestRegistry()
		// This constructor bypasses resolve entirely and hand-assembles a
		// *LegalPredicate whose Sub contains a manually-constructed nested
		// "any" — resolve's insideAny tracking never sees this, since it
		// never routes through resolve. checkNoNestedAny's post-resolution
		// walk is the only thing that can catch it.
		registry["handBuildsNestedAny"] = &LegalPredicateConstructor{
			Name: "handBuildsNestedAny",
			Constructor: func(spec LegalSpec, chest *ComponentChest, resolve func(LegalSpec) (*LegalPredicate, error)) (*LegalPredicate, error) {
				return &LegalPredicate{
					Name: legalAnyCompositorName,
					Sub: []*LegalPredicate{
						{
							Name: legalAnyCompositorName,
							Sub: []*LegalPredicate{
								{Name: "innerA", Evaluate: func(ctx LegalContext) LegalVerdict { return LegalVerdict{Outcome: LegalPass} }},
								{Name: "innerB", Evaluate: func(ctx LegalContext) LegalVerdict { return LegalVerdict{Outcome: LegalFail} }},
							},
						},
						{Name: "outerB", Evaluate: func(ctx LegalContext) LegalVerdict { return LegalVerdict{Outcome: LegalPass} }},
					},
				}, nil
			},
		}

		_, err := resolveLegalSpecs([]LegalSpec{{Name: "handBuildsNestedAny"}}, registry, nil, nil, nil)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(strings.Contains(err.Error(), "nested")).Equals(true)
	})

	t.Run("legitimate single any still resolves", func(t *testing.T) {
		registry := legalTestRegistry()
		spec := LegalSpec{
			Name: "any",
			Sub: []LegalSpec{
				{Name: "alwaysFail"},
				{Name: "alwaysPass"},
			},
		}
		preds, err := resolveLegalSpecs([]LegalSpec{spec}, registry, nil, nil, nil)
		assert.For(t).ThatActual(err).IsNil()
		v := evalLegalPredicate(preds[0], LegalContext{})
		assert.For(t).ThatActual(v.Outcome).Equals(LegalPass)
	})
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

	// Footgun-batch F6: an any(Fail, Unknown) Unknown verdict must carry the
	// same Message the all-Fail case would (LegalVerdict explicitly permits a
	// Message on LegalUnknown), and its Reason must identify WHICH
	// sub-predicate was unknown — not a bare "something was unknown".
	t.Run("fail-unknown Unknown carries default template and names the unknown child", func(t *testing.T) {
		preds, err := resolveLegalSpecs([]LegalSpec{{Name: "any", Sub: []LegalSpec{{Name: "alwaysFail"}, {Name: "alwaysUnknown"}}}}, registry, nil, nil, nil)
		assert.For(t).ThatActual(err).IsNil()
		v := evalLegalPredicate(preds[0], LegalContext{})
		assert.For(t).ThatActual(v.Outcome).Equals(LegalUnknown)
		assert.For(t).ThatActual(v.Message).IsNotNil()
		if v.Message != nil {
			assert.For(t).ThatActual(v.Message.Template).Equals("legal.any_failed")
		}
		assert.For(t).ThatActual(strings.Contains(v.Reason, "alwaysUnknown")).Equals(true)
	})

	t.Run("fail-unknown Unknown carries spec Message override when set", func(t *testing.T) {
		spec := LegalSpec{Name: "any", Sub: []LegalSpec{{Name: "alwaysUnknown"}, {Name: "alwaysFail"}}}.WithMessage("custom.failed")
		preds, err := resolveLegalSpecs([]LegalSpec{spec}, registry, nil, nil, nil)
		assert.For(t).ThatActual(err).IsNil()
		v := evalLegalPredicate(preds[0], LegalContext{})
		assert.For(t).ThatActual(v.Outcome).Equals(LegalUnknown)
		assert.For(t).ThatActual(v.Message).IsNotNil()
		if v.Message != nil {
			assert.For(t).ThatActual(v.Message.Template).Equals("custom.failed")
		}
		assert.For(t).ThatActual(strings.Contains(v.Reason, "alwaysUnknown")).Equals(true)
	})
}

// TestLegalAnyKleeneTruthTableExhaustiveThreeWay covers Finding 3: rather
// than a handful of hand-picked 3-way cases, exercise all 27 combinations of
// three children each independently Pass/Unknown/Fail, computing the
// expected Kleene outcome the same way evalLegalAnyKleene is documented to:
// any Pass -> Pass; else any Unknown -> Unknown; else Fail.
func TestLegalAnyKleeneTruthTableExhaustiveThreeWay(t *testing.T) {
	registry := legalTestRegistry()

	outcomes := []struct {
		ctorName string
		outcome  LegalOutcome
	}{
		{"alwaysPass", LegalPass},
		{"alwaysUnknown", LegalUnknown},
		{"alwaysFail", LegalFail},
	}

	for _, a := range outcomes {
		for _, b := range outcomes {
			for _, c := range outcomes {
				name := a.ctorName + "_" + b.ctorName + "_" + c.ctorName
				t.Run(name, func(t *testing.T) {
					anyPass := a.outcome == LegalPass || b.outcome == LegalPass || c.outcome == LegalPass
					anyUnknown := a.outcome == LegalUnknown || b.outcome == LegalUnknown || c.outcome == LegalUnknown

					var want LegalOutcome
					switch {
					case anyPass:
						want = LegalPass
					case anyUnknown:
						want = LegalUnknown
					default:
						want = LegalFail
					}

					spec := LegalSpec{
						Name: "any",
						Sub: []LegalSpec{
							{Name: a.ctorName},
							{Name: b.ctorName},
							{Name: c.ctorName},
						},
					}
					preds, err := resolveLegalSpecs([]LegalSpec{spec}, registry, nil, nil, nil)
					assert.For(t).ThatActual(err).IsNil()
					v := evalLegalPredicate(preds[0], LegalContext{})
					assert.For(t).ThatActual(v.Outcome).Equals(want)
				})
			}
		}
	}
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

func TestAdminBypassIsExplicitPredicatePolicy(t *testing.T) {
	calls := 0
	pred := &LegalPredicate{
		Name:         "actorEligibility",
		Reads:        []LegalRead{{Path: "proposer.SeatFilled", Facet: LegalFacetValues}},
		AdminPolicy:  LegalAdminBypass,
		UsesProposer: true,
		Evaluate: func(ctx LegalContext) LegalVerdict {
			calls++
			return LegalVerdict{Outcome: LegalFail}
		},
	}

	v := evalLegalPredicate(pred, LegalContext{ProposerPlayerIndex: AdminPlayerIndex})
	assert.For(t).ThatActual(v.Outcome).Equals(LegalPass)
	assert.For(t).ThatActual(calls).Equals(0)

	v = evalLegalPredicate(pred, LegalContext{ProposerPlayerIndex: PlayerIndex(0)})
	assert.For(t).ThatActual(v.Outcome).Equals(LegalFail)
	assert.For(t).ThatActual(calls).Equals(1)
}

func TestAdminBypassRejectsNonProposerReads(t *testing.T) {
	pred := &LegalPredicate{
		Name:         "unsafeBypass",
		Reads:        []LegalRead{{Path: "game.Phase", Facet: LegalFacetValues}},
		AdminPolicy:  LegalAdminBypass,
		UsesProposer: true,
	}
	err := validateLegalPredicateForBoot(pred, nil, nil)
	assert.For(t).ThatActual(err).IsNotNil()
	if err == nil || !strings.Contains(err.Error(), "neither proposer-scoped nor a move field") {
		t.Fatalf("error = %v, want proposer-or-move diagnostic", err)
	}
}

func TestReadTypeContractsMustNameDeclaredReads(t *testing.T) {
	pred := &LegalPredicate{
		Name:              "badMetadata",
		RequiredReadTypes: map[LegalPropPath]PropertyType{"game.Score": TypeInt},
	}
	err := validateLegalPredicateForBoot(pred, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "undeclared read") {
		t.Fatalf("error = %v, want undeclared-read diagnostic", err)
	}
}

func TestUnionLegalReadTypesNarrowsPolymorphicContract(t *testing.T) {
	path := LegalPropPath("game.Score")
	required, allowed, err := unionLegalReadTypes([]*LegalPredicate{
		{AllowedReadTypes: map[LegalPropPath][]PropertyType{path: {TypeInt, TypeBool}}},
		{RequiredReadTypes: map[LegalPropPath]PropertyType{path: TypeInt}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if required[path] != TypeInt || len(allowed) != 0 {
		t.Fatalf("required=%v allowed=%v, want exact TypeInt", required, allowed)
	}
}

func TestAllowedReadTypesRejectsInvalidMetadata(t *testing.T) {
	path := LegalPropPath("game.Score")
	for _, allowed := range [][]PropertyType{{}, {TypeIllegal}, {TypeInt, TypeInt}} {
		pred := &LegalPredicate{
			Name:             "badMetadata",
			Reads:            []LegalRead{{Path: path, Facet: LegalFacetValues}},
			AllowedReadTypes: map[LegalPropPath][]PropertyType{path: allowed},
		}
		if err := validateLegalPredicateForBoot(pred, nil, nil); err == nil {
			t.Errorf("AllowedReadTypes %v unexpectedly passed", allowed)
		}
	}
}
