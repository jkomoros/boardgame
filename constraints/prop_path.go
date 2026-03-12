package constraints

import (
	"fmt"
	"strings"

	"github.com/jkomoros/boardgame"
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
