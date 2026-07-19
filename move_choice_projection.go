package boardgame

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

const moveChoiceProjectionConfigKey = "github.com/jkomoros/boardgame.MoveChoiceProjection"

// MoveChoiceProjectionSchemaVersion identifies the generated/runtime contract
// for finite move-choice projections. It is deliberately independent from the
// creator-input proposal protocol version and fingerprint.
const MoveChoiceProjectionSchemaVersion = 1

// Move-choice projection limits are part of the protocol's resource-safety
// contract. Static enum universes are rejected during manager boot; dynamic
// player universes are checked again by the server projector.
const (
	MoveChoiceProjectionMaxSets             = 8
	MoveChoiceProjectionMaxCandidatesPerSet = 64
	MoveChoiceProjectionMaxLegalEvaluations = 128
	// Static values get half of the total wire budget; the remainder is
	// reserved for move/field identity, candidate status, and envelope data.
	// The server still measures the complete viewer-specific JSON before Legal.
	MoveChoiceProjectionMaxStaticCandidateBytes = 32 << 10
	MoveChoiceProjectionMaxWireBytes            = 64 << 10
)

// MoveChoiceSource is a sealed framework-owned finite candidate universe.
// Sources enumerate possible values; ordinary move legality remains the sole
// authority for whether each complete binding is currently available.
type MoveChoiceSource string

const (
	MoveChoiceSourcePlayers    MoveChoiceSource = "players"
	MoveChoiceSourceEnumValues MoveChoiceSource = "enum-values"
)

// MoveChoiceDisclosure records the actor-visible semantics of a projection.
// Version one only supports exact candidate membership and availability
// disclosed to the player who would propose the move.
type MoveChoiceDisclosure string

const MoveChoiceDisclosureActorExact MoveChoiceDisclosure = "actor-exact"

// MoveChoiceProjection is the immutable declaration stored on a configured
// move. Public authoring goes through moves.WithChoices; Source may be empty and
// is then inferred from the resolved creator-input codec.
type MoveChoiceProjection struct {
	FieldName      string
	Source         MoveChoiceSource
	ExcludedValues []string
	Disclosure     MoveChoiceDisclosure
}

// MoveChoiceProjectionSchema is the frozen generated/runtime contract for one
// projected move. CandidateValues is populated for enum sources and already
// excludes implementation sentinels. Player candidates are state-dependent.
type MoveChoiceProjectionSchema struct {
	MoveName        string               `json:"moveName"`
	FieldName       string               `json:"fieldName"`
	Source          MoveChoiceSource     `json:"source"`
	CandidateValues []string             `json:"candidateValues,omitempty"`
	Disclosure      MoveChoiceDisclosure `json:"disclosure"`
}

// SetMoveChoiceProjection stores one choice-projection declaration in a move
// configuration. Game authors normally use moves.WithChoices.
func SetMoveChoiceProjection(config PropertyCollection, projection MoveChoiceProjection) error {
	if config == nil {
		return fmt.Errorf("move choice projection configuration is nil")
	}
	if _, exists := config[moveChoiceProjectionConfigKey]; exists {
		return fmt.Errorf("move has more than one choice projection")
	}
	config[moveChoiceProjectionConfigKey] = cloneMoveChoiceProjection(projection)
	return nil
}

// ConfiguredMoveChoiceProjection returns a defensive copy of a move's optional
// choice projection.
func ConfiguredMoveChoiceProjection(move Move) (*MoveChoiceProjection, error) {
	if move == nil || move.Info() == nil {
		return nil, nil
	}
	raw, exists := move.Info().CustomConfiguration()[moveChoiceProjectionConfigKey]
	if !exists {
		return nil, nil
	}
	projection, ok := raw.(MoveChoiceProjection)
	if !ok {
		return nil, fmt.Errorf("move %q has malformed choice projection configuration", move.Info().Name())
	}
	result := cloneMoveChoiceProjection(projection)
	return &result, nil
}

func cloneMoveChoiceProjection(projection MoveChoiceProjection) MoveChoiceProjection {
	projection.ExcludedValues = append([]string(nil), projection.ExcludedValues...)
	return projection
}

