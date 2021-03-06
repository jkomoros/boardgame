/*

Package dice is a simple package that defines die components with variable
numbers of sides.

*/
package dice

import (
	"math"
	"math/rand"

	"github.com/jkomoros/boardgame/base"
)

//go:generate boardgame-util codegen

//Value is the value component of a dice, encoding what the allowable values
//are.
//
//boardgame:codegen
type Value struct {
	base.ComponentValues
	Faces []int
}

//DynamicValue encodes which face is currently selected.
//
//boardgame:codegen
type DynamicValue struct {
	base.SubState
	base.ComponentValues
	Value        int
	SelectedFace int
}

//DefaultDie returns a die configured as as a typical six-sided die.
func DefaultDie() *Value {
	return BasicDie(1, 6)
}

//BasicDie returns a die with a face each for min through max, inclusive.
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

//Min returns the lowest value face for this die
func (v *Value) Min() int {
	min := math.MaxInt64
	for _, face := range v.Faces {
		if face < min {
			min = face
		}
	}
	return min
}

//Max returns the maximum value face for this die
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
//die Value it is associated with. Accepts a source of randomness it will use.
//You almost always should pass state.Rand() for this to have outcomes that
//are deterministic for this state (which can be useful for testing
//scenarios). If r is nil, a generic source of randomness will be used.
func (d *DynamicValue) Roll(r *rand.Rand) {

	values, ok := d.ContainingComponent().Values().(*Value)

	if !ok {
		//This shouldn't happen, unless someone has called
		//SetContainingComponent themselves after the dynamic values was
		//called, in which case they get what they deserve for not having the dice actually roll.
		return
	}

	var val int

	if r == nil {
		val = rand.Intn(len(values.Faces))
	} else {
		val = r.Intn(len(values.Faces))
	}

	d.SelectedFace = val
	d.Value = values.Faces[val]

}
