package boardgame

type Timer struct {
	//TODO: actually store fields
}

func NewTimer() *Timer {
	return &Timer{}
}

func (t *Timer) Copy() *Timer {
	var result Timer
	result = *t
	return &result
}
