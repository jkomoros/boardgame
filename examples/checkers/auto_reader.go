/************************************
 *
 * This file contains auto-generated methods to help certain structs
 * implement boardgame.PropertyReader and friends. It was generated
 * by the codegen package via 'boardgame-util codegen'.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package checkers

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// Implementation for token

var __tokenReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"Color": boardgame.TypeEnum,
}

type __tokenReader struct {
	data *token
}

func (t *__tokenReader) Props() map[string]boardgame.PropertyType {
	return __tokenReaderProps
}

func (t *__tokenReader) Prop(name string) (interface{}, error) {
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

func (t *__tokenReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (t *__tokenReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (t *__tokenReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (t *__tokenReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "Color":
		return t.data.Color, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (t *__tokenReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (t *__tokenReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (t *__tokenReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (t *__tokenReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (t *__tokenReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (t *__tokenReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (t *__tokenReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (t *__tokenReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (t *token) Reader() boardgame.PropertyReader {
	return &__tokenReader{t}
}

// Implementation for tokenDynamic

var __tokenDynamicReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"Crowned": boardgame.TypeBool,
}

type __tokenDynamicReader struct {
	data *tokenDynamic
}

func (t *__tokenDynamicReader) Props() map[string]boardgame.PropertyType {
	return __tokenDynamicReaderProps
}

func (t *__tokenDynamicReader) Prop(name string) (interface{}, error) {
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

func (t *__tokenDynamicReader) PropMutable(name string) bool {
	switch name {
	case "Crowned":
		return true
	}

	return false
}

func (t *__tokenDynamicReader) SetProp(name string, value interface{}) error {
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

func (t *__tokenDynamicReader) ConfigureProp(name string, value interface{}) error {
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

func (t *__tokenDynamicReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (t *__tokenDynamicReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (t *__tokenDynamicReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (t *__tokenDynamicReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (t *__tokenDynamicReader) BoolProp(name string) (bool, error) {

	switch name {
	case "Crowned":
		return t.data.Crowned, nil

	}

	return false, errors.New("No such Bool prop: " + name)

}

func (t *__tokenDynamicReader) SetBoolProp(name string, value bool) error {

	switch name {
	case "Crowned":
		t.data.Crowned = value
		return nil

	}

	return errors.New("No such Bool prop: " + name)

}

func (t *__tokenDynamicReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (t *__tokenDynamicReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (t *__tokenDynamicReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (t *__tokenDynamicReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (t *__tokenDynamicReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (t *__tokenDynamicReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (t *__tokenDynamicReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (t *__tokenDynamicReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (t *__tokenDynamicReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (t *__tokenDynamicReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (t *__tokenDynamicReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (t *__tokenDynamicReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (t *__tokenDynamicReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (t *__tokenDynamicReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (t *__tokenDynamicReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (t *__tokenDynamicReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (t *__tokenDynamicReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (t *__tokenDynamicReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (t *__tokenDynamicReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (t *__tokenDynamicReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (t *__tokenDynamicReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (t *__tokenDynamicReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (t *__tokenDynamicReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (t *__tokenDynamicReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (t *__tokenDynamicReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (t *__tokenDynamicReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (t *tokenDynamic) Reader() boardgame.PropertyReader {
	return &__tokenDynamicReader{t}
}

func (t *tokenDynamic) ReadSetter() boardgame.PropertyReadSetter {
	return &__tokenDynamicReader{t}
}

func (t *tokenDynamic) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__tokenDynamicReader{t}
}

// Implementation for movePlaceToken

var __movePlaceTokenReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"TargetIndex": boardgame.TypeEnum,
}

type __movePlaceTokenReader struct {
	data *movePlaceToken
}

func (m *__movePlaceTokenReader) Props() map[string]boardgame.PropertyType {
	return __movePlaceTokenReaderProps
}

func (m *__movePlaceTokenReader) Prop(name string) (interface{}, error) {
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

func (m *__movePlaceTokenReader) PropMutable(name string) bool {
	switch name {
	case "TargetIndex":
		return true
	}

	return false
}

func (m *__movePlaceTokenReader) SetProp(name string, value interface{}) error {
	props := m.Props()
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
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
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
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
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
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *__movePlaceTokenReader) ConfigureProp(name string, value interface{}) error {
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return m.ConfigureImmutableBoardProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return m.ConfigureImmutableEnumProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return m.ConfigureImmutableStackProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return m.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *__movePlaceTokenReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__movePlaceTokenReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__movePlaceTokenReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__movePlaceTokenReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__movePlaceTokenReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__movePlaceTokenReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__movePlaceTokenReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__movePlaceTokenReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__movePlaceTokenReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "TargetIndex":
		return m.data.TargetIndex, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__movePlaceTokenReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "TargetIndex":
		slotValue := value.RangeVal()
		if slotValue == nil {
			return errors.New("TargetIndex couldn't be upconverted, returned nil.")
		}
		m.data.TargetIndex = slotValue
		return nil

	}

	return errors.New("No such Enum prop: " + name)

}

func (m *__movePlaceTokenReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "TargetIndex":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__movePlaceTokenReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "TargetIndex":
		return m.data.TargetIndex, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__movePlaceTokenReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__movePlaceTokenReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__movePlaceTokenReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__movePlaceTokenReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__movePlaceTokenReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__movePlaceTokenReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__movePlaceTokenReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__movePlaceTokenReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__movePlaceTokenReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__movePlaceTokenReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *__movePlaceTokenReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__movePlaceTokenReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__movePlaceTokenReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__movePlaceTokenReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__movePlaceTokenReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__movePlaceTokenReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__movePlaceTokenReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__movePlaceTokenReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__movePlaceTokenReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__movePlaceTokenReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *movePlaceToken) Reader() boardgame.PropertyReader {
	return &__movePlaceTokenReader{m}
}

func (m *movePlaceToken) ReadSetter() boardgame.PropertyReadSetter {
	return &__movePlaceTokenReader{m}
}

func (m *movePlaceToken) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__movePlaceTokenReader{m}
}

// Implementation for moveMoveToken

var __moveMoveTokenReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"SpaceIndex":        boardgame.TypeEnum,
	"TargetPlayerIndex": boardgame.TypePlayerIndex,
	"TokenIndexToMove":  boardgame.TypeEnum,
}

type __moveMoveTokenReader struct {
	data *moveMoveToken
}

func (m *__moveMoveTokenReader) Props() map[string]boardgame.PropertyType {
	return __moveMoveTokenReaderProps
}

func (m *__moveMoveTokenReader) Prop(name string) (interface{}, error) {
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

func (m *__moveMoveTokenReader) PropMutable(name string) bool {
	switch name {
	case "SpaceIndex":
		return true
	case "TargetPlayerIndex":
		return true
	case "TokenIndexToMove":
		return true
	}

	return false
}

func (m *__moveMoveTokenReader) SetProp(name string, value interface{}) error {
	props := m.Props()
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
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
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
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
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
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *__moveMoveTokenReader) ConfigureProp(name string, value interface{}) error {
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return m.ConfigureImmutableBoardProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return m.ConfigureImmutableEnumProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return m.ConfigureImmutableStackProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return m.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *__moveMoveTokenReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveMoveTokenReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__moveMoveTokenReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__moveMoveTokenReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveMoveTokenReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__moveMoveTokenReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__moveMoveTokenReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveMoveTokenReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveMoveTokenReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "SpaceIndex":
		return m.data.SpaceIndex, nil
	case "TokenIndexToMove":
		return m.data.TokenIndexToMove, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveMoveTokenReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "SpaceIndex":
		slotValue := value.RangeVal()
		if slotValue == nil {
			return errors.New("SpaceIndex couldn't be upconverted, returned nil.")
		}
		m.data.SpaceIndex = slotValue
		return nil
	case "TokenIndexToMove":
		slotValue := value.RangeVal()
		if slotValue == nil {
			return errors.New("TokenIndexToMove couldn't be upconverted, returned nil.")
		}
		m.data.TokenIndexToMove = slotValue
		return nil

	}

	return errors.New("No such Enum prop: " + name)

}

func (m *__moveMoveTokenReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "SpaceIndex":
		return boardgame.ErrPropertyImmutable
	case "TokenIndexToMove":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__moveMoveTokenReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "SpaceIndex":
		return m.data.SpaceIndex, nil
	case "TokenIndexToMove":
		return m.data.TokenIndexToMove, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveMoveTokenReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__moveMoveTokenReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__moveMoveTokenReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__moveMoveTokenReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__moveMoveTokenReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "TargetPlayerIndex":
		return m.data.TargetPlayerIndex, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveMoveTokenReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "TargetPlayerIndex":
		m.data.TargetPlayerIndex = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveMoveTokenReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveMoveTokenReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveMoveTokenReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveMoveTokenReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *__moveMoveTokenReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__moveMoveTokenReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveMoveTokenReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__moveMoveTokenReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__moveMoveTokenReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__moveMoveTokenReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__moveMoveTokenReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__moveMoveTokenReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__moveMoveTokenReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__moveMoveTokenReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *moveMoveToken) Reader() boardgame.PropertyReader {
	return &__moveMoveTokenReader{m}
}

func (m *moveMoveToken) ReadSetter() boardgame.PropertyReadSetter {
	return &__moveMoveTokenReader{m}
}

func (m *moveMoveToken) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__moveMoveTokenReader{m}
}

// Implementation for moveCrownToken

var __moveCrownTokenReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"ComponentIndex": boardgame.TypeInt,
}

type __moveCrownTokenReader struct {
	data *moveCrownToken
}

func (m *__moveCrownTokenReader) Props() map[string]boardgame.PropertyType {
	return __moveCrownTokenReaderProps
}

func (m *__moveCrownTokenReader) Prop(name string) (interface{}, error) {
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

func (m *__moveCrownTokenReader) PropMutable(name string) bool {
	switch name {
	case "ComponentIndex":
		return true
	}

	return false
}

func (m *__moveCrownTokenReader) SetProp(name string, value interface{}) error {
	props := m.Props()
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
		return m.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return m.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
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
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
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
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *__moveCrownTokenReader) ConfigureProp(name string, value interface{}) error {
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return m.ConfigureImmutableBoardProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return m.ConfigureImmutableEnumProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return m.ConfigureImmutableStackProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return m.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *__moveCrownTokenReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveCrownTokenReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__moveCrownTokenReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__moveCrownTokenReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveCrownTokenReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__moveCrownTokenReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__moveCrownTokenReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveCrownTokenReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveCrownTokenReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveCrownTokenReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *__moveCrownTokenReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__moveCrownTokenReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveCrownTokenReader) IntProp(name string) (int, error) {

	switch name {
	case "ComponentIndex":
		return m.data.ComponentIndex, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__moveCrownTokenReader) SetIntProp(name string, value int) error {

	switch name {
	case "ComponentIndex":
		m.data.ComponentIndex = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (m *__moveCrownTokenReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__moveCrownTokenReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__moveCrownTokenReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveCrownTokenReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveCrownTokenReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveCrownTokenReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveCrownTokenReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveCrownTokenReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *__moveCrownTokenReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__moveCrownTokenReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveCrownTokenReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__moveCrownTokenReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__moveCrownTokenReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__moveCrownTokenReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__moveCrownTokenReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__moveCrownTokenReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__moveCrownTokenReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__moveCrownTokenReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *moveCrownToken) Reader() boardgame.PropertyReader {
	return &__moveCrownTokenReader{m}
}

func (m *moveCrownToken) ReadSetter() boardgame.PropertyReadSetter {
	return &__moveCrownTokenReader{m}
}

func (m *moveCrownToken) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__moveCrownTokenReader{m}
}

// Implementation for gameState

var __gameStateReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"CurrentPlayer": boardgame.TypePlayerIndex,
	"Phase":         boardgame.TypeEnum,
	"Spaces":        boardgame.TypeStack,
	"UnusedTokens":  boardgame.TypeStack,
}

type __gameStateReader struct {
	data *gameState
}

func (g *__gameStateReader) Props() map[string]boardgame.PropertyType {
	return __gameStateReaderProps
}

func (g *__gameStateReader) Prop(name string) (interface{}, error) {
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

func (g *__gameStateReader) PropMutable(name string) bool {
	switch name {
	case "CurrentPlayer":
		return true
	case "Phase":
		return true
	case "Spaces":
		return true
	case "UnusedTokens":
		return true
	}

	return false
}

func (g *__gameStateReader) SetProp(name string, value interface{}) error {
	props := g.Props()
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
		return g.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return g.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
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
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
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
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (g *__gameStateReader) ConfigureProp(name string, value interface{}) error {
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return g.ConfigureImmutableBoardProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return g.ConfigureImmutableEnumProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return g.ConfigureImmutableStackProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return g.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (g *__gameStateReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (g *__gameStateReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (g *__gameStateReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (g *__gameStateReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (g *__gameStateReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (g *__gameStateReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (g *__gameStateReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (g *__gameStateReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (g *__gameStateReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "Phase":
		return g.data.Phase, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (g *__gameStateReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "Phase":
		g.data.Phase = value
		return nil

	}

	return errors.New("No such Enum prop: " + name)

}

func (g *__gameStateReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "Phase":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (g *__gameStateReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "Phase":
		return g.data.Phase, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (g *__gameStateReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (g *__gameStateReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (g *__gameStateReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (g *__gameStateReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (g *__gameStateReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "CurrentPlayer":
		return g.data.CurrentPlayer, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (g *__gameStateReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "CurrentPlayer":
		g.data.CurrentPlayer = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (g *__gameStateReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (g *__gameStateReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (g *__gameStateReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "Spaces":
		return g.data.Spaces, nil
	case "UnusedTokens":
		return g.data.UnusedTokens, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (g *__gameStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "Spaces":
		slotValue := value.SizedStack()
		if slotValue == nil {
			return errors.New("Spaces couldn't be upconverted, returned nil.")
		}
		g.data.Spaces = slotValue
		return nil
	case "UnusedTokens":
		g.data.UnusedTokens = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (g *__gameStateReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "Spaces":
		return boardgame.ErrPropertyImmutable
	case "UnusedTokens":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (g *__gameStateReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "Spaces":
		return g.data.Spaces, nil
	case "UnusedTokens":
		return g.data.UnusedTokens, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (g *__gameStateReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (g *__gameStateReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (g *__gameStateReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (g *__gameStateReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (g *__gameStateReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (g *__gameStateReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (g *__gameStateReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (g *__gameStateReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (g *gameState) Reader() boardgame.PropertyReader {
	return &__gameStateReader{g}
}

func (g *gameState) ReadSetter() boardgame.PropertyReadSetter {
	return &__gameStateReader{g}
}

func (g *gameState) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__gameStateReader{g}
}

// Implementation for playerState

var __playerStateReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"CapturedTokens": boardgame.TypeStack,
	"Color":          boardgame.TypeEnum,
	"FinishedTurn":   boardgame.TypeBool,
}

type __playerStateReader struct {
	data *playerState
}

func (p *__playerStateReader) Props() map[string]boardgame.PropertyType {
	return __playerStateReaderProps
}

func (p *__playerStateReader) Prop(name string) (interface{}, error) {
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

func (p *__playerStateReader) PropMutable(name string) bool {
	switch name {
	case "CapturedTokens":
		return true
	case "Color":
		return true
	case "FinishedTurn":
		return true
	}

	return false
}

func (p *__playerStateReader) SetProp(name string, value interface{}) error {
	props := p.Props()
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
		return p.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return p.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
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
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
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
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (p *__playerStateReader) ConfigureProp(name string, value interface{}) error {
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return p.ConfigureImmutableBoardProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return p.ConfigureImmutableEnumProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return p.ConfigureImmutableStackProp(name, val)
		}
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return p.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (p *__playerStateReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (p *__playerStateReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (p *__playerStateReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (p *__playerStateReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (p *__playerStateReader) BoolProp(name string) (bool, error) {

	switch name {
	case "FinishedTurn":
		return p.data.FinishedTurn, nil

	}

	return false, errors.New("No such Bool prop: " + name)

}

func (p *__playerStateReader) SetBoolProp(name string, value bool) error {

	switch name {
	case "FinishedTurn":
		p.data.FinishedTurn = value
		return nil

	}

	return errors.New("No such Bool prop: " + name)

}

func (p *__playerStateReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (p *__playerStateReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (p *__playerStateReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "Color":
		return p.data.Color, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (p *__playerStateReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "Color":
		p.data.Color = value
		return nil

	}

	return errors.New("No such Enum prop: " + name)

}

func (p *__playerStateReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "Color":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (p *__playerStateReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "Color":
		return p.data.Color, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (p *__playerStateReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (p *__playerStateReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (p *__playerStateReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (p *__playerStateReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (p *__playerStateReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (p *__playerStateReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (p *__playerStateReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (p *__playerStateReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (p *__playerStateReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "CapturedTokens":
		return p.data.CapturedTokens, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (p *__playerStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "CapturedTokens":
		p.data.CapturedTokens = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (p *__playerStateReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "CapturedTokens":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (p *__playerStateReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "CapturedTokens":
		return p.data.CapturedTokens, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (p *__playerStateReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (p *__playerStateReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (p *__playerStateReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (p *__playerStateReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (p *__playerStateReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (p *__playerStateReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (p *__playerStateReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (p *__playerStateReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (p *playerState) Reader() boardgame.PropertyReader {
	return &__playerStateReader{p}
}

func (p *playerState) ReadSetter() boardgame.PropertyReadSetter {
	return &__playerStateReader{p}
}

func (p *playerState) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__playerStateReader{p}
}
