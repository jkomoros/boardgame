/************************************
 *
 * This file contains auto-generated methods to help certain structs
 * implement boardgame.PropertyReader and friends. It was generated
 * by the codegen package via 'boardgame-util codegen'.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// Implementation for moveShuffleStack

var __moveShuffleStackReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __moveShuffleStackReader struct {
	data *moveShuffleStack
}

func (m *__moveShuffleStackReader) Props() map[string]boardgame.PropertyType {
	return __moveShuffleStackReaderProps
}

func (m *__moveShuffleStackReader) Prop(name string) (interface{}, error) {
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

func (m *__moveShuffleStackReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *__moveShuffleStackReader) SetProp(name string, value interface{}) error {
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

func (m *__moveShuffleStackReader) ConfigureProp(name string, value interface{}) error {
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

func (m *__moveShuffleStackReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveShuffleStackReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__moveShuffleStackReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__moveShuffleStackReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveShuffleStackReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__moveShuffleStackReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__moveShuffleStackReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveShuffleStackReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveShuffleStackReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveShuffleStackReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *__moveShuffleStackReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__moveShuffleStackReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveShuffleStackReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__moveShuffleStackReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__moveShuffleStackReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__moveShuffleStackReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__moveShuffleStackReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveShuffleStackReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveShuffleStackReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveShuffleStackReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveShuffleStackReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveShuffleStackReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *__moveShuffleStackReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__moveShuffleStackReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveShuffleStackReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__moveShuffleStackReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__moveShuffleStackReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__moveShuffleStackReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__moveShuffleStackReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__moveShuffleStackReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__moveShuffleStackReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__moveShuffleStackReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *moveShuffleStack) Reader() boardgame.PropertyReader {
	return &__moveShuffleStackReader{m}
}

func (m *moveShuffleStack) ReadSetter() boardgame.PropertyReadSetter {
	return &__moveShuffleStackReader{m}
}

func (m *moveShuffleStack) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__moveShuffleStackReader{m}
}

// Implementation for moveNoOpFixUp

var __moveNoOpFixUpReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __moveNoOpFixUpReader struct {
	data *moveNoOpFixUp
}

func (m *__moveNoOpFixUpReader) Props() map[string]boardgame.PropertyType {
	return __moveNoOpFixUpReaderProps
}

func (m *__moveNoOpFixUpReader) Prop(name string) (interface{}, error) {
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

func (m *__moveNoOpFixUpReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *__moveNoOpFixUpReader) SetProp(name string, value interface{}) error {
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

func (m *__moveNoOpFixUpReader) ConfigureProp(name string, value interface{}) error {
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

func (m *__moveNoOpFixUpReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveNoOpFixUpReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__moveNoOpFixUpReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__moveNoOpFixUpReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveNoOpFixUpReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__moveNoOpFixUpReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__moveNoOpFixUpReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveNoOpFixUpReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveNoOpFixUpReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveNoOpFixUpReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *__moveNoOpFixUpReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__moveNoOpFixUpReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveNoOpFixUpReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__moveNoOpFixUpReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__moveNoOpFixUpReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__moveNoOpFixUpReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__moveNoOpFixUpReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveNoOpFixUpReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveNoOpFixUpReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveNoOpFixUpReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveNoOpFixUpReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveNoOpFixUpReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *__moveNoOpFixUpReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__moveNoOpFixUpReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveNoOpFixUpReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__moveNoOpFixUpReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__moveNoOpFixUpReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__moveNoOpFixUpReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__moveNoOpFixUpReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__moveNoOpFixUpReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__moveNoOpFixUpReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__moveNoOpFixUpReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *moveNoOpFixUp) Reader() boardgame.PropertyReader {
	return &__moveNoOpFixUpReader{m}
}

func (m *moveNoOpFixUp) ReadSetter() boardgame.PropertyReadSetter {
	return &__moveNoOpFixUpReader{m}
}

func (m *moveNoOpFixUp) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__moveNoOpFixUpReader{m}
}

// Implementation for moveNoOpFixUpMulti

var __moveNoOpFixUpMultiReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __moveNoOpFixUpMultiReader struct {
	data *moveNoOpFixUpMulti
}

func (m *__moveNoOpFixUpMultiReader) Props() map[string]boardgame.PropertyType {
	return __moveNoOpFixUpMultiReaderProps
}

func (m *__moveNoOpFixUpMultiReader) Prop(name string) (interface{}, error) {
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

func (m *__moveNoOpFixUpMultiReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *__moveNoOpFixUpMultiReader) SetProp(name string, value interface{}) error {
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

func (m *__moveNoOpFixUpMultiReader) ConfigureProp(name string, value interface{}) error {
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

func (m *__moveNoOpFixUpMultiReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__moveNoOpFixUpMultiReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *moveNoOpFixUpMulti) Reader() boardgame.PropertyReader {
	return &__moveNoOpFixUpMultiReader{m}
}

func (m *moveNoOpFixUpMulti) ReadSetter() boardgame.PropertyReadSetter {
	return &__moveNoOpFixUpMultiReader{m}
}

func (m *moveNoOpFixUpMulti) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__moveNoOpFixUpMultiReader{m}
}

// Implementation for gameState

var __gameStateReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"Counter":         boardgame.TypeInt,
	"CurrentPlayer":   boardgame.TypePlayerIndex,
	"DiscardStack":    boardgame.TypeStack,
	"DrawStack":       boardgame.TypeStack,
	"Phase":           boardgame.TypeEnum,
	"RRHasStarted":    boardgame.TypeBool,
	"RRLastPlayer":    boardgame.TypePlayerIndex,
	"RRRoundCount":    boardgame.TypeInt,
	"RRStarterPlayer": boardgame.TypePlayerIndex,
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
	case "Counter":
		return true
	case "CurrentPlayer":
		return true
	case "DiscardStack":
		return true
	case "DrawStack":
		return true
	case "Phase":
		return true
	case "RRHasStarted":
		return true
	case "RRLastPlayer":
		return true
	case "RRRoundCount":
		return true
	case "RRStarterPlayer":
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

	switch name {
	case "RRHasStarted":
		return g.data.RRHasStarted, nil

	}

	return false, errors.New("No such Bool prop: " + name)

}

func (g *__gameStateReader) SetBoolProp(name string, value bool) error {

	switch name {
	case "RRHasStarted":
		g.data.RRHasStarted = value
		return nil

	}

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

	switch name {
	case "Counter":
		return g.data.Counter, nil
	case "RRRoundCount":
		return g.data.RRRoundCount, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (g *__gameStateReader) SetIntProp(name string, value int) error {

	switch name {
	case "Counter":
		g.data.Counter = value
		return nil
	case "RRRoundCount":
		g.data.RRRoundCount = value
		return nil

	}

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
	case "RRLastPlayer":
		return g.data.RRLastPlayer, nil
	case "RRStarterPlayer":
		return g.data.RRStarterPlayer, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (g *__gameStateReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "CurrentPlayer":
		g.data.CurrentPlayer = value
		return nil
	case "RRLastPlayer":
		g.data.RRLastPlayer = value
		return nil
	case "RRStarterPlayer":
		g.data.RRStarterPlayer = value
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
	case "DiscardStack":
		return g.data.DiscardStack, nil
	case "DrawStack":
		return g.data.DrawStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (g *__gameStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "DiscardStack":
		g.data.DiscardStack = value
		return nil
	case "DrawStack":
		g.data.DrawStack = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (g *__gameStateReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "DiscardStack":
		return boardgame.ErrPropertyImmutable
	case "DrawStack":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (g *__gameStateReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "DiscardStack":
		return g.data.DiscardStack, nil
	case "DrawStack":
		return g.data.DrawStack, nil

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
	"Counter":   boardgame.TypeInt,
	"Hand":      boardgame.TypeStack,
	"OtherHand": boardgame.TypeStack,
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
	case "Counter":
		return true
	case "Hand":
		return true
	case "OtherHand":
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

	return false, errors.New("No such Bool prop: " + name)

}

func (p *__playerStateReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (p *__playerStateReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (p *__playerStateReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (p *__playerStateReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (p *__playerStateReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (p *__playerStateReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (p *__playerStateReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (p *__playerStateReader) IntProp(name string) (int, error) {

	switch name {
	case "Counter":
		return p.data.Counter, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (p *__playerStateReader) SetIntProp(name string, value int) error {

	switch name {
	case "Counter":
		p.data.Counter = value
		return nil

	}

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
	case "Hand":
		return p.data.Hand, nil
	case "OtherHand":
		return p.data.OtherHand, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (p *__playerStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "Hand":
		p.data.Hand = value
		return nil
	case "OtherHand":
		p.data.OtherHand = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (p *__playerStateReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "Hand":
		return boardgame.ErrPropertyImmutable
	case "OtherHand":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (p *__playerStateReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "Hand":
		return p.data.Hand, nil
	case "OtherHand":
		return p.data.OtherHand, nil

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

// Implementation for moveDealCards

var __moveDealCardsReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __moveDealCardsReader struct {
	data *moveDealCards
}

func (m *__moveDealCardsReader) Props() map[string]boardgame.PropertyType {
	return __moveDealCardsReaderProps
}

func (m *__moveDealCardsReader) Prop(name string) (interface{}, error) {
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

func (m *__moveDealCardsReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *__moveDealCardsReader) SetProp(name string, value interface{}) error {
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

func (m *__moveDealCardsReader) ConfigureProp(name string, value interface{}) error {
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

func (m *__moveDealCardsReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveDealCardsReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__moveDealCardsReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__moveDealCardsReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveDealCardsReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__moveDealCardsReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__moveDealCardsReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveDealCardsReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveDealCardsReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveDealCardsReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *__moveDealCardsReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__moveDealCardsReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveDealCardsReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__moveDealCardsReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__moveDealCardsReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__moveDealCardsReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__moveDealCardsReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveDealCardsReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveDealCardsReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveDealCardsReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveDealCardsReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveDealCardsReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *__moveDealCardsReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__moveDealCardsReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveDealCardsReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__moveDealCardsReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__moveDealCardsReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__moveDealCardsReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__moveDealCardsReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__moveDealCardsReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__moveDealCardsReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__moveDealCardsReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *moveDealCards) Reader() boardgame.PropertyReader {
	return &__moveDealCardsReader{m}
}

func (m *moveDealCards) ReadSetter() boardgame.PropertyReadSetter {
	return &__moveDealCardsReader{m}
}

func (m *moveDealCards) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__moveDealCardsReader{m}
}

// Implementation for moveDealOtherCards

var __moveDealOtherCardsReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __moveDealOtherCardsReader struct {
	data *moveDealOtherCards
}

func (m *__moveDealOtherCardsReader) Props() map[string]boardgame.PropertyType {
	return __moveDealOtherCardsReaderProps
}

func (m *__moveDealOtherCardsReader) Prop(name string) (interface{}, error) {
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

func (m *__moveDealOtherCardsReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *__moveDealOtherCardsReader) SetProp(name string, value interface{}) error {
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

func (m *__moveDealOtherCardsReader) ConfigureProp(name string, value interface{}) error {
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

func (m *__moveDealOtherCardsReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveDealOtherCardsReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__moveDealOtherCardsReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__moveDealOtherCardsReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveDealOtherCardsReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__moveDealOtherCardsReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__moveDealOtherCardsReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveDealOtherCardsReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveDealOtherCardsReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveDealOtherCardsReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *__moveDealOtherCardsReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__moveDealOtherCardsReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveDealOtherCardsReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__moveDealOtherCardsReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__moveDealOtherCardsReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__moveDealOtherCardsReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__moveDealOtherCardsReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveDealOtherCardsReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveDealOtherCardsReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveDealOtherCardsReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveDealOtherCardsReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveDealOtherCardsReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *__moveDealOtherCardsReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__moveDealOtherCardsReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveDealOtherCardsReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__moveDealOtherCardsReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__moveDealOtherCardsReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__moveDealOtherCardsReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__moveDealOtherCardsReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__moveDealOtherCardsReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__moveDealOtherCardsReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__moveDealOtherCardsReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *moveDealOtherCards) Reader() boardgame.PropertyReader {
	return &__moveDealOtherCardsReader{m}
}

func (m *moveDealOtherCards) ReadSetter() boardgame.PropertyReadSetter {
	return &__moveDealOtherCardsReader{m}
}

func (m *moveDealOtherCards) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__moveDealOtherCardsReader{m}
}

// Implementation for moveCurrentPlayerDraw

var __moveCurrentPlayerDrawReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"TargetPlayerIndex": boardgame.TypePlayerIndex,
}

type __moveCurrentPlayerDrawReader struct {
	data *moveCurrentPlayerDraw
}

func (m *__moveCurrentPlayerDrawReader) Props() map[string]boardgame.PropertyType {
	return __moveCurrentPlayerDrawReaderProps
}

func (m *__moveCurrentPlayerDrawReader) Prop(name string) (interface{}, error) {
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

func (m *__moveCurrentPlayerDrawReader) PropMutable(name string) bool {
	switch name {
	case "TargetPlayerIndex":
		return true
	}

	return false
}

func (m *__moveCurrentPlayerDrawReader) SetProp(name string, value interface{}) error {
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

func (m *__moveCurrentPlayerDrawReader) ConfigureProp(name string, value interface{}) error {
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

func (m *__moveCurrentPlayerDrawReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "TargetPlayerIndex":
		return m.data.TargetPlayerIndex, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "TargetPlayerIndex":
		m.data.TargetPlayerIndex = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__moveCurrentPlayerDrawReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *moveCurrentPlayerDraw) Reader() boardgame.PropertyReader {
	return &__moveCurrentPlayerDrawReader{m}
}

func (m *moveCurrentPlayerDraw) ReadSetter() boardgame.PropertyReadSetter {
	return &__moveCurrentPlayerDrawReader{m}
}

func (m *moveCurrentPlayerDraw) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__moveCurrentPlayerDrawReader{m}
}

// Implementation for moveStartPhaseDrawAgain

var __moveStartPhaseDrawAgainReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __moveStartPhaseDrawAgainReader struct {
	data *moveStartPhaseDrawAgain
}

func (m *__moveStartPhaseDrawAgainReader) Props() map[string]boardgame.PropertyType {
	return __moveStartPhaseDrawAgainReaderProps
}

func (m *__moveStartPhaseDrawAgainReader) Prop(name string) (interface{}, error) {
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

func (m *__moveStartPhaseDrawAgainReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *__moveStartPhaseDrawAgainReader) SetProp(name string, value interface{}) error {
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

func (m *__moveStartPhaseDrawAgainReader) ConfigureProp(name string, value interface{}) error {
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

func (m *__moveStartPhaseDrawAgainReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__moveStartPhaseDrawAgainReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *moveStartPhaseDrawAgain) Reader() boardgame.PropertyReader {
	return &__moveStartPhaseDrawAgainReader{m}
}

func (m *moveStartPhaseDrawAgain) ReadSetter() boardgame.PropertyReadSetter {
	return &__moveStartPhaseDrawAgainReader{m}
}

func (m *moveStartPhaseDrawAgain) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__moveStartPhaseDrawAgainReader{m}
}

// Implementation for moveStartPhaseIllegal

var __moveStartPhaseIllegalReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __moveStartPhaseIllegalReader struct {
	data *moveStartPhaseIllegal
}

func (m *__moveStartPhaseIllegalReader) Props() map[string]boardgame.PropertyType {
	return __moveStartPhaseIllegalReaderProps
}

func (m *__moveStartPhaseIllegalReader) Prop(name string) (interface{}, error) {
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

func (m *__moveStartPhaseIllegalReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *__moveStartPhaseIllegalReader) SetProp(name string, value interface{}) error {
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

func (m *__moveStartPhaseIllegalReader) ConfigureProp(name string, value interface{}) error {
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

func (m *__moveStartPhaseIllegalReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__moveStartPhaseIllegalReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *moveStartPhaseIllegal) Reader() boardgame.PropertyReader {
	return &__moveStartPhaseIllegalReader{m}
}

func (m *moveStartPhaseIllegal) ReadSetter() boardgame.PropertyReadSetter {
	return &__moveStartPhaseIllegalReader{m}
}

func (m *moveStartPhaseIllegal) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__moveStartPhaseIllegalReader{m}
}
