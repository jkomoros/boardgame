package enum

import (
	"slices"
	"testing"

	"github.com/workfit/tester/assert"
)

func TestRangedEnum(t *testing.T) {

	tests := []struct {
		indexes        []int
		errExpected    bool
		expectedValues map[EnumKey]string
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
			map[EnumKey]string{
				0: "0",
				1: "1",
			},
		},
		{
			[]int{2, 3},
			false,
			map[EnumKey]string{
				0: "0_0",
				1: "0_1",
				2: "0_2",
				3: "1_0",
				4: "1_1",
				5: "1_2",
			},
		},
		{
			[]int{1, 2, 2},
			false,
			map[EnumKey]string{
				0: "0_0_0",
				1: "0_0_1",
				2: "0_1_0",
				3: "0_1_1",
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

	val := theEnum.NewRangeVal()
	assert.For(t).ThatActual(val.RangeValue()).Equals([]int{0, 0, 0})

	err := val.SetRangeValue(0, 1, 1)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(val.RangeValue()).Equals([]int{0, 1, 1})

	//The first index of 1 is illegal, should fail
	err = val.SetRangeValue(1, 1, 1)

	assert.For(t).ThatActual(err).IsNotNil()

	//Verify that after a failed set the value didn't change.
	assert.For(t).ThatActual(val.RangeValue()).Equals([]int{0, 1, 1})

	assert.For(t).ThatActual(theEnum.RangeToValue(0, 1, 1)).Equals(EnumKey(3))

	assert.For(t).ThatActual(theEnum.ValueToRange(3)).Equals([]int{0, 1, 1})

}

func TestCombine(t *testing.T) {
	setOne := NewSet()

	a, err := setOne.Add("a", map[EnumKey]string{
		0: "Zero",
	})

	assert.For(t).ThatActual(err).IsNil()

	setTwo := NewSet()

	b, err := setTwo.Add("b", map[EnumKey]string{
		1: "One",
	})

	assert.For(t).ThatActual(err).IsNil()

	c, err := setTwo.Add("c", map[EnumKey]string{
		1: "Not One",
	})

	assert.For(t).ThatActual(err).IsNil()

	d, err := setTwo.Add("d", map[EnumKey]string{
		2: "Zero",
	})

	assert.For(t).ThatActual(err).IsNil()

	ab, err := setTwo.Combine("a+b", a, b)

	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(a.SubsetOf(ab)).IsTrue()
	assert.For(t).ThatActual(b.SubsetOf(ab)).IsTrue()
	assert.For(t).ThatActual(c.SubsetOf(ab)).IsFalse()
	assert.For(t).ThatActual(ab.SubsetOf(ab)).IsTrue()

	intValues := ab.Values()
	slices.Sort(intValues)

	assert.For(t).ThatActual(intValues).Equals([]EnumKey{0, 1})
	assert.For(t).ThatActual(ab.String(1)).Equals("One")

	//c overlaps with b's int key
	_, err = setTwo.Combine("a+b+c", a, b, c)
	assert.For(t).ThatActual(err).IsNotNil()

	//d overlaps with string key Zero
	_, err = setTwo.Combine("a+b+d", a, b, d)
	assert.For(t).ThatActual(err).IsNotNil()

}

func TestNormalizeStringKey(t *testing.T) {
	enums := NewSet()

	//Two values that have the same normalized key may not be included
	_, err := enums.Add("A", map[EnumKey]string{
		0: "Zero",
		1: "zero ",
	})

	assert.For(t).ThatActual(err).IsNotNil()

	theEnum, err := enums.Add("B", map[EnumKey]string{
		0: "Zero",
	})

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(theEnum.String(0)).Equals("Zero")

	assert.For(t).ThatActual(theEnum.ValueFromString(" zero")).Equals(EnumKey(0))

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

	colorEnum, err := enums.Add("Color", map[EnumKey]string{
		ColorBlue:  "Blue",
		ColorGreen: "Green",
		ColorRed:   "Red",
	})

	assert.For(t).ThatActual(colorEnum).IsNotNil()

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(colorEnum.Name()).Equals("Color")

	assert.For(t).ThatActual(len(enums.EnumNames())).Equals(1)

	assert.For(t).ThatActual(enums.Enum("Color")).Equals(colorEnum)

	assert.For(t).ThatActual(colorEnum.DefaultValue()).Equals(EnumKey(ColorBlue))

	assert.For(t).ThatActual(colorEnum.String(ColorBlue)).Equals("Blue")

	assert.For(t).ThatActual(colorEnum.String(125)).Equals("")

	assert.For(t).ThatActual(colorEnum.MaxValue()).Equals(EnumKey(2))

	_, err = enums.Add("Color", map[EnumKey]string{
		ColorBlue: "Blue",
	})

	assert.For(t).ThatActual(err).IsNotNil()

	cardEnum, err := enums.Add("Card", map[EnumKey]string{
		CardSpade:   "Spade",
		CardClub:    "Club",
		CardDiamond: "Diamond",
		CardHeart:   "Heart",
	})

	assert.For(t).ThatActual(cardEnum).IsNotNil()

	assert.For(t).ThatActual(err).IsNil()

	val := colorEnum.ValueFromString("Blue")

	assert.For(t).ThatActual(val).Equals(EnumKey(ColorBlue))

	val = colorEnum.ValueFromString("Turquoise")

	assert.For(t).ThatActual(val).Equals(IllegalValue)

	eVal := colorEnum.NewVal()

	assert.For(t).ThatActual(eVal.Value()).Equals(EnumKey(ColorBlue))

	err = eVal.SetValue(ColorGreen)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(eVal.Value()).Equals(EnumKey(ColorGreen))

	otherVal := colorEnum.NewVal()

	otherVal.SetValue(ColorGreen)

	assert.For(t).ThatActual(eVal.Equals(otherVal)).IsTrue()
	assert.For(t).ThatActual(otherVal.Equals(eVal)).IsTrue()

	otherVal.SetValue(ColorBlue)

	assert.For(t).ThatActual(eVal.Equals(otherVal)).IsFalse()

	err = eVal.SetStringValue("Blue")

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(eVal.Value()).Equals(EnumKey(ColorBlue))

	err = eVal.SetStringValue("Turquoise")

	assert.For(t).ThatActual(err).IsNotNil()

	constant, err := colorEnum.NewImmutableVal(ColorGreen)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(constant.Value()).Equals(EnumKey(ColorGreen))

	constant, err = colorEnum.NewImmutableVal(150)

	assert.For(t).ThatActual(err).IsNotNil()

	//Do a new manager to check that adding enums after finished doesn't work

	enums = NewSet()

	_, err = enums.Add("Color", map[EnumKey]string{
		ColorBlue:  "Blue",
		ColorGreen: "Green",
		ColorRed:   "Blue",
	})

	assert.For(t).ThatActual(err).IsNotNil()

	enums.Finish()

	_, err = enums.Add("Card", map[EnumKey]string{
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

	colorEnum, err := firstSet.Add("Color", map[EnumKey]string{
		ColorBlue:  "Blue",
		ColorGreen: "Green",
		ColorRed:   "Red",
	})

	cardEnum, err := secondSet.Add("Card", map[EnumKey]string{
		CardSpade:   "Spade",
		CardClub:    "Club",
		CardDiamond: "Diamond",
		CardHeart:   "Heart",
	})

	combinedSet, err := CombineSets(firstSet, secondSet)

	assert.For(t).ThatActual(err).IsNil()

	enumNames := combinedSet.EnumNames()

	slices.Sort(enumNames)

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
	_, err := set.Add("Color", map[EnumKey]string{
		ColorBlue:  "Blue",
		ColorGreen: "2",
		ColorRed:   "Red",
	})

	assert.For(t).ThatActual(err).IsNotNil()

	//Illegal becuase ColorGreen's string value overlaps with already-existing
	//int ColorBlue.
	_, err = set.Add("Color", map[EnumKey]string{
		ColorBlue:  "Blue",
		ColorGreen: "0",
		ColorRed:   "Red",
	})

	assert.For(t).ThatActual(err).IsNotNil()

	//Legal because ColorGreen is 1, so it may have the string value of 1.
	_, err = set.Add("Color", map[EnumKey]string{
		ColorBlue:  "Blue",
		ColorGreen: "1",
		ColorRed:   "Red",
	})

	assert.For(t).ThatActual(err).IsNil()

}

func TestEnumSlice(t *testing.T) {
	set := NewSet()

	const (
		ColorBlue = iota
		ColorGreen
		ColorRed
	)

	colorEnum, err := set.Add("Color", map[EnumKey]string{
		ColorBlue:  "Blue",
		ColorGreen: "Green",
		ColorRed:   "Red",
	})

	assert.For(t).ThatActual(err).IsNil()

	// NewEnumSlice returns an empty slice
	s := colorEnum.NewEnumSlice()
	assert.For(t).ThatActual(s.Len()).Equals(0)
	assert.For(t).ThatActual(s.Enum()).Equals(colorEnum)

	// Append and access
	s.Append(ColorRed, ColorBlue)
	assert.For(t).ThatActual(s.Len()).Equals(2)
	assert.For(t).ThatActual(s.Value(0)).Equals(EnumKey(ColorRed))
	assert.For(t).ThatActual(s.Value(1)).Equals(EnumKey(ColorBlue))

	// Values returns a copy
	vals := s.Values()
	assert.For(t).ThatActual(vals).Equals([]EnumKey{ColorRed, ColorBlue})
	vals[0] = ColorGreen // mutating the copy should not affect the original
	assert.For(t).ThatActual(s.Value(0)).Equals(EnumKey(ColorRed))

	// SetValue
	s.SetValue(0, ColorGreen)
	assert.For(t).ThatActual(s.Value(0)).Equals(EnumKey(ColorGreen))

	// SetValues replaces entirely
	s.SetValues([]EnumKey{ColorBlue, ColorGreen, ColorRed})
	assert.For(t).ThatActual(s.Len()).Equals(3)
	assert.For(t).ThatActual(s.Value(2)).Equals(EnumKey(ColorRed))

	// Truncate
	s.Truncate(1)
	assert.For(t).ThatActual(s.Len()).Equals(1)
	assert.For(t).ThatActual(s.Value(0)).Equals(EnumKey(ColorBlue))

	// Copy returns an independent mutable copy
	s.Append(ColorRed)
	c := s.Copy()
	assert.For(t).ThatActual(c.Len()).Equals(2)
	c.SetValue(0, ColorGreen)
	assert.For(t).ThatActual(s.Value(0)).Equals(EnumKey(ColorBlue)) // original unchanged

	// ImmutableCopy returns an independent immutable copy
	ic := s.ImmutableCopy()
	assert.For(t).ThatActual(ic.Len()).Equals(2)
	assert.For(t).ThatActual(ic.Value(0)).Equals(EnumKey(ColorBlue))

	// JSON round-trip
	data, err := s.(*enumSlice).MarshalJSON()
	assert.For(t).ThatActual(err).IsNil()

	s2 := colorEnum.NewEnumSlice()
	err = s2.(*enumSlice).UnmarshalJSON(data)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(s2.Values()).Equals(s.Values())
}

func TestMembershipSet(t *testing.T) {
	set := NewSet()

	const (
		GroupA = iota
		GroupB
		GroupC
	)

	groupEnum, err := set.Add("Group", map[EnumKey]string{
		GroupA: "A",
		GroupB: "B",
		GroupC: "C",
	})

	assert.For(t).ThatActual(err).IsNil()

	e := groupEnum.(*enum)

	// NewMembershipSet with initial members
	ms := e.NewMembershipSet(GroupA, GroupC)
	assert.For(t).ThatActual(ms.Len()).Equals(2)
	assert.For(t).ThatActual(ms.Contains(GroupA)).IsTrue()
	assert.For(t).ThatActual(ms.Contains(GroupB)).IsFalse()
	assert.For(t).ThatActual(ms.Contains(GroupC)).IsTrue()

	// Members returns sorted keys
	members := ms.Members()
	assert.For(t).ThatActual(members).Equals([]EnumKey{GroupA, GroupC})

	// Enum reference
	assert.For(t).ThatActual(ms.Enum()).Equals(groupEnum)

	// ContainsVal
	val := groupEnum.MustNewImmutableVal(GroupB)
	assert.For(t).ThatActual(ms.ContainsVal(val)).IsFalse()
	val = groupEnum.MustNewImmutableVal(GroupA)
	assert.For(t).ThatActual(ms.ContainsVal(val)).IsTrue()
	assert.For(t).ThatActual(ms.ContainsVal(nil)).IsFalse()

	// Add
	ms.Add(GroupB)
	assert.For(t).ThatActual(ms.Contains(GroupB)).IsTrue()
	assert.For(t).ThatActual(ms.Len()).Equals(3)

	// Add invalid key is silently ignored
	ms.Add(99)
	assert.For(t).ThatActual(ms.Len()).Equals(3)

	// Remove
	ms.Remove(GroupA)
	assert.For(t).ThatActual(ms.Contains(GroupA)).IsFalse()
	assert.For(t).ThatActual(ms.Len()).Equals(2)

	// Remove non-existent key is a no-op
	ms.Remove(GroupA)
	assert.For(t).ThatActual(ms.Len()).Equals(2)

	// Empty set
	empty := e.NewMembershipSet()
	assert.For(t).ThatActual(empty.Len()).Equals(0)
	assert.For(t).ThatActual(len(empty.Members())).Equals(0)

	// NewMembershipSet silently ignores invalid keys
	withInvalid := e.NewMembershipSet(GroupA, 99, GroupB)
	assert.For(t).ThatActual(withInvalid.Len()).Equals(2)
	assert.For(t).ThatActual(withInvalid.Contains(GroupA)).IsTrue()
	assert.For(t).ThatActual(withInvalid.Contains(GroupB)).IsTrue()
}
