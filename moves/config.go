package moves

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

//CustomConfigurationOption is a function that takes a PropertyCollection and
//modifies a key on it. This package defines a number of functions that return
//funcs that satisfy this interface and can be used in DefaultConfig to pass
//in configuration to the base moves without requiring verbose embedding and
//method overriding. All of those functions in this package start with "With".
type CustomConfigurationOption func(boardgame.PropertyCollection)

const fullyQualifiedPackageName = "github.com/jkomoros/boardgame/moves."

const configNameStartPhase = fullyQualifiedPackageName + "StartPhase"
const configNameStartPhaseEnum = fullyQualifiedPackageName + "StartPhaseEnum"
const configNameSourceStack = fullyQualifiedPackageName + "SourceStack"
const configNameDestinationStack = fullyQualifiedPackageName + "DestinationStack"
const configNameTargetCount = fullyQualifiedPackageName + "TargetCount"
const configNameNumRounds = fullyQualifiedPackageName + "NumRounds"
const configNameGameStack = fullyQualifiedPackageName + "GameStack"
const configNamePlayerStack = fullyQualifiedPackageName + "PlayerStack"
const configNameMoveName = fullyQualifiedPackageName + "MoveName"
const configNameHelpText = fullyQualifiedPackageName + "HelpText"
const configNameIsFixUp = fullyQualifiedPackageName + "IsFixUp"
const configNameLegalPhases = fullyQualifiedPackageName + "LegalPhases"

//WithMoveName returns a function configuration option suitable for being
//passed to DefaultConfig. moves.Base uses this, if provided, to power
//MoveTypeName, which means that DefaultConfig will use this name in some
//cases. If you're passing a move struct that not's from this package, the
//auto-generated move name is likely sufficient and you don't need this. See
//the documentation for moves.Base.MoveTypeName for more information.
func WithMoveName(moveName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configNameMoveName] = moveName
	}
}

//WithHelpText returns a function configuration option suitable for being
//passed to DefaultConfig. moves.Base uses this, if provided, to power
//MoveTypeHelpText, which means that DefaultConfig will use this name in some
//cases. See the documentation for moves.Base.MoveTypeHelpText for more
//information.
func WithHelpText(helpText string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configNameHelpText] = helpText
	}
}

//WithLegalPhases returns a function configuration option suitable for being
//passed to DefaultConfig. moves.Base will return whatever is passed via this
//for MoveTypeLegalPhases().
func WithLegalPhases(legalPhases []int) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configNameLegalPhases] = legalPhases
	}
}

//WithIsFixUp returns a function configuration option suitable for being
//passed to DefaultConfig. moves.Base uses this, if provided, to power
//MoveTypeIsFixUp, which means that DefaultConfig will use this name in some
//cases. See the documentation for moves.Base.MoveTypeIsFixup for more
//information. All moves in this package will return reasonable values for
//MoveTypeIsFixUp on their own, so it is much more rare to use this than other
//config options in this package.
func WithIsFixUp(isFixUp bool) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configNameIsFixUp] = isFixUp
	}
}

//WithPhaseToStart returns a function configuration option suitable for being
//passed to DefaultConfig. PhaseEnum should be the enum that is used for
//phases, and phaseToStart is the value within that phase to start. The
//phaseEnum is optional; if not provided, the name of the move and help text
//will just use the int value of the phase instead.
func WithPhaseToStart(phaseToStart int, optionalPhaseEnum enum.Enum) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configNameStartPhase] = phaseToStart
		config[configNameStartPhaseEnum] = optionalPhaseEnum
	}
}

//WithSourceStack returns a function configuration option suitable for being
//passed to DefaultConfig. The stackPropName is assumed to be on the GameState
//object. If it isn't, you'll need to embed the move and override Sourcetack
//yourself.
func WithSourceStack(stackPropName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configNameSourceStack] = stackPropName
	}
}

//WithDestinationStack returns a function configuration option suitable for
//being passed to DefaultConfig. The stackPropName is assumed to be on the
//GameState object. If it isn't, you'll need to embed the move and override
//DestinationStack yourself.
func WithDestinationStack(stackPropName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configNameDestinationStack] = stackPropName
	}
}

//WithGameStack returns a function configuration option suitable for being
//passed to DefaultConfig.
func WithGameStack(stackPropName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configNameGameStack] = stackPropName
	}
}

//WithPlayerStack returns a function configuration option suitable for being
//passed to DefaultConfig.
func WithPlayerStack(stackPropName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configNamePlayerStack] = stackPropName
	}
}

//WithNumRounds returns a function configuration option suitable for being
//passed to DefaultConfig.
func WithNumRounds(numRounds int) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configNameNumRounds] = numRounds
	}
}

//WithTargetCount returns a function configuration option suitable for being
//passed to DefaultConfig.
func WithTargetCount(targetCount int) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configNameTargetCount] = targetCount
	}
}
