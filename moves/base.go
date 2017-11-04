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

//Legal checks whether the game's CurrentPhase (as determined by the delegate)
//is one of the LegalPhases for this moveType. A nil LegalPhases is
//interpreted as the move being legal in all phases. The string for the
//current phase will be based on the enum value of the PhaseEnum named by
//delegate.PhaseEnumName(), if it exists. Next, it checks to see if the give
//move is at a legal point in the move progression for this phase, if it
//exists.
func (d *Base) Legal(state boardgame.State) error {

	if err := d.legalInPhase(state); err != nil {
		return err
	}

	return d.legalMoveInProgression(state)

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

func (d *Base) legalMoveInProgression(state boardgame.State) error {
	currentPhase := state.Game().Manager().Delegate().CurrentPhase(state)

	legalMoveProgression := state.Game().Manager().Delegate().PhaseMoveProgression(currentPhase)

	//If there is no legal move progression then moves are legal in the phase at any time
	if legalMoveProgression == nil {
		return nil
	}

	//TODO: do proper progression support.
	return errors.New("Support for move progressions is not yet implemented")

}
