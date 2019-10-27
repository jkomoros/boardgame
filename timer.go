package boardgame

import (
	"container/heap"
	"time"
)

//ImmutableTimer is a Timer that does not have any mutator methods. See Timer
//for more.
type ImmutableTimer interface {
	//Active returns true if the given timer has been Start()'d and has not
	//yet fired or been canceled.
	Active() bool
	//TimerLeft reutrns the amount of time until the timer fires, if it is
	//active.
	TimeLeft() time.Duration
	id() string
	state() *state
	setState(*state)
}

//Timer is a type of property that can be used in states that represents a
//countdown. Timers must exist in a given SubState in your state, and must
//always be non-nil, even if they aren't actively being used.
type Timer interface {
	ImmutableTimer
	//Start begins a timer that will automatically call game.ProposeMove(Move,
	//AdminPlayerIndex) after the given duration has elapsed. Generally called
	//from within a move.Apply() method.
	Start(time.Duration, Move)
	//Cancel cancels a previously Start()'d timer, so that it will no longer
	//fire. If the timer was not active, it's a no-op. The return value is
	//whether the timer was active before being canceled. Generally called
	//during a move.Apply() method.
	Cancel() bool
	importFrom(other ImmutableTimer) error
}

type timer struct {
	//Id will be an opaque identifier that is used to keep track of the
	//corresponding underlying Timer object in the game engine. It is not
	//meaningful to inspect yourself and should not be modified.
	Id       string
	statePtr *state
}

//NewTimer returns a new blank timer, ready for use. Typically this would be
//used inside of GameDelegate.GameStateConstructor and friends. In practice
//however this is not necessary because the auto-crated StructInflaters for
//your structs will install a non-nil Timer even if not struct tags are
//provided, because no configuration is necessary. See StructInflater for
//more.
func NewTimer() Timer {
	return &timer{}
}

func (t *timer) importFrom(other ImmutableTimer) error {
	t.Id = other.id()
	return nil
}

func (t *timer) id() string {
	return t.Id
}

func (t *timer) state() *state {
	return t.statePtr
}

func (t *timer) setState(state *state) {
	t.statePtr = state
}

func (t *timer) MarshalJSON() ([]byte, error) {
	obj := map[string]interface{}{
		"Id": t.Id,
		//TODO: this is a hack so client can find it easily
		"IsTimer": true,
	}

	return DefaultMarshalJSON(obj)
}

//Active returns true if the timer is active and counting down.
func (t *timer) Active() bool {
	return t.statePtr.game.manager.timers.TimerActive(t.Id)
}

//TimeLeft returns the number of nanoseconds left until this timer fires.
func (t *timer) TimeLeft() time.Duration {
	return t.statePtr.game.manager.timers.GetTimerRemaining(t.Id)
}

//Start starts the timer. After duration has passed, the Move will be proposed
//via proposeMove. If the timer is already active, it will be canceled before
//the new timer is configured.
func (t *timer) Start(duration time.Duration, move Move) {

	if t.Active() {
		t.Cancel()
	}

	game := t.statePtr.game
	manager := game.manager

	t.Id = manager.timers.PrepareTimer(duration, t.statePtr, move)

	t.statePtr.timersToStart = append(t.statePtr.timersToStart, t.Id)
}

//Cancel cancels an active timer. If the timer is not active, it has no
//effect. Returns true if the timer was active and canceled, false if the
//timer was not active.
func (t *timer) Cancel() bool {

	wasActive := t.Active()

	manager := t.statePtr.game.manager

	manager.timers.CancelTimer(t.Id)

	//Technically there's a case where Start() was called, but the state was
	//never fully committed. However, StartTimer() on a canceled timer is a
	//no-op so it's fine.

	t.Id = ""

	return wasActive
}

type timerRecord struct {
	id     string
	gameId string
	index  int
	//When the timer should fire. Set after the timer is fully Started().
	//Before that it is impossibly far in the future.
	fireTime time.Time
	//The duration we were configured with. Only set when the timer is
	//Prepared() but not yet Started().
	duration time.Duration
	game     *Game
	move     Move
}

func (t *timerRecord) MarshalJSON() ([]byte, error) {
	//TODO: return other information, evertyhing in TimerRecord struct in issue 698.
	obj := map[string]interface{}{
		//TimeLeft is only ever for the client (it's not read back in when
		//deserialized), so put it in the more traditional milliseconds units,
		//not nanoseconds.
		"TimeLeft": t.TimeRemaining() / time.Millisecond,
	}

	return DefaultMarshalJSON(obj)
}

func (t *timerRecord) TimeRemaining() time.Duration {

	//Before a timer is Started(), just say its duration as the time
	//remaining.
	if t.duration > 0 {
		return t.duration
	}

	duration := t.fireTime.Sub(time.Now())

	if duration < 0 {
		duration = 0
	}

	return duration
}

