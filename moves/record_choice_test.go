package moves

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

type moveRecordInheritedTarget struct {
	RecordCurrentPlayerChoice
}

func TestWithRecordedChoiceExpandsToChoiceAndDestination(t *testing.T) {
	config := boardgame.PropertyCollection{}
	WithRecordedChoice(
		"ChosenCard",
		InPlayer("PendingCard"),
		ExcludeChoices("Unknown", "Guard"),
	)(config)

	recorded, ok := config[configPropRecordedChoices].([]recordedChoiceDeclaration)
	if !ok || len(recorded) != 1 {
		t.Fatalf("recorded choice declarations = %#v", config[configPropRecordedChoices])
	}
	if recorded[0].field != "ChosenCard" || recorded[0].target.scope != recordedChoicePlayer || recorded[0].target.property != "PendingCard" {
		t.Fatalf("recorded choice = %#v", recorded[0])
	}
	choices, ok := config[configPropMoveChoiceProjections].([]moveChoiceProjectionDeclaration)
	if !ok || len(choices) != 1 || choices[0].projection.FieldName != "ChosenCard" {
		t.Fatalf("choice projection declarations = %#v", config[configPropMoveChoiceProjections])
	}
	if specs, _ := config[configPropPreconditions].([]legal.Spec); len(specs) != 2 {
		t.Fatalf("exclusion preconditions = %#v", config[configPropPreconditions])
	}
}

func TestRecordedChoiceDestinationsAreSealedValues(t *testing.T) {
	tests := []struct {
		name     string
		target   ChoiceDestination
		scope    recordedChoiceScope
		property string
	}{
		{"player", InPlayer("SelectedPlayer"), recordedChoicePlayer, "SelectedPlayer"},
		{"game", InGame("PendingSuit"), recordedChoiceGame, "PendingSuit"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.target.recordedChoiceDestination()
			if got.scope != test.scope || got.property != test.property {
				t.Fatalf("destination = %#v", got)
			}
		})
	}
}

func TestRecordCurrentPlayerChoiceUsesTopLevelMoveConvention(t *testing.T) {
	manager, err := newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return []boardgame.MoveConfig{auto.MustConfig(
			new(moveRecordInheritedTarget),
			WithMoveName("Record Current Target"),
			WithMoveInputFieldOverride("TargetPlayerIndex", boardgame.MoveInputRequired),
			WithRecordedChoice("TargetPlayerIndex", InGame("CurrentPlayer")),
		)}
	})
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewGame(2, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	actor := game.CurrentState().CurrentPlayerIndex()
	move := game.MoveByName("Record Current Target").(*moveRecordInheritedTarget)
	move.TargetPlayerIndex = actor
	if err := <-game.ProposeMove(move, actor); err != nil {
		t.Fatal(err)
	}
}

func TestRecordCurrentPlayerChoiceRejectsInvalidUsageAtBoot(t *testing.T) {
	manager, err := newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return []boardgame.MoveConfig{auto.MustConfig(new(moveContribNone), WithMoveName("Valid"))}
	})
	if err != nil {
		t.Fatal(err)
	}
	auto := NewAutoConfigurer(manager.Delegate())
	_, err = auto.Config(
		new(moveContribCurrentPlayer),
		WithMoveName("Wrong Record Base"),
		WithMoveInputFieldOverride("TargetPlayerIndex", boardgame.MoveInputRequired),
		WithRecordedChoice("TargetPlayerIndex", InGame("CurrentPlayer")),
	)
	if err == nil || !strings.Contains(err.Error(), "has no effect") {
		t.Fatalf("wrong-base error = %v", err)
	}

	_, err = newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return []boardgame.MoveConfig{auto.MustConfig(
			new(moveRecordInheritedTarget),
			WithMoveName("Mismatched Recorded Choice"),
			WithMoveInputFieldOverride("TargetPlayerIndex", boardgame.MoveInputRequired),
			WithRecordedChoice("TargetPlayerIndex", InGame("Counter")),
		)}
	})
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("type-mismatch error = %v", err)
	}
}
