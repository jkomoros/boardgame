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

	err := enum.Add("Color", ColorBlue, ColorGreen, ColorRed)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(enum.Membership(ColorBlue)).Equals("Color")

	assert.For(t).ThatActual(enum.DefaultValue("Color")).Equals(ColorBlue)

	err = enum.Add("Color", ColorBlue)

	assert.For(t).ThatActual(err).IsNotNil()

	err = enum.Add("Card", CardSpade, CardClub, CardDiamond, CardHeart)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(enum.Membership(CardDiamond)).Equals("Card")

	err = enum.Add("Another", ConstDuplicate)

	assert.For(t).ThatActual(err).IsNotNil()

}
