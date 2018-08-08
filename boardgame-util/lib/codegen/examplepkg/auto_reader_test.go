/************************************
 *
 * This file contains auto-generated methods to help certain structs
 * implement boardgame.PropertyReader and friends. It was generated
 * by the codegen package via 'boardgame-util codegen'.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package examplepkg

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// Implementation for testStruct

var __testStructReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"A": boardgame.TypeInt,
	"B": boardgame.TypeString,
}

type __testStructReader struct {
	data *testStruct
}

func (t *__testStructReader) Props() map[string]boardgame.PropertyType {
	return __testStructReaderProps
}

func (t *__testStructReader) Prop(name string) (interface{}, error) {
	props := t.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return t.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return t.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return t.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return t.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return t.IntProp(name)
	case boardgame.TypeIntSlice:
		return t.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return t.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return t.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return t.ImmutableStackProp(name)
	case boardgame.TypeString:
		return t.StringProp(name)
	case boardgame.TypeStringSlice:
		return t.StringSliceProp(name)
	case boardgame.TypeTimer:
		return t.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (t *__testStructReader) PropMutable(name string) bool {
	switch name {
	case "A":
		return true
	case "B":
		return true
	}

	return false
}

func (t *__testStructReader) SetProp(name string, value interface{}) error {
	props := t.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return t.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return t.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return t.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return t.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return t.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return t.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return t.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return t.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (t *__testStructReader) ConfigureProp(name string, value interface{}) error {
	props := t.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if t.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return t.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return t.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return t.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return t.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if t.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return t.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return t.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return t.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return t.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return t.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return t.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if t.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return t.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return t.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return t.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return t.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if t.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return t.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return t.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (t *__testStructReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (t *__testStructReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (t *__testStructReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (t *__testStructReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (t *__testStructReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (t *__testStructReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (t *__testStructReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (t *__testStructReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (t *__testStructReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (t *__testStructReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (t *__testStructReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (t *__testStructReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (t *__testStructReader) IntProp(name string) (int, error) {

	switch name {
	case "A":
		return t.data.A, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (t *__testStructReader) SetIntProp(name string, value int) error {

	switch name {
	case "A":
		t.data.A = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (t *__testStructReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (t *__testStructReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (t *__testStructReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (t *__testStructReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (t *__testStructReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (t *__testStructReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (t *__testStructReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (t *__testStructReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (t *__testStructReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (t *__testStructReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (t *__testStructReader) StringProp(name string) (string, error) {

	switch name {
	case "B":
		return t.data.B, nil

	}

	return "", errors.New("No such String prop: " + name)

}

func (t *__testStructReader) SetStringProp(name string, value string) error {

	switch name {
	case "B":
		t.data.B = value
		return nil

	}

	return errors.New("No such String prop: " + name)

}

func (t *__testStructReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (t *__testStructReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (t *__testStructReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (t *__testStructReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (t *__testStructReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (t *__testStructReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (t *testStruct) Reader() boardgame.PropertyReader {
	return &__testStructReader{t}
}

func (t *testStruct) ReadSetter() boardgame.PropertyReadSetter {
	return &__testStructReader{t}
}

func (t *testStruct) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__testStructReader{t}
}
