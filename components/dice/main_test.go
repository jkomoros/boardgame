package dice

import (
	"github.com/jkomoros/boardgame"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestBasicDie(t *testing.T) {

	tests := []struct {
		die         *Value
		expectEmpty bool
		min         int
		max         int
		numFaces    int
	}{
		{
			DefaultDie(),
			false,
			1,
			6,
			6,
		},
		{
			BasicDie(6, 1),
			true,
			0,
			0,
			0,
		},
		{
			BasicDie(0, 100),
			false,
			0,
			100,
			101,
		},
	}

	var nilDie *Value

	for i, test := range tests {
		if test.expectEmpty {
			assert.For(t, i).ThatActual(test.die).Equals(nilDie)
			continue
		} else {
			assert.For(t, i).ThatActual(test.die).IsNotNil()
		}
		assert.For(t, i).ThatActual(test.die.Min()).Equals(test.min)
		assert.For(t, i).ThatActual(test.die.Max()).Equals(test.max)
		assert.For(t, i).ThatActual(len(test.die.Faces)).Equals(test.numFaces)
	}
}

func TestDieRoll(t *testing.T) {

	dynamic := &DynamicValue{
		Value: 0,
	}

	values := DefaultDie()

	die := &boardgame.Component{
		Values: values,
	}

	seenValues := make(map[int]bool)

	min := values.Min()
	max := values.Max()

	for i := 0; i < 10; i++ {
		if err := dynamic.Roll(die); err != nil {
			t.Fatal("Got non-nill err for roll: ", err)
		}

		assert.For(t).ThatActual(dynamic.Value).Equals(values.Faces[dynamic.SelectedFace])

		if dynamic.Value < min || dynamic.Value > max {
			t.Error("Invalid Value after roll: ", dynamic.Value)
		}
		seenValues[dynamic.Value] = true
	}

	if len(seenValues) < 3 {
		t.Error("We didn't see enough different values across 10 rolls, which is suspicious.", len(seenValues))
	}

}
