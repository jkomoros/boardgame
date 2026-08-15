package gametypes

import (
	"strings"
	"testing"
)

// enumValuesForTest builds the ordered value list an enum with no presentation
// data produces: ascending keys from zero, label equal to the string value, and
// the lowest value flagged as the enum's default.
func enumValuesForTest(values ...string) []EnumValueInfo {
	result := make([]EnumValueInfo, len(values))
	for i, value := range values {
		result[i] = EnumValueInfo{Key: i, Value: value, Label: value, IsDefault: i == 0}
	}
	return result
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"cards", "Cards"},
		{"playing cards", "PlayingCards"},
		{"playing_cards", "PlayingCards"},
		{"playing-cards", "PlayingCards"},
		{"UPPER", "UPPER"},
		{"a", "A"},
		{"with spaces and_underscores-and-hyphens", "WithSpacesAndUnderscoresAndHyphens"},
		{"123start", "123start"},
		{"  extra   spaces  ", "ExtraSpaces"},
		{"", ""},
		{"   ", ""},
		{"!@#$%", ""},
		{"hello!world", "Helloworld"},
	}

	for _, tc := range tests {
		result := toPascalCase(tc.input)
		if result != tc.expected {
			t.Errorf("toPascalCase(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestEscapeForTS(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{`he said "hi"`, `he said \"hi\"`},
		{"line1\nline2", `line1\nline2`},
		{`back\slash`, `back\\slash`},
		{"tab\there", "tab\there"},
		{"", ""},
	}

	for _, tc := range tests {
		result := escapeForTS(tc.input)
		if result != tc.expected {
			t.Errorf("escapeForTS(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestBaseFieldTypeToTS(t *testing.T) {
	enums := []EnumInfo{
		{Name: "color", Values: enumValuesForTest("Red", "Blue")},
	}

	tests := []struct {
		field    FieldInfo
		expected string
	}{
		{FieldInfo{Type: "TypeBool"}, "boolean"},
		{FieldInfo{Type: "TypeInt"}, "number"},
		{FieldInfo{Type: "TypeString"}, "string"},
		{FieldInfo{Type: "TypePlayerIndex"}, "number"},
		{FieldInfo{Type: "TypeIntSlice"}, "readonly number[]"},
		{FieldInfo{Type: "TypeBoolSlice"}, "readonly boolean[]"},
		{FieldInfo{Type: "TypeStringSlice"}, "readonly string[]"},
		{FieldInfo{Type: "TypePlayerIndexSlice"}, "readonly number[]"},
		{FieldInfo{Type: "TypeEnum", EnumName: "color"}, "ColorValue"},
		{FieldInfo{Type: "TypeEnum"}, "string"},
		{FieldInfo{Type: "TypeEnumSlice", EnumName: "color"}, "readonly ColorValue[]"},
		{FieldInfo{Type: "TypeEnumSlice"}, "readonly string[]"},
		{FieldInfo{Type: "TypeSomethingUnknown"}, "never"},
	}

	for _, tc := range tests {
		result := baseFieldTypeToTS(tc.field, enums)
		if result != tc.expected {
			t.Errorf("baseFieldTypeToTS(%+v) = %q, want %q", tc.field, result, tc.expected)
		}
	}
}

func TestDynamicFieldTypeToTS(t *testing.T) {
	enums := []EnumInfo{
		{Name: "color", Values: enumValuesForTest("Red", "Blue")},
	}

	tests := []struct {
		field    FieldInfo
		expected string
	}{
		{FieldInfo{Type: "TypeBool"}, "boolean"},
		{FieldInfo{Type: "TypeInt"}, "number"},
		{FieldInfo{Type: "TypeString"}, "string"},
		{FieldInfo{Type: "TypeStack"}, "RawStack"},
		{FieldInfo{Type: "TypeTimer"}, "ExpandedTimer"},
		{FieldInfo{Type: "TypeBoard"}, "Board"},
		{FieldInfo{Type: "TypeEnum", EnumName: "color"}, "ColorValue"},
		{FieldInfo{Type: "TypeEnum"}, "string"},
	}

	for _, tc := range tests {
		result := dynamicFieldTypeToTS(tc.field, enums)
		if result != tc.expected {
			t.Errorf("dynamicFieldTypeToTS(%+v) = %q, want %q", tc.field, result, tc.expected)
		}
	}
}

func TestStateFieldTypeToTS(t *testing.T) {
	decks := []DeckInfo{
		{Name: "cards", Fields: []FieldInfo{{Name: "Rank", Type: "TypeString"}}},
		{Name: "tokens", Fields: nil},
		{Name: "dice", Fields: []FieldInfo{{Name: "Faces", Type: "TypeIntSlice"}}, DynamicFields: []FieldInfo{{Name: "Value", Type: "TypeInt"}}},
		{Name: "pieces", DynamicFields: []FieldInfo{{Name: "Crowned", Type: "TypeBool"}}},
	}
	enums := []EnumInfo{
		{Name: "phase", Values: enumValuesForTest("Setup", "Playing")},
	}

	tests := []struct {
		field    FieldInfo
		expected string
	}{
		{FieldInfo{Type: "TypeTimer"}, "ExpandedTimer"},
		{FieldInfo{Type: "TypeStack", DeckName: "cards"}, "ExpandedStack<CardsComponentValues, Readonly<Record<string, never>>>"},
		{FieldInfo{Type: "TypeStack", DeckName: "tokens"}, "ExpandedStack<Readonly<Record<string, never>>, Readonly<Record<string, never>>>"},
		{FieldInfo{Type: "TypeStack"}, "ExpandedStack"},
		{FieldInfo{Type: "TypeBoard"}, "ExpandedBoard"},
		{FieldInfo{Type: "TypeEnum", EnumName: "phase"}, "PhaseValue"},
		{FieldInfo{Type: "TypeBool"}, "boolean"},
		// Deck with both static and dynamic fields
		{FieldInfo{Type: "TypeStack", DeckName: "dice"}, "ExpandedStack<DiceComponentValues, DiceDynamicComponentValues>"},
		// Deck with dynamic fields only
		{FieldInfo{Type: "TypeStack", DeckName: "pieces"}, "ExpandedStack<Readonly<Record<string, never>>, PiecesDynamicComponentValues>"},
	}

	for _, tc := range tests {
		result := stateFieldTypeToTS(tc.field, decks, enums)
		if result != tc.expected {
			t.Errorf("stateFieldTypeToTS(%+v) = %q, want %q", tc.field, result, tc.expected)
		}
	}
}

func TestGenerateTypeScript(t *testing.T) {
	result := TypeResult{
		PackageName: "testgame",
		ImportPath:  "github.com/test/testgame",
		GameFields: []FieldInfo{
			{Name: "CurrentPlayer", Type: "TypePlayerIndex"},
			{Name: "DrawStack", Type: "TypeStack", DeckName: "cards"},
		},
		PlayerFields: []FieldInfo{
			{Name: "Hand", Type: "TypeStack", DeckName: "cards"},
			{Name: "Score", Type: "TypeInt"},
		},
		Decks: []DeckInfo{
			{Name: "cards", Fields: []FieldInfo{
				{Name: "Rank", Type: "TypeString"},
				{Name: "Suit", Type: "TypeString"},
			}},
		},
		Enums: []EnumInfo{
			{Name: "phase", Values: enumValuesForTest("Setup", "Playing")},
		},
		Constants: []ConstantInfo{
			{Name: "numCards", Kind: "number", Value: "9"},
			{Name: "friendly", Kind: "boolean", Value: "true"},
			{Name: "display-label", Kind: "string", Value: "Cards \"left\""},
		},
		GameComputedFields: []FieldInfo{
			{Name: "cards-done", Type: "TypeBool"},
		},
		PlayerComputedFields: []FieldInfo{
			{Name: "HandValue", Type: "TypeInt"},
			{Name: "Mood", Type: "TypeEnum", EnumName: "phase"},
		},
	}

	ts := GenerateTypeScript(result)

	// Check header
	if !strings.Contains(ts, "Auto-generated by boardgame-util") {
		t.Error("missing header")
	}

	// Check imports
	if !strings.Contains(ts, "import type { CatalogComponent, ExpandedStack, FullGameState }") {
		t.Errorf("wrong imports, got:\n%s", ts)
	}

	// Check enum
	if !strings.Contains(ts, `export type PhaseValue = "Setup" | "Playing";`) {
		t.Error("missing or wrong PhaseValue enum")
	}

	for _, want := range []string{
		`export interface GameConstants {`,
		`readonly "numCards": 9;`,
		`readonly "friendly": true;`,
		`readonly "display-label": "Cards \"left\"";`,
	} {
		if !strings.Contains(ts, want) {
			t.Errorf("missing generated constant %q:\n%s", want, ts)
		}
	}

	for _, want := range []string{
		`export interface GameComputed {`,
		`export interface GameEnums {`,
		`readonly "cards-done": boolean;`,
		`export interface PlayerComputed {`,
		`readonly Color: string;`,
		`readonly "HandValue": number;`,
		`readonly "Mood": PhaseValue;`,
	} {
		if !strings.Contains(ts, want) {
			t.Errorf("missing generated computed contract %q:\n%s", want, ts)
		}
	}
	if strings.Contains(ts, "GameComputed extends Readonly<Record<string, unknown>>") ||
		strings.Contains(ts, "PlayerComputed extends Readonly<Record<string, unknown>>") {
		t.Fatal("computed contracts must not advertise dynamic keys")
	}

	// Check component values interface
	if !strings.Contains(ts, "export interface CardsComponentValues {") {
		t.Error("missing CardsComponentValues interface")
	}

	// Check GameState
	if !strings.Contains(ts, "readonly DrawStack: ExpandedStack<CardsComponentValues, Readonly<Record<string, never>>>;") {
		t.Error("missing typed DrawStack in GameState")
	}

	// Check PlayerState
	if !strings.Contains(ts, "readonly Hand: ExpandedStack<CardsComponentValues, Readonly<Record<string, never>>>;") {
		t.Error("missing typed Hand in PlayerState")
	}
	if !strings.Contains(ts, "readonly Score: number;") {
		t.Error("missing Score in PlayerState")
	}

	// Check Computed field
	if !strings.Contains(ts, "readonly Computed?: GameComputed;") || !strings.Contains(ts, "readonly Computed?: PlayerComputed;") {
		t.Error("missing Computed field")
	}

	// Check State type alias
	if !strings.Contains(ts, "export type State = FullGameState<GameState, PlayerState, GameComputed, PlayerComputed, DynamicComponentValues>;") {
		t.Error("missing State type alias")
	}
}

// climateResult mirrors a real game enum: declared with iota, so key order is
// declaration order, and deliberately neither alphabetical nor reverse
// alphabetical.
func climateResult() TypeResult {
	return TypeResult{Enums: []EnumInfo{{Name: "climate", Values: []EnumValueInfo{
		{Key: 0, Value: "Unknown", Label: "Unknown", IsDefault: true},
		{Key: 1, Value: "Ice Age", Label: "Ice Age"},
		{Key: 2, Value: "Freezing", Label: "Freezing"},
		{Key: 3, Value: "Cold", Label: "Cold"},
		{Key: 4, Value: "Temperate", Label: "Temperate"},
		{Key: 5, Value: "Warm", Label: "Warm"},
		{Key: 6, Value: "Scorching", Label: "Scorching"},
	}}}}
}

// TestGenerateTypeScriptEmitsEnumValuesInDeclaredOrder asserts the whole list
// as one contiguous block rather than checking for each entry separately. A
// per-entry check would pass for a reordered or partially-emitted list; this
// one fails if any value is missing, extra, or out of place.
func TestGenerateTypeScriptEmitsEnumValuesInDeclaredOrder(t *testing.T) {

	ts := GenerateTypeScript(climateResult())

	want := `export const ClimateValues = [
  { Key: 0, Value: "Unknown", Label: "Unknown", IsDefault: true },
  { Key: 1, Value: "Ice Age", Label: "Ice Age", IsDefault: false },
  { Key: 2, Value: "Freezing", Label: "Freezing", IsDefault: false },
  { Key: 3, Value: "Cold", Label: "Cold", IsDefault: false },
  { Key: 4, Value: "Temperate", Label: "Temperate", IsDefault: false },
  { Key: 5, Value: "Warm", Label: "Warm", IsDefault: false },
  { Key: 6, Value: "Scorching", Label: "Scorching", IsDefault: false },
] as const satisfies readonly EnumValueInfo<ClimateValue>[];`

	if !strings.Contains(ts, want) {
		t.Fatalf("generated enum metadata is missing, empty, or misordered.\nwant block:\n%s\ngot:\n%s", want, ts)
	}

	for _, wantFragment := range []string{
		"export interface EnumValueInfo<V extends string = string> {",
		"readonly Key: number;",
		"readonly Value: V;",
		"readonly Label: string;",
		"readonly IsDefault: boolean;",
		"export const ClimateValueInfo: Readonly<Record<ClimateValue, EnumValueInfo<ClimateValue>>> =",
		"Object.fromEntries(ClimateValues.map((value) => [value.Value, value])) as Readonly<Record<ClimateValue, EnumValueInfo<ClimateValue>>>;",
	} {
		if !strings.Contains(ts, wantFragment) {
			t.Errorf("missing %q:\n%s", wantFragment, ts)
		}
	}

	// The union and the ordered list must agree, so a renderer that indexes
	// the lookup with a union member always finds an entry.
	if !strings.Contains(ts, `export type ClimateValue = "Unknown" | "Ice Age" | "Freezing" | "Cold" | "Temperate" | "Warm" | "Scorching";`) {
		t.Errorf("union does not match the ordered list:\n%s", ts)
	}
}

// TestGenerateTypeScriptMarksTheDefaultTheExtractorFound pins the flag to the
// value the extractor marked rather than to a position in the list. A TreeEnum's
// DefaultValue is its first leaf, which need not be the lowest leaf key, so a
// generator that flagged "the first entry" instead would be wrong for real
// enums while still passing any fixture whose default happens to come first.
func TestGenerateTypeScriptMarksTheDefaultTheExtractorFound(t *testing.T) {

	ts := GenerateTypeScript(TypeResult{Enums: []EnumInfo{{Name: "climate", Values: []EnumValueInfo{
		{Key: 0, Value: "Cold", Label: "Cold"},
		{Key: 1, Value: "Ice Age", Label: "Ice Age", IsDefault: true},
		{Key: 2, Value: "Warm", Label: "Warm"},
	}}}})

	// Asserted as one block: the flag has to be present on every entry, true on
	// exactly the marked one, and false on the rest. An entry missing it would
	// not even compile downstream, because `as const` gives each entry its own
	// literal type and a key on only some of them cannot be read off the union.
	want := `export const ClimateValues = [
  { Key: 0, Value: "Cold", Label: "Cold", IsDefault: false },
  { Key: 1, Value: "Ice Age", Label: "Ice Age", IsDefault: true },
  { Key: 2, Value: "Warm", Label: "Warm", IsDefault: false },
] as const satisfies readonly EnumValueInfo<ClimateValue>[];`

	if !strings.Contains(ts, want) {
		t.Fatalf("default flag is missing, on the wrong value, or not on every value.\nwant block:\n%s\ngot:\n%s", want, ts)
	}

	if !strings.Contains(ts, `export const ClimateValueDefault = "Ice Age" satisfies ClimateValue;`) {
		t.Errorf("default value is not exposed directly, or names the wrong value:\n%s", ts)
	}
}

// TestGenerateTypeScriptEmitsDefaultForTheShippedShape checks the ordinary case
// alongside the awkward one above, so the two together fail for both a default
// stuck at index zero and one derived from the wrong entry.
func TestGenerateTypeScriptEmitsDefaultForTheShippedShape(t *testing.T) {

	ts := GenerateTypeScript(climateResult())

	if !strings.Contains(ts, `export const ClimateValueDefault = "Unknown" satisfies ClimateValue;`) {
		t.Errorf("missing ClimateValueDefault:\n%s", ts)
	}
}

// TestGenerateTypeScriptEmitsValueNames covers the by-name constants, including
// the multi-word value that has no identifier-style form and so can only be
// reached with brackets.
func TestGenerateTypeScriptEmitsValueNames(t *testing.T) {

	ts := GenerateTypeScript(climateResult())

	want := `export const ClimateValueName = {
  "Unknown": "Unknown",
  "Ice Age": "Ice Age",
  "Freezing": "Freezing",
  "Cold": "Cold",
  "Temperate": "Temperate",
  "Warm": "Warm",
  "Scorching": "Scorching",
} as const satisfies Readonly<Record<ClimateValue, ClimateValue>>;`

	if !strings.Contains(ts, want) {
		t.Fatalf("by-name constants are missing or incomplete.\nwant block:\n%s\ngot:\n%s", want, ts)
	}

	// The caveat is the whole reason a game cannot assume a dotted form exists.
	if !strings.Contains(ts, `bracket access, e.g. ClimateValueName["Two Words"]`) {
		t.Errorf("generated docs do not warn that non-identifier values need bracket access:\n%s", ts)
	}
}

// TestGenerateTypeScriptSkipsDefaultAndNamesForValuelessEnum: an enum whose
// values could not be resolved generates a `string` union, so there is no value
// to name and no default to point at. Emitting either would be a lie.
func TestGenerateTypeScriptSkipsDefaultAndNamesForValuelessEnum(t *testing.T) {

	ts := GenerateTypeScript(TypeResult{Enums: []EnumInfo{{Name: "climate"}}})

	for _, unwanted := range []string{"ClimateValueDefault", "ClimateValueName"} {
		if strings.Contains(ts, unwanted) {
			t.Errorf("emitted %q for an enum with no values:\n%s", unwanted, ts)
		}
	}
}

// TestGenerateTypeScriptOmitsAbsentPresentation is the "pays nothing" half of
// the contract: an enum that attached no presentation must not emit optional
// keys at all, not even empty ones.
func TestGenerateTypeScriptOmitsAbsentPresentation(t *testing.T) {

	ts := GenerateTypeScript(climateResult())

	for _, unwanted := range []string{"Description:", "CSSColor:", "Art:", "new URL("} {
		if strings.Contains(ts, unwanted) {
			t.Errorf("emitted %q for an enum with no presentation:\n%s", unwanted, ts)
		}
	}
}

func TestGenerateTypeScriptEmitsPresentation(t *testing.T) {

	ts := GenerateTypeScript(TypeResult{Enums: []EnumInfo{{Name: "card", Values: []EnumValueInfo{
		{Key: 0, Value: "Unknown", Label: "Unknown", IsDefault: true},
		{Key: 1, Value: "Guard", Label: "Guard", Description: `Guess a "hand"`, CSSColor: "#c62828"},
		{Key: 2, Value: "Princess", Label: "The Princess", Art: "assets/princess.jpg"},
	}}}})

	want := `export const CardValues = [
  { Key: 0, Value: "Unknown", Label: "Unknown", IsDefault: true },
  { Key: 1, Value: "Guard", Label: "Guard", IsDefault: false, Description: "Guess a \"hand\"", CSSColor: "#c62828" },
  { Key: 2, Value: "Princess", Label: "The Princess", IsDefault: false, Art: new URL("assets/princess.jpg", import.meta.url).href },
] as const satisfies readonly EnumValueInfo<CardValue>[];`

	if !strings.Contains(ts, want) {
		t.Fatalf("presentation not emitted as expected.\nwant block:\n%s\ngot:\n%s", want, ts)
	}
}

func TestGenerateTypeScriptWithoutEnumsEmitsNoMetadata(t *testing.T) {

	ts := GenerateTypeScript(TypeResult{})

	for _, unwanted := range []string{"EnumValueInfo", "Object.fromEntries", "ValueDefault", "ValueName"} {
		if strings.Contains(ts, unwanted) {
			t.Errorf("emitted %q for a game with no enums:\n%s", unwanted, ts)
		}
	}
}

func TestGenerateTypeScriptEmitsHonestComponentCatalog(t *testing.T) {
	ts := GenerateTypeScript(TypeResult{Decks: []DeckInfo{
		{Name: "cards", Fields: []FieldInfo{{Name: "Suit", Type: "TypeString"}}},
		{Name: "tokens", DynamicFields: []FieldInfo{{Name: "Active", Type: "TypeBool"}}},
		{Name: "markers"},
	}})
	for _, want := range []string{
		`readonly "cards": readonly CatalogComponent<CardsComponentValues>[];`,
		`readonly "tokens": readonly CatalogComponent<Readonly<Record<string, never>>>[];`,
		`readonly "markers": readonly CatalogComponent<Readonly<Record<string, never>>>[];`,
		`readonly "tokens": readonly TokensDynamicComponentValues[];`,
	} {
		if !strings.Contains(ts, want) {
			t.Errorf("missing %q:\n%s", want, ts)
		}
	}
}

func TestGenerateTypeScriptDoesNotRedeclareFrameworkComputedOverride(t *testing.T) {
	ts := GenerateTypeScript(TypeResult{PlayerComputedFields: []FieldInfo{
		{Name: "Color", Type: "TypeString"},
	}})
	if count := strings.Count(ts, "readonly Color: string;"); count != 1 {
		t.Fatalf("framework Color declarations = %d, want one:\n%s", count, ts)
	}
}

func TestGenerateRendererTypeScriptBindsCompleteContractAndExactRegistration(t *testing.T) {
	ts := GenerateRendererTypeScript("sample")
	for _, want := range []string{
		"export interface GameClientContract",
		"export abstract class GameRenderer extends BoardgameBaseGameRenderer<",
		"export abstract class TableRenderer extends BoardgameTableViewBase<",
		"export abstract class HandRenderer extends BoardgameHandViewBase<",
		"export abstract class PlayerInfoRenderer extends BoardgameBasePlayerInfoRenderer<",
		"export function registerGameRenderer",
		"export function registerTableRenderer",
		"export function registerHandRenderer",
		"export function registerPlayerInfoRenderer",
		"function registerRenderer<Base extends HTMLElement>(",
		"'[sample] cannot register '",
		"must extend the generated ' + expectedBaseName + ' base",
		"customElements.define(tagName, constructor);",
		"'boardgame-render-game-sample', 'game', 'GameRenderer', GameRenderer, constructor",
		"protected override readonly moveInputSchema = moveInputSchema;",
		"protected override readonly moveChoiceProjectionSchema = moveChoiceProjectionSchema;",
		"protected override readonly moveChoiceProjectionSchemaFingerprint = moveChoiceProjectionSchemaFingerprint;",
		"readonly Components: ComponentCatalog;",
		"readonly Constants: GameConstants;",
		"readonly Enums: GameEnums;",
		"readonly MoveChoiceProjections: MoveChoiceProjections;",
		"readonly RendererTag:",
		"'boardgame-render-game-sample-table'",
		"'boardgame-render-game-sample-hand'",
		"GameClientContract['State']",
		"GameClientContract['Constants']",
		"GameClientContract['Enums']",
		"GameClientContract['MoveChoiceProjections']",
	} {
		if !strings.Contains(ts, want) {
			t.Errorf("missing %q:\n%s", want, ts)
		}
	}
	if strings.Contains(ts, "customElements.define('boardgame-render-game-sample', GameRenderer)") {
		t.Fatal("generated module registered an abstract renderer at module evaluation")
	}
}

func TestGenerateTypeScriptWithBoard(t *testing.T) {
	result := TypeResult{
		PackageName: "boardgame",
		GameFields: []FieldInfo{
			{Name: "Spaces", Type: "TypeBoard", DeckName: "tokens"},
		},
		Decks: []DeckInfo{
			{Name: "tokens", Fields: []FieldInfo{{Name: "Color", Type: "TypeString"}}},
		},
	}

	ts := GenerateTypeScript(result)

	if !strings.Contains(ts, "CatalogComponent, ExpandedBoard, FullGameState") {
		t.Errorf("missing Board import, got:\n%s", ts)
	}
	if !strings.Contains(ts, "Spaces: ExpandedBoard<TokensComponentValues, Readonly<Record<string, never>>>;") {
		t.Errorf("Board field not typed correctly, got:\n%s", ts)
	}
}

func TestGenerateTypeScriptWithTimer(t *testing.T) {
	result := TypeResult{
		PackageName: "timergame",
		GameFields: []FieldInfo{
			{Name: "MyTimer", Type: "TypeTimer"},
		},
	}

	ts := GenerateTypeScript(result)

	if !strings.Contains(ts, "ExpandedTimer, FullGameState") {
		t.Errorf("missing ExpandedTimer import, got:\n%s", ts)
	}
	if !strings.Contains(ts, "MyTimer: ExpandedTimer;") {
		t.Error("Timer field not typed correctly")
	}
}

func TestGenerateTypeScriptEmpty(t *testing.T) {
	result := TypeResult{
		PackageName: "empty",
	}

	ts := GenerateTypeScript(result)

	if !strings.Contains(ts, "export interface GameState {") {
		t.Error("missing GameState interface")
	}
	if !strings.Contains(ts, "export interface PlayerState {") {
		t.Error("missing PlayerState interface")
	}
	if !strings.Contains(ts, "export type GameConstants = Readonly<Record<string, never>>;") {
		t.Error("missing closed empty constants contract")
	}
	// Should only import FullGameState
	if !strings.Contains(ts, "import type { FullGameState }") {
		t.Errorf("wrong imports for empty game, got:\n%s", ts)
	}
}

func TestGenerateTypeScriptWithDynamicValues(t *testing.T) {
	result := TypeResult{
		PackageName: "checkerslike",
		GameFields: []FieldInfo{
			{Name: "Tokens", Type: "TypeStack", DeckName: "tokens"},
		},
		Decks: []DeckInfo{
			{
				Name:   "tokens",
				Fields: []FieldInfo{{Name: "Color", Type: "TypeString"}},
				DynamicFields: []FieldInfo{
					{Name: "Crowned", Type: "TypeBool"},
					{Name: "MoveCount", Type: "TypeInt"},
				},
			},
		},
	}

	ts := GenerateTypeScript(result)

	// Check static component values interface
	if !strings.Contains(ts, "export interface TokensComponentValues {") {
		t.Errorf("missing TokensComponentValues interface, got:\n%s", ts)
	}

	// Check dynamic component values interface
	if !strings.Contains(ts, "export interface TokensDynamicComponentValues {") {
		t.Errorf("missing TokensDynamicComponentValues interface, got:\n%s", ts)
	}
	if !strings.Contains(ts, "Crowned: boolean;") {
		t.Errorf("missing Crowned field in dynamic interface, got:\n%s", ts)
	}
	if !strings.Contains(ts, "MoveCount: number;") {
		t.Errorf("missing MoveCount field in dynamic interface, got:\n%s", ts)
	}

	// Check two-param ExpandedStack
	if !strings.Contains(ts, "Tokens: ExpandedStack<TokensComponentValues, TokensDynamicComponentValues>;") {
		t.Errorf("missing two-param ExpandedStack, got:\n%s", ts)
	}
}

func TestGenerateTypeScriptWithDynamicOnly(t *testing.T) {
	result := TypeResult{
		PackageName: "dynonly",
		GameFields: []FieldInfo{
			{Name: "Pieces", Type: "TypeStack", DeckName: "pieces"},
		},
		Decks: []DeckInfo{
			{
				Name:          "pieces",
				DynamicFields: []FieldInfo{{Name: "Active", Type: "TypeBool"}},
			},
		},
	}

	ts := GenerateTypeScript(result)

	// Should NOT have static component values interface
	if strings.Contains(ts, "PiecesComponentValues") {
		t.Errorf("should not have PiecesComponentValues, got:\n%s", ts)
	}

	// Should have dynamic component values interface
	if !strings.Contains(ts, "export interface PiecesDynamicComponentValues {") {
		t.Errorf("missing PiecesDynamicComponentValues, got:\n%s", ts)
	}

	// No static values means the visible Values object is exactly empty.
	if !strings.Contains(ts, "Pieces: ExpandedStack<Readonly<Record<string, never>>, PiecesDynamicComponentValues>;") {
		t.Errorf("missing dynamic-only ExpandedStack, got:\n%s", ts)
	}
}

func TestGenerateTypeScriptWithDynamicRawStack(t *testing.T) {
	result := TypeResult{
		PackageName: "stackdyn",
		GameFields: []FieldInfo{
			{Name: "Bag", Type: "TypeStack", DeckName: "items"},
		},
		Decks: []DeckInfo{
			{
				Name:          "items",
				Fields:        []FieldInfo{{Name: "Name", Type: "TypeString"}},
				DynamicFields: []FieldInfo{{Name: "Container", Type: "TypeStack"}},
			},
		},
	}

	ts := GenerateTypeScript(result)

	// Should import RawStack
	if !strings.Contains(ts, "RawStack") {
		t.Errorf("missing RawStack import, got:\n%s", ts)
	}
	// Dynamic field should use RawStack
	if !strings.Contains(ts, "Container: RawStack;") {
		t.Errorf("dynamic stack field not typed as RawStack, got:\n%s", ts)
	}
}
