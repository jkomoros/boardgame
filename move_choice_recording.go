package boardgame

import "fmt"

const moveChoiceRecordingConfigKey = "github.com/jkomoros/boardgame.MoveChoiceRecording"

// MoveChoiceRecordingScope identifies the state object that receives a
// configured recorded choice. Public game authoring goes through
// moves.InPlayer and moves.InGame.
type MoveChoiceRecordingScope string

const (
	MoveChoiceRecordingPlayer MoveChoiceRecordingScope = "player"
	MoveChoiceRecordingGame   MoveChoiceRecordingScope = "game"
)

// MoveChoiceRecording is the engine-owned state-effect descriptor installed
// by moves.WithRecordedChoice. It is data rather than a promoted method so a
// concrete move cannot accidentally shadow the recording behavior.
type MoveChoiceRecording struct {
	FieldName            string
	DestinationScope     MoveChoiceRecordingScope
	DestinationProperty  string
	DestinationPlayerKey string
}

// SetMoveChoiceRecording stores a recording descriptor in move configuration.
// Game authors normally use moves.WithRecordedChoice.
func SetMoveChoiceRecording(config PropertyCollection, recording MoveChoiceRecording) error {
	if config == nil {
		return fmt.Errorf("move choice recording configuration is nil")
	}
	if _, exists := config[moveChoiceRecordingConfigKey]; exists {
		return fmt.Errorf("move has more than one choice recording")
	}
	config[moveChoiceRecordingConfigKey] = recording
	return nil
}

// ConfiguredMoveChoiceRecording returns the engine-owned recording descriptor
// installed on move, if any. The defensive value is suitable for manager-time
// consistency validation by higher-level authoring helpers.
func ConfiguredMoveChoiceRecording(move Move) (*MoveChoiceRecording, error) {
	if move == nil || move.Info() == nil {
		return nil, nil
	}
	raw, exists := move.Info().CustomConfiguration()[moveChoiceRecordingConfigKey]
	if !exists {
		return nil, nil
	}
	recording, ok := raw.(MoveChoiceRecording)
	if !ok {
		return nil, fmt.Errorf("move %q has malformed choice recording configuration", move.Info().Name())
	}
	return &recording, nil
}

// validateMoveChoiceRecordingConfiguration validates the engine-owned state
// effect independently of a move's ValidConfiguration method. That separation
// is intentional: an outer game move can override a promoted validation
// method, but it cannot bypass an effect the engine itself will apply.
func validateMoveChoiceRecordingConfiguration(move Move, state State) error {
	recording, err := ConfiguredMoveChoiceRecording(move)
	if err != nil || recording == nil {
		return err
	}
	if state == nil || state.GameState() == nil || move.ReadSetter() == nil {
		return fmt.Errorf("recorded choice requires initialized move and state properties")
	}

	source := move.ReadSetter()
	sourceType, ok := source.Props()[recording.FieldName]
	if !ok {
		return fmt.Errorf("recorded choice source %q does not exist", recording.FieldName)
	}
	if sourceType != TypePlayerIndex && sourceType != TypeEnum {
		return fmt.Errorf("recorded choice source %q has unsupported type %v", recording.FieldName, sourceType)
	}
	fields, err := ResolveMoveInputFields(move)
	if err != nil {
		return fmt.Errorf("resolve recorded choice input: %w", err)
	}
	required := false
	for _, field := range fields {
		if field.Name == recording.FieldName {
			required = field.Disposition == MoveInputRequired
			break
		}
	}
	if !required {
		return fmt.Errorf("recorded choice source %q must be a required creator input", recording.FieldName)
	}

	var destination PropertyReadSetter
	switch recording.DestinationScope {
	case MoveChoiceRecordingGame:
		destination = state.GameState().ReadSetter()
	case MoveChoiceRecordingPlayer:
		if keyType, ok := source.Props()[recording.DestinationPlayerKey]; !ok || keyType != TypePlayerIndex {
			return fmt.Errorf("recorded choice destination player key %q must be a player-index property", recording.DestinationPlayerKey)
		}
		players := state.PlayerStates()
		if len(players) == 0 {
			return fmt.Errorf("recorded choice requires at least one player state")
		}
		destination = players[0].ReadSetter()
	default:
		return fmt.Errorf("recorded choice has unsupported destination scope %q", recording.DestinationScope)
	}
	if destination == nil {
		return fmt.Errorf("recorded choice destination has no property reader")
	}
	destinationType, ok := destination.Props()[recording.DestinationProperty]
	if !ok {
		return fmt.Errorf("recorded choice destination %q does not exist", recording.DestinationProperty)
	}
	if destinationType != sourceType {
		return fmt.Errorf("recorded choice source %q type %v does not match destination %q type %v", recording.FieldName, sourceType, recording.DestinationProperty, destinationType)
	}
	if sourceType == TypeEnum {
		sourceEnum, err := source.ImmutableEnumProp(recording.FieldName)
		if err != nil {
			return fmt.Errorf("read recorded choice source enum: %w", err)
		}
		destinationEnum, err := destination.ImmutableEnumProp(recording.DestinationProperty)
		if err != nil {
			return fmt.Errorf("read recorded choice destination enum: %w", err)
		}
		if sourceEnum == nil || destinationEnum == nil || sourceEnum.Enum() != destinationEnum.Enum() {
			return fmt.Errorf("recorded choice source and destination use different enums")
		}
	}
	return nil
}

func applyMoveChoiceRecording(move Move, state State) error {
	recording, err := ConfiguredMoveChoiceRecording(move)
	if err != nil || recording == nil {
		return err
	}
	source := move.ReadSetter()
	if source == nil {
		return fmt.Errorf("recorded choice move has no property reader")
	}
	sourceType, ok := source.Props()[recording.FieldName]
	if !ok {
		return fmt.Errorf("recorded choice source %q does not exist", recording.FieldName)
	}

	var destination PropertyReadSetter
	switch recording.DestinationScope {
	case MoveChoiceRecordingGame:
		destination = state.GameState().ReadSetter()
	case MoveChoiceRecordingPlayer:
		player, err := source.PlayerIndexProp(recording.DestinationPlayerKey)
		if err != nil {
			return fmt.Errorf("read recorded choice destination player: %w", err)
		}
		players := state.PlayerStates()
		if player < 0 || int(player) >= len(players) {
			return fmt.Errorf("recorded choice destination player %d is invalid", player)
		}
		destination = players[player].ReadSetter()
	default:
		return fmt.Errorf("recorded choice has unsupported destination scope %q", recording.DestinationScope)
	}

	switch sourceType {
	case TypePlayerIndex:
		value, err := source.PlayerIndexProp(recording.FieldName)
		if err != nil {
			return fmt.Errorf("read recorded player choice: %w", err)
		}
		if err := destination.SetPlayerIndexProp(recording.DestinationProperty, value); err != nil {
			return fmt.Errorf("store recorded player choice: %w", err)
		}
	case TypeEnum:
		value, err := source.ImmutableEnumProp(recording.FieldName)
		if err != nil {
			return fmt.Errorf("read recorded enum choice: %w", err)
		}
		target, err := destination.EnumProp(recording.DestinationProperty)
		if err != nil {
			return fmt.Errorf("fetch recorded enum destination: %w", err)
		}
		if err := target.SetValue(value.Value()); err != nil {
			return fmt.Errorf("store recorded enum choice: %w", err)
		}
	default:
		return fmt.Errorf("recorded choice source %q has unsupported type %v", recording.FieldName, sourceType)
	}
	return nil
}
