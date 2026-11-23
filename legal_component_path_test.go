package boardgame

import (
	"strings"
	"testing"

	"github.com/workfit/tester/assert"
)

/*
Tests for the component hop of the legal path grammar
(legal_component_path.go). The fixture is the package's own test game:
deck "test" carries static Values (testingComponent: String, Integer) and
DynamicComponentValues (testingComponentDynamic: IntVar, Stack, Enum), so one
deck exercises both value sources. DownSizeStack is sanitize:"all:visible",
DrawDeck is sanitize:"len", and player Hand is untagged (visible).
*/

func TestParseLegalComponentPath(t *testing.T) {
	tests := []struct {
		name          string
		path          LegalPropPath
		wantErr       bool
		wantKind      legalPathKind
		wantProp      string
		wantBoard     string
		wantSelector  legalComponentSelector
		wantComponent string
	}{
		{"first", "game.DrawDeck[first].IntVar", false, pathGame, "DrawDeck", "", legalComponentSelector{kind: legalSelectFirst}, "IntVar"},
		{"occupied ordinal", "game.DrawDeck[occupied:2].Integer", false, pathGame, "DrawDeck", "", legalComponentSelector{kind: legalSelectOrdinal, ordinal: 2}, "Integer"},
		{"move field", "player.Hand[move.ScoreIncrement].Integer", false, pathPlayer, "Hand", "", legalComponentSelector{kind: legalSelectMoveField, moveField: "ScoreIncrement"}, "Integer"},
		{"proposer", "proposer.Hand[first].IntVar", false, pathProposer, "Hand", "", legalComponentSelector{kind: legalSelectFirst}, "IntVar"},
		{"players[move.F]", "players[move.TargetPlayerIndex].Hand[move.ScoreIncrement].Stack", false, pathPlayersMoveField, "Hand", "", legalComponentSelector{kind: legalSelectMoveField, moveField: "ScoreIncrement"}, "Stack"},
		{"board space then component", "game.MyBoard[move.ScoreIncrement][first].IntVar", false, pathGame, "MyBoard", "ScoreIncrement", legalComponentSelector{kind: legalSelectFirst}, "IntVar"},

		{"selector without property", "game.DrawDeck[first]", true, 0, "", "", legalComponentSelector{}, ""},
		{"unknown selector", "game.DrawDeck[last].IntVar", true, 0, "", "", legalComponentSelector{}, ""},
		{"literal slot index", "game.DrawDeck[3].IntVar", true, 0, "", "", legalComponentSelector{}, ""},
		{"negative ordinal", "game.DrawDeck[occupied:-1].IntVar", true, 0, "", "", legalComponentSelector{}, ""},
		{"non-canonical ordinal", "game.DrawDeck[occupied:+1].IntVar", true, 0, "", "", legalComponentSelector{}, ""},
		{"empty move field", "game.DrawDeck[move.].IntVar", true, 0, "", "", legalComponentSelector{}, ""},
		{"second component hop", "game.DrawDeck[first].Stack[first].IntVar", true, 0, "", "", legalComponentSelector{}, ""},
		{"dotted component property", "game.DrawDeck[first].Stack.Len", true, 0, "", "", legalComponentSelector{}, ""},
		{"three brackets", "game.MyBoard[move.A][first][first].X", true, 0, "", "", legalComponentSelector{}, ""},
		{"board bracket must be a move field", "game.MyBoard[first][first].X", true, 0, "", "", legalComponentSelector{}, ""},
		{"missing close bracket", "game.DrawDeck[first.IntVar", true, 0, "", "", legalComponentSelector{}, ""},
		{"move kind cannot select", "move.Thing[first].X", true, 0, "", "", legalComponentSelector{}, ""},
		{"players[*] cannot select", "players[*].Hand[first].X", true, 0, "", "", legalComponentSelector{}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseLegalPath(tc.path)
			if tc.wantErr {
				assert.For(t).ThatActual(err).IsNotNil()
				return
			}
			assert.For(t).ThatActual(err).IsNil()
			assert.For(t).ThatActual(got.kind).Equals(tc.wantKind)
			assert.For(t).ThatActual(got.prop).Equals(tc.wantProp)
			assert.For(t).ThatActual(got.boardIndexField).Equals(tc.wantBoard)
			assert.For(t).ThatActual(got.component).Equals(tc.wantSelector)
			assert.For(t).ThatActual(got.componentProp).Equals(tc.wantComponent)
		})
	}
}

