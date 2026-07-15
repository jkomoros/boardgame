package gametypes

import (
	"fmt"
	"regexp"
)

var typeScriptIdentifier = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

// ValidateTypeResult rejects extractor output that cannot be represented by the
// generated TypeScript API. Validation is deliberately separate from rendering
// so callers can validate a complete generation before touching any files.
func ValidateTypeResult(result TypeResult) error {
	declared := map[string]string{
		"ComponentCatalog":       "framework component catalog",
		"DynamicComponentValues": "framework dynamic component values",
		"GameComputed":           "framework game computed values",
		"PlayerComputed":         "framework player computed values",
		"GameState":              "framework game state",
		"PlayerState":            "framework player state",
		"State":                  "framework full state",
	}
	declare := func(name, source string) error {
		if !typeScriptIdentifier.MatchString(name) {
			return fmt.Errorf("%s generates invalid TypeScript identifier %q", source, name)
		}
		if previous, ok := declared[name]; ok {
			return fmt.Errorf("%s and %s both generate TypeScript declaration %q", previous, source, name)
		}
		declared[name] = source
		return nil
	}
	validateFields := func(owner string, fields []FieldInfo) error {
		seen := make(map[string]bool)
		for _, field := range fields {
			if !typeScriptIdentifier.MatchString(field.Name) {
				return fmt.Errorf("%s field %q is not a valid TypeScript identifier", owner, field.Name)
			}
			if seen[field.Name] {
				return fmt.Errorf("%s contains duplicate field %q", owner, field.Name)
			}
			seen[field.Name] = true
			if (field.Type == "TypeEnum" || field.Type == "TypeEnumSlice") && field.EnumName != "" {
				name := toPascalCase(field.EnumName) + "Value"
				if !typeScriptIdentifier.MatchString(name) {
					return fmt.Errorf("%s field %q references enum %q, which generates invalid TypeScript identifier %q", owner, field.Name, field.EnumName, name)
				}
			}
		}
		return nil
	}

	if err := validateFields("game state", result.GameFields); err != nil {
		return err
	}
	if err := validateFields("player state", result.PlayerFields); err != nil {
		return err
	}
	for _, enum := range result.Enums {
		if err := declare(toPascalCase(enum.Name)+"Value", fmt.Sprintf("enum %q", enum.Name)); err != nil {
			return err
		}
	}
	for _, deck := range result.Decks {
		base := toPascalCase(deck.Name)
		if base == "" || !typeScriptIdentifier.MatchString(base) {
			return fmt.Errorf("deck %q generates invalid TypeScript identifier %q", deck.Name, base)
		}
		if len(deck.Fields) > 0 {
			if err := declare(base+"ComponentValues", fmt.Sprintf("deck %q static values", deck.Name)); err != nil {
				return err
			}
		}
		if len(deck.DynamicFields) > 0 {
			if err := declare(base+"DynamicComponentValues", fmt.Sprintf("deck %q dynamic values", deck.Name)); err != nil {
				return err
			}
		}
		if err := validateFields(fmt.Sprintf("deck %q static values", deck.Name), deck.Fields); err != nil {
			return err
		}
		if err := validateFields(fmt.Sprintf("deck %q dynamic values", deck.Name), deck.DynamicFields); err != nil {
			return err
		}
	}
	return nil
}
