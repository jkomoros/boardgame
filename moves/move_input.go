package moves

import (
	"fmt"
	"reflect"

	"github.com/jkomoros/boardgame"
)

const configPropMoveInputDeclarations = fullyQualifiedPackageName + "MoveInputDeclarations"
const configPropMoveChoiceProjections = fullyQualifiedPackageName + "MoveChoiceProjections"

type moveInputDeclaration struct {
	field    boardgame.MoveInputField
	override bool
	err      string
}

type moveChoiceProjectionDeclaration struct {
	projection boardgame.MoveChoiceProjection
	err        string
}

type currentPlayerMoveInputBehavior interface{ moveInputCurrentPlayerBehavior() }
type anyPlayerMoveInputBehavior interface{ moveInputAnyPlayerBehavior() }

type sourcedMoveInputField struct {
	field  boardgame.MoveInputField
	source string
}

// WithMoveInputField explicitly configures one creator-facing move field.
// Supported unclassified fields are required, and standard embedded behaviors
// declare fields they own automatically. Repeating a field is an error.
func WithMoveInputField(name string, disposition boardgame.MoveInputDisposition, codec ...boardgame.MoveInputCodec) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		declarations, _ := config[configPropMoveInputDeclarations].([]moveInputDeclaration)
		field := boardgame.MoveInputField{Name: name, Disposition: disposition}
		declaration := moveInputDeclaration{field: field}
		if len(codec) > 1 {
			declaration.err = fmt.Sprintf("creator-input field %q received %d codecs; provide at most one", name, len(codec))
		}
		if len(codec) > 0 {
			declaration.field.Codec = codec[0]
		}
		config[configPropMoveInputDeclarations] = append(declarations, declaration)
	}
}

// WithMoveInputFieldOverride deliberately replaces a standard behavior's
// declaration. The loud name is intentional: changing ownership of a behavior
// field (for example allowing an explicit TargetPlayerIndex) is an advanced
// escape hatch and should stand out in configuration review.
func WithMoveInputFieldOverride(name string, disposition boardgame.MoveInputDisposition, codec ...boardgame.MoveInputCodec) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		declarations, _ := config[configPropMoveInputDeclarations].([]moveInputDeclaration)
		field := boardgame.MoveInputField{Name: name, Disposition: disposition}
		declaration := moveInputDeclaration{field: field, override: true}
		if len(codec) > 1 {
			declaration.err = fmt.Sprintf("creator-input field override %q received %d codecs; provide at most one", name, len(codec))
		}
		if len(codec) > 0 {
			declaration.field.Codec = codec[0]
		}
		config[configPropMoveInputDeclarations] = append(declarations, declaration)
	}
}

// WithMoveInputDefault marks a field as server-defaulted and overrideable.
func WithMoveInputDefault(name string, codec ...boardgame.MoveInputCodec) CustomConfigurationOption {
	return WithMoveInputField(name, boardgame.MoveInputServerDefaulted, codec...)
}

func collectMoveInputFields(move AutoConfigurableMove, config boardgame.PropertyCollection) error {
	var contributions []sourcedMoveInputField
	if _, ok := move.(currentPlayerMoveInputBehavior); ok {
		contributions = append(contributions, sourcedMoveInputField{contextOwnedTargetPlayerField(), "moves.CurrentPlayer"})
	} else if _, ok := move.(anyPlayerMoveInputBehavior); ok {
		contributions = append(contributions, sourcedMoveInputField{contextOwnedTargetPlayerField(), "moves.AnyPlayer"})
	}
	provided, err := providedMoveInputContributions(move)
	if err != nil {
		return err
	}
	contributions = append(contributions, provided...)
	props := move.ReadSetter().Props()
	contributions = representableMoveInputContributions(contributions, props)
	explicit, _ := config[configPropMoveInputDeclarations].([]moveInputDeclaration)

	seen := make(map[string]string, len(contributions)+len(explicit))
	collected := make([]boardgame.MoveInputField, 0, len(contributions)+len(explicit))
	for _, contribution := range contributions {
		field := contribution.field
		if previous, exists := seen[field.Name]; exists {
			return inputFieldConflict(field.Name, previous, contribution.source)
		}
		seen[field.Name] = contribution.source
		collected = append(collected, field)
	}
	for _, declaration := range explicit {
		if declaration.err != "" {
			return fmt.Errorf("%s", declaration.err)
		}
		field := declaration.field
		if previous, exists := seen[field.Name]; exists {
			if !declaration.override || previous == "WithMoveInputField" {
				return inputFieldConflict(field.Name, previous, "WithMoveInputField")
			}
			for i := range collected {
				if collected[i].Name == field.Name {
					collected[i] = field
					break
				}
			}
			seen[field.Name] = "WithMoveInputField"
			continue
		}
		if declaration.override {
			return fmt.Errorf("creator-input field override %q did not match a behavior-provided field", field.Name)
		}
		seen[field.Name] = "WithMoveInputField"
		collected = append(collected, field)
	}

	choiceProjections, _ := config[configPropMoveChoiceProjections].([]moveChoiceProjectionDeclaration)
	if len(choiceProjections) > 1 {
		return fmt.Errorf("move has more than one choice projection; version one permits one")
	}
	if len(choiceProjections) == 1 {
		if choiceProjections[0].err != "" {
			return fmt.Errorf("move choice projection: %s", choiceProjections[0].err)
		}
		if err := boardgame.SetMoveChoiceProjection(config, choiceProjections[0].projection); err != nil {
			return err
		}
	}

	boardgame.SetMoveInputFields(config, collected)
	for _, field := range collected {
		propType, ok := props[field.Name]
		if !ok {
			return fmt.Errorf("move configures creator input for unknown field %q", field.Name)
		}
		if err := boardgame.ValidateMoveInputFieldDeclaration(field, propType); err != nil {
			return fmt.Errorf("creator-input field %q: %w", field.Name, err)
		}
	}
	return nil
}

