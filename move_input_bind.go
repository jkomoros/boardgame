package boardgame

import (
	"encoding/json"
	"fmt"

	"github.com/jkomoros/boardgame/enum"
)

// MoveWithInput creates a fresh move with state defaults, then binds only its
// declared creator input. Input uses JSON scalar representations (enum strings,
// booleans, integer numbers), matching the generated authoring contract. It
// rejects missing required fields, unknown/context-owned fields, and null.
// It does not test legality or submit the move. On error the partial move is
// discarded, so callers never receive a partially bound action.
func (g *Game) MoveWithInput(name string, input map[string]interface{}) (Move, error) {
	move := g.MoveByName(name)
	if move == nil {
		return nil, fmt.Errorf("unknown move %q", name)
	}
	fields, err := ResolveMoveInputFields(move)
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]MoveInputField, len(fields))
	for _, field := range fields {
		allowed[field.Name] = field
	}
	for name := range input {
		field, ok := allowed[name]
		if !ok || (field.Disposition != MoveInputRequired && field.Disposition != MoveInputServerDefaulted) {
			return nil, fmt.Errorf("move %q does not accept creator field %q", move.Info().Name(), name)
		}
	}
	for _, field := range fields {
		value, supplied := input[field.Name]
		if !supplied {
			if field.Disposition == MoveInputRequired {
				return nil, fmt.Errorf("move %q requires field %q", name, field.Name)
			}
			continue
		}
		blob, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", field.Name, err)
		}
		if string(blob) == "null" {
			return nil, fmt.Errorf("field %q must not be null", field.Name)
		}
		setter := move.ReadSetter()
		switch field.Codec {
		case MoveInputCodecInteger:
			var number int
			if err = json.Unmarshal(blob, &number); err == nil {
				if setter.Props()[field.Name] == TypeEnum {
					value, readErr := setter.EnumProp(field.Name)
					if readErr != nil {
						err = readErr
					} else {
						err = value.SetValue(enum.EnumKey(number))
					}
				} else {
					err = setter.SetIntProp(field.Name, number)
				}
			}
		case MoveInputCodecPlayerIndex:
			var number PlayerIndex
			if err = json.Unmarshal(blob, &number); err == nil {
				err = setter.SetPlayerIndexProp(field.Name, number)
			}
		case MoveInputCodecBoolean:
			var boolean bool
			if err = json.Unmarshal(blob, &boolean); err == nil {
				err = setter.SetBoolProp(field.Name, boolean)
			}
		case MoveInputCodecString:
			var text string
			if err = json.Unmarshal(blob, &text); err == nil {
				err = setter.SetStringProp(field.Name, text)
			}
		case MoveInputCodecEnum:
			var text string
			if err = json.Unmarshal(blob, &text); err == nil {
				value, readErr := setter.EnumProp(field.Name)
				if readErr != nil {
					err = readErr
				} else {
					err = value.SetStringValue(text)
				}
			}
		default:
			err = fmt.Errorf("unsupported input codec %q", field.Codec)
		}
		if err != nil {
			return nil, fmt.Errorf("move %q field %q: %w", name, field.Name, err)
		}
	}
	return move, nil
}
