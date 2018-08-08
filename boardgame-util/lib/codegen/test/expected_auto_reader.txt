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

// Implementation for myStruct

var __myStructReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"EnumVar":            boardgame.TypeEnum,
	"MyBool":             boardgame.TypeBool,
	"MyBoolSlice":        boardgame.TypeBoolSlice,
	"MyInt":              boardgame.TypeInt,
	"MyIntSlice":         boardgame.TypeIntSlice,
	"MyPlayerIndexSlice": boardgame.TypePlayerIndexSlice,
	"MySizedStack":       boardgame.TypeStack,
	"MyStringSlice":      boardgame.TypeStringSlice,
	"TheTimer":           boardgame.TypeTimer,
}

type __myStructReader struct {
	data *myStruct
}

func (m *__myStructReader) Props() map[string]boardgame.PropertyType {
	return __myStructReaderProps
}

func (m *__myStructReader) Prop(name string) (interface{}, error) {
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

func (m *__myStructReader) PropMutable(name string) bool {
	switch name {
	case "EnumVar":
		return true
	case "MyBool":
		return true
	case "MyBoolSlice":
		return true
	case "MyInt":
		return true
	case "MyIntSlice":
		return true
	case "MyPlayerIndexSlice":
		return true
	case "MySizedStack":
		return true
	case "MyStringSlice":
		return true
	case "TheTimer":
		return true
	}

	return false
}

func (m *__myStructReader) SetProp(name string, value interface{}) error {
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

func (m *__myStructReader) ConfigureProp(name string, value interface{}) error {
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

func (m *__myStructReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__myStructReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__myStructReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__myStructReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__myStructReader) BoolProp(name string) (bool, error) {

	switch name {
	case "MyBool":
		return m.data.MyBool, nil

	}

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__myStructReader) SetBoolProp(name string, value bool) error {

	switch name {
	case "MyBool":
		m.data.MyBool = value
		return nil

	}

	return errors.New("No such Bool prop: " + name)

}

func (m *__myStructReader) BoolSliceProp(name string) ([]bool, error) {

	switch name {
	case "MyBoolSlice":
		return m.data.MyBoolSlice, nil

	}

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__myStructReader) SetBoolSliceProp(name string, value []bool) error {

	switch name {
	case "MyBoolSlice":
		m.data.MyBoolSlice = value
		return nil

	}

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__myStructReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "EnumVar":
		return m.data.EnumVar, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__myStructReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "EnumVar":
		m.data.EnumVar = value
		return nil

	}

	return errors.New("No such Enum prop: " + name)

}

func (m *__myStructReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "EnumVar":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__myStructReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "EnumVar":
		return m.data.EnumVar, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__myStructReader) IntProp(name string) (int, error) {

	switch name {
	case "MyInt":
		return m.data.MyInt, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__myStructReader) SetIntProp(name string, value int) error {

	switch name {
	case "MyInt":
		m.data.MyInt = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (m *__myStructReader) IntSliceProp(name string) ([]int, error) {

	switch name {
	case "MyIntSlice":
		return m.data.MyIntSlice, nil

	}

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__myStructReader) SetIntSliceProp(name string, value []int) error {

	switch name {
	case "MyIntSlice":
		m.data.MyIntSlice = value
		return nil

	}

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__myStructReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__myStructReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__myStructReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	switch name {
	case "MyPlayerIndexSlice":
		return m.data.MyPlayerIndexSlice, nil

	}

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__myStructReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	switch name {
	case "MyPlayerIndexSlice":
		m.data.MyPlayerIndexSlice = value
		return nil

	}

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__myStructReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "MySizedStack":
		return m.data.MySizedStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__myStructReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "MySizedStack":
		m.data.MySizedStack = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (m *__myStructReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "MySizedStack":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__myStructReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "MySizedStack":
		return m.data.MySizedStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__myStructReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__myStructReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__myStructReader) StringSliceProp(name string) ([]string, error) {

	switch name {
	case "MyStringSlice":
		return m.data.MyStringSlice, nil

	}

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__myStructReader) SetStringSliceProp(name string, value []string) error {

	switch name {
	case "MyStringSlice":
		m.data.MyStringSlice = value
		return nil

	}

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__myStructReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	switch name {
	case "TheTimer":
		return m.data.TheTimer, nil

	}

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__myStructReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	switch name {
	case "TheTimer":
		m.data.TheTimer = value
		return nil

	}

	return errors.New("No such Timer prop: " + name)

}

func (m *__myStructReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	switch name {
	case "TheTimer":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__myStructReader) TimerProp(name string) (boardgame.Timer, error) {

	switch name {
	case "TheTimer":
		return m.data.TheTimer, nil

	}

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *myStruct) Reader() boardgame.PropertyReader {
	return &__myStructReader{m}
}

func (m *myStruct) ReadSetter() boardgame.PropertyReadSetter {
	return &__myStructReader{m}
}

func (m *myStruct) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__myStructReader{m}
}

// Implementation for roundRobinStruct

var __roundRobinStructReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"MyBool":          boardgame.TypeBool,
	"RRHasStarted":    boardgame.TypeBool,
	"RRLastPlayer":    boardgame.TypePlayerIndex,
	"RRRoundCount":    boardgame.TypeInt,
	"RRStarterPlayer": boardgame.TypePlayerIndex,
}

type __roundRobinStructReader struct {
	data *roundRobinStruct
}

func (r *__roundRobinStructReader) Props() map[string]boardgame.PropertyType {
	return __roundRobinStructReaderProps
}

func (r *__roundRobinStructReader) Prop(name string) (interface{}, error) {
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

func (r *__roundRobinStructReader) PropMutable(name string) bool {
	switch name {
	case "MyBool":
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

func (r *__roundRobinStructReader) SetProp(name string, value interface{}) error {
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

func (r *__roundRobinStructReader) ConfigureProp(name string, value interface{}) error {
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

func (r *__roundRobinStructReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (r *__roundRobinStructReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (r *__roundRobinStructReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (r *__roundRobinStructReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (r *__roundRobinStructReader) BoolProp(name string) (bool, error) {

	switch name {
	case "MyBool":
		return r.data.MyBool, nil
	case "RRHasStarted":
		return r.data.RRHasStarted, nil

	}

	return false, errors.New("No such Bool prop: " + name)

}

func (r *__roundRobinStructReader) SetBoolProp(name string, value bool) error {

	switch name {
	case "MyBool":
		r.data.MyBool = value
		return nil
	case "RRHasStarted":
		r.data.RRHasStarted = value
		return nil

	}

	return errors.New("No such Bool prop: " + name)

}

func (r *__roundRobinStructReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (r *__roundRobinStructReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (r *__roundRobinStructReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (r *__roundRobinStructReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (r *__roundRobinStructReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (r *__roundRobinStructReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (r *__roundRobinStructReader) IntProp(name string) (int, error) {

	switch name {
	case "RRRoundCount":
		return r.data.RRRoundCount, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (r *__roundRobinStructReader) SetIntProp(name string, value int) error {

	switch name {
	case "RRRoundCount":
		r.data.RRRoundCount = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (r *__roundRobinStructReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (r *__roundRobinStructReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (r *__roundRobinStructReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "RRLastPlayer":
		return r.data.RRLastPlayer, nil
	case "RRStarterPlayer":
		return r.data.RRStarterPlayer, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (r *__roundRobinStructReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "RRLastPlayer":
		r.data.RRLastPlayer = value
		return nil
	case "RRStarterPlayer":
		r.data.RRStarterPlayer = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (r *__roundRobinStructReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (r *__roundRobinStructReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (r *__roundRobinStructReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (r *__roundRobinStructReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (r *__roundRobinStructReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (r *__roundRobinStructReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (r *__roundRobinStructReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (r *__roundRobinStructReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (r *__roundRobinStructReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (r *__roundRobinStructReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (r *__roundRobinStructReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (r *__roundRobinStructReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (r *__roundRobinStructReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (r *__roundRobinStructReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (r *roundRobinStruct) Reader() boardgame.PropertyReader {
	return &__roundRobinStructReader{r}
}

func (r *roundRobinStruct) ReadSetter() boardgame.PropertyReadSetter {
	return &__roundRobinStructReader{r}
}

func (r *roundRobinStruct) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__roundRobinStructReader{r}
}

// Implementation for structWithManyKeys

var __structWithManyKeysReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"A": boardgame.TypeInt,
	"B": boardgame.TypeInt,
	"C": boardgame.TypeInt,
	"D": boardgame.TypeInt,
	"E": boardgame.TypeInt,
	"F": boardgame.TypeInt,
	"G": boardgame.TypeInt,
	"H": boardgame.TypeInt,
	"I": boardgame.TypeInt,
}

type __structWithManyKeysReader struct {
	data *structWithManyKeys
}

func (s *__structWithManyKeysReader) Props() map[string]boardgame.PropertyType {
	return __structWithManyKeysReaderProps
}

func (s *__structWithManyKeysReader) Prop(name string) (interface{}, error) {
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

func (s *__structWithManyKeysReader) PropMutable(name string) bool {
	switch name {
	case "A":
		return true
	case "B":
		return true
	case "C":
		return true
	case "D":
		return true
	case "E":
		return true
	case "F":
		return true
	case "G":
		return true
	case "H":
		return true
	case "I":
		return true
	}

	return false
}

func (s *__structWithManyKeysReader) SetProp(name string, value interface{}) error {
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

func (s *__structWithManyKeysReader) ConfigureProp(name string, value interface{}) error {
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

func (s *__structWithManyKeysReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (s *__structWithManyKeysReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (s *__structWithManyKeysReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (s *__structWithManyKeysReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (s *__structWithManyKeysReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (s *__structWithManyKeysReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (s *__structWithManyKeysReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (s *__structWithManyKeysReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (s *__structWithManyKeysReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (s *__structWithManyKeysReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (s *__structWithManyKeysReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (s *__structWithManyKeysReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (s *__structWithManyKeysReader) IntProp(name string) (int, error) {

	switch name {
	case "A":
		return s.data.A, nil
	case "B":
		return s.data.B, nil
	case "C":
		return s.data.C, nil
	case "D":
		return s.data.D, nil
	case "E":
		return s.data.E, nil
	case "F":
		return s.data.F, nil
	case "G":
		return s.data.G, nil
	case "H":
		return s.data.H, nil
	case "I":
		return s.data.I, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (s *__structWithManyKeysReader) SetIntProp(name string, value int) error {

	switch name {
	case "A":
		s.data.A = value
		return nil
	case "B":
		s.data.B = value
		return nil
	case "C":
		s.data.C = value
		return nil
	case "D":
		s.data.D = value
		return nil
	case "E":
		s.data.E = value
		return nil
	case "F":
		s.data.F = value
		return nil
	case "G":
		s.data.G = value
		return nil
	case "H":
		s.data.H = value
		return nil
	case "I":
		s.data.I = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (s *__structWithManyKeysReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (s *__structWithManyKeysReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (s *__structWithManyKeysReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (s *__structWithManyKeysReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (s *__structWithManyKeysReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (s *__structWithManyKeysReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (s *__structWithManyKeysReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (s *__structWithManyKeysReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (s *__structWithManyKeysReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (s *__structWithManyKeysReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (s *__structWithManyKeysReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (s *__structWithManyKeysReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (s *__structWithManyKeysReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (s *__structWithManyKeysReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (s *__structWithManyKeysReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (s *__structWithManyKeysReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (s *__structWithManyKeysReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (s *__structWithManyKeysReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (s *structWithManyKeys) Reader() boardgame.PropertyReader {
	return &__structWithManyKeysReader{s}
}

func (s *structWithManyKeys) ReadSetter() boardgame.PropertyReadSetter {
	return &__structWithManyKeysReader{s}
}

func (s *structWithManyKeys) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__structWithManyKeysReader{s}
}

// Implementation for embeddedStruct

var __embeddedStructReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"MyInt":             boardgame.TypeInt,
	"TargetPlayerIndex": boardgame.TypePlayerIndex,
}

type __embeddedStructReader struct {
	data *embeddedStruct
}

func (e *__embeddedStructReader) Props() map[string]boardgame.PropertyType {
	return __embeddedStructReaderProps
}

func (e *__embeddedStructReader) Prop(name string) (interface{}, error) {
	props := e.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return e.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return e.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return e.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return e.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return e.IntProp(name)
	case boardgame.TypeIntSlice:
		return e.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return e.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return e.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return e.ImmutableStackProp(name)
	case boardgame.TypeString:
		return e.StringProp(name)
	case boardgame.TypeStringSlice:
		return e.StringSliceProp(name)
	case boardgame.TypeTimer:
		return e.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (e *__embeddedStructReader) PropMutable(name string) bool {
	switch name {
	case "MyInt":
		return true
	case "TargetPlayerIndex":
		return true
	}

	return false
}

func (e *__embeddedStructReader) SetProp(name string, value interface{}) error {
	props := e.Props()
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
		return e.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return e.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return e.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return e.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return e.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return e.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return e.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return e.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (e *__embeddedStructReader) ConfigureProp(name string, value interface{}) error {
	props := e.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if e.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return e.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return e.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return e.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return e.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if e.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return e.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return e.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return e.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return e.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return e.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return e.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if e.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return e.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return e.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return e.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return e.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if e.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return e.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return e.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (e *__embeddedStructReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (e *__embeddedStructReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (e *__embeddedStructReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (e *__embeddedStructReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (e *__embeddedStructReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (e *__embeddedStructReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (e *__embeddedStructReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (e *__embeddedStructReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (e *__embeddedStructReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (e *__embeddedStructReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (e *__embeddedStructReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (e *__embeddedStructReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (e *__embeddedStructReader) IntProp(name string) (int, error) {

	switch name {
	case "MyInt":
		return e.data.MyInt, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (e *__embeddedStructReader) SetIntProp(name string, value int) error {

	switch name {
	case "MyInt":
		e.data.MyInt = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (e *__embeddedStructReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (e *__embeddedStructReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (e *__embeddedStructReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "TargetPlayerIndex":
		return e.data.TargetPlayerIndex, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (e *__embeddedStructReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "TargetPlayerIndex":
		e.data.TargetPlayerIndex = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (e *__embeddedStructReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (e *__embeddedStructReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (e *__embeddedStructReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (e *__embeddedStructReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (e *__embeddedStructReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (e *__embeddedStructReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (e *__embeddedStructReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (e *__embeddedStructReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (e *__embeddedStructReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (e *__embeddedStructReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (e *__embeddedStructReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (e *__embeddedStructReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (e *__embeddedStructReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (e *__embeddedStructReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (e *embeddedStruct) Reader() boardgame.PropertyReader {
	return &__embeddedStructReader{e}
}

func (e *embeddedStruct) ReadSetter() boardgame.PropertyReadSetter {
	return &__embeddedStructReader{e}
}

func (e *embeddedStruct) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__embeddedStructReader{e}
}

// Implementation for doubleEmbeddedStruct

var __doubleEmbeddedStructReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"MyInt":             boardgame.TypeInt,
	"TargetPlayerIndex": boardgame.TypePlayerIndex,
}

type __doubleEmbeddedStructReader struct {
	data *doubleEmbeddedStruct
}

func (d *__doubleEmbeddedStructReader) Props() map[string]boardgame.PropertyType {
	return __doubleEmbeddedStructReaderProps
}

func (d *__doubleEmbeddedStructReader) Prop(name string) (interface{}, error) {
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

func (d *__doubleEmbeddedStructReader) PropMutable(name string) bool {
	switch name {
	case "MyInt":
		return true
	case "TargetPlayerIndex":
		return true
	}

	return false
}

func (d *__doubleEmbeddedStructReader) SetProp(name string, value interface{}) error {
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

func (d *__doubleEmbeddedStructReader) ConfigureProp(name string, value interface{}) error {
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

func (d *__doubleEmbeddedStructReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *__doubleEmbeddedStructReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (d *__doubleEmbeddedStructReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (d *__doubleEmbeddedStructReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *__doubleEmbeddedStructReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (d *__doubleEmbeddedStructReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (d *__doubleEmbeddedStructReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (d *__doubleEmbeddedStructReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (d *__doubleEmbeddedStructReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *__doubleEmbeddedStructReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (d *__doubleEmbeddedStructReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (d *__doubleEmbeddedStructReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *__doubleEmbeddedStructReader) IntProp(name string) (int, error) {

	switch name {
	case "MyInt":
		return d.data.MyInt, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (d *__doubleEmbeddedStructReader) SetIntProp(name string, value int) error {

	switch name {
	case "MyInt":
		d.data.MyInt = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (d *__doubleEmbeddedStructReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (d *__doubleEmbeddedStructReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (d *__doubleEmbeddedStructReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "TargetPlayerIndex":
		return d.data.TargetPlayerIndex, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (d *__doubleEmbeddedStructReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "TargetPlayerIndex":
		d.data.TargetPlayerIndex = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (d *__doubleEmbeddedStructReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *__doubleEmbeddedStructReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *__doubleEmbeddedStructReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *__doubleEmbeddedStructReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (d *__doubleEmbeddedStructReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (d *__doubleEmbeddedStructReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *__doubleEmbeddedStructReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (d *__doubleEmbeddedStructReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (d *__doubleEmbeddedStructReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (d *__doubleEmbeddedStructReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (d *__doubleEmbeddedStructReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (d *__doubleEmbeddedStructReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (d *__doubleEmbeddedStructReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (d *__doubleEmbeddedStructReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (d *doubleEmbeddedStruct) Reader() boardgame.PropertyReader {
	return &__doubleEmbeddedStructReader{d}
}

func (d *doubleEmbeddedStruct) ReadSetter() boardgame.PropertyReadSetter {
	return &__doubleEmbeddedStructReader{d}
}

func (d *doubleEmbeddedStruct) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__doubleEmbeddedStructReader{d}
}

// Implementation for myOtherStruct

var __myOtherStructReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"MyGrowableStack": boardgame.TypeStack,
	"ThePlayerIndex":  boardgame.TypePlayerIndex,
}

type __myOtherStructReader struct {
	data *myOtherStruct
}

func (m *__myOtherStructReader) Props() map[string]boardgame.PropertyType {
	return __myOtherStructReaderProps
}

func (m *__myOtherStructReader) Prop(name string) (interface{}, error) {
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

func (m *__myOtherStructReader) PropMutable(name string) bool {
	switch name {
	case "MyGrowableStack":
		return true
	case "ThePlayerIndex":
		return true
	}

	return false
}

func (m *__myOtherStructReader) SetProp(name string, value interface{}) error {
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

func (m *__myOtherStructReader) ConfigureProp(name string, value interface{}) error {
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

func (m *__myOtherStructReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__myOtherStructReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__myOtherStructReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__myOtherStructReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__myOtherStructReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__myOtherStructReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__myOtherStructReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__myOtherStructReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__myOtherStructReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__myOtherStructReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *__myOtherStructReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__myOtherStructReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__myOtherStructReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__myOtherStructReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__myOtherStructReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__myOtherStructReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__myOtherStructReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "ThePlayerIndex":
		return m.data.ThePlayerIndex, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__myOtherStructReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "ThePlayerIndex":
		m.data.ThePlayerIndex = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__myOtherStructReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__myOtherStructReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__myOtherStructReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "MyGrowableStack":
		return m.data.MyGrowableStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__myOtherStructReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "MyGrowableStack":
		m.data.MyGrowableStack = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (m *__myOtherStructReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "MyGrowableStack":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__myOtherStructReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "MyGrowableStack":
		return m.data.MyGrowableStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__myOtherStructReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__myOtherStructReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__myOtherStructReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__myOtherStructReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__myOtherStructReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__myOtherStructReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__myOtherStructReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__myOtherStructReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *myOtherStruct) Reader() boardgame.PropertyReader {
	return &__myOtherStructReader{m}
}

func (m *myOtherStruct) ReadSetter() boardgame.PropertyReadSetter {
	return &__myOtherStructReader{m}
}

func (m *myOtherStruct) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__myOtherStructReader{m}
}

// Implementation for onlyReader

var __onlyReaderReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"MyString": boardgame.TypeString,
}

type __onlyReaderReader struct {
	data *onlyReader
}

func (o *__onlyReaderReader) Props() map[string]boardgame.PropertyType {
	return __onlyReaderReaderProps
}

func (o *__onlyReaderReader) Prop(name string) (interface{}, error) {
	props := o.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return o.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return o.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return o.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return o.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return o.IntProp(name)
	case boardgame.TypeIntSlice:
		return o.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return o.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return o.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return o.ImmutableStackProp(name)
	case boardgame.TypeString:
		return o.StringProp(name)
	case boardgame.TypeStringSlice:
		return o.StringSliceProp(name)
	case boardgame.TypeTimer:
		return o.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (o *__onlyReaderReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (o *__onlyReaderReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (o *__onlyReaderReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (o *__onlyReaderReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (o *__onlyReaderReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (o *__onlyReaderReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (o *__onlyReaderReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (o *__onlyReaderReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (o *__onlyReaderReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (o *__onlyReaderReader) StringProp(name string) (string, error) {

	switch name {
	case "MyString":
		return o.data.MyString, nil

	}

	return "", errors.New("No such String prop: " + name)

}

func (o *__onlyReaderReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (o *__onlyReaderReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (o *onlyReader) Reader() boardgame.PropertyReader {
	return &__onlyReaderReader{o}
}

// Implementation for includesImmutable

var __includesImmutableReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"MyEnum":           boardgame.TypeEnum,
	"MyImmutableTimer": boardgame.TypeTimer,
	"MyMutableEnum":    boardgame.TypeEnum,
	"MyMutableStack":   boardgame.TypeStack,
	"MyStack":          boardgame.TypeStack,
	"MyTimer":          boardgame.TypeTimer,
}

type __includesImmutableReader struct {
	data *includesImmutable
}

func (i *__includesImmutableReader) Props() map[string]boardgame.PropertyType {
	return __includesImmutableReaderProps
}

func (i *__includesImmutableReader) Prop(name string) (interface{}, error) {
	props := i.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return i.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return i.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return i.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return i.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return i.IntProp(name)
	case boardgame.TypeIntSlice:
		return i.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return i.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return i.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return i.ImmutableStackProp(name)
	case boardgame.TypeString:
		return i.StringProp(name)
	case boardgame.TypeStringSlice:
		return i.StringSliceProp(name)
	case boardgame.TypeTimer:
		return i.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (i *__includesImmutableReader) PropMutable(name string) bool {
	switch name {
	case "MyEnum":
		return false
	case "MyImmutableTimer":
		return false
	case "MyMutableEnum":
		return true
	case "MyMutableStack":
		return true
	case "MyStack":
		return false
	case "MyTimer":
		return true
	}

	return false
}

func (i *__includesImmutableReader) SetProp(name string, value interface{}) error {
	props := i.Props()
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
		return i.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return i.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return i.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return i.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return i.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return i.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return i.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return i.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (i *__includesImmutableReader) ConfigureProp(name string, value interface{}) error {
	props := i.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if i.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return i.ConfigureBoardProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableBoard)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableBoard")
			}
			return i.ConfigureImmutableBoardProp(name, val)
		}
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return i.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return i.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if i.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return i.ConfigureEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.ImmutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.ImmutableVal")
			}
			return i.ConfigureImmutableEnumProp(name, val)
		}
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return i.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return i.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return i.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return i.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if i.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return i.ConfigureStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableStack")
			}
			return i.ConfigureImmutableStackProp(name, val)
		}
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return i.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return i.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if i.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return i.ConfigureTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.ImmutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.ImmutableTimer")
			}
			return i.ConfigureImmutableTimerProp(name, val)
		}

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (i *__includesImmutableReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (i *__includesImmutableReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (i *__includesImmutableReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (i *__includesImmutableReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (i *__includesImmutableReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (i *__includesImmutableReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (i *__includesImmutableReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (i *__includesImmutableReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (i *__includesImmutableReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "MyEnum":
		return i.data.MyEnum, nil
	case "MyMutableEnum":
		return i.data.MyMutableEnum, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (i *__includesImmutableReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "MyEnum":
		return boardgame.ErrPropertyImmutable
	case "MyMutableEnum":
		i.data.MyMutableEnum = value
		return nil

	}

	return errors.New("No such Enum prop: " + name)

}

func (i *__includesImmutableReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "MyEnum":
		i.data.MyEnum = value
		return nil
	case "MyMutableEnum":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (i *__includesImmutableReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "MyEnum":
		return nil, boardgame.ErrPropertyImmutable
	case "MyMutableEnum":
		return i.data.MyMutableEnum, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (i *__includesImmutableReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (i *__includesImmutableReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (i *__includesImmutableReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (i *__includesImmutableReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (i *__includesImmutableReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (i *__includesImmutableReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (i *__includesImmutableReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (i *__includesImmutableReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (i *__includesImmutableReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "MyMutableStack":
		return i.data.MyMutableStack, nil
	case "MyStack":
		return i.data.MyStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (i *__includesImmutableReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "MyMutableStack":
		i.data.MyMutableStack = value
		return nil
	case "MyStack":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such Stack prop: " + name)

}

func (i *__includesImmutableReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "MyMutableStack":
		return boardgame.ErrPropertyImmutable
	case "MyStack":
		i.data.MyStack = value
		return nil

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (i *__includesImmutableReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "MyMutableStack":
		return i.data.MyMutableStack, nil
	case "MyStack":
		return nil, boardgame.ErrPropertyImmutable

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (i *__includesImmutableReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (i *__includesImmutableReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (i *__includesImmutableReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (i *__includesImmutableReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (i *__includesImmutableReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	switch name {
	case "MyImmutableTimer":
		return i.data.MyImmutableTimer, nil
	case "MyTimer":
		return i.data.MyTimer, nil

	}

	return nil, errors.New("No such Timer prop: " + name)

}

func (i *__includesImmutableReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	switch name {
	case "MyImmutableTimer":
		return boardgame.ErrPropertyImmutable
	case "MyTimer":
		i.data.MyTimer = value
		return nil

	}

	return errors.New("No such Timer prop: " + name)

}

func (i *__includesImmutableReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	switch name {
	case "MyImmutableTimer":
		i.data.MyImmutableTimer = value
		return nil
	case "MyTimer":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (i *__includesImmutableReader) TimerProp(name string) (boardgame.Timer, error) {

	switch name {
	case "MyImmutableTimer":
		return nil, boardgame.ErrPropertyImmutable
	case "MyTimer":
		return i.data.MyTimer, nil

	}

	return nil, errors.New("No such Timer prop: " + name)

}

func (i *includesImmutable) Reader() boardgame.PropertyReader {
	return &__includesImmutableReader{i}
}

func (i *includesImmutable) ReadSetter() boardgame.PropertyReadSetter {
	return &__includesImmutableReader{i}
}

func (i *includesImmutable) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__includesImmutableReader{i}
}

// Implementation for upToReadSetter

var __upToReadSetterReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"MyInt": boardgame.TypeInt,
}

type __upToReadSetterReader struct {
	data *upToReadSetter
}

func (u *__upToReadSetterReader) Props() map[string]boardgame.PropertyType {
	return __upToReadSetterReaderProps
}

func (u *__upToReadSetterReader) Prop(name string) (interface{}, error) {
	props := u.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return u.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return u.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return u.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return u.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return u.IntProp(name)
	case boardgame.TypeIntSlice:
		return u.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return u.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return u.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return u.ImmutableStackProp(name)
	case boardgame.TypeString:
		return u.StringProp(name)
	case boardgame.TypeStringSlice:
		return u.StringSliceProp(name)
	case boardgame.TypeTimer:
		return u.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (u *__upToReadSetterReader) PropMutable(name string) bool {
	switch name {
	case "MyInt":
		return true
	}

	return false
}

func (u *__upToReadSetterReader) SetProp(name string, value interface{}) error {
	props := u.Props()
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
		return u.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return u.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return u.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return u.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return u.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return u.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return u.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return u.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (u *__upToReadSetterReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (u *__upToReadSetterReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (u *__upToReadSetterReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (u *__upToReadSetterReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (u *__upToReadSetterReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (u *__upToReadSetterReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (u *__upToReadSetterReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (u *__upToReadSetterReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (u *__upToReadSetterReader) IntProp(name string) (int, error) {

	switch name {
	case "MyInt":
		return u.data.MyInt, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (u *__upToReadSetterReader) SetIntProp(name string, value int) error {

	switch name {
	case "MyInt":
		u.data.MyInt = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (u *__upToReadSetterReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (u *__upToReadSetterReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (u *__upToReadSetterReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (u *__upToReadSetterReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (u *__upToReadSetterReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (u *__upToReadSetterReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (u *__upToReadSetterReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (u *__upToReadSetterReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (u *__upToReadSetterReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (u *__upToReadSetterReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (u *__upToReadSetterReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (u *__upToReadSetterReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (u *__upToReadSetterReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (u *__upToReadSetterReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (u *upToReadSetter) Reader() boardgame.PropertyReader {
	return &__upToReadSetterReader{u}
}

func (u *upToReadSetter) ReadSetter() boardgame.PropertyReadSetter {
	return &__upToReadSetterReader{u}
}

// Implementation for sizedStackExample

var __sizedStackExampleReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"MyMutableSizedStack": boardgame.TypeStack,
	"MySizedStack":        boardgame.TypeStack,
}

type __sizedStackExampleReader struct {
	data *sizedStackExample
}

func (s *__sizedStackExampleReader) Props() map[string]boardgame.PropertyType {
	return __sizedStackExampleReaderProps
}

func (s *__sizedStackExampleReader) Prop(name string) (interface{}, error) {
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

func (s *__sizedStackExampleReader) PropMutable(name string) bool {
	switch name {
	case "MyMutableSizedStack":
		return true
	case "MySizedStack":
		return false
	}

	return false
}

func (s *__sizedStackExampleReader) SetProp(name string, value interface{}) error {
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

func (s *__sizedStackExampleReader) ConfigureProp(name string, value interface{}) error {
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

func (s *__sizedStackExampleReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (s *__sizedStackExampleReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (s *__sizedStackExampleReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (s *__sizedStackExampleReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (s *__sizedStackExampleReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (s *__sizedStackExampleReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (s *__sizedStackExampleReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (s *__sizedStackExampleReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (s *__sizedStackExampleReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (s *__sizedStackExampleReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (s *__sizedStackExampleReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (s *__sizedStackExampleReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (s *__sizedStackExampleReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (s *__sizedStackExampleReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (s *__sizedStackExampleReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (s *__sizedStackExampleReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (s *__sizedStackExampleReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (s *__sizedStackExampleReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (s *__sizedStackExampleReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (s *__sizedStackExampleReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (s *__sizedStackExampleReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "MyMutableSizedStack":
		return s.data.MyMutableSizedStack, nil
	case "MySizedStack":
		return s.data.MySizedStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (s *__sizedStackExampleReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "MyMutableSizedStack":
		slotValue := value.SizedStack()
		if slotValue == nil {
			return errors.New("MyMutableSizedStack couldn't be upconverted, returned nil.")
		}
		s.data.MyMutableSizedStack = slotValue
		return nil
	case "MySizedStack":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such Stack prop: " + name)

}

func (s *__sizedStackExampleReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "MyMutableSizedStack":
		return boardgame.ErrPropertyImmutable
	case "MySizedStack":
		slotValue := value.ImmutableSizedStack()
		if slotValue == nil {
			return errors.New("MySizedStack couldn't be upconverted, returned nil.")
		}
		s.data.MySizedStack = slotValue
		return nil

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (s *__sizedStackExampleReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "MyMutableSizedStack":
		return s.data.MyMutableSizedStack, nil
	case "MySizedStack":
		return nil, boardgame.ErrPropertyImmutable

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (s *__sizedStackExampleReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (s *__sizedStackExampleReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (s *__sizedStackExampleReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (s *__sizedStackExampleReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (s *__sizedStackExampleReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (s *__sizedStackExampleReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (s *__sizedStackExampleReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (s *__sizedStackExampleReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (s *sizedStackExample) Reader() boardgame.PropertyReader {
	return &__sizedStackExampleReader{s}
}

func (s *sizedStackExample) ReadSetter() boardgame.PropertyReadSetter {
	return &__sizedStackExampleReader{s}
}

func (s *sizedStackExample) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__sizedStackExampleReader{s}
}

// Implementation for mergedStackExample

var __mergedStackExampleReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"MyMergedStack": boardgame.TypeStack,
}

type __mergedStackExampleReader struct {
	data *mergedStackExample
}

func (m *__mergedStackExampleReader) Props() map[string]boardgame.PropertyType {
	return __mergedStackExampleReaderProps
}

func (m *__mergedStackExampleReader) Prop(name string) (interface{}, error) {
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

func (m *__mergedStackExampleReader) PropMutable(name string) bool {
	switch name {
	case "MyMergedStack":
		return false
	}

	return false
}

func (m *__mergedStackExampleReader) SetProp(name string, value interface{}) error {
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

func (m *__mergedStackExampleReader) ConfigureProp(name string, value interface{}) error {
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

func (m *__mergedStackExampleReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__mergedStackExampleReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *__mergedStackExampleReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *__mergedStackExampleReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *__mergedStackExampleReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *__mergedStackExampleReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *__mergedStackExampleReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *__mergedStackExampleReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *__mergedStackExampleReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__mergedStackExampleReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *__mergedStackExampleReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *__mergedStackExampleReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *__mergedStackExampleReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *__mergedStackExampleReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *__mergedStackExampleReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *__mergedStackExampleReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *__mergedStackExampleReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *__mergedStackExampleReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *__mergedStackExampleReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__mergedStackExampleReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *__mergedStackExampleReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "MyMergedStack":
		return m.data.MyMergedStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__mergedStackExampleReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "MyMergedStack":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such Stack prop: " + name)

}

func (m *__mergedStackExampleReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "MyMergedStack":
		slotValue := value.MergedStack()
		if slotValue == nil {
			return errors.New("MyMergedStack couldn't be upconverted, returned nil.")
		}
		m.data.MyMergedStack = slotValue
		return nil

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *__mergedStackExampleReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "MyMergedStack":
		return nil, boardgame.ErrPropertyImmutable

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *__mergedStackExampleReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *__mergedStackExampleReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *__mergedStackExampleReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *__mergedStackExampleReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *__mergedStackExampleReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *__mergedStackExampleReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *__mergedStackExampleReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *__mergedStackExampleReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *mergedStackExample) Reader() boardgame.PropertyReader {
	return &__mergedStackExampleReader{m}
}

func (m *mergedStackExample) ReadSetter() boardgame.PropertyReadSetter {
	return &__mergedStackExampleReader{m}
}

func (m *mergedStackExample) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__mergedStackExampleReader{m}
}

// Implementation for rangeValExample

var __rangeValExampleReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"MyMutableRangeVal": boardgame.TypeEnum,
	"MyRangeVal":        boardgame.TypeEnum,
}

type __rangeValExampleReader struct {
	data *rangeValExample
}

func (r *__rangeValExampleReader) Props() map[string]boardgame.PropertyType {
	return __rangeValExampleReaderProps
}

func (r *__rangeValExampleReader) Prop(name string) (interface{}, error) {
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

func (r *__rangeValExampleReader) PropMutable(name string) bool {
	switch name {
	case "MyMutableRangeVal":
		return true
	case "MyRangeVal":
		return false
	}

	return false
}

func (r *__rangeValExampleReader) SetProp(name string, value interface{}) error {
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

func (r *__rangeValExampleReader) ConfigureProp(name string, value interface{}) error {
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

func (r *__rangeValExampleReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (r *__rangeValExampleReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (r *__rangeValExampleReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (r *__rangeValExampleReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (r *__rangeValExampleReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (r *__rangeValExampleReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (r *__rangeValExampleReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (r *__rangeValExampleReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (r *__rangeValExampleReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "MyMutableRangeVal":
		return r.data.MyMutableRangeVal, nil
	case "MyRangeVal":
		return r.data.MyRangeVal, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (r *__rangeValExampleReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "MyMutableRangeVal":
		slotValue := value.RangeVal()
		if slotValue == nil {
			return errors.New("MyMutableRangeVal couldn't be upconverted, returned nil.")
		}
		r.data.MyMutableRangeVal = slotValue
		return nil
	case "MyRangeVal":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such Enum prop: " + name)

}

func (r *__rangeValExampleReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "MyMutableRangeVal":
		return boardgame.ErrPropertyImmutable
	case "MyRangeVal":
		slotValue := value.ImmutableRangeVal()
		if slotValue == nil {
			return errors.New("MyRangeVal couldn't be upconverted, returned nil.")
		}
		r.data.MyRangeVal = slotValue
		return nil

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (r *__rangeValExampleReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "MyMutableRangeVal":
		return r.data.MyMutableRangeVal, nil
	case "MyRangeVal":
		return nil, boardgame.ErrPropertyImmutable

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (r *__rangeValExampleReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (r *__rangeValExampleReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (r *__rangeValExampleReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (r *__rangeValExampleReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (r *__rangeValExampleReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (r *__rangeValExampleReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (r *__rangeValExampleReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (r *__rangeValExampleReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (r *__rangeValExampleReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (r *__rangeValExampleReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (r *__rangeValExampleReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (r *__rangeValExampleReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (r *__rangeValExampleReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (r *__rangeValExampleReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (r *__rangeValExampleReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (r *__rangeValExampleReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (r *__rangeValExampleReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (r *__rangeValExampleReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (r *__rangeValExampleReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (r *__rangeValExampleReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (r *rangeValExample) Reader() boardgame.PropertyReader {
	return &__rangeValExampleReader{r}
}

func (r *rangeValExample) ReadSetter() boardgame.PropertyReadSetter {
	return &__rangeValExampleReader{r}
}

func (r *rangeValExample) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__rangeValExampleReader{r}
}

// Implementation for treeValExample

var __treeValExampleReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"MyImmutableTreeVal": boardgame.TypeEnum,
	"MyTreeVal":          boardgame.TypeEnum,
}

type __treeValExampleReader struct {
	data *treeValExample
}

func (t *__treeValExampleReader) Props() map[string]boardgame.PropertyType {
	return __treeValExampleReaderProps
}

func (t *__treeValExampleReader) Prop(name string) (interface{}, error) {
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

func (t *__treeValExampleReader) PropMutable(name string) bool {
	switch name {
	case "MyImmutableTreeVal":
		return false
	case "MyTreeVal":
		return true
	}

	return false
}

func (t *__treeValExampleReader) SetProp(name string, value interface{}) error {
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

func (t *__treeValExampleReader) ConfigureProp(name string, value interface{}) error {
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

func (t *__treeValExampleReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (t *__treeValExampleReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (t *__treeValExampleReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (t *__treeValExampleReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (t *__treeValExampleReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (t *__treeValExampleReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (t *__treeValExampleReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (t *__treeValExampleReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (t *__treeValExampleReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "MyImmutableTreeVal":
		return t.data.MyImmutableTreeVal, nil
	case "MyTreeVal":
		return t.data.MyTreeVal, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (t *__treeValExampleReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "MyImmutableTreeVal":
		return boardgame.ErrPropertyImmutable
	case "MyTreeVal":
		slotValue := value.TreeVal()
		if slotValue == nil {
			return errors.New("MyTreeVal couldn't be upconverted, returned nil.")
		}
		t.data.MyTreeVal = slotValue
		return nil

	}

	return errors.New("No such Enum prop: " + name)

}

func (t *__treeValExampleReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "MyImmutableTreeVal":
		slotValue := value.ImmutableTreeVal()
		if slotValue == nil {
			return errors.New("MyImmutableTreeVal couldn't be upconverted, returned nil.")
		}
		t.data.MyImmutableTreeVal = slotValue
		return nil
	case "MyTreeVal":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (t *__treeValExampleReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "MyImmutableTreeVal":
		return nil, boardgame.ErrPropertyImmutable
	case "MyTreeVal":
		return t.data.MyTreeVal, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (t *__treeValExampleReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (t *__treeValExampleReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (t *__treeValExampleReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (t *__treeValExampleReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (t *__treeValExampleReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (t *__treeValExampleReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (t *__treeValExampleReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (t *__treeValExampleReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (t *__treeValExampleReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (t *__treeValExampleReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (t *__treeValExampleReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (t *__treeValExampleReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (t *__treeValExampleReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (t *__treeValExampleReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (t *__treeValExampleReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (t *__treeValExampleReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (t *__treeValExampleReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (t *__treeValExampleReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (t *__treeValExampleReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (t *__treeValExampleReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (t *treeValExample) Reader() boardgame.PropertyReader {
	return &__treeValExampleReader{t}
}

func (t *treeValExample) ReadSetter() boardgame.PropertyReadSetter {
	return &__treeValExampleReader{t}
}

func (t *treeValExample) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &__treeValExampleReader{t}
}
