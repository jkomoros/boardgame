package moves

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

func TestWithChoicesInstallsActorExactProjection(t *testing.T) {
	config := boardgame.PropertyCollection{}
	WithChoices("Target")(config)

	declarations, ok := config[configPropMoveChoiceProjections].([]moveChoiceProjectionDeclaration)
	if !ok || len(declarations) != 1 {
		t.Fatalf("choice declarations = %#v", config[configPropMoveChoiceProjections])
	}
	projection := declarations[0].projection
	if projection.FieldName != "Target" || projection.Source != "" || projection.Disclosure != boardgame.MoveChoiceDisclosureActorExact {
		t.Fatalf("projection = %#v", projection)
	}
}

func TestExcludeChoicesIsImmutableAndAddsCanonicalLegality(t *testing.T) {
	values := []string{"Unknown", "Guard"}
	option := ExcludeChoices(values...)
	values[0] = "changed"

	config := boardgame.PropertyCollection{}
	WithChoices("Card", option)(config)
	declarations := config[configPropMoveChoiceProjections].([]moveChoiceProjectionDeclaration)
	projection := declarations[0].projection
	if got := projection.ExcludedValues; len(got) != 2 || got[0] != "Unknown" || got[1] != "Guard" {
		t.Fatalf("excluded values = %v", got)
	}
	if enabled, _ := config[configPropLegalPlanEnabled].(bool); !enabled {
		t.Fatal("excluded choices did not enable canonical legality")
	}
	if specs, _ := config[configPropPreconditions].([]legal.Spec); len(specs) != 2 {
		t.Fatalf("legal preconditions = %#v, want one per exclusion", config[configPropPreconditions])
	}
}
