package moves

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jkomoros/boardgame"
)

const configPropRecordedChoices = fullyQualifiedPackageName + "RecordedChoices"

type recordedChoiceScope string

const (
	recordedChoicePlayer recordedChoiceScope = "player"
	recordedChoiceGame   recordedChoiceScope = "game"
)

// ChoiceDestination is a sealed destination for WithRecordedChoice.
type ChoiceDestination interface {
	recordedChoiceDestination() recordedChoiceTarget
}

type recordedChoiceTarget struct {
	scope    recordedChoiceScope
	property string
	err      string
}

func (t recordedChoiceTarget) recordedChoiceDestination() recordedChoiceTarget { return t }

// InPlayer stores the committed choice in a property on the proposing/current
// player's state.
func InPlayer(property string) ChoiceDestination {
	return recordedChoiceTarget{scope: recordedChoicePlayer, property: property}
}

// InGame stores the committed choice in a property on the game state.
func InGame(property string) ChoiceDestination {
	return recordedChoiceTarget{scope: recordedChoiceGame, property: property}
}

type recordedChoiceDeclaration struct {
	field  string
	target recordedChoiceTarget
}

// WithRecordedChoice configures a RecordCurrentPlayerChoice move to present a
// finite field and copy the committed value into state. It expands to the same
// actor-exact choice projection as WithChoices; canonical Legal remains the
// sole authority for candidate availability.
func WithRecordedChoice(field string, destination ChoiceDestination, options ...ChoiceOption) CustomConfigurationOption {
	var target recordedChoiceTarget
	if destination == nil {
		target.err = "recorded choice destination is nil"
	} else {
		target = destination.recordedChoiceDestination()
	}
	return func(config boardgame.PropertyCollection) {
		declarations, _ := config[configPropRecordedChoices].([]recordedChoiceDeclaration)
		config[configPropRecordedChoices] = append(declarations, recordedChoiceDeclaration{field: field, target: target})
		_ = boardgame.SetMoveChoiceRecording(config, boardgame.MoveChoiceRecording{
			FieldName: field, DestinationProperty: target.property,
			DestinationScope:     boardgame.MoveChoiceRecordingScope(target.scope),
			DestinationPlayerKey: "TargetPlayerIndex",
		})
		WithChoices(field, options...)(config)
	}
}

// RecordCurrentPlayerChoice is a complete current-player move for the common
// semantic operation "commit one finite scalar answer for later rules to
// consume." Embed it in a named game move and configure WithRecordedChoice.
// The framework-provided TopLevelStruct back-reference lets Apply read the
// choice field from the outer move without replacing its move identity.
type RecordCurrentPlayerChoice struct {
	CurrentPlayer
}

func (*RecordCurrentPlayerChoice) consumesRecordedChoiceConfiguration() {}

type recordedChoiceBinding struct {
	field    string
	target   recordedChoiceTarget
	propType boardgame.PropertyType
	source   boardgame.PropertyReadSetter
	dest     boardgame.PropertyReadSetter
}

