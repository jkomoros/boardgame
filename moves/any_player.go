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

The target player is encoded as TargetPlayerIndex. DefaultsForState sets this to
ObserverPlayerIndex as a sentinel. If the proposer is a concrete player and the
target is still the sentinel, Legal auto-corrects the target to the proposer.
This means agents work without manually setting TargetPlayerIndex, matching
the ergonomics of [CurrentPlayer].

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

// moveInputAnyPlayerBehavior is an unshadowable package-private marker used by
// auto.Config to recognize this embedded behavior.
func (a *AnyPlayer) moveInputAnyPlayerBehavior() {}

// Legal checks that the proposer is making a move for themselves and that their
// seat is filled.
func (a *AnyPlayer) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := a.Default.Legal(state, proposer); err != nil {
		return err
	}

	target := a.TargetPlayerIndex

	// If the target is still the default sentinel (ObserverPlayerIndex) and
	// the proposer is a concrete player, auto-correct to the proposer. This
	// makes agents "just work" without manually setting TargetPlayerIndex,
	// matching the ergonomics of CurrentPlayer.DefaultsForState.
	if target == boardgame.ObserverPlayerIndex && proposer >= 0 {
		target = proposer
		a.TargetPlayerIndex = proposer
	}

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

// DefaultsForState sets TargetPlayerIndex to ObserverPlayerIndex as a sentinel.
// If the caller forgets to set it and the proposer is a concrete player, Legal
// auto-corrects the target to the proposer (making agents work without manual
// setup). If the proposer is not a concrete player (e.g., AdminPlayerIndex),
// the sentinel remains and Legal rejects the move with a clear error.
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
