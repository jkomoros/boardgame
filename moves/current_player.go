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
return boardgame.ObserverPlayerIndex. If there are simultaneous phases where
anyone may move, return boardgame.AnyPlayerIndex (which acts as a wildcard in
Equivalent() but does not grant omniscient access to hidden state, unlike
boardgame.AdminPlayerIndex which should be reserved for engine-initiated
actions like fix-up moves and timers).

Typically you'd implement your own Legal method that calls
CurrentPlayer.Legal() first, then do your own specific checking after that,
too.

boardgame:codegen
*/
type CurrentPlayer struct {
	Default
	TargetPlayerIndex boardgame.PlayerIndex
}

// moveInputCurrentPlayerBehavior is an unshadowable package-private marker
// used by auto.Config to recognize this embedded behavior.
func (c *CurrentPlayer) moveInputCurrentPlayerBehavior() {}

// Legal will return an error if the TargetPlayerIndex is not the
// CurrentPlayerIndex, if the TargetPlayerIndex is not equivalent to the
// proposer, or if the TargetPlayerIndex is not one of the players.
func (c *CurrentPlayer) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := c.Default.Legal(state, proposer); err != nil {
		return err
	}
	if manager := state.Manager(); manager != nil {
		// Default.Legal already evaluated the assembled plan, including this
		// type's contributed proposer predicate. During boot probing it also
		// intentionally returns before ordinary legality. Do not run the frozen
		// imperative copy a second time in either case.
		if manager.LegalProbeActive() || manager.LegalPlanAssembled(c.Name()) {
			return nil
		}
	}

	currentPlayer := state.CurrentPlayerIndex()

	targetPlayerIndex := c.TargetPlayerIndex.EnsureValid(state)

	if !targetPlayerIndex.Valid(state) {
		return errors.New("The specified target player is not valid")
	}

	if targetPlayerIndex < 0 {
		return errors.New("The specified target player is not valid")
	}

	if !targetPlayerIndex.Equivalent(currentPlayer) {
		return errors.New("it's not your turn")
	}

	if !targetPlayerIndex.Equivalent(proposer) {
		return errors.New("it's not your turn")
	}

	return nil

}

// DefaultsForState will set the TargetPlayerIndex to be the CurrentPlayerIndex.
func (c *CurrentPlayer) DefaultsForState(state boardgame.ImmutableState) {
	c.TargetPlayerIndex = state.CurrentPlayerIndex().EnsureValid(state)
}

// FallbackName returns "Current Player Move"
func (c *CurrentPlayer) FallbackName(m *boardgame.GameManager) string {
	return "Current Player Move"
}

// FallbackHelpText returns "A move by the current player."
func (c *CurrentPlayer) FallbackHelpText() string {
	return "A move by the current player."
}
