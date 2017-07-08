package boardgame

import (
	"github.com/workfit/tester/assert"
	"testing"
)

func TestEnum(t *testing.T) {
	enum := NewEnumManager()

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

	const ConstDuplicate = iota

	assert.For(t).ThatActual(enum).IsNotNil()

	err := enum.Add("Color", map[int]string{
		ColorBlue:  "Blue",
		ColorGreen: "Green",
		ColorRed:   "Red",
	})

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(enum.Membership(ColorBlue)).Equals("Color")

	assert.For(t).ThatActual(enum.DefaultValue("Color")).Equals(ColorBlue)

	assert.For(t).ThatActual(enum.String(ColorBlue)).Equals("Blue")

	assert.For(t).ThatActual(enum.String(125)).Equals("")

	err = enum.Add("Color", map[int]string{
		ColorBlue: "Blue",
	})

	assert.For(t).ThatActual(err).IsNotNil()

	err = enum.Add("Card", map[int]string{
		CardSpade:   "Spade",
		CardClub:    "Club",
		CardDiamond: "Diamond",
		CardHeart:   "Heart",
	})

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(enum.Membership(CardDiamond)).Equals("Card")

	err = enum.Add("Another", map[int]string{
		ConstDuplicate: "Duplicate",
	})

	assert.For(t).ThatActual(err).IsNotNil()

	val := enum.ValueFromString("Color", "Blue")

	assert.For(t).ThatActual(val).Equals(ColorBlue)

	val = enum.ValueFromString("Color", "Turquoise")

	assert.For(t).ThatActual(val).Equals(0)

	val = enum.ValueFromString("InvalidEnum", "Blue")

	assert.For(t).ThatActual(val).Equals(-1)

	eVal := NewEnumValue("Color")

	assert.For(t).ThatActual(eVal.Inflated()).IsFalse()

	eVal.Inflate(enum)

	assert.For(t).ThatActual(eVal.Inflated()).IsTrue()

	assert.For(t).ThatActual(eVal.Value()).Equals(ColorBlue)

	result := eVal.SetValue(ColorGreen)

	assert.For(t).ThatActual(result).IsTrue()

	assert.For(t).ThatActual(eVal.Value()).Equals(ColorGreen)

	eVal.Lock()

	result = eVal.SetValue(ColorRed)

	assert.For(t).ThatActual(result).IsFalse()

	assert.For(t).ThatActual(eVal.Value()).Equals(ColorGreen)

	//Do a new manager to check that adding enums after finished doesn't work

	enum = NewEnumManager()

	err = enum.Add("Color", map[int]string{
		ColorBlue:  "Blue",
		ColorGreen: "Green",
		ColorRed:   "Blue",
	})

	assert.For(t).ThatActual(err).IsNotNil()

	enum.Finish()

	err = enum.Add("Card", map[int]string{
		CardSpade: "Spade",
		CardClub:  "Club",
	})

	assert.For(t).ThatActual(err).IsNotNil()
}
