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
	needsExpandedStack := false
	needsExpandedTimer := false

	for _, f := range result.GameFields {
		checkFieldImports(f, &needsExpandedStack, &needsExpandedTimer)
	}
	for _, f := range result.PlayerFields {
		checkFieldImports(f, &needsExpandedStack, &needsExpandedTimer)
	}

	// Build import line
	var imports []string
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
		b.WriteString("export type ")
		b.WriteString(toPascalCase(e.Name))
		b.WriteString("Value = ")
		if len(e.Values) == 0 {
			b.WriteString("string")
		} else {
			for i, v := range e.Values {
				if i > 0 {
					b.WriteString(" | ")
				}
				b.WriteString("\"")
				b.WriteString(v)
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
		interfaceName := toPascalCase(deck.Name) + "ComponentValues"
		b.WriteString("export interface ")
		b.WriteString(interfaceName)
		b.WriteString(" {\n")
		for _, f := range deck.Fields {
			b.WriteString("  ")
			b.WriteString(f.Name)
			b.WriteString(": ")
			b.WriteString(fieldTypeToTS(f, result.Enums))
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

func checkFieldImports(f FieldInfo, needsStack, needsTimer *bool) {
	switch f.Type {
	case "TypeStack", "TypeBoard":
		*needsStack = true
	case "TypeTimer":
		*needsTimer = true
	}
}

// stateFieldTypeToTS maps a state field (with deck/enum context) to a TypeScript type string.
func stateFieldTypeToTS(f FieldInfo, decks []DeckInfo, enums []EnumInfo) string {
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
	case "TypeTimer":
		return "ExpandedTimer"
	case "TypeStack", "TypeBoard":
		if f.DeckName != "" {
			// Check if this deck has component fields
			for _, d := range decks {
				if d.Name == f.DeckName && len(d.Fields) > 0 {
					return "ExpandedStack<" + toPascalCase(f.DeckName) + "ComponentValues>"
				}
			}
		}
		return "ExpandedStack"
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

// fieldTypeToTS maps a component value field to a TypeScript type string.
// Component value fields don't have deck/enum context in their tags.
func fieldTypeToTS(f FieldInfo, enums []EnumInfo) string {
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

// toPascalCase converts a name like "playing cards" or "cards" to "PlayingCards" or "Cards".
func toPascalCase(name string) string {
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
