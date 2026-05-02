package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
)

// GatheringMoves auto-detects which selection behaviors are present on the
// player state and returns corresponding move configs. It inspects the example
// player state for [behaviors.HasPlayerTeam], [behaviors.HasPlayerRole], and
// [behaviors.HasPlayerColor], returning configs for [SelectTeam], [SelectRole],
// and [SelectColor] respectively.
//
// Usage:
//
//	moves.AddForPhase(phaseGathering, moves.GatheringMoves(auto)...)
//
// Returns nil (empty slice) if no selection behaviors are detected. AddForPhase
// with an empty slice is a no-op, so calling GatheringMoves is always safe.
//
// For more control over individual move configuration (e.g., WithUnique on
// SelectRole), configure moves directly instead of using this helper.
// selectionIsUnique is the shared logic for checking whether uniqueness should
// be enforced on a selection move. It checks both WithUnique (opt-in) and
// WithAllowDuplicates (opt-out), with defaultUnique determining behavior when
// neither is set. This allows both options to work on all selection moves.
// checkRequireAdmin verifies that the proposer is the game administrator, if
// the move was configured with WithRequireAdmin. Returns nil if the check
// passes or is not configured. Used by CloseAllSeats and selection moves.
func checkRequireAdmin(config boardgame.PropertyCollection, state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	val, ok := config[configPropRequireAdmin]
	if !ok {
		return nil
	}
	requireAdmin, ok := val.(bool)
	if !ok || !requireAdmin {
		return nil
	}

	// AdminPlayerIndex always passes (engine-initiated actions)
	if proposer == boardgame.AdminPlayerIndex {
		return nil
	}

	if proposer < 0 {
		return errors.New("only the game administrator can make this move")
	}

	player := state.ImmutablePlayerStates()[proposer]
	// If the playerState doesn't have GameAdministrator, skip the check
	// (backward compatible with games that don't use admin)
	if _, ok := player.(behaviors.HasGameAdministrator); !ok {
		return nil
	}
	if !behaviors.PlayerIsAdmin(player) {
		return errors.New("only the game administrator can make this move")
	}
	return nil
}

func selectionIsUnique(config boardgame.PropertyCollection, defaultUnique bool) bool {
	// Explicit WithUnique overrides everything
	if val, ok := config[configPropUnique]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	// Explicit WithAllowDuplicates overrides default
	if val, ok := config[configPropAllowDuplicates]; ok {
		if boolVal, ok := val.(bool); ok {
			return !boolVal
		}
	}
	return defaultUnique
}

func GatheringMoves(auto *AutoConfigurer) []boardgame.MoveConfig {
	exampleState := auto.delegate.Manager().ExampleState()
	playerState := exampleState.ImmutablePlayerStates()[0]

	var result []boardgame.MoveConfig

	if _, ok := playerState.(behaviors.HasPlayerTeam); ok {
		result = append(result, auto.MustConfig(new(SelectTeam)))
	}
	if _, ok := playerState.(behaviors.HasPlayerRole); ok {
		result = append(result, auto.MustConfig(new(SelectRole)))
	}
	if _, ok := playerState.(behaviors.HasPlayerColor); ok {
		result = append(result, auto.MustConfig(new(SelectColor)))
	}

	return result
}
