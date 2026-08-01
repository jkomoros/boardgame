package memory

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/storage/memory"
)

// deadlockedDelegate is memory's real delegate with its activation move
// removed from its move list and nothing else changed. That single omission is
// the shape that actually deadlocked memory and pig: moves.SeatPlayer.Apply
// unconditionally calls SetPlayerInactive on whoever it seats, an inactive
// player fails PlayerIndex.Valid, and every moves.CurrentPlayer-gated move is
// therefore permanently illegal. Only an activation move clears the flag.
type deadlockedDelegate struct {
	gameDelegate
}

// ConfigureMoves drops the activation move by asking each move whether it IS
// one (interfaces.SeatedPlayerActivator), not by matching its display name.
// The display name is prose that the framework is free to reword -- and did,
// when "Activate Inactive Players" lost its stray plural -- at which point a
// name-matching filter would quietly stop removing anything and this test
// would pass for the opposite of its reason.
func (d *deadlockedDelegate) ConfigureMoves() []boardgame.MoveConfig {
	var result []boardgame.MoveConfig
	for _, config := range d.gameDelegate.ConfigureMoves() {
		if activator, ok := config.Constructor()().(interfaces.SeatedPlayerActivator); ok && activator.ActivatesSeatedPlayers() {
			continue
		}
		result = append(result, config)
	}
	return result
}

// TestSeatingWithoutActivationRefusesToBoot is the end-to-end proof that the
// Seat/InactivePlayer pairing is enforced by the framework rather than by
// prose plus a hand-copied per-game test.
//
// Before this check the deadlocked delegate below built a GameManager without
// complaint and shipped a game whose every turn-taking move was illegal from
// the moment auto-seating filled the seats. Ordinary tests miss it because
// manager.NewDefaultGame() leaves players unseated, so nothing is ever marked
// inactive in-process.
func TestSeatingWithoutActivationRefusesToBoot(t *testing.T) {

	// Sanity: memory's real delegate, which does configure the activation
	// move, still boots. Otherwise a check that rejects everything would
	// pass the negative case below for the wrong reason.
	if _, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager()); err != nil {
		t.Fatalf("memory's real delegate should boot, got: %v", err)
	}

	_, err := boardgame.NewGameManager(&deadlockedDelegate{}, memory.NewStorageManager())

	if err == nil {
		t.Fatal("a game that seats players (marking them inactive) and gates " +
			"moves on CurrentPlayer, but configures no move that " +
			"ActivatesSeatedPlayers, booted without complaint: every seated " +
			"player stays inactive forever and those moves are permanently " +
			"illegal")
	}

	if !strings.Contains(err.Error(), "ActivatesSeatedPlayers") || !strings.Contains(err.Error(), "moves.ActivateFilledSeat") {
		t.Errorf("the boot error should name the missing move so the author "+
			"knows what to add, got: %v", err)
	}
}
