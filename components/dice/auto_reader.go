/************************************
 *
 * This file contains auto-generated methods to help certain structs
 * implement boardgame.PropertyReader and friends. It was generated
 * by the codegen package via 'boardgame-util codegen'.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package dice

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// Implementation for Value

var ȧutoGeneratedValueReaderProps = map[string]boardgame.PropertyType{
	"Faces": boardgame.TypeIntSlice,
}

type ȧutoGeneratedValueReader struct {
	data *Value
}

func (v *ȧutoGeneratedValueReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedValueReaderProps
}

func (v *ȧutoGeneratedValueReader) Prop(name string) (interface{}, error) {
	props := v.Props()
	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		return v.ImmutableBoardProp(name)
	case boardgame.TypeBool:
		return v.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return v.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return v.ImmutableEnumProp(name)
	case boardgame.TypeInt:
		return v.IntProp(name)
	case boardgame.TypeIntSlice:
		return v.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return v.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return v.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return v.ImmutableStackProp(name)
	case boardgame.TypeString:
		return v.StringProp(name)
	case boardgame.TypeStringSlice:
		return v.StringSliceProp(name)
	case boardgame.TypeTimer:
		return v.ImmutableTimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (v *ȧutoGeneratedValueReader) PropMutable(name string) bool {
	switch name {
	case "Faces":
		return true
	}

	return false
}

func (v *ȧutoGeneratedValueReader) SetProp(name string, value interface{}) error {
	props := v.Props()
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
		return v.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return v.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return v.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return v.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return v.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return v.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return v.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return v.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (v *ȧutoGeneratedValueReader) ConfigureProp(name string, value interface{}) error {
	props := v.Props()
	propType, ok := props[name]

	if !ok {
		return errors.New("No such property with that name: " + name)
	}

	switch propType {
	case boardgame.TypeBoard:
		if v.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return v.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return v.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return v.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return v.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		if v.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return v.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return v.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeInt:
		val, ok := value.(int)
		if !ok {
			return errors.New("Provided value was not of type int")
		}
		return v.SetIntProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return v.SetIntSliceProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return v.SetPlayerIndexProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return v.SetPlayerIndexSliceProp(name, val)
	case boardgame.TypeStack:
		if v.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return v.ConfigureStackProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return v.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return v.SetStringProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return v.SetStringSliceProp(name, val)
	case boardgame.TypeTimer:
		if v.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return v.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return v.ConfigureImmutableTimerProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (v *ȧutoGeneratedValueReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (v *ȧutoGeneratedValueReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (v *ȧutoGeneratedValueReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (v *ȧutoGeneratedValueReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (v *ȧutoGeneratedValueReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (v *ȧutoGeneratedValueReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (v *ȧutoGeneratedValueReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (v *ȧutoGeneratedValueReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (v *ȧutoGeneratedValueReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (v *ȧutoGeneratedValueReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (v *ȧutoGeneratedValueReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (v *ȧutoGeneratedValueReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (v *ȧutoGeneratedValueReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (v *ȧutoGeneratedValueReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (v *ȧutoGeneratedValueReader) IntSliceProp(name string) ([]int, error) {

	switch name {
	case "Faces":
		return v.data.Faces, nil

	}

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (v *ȧutoGeneratedValueReader) SetIntSliceProp(name string, value []int) error {

	switch name {
	case "Faces":
		v.data.Faces = value
		return nil

	}

	return errors.New("No such IntSlice prop: " + name)

}

func (v *ȧutoGeneratedValueReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (v *ȧutoGeneratedValueReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (v *ȧutoGeneratedValueReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (v *ȧutoGeneratedValueReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (v *ȧutoGeneratedValueReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (v *ȧutoGeneratedValueReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (v *ȧutoGeneratedValueReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (v *ȧutoGeneratedValueReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (v *ȧutoGeneratedValueReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (v *ȧutoGeneratedValueReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (v *ȧutoGeneratedValueReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (v *ȧutoGeneratedValueReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (v *ȧutoGeneratedValueReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (v *ȧutoGeneratedValueReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (v *ȧutoGeneratedValueReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (v *ȧutoGeneratedValueReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for Value
func (v *Value) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedValueReader{v}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for Value
func (v *Value) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedValueReader{v}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for Value
func (v *Value) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedValueReader{v}
}

// Implementation for DynamicValue

var ȧutoGeneratedDynamicValueReaderProps = map[string]boardgame.PropertyType{
	"SelectedFace": boardgame.TypeInt,
	"Value":        boardgame.TypeInt,
}

type ȧutoGeneratedDynamicValueReader struct {
	data *DynamicValue
}

func (d *ȧutoGeneratedDynamicValueReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedDynamicValueReaderProps
}

func (d *ȧutoGeneratedDynamicValueReader) Prop(name string) (interface{}, error) {
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

func (d *ȧutoGeneratedDynamicValueReader) PropMutable(name string) bool {
	switch name {
	case "SelectedFace":
		return true
	case "Value":
		return true
	}

	return false
}

func (d *ȧutoGeneratedDynamicValueReader) SetProp(name string, value interface{}) error {
	props := d.Props()
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
		return d.SetBoolProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return d.SetBoolSliceProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
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
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
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
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (d *ȧutoGeneratedDynamicValueReader) ConfigureProp(name string, value interface{}) error {
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
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return d.ConfigureImmutableBoardProp(name, val)
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
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return d.ConfigureImmutableEnumProp(name, val)
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
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return d.ConfigureImmutableStackProp(name, val)
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

func (d *ȧutoGeneratedDynamicValueReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) IntProp(name string) (int, error) {

	switch name {
	case "SelectedFace":
		return d.data.SelectedFace, nil
	case "Value":
		return d.data.Value, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) SetIntProp(name string, value int) error {

	switch name {
	case "SelectedFace":
		d.data.SelectedFace = value
		return nil
	case "Value":
		d.data.Value = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (d *ȧutoGeneratedDynamicValueReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

//Reader returns an autp-generated boardgame.PropertyReader for DynamicValue
func (d *DynamicValue) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedDynamicValueReader{d}
}

//ReadSetter returns an autp-generated boardgame.PropertyReadSetter for DynamicValue
func (d *DynamicValue) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedDynamicValueReader{d}
}

//ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for DynamicValue
func (d *DynamicValue) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedDynamicValueReader{d}
}
