package moves

import (
	"errors"
	"strconv"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

// AllPlayersSubmitted is a FixUp move that extends [StartPhase]. It becomes
// legal when all active (non-inactive) players have submitted their selection
// (as tracked by [interfaces.PlayerSubmitter]). When it fires, it transitions
// the game to the next phase via StartPhase.Apply, which calls
// BeforeLeavePhase, BeforeEnterPhase, and SetCurrentPhase.
//
// Use this in a simultaneous selection phase: configure it with
// [WithPhaseToStart] pointing at the reveal/resolution phase. Player-facing
// moves call SetPlayerSubmitted() on each player's state. Once all active
// players have submitted, this FixUp fires and advances the phase.
//
// Games that use BeforeLeavePhase on their gameState can perform reveal or
// scoring logic there, since it runs as part of StartPhase.Apply before the
// phase actually changes.
//
//boardgame:codegen
type AllPlayersSubmitted struct {
	StartPhase
}

// ValidConfiguration checks that the embedding move satisfies StartPhase's
// configuration and that playerState implements interfaces.PlayerSubmitter.
func (a *AllPlayersSubmitted) ValidConfiguration(exampleState boardgame.State) error {
	if err := a.StartPhase.ValidConfiguration(exampleState); err != nil {
		return err
	}
	if _, ok := exampleState.PlayerStates()[0].(interfaces.PlayerSubmitter); !ok {
		return errors.New("PlayerState does not implement interfaces.PlayerSubmitter. behaviors.PlayerSubmission implements it for free")
	}
	return nil
}

// Legal checks that the move is legal in this phase (via StartPhase.Legal) and
// that all active players have submitted.
func (a *AllPlayersSubmitted) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := a.StartPhase.Legal(state, proposer); err != nil {
		return err
	}
	for i, player := range state.ImmutablePlayerStates() {
		if behaviors.PlayerIsInactive(player) {
			continue
		}
		submitter, ok := player.(interfaces.PlayerSubmitter)
		if !ok {
			return errors.New("player " + strconv.Itoa(i) + " does not implement PlayerSubmitter")
		}
		if !submitter.HasSubmitted() {
			return errors.New("player " + strconv.Itoa(i) + " has not yet submitted")
		}
	}
	return nil
}

// FallbackName returns "All Players Submitted"
func (a *AllPlayersSubmitted) FallbackName(m *boardgame.GameManager) string {
	return "All Players Submitted"
}

// FallbackHelpText returns a description of the move.
func (a *AllPlayersSubmitted) FallbackHelpText() string {
	return "Advances the phase once all active players have submitted their selection"
}
