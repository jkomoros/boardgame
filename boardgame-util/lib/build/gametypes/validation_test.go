package gametypes

import (
	"strings"
	"testing"
)

func TestValidateTypeResultRejectsInvalidGeneratedIdentifier(t *testing.T) {
	err := ValidateTypeResult(TypeResult{Enums: []EnumInfo{{Name: "123start"}}})
	if err == nil || !strings.Contains(err.Error(), `invalid TypeScript identifier "123startValue"`) {
		t.Fatalf("ValidateTypeResult() error = %v, want invalid generated identifier", err)
	}
}

func TestValidateTypeResultRejectsPascalCaseCollision(t *testing.T) {
	err := ValidateTypeResult(TypeResult{Decks: []DeckInfo{
		{Name: "playing-cards", Fields: []FieldInfo{{Name: "Rank", Type: "TypeInt"}}},
		{Name: "playing_cards", Fields: []FieldInfo{{Name: "Suit", Type: "TypeInt"}}},
	}})
	if err == nil || !strings.Contains(err.Error(), `both generate TypeScript declaration "PlayingCardsComponentValues"`) {
		t.Fatalf("ValidateTypeResult() error = %v, want generated declaration collision", err)
	}
}

func TestValidateTypeResultRejectsEnumMetadataCollisionsAndDuplicates(t *testing.T) {
	for name, want := range map[string]struct {
		result   TypeResult
		fragment string
	}{
		// Two enums whose names pascal-case to the same thing would generate
		// three colliding declarations each; the ordered list and the by-value
		// lookup are registered alongside the union so the collision is
		// reported rather than one enum silently shadowing the other.
		"metadata name collision": {
			result: TypeResult{Enums: []EnumInfo{
				{Name: "climate-card"},
				{Name: "climate_card"},
			}},
			fragment: `both generate TypeScript declaration "ClimateCardValue"`,
		},
		"duplicate key": {
			result: TypeResult{Enums: []EnumInfo{{Name: "color", Values: []EnumValueInfo{
				{Key: 0, Value: "Red"},
				{Key: 0, Value: "Blue"},
			}}}},
			fragment: `duplicate key 0`,
		},
		"duplicate value": {
			result: TypeResult{Enums: []EnumInfo{{Name: "color", Values: []EnumValueInfo{
				{Key: 0, Value: "Red"},
				{Key: 1, Value: "Red"},
			}}}},
			fragment: `duplicate value "Red"`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := ValidateTypeResult(want.result)
			if err == nil || !strings.Contains(err.Error(), want.fragment) {
				t.Fatalf("ValidateTypeResult() error = %v, want one containing %q", err, want.fragment)
			}
		})
	}
}

func TestValidateTypeResultRejectsInvalidAndDuplicateFields(t *testing.T) {
	for name, result := range map[string]TypeResult{
		"invalid": {GameFields: []FieldInfo{{Name: "not-valid", Type: "TypeInt"}}},
		"duplicate": {PlayerFields: []FieldInfo{
			{Name: "Score", Type: "TypeInt"},
			{Name: "Score", Type: "TypeInt"},
		}},
		"unknown type":              {GameFields: []FieldInfo{{Name: "Mystery", Type: "TypeFuture"}}},
		"computed interface":        {GameComputedFields: []FieldInfo{{Name: "Cards", Type: "TypeStack"}}},
		"stateful static component": {Decks: []DeckInfo{{Name: "cards", Fields: []FieldInfo{{Name: "Nested", Type: "TypeBoard"}}}}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := ValidateTypeResult(result); err == nil {
				t.Fatal("ValidateTypeResult() succeeded, want error")
			}
		})
	}
}

func TestValidateTypeResultAcceptsDistinctContract(t *testing.T) {
	err := ValidateTypeResult(TypeResult{
		GameFields:   []FieldInfo{{Name: "CurrentPlayer", Type: "TypePlayerIndex"}},
		PlayerFields: []FieldInfo{{Name: "Score", Type: "TypeInt"}},
		Enums:        []EnumInfo{{Name: "phase"}},
		Decks: []DeckInfo{{
			Name:          "playing_cards",
			Fields:        []FieldInfo{{Name: "Rank", Type: "TypeEnum", EnumName: "rank"}},
			DynamicFields: []FieldInfo{{Name: "FaceUp", Type: "TypeBool"}},
		}},
	})
	if err != nil {
		t.Fatalf("ValidateTypeResult() unexpected error: %v", err)
	}
}

func TestValidateTypeResultRejectsInvalidConstants(t *testing.T) {
	tests := map[string][]ConstantInfo{
		"empty name":     {{Name: "", Kind: "string", Value: "x"}},
		"duplicate name": {{Name: "same", Kind: "string", Value: "x"}, {Name: "same", Kind: "string", Value: "y"}},
		"unknown kind":   {{Name: "x", Kind: "object", Value: "{}"}},
		"bad integer":    {{Name: "x", Kind: "number", Value: "1.5"}},
		"unsafe integer": {{Name: "x", Kind: "number", Value: "9007199254740992"}},
		"bad boolean":    {{Name: "x", Kind: "boolean", Value: "yes"}},
	}
	for name, constants := range tests {
		t.Run(name, func(t *testing.T) {
			if err := ValidateTypeResult(TypeResult{Constants: constants}); err == nil {
				t.Fatal("ValidateTypeResult() succeeded, want error")
			}
		})
	}
}
