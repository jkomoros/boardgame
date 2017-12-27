package enum

import (
	"github.com/workfit/tester/assert"
	"sort"
	"testing"
)

func TestRangedEnum(t *testing.T) {

	tests := []struct {
		indexes        []int
		errExpected    bool
		expectedValues map[int]string
	}{
		{
			[]int{},
			true,
			nil,
		},
		{
			[]int{3, 0, 1},
			true,
			nil,
		},
		{
			[]int{2},
			false,
			map[int]string{
				0: "0",
				1: "1",
			},
		},
		{
			[]int{2, 3},
			false,
			map[int]string{
				0: "0,0",
				1: "0,1",
				2: "0,2",
				3: "1,0",
				4: "1,1",
				5: "1,2",
			},
		},
		{
			[]int{1, 2, 2},
			false,
			map[int]string{
				0: "0,0,0",
				1: "0,0,1",
				2: "0,1,0",
				3: "0,1,1",
			},
		},
	}

	for i, test := range tests {
		set := NewSet()
		theEnumRaw, err := set.AddRange("theEnum", test.indexes...)
		if test.errExpected {
			assert.For(t, i).ThatActual(err).IsNotNil()
			continue
		} else {
			assert.For(t, i).ThatActual(err).IsNil()
		}

		theEnum := theEnumRaw.(*enum)

		assert.For(t, i).ThatActual(len(theEnum.values)).Equals(len(test.expectedValues))

		for key, val := range test.expectedValues {
			realVal := theEnum.String(key)
			assert.For(t, i).ThatActual(realVal).Equals(val)
		}
	}

	set := NewSet()
	theEnum, _ := set.AddRange("theEnum", 1, 2, 2)

	val := theEnum.NewMutableVal()
	assert.For(t).ThatActual(val.RangeValue()).Equals([]int{0, 0, 0})

	err := val.SetRangeValue(0, 1, 1)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(val.RangeValue()).Equals([]int{0, 1, 1})

	//The first index of 1 is illegal, should fail
	err = val.SetRangeValue(1, 1, 1)

	assert.For(t).ThatActual(err).IsNotNil()

	//Verify that after a failed set the value didn't change.
	assert.For(t).ThatActual(val.RangeValue()).Equals([]int{0, 1, 1})

}

func TestEnum(t *testing.T) {
	enums := NewSet()

	assert.For(t).ThatActual(len(enums.EnumNames())).Equals(0)

	const (
		ColorBlue = iota
		ColorGreen
		ColorRed
	)

	const (
		CardSpade = iota
		CardClub
		CardDiamond
		CardHeart
	)

	assert.For(t).ThatActual(enums).IsNotNil()

	colorEnum, err := enums.Add("Color", map[int]string{
		ColorBlue:  "Blue",
		ColorGreen: "Green",
		ColorRed:   "Red",
	})

	assert.For(t).ThatActual(colorEnum).IsNotNil()

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(colorEnum.Name()).Equals("Color")

	assert.For(t).ThatActual(len(enums.EnumNames())).Equals(1)

	assert.For(t).ThatActual(enums.Enum("Color")).Equals(colorEnum)

	assert.For(t).ThatActual(colorEnum.DefaultValue()).Equals(ColorBlue)

	assert.For(t).ThatActual(colorEnum.String(ColorBlue)).Equals("Blue")

	assert.For(t).ThatActual(colorEnum.String(125)).Equals("")

	_, err = enums.Add("Color", map[int]string{
		ColorBlue: "Blue",
	})

	assert.For(t).ThatActual(err).IsNotNil()

	cardEnum, err := enums.Add("Card", map[int]string{
		CardSpade:   "Spade",
		CardClub:    "Club",
		CardDiamond: "Diamond",
		CardHeart:   "Heart",
	})

	assert.For(t).ThatActual(cardEnum).IsNotNil()

	assert.For(t).ThatActual(err).IsNil()

	val := colorEnum.ValueFromString("Blue")

	assert.For(t).ThatActual(val).Equals(ColorBlue)

	val = colorEnum.ValueFromString("Turquoise")

	assert.For(t).ThatActual(val).Equals(IllegalValue)

	eVal := colorEnum.NewMutableVal()

	assert.For(t).ThatActual(eVal.Value()).Equals(ColorBlue)

	err = eVal.SetValue(ColorGreen)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(eVal.Value()).Equals(ColorGreen)

	otherVal := colorEnum.NewMutableVal()

	otherVal.SetValue(ColorGreen)

	assert.For(t).ThatActual(eVal.Equals(otherVal)).IsTrue()
	assert.For(t).ThatActual(otherVal.Equals(eVal)).IsTrue()

	otherVal.SetValue(ColorBlue)

	assert.For(t).ThatActual(eVal.Equals(otherVal)).IsFalse()

	err = eVal.SetStringValue("Blue")

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(eVal.Value()).Equals(ColorBlue)

	err = eVal.SetStringValue("Turquoise")

	assert.For(t).ThatActual(err).IsNotNil()

	constant, err := colorEnum.NewVal(ColorGreen)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(constant.Value()).Equals(ColorGreen)

	constant, err = colorEnum.NewVal(150)

	assert.For(t).ThatActual(err).IsNotNil()

	//Do a new manager to check that adding enums after finished doesn't work

	enums = NewSet()

	_, err = enums.Add("Color", map[int]string{
		ColorBlue:  "Blue",
		ColorGreen: "Green",
		ColorRed:   "Blue",
	})

	assert.For(t).ThatActual(err).IsNotNil()

	enums.Finish()

	_, err = enums.Add("Card", map[int]string{
		CardSpade: "Spade",
		CardClub:  "Club",
	})

	assert.For(t).ThatActual(err).IsNotNil()
}

