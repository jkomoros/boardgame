/************************************
 *
 * This file contains auto-generated methods to help certain structs
 * implement boardgame.PropertyReader and friends. It was generated
 * by the codegen package via 'boardgame-util codegen'.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package cgame

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// Implementation for cardValue

var ȧutoGeneratedCardValueReaderProps = map[string]boardgame.PropertyType{
	"Color": boardgame.TypeEnum,
	"Face":  boardgame.TypeBool,
	"Pips":  boardgame.TypeIntSlice,
	"Rank":  boardgame.TypeInt,
	"Suit":  boardgame.TypeString,
}

type ȧutoGeneratedCardValueReader struct {
	data *cardValue
}

func (c *ȧutoGeneratedCardValueReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedCardValueReaderProps
}

func (c *ȧutoGeneratedCardValueReader) Prop(name string) (interface{}, error) {
	props := c.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return c.IntProp(name)
	case boardgame.TypeBool:
		return c.BoolProp(name)
	case boardgame.TypeString:
		return c.StringProp(name)
	case boardgame.TypePlayerIndex:
		return c.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return c.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return c.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return c.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return c.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return c.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return c.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return c.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return c.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return c.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (c *ȧutoGeneratedCardValueReader) IntProp(name string) (int, error) {

	switch name {
	case "Rank":
		return c.data.Rank, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) BoolProp(name string) (bool, error) {

	switch name {
	case "Face":
		return c.data.Face, nil

	}

	return false, errors.New("No such Bool prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) StringProp(name string) (string, error) {

	switch name {
	case "Suit":
		return c.data.Suit, nil

	}

	return "", errors.New("No such String prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "Color":
		return c.data.Color, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) IntSliceProp(name string) ([]int, error) {

	switch name {
	case "Pips":
		return c.data.Pips, nil

	}

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for cardValue
func (c *cardValue) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedCardValueReader{c}
}

// Implementation for chipValue

var ȧutoGeneratedChipValueReaderProps = map[string]boardgame.PropertyType{
	"Value": boardgame.TypeInt,
}

type ȧutoGeneratedChipValueReader struct {
	data *chipValue
}

func (c *ȧutoGeneratedChipValueReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedChipValueReaderProps
}

func (c *ȧutoGeneratedChipValueReader) Prop(name string) (interface{}, error) {
	props := c.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return c.IntProp(name)
	case boardgame.TypeBool:
		return c.BoolProp(name)
	case boardgame.TypeString:
		return c.StringProp(name)
	case boardgame.TypePlayerIndex:
		return c.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return c.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return c.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return c.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return c.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return c.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return c.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return c.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return c.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return c.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (c *ȧutoGeneratedChipValueReader) IntProp(name string) (int, error) {

	switch name {
	case "Value":
		return c.data.Value, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (c *ȧutoGeneratedChipValueReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (c *ȧutoGeneratedChipValueReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (c *ȧutoGeneratedChipValueReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (c *ȧutoGeneratedChipValueReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *ȧutoGeneratedChipValueReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (c *ȧutoGeneratedChipValueReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (c *ȧutoGeneratedChipValueReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (c *ȧutoGeneratedChipValueReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *ȧutoGeneratedChipValueReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *ȧutoGeneratedChipValueReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *ȧutoGeneratedChipValueReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (c *ȧutoGeneratedChipValueReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for chipValue
func (c *chipValue) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedChipValueReader{c}
}

// Implementation for cardDynamic

var ȧutoGeneratedCardDynamicReaderProps = map[string]boardgame.PropertyType{
	"Mark": boardgame.TypeInt,
	"Nest": boardgame.TypeStack,
	"Tag":  boardgame.TypeString,
}

type ȧutoGeneratedCardDynamicReader struct {
	data *cardDynamic
}

func (c *ȧutoGeneratedCardDynamicReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedCardDynamicReaderProps
}

func (c *ȧutoGeneratedCardDynamicReader) Prop(name string) (interface{}, error) {
	props := c.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return c.IntProp(name)
	case boardgame.TypeBool:
		return c.BoolProp(name)
	case boardgame.TypeString:
		return c.StringProp(name)
	case boardgame.TypePlayerIndex:
		return c.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return c.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return c.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return c.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return c.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return c.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return c.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return c.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return c.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return c.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (c *ȧutoGeneratedCardDynamicReader) PropMutable(name string) bool {
	switch name {
	case "Mark":
		return true
	case "Nest":
		return true
	case "Tag":
		return true
	}

	return false
}

func (c *ȧutoGeneratedCardDynamicReader) SetProp(name string, value interface{}) error {
	props := c.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return c.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return c.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return c.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return c.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return c.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return c.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureProp(name string, value interface{}) error {
	props := c.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return c.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return c.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return c.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return c.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return c.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return c.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return c.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return c.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return c.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return c.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return c.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return c.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return c.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return c.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return c.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return c.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (c *ȧutoGeneratedCardDynamicReader) IntProp(name string) (int, error) {

	switch name {
	case "Mark":
		return c.data.Mark, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) SetIntProp(name string, value int) error {

	switch name {
	case "Mark":
		c.data.Mark = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) StringProp(name string) (string, error) {

	switch name {
	case "Tag":
		return c.data.Tag, nil

	}

	return "", errors.New("No such String prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) SetStringProp(name string, value string) error {

	switch name {
	case "Tag":
		c.data.Tag = value
		return nil

	}

	return errors.New("No such String prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "Nest":
		return c.data.Nest, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "Nest":
		c.data.Nest = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "Nest":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "Nest":
		return c.data.Nest, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for cardDynamic
func (c *cardDynamic) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedCardDynamicReader{c}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for cardDynamic
func (c *cardDynamic) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedCardDynamicReader{c}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for cardDynamic
func (c *cardDynamic) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedCardDynamicReader{c}
}

// Implementation for gameState

var ȧutoGeneratedGameStateReaderProps = map[string]boardgame.PropertyType{
	"Both":       boardgame.TypeStack,
	"CapSized":   boardgame.TypeStack,
	"Chips":      boardgame.TypeStack,
	"ChipsColor": boardgame.TypeStack,
	"ChipsValue": boardgame.TypeStack,
	"Combo":      boardgame.TypeStack,
	"Draw":       boardgame.TypeStack,
	"FallMark":   boardgame.TypeStack,
	"MaxTwo":     boardgame.TypeStack,
	"Pool":       boardgame.TypeStack,
	"SameColor":  boardgame.TypeStack,
	"SameMark":   boardgame.TypeStack,
	"SamePips":   boardgame.TypeStack,
	"Spaces":     boardgame.TypeBoard,
	"TwoSuits":   boardgame.TypeStack,
	"UniqueRank": boardgame.TypeStack,
}

type ȧutoGeneratedGameStateReader struct {
	data *gameState
}

func (g *ȧutoGeneratedGameStateReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedGameStateReaderProps
}

func (g *ȧutoGeneratedGameStateReader) Prop(name string) (interface{}, error) {
	props := g.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return g.IntProp(name)
	case boardgame.TypeBool:
		return g.BoolProp(name)
	case boardgame.TypeString:
		return g.StringProp(name)
	case boardgame.TypePlayerIndex:
		return g.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return g.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return g.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return g.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return g.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return g.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return g.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return g.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return g.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return g.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (g *ȧutoGeneratedGameStateReader) PropMutable(name string) bool {
	switch name {
	case "Both":
		return false
	case "CapSized":
		return true
	case "Chips":
		return true
	case "ChipsColor":
		return true
	case "ChipsValue":
		return true
	case "Combo":
		return true
	case "Draw":
		return true
	case "FallMark":
		return true
	case "MaxTwo":
		return true
	case "Pool":
		return true
	case "SameColor":
		return true
	case "SameMark":
		return true
	case "SamePips":
		return true
	case "Spaces":
		return true
	case "TwoSuits":
		return true
	case "UniqueRank":
		return true
	}

	return false
}

func (g *ȧutoGeneratedGameStateReader) SetProp(name string, value interface{}) error {
	props := g.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return g.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return g.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return g.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return g.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return g.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return g.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return g.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return g.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (g *ȧutoGeneratedGameStateReader) ConfigureProp(name string, value interface{}) error {
	props := g.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return g.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return g.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return g.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return g.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if g.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return g.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return g.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return g.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return g.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return g.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return g.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if g.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return g.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return g.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if g.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return g.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return g.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if g.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return g.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return g.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if g.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return g.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return g.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (g *ȧutoGeneratedGameStateReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "Both":
		return g.data.Both, nil
	case "CapSized":
		return g.data.CapSized, nil
	case "Chips":
		return g.data.Chips, nil
	case "ChipsColor":
		return g.data.ChipsColor, nil
	case "ChipsValue":
		return g.data.ChipsValue, nil
	case "Combo":
		return g.data.Combo, nil
	case "Draw":
		return g.data.Draw, nil
	case "FallMark":
		return g.data.FallMark, nil
	case "MaxTwo":
		return g.data.MaxTwo, nil
	case "Pool":
		return g.data.Pool, nil
	case "SameColor":
		return g.data.SameColor, nil
	case "SameMark":
		return g.data.SameMark, nil
	case "SamePips":
		return g.data.SamePips, nil
	case "TwoSuits":
		return g.data.TwoSuits, nil
	case "UniqueRank":
		return g.data.UniqueRank, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "Both":
		return boardgame.ErrPropertyImmutable
	case "CapSized":
		slotValue := value.SizedStack()
		if slotValue == nil {
			return errors.New("CapSized couldn't be upconverted, returned nil")
		}
		g.data.CapSized = slotValue
		return nil
	case "Chips":
		g.data.Chips = value
		return nil
	case "ChipsColor":
		g.data.ChipsColor = value
		return nil
	case "ChipsValue":
		g.data.ChipsValue = value
		return nil
	case "Combo":
		g.data.Combo = value
		return nil
	case "Draw":
		g.data.Draw = value
		return nil
	case "FallMark":
		g.data.FallMark = value
		return nil
	case "MaxTwo":
		g.data.MaxTwo = value
		return nil
	case "Pool":
		g.data.Pool = value
		return nil
	case "SameColor":
		g.data.SameColor = value
		return nil
	case "SameMark":
		g.data.SameMark = value
		return nil
	case "SamePips":
		g.data.SamePips = value
		return nil
	case "TwoSuits":
		g.data.TwoSuits = value
		return nil
	case "UniqueRank":
		slotValue := value.SizedStack()
		if slotValue == nil {
			return errors.New("UniqueRank couldn't be upconverted, returned nil")
		}
		g.data.UniqueRank = slotValue
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "Both":
		slotValue := value.MergedStack()
		if slotValue == nil {
			return errors.New("Both couldn't be upconverted, returned nil")
		}
		g.data.Both = slotValue
		return nil
	case "CapSized":
		return boardgame.ErrPropertyImmutable
	case "Chips":
		return boardgame.ErrPropertyImmutable
	case "ChipsColor":
		return boardgame.ErrPropertyImmutable
	case "ChipsValue":
		return boardgame.ErrPropertyImmutable
	case "Combo":
		return boardgame.ErrPropertyImmutable
	case "Draw":
		return boardgame.ErrPropertyImmutable
	case "FallMark":
		return boardgame.ErrPropertyImmutable
	case "MaxTwo":
		return boardgame.ErrPropertyImmutable
	case "Pool":
		return boardgame.ErrPropertyImmutable
	case "SameColor":
		return boardgame.ErrPropertyImmutable
	case "SameMark":
		return boardgame.ErrPropertyImmutable
	case "SamePips":
		return boardgame.ErrPropertyImmutable
	case "TwoSuits":
		return boardgame.ErrPropertyImmutable
	case "UniqueRank":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "Both":
		return nil, boardgame.ErrPropertyImmutable
	case "CapSized":
		return g.data.CapSized, nil
	case "Chips":
		return g.data.Chips, nil
	case "ChipsColor":
		return g.data.ChipsColor, nil
	case "ChipsValue":
		return g.data.ChipsValue, nil
	case "Combo":
		return g.data.Combo, nil
	case "Draw":
		return g.data.Draw, nil
	case "FallMark":
		return g.data.FallMark, nil
	case "MaxTwo":
		return g.data.MaxTwo, nil
	case "Pool":
		return g.data.Pool, nil
	case "SameColor":
		return g.data.SameColor, nil
	case "SameMark":
		return g.data.SameMark, nil
	case "SamePips":
		return g.data.SamePips, nil
	case "TwoSuits":
		return g.data.TwoSuits, nil
	case "UniqueRank":
		return g.data.UniqueRank, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	switch name {
	case "Spaces":
		return g.data.Spaces, nil

	}

	return nil, errors.New("No such Board prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	switch name {
	case "Spaces":
		g.data.Spaces = value
		return nil

	}

	return errors.New("No such Board prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	switch name {
	case "Spaces":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) BoardProp(name string) (boardgame.Board, error) {

	switch name {
	case "Spaces":
		return g.data.Spaces, nil

	}

	return nil, errors.New("No such Board prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for gameState
func (g *gameState) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedGameStateReader{g}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for gameState
func (g *gameState) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedGameStateReader{g}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for gameState
func (g *gameState) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedGameStateReader{g}
}

// Implementation for playerState

var ȧutoGeneratedPlayerStateReaderProps = map[string]boardgame.PropertyType{
	"Hand":  boardgame.TypeStack,
	"Slots": boardgame.TypeStack,
}

type ȧutoGeneratedPlayerStateReader struct {
	data *playerState
}

func (p *ȧutoGeneratedPlayerStateReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedPlayerStateReaderProps
}

func (p *ȧutoGeneratedPlayerStateReader) Prop(name string) (interface{}, error) {
	props := p.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return p.IntProp(name)
	case boardgame.TypeBool:
		return p.BoolProp(name)
	case boardgame.TypeString:
		return p.StringProp(name)
	case boardgame.TypePlayerIndex:
		return p.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return p.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return p.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return p.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return p.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return p.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return p.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return p.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return p.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return p.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (p *ȧutoGeneratedPlayerStateReader) PropMutable(name string) bool {
	switch name {
	case "Hand":
		return true
	case "Slots":
		return true
	}

	return false
}

func (p *ȧutoGeneratedPlayerStateReader) SetProp(name string, value interface{}) error {
	props := p.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return p.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return p.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return p.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return p.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return p.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return p.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return p.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return p.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureProp(name string, value interface{}) error {
	props := p.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return p.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return p.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return p.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return p.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if p.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return p.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return p.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return p.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return p.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return p.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return p.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if p.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return p.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return p.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if p.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return p.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return p.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if p.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return p.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return p.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if p.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return p.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return p.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (p *ȧutoGeneratedPlayerStateReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "Hand":
		return p.data.Hand, nil
	case "Slots":
		return p.data.Slots, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "Hand":
		p.data.Hand = value
		return nil
	case "Slots":
		slotValue := value.SizedStack()
		if slotValue == nil {
			return errors.New("Slots couldn't be upconverted, returned nil")
		}
		p.data.Slots = slotValue
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "Hand":
		return boardgame.ErrPropertyImmutable
	case "Slots":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "Hand":
		return p.data.Hand, nil
	case "Slots":
		return p.data.Slots, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for playerState
func (p *playerState) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedPlayerStateReader{p}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for playerState
func (p *playerState) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedPlayerStateReader{p}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for playerState
func (p *playerState) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedPlayerStateReader{p}
}

// Implementation for moveNever

var ȧutoGeneratedMoveNeverReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveNeverReader struct {
	data *moveNever
}

func (m *ȧutoGeneratedMoveNeverReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveNeverReaderProps
}

func (m *ȧutoGeneratedMoveNeverReader) Prop(name string) (interface{}, error) {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return m.IntProp(name)
	case boardgame.TypeBool:
		return m.BoolProp(name)
	case boardgame.TypeString:
		return m.StringProp(name)
	case boardgame.TypePlayerIndex:
		return m.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return m.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return m.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return m.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return m.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return m.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return m.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return m.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return m.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return m.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveNeverReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveNeverReader) SetProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveNeverReader) ConfigureProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if m.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return m.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return m.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if m.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return m.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return m.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if m.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return m.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return m.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if m.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return m.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return m.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if m.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return m.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return m.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveNeverReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveNeverReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for moveNever
func (m *moveNever) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveNeverReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveNever
func (m *moveNever) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveNeverReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveNever
func (m *moveNever) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveNeverReader{m}
}

// Implementation for bad1

var ȧutoGeneratedBad1ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad1Reader struct {
	data *bad1
}

func (b *ȧutoGeneratedBad1Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad1ReaderProps
}

func (b *ȧutoGeneratedBad1Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad1Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad1Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad1Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad1Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad1Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad1
func (b *bad1) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad1Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad1
func (b *bad1) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad1Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad1
func (b *bad1) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad1Reader{b}
}

// Implementation for bad2

var ȧutoGeneratedBad2ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad2Reader struct {
	data *bad2
}

func (b *ȧutoGeneratedBad2Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad2ReaderProps
}

func (b *ȧutoGeneratedBad2Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad2Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad2Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad2Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad2Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad2Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad2
func (b *bad2) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad2Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad2
func (b *bad2) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad2Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad2
func (b *bad2) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad2Reader{b}
}

// Implementation for bad3

var ȧutoGeneratedBad3ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad3Reader struct {
	data *bad3
}

func (b *ȧutoGeneratedBad3Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad3ReaderProps
}

func (b *ȧutoGeneratedBad3Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad3Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad3Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad3Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad3Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad3Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad3
func (b *bad3) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad3Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad3
func (b *bad3) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad3Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad3
func (b *bad3) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad3Reader{b}
}

// Implementation for bad4

var ȧutoGeneratedBad4ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad4Reader struct {
	data *bad4
}

func (b *ȧutoGeneratedBad4Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad4ReaderProps
}

func (b *ȧutoGeneratedBad4Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad4Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad4Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad4Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad4Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad4Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad4
func (b *bad4) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad4Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad4
func (b *bad4) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad4Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad4
func (b *bad4) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad4Reader{b}
}

// Implementation for bad5

var ȧutoGeneratedBad5ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad5Reader struct {
	data *bad5
}

func (b *ȧutoGeneratedBad5Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad5ReaderProps
}

func (b *ȧutoGeneratedBad5Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad5Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad5Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad5Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad5Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad5Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad5
func (b *bad5) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad5Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad5
func (b *bad5) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad5Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad5
func (b *bad5) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad5Reader{b}
}

// Implementation for bad6

var ȧutoGeneratedBad6ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad6Reader struct {
	data *bad6
}

func (b *ȧutoGeneratedBad6Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad6ReaderProps
}

func (b *ȧutoGeneratedBad6Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad6Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad6Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad6Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad6Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad6Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad6
func (b *bad6) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad6Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad6
func (b *bad6) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad6Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad6
func (b *bad6) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad6Reader{b}
}

// Implementation for bad7

var ȧutoGeneratedBad7ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad7Reader struct {
	data *bad7
}

func (b *ȧutoGeneratedBad7Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad7ReaderProps
}

func (b *ȧutoGeneratedBad7Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad7Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad7Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad7Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad7Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad7Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad7
func (b *bad7) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad7Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad7
func (b *bad7) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad7Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad7
func (b *bad7) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad7Reader{b}
}

// Implementation for bad8

var ȧutoGeneratedBad8ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad8Reader struct {
	data *bad8
}

func (b *ȧutoGeneratedBad8Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad8ReaderProps
}

func (b *ȧutoGeneratedBad8Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad8Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad8Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad8Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad8Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad8Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad8
func (b *bad8) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad8Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad8
func (b *bad8) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad8Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad8
func (b *bad8) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad8Reader{b}
}

// Implementation for bad9

var ȧutoGeneratedBad9ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad9Reader struct {
	data *bad9
}

func (b *ȧutoGeneratedBad9Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad9ReaderProps
}

func (b *ȧutoGeneratedBad9Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad9Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad9Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad9Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad9Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad9Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad9
func (b *bad9) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad9Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad9
func (b *bad9) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad9Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad9
func (b *bad9) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad9Reader{b}
}

// Implementation for bad10

var ȧutoGeneratedBad10ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad10Reader struct {
	data *bad10
}

func (b *ȧutoGeneratedBad10Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad10ReaderProps
}

func (b *ȧutoGeneratedBad10Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad10Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad10Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad10Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad10Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad10Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad10
func (b *bad10) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad10Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad10
func (b *bad10) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad10Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad10
func (b *bad10) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad10Reader{b}
}

// Implementation for bad11

var ȧutoGeneratedBad11ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad11Reader struct {
	data *bad11
}

func (b *ȧutoGeneratedBad11Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad11ReaderProps
}

func (b *ȧutoGeneratedBad11Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad11Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad11Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad11Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad11Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad11Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad11
func (b *bad11) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad11Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad11
func (b *bad11) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad11Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad11
func (b *bad11) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad11Reader{b}
}

// Implementation for bad12

var ȧutoGeneratedBad12ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad12Reader struct {
	data *bad12
}

func (b *ȧutoGeneratedBad12Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad12ReaderProps
}

func (b *ȧutoGeneratedBad12Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad12Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad12Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad12Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad12Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad12Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad12
func (b *bad12) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad12Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad12
func (b *bad12) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad12Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad12
func (b *bad12) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad12Reader{b}
}

// Implementation for bad13

var ȧutoGeneratedBad13ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad13Reader struct {
	data *bad13
}

func (b *ȧutoGeneratedBad13Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad13ReaderProps
}

func (b *ȧutoGeneratedBad13Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad13Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad13Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad13Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad13Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad13Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad13
func (b *bad13) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad13Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad13
func (b *bad13) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad13Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad13
func (b *bad13) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad13Reader{b}
}

// Implementation for bad14

var ȧutoGeneratedBad14ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad14Reader struct {
	data *bad14
}

func (b *ȧutoGeneratedBad14Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad14ReaderProps
}

func (b *ȧutoGeneratedBad14Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad14Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad14Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad14Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad14Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad14Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad14
func (b *bad14) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad14Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad14
func (b *bad14) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad14Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad14
func (b *bad14) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad14Reader{b}
}

// Implementation for bad15

var ȧutoGeneratedBad15ReaderProps = map[string]boardgame.PropertyType{
	"S": boardgame.TypeStack,
}

type ȧutoGeneratedBad15Reader struct {
	data *bad15
}

func (b *ȧutoGeneratedBad15Reader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedBad15ReaderProps
}

func (b *ȧutoGeneratedBad15Reader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return b.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad15Reader) PropMutable(name string) bool {
	switch name {
	case "S":
		return true
	}

	return false
}

func (b *ȧutoGeneratedBad15Reader) SetProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeEnumSlice:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad15Reader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return b.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return b.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return b.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return b.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return b.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return b.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return b.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *ȧutoGeneratedBad15Reader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "S":
		b.data.S = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "S":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "S":
		return b.data.S, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (b *ȧutoGeneratedBad15Reader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for bad15
func (b *bad15) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedBad15Reader{b}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for bad15
func (b *bad15) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedBad15Reader{b}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for bad15
func (b *bad15) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedBad15Reader{b}
}
