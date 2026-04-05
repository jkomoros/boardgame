package constraints

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/errors"
)

// Same returns a StackConstraint that rejects a move if not all components in
// the destination stack (including the proposed additions) have the same value
// for the named property. Components that don't have the named property are
// skipped. A stack with 0 or 1 resolvable components always passes. The
// destination stack is in its pre-insertion state, so proposed components are
// checked separately against the established value.
func Same(propPath string) boardgame.StackConstraint {
	return func(dest boardgame.ImmutableStack, proposed []boardgame.ImmutableComponentInstance, state boardgame.ImmutableState) error {
		var firstVal string
		firstSet := false
		// Establish the common value from existing components.
		for _, c := range dest.ImmutableComponents() {
			if c == nil {
				continue
			}
			val, ok := resolvePropValue(c, propPath)
			if !ok {
				continue
			}
			if !firstSet {
				firstVal = val
				firstSet = true
				continue
			}
			if val != firstVal {
				return errors.New("property " + propPath + " has mixed values: " + firstVal + " and " + val)
			}
		}
		// Check proposed additions match the established value.
		for _, c := range proposed {
			if c == nil {
				continue
			}
			val, ok := resolvePropValue(c, propPath)
			if !ok {
				continue
			}
			if !firstSet {
				firstVal = val
				firstSet = true
				continue
			}
			if val != firstVal {
				return errors.New("property " + propPath + " has mixed values: " + firstVal + " and " + val)
			}
		}
		return nil
	}
}

// SameConstructor returns a StackConstraintConstructor for use in struct
// tags. Usage: same(propPath).
func SameConstructor() *boardgame.StackConstraintConstructor {
	return &boardgame.StackConstraintConstructor{
		Name: "same",
		Constructor: func(args []string, chest *boardgame.ComponentChest) (boardgame.StackConstraint, error) {
			if len(args) != 1 {
				return nil, errors.New("same constraint requires exactly 1 argument")
			}
			if err := validatePropPath(args[0], chest); err != nil {
				return nil, errors.New("same constraint: " + err.Error())
			}
			return Same(args[0]), nil
		},
	}
}
