package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

const fullyQualifiedPackageName = "github.com/jkomoros/boardgame/moves."

const configPropStartPhase = fullyQualifiedPackageName + "StartPhase"
const configPropStartPhaseEnum = fullyQualifiedPackageName + "StartPhaseEnum"
const configPropSourceProperty = fullyQualifiedPackageName + "SourceProperty"
const configPropDestinationProperty = fullyQualifiedPackageName + "DestinationProperty"
const configPropTargetCount = fullyQualifiedPackageName + "TargetCount"
const configPropNumRounds = fullyQualifiedPackageName + "NumRounds"
const configPropGameProperty = fullyQualifiedPackageName + "GameProperty"
const configPropPlayerProperty = fullyQualifiedPackageName + "PlayerPropety"
const configPropMoveName = fullyQualifiedPackageName + "MoveName"
const configPropMoveNameSuffix = fullyQualifiedPackageName + "MoveNameSuffix"
const configPropHelpText = fullyQualifiedPackageName + "HelpText"
const configPropIsFixUp = fullyQualifiedPackageName + "IsFixUp"
const configPropLegalPhases = fullyQualifiedPackageName + "LegalPhases"
const configPropLegalMoveProgression = fullyQualifiedPackageName + "LegalMoveProgression"
const configPropLegalType = fullyQualifiedPackageName + "LegalType"
const configPropAmount = fullyQualifiedPackageName + "Amount"

//CustomConfigurationOption is a function that takes a PropertyCollection and
//modifies a key on it. This package defines a number of functions that return
//funcs that satisfy this interface and can be used in auto.Config to pass in
//configuration to the base moves without requiring verbose embedding and
//method overriding. All of those functions in this package start with "With".
type CustomConfigurationOption func(boardgame.PropertyCollection)

//LegalType returns a function configuration option suitable for being
//passed to auto.Config. The legalType will be bassed to the components'
//Legal() method. Idiomatically this should be a value from an enum that is
//related to the legalType for that type of component. However, if you only
//have one DefaultComponent move for that type of component, it's fine to just
//skip this to use 0 instead.
func WithLegalType(legalType int) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropLegalType] = legalType
	}
}

//MoveName returns a function configuration option suitable for being
//passed to auto.Config. moves.Base uses this, if provided, to power
//MoveTypeName, which means that auto.Config will use this name whenever it is
//passed. If you're passing a move struct that not's from this package, the
//auto-generated move name is likely sufficient and you don't need this. See
//the documentation for moves.Base.MoveTypeName for more information.
func WithMoveName(moveName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropMoveName] = moveName
	}
}

//MoveNameSuffix returns a function configuration option suitable for being
//passed to auto.Config. The suffix, if provided, will be appended to whatever
//the Move's name would have been (see the behavior for DeriveName on
//move.Base). This is useful because every move must have a unique name, but
//sometimes you have the same underlying move struct who is legal in different
//points in different progressions. This makes it easy to provide a suffix for
//subsequent uses of the same move to ensure the names are all unique.
func WithMoveNameSuffix(suffix string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropMoveNameSuffix] = suffix
	}
}

//HelpText returns a function configuration option suitable for being
//passed to auto.Config. moves.Base uses this, if provided, to power
//MoveTypeHelpText, which means that auto.Config will use this text whenever
//it is passed. See the documentation for moves.Base.MoveTypeHelpText for more
//information.
func WithHelpText(helpText string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropHelpText] = helpText
	}
}

//LegalPhases returns a function configuration option suitable for being
//passed to auto.Config. legalPhases will extend whatever has already been
//passed before. move.Base will use the result of this to determine if a given
//move is legal in the current phase. Typically you don't use this directly,
//and instead use moves.AddForPhase to use this implicitly.
func WithLegalPhases(legalPhases ...int) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		previousLegalPhases := config[configPropLegalPhases]

		if ints, ok := previousLegalPhases.([]int); ok {
			legalPhases = append(ints, legalPhases...)
		}

		config[configPropLegalPhases] = legalPhases
	}
}

//LegalMoveProgression returns a function configuration option suitable
//for being passed to auto.Config. moves.Base's Legal() will use this for this
//move type to determine if the move is legal in the order it's being applied.
//Typically you don't use this directly, and instead use moves.AddOrderedForPhase to
//use this implicitly.
func WithLegalMoveProgression(group MoveProgressionGroup) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropLegalMoveProgression] = group
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
//WithIsFixUp, which is easy to forget.
func WithIsFixUp(isFixUp bool) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropIsFixUp] = isFixUp
	}
}

//PhaseToStart returns a function configuration option suitable for being
//passed to auto.Config. PhaseEnum should be the enum that is used for phases,
//and phaseToStart is the value within that phase to start. The phaseEnum is
//optional; if not provided, the name of the move and help text will just use
//the int value of the phase instead.
func WithPhaseToStart(phaseToStart int, optionalPhaseEnum enum.Enum) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropStartPhase] = phaseToStart
		config[configPropStartPhaseEnum] = optionalPhaseEnum
	}
}

//SourceProperty returns a function configuration option suitable for being
//passed to auto.Config. The stackPropName is assumed to be on the GameState
//object. If it isn't, you'll need to embed the move and override SourceStack
//yourself.
func WithSourceProperty(stackPropName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropSourceProperty] = stackPropName
	}
}

//DestinationProperty returns a function configuration option suitable for
//being passed to auto.Config. The stackPropName is assumed to be on the
//GameState object. If it isn't, you'll need to embed the move and override
//DestinationStack yourself.
func WithDestinationProperty(stackPropName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropDestinationProperty] = stackPropName
	}
}

//GameProperty returns a function configuration option suitable for being
//passed to auto.Config. Often used to configure what a move's GameStack()
//will return, but other moves use it for non-stack properties.
func WithGameProperty(stackPropName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropGameProperty] = stackPropName
	}
}

//PlayerProperty returns a function configuration option suitable for being
//passed to auto.Config. Often used to configure what a move's PlayerStack()
//will return, but other moves use it for non-stack properties.
func WithPlayerProperty(stackPropName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropPlayerProperty] = stackPropName
	}
}

//NumRounds returns a function configuration option suitable for being
//passed to auto.Config.
func WithNumRounds(numRounds int) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropNumRounds] = numRounds
	}
}

//TargetCount returns a function configuration option suitable for being
//passed to auto.Config.
func WithTargetCount(targetCount int) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropTargetCount] = targetCount
	}
}

//Amount returns a function configuration option suitable for being
//passed to auto.Config.
func WithAmount(amount int) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configPropAmount] = amount
	}
}