// BuildMoveChoiceProjectionSchema validates and freezes every opted-in player
// move projection. It is separate from BuildMoveInputSchema so presentation or
// safe-choice evolution cannot invalidate the creator proposal protocol.
func BuildMoveChoiceProjectionSchema(manager *GameManager) ([]MoveChoiceProjectionSchema, error) {
	if manager == nil {
		return nil, fmt.Errorf("manager is nil")
	}
	if manager.moveChoiceProjectionSchema != nil {
		return cloneMoveChoiceProjectionSchema(manager.moveChoiceProjectionSchema), nil
	}
	inputSchema, err := BuildMoveInputSchema(manager)
	if err != nil {
		return nil, err
	}
	inputByMove := make(map[string]MoveInputSchemaMove, len(inputSchema))
	for _, move := range inputSchema {
		inputByMove[move.Name] = move
	}

	result := make([]MoveChoiceProjectionSchema, 0)
	for _, moveType := range manager.moves {
		move := moveType.NewMove(nil)
		if move == nil {
			return nil, fmt.Errorf("could not construct move %q while building choice-projection schema", moveType.Name())
		}
		projection, err := ConfiguredMoveChoiceProjection(move)
		if err != nil {
			return nil, err
		}
		if projection == nil {
			continue
		}
		if fixUp, ok := move.(interface{ IsFixUp() bool }); ok && fixUp.IsFixUp() {
			return nil, fmt.Errorf("fix-up move %q cannot publish player choices", moveType.Name())
		}
		moveSchema, ok := inputByMove[move.Info().Name()]
		if !ok {
			return nil, fmt.Errorf("choice projection move %q has no creator-input schema", move.Info().Name())
		}
		item, err := resolveMoveChoiceProjection(moveSchema, *projection)
		if err != nil {
			return nil, fmt.Errorf("move %q choice projection: %w", move.Info().Name(), err)
		}
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].MoveName < result[j].MoveName })
	if err := validateMoveChoiceProjectionSchemaLimits(result); err != nil {
		return nil, err
	}
	return result, nil
}

func validateMoveChoiceProjectionSchemaLimits(schema []MoveChoiceProjectionSchema) error {
	if len(schema) > MoveChoiceProjectionMaxSets {
		return fmt.Errorf("game configures %d move choice projections; limit is %d", len(schema), MoveChoiceProjectionMaxSets)
	}
	staticCandidates := 0
	staticCandidateBytes := 0
	for _, projection := range schema {
		if projection.Source != MoveChoiceSourceEnumValues {
			continue
		}
		if len(projection.CandidateValues) > MoveChoiceProjectionMaxCandidatesPerSet {
			return fmt.Errorf("move %q has %d static choice candidates; limit is %d", projection.MoveName, len(projection.CandidateValues), MoveChoiceProjectionMaxCandidatesPerSet)
		}
		staticCandidates += len(projection.CandidateValues)
		if staticCandidates > MoveChoiceProjectionMaxLegalEvaluations {
			return fmt.Errorf("move choice projections have %d static candidates; total limit is %d", staticCandidates, MoveChoiceProjectionMaxLegalEvaluations)
		}
		for _, value := range projection.CandidateValues {
			// Count JSON encoding rather than raw string bytes so quotes,
			// backslashes, and control characters cannot expand past the cap.
			encoded, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("encode move %q candidate value: %w", projection.MoveName, err)
			}
			staticCandidateBytes += len(encoded)
			if staticCandidateBytes > MoveChoiceProjectionMaxStaticCandidateBytes {
				return fmt.Errorf("move choice projections have %d encoded static candidate bytes; limit is %d", staticCandidateBytes, MoveChoiceProjectionMaxStaticCandidateBytes)
			}
		}
	}
	return nil
}

