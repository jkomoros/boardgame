package boardgame

import (
	"strings"
	"testing"
)

func choiceProjectionMoveSchema() MoveInputSchemaMove {
	return MoveInputSchemaMove{Name: "Guess Card", Fields: []MoveInputSchemaField{
		{
			Name: "GuessedCard", WireType: "enum", Disposition: string(MoveInputRequired),
			Codec: string(MoveInputCodecEnum), EnumName: "card",
			EnumValues: []string{"Guard", "Priest", "Unknown"},
		},
		{
			Name: "TargetPlayerIndex", WireType: "playerIndex", Disposition: string(MoveInputContextOwned),
			Codec: string(MoveInputCodecPlayerIndex),
		},
	}}
}

func TestResolveMoveChoiceProjectionNarrowsEnumUniverse(t *testing.T) {
	projection := MoveChoiceProjection{
		FieldName:      "GuessedCard",
		ExcludedValues: []string{"Unknown"}, Disclosure: MoveChoiceDisclosureActorExact,
	}
	got, err := resolveMoveChoiceProjection(choiceProjectionMoveSchema(), projection)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.CandidateValues) != 2 || got.CandidateValues[0] != "Guard" || got.CandidateValues[1] != "Priest" {
		t.Fatalf("candidate values = %v", got.CandidateValues)
	}
	if got.Source != MoveChoiceSourceEnumValues {
		t.Fatalf("source = %q, want inferred enum source", got.Source)
	}
}

func TestResolveMoveChoiceProjectionValidatesDisclosureAndExclusions(t *testing.T) {
	base := MoveChoiceProjection{FieldName: "GuessedCard"}
	tests := []struct {
		name       string
		projection MoveChoiceProjection
		want       string
	}{
		{"missing disclosure", MoveChoiceProjection{FieldName: base.FieldName}, "unsupported disclosure"},
		{"unknown exclusion", MoveChoiceProjection{FieldName: base.FieldName, Disclosure: MoveChoiceDisclosureActorExact, ExcludedValues: []string{"Baron"}}, "not canonical"},
		{"duplicate exclusion", MoveChoiceProjection{FieldName: base.FieldName, Disclosure: MoveChoiceDisclosureActorExact, ExcludedValues: []string{"Unknown", "Unknown"}}, "duplicated"},
		{"exhaustive exclusions", MoveChoiceProjection{FieldName: base.FieldName, Disclosure: MoveChoiceDisclosureActorExact, ExcludedValues: []string{"Guard", "Priest", "Unknown"}}, "entire candidate universe"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := resolveMoveChoiceProjection(choiceProjectionMoveSchema(), test.projection)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestResolveMoveChoiceProjectionInfersPlayerSourceAndRejectsUnsupportedCodecs(t *testing.T) {
	playerMove := MoveInputSchemaMove{Name: "Choose Player", Fields: []MoveInputSchemaField{{
		Name: "Target", Disposition: string(MoveInputRequired), Codec: string(MoveInputCodecPlayerIndex),
	}}}
	got, err := resolveMoveChoiceProjection(playerMove, MoveChoiceProjection{
		FieldName: "Target", Disclosure: MoveChoiceDisclosureActorExact,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != MoveChoiceSourcePlayers {
		t.Fatalf("source = %q, want players", got.Source)
	}

	tests := []struct {
		name       string
		codec      MoveInputCodec
		exclusions []string
		want       string
	}{
		{"integer", MoveInputCodecInteger, nil, "unsupported choice codec"},
		{"boolean", MoveInputCodecBoolean, nil, "unsupported choice codec"},
		{"string", MoveInputCodecString, nil, "unsupported choice codec"},
		{"player exclusions", MoveInputCodecPlayerIndex, []string{"0"}, "does not support excluded values"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			move := MoveInputSchemaMove{Name: "Choose", Fields: []MoveInputSchemaField{{
				Name: "Value", Disposition: string(MoveInputRequired), Codec: string(test.codec),
			}}}
			_, err := resolveMoveChoiceProjection(move, MoveChoiceProjection{
				FieldName: "Value", ExcludedValues: test.exclusions, Disclosure: MoveChoiceDisclosureActorExact,
			})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestMoveChoiceProjectionFingerprintTracksCandidateUniverse(t *testing.T) {
	first := []MoveChoiceProjectionSchema{{
		MoveName: "Guess Card", FieldName: "GuessedCard", Source: MoveChoiceSourceEnumValues,
		CandidateValues: []string{"Guard", "Priest"}, Disclosure: MoveChoiceDisclosureActorExact,
	}}
	second := cloneMoveChoiceProjectionSchema(first)
	second[0].CandidateValues = []string{"Priest"}
	if FingerprintMoveChoiceProjectionSchema(first) == FingerprintMoveChoiceProjectionSchema(second) {
		t.Fatal("candidate-universe edit did not change choice-projection fingerprint")
	}
}

func TestMoveChoiceProjectionSchemaResourceLimits(t *testing.T) {
	projection := func(name string, candidates int) MoveChoiceProjectionSchema {
		values := make([]string, candidates)
		for i := range values {
			values[i] = "value"
		}
		return MoveChoiceProjectionSchema{
			MoveName: name, FieldName: "Choice", Source: MoveChoiceSourceEnumValues,
			CandidateValues: values, Disclosure: MoveChoiceDisclosureActorExact,
		}
	}
	tests := []struct {
		name   string
		schema []MoveChoiceProjectionSchema
		want   string
	}{
		{"too many sets", make([]MoveChoiceProjectionSchema, MoveChoiceProjectionMaxSets+1), "projections"},
		{"too many candidates in one set", []MoveChoiceProjectionSchema{projection("Large", MoveChoiceProjectionMaxCandidatesPerSet+1)}, "static choice candidates"},
		{"too many candidates in total", []MoveChoiceProjectionSchema{projection("First", 64), projection("Second", 64), projection("Third", 1)}, "total limit"},
		{"too many encoded candidate bytes", []MoveChoiceProjectionSchema{{
			MoveName: "Huge", FieldName: "Choice", Source: MoveChoiceSourceEnumValues,
			CandidateValues: []string{strings.Repeat("\\", MoveChoiceProjectionMaxStaticCandidateBytes)},
		}}, "encoded static candidate bytes"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateMoveChoiceProjectionSchemaLimits(test.schema)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
	if err := validateMoveChoiceProjectionSchemaLimits([]MoveChoiceProjectionSchema{projection("Valid", 64)}); err != nil {
		t.Fatalf("valid bounded schema failed: %v", err)
	}
}
