package boardgame

import (
	"container/heap"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/jkomoros/boardgame/enum"
)

// ImmutableTimer is a Timer that does not have any mutator methods. See Timer
// for more.
type ImmutableTimer interface {
	// Active reports whether the timer has started and has neither committed its
	// completion nor been canceled.
	Active() bool
	// TimeLeft returns the time until an active timer's persisted deadline. It
	// returns zero for inactive and overdue timers.
	TimeLeft() time.Duration
	id() string
	state() *state
	setState(*state)
}

// Timer is a state property representing a durable countdown. Timers must
// exist in a SubState and must always be non-nil, even when inactive.
//
// Start and Cancel only mutate the candidate State passed to Move.Apply. The
// scheduler is reconciled after that State commits, so a rejected move has no
// timer side effects. Active timers persist their deadline and completion move
// and resume after a game is reloaded.
type Timer interface {
	ImmutableTimer
	// Start begins a timer that proposes move as AdminPlayerIndex after duration.
	// The duration begins at the candidate state's save boundary, after Apply
	// and validation finish. It is generally called from Move.Apply.
	Start(time.Duration, Move)
	// Cancel cancels a started timer. It returns whether the timer was active.
	// It is generally called from Move.Apply.
	Cancel() bool
	importFrom(other ImmutableTimer) error
}

type timerStatus string

const (
	timerInactive timerStatus = "inactive"
	timerActive   timerStatus = "active"
	timerCanceled timerStatus = "canceled"
	timerFired    timerStatus = "fired"
)

type timer struct {
	// ID is an opaque public identifier used to join a Timer property to the
	// ActiveTimers payload. Lifecycle data is persisted separately in the state
	// storage record and is deliberately omitted from viewer JSON.
	ID         string
	Generation uint64
	Deadline   time.Time
	Status     timerStatus
	Move       *MoveStorageRecord
	statePtr   *state
	// pendingDuration exists only on a candidate state between Apply and the
	// atomic save boundary. It is never persisted or exposed to viewers.
	pendingDuration time.Duration
	pendingDeadline bool
}

// NewTimer returns a new blank timer, ready for use. StructInflater normally
// installs timers automatically for Timer properties.
func NewTimer() Timer { return &timer{Status: timerInactive} }

func (t *timer) importFrom(other ImmutableTimer) error {
	otherTimer, ok := other.(*timer)
	if !ok {
		return errors.New("timer had an unexpected implementation")
	}
	t.ID = otherTimer.ID
	t.Generation = otherTimer.Generation
	t.Deadline = otherTimer.Deadline
	t.Status = otherTimer.Status
	t.pendingDuration = otherTimer.pendingDuration
	t.pendingDeadline = otherTimer.pendingDeadline
	if otherTimer.Move == nil {
		t.Move = nil
	} else {
		moveCopy := *otherTimer.Move
		moveCopy.Blob = append([]byte(nil), otherTimer.Move.Blob...)
		t.Move = &moveCopy
	}
	return nil
}

func (t *timer) id() string        { return t.ID }
func (t *timer) state() *state     { return t.statePtr }
func (t *timer) setState(s *state) { t.statePtr = s }

// MarshalJSON intentionally preserves the historical public Timer shape. The
// durable deadline, generation, status, and move intent are state-storage
// metadata and never travel in a viewer state payload.
func (t *timer) MarshalJSON() ([]byte, error) {
	return DefaultMarshalJSON(map[string]interface{}{
		"ID":      t.ID,
		"IsTimer": true,
	})
}

func (t *timer) Active() bool { return t != nil && t.Status == timerActive }

