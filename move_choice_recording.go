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

func configuredMoveChoiceRecording(move Move) (*MoveChoiceRecording, error) {
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

func applyMoveChoiceRecording(move Move, state State) error {
	recording, err := configuredMoveChoiceRecording(move)
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
