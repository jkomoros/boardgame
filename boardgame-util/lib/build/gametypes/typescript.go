package gametypes

import (
	"sort"
	"strings"
	"unicode"
)

// GenerateTypeScript produces the contents of a _types.ts file given a
// TypeResult.
func GenerateTypeScript(result TypeResult) string {

	var b strings.Builder

	b.WriteString(typeScriptHeader)

	// Determine which imports we need
	needsBoard := false
	needsExpandedBoard := false
	needsExpandedStack := false
	needsExpandedTimer := false
	needsRawStack := false
	needsCatalogComponent := len(result.Decks) > 0

	for _, f := range result.GameFields {
		checkFieldImports(f, &needsExpandedStack, &needsExpandedTimer, &needsExpandedBoard)
	}
	for _, f := range result.PlayerFields {
		checkFieldImports(f, &needsExpandedStack, &needsExpandedTimer, &needsExpandedBoard)
	}
	// Check dynamic fields for RawStack imports
	for _, deck := range result.Decks {
		for _, f := range deck.DynamicFields {
			if f.Type == "TypeStack" {
				needsRawStack = true
			}
			if f.Type == "TypeBoard" {
				needsBoard = true
			}
		}
	}

	// Build import line (alphabetical order)
	var imports []string
	if needsCatalogComponent {
		imports = append(imports, "CatalogComponent")
	}
	if needsBoard {
		imports = append(imports, "Board")
	}
	if needsExpandedStack {
		imports = append(imports, "ExpandedStack")
	}
	if needsExpandedBoard {
		imports = append(imports, "ExpandedBoard")
	}
	if needsExpandedTimer {
		imports = append(imports, "ExpandedTimer")
	}
	imports = append(imports, "FullGameState")
	if needsRawStack {
		imports = append(imports, "RawStack")
	}
	sort.Strings(imports)

	b.WriteString("import type { ")
	b.WriteString(strings.Join(imports, ", "))
	b.WriteString(" } from '../../src/types/boardgame-types.js';\n")
	b.WriteString("\n")

	// Generate enum types
	for _, e := range result.Enums {
		name := toPascalCase(e.Name)
		if name == "" {
			continue
		}
		b.WriteString("export type ")
		b.WriteString(name)
		b.WriteString("Value = ")
		if len(e.Values) == 0 {
			b.WriteString("string")
		} else {
			for i, v := range e.Values {
				if i > 0 {
					b.WriteString(" | ")
				}
				b.WriteString("\"")
				b.WriteString(escapeForTS(v))
				b.WriteString("\"")
			}
		}
		b.WriteString(";\n\n")
	}

	// Generate component value interfaces (one per deck that has fields)
	for _, deck := range result.Decks {
		if len(deck.Fields) == 0 {
			continue
		}
		name := toPascalCase(deck.Name)
		if name == "" {
			continue
		}
		interfaceName := name + "ComponentValues"
		b.WriteString("export interface ")
		b.WriteString(interfaceName)
		b.WriteString(" {\n")
		for _, f := range deck.Fields {
			b.WriteString("  readonly ")
			b.WriteString(f.Name)
			b.WriteString(": ")
			b.WriteString(baseFieldTypeToTS(f, result.Enums))
			b.WriteString(";\n")
		}
		b.WriteString("}\n\n")
	}

	if len(result.Decks) > 0 {
		b.WriteString("export interface ComponentCatalog {\n")
		for _, deck := range result.Decks {
			b.WriteString("  readonly ")
			b.WriteString(tsQuoted(deck.Name))
			b.WriteString(": readonly ")
			b.WriteString(catalogComponentTypeForDeck(deck))
			b.WriteString("[];\n")
		}
		b.WriteString("}\n\n")
	} else {
		b.WriteString("export type ComponentCatalog = Readonly<Record<string, never>>;\n\n")
	}

	// Generate dynamic component value interfaces (one per deck that has dynamic fields)
	for _, deck := range result.Decks {
		if len(deck.DynamicFields) == 0 {
			continue
		}
		name := toPascalCase(deck.Name)
		if name == "" {
			continue
		}
		interfaceName := name + "DynamicComponentValues"
		b.WriteString("export interface ")
		b.WriteString(interfaceName)
		b.WriteString(" {\n")
		for _, f := range deck.DynamicFields {
			b.WriteString("  readonly ")
			b.WriteString(f.Name)
			b.WriteString(": ")
			b.WriteString(dynamicFieldTypeToTS(f, result.Enums))
			b.WriteString(";\n")
		}
		b.WriteString("}\n\n")
	}

	// Components in the serialized state contain per-component dynamic
	// values, keyed by deck. This is deliberately separate from the resolved
	// component catalogue used by expanded stacks.
	hasDynamicComponents := false
	for _, deck := range result.Decks {
		if len(deck.DynamicFields) > 0 {
			hasDynamicComponents = true
			break
		}
	}
	if hasDynamicComponents {
		b.WriteString("export interface DynamicComponentValues {\n")
		for _, deck := range result.Decks {
			if len(deck.DynamicFields) == 0 {
				continue
			}
			b.WriteString("  readonly ")
			b.WriteString(tsQuoted(deck.Name))
			b.WriteString(": readonly ")
			b.WriteString(toPascalCase(deck.Name))
			b.WriteString("DynamicComponentValues[];\n")
		}
		b.WriteString("}\n\n")
	} else {
		b.WriteString("export type DynamicComponentValues = Readonly<Record<string, never>>;\n\n")
	}

	if len(result.Constants) > 0 {
		b.WriteString("export interface GameConstants {\n")
		for _, constant := range result.Constants {
			b.WriteString("  readonly ")
			b.WriteString(tsQuoted(constant.Name))
			b.WriteString(": ")
			switch constant.Kind {
			case "number", "boolean":
				b.WriteString(constant.Value)
			case "string":
				b.WriteString(tsQuoted(constant.Value))
			default:
				b.WriteString("never")
			}
			b.WriteString(";\n")
		}
		b.WriteString("}\n\n")
	} else {
		b.WriteString("export type GameConstants = Readonly<Record<string, never>>;\n\n")
	}

	b.WriteString("export interface ComputedEnumOption {\n")
	b.WriteString("  readonly Key: number;\n")
	b.WriteString("  readonly Name: string;\n")
	b.WriteString("  readonly CSSColor?: string;\n")
	b.WriteString("}\n\n")

	b.WriteString("export interface GameComputed {\n")
	b.WriteString("  readonly PlayerOrder?: readonly number[];\n")
	b.WriteString("  readonly AvailableTeams?: readonly ComputedEnumOption[];\n")
	b.WriteString("  readonly AvailableRoles?: readonly ComputedEnumOption[];\n")
	b.WriteString("  readonly AvailableColors?: readonly ComputedEnumOption[];\n")
	b.WriteString("  readonly ReadyToStartError?: string;\n")
	writeComputedFields(&b, result.GameComputedFields, result.Enums, map[string]bool{
		"PlayerOrder": true, "AvailableTeams": true, "AvailableRoles": true,
		"AvailableColors": true, "ReadyToStartError": true,
	})
	b.WriteString("}\n\n")

	b.WriteString("export interface PlayerComputed {\n")
	b.WriteString("  readonly Color: string;\n")
	b.WriteString("  readonly MayBeActive: boolean;\n")
	b.WriteString("  readonly GameScore?: number;\n")
	b.WriteString("  readonly TeamValue?: string;\n")
	b.WriteString("  readonly RoleValue?: string;\n")
	b.WriteString("  readonly ColorValue?: string;\n")
	b.WriteString("  readonly IsGameAdmin?: boolean;\n")
	writeComputedFields(&b, result.PlayerComputedFields, result.Enums, map[string]bool{
		"Color": true, "MayBeActive": true, "GameScore": true,
		"TeamValue": true, "RoleValue": true, "ColorValue": true,
		"IsGameAdmin": true,
	})
	b.WriteString("}\n\n")

	// Generate GameState interface

	b.WriteString("export interface GameState {\n")
	for _, f := range result.GameFields {
		b.WriteString("  readonly ")
		b.WriteString(f.Name)
		b.WriteString(": ")
		b.WriteString(stateFieldTypeToTS(f, result.Decks, result.Enums))
		b.WriteString(";\n")
	}
	b.WriteString("  readonly Computed?: GameComputed;\n")
	b.WriteString("}\n\n")

	// Generate PlayerState interface
	b.WriteString("export interface PlayerState {\n")
	for _, f := range result.PlayerFields {
		b.WriteString("  readonly ")
		b.WriteString(f.Name)
		b.WriteString(": ")
		b.WriteString(stateFieldTypeToTS(f, result.Decks, result.Enums))
		b.WriteString(";\n")
	}
	b.WriteString("  readonly Computed?: PlayerComputed;\n")
	b.WriteString("}\n\n")

	// Generate State type alias
	b.WriteString("export type State = FullGameState<GameState, PlayerState, GameComputed, PlayerComputed, DynamicComponentValues>;\n")

	return b.String()
}

