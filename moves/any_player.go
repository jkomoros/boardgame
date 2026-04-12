package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

/*
AnyPlayer is a convenience embeddable move for moves that any seated player can
make for themselves. It is the gathering-phase equivalent of [CurrentPlayer]:
instead of checking that the proposer is the current player, it checks that the
proposer is a valid seated player making a move on their own behalf.

The target player is encoded as TargetPlayerIndex. Unlike CurrentPlayer,
DefaultsForState sets this to ObserverPlayerIndex as a fail-safe sentinel. The
client must provide the actual target player index when proposing the move.
Agents must explicitly set TargetPlayerIndex to their own player index.

Legal checks that TargetPlayerIndex is a concrete player (>= 0), that the
proposer matches the target (using Equivalent, which allows AdminPlayerIndex
to act on behalf of any player), and that the target's seat is filled (if the
game uses seating).

boardgame:codegen
*/
type AnyPlayer struct {
	Default
	TargetPlayerIndex boardgame.PlayerIndex
}

// Legal checks that the proposer is making a move for themselves and that their
// seat is filled.
func (a *AnyPlayer) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := a.Default.Legal(state, proposer); err != nil {
		return err
	}

	target := a.TargetPlayerIndex

	// Target must be a concrete player index (not a sentinel like Observer or Any)
	if target < 0 {
		return errors.New("target player must be a seated player, not a special index")
	}

	// Target must be within bounds
	if int(target) >= len(state.ImmutablePlayerStates()) {
		return errors.New("target player index is out of bounds")
	}

	// Proposer must match target (self-selection) OR be admin.
	// Equivalent allows AdminPlayerIndex to match any concrete index.
	if !target.Equivalent(proposer) {
		return errors.New("you can only make this move for yourself")
	}

	// Target's seat must be filled (if the game uses seating)
	player := state.ImmutablePlayerStates()[target]
	if seater, ok := player.(interfaces.Seater); ok {
		if !seater.SeatIsFilled() {
			return errors.New("your seat is not yet filled")
		}
	}

	return nil
}

// DefaultsForState sets TargetPlayerIndex to ObserverPlayerIndex as a fail-safe.
// If the caller (client or agent) forgets to set it, Legal will reject the move
// rather than silently targeting player 0.
func (a *AnyPlayer) DefaultsForState(state boardgame.ImmutableState) {
	a.TargetPlayerIndex = boardgame.ObserverPlayerIndex
}

// FallbackName returns "Any Player Move"
func (a *AnyPlayer) FallbackName(m *boardgame.GameManager) string {
	return "Any Player Move"
}

// FallbackHelpText returns a description of the move.
func (a *AnyPlayer) FallbackHelpText() string {
	return "A move that any seated player can make for themselves."
}
