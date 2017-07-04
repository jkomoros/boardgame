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
)

/*
Base is an optional, convenience struct designed to be embedded
anonymously in your own Moves. It implements no-op methods for many of the
required methods on Moves. Legal and Apply are not covered, because every Move
should implement their own, and if this implemented them it would obscure
errors where for example your Legal() was incorrectly named and thus not used.
In general your MoveConstructor can always be exactly the same, modulo the
name of your underlying move type:

	MoveConstructor: func() boardgame.Move {
 		return new(myMoveStruct)
	}

Base cannot help your move implement PropertyReadSetter; use autoreader to
generate that code for you.

*/
type Base struct {
	moveType *boardgame.MoveType
}

func (d *Base) SetType(m *boardgame.MoveType) {
	d.moveType = m
}

//Type simply returns BaseMove.MoveType
func (d *Base) Type() *boardgame.MoveType {
	return d.moveType
}

//DefaultsForState doesn't do anything
func (d *Base) DefaultsForState(state boardgame.State) {
	return
}

//Description defaults to returning the Type's HelpText()
func (d *Base) Description() string {
	return d.Type().HelpText()
}

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

	currentPlayer := state.Game().CurrentPlayerIndex()

	if !c.TargetPlayerIndex.Valid(state) {
		return errors.New("The specified target player is not valid")
	}

	if c.TargetPlayerIndex < 0 {
		return errors.New("The specified target player is not valid")
	}

	if !c.TargetPlayerIndex.Equivalent(currentPlayer) {
		return errors.New("It's not your turn")
	}

	if !c.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("It's not your turn")
	}

	return nil

}

//DefaultsForState will set the TargetPlayerIndex to be the CurrentPlayerIndex.
func (c *CurrentPlayer) DefaultsForState(state boardgame.State) {
	c.TargetPlayerIndex = state.Game().CurrentPlayerIndex()
}
