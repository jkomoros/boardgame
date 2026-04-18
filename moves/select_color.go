package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

/*
SelectColor is a player move that allows a seated player to choose their color.
It embeds [AnyPlayer], so any seated player can propose it for themselves during
any phase where it is legal.

The player state must embed [behaviors.PlayerColor], and the game must have a
"color" enum in its chest.

By default, SelectColor enforces uniqueness: no two seated players may share the
same color. This is the safe default since in virtually every real game with
player colors, colors are unique. To disable uniqueness enforcement, configure
with [WithAllowDuplicates].

Any valid value in the "color" enum is accepted. See [SelectTeam] for the
sentinel convention if your game needs an "unset" value.

boardgame:codegen
*/
type SelectColor struct {
	AnyPlayer
	SelectedColor enum.Val `enum:"color"`
}

// Legal verifies the parent AnyPlayer checks pass, that SelectedColor is set and
// valid, that it belongs to the correct enum, and enforces uniqueness unless
// WithAllowDuplicates is configured.
func (s *SelectColor) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := s.AnyPlayer.Legal(state, proposer); err != nil {
		return err
	}

	if s.SelectedColor == nil {
		return errors.New("no color selected")
	}

	colorEnum := state.Manager().Chest().Enums().Enum("color")
	if colorEnum == nil {
		return errors.New("no 'color' enum found")
	}
	if s.SelectedColor.Enum() != colorEnum {
		return errors.New("selected color value is from a different enum")
	}
	if !colorEnum.Valid(s.SelectedColor.Value()) {
		return errors.New("selected color is not a valid value")
	}

	// Enforce uniqueness by default (unless WithAllowDuplicates)
	if !s.allowDuplicates() {
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
			if colorHolder, ok := p.(behaviors.HasPlayerColor); ok {
				if colorHolder.GetPlayerColor().Color.Value() == s.SelectedColor.Value() {
					return errors.New("another player already has that color")
				}
			}
		}
	}

	return nil
}

// allowDuplicates returns true if WithAllowDuplicates was configured.
func (s *SelectColor) allowDuplicates() bool {
	config := s.CustomConfiguration()
	val, ok := config[configPropAllowDuplicates]
	if !ok {
		return false
	}
	boolVal, ok := val.(bool)
	return ok && boolVal
}

// Apply sets the target player's color to the selected value.
func (s *SelectColor) Apply(state boardgame.State) error {
	target := s.TargetPlayerIndex
	player := state.PlayerStates()[target]

	colorHolder, ok := player.(behaviors.HasPlayerColor)
	if !ok {
		return errors.New("player state does not implement HasPlayerColor")
	}

	colorHolder.GetPlayerColor().Color.SetValue(s.SelectedColor.Value())
	return nil
}

// ValidConfiguration checks that the player state implements HasPlayerColor and
// that a "color" enum exists in the chest.
func (s *SelectColor) ValidConfiguration(exampleState boardgame.State) error {
	if err := s.AnyPlayer.ValidConfiguration(exampleState); err != nil {
		return err
	}

	playerState := exampleState.ImmutablePlayerStates()[0]
	if _, ok := playerState.(behaviors.HasPlayerColor); !ok {
		return errors.New("player state does not implement HasPlayerColor. Embed behaviors.PlayerColor in your player state")
	}

	if exampleState.Manager().Chest().Enums().Enum("color") == nil {
		return errors.New("no 'color' enum found in the chest. Define an enum named 'color'")
	}

	return nil
}

// FallbackName returns "Select Color"
func (s *SelectColor) FallbackName(m *boardgame.GameManager) string {
	return "Select Color"
}

// FallbackHelpText returns a description of the move.
func (s *SelectColor) FallbackHelpText() string {
	return "Choose your player color."
}
