/*

	with is a package of configuration functions designed to be passed in as
	options in auto.Config.

*/
package with

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/moves/internal/privateconstants"
)

//LegalType returns a function configuration option suitable for being
//passed to auto.Config. The legalType will be bassed to the components'
//Legal() method. Idiomatically this should be a value from an enum that is
//related to the legalType for that type of component. However, if you only
//have one DefaultComponent move for that type of component, it's fine to just
//skip this to use 0 instead.
func LegalType(legalType int) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.LegalType] = legalType
	}
}

//MoveName returns a function configuration option suitable for being
//passed to auto.Config. moves.Base uses this, if provided, to power
//MoveTypeName, which means that auto.Config will use this name whenever it is
//passed. If you're passing a move struct that not's from this package, the
//auto-generated move name is likely sufficient and you don't need this. See
//the documentation for moves.Base.MoveTypeName for more information.
func MoveName(moveName string) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.MoveName] = moveName
	}
}

//MoveNameSuffix returns a function configuration option suitable for being
//passed to auto.Config. The suffix, if provided, will be appended to whatever
//the Move's name would have been (see the behavior for DeriveName on
//move.Base). This is useful because every move must have a unique name, but
//sometimes you have the same underlying move struct who is legal in different
//points in different progressions. This makes it easy to provide a suffix for
//subsequent uses of the same move to ensure the names are all unique.
func MoveNameSuffix(suffix string) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.MoveNameSuffix] = suffix
	}
}

//HelpText returns a function configuration option suitable for being
//passed to auto.Config. moves.Base uses this, if provided, to power
//MoveTypeHelpText, which means that auto.Config will use this text whenever
//it is passed. See the documentation for moves.Base.MoveTypeHelpText for more
//information.
func HelpText(helpText string) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.HelpText] = helpText
	}
}

//LegalPhases returns a function configuration option suitable for being
//passed to auto.Config. legalPhases will extend whatever has already been
//passed before. move.Base will use the result of this to determine if a given
//move is legal in the current phase. Typically you don't use this directly,
//and instead use moves.AddForPhase to use this implicitly.
func LegalPhases(legalPhases ...int) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		previousLegalPhases := config[privateconstants.LegalPhases]

		if ints, ok := previousLegalPhases.([]int); ok {
			legalPhases = append(ints, legalPhases...)
		}

		config[privateconstants.LegalPhases] = legalPhases
	}
}

//LegalMoveProgression returns a function configuration option suitable
//for being passed to auto.Config. moves.Base's Legal() will use this for this
//move type to determine if the move is legal in the order it's being applied.
//Typically you don't use this directly, and instead use moves.AddOrderedForPhase to
//use this implicitly.
func LegalMoveProgression(group interfaces.MoveProgressionGroup) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.LegalMoveProgression] = group
	}
}

//IsFixUp returns a function configuration option suitable for being passed to
//auto.Config. moves.Base uses this, if provided, to power MoveTypeIsFixUp,
//which means that auto.Config will use this if it is passed. See the
//documentation for moves.Base.IsFixUp for more information. All moves in this
//package will return reasonable values for IsFixUp on their own, so it is
//much more rare to use this than other config options in this package. In
//general, instead of using this option you should simply embed FixUp (or a
//move that itself embedds IsFixUp), so you don't have to remember to pass
//with.IsFixUp, which is easy to forget.
func IsFixUp(isFixUp bool) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.IsFixUp] = isFixUp
	}
}

//PhaseToStart returns a function configuration option suitable for being
//passed to auto.Config. PhaseEnum should be the enum that is used for phases,
//and phaseToStart is the value within that phase to start. The phaseEnum is
//optional; if not provided, the name of the move and help text will just use
//the int value of the phase instead.
func PhaseToStart(phaseToStart int, optionalPhaseEnum enum.Enum) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.StartPhase] = phaseToStart
		config[privateconstants.StartPhaseEnum] = optionalPhaseEnum
	}
}

//SourceStack returns a function configuration option suitable for being
//passed to auto.Config. The stackPropName is assumed to be on the GameState
//object. If it isn't, you'll need to embed the move and override SourceStack
//yourself.
func SourceStack(stackPropName string) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.SourceStack] = stackPropName
	}
}

//DestinationStack returns a function configuration option suitable for
//being passed to auto.Config. The stackPropName is assumed to be on the
//GameState object. If it isn't, you'll need to embed the move and override
//DestinationStack yourself.
func DestinationStack(stackPropName string) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.DestinationStack] = stackPropName
	}
}

//GameStack returns a function configuration option suitable for being passed
//to auto.Config.
func GameStack(stackPropName string) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.GameStack] = stackPropName
	}
}

//PlayerStack returns a function configuration option suitable for being
//passed to auto.Config.
func PlayerStack(stackPropName string) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.PlayerStack] = stackPropName
	}
}

//NumRounds returns a function configuration option suitable for being
//passed to auto.Config.
func NumRounds(numRounds int) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.NumRounds] = numRounds
	}
}

//TargetCount returns a function configuration option suitable for being
//passed to auto.Config.
func TargetCount(targetCount int) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.TargetCount] = targetCount
	}
}
