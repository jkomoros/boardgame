package moveargs

import (
	"strings"
	"testing"
)

func TestGenerateTypeScriptSeparatesInputResolvedAndWire(t *testing.T) {
	moves := []MoveInfo{{
		Name: "Choose Target",
		Fields: []MoveFieldInfo{
			{Name: "RequiredCount", WireType: "int", Disposition: "required", Codec: "integer"},
			{Name: "OptionalLabel", WireType: "string", Disposition: "server-defaulted", Codec: "string"},
			{Name: "TargetPlayerIndex", WireType: "playerIndex", Disposition: "context-owned", Codec: "player-index"},
			{Name: "Mode", WireType: "enum", Disposition: "required", Codec: "enum", EnumName: "mode", EnumValues: []string{"Fast", "Careful"}},
		},
	}}

	got := GenerateTypeScript(moves)
	for _, want := range []string{
		"export interface ChooseTargetInput",
		"RequiredCount: number;",
		"OptionalLabel?: string;",
		`Mode: "Fast" | "Careful";`,
		"export interface ChooseTargetResolved",
		"TargetPlayerIndex: number;",
		"export interface ChooseTargetWire",
		"OptionalLabel?: string;",
		`"disposition": "context-owned"`,
		"moveInputSchemaFingerprint",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output did not contain %q:\n%s", want, got)
		}
	}

	inputSection := got[strings.Index(got, "export interface ChooseTargetInput"):strings.Index(got, "export interface ChooseTargetResolved")]
	if strings.Contains(inputSection, "TargetPlayerIndex") {
		t.Errorf("context-owned field leaked into creator input:\n%s", inputSection)
	}
	resolvedSection := got[strings.Index(got, "export interface ChooseTargetResolved"):strings.Index(got, "export interface ChooseTargetWire")]
	if strings.Contains(resolvedSection, "OptionalLabel?") || !strings.Contains(resolvedSection, "OptionalLabel: string") {
		t.Errorf("resolved defaulted field was not required:\n%s", resolvedSection)
	}
}

func TestGenerateTypeScriptRangeEnumUsesNumericAuthorCodec(t *testing.T) {
	got := GenerateTypeScript([]MoveInfo{{
		Name: "Move Token",
		Fields: []MoveFieldInfo{{
			Name: "SpaceIndex", WireType: "enum", Disposition: "required", Codec: "integer", EnumName: "spaces",
		}},
	}})
	if !strings.Contains(got, "SpaceIndex: number;") {
		t.Fatalf("range enum was not numeric:\n%s", got)
	}
	if !strings.Contains(got, "SpaceIndex: string;") {
		t.Fatalf("range enum wire field was not string encoded:\n%s", got)
	}
}

func TestGenerateTypeScriptZeroInputRejectsValues(t *testing.T) {
	got := GenerateTypeScript([]MoveInfo{{
		Name: "Roll Dice",
		Fields: []MoveFieldInfo{{
			Name: "TargetPlayerIndex", WireType: "playerIndex", Disposition: "context-owned", Codec: "player-index",
		}},
	}})
	if !strings.Contains(got, "export type RollDiceInput = Record<string, never>;") {
		t.Fatalf("zero-input move did not get exact empty record:\n%s", got)
	}
}

func TestGenerateTypeScriptIsDeterministic(t *testing.T) {
	moveA := MoveInfo{Name: "A", Fields: []MoveFieldInfo{
		{Name: "Z", WireType: "string", Disposition: "required", Codec: "string"},
		{Name: "A", WireType: "bool", Disposition: "required", Codec: "boolean"},
	}}
	moveB := MoveInfo{Name: "B", Fields: []MoveFieldInfo{{Name: "Value", WireType: "int", Disposition: "required", Codec: "integer"}}}
	first := GenerateTypeScript([]MoveInfo{moveB, moveA})
	moveAReordered := MoveInfo{Name: "A", Fields: []MoveFieldInfo{
		{Name: "A", WireType: "bool", Disposition: "required", Codec: "boolean"},
		{Name: "Z", WireType: "string", Disposition: "required", Codec: "string"},
	}}
	second := GenerateTypeScript([]MoveInfo{moveAReordered, moveB})
	if first != second {
		t.Fatal("output depended on field order")
	}
}

func TestValidateTypeScriptSchemaRejectsMalformedAndCollidingNames(t *testing.T) {
	tests := []struct {
		name  string
		moves []MoveInfo
		want  string
	}{
		{"empty symbol", []MoveInfo{{Name: "---"}}, "invalid TypeScript identifier"},
		{"numeric symbol", []MoveInfo{{Name: "123 move"}}, "invalid TypeScript identifier"},
		{"collision", []MoveInfo{{Name: "Move-Token"}, {Name: "Move Token"}}, "both generate"},
		{"duplicate", []MoveInfo{{Name: "Move"}, {Name: "Move"}}, "appears more than once"},
		{"unsupported codec", []MoveInfo{{Name: "Move", Fields: []MoveFieldInfo{{Name: "Mystery", Codec: "future"}}}}, "unsupported creator codec"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateTypeScriptSchema(test.moves)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}
