package moves

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
)

type moveRecordInheritedTarget struct {
	RecordCurrentPlayerChoice
}

type moveRecordWithCustomApply struct {
	RecordCurrentPlayerChoice
}

func (*moveRecordWithCustomApply) Apply(state boardgame.State) error {
	game := state.GameState().(*gameState)
	game.Counter++
	return nil
}

type moveRecordShadowedValidation struct {
	RecordCurrentPlayerChoice
}

// ValidConfiguration deliberately shadows the embedded helper's method. The
// engine must still validate the state effect it owns.
func (*moveRecordShadowedValidation) ValidConfiguration(boardgame.State) error { return nil }

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
	if config[configPropPreconditions] != nil {
		t.Fatalf("recorded-choice exclusions unexpectedly enabled declarative legality: %#v", config[configPropPreconditions])
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

func TestRecordCurrentPlayerChoiceUsesConcreteMoveAffiliation(t *testing.T) {
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
	game, err := manager.NewGame(4, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	game.CurrentState().ImmutableGameState().(*gameState).CurrentPlayer = boardgame.AnyPlayerIndex
	actor := boardgame.PlayerIndex(0)
	move := game.MoveByName("Record Current Target").(*moveRecordInheritedTarget)
	move.TargetPlayerIndex = actor
	if err := <-game.ProposeMove(move, actor); err != nil {
		t.Fatal(err)
	}
	if got := game.CurrentState().ImmutableGameState().(*gameState).CurrentPlayer; got != actor {
		t.Fatalf("recorded current player = %d, want %d", got, actor)
	}
}

func TestRecordedChoiceComposesWithOuterApply(t *testing.T) {
	manager, err := newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return []boardgame.MoveConfig{auto.MustConfig(
			new(moveRecordWithCustomApply),
			WithMoveName("Record With Custom Apply"),
			WithMoveInputFieldOverride("TargetPlayerIndex", boardgame.MoveInputRequired),
			WithRecordedChoice("TargetPlayerIndex", InGame("CurrentPlayer")),
		)}
	})
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewGame(4, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	game.CurrentState().ImmutableGameState().(*gameState).CurrentPlayer = boardgame.AnyPlayerIndex
	actor := boardgame.PlayerIndex(0)
	move := game.MoveByName("Record With Custom Apply").(*moveRecordWithCustomApply)
	move.TargetPlayerIndex = actor
	if err := <-game.ProposeMove(move, actor); err != nil {
		t.Fatal(err)
	}
	state := game.CurrentState().ImmutableGameState().(*gameState)
	if state.CurrentPlayer != actor || state.Counter != 1 {
		t.Fatalf("state after composed record/apply = current %d counter %d, want %d/1", state.CurrentPlayer, state.Counter, actor)
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

func TestRecordedChoiceValidationCannotBeShadowed(t *testing.T) {
	_, err := newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return []boardgame.MoveConfig{auto.MustConfig(
			new(moveRecordShadowedValidation),
			WithMoveName("Shadowed Invalid Recording"),
			WithMoveInputFieldOverride("TargetPlayerIndex", boardgame.MoveInputRequired),
			WithRecordedChoice("TargetPlayerIndex", InGame("Counter")),
		)}
	})
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("shadowed validation error = %v", err)
	}
}

func TestWithRecordedChoiceRejectsAnExistingLowLevelDescriptor(t *testing.T) {
	_, err := newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return []boardgame.MoveConfig{auto.MustConfig(
			new(moveRecordInheritedTarget),
			WithMoveName("Duplicate Record Descriptor"),
			WithMoveInputFieldOverride("TargetPlayerIndex", boardgame.MoveInputRequired),
			func(config boardgame.PropertyCollection) {
				if err := boardgame.SetMoveChoiceRecording(config, boardgame.MoveChoiceRecording{
					FieldName: "TargetPlayerIndex", DestinationScope: boardgame.MoveChoiceRecordingGame,
					DestinationProperty: "Counter",
				}); err != nil {
					t.Fatal(err)
				}
			},
			WithRecordedChoice("TargetPlayerIndex", InGame("CurrentPlayer")),
		)}
	})
	if err == nil || !strings.Contains(err.Error(), "more than one choice recording") {
		t.Fatalf("duplicate descriptor error = %v", err)
	}
}
