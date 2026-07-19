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
		FieldName: "GuessedCard", Source: MoveChoiceSourceEnumValues,
		ExcludedValues: []string{"Unknown"}, Disclosure: MoveChoiceDisclosureActorExact,
		AuditRationale: "the card catalogue and protection state are public",
	}
	got, err := resolveMoveChoiceProjection(choiceProjectionMoveSchema(), projection)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.CandidateValues) != 2 || got.CandidateValues[0] != "Guard" || got.CandidateValues[1] != "Priest" {
		t.Fatalf("candidate values = %v", got.CandidateValues)
	}
	if got.AuditRationale != projection.AuditRationale {
		t.Fatalf("audit rationale was not retained: %#v", got)
	}
}

func TestResolveMoveChoiceProjectionRequiresExplicitAuditedActorDisclosure(t *testing.T) {
	base := MoveChoiceProjection{FieldName: "GuessedCard", Source: MoveChoiceSourceEnumValues}
	tests := []struct {
		name       string
		projection MoveChoiceProjection
		want       string
	}{
		{"missing rationale", MoveChoiceProjection{FieldName: base.FieldName, Source: base.Source, Disclosure: MoveChoiceDisclosureActorExact}, "audit rationale"},
		{"missing disclosure", MoveChoiceProjection{FieldName: base.FieldName, Source: base.Source, AuditRationale: "reviewed"}, "unsupported disclosure"},
		{"unknown exclusion", MoveChoiceProjection{FieldName: base.FieldName, Source: base.Source, Disclosure: MoveChoiceDisclosureActorExact, AuditRationale: "reviewed", ExcludedValues: []string{"Baron"}}, "not canonical"},
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

func TestMoveChoiceProjectionFingerprintExcludesAuditProse(t *testing.T) {
	first := []MoveChoiceProjectionSchema{{
		MoveName: "Guess Card", FieldName: "GuessedCard", Source: MoveChoiceSourceEnumValues,
		CandidateValues: []string{"Guard", "Priest"}, Disclosure: MoveChoiceDisclosureActorExact,
		AuditRationale: "first review",
	}}
	second := cloneMoveChoiceProjectionSchema(first)
	second[0].AuditRationale = "rewritten review prose"
	if FingerprintMoveChoiceProjectionSchema(first) != FingerprintMoveChoiceProjectionSchema(second) {
		t.Fatal("audit-only edit changed choice-projection protocol fingerprint")
	}
	second[0].CandidateValues = []string{"Priest"}
	if FingerprintMoveChoiceProjectionSchema(first) == FingerprintMoveChoiceProjectionSchema(second) {
		t.Fatal("candidate-universe edit did not change choice-projection fingerprint")
	}
}