func resolveMoveChoiceProjection(move MoveInputSchemaMove, projection MoveChoiceProjection) (MoveChoiceProjectionSchema, error) {
	if projection.Disclosure != MoveChoiceDisclosureActorExact {
		return MoveChoiceProjectionSchema{}, fmt.Errorf("unsupported disclosure %q", projection.Disclosure)
	}
	required := 0
	var field *MoveInputSchemaField
	for i := range move.Fields {
		if move.Fields[i].Disposition == string(MoveInputRequired) {
			required++
		}
		if move.Fields[i].Name == projection.FieldName {
			candidate := move.Fields[i]
			field = &candidate
		}
	}
	if field == nil {
		return MoveChoiceProjectionSchema{}, fmt.Errorf("field %q is not a configured creator input", projection.FieldName)
	}
	if required != 1 || field.Disposition != string(MoveInputRequired) {
		return MoveChoiceProjectionSchema{}, fmt.Errorf("version one requires exactly one required creator field and it must be %q; got %d required fields", projection.FieldName, required)
	}

	source := projection.Source
	if source == "" {
		switch field.Codec {
		case string(MoveInputCodecPlayerIndex):
			source = MoveChoiceSourcePlayers
		case string(MoveInputCodecEnum):
			source = MoveChoiceSourceEnumValues
		default:
			return MoveChoiceProjectionSchema{}, fmt.Errorf("field %q uses unsupported choice codec %q; choices require player-index or ordinary enum input", field.Name, field.Codec)
		}
	}

	item := MoveChoiceProjectionSchema{
		MoveName:   move.Name,
		FieldName:  field.Name,
		Source:     source,
		Disclosure: projection.Disclosure,
	}
	switch source {
	case MoveChoiceSourcePlayers:
		if field.Codec != string(MoveInputCodecPlayerIndex) {
			return MoveChoiceProjectionSchema{}, fmt.Errorf("player source requires codec %q, got %q", MoveInputCodecPlayerIndex, field.Codec)
		}
		if len(projection.ExcludedValues) != 0 {
			return MoveChoiceProjectionSchema{}, fmt.Errorf("player source does not support excluded values")
		}
	case MoveChoiceSourceEnumValues:
		if field.Codec != string(MoveInputCodecEnum) {
			return MoveChoiceProjectionSchema{}, fmt.Errorf("enum source requires codec %q, got %q", MoveInputCodecEnum, field.Codec)
		}
		values := make(map[string]bool, len(field.EnumValues))
		for _, value := range field.EnumValues {
			values[value] = true
		}
		excluded := make(map[string]bool, len(projection.ExcludedValues))
		for _, value := range projection.ExcludedValues {
			if excluded[value] {
				return MoveChoiceProjectionSchema{}, fmt.Errorf("excluded enum value %q is duplicated", value)
			}
			if !values[value] {
				return MoveChoiceProjectionSchema{}, fmt.Errorf("excluded enum value %q is not canonical", value)
			}
			excluded[value] = true
		}
		for _, value := range field.EnumValues {
			if !excluded[value] {
				item.CandidateValues = append(item.CandidateValues, value)
			}
		}
		if len(item.CandidateValues) == 0 {
			return MoveChoiceProjectionSchema{}, fmt.Errorf("excluded enum values remove the entire candidate universe")
		}
	default:
		return MoveChoiceProjectionSchema{}, fmt.Errorf("unsupported source %q", projection.Source)
	}
	return item, nil
}

// MoveChoiceProjectionSchemaFingerprint returns the fingerprint used by
// generated clients to validate projected-choice metadata.
func MoveChoiceProjectionSchemaFingerprint(manager *GameManager) (string, error) {
	if manager != nil && manager.moveChoiceProjectionSchemaFingerprint != "" {
		return manager.moveChoiceProjectionSchemaFingerprint, nil
	}
	schema, err := BuildMoveChoiceProjectionSchema(manager)
	if err != nil {
		return "", err
	}
	return FingerprintMoveChoiceProjectionSchema(schema), nil
}

// FingerprintMoveChoiceProjectionSchema fingerprints semantic projection data.
func FingerprintMoveChoiceProjectionSchema(schema []MoveChoiceProjectionSchema) string {
	canonical := cloneMoveChoiceProjectionSchema(schema)
	for i := range canonical {
		sort.Strings(canonical[i].CandidateValues)
	}
	sort.Slice(canonical, func(i, j int) bool { return canonical[i].MoveName < canonical[j].MoveName })
	encoded, err := json.Marshal(struct {
		Version     int                          `json:"version"`
		Projections []MoveChoiceProjectionSchema `json:"projections"`
	}{MoveChoiceProjectionSchemaVersion, canonical})
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func cloneMoveChoiceProjectionSchema(schema []MoveChoiceProjectionSchema) []MoveChoiceProjectionSchema {
	result := append([]MoveChoiceProjectionSchema(nil), schema...)
	for i := range result {
		result[i].CandidateValues = append([]string(nil), schema[i].CandidateValues...)
	}
	return result
}