type timerQueue []*timerRecord

type timerManager struct {
	records     timerQueue
	recordsById map[string]*timerRecord
	//TODO: recordsByGameId for efficiency so we don't have to search
	manager *GameManager
}

func newTimerManager(gameManager *GameManager) *timerManager {
	return &timerManager{
		//the default id in TimerProps is 0, so we should start beyond that.
		records:     make(timerQueue, 0),
		recordsById: make(map[string]*timerRecord),
		manager:     gameManager,
	}
}

const timerIDLength = 16

func (t *timerManager) ActiveTimersForGame(gameID string) map[string]*timerRecord {
	result := make(map[string]*timerRecord)
	for _, rec := range t.recordsById {
		if rec.gameId == gameID {
			result[rec.id] = rec
		}
	}
	return result
}

//PrepareTimer creates a timer entry and gets it ready and an Id allocated.
//However, the timer doesn't actually start counting down until
//manager.StartTimer(id) is called.
func (t *timerManager) PrepareTimer(duration time.Duration, state *state, move Move) string {

	record := &timerRecord{
		id:       randomString(timerIDLength, state.Rand()),
		gameId:   state.Game().Id(),
		index:    -1,
		duration: duration,
		//fireTime will be set when StartTimer is called. For now, set it to
		//something impossibly far in the future.
		fireTime: time.Now().Add(time.Hour * 100000),
		game:     state.Game(),
		move:     move,
	}

	t.recordsById[record.id] = record

	heap.Push(&t.records, record)

	return record.id
}

//StartTimer actually triggers a timer that was previously PrepareTimer'd to
//start counting down.
func (t *timerManager) StartTimer(id string) {

	if t.TimerActive(id) {
		return
	}

	record := t.recordsById[id]

	if record == nil {
		return
	}

	record.fireTime = time.Now().Add(record.duration)
	record.duration = 0

	heap.Fix(&t.records, record.index)
}

//TimerActive returns if the timer is active and counting down.
func (t *timerManager) TimerActive(id string) bool {
	record := t.recordsById[id]

	if record == nil {
		return false
	}

	if record.duration > 0 {
		return false
	}

	return true
}

func (t *timerManager) GetTimerRemaining(id string) time.Duration {
	record := t.recordsById[id]

	if record == nil {
		return 0
	}

	return record.TimeRemaining()
}

func (t *timerManager) CancelTimer(id string) {
	record := t.recordsById[id]

	if record == nil {
		return
	}

	heap.Remove(&t.records, record.index)

	record.index = -1

	delete(t.recordsById, record.id)

}

//ForceNextTimer is designed to force fire the next timer no matter when it's
//_supposed_ to fire. Will return true if a timer was fired. Primarily exists
//for debug purposes.
func (t *timerManager) ForceNextTimer() bool {
	record := t.popNext(true)
	if record == nil {
		return false
	}

	<-record.game.ProposeMove(record.move, AdminPlayerIndex)

	return true
}

//Should be called regularly by the manager to tell this to check and see if
//any timers have fired, and execute them if so.
func (t *timerManager) Tick() {
	for t.nextTimerFired() {
		record := t.popNext(false)
		if record == nil {
			continue
		}

		if err := <-record.game.ProposeMove(record.move, AdminPlayerIndex); err != nil {
			//TODO: log the error or something
			t.manager.Logger().Info("When timer failed the move could not be made: ", err, record.move)
		}
	}
}

//Whether the next timer in the queue is already fired
func (t *timerManager) nextTimerFired() bool {
	if len(t.records) == 0 {
		return false
	}

	record := t.records[0]

	return record.TimeRemaining() <= 0
}

func (t *timerManager) popNext(force bool) *timerRecord {
	if force {
		if len(t.records) == 0 {
			return nil
		}
	} else {
		if !t.nextTimerFired() {
			return nil
		}
	}

	x := heap.Pop(&t.records)

	record := x.(*timerRecord)

	delete(t.recordsById, record.id)

	return record
}

func (t timerQueue) Len() int {
	return len(t)
}

func (t timerQueue) Less(i, j int) bool {
	return t[i].fireTime.Sub(t[j].fireTime) < 0
}

func (t timerQueue) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
	t[i].index = i
	t[j].index = j
}

//DO NOT USE THIS DIRECTLY. Use heap.Push(t)
func (t *timerQueue) Push(x interface{}) {
	n := len(*t)
	item := x.(*timerRecord)
	item.index = n
	*t = append(*t, item)
}

//DO NOT USE THIS DIRECTLY. Use heap.Pop()
func (t *timerQueue) Pop() interface{} {
	old := *t
	n := len(old)
	item := old[n-1]
	item.index = -1
	*t = old[0 : n-1]
	return item
}
