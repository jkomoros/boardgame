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

// Implementation for ApplyUntil

var __ApplyUntilReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __ApplyUntilReader struct {
	data *ApplyUntil
}

func (a *__ApplyUntilReader) Props() map[string]boardgame.PropertyType {
	return __ApplyUntilReaderProps
}

func (a *__ApplyUntilReader) Prop(name string) (interface{}, error) {
	props := a.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return a.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return a.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return a.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return a.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return a.IntProp(name)
	case boardgame.TypeIntSlice:
		return a.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return a.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return a.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return a.ImmutableStackProp(name)
	case boardgame.TypeString:
		return a.StringProp(name)
	case boardgame.TypeStringSlice:
		return a.StringSliceProp(name)
	case boardgame.TypeTimer:
		return a.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (a *__ApplyUntilReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (a *__ApplyUntilReader) SetProp(name string, value interface{}) error {
	props := a.Props()
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
		return a.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return a.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return a.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return a.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return a.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return a.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return a.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return a.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (a *__ApplyUntilReader) ConfigureProp(name string, value interface{}) error {
	props := a.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if a.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return a.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return a.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return a.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return a.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if a.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return a.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return a.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return a.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return a.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return a.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return a.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if a.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return a.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return a.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return a.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return a.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if a.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return a.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return a.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (a *__ApplyUntilReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (a *__ApplyUntilReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (a *__ApplyUntilReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (a *__ApplyUntilReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (a *__ApplyUntilReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (a *__ApplyUntilReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (a *__ApplyUntilReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (a *__ApplyUntilReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (a *__ApplyUntilReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (a *__ApplyUntilReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (a *__ApplyUntilReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (a *__ApplyUntilReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (a *__ApplyUntilReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (a *__ApplyUntilReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (a *__ApplyUntilReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (a *__ApplyUntilReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (a *__ApplyUntilReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (a *__ApplyUntilReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (a *__ApplyUntilReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (a *__ApplyUntilReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (a *__ApplyUntilReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (a *__ApplyUntilReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (a *__ApplyUntilReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (a *__ApplyUntilReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (a *__ApplyUntilReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (a *__ApplyUntilReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (a *__ApplyUntilReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (a *__ApplyUntilReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (a *__ApplyUntilReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (a *__ApplyUntilReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (a *__ApplyUntilReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (a *__ApplyUntilReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (a *ApplyUntil) Reader() boardgame.PropertyReader {
	return &__ApplyUntilReader{a}
}

func (a *ApplyUntil) ReadSetter() boardgame.PropertyReadSetter {
	return &__ApplyUntilReader{a}
}

func (a *ApplyUntil) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__ApplyUntilReader{a}
}

// Implementation for ApplyUntilCount

var __ApplyUntilCountReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __ApplyUntilCountReader struct {
	data *ApplyUntilCount
}

func (a *__ApplyUntilCountReader) Props() map[string]boardgame.PropertyType {
	return __ApplyUntilCountReaderProps
}

func (a *__ApplyUntilCountReader) Prop(name string) (interface{}, error) {
	props := a.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return a.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return a.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return a.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return a.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return a.IntProp(name)
	case boardgame.TypeIntSlice:
		return a.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return a.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return a.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return a.ImmutableStackProp(name)
	case boardgame.TypeString:
		return a.StringProp(name)
	case boardgame.TypeStringSlice:
		return a.StringSliceProp(name)
	case boardgame.TypeTimer:
		return a.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (a *__ApplyUntilCountReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (a *__ApplyUntilCountReader) SetProp(name string, value interface{}) error {
	props := a.Props()
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
		return a.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return a.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return a.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return a.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return a.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return a.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return a.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return a.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (a *__ApplyUntilCountReader) ConfigureProp(name string, value interface{}) error {
	props := a.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if a.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return a.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return a.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return a.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return a.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if a.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return a.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return a.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return a.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return a.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return a.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return a.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if a.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return a.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return a.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return a.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return a.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if a.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return a.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return a.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (a *__ApplyUntilCountReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (a *__ApplyUntilCountReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (a *__ApplyUntilCountReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (a *__ApplyUntilCountReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (a *__ApplyUntilCountReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (a *__ApplyUntilCountReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (a *__ApplyUntilCountReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (a *__ApplyUntilCountReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (a *__ApplyUntilCountReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (a *__ApplyUntilCountReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (a *__ApplyUntilCountReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (a *__ApplyUntilCountReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (a *__ApplyUntilCountReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (a *__ApplyUntilCountReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (a *__ApplyUntilCountReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (a *__ApplyUntilCountReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (a *__ApplyUntilCountReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (a *__ApplyUntilCountReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (a *__ApplyUntilCountReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (a *__ApplyUntilCountReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (a *__ApplyUntilCountReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (a *__ApplyUntilCountReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (a *__ApplyUntilCountReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (a *__ApplyUntilCountReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (a *__ApplyUntilCountReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (a *__ApplyUntilCountReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (a *__ApplyUntilCountReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (a *__ApplyUntilCountReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (a *__ApplyUntilCountReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (a *__ApplyUntilCountReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (a *__ApplyUntilCountReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (a *__ApplyUntilCountReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (a *ApplyUntilCount) Reader() boardgame.PropertyReader {
	return &__ApplyUntilCountReader{a}
}

func (a *ApplyUntilCount) ReadSetter() boardgame.PropertyReadSetter {
	return &__ApplyUntilCountReader{a}
}

func (a *ApplyUntilCount) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__ApplyUntilCountReader{a}
}

// Implementation for ApplyCountTimes

var __ApplyCountTimesReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __ApplyCountTimesReader struct {
	data *ApplyCountTimes
}

func (a *__ApplyCountTimesReader) Props() map[string]boardgame.PropertyType {
	return __ApplyCountTimesReaderProps
}

func (a *__ApplyCountTimesReader) Prop(name string) (interface{}, error) {
	props := a.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return a.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return a.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return a.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return a.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return a.IntProp(name)
	case boardgame.TypeIntSlice:
		return a.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return a.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return a.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return a.ImmutableStackProp(name)
	case boardgame.TypeString:
		return a.StringProp(name)
	case boardgame.TypeStringSlice:
		return a.StringSliceProp(name)
	case boardgame.TypeTimer:
		return a.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (a *__ApplyCountTimesReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (a *__ApplyCountTimesReader) SetProp(name string, value interface{}) error {
	props := a.Props()
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
		return a.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return a.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return a.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return a.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return a.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return a.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return a.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return a.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (a *__ApplyCountTimesReader) ConfigureProp(name string, value interface{}) error {
	props := a.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if a.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return a.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return a.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return a.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return a.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if a.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return a.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return a.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return a.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return a.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return a.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return a.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if a.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return a.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return a.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return a.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return a.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if a.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return a.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return a.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (a *__ApplyCountTimesReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (a *__ApplyCountTimesReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (a *__ApplyCountTimesReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (a *__ApplyCountTimesReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (a *__ApplyCountTimesReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (a *__ApplyCountTimesReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (a *__ApplyCountTimesReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (a *__ApplyCountTimesReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (a *__ApplyCountTimesReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (a *__ApplyCountTimesReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (a *__ApplyCountTimesReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (a *__ApplyCountTimesReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (a *__ApplyCountTimesReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (a *__ApplyCountTimesReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (a *__ApplyCountTimesReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (a *__ApplyCountTimesReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (a *__ApplyCountTimesReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (a *__ApplyCountTimesReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (a *__ApplyCountTimesReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (a *__ApplyCountTimesReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (a *__ApplyCountTimesReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (a *__ApplyCountTimesReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (a *__ApplyCountTimesReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (a *__ApplyCountTimesReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (a *__ApplyCountTimesReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (a *__ApplyCountTimesReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (a *__ApplyCountTimesReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (a *__ApplyCountTimesReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (a *__ApplyCountTimesReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (a *__ApplyCountTimesReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (a *__ApplyCountTimesReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (a *__ApplyCountTimesReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (a *ApplyCountTimes) Reader() boardgame.PropertyReader {
	return &__ApplyCountTimesReader{a}
}

func (a *ApplyCountTimes) ReadSetter() boardgame.PropertyReadSetter {
	return &__ApplyCountTimesReader{a}
}

func (a *ApplyCountTimes) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__ApplyCountTimesReader{a}
}

// Implementation for Base

var __BaseReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __BaseReader struct {
	data *Base
}

func (b *__BaseReader) Props() map[string]boardgame.PropertyType {
	return __BaseReaderProps
}

func (b *__BaseReader) Prop(name string) (interface{}, error) {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return b.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return b.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return b.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return b.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return b.IntProp(name)
	case boardgame.TypeIntSlice:
		return b.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return b.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return b.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return b.ImmutableStackProp(name)
	case boardgame.TypeString:
		return b.StringProp(name)
	case boardgame.TypeStringSlice:
		return b.StringSliceProp(name)
	case boardgame.TypeTimer:
		return b.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (b *__BaseReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (b *__BaseReader) SetProp(name string, value interface{}) error {
	props := b.Props()
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
		return b.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *__BaseReader) ConfigureProp(name string, value interface{}) error {
	props := b.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return b.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return b.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return b.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return b.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return b.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return b.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return b.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return b.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return b.SetPlayerIndexProp(name, val)
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return b.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return b.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return b.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if b.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return b.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return b.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (b *__BaseReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *__BaseReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (b *__BaseReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (b *__BaseReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (b *__BaseReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (b *__BaseReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (b *__BaseReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (b *__BaseReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (b *__BaseReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *__BaseReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (b *__BaseReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (b *__BaseReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (b *__BaseReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (b *__BaseReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (b *__BaseReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (b *__BaseReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (b *__BaseReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (b *__BaseReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (b *__BaseReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *__BaseReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (b *__BaseReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *__BaseReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (b *__BaseReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (b *__BaseReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (b *__BaseReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (b *__BaseReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (b *__BaseReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (b *__BaseReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (b *__BaseReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *__BaseReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (b *__BaseReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (b *__BaseReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (b *Base) Reader() boardgame.PropertyReader {
	return &__BaseReader{b}
}

func (b *Base) ReadSetter() boardgame.PropertyReadSetter {
	return &__BaseReader{b}
}

func (b *Base) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__BaseReader{b}
}

// Implementation for CollectCountComponents

var __CollectCountComponentsReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __CollectCountComponentsReader struct {
	data *CollectCountComponents
}

func (c *__CollectCountComponentsReader) Props() map[string]boardgame.PropertyType {
	return __CollectCountComponentsReaderProps
}

func (c *__CollectCountComponentsReader) Prop(name string) (interface{}, error) {
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

func (c *__CollectCountComponentsReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (c *__CollectCountComponentsReader) SetProp(name string, value interface{}) error {
	props := c.Props()
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
		return c.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return c.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return c.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return c.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return c.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return c.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (c *__CollectCountComponentsReader) ConfigureProp(name string, value interface{}) error {
	props := c.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return c.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return c.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return c.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return c.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return c.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return c.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return c.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return c.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexProp(name, val)
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return c.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return c.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return c.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return c.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return c.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (c *__CollectCountComponentsReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *__CollectCountComponentsReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (c *__CollectCountComponentsReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (c *__CollectCountComponentsReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *__CollectCountComponentsReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (c *__CollectCountComponentsReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (c *__CollectCountComponentsReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (c *__CollectCountComponentsReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (c *__CollectCountComponentsReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *__CollectCountComponentsReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (c *__CollectCountComponentsReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (c *__CollectCountComponentsReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *__CollectCountComponentsReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (c *__CollectCountComponentsReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (c *__CollectCountComponentsReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (c *__CollectCountComponentsReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (c *__CollectCountComponentsReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (c *__CollectCountComponentsReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (c *__CollectCountComponentsReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *__CollectCountComponentsReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *__CollectCountComponentsReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *__CollectCountComponentsReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (c *__CollectCountComponentsReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (c *__CollectCountComponentsReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *__CollectCountComponentsReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (c *__CollectCountComponentsReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (c *__CollectCountComponentsReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (c *__CollectCountComponentsReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (c *__CollectCountComponentsReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (c *__CollectCountComponentsReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (c *__CollectCountComponentsReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (c *__CollectCountComponentsReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (c *CollectCountComponents) Reader() boardgame.PropertyReader {
	return &__CollectCountComponentsReader{c}
}

func (c *CollectCountComponents) ReadSetter() boardgame.PropertyReadSetter {
	return &__CollectCountComponentsReader{c}
}

func (c *CollectCountComponents) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__CollectCountComponentsReader{c}
}

// Implementation for CollectComponentsUntilPlayerCountLeft

var __CollectComponentsUntilPlayerCountLeftReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __CollectComponentsUntilPlayerCountLeftReader struct {
	data *CollectComponentsUntilPlayerCountLeft
}

func (c *__CollectComponentsUntilPlayerCountLeftReader) Props() map[string]boardgame.PropertyType {
	return __CollectComponentsUntilPlayerCountLeftReaderProps
}

func (c *__CollectComponentsUntilPlayerCountLeftReader) Prop(name string) (interface{}, error) {
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

func (c *__CollectComponentsUntilPlayerCountLeftReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (c *__CollectComponentsUntilPlayerCountLeftReader) SetProp(name string, value interface{}) error {
	props := c.Props()
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
		return c.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return c.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return c.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return c.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return c.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return c.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (c *__CollectComponentsUntilPlayerCountLeftReader) ConfigureProp(name string, value interface{}) error {
	props := c.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return c.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return c.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return c.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return c.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return c.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return c.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return c.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return c.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexProp(name, val)
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return c.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return c.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return c.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return c.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return c.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (c *__CollectComponentsUntilPlayerCountLeftReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (c *__CollectComponentsUntilPlayerCountLeftReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (c *CollectComponentsUntilPlayerCountLeft) Reader() boardgame.PropertyReader {
	return &__CollectComponentsUntilPlayerCountLeftReader{c}
}

func (c *CollectComponentsUntilPlayerCountLeft) ReadSetter() boardgame.PropertyReadSetter {
	return &__CollectComponentsUntilPlayerCountLeftReader{c}
}

func (c *CollectComponentsUntilPlayerCountLeft) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__CollectComponentsUntilPlayerCountLeftReader{c}
}

// Implementation for CollectComponentsUntilGameCountReached

var __CollectComponentsUntilGameCountReachedReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __CollectComponentsUntilGameCountReachedReader struct {
	data *CollectComponentsUntilGameCountReached
}

func (c *__CollectComponentsUntilGameCountReachedReader) Props() map[string]boardgame.PropertyType {
	return __CollectComponentsUntilGameCountReachedReaderProps
}

func (c *__CollectComponentsUntilGameCountReachedReader) Prop(name string) (interface{}, error) {
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

func (c *__CollectComponentsUntilGameCountReachedReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (c *__CollectComponentsUntilGameCountReachedReader) SetProp(name string, value interface{}) error {
	props := c.Props()
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
		return c.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return c.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return c.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return c.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return c.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return c.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (c *__CollectComponentsUntilGameCountReachedReader) ConfigureProp(name string, value interface{}) error {
	props := c.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return c.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return c.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return c.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return c.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return c.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return c.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return c.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return c.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexProp(name, val)
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return c.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return c.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return c.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return c.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return c.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (c *__CollectComponentsUntilGameCountReachedReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (c *__CollectComponentsUntilGameCountReachedReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (c *CollectComponentsUntilGameCountReached) Reader() boardgame.PropertyReader {
	return &__CollectComponentsUntilGameCountReachedReader{c}
}

func (c *CollectComponentsUntilGameCountReached) ReadSetter() boardgame.PropertyReadSetter {
	return &__CollectComponentsUntilGameCountReachedReader{c}
}

func (c *CollectComponentsUntilGameCountReached) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__CollectComponentsUntilGameCountReachedReader{c}
}

// Implementation for CurrentPlayer

var __CurrentPlayerReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"TargetPlayerIndex": boardgame.TypePlayerIndex,
}

type __CurrentPlayerReader struct {
	data *CurrentPlayer
}

func (c *__CurrentPlayerReader) Props() map[string]boardgame.PropertyType {
	return __CurrentPlayerReaderProps
}

func (c *__CurrentPlayerReader) Prop(name string) (interface{}, error) {
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

func (c *__CurrentPlayerReader) PropMutable(name string) bool {
	switch name {
	case "TargetPlayerIndex":
		return true
	}

	return false
}

func (c *__CurrentPlayerReader) SetProp(name string, value interface{}) error {
	props := c.Props()
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
		return c.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return c.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return c.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return c.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return c.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return c.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (c *__CurrentPlayerReader) ConfigureProp(name string, value interface{}) error {
	props := c.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return c.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return c.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return c.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return c.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return c.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return c.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return c.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return c.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexProp(name, val)
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
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return c.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return c.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return c.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return c.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return c.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (c *__CurrentPlayerReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *__CurrentPlayerReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (c *__CurrentPlayerReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (c *__CurrentPlayerReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *__CurrentPlayerReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (c *__CurrentPlayerReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (c *__CurrentPlayerReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (c *__CurrentPlayerReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (c *__CurrentPlayerReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *__CurrentPlayerReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (c *__CurrentPlayerReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (c *__CurrentPlayerReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *__CurrentPlayerReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (c *__CurrentPlayerReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (c *__CurrentPlayerReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (c *__CurrentPlayerReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (c *__CurrentPlayerReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "TargetPlayerIndex":
		return c.data.TargetPlayerIndex, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (c *__CurrentPlayerReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "TargetPlayerIndex":
		c.data.TargetPlayerIndex = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (c *__CurrentPlayerReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *__CurrentPlayerReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *__CurrentPlayerReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *__CurrentPlayerReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (c *__CurrentPlayerReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (c *__CurrentPlayerReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *__CurrentPlayerReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (c *__CurrentPlayerReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (c *__CurrentPlayerReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (c *__CurrentPlayerReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (c *__CurrentPlayerReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (c *__CurrentPlayerReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (c *__CurrentPlayerReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (c *__CurrentPlayerReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (c *CurrentPlayer) Reader() boardgame.PropertyReader {
	return &__CurrentPlayerReader{c}
}

func (c *CurrentPlayer) ReadSetter() boardgame.PropertyReadSetter {
	return &__CurrentPlayerReader{c}
}

func (c *CurrentPlayer) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__CurrentPlayerReader{c}
}

// Implementation for DealCountComponents

var __DealCountComponentsReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __DealCountComponentsReader struct {
	data *DealCountComponents
}

func (d *__DealCountComponentsReader) Props() map[string]boardgame.PropertyType {
	return __DealCountComponentsReaderProps
}

func (d *__DealCountComponentsReader) Prop(name string) (interface{}, error) {
	props := d.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return d.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return d.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return d.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return d.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return d.IntProp(name)
	case boardgame.TypeIntSlice:
		return d.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return d.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return d.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return d.ImmutableStackProp(name)
	case boardgame.TypeString:
		return d.StringProp(name)
	case boardgame.TypeStringSlice:
		return d.StringSliceProp(name)
	case boardgame.TypeTimer:
		return d.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (d *__DealCountComponentsReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (d *__DealCountComponentsReader) SetProp(name string, value interface{}) error {
	props := d.Props()
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
		return d.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return d.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return d.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return d.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return d.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return d.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (d *__DealCountComponentsReader) ConfigureProp(name string, value interface{}) error {
	props := d.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return d.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return d.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return d.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return d.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return d.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return d.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return d.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return d.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return d.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return d.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return d.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return d.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return d.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return d.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (d *__DealCountComponentsReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *__DealCountComponentsReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (d *__DealCountComponentsReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (d *__DealCountComponentsReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *__DealCountComponentsReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (d *__DealCountComponentsReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (d *__DealCountComponentsReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (d *__DealCountComponentsReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (d *__DealCountComponentsReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *__DealCountComponentsReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (d *__DealCountComponentsReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (d *__DealCountComponentsReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *__DealCountComponentsReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (d *__DealCountComponentsReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (d *__DealCountComponentsReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (d *__DealCountComponentsReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (d *__DealCountComponentsReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (d *__DealCountComponentsReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (d *__DealCountComponentsReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *__DealCountComponentsReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *__DealCountComponentsReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *__DealCountComponentsReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (d *__DealCountComponentsReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (d *__DealCountComponentsReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *__DealCountComponentsReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (d *__DealCountComponentsReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (d *__DealCountComponentsReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (d *__DealCountComponentsReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (d *__DealCountComponentsReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (d *__DealCountComponentsReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (d *__DealCountComponentsReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (d *__DealCountComponentsReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (d *DealCountComponents) Reader() boardgame.PropertyReader {
	return &__DealCountComponentsReader{d}
}

func (d *DealCountComponents) ReadSetter() boardgame.PropertyReadSetter {
	return &__DealCountComponentsReader{d}
}

func (d *DealCountComponents) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__DealCountComponentsReader{d}
}

// Implementation for DealComponentsUntilPlayerCountReached

var __DealComponentsUntilPlayerCountReachedReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __DealComponentsUntilPlayerCountReachedReader struct {
	data *DealComponentsUntilPlayerCountReached
}

func (d *__DealComponentsUntilPlayerCountReachedReader) Props() map[string]boardgame.PropertyType {
	return __DealComponentsUntilPlayerCountReachedReaderProps
}

func (d *__DealComponentsUntilPlayerCountReachedReader) Prop(name string) (interface{}, error) {
	props := d.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return d.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return d.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return d.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return d.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return d.IntProp(name)
	case boardgame.TypeIntSlice:
		return d.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return d.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return d.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return d.ImmutableStackProp(name)
	case boardgame.TypeString:
		return d.StringProp(name)
	case boardgame.TypeStringSlice:
		return d.StringSliceProp(name)
	case boardgame.TypeTimer:
		return d.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (d *__DealComponentsUntilPlayerCountReachedReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (d *__DealComponentsUntilPlayerCountReachedReader) SetProp(name string, value interface{}) error {
	props := d.Props()
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
		return d.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return d.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return d.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return d.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return d.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return d.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (d *__DealComponentsUntilPlayerCountReachedReader) ConfigureProp(name string, value interface{}) error {
	props := d.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return d.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return d.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return d.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return d.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return d.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return d.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return d.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return d.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return d.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return d.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return d.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return d.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return d.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return d.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (d *__DealComponentsUntilPlayerCountReachedReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (d *__DealComponentsUntilPlayerCountReachedReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (d *DealComponentsUntilPlayerCountReached) Reader() boardgame.PropertyReader {
	return &__DealComponentsUntilPlayerCountReachedReader{d}
}

func (d *DealComponentsUntilPlayerCountReached) ReadSetter() boardgame.PropertyReadSetter {
	return &__DealComponentsUntilPlayerCountReachedReader{d}
}

func (d *DealComponentsUntilPlayerCountReached) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__DealComponentsUntilPlayerCountReachedReader{d}
}

// Implementation for DealComponentsUntilGameCountLeft

var __DealComponentsUntilGameCountLeftReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __DealComponentsUntilGameCountLeftReader struct {
	data *DealComponentsUntilGameCountLeft
}

func (d *__DealComponentsUntilGameCountLeftReader) Props() map[string]boardgame.PropertyType {
	return __DealComponentsUntilGameCountLeftReaderProps
}

func (d *__DealComponentsUntilGameCountLeftReader) Prop(name string) (interface{}, error) {
	props := d.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return d.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return d.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return d.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return d.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return d.IntProp(name)
	case boardgame.TypeIntSlice:
		return d.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return d.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return d.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return d.ImmutableStackProp(name)
	case boardgame.TypeString:
		return d.StringProp(name)
	case boardgame.TypeStringSlice:
		return d.StringSliceProp(name)
	case boardgame.TypeTimer:
		return d.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (d *__DealComponentsUntilGameCountLeftReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (d *__DealComponentsUntilGameCountLeftReader) SetProp(name string, value interface{}) error {
	props := d.Props()
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
		return d.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return d.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return d.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return d.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return d.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return d.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (d *__DealComponentsUntilGameCountLeftReader) ConfigureProp(name string, value interface{}) error {
	props := d.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return d.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return d.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return d.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return d.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return d.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return d.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return d.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return d.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return d.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return d.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return d.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return d.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return d.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return d.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (d *__DealComponentsUntilGameCountLeftReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (d *__DealComponentsUntilGameCountLeftReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (d *DealComponentsUntilGameCountLeft) Reader() boardgame.PropertyReader {
	return &__DealComponentsUntilGameCountLeftReader{d}
}

func (d *DealComponentsUntilGameCountLeft) ReadSetter() boardgame.PropertyReadSetter {
	return &__DealComponentsUntilGameCountLeftReader{d}
}

func (d *DealComponentsUntilGameCountLeft) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__DealComponentsUntilGameCountLeftReader{d}
}

// Implementation for DefaultComponent

var __DefaultComponentReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"ComponentIndex": boardgame.TypeInt,
}

type __DefaultComponentReader struct {
	data *DefaultComponent
}

func (d *__DefaultComponentReader) Props() map[string]boardgame.PropertyType {
	return __DefaultComponentReaderProps
}

func (d *__DefaultComponentReader) Prop(name string) (interface{}, error) {
	props := d.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return d.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return d.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return d.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return d.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return d.IntProp(name)
	case boardgame.TypeIntSlice:
		return d.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return d.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return d.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return d.ImmutableStackProp(name)
	case boardgame.TypeString:
		return d.StringProp(name)
	case boardgame.TypeStringSlice:
		return d.StringSliceProp(name)
	case boardgame.TypeTimer:
		return d.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (d *__DefaultComponentReader) PropMutable(name string) bool {
	switch name {
	case "ComponentIndex":
		return true
	}

	return false
}

func (d *__DefaultComponentReader) SetProp(name string, value interface{}) error {
	props := d.Props()
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
		return d.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return d.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return d.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return d.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return d.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return d.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (d *__DefaultComponentReader) ConfigureProp(name string, value interface{}) error {
	props := d.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return d.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return d.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return d.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return d.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return d.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return d.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return d.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return d.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return d.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return d.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return d.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return d.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return d.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return d.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (d *__DefaultComponentReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *__DefaultComponentReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (d *__DefaultComponentReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (d *__DefaultComponentReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *__DefaultComponentReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (d *__DefaultComponentReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (d *__DefaultComponentReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (d *__DefaultComponentReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (d *__DefaultComponentReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *__DefaultComponentReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (d *__DefaultComponentReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (d *__DefaultComponentReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *__DefaultComponentReader) IntProp(name string) (int, error) {

	switch name {
	case "ComponentIndex":
		return d.data.ComponentIndex, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (d *__DefaultComponentReader) SetIntProp(name string, value int) error {

	switch name {
	case "ComponentIndex":
		d.data.ComponentIndex = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (d *__DefaultComponentReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (d *__DefaultComponentReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (d *__DefaultComponentReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (d *__DefaultComponentReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (d *__DefaultComponentReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *__DefaultComponentReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *__DefaultComponentReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *__DefaultComponentReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (d *__DefaultComponentReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (d *__DefaultComponentReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *__DefaultComponentReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (d *__DefaultComponentReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (d *__DefaultComponentReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (d *__DefaultComponentReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (d *__DefaultComponentReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (d *__DefaultComponentReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (d *__DefaultComponentReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (d *__DefaultComponentReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (d *DefaultComponent) Reader() boardgame.PropertyReader {
	return &__DefaultComponentReader{d}
}

func (d *DefaultComponent) ReadSetter() boardgame.PropertyReadSetter {
	return &__DefaultComponentReader{d}
}

func (d *DefaultComponent) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__DefaultComponentReader{d}
}

// Implementation for FinishTurn

var __FinishTurnReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __FinishTurnReader struct {
	data *FinishTurn
}

func (f *__FinishTurnReader) Props() map[string]boardgame.PropertyType {
	return __FinishTurnReaderProps
}

func (f *__FinishTurnReader) Prop(name string) (interface{}, error) {
	props := f.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return f.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return f.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return f.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return f.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return f.IntProp(name)
	case boardgame.TypeIntSlice:
		return f.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return f.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return f.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return f.ImmutableStackProp(name)
	case boardgame.TypeString:
		return f.StringProp(name)
	case boardgame.TypeStringSlice:
		return f.StringSliceProp(name)
	case boardgame.TypeTimer:
		return f.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (f *__FinishTurnReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (f *__FinishTurnReader) SetProp(name string, value interface{}) error {
	props := f.Props()
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
		return f.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return f.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return f.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return f.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return f.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return f.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return f.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return f.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (f *__FinishTurnReader) ConfigureProp(name string, value interface{}) error {
	props := f.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if f.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return f.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return f.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return f.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return f.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if f.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return f.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return f.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return f.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return f.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return f.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return f.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if f.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return f.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return f.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return f.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return f.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if f.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return f.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return f.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (f *__FinishTurnReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (f *__FinishTurnReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (f *__FinishTurnReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (f *__FinishTurnReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (f *__FinishTurnReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (f *__FinishTurnReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (f *__FinishTurnReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (f *__FinishTurnReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (f *__FinishTurnReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (f *__FinishTurnReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (f *__FinishTurnReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (f *__FinishTurnReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (f *__FinishTurnReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (f *__FinishTurnReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (f *__FinishTurnReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (f *__FinishTurnReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (f *__FinishTurnReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (f *__FinishTurnReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (f *__FinishTurnReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (f *__FinishTurnReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (f *__FinishTurnReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (f *__FinishTurnReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (f *__FinishTurnReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (f *__FinishTurnReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (f *__FinishTurnReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (f *__FinishTurnReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (f *__FinishTurnReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (f *__FinishTurnReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (f *__FinishTurnReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (f *__FinishTurnReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (f *__FinishTurnReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (f *__FinishTurnReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (f *FinishTurn) Reader() boardgame.PropertyReader {
	return &__FinishTurnReader{f}
}

func (f *FinishTurn) ReadSetter() boardgame.PropertyReadSetter {
	return &__FinishTurnReader{f}
}

func (f *FinishTurn) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__FinishTurnReader{f}
}

// Implementation for FixUp

var __FixUpReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __FixUpReader struct {
	data *FixUp
}

func (f *__FixUpReader) Props() map[string]boardgame.PropertyType {
	return __FixUpReaderProps
}

func (f *__FixUpReader) Prop(name string) (interface{}, error) {
	props := f.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return f.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return f.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return f.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return f.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return f.IntProp(name)
	case boardgame.TypeIntSlice:
		return f.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return f.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return f.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return f.ImmutableStackProp(name)
	case boardgame.TypeString:
		return f.StringProp(name)
	case boardgame.TypeStringSlice:
		return f.StringSliceProp(name)
	case boardgame.TypeTimer:
		return f.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (f *__FixUpReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (f *__FixUpReader) SetProp(name string, value interface{}) error {
	props := f.Props()
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
		return f.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return f.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return f.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return f.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return f.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return f.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return f.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return f.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (f *__FixUpReader) ConfigureProp(name string, value interface{}) error {
	props := f.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if f.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return f.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return f.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return f.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return f.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if f.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return f.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return f.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return f.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return f.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return f.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return f.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if f.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return f.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return f.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return f.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return f.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if f.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return f.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return f.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (f *__FixUpReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (f *__FixUpReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (f *__FixUpReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (f *__FixUpReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (f *__FixUpReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (f *__FixUpReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (f *__FixUpReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (f *__FixUpReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (f *__FixUpReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (f *__FixUpReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (f *__FixUpReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (f *__FixUpReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (f *__FixUpReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (f *__FixUpReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (f *__FixUpReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (f *__FixUpReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (f *__FixUpReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (f *__FixUpReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (f *__FixUpReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (f *__FixUpReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (f *__FixUpReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (f *__FixUpReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (f *__FixUpReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (f *__FixUpReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (f *__FixUpReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (f *__FixUpReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (f *__FixUpReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (f *__FixUpReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (f *__FixUpReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (f *__FixUpReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (f *__FixUpReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (f *__FixUpReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (f *FixUp) Reader() boardgame.PropertyReader {
	return &__FixUpReader{f}
}

func (f *FixUp) ReadSetter() boardgame.PropertyReadSetter {
	return &__FixUpReader{f}
}

func (f *FixUp) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__FixUpReader{f}
}

// Implementation for FixUpMulti

var __FixUpMultiReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __FixUpMultiReader struct {
	data *FixUpMulti
}

func (f *__FixUpMultiReader) Props() map[string]boardgame.PropertyType {
	return __FixUpMultiReaderProps
}

func (f *__FixUpMultiReader) Prop(name string) (interface{}, error) {
	props := f.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return f.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return f.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return f.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return f.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return f.IntProp(name)
	case boardgame.TypeIntSlice:
		return f.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return f.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return f.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return f.ImmutableStackProp(name)
	case boardgame.TypeString:
		return f.StringProp(name)
	case boardgame.TypeStringSlice:
		return f.StringSliceProp(name)
	case boardgame.TypeTimer:
		return f.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (f *__FixUpMultiReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (f *__FixUpMultiReader) SetProp(name string, value interface{}) error {
	props := f.Props()
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
		return f.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return f.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return f.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return f.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return f.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return f.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return f.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return f.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (f *__FixUpMultiReader) ConfigureProp(name string, value interface{}) error {
	props := f.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if f.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return f.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return f.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return f.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return f.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if f.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return f.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return f.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return f.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return f.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return f.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return f.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if f.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return f.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return f.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return f.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return f.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if f.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return f.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return f.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (f *__FixUpMultiReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (f *__FixUpMultiReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (f *__FixUpMultiReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (f *__FixUpMultiReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (f *__FixUpMultiReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (f *__FixUpMultiReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (f *__FixUpMultiReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (f *__FixUpMultiReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (f *__FixUpMultiReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (f *__FixUpMultiReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (f *__FixUpMultiReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (f *__FixUpMultiReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (f *__FixUpMultiReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (f *__FixUpMultiReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (f *__FixUpMultiReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (f *__FixUpMultiReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (f *__FixUpMultiReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (f *__FixUpMultiReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (f *__FixUpMultiReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (f *__FixUpMultiReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (f *__FixUpMultiReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (f *__FixUpMultiReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (f *__FixUpMultiReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (f *__FixUpMultiReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (f *__FixUpMultiReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (f *__FixUpMultiReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (f *__FixUpMultiReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (f *__FixUpMultiReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (f *__FixUpMultiReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (f *__FixUpMultiReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (f *__FixUpMultiReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (f *__FixUpMultiReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (f *FixUpMulti) Reader() boardgame.PropertyReader {
	return &__FixUpMultiReader{f}
}

func (f *FixUpMulti) ReadSetter() boardgame.PropertyReadSetter {
	return &__FixUpMultiReader{f}
}

func (f *FixUpMulti) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__FixUpMultiReader{f}
}

// Implementation for MoveCountComponents

var __MoveCountComponentsReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __MoveCountComponentsReader struct {
	data *MoveCountComponents
}

func (m *__MoveCountComponentsReader) Props() map[string]boardgame.PropertyType {
	return __MoveCountComponentsReaderProps
}

func (m *__MoveCountComponentsReader) Prop(name string) (interface{}, error) {
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

func (m *__MoveCountComponentsReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *__MoveCountComponentsReader) SetProp(name string, value interface{}) error {
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

func (m *__MoveCountComponentsReader) ConfigureProp(name string, value interface{}) error {
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

func (m *__MoveCountComponentsReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__MoveCountComponentsReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__MoveCountComponentsReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__MoveCountComponentsReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__MoveCountComponentsReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__MoveCountComponentsReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__MoveCountComponentsReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__MoveCountComponentsReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__MoveCountComponentsReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__MoveCountComponentsReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *__MoveCountComponentsReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__MoveCountComponentsReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__MoveCountComponentsReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__MoveCountComponentsReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__MoveCountComponentsReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__MoveCountComponentsReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__MoveCountComponentsReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__MoveCountComponentsReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__MoveCountComponentsReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__MoveCountComponentsReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__MoveCountComponentsReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__MoveCountComponentsReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *__MoveCountComponentsReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__MoveCountComponentsReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__MoveCountComponentsReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__MoveCountComponentsReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__MoveCountComponentsReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__MoveCountComponentsReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__MoveCountComponentsReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__MoveCountComponentsReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__MoveCountComponentsReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__MoveCountComponentsReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *MoveCountComponents) Reader() boardgame.PropertyReader {
	return &__MoveCountComponentsReader{m}
}

func (m *MoveCountComponents) ReadSetter() boardgame.PropertyReadSetter {
	return &__MoveCountComponentsReader{m}
}

func (m *MoveCountComponents) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__MoveCountComponentsReader{m}
}

// Implementation for MoveComponentsUntilCountReached

var __MoveComponentsUntilCountReachedReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __MoveComponentsUntilCountReachedReader struct {
	data *MoveComponentsUntilCountReached
}

func (m *__MoveComponentsUntilCountReachedReader) Props() map[string]boardgame.PropertyType {
	return __MoveComponentsUntilCountReachedReaderProps
}

func (m *__MoveComponentsUntilCountReachedReader) Prop(name string) (interface{}, error) {
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

func (m *__MoveComponentsUntilCountReachedReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *__MoveComponentsUntilCountReachedReader) SetProp(name string, value interface{}) error {
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

func (m *__MoveComponentsUntilCountReachedReader) ConfigureProp(name string, value interface{}) error {
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

func (m *__MoveComponentsUntilCountReachedReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__MoveComponentsUntilCountReachedReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *MoveComponentsUntilCountReached) Reader() boardgame.PropertyReader {
	return &__MoveComponentsUntilCountReachedReader{m}
}

func (m *MoveComponentsUntilCountReached) ReadSetter() boardgame.PropertyReadSetter {
	return &__MoveComponentsUntilCountReachedReader{m}
}

func (m *MoveComponentsUntilCountReached) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__MoveComponentsUntilCountReachedReader{m}
}

// Implementation for MoveComponentsUntilCountLeft

var __MoveComponentsUntilCountLeftReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __MoveComponentsUntilCountLeftReader struct {
	data *MoveComponentsUntilCountLeft
}

func (m *__MoveComponentsUntilCountLeftReader) Props() map[string]boardgame.PropertyType {
	return __MoveComponentsUntilCountLeftReaderProps
}

func (m *__MoveComponentsUntilCountLeftReader) Prop(name string) (interface{}, error) {
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

func (m *__MoveComponentsUntilCountLeftReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *__MoveComponentsUntilCountLeftReader) SetProp(name string, value interface{}) error {
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

func (m *__MoveComponentsUntilCountLeftReader) ConfigureProp(name string, value interface{}) error {
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

func (m *__MoveComponentsUntilCountLeftReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__MoveComponentsUntilCountLeftReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *MoveComponentsUntilCountLeft) Reader() boardgame.PropertyReader {
	return &__MoveComponentsUntilCountLeftReader{m}
}

func (m *MoveComponentsUntilCountLeft) ReadSetter() boardgame.PropertyReadSetter {
	return &__MoveComponentsUntilCountLeftReader{m}
}

func (m *MoveComponentsUntilCountLeft) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__MoveComponentsUntilCountLeftReader{m}
}

// Implementation for NoOp

var __NoOpReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __NoOpReader struct {
	data *NoOp
}

func (n *__NoOpReader) Props() map[string]boardgame.PropertyType {
	return __NoOpReaderProps
}

func (n *__NoOpReader) Prop(name string) (interface{}, error) {
	props := n.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return n.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return n.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return n.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return n.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return n.IntProp(name)
	case boardgame.TypeIntSlice:
		return n.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return n.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return n.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return n.ImmutableStackProp(name)
	case boardgame.TypeString:
		return n.StringProp(name)
	case boardgame.TypeStringSlice:
		return n.StringSliceProp(name)
	case boardgame.TypeTimer:
		return n.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (n *__NoOpReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (n *__NoOpReader) SetProp(name string, value interface{}) error {
	props := n.Props()
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
		return n.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return n.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return n.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return n.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return n.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return n.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return n.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return n.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (n *__NoOpReader) ConfigureProp(name string, value interface{}) error {
	props := n.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if n.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return n.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return n.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return n.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return n.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if n.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return n.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return n.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return n.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return n.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return n.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return n.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if n.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return n.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return n.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return n.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return n.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if n.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return n.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return n.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (n *__NoOpReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (n *__NoOpReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (n *__NoOpReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (n *__NoOpReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (n *__NoOpReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (n *__NoOpReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (n *__NoOpReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (n *__NoOpReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (n *__NoOpReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (n *__NoOpReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (n *__NoOpReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (n *__NoOpReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (n *__NoOpReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (n *__NoOpReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (n *__NoOpReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (n *__NoOpReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (n *__NoOpReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (n *__NoOpReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (n *__NoOpReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (n *__NoOpReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (n *__NoOpReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (n *__NoOpReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (n *__NoOpReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (n *__NoOpReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (n *__NoOpReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (n *__NoOpReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (n *__NoOpReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (n *__NoOpReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (n *__NoOpReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (n *__NoOpReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (n *__NoOpReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (n *__NoOpReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (n *NoOp) Reader() boardgame.PropertyReader {
	return &__NoOpReader{n}
}

func (n *NoOp) ReadSetter() boardgame.PropertyReadSetter {
	return &__NoOpReader{n}
}

func (n *NoOp) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__NoOpReader{n}
}

// Implementation for Done

var __DoneReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __DoneReader struct {
	data *Done
}

func (d *__DoneReader) Props() map[string]boardgame.PropertyType {
	return __DoneReaderProps
}

func (d *__DoneReader) Prop(name string) (interface{}, error) {
	props := d.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return d.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return d.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return d.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return d.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return d.IntProp(name)
	case boardgame.TypeIntSlice:
		return d.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return d.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return d.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return d.ImmutableStackProp(name)
	case boardgame.TypeString:
		return d.StringProp(name)
	case boardgame.TypeStringSlice:
		return d.StringSliceProp(name)
	case boardgame.TypeTimer:
		return d.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (d *__DoneReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (d *__DoneReader) SetProp(name string, value interface{}) error {
	props := d.Props()
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
		return d.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return d.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return d.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return d.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return d.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return d.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (d *__DoneReader) ConfigureProp(name string, value interface{}) error {
	props := d.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return d.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return d.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return d.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return d.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return d.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return d.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return d.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return d.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return d.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return d.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return d.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return d.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return d.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return d.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (d *__DoneReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *__DoneReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (d *__DoneReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (d *__DoneReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *__DoneReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (d *__DoneReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (d *__DoneReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (d *__DoneReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (d *__DoneReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *__DoneReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (d *__DoneReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (d *__DoneReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *__DoneReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (d *__DoneReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (d *__DoneReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (d *__DoneReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (d *__DoneReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (d *__DoneReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (d *__DoneReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *__DoneReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *__DoneReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *__DoneReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (d *__DoneReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (d *__DoneReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *__DoneReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (d *__DoneReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (d *__DoneReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (d *__DoneReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (d *__DoneReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (d *__DoneReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (d *__DoneReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (d *__DoneReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (d *Done) Reader() boardgame.PropertyReader {
	return &__DoneReader{d}
}

func (d *Done) ReadSetter() boardgame.PropertyReadSetter {
	return &__DoneReader{d}
}

func (d *Done) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__DoneReader{d}
}

// Implementation for RoundRobin

var __RoundRobinReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __RoundRobinReader struct {
	data *RoundRobin
}

func (r *__RoundRobinReader) Props() map[string]boardgame.PropertyType {
	return __RoundRobinReaderProps
}

func (r *__RoundRobinReader) Prop(name string) (interface{}, error) {
	props := r.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return r.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return r.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return r.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return r.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return r.IntProp(name)
	case boardgame.TypeIntSlice:
		return r.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return r.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return r.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return r.ImmutableStackProp(name)
	case boardgame.TypeString:
		return r.StringProp(name)
	case boardgame.TypeStringSlice:
		return r.StringSliceProp(name)
	case boardgame.TypeTimer:
		return r.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (r *__RoundRobinReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (r *__RoundRobinReader) SetProp(name string, value interface{}) error {
	props := r.Props()
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
		return r.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return r.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return r.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return r.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return r.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return r.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return r.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return r.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (r *__RoundRobinReader) ConfigureProp(name string, value interface{}) error {
	props := r.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if r.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return r.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return r.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return r.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return r.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if r.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return r.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return r.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return r.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return r.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return r.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return r.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if r.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return r.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return r.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return r.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return r.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if r.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return r.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return r.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (r *__RoundRobinReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (r *__RoundRobinReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (r *__RoundRobinReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (r *__RoundRobinReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (r *__RoundRobinReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (r *__RoundRobinReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (r *__RoundRobinReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (r *__RoundRobinReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (r *__RoundRobinReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (r *__RoundRobinReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (r *__RoundRobinReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (r *__RoundRobinReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (r *__RoundRobinReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (r *__RoundRobinReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (r *__RoundRobinReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (r *__RoundRobinReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (r *__RoundRobinReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (r *__RoundRobinReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (r *__RoundRobinReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (r *__RoundRobinReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (r *__RoundRobinReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (r *__RoundRobinReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (r *__RoundRobinReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (r *__RoundRobinReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (r *__RoundRobinReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (r *__RoundRobinReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (r *__RoundRobinReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (r *__RoundRobinReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (r *__RoundRobinReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (r *__RoundRobinReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (r *__RoundRobinReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (r *__RoundRobinReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (r *RoundRobin) Reader() boardgame.PropertyReader {
	return &__RoundRobinReader{r}
}

func (r *RoundRobin) ReadSetter() boardgame.PropertyReadSetter {
	return &__RoundRobinReader{r}
}

func (r *RoundRobin) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__RoundRobinReader{r}
}

// Implementation for RoundRobinNumRounds

var __RoundRobinNumRoundsReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __RoundRobinNumRoundsReader struct {
	data *RoundRobinNumRounds
}

func (r *__RoundRobinNumRoundsReader) Props() map[string]boardgame.PropertyType {
	return __RoundRobinNumRoundsReaderProps
}

func (r *__RoundRobinNumRoundsReader) Prop(name string) (interface{}, error) {
	props := r.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return r.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return r.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return r.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return r.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return r.IntProp(name)
	case boardgame.TypeIntSlice:
		return r.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return r.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return r.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return r.ImmutableStackProp(name)
	case boardgame.TypeString:
		return r.StringProp(name)
	case boardgame.TypeStringSlice:
		return r.StringSliceProp(name)
	case boardgame.TypeTimer:
		return r.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (r *__RoundRobinNumRoundsReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (r *__RoundRobinNumRoundsReader) SetProp(name string, value interface{}) error {
	props := r.Props()
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
		return r.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return r.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return r.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return r.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return r.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return r.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return r.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return r.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (r *__RoundRobinNumRoundsReader) ConfigureProp(name string, value interface{}) error {
	props := r.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if r.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return r.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return r.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return r.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return r.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if r.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return r.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return r.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return r.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return r.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return r.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return r.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if r.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return r.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return r.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return r.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return r.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if r.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return r.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return r.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (r *__RoundRobinNumRoundsReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (r *__RoundRobinNumRoundsReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (r *RoundRobinNumRounds) Reader() boardgame.PropertyReader {
	return &__RoundRobinNumRoundsReader{r}
}

func (r *RoundRobinNumRounds) ReadSetter() boardgame.PropertyReadSetter {
	return &__RoundRobinNumRoundsReader{r}
}

func (r *RoundRobinNumRounds) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__RoundRobinNumRoundsReader{r}
}

// Implementation for ShuffleStack

var __ShuffleStackReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __ShuffleStackReader struct {
	data *ShuffleStack
}

func (s *__ShuffleStackReader) Props() map[string]boardgame.PropertyType {
	return __ShuffleStackReaderProps
}

func (s *__ShuffleStackReader) Prop(name string) (interface{}, error) {
	props := s.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return s.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return s.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return s.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return s.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return s.IntProp(name)
	case boardgame.TypeIntSlice:
		return s.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return s.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return s.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return s.ImmutableStackProp(name)
	case boardgame.TypeString:
		return s.StringProp(name)
	case boardgame.TypeStringSlice:
		return s.StringSliceProp(name)
	case boardgame.TypeTimer:
		return s.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (s *__ShuffleStackReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (s *__ShuffleStackReader) SetProp(name string, value interface{}) error {
	props := s.Props()
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
		return s.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return s.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return s.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return s.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return s.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return s.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return s.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return s.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (s *__ShuffleStackReader) ConfigureProp(name string, value interface{}) error {
	props := s.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if s.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return s.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return s.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return s.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return s.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if s.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return s.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return s.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return s.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return s.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return s.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return s.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if s.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return s.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return s.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return s.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return s.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if s.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return s.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return s.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (s *__ShuffleStackReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (s *__ShuffleStackReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (s *__ShuffleStackReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (s *__ShuffleStackReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (s *__ShuffleStackReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (s *__ShuffleStackReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (s *__ShuffleStackReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (s *__ShuffleStackReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (s *__ShuffleStackReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (s *__ShuffleStackReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (s *__ShuffleStackReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (s *__ShuffleStackReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (s *__ShuffleStackReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (s *__ShuffleStackReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (s *__ShuffleStackReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (s *__ShuffleStackReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (s *__ShuffleStackReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (s *__ShuffleStackReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (s *__ShuffleStackReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (s *__ShuffleStackReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (s *__ShuffleStackReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (s *__ShuffleStackReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (s *__ShuffleStackReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (s *__ShuffleStackReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (s *__ShuffleStackReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (s *__ShuffleStackReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (s *__ShuffleStackReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (s *__ShuffleStackReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (s *__ShuffleStackReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (s *__ShuffleStackReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (s *__ShuffleStackReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (s *__ShuffleStackReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (s *ShuffleStack) Reader() boardgame.PropertyReader {
	return &__ShuffleStackReader{s}
}

func (s *ShuffleStack) ReadSetter() boardgame.PropertyReadSetter {
	return &__ShuffleStackReader{s}
}

func (s *ShuffleStack) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__ShuffleStackReader{s}
}

// Implementation for StartPhase

var __StartPhaseReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{}

type __StartPhaseReader struct {
	data *StartPhase
}

func (s *__StartPhaseReader) Props() map[string]boardgame.PropertyType {
	return __StartPhaseReaderProps
}

func (s *__StartPhaseReader) Prop(name string) (interface{}, error) {
	props := s.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return s.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return s.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return s.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return s.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return s.IntProp(name)
	case boardgame.TypeIntSlice:
		return s.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return s.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return s.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return s.ImmutableStackProp(name)
	case boardgame.TypeString:
		return s.StringProp(name)
	case boardgame.TypeStringSlice:
		return s.StringSliceProp(name)
	case boardgame.TypeTimer:
		return s.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (s *__StartPhaseReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (s *__StartPhaseReader) SetProp(name string, value interface{}) error {
	props := s.Props()
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
		return s.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return s.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return s.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return s.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return s.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return s.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return s.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return s.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (s *__StartPhaseReader) ConfigureProp(name string, value interface{}) error {
	props := s.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if s.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return s.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return s.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return s.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return s.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if s.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return s.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return s.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return s.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return s.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return s.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return s.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if s.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return s.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return s.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return s.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return s.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if s.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return s.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return s.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (s *__StartPhaseReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (s *__StartPhaseReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (s *__StartPhaseReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (s *__StartPhaseReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (s *__StartPhaseReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (s *__StartPhaseReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (s *__StartPhaseReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (s *__StartPhaseReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (s *__StartPhaseReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (s *__StartPhaseReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (s *__StartPhaseReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (s *__StartPhaseReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (s *__StartPhaseReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (s *__StartPhaseReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (s *__StartPhaseReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (s *__StartPhaseReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (s *__StartPhaseReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (s *__StartPhaseReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (s *__StartPhaseReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (s *__StartPhaseReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (s *__StartPhaseReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (s *__StartPhaseReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (s *__StartPhaseReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (s *__StartPhaseReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (s *__StartPhaseReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (s *__StartPhaseReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (s *__StartPhaseReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (s *__StartPhaseReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (s *__StartPhaseReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (s *__StartPhaseReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (s *__StartPhaseReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (s *__StartPhaseReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (s *StartPhase) Reader() boardgame.PropertyReader {
	return &__StartPhaseReader{s}
}

func (s *StartPhase) ReadSetter() boardgame.PropertyReadSetter {
	return &__StartPhaseReader{s}
}

func (s *StartPhase) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__StartPhaseReader{s}
}
