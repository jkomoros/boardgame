package moveargs

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/jkomoros/boardgame"
)

// MoveFieldInfo describes a single field on a move struct.
type MoveFieldInfo = boardgame.MoveInputSchemaField

// MoveInfo describes a single move's name and fields.
type MoveInfo = boardgame.MoveInputSchemaMove

// ValidateTypeScriptSchema rejects move names that cannot produce one unique,
// legal TypeScript declaration name. Without this preflight, punctuation-only
// names or punctuation-equivalent names could create malformed declarations or
// silently merge interfaces.
func ValidateTypeScriptSchema(moves []MoveInfo) error {
	seenMoves := make(map[string]bool, len(moves))
	seenSymbols := make(map[string]string, len(moves))
	for _, move := range moves {
		if seenMoves[move.Name] {
			return fmt.Errorf("move name %q appears more than once", move.Name)
		}
		seenMoves[move.Name] = true
		symbol := toPascalCase(move.Name)
		if !validTypeScriptIdentifier(symbol) {
			return fmt.Errorf("move name %q generates invalid TypeScript identifier %q", move.Name, symbol)
		}
		if previous, ok := seenSymbols[symbol]; ok {
			return fmt.Errorf("move names %q and %q both generate TypeScript identifier %q", previous, move.Name, symbol)
		}
		seenSymbols[symbol] = move.Name
		seenFields := make(map[string]bool, len(move.Fields))
		for _, field := range move.Fields {
			if !validTypeScriptIdentifier(field.Name) {
				return fmt.Errorf("move %q field %q is not a valid TypeScript identifier", move.Name, field.Name)
			}
			if seenFields[field.Name] {
				return fmt.Errorf("move %q contains duplicate field %q", move.Name, field.Name)
			}
			seenFields[field.Name] = true
			switch field.Codec {
			case "integer", "player-index", "boolean", "enum", "string":
			default:
				return fmt.Errorf("move %q field %q has unsupported creator codec %q", move.Name, field.Name, field.Codec)
			}
		}
	}
	return nil
}

func validTypeScriptIdentifier(value string) bool {
	runes := []rune(value)
	if len(runes) == 0 || !(unicode.IsLetter(runes[0]) || runes[0] == '_') {
		return false
	}
	for _, r := range runes[1:] {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			return false
		}
	}
	return true
}

// GenerateTypeScript produces the contents of a _move_args.ts file given a
// list of moves with their fields. This provides typed argument interfaces for
// each move and a mapped type that connects move names to their args.
func GenerateTypeScript(moves []MoveInfo) string {
	if err := ValidateTypeScriptSchema(moves); err != nil {
		panic(err)
	}
	if len(moves) == 0 {
		return typeScriptHeader + typeScriptEmptyBody
	}

	var b strings.Builder
	b.WriteString(typeScriptHeader)
	sort.Slice(moves, func(i, j int) bool { return moves[i].Name < moves[j].Name })

	// Sort fields within each move for deterministic output (Go maps iterate
	// in random order, and we want stable git diffs).
	for i := range moves {
		sort.Slice(moves[i].Fields, func(a, c int) bool {
			return moves[i].Fields[a].Name < moves[i].Fields[c].Name
		})
	}

	// Generate distinct creator-input, resolved, and form-wire contracts.
	for _, move := range moves {
		baseName := toPascalCase(move.Name)
		inputFields := fieldsFor(move.Fields, false)
		resolvedFields := fieldsFor(move.Fields, true)
		wireFields := fieldsFor(move.Fields, false)
		writeInterface(&b, baseName+"Input", inputFields, false)
		writeInterface(&b, baseName+"Resolved", resolvedFields, false, false)
		writeInterface(&b, baseName+"Wire", wireFields, true, true)
	}

	b.WriteString("/** Maps move names to creator-facing native inputs. */\n")
	b.WriteString("export type MoveInputs = {\n")
	for _, move := range moves {
		b.WriteString("  " + tsString(move.Name) + ": " + toPascalCase(move.Name) + "Input;\n")
	}
	b.WriteString("};\n")
	b.WriteString("\n/** @deprecated Use MoveInputs. */\nexport type MoveArgs = MoveInputs;\n\n")
	b.WriteString("export type ResolvedMoveInputs = {\n")
	for _, move := range moves {
		b.WriteString("  " + tsString(move.Name) + ": " + toPascalCase(move.Name) + "Resolved;\n")
	}
	b.WriteString("};\n\nexport type MoveWireInputs = {\n")
	for _, move := range moves {
		b.WriteString("  " + tsString(move.Name) + ": " + toPascalCase(move.Name) + "Wire;\n")
	}
	b.WriteString("};\n\n")

	b.WriteString("export const moveInputSchema = ")
	b.WriteString(schemaJSON(moves))
	b.WriteString(" as const;\n\n")
	b.WriteString("export const moveInputSchemaFingerprint = ")
	b.WriteString(tsString(schemaFingerprint(moves)))
	b.WriteString(";\n")

	return b.String()
}

func fieldsFor(fields []MoveFieldInfo, includeContext bool) []MoveFieldInfo {
	result := make([]MoveFieldInfo, 0, len(fields))
	for _, field := range fields {
		if field.Disposition == "unsupported" || (!includeContext && field.Disposition == "context-owned") {
			continue
		}
		result = append(result, field)
	}
	return result
}

func writeInterface(b *strings.Builder, name string, fields []MoveFieldInfo, wire bool, defaultsOptional ...bool) {
	if len(fields) == 0 {
		b.WriteString("export type " + name + " = Record<string, never>;\n\n")
		return
	}
	b.WriteString("export interface " + name + " {\n")
	for _, field := range fields {
		b.WriteString("  readonly " + field.Name)
		optional := len(defaultsOptional) == 0 || defaultsOptional[0]
		if optional && field.Disposition == "server-defaulted" {
			b.WriteString("?")
		}
		if wire {
			b.WriteString(": string;\n")
		} else {
			b.WriteString(": " + fieldTypeToTS(field) + ";\n")
		}
	}
	b.WriteString("}\n\n")
}

// fieldTypeToTS maps a Go property type string to a TypeScript type.
func fieldTypeToTS(field MoveFieldInfo) string {
	switch field.Codec {
	case "integer", "player-index":
		return "number"
	case "boolean":
		return "boolean"
	case "enum":
		if len(field.EnumValues) == 0 {
			return "string"
		}
		values := make([]string, len(field.EnumValues))
		for i, value := range field.EnumValues {
			values[i] = tsString(value)
		}
		return strings.Join(values, " | ")
	case "string":
		return "string"
	default:
		return "never"
	}
}

func tsString(value string) string {
	return strconv.Quote(value)
}

func schemaJSON(moves []MoveInfo) string {
	encoded, err := json.MarshalIndent(moves, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(encoded)
}

func schemaFingerprint(moves []MoveInfo) string {
	return boardgame.FingerprintMoveInputSchema(moves)
}

// toPascalCase converts a move name like "Reveal Card" to "RevealCard".
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

const typeScriptEmptyBody = `export type MoveInputs = Record<string, Record<string, never>>;
/** @deprecated Use MoveInputs. */
export type MoveArgs = MoveInputs;
export type ResolvedMoveInputs = MoveInputs;
export type MoveWireInputs = MoveInputs;
export const moveInputSchema = [] as const;
export const moveInputSchemaFingerprint = "sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945";
`
