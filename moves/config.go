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

const configNameStartPhase = "__moves.StartPhaseConfigProp"
const configNameSourceStack = "__moves.SourceStackConfigProp"
const configNameDestinationStack = "__moves.DestinationStackConfigProp"
const configNameTargetCount = "__moves.TargetCountConfigProp"
const configNameNumRounds = "__moves.NumRoundsConfigProp"
const configNameGameStack = "__moves.GameStackConfigProp"
const configNamePlayerStack = "__moves.PlayerStackConfigProp"

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
