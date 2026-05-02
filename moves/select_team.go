package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

/*
SelectTeam is a player move that allows a seated player to choose their team.
It embeds [AnyPlayer], so any seated player can propose it for themselves during
any phase where it is legal.

The player state must embed [behaviors.PlayerTeam], and the game must have a
"team" enum in its chest.

Any valid value in the "team" enum is accepted. If your game needs an "unset"
sentinel to detect players who haven't picked a team yet, define your enum
with a sentinel first value:

	const (
	    teamUnset = iota // sentinel: no team selected yet
	    teamRed
	    teamBlue
	)

Then check for the sentinel in your [boardgame.GameDelegate.ReadyToStart]
implementation.

boardgame:codegen
*/
type SelectTeam struct {
	AnyPlayer
	SelectedTeam enum.Val `enum:"team"`
}

// Legal verifies the parent AnyPlayer checks pass, that SelectedTeam is set and
// valid, and that it belongs to the correct enum.
func (s *SelectTeam) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := s.AnyPlayer.Legal(state, proposer); err != nil {
		return err
	}

	if s.SelectedTeam == nil {
		return errors.New("no team selected")
	}

	// Verify the value belongs to the correct enum (prevents cross-enum corruption)
	teamEnum := state.Manager().Chest().Enums().Enum("team")
	if teamEnum == nil {
		return errors.New("no 'team' enum found")
	}
	if s.SelectedTeam.Enum() != teamEnum {
		return errors.New("selected team value is from a different enum")
	}

	// Verify the value is valid within the enum
	if !teamEnum.Valid(s.SelectedTeam.Value()) {
		return errors.New("selected team is not a valid value")
	}

	// If WithUnique is configured, check no other seated player has this team
	if s.isUniqueEnforced() {
		target := s.TargetPlayerIndex
		for i, p := range state.ImmutablePlayerStates() {
			if boardgame.PlayerIndex(i) == target {
				continue
			}
			if seater, ok := p.(interfaces.Seater); ok {
				if !seater.SeatIsFilled() {
					continue
				}
			}
			if teamHolder, ok := p.(behaviors.HasPlayerTeam); ok {
				if teamHolder.GetPlayerTeam().Team.Value() == s.SelectedTeam.Value() {
					return errors.New("another player already has that team")
				}
			}
		}
	}

	return nil
}

// isUniqueEnforced returns true if uniqueness should be enforced.
// SelectTeam defaults to non-unique (multiple players on one team is normal).
func (s *SelectTeam) isUniqueEnforced() bool {
	return selectionIsUnique(s.CustomConfiguration(), false)
}

// Apply sets the target player's team to the selected value.
func (s *SelectTeam) Apply(state boardgame.State) error {
	target := s.TargetPlayerIndex
	player := state.PlayerStates()[target]

	teamHolder, ok := player.(behaviors.HasPlayerTeam)
	if !ok {
		return errors.New("player state does not implement HasPlayerTeam")
	}

	teamHolder.GetPlayerTeam().Team.SetValue(s.SelectedTeam.Value())
	return nil
}

// ValidConfiguration checks that the player state implements HasPlayerTeam and
// that a "team" enum exists in the chest.
func (s *SelectTeam) ValidConfiguration(exampleState boardgame.State) error {
	if err := s.AnyPlayer.ValidConfiguration(exampleState); err != nil {
		return err
	}

	playerState := exampleState.ImmutablePlayerStates()[0]
	if _, ok := playerState.(behaviors.HasPlayerTeam); !ok {
		return errors.New("player state does not implement HasPlayerTeam. Embed behaviors.PlayerTeam in your player state")
	}

	if exampleState.Manager().Chest().Enums().Enum("team") == nil {
		return errors.New("no 'team' enum found in the chest. Define an enum named 'team'")
	}

	return nil
}

// FallbackName returns "Select Team"
func (s *SelectTeam) FallbackName(m *boardgame.GameManager) string {
	return "Select Team"
}

// FallbackHelpText returns a description of the move.
func (s *SelectTeam) FallbackHelpText() string {
	return "Choose which team to join."
}
