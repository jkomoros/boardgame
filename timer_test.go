package boardgame

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/workfit/tester/assert"
)

func TestTimerManager(t *testing.T) {

	game := testDefaultGame(t, true)

	currentVersion := game.Version()

	move := game.MoveByName("Draw Card")

	assert.For(t).ThatActual(move).IsNotNil()

	assert.For(t).ThatActual(move.(*testMoveDrawCard).TargetPlayerIndex).Equals(PlayerIndex(0))

	timer := newTimerManager(game.manager)

	assert.For(t).ThatActual(timer.nextTimerFired()).Equals(false)

	registeredDuration := time.Duration(50) * time.Millisecond

	id := timer.PrepareTimer(registeredDuration, game.CurrentState().(*state), move)

	assert.For(t).ThatActual(id).Equals("D732D7BBF5331D57")

	remaining := timer.GetTimerRemaining(id)

	assert.For(t).ThatActual(remaining).Equals(registeredDuration)

	assert.For(t).ThatActual(timer.records[0].fireTime.Sub(time.Now()) > time.Hour)

	timer.StartTimer(id)

	assert.For(t).ThatActual(registeredDuration-remaining < time.Millisecond*10).IsTrue()

	assert.For(t).ThatActual(timer.nextTimerFired()).Equals(false)

	timer.Tick()

	//Ticking before any time has really passed shouldn't trigger the next timer.
	assert.For(t).ThatActual(timer.nextTimerFired()).Equals(false)

	<-time.After(60 * time.Millisecond)

	assert.For(t).ThatActual(timer.nextTimerFired()).IsTrue()

	timer.Tick()

	//Make sure the game was actually made after the time had elapsed.
	assert.For(t).ThatActual(game.Version()).Equals(currentVersion + 1)

	//Now that the tick has happened, there should be not more records.
	assert.For(t).ThatActual(len(timer.records)).Equals(0)
	assert.For(t).ThatActual(len(timer.recordsByID)).Equals(0)

	assert.For(t).ThatActual(timer.GetTimerRemaining(id)).Equals(time.Duration(0))

}

func TestTimerManagerMultiple(t *testing.T) {

	game := testDefaultGame(t, false)

	move := game.MoveByName("Draw Card")

	assert.For(t).ThatActual(move).IsNotNil()

	currentVersion := game.Version()

	timer := newTimerManager(game.manager)

	state := game.CurrentState().(*state)

	firstID := timer.PrepareTimer(time.Duration(50)*time.Millisecond, state, move)
	timer.StartTimer(firstID)
	secondID := timer.PrepareTimer(time.Duration(10)*time.Millisecond, state, move)
	timer.StartTimer(secondID)
	thirdID := timer.PrepareTimer(time.Duration(70)*time.Millisecond, state, move)
	timer.StartTimer(thirdID)

	//Make sure that even though the second timer was added second, it is first.
	assert.For(t).ThatActual(timer.records[0].id).Equals(secondID)
	assert.For(t).ThatActual(timer.records[1].id).Equals(firstID)
	assert.For(t).ThatActual(timer.records[2].id).Equals(thirdID)

	timer.CancelTimer(secondID)

	assert.For(t).ThatActual(len(timer.records)).Equals(2)

	assert.For(t).ThatActual(timer.records[0].id).Equals(firstID)
	assert.For(t).ThatActual(timer.records[1].id).Equals(thirdID)

	assert.For(t).ThatActual(timer.nextTimerFired()).IsFalse()

	<-time.After(70 * time.Millisecond)

	timer.Tick()

	assert.For(t).ThatActual(len(timer.records)).Equals(0)

	assert.For(t).ThatActual(game.Version()).Equals(currentVersion + 2)

}

func TestCancelTimersForGame(t *testing.T) {
	game := testDefaultGame(t, false)

	move := game.MoveByName("Draw Card")
	assert.For(t).ThatActual(move).IsNotNil()

	timer := newTimerManager(game.manager)

	state := game.CurrentState().(*state)

	firstID := timer.PrepareTimer(time.Duration(50)*time.Millisecond, state, move)
	timer.StartTimer(firstID)
	secondID := timer.PrepareTimer(time.Duration(100)*time.Millisecond, state, move)
	timer.StartTimer(secondID)

	assert.For(t).ThatActual(len(timer.records)).Equals(2)
	assert.For(t).ThatActual(len(timer.recordsByID)).Equals(2)

	// Cancel all timers for this game
	timer.CancelTimersForGame(game.ID())

	assert.For(t).ThatActual(len(timer.records)).Equals(0)
	assert.For(t).ThatActual(len(timer.recordsByID)).Equals(0)
}