// A wrapper move may inherit CurrentPlayer's private marker through multiple
// embedded bases even when its generated reader deliberately does not flatten
// that base's TargetPlayerIndex. The field is framework-owned and still gets
// its Go default, so omitting only this implicit context metadata is safe.
// Provider and explicit creator fields remain subject to the loud unknown-field
// validation below.
func representableMoveInputContributions(
	contributions []sourcedMoveInputField,
	props map[string]boardgame.PropertyType,
) []sourcedMoveInputField {
	result := make([]sourcedMoveInputField, 0, len(contributions))
	for _, contribution := range contributions {
		_, represented := props[contribution.field.Name]
		implicitContext := contribution.field.Disposition == boardgame.MoveInputContextOwned &&
			(contribution.source == "moves.CurrentPlayer" || contribution.source == "moves.AnyPlayer")
		if !represented && implicitContext {
			continue
		}
		result = append(result, contribution)
	}
	return result
}

// providedMoveInputContributions uses Go's ordinary method set. A move may
// implement MoveInputFieldsProvider directly or inherit it by embedding one
// behavior. A wrapper behavior that embeds several providers composes their
// results in its own MoveInputFields method. This is deterministic across
// package visibility and avoids reflective access to unexported or nil fields.
func providedMoveInputContributions(move AutoConfigurableMove) (result []sourcedMoveInputField, err error) {
	provider, ok := move.(boardgame.MoveInputFieldsProvider)
	if !ok {
		if hidden := embeddedMoveInputProviderTypes(reflect.TypeOf(move)); len(hidden) > 0 {
			return nil, fmt.Errorf("embedded creator-input providers %v are not composed by %T; add MoveInputFields to a wrapper behavior and return its complete contract", hidden, move)
		}
		return nil, nil
	}
	source := fmt.Sprintf("%T.MoveInputFields", move)
	defer func() {
		if recovered := recover(); recovered != nil {
			result = nil
			err = fmt.Errorf("creator-input provider %s panicked: %v; initialize pointer behaviors or use a value embedding", source, recovered)
		}
	}()
	for _, field := range provider.MoveInputFields() {
		result = append(result, sourcedMoveInputField{field: field, source: source})
	}
	return result, nil
}

// embeddedMoveInputProviderTypes inspects method sets only. It never reads or
// interfaces field values, so unexported and nil pointer behaviors are safe.
// It is used solely to turn Go's ambiguous-method case into a loud diagnostic.
func embeddedMoveInputProviderTypes(moveType reflect.Type) []string {
	providerType := reflect.TypeOf((*boardgame.MoveInputFieldsProvider)(nil)).Elem()
	var result []string
	visited := make(map[reflect.Type]bool)
	var visit func(reflect.Type)
	visit = func(valueType reflect.Type) {
		if valueType.Kind() == reflect.Ptr {
			valueType = valueType.Elem()
		}
		if valueType.Kind() != reflect.Struct {
			return
		}
		if visited[valueType] {
			return
		}
		visited[valueType] = true
		for i := 0; i < valueType.NumField(); i++ {
			field := valueType.Field(i)
			if !field.Anonymous {
				continue
			}
			candidate := field.Type
			implements := candidate.Implements(providerType)
			if candidate.Kind() != reflect.Ptr {
				implements = implements || reflect.PointerTo(candidate).Implements(providerType)
			}
			if implements {
				result = append(result, candidate.String())
				continue
			}
			visit(candidate)
		}
	}
	visit(moveType)
	return result
}

func contextOwnedTargetPlayerField() boardgame.MoveInputField {
	return boardgame.MoveInputField{
		Name:        "TargetPlayerIndex",
		Disposition: boardgame.MoveInputContextOwned,
		Codec:       boardgame.MoveInputCodecPlayerIndex,
	}
}

func inputFieldConflict(name, first, second string) error {
	return fmt.Errorf("creator-input field %q was configured by both %s and %s", name, first, second)
}
