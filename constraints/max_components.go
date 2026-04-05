package constraints

import (
	"strconv"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/errors"
)

// MaxNumComponents returns a StackConstraint that rejects a move if the
// destination stack would contain more than max components after the
// addition. The destination stack is in its pre-insertion state, so the
// constraint adds len(proposed) to the current count.
func MaxNumComponents(max int) boardgame.StackConstraint {
	return func(dest boardgame.ImmutableStack, proposed []boardgame.ImmutableComponentInstance, state boardgame.ImmutableState) error {
		afterCount := dest.NumComponents() + len(proposed)
		if afterCount > max {
			return errors.New("stack would have " + strconv.Itoa(afterCount) + " components, which exceeds max of " + strconv.Itoa(max))
		}
		return nil
	}
}

// MaxNumComponentsConstructor returns a StackConstraintConstructor for use in
// struct tags. Usage: max(N), where N is the maximum number of components.
func MaxNumComponentsConstructor() *boardgame.StackConstraintConstructor {
	return &boardgame.StackConstraintConstructor{
		Name: "max",
		Constructor: func(args []string, chest *boardgame.ComponentChest) (boardgame.StackConstraint, error) {
			if len(args) != 1 {
				return nil, errors.New("max constraint requires exactly 1 argument")
			}
			max, err := intEffectiveValue(args[0], chest)
			if err != nil {
				return nil, errors.New("max constraint argument is not a valid int: " + err.Error())
			}
			return MaxNumComponents(max), nil
		},
	}
}