func TestCombinedEnumSets(t *testing.T) {

	firstSet := NewSet()
	secondSet := NewSet()

	const (
		ColorBlue = iota
		ColorGreen
		ColorRed
	)

	const (
		CardSpade = ColorRed + 1 + iota
		CardClub
		CardDiamond
		CardHeart
	)

	colorEnum, err := firstSet.Add("Color", map[int]string{
		ColorBlue:  "Blue",
		ColorGreen: "Green",
		ColorRed:   "Red",
	})

	cardEnum, err := secondSet.Add("Card", map[int]string{
		CardSpade:   "Spade",
		CardClub:    "Club",
		CardDiamond: "Diamond",
		CardHeart:   "Heart",
	})

	combinedSet, err := CombineSets(firstSet, secondSet)

	assert.For(t).ThatActual(err).IsNil()

	enumNames := combinedSet.EnumNames()

	sort.Strings(enumNames)

	assert.For(t).ThatActual(enumNames).Equals([]string{"Card", "Color"})

	assert.For(t).ThatActual(combinedSet.Enum("Color")).Equals(colorEnum)
	assert.For(t).ThatActual(combinedSet.Enum("Card")).Equals(cardEnum)
}

func TestIntStringOverlap(t *testing.T) {

	set := NewSet()

	const (
		ColorBlue = iota
		ColorGreen
		ColorRed
	)

	//Illegal because ColorRed value will overlap with ColorGreen's string
	//value.
	_, err := set.Add("Color", map[int]string{
		ColorBlue:  "Blue",
		ColorGreen: "2",
		ColorRed:   "Red",
	})

	assert.For(t).ThatActual(err).IsNotNil()

	//Illegal becuase ColorGreen's string value overlaps with already-existing
	//int ColorBlue.
	_, err = set.Add("Color", map[int]string{
		ColorBlue:  "Blue",
		ColorGreen: "0",
		ColorRed:   "Red",
	})

	assert.For(t).ThatActual(err).IsNotNil()

	//Legal because ColorGreen is 1, so it may have the string value of 1.
	_, err = set.Add("Color", map[int]string{
		ColorBlue:  "Blue",
		ColorGreen: "1",
		ColorRed:   "Red",
	})

	assert.For(t).ThatActual(err).IsNil()

}
