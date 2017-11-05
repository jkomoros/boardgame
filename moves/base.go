/*

moves is a convenience package that implements composable Moves to make it
easy to implement common logic. The Base move type is a very simple move that
implements the basic stubs necessary for your straightforward moves to have
minimal boilerplate.

*/
package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"strconv"
)

//Moves should implement AllowMultipleInProgression if they want to
//affirmatively communicate to moves.Base that in a move progression is it
//legal to apply multiple. If the move does not implement this interface then
//it is considered to only allow one.
type AllowMultipleInProgression interface {
	//AllowMultipleInProgression should return true if the given move is
	//allowed to apply multiple times in order in a move progression.
	AllowMultipleInProgression() bool
}

/*
Base is an optional, convenience struct designed to be embedded
anonymously in your own Moves. It implements no-op methods for many of the
required methods on Moves. Apply is not covered, because every Move
should implement their own, and if this implemented them it would obscure
errors where for example your Apply() was incorrectly named and thus not used.
In general your MoveConstructor can always be exactly the same, modulo the
name of your underlying move type:

	MoveConstructor: func() boardgame.Move {
 		return new(myMoveStruct)
	}

Base's Legal() method does basic checking for whehter the move is legal in
this phase, so your own Legal() method should always call Base.Legal() at the
top of its own method.

Base cannot help your move implement PropertyReadSetter; use autoreader to
generate that code for you.

*/
type Base struct {
	info *boardgame.MoveInfo
}

func (d *Base) SetInfo(m *boardgame.MoveInfo) {
	d.info = m
}

//Type simply returns BaseMove.MoveInfo
func (d *Base) Info() *boardgame.MoveInfo {
	return d.info
}

//DefaultsForState doesn't do anything
func (d *Base) DefaultsForState(state boardgame.State) {
	return
}

//Description defaults to returning the Type's HelpText()
func (d *Base) Description() string {
	return d.Info().Type().HelpText()
}

//ValidConfiguration always returns nil because there is no required
//configuration for moves.Base.
func (d *Base) ValidConfiguration() error {
	return nil
}

//Legal checks whether the game's CurrentPhase (as determined by the delegate)
//is one of the LegalPhases for this moveType. A nil LegalPhases is
//interpreted as the move being legal in all phases. The string for the
//current phase will be based on the enum value of the PhaseEnum named by
//delegate.PhaseEnumName(), if it exists. Next, it checks to see if the give
//move is at a legal point in the move progression for this phase, if it
//exists. Each move in the move progression must show up 1 or more times. The
//method checks to see if we were to make this move, would the moves since the
//last phase change match the pattern?
func (d *Base) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := d.legalInPhase(state); err != nil {
		return err
	}

	return d.legalMoveInProgression(state, proposer)

}

//legalInPhase will return a descriptive error if this move is not legal in
//the current phase of hte game.
func (d *Base) legalInPhase(state boardgame.State) error {

	legalPhases := d.Info().Type().LegalPhases()

	if len(legalPhases) == 0 {
		return nil
	}

	currentPhase := state.Game().Manager().Delegate().CurrentPhase(state)

	for _, phase := range legalPhases {
		if phase == currentPhase {
			return nil
		}
	}

	phaseName := strconv.Itoa(currentPhase)

	if phaseEnum := state.Game().Manager().Delegate().PhaseEnum(); phaseEnum != nil {
		phaseName = phaseEnum.String(currentPhase)
	}

	return errors.New("Move is not legal in phase " + phaseName)
}

func (d *Base) legalMoveInProgression(state boardgame.State, proposer boardgame.PlayerIndex) error {
	currentPhase := state.Game().Manager().Delegate().CurrentPhase(state)

	pattern := state.Game().Manager().Delegate().PhaseMoveProgression(currentPhase)

	//If there is no legal move progression then moves are legal in the phase at any time
	if pattern == nil {
		return nil
	}

	historicalMoves := state.Game().HistoricalMovesSincePhaseTransition(state.Version())

	progression := make([]string, len(historicalMoves))

	for i, move := range historicalMoves {
		progression[i] = move.Name
	}

	//If we were to add our target move to the historical progression, would it match the pattern?
	if !progressionMatches(append(progression, d.Info().Type().Name()), pattern) {
		return errors.New("This move is not legal at this point in the current phase.")
	}

	//Are we a new type of move in the progression? if so, is the move before
	//us still legal?

	if len(historicalMoves) == 0 {
		//We're the first move, it's fine.
		return nil
	}

	lastMoveRecord := historicalMoves[len(historicalMoves)-1]

	if lastMoveRecord.Name == d.Info().Type().Name() {

		//We're applying multiple in a row. Is that legal?

		//We can't check ourselves because we're embedded in the real move type.
		testMove := d.Info().Type().NewMove(state)

		allowMultiple, ok := testMove.(AllowMultipleInProgression)

		if !ok || !allowMultiple.AllowMultipleInProgression() {
			return errors.New("This move was just applied and is not configured to allow multiple in a row in this phase.")
		}

		return nil
	}

	lastMoveType := state.Game().Manager().FixUpMoveTypeByName(lastMoveRecord.Name)

	if lastMoveType == nil {
		lastMoveType = state.Game().Manager().PlayerMoveTypeByName(lastMoveRecord.Name)
	}

	if lastMoveType == nil {
		return errors.New("Unexpected error: couldn't find a historical move type")
	}

	//LastMove will have all of the defaults set.
	lastMove := lastMoveType.NewMove(state)

	if lastMove.Legal(state, proposer) == nil {
		return errors.New("A move that needs to happen earlier in the phase is still legal to apply.")
	}

	return nil

}

//progressionMatches returns true if the given history matches the pattern.
func progressionMatches(input []string, pattern []string) bool {

	inputPosition := 0
	patternPosition := 0

	for inputPosition < len(input) {

		inputItem := input[inputPosition]
		patternItem := pattern[patternPosition]

		if inputItem != patternItem {
			//Perhaps we just passed to the next part of the pattern?

			//that's not legal at the very front of input
			if inputPosition == 0 {
				return false
			}

			patternPosition++

			if patternPosition >= len(pattern) {
				//No more pattern, I guess we didn't match.
				return false
			}

			patternItem = pattern[patternPosition]

			if inputItem != patternItem {
				//Nope, we didn't match the next part of the pattern, we just don't match
				return false
			}

		}

		inputPosition++

	}

	//If we got to the end of the input without invalidating then it passes.
	return true

}
