package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/moves/internal/privateconstants"
)

//WithLegalType returns a function configuration option suitable for being
//passed to auto.Config. The legalType will be bassed to the components'
//Legal() method. Idiomatically this should be a value from an enum that is
//related to the legalType for that type of component. However, if you only
//have one DefaultComponent move for that type of component, it's fine to just
//skip this to use 0 instead.
func WithLegalType(legalType int) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.LegalType] = legalType
	}
}

//WithMoveName returns a function configuration option suitable for being
//passed to auto.Config. moves.Base uses this, if provided, to power
//MoveTypeName, which means that auto.Config will use this name whenever it is
//passed. If you're passing a move struct that not's from this package, the
//auto-generated move name is likely sufficient and you don't need this. See
//the documentation for moves.Base.MoveTypeName for more information.
func WithMoveName(moveName string) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.MoveName] = moveName
	}
}

//WithHelpText returns a function configuration option suitable for being
//passed to auto.Config. moves.Base uses this, if provided, to power
//MoveTypeHelpText, which means that auto.Config will use this text whenever
//it is passed. See the documentation for moves.Base.MoveTypeHelpText for more
//information.
func WithHelpText(helpText string) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.HelpText] = helpText
	}
}

//WithLegalPhases returns a function configuration option suitable for being
//passed to auto.Config. legalPhases will extend whatever has already been
//passed before. move.Base will use the result of this to determine if a given
//move is legal in the current phase. Typically you don't use this directly,
//and instead use moves.AddForPhase to use this implicitly.
func WithLegalPhases(legalPhases ...int) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		previousLegalPhases := config[privateconstants.LegalPhases]

		if ints, ok := previousLegalPhases.([]int); ok {
			legalPhases = append(ints, legalPhases...)
		}

		config[privateconstants.LegalPhases] = legalPhases
	}
}

//WithLegalMoveProgression returns a function configuration option suitable
//for being passed to auto.Config. moves.Base's Legal() will use this for this
//move type to determine if the move is legal in the order it's being applied.
//Typically you don't use this directly, and instead use moves.AddOrderedForPhase to
//use this implicitly.
func WithLegalMoveProgression(moveProgression []string) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.LegalMoveProgression] = moveProgression
	}
}

//WithIsFixUp returns a function configuration option suitable for being
//passed to auto.Config. moves.Base uses this, if provided, to power
//MoveTypeIsFixUp, which means that auto.Config will use this if it is passed.
//See the documentation for moves.Base.MoveTypeIsFixup for more information.
//All moves in this package will return reasonable values for MoveTypeIsFixUp
//on their own, so it is much more rare to use this than other config options
//in this package. In general, instead of using this option you should simply
//embed FixUp (or a move that itself embedds IsFixUp), so you don't have to
//remember to pass WithIsFixUp, which is easy to forget.
func WithIsFixUp(isFixUp bool) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.IsFixUp] = isFixUp
	}
}

//WithPhaseToStart returns a function configuration option suitable for being
//passed to auto.Config. PhaseEnum should be the enum that is used for phases,
//and phaseToStart is the value within that phase to start. The phaseEnum is
//optional; if not provided, the name of the move and help text will just use
//the int value of the phase instead.
func WithPhaseToStart(phaseToStart int, optionalPhaseEnum enum.Enum) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.StartPhase] = phaseToStart
		config[privateconstants.StartPhaseEnum] = optionalPhaseEnum
	}
}

//WithSourceStack returns a function configuration option suitable for being
//passed to auto.Config. The stackPropName is assumed to be on the GameState
//object. If it isn't, you'll need to embed the move and override SourceStack
//yourself.
func WithSourceStack(stackPropName string) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.SourceStack] = stackPropName
	}
}

//WithDestinationStack returns a function configuration option suitable for
//being passed to auto.Config. The stackPropName is assumed to be on the
//GameState object. If it isn't, you'll need to embed the move and override
//DestinationStack yourself.
func WithDestinationStack(stackPropName string) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.DestinationStack] = stackPropName
	}
}

//WithGameStack returns a function configuration option suitable for being
//passed to auto.Config.
func WithGameStack(stackPropName string) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.GameStack] = stackPropName
	}
}

//WithPlayerStack returns a function configuration option suitable for being
//passed to auto.Config.
func WithPlayerStack(stackPropName string) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.PlayerStack] = stackPropName
	}
}

//WithNumRounds returns a function configuration option suitable for being
//passed to auto.Config.
func WithNumRounds(numRounds int) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.NumRounds] = numRounds
	}
}

//WithTargetCount returns a function configuration option suitable for being
//passed to auto.Config.
func WithTargetCount(targetCount int) interfaces.CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[privateconstants.TargetCount] = targetCount
	}
}
