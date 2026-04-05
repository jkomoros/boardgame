package constraints

import (
	"strconv"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/errors"
)

// MaxDistinctValues returns a StackConstraint that rejects a move if the
// destination stack (including the proposed additions) has more than max
// distinct values for the named property. Components that don't have the
// named property are skipped. The destination stack is in its pre-insertion
// state, so proposed components are counted separately.
func MaxDistinctValues(propPath string, max int) boardgame.StackConstraint {
	return func(dest boardgame.ImmutableStack, proposed []boardgame.ImmutableComponentInstance, state boardgame.ImmutableState) error {
		seen := make(map[string]bool)
		// Collect distinct values from existing components.
		for _, c := range dest.ImmutableComponents() {
			if c == nil {
				continue
			}
			val, ok := resolvePropValue(c, propPath)
			if !ok {
				continue
			}
			seen[val] = true
		}
		// Include proposed additions.
		for _, c := range proposed {
			if c == nil {
				continue
			}
			val, ok := resolvePropValue(c, propPath)
			if !ok {
				continue
			}
			seen[val] = true
		}
		if len(seen) > max {
			return errors.New("property " + propPath + " would have " + strconv.Itoa(len(seen)) + " distinct values, which exceeds max of " + strconv.Itoa(max))
		}
		return nil
	}
}

// MaxDistinctValuesConstructor returns a StackConstraintConstructor for use
// in struct tags. Usage: maxdistinct(propPath,N).
func MaxDistinctValuesConstructor() *boardgame.StackConstraintConstructor {
	return &boardgame.StackConstraintConstructor{
		Name: "maxdistinct",
		Constructor: func(args []string, chest *boardgame.ComponentChest) (boardgame.StackConstraint, error) {
			if len(args) != 2 {
				return nil, errors.New("maxdistinct constraint requires exactly 2 arguments: propPath and max")
			}
			if err := validatePropPath(args[0], chest); err != nil {
				return nil, errors.New("maxdistinct constraint: " + err.Error())
			}
			max, err := intEffectiveValue(args[1], chest)
			if err != nil {
				return nil, errors.New("maxdistinct second argument is not a valid int: " + err.Error())
			}
			return MaxDistinctValues(args[0], max), nil
		},
	}
}
