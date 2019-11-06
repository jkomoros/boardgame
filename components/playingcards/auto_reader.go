/************************************
 *
 * This file contains auto-generated methods to help certain structs
 * implement boardgame.PropertyReader and friends. It was generated
 * by the codegen package via 'boardgame-util codegen'.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package playingcards

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// Implementation for Card

var ȧutoGeneratedCardReaderProps = map[string]boardgame.PropertyType{
	"Rank": boardgame.TypeEnum,
	"Suit": boardgame.TypeEnum,
}

type ȧutoGeneratedCardReader struct {
	data *Card
}

func (c *ȧutoGeneratedCardReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedCardReaderProps
}

func (c *ȧutoGeneratedCardReader) Prop(name string) (interface{}, error) {
	props := c.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return c.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return c.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return c.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return c.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return c.IntProp(name)
	case boardgame.TypeIntSlice:
		return c.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return c.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return c.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return c.ImmutableStackProp(name)
	case boardgame.TypeString:
		return c.StringProp(name)
	case boardgame.TypeStringSlice:
		return c.StringSliceProp(name)
	case boardgame.TypeTimer:
		return c.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (c *ȧutoGeneratedCardReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *ȧutoGeneratedCardReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (c *ȧutoGeneratedCardReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (c *ȧutoGeneratedCardReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "Rank":
		return c.data.Rank, nil
	case "Suit":
		return c.data.Suit, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *ȧutoGeneratedCardReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (c *ȧutoGeneratedCardReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (c *ȧutoGeneratedCardReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (c *ȧutoGeneratedCardReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *ȧutoGeneratedCardReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *ȧutoGeneratedCardReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (c *ȧutoGeneratedCardReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (c *ȧutoGeneratedCardReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for Card
func (c *Card) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedCardReader{c}
}