func TestValidateLegalComponentPath(t *testing.T) {
	manager := newTestGameManger(t)
	exampleState := manager.ExampleState()
	moveReader := manager.ExampleMoveByName("Test").Reader()

	valid := []struct {
		path LegalPropPath
		want PropertyType
	}{
		{"game.DrawDeck[first].Integer", TypeInt},
		{"game.DrawDeck[first].String", TypeString},
		{"game.DrawDeck[first].IntVar", TypeInt},
		{"game.DrawDeck[first].Enum", TypeEnum},
		{"game.DrawDeck[first].Stack", TypeStack},
		{"game.DrawDeck[occupied:1].IntVar", TypeInt},
		{"game.DownSizeStack[move.ScoreIncrement].IntVar", TypeInt},
		{"player.Hand[first].Integer", TypeInt},
		{"proposer.Hand[first].IntVar", TypeInt},
		{"players[move.TargetPlayerIndex].Hand[move.ScoreIncrement].Stack", TypeStack},
		{"game.MyBoard[move.ScoreIncrement][first].IntVar", TypeInt},
	}
	for _, tc := range valid {
		t.Run(string(tc.path), func(t *testing.T) {
			assert.For(t).ThatActual(validateLegalPath(tc.path, exampleState, moveReader)).IsNil()
			assert.For(t).ThatActual(validateLegalPathType(tc.path, tc.want, exampleState, moveReader)).IsNil()
		})
	}

	invalid := []struct {
		name    string
		path    LegalPropPath
		mention string
	}{
		{"missing component property", "game.DrawDeck[first].Nope", "not defined in deck \"test\""},
		{"stack hop is not a stack", "game.CurrentPlayer[first].IntVar", "expected TypeStack"},
		{"missing stack", "game.NoSuchStack[first].IntVar", "does not exist"},
		{"missing move field", "game.DrawDeck[move.NoSuchField].IntVar", "does not exist"},
		{"string move field", "game.DrawDeck[move.AString].IntVar", "expected TypeInt or TypeEnum"},
		{"bool move field", "game.DrawDeck[move.ABool].IntVar", "expected TypeInt or TypeEnum"},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			err := validateLegalPath(tc.path, exampleState, moveReader)
			assert.For(t).ThatActual(err).IsNotNil()
			assert.For(t, err.Error()).ThatActual(strings.Contains(err.Error(), tc.mention)).Equals(true)
		})
	}

	t.Run("move selector needs a move reader", func(t *testing.T) {
		assert.For(t).ThatActual(validateLegalPath("game.DrawDeck[move.ScoreIncrement].IntVar", exampleState, nil)).IsNotNil()
	})

	t.Run("type contract sees the component property type", func(t *testing.T) {
		err := validateLegalPathType("game.DrawDeck[first].Stack", TypeInt, exampleState, moveReader)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(strings.Contains(err.Error(), "TypeStack")).Equals(true)
		assert.For(t).ThatActual(validateLegalPathTypes("game.DrawDeck[first].IntVar", []PropertyType{TypeBool, TypeInt}, exampleState, moveReader)).IsNil()
	})
}

// ambiguousComponentValues is a static Values type that deliberately reuses
// the dynamic property name IntVar, and a second static type that lacks
// Integer, to exercise legalDeckComponentProp's error arms.
type ambiguousComponentValues struct {
	c      Component
	IntVar int
}

func (a *ambiguousComponentValues) Reader() PropertyReader         { return getDefaultReader(a) }
func (a *ambiguousComponentValues) ContainingComponent() Component { return a.c }
func (a *ambiguousComponentValues) SetContainingComponent(c Component) {
	a.c = c
}

