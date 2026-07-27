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

	deck := boardgame.NewDeck()
	deck.AddComponent(values)
	die := deck.ComponentAt(0)

	dynamic.SetContainingComponent(die)

	seenValues := make(map[int]bool)

	min := values.Min()
	max := values.Max()

	for i := 0; i < 10; i++ {
		dynamic.Roll(nil)

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

//TestRollCountCountsEveryThrow pins the one thing RollCount exists for: a throw
//that lands on the face already showing changes neither SelectedFace nor Value,
//so without the counter a client cannot tell it from a state in which this die
//was not thrown at all. About one throw in six for a six-sided die -- and a
//renderer that animates rolls simply did not animate those.
func TestRollCountCountsEveryThrow(t *testing.T) {

	dynamic := &DynamicValue{}

	values := DefaultDie()

	deck := boardgame.NewDeck()
	deck.AddComponent(values)
	die := deck.ComponentAt(0)

	dynamic.SetContainingComponent(die)

	assert.For(t).ThatActual(dynamic.RollCount).Equals(0)

	sameFaceThrows := 0

	for i := 0; i < 200; i++ {
		beforeFace := dynamic.SelectedFace
		beforeValue := dynamic.Value
		beforeCount := dynamic.RollCount

		dynamic.Roll(nil)

		//Every throw, without exception.
		assert.For(t, i).ThatActual(dynamic.RollCount).Equals(beforeCount + 1)

		if i > 0 && dynamic.SelectedFace == beforeFace {
			sameFaceThrows++
			//The case the counter exists for: nothing else moved.
			assert.For(t, i).ThatActual(dynamic.Value).Equals(beforeValue)
		}
	}

	//If this ever comes back zero the test has stopped exercising the case it
	//was written for, which is louder than a silently vacuous pass. The
	//expected count over 199 throws of a d6 is about 33.
	if sameFaceThrows == 0 {
		t.Error("no throw in 200 landed on the face already showing, so the case RollCount exists for was never exercised")
	}

}
