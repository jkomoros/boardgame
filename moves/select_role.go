package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

/*
SelectRole is a player move that allows a seated player to choose their role.
It embeds [AnyPlayer], so any seated player can propose it for themselves during
any phase where it is legal.

The player state must embed [behaviors.PlayerRole], and the game must have a
"role" enum in its chest.

If configured with [WithUnique], Legal will reject values already claimed by
another seated player (for games like Spirit Island where each player must have
a unique spirit/role). By default, duplicate roles are allowed (for games like
Captain Sonar where roles are unique per team but shared across teams, validated
via [boardgame.GameDelegate.ReadyToStart]).

Any valid value in the "role" enum is accepted. See [SelectTeam] for the
sentinel convention if your game needs an "unset" value.

boardgame:codegen
*/
type SelectRole struct {
	AnyPlayer
	SelectedRole enum.Val `enum:"role"`
}

// Legal verifies the parent AnyPlayer checks pass, that SelectedRole is set and
// valid, that it belongs to the correct enum, and optionally enforces uniqueness.
func (s *SelectRole) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := s.AnyPlayer.Legal(state, proposer); err != nil {
		return err
	}

	if s.SelectedRole == nil {
		return errors.New("no role selected")
	}

	roleEnum := state.Manager().Chest().Enums().Enum("role")
	if roleEnum == nil {
		return errors.New("no 'role' enum found")
	}
	if s.SelectedRole.Enum() != roleEnum {
		return errors.New("selected role value is from a different enum")
	}
	if !roleEnum.Valid(s.SelectedRole.Value()) {
		return errors.New("selected role is not a valid value")
	}

	// If WithUnique is configured, check no other seated player has this role
	if s.isUnique() {
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
			if roleHolder, ok := p.(behaviors.HasPlayerRole); ok {
				if roleHolder.GetPlayerRole().Role.Value() == s.SelectedRole.Value() {
					return errors.New("another player already has that role")
				}
			}
		}
	}

	return nil
}

// isUnique returns true if WithUnique was configured.
func (s *SelectRole) isUnique() bool {
	config := s.CustomConfiguration()
	val, ok := config[configPropUnique]
	if !ok {
		return false
	}
	boolVal, ok := val.(bool)
	return ok && boolVal
}

// Apply sets the target player's role to the selected value.
func (s *SelectRole) Apply(state boardgame.State) error {
	target := s.TargetPlayerIndex
	player := state.PlayerStates()[target]

	roleHolder, ok := player.(behaviors.HasPlayerRole)
	if !ok {
		return errors.New("player state does not implement HasPlayerRole")
	}

	roleHolder.GetPlayerRole().Role.SetValue(s.SelectedRole.Value())
	return nil
}

// ValidConfiguration checks that the player state implements HasPlayerRole and
// that a "role" enum exists in the chest.
func (s *SelectRole) ValidConfiguration(exampleState boardgame.State) error {
	if err := s.AnyPlayer.ValidConfiguration(exampleState); err != nil {
		return err
	}

	playerState := exampleState.ImmutablePlayerStates()[0]
	if _, ok := playerState.(behaviors.HasPlayerRole); !ok {
		return errors.New("player state does not implement HasPlayerRole. Embed behaviors.PlayerRole in your player state")
	}

	if exampleState.Manager().Chest().Enums().Enum("role") == nil {
		return errors.New("no 'role' enum found in the chest. Define an enum named 'role'")
	}

	return nil
}

// FallbackName returns "Select Role"
func (s *SelectRole) FallbackName(m *boardgame.GameManager) string {
	return "Select Role"
}

// FallbackHelpText returns a description of the move.
func (s *SelectRole) FallbackHelpText() string {
	return "Choose your role for this game."
}