func (r *RecordCurrentPlayerChoice) binding(state boardgame.State, validating bool) (*recordedChoiceBinding, error) {
	top := r.TopLevelStruct()
	if top == nil || top.ReadSetter() == nil {
		return nil, errors.New("RecordCurrentPlayerChoice: top-level move is unavailable")
	}
	declarations, _ := r.CustomConfiguration()[configPropRecordedChoices].([]recordedChoiceDeclaration)
	if len(declarations) != 1 {
		return nil, fmt.Errorf("RecordCurrentPlayerChoice: expected exactly one WithRecordedChoice declaration, got %d", len(declarations))
	}
	declaration := declarations[0]
	if declaration.target.err != "" {
		return nil, errors.New("RecordCurrentPlayerChoice: " + declaration.target.err)
	}
	if strings.TrimSpace(declaration.field) == "" || strings.TrimSpace(declaration.target.property) == "" {
		return nil, errors.New("RecordCurrentPlayerChoice: source and destination properties must be non-empty")
	}

	fields, err := boardgame.ResolveMoveInputFields(top)
	if err != nil {
		return nil, fmt.Errorf("RecordCurrentPlayerChoice: resolve move inputs: %w", err)
	}
	required := 0
	for _, field := range fields {
		if field.Disposition != boardgame.MoveInputRequired {
			continue
		}
		required++
		if field.Name != declaration.field {
			return nil, fmt.Errorf("RecordCurrentPlayerChoice: required creator field is %q, not configured field %q", field.Name, declaration.field)
		}
	}
	if required != 1 {
		return nil, fmt.Errorf("RecordCurrentPlayerChoice: expected exactly one required creator field, got %d", required)
	}

	sourceType, ok := top.ReadSetter().Props()[declaration.field]
	if !ok {
		return nil, fmt.Errorf("RecordCurrentPlayerChoice: source property %q does not exist on the top-level move", declaration.field)
	}
	if sourceType != boardgame.TypePlayerIndex && sourceType != boardgame.TypeEnum {
		return nil, fmt.Errorf("RecordCurrentPlayerChoice: source property %q has unsupported type %v", declaration.field, sourceType)
	}

	var destination boardgame.PropertyReadSetter
	switch declaration.target.scope {
	case recordedChoicePlayer:
		players := state.PlayerStates()
		if len(players) == 0 {
			return nil, errors.New("RecordCurrentPlayerChoice: game has no player states")
		}
		index := r.TargetPlayerIndex
		if index < 0 || int(index) >= len(players) {
			if !validating {
				return nil, fmt.Errorf("RecordCurrentPlayerChoice: target player index %d is invalid", index)
			}
			// During manager validation a throwaway move may not yet have current
			// player defaults. Any player state has the same schema.
			index = 0
		}
		destination = players[index].ReadSetter()
	case recordedChoiceGame:
		destination = state.GameState().ReadSetter()
	default:
		return nil, fmt.Errorf("RecordCurrentPlayerChoice: unsupported destination scope %q", declaration.target.scope)
	}
	destinationType, ok := destination.Props()[declaration.target.property]
	if !ok {
		return nil, fmt.Errorf("RecordCurrentPlayerChoice: destination property %q does not exist on %s state", declaration.target.property, declaration.target.scope)
	}
	if destinationType != sourceType {
		return nil, fmt.Errorf("RecordCurrentPlayerChoice: source %q type %v does not match destination %q type %v", declaration.field, sourceType, declaration.target.property, destinationType)
	}

	binding := &recordedChoiceBinding{
		field: declaration.field, target: declaration.target, propType: sourceType,
		source: top.ReadSetter(), dest: destination,
	}
	if sourceType == boardgame.TypeEnum {
		source, err := binding.source.ImmutableEnumProp(binding.field)
		if err != nil {
			return nil, fmt.Errorf("RecordCurrentPlayerChoice: read source enum: %w", err)
		}
		dest, err := binding.dest.ImmutableEnumProp(binding.target.property)
		if err != nil {
			return nil, fmt.Errorf("RecordCurrentPlayerChoice: read destination enum: %w", err)
		}
		if source == nil || dest == nil || source.Enum() != dest.Enum() {
			return nil, errors.New("RecordCurrentPlayerChoice: source and destination use different enums")
		}
	}
	return binding, nil
}

// ValidConfiguration verifies the source/destination contract at manager boot.
func (r *RecordCurrentPlayerChoice) ValidConfiguration(exampleState boardgame.State) error {
	if err := r.CurrentPlayer.ValidConfiguration(exampleState); err != nil {
		return err
	}
	_, err := r.binding(exampleState, true)
	return err
}

// Apply completes the Move interface for the common recording-only case. The
// actual configured recording is performed by the engine-owned descriptor.
func (*RecordCurrentPlayerChoice) Apply(boardgame.State) error { return nil }
