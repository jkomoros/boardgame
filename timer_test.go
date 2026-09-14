package boardgame

import (
	"bytes"
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
	if err := restarted.Internals().RestoreTimers(); err != nil {
		t.Fatal(err)
	}
	if !restarted.timers.TimerActive(id) {
		t.Fatal("restart did not discover the durable timer before loading the game")
	}
	if err := restarted.Internals().RestoreTimers(); err == nil {
		t.Fatal("live restore unexpectedly replaced an already initialized scheduler")
	}
	if !restarted.timers.TimerActive(id) {
		t.Fatal("rejected live restore removed the existing durable timer")
	}
	if !restarted.timers.ForceNextTimer() {
		t.Fatal("restored timer was not scheduled")
	}
	reloaded := restarted.ModifiableGame(game.ID())
	if reloaded == nil {
		t.Fatal("timer did not load the cold game")
	}
	if reloaded.Version() != version+1 {
		t.Fatalf("timer completion advanced to version %d; want %d", reloaded.Version(), version+1)
	}
	reloadedState, _ := concreteStates(reloaded.CurrentState())
	restored := reloadedState.Timer.(*timer)
	if restored.ID != id || restored.Generation != generation || !restored.Deadline.Equal(deadline) {
		t.Fatalf("restored timer identity changed: %+v", restored)
	}
	if reloadedState.Timer.Active() {
		t.Fatal("successfully fired timer remained active")
	}
	if restarted.timers.ForceNextTimer() {
		t.Fatal("timer completion was scheduled more than once")
	}
}

func TestPoppedDurableTimerDoesNotRetryAfterGameFinishes(t *testing.T) {
	manager, game := durableTimerTestGame(t, newTestStorageManager())
	record := manager.timers.popNext(true)
	if record == nil || record.guard == nil {
		t.Fatal("fixture did not pop its durable timer")
	}

	move := game.MoveByName("Test").(*testMove)
	move.ScoreIncrement = 5
	if err := <-game.ProposeMove(move, game.CurrentState().CurrentPlayerIndex()); err != nil {
		t.Fatal(err)
	}
	if !game.Finished() {
		t.Fatal("fixture move did not finish the game")
	}
	gameState, _ := concreteStates(game.CurrentState())
	if !gameState.Timer.Active() {
		t.Fatal("fixture must retain the persisted active timer after game end")
	}

	if err := manager.timers.fire(record); err == nil {
		t.Fatal("popped timer unexpectedly fired after game end")
	}
	if manager.timers.ForceNextTimer() {
		t.Fatal("finished game's still-active timer was queued for retry")
	}
}

func TestDurableTimerFirstFailureSchedulesInitialRetry(t *testing.T) {
	manager, game := durableTimerTestGame(t, newTestStorageManager())
	defer manager.Internals().Close()
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	manager.timers.now = func() time.Time { return now }

	gameState, _ := concreteStates(game.CurrentState())
	persisted := gameState.Timer.(*timer)
	illegal := game.MoveByName("Make Illegal Phase")
	persisted.Move = StorageRecordForMove(illegal, persisted.Move.Phase, AdminPlayerIndex)
	if persisted.Move == nil {
		t.Fatal("could not serialize failing timer move")
	}

	var logs bytes.Buffer
	manager.Logger().SetOutput(&logs)
	record := manager.timers.popNext(true)
	if record == nil {
		t.Fatal("fixture did not pop its durable timer")
	}
	if err := manager.timers.fire(record); err == nil {
		t.Fatal("invalid timer completion unexpectedly committed")
	}
	manager.timers.mu.Lock()
	retried := manager.timers.recordsByID[record.id]
	manager.timers.mu.Unlock()
	if retried == nil || retried.fireTime.Sub(now) != timerRetryInitialDelay {
		t.Fatalf("first failed completion retry = %+v; want %v", retried, timerRetryInitialDelay)
	}
	if !strings.Contains(logs.String(), game.ID()) || !strings.Contains(logs.String(), record.id) {
		t.Fatalf("timer failure log did not identify game and timer: %s", logs.String())
	}
	if strings.Contains(logs.String(), illegal.Info().Name()) || strings.Contains(logs.String(), string(persisted.Move.Blob)) {
		t.Fatalf("timer failure log disclosed completion move input: %s", logs.String())
	}
}

