package behaviors

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

/*
MoveBudget is a struct designed to be embedded anonymously in your PlayerStates.
It tracks how many actions a player has remaining in their current turn. Games
that use a simple one-action-per-turn model set the budget to 1; games with
multiple actions per turn use a higher value.

The bool-action pattern (e.g. "has this player made their move?") is handled by
a budget of 1: HasMovesLeft() returns true when MovesLeft > 0, which is
equivalent to "has not yet acted."

Example:

	type playerState struct {
	    base.SubState
	    behaviors.MoveBudget
	}

	func (p *playerState) ResetForTurnStart() error {
	    p.ResetMovesTo(1) // or 2, or variable
	    return nil
	}

	// In a move's Legal:
	if !p.HasMovesLeft() { return errors.New("no actions remaining") }

	// In a move's Apply:
	p.ConsumeMove()
*/
type MoveBudget struct {
	MovesLeft int
}

// HasMovesLeft returns true if MovesLeft > 0. Satisfies the
// [interfaces.TurnBudgeter] interface. Use this in Legal() checks.
func (m *MoveBudget) HasMovesLeft() bool {
	return m.MovesLeft > 0
}

// ConsumeMove decrements MovesLeft by 1. Satisfies the
// [interfaces.TurnBudgeter] interface. Call this in Apply() after the action
// succeeds. The caller should check HasMovesLeft() in Legal() first.
func (m *MoveBudget) ConsumeMove() {
	m.MovesLeft--
}

// ResetMovesTo sets MovesLeft to the given value. Satisfies the
// [interfaces.TurnBudgeter] interface. Call this in ResetForTurnStart() with
// the appropriate budget for the game.
func (m *MoveBudget) ResetMovesTo(n int) {
	m.MovesLeft = n
}

// PlayerHasMovesLeft is a convenience function that returns true if the given
// player state implements [interfaces.TurnBudgeter] and HasMovesLeft() returns
// true. Returns false otherwise.
func PlayerHasMovesLeft(playerState boardgame.ImmutableSubState) bool {
	budgeter, ok := playerState.(interfaces.TurnBudgeter)
	if !ok {
		return false
	}
	return budgeter.HasMovesLeft()
}
