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

}
