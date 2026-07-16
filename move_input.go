package boardgame

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

const moveInputFieldsConfigKey = "github.com/jkomoros/boardgame.MoveInputFields"

var reservedMoveInputWireNames = map[string]struct{}{
	"MoveType": {}, "admin": {}, "player": {}, "ExpectedVersion": {},
}

func validateMoveInputWireName(name string, disposition MoveInputDisposition) error {
	if _, reserved := reservedMoveInputWireNames[name]; reserved && disposition != MoveInputContextOwned {
		return fmt.Errorf("field %q collides with a reserved proposal protocol field", name)
	}
	return nil
}

// MoveInputDisposition says who is responsible for supplying a persisted move
// field. It is deliberately independent of DefaultsForState: responsibility is
// configuration, not something inferred by executing a default on one state.
type MoveInputDisposition string

const (
	MoveInputRequired        MoveInputDisposition = "required"
	MoveInputServerDefaulted MoveInputDisposition = "server-defaulted"
	MoveInputContextOwned    MoveInputDisposition = "context-owned"
	MoveInputUnsupported     MoveInputDisposition = "unsupported"
)

// MoveInputCodec is the creator-facing representation and validation rule for
// a move field. The persisted Go property type is recorded separately.
type MoveInputCodec string

const (
	MoveInputCodecInteger     MoveInputCodec = "integer"
	MoveInputCodecBoolean     MoveInputCodec = "boolean"
	MoveInputCodecEnum        MoveInputCodec = "enum"
	MoveInputCodecPlayerIndex MoveInputCodec = "player-index"
	MoveInputCodecString      MoveInputCodec = "string"
)

// MoveInputField describes one field in a configured move's creator-input
// contract. An empty Codec asks the framework to infer the standard codec.
type MoveInputField struct {
	Name        string
	Disposition MoveInputDisposition
	Codec       MoveInputCodec
}

// MoveInputSchemaField is the authoritative generated/runtime description of
// one configured move field.
type MoveInputSchemaField struct {
	Name        string   `json:"name"`
	WireType    string   `json:"wireType"`
	Disposition string   `json:"disposition"`
	Codec       string   `json:"codec,omitempty"`
	EnumName    string   `json:"enumName,omitempty"`
	EnumValues  []string `json:"enumValues,omitempty"`
}

// MoveInputSchemaMove describes the creator-input contract for one move type.
type MoveInputSchemaMove struct {
	Name   string                 `json:"name"`
	Fields []MoveInputSchemaField `json:"fields"`
}

// MoveInputFieldsProvider is implemented by embeddable move behaviors that
// own fields. auto.Config collects a single promoted provider automatically,
// so game authors normally do not add methods to their move structs. A wrapper
// behavior that embeds multiple providers implements this method to return its
// complete, composed contract using ordinary Go method-set rules.
type MoveInputFieldsProvider interface {
	MoveInputFields() []MoveInputField
}

// SetMoveInputFields stores a resolved field contract in a MoveConfig's custom
// configuration. It is primarily for configuration helpers.
func SetMoveInputFields(config PropertyCollection, fields []MoveInputField) {
	config[moveInputFieldsConfigKey] = append([]MoveInputField(nil), fields...)
}

// ConfiguredMoveInputFields returns the declarations stored for a move.
func ConfiguredMoveInputFields(move Move) []MoveInputField {
	if move == nil || move.Info() == nil {
		return nil
	}
	fields, _ := move.Info().CustomConfiguration()[moveInputFieldsConfigKey].([]MoveInputField)
	return append([]MoveInputField(nil), fields...)
}

