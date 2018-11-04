package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"strconv"
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

	phaseInstaller := WithLegalPhases(phase)

	for _, move := range moves {
		phaseInstaller(move.CustomConfiguration())
	}

	return moves

}

//errorMove is a special type of move used entirely to signal an error in
//ValidConfiguration.
//
//boardgame:codegen
type errorMove struct {
	NoOp
	Message string
}

type moveFactory func() boardgame.Move

func errorMoveWithMessage(message string) moveFactory {
	return func() boardgame.Move {
		return &errorMove{
			Message: message,
		}
	}
}

func (e *errorMove) ValidConfiguration(exampleState boardgame.State) error {
	return errors.New("Error earlier in move processing: " + e.Message)
}

//AddOrderedForPhase is designed to be used within Combine. It calls
//WithLegalPhases() and also WithLegalMoveProgression() on the config for each
//config passed in, which means that the moves' Legal() will only be Legal in
//that phase, in that point in the move progression. It's a convenience to
//make it less error-prone and more clear what the intent is for phase-locked,
//ordered moves. All moveTypes passed must be legal auto-configurable moves.
//You may pass configs generated from AutoConfigurer.Config(), or any of the
//MoveProgressionGroup types defined in this package. All of the top level
//groups passed will be treated implicitly like a single Serial group. All
//moves contained within the provided groups will be registered. If your
//PhaseEnum is a Tree, then phase must be a leaf enum value, or the moves will
//fail to pass the ValidConfiguration check. This will also sanity check that
//the last Move enumerated is a StartPhase move, which is almost always what
//you want, and omission is likely an error. Check out the package doc for an
//example of using groups in this function.
func AddOrderedForPhase(phase int, groups ...MoveProgressionGroup) []boardgame.MoveConfig {

	//Technically it's illegal to attach a move progression to a non-leaf
	//phase enum val, but at this point we don't have a reference to delegate
	//so we can't check. moves.Default.ValidConfiguration will check.

	if len(groups) == 0 {
		return nil
	}

	lastGroup := groups[len(groups)-1]
	lastGroupItems := lastGroup.MoveConfigs()
	if len(lastGroupItems) == 0 {
		return nil
	}

	lastGroupLastItem := lastGroupItems[len(lastGroupItems)-1]

	lastGroupLastItemStruct := lastGroupLastItem.Constructor()()
	if lastGroupLastItemStruct == nil {
		return nil
	}

	if _, ok := lastGroupLastItemStruct.(phaseToStarter); !ok {
		//It's not a StartPhase. If it's a NoOp, that's our signal it was
		//intentional and we should let it slide.
		if _, noOpOk := lastGroupLastItemStruct.(isNoOper); !noOpOk {
			//Not a no op either. Error and tell them how to signal it was
			//intentional.

			//TODO: in the future it'd be nice if we could use the human-
			//readable name for the phase here.
			message := "The end of your phase run for phase " + strconv.Itoa(phase) + " did not end with a StartPhase move, which is typical. If this was intentional, end that phase with a moves.NoOp to override this error."

			return []boardgame.MoveConfig{
				boardgame.NewMoveConfig("AddOrderedForPhase Error", errorMoveWithMessage(message), nil),
			}
		}
	}

	//Every move in the phase shares the same group to match against. That's
	//because the group matches as long as the whole tape is consumed, and a
	//move tests if it is legal by speculatively adding itself to the
	//historical tape and seing if the progression still matches. This means
	//that the same progression can be shared.
	impliedSerialGroup := Serial(groups...)

	moves := impliedSerialGroup.MoveConfigs()

	installer := WithLegalMoveProgression(impliedSerialGroup)

	for _, move := range moves {
		installer(move.CustomConfiguration())
	}

	return AddForPhase(phase, moves...)

}
