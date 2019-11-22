package behaviors

import (
	"github.com/jkomoros/boardgame"
)

/*
CurrentPlayerBehavior is a struct designed to be embedded anonymously in your
gameState. It encodes the current player for the game. base.GameDelegate's
CurrentPlayerIndex works well with this. It's named CurrentPlayerBehavior and
not CurrentPlayer because otherwise it would conflict with the internal property
name when accessing it from your SubState.
*/
type CurrentPlayerBehavior struct {
	CurrentPlayer boardgame.PlayerIndex
}

//SetCurrentPlayer sets the CurrentPlayer value to the given value. This
//satisfies the moves/interfaces.CurrentPlayerSetter interface, allowing you to
//use moves.FinishTurn.
func (c *CurrentPlayerBehavior) SetCurrentPlayer(currentPlayer boardgame.PlayerIndex) {
	c.CurrentPlayer = currentPlayer
}
