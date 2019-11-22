package behaviors

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

/*
PhaseBehavior is a struct designed to be embedded anonymously in your gameState.
It encodes the current phase for the game. base.GameDelegate's CurrentPhase
works well with this. It expects your phase enum to be named 'phase'. It's named
PhaseBehavior and not Phase because otherwise it would conflict with the
internal property name when accessing it from your SubState.
*/
type PhaseBehavior struct {
	Phase enum.Val `enum:"phase"`
}

//SetCurrentPhase sets the phase value to the given value. This
//satisfies the moves/interfaces.CurrentPhaseSetter interface, allowing you to
//use moves.StartPhase.
func (p *PhaseBehavior) SetCurrentPhase(phase int) {
	p.Phase.SetValue(phase)
}

//ConnectBehavior doesn't do anything
func (p *PhaseBehavior) ConnectBehavior(containingStruct boardgame.SubState) {
	//Pass
}
