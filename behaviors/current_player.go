package behaviors

import (
	"github.com/jkomoros/boardgame"
)

/*
CurrentPlayer is a struct designed to be embedded anonymously in your gameState.
It encodes the current player for the game. base.GameDelegate's
CurrentPlayerIndex works well with this.
*/
type CurrentPlayer struct {
	CurrentPlayer boardgame.PlayerIndex
}

//SetCurrentPlayer sets the CurrentPlayer value to the given value. This
//satisfies the moves/interfaces.CurrentPlayerSetter interface, allowing you to
//use moves.FinishTurn.
func (c *CurrentPlayer) SetCurrentPlayer(currentPlayer boardgame.PlayerIndex) {
	c.CurrentPlayer = currentPlayer
}

//ConnectBehavior doesn't do anything
func (c *CurrentPlayer) ConnectBehavior(containingStruct boardgame.SubState) {
	//Pass
}
