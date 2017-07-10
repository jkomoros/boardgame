package enum

import (
	"github.com/workfit/tester/assert"
	"sort"
	"testing"
)

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

	assert.For(t).ThatActual(val).Equals(-1)

	eVal := colorEnum.NewVar()

	assert.For(t).ThatActual(eVal.Value()).Equals(ColorBlue)

	err = eVal.SetValue(ColorGreen)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(eVal.Value()).Equals(ColorGreen)

	constant, err := colorEnum.NewConst(ColorGreen)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(constant.Value()).Equals(ColorGreen)

	constant, err = colorEnum.NewConst(150)

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