func TestLegalDeckComponentPropClassification(t *testing.T) {
	manager := newTestGameManger(t)
	exampleState := manager.ExampleState()
	dynamic := exampleState.ImmutableDynamicComponentValues()
	deck := manager.Chest().Deck("test")

	typeOf, source, err := legalDeckComponentProp(deck, dynamic, "Integer")
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(typeOf).Equals(TypeInt)
	assert.For(t).ThatActual(source).Equals(legalComponentStatic)

	typeOf, source, err = legalDeckComponentProp(deck, dynamic, "IntVar")
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(typeOf).Equals(TypeInt)
	assert.For(t).ThatActual(source).Equals(legalComponentDynamic)

	t.Run("static and dynamic both define it", func(t *testing.T) {
		mixed := NewDeck()
		mixed.name = "test"
		mixed.AddComponent(&ambiguousComponentValues{})
		_, _, err := legalDeckComponentProp(mixed, dynamic, "IntVar")
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(strings.Contains(err.Error(), "both")).Equals(true)
	})

	t.Run("only some components define a static property", func(t *testing.T) {
		mixed := NewDeck()
		mixed.name = "partial"
		mixed.AddComponent(&testingComponent{Integer: 1})
		mixed.AddComponent(&ambiguousComponentValues{})
		_, _, err := legalDeckComponentProp(mixed, dynamic, "Integer")
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(strings.Contains(err.Error(), "only some")).Equals(true)
	})
}

// componentPathFixture returns a set-up test game whose DownSizeStack holds
// the "test" deck's first two components in slots 0 and 2 (slot 1 empty),
// with distinct dynamic IntVar values, plus a Test move.
func componentPathFixture(t *testing.T) (*Game, ImmutableState, *testGameState, Move) {
	t.Helper()
	game := testDefaultGame(t, false)
	state := game.CurrentState()
	gameState, _ := concreteStates(state)
	assert.For(t).ThatActual(gameState).IsNotNil()

	first := gameState.DrawDeck.ComponentAt(0)
	assert.For(t).ThatActual(first).IsNotNil()
	assert.For(t).ThatActual(first.MoveTo(gameState.DownSizeStack, 0)).IsNil()
	second := gameState.DrawDeck.ComponentAt(0)
	assert.For(t).ThatActual(second).IsNotNil()
	assert.For(t).ThatActual(second.MoveTo(gameState.DownSizeStack, 2)).IsNil()

	first.DynamicValues().(*testingComponentDynamic).IntVar = 7
	second.DynamicValues().(*testingComponentDynamic).IntVar = 9

	move := game.Manager().ExampleMoveByName("Test")
	return game, state, gameState, move
}

func TestResolveLegalComponentPath(t *testing.T) {
	_, state, gameState, move := componentPathFixture(t)

	t.Run("static value via first", func(t *testing.T) {
		val, propType, err := resolveLegalPath("game.DownSizeStack[first].Integer", state, nil)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(propType).Equals(TypeInt)
		want, _ := gameState.DownSizeStack.ComponentAt(0).Values().Reader().IntProp("Integer")
		assert.For(t).ThatActual(val).Equals(want)
	})

	t.Run("dynamic value via first", func(t *testing.T) {
		val, propType, err := resolveLegalPath("game.DownSizeStack[first].IntVar", state, nil)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(propType).Equals(TypeInt)
		assert.For(t).ThatActual(val).Equals(7)
	})

	t.Run("occupied ordinal skips holes", func(t *testing.T) {
		val, _, err := resolveLegalPath("game.DownSizeStack[occupied:1].IntVar", state, nil)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(val).Equals(9)
		_, _, err = resolveLegalPath("game.DownSizeStack[occupied:2].IntVar", state, nil)
		assert.For(t).ThatActual(err).IsNotNil()
	})

	t.Run("dynamic stack is a stack", func(t *testing.T) {
		val, propType, err := resolveLegalPath("game.DownSizeStack[first].Stack", state, nil)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(propType).Equals(TypeStack)
		_, ok := val.(ImmutableStack)
		assert.For(t).ThatActual(ok).Equals(true)
	})

	setIndex := func(i int) {
		assert.For(t).ThatActual(move.ReadSetter().SetIntProp("ScoreIncrement", i)).IsNil()
	}

	t.Run("move field selects a slot", func(t *testing.T) {
		setIndex(2)
		val, _, err := resolveLegalPath("game.DownSizeStack[move.ScoreIncrement].IntVar", state, move)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(val).Equals(9)
	})

	t.Run("empty slot is an error", func(t *testing.T) {
		setIndex(1)
		_, _, err := resolveLegalPath("game.DownSizeStack[move.ScoreIncrement].IntVar", state, move)
		assert.For(t).ThatActual(err).IsNotNil()
	})

	t.Run("out of range is an error, not a panic", func(t *testing.T) {
		for _, i := range []int{-1, 1000} {
			setIndex(i)
			_, _, err := resolveLegalPath("game.DownSizeStack[move.ScoreIncrement].IntVar", state, move)
			assert.For(t, i).ThatActual(err).IsNotNil()
		}
	})

	t.Run("nil move is an error", func(t *testing.T) {
		_, _, err := resolveLegalPath("game.DownSizeStack[move.ScoreIncrement].IntVar", state, nil)
		assert.For(t).ThatActual(err).IsNotNil()
	})

	t.Run("empty stack is an error", func(t *testing.T) {
		_, _, err := resolveLegalPath("game.OtherStack[first].IntVar", state, nil)
		assert.For(t).ThatActual(err).IsNotNil()
	})
}

