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

// ExtendDefaults returns DefaultConstructors() with the provided custom
// constructors appended. This is a convenience for
// GameDelegate.ConfigureStackConstraintConstructors when you want the
// pre-built constraints plus your own.
func ExtendDefaults(custom ...*boardgame.StackConstraintConstructor) []*boardgame.StackConstraintConstructor {
	return append(DefaultConstructors(), custom...)
}
