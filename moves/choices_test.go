package moves

import (
	"testing"

	"github.com/jkomoros/boardgame"
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

func TestExcludeChoicesIsImmutableAndDoesNotEnableDeclarativeLegality(t *testing.T) {
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
	if config[configPropLegalPlanEnabled] != nil || config[configPropPreconditions] != nil {
		t.Fatalf("choice-domain exclusions unexpectedly enabled declarative legality: %#v", config)
	}
}