func TestLegalComponentPathFieldDependence(t *testing.T) {
	tests := []struct {
		path LegalPropPath
		want bool
	}{
		{"game.DownSizeStack[first].IntVar", false},
		{"game.DownSizeStack[occupied:1].Integer", false},
		{"player.Hand[first].IntVar", false},
		{"game.DownSizeStack[move.ScoreIncrement].IntVar", true},
		{"players[move.TargetPlayerIndex].Hand[first].IntVar", true},
		// The board index reads the move even with a state-only component
		// selector, and on its own (previously misclassified).
		{"game.MyBoard[move.ScoreIncrement][first].IntVar", true},
		{"game.MyBoard[move.ScoreIncrement]", true},
	}
	for _, tc := range tests {
		got := legalReadsIncludeMovePath([]LegalRead{{Path: tc.path, Facet: LegalFacetValues}})
		assert.For(t, string(tc.path)).ThatActual(got).Equals(tc.want)
	}
}

func TestLegalComponentReadEvaluable(t *testing.T) {
	read := func(path LegalPropPath, facet LegalFacet) LegalRead {
		return LegalRead{Path: path, Facet: facet}
	}

	t.Run("default policies", func(t *testing.T) {
		_, state, _, _ := componentPathFixture(t)
		// DownSizeStack is all:visible and the test deck's dynamic values
		// are untagged (visible): both static and dynamic reads survive.
		assert.For(t).ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read("game.DownSizeStack[first].Integer", LegalFacetValues))).Equals(true)
		assert.For(t).ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read("game.DownSizeStack[first].IntVar", LegalFacetValues))).Equals(true)
		// DrawDeck is sanitize:"len": the stack hop needs FacetValues, so
		// neither a static nor a dynamic value read survives.
		assert.For(t).ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read("game.DrawDeck[first].Integer", LegalFacetValues))).Equals(false)
		assert.For(t).ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read("game.DrawDeck[first].IntVar", LegalFacetCount))).Equals(false)
		// Admin is omniscient.
		assert.For(t).ThatActual(LegalReadEvaluable(state, AdminPlayerIndex, read("game.DrawDeck[first].IntVar", LegalFacetValues))).Equals(true)
	})

	t.Run("dynamic value policies", func(t *testing.T) {
		game, state, _, _ := componentPathFixture(t)
		policy := &sanitizationTestConfig{
			Game: map[string]string{"DownSizeStack": "visible"},
			DynamicComponentValues: map[string]map[string]string{
				"test": {"IntVar": "hidden", "Stack": "len"},
			},
		}
		policy.Install(game.Manager())

		// A hidden dynamic property is not evaluable even though the
		// component (and so its static values) is visible.
		assert.For(t).ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read("game.DownSizeStack[first].IntVar", LegalFacetValues))).Equals(false)
		assert.For(t).ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read("game.DownSizeStack[first].Integer", LegalFacetValues))).Equals(true)
		// The read's own facet is judged against the dynamic policy: a
		// dynamic stack under PolicyLen keeps its count, not its values.
		assert.For(t).ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read("game.DownSizeStack[first].Stack", LegalFacetCount))).Equals(true)
		assert.For(t).ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read("game.DownSizeStack[first].Stack", LegalFacetNonEmpty))).Equals(true)
		assert.For(t).ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read("game.DownSizeStack[first].Stack", LegalFacetValues))).Equals(false)
		assert.For(t).ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read("game.DownSizeStack[first].Stack", LegalFacetOccupancy))).Equals(false)
	})

	t.Run("stack hop policy dominates", func(t *testing.T) {
		game, state, _, _ := componentPathFixture(t)
		policy := &sanitizationTestConfig{
			Game: map[string]string{"DownSizeStack": "order"},
		}
		policy.Install(game.Manager())
		// PolicyOrder reveals identities but not values: no component
		// value is evaluable, static or dynamic, whatever the facet.
		for _, facet := range []LegalFacet{LegalFacetValues, LegalFacetCount, LegalFacetNonEmpty} {
			assert.For(t, facet).ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read("game.DownSizeStack[first].Integer", facet))).Equals(false)
			assert.For(t, facet).ThatActual(LegalReadEvaluable(state, ObserverPlayerIndex, read("game.DownSizeStack[first].Stack", facet))).Equals(false)
		}
	})

	t.Run("player-owned stacks", func(t *testing.T) {
		game := testDefaultGame(t, false)
		state := game.CurrentState()
		gameState, players := concreteStates(state)
		gameState.CurrentPlayer = 0
		assert.For(t).ThatActual(gameState.DrawDeck.ComponentAt(0).MoveTo(players[0].Hand, 0)).IsNil()
		policy := &sanitizationTestConfig{
			Player: map[string]string{"Hand": "other:hidden"},
		}
		policy.Install(game.Manager())

		// player.X judges the current player's policy for this viewer.
		assert.For(t, "self").ThatActual(LegalReadEvaluable(state, 0, read("player.Hand[first].IntVar", LegalFacetValues))).Equals(true)
		assert.For(t, "other").ThatActual(LegalReadEvaluable(state, 1, read("player.Hand[first].IntVar", LegalFacetValues))).Equals(false)
		// proposer.X and players[move.F] are conservative over every
		// player: some player's Hand is hidden from every concrete viewer.
		assert.For(t, "proposer").ThatActual(LegalReadEvaluable(state, 0, read("proposer.Hand[first].IntVar", LegalFacetValues))).Equals(false)
		assert.For(t, "move player").ThatActual(LegalReadEvaluable(state, 0, read("players[move.TargetPlayerIndex].Hand[move.ScoreIncrement].IntVar", LegalFacetValues))).Equals(false)
	})
}