// ResolveMoveInputFields returns the complete creator-input contract for a
// move. Supported unclassified fields are required. Unsupported property types
// fail loudly instead of silently disappearing from generated APIs.
func ResolveMoveInputFields(move Move) ([]MoveInputField, error) {
	if move == nil || move.ReadSetter() == nil || move.Info() == nil {
		return nil, fmt.Errorf("move is nil or is not initialized")
	}

	props := move.ReadSetter().Props()
	configured := ConfiguredMoveInputFields(move)
	byName := make(map[string]MoveInputField, len(configured))
	for _, field := range configured {
		if _, ok := props[field.Name]; !ok {
			return nil, fmt.Errorf("move %q configures creator input for unknown field %q", move.Info().Name(), field.Name)
		}
		if _, exists := byName[field.Name]; exists {
			return nil, fmt.Errorf("move %q configures creator input field %q more than once", move.Info().Name(), field.Name)
		}
		byName[field.Name] = field
	}

	result := make([]MoveInputField, 0, len(props))
	for name, propType := range props {
		field, configured := byName[name]
		if !configured {
			field = MoveInputField{Name: name, Disposition: MoveInputRequired}
		}
		if err := validateMoveInputDisposition(field.Disposition); err != nil {
			return nil, fmt.Errorf("move %q field %q: %w", move.Info().Name(), name, err)
		}
		if field.Codec == "" && field.Disposition != MoveInputUnsupported {
			codec, err := inferMoveInputCodec(move, name, propType)
			if err != nil {
				return nil, fmt.Errorf("move %q field %q: %w; mark it unsupported explicitly if intentional", move.Info().Name(), name, err)
			}
			field.Codec = codec
		}
		if field.Disposition != MoveInputUnsupported {
			if err := validateMoveInputCodec(field.Codec, propType); err != nil {
				return nil, fmt.Errorf("move %q field %q: %w", move.Info().Name(), name, err)
			}
		}
		result = append(result, field)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func validateMoveInputDisposition(disposition MoveInputDisposition) error {
	switch disposition {
	case MoveInputRequired, MoveInputServerDefaulted, MoveInputContextOwned, MoveInputUnsupported:
		return nil
	default:
		return fmt.Errorf("unknown creator-input disposition %q", disposition)
	}
}

func inferMoveInputCodec(move Move, fieldName string, propType PropertyType) (MoveInputCodec, error) {
	switch propType {
	case TypeInt:
		return MoveInputCodecInteger, nil
	case TypeBool:
		return MoveInputCodecBoolean, nil
	case TypeString:
		return MoveInputCodecString, nil
	case TypePlayerIndex:
		return MoveInputCodecPlayerIndex, nil
	case TypeEnum:
		val, err := move.ReadSetter().ImmutableEnumProp(fieldName)
		if err != nil || val == nil || val.Enum() == nil {
			return "", fmt.Errorf("could not inspect enum metadata")
		}
		if ranged := val.Enum().RangeEnum(); ranged != nil {
			return MoveInputCodecInteger, nil
		}
		return MoveInputCodecEnum, nil
	default:
		return "", fmt.Errorf("property type %v has no standard creator-input codec", propType)
	}
}

func validateMoveInputCodec(codec MoveInputCodec, propType PropertyType) error {
	valid := false
	switch codec {
	case MoveInputCodecInteger:
		valid = propType == TypeInt || propType == TypeEnum
	case MoveInputCodecBoolean:
		valid = propType == TypeBool
	case MoveInputCodecEnum:
		valid = propType == TypeEnum
	case MoveInputCodecPlayerIndex:
		valid = propType == TypePlayerIndex
	case MoveInputCodecString:
		valid = propType == TypeString
	default:
		return fmt.Errorf("unknown creator-input codec %q", codec)
	}
	if !valid {
		return fmt.Errorf("creator-input codec %q is incompatible with property type %v", codec, propType)
	}
	return nil
}

// ValidateMoveInputFieldDeclaration validates a configured declaration without
// requiring state-dependent inflation. Enum membership metadata is resolved
// later by ResolveMoveInputFields once the move is installed.
func ValidateMoveInputFieldDeclaration(field MoveInputField, propType PropertyType) error {
	if field.Name == "" {
		return fmt.Errorf("creator-input field name is empty")
	}
	if err := validateMoveInputDisposition(field.Disposition); err != nil {
		return err
	}
	if field.Codec != "" && field.Disposition != MoveInputUnsupported {
		return validateMoveInputCodec(field.Codec, propType)
	}
	return nil
}

// BuildMoveInputSchema derives the deterministic player-move schema used by
// both code generation and the server's stale-generation fingerprint.
func BuildMoveInputSchema(manager *GameManager) ([]MoveInputSchemaMove, error) {
	if manager == nil {
		return nil, fmt.Errorf("manager is nil")
	}
	if manager.moveInputSchema != nil {
		return cloneMoveInputSchema(manager.moveInputSchema), nil
	}
	schema := make([]MoveInputSchemaMove, 0)
	// NewGameManager validates and freezes this schema before setting
	// initialized. The public ExampleMoves helper intentionally returns nil
	// until initialization is complete, so walk the already-installed move
	// types directly here. This also keeps boot-time validation and later
	// generation on exactly the same path.
	for _, moveType := range manager.moves {
		move := moveType.NewMove(nil)
		if move == nil {
			return nil, fmt.Errorf("could not construct move %q while building creator-input schema", moveType.Name())
		}
		if fixUp, ok := move.(interface{ IsFixUp() bool }); ok && fixUp.IsFixUp() {
			continue
		}
		resolved, err := ResolveMoveInputFields(move)
		if err != nil {
			return nil, err
		}
		fields := make([]MoveInputSchemaField, 0, len(resolved))
		props := move.ReadSetter().Props()
		for _, field := range resolved {
			if err := validateMoveInputWireName(field.Name, field.Disposition); err != nil {
				return nil, fmt.Errorf("move %q: %w", moveType.Name(), err)
			}
			propType := props[field.Name]
			item := MoveInputSchemaField{
				Name:        field.Name,
				WireType:    moveInputWireType(propType),
				Disposition: string(field.Disposition),
				Codec:       string(field.Codec),
			}
			if propType == TypeEnum {
				if enumVal, enumErr := move.ReadSetter().ImmutableEnumProp(field.Name); enumErr == nil && enumVal != nil {
					item.EnumName = enumVal.Enum().Name()
					if field.Codec == MoveInputCodecEnum {
						for _, value := range enumVal.Enum().Values() {
							item.EnumValues = append(item.EnumValues, enumVal.Enum().String(value))
						}
						sort.Strings(item.EnumValues)
					}
				}
			}
			fields = append(fields, item)
		}
		schema = append(schema, MoveInputSchemaMove{Name: move.Info().Name(), Fields: fields})
	}
	sort.Slice(schema, func(i, j int) bool { return schema[i].Name < schema[j].Name })
	return schema, nil
}

// MoveInputSchemaFingerprint returns the fingerprint a matching generated
// client must present before using the safe proposal API.
func MoveInputSchemaFingerprint(manager *GameManager) (string, error) {
	if manager != nil && manager.moveInputSchemaFingerprint != "" {
		return manager.moveInputSchemaFingerprint, nil
	}
	schema, err := BuildMoveInputSchema(manager)
	if err != nil {
		return "", err
	}
	return FingerprintMoveInputSchema(schema), nil
}

func cloneMoveInputSchema(schema []MoveInputSchemaMove) []MoveInputSchemaMove {
	result := make([]MoveInputSchemaMove, len(schema))
	for i, move := range schema {
		result[i] = move
		result[i].Fields = append([]MoveInputSchemaField(nil), move.Fields...)
		for j := range result[i].Fields {
			result[i].Fields[j].EnumValues = append([]string(nil), move.Fields[j].EnumValues...)
		}
	}
	return result
}

// FingerprintMoveInputSchema fingerprints canonical JSON for a sorted schema.
func FingerprintMoveInputSchema(schema []MoveInputSchemaMove) string {
	canonical := make([]MoveInputSchemaMove, len(schema))
	for i, move := range schema {
		canonical[i] = move
		canonical[i].Fields = append([]MoveInputSchemaField(nil), move.Fields...)
		sort.Slice(canonical[i].Fields, func(a, b int) bool {
			return canonical[i].Fields[a].Name < canonical[i].Fields[b].Name
		})
	}
	sort.Slice(canonical, func(i, j int) bool { return canonical[i].Name < canonical[j].Name })
	encoded, err := json.MarshalIndent(canonical, "", "  ")
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func moveInputWireType(propType PropertyType) string {
	switch propType {
	case TypeInt:
		return "int"
	case TypeBool:
		return "bool"
	case TypeString:
		return "string"
	case TypePlayerIndex:
		return "playerIndex"
	case TypeEnum:
		return "enum"
	default:
		return ""
	}
}
