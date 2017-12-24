package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

/*

CurrentPlayer is a convenience embeddable move that represents a move made by
the CurrentPlayer.

The target player is encoded as TargetPlayerIndex. This is checked to make
sure it is equivalent to the delegate's CurrentPlayerIndex, as well as to the
proposer. This means that your Delegate should return a reasonable result from
CurrentPlayerIndex. If your game has different rounds where no one may move,
return boardgame.ObserverPlayerIndex. If there are rounds where anyone may
move, return boardgame.AdminPlayerIndex.

Typically you'd implement your own Legal method that calls
CurrentPlayer.Legal() first, then do your own specific checking after that,
too.

*/
type CurrentPlayer struct {
	Base
	TargetPlayerIndex boardgame.PlayerIndex
}

//Legal will return an error if the TargetPlayerIndex is not the
//CurrentPlayerIndex, if the TargetPlayerIndex is not equivalent to the
//proposer, or if the TargetPlayerIndex is not one of the players.
func (c *CurrentPlayer) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := c.Base.Legal(state, proposer); err != nil {
		return err
	}

	currentPlayer := state.CurrentPlayerIndex()

	if !c.TargetPlayerIndex.Valid(state) {
		return errors.New("The specified target player is not valid")
	}

	if c.TargetPlayerIndex < 0 {
		return errors.New("The specified target player is not valid")
	}

	if !c.TargetPlayerIndex.Equivalent(currentPlayer) {
		return errors.New("It's not your turn!")
	}

	if !c.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("It's not your turn!")
	}

	return nil

}

//DefaultsForState will set the TargetPlayerIndex to be the CurrentPlayerIndex.
func (c *CurrentPlayer) DefaultsForState(state boardgame.State) {
	c.TargetPlayerIndex = state.CurrentPlayerIndex()
}

func (c *CurrentPlayer) MoveTypeFallbackName(manager *boardgame.GameManager) string {
	return "Current Player Move"
}

func (c *CurrentPlayer) MoveTypeHelpText(manager *boardgame.GameManager) string {
	return "A move by the current player."
}

func (c *CurrentPlayer) MoveTypeIsFixUp(manager *boardgame.GameManager) bool {
	return false
}