func writeComputedFields(b *strings.Builder, fields []FieldInfo, enums []EnumInfo, frameworkNames map[string]bool) {
	for _, field := range fields {
		if frameworkNames[field.Name] {
			continue
		}
		b.WriteString("  readonly ")
		b.WriteString(tsQuoted(field.Name))
		b.WriteString(": ")
		b.WriteString(baseFieldTypeToTS(field, enums))
		b.WriteString(";\n")
	}
}

// GenerateRendererTypeScript produces the thin, unregistered game-bound base.
// It contains no framework internals and installs the generated runtime schema
// exactly once for every renderer that extends it.
func GenerateRendererTypeScript(gameName string) string {
	source := `/*
 * Auto-generated by boardgame-util. DO NOT EDIT.
 */

import {
  BoardgameBaseGameRenderer,
  BoardgameBasePlayerInfoRenderer,
  BoardgameHandViewBase,
  BoardgameTableViewBase,
} from '../../src/client.js';
import {
  moveInputSchema,
  moveInputSchemaFingerprint,
  type MoveInputs,
} from './_move_args.js';
import type { MoveName } from './_move_names.js';
import type {
  ComponentCatalog,
  DynamicComponentValues,
  GameConstants,
  GameComputed,
  GameState,
  PlayerComputed,
  PlayerState,
  State,
} from './_types.js';

/** Complete compile-time contract for this game's renderer surface. */
export interface GameClientContract {
  readonly State: State;
  readonly GameState: GameState;
  readonly PlayerState: PlayerState;
  readonly GameComputed: GameComputed;
  readonly PlayerComputed: PlayerComputed;
  readonly Components: ComponentCatalog;
  readonly DynamicComponents: DynamicComponentValues;
  readonly Constants: GameConstants;
  readonly MoveName: MoveName;
  readonly MoveInputs: MoveInputs;
  readonly RendererTag:
    | 'boardgame-render-game-__GAME_NAME__'
    | 'boardgame-render-game-__GAME_NAME__-table'
    | 'boardgame-render-game-__GAME_NAME__-hand';
}

/** Extend this class, then register only your concrete renderer element. */
export abstract class GameRenderer extends BoardgameBaseGameRenderer<
  GameClientContract['State'],
  GameClientContract['Components'],
  GameClientContract['MoveName'],
  GameClientContract['MoveInputs'],
  GameClientContract['Constants']
> {
  protected override readonly moveInputSchema = moveInputSchema;
  protected override readonly moveInputSchemaFingerprint = moveInputSchemaFingerprint;
}

/** Extend this class for the shared-screen companion surface. */
export abstract class TableRenderer extends BoardgameTableViewBase<
  GameClientContract['State'],
  GameClientContract['Components'],
  GameClientContract['MoveName'],
  GameClientContract['MoveInputs'],
  GameClientContract['Constants']
> {
  protected override readonly moveInputSchema = moveInputSchema;
  protected override readonly moveInputSchemaFingerprint = moveInputSchemaFingerprint;
}

/** Extend this class for the private per-player companion surface. */
export abstract class HandRenderer extends BoardgameHandViewBase<
  GameClientContract['State'],
  GameClientContract['Components'],
  GameClientContract['MoveName'],
  GameClientContract['MoveInputs'],
  GameClientContract['Constants']
> {
  protected override readonly moveInputSchema = moveInputSchema;
  protected override readonly moveInputSchemaFingerprint = moveInputSchemaFingerprint;
}

/** Extend this class for a strictly typed player-info renderer. */
export abstract class PlayerInfoRenderer extends BoardgameBasePlayerInfoRenderer<
  GameClientContract['State'],
  GameClientContract['PlayerState']
> {}

type RendererConstructor<Base extends HTMLElement> = new () => Base;
type AbstractRendererConstructor<Base extends HTMLElement> = abstract new () => Base;

function rendererConstructorName(constructor: object): string {
  const name = (constructor as { readonly name?: unknown }).name;
  return typeof name === 'string' && name ? name : '(anonymous renderer)';
}

function registerRenderer<Base extends HTMLElement>(
  tagName: string,
  surfaceName: string,
  expectedBaseName: string,
  expectedBase: AbstractRendererConstructor<Base>,
  constructor: RendererConstructor<Base>,
): void {
  const constructorName = rendererConstructorName(constructor);
  if (!(constructor.prototype instanceof expectedBase)) {
    throw new Error(
      '[__GAME_NAME__] ' + surfaceName + ' renderer ' + constructorName +
      ' must extend the generated ' + expectedBaseName + ' base',
    );
  }
  const existing = customElements.get(tagName);
  if (existing) {
    throw new Error(
      '[__GAME_NAME__] cannot register ' + surfaceName + ' renderer ' + constructorName +
      ' as <' + tagName + '>: that tag is already registered by ' + rendererConstructorName(existing),
    );
  }
  customElements.define(tagName, constructor);
}

/** Register the ordinary game surface under its generated exact tag. */
export function registerGameRenderer<T extends RendererConstructor<GameRenderer>>(
  constructor: T,
): void {
  registerRenderer(
    'boardgame-render-game-__GAME_NAME__', 'game', 'GameRenderer', GameRenderer, constructor,
  );
}

/** Register the companion shared-screen surface under its generated exact tag. */
export function registerTableRenderer<T extends RendererConstructor<TableRenderer>>(
  constructor: T,
): void {
  registerRenderer(
    'boardgame-render-game-__GAME_NAME__-table', 'table', 'TableRenderer', TableRenderer, constructor,
  );
}

/** Register the companion private-player surface under its generated exact tag. */
export function registerHandRenderer<T extends RendererConstructor<HandRenderer>>(
  constructor: T,
): void {
  registerRenderer(
    'boardgame-render-game-__GAME_NAME__-hand', 'hand', 'HandRenderer', HandRenderer, constructor,
  );
}

/** Register the player-info surface under its generated exact tag. */
export function registerPlayerInfoRenderer<T extends RendererConstructor<PlayerInfoRenderer>>(
  constructor: T,
): void {
  registerRenderer(
    'boardgame-render-player-info-__GAME_NAME__',
    'player info',
    'PlayerInfoRenderer',
    PlayerInfoRenderer,
    constructor,
  );
}
`
	return strings.ReplaceAll(source, "__GAME_NAME__", escapeForTS(gameName))
}

