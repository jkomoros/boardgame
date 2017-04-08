package boardgame

import (
	"container/heap"
	"time"
)

type Timer struct {
	//Id will be an opaque identifier that is used to keep track of the
	//corresponding underlying Timer object in the game engine. It is not
	//meaningful to inspect yourself and should not be modified.
	Id       int
	statePtr *state
}

func NewTimer() *Timer {
	return &Timer{}
}

func (t *Timer) Copy() *Timer {
	var result Timer
	result = *t
	return &result
}

func (t *Timer) MarshalJSON() ([]byte, error) {
	obj := map[string]interface{}{
		"Id":       t.Id,
		"TimeLeft": t.TimeLeft(),
	}

	return DefaultMarshalJSON(obj)
}

//Active returns true if the timer is active and counting down.
func (t *Timer) Active() bool {
	return t.Id == 0 && t.TimeLeft() > 0
}

//TimeLeft returns the number of nanoseconds left until this timer fires.
func (t *Timer) TimeLeft() time.Duration {

	return t.statePtr.game.manager.timers.GetTimerRemaining(t.Id)
}

//Start starts the timer. After duration has passed, the Move
//will be proposed via proposeMove.
func (t *Timer) Start(duration time.Duration, move Move) {
	game := t.statePtr.game
	manager := game.manager

	t.Id = manager.timers.RegisterTimer(duration, game, move)
}

//Cancel cancels an active timer. If the timer is not active, it has no
//effect. Returns true if the timer was active and canceled, false if the
//timer was not active.
func (t *Timer) Cancel() bool {
	if !t.Active() {
		return false
	}
	manager := t.statePtr.game.manager

	manager.timers.CancelTimer(t.Id)

	t.Id = 0

	return true
}

type timerRecord struct {
	id       int
	index    int
	fireTime time.Time
	game     *Game
	move     Move
}

func (t *timerRecord) TimeRemaining() time.Duration {
	duration := t.fireTime.Sub(time.Now())

	if duration < 0 {
		duration = 0
	}

	return duration
}

type timerQueue []*timerRecord

type timerManager struct {
	nextId      int
	records     timerQueue
	recordsById map[int]*timerRecord
}

func newTimerManager() *timerManager {
	return &timerManager{
		//the default id in TimerProps is 0, so we should start beyond that.
		nextId:      1,
		records:     make(timerQueue, 0),
		recordsById: make(map[int]*timerRecord),
	}
}

func (t *timerManager) RegisterTimer(duration time.Duration, game *Game, move Move) int {
	record := &timerRecord{
		id:       t.nextId,
		index:    -1,
		fireTime: time.Now().Add(duration),
		game:     game,
		move:     move,
	}
	t.nextId++

	t.recordsById[record.id] = record

	heap.Push(&t.records, record)

	return record.id
}

func (t *timerManager) GetTimerRemaining(id int) time.Duration {
	record := t.recordsById[id]

	if record == nil {
		return 0
	}

	return record.TimeRemaining()
}

func (t *timerManager) CancelTimer(id int) {
	record := t.recordsById[id]

	if record == nil {
		return
	}

	heap.Remove(&t.records, record.index)

	record.index = -1

	delete(t.recordsById, record.id)

}

//Should be called regularly by the manager to tell this to check and see if
//any timers have fired, and execute them if so.
func (t *timerManager) Tick() {
	for t.nextTimerFired() {
		record := t.popNext()
		if record == nil {
			continue
		}
		if err := <-record.game.ProposeMove(record.move); err != nil {
			//TODO: log the error or something
			panic(err)
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

func (t *timerManager) popNext() *timerRecord {
	if !t.nextTimerFired() {
		return nil
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
