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

func TestValidateTypeResultRejectsInvalidAndDuplicateFields(t *testing.T) {
	for name, result := range map[string]TypeResult{
		"invalid": {GameFields: []FieldInfo{{Name: "not-valid", Type: "TypeInt"}}},
		"duplicate": {PlayerFields: []FieldInfo{
			{Name: "Score", Type: "TypeInt"},
			{Name: "Score", Type: "TypeInt"},
		}},
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
