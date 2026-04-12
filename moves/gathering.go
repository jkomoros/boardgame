package moves

import (
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
