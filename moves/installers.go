package moves

import (
	"github.com/jkomoros/boardgame"
)

//Combine takes a series of lists of moveTypeConfigs and flattens them into a
//single list, appropraite for being retunrned from delegate.ConfigureMoves().
func Combine(moves ...[]boardgame.MoveTypeConfig) []boardgame.MoveTypeConfig {

	var result []boardgame.MoveTypeConfig

	for _, list := range moves {
		result = append(result, list...)
	}

	return result

}

//Add is designed to be used inside of Combine. It is a parallel of
//AddForPhase and AddOrderedForPhase. It doesn't actually do any processing,
//and is effectively equivalent to wrapping your config bundles in
//[]*boardgame.MoveTypeConfig{}.
func Add(moves ...boardgame.MoveTypeConfig) []boardgame.MoveTypeConfig {
	return moves
}

//AddForPhase is designed to be used within Combine. It calls
//WithLegalPhases() on the config for each config passed in. It's a
//convenience to make it less error-prone and more clear what the intent is
//for phase-locked moves.
func AddForPhase(phase int, moves ...boardgame.MoveTypeConfig) []boardgame.MoveTypeConfig {

	phaseInstaller := WithLegalPhases(phase)

	for _, move := range moves {
		phaseInstaller(move.CustomConfiguration)
	}

	return moves

}

//AddOrderedForPhase is designed to be used within Combine. It calls
//WithLegalPhases() and also WithLegalMoveProgression() on the config for each
//config passed in. It's a convenience to make it less error-prone and more
//clear what the intent is for phase-locked, ordered moves. All moveTypes
//passed must be legal auto-configurable moves.
func AddOrderedForPhase(phase int, moves ...boardgame.MoveTypeConfig) []boardgame.MoveTypeConfig {

	progression := make([]string, len(moves))

	for i, move := range moves {
		progression[i] = move.Name
	}

	installer := WithLegalMoveProgression(progression)

	for _, move := range moves {
		installer(move.CustomConfiguration)
	}

	return AddForPhase(phase, moves...)

}
