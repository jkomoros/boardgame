package constraints

import "github.com/jkomoros/boardgame"

// DefaultConstructors returns the full set of pre-built
// StackConstraintConstructors provided by this package. This is designed to
// be returned directly from GameDelegate.ConfigureStackConstraintConstructors.
func DefaultConstructors() []*boardgame.StackConstraintConstructor {
	return []*boardgame.StackConstraintConstructor{
		MaxNumComponentsConstructor(),
		UniqueConstructor(),
		SameConstructor(),
		MaxDistinctValuesConstructor(),
	}
}
