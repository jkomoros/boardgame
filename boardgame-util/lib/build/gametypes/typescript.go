package gametypes

import (
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
	needsExpandedStack := false
	needsExpandedTimer := false

	for _, f := range result.GameFields {
		checkFieldImports(f, &needsExpandedStack, &needsExpandedTimer, &needsBoard)
	}
	for _, f := range result.PlayerFields {
		checkFieldImports(f, &needsExpandedStack, &needsExpandedTimer, &needsBoard)
	}

	// Build import line (alphabetical order)
	var imports []string
	if needsBoard {
		imports = append(imports, "Board")
	}
	if needsExpandedStack {
		imports = append(imports, "ExpandedStack")
	}
	if needsExpandedTimer {
		imports = append(imports, "ExpandedTimer")
	}
	imports = append(imports, "FullGameState")

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
			b.WriteString("  ")
			b.WriteString(f.Name)
			b.WriteString(": ")
			b.WriteString(baseFieldTypeToTS(f, result.Enums))
			b.WriteString(";\n")
		}
		b.WriteString("}\n\n")
	}

	// Generate GameState interface
	b.WriteString("export interface GameState {\n")
	for _, f := range result.GameFields {
		b.WriteString("  ")
		b.WriteString(f.Name)
		b.WriteString(": ")
		b.WriteString(stateFieldTypeToTS(f, result.Decks, result.Enums))
		b.WriteString(";\n")
	}
	b.WriteString("  Computed?: Record<string, unknown>;\n")
	b.WriteString("}\n\n")

	// Generate PlayerState interface
	b.WriteString("export interface PlayerState {\n")
	for _, f := range result.PlayerFields {
		b.WriteString("  ")
		b.WriteString(f.Name)
		b.WriteString(": ")
		b.WriteString(stateFieldTypeToTS(f, result.Decks, result.Enums))
		b.WriteString(";\n")
	}
	b.WriteString("  Computed?: Record<string, unknown>;\n")
	b.WriteString("}\n\n")

	// Generate State type alias
	b.WriteString("export type State = FullGameState<GameState, PlayerState>;\n")

	return b.String()
}

func checkFieldImports(f FieldInfo, needsStack, needsTimer, needsBoard *bool) {
	switch f.Type {
	case "TypeStack":
		*needsStack = true
	case "TypeBoard":
		*needsBoard = true
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
		return "number[]"
	case "TypeBoolSlice":
		return "boolean[]"
	case "TypeStringSlice":
		return "string[]"
	case "TypePlayerIndexSlice":
		return "number[]"
	case "TypeEnum":
		if f.EnumName != "" {
			return toPascalCase(f.EnumName) + "Value"
		}
		return "string"
	case "TypeEnumSlice":
		if f.EnumName != "" {
			return toPascalCase(f.EnumName) + "Value[]"
		}
		return "string[]"
	default:
		return "unknown"
	}
}

// stateFieldTypeToTS maps a state field (with deck/enum context) to a TypeScript type string.
// Extends baseFieldTypeToTS with stack, board, and timer support.
func stateFieldTypeToTS(f FieldInfo, decks []DeckInfo, enums []EnumInfo) string {
	switch f.Type {
	case "TypeTimer":
		return "ExpandedTimer"
	case "TypeStack":
		if f.DeckName != "" {
			// Check if this deck has component fields
			for _, d := range decks {
				if d.Name == f.DeckName && len(d.Fields) > 0 {
					return "ExpandedStack<" + toPascalCase(f.DeckName) + "ComponentValues>"
				}
			}
		}
		return "ExpandedStack"
	case "TypeBoard":
		return "Board"
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
