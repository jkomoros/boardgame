package moves

import (
	"github.com/jkomoros/boardgame"
	gr "github.com/jkomoros/boardgame/moves/groups"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/moves/with"
)

//Combine takes a series of lists of moveTypeConfigs and flattens them into a
//single list, appropriate for being retunrned from delegate.ConfigureMoves().
//It doesn't do anything special, but instead exists entirely as a convenience
//to make writing your ConfigureMoves easier.
func Combine(moves ...[]boardgame.MoveConfig) []boardgame.MoveConfig {

	var result []boardgame.MoveConfig

	for _, list := range moves {
		result = append(result, list...)
	}

	return result

}

//Add is designed to be used inside of Combine for moves that can apply in any
//phase in any order. It is a parallel of AddForPhase and AddOrderedForPhase.
//It doesn't actually do any processing, and is effectively equivalent to
//wrapping your config bundles in []boardgame.MoveConfig{}. However, it makes
//the intent of your move installers clearer.
func Add(moves ...boardgame.MoveConfig) []boardgame.MoveConfig {
	return moves
}

//AddForPhase is designed to be used within Combine. It calls
//WithLegalPhases() on the config for each config passed in, so that those
//moves will only be Legal() in that phase. It's a convenience to make it less
//error-prone and more clear what the intent is for phase-locked moves.
func AddForPhase(phase int, moves ...boardgame.MoveConfig) []boardgame.MoveConfig {

	phaseInstaller := with.LegalPhases(phase)

	for _, move := range moves {
		phaseInstaller(move.CustomConfiguration())
	}

	return moves

}

//AddOrderedForPhase is designed to be used within Combine. It calls
//WithLegalPhases() and also WithLegalMoveProgression() on the config for each
//config passed in, which means that the moves' Legal() will only be Legal in
//that phase, in that point in the move progression. It's a convenience to
//make it less error-prone and more clear what the intent is for phase-locked,
//ordered moves. All moveTypes passed must be legal auto-configurable moves.
//You may pass configs generated from AutoConfigurer.Config(), or any of the
//group types defined in moves/groups. All of the top level groups passed will
//be treated implicitly like a single group.Serial. All moves contained within
//the provided groups will be registered.
func AddOrderedForPhase(phase int, groups ...interfaces.MoveProgressionGroup) []boardgame.MoveConfig {

	//Technically it's illegal to attach a move progression to a non-leaf
	//phase enum val, but at this point we don't have a reference to delegate
	//so we can't check. moves.Base.ValidConfiguration will check.

	impliedSerialGroup := gr.Serial(groups)

	moves := impliedSerialGroup.MoveConfigs()

	progression := make([]string, len(moves))

	for i, move := range moves {
		progression[i] = move.Name()
	}

	installer := with.LegalMoveProgression(progression)

	for _, move := range moves {
		installer(move.CustomConfiguration())
	}

	return AddForPhase(phase, moves...)

}
