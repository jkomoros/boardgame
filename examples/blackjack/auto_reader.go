/************************************
 *
 * This file contains auto-generated methods to help certain structs
 * implement boardgame.PropertyReader and friends. It was generated
 * by the codegen package via 'boardgame-util codegen'.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package blackjack

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// Implementation for moveFinishTurn

var ȧutoGeneratedMoveFinishTurnReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveFinishTurnReader struct {
	data *moveFinishTurn
}

func (m *ȧutoGeneratedMoveFinishTurnReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveFinishTurnReaderProps
}

func (m *ȧutoGeneratedMoveFinishTurnReader) Prop(name string) (interface{}, error) {
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

func (m *ȧutoGeneratedMoveFinishTurnReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveFinishTurnReader) SetProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveFinishTurnReader) ConfigureProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveFinishTurnReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishTurnReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for moveFinishTurn
func (m *moveFinishTurn) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveFinishTurnReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveFinishTurn
func (m *moveFinishTurn) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveFinishTurnReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveFinishTurn
func (m *moveFinishTurn) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveFinishTurnReader{m}
}

// Implementation for moveRevealHiddenCard

var ȧutoGeneratedMoveRevealHiddenCardReaderProps = map[string]boardgame.PropertyType{
	"TargetPlayerIndex": boardgame.TypePlayerIndex,
}

type ȧutoGeneratedMoveRevealHiddenCardReader struct {
	data *moveRevealHiddenCard
}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveRevealHiddenCardReaderProps
}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) Prop(name string) (interface{}, error) {
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

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) PropMutable(name string) bool {
	switch name {
	case "TargetPlayerIndex":
		return true
	}

	return false
}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) SetProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ConfigureProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "TargetPlayerIndex":
		return m.data.TargetPlayerIndex, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "TargetPlayerIndex":
		m.data.TargetPlayerIndex = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveRevealHiddenCardReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for moveRevealHiddenCard
func (m *moveRevealHiddenCard) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveRevealHiddenCardReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveRevealHiddenCard
func (m *moveRevealHiddenCard) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveRevealHiddenCardReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveRevealHiddenCard
func (m *moveRevealHiddenCard) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveRevealHiddenCardReader{m}
}

// Implementation for moveCurrentPlayerHit

var ȧutoGeneratedMoveCurrentPlayerHitReaderProps = map[string]boardgame.PropertyType{
	"TargetPlayerIndex": boardgame.TypePlayerIndex,
}

type ȧutoGeneratedMoveCurrentPlayerHitReader struct {
	data *moveCurrentPlayerHit
}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveCurrentPlayerHitReaderProps
}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) Prop(name string) (interface{}, error) {
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

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) PropMutable(name string) bool {
	switch name {
	case "TargetPlayerIndex":
		return true
	}

	return false
}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) SetProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ConfigureProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "TargetPlayerIndex":
		return m.data.TargetPlayerIndex, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "TargetPlayerIndex":
		m.data.TargetPlayerIndex = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerHitReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for moveCurrentPlayerHit
func (m *moveCurrentPlayerHit) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveCurrentPlayerHitReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveCurrentPlayerHit
func (m *moveCurrentPlayerHit) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveCurrentPlayerHitReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveCurrentPlayerHit
func (m *moveCurrentPlayerHit) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveCurrentPlayerHitReader{m}
}

// Implementation for moveCurrentPlayerStand

var ȧutoGeneratedMoveCurrentPlayerStandReaderProps = map[string]boardgame.PropertyType{
	"TargetPlayerIndex": boardgame.TypePlayerIndex,
}

type ȧutoGeneratedMoveCurrentPlayerStandReader struct {
	data *moveCurrentPlayerStand
}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveCurrentPlayerStandReaderProps
}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) Prop(name string) (interface{}, error) {
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

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) PropMutable(name string) bool {
	switch name {
	case "TargetPlayerIndex":
		return true
	}

	return false
}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) SetProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ConfigureProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "TargetPlayerIndex":
		return m.data.TargetPlayerIndex, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "TargetPlayerIndex":
		m.data.TargetPlayerIndex = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCurrentPlayerStandReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for moveCurrentPlayerStand
func (m *moveCurrentPlayerStand) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveCurrentPlayerStandReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveCurrentPlayerStand
func (m *moveCurrentPlayerStand) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveCurrentPlayerStandReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveCurrentPlayerStand
func (m *moveCurrentPlayerStand) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveCurrentPlayerStandReader{m}
}

// Implementation for moveStartRoundCleanup

var ȧutoGeneratedMoveStartRoundCleanupReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveStartRoundCleanupReader struct {
	data *moveStartRoundCleanup
}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveStartRoundCleanupReaderProps
}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) Prop(name string) (interface{}, error) {
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

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) SetProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ConfigureProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveStartRoundCleanupReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for moveStartRoundCleanup
func (m *moveStartRoundCleanup) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveStartRoundCleanupReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveStartRoundCleanup
func (m *moveStartRoundCleanup) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveStartRoundCleanupReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveStartRoundCleanup
func (m *moveStartRoundCleanup) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveStartRoundCleanupReader{m}
}

// Implementation for moveAccumulateScores

var ȧutoGeneratedMoveAccumulateScoresReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveAccumulateScoresReader struct {
	data *moveAccumulateScores
}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveAccumulateScoresReaderProps
}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) Prop(name string) (interface{}, error) {
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

func (m *ȧutoGeneratedMoveAccumulateScoresReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) SetProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ConfigureProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveAccumulateScoresReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAccumulateScoresReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for moveAccumulateScores
func (m *moveAccumulateScores) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveAccumulateScoresReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveAccumulateScores
func (m *moveAccumulateScores) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveAccumulateScoresReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveAccumulateScores
func (m *moveAccumulateScores) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveAccumulateScoresReader{m}
}

// Implementation for moveCollectCards

var ȧutoGeneratedMoveCollectCardsReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveCollectCardsReader struct {
	data *moveCollectCards
}

func (m *ȧutoGeneratedMoveCollectCardsReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveCollectCardsReaderProps
}

func (m *ȧutoGeneratedMoveCollectCardsReader) Prop(name string) (interface{}, error) {
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

func (m *ȧutoGeneratedMoveCollectCardsReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveCollectCardsReader) SetProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveCollectCardsReader) ConfigureProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveCollectCardsReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveCollectCardsReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for moveCollectCards
func (m *moveCollectCards) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveCollectCardsReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveCollectCards
func (m *moveCollectCards) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveCollectCardsReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveCollectCards
func (m *moveCollectCards) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveCollectCardsReader{m}
}

// Implementation for moveResetPlayerForNewRound

var ȧutoGeneratedMoveResetPlayerForNewRoundReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveResetPlayerForNewRoundReader struct {
	data *moveResetPlayerForNewRound
}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveResetPlayerForNewRoundReaderProps
}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) Prop(name string) (interface{}, error) {
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

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) SetProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ConfigureProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveResetPlayerForNewRoundReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for moveResetPlayerForNewRound
func (m *moveResetPlayerForNewRound) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveResetPlayerForNewRoundReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveResetPlayerForNewRound
func (m *moveResetPlayerForNewRound) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveResetPlayerForNewRoundReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveResetPlayerForNewRound
func (m *moveResetPlayerForNewRound) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveResetPlayerForNewRoundReader{m}
}

// Implementation for moveIncrementRoundsCompleted

var ȧutoGeneratedMoveIncrementRoundsCompletedReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveIncrementRoundsCompletedReader struct {
	data *moveIncrementRoundsCompleted
}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveIncrementRoundsCompletedReaderProps
}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) Prop(name string) (interface{}, error) {
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

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) SetProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ConfigureProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveIncrementRoundsCompletedReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for moveIncrementRoundsCompleted
func (m *moveIncrementRoundsCompleted) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveIncrementRoundsCompletedReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveIncrementRoundsCompleted
func (m *moveIncrementRoundsCompleted) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveIncrementRoundsCompletedReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveIncrementRoundsCompleted
func (m *moveIncrementRoundsCompleted) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveIncrementRoundsCompletedReader{m}
}

// Implementation for gameState

var ȧutoGeneratedGameStateReaderProps = map[string]boardgame.PropertyType{
	"CurrentPlayer":   boardgame.TypePlayerIndex,
	"DiscardStack":    boardgame.TypeStack,
	"DrawStack":       boardgame.TypeStack,
	"MaxRounds":       boardgame.TypeInt,
	"Phase":           boardgame.TypeEnum,
	"RRHasStarted":    boardgame.TypeBool,
	"RRLastPlayer":    boardgame.TypePlayerIndex,
	"RRRoundCount":    boardgame.TypeInt,
	"RRStarterPlayer": boardgame.TypePlayerIndex,
	"RoundsCompleted": boardgame.TypeInt,
	"UnusedCards":     boardgame.TypeStack,
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
	case "CurrentPlayer":
		return true
	case "DiscardStack":
		return true
	case "DrawStack":
		return true
	case "MaxRounds":
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
	case "RoundsCompleted":
		return true
	case "UnusedCards":
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
	case "MaxRounds":
		return g.data.MaxRounds, nil
	case "RRRoundCount":
		return g.data.RRRoundCount, nil
	case "RoundsCompleted":
		return g.data.RoundsCompleted, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetIntProp(name string, value int) error {

	switch name {
	case "MaxRounds":
		g.data.MaxRounds = value
		return nil
	case "RRRoundCount":
		g.data.RRRoundCount = value
		return nil
	case "RoundsCompleted":
		g.data.RoundsCompleted = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) BoolProp(name string) (bool, error) {

	switch name {
	case "RRHasStarted":
		return g.data.RRHasStarted, nil

	}

	return false, errors.New("No such Bool prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetBoolProp(name string, value bool) error {

	switch name {
	case "RRHasStarted":
		g.data.RRHasStarted = value
		return nil

	}

	return errors.New("No such Bool prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

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

func (g *ȧutoGeneratedGameStateReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

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
	case "DiscardStack":
		return g.data.DiscardStack, nil
	case "DrawStack":
		return g.data.DrawStack, nil
	case "UnusedCards":
		return g.data.UnusedCards, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "DiscardStack":
		g.data.DiscardStack = value
		return nil
	case "DrawStack":
		g.data.DrawStack = value
		return nil
	case "UnusedCards":
		g.data.UnusedCards = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "DiscardStack":
		return boardgame.ErrPropertyImmutable
	case "DrawStack":
		return boardgame.ErrPropertyImmutable
	case "UnusedCards":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "DiscardStack":
		return g.data.DiscardStack, nil
	case "DrawStack":
		return g.data.DrawStack, nil
	case "UnusedCards":
		return g.data.UnusedCards, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

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
	"Eliminated":     boardgame.TypeBool,
	"Hand":           boardgame.TypeStack,
	"HiddenHand":     boardgame.TypeStack,
	"PlayerInactive": boardgame.TypeBool,
	"Score":          boardgame.TypeInt,
	"SeatClosed":     boardgame.TypeBool,
	"SeatFilled":     boardgame.TypeBool,
	"Stood":          boardgame.TypeBool,
	"VisibleHand":    boardgame.TypeStack,
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
	case "Eliminated":
		return true
	case "Hand":
		return false
	case "HiddenHand":
		return true
	case "PlayerInactive":
		return true
	case "Score":
		return true
	case "SeatClosed":
		return true
	case "SeatFilled":
		return true
	case "Stood":
		return true
	case "VisibleHand":
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
	case "Score":
		return p.data.Score, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetIntProp(name string, value int) error {

	switch name {
	case "Score":
		p.data.Score = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) BoolProp(name string) (bool, error) {

	switch name {
	case "Eliminated":
		return p.data.Eliminated, nil
	case "PlayerInactive":
		return p.data.PlayerInactive, nil
	case "SeatClosed":
		return p.data.SeatClosed, nil
	case "SeatFilled":
		return p.data.SeatFilled, nil
	case "Stood":
		return p.data.Stood, nil

	}

	return false, errors.New("No such Bool prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetBoolProp(name string, value bool) error {

	switch name {
	case "Eliminated":
		p.data.Eliminated = value
		return nil
	case "PlayerInactive":
		p.data.PlayerInactive = value
		return nil
	case "SeatClosed":
		p.data.SeatClosed = value
		return nil
	case "SeatFilled":
		p.data.SeatFilled = value
		return nil
	case "Stood":
		p.data.Stood = value
		return nil

	}

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
	case "HiddenHand":
		return p.data.HiddenHand, nil
	case "VisibleHand":
		return p.data.VisibleHand, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "Hand":
		return boardgame.ErrPropertyImmutable
	case "HiddenHand":
		p.data.HiddenHand = value
		return nil
	case "VisibleHand":
		p.data.VisibleHand = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "Hand":
		slotValue := value.MergedStack()
		if slotValue == nil {
			return errors.New("Hand couldn't be upconverted, returned nil")
		}
		p.data.Hand = slotValue
		return nil
	case "HiddenHand":
		return boardgame.ErrPropertyImmutable
	case "VisibleHand":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "Hand":
		return nil, boardgame.ErrPropertyImmutable
	case "HiddenHand":
		return p.data.HiddenHand, nil
	case "VisibleHand":
		return p.data.VisibleHand, nil

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
