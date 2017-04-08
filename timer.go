package boardgame

type Timer struct {
	//ID will be an opaque identifier that is used to keep track of the
	//corresponding underlying Timer object in the game engine. It is not
	//meaningful to inspect yourself and should not be modified.
	ID int
}

func NewTimer() *Timer {
	return &Timer{}
}

func (t *Timer) Copy() *Timer {
	var result Timer
	result = *t
	return &result
}

//Active returns true if the timer is active and counting down.
func (t *Timer) Active() bool {
	return t.ID == 0
}
