package legal_test

import (
	"reflect"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/legal"
)

func TestComponentSelectorWireShapes(t *testing.T) {
	tests := []struct {
		name string
		spec legal.Spec
		want []string
	}{
		{"first", legal.ComponentPropEquals(legal.FirstOccupied("game.Cards"), "Type", "Guard"), []string{"game.Cards", "first-occupied", "Type", "Guard"}},
		{"ordinal", legal.ComponentPropEquals(legal.OccupiedOrdinal("game.Cards", 2), "Type", "Guard"), []string{"game.Cards", "occupied-ordinal:2", "Type", "Guard"}},
		{"move index", legal.ComponentPropEquals(legal.MoveIndex("game.Cards", "move.CardIndex"), "Type", "Guard"), []string{"game.Cards", "move-index:move.CardIndex", "Type", "Guard"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !reflect.DeepEqual(test.spec.Args, test.want) {
				t.Fatalf("Args = %v, want %v", test.spec.Args, test.want)
			}
		})
	}
}

func TestComponentScalarMetadataAndSparseSelectors(t *testing.T) {
	spec := legal.ComponentPropsEqual(
		legal.FirstOccupied("game.VisibleCards"),
		legal.OccupiedOrdinal("game.VisibleCards", 1),
		"Type",
	)
	pred := resolvePredicateForTest(t, spec)
	if pred.ClientEvaluable {
		t.Fatal("component scalar predicate claimed client evaluability without a TypeScript evaluator")
	}
	if len(pred.Reads) != 1 || pred.Reads[0] != (boardgame.LegalRead{Path: "game.VisibleCards", Facet: boardgame.LegalFacetValues}) {
		t.Fatalf("Reads = %+v, want one Values read", pred.Reads)
	}
	if pred.RequiredReadTypes["game.VisibleCards"] != boardgame.TypeStack {
		t.Fatalf("RequiredReadTypes = %+v", pred.RequiredReadTypes)
	}
	if len(pred.RequiredComponentFields) != 1 || pred.RequiredComponentFields[0].Field != "Type" {
		t.Fatalf("RequiredComponentFields = %+v", pred.RequiredComponentFields)
	}
	if got := pred.Evaluate(buildLegalFixture(t, "memoryVisibleMatchingSparse").context(0)); got.Outcome != legal.Pass {
		t.Fatalf("matching sparse pair = %+v, want Pass", got)
	}
	if got := pred.Evaluate(buildLegalFixture(t, "memoryVisibleDifferentSparse").context(0)); got.Outcome != legal.Fail {
		t.Fatalf("different sparse pair = %+v, want Fail", got)
	}
}

func TestComponentMoveIndexMetadata(t *testing.T) {
	pred := resolvePredicateForTest(t, legal.ComponentPropEquals(
		legal.MoveIndex("game.HiddenCards", "move.CardIndex"), "Type", "unused",
	))
	if len(pred.Reads) != 2 || pred.Reads[0].Facet != boardgame.LegalFacetValues || pred.Reads[1].Path != "move.CardIndex" {
		t.Fatalf("Reads = %+v", pred.Reads)
	}
	if pred.RequiredReadTypes["move.CardIndex"] != boardgame.TypeInt {
		t.Fatalf("MoveIndex type contract = %+v", pred.RequiredReadTypes)
	}
}

func TestComponentSelectorsRejectMalformedSpecs(t *testing.T) {
	bad := []legal.Spec{
		legal.ComponentPropEquals(legal.OccupiedOrdinal("game.VisibleCards", -1), "Type", "x"),
		legal.ComponentPropEquals(legal.MoveIndex("game.VisibleCards", "player.CardIndex"), "Type", "x"),
		{Name: "componentPropEquals", Args: []string{"game.VisibleCards", "last-occupied", "Type", "x"}},
	}
	for _, spec := range bad {
		if _, err := resolveSpecViaRegistry(spec, legal.DefaultConstructors(), nil); err == nil {
			t.Errorf("malformed spec unexpectedly constructed: %+v", spec)
		}
	}
}

func TestComponentSelectorProp(t *testing.T) {
	tests := []struct {
		name     string
		selector legal.ComponentSelector
		want     string
	}{
		{"first", legal.FirstOccupied("game.Cards"), "game.Cards[first].Type"},
		{"ordinal", legal.OccupiedOrdinal("game.Cards", 2), "game.Cards[occupied:2].Type"},
		{"move index", legal.MoveIndex("players[move.TargetPlayerIndex].Species", "move.TargetSpecies"), "players[move.TargetPlayerIndex].Species[move.TargetSpecies].Type"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.selector.Prop("Type"); got != test.want {
				t.Fatalf("Prop = %q, want %q", got, test.want)
			}
		})
	}
}

// TestComponentPathMatchesComponentPropEquals pins that PropEquals over a
// selector's component path is the same check as the older
// ComponentPropEquals for a static enum field, on both the pass and the
// fail branch.
func TestComponentPathMatchesComponentPropEquals(t *testing.T) {
	fixture := buildLegalFixture(t, "checkersDefault")
	ctx := fixture.context(0)
	selector := legal.FirstOccupied("game.Spaces")
	path := selector.Prop("Color")

	val, propType, err := ctx.ResolvePath(legal.PropPath(path))
	if err != nil || propType != boardgame.TypeEnum {
		t.Fatalf("resolving %q: %v (type %v)", path, err, propType)
	}
	actual := val.(enum.ImmutableVal)
	other := ""
	for _, key := range actual.Enum().Values() {
		if key != actual.Value() {
			other = actual.Enum().String(key)
			break
		}
	}
	if other == "" {
		t.Fatal("color enum has only one value")
	}

	for _, want := range []string{actual.String(), other} {
		viaPath := resolvePredicateForTest(t, legal.PropEquals(path, want)).Evaluate(ctx)
		viaComponent := resolvePredicateForTest(t, legal.ComponentPropEquals(selector, "Color", want)).Evaluate(ctx)
		if viaPath.Outcome == legal.Unknown || viaPath.Outcome != viaComponent.Outcome {
			t.Fatalf("want %q: PropEquals outcome %v, ComponentPropEquals outcome %v", want, viaPath.Outcome, viaComponent.Outcome)
		}
	}
}
