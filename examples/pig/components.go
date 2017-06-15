package pig

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"math/rand"
)

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

//Roll sets the Value of the Die randomly to a new legal value. The component
//you pass should be the same Die component that we're rolling.
func (d *dieDynamicValue) Roll(c *boardgame.Component) error {

	if c == nil {
		return errors.New("No component provided")
	}

	values, ok := c.Values.(*dieValue)

	if !ok {
		return errors.New("Component passed was not a die")
	}

	random := rand.Intn(values.Max - values.Min)

	d.Value = random + values.Min

	return nil

}
