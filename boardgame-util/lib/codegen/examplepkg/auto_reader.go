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

var ȧutoGeneratedMyStructReaderProps = map[string]boardgame.PropertyType{
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

type ȧutoGeneratedMyStructReader struct {
	data *myStruct
}

func (m *ȧutoGeneratedMyStructReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMyStructReaderProps
}

func (m *ȧutoGeneratedMyStructReader) Prop(name string) (interface{}, error) {
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

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMyStructReader) PropMutable(name string) bool {
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

func (m *ȧutoGeneratedMyStructReader) SetProp(name string, value interface{}) error {
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMyStructReader) ConfigureProp(name string, value interface{}) error {
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMyStructReader) IntProp(name string) (int, error) {

	switch name {
	case "MyInt":
		return m.data.MyInt, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) SetIntProp(name string, value int) error {

	switch name {
	case "MyInt":
		m.data.MyInt = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) BoolProp(name string) (bool, error) {

	switch name {
	case "MyBool":
		return m.data.MyBool, nil

	}

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) SetBoolProp(name string, value bool) error {

	switch name {
	case "MyBool":
		m.data.MyBool = value
		return nil

	}

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "EnumVar":
		return m.data.EnumVar, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "EnumVar":
		m.data.EnumVar = value
		return nil

	}

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "EnumVar":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "EnumVar":
		return m.data.EnumVar, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) IntSliceProp(name string) ([]int, error) {

	switch name {
	case "MyIntSlice":
		return m.data.MyIntSlice, nil

	}

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) SetIntSliceProp(name string, value []int) error {

	switch name {
	case "MyIntSlice":
		m.data.MyIntSlice = value
		return nil

	}

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) BoolSliceProp(name string) ([]bool, error) {

	switch name {
	case "MyBoolSlice":
		return m.data.MyBoolSlice, nil

	}

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) SetBoolSliceProp(name string, value []bool) error {

	switch name {
	case "MyBoolSlice":
		m.data.MyBoolSlice = value
		return nil

	}

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) StringSliceProp(name string) ([]string, error) {

	switch name {
	case "MyStringSlice":
		return m.data.MyStringSlice, nil

	}

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) SetStringSliceProp(name string, value []string) error {

	switch name {
	case "MyStringSlice":
		m.data.MyStringSlice = value
		return nil

	}

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	switch name {
	case "MyPlayerIndexSlice":
		return m.data.MyPlayerIndexSlice, nil

	}

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	switch name {
	case "MyPlayerIndexSlice":
		m.data.MyPlayerIndexSlice = value
		return nil

	}

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "MySizedStack":
		return m.data.MySizedStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "MySizedStack":
		m.data.MySizedStack = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "MySizedStack":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "MySizedStack":
		return m.data.MySizedStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	switch name {
	case "TheTimer":
		return m.data.TheTimer, nil

	}

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	switch name {
	case "TheTimer":
		m.data.TheTimer = value
		return nil

	}

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	switch name {
	case "TheTimer":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMyStructReader) TimerProp(name string) (boardgame.Timer, error) {

	switch name {
	case "TheTimer":
		return m.data.TheTimer, nil

	}

	return nil, errors.New("No such Timer prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for myStruct
func (m *myStruct) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMyStructReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for myStruct
func (m *myStruct) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMyStructReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for myStruct
func (m *myStruct) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMyStructReader{m}
}

// Implementation for roundRobinStruct

var ȧutoGeneratedRoundRobinStructReaderProps = map[string]boardgame.PropertyType{
	"MyBool":          boardgame.TypeBool,
	"RRHasStarted":    boardgame.TypeBool,
	"RRLastPlayer":    boardgame.TypePlayerIndex,
	"RRRoundCount":    boardgame.TypeInt,
	"RRStarterPlayer": boardgame.TypePlayerIndex,
}

type ȧutoGeneratedRoundRobinStructReader struct {
	data *roundRobinStruct
}

func (r *ȧutoGeneratedRoundRobinStructReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedRoundRobinStructReaderProps
}

func (r *ȧutoGeneratedRoundRobinStructReader) Prop(name string) (interface{}, error) {
	props := r.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return r.IntProp(name)
	case boardgame.TypeBool:
		return r.BoolProp(name)
	case boardgame.TypeString:
		return r.StringProp(name)
	case boardgame.TypePlayerIndex:
		return r.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return r.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return r.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return r.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return r.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return r.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return r.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return r.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return r.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (r *ȧutoGeneratedRoundRobinStructReader) PropMutable(name string) bool {
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

func (r *ȧutoGeneratedRoundRobinStructReader) SetProp(name string, value interface{}) error {
	props := r.Props()
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
		return r.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return r.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return r.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return r.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return r.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return r.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return r.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return r.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (r *ȧutoGeneratedRoundRobinStructReader) ConfigureProp(name string, value interface{}) error {
	props := r.Props()
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
		return r.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return r.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return r.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return r.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if r.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return r.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return r.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return r.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return r.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return r.SetStringSliceProp(name, val)
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
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return r.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if r.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return r.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return r.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if r.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return r.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return r.ConfigureImmutableTimerProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (r *ȧutoGeneratedRoundRobinStructReader) IntProp(name string) (int, error) {

	switch name {
	case "RRRoundCount":
		return r.data.RRRoundCount, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) SetIntProp(name string, value int) error {

	switch name {
	case "RRRoundCount":
		r.data.RRRoundCount = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) BoolProp(name string) (bool, error) {

	switch name {
	case "MyBool":
		return r.data.MyBool, nil
	case "RRHasStarted":
		return r.data.RRHasStarted, nil

	}

	return false, errors.New("No such Bool prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) SetBoolProp(name string, value bool) error {

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

func (r *ȧutoGeneratedRoundRobinStructReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "RRLastPlayer":
		return r.data.RRLastPlayer, nil
	case "RRStarterPlayer":
		return r.data.RRStarterPlayer, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

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

func (r *ȧutoGeneratedRoundRobinStructReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (r *ȧutoGeneratedRoundRobinStructReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for roundRobinStruct
func (r *roundRobinStruct) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedRoundRobinStructReader{r}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for roundRobinStruct
func (r *roundRobinStruct) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedRoundRobinStructReader{r}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for roundRobinStruct
func (r *roundRobinStruct) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedRoundRobinStructReader{r}
}

// Implementation for structWithManyKeys

var ȧutoGeneratedStructWithManyKeysReaderProps = map[string]boardgame.PropertyType{
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

type ȧutoGeneratedStructWithManyKeysReader struct {
	data *structWithManyKeys
}

func (s *ȧutoGeneratedStructWithManyKeysReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedStructWithManyKeysReaderProps
}

func (s *ȧutoGeneratedStructWithManyKeysReader) Prop(name string) (interface{}, error) {
	props := s.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return s.IntProp(name)
	case boardgame.TypeBool:
		return s.BoolProp(name)
	case boardgame.TypeString:
		return s.StringProp(name)
	case boardgame.TypePlayerIndex:
		return s.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return s.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return s.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return s.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return s.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return s.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return s.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return s.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return s.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (s *ȧutoGeneratedStructWithManyKeysReader) PropMutable(name string) bool {
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

func (s *ȧutoGeneratedStructWithManyKeysReader) SetProp(name string, value interface{}) error {
	props := s.Props()
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
		return s.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return s.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return s.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return s.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return s.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return s.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return s.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return s.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (s *ȧutoGeneratedStructWithManyKeysReader) ConfigureProp(name string, value interface{}) error {
	props := s.Props()
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
		return s.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return s.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return s.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return s.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if s.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return s.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return s.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return s.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return s.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return s.SetStringSliceProp(name, val)
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
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return s.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if s.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return s.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return s.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if s.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return s.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return s.ConfigureImmutableTimerProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (s *ȧutoGeneratedStructWithManyKeysReader) IntProp(name string) (int, error) {

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

func (s *ȧutoGeneratedStructWithManyKeysReader) SetIntProp(name string, value int) error {

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

func (s *ȧutoGeneratedStructWithManyKeysReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (s *ȧutoGeneratedStructWithManyKeysReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for structWithManyKeys
func (s *structWithManyKeys) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedStructWithManyKeysReader{s}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for structWithManyKeys
func (s *structWithManyKeys) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedStructWithManyKeysReader{s}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for structWithManyKeys
func (s *structWithManyKeys) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedStructWithManyKeysReader{s}
}

// Implementation for embeddedStruct

var ȧutoGeneratedEmbeddedStructReaderProps = map[string]boardgame.PropertyType{
	"MyInt":             boardgame.TypeInt,
	"TargetPlayerIndex": boardgame.TypePlayerIndex,
}

type ȧutoGeneratedEmbeddedStructReader struct {
	data *embeddedStruct
}

func (e *ȧutoGeneratedEmbeddedStructReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedEmbeddedStructReaderProps
}

func (e *ȧutoGeneratedEmbeddedStructReader) Prop(name string) (interface{}, error) {
	props := e.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return e.IntProp(name)
	case boardgame.TypeBool:
		return e.BoolProp(name)
	case boardgame.TypeString:
		return e.StringProp(name)
	case boardgame.TypePlayerIndex:
		return e.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return e.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return e.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return e.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return e.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return e.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return e.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return e.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return e.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (e *ȧutoGeneratedEmbeddedStructReader) PropMutable(name string) bool {
	switch name {
	case "MyInt":
		return true
	case "TargetPlayerIndex":
		return true
	}

	return false
}

func (e *ȧutoGeneratedEmbeddedStructReader) SetProp(name string, value interface{}) error {
	props := e.Props()
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
		return e.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return e.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return e.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return e.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return e.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return e.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return e.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return e.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (e *ȧutoGeneratedEmbeddedStructReader) ConfigureProp(name string, value interface{}) error {
	props := e.Props()
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
		return e.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return e.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return e.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return e.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if e.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return e.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return e.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return e.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return e.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return e.SetStringSliceProp(name, val)
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
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return e.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if e.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return e.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return e.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if e.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return e.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return e.ConfigureImmutableTimerProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (e *ȧutoGeneratedEmbeddedStructReader) IntProp(name string) (int, error) {

	switch name {
	case "MyInt":
		return e.data.MyInt, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) SetIntProp(name string, value int) error {

	switch name {
	case "MyInt":
		e.data.MyInt = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "TargetPlayerIndex":
		return e.data.TargetPlayerIndex, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "TargetPlayerIndex":
		e.data.TargetPlayerIndex = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (e *ȧutoGeneratedEmbeddedStructReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for embeddedStruct
func (e *embeddedStruct) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedEmbeddedStructReader{e}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for embeddedStruct
func (e *embeddedStruct) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedEmbeddedStructReader{e}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for embeddedStruct
func (e *embeddedStruct) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedEmbeddedStructReader{e}
}

// Implementation for doubleEmbeddedStruct

var ȧutoGeneratedDoubleEmbeddedStructReaderProps = map[string]boardgame.PropertyType{
	"MyInt":             boardgame.TypeInt,
	"TargetPlayerIndex": boardgame.TypePlayerIndex,
}

type ȧutoGeneratedDoubleEmbeddedStructReader struct {
	data *doubleEmbeddedStruct
}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedDoubleEmbeddedStructReaderProps
}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) Prop(name string) (interface{}, error) {
	props := d.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return d.IntProp(name)
	case boardgame.TypeBool:
		return d.BoolProp(name)
	case boardgame.TypeString:
		return d.StringProp(name)
	case boardgame.TypePlayerIndex:
		return d.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return d.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return d.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return d.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return d.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return d.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return d.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return d.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return d.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) PropMutable(name string) bool {
	switch name {
	case "MyInt":
		return true
	case "TargetPlayerIndex":
		return true
	}

	return false
}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) SetProp(name string, value interface{}) error {
	props := d.Props()
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
		return d.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return d.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return d.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return d.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return d.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return d.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) ConfigureProp(name string, value interface{}) error {
	props := d.Props()
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
		return d.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return d.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return d.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return d.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return d.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return d.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return d.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return d.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return d.SetStringSliceProp(name, val)
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
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return d.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return d.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return d.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if d.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return d.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return d.ConfigureImmutableTimerProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) IntProp(name string) (int, error) {

	switch name {
	case "MyInt":
		return d.data.MyInt, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) SetIntProp(name string, value int) error {

	switch name {
	case "MyInt":
		d.data.MyInt = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "TargetPlayerIndex":
		return d.data.TargetPlayerIndex, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "TargetPlayerIndex":
		d.data.TargetPlayerIndex = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (d *ȧutoGeneratedDoubleEmbeddedStructReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for doubleEmbeddedStruct
func (d *doubleEmbeddedStruct) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedDoubleEmbeddedStructReader{d}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for doubleEmbeddedStruct
func (d *doubleEmbeddedStruct) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedDoubleEmbeddedStructReader{d}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for doubleEmbeddedStruct
func (d *doubleEmbeddedStruct) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedDoubleEmbeddedStructReader{d}
}

// Implementation for myOtherStruct

var ȧutoGeneratedMyOtherStructReaderProps = map[string]boardgame.PropertyType{
	"MyGrowableStack": boardgame.TypeStack,
	"ThePlayerIndex":  boardgame.TypePlayerIndex,
}

type ȧutoGeneratedMyOtherStructReader struct {
	data *myOtherStruct
}

func (m *ȧutoGeneratedMyOtherStructReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMyOtherStructReaderProps
}

func (m *ȧutoGeneratedMyOtherStructReader) Prop(name string) (interface{}, error) {
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

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMyOtherStructReader) PropMutable(name string) bool {
	switch name {
	case "MyGrowableStack":
		return true
	case "ThePlayerIndex":
		return true
	}

	return false
}

func (m *ȧutoGeneratedMyOtherStructReader) SetProp(name string, value interface{}) error {
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMyOtherStructReader) ConfigureProp(name string, value interface{}) error {
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMyOtherStructReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "ThePlayerIndex":
		return m.data.ThePlayerIndex, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "ThePlayerIndex":
		m.data.ThePlayerIndex = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "MyGrowableStack":
		return m.data.MyGrowableStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "MyGrowableStack":
		m.data.MyGrowableStack = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "MyGrowableStack":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "MyGrowableStack":
		return m.data.MyGrowableStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMyOtherStructReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for myOtherStruct
func (m *myOtherStruct) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMyOtherStructReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for myOtherStruct
func (m *myOtherStruct) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMyOtherStructReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for myOtherStruct
func (m *myOtherStruct) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMyOtherStructReader{m}
}

// Implementation for onlyReader

var ȧutoGeneratedOnlyReaderReaderProps = map[string]boardgame.PropertyType{
	"MyString": boardgame.TypeString,
}

type ȧutoGeneratedOnlyReaderReader struct {
	data *onlyReader
}

func (o *ȧutoGeneratedOnlyReaderReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedOnlyReaderReaderProps
}

func (o *ȧutoGeneratedOnlyReaderReader) Prop(name string) (interface{}, error) {
	props := o.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return o.IntProp(name)
	case boardgame.TypeBool:
		return o.BoolProp(name)
	case boardgame.TypeString:
		return o.StringProp(name)
	case boardgame.TypePlayerIndex:
		return o.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return o.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return o.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return o.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return o.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return o.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return o.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return o.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return o.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (o *ȧutoGeneratedOnlyReaderReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (o *ȧutoGeneratedOnlyReaderReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (o *ȧutoGeneratedOnlyReaderReader) StringProp(name string) (string, error) {

	switch name {
	case "MyString":
		return o.data.MyString, nil

	}

	return "", errors.New("No such String prop: " + name)

}

func (o *ȧutoGeneratedOnlyReaderReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (o *ȧutoGeneratedOnlyReaderReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (o *ȧutoGeneratedOnlyReaderReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (o *ȧutoGeneratedOnlyReaderReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (o *ȧutoGeneratedOnlyReaderReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (o *ȧutoGeneratedOnlyReaderReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (o *ȧutoGeneratedOnlyReaderReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (o *ȧutoGeneratedOnlyReaderReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (o *ȧutoGeneratedOnlyReaderReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for onlyReader
func (o *onlyReader) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedOnlyReaderReader{o}
}

// Implementation for includesImmutable

var ȧutoGeneratedIncludesImmutableReaderProps = map[string]boardgame.PropertyType{
	"MyEnum":           boardgame.TypeEnum,
	"MyImmutableTimer": boardgame.TypeTimer,
	"MyMutableEnum":    boardgame.TypeEnum,
	"MyMutableStack":   boardgame.TypeStack,
	"MyStack":          boardgame.TypeStack,
	"MyTimer":          boardgame.TypeTimer,
}

type ȧutoGeneratedIncludesImmutableReader struct {
	data *includesImmutable
}

func (i *ȧutoGeneratedIncludesImmutableReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedIncludesImmutableReaderProps
}

func (i *ȧutoGeneratedIncludesImmutableReader) Prop(name string) (interface{}, error) {
	props := i.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return i.IntProp(name)
	case boardgame.TypeBool:
		return i.BoolProp(name)
	case boardgame.TypeString:
		return i.StringProp(name)
	case boardgame.TypePlayerIndex:
		return i.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return i.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return i.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return i.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return i.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return i.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return i.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return i.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return i.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (i *ȧutoGeneratedIncludesImmutableReader) PropMutable(name string) bool {
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

func (i *ȧutoGeneratedIncludesImmutableReader) SetProp(name string, value interface{}) error {
	props := i.Props()
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
		return i.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return i.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return i.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return i.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return i.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return i.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return i.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return i.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (i *ȧutoGeneratedIncludesImmutableReader) ConfigureProp(name string, value interface{}) error {
	props := i.Props()
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
		return i.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return i.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return i.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return i.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if i.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return i.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return i.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return i.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return i.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return i.SetStringSliceProp(name, val)
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
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return i.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if i.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return i.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return i.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if i.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return i.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return i.ConfigureImmutableTimerProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (i *ȧutoGeneratedIncludesImmutableReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "MyEnum":
		return i.data.MyEnum, nil
	case "MyMutableEnum":
		return i.data.MyMutableEnum, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "MyEnum":
		return boardgame.ErrPropertyImmutable
	case "MyMutableEnum":
		i.data.MyMutableEnum = value
		return nil

	}

	return errors.New("No such Enum prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "MyEnum":
		i.data.MyEnum = value
		return nil
	case "MyMutableEnum":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "MyEnum":
		return nil, boardgame.ErrPropertyImmutable
	case "MyMutableEnum":
		return i.data.MyMutableEnum, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "MyMutableStack":
		return i.data.MyMutableStack, nil
	case "MyStack":
		return i.data.MyStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "MyMutableStack":
		i.data.MyMutableStack = value
		return nil
	case "MyStack":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such Stack prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "MyMutableStack":
		return boardgame.ErrPropertyImmutable
	case "MyStack":
		i.data.MyStack = value
		return nil

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "MyMutableStack":
		return i.data.MyMutableStack, nil
	case "MyStack":
		return nil, boardgame.ErrPropertyImmutable

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	switch name {
	case "MyImmutableTimer":
		return i.data.MyImmutableTimer, nil
	case "MyTimer":
		return i.data.MyTimer, nil

	}

	return nil, errors.New("No such Timer prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	switch name {
	case "MyImmutableTimer":
		return boardgame.ErrPropertyImmutable
	case "MyTimer":
		i.data.MyTimer = value
		return nil

	}

	return errors.New("No such Timer prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	switch name {
	case "MyImmutableTimer":
		i.data.MyImmutableTimer = value
		return nil
	case "MyTimer":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (i *ȧutoGeneratedIncludesImmutableReader) TimerProp(name string) (boardgame.Timer, error) {

	switch name {
	case "MyImmutableTimer":
		return nil, boardgame.ErrPropertyImmutable
	case "MyTimer":
		return i.data.MyTimer, nil

	}

	return nil, errors.New("No such Timer prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for includesImmutable
func (i *includesImmutable) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedIncludesImmutableReader{i}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for includesImmutable
func (i *includesImmutable) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedIncludesImmutableReader{i}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for includesImmutable
func (i *includesImmutable) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedIncludesImmutableReader{i}
}

// Implementation for upToReadSetter

var ȧutoGeneratedUpToReadSetterReaderProps = map[string]boardgame.PropertyType{
	"MyInt": boardgame.TypeInt,
}

type ȧutoGeneratedUpToReadSetterReader struct {
	data *upToReadSetter
}

func (u *ȧutoGeneratedUpToReadSetterReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedUpToReadSetterReaderProps
}

func (u *ȧutoGeneratedUpToReadSetterReader) Prop(name string) (interface{}, error) {
	props := u.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return u.IntProp(name)
	case boardgame.TypeBool:
		return u.BoolProp(name)
	case boardgame.TypeString:
		return u.StringProp(name)
	case boardgame.TypePlayerIndex:
		return u.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return u.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return u.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return u.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return u.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return u.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return u.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return u.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return u.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (u *ȧutoGeneratedUpToReadSetterReader) PropMutable(name string) bool {
	switch name {
	case "MyInt":
		return true
	}

	return false
}

func (u *ȧutoGeneratedUpToReadSetterReader) SetProp(name string, value interface{}) error {
	props := u.Props()
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
		return u.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return u.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return u.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return u.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return u.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return u.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return u.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return u.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (u *ȧutoGeneratedUpToReadSetterReader) IntProp(name string) (int, error) {

	switch name {
	case "MyInt":
		return u.data.MyInt, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) SetIntProp(name string, value int) error {

	switch name {
	case "MyInt":
		u.data.MyInt = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (u *ȧutoGeneratedUpToReadSetterReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for upToReadSetter
func (u *upToReadSetter) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedUpToReadSetterReader{u}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for upToReadSetter
func (u *upToReadSetter) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedUpToReadSetterReader{u}
}

// Implementation for sizedStackExample

var ȧutoGeneratedSizedStackExampleReaderProps = map[string]boardgame.PropertyType{
	"MyMutableSizedStack": boardgame.TypeStack,
	"MySizedStack":        boardgame.TypeStack,
}

type ȧutoGeneratedSizedStackExampleReader struct {
	data *sizedStackExample
}

func (s *ȧutoGeneratedSizedStackExampleReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedSizedStackExampleReaderProps
}

func (s *ȧutoGeneratedSizedStackExampleReader) Prop(name string) (interface{}, error) {
	props := s.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return s.IntProp(name)
	case boardgame.TypeBool:
		return s.BoolProp(name)
	case boardgame.TypeString:
		return s.StringProp(name)
	case boardgame.TypePlayerIndex:
		return s.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return s.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return s.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return s.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return s.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return s.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return s.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return s.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return s.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (s *ȧutoGeneratedSizedStackExampleReader) PropMutable(name string) bool {
	switch name {
	case "MyMutableSizedStack":
		return true
	case "MySizedStack":
		return false
	}

	return false
}

func (s *ȧutoGeneratedSizedStackExampleReader) SetProp(name string, value interface{}) error {
	props := s.Props()
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
		return s.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return s.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return s.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return s.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return s.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return s.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return s.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return s.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (s *ȧutoGeneratedSizedStackExampleReader) ConfigureProp(name string, value interface{}) error {
	props := s.Props()
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
		return s.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return s.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return s.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return s.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if s.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return s.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return s.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return s.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return s.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return s.SetStringSliceProp(name, val)
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
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return s.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if s.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return s.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return s.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if s.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return s.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return s.ConfigureImmutableTimerProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (s *ȧutoGeneratedSizedStackExampleReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "MyMutableSizedStack":
		return s.data.MyMutableSizedStack, nil
	case "MySizedStack":
		return s.data.MySizedStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "MyMutableSizedStack":
		slotValue := value.SizedStack()
		if slotValue == nil {
			return errors.New("MyMutableSizedStack couldn't be upconverted, returned nil")
		}
		s.data.MyMutableSizedStack = slotValue
		return nil
	case "MySizedStack":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such Stack prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "MyMutableSizedStack":
		return boardgame.ErrPropertyImmutable
	case "MySizedStack":
		slotValue := value.ImmutableSizedStack()
		if slotValue == nil {
			return errors.New("MySizedStack couldn't be upconverted, returned nil")
		}
		s.data.MySizedStack = slotValue
		return nil

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "MyMutableSizedStack":
		return s.data.MyMutableSizedStack, nil
	case "MySizedStack":
		return nil, boardgame.ErrPropertyImmutable

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (s *ȧutoGeneratedSizedStackExampleReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for sizedStackExample
func (s *sizedStackExample) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedSizedStackExampleReader{s}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for sizedStackExample
func (s *sizedStackExample) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedSizedStackExampleReader{s}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for sizedStackExample
func (s *sizedStackExample) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedSizedStackExampleReader{s}
}

// Implementation for mergedStackExample

var ȧutoGeneratedMergedStackExampleReaderProps = map[string]boardgame.PropertyType{
	"MyMergedStack": boardgame.TypeStack,
}

type ȧutoGeneratedMergedStackExampleReader struct {
	data *mergedStackExample
}

func (m *ȧutoGeneratedMergedStackExampleReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMergedStackExampleReaderProps
}

func (m *ȧutoGeneratedMergedStackExampleReader) Prop(name string) (interface{}, error) {
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

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMergedStackExampleReader) PropMutable(name string) bool {
	switch name {
	case "MyMergedStack":
		return false
	}

	return false
}

func (m *ȧutoGeneratedMergedStackExampleReader) SetProp(name string, value interface{}) error {
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMergedStackExampleReader) ConfigureProp(name string, value interface{}) error {
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

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (m *ȧutoGeneratedMergedStackExampleReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	switch name {
	case "MyMergedStack":
		return m.data.MyMergedStack, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "MyMergedStack":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "MyMergedStack":
		slotValue := value.MergedStack()
		if slotValue == nil {
			return errors.New("MyMergedStack couldn't be upconverted, returned nil")
		}
		m.data.MyMergedStack = slotValue
		return nil

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "MyMergedStack":
		return nil, boardgame.ErrPropertyImmutable

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMergedStackExampleReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for mergedStackExample
func (m *mergedStackExample) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMergedStackExampleReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for mergedStackExample
func (m *mergedStackExample) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMergedStackExampleReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for mergedStackExample
func (m *mergedStackExample) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMergedStackExampleReader{m}
}

// Implementation for rangeValExample

var ȧutoGeneratedRangeValExampleReaderProps = map[string]boardgame.PropertyType{
	"MyMutableRangeVal": boardgame.TypeEnum,
	"MyRangeVal":        boardgame.TypeEnum,
}

type ȧutoGeneratedRangeValExampleReader struct {
	data *rangeValExample
}

func (r *ȧutoGeneratedRangeValExampleReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedRangeValExampleReaderProps
}

func (r *ȧutoGeneratedRangeValExampleReader) Prop(name string) (interface{}, error) {
	props := r.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeInt:
		return r.IntProp(name)
	case boardgame.TypeBool:
		return r.BoolProp(name)
	case boardgame.TypeString:
		return r.StringProp(name)
	case boardgame.TypePlayerIndex:
		return r.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return r.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return r.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return r.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return r.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return r.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return r.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return r.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return r.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (r *ȧutoGeneratedRangeValExampleReader) PropMutable(name string) bool {
	switch name {
	case "MyMutableRangeVal":
		return true
	case "MyRangeVal":
		return false
	}

	return false
}

func (r *ȧutoGeneratedRangeValExampleReader) SetProp(name string, value interface{}) error {
	props := r.Props()
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
		return r.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return r.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return r.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return r.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return r.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return r.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return r.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return r.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (r *ȧutoGeneratedRangeValExampleReader) ConfigureProp(name string, value interface{}) error {
	props := r.Props()
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
		return r.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return r.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return r.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return r.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if r.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return r.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return r.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return r.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return r.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return r.SetStringSliceProp(name, val)
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
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return r.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if r.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return r.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return r.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if r.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return r.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return r.ConfigureImmutableTimerProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (r *ȧutoGeneratedRangeValExampleReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "MyMutableRangeVal":
		return r.data.MyMutableRangeVal, nil
	case "MyRangeVal":
		return r.data.MyRangeVal, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "MyMutableRangeVal":
		slotValue := value.RangeVal()
		if slotValue == nil {
			return errors.New("MyMutableRangeVal couldn't be upconverted, returned nil")
		}
		r.data.MyMutableRangeVal = slotValue
		return nil
	case "MyRangeVal":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such Enum prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "MyMutableRangeVal":
		return boardgame.ErrPropertyImmutable
	case "MyRangeVal":
		slotValue := value.ImmutableRangeVal()
		if slotValue == nil {
			return errors.New("MyRangeVal couldn't be upconverted, returned nil")
		}
		r.data.MyRangeVal = slotValue
		return nil

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "MyMutableRangeVal":
		return r.data.MyMutableRangeVal, nil
	case "MyRangeVal":
		return nil, boardgame.ErrPropertyImmutable

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (r *ȧutoGeneratedRangeValExampleReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for rangeValExample
func (r *rangeValExample) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedRangeValExampleReader{r}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for rangeValExample
func (r *rangeValExample) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedRangeValExampleReader{r}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for rangeValExample
func (r *rangeValExample) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedRangeValExampleReader{r}
}

// Implementation for treeValExample

var ȧutoGeneratedTreeValExampleReaderProps = map[string]boardgame.PropertyType{
	"MyImmutableTreeVal": boardgame.TypeEnum,
	"MyTreeVal":          boardgame.TypeEnum,
}

type ȧutoGeneratedTreeValExampleReader struct {
	data *treeValExample
}

func (t *ȧutoGeneratedTreeValExampleReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedTreeValExampleReaderProps
}

func (t *ȧutoGeneratedTreeValExampleReader) Prop(name string) (interface{}, error) {
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

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (t *ȧutoGeneratedTreeValExampleReader) PropMutable(name string) bool {
	switch name {
	case "MyImmutableTreeVal":
		return false
	case "MyTreeVal":
		return true
	}

	return false
}

func (t *ȧutoGeneratedTreeValExampleReader) SetProp(name string, value interface{}) error {
	props := t.Props()
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
		return t.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return t.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return t.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return t.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return t.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return t.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return t.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return t.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeBoard:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (t *ȧutoGeneratedTreeValExampleReader) ConfigureProp(name string, value interface{}) error {
	props := t.Props()
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
		return t.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return t.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return t.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return t.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if t.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return t.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return t.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return t.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return t.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return t.SetStringSliceProp(name, val)
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
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return t.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if t.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return t.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return t.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if t.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return t.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return t.ConfigureImmutableTimerProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (t *ȧutoGeneratedTreeValExampleReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	switch name {
	case "MyImmutableTreeVal":
		return t.data.MyImmutableTreeVal, nil
	case "MyTreeVal":
		return t.data.MyTreeVal, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "MyImmutableTreeVal":
		return boardgame.ErrPropertyImmutable
	case "MyTreeVal":
		slotValue := value.TreeVal()
		if slotValue == nil {
			return errors.New("MyTreeVal couldn't be upconverted, returned nil")
		}
		t.data.MyTreeVal = slotValue
		return nil

	}

	return errors.New("No such Enum prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	switch name {
	case "MyImmutableTreeVal":
		slotValue := value.ImmutableTreeVal()
		if slotValue == nil {
			return errors.New("MyImmutableTreeVal couldn't be upconverted, returned nil")
		}
		t.data.MyImmutableTreeVal = slotValue
		return nil
	case "MyTreeVal":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "MyImmutableTreeVal":
		return nil, boardgame.ErrPropertyImmutable
	case "MyTreeVal":
		return t.data.MyTreeVal, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (t *ȧutoGeneratedTreeValExampleReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for treeValExample
func (t *treeValExample) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedTreeValExampleReader{t}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for treeValExample
func (t *treeValExample) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedTreeValExampleReader{t}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for treeValExample
func (t *treeValExample) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedTreeValExampleReader{t}
}
