package moves

import (
	"github.com/jkomoros/boardgame"
)

//CustomConfigurationOption is a function that takes a PropertyCollection and
//modifies a key on it. This package defines a number of functions that return
//funcs that satisfy this interface and can be used in DefaultConfig to pass
//in configuration to the base moves without requiring verbose embedding and
//method overriding. All of those functions in this package start with "With".
type CustomConfigurationOption func(boardgame.PropertyCollection)

const fullyQualifiedPackageName = "github.com/jkomoros/boardgame/moves."

const configNameStartPhase = fullyQualifiedPackageName + "StartPhase"
const configNameSourceStack = fullyQualifiedPackageName + "SourceStack"
const configNameDestinationStack = fullyQualifiedPackageName + "DestinationStack"
const configNameTargetCount = fullyQualifiedPackageName + "TargetCount"
const configNameNumRounds = fullyQualifiedPackageName + "NumRounds"
const configNameGameStack = fullyQualifiedPackageName + "GameStack"
const configNamePlayerStack = fullyQualifiedPackageName + "PlayerStack"
const configNameMoveName = fullyQualifiedPackageName + "MoveName"

//WithPhaseToStart returns a function configuration option suitable for being
//passed to DefaultConfig. moves.Base uses this, if provided, to power
//MoveTypeName, which means that DefaultConfig will use this name in some
//cases. See the documentation for moves.Base.MoveTypeName for more
//information.
func WithMoveName(moveName string) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configNameMoveName] = moveName
	}
}

//WithPhaseToStart returns a function configuration option suitable for being
//passed to DefaultConfig.
func WithPhaseToStart(phaseToStart int) CustomConfigurationOption {
	return func(config boardgame.PropertyCollection) {
		config[configNameStartPhase] = phaseToStart
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
