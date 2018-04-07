/*

dice is a simple package that defines die components with variable numbers of sides.

*/
package dice

import (
	"github.com/jkomoros/boardgame"
	"math"
	"math/rand"
)

//go:generate autoreader

//+autoreader
type Value struct {
	boardgame.BaseComponentValues
	Faces []int
}

//+autoreader
type DynamicValue struct {
	boardgame.BaseSubState
	boardgame.BaseComponentValues
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

//Roll sets the Value of the Die randomly to a new value that is legal for the
//die Value it is associated with.
func (d *DynamicValue) Roll() {

	values, ok := d.ContainingComponent().Values.(*Value)

	if !ok {
		//This shouldn't happen, unless someone has called
		//SetContainingComponent themselves after the dynamic values was
		//called, in which case they get what they deserve for not having the dice actually roll.
		return
	}

	random := rand.Intn(len(values.Faces))

	d.SelectedFace = random
	d.Value = values.Faces[random]

}