func catalogComponentTypeForDeck(deck DeckInfo) string {
	staticType, _ := deckValueTypes(deck)
	return "CatalogComponent<" + staticType + ">"
}

func deckValueTypes(deck DeckInfo) (string, string) {
	staticType := "Readonly<Record<string, never>>"
	dynamicType := "Readonly<Record<string, never>>"
	name := toPascalCase(deck.Name)
	if len(deck.Fields) > 0 && name != "" {
		staticType = name + "ComponentValues"
	}
	if len(deck.DynamicFields) > 0 && name != "" {
		dynamicType = name + "DynamicComponentValues"
	}
	return staticType, dynamicType
}

func findDeck(name string, decks []DeckInfo) (DeckInfo, bool) {
	for _, deck := range decks {
		if deck.Name == name {
			return deck, true
		}
	}
	return DeckInfo{}, false
}

func tsQuoted(value string) string {
	return `"` + escapeForTS(value) + `"`
}

func checkFieldImports(f FieldInfo, needsStack, needsTimer, needsExpandedBoard *bool) {
	switch f.Type {
	case "TypeStack":
		*needsStack = true
	case "TypeBoard":
		*needsExpandedBoard = true
	case "TypeTimer":
		*needsTimer = true
	}
}

// baseFieldTypeToTS maps a field to a TypeScript type string for basic types
// (no stack/board/timer support). Used for component value fields.
func baseFieldTypeToTS(f FieldInfo, enums []EnumInfo) string {
	switch f.Type {
	case "TypeBool":
		return "boolean"
	case "TypeInt":
		return "number"
	case "TypeString":
		return "string"
	case "TypePlayerIndex":
		return "number"
	case "TypeIntSlice":
		return "readonly number[]"
	case "TypeBoolSlice":
		return "readonly boolean[]"
	case "TypeStringSlice":
		return "readonly string[]"
	case "TypePlayerIndexSlice":
		return "readonly number[]"
	case "TypeEnum":
		if f.EnumName != "" {
			return toPascalCase(f.EnumName) + "Value"
		}
		return "string"
	case "TypeEnumSlice":
		if f.EnumName != "" {
			return "readonly " + toPascalCase(f.EnumName) + "Value[]"
		}
		return "readonly string[]"
	default:
		return "unknown"
	}
}