func (t *timer) TimeLeft() time.Duration {
	if !t.Active() || t.statePtr == nil || t.statePtr.game == nil || t.statePtr.game.manager == nil {
		return 0
	}
	if t.pendingDeadline {
		if t.pendingDuration < 0 {
			return 0
		}
		return t.pendingDuration
	}
	remaining := t.Deadline.Sub(t.statePtr.game.manager.timers.now())
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (t *timer) Start(duration time.Duration, move Move) {
	if t.statePtr == nil || t.statePtr.game == nil {
		panic("cannot start a timer that is not attached to a game state")
	}
	if move == nil || move.Info() == nil {
		panic("cannot start a timer with a nil or unaffiliated move")
	}
	var phase enum.EnumKey
	if current := t.statePtr.game.manager.delegate.CurrentPhase(t.statePtr); current != nil {
		phase = current.Value()
	}
	intent := StorageRecordForMove(move, phase, AdminPlayerIndex)
	if intent == nil {
		panic("cannot start a timer with a move that cannot be serialized")
	}

	t.Generation++
	t.ID = randomString(timerIDLength, t.statePtr.Rand())
	t.Deadline = time.Time{}
	t.Status = timerActive
	t.Move = intent
	t.pendingDuration = duration
	t.pendingDeadline = true
}

func (t *timer) Cancel() bool {
	if !t.Active() {
		// Preserve the historical concrete behavior of clearing the public ID
		// even when cancellation reports false.
		t.ID = ""
		return false
	}
	t.ID = ""
	t.Deadline = time.Time{}
	t.Status = timerCanceled
	t.Move = nil
	t.pendingDuration = 0
	t.pendingDeadline = false
	return true
}

type persistedTimerRecord struct {
	Ref        StatePropertyRef
	ID         string
	Generation uint64
	Deadline   time.Time
	Status     timerStatus
	Move       *MoveStorageRecord `json:",omitempty"`
}

type timerGuard struct {
	Ref        StatePropertyRef
	ID         string
	Generation uint64
}

// TimerWakeup is the minimal non-secret record a storage backend returns to
// rebuild the process-local timer heap. Move intent remains only in the
// durable state and is inflated after the deadline when the game is checked
// out for modification.
type TimerWakeup struct {
	GameID     string
	Ref        StatePropertyRef
	ID         string
	Generation uint64
	Deadline   time.Time
}

// TimerWakeupStorage is an optional storage capability for discovering active
// durable timers without inflating every game into the manager's warm cache.
type TimerWakeupStorage interface {
	TimerWakeups(gameName string) ([]TimerWakeup, error)
}

// TimerWakeupStorageAvailability lets wrappers preserve optional support.
type TimerWakeupStorageAvailability interface {
	TimerWakeupStorageAvailable() bool
}

// SupportsTimerWakeupStorage reports whether storage can discover durable
// timer wakeups, including through a capability-preserving wrapper.
func SupportsTimerWakeupStorage(storage StorageManager) bool {
	if storage == nil {
		return false
	}
	if _, ok := storage.(TimerWakeupStorage); !ok {
		return false
	}
	if availability, ok := storage.(TimerWakeupStorageAvailability); ok {
		return availability.TimerWakeupStorageAvailable()
	}
	return true
}

// TimerWakeupsFromStateStorage extracts only active scheduler metadata from a
// durable state blob. Storage implementations can use it while scanning their
// game-head index; completion move intent is deliberately not returned.
func TimerWakeupsFromStateStorage(gameID string, record StateStorageRecord) ([]TimerWakeup, error) {
	var envelope struct {
		Timers []persistedTimerRecord
	}
	if err := json.Unmarshal(record, &envelope); err != nil {
		return nil, err
	}
	result := make([]TimerWakeup, 0, len(envelope.Timers))
	for _, timer := range envelope.Timers {
		if timer.Status != timerActive || timer.ID == "" || timer.Move == nil || timer.Deadline.IsZero() {
			continue
		}
		result = append(result, TimerWakeup{
			GameID: gameID, Ref: timer.Ref, ID: timer.ID,
			Generation: timer.Generation, Deadline: timer.Deadline,
		})
	}
	return result, nil
}

type timerLocation struct {
	ref   StatePropertyRef
	timer *timer
}

func timerLocations(state *state) ([]timerLocation, error) {
	if state == nil {
		return nil, errors.New("nil state")
	}
	var result []timerLocation
	visit := func(subState ConfigurableSubState) error {
		if subState == nil {
			return nil
		}
		reader := subState.Reader()
		for name, propType := range reader.Props() {
			if propType != TypeTimer {
				continue
			}
			immutable, err := reader.ImmutableTimerProp(name)
			if err != nil {
				return err
			}
			concrete, ok := immutable.(*timer)
			if !ok {
				return fmt.Errorf("timer property %s had an unexpected implementation", name)
			}
			ref := subState.StatePropertyRef()
			ref.PropName = name
			result = append(result, timerLocation{ref: ref, timer: concrete})
		}
		return nil
	}
	if err := visit(state.gameState); err != nil {
		return nil, err
	}
	for _, player := range state.playerStates {
		if err := visit(player); err != nil {
			return nil, err
		}
	}
	for _, values := range state.dynamicComponentValues {
		for _, value := range values {
			if err := visit(value); err != nil {
				return nil, err
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i].ref, result[j].ref
		if left.Group != right.Group {
			return left.Group < right.Group
		}
		if left.PlayerIndex != right.PlayerIndex {
			return left.PlayerIndex < right.PlayerIndex
		}
		if left.DeckName != right.DeckName {
			return left.DeckName < right.DeckName
		}
		if left.DeckIndex != right.DeckIndex {
			return left.DeckIndex < right.DeckIndex
		}
		return left.PropName < right.PropName
	})
	return result, nil
}

func timerAtRef(state *state, ref StatePropertyRef) (*timer, error) {
	locations, err := timerLocations(state)
	if err != nil {
		return nil, err
	}
	for _, location := range locations {
		if location.ref == ref {
			return location.timer, nil
		}
	}
	return nil, errors.New("timer property no longer exists")
}

func timerStorageRecords(state *state) []persistedTimerRecord {
	locations, err := timerLocations(state)
	if err != nil {
		return nil
	}
	result := make([]persistedTimerRecord, 0, len(locations))
	for _, location := range locations {
		timer := location.timer
		result = append(result, persistedTimerRecord{
			Ref: location.ref, ID: timer.ID, Generation: timer.Generation,
			Deadline: timer.Deadline, Status: timer.Status, Move: timer.Move,
		})
	}
	return result
}

func finalizeTimerDeadlines(state *state, now time.Time) error {
	locations, err := timerLocations(state)
	if err != nil {
		return err
	}
	for _, location := range locations {
		timer := location.timer
		if !timer.Active() || !timer.pendingDeadline {
			continue
		}
		timer.Deadline = now.Add(timer.pendingDuration)
		timer.pendingDuration = 0
		timer.pendingDeadline = false
	}
	return nil
}

func restoreTimerStorageRecords(state *state, records []persistedTimerRecord) error {
	for _, record := range records {
		timer, err := timerAtRef(state, record.Ref)
		if err != nil {
			return err
		}
		timer.ID = record.ID
		timer.Generation = record.Generation
		timer.Deadline = record.Deadline
		timer.Status = record.Status
		timer.Move = record.Move
	}
	return nil
}

func validateTimerGuard(state *state, guard *timerGuard) error {
	timer, err := timerAtRef(state, guard.Ref)
	if err != nil {
		return err
	}
	if !timer.Active() || timer.ID != guard.ID || timer.Generation != guard.Generation {
		return errors.New("timer completion was stale or already handled")
	}
	return nil
}

func consumeTimer(state *state, guard *timerGuard) error {
	if err := validateTimerGuard(state, guard); err != nil {
		return err
	}
	timer, _ := timerAtRef(state, guard.Ref)
	timer.Status = timerFired
	timer.Move = nil
	return nil
}

type timerRecord struct {
	id       string
	gameID   string
	index    int
	fireTime time.Time
	deadline time.Time
	// duration is non-zero only between the legacy PrepareTimer and StartTimer
	// calls retained for internal debugging compatibility.
	duration time.Duration
	game     *Game
	move     Move
	guard    *timerGuard
	now      func() time.Time
}

func (t *timerRecord) MarshalJSON() ([]byte, error) {
	return DefaultMarshalJSON(map[string]interface{}{
		"TimeLeft": t.TimeRemaining() / time.Millisecond,
	})
}

func (t *timerRecord) TimeRemaining() time.Duration {
	if t.duration > 0 {
		return t.duration
	}
	now := time.Now
	if t.now != nil {
		now = t.now
	}
	remaining := t.deadline.Sub(now())
	if remaining < 0 {
		return 0
	}
	return remaining
}

type timerQueue []*timerRecord

type timerManager struct {
	mu          sync.Mutex
	records     timerQueue
	recordsByID map[string]*timerRecord
	retryStates map[string]timerRetryState
	manager     *GameManager
	now         func() time.Time
}

type timerRetryState struct {
	gameID     string
	generation uint64
	failures   uint
	fireTime   time.Time
}

func newTimerManager(gameManager *GameManager) *timerManager {
	return &timerManager{
		records: make(timerQueue, 0), recordsByID: make(map[string]*timerRecord),
		retryStates: make(map[string]timerRetryState), manager: gameManager, now: time.Now,
	}
}

const (
	timerIDLength          = 16
	timerRetryInitialDelay = 250 * time.Millisecond
	timerRetryMaximumDelay = 30 * time.Second
)

func (t *timerManager) ActiveTimersForGame(gameID string) map[string]*timerRecord {
	t.mu.Lock()
	defer t.mu.Unlock()
	result := make(map[string]*timerRecord)
	for _, rec := range t.recordsByID {
		if rec.gameID == gameID && rec.duration == 0 {
			copy := *rec
			result[rec.id] = &copy
		}
	}
	return result
}

// PrepareTimer and StartTimer are retained for the manager's debugging API and
// compatibility with older internal callers. State Timer.Start uses durable
// state plus ReconcileState instead.
func (t *timerManager) PrepareTimer(duration time.Duration, state *state, move Move) string {
	id := randomString(timerIDLength, state.Rand())
	now := t.now()
	record := &timerRecord{
		id: id, gameID: state.Game().ID(), index: -1, duration: duration,
		fireTime: now.Add(100000 * time.Hour), deadline: now.Add(duration),
		game: state.Game(), move: move, now: t.now,
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.recordsByID[id] = record
	heap.Push(&t.records, record)
	return id
}

func (t *timerManager) StartTimer(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	record := t.recordsByID[id]
	if record == nil || record.duration == 0 {
		return
	}
	record.deadline = t.now().Add(record.duration)
	record.fireTime = record.deadline
	record.duration = 0
	heap.Fix(&t.records, record.index)
}

func (t *timerManager) TimerActive(id string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	record := t.recordsByID[id]
	return record != nil && record.duration == 0
}

func (t *timerManager) GetTimerRemaining(id string) time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	record := t.recordsByID[id]
	if record == nil {
		return 0
	}
	return record.TimeRemaining()
}

// ReconcileState makes the in-process heap an exact projection of a game's
// committed durable timers. It is safe to call repeatedly.
func (t *timerManager) ReconcileState(state *state) {
	locations, err := timerLocations(state)
	if err != nil {
		t.manager.Logger().Error("Could not inspect committed timers: ", err)
		return
	}
	var desired []*timerRecord
	if state.game.Finished() {
		t.replaceGameRecords(state.game.ID(), desired)
		return
	}
	for _, location := range locations {
		persisted := location.timer
		if !persisted.Active() || persisted.Move == nil || persisted.ID == "" {
			continue
		}
		guard := &timerGuard{Ref: location.ref, ID: persisted.ID, Generation: persisted.Generation}
		desired = append(desired, &timerRecord{
			id: persisted.ID, gameID: state.game.ID(), index: -1,
			fireTime: persisted.Deadline, deadline: persisted.Deadline,
			guard: guard, now: t.now,
		})
	}
	t.replaceGameRecords(state.game.ID(), desired)
}

func (t *timerManager) replaceGameRecords(gameID string, desired []*timerRecord) {
	t.mu.Lock()
	defer t.mu.Unlock()
	preserved := make(map[string]timerRetryState)
	for id, retry := range t.retryStates {
		if retry.gameID == gameID {
			preserved[id] = retry
		}
	}
	t.cancelTimersForGameLocked(gameID)
	for _, record := range desired {
		if record.guard != nil {
			retry := timerRetryState{gameID: gameID, generation: record.guard.Generation}
			if previous, ok := preserved[record.id]; ok && previous.generation == record.guard.Generation {
				retry = previous
				if retry.failures > 0 {
					record.fireTime = retry.fireTime
				}
			}
			t.retryStates[record.id] = retry
		}
		t.recordsByID[record.id] = record
		heap.Push(&t.records, record)
	}
}

func (t *timerManager) RestoreWakeups(wakeups []TimerWakeup) error {
	records := make([]*timerRecord, 0, len(wakeups))
	seenIDs := make(map[string]bool, len(wakeups))
	for _, wakeup := range wakeups {
		if wakeup.GameID == "" || wakeup.ID == "" || wakeup.Deadline.IsZero() {
			return errors.New("durable timer wakeup was incomplete")
		}
		if seenIDs[wakeup.ID] {
			return fmt.Errorf("durable timer wakeup ID %q was duplicated", wakeup.ID)
		}
		seenIDs[wakeup.ID] = true
		guard := &timerGuard{Ref: wakeup.Ref, ID: wakeup.ID, Generation: wakeup.Generation}
		records = append(records, &timerRecord{
			id: wakeup.ID, gameID: wakeup.GameID, index: -1,
			fireTime: wakeup.Deadline, deadline: wakeup.Deadline,
			guard: guard, now: t.now,
		})
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.recordsByID) != 0 || len(t.retryStates) != 0 {
		return errors.New("durable timers may only be restored into an empty startup scheduler")
	}
	for _, record := range records {
		t.retryStates[record.id] = timerRetryState{
			gameID: record.gameID, generation: record.guard.Generation,
		}
		t.recordsByID[record.id] = record
		heap.Push(&t.records, record)
	}
	return nil
}

func (t *timerManager) CancelTimersForGame(gameID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cancelTimersForGameLocked(gameID)
}

func (t *timerManager) cancelTimersForGameLocked(gameID string) {
	for id, rec := range t.recordsByID {
		if rec.gameID == gameID {
			heap.Remove(&t.records, rec.index)
			delete(t.recordsByID, id)
		}
	}
	for id, retry := range t.retryStates {
		if retry.gameID == gameID {
			delete(t.retryStates, id)
		}
	}
}

func (t *timerManager) CancelTimer(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	record := t.recordsByID[id]
	delete(t.retryStates, id)
	if record == nil {
		return
	}
	heap.Remove(&t.records, record.index)
	delete(t.recordsByID, id)
}

func (t *timerManager) ForceNextTimer() bool {
	fired, _ := t.ForceNextTimerWithError()
	return fired
}

func (t *timerManager) ForceNextTimerWithError() (bool, error) {
	record := t.popNext(true)
	if record == nil {
		return false, nil
	}
	return true, t.fire(record)
}

func (t *timerManager) Tick() {
	for {
		record := t.popNext(false)
		if record == nil {
			return
		}
		_ = t.fire(record)
	}
}

func (t *timerManager) fire(record *timerRecord) error {
	game, move, retryable, err := t.resolveRecord(record)
	if err != nil {
		if retryable {
			if delay, scheduled := t.retry(record); scheduled {
				t.logTimerFailure(record, "Timer completion could not be prepared; durable intent will retry", delay)
			}
		}
		return err
	}
	err = <-game.ProposeMove(move, AdminPlayerIndex)
	if err == nil {
		return nil
	}
	if game.Finished() {
		t.logTimerFailure(record, "Timer completion move was rejected after the game finished", 0)
		return err
	}
	// A storage or transient move failure leaves the durable timer active. Retry
	// later, but do not resurrect a stale generation after cancel/restart.
	if record.guard != nil {
		state, ok := game.CurrentState().(*state)
		if !ok || validateTimerGuard(state, record.guard) != nil {
			return err
		}
		if delay, scheduled := t.retry(record); scheduled {
			t.logTimerFailure(record, "Timer completion move was rejected; durable intent will retry", delay)
		}
	}
	return err
}

func timerRetryDelay(failures uint) time.Duration {
	delay := timerRetryInitialDelay
	for i := uint(0); i < failures && delay < timerRetryMaximumDelay; i++ {
		delay *= 2
		if delay > timerRetryMaximumDelay {
			delay = timerRetryMaximumDelay
		}
	}
	return delay
}

// retry reschedules only an identity which remains known to the scheduler.
// Reconciliation removes that identity when the durable timer is canceled or
// replaces it when its generation changes, preventing a concurrent failed
// completion from resurrecting stale work.
func (t *timerManager) retry(record *timerRecord) (time.Duration, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if record == nil || record.guard == nil {
		return 0, false
	}
	retry, ok := t.retryStates[record.id]
	if !ok || retry.gameID != record.gameID || retry.generation != record.guard.Generation {
		return 0, false
	}
	existing := t.recordsByID[record.id]
	if existing != nil && (existing.guard == nil || existing.guard.Generation != record.guard.Generation) {
		return 0, false
	}
	delay := timerRetryDelay(retry.failures)
	if delay < timerRetryMaximumDelay {
		retry.failures++
	}
	retry.fireTime = t.now().Add(delay)
	t.retryStates[record.id] = retry

	if existing != nil {
		existing.fireTime = retry.fireTime
		heap.Fix(&t.records, existing.index)
		return delay, true
	}
	record.fireTime = retry.fireTime
	t.recordsByID[record.id] = record
	heap.Push(&t.records, record)
	return delay, true
}

func (t *timerManager) logTimerFailure(record *timerRecord, message string, retryDelay time.Duration) {
	fields := map[string]interface{}{
		"game_id":  record.gameID,
		"timer_id": record.id,
	}
	if record.guard != nil {
		fields["timer_generation"] = record.guard.Generation
	}
	if retryDelay > 0 {
		fields["retry_delay"] = retryDelay
	}
	t.manager.Logger().WithFields(fields).Info(message)
}

func (t *timerManager) resolveRecord(record *timerRecord) (*Game, Move, bool, error) {
	if record.guard == nil {
		if record.game == nil || record.move == nil {
			return nil, nil, false, errors.New("timer scheduler record was incomplete")
		}
		return record.game, record.move, false, nil
	}
	game := t.manager.ModifiableGame(record.gameID)
	if game == nil {
		return nil, nil, true, errors.New("timer game could not be loaded")
	}
	if game.Finished() {
		return nil, nil, false, errors.New("timer game was already finished")
	}
	// Loading a cold game reconciles its state and may recreate this record.
	// Remove that duplicate before proposing the one already popped.
	t.removeMatching(record.id, record.guard.Generation)
	state, ok := game.CurrentState().(*state)
	if !ok || state == nil {
		return nil, nil, true, errors.New("timer game had no mutable current state")
	}
	if err := validateTimerGuard(state, record.guard); err != nil {
		return nil, nil, false, err
	}
	persisted, err := timerAtRef(state, record.guard.Ref)
	if err != nil || persisted.Move == nil {
		return nil, nil, false, errors.New("timer completion move was unavailable")
	}
	move, err := persisted.Move.inflateForState(game, state)
	if err != nil {
		return nil, nil, true, err
	}
	move.Info().timerGuard = record.guard
	return game, move, false, nil
}

func (t *timerManager) removeMatching(id string, generation uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	record := t.recordsByID[id]
	if record == nil || record.guard == nil || record.guard.Generation != generation {
		return
	}
	heap.Remove(&t.records, record.index)
	delete(t.recordsByID, id)
}

func (t *timerManager) nextTimerFired() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.records) > 0 && !t.records[0].fireTime.After(t.now())
}

func (t *timerManager) popNext(force bool) *timerRecord {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.records) == 0 || (!force && t.records[0].fireTime.After(t.now())) {
		return nil
	}
	record := heap.Pop(&t.records).(*timerRecord)
	delete(t.recordsByID, record.id)
	return record
}

func (t timerQueue) Len() int           { return len(t) }
func (t timerQueue) Less(i, j int) bool { return t[i].fireTime.Before(t[j].fireTime) }
func (t timerQueue) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
	t[i].index = i
	t[j].index = j
}
func (t *timerQueue) Push(x interface{}) {
	item := x.(*timerRecord)
	item.index = len(*t)
	*t = append(*t, item)
}
func (t *timerQueue) Pop() interface{} {
	old := *t
	item := old[len(old)-1]
	item.index = -1
	*t = old[:len(old)-1]
	return item
}