func TestDurableTimerRetryBackoffAndGenerationReset(t *testing.T) {
	manager := newTimerManager(nil)
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }
	const id = "retry-timer"
	const gameID = "retry-game"

	recordForGeneration := func(generation uint64) *timerRecord {
		return &timerRecord{
			id: id, gameID: gameID, index: -1, fireTime: now,
			guard: &timerGuard{ID: id, Generation: generation}, now: manager.now,
		}
	}
	record := recordForGeneration(1)
	manager.replaceGameRecords(gameID, []*timerRecord{record})

	wants := []time.Duration{
		250 * time.Millisecond,
		500 * time.Millisecond,
		time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
		30 * time.Second,
		30 * time.Second,
	}
	for i, want := range wants {
		popped := manager.popNext(true)
		if popped == nil {
			t.Fatalf("retry %d had no queued timer", i+1)
		}
		delay, scheduled := manager.retry(popped)
		if !scheduled || delay != want {
			t.Fatalf("retry %d = %v, %t; want %v, true", i+1, delay, scheduled, want)
		}
		manager.mu.Lock()
		fireTime := manager.recordsByID[id].fireTime
		manager.mu.Unlock()
		if got := fireTime.Sub(now); got != want {
			t.Fatalf("retry %d scheduled after %v; want %v", i+1, got, want)
		}
	}

	// Reconciliation of the same durable generation keeps its accumulated
	// backoff instead of reverting an overdue deadline to an immediate retry.
	sameGeneration := recordForGeneration(1)
	manager.replaceGameRecords(gameID, []*timerRecord{sameGeneration})
	manager.mu.Lock()
	gotFireTime := manager.recordsByID[id].fireTime
	manager.mu.Unlock()
	if got := gotFireTime.Sub(now); got != timerRetryMaximumDelay {
		t.Fatalf("same-generation reconciliation reset backoff to %v", got)
	}

	// A replacement generation starts at the initial delay and prevents the
	// already-popped old generation from putting itself back on the heap.
	oldGeneration := manager.popNext(true)
	replacement := recordForGeneration(2)
	manager.replaceGameRecords(gameID, []*timerRecord{replacement})
	if delay, scheduled := manager.retry(oldGeneration); scheduled || delay != 0 {
		t.Fatalf("old generation was resurrected with retry %v", delay)
	}
	current := manager.popNext(true)
	if delay, scheduled := manager.retry(current); !scheduled || delay != timerRetryInitialDelay {
		t.Fatalf("replacement generation first retry = %v, %t; want %v, true", delay, scheduled, timerRetryInitialDelay)
	}
}

func TestCanceledPoppedTimerGenerationIsNotResurrected(t *testing.T) {
	manager := newTimerManager(nil)
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }
	record := &timerRecord{
		id: "canceled-timer", gameID: "canceled-game", index: -1, fireTime: now,
		guard: &timerGuard{ID: "canceled-timer", Generation: 3}, now: manager.now,
	}
	manager.replaceGameRecords(record.gameID, []*timerRecord{record})
	popped := manager.popNext(true)
	manager.replaceGameRecords(record.gameID, nil)
	if delay, scheduled := manager.retry(popped); scheduled || delay != 0 {
		t.Fatalf("canceled generation was resurrected with retry %v", delay)
	}
	if manager.popNext(true) != nil {
		t.Fatal("canceled generation remained queued")
	}
}

func TestDurableTimerSurvivesIdleEviction(t *testing.T) {
	manager, game := durableTimerTestGame(t, newTestStorageManager())
	version := game.Version()
	manager.freezeGame(game)
	if !game.Frozen() {
		t.Fatal("game was not frozen")
	}
	if fired, err := manager.Internals().ForceNextTimerWithError(); !fired || err != nil {
		t.Fatalf("cold timer firing = %t, %v", fired, err)
	}
	reloaded := manager.ModifiableGame(game.ID())
	if reloaded == nil || reloaded.Version() != version+1 {
		t.Fatalf("cold timer did not advance the stored game")
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
	if err := manager.timers.fire(&record); err == nil {
		t.Fatal("stale timer completion unexpectedly committed")
	}
	if game.Version() != version {
		t.Fatal("stale timer completion advanced the game")
	}
}

func TestTimerDeadlineStartsAtSaveBoundary(t *testing.T) {
	_, game := durableTimerTestGame(t, newTestStorageManager())
	candidate, err := game.CurrentState().(*state).copy(false)
	if err != nil {
		t.Fatal(err)
	}
	gameState, _ := concreteStates(candidate)
	completion := game.MoveByNameForState("Draw Card", candidate)
	gameState.Timer.Start(time.Minute, completion)
	timer := gameState.Timer.(*timer)
	if !timer.Deadline.IsZero() || timer.TimeLeft() != time.Minute {
		t.Fatalf("candidate timer resolved before save boundary: %+v", timer)
	}
	boundary := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	if err := finalizeTimerDeadlines(candidate, boundary); err != nil {
		t.Fatal(err)
	}
	if want := boundary.Add(time.Minute); !timer.Deadline.Equal(want) || timer.pendingDeadline {
		t.Fatalf("deadline = %v pending=%t; want %v finalized", timer.Deadline, timer.pendingDeadline, want)
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