// dynamicFieldTypeToTS maps a dynamic component value field to a TypeScript type string.
// Like baseFieldTypeToTS but maps TypeStack to RawStack (dynamic value stacks are not
// expanded by the client selector) and TypeTimer to Record<string, unknown>.
func dynamicFieldTypeToTS(f FieldInfo, enums []EnumInfo) string {
	switch f.Type {
	case "TypeStack":
		return "RawStack"
	case "TypeTimer":
		return "Record<string, unknown>"
	case "TypeBoard":
		return "Board"
	default:
		return baseFieldTypeToTS(f, enums)
	}
}

// stateFieldTypeToTS maps a state field (with deck/enum context) to a TypeScript type string.
// Extends baseFieldTypeToTS with stack, board, and timer support.
func stateFieldTypeToTS(f FieldInfo, decks []DeckInfo, enums []EnumInfo) string {
	switch f.Type {
	case "TypeTimer":
		return "ExpandedTimer"
	case "TypeStack":
		if deck, ok := findDeck(f.DeckName, decks); ok {
			staticType, dynamicType := deckValueTypes(deck)
			return "ExpandedStack<" + staticType + ", " + dynamicType + ">"
		}
		return "ExpandedStack"
	case "TypeBoard":
		if deck, ok := findDeck(f.DeckName, decks); ok {
			staticType, dynamicType := deckValueTypes(deck)
			return "ExpandedBoard<" + staticType + ", " + dynamicType + ">"
		}
		return "ExpandedBoard"
	default:
		return baseFieldTypeToTS(f, enums)
	}
}

// escapeForTS escapes a string for use inside a TypeScript double-quoted string literal.
func escapeForTS(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}

// toPascalCase converts a name like "playing cards" or "playing_cards" to "PlayingCards".
// Splits on whitespace, underscores, and hyphens.
func toPascalCase(name string) string {
	// Replace underscores and hyphens with spaces so Fields splits on them
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")

	words := strings.Fields(name)
	var b strings.Builder

	for _, word := range words {
		var cleaned []rune
		for _, r := range word {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				cleaned = append(cleaned, r)
			}
		}
		if len(cleaned) == 0 {
			continue
		}
		cleaned[0] = unicode.ToUpper(cleaned[0])
		b.WriteString(string(cleaned))
	}

	return b.String()
}

const typeScriptHeader = `/*
 * Auto-generated by boardgame-util. DO NOT EDIT.
 */

`
