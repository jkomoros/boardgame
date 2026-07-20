package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
)

/*
HopAlongPath is a framework-provided FixUp that executes one hop of a
multi-hop path per application. It works in conjunction with MoveOnGraph,
which stores the computed path on the player's LocationBehavior.LocRemainingPath.

Each application is a separate version, which means a separate animation
cycle in the client's bundle queue (like DealCards). This produces smooth
hop-by-hop animation.

Legal while any LocationBehavior (on player or game state) has
len(LocRemainingPath) > 1.

Games should register HopAlongPath before other FixUps in ConfigureMoves()
to ensure hops complete before other FixUps fire.

boardgame:codegen
*/
type HopAlongPath struct {
	FixUp
}

// Legal returns nil if there is a LocationBehavior with remaining hops.
func (h *HopAlongPath) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := h.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	behavior := findBehaviorWithRemainingPath(state)
	if behavior == nil {
		return errors.New("no location behavior has remaining path hops")
	}

	if err := behavior.MayMoveTo(behavior.LocRemainingPath[1]); err != nil {
		return errors.New("cannot execute next path hop: " + err.Error())
	}

	return nil
}

// Apply moves the token one hop along the remaining path.
func (h *HopAlongPath) Apply(state boardgame.State) error {

	behavior := findBehaviorWithRemainingPath(state)
	if behavior == nil {
		return errors.New("no location behavior has remaining path hops")
	}

	// Move to the next hop
	nextIndex := behavior.LocRemainingPath[1]
	if err := behavior.MoveTo(nextIndex); err != nil {
		return err
	}

	// Consume the hop
	behavior.LocRemainingPath = behavior.LocRemainingPath[1:]

	return nil
}

// FallbackName returns "Hop Along Path"
func (h *HopAlongPath) FallbackName(m *boardgame.GameManager) string {
	return "Hop Along Path"
}

// FallbackHelpText returns a description of the FixUp.
func (h *HopAlongPath) FallbackHelpText() string {
	return "Execute one hop of a multi-hop movement path."
}

// findBehaviorWithRemainingPath scans all player states and the game state
// for a LocationBehavior that has remaining path hops (len > 1, meaning there
// is at least one more hop to execute).
func findBehaviorWithRemainingPath(state boardgame.ImmutableState) *behaviors.LocationBehavior {
	// Check player states
	for _, ps := range state.ImmutablePlayerStates() {
		if provider, ok := ps.(behaviors.HasLocationBehavior); ok {
			behavior := provider.GetLocationBehavior()
			if len(behavior.LocRemainingPath) > 1 {
				return behavior
			}
		}
	}

	// Check game state
	if provider, ok := state.ImmutableGameState().(behaviors.HasLocationBehavior); ok {
		behavior := provider.GetLocationBehavior()
		if len(behavior.LocRemainingPath) > 1 {
			return behavior
		}
	}

	return nil
}
