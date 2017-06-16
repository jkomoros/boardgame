/*
 * dice is a simple package that defines die components with variable numbers of sides.
 */
package dice

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"math"
	"math/rand"
)

//go:generate autoreader

//+autoreader
type Value struct {
	Faces []int
}

//+autoreader
type DynamicValue struct {
	Value        int
	SelectedFace int
}

func DefaultDie() *Value {
	return BasicDie(1, 6)
}

func BasicDie(min, max int) *Value {

	if min >= max {
		return nil
	}

	var faces []int
	for i := min; i <= max; i++ {
		faces = append(faces, i)
	}
	return &Value{
		Faces: faces,
	}
}

func (v *Value) Min() int {
	min := math.MaxInt64
	for _, face := range v.Faces {
		if face < min {
			min = face
		}
	}
	return min
}

func (v *Value) Max() int {
	max := math.MinInt64
	for _, face := range v.Faces {
		if face > max {
			max = face
		}
	}
	return max
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

	random := rand.Intn(len(values.Faces))

	d.SelectedFace = random
	d.Value = values.Faces[random]

	return nil

}