// TestLegalComponentPathBootContracts pins that a predicate's ordinary
// RequiredReadTypes/AllowedReadTypes contract sees through the component
// hop: an int relation over a dynamic stack, or over a missing component
// property, is a boot error rather than a runtime Unknown.
func TestLegalComponentPathBootContracts(t *testing.T) {
	manager := newTestGameManger(t)
	exampleState := manager.ExampleState()
	moveReader := manager.ExampleMoveByName("Test").Reader()

	intRelation := func(path LegalPropPath) *LegalPredicate {
		return &LegalPredicate{
			Name:              "intRelation",
			Reads:             []LegalRead{{Path: path, Facet: LegalFacetValues}},
			RequiredReadTypes: map[LegalPropPath]PropertyType{path: TypeInt},
			Evaluate:          func(LegalContext) LegalVerdict { return LegalVerdict{Outcome: LegalPass} },
		}
	}

	assert.For(t).ThatActual(validateLegalPredicateForBoot(intRelation("game.DrawDeck[first].IntVar"), exampleState, moveReader)).IsNil()
	assert.For(t).ThatActual(validateLegalPredicateForBoot(intRelation("game.DrawDeck[move.ScoreIncrement].Integer"), exampleState, moveReader)).IsNil()
	assert.For(t).ThatActual(validateLegalPredicateForBoot(intRelation("game.DrawDeck[first].Stack"), exampleState, moveReader)).IsNotNil()
	assert.For(t).ThatActual(validateLegalPredicateForBoot(intRelation("game.DrawDeck[first].Missing"), exampleState, moveReader)).IsNotNil()
	assert.For(t).ThatActual(validateLegalPredicateForBoot(intRelation("game.DrawDeck[move.AString].IntVar"), exampleState, moveReader)).IsNotNil()

	t.Run("component field requirements stay one hop", func(t *testing.T) {
		pred := &LegalPredicate{
			Name:  "nested",
			Reads: []LegalRead{{Path: "game.DrawDeck[first].Stack", Facet: LegalFacetValues}},
			RequiredComponentFields: []LegalComponentFieldRequirement{{
				StackPath:    "game.DrawDeck[first].Stack",
				Field:        "Integer",
				AllowedTypes: []PropertyType{TypeInt},
			}},
			Evaluate: func(LegalContext) LegalVerdict { return LegalVerdict{Outcome: LegalPass} },
		}
		err := validateLegalPredicateForBoot(pred, exampleState, moveReader)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(strings.Contains(err.Error(), "one component hop")).Equals(true)
	})
}