func TestTimerProp(t *testing.T) {
	game := testDefaultGame(t, false)
	game.Manager().Internals().UseManualTimers()

	move := game.MoveByName("Draw Card")

	assert.For(t).ThatActual(move).IsNotNil()

	currentVersion := game.Version()

	gameState, _ := concreteStates(game.CurrentState())

	gameState.Timer.Start(time.Millisecond*5, move)

	assert.For(t).ThatActual(gameState.Timer.Active()).IsTrue()

	assert.For(t).ThatActual(gameState.Timer.TimeLeft() <= time.Millisecond*5).IsTrue()

	//Trigger the timers to actually be added
	game.CurrentState().(*state).committed()

	assert.For(t).ThatActual(gameState.Timer.Active()).IsTrue()

	assert.For(t).ThatActual(gameState.Timer.TimeLeft() < time.Millisecond*50).IsTrue()

	if fired, err := game.Manager().Internals().ForceNextTimerWithError(); !fired || err != nil {
		t.Fatalf("forcing timer = %t, %v", fired, err)
	}

	assert.For(t).ThatActual(game.Version()).Equals(currentVersion + 1)

	gameState, _ = concreteStates(game.CurrentState())

	assert.For(t).ThatActual(gameState.Timer.Active()).IsFalse()

	assert.For(t).ThatActual(gameState.Timer.TimeLeft()).Equals(time.Duration(0))

	gameState.Timer.Cancel()

	assert.For(t).ThatActual(gameState.Timer.id()).Equals("")

}

type testMoveStartDurableTimer struct{ baseMove }

var testMoveStartDurableTimerConfig = NewMoveConfig(
	"Start Durable Timer",
	func() Move { return new(testMoveStartDurableTimer) },
	nil,
)

func (t *testMoveStartDurableTimer) Reader() PropertyReader {
	return getDefaultReader(t)
}
func (t *testMoveStartDurableTimer) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}
func (t *testMoveStartDurableTimer) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(t)
}
func (t *testMoveStartDurableTimer) Legal(ImmutableState, PlayerIndex) error { return nil }
func (t *testMoveStartDurableTimer) Apply(state State) error {
	gameState, _ := concreteStates(state)
	completion := state.Game().MoveByNameForState("Draw Card", state)
	gameState.Timer.Start(time.Hour, completion)
	return nil
}

type testMoveCancelDurableTimer struct {
	baseMove
	Fail bool
}

var testMoveCancelDurableTimerConfig = NewMoveConfig(
	"Cancel Durable Timer",
	func() Move { return new(testMoveCancelDurableTimer) },
	nil,
)

func (t *testMoveCancelDurableTimer) Reader() PropertyReader {
	return getDefaultReader(t)
}
func (t *testMoveCancelDurableTimer) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(t)
}
func (t *testMoveCancelDurableTimer) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(t)
}
func (t *testMoveCancelDurableTimer) Legal(ImmutableState, PlayerIndex) error { return nil }
func (t *testMoveCancelDurableTimer) Apply(state State) error {
	gameState, _ := concreteStates(state)
	gameState.Timer.Cancel()
	if t.Fail {
		return errors.New("deliberate cancellation rejection")
	}
	return nil
}

func durableTimerTestDelegate() *testGameDelegate {
	delegate := defaultTestGameDelegate(0)
	baseInstaller := delegate.moveInstaller
	delegate.moveInstaller = func(manager *GameManager) []MoveConfig {
		result := baseInstaller(manager)
		return append(result, testMoveStartDurableTimerConfig, testMoveCancelDurableTimerConfig)
	}
	return delegate
}

func durableTimerTestGame(t *testing.T, storage StorageManager) (*GameManager, *Game) {
	t.Helper()
	manager, err := NewGameManager(durableTimerTestDelegate(), storage)
	if err != nil {
		t.Fatal(err)
	}
	manager.Internals().UseManualTimers()
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	if err := <-game.ProposeMove(game.MoveByName("Start Durable Timer"), AdminPlayerIndex); err != nil {
		t.Fatal(err)
	}
	return manager, game
}

func TestRejectedCancellationHasNoLiveEffect(t *testing.T) {
	manager, game := durableTimerTestGame(t, newTestStorageManager())
	before := game.Version()
	gameState, _ := concreteStates(game.CurrentState())
	id := gameState.Timer.id()

	cancel := game.MoveByName("Cancel Durable Timer").(*testMoveCancelDurableTimer)
	cancel.Fail = true
	if err := <-game.ProposeMove(cancel, AdminPlayerIndex); err == nil {
		t.Fatal("canceling move unexpectedly committed")
	}
	gameState, _ = concreteStates(game.CurrentState())
	if !gameState.Timer.Active() || gameState.Timer.id() != id {
		t.Fatal("rejected cancellation changed the durable timer")
	}
	if !manager.timers.TimerActive(id) {
		t.Fatal("rejected cancellation removed the live scheduler record")
	}
	if game.Version() != before {
		t.Fatalf("rejected cancellation advanced version to %d; want %d", game.Version(), before)
	}
}

