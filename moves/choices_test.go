package moves

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
)

type stackDomainMove struct{ MoveOnGraph }

func (*stackDomainMove) PlayerLocationBehavior(player boardgame.ImmutableSubState) *behaviors.LocationBehavior {
	return &player.(*playerState).LocationBehavior
}

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

func TestStackChoiceSourcesAreSealedAndDefensivelyCopied(t *testing.T) {
	tests := []struct {
		name     string
		option   ChoiceOption
		scope    boardgame.MoveChoiceStackScope
		property string
	}{
		{"current player", FromCurrentPlayerStack("Hand"), boardgame.MoveChoiceStackScopeActorPlayer, "Hand"},
		{"game", FromGameStack("Market"), boardgame.MoveChoiceStackScopeGame, "Market"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := boardgame.PropertyCollection{}
			WithChoices("Slot", test.option)(config)
			declaration := config[configPropMoveChoiceProjections].([]moveChoiceProjectionDeclaration)[0]
			if declaration.err != "" || declaration.projection.StackSource == nil {
				t.Fatalf("declaration = %#v", declaration)
			}
			got := declaration.projection.StackSource
			if got.Scope != test.scope || got.Property != test.property {
				t.Fatalf("stack source = %#v", got)
			}
			got.Property = "Changed"
			second := boardgame.PropertyCollection{}
			WithChoices("Slot", test.option)(second)
			if source := second[configPropMoveChoiceProjections].([]moveChoiceProjectionDeclaration)[0].projection.StackSource; source.Property != test.property {
				t.Fatalf("choice option aliased prior projection: %#v", source)
			}
		})
	}
}

func TestWithChoicesRejectsMultipleStackSources(t *testing.T) {
	config := boardgame.PropertyCollection{}
	WithChoices("Slot", FromCurrentPlayerStack("Hand"), FromGameStack("Market"))(config)
	declaration := config[configPropMoveChoiceProjections].([]moveChoiceProjectionDeclaration)[0]
	if declaration.err == "" {
		t.Fatalf("multiple sources were silently accepted: %#v", declaration)
	}
}

func TestStackChoiceDomainRejectsForgedEmptySlotBeforeLegal(t *testing.T) {
	manager, err := newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return []boardgame.MoveConfig{auto.MustConfig(
			new(stackDomainMove),
			WithMoveName("Choose Token Slot"),
			WithChoices("TargetLocation", FromCurrentPlayerStack("TokenLocation")),
		)}
	})
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	move := game.MoveByName("Choose Token Slot").(*stackDomainMove)
	move.TargetLocation = 1 // TokenLocation is sized to four; only slot zero is occupied.
	err = <-game.ProposeMove(move, 0)
	if err == nil || !strings.Contains(err.Error(), "not an occupied slot") {
		t.Fatalf("forged empty-slot proposal error = %v", err)
	}
}

func TestStackChoiceLocatorFailsManagerBoot(t *testing.T) {
	_, err := newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return []boardgame.MoveConfig{auto.MustConfig(
			new(stackDomainMove),
			WithMoveName("Invalid Stack Choice"),
			WithChoices("TargetLocation", FromCurrentPlayerStack("Counter")),
		)}
	})
	if err == nil || !strings.Contains(err.Error(), "not a stack") {
		t.Fatalf("invalid stack locator manager error = %v", err)
	}
}
