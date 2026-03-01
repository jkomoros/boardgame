package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

// ResetAllPlayerSubmissions is a FixUp move that resets all players'
// submission flags to false. It is legal when any player currently has their
// PlayerSubmitted flag set to true.
//
// Typically placed at the start of a simultaneous selection phase (or in the
// reveal phase) to clear flags from a previous round before players begin
// making new selections.
//
//boardgame:codegen
type ResetAllPlayerSubmissions struct {
	FixUp
}

// ValidConfiguration checks that playerState implements
// interfaces.PlayerSubmitter.
func (r *ResetAllPlayerSubmissions) ValidConfiguration(exampleState boardgame.State) error {
	if err := r.FixUp.ValidConfiguration(exampleState); err != nil {
		return err
	}
	if _, ok := exampleState.PlayerStates()[0].(interfaces.PlayerSubmitter); !ok {
		return errors.New("PlayerState does not implement interfaces.PlayerSubmitter. behaviors.PlayerSubmission implements it for free")
	}
	return nil
}

// Legal checks that the move is legal in this phase and that at least one
// player has a pending submission to reset.
func (r *ResetAllPlayerSubmissions) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := r.FixUp.Legal(state, proposer); err != nil {
		return err
	}
	for _, player := range state.ImmutablePlayerStates() {
		if behaviors.PlayerHasSubmitted(player) {
			return nil
		}
	}
	return errors.New("no players have pending submissions to reset")
}

// Apply resets all players' submission flags to false.
func (r *ResetAllPlayerSubmissions) Apply(state boardgame.State) error {
	for _, player := range state.PlayerStates() {
		if submitter, ok := player.(interfaces.PlayerSubmitter); ok {
			submitter.ResetSubmission()
		}
	}
	return nil
}

// FallbackName returns "Reset All Player Submissions"
func (r *ResetAllPlayerSubmissions) FallbackName(m *boardgame.GameManager) string {
	return "Reset All Player Submissions"
}

// FallbackHelpText returns a description of the move.
func (r *ResetAllPlayerSubmissions) FallbackHelpText() string {
	return "Resets all players' submission flags to prepare for a new simultaneous selection round"
}