func TestDurableTimerRestoresAndFiresOnce(t *testing.T) {
	storage := newTestStorageManager()
	manager, game := durableTimerTestGame(t, storage)
	gameState, _ := concreteStates(game.CurrentState())
	persisted := gameState.Timer.(*timer)
	id, generation, deadline := persisted.ID, persisted.Generation, persisted.Deadline
	version := game.Version()

	manager.freezeGame(game)
	restarted, err := NewGameManager(durableTimerTestDelegate(), storage)
	if err != nil {
		t.Fatal(err)
	}
	restarted.Internals().UseManualTimers()
	reloaded := restarted.ModifiableGame(game.ID())
	if reloaded == nil {
		t.Fatal("reloading game failed")
	}
	reloadedState, _ := concreteStates(reloaded.CurrentState())
	restored := reloadedState.Timer.(*timer)
	if restored.ID != id || restored.Generation != generation || !restored.Deadline.Equal(deadline) ||
		restored.Move == nil || restored.Move.Name != "Draw Card" {
		t.Fatalf("restored timer = %+v; want id=%s generation=%d deadline=%v Draw Card", restored, id, generation, deadline)
	}
	if !restarted.timers.ForceNextTimer() {
		t.Fatal("restored timer was not scheduled")
	}
	if reloaded.Version() != version+1 {
		t.Fatalf("timer completion advanced to version %d; want %d", reloaded.Version(), version+1)
	}
	reloadedState, _ = concreteStates(reloaded.CurrentState())
	if reloadedState.Timer.Active() {
		t.Fatal("successfully fired timer remained active")
	}
	if restarted.timers.ForceNextTimer() {
		t.Fatal("timer completion was scheduled more than once")
	}
}

func TestStaleTimerCompletionRejectedAfterCancel(t *testing.T) {
	manager, game := durableTimerTestGame(t, newTestStorageManager())
	manager.timers.mu.Lock()
	record := *manager.timers.records[0]
	manager.timers.mu.Unlock()

	cancel := game.MoveByName("Cancel Durable Timer").(*testMoveCancelDurableTimer)
	if err := <-game.ProposeMove(cancel, AdminPlayerIndex); err != nil {
		t.Fatal(err)
	}
	version := game.Version()
	if err := <-game.ProposeMove(record.move, AdminPlayerIndex); err == nil {
		t.Fatal("stale timer completion unexpectedly committed")
	}
	if game.Version() != version {
		t.Fatal("stale timer completion advanced the game")
	}
}

func TestTimerStorageMetadataIsNotViewerJSON(t *testing.T) {
	_, game := durableTimerTestGame(t, newTestStorageManager())
	stored := string(game.CurrentState().StorageRecord())
	if !strings.Contains(stored, `"Timers"`) || !strings.Contains(stored, `"Deadline"`) ||
		!strings.Contains(stored, `"Generation"`) || !strings.Contains(stored, `"Draw Card"`) {
		t.Fatalf("timer lifecycle metadata missing from storage: %s", stored)
	}
	viewer, err := json.Marshal(game.CurrentState())
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{`"Timers"`, `"Deadline"`, `"Generation"`, `"Draw Card"`} {
		if strings.Contains(string(viewer), secret) {
			t.Fatalf("viewer JSON disclosed timer metadata %s: %s", secret, viewer)
		}
	}
}

func TestManualTimerLifecycle(t *testing.T) {
	game := testDefaultGame(t, false)
	manager := game.Manager()
	manager.Internals().UseManualTimers()
	select {
	case <-manager.timerTickerDone:
	default:
		t.Fatal("UseManualTimers returned before the ticker stopped")
	}

	manager.Internals().Close()
	manager.Internals().Close()
	if !game.Frozen() {
		t.Fatal("Close did not freeze a resident modifiable game")
	}
}

func TestForceNextTimerWithError(t *testing.T) {
	game := testDefaultGame(t, false)
	manager := game.Manager()
	manager.Internals().UseManualTimers()
	state := game.CurrentState().(*state)
	move := game.MoveByName("Make Illegal Phase")
	id := manager.timers.PrepareTimer(time.Hour, state, move)
	manager.timers.StartTimer(id)

	fired, err := manager.Internals().ForceNextTimerWithError()
	if !fired || err == nil {
		t.Fatalf("ForceNextTimerWithError = %t, %v; want attempted failure", fired, err)
	}
	fired, err = manager.Internals().ForceNextTimerWithError()
	if fired || err != nil {
		t.Fatalf("second ForceNextTimerWithError = %t, %v; want empty queue", fired, err)
	}
	manager.Internals().Close()
}
