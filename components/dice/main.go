/*
 * dice is a simple package that defines die components with variable numbers of sides.
 */
package dice

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"math/rand"
)

//go:generate autoreader

//+autoreader
type Value struct {
	Min int
	Max int
}

//+autoreader
type DynamicValue struct {
	Value int
}

func DefaultDie() *Value {
	return &Value{
		Min: 1,
		Max: 6,
	}
}

//Roll sets the Value of the Die randomly to a new legal value. The component
//you pass should be the same Die component that we're rolling.
func (d *DynamicValue) Roll(c *boardgame.Component) error {

	if c == nil {
		return errors.New("No component provided")
	}

	values, ok := c.Values.(*Value)

	if !ok {
		return errors.New("Component passed was not a die")
	}

	random := rand.Intn(values.Max - values.Min)

	d.Value = random + values.Min

	return nil

}
