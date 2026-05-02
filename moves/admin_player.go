package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
)

/*
AdminPlayer is a convenience embeddable move for moves that only the game
administrator can make. It is the admin-restricted parallel of [AnyPlayer] and
[CurrentPlayer]:

  - [CurrentPlayer] — only the current player can propose
  - [AnyPlayer] — any seated player can propose (for themselves)
  - AdminPlayer — only the game admin can propose (for themselves)

The player state must embed [behaviors.GameAdministrator] for admin checks to
work. If the player state does not embed it, AdminPlayer behaves identically to
[AnyPlayer] (all seated players can propose).

Like [AnyPlayer], the target player is encoded as TargetPlayerIndex, which
defaults to ObserverPlayerIndex as a sentinel and is auto-corrected to the
proposer in Legal.

boardgame:codegen
*/
type AdminPlayer struct {
	AnyPlayer
}

// Legal checks all AnyPlayer constraints (seated, proposer matches target),
// then additionally verifies that the target player is the game administrator.
func (a *AdminPlayer) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := a.AnyPlayer.Legal(state, proposer); err != nil {
		return err
	}

	target := a.TargetPlayerIndex
	player := state.ImmutablePlayerStates()[target]

	// Only enforce admin check if the playerState has the GameAdministrator
	// behavior. Games without it get AnyPlayer behavior (no restriction).
	if _, ok := player.(behaviors.HasGameAdministrator); ok {
		if !behaviors.PlayerIsAdmin(player) {
			return errors.New("only the game administrator can make this move")
		}
	}

	return nil
}

// FallbackName returns "Admin Player Move"
func (a *AdminPlayer) FallbackName(m *boardgame.GameManager) string {
	return "Admin Player Move"
}

// FallbackHelpText returns a description of the move.
func (a *AdminPlayer) FallbackHelpText() string {
	return "A move that only the game administrator can make."
}
