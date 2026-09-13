package legal_test

import (
	"reflect"
	"testing"

	"github.com/jkomoros/boardgame"
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
