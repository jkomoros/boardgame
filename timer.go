package boardgame

type Timer struct {
	//ID will be an opaque identifier that is used to keep track of the
	//corresponding underlying Timer object in the game engine. It is not
	//meaningful to inspect yourself and should not be modified.
	ID       int
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
		"ID":       t.ID,
		"TimeLeft": t.TimeLeft(),
	}

	return DefaultMarshalJSON(obj)
}

//Active returns true if the timer is active and counting down.
func (t *Timer) Active() bool {
	return t.ID == 0
}

//TimeLeft returns the number of nanoseconds left until this timer fires.
func (t *Timer) TimeLeft() int {
	if !t.Active() {
		return 0
	}
	//TODO: when this is actually hooked up, return a real value.
	return 0
}

//Start starts the timer. After nanoseconds nanoseconds have passed, the Move
//will be proposed via proposeMove.
func (t *Timer) Start(nanoseconds int, move Move) {
	//TODO: actually do something here
}

//Cancel cancels an active timer. If the timer is not active, it has no
//effect. Returns true if the timer was active and canceled, false if the
//timer was not active.
func (t *Timer) Cancel() bool {
	if !t.Active() {
		return false
	}
	//TODO: tell the manager to cancel here.
	return true
}
