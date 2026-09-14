package boardgame

import (
	"fmt"
	"time"
)

// UseManualClock fixes timer time for an isolated replay/scenario manager.
// Call it before creating/loading any games. Timer steps still run through
// ForceNextTimerWithError; this does not replace creator callbacks' wall clocks
// or move timestamps. It is not a live-game pause/resume API.
func (m *ManagerInternals) UseManualClock(at time.Time) error {
	m.manager.modifiableGamesLock.RLock()
	count := len(m.manager.modifiableGames)
	m.manager.modifiableGamesLock.RUnlock()
	if count != 0 || at.IsZero() {
		return fmt.Errorf("manual clock requires a nonzero time and a manager with no loaded games")
	}
	m.UseManualTimers()
	m.manager.timers.now = func() time.Time { return at }
	return nil
}
