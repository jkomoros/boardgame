package pig

const diceDeckName = "dice"

//+autoreader
type dieValue struct {
	Min int
	Max int
}

//+autoreader
type dieDynamicValue struct {
	Value int
}

func DefaultDie() *dieValue {
	return &dieValue{
		Min: 1,
		Max: 6,
	}
}
