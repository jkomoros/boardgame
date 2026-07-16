package gametypes

import (
	"strings"
	"testing"
)

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
		{Name: "color", Values: []string{"Red", "Blue"}},
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
		{FieldInfo{Type: "TypeSomethingUnknown"}, "unknown"},
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
		{Name: "color", Values: []string{"Red", "Blue"}},
	}

	tests := []struct {
		field    FieldInfo
		expected string
	}{
		{FieldInfo{Type: "TypeBool"}, "boolean"},
		{FieldInfo{Type: "TypeInt"}, "number"},
		{FieldInfo{Type: "TypeString"}, "string"},
		{FieldInfo{Type: "TypeStack"}, "RawStack"},
		{FieldInfo{Type: "TypeTimer"}, "Record<string, unknown>"},
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
		{Name: "phase", Values: []string{"Setup", "Playing"}},
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
			{Name: "phase", Values: []string{"Setup", "Playing"}},
		},
		Constants: []ConstantInfo{
			{Name: "numCards", Kind: "number", Value: "9"},
			{Name: "friendly", Kind: "boolean", Value: "true"},
			{Name: "display-label", Kind: "string", Value: "Cards \"left\""},
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
		"readonly Components: ComponentCatalog;",
		"readonly Constants: GameConstants;",
		"readonly RendererTag:",
		"'boardgame-render-game-sample-table'",
		"'boardgame-render-game-sample-hand'",
		"GameClientContract['State']",
		"GameClientContract['Constants']",
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
