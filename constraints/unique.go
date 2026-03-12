package constraints

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/errors"
)

// Unique returns a StackConstraint that rejects a move if any two components
// in the destination stack share the same value for the named property.
// Components that don't have the named property are skipped.
func Unique(propPath string) boardgame.StackConstraint {
	return func(dest boardgame.ImmutableStack, justAdded []boardgame.ImmutableComponentInstance, state boardgame.ImmutableState) error {
		seen := make(map[string]bool)
		for _, c := range dest.ImmutableComponents() {
			if c == nil {
				continue
			}
			val, ok := resolvePropValue(c, propPath)
			if !ok {
				continue
			}
			if seen[val] {
				return errors.New("duplicate value " + val + " for property " + propPath)
			}
			seen[val] = true
		}
		return nil
	}
}

// UniqueConstructor returns a StackConstraintConstructor for use in struct
// tags. Usage: unique(propPath).
func UniqueConstructor() *boardgame.StackConstraintConstructor {
	return &boardgame.StackConstraintConstructor{
		Name: "unique",
		Constructor: func(args []string, chest *boardgame.ComponentChest) (boardgame.StackConstraint, error) {
			if len(args) != 1 {
				return nil, errors.New("unique constraint requires exactly 1 argument")
			}
			return Unique(args[0]), nil
		},
	}
}
