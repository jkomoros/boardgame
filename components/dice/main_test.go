package dice

import (
	"github.com/jkomoros/boardgame"
	"testing"
)

func TestDieRoll(t *testing.T) {

	dynamic := &DynamicValue{
		Value: 0,
	}

	values := DefaultDie()

	die := &boardgame.Component{
		Values: values,
	}

	seenValues := make(map[int]bool)

	for i := 0; i < 10; i++ {
		if err := dynamic.Roll(die); err != nil {
			t.Fatal("Got non-nill err for roll: ", err)
		}
		if dynamic.Value < values.Min || dynamic.Value > values.Max {
			t.Error("Invalid Value after roll: ", dynamic.Value)
		}
		seenValues[dynamic.Value] = true
	}

	if len(seenValues) < 3 {
		t.Error("We didn't see enough different values across 10 rolls, which is suspicious.", len(seenValues))
	}

}
