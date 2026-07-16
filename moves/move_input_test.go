package moves

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
)

type MoveInputProviderA struct{}

func (m *MoveInputProviderA) MoveInputFields() []boardgame.MoveInputField {
	return []boardgame.MoveInputField{{Name: "A", Disposition: boardgame.MoveInputServerDefaulted}}
}

type MoveInputProviderB struct{}

func (m *MoveInputProviderB) MoveInputFields() []boardgame.MoveInputField {
	return []boardgame.MoveInputField{{Name: "B", Disposition: boardgame.MoveInputContextOwned}}
}

type moveMultipleInputProviders struct {
	Default
	MoveInputProviderA
	MoveInputProviderB
	A int
	B bool
}

func (m *moveMultipleInputProviders) Apply(state boardgame.State) error { return nil }

// Go does not promote two methods with the same name. A wrapper or move that
// deliberately embeds multiple providers composes their complete contract.
func (m *moveMultipleInputProviders) MoveInputFields() []boardgame.MoveInputField {
	result := append([]boardgame.MoveInputField(nil), m.MoveInputProviderA.MoveInputFields()...)
	return append(result, m.MoveInputProviderB.MoveInputFields()...)
}

type moveUncomposedInputProviders struct {
	Default
	MoveInputProviderA
	MoveInputProviderB
	A int
	B bool
}

func (m *moveUncomposedInputProviders) Apply(state boardgame.State) error { return nil }

type recursiveInputBehavior struct{ *recursiveInputBehavior }

type moveRecursiveInputBehavior struct {
	Default
	*recursiveInputBehavior
}

func (m *moveRecursiveInputBehavior) Apply(state boardgame.State) error { return nil }

type moveInputWrapperProvider struct{ MoveInputProviderA }

func (m *moveInputWrapperProvider) MoveInputFields() []boardgame.MoveInputField {
	return []boardgame.MoveInputField{{Name: "B", Disposition: boardgame.MoveInputContextOwned}}
}

type moveWrappedInputProvider struct {
	Default
	moveInputWrapperProvider
}

func (m *moveWrappedInputProvider) Apply(state boardgame.State) error { return nil }

type nilMoveInputProvider struct{ value int }

func (m *nilMoveInputProvider) MoveInputFields() []boardgame.MoveInputField {
	_ = m.value
	return nil
}

type moveNilInputProvider struct {
	Default
	*nilMoveInputProvider
}

func (m *moveNilInputProvider) Apply(state boardgame.State) error { return nil }

type MoveInputConflictingProvider struct{}

func (m *MoveInputConflictingProvider) MoveInputFields() []boardgame.MoveInputField {
	return []boardgame.MoveInputField{{Name: "A", Disposition: boardgame.MoveInputRequired}}
}

type moveConflictingInputProviders struct {
	Default
	MoveInputProviderA
	MoveInputConflictingProvider
	A int
}

func (m *moveConflictingInputProviders) Apply(state boardgame.State) error { return nil }

func (m *moveConflictingInputProviders) MoveInputFields() []boardgame.MoveInputField {
	result := append([]boardgame.MoveInputField(nil), m.MoveInputProviderA.MoveInputFields()...)
	return append(result, m.MoveInputConflictingProvider.MoveInputFields()...)
}

func TestCurrentPlayerContributesContextOwnedTarget(t *testing.T) {
	manager, err := newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return []boardgame.MoveConfig{auto.MustConfig(new(moveContribCurrentPlayer), WithMoveName("Input Current Player"))}
	})
	if err != nil {
		t.Fatal(err)
	}
	fields, err := boardgame.ResolveMoveInputFields(manager.ExampleMoveByName("Input Current Player"))
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 1 || fields[0].Name != "TargetPlayerIndex" || fields[0].Disposition != boardgame.MoveInputContextOwned {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}

func TestUnrepresentableInheritedContextFieldIsOmittedWithoutHidingCreatorFields(t *testing.T) {
	contributions := []sourcedMoveInputField{
		{field: contextOwnedTargetPlayerField(), source: "moves.CurrentPlayer"},
		{field: boardgame.MoveInputField{Name: "TargetLocation", Disposition: boardgame.MoveInputRequired}, source: "provider"},
	}
	got := representableMoveInputContributions(contributions, map[string]boardgame.PropertyType{
		"TargetLocation": boardgame.TypeInt,
	})
	if len(got) != 1 || got[0].field.Name != "TargetLocation" {
		t.Fatalf("unexpected representable fields: %#v", got)
	}
	// A provider-owned field is retained so the ordinary unknown-field check
	// still rejects it instead of silently weakening a creator contract.
	providerContext := sourcedMoveInputField{
		field:  contextOwnedTargetPlayerField(),
		source: "custom.MoveInputFields",
	}
	got = representableMoveInputContributions([]sourcedMoveInputField{providerContext}, nil)
	if len(got) != 1 {
		t.Fatalf("provider field was incorrectly hidden: %#v", got)
	}
}

