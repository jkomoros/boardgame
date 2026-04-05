package constraints

import (
	"fmt"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/errors"
)

// resolvePropValue resolves a property path for a given component instance.
// The path syntax is:
//   - "name" — check Values() first, fall back to ImmutableDynamicValues()
//   - "component.name" — only Values()
//   - "dynamic.name" — only ImmutableDynamicValues()
//
// Returns the value as a string (via fmt.Sprintf) and true if found, or
// ("", false) if the property doesn't exist on this component.
func resolvePropValue(c boardgame.ImmutableComponentInstance, propPath string) (string, bool) {
	prefix, propName := splitPropPath(propPath)

	switch prefix {
	case "component":
		return readPropFromValues(c, propName)
	case "dynamic":
		return readPropFromDynamic(c, propName)
	default:
		// No prefix: try Values() first, then DynamicValues().
		if val, ok := readPropFromValues(c, propName); ok {
			return val, true
		}
		return readPropFromDynamic(c, propName)
	}
}

func splitPropPath(propPath string) (prefix, name string) {
	parts := strings.SplitN(propPath, ".", 2)
	if len(parts) == 2 {
		lower := strings.ToLower(parts[0])
		if lower == "component" || lower == "dynamic" {
			return lower, parts[1]
		}
	}
	return "", propPath
}

func readPropFromValues(c boardgame.ImmutableComponentInstance, propName string) (string, bool) {
	vals := c.Values()
	if vals == nil {
		return "", false
	}
	reader := vals.Reader()
	if reader == nil {
		return "", false
	}
	v, err := reader.Prop(propName)
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("%v", v), true
}

func readPropFromDynamic(c boardgame.ImmutableComponentInstance, propName string) (string, bool) {
	dyn := c.ImmutableDynamicValues()
	if dyn == nil {
		return "", false
	}
	reader := dyn.Reader()
	if reader == nil {
		return "", false
	}
	v, err := reader.Prop(propName)
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("%v", v), true
}

// validatePropPath checks at construction time that the given propPath refers
// to a property that exists on at least one deck's components. This catches
// typos like "colour" instead of "color" that would otherwise cause the
// constraint to silently skip every component.
//
// chest may be nil (when used programmatically without a chest), in which case
// validation is skipped.
func validatePropPath(propPath string, chest *boardgame.ComponentChest) error {
	if chest == nil {
		return nil
	}

	prefix, propName := splitPropPath(propPath)

	checkComponent := prefix == "component" || prefix == ""
	checkDynamic := prefix == "dynamic" || prefix == ""

	for _, deckName := range chest.DeckNames() {
		deck := chest.Deck(deckName)
		if deck == nil {
			continue
		}

		if checkComponent {
			if deckHasComponentProp(deck, propName) {
				return nil
			}
		}

		if checkDynamic {
			if deckHasDynamicProp(deck, propName, chest) {
				return nil
			}
		}
	}

	return errors.New("property " + propPath + " does not exist on any deck's components; check for typos in the property name")
}

// deckHasComponentProp returns true if the deck has at least one component
// whose Values() reader exposes the named property.
func deckHasComponentProp(deck *boardgame.Deck, propName string) bool {
	components := deck.Components()
	if len(components) == 0 {
		return false
	}
	c := components[0]
	if c == nil {
		return false
	}
	vals := c.Values()
	if vals == nil {
		return false
	}
	reader := vals.Reader()
	if reader == nil {
		return false
	}
	props := reader.Props()
	_, exists := props[propName]
	return exists
}

// deckHasDynamicProp returns true if the deck's DynamicComponentValues (as
// provided by the game delegate) exposes the named property.
func deckHasDynamicProp(deck *boardgame.Deck, propName string, chest *boardgame.ComponentChest) bool {
	manager := chest.Manager()
	if manager == nil {
		return false
	}
	delegate := manager.Delegate()
	if delegate == nil {
		return false
	}
	dyn := delegate.DynamicComponentValuesConstructor(deck)
	if dyn == nil {
		return false
	}
	reader := dyn.Reader()
	if reader == nil {
		return false
	}
	props := reader.Props()
	_, exists := props[propName]
	return exists
}
