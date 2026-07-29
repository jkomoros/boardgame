package memory

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
)

// deadlockedDelegate is memory's real delegate with moves.ActivateInactivePlayer
// removed from its move list and nothing else changed. That single omission is
// the shape that actually deadlocked memory and pig: moves.SeatPlayer.Apply
// unconditionally calls SetPlayerInactive on whoever it seats, an inactive
// player fails PlayerIndex.Valid, and every moves.CurrentPlayer-gated move is
// therefore permanently illegal. Only ActivateInactivePlayer clears the flag.
type deadlockedDelegate struct {
	gameDelegate
}

func (d *deadlockedDelegate) ConfigureMoves() []boardgame.MoveConfig {
	var result []boardgame.MoveConfig
	for _, config := range d.gameDelegate.ConfigureMoves() {
		if strings.Contains(config.Name(), "Activate Inactive") {
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
			"moves on CurrentPlayer, but configures no ActivateInactivePlayer, " +
			"booted without complaint: every seated player stays inactive " +
			"forever and those moves are permanently illegal")
	}

	if !strings.Contains(err.Error(), "ActivateInactivePlayer") {
		t.Errorf("the boot error should name the missing move so the author "+
			"knows what to add, got: %v", err)
	}
}