func TestMoveInputConfigurationFailsLoudly(t *testing.T) {
	tests := []struct {
		name    string
		options []CustomConfigurationOption
		want    string
	}{
		{"unknown", []CustomConfigurationOption{WithMoveInputDefault("NotAField")}, "unknown field"},
		{"behavior conflict", []CustomConfigurationOption{WithMoveInputDefault("TargetPlayerIndex")}, "configured by both"},
		{"bad codec", []CustomConfigurationOption{WithMoveInputFieldOverride("TargetPlayerIndex", boardgame.MoveInputRequired, boardgame.MoveInputCodecBoolean)}, "incompatible"},
		{"multiple codecs", []CustomConfigurationOption{WithMoveInputFieldOverride("TargetPlayerIndex", boardgame.MoveInputRequired, boardgame.MoveInputCodecInteger, boardgame.MoveInputCodecPlayerIndex)}, "received 2 codecs"},
		{"unmatched override", []CustomConfigurationOption{WithMoveInputFieldOverride("NotAField", boardgame.MoveInputRequired)}, "did not match"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager, err := newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
				auto := NewAutoConfigurer(manager.Delegate())
				return []boardgame.MoveConfig{auto.MustConfig(new(moveContribNone), WithMoveName("Valid Input"))}
			})
			if err != nil {
				t.Fatal(err)
			}
			auto := NewAutoConfigurer(manager.Delegate())
			options := append([]CustomConfigurationOption{WithMoveName("Bad Input")}, test.options...)
			_, err = auto.Config(new(moveContribCurrentPlayer), options...)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestMoveInputOverrideIsExplicitEscapeHatch(t *testing.T) {
	manager, err := newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return []boardgame.MoveConfig{auto.MustConfig(
			new(moveContribCurrentPlayer),
			WithMoveName("Explicit Target"),
			WithMoveInputFieldOverride("TargetPlayerIndex", boardgame.MoveInputServerDefaulted),
		)}
	})
	if err != nil {
		t.Fatal(err)
	}
	fields, err := boardgame.ResolveMoveInputFields(manager.ExampleMoveByName("Explicit Target"))
	if err != nil {
		t.Fatal(err)
	}
	if fields[0].Disposition != boardgame.MoveInputServerDefaulted {
		t.Fatalf("unexpected override: %#v", fields)
	}
}

func TestMultipleMoveInputProvidersComposeThroughOrdinaryGoMethodSet(t *testing.T) {
	contributions, err := providedMoveInputContributions(new(moveMultipleInputProviders))
	if err != nil {
		t.Fatal(err)
	}
	if len(contributions) != 2 || contributions[0].field.Name != "A" || contributions[1].field.Name != "B" {
		t.Fatalf("embedded contributions were lost: %#v", contributions)
	}
}

func TestUncomposedMoveInputProvidersFailLoudly(t *testing.T) {
	manager, err := newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return []boardgame.MoveConfig{auto.MustConfig(new(moveContribNone), WithMoveName("Valid Provider"))}
	})
	if err != nil {
		t.Fatal(err)
	}
	auto := NewAutoConfigurer(manager.Delegate())
	_, err = auto.Config(new(moveUncomposedInputProviders), WithMoveName("Uncomposed Providers"))
	if err == nil {
		t.Fatal("expected ambiguous providers to fail")
	}
	for _, want := range []string{"Uncomposed Providers", "MoveInputProviderA", "MoveInputProviderB", "complete contract"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q did not contain %q", err, want)
		}
	}
}

func TestEmbeddedMoveInputProviderTypeScanHandlesRecursiveEmbedding(t *testing.T) {
	got := embeddedMoveInputProviderTypes(reflect.TypeOf(new(moveRecursiveInputBehavior)))
	if len(got) != 0 {
		t.Fatalf("recursive non-provider reported providers: %v", got)
	}
}

func TestWrapperMoveInputProviderOwnsItsCompleteContract(t *testing.T) {
	contributions, err := providedMoveInputContributions(new(moveWrappedInputProvider))
	if err != nil {
		t.Fatal(err)
	}
	if len(contributions) != 1 || contributions[0].field.Name != "B" {
		t.Fatalf("wrapper contract was not authoritative: %#v", contributions)
	}
}

func TestNilPointerMoveInputProviderFailsLoudly(t *testing.T) {
	manager, err := newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return []boardgame.MoveConfig{auto.MustConfig(new(moveContribNone), WithMoveName("Valid Provider"))}
	})
	if err != nil {
		t.Fatal(err)
	}
	auto := NewAutoConfigurer(manager.Delegate())
	_, err = auto.Config(new(moveNilInputProvider), WithMoveName("Nil Provider"))
	if err == nil || !strings.Contains(err.Error(), "panicked") {
		t.Fatalf("error = %v, want loud nil-provider diagnostic", err)
	}
}

func TestConflictingEmbeddedMoveInputProvidersFailWithProvenance(t *testing.T) {
	manager, err := newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return []boardgame.MoveConfig{auto.MustConfig(new(moveContribNone), WithMoveName("Valid Provider"))}
	})
	if err != nil {
		t.Fatal(err)
	}
	auto := NewAutoConfigurer(manager.Delegate())
	_, err = auto.Config(new(moveConflictingInputProviders), WithMoveName("Conflicting Providers"))
	if err == nil {
		t.Fatal("expected provider conflict")
	}
	message := err.Error()
	for _, want := range []string{"Conflicting Providers", "moveConflictingInputProviders"} {
		if !strings.Contains(message, want) {
			t.Fatalf("error %q did not contain %q", message, want)
		}
	}
}
