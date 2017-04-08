package boardgame

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
	return t.Id == 0
}

//TimeLeft returns the number of nanoseconds left until this timer fires.
func (t *Timer) TimeLeft() int {
	if !t.Active() {
		return 0
	}

	return t.statePtr.game.manager.getTimer(t.Id)
}

//Start starts the timer. After nanoseconds nanoseconds have passed, the Move
//will be proposed via proposeMove.
func (t *Timer) Start(nanoseconds int, move Move) {
	game := t.statePtr.game
	manager := game.manager

	t.Id = manager.registerTimer(nanoseconds, game, move)
}

//Cancel cancels an active timer. If the timer is not active, it has no
//effect. Returns true if the timer was active and canceled, false if the
//timer was not active.
func (t *Timer) Cancel() bool {
	if !t.Active() {
		return false
	}
	manager := t.statePtr.game.manager

	manager.cancelTimer(t.Id)

	t.Id = 0

	return true
}

/**************************************
 *
 * GameManager timer-related methods
 *
 **************************************/

//registerTimer registers an implementation level-timer, so that after
//nanoseconds have passed the given move will be proposed to the given game.
//Returns an Id that can be used ot cancel or fetch information on the timer
//in the future. Only designed to be called from Timer.Start()
func (g *GameManager) registerTimer(nanoseconds int, game *Game, move Move) int {
	//TOOD: actually implement this
	return 0
}

//cancelTimer cancels the timer with the given Id. IF there there is no such
//timer, or that timer is already over, it is a no-op.
func (g *GameManager) cancelTimer(id int) {
	//TODO: actually implement this.
}

//getTimer returns the number of nanoseconds left for the timer with the given
//id.
func (g *GameManager) getTimer(id int) int {
	//TODO: actually implement this.
	return 0
}
