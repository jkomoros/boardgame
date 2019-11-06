/************************************
 *
 * This file contains auto-generated methods to help certain structs
 * implement boardgame.PropertyReader and friends. It was generated
 * by the codegen package via 'boardgame-util codegen'.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package debuganimations

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// Implementation for cardValue

var ȧutoGeneratedCardValueReaderProps = map[string]boardgame.PropertyType{
	"Type": boardgame.TypeString,
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

func (c *ȧutoGeneratedCardValueReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) StringProp(name string) (string, error) {

	switch name {
	case "Type":
		return c.data.Type, nil

	}

	return "", errors.New("No such String prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for cardValue
func (c *cardValue) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedCardValueReader{c}
}

// Implementation for moveMoveCardBetweenShortStacks

var ȧutoGeneratedMoveMoveCardBetweenShortStacksReaderProps = map[string]boardgame.PropertyType{
	"FromFirst": boardgame.TypeBool,
}

type ȧutoGeneratedMoveMoveCardBetweenShortStacksReader struct {
	data *moveMoveCardBetweenShortStacks
}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveMoveCardBetweenShortStacksReaderProps
}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) Prop(name string) (interface{}, error) {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return m.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return m.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return m.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return m.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return m.IntProp(name)
	case boardgame.TypeIntSlice:
		return m.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return m.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return m.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return m.ImmutableStackProp(name)
	case boardgame.TypeString:
		return m.StringProp(name)
	case boardgame.TypeStringSlice:
		return m.StringSliceProp(name)
	case boardgame.TypeTimer:
		return m.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) PropMutable(name string) bool {
	switch name {
	case "FromFirst":
		return true
	}

	return false
}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) SetProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) ConfigureProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
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
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
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
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
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
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) BoolProp(name string) (bool, error) {

	switch name {
	case "FromFirst":
		return m.data.FromFirst, nil

	}

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) SetBoolProp(name string, value bool) error {

	switch name {
	case "FromFirst":
		m.data.FromFirst = value
		return nil

	}

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenShortStacksReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for moveMoveCardBetweenShortStacks
func (m *moveMoveCardBetweenShortStacks) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveMoveCardBetweenShortStacksReader{m}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveMoveCardBetweenShortStacks
func (m *moveMoveCardBetweenShortStacks) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveMoveCardBetweenShortStacksReader{m}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveMoveCardBetweenShortStacks
func (m *moveMoveCardBetweenShortStacks) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveMoveCardBetweenShortStacksReader{m}
}

// Implementation for moveMoveCardBetweenDrawAndDiscardStacks

var ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReaderProps = map[string]boardgame.PropertyType{
	"FromDraw": boardgame.TypeBool,
}

type ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader struct {
	data *moveMoveCardBetweenDrawAndDiscardStacks
}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReaderProps
}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) Prop(name string) (interface{}, error) {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return m.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return m.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return m.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return m.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return m.IntProp(name)
	case boardgame.TypeIntSlice:
		return m.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return m.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return m.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return m.ImmutableStackProp(name)
	case boardgame.TypeString:
		return m.StringProp(name)
	case boardgame.TypeStringSlice:
		return m.StringSliceProp(name)
	case boardgame.TypeTimer:
		return m.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) PropMutable(name string) bool {
	switch name {
	case "FromDraw":
		return true
	}

	return false
}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) SetProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) ConfigureProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
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
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
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
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
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
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) BoolProp(name string) (bool, error) {

	switch name {
	case "FromDraw":
		return m.data.FromDraw, nil

	}

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) SetBoolProp(name string, value bool) error {

	switch name {
	case "FromDraw":
		m.data.FromDraw = value
		return nil

	}

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for moveMoveCardBetweenDrawAndDiscardStacks
func (m *moveMoveCardBetweenDrawAndDiscardStacks) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader{m}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveMoveCardBetweenDrawAndDiscardStacks
func (m *moveMoveCardBetweenDrawAndDiscardStacks) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader{m}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveMoveCardBetweenDrawAndDiscardStacks
func (m *moveMoveCardBetweenDrawAndDiscardStacks) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveMoveCardBetweenDrawAndDiscardStacksReader{m}
}

// Implementation for moveFlipHiddenCard

var ȧutoGeneratedMoveFlipHiddenCardReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveFlipHiddenCardReader struct {
	data *moveFlipHiddenCard
}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveFlipHiddenCardReaderProps
}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) Prop(name string) (interface{}, error) {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return m.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return m.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return m.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return m.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return m.IntProp(name)
	case boardgame.TypeIntSlice:
		return m.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return m.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return m.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return m.ImmutableStackProp(name)
	case boardgame.TypeString:
		return m.StringProp(name)
	case boardgame.TypeStringSlice:
		return m.StringSliceProp(name)
	case boardgame.TypeTimer:
		return m.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) SetProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) ConfigureProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
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
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
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
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
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
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveFlipHiddenCardReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for moveFlipHiddenCard
func (m *moveFlipHiddenCard) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveFlipHiddenCardReader{m}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveFlipHiddenCard
func (m *moveFlipHiddenCard) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveFlipHiddenCardReader{m}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveFlipHiddenCard
func (m *moveFlipHiddenCard) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveFlipHiddenCardReader{m}
}

// Implementation for moveMoveCardBetweenFanStacks

var ȧutoGeneratedMoveMoveCardBetweenFanStacksReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveMoveCardBetweenFanStacksReader struct {
	data *moveMoveCardBetweenFanStacks
}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveMoveCardBetweenFanStacksReaderProps
}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) Prop(name string) (interface{}, error) {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return m.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return m.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return m.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return m.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return m.IntProp(name)
	case boardgame.TypeIntSlice:
		return m.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return m.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return m.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return m.ImmutableStackProp(name)
	case boardgame.TypeString:
		return m.StringProp(name)
	case boardgame.TypeStringSlice:
		return m.StringSliceProp(name)
	case boardgame.TypeTimer:
		return m.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) SetProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) ConfigureProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
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
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
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
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
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
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveCardBetweenFanStacksReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for moveMoveCardBetweenFanStacks
func (m *moveMoveCardBetweenFanStacks) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveMoveCardBetweenFanStacksReader{m}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveMoveCardBetweenFanStacks
func (m *moveMoveCardBetweenFanStacks) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveMoveCardBetweenFanStacksReader{m}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveMoveCardBetweenFanStacks
func (m *moveMoveCardBetweenFanStacks) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveMoveCardBetweenFanStacksReader{m}
}

// Implementation for moveVisibleShuffleCards

var ȧutoGeneratedMoveVisibleShuffleCardsReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveVisibleShuffleCardsReader struct {
	data *moveVisibleShuffleCards
}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveVisibleShuffleCardsReaderProps
}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) Prop(name string) (interface{}, error) {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return m.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return m.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return m.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return m.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return m.IntProp(name)
	case boardgame.TypeIntSlice:
		return m.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return m.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return m.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return m.ImmutableStackProp(name)
	case boardgame.TypeString:
		return m.StringProp(name)
	case boardgame.TypeStringSlice:
		return m.StringSliceProp(name)
	case boardgame.TypeTimer:
		return m.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) SetProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) ConfigureProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
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
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
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
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
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
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveVisibleShuffleCardsReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for moveVisibleShuffleCards
func (m *moveVisibleShuffleCards) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveVisibleShuffleCardsReader{m}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveVisibleShuffleCards
func (m *moveVisibleShuffleCards) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveVisibleShuffleCardsReader{m}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveVisibleShuffleCards
func (m *moveVisibleShuffleCards) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveVisibleShuffleCardsReader{m}
}

// Implementation for moveShuffleCards

var ȧutoGeneratedMoveShuffleCardsReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveShuffleCardsReader struct {
	data *moveShuffleCards
}

func (m *ȧutoGeneratedMoveShuffleCardsReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveShuffleCardsReaderProps
}

func (m *ȧutoGeneratedMoveShuffleCardsReader) Prop(name string) (interface{}, error) {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return m.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return m.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return m.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return m.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return m.IntProp(name)
	case boardgame.TypeIntSlice:
		return m.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return m.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return m.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return m.ImmutableStackProp(name)
	case boardgame.TypeString:
		return m.StringProp(name)
	case boardgame.TypeStringSlice:
		return m.StringSliceProp(name)
	case boardgame.TypeTimer:
		return m.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveShuffleCardsReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveShuffleCardsReader) SetProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveShuffleCardsReader) ConfigureProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
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
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
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
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
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
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveShuffleCardsReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveShuffleCardsReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for moveShuffleCards
func (m *moveShuffleCards) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveShuffleCardsReader{m}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveShuffleCards
func (m *moveShuffleCards) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveShuffleCardsReader{m}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveShuffleCards
func (m *moveShuffleCards) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveShuffleCardsReader{m}
}

// Implementation for moveMoveBetweenHidden

var ȧutoGeneratedMoveMoveBetweenHiddenReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveMoveBetweenHiddenReader struct {
	data *moveMoveBetweenHidden
}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveMoveBetweenHiddenReaderProps
}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) Prop(name string) (interface{}, error) {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return m.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return m.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return m.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return m.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return m.IntProp(name)
	case boardgame.TypeIntSlice:
		return m.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return m.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return m.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return m.ImmutableStackProp(name)
	case boardgame.TypeString:
		return m.StringProp(name)
	case boardgame.TypeStringSlice:
		return m.StringSliceProp(name)
	case boardgame.TypeTimer:
		return m.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) SetProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) ConfigureProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
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
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
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
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
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
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveBetweenHiddenReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for moveMoveBetweenHidden
func (m *moveMoveBetweenHidden) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveMoveBetweenHiddenReader{m}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveMoveBetweenHidden
func (m *moveMoveBetweenHidden) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveMoveBetweenHiddenReader{m}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveMoveBetweenHidden
func (m *moveMoveBetweenHidden) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveMoveBetweenHiddenReader{m}
}

// Implementation for moveMoveToken

var ȧutoGeneratedMoveMoveTokenReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveMoveTokenReader struct {
	data *moveMoveToken
}

func (m *ȧutoGeneratedMoveMoveTokenReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveMoveTokenReaderProps
}

func (m *ȧutoGeneratedMoveMoveTokenReader) Prop(name string) (interface{}, error) {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return m.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return m.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return m.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return m.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return m.IntProp(name)
	case boardgame.TypeIntSlice:
		return m.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return m.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return m.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return m.ImmutableStackProp(name)
	case boardgame.TypeString:
		return m.StringProp(name)
	case boardgame.TypeStringSlice:
		return m.StringSliceProp(name)
	case boardgame.TypeTimer:
		return m.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveTokenReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveMoveTokenReader) SetProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveTokenReader) ConfigureProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
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
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
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
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
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
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveTokenReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for moveMoveToken
func (m *moveMoveToken) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveMoveTokenReader{m}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveMoveToken
func (m *moveMoveToken) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveMoveTokenReader{m}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveMoveToken
func (m *moveMoveToken) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveMoveTokenReader{m}
}

// Implementation for moveMoveTokenSanitized

var ȧutoGeneratedMoveMoveTokenSanitizedReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveMoveTokenSanitizedReader struct {
	data *moveMoveTokenSanitized
}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveMoveTokenSanitizedReaderProps
}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) Prop(name string) (interface{}, error) {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return m.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return m.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return m.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return m.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return m.IntProp(name)
	case boardgame.TypeIntSlice:
		return m.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return m.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return m.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return m.ImmutableStackProp(name)
	case boardgame.TypeString:
		return m.StringProp(name)
	case boardgame.TypeStringSlice:
		return m.StringSliceProp(name)
	case boardgame.TypeTimer:
		return m.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) SetProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) ConfigureProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
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
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
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
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
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
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveMoveTokenSanitizedReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for moveMoveTokenSanitized
func (m *moveMoveTokenSanitized) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveMoveTokenSanitizedReader{m}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveMoveTokenSanitized
func (m *moveMoveTokenSanitized) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveMoveTokenSanitizedReader{m}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveMoveTokenSanitized
func (m *moveMoveTokenSanitized) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveMoveTokenSanitizedReader{m}
}

// Implementation for moveStartMoveAllComponentsToHidden

var ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader struct {
	data *moveStartMoveAllComponentsToHidden
}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReaderProps
}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) Prop(name string) (interface{}, error) {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return m.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return m.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return m.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return m.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return m.IntProp(name)
	case boardgame.TypeIntSlice:
		return m.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return m.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return m.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return m.ImmutableStackProp(name)
	case boardgame.TypeString:
		return m.StringProp(name)
	case boardgame.TypeStringSlice:
		return m.StringSliceProp(name)
	case boardgame.TypeTimer:
		return m.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) SetProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) ConfigureProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
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
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
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
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
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
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for moveStartMoveAllComponentsToHidden
func (m *moveStartMoveAllComponentsToHidden) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader{m}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveStartMoveAllComponentsToHidden
func (m *moveStartMoveAllComponentsToHidden) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader{m}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveStartMoveAllComponentsToHidden
func (m *moveStartMoveAllComponentsToHidden) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveStartMoveAllComponentsToHiddenReader{m}
}

// Implementation for moveStartMoveAllComponentsToVisible

var ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader struct {
	data *moveStartMoveAllComponentsToVisible
}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReaderProps
}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) Prop(name string) (interface{}, error) {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return m.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return m.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return m.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return m.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return m.IntProp(name)
	case boardgame.TypeIntSlice:
		return m.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return m.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return m.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return m.ImmutableStackProp(name)
	case boardgame.TypeString:
		return m.StringProp(name)
	case boardgame.TypeStringSlice:
		return m.StringSliceProp(name)
	case boardgame.TypeTimer:
		return m.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) SetProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) ConfigureProp(name string, value interface{}) error {
	props := m.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
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
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
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
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return m.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return m.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return m.SetPlayerIndexProp(name, val)
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
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return m.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return m.SetStringSliceProp(name, val)
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for moveStartMoveAllComponentsToVisible
func (m *moveStartMoveAllComponentsToVisible) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader{m}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveStartMoveAllComponentsToVisible
func (m *moveStartMoveAllComponentsToVisible) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader{m}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveStartMoveAllComponentsToVisible
func (m *moveStartMoveAllComponentsToVisible) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveStartMoveAllComponentsToVisibleReader{m}
}

// Implementation for gameState

var ȧutoGeneratedGameStateReaderProps = map[string]boardgame.PropertyType{
	"AllHiddenStack":      boardgame.TypeStack,
	"AllVisibleStack":     boardgame.TypeStack,
	"Card":                boardgame.TypeStack,
	"CurrentPlayer":       boardgame.TypePlayerIndex,
	"DiscardStack":        boardgame.TypeStack,
	"DrawStack":           boardgame.TypeStack,
	"FanDiscard":          boardgame.TypeStack,
	"FanShuffleCount":     boardgame.TypeInt,
	"FanStack":            boardgame.TypeStack,
	"FirstShortStack":     boardgame.TypeStack,
	"HiddenCard":          boardgame.TypeStack,
	"HiddenStack":         boardgame.TypeStack,
	"Phase":               boardgame.TypeEnum,
	"SanitizedTokensFrom": boardgame.TypeStack,
	"SanitizedTokensTo":   boardgame.TypeStack,
	"SecondShortStack":    boardgame.TypeStack,
	"TokensFrom":          boardgame.TypeStack,
	"TokensTo":            boardgame.TypeStack,
	"VisibleCard":         boardgame.TypeStack,
	"VisibleStack":        boardgame.TypeStack,
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
	case boardgame.TypeBoard:
		return g.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return g.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return g.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return g.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return g.IntProp(name)
	case boardgame.TypeIntSlice:
		return g.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return g.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return g.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return g.ImmutableStackProp(name)
	case boardgame.TypeString:
		return g.StringProp(name)
	case boardgame.TypeStringSlice:
		return g.StringSliceProp(name)
	case boardgame.TypeTimer:
		return g.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (g *ȧutoGeneratedGameStateReader) PropMutable(name string) bool {
	switch name {
	case "AllHiddenStack":
		return true
	case "AllVisibleStack":
		return true
	case "Card":
		return false
	case "CurrentPlayer":
		return true
	case "DiscardStack":
		return true
	case "DrawStack":
		return true
	case "FanDiscard":
		return true
	case "FanShuffleCount":
		return true
	case "FanStack":
		return true
	case "FirstShortStack":
		return true
	case "HiddenCard":
		return true
	case "HiddenStack":
		return true
	case "Phase":
		return true
	case "SanitizedTokensFrom":
		return true
	case "SanitizedTokensTo":
		return true
	case "SecondShortStack":
		return true
	case "TokensFrom":
		return true
	case "TokensTo":
		return true
	case "VisibleCard":
		return true
	case "VisibleStack":
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
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return g.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return g.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return g.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return g.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return g.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return g.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return g.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return g.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
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
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return g.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return g.SetBoolSliceProp(name, val)
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
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return g.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return g.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return g.SetPlayerIndexProp(name, val)
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
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return g.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return g.SetStringSliceProp(name, val)
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (g *ȧutoGeneratedGameStateReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "Phase":
		return g.data.Phase, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "Phase":
		g.data.Phase = value
		return nil

	}

	return errors.New("No such Enum prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "Phase":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "Phase":
		return g.data.Phase, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) IntProp(name string) (int, error) {

	switch name {
	case "FanShuffleCount":
		return g.data.FanShuffleCount, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetIntProp(name string, value int) error {

	switch name {
	case "FanShuffleCount":
		g.data.FanShuffleCount = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "CurrentPlayer":
		return g.data.CurrentPlayer, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "CurrentPlayer":
		g.data.CurrentPlayer = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "AllHiddenStack":
		return g.data.AllHiddenStack, nil
	case "AllVisibleStack":
		return g.data.AllVisibleStack, nil
	case "Card":
		return g.data.Card, nil
	case "DiscardStack":
		return g.data.DiscardStack, nil
	case "DrawStack":
		return g.data.DrawStack, nil
	case "FanDiscard":
		return g.data.FanDiscard, nil
	case "FanStack":
		return g.data.FanStack, nil
	case "FirstShortStack":
		return g.data.FirstShortStack, nil
	case "HiddenCard":
		return g.data.HiddenCard, nil
	case "HiddenStack":
		return g.data.HiddenStack, nil
	case "SanitizedTokensFrom":
		return g.data.SanitizedTokensFrom, nil
	case "SanitizedTokensTo":
		return g.data.SanitizedTokensTo, nil
	case "SecondShortStack":
		return g.data.SecondShortStack, nil
	case "TokensFrom":
		return g.data.TokensFrom, nil
	case "TokensTo":
		return g.data.TokensTo, nil
	case "VisibleCard":
		return g.data.VisibleCard, nil
	case "VisibleStack":
		return g.data.VisibleStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "AllHiddenStack":
		g.data.AllHiddenStack = value
		return nil
	case "AllVisibleStack":
		g.data.AllVisibleStack = value
		return nil
	case "Card":
		return boardgame.ErrPropertyImmutable
	case "DiscardStack":
		g.data.DiscardStack = value
		return nil
	case "DrawStack":
		g.data.DrawStack = value
		return nil
	case "FanDiscard":
		g.data.FanDiscard = value
		return nil
	case "FanStack":
		g.data.FanStack = value
		return nil
	case "FirstShortStack":
		g.data.FirstShortStack = value
		return nil
	case "HiddenCard":
		slotValue := value.SizedStack()
		if slotValue == nil {
			return errors.New("HiddenCard couldn't be upconverted, returned nil")
		}
		g.data.HiddenCard = slotValue
		return nil
	case "HiddenStack":
		g.data.HiddenStack = value
		return nil
	case "SanitizedTokensFrom":
		g.data.SanitizedTokensFrom = value
		return nil
	case "SanitizedTokensTo":
		g.data.SanitizedTokensTo = value
		return nil
	case "SecondShortStack":
		g.data.SecondShortStack = value
		return nil
	case "TokensFrom":
		g.data.TokensFrom = value
		return nil
	case "TokensTo":
		g.data.TokensTo = value
		return nil
	case "VisibleCard":
		slotValue := value.SizedStack()
		if slotValue == nil {
			return errors.New("VisibleCard couldn't be upconverted, returned nil")
		}
		g.data.VisibleCard = slotValue
		return nil
	case "VisibleStack":
		g.data.VisibleStack = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "AllHiddenStack":
		return boardgame.ErrPropertyImmutable
	case "AllVisibleStack":
		return boardgame.ErrPropertyImmutable
	case "Card":
		slotValue := value.MergedStack()
		if slotValue == nil {
			return errors.New("Card couldn't be upconverted, returned nil")
		}
		g.data.Card = slotValue
		return nil
	case "DiscardStack":
		return boardgame.ErrPropertyImmutable
	case "DrawStack":
		return boardgame.ErrPropertyImmutable
	case "FanDiscard":
		return boardgame.ErrPropertyImmutable
	case "FanStack":
		return boardgame.ErrPropertyImmutable
	case "FirstShortStack":
		return boardgame.ErrPropertyImmutable
	case "HiddenCard":
		return boardgame.ErrPropertyImmutable
	case "HiddenStack":
		return boardgame.ErrPropertyImmutable
	case "SanitizedTokensFrom":
		return boardgame.ErrPropertyImmutable
	case "SanitizedTokensTo":
		return boardgame.ErrPropertyImmutable
	case "SecondShortStack":
		return boardgame.ErrPropertyImmutable
	case "TokensFrom":
		return boardgame.ErrPropertyImmutable
	case "TokensTo":
		return boardgame.ErrPropertyImmutable
	case "VisibleCard":
		return boardgame.ErrPropertyImmutable
	case "VisibleStack":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "AllHiddenStack":
		return g.data.AllHiddenStack, nil
	case "AllVisibleStack":
		return g.data.AllVisibleStack, nil
	case "Card":
		return nil, boardgame.ErrPropertyImmutable
	case "DiscardStack":
		return g.data.DiscardStack, nil
	case "DrawStack":
		return g.data.DrawStack, nil
	case "FanDiscard":
		return g.data.FanDiscard, nil
	case "FanStack":
		return g.data.FanStack, nil
	case "FirstShortStack":
		return g.data.FirstShortStack, nil
	case "HiddenCard":
		return g.data.HiddenCard, nil
	case "HiddenStack":
		return g.data.HiddenStack, nil
	case "SanitizedTokensFrom":
		return g.data.SanitizedTokensFrom, nil
	case "SanitizedTokensTo":
		return g.data.SanitizedTokensTo, nil
	case "SecondShortStack":
		return g.data.SecondShortStack, nil
	case "TokensFrom":
		return g.data.TokensFrom, nil
	case "TokensTo":
		return g.data.TokensTo, nil
	case "VisibleCard":
		return g.data.VisibleCard, nil
	case "VisibleStack":
		return g.data.VisibleStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

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

//Reader returns an autp-generated boardgame.PropertyReader for gameState
func (g *gameState) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedGameStateReader{g}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for gameState
func (g *gameState) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedGameStateReader{g}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for gameState
func (g *gameState) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedGameStateReader{g}
}

// Implementation for playerState

var ȧutoGeneratedPlayerStateReaderProps = map[string]boardgame.PropertyType{
	"Hand": boardgame.TypeStack,
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
	case boardgame.TypeBoard:
		return p.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return p.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return p.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return p.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return p.IntProp(name)
	case boardgame.TypeIntSlice:
		return p.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return p.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return p.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return p.ImmutableStackProp(name)
	case boardgame.TypeString:
		return p.StringProp(name)
	case boardgame.TypeStringSlice:
		return p.StringSliceProp(name)
	case boardgame.TypeTimer:
		return p.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (p *ȧutoGeneratedPlayerStateReader) PropMutable(name string) bool {
	switch name {
	case "Hand":
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
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return p.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return p.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return p.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return p.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return p.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return p.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return p.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return p.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
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
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return p.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return p.SetBoolSliceProp(name, val)
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
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return p.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return p.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return p.SetPlayerIndexProp(name, val)
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
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return p.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return p.SetStringSliceProp(name, val)
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

	}

	return errors.New("Unexpected property type: " + propType.String())
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

func (p *ȧutoGeneratedPlayerStateReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

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

func (p *ȧutoGeneratedPlayerStateReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

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

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "Hand":
		p.data.Hand = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "Hand":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "Hand":
		return p.data.Hand, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

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

//Reader returns an autp-generated boardgame.PropertyReader for playerState
func (p *playerState) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedPlayerStateReader{p}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for playerState
func (p *playerState) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedPlayerStateReader{p}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for playerState
func (p *playerState) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedPlayerStateReader{p}
}
