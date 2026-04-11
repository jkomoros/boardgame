package behaviors

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

/*
PlayerElimination is a struct designed to be embedded anonymously in your
PlayerStates. It tracks whether a player has been eliminated (knocked out) from
play. This is distinct from [InactivePlayer], which controls whether a player is
skipped in turn order.

PlayerElimination is scope-agnostic: the game decides when to set and clear the
flag.

  - Round-scoped (e.g. Love Letter): call ClearEliminated() at round start
  - Turn-scoped (e.g. Blackjack bust): call ClearEliminated() in ResetForTurnStart()
  - Game-scoped (permanent): never call ClearEliminated()

PlayerElimination does NOT automatically set the player as inactive. If
eliminated players should be skipped in turn order, the game should call both
SetEliminated() and SetPlayerInactive() explicitly.

Example:

	type playerState struct {
	    base.SubState
	    behaviors.PlayerElimination
	    behaviors.InactivePlayer
	}

	// In a move's Apply when a player is knocked out:
	p.SetEliminated()
*/
type PlayerElimination struct {
	Eliminated bool
}

// IsEliminated returns whether Eliminated is true. Satisfies the
// [interfaces.PlayerEliminator] interface.
func (p *PlayerElimination) IsEliminated() bool {
	return p.Eliminated
}

// SetEliminated marks the player as eliminated. Satisfies the
// [interfaces.PlayerEliminator] interface.
func (p *PlayerElimination) SetEliminated() {
	p.Eliminated = true
}

// ClearEliminated marks the player as not eliminated. Satisfies the
// [interfaces.PlayerEliminator] interface. Use this at round or turn boundaries
// for non-permanent elimination.
func (p *PlayerElimination) ClearEliminated() {
	p.Eliminated = false
}

// PlayerIsEliminated is a convenience method that does the cast to
// [interfaces.PlayerEliminator], so you don't have to. You can pass any
// playerState to it and it will return true if the player state implements
// [interfaces.PlayerEliminator] and IsEliminated returns true, false otherwise.
func PlayerIsEliminated(playerState boardgame.ImmutableSubState) bool {
	eliminator, ok := playerState.(interfaces.PlayerEliminator)
	if !ok {
		return false
	}
	return eliminator.IsEliminated()
}
