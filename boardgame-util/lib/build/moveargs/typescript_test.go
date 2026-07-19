package moveargs

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
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

func TestGenerateTypeScriptEmitsExactNarrowedChoiceProjectionMap(t *testing.T) {
	moves := []MoveInfo{{
		Name: "Guess Card",
		Fields: []MoveFieldInfo{{
			Name: "GuessedCard", WireType: "enum", Disposition: "required", Codec: "enum",
			EnumName: "card", EnumValues: []string{"Guard", "Priest", "Unknown"},
		}},
	}, {
		Name: "Select Player",
		Fields: []MoveFieldInfo{{
			Name: "OtherPlayerIndex", WireType: "playerIndex", Disposition: "required", Codec: "player-index",
		}},
	}, {
		Name: "Choose Card",
		Fields: []MoveFieldInfo{{
			Name: "TargetCard", WireType: "int", Disposition: "required", Codec: "integer",
		}},
	}}
	choices := []ChoiceProjectionInfo{{
		MoveName: "Guess Card", FieldName: "GuessedCard", Source: boardgame.MoveChoiceSourceEnumValues,
		CandidateValues: []string{"Guard", "Priest"}, Disclosure: boardgame.MoveChoiceDisclosureActorExact,
	}, {
		MoveName: "Select Player", FieldName: "OtherPlayerIndex", Source: boardgame.MoveChoiceSourcePlayers,
		Disclosure: boardgame.MoveChoiceDisclosureActorExact,
	}, {
		MoveName: "Choose Card", FieldName: "TargetCard", Source: boardgame.MoveChoiceSourceStackSlots,
		StackSource: &boardgame.MoveChoiceStackSource{Scope: boardgame.MoveChoiceStackScopeActorPlayer, Property: "Hand"},
		Disclosure:  boardgame.MoveChoiceDisclosureActorExact,
	}}

	got := GenerateTypeScript(moves, choices)
	for _, want := range []string{
		"export type MoveChoiceProjections = {",
		`readonly field: "GuessedCard";`,
		`readonly value: "Guard" | "Priest";`,
		"readonly input: GuessCardInput;",
		`readonly field: "OtherPlayerIndex";`,
		"readonly value: number;",
		`readonly field: "TargetCard";`,
		`"source": "stack-slots"`,
		`"scope": "actor-player"`,
		`"property": "Hand"`,
		"moveChoiceProjectionSchemaFingerprint",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output did not contain %q:\n%s", want, got)
		}
	}
	if strings.Contains(got[strings.Index(got, "export type MoveChoiceProjections"):strings.Index(got, "export const moveInputSchema")], `"Unknown"`) {
		t.Fatalf("excluded enum sentinel leaked into projection value union:\n%s", got)
	}
	if strings.Contains(got[strings.Index(got, "export const moveChoiceProjectionSchema"):], "not generated") {
		t.Fatalf("audit rationale leaked into generated schema:\n%s", got)
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
	if !strings.Contains(got, "export const moveChoiceProjectionSchema = [] as const;") {
		t.Fatalf("empty choice schema was not an array literal:\n%s", got)
	}
	if strings.Contains(got, "moveChoiceProjectionSchema = null as const") {
		t.Fatalf("empty choice schema emitted invalid null const assertion:\n%s", got)
	}
}

func TestValidateTypeScriptSchemaRejectsMultipleProjectionSlices(t *testing.T) {
	err := ValidateTypeScriptSchema(nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "provide exactly one") {
		t.Fatalf("error = %v, want multiple-schema diagnostic", err)
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
