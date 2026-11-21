/************************************
 *
 * This file contains auto-generated methods to help certain structs
 * implement boardgame.PropertyReader and friends. It was generated
 * by the codegen package via 'boardgame-util codegen'.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package vecgame

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// Implementation for cardValue

var ȧutoGeneratedCardValueReaderProps = map[string]boardgame.PropertyType{
	"Rank": boardgame.TypeInt,
	"Suit": boardgame.TypeString,
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

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) IntSliceProp(name string) ([]int, error) {

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

// Implementation for tokenValue

var ȧutoGeneratedTokenValueReaderProps = map[string]boardgame.PropertyType{
	"Kind": boardgame.TypeString,
}

type ȧutoGeneratedTokenValueReader struct {
	data *tokenValue
}

func (t *ȧutoGeneratedTokenValueReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedTokenValueReaderProps
}

func (t *ȧutoGeneratedTokenValueReader) Prop(name string) (interface{}, error) {
	props := t.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return t.IntProp(name)
	case boardgame.TypeBool:
		return t.BoolProp(name)
	case boardgame.TypeString:
		return t.StringProp(name)
	case boardgame.TypePlayerIndex:
		return t.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return t.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return t.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return t.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return t.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return t.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return t.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return t.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return t.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return t.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (t *ȧutoGeneratedTokenValueReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (t *ȧutoGeneratedTokenValueReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (t *ȧutoGeneratedTokenValueReader) StringProp(name string) (string, error) {

	switch name {
	case "Kind":
		return t.data.Kind, nil

	}

	return "", errors.New("No such String prop: " + name)

}

func (t *ȧutoGeneratedTokenValueReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (t *ȧutoGeneratedTokenValueReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (t *ȧutoGeneratedTokenValueReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (t *ȧutoGeneratedTokenValueReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (t *ȧutoGeneratedTokenValueReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (t *ȧutoGeneratedTokenValueReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (t *ȧutoGeneratedTokenValueReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (t *ȧutoGeneratedTokenValueReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (t *ȧutoGeneratedTokenValueReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (t *ȧutoGeneratedTokenValueReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for tokenValue
func (t *tokenValue) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedTokenValueReader{t}
}

// Implementation for cardDynamic

var ȧutoGeneratedCardDynamicReaderProps = map[string]boardgame.PropertyType{
	"Counter": boardgame.TypeInt,
	"Held":    boardgame.TypeStack,
	"Mark":    boardgame.TypeEnum,
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
	case "Counter":
		return true
	case "Held":
		return true
	case "Mark":
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
	case "Counter":
		return c.data.Counter, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) SetIntProp(name string, value int) error {

	switch name {
	case "Counter":
		c.data.Counter = value
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

	return "", errors.New("No such String prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "Mark":
		return c.data.Mark, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "Mark":
		c.data.Mark = value
		return nil

	}

	return errors.New("No such Enum prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "Mark":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "Mark":
		return c.data.Mark, nil

	}

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
	case "Held":
		return c.data.Held, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "Held":
		slotValue := value.SizedStack()
		if slotValue == nil {
			return errors.New("Held couldn't be upconverted, returned nil")
		}
		c.data.Held = slotValue
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "Held":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (c *ȧutoGeneratedCardDynamicReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "Held":
		return c.data.Held, nil

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
	"All":         boardgame.TypeStack,
	"Bools":       boardgame.TypeBoolSlice,
	"Clock":       boardgame.TypeTimer,
	"Color":       boardgame.TypeEnum,
	"Colors":      boardgame.TypeEnumSlice,
	"Count":       boardgame.TypeInt,
	"Draw":        boardgame.TypeStack,
	"Flag":        boardgame.TypeBool,
	"Ints":        boardgame.TypeIntSlice,
	"Limited":     boardgame.TypeStack,
	"Name":        boardgame.TypeString,
	"Overlap":     boardgame.TypeStack,
	"Row":         boardgame.TypeStack,
	"Spaces":      boardgame.TypeBoard,
	"Strings":     boardgame.TypeStringSlice,
	"TagHidden":   boardgame.TypeInt,
	"TagLen":      boardgame.TypeStack,
	"TagNonEmpty": boardgame.TypeStack,
	"TagOrder":    boardgame.TypeStack,
	"TagRedOnly":  boardgame.TypeString,
	"TagSpaced":   boardgame.TypeInt,
	"Tokens":      boardgame.TypeStack,
	"Under":       boardgame.TypeStack,
	"Who":         boardgame.TypePlayerIndex,
	"Whos":        boardgame.TypePlayerIndexSlice,
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
	case "All":
		return false
	case "Bools":
		return true
	case "Clock":
		return true
	case "Color":
		return true
	case "Colors":
		return true
	case "Count":
		return true
	case "Draw":
		return true
	case "Flag":
		return true
	case "Ints":
		return true
	case "Limited":
		return true
	case "Name":
		return true
	case "Overlap":
		return false
	case "Row":
		return true
	case "Spaces":
		return true
	case "Strings":
		return true
	case "TagHidden":
		return true
	case "TagLen":
		return true
	case "TagNonEmpty":
		return true
	case "TagOrder":
		return true
	case "TagRedOnly":
		return true
	case "TagSpaced":
		return true
	case "Tokens":
		return true
	case "Under":
		return true
	case "Who":
		return true
	case "Whos":
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

	switch name {
	case "Count":
		return g.data.Count, nil
	case "TagHidden":
		return g.data.TagHidden, nil
	case "TagSpaced":
		return g.data.TagSpaced, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetIntProp(name string, value int) error {

	switch name {
	case "Count":
		g.data.Count = value
		return nil
	case "TagHidden":
		g.data.TagHidden = value
		return nil
	case "TagSpaced":
		g.data.TagSpaced = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) BoolProp(name string) (bool, error) {

	switch name {
	case "Flag":
		return g.data.Flag, nil

	}

	return false, errors.New("No such Bool prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetBoolProp(name string, value bool) error {

	switch name {
	case "Flag":
		g.data.Flag = value
		return nil

	}

	return errors.New("No such Bool prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) StringProp(name string) (string, error) {

	switch name {
	case "Name":
		return g.data.Name, nil
	case "TagRedOnly":
		return g.data.TagRedOnly, nil

	}

	return "", errors.New("No such String prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetStringProp(name string, value string) error {

	switch name {
	case "Name":
		g.data.Name = value
		return nil
	case "TagRedOnly":
		g.data.TagRedOnly = value
		return nil

	}

	return errors.New("No such String prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "Who":
		return g.data.Who, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "Who":
		g.data.Who = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "Color":
		return g.data.Color, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "Color":
		g.data.Color = value
		return nil

	}

	return errors.New("No such Enum prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "Color":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "Color":
		return g.data.Color, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) IntSliceProp(name string) ([]int, error) {

	switch name {
	case "Ints":
		return g.data.Ints, nil

	}

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetIntSliceProp(name string, value []int) error {

	switch name {
	case "Ints":
		g.data.Ints = value
		return nil

	}

	return errors.New("No such IntSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) BoolSliceProp(name string) ([]bool, error) {

	switch name {
	case "Bools":
		return g.data.Bools, nil

	}

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetBoolSliceProp(name string, value []bool) error {

	switch name {
	case "Bools":
		g.data.Bools = value
		return nil

	}

	return errors.New("No such BoolSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) StringSliceProp(name string) ([]string, error) {

	switch name {
	case "Strings":
		return g.data.Strings, nil

	}

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetStringSliceProp(name string, value []string) error {

	switch name {
	case "Strings":
		g.data.Strings = value
		return nil

	}

	return errors.New("No such StringSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	switch name {
	case "Whos":
		return g.data.Whos, nil

	}

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	switch name {
	case "Whos":
		g.data.Whos = value
		return nil

	}

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "All":
		return g.data.All, nil
	case "Draw":
		return g.data.Draw, nil
	case "Limited":
		return g.data.Limited, nil
	case "Overlap":
		return g.data.Overlap, nil
	case "Row":
		return g.data.Row, nil
	case "TagLen":
		return g.data.TagLen, nil
	case "TagNonEmpty":
		return g.data.TagNonEmpty, nil
	case "TagOrder":
		return g.data.TagOrder, nil
	case "Tokens":
		return g.data.Tokens, nil
	case "Under":
		return g.data.Under, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "All":
		return boardgame.ErrPropertyImmutable
	case "Draw":
		g.data.Draw = value
		return nil
	case "Limited":
		g.data.Limited = value
		return nil
	case "Overlap":
		return boardgame.ErrPropertyImmutable
	case "Row":
		slotValue := value.SizedStack()
		if slotValue == nil {
			return errors.New("Row couldn't be upconverted, returned nil")
		}
		g.data.Row = slotValue
		return nil
	case "TagLen":
		g.data.TagLen = value
		return nil
	case "TagNonEmpty":
		g.data.TagNonEmpty = value
		return nil
	case "TagOrder":
		g.data.TagOrder = value
		return nil
	case "Tokens":
		g.data.Tokens = value
		return nil
	case "Under":
		slotValue := value.SizedStack()
		if slotValue == nil {
			return errors.New("Under couldn't be upconverted, returned nil")
		}
		g.data.Under = slotValue
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "All":
		slotValue := value.MergedStack()
		if slotValue == nil {
			return errors.New("All couldn't be upconverted, returned nil")
		}
		g.data.All = slotValue
		return nil
	case "Draw":
		return boardgame.ErrPropertyImmutable
	case "Limited":
		return boardgame.ErrPropertyImmutable
	case "Overlap":
		slotValue := value.MergedStack()
		if slotValue == nil {
			return errors.New("Overlap couldn't be upconverted, returned nil")
		}
		g.data.Overlap = slotValue
		return nil
	case "Row":
		return boardgame.ErrPropertyImmutable
	case "TagLen":
		return boardgame.ErrPropertyImmutable
	case "TagNonEmpty":
		return boardgame.ErrPropertyImmutable
	case "TagOrder":
		return boardgame.ErrPropertyImmutable
	case "Tokens":
		return boardgame.ErrPropertyImmutable
	case "Under":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "All":
		return nil, boardgame.ErrPropertyImmutable
	case "Draw":
		return g.data.Draw, nil
	case "Limited":
		return g.data.Limited, nil
	case "Overlap":
		return nil, boardgame.ErrPropertyImmutable
	case "Row":
		return g.data.Row, nil
	case "TagLen":
		return g.data.TagLen, nil
	case "TagNonEmpty":
		return g.data.TagNonEmpty, nil
	case "TagOrder":
		return g.data.TagOrder, nil
	case "Tokens":
		return g.data.Tokens, nil
	case "Under":
		return g.data.Under, nil

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

	switch name {
	case "Clock":
		return g.data.Clock, nil

	}

	return nil, errors.New("No such Timer prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	switch name {
	case "Clock":
		g.data.Clock = value
		return nil

	}

	return errors.New("No such Timer prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	switch name {
	case "Clock":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) TimerProp(name string) (boardgame.Timer, error) {

	switch name {
	case "Clock":
		return g.data.Clock, nil

	}

	return nil, errors.New("No such Timer prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	switch name {
	case "Colors":
		return g.data.Colors, nil

	}

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	switch name {
	case "Colors":
		g.data.Colors = value
		return nil

	}

	return errors.New("No such EnumSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	switch name {
	case "Colors":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	switch name {
	case "Colors":
		return g.data.Colors, nil

	}

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
	"Color":      boardgame.TypeEnum,
	"Dyn":        boardgame.TypeInt,
	"Hand":       boardgame.TypeStack,
	"OwnOnly":    boardgame.TypeIntSlice,
	"RivalNote":  boardgame.TypeString,
	"Score":      boardgame.TypeInt,
	"Secret":     boardgame.TypeInt,
	"Slots":      boardgame.TypeStack,
	"TeamSecret": boardgame.TypeString,
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
	case "Color":
		return true
	case "Dyn":
		return true
	case "Hand":
		return true
	case "OwnOnly":
		return true
	case "RivalNote":
		return true
	case "Score":
		return true
	case "Secret":
		return true
	case "Slots":
		return true
	case "TeamSecret":
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

	switch name {
	case "Dyn":
		return p.data.Dyn, nil
	case "Score":
		return p.data.Score, nil
	case "Secret":
		return p.data.Secret, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetIntProp(name string, value int) error {

	switch name {
	case "Dyn":
		p.data.Dyn = value
		return nil
	case "Score":
		p.data.Score = value
		return nil
	case "Secret":
		p.data.Secret = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) StringProp(name string) (string, error) {

	switch name {
	case "RivalNote":
		return p.data.RivalNote, nil
	case "TeamSecret":
		return p.data.TeamSecret, nil

	}

	return "", errors.New("No such String prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetStringProp(name string, value string) error {

	switch name {
	case "RivalNote":
		p.data.RivalNote = value
		return nil
	case "TeamSecret":
		p.data.TeamSecret = value
		return nil

	}

	return errors.New("No such String prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "Color":
		return p.data.Color, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "Color":
		p.data.Color = value
		return nil

	}

	return errors.New("No such Enum prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "Color":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "Color":
		return p.data.Color, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) IntSliceProp(name string) ([]int, error) {

	switch name {
	case "OwnOnly":
		return p.data.OwnOnly, nil

	}

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetIntSliceProp(name string, value []int) error {

	switch name {
	case "OwnOnly":
		p.data.OwnOnly = value
		return nil

	}

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
