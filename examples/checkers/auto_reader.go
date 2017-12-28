/************************************
 *
 * This file contains auto-generated methods to help certain structs
 * implement boardgame.SubState and boardgame.MutableSubState. It was
 * generated by autoreader.
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
	case boardgame.TypeBool:
		return t.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return t.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return t.EnumProp(name)
	case boardgame.TypeInt:
		return t.IntProp(name)
	case boardgame.TypeIntSlice:
		return t.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return t.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return t.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return t.StackProp(name)
	case boardgame.TypeString:
		return t.StringProp(name)
	case boardgame.TypeStringSlice:
		return t.StringSliceProp(name)
	case boardgame.TypeTimer:
		return t.TimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (t *__tokenReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (t *__tokenReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (t *__tokenReader) EnumProp(name string) (enum.Val, error) {

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

func (t *__tokenReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (t *__tokenReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (t *__tokenReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (t *__tokenReader) TimerProp(name string) (boardgame.Timer, error) {

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
	case boardgame.TypeBool:
		return t.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return t.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return t.EnumProp(name)
	case boardgame.TypeInt:
		return t.IntProp(name)
	case boardgame.TypeIntSlice:
		return t.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return t.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return t.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return t.StackProp(name)
	case boardgame.TypeString:
		return t.StringProp(name)
	case boardgame.TypeStringSlice:
		return t.StringSliceProp(name)
	case boardgame.TypeTimer:
		return t.TimerProp(name)

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
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
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
	case boardgame.TypeEnum:
		if t.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.MutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.MutableVal")
			}
			return t.ConfigureMutableEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return t.ConfigureEnumProp(name, val)
		}
	case boardgame.TypeStack:
		if t.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.MutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.MutableStack")
			}
			return t.ConfigureMutableStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return t.ConfigureStackProp(name, val)
		}
	case boardgame.TypeTimer:
		if t.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.MutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.MutableTimer")
			}
			return t.ConfigureMutableTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return t.ConfigureTimerProp(name, val)
		}
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

	}

	return errors.New("Unexpected property type: " + propType.String())
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

func (t *__tokenDynamicReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (t *__tokenDynamicReader) ConfigureMutableEnumProp(name string, value enum.MutableVal) error {

	return errors.New("No such MutableEnum prop: " + name)

}

func (t *__tokenDynamicReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (t *__tokenDynamicReader) MutableEnumProp(name string) (enum.MutableVal, error) {

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

func (t *__tokenDynamicReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (t *__tokenDynamicReader) ConfigureMutableStackProp(name string, value boardgame.MutableStack) error {

	return errors.New("No such MutableStack prop: " + name)

}

func (t *__tokenDynamicReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (t *__tokenDynamicReader) MutableStackProp(name string) (boardgame.MutableStack, error) {

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

func (t *__tokenDynamicReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (t *__tokenDynamicReader) ConfigureMutableTimerProp(name string, value boardgame.MutableTimer) error {

	return errors.New("No such MutableTimer prop: " + name)

}

func (t *__tokenDynamicReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (t *__tokenDynamicReader) MutableTimerProp(name string) (boardgame.MutableTimer, error) {

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

// Implementation for gameState

var __gameStateReaderProps map[string]boardgame.PropertyType = map[string]boardgame.PropertyType{
	"Phase":        boardgame.TypeEnum,
	"Spaces":       boardgame.TypeStack,
	"UnusedTokens": boardgame.TypeStack,
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
	case boardgame.TypeBool:
		return g.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return g.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return g.EnumProp(name)
	case boardgame.TypeInt:
		return g.IntProp(name)
	case boardgame.TypeIntSlice:
		return g.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return g.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return g.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return g.StackProp(name)
	case boardgame.TypeString:
		return g.StringProp(name)
	case boardgame.TypeStringSlice:
		return g.StringSliceProp(name)
	case boardgame.TypeTimer:
		return g.TimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (g *__gameStateReader) PropMutable(name string) bool {
	switch name {
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
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
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
	case boardgame.TypeEnum:
		if g.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.MutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.MutableVal")
			}
			return g.ConfigureMutableEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return g.ConfigureEnumProp(name, val)
		}
	case boardgame.TypeStack:
		if g.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.MutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.MutableStack")
			}
			return g.ConfigureMutableStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return g.ConfigureStackProp(name, val)
		}
	case boardgame.TypeTimer:
		if g.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.MutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.MutableTimer")
			}
			return g.ConfigureMutableTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return g.ConfigureTimerProp(name, val)
		}
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

	}

	return errors.New("Unexpected property type: " + propType.String())
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

func (g *__gameStateReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "Phase":
		return g.data.Phase, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (g *__gameStateReader) ConfigureMutableEnumProp(name string, value enum.MutableVal) error {

	switch name {
	case "Phase":
		g.data.Phase = value
		return nil

	}

	return errors.New("No such MutableEnum prop: " + name)

}

func (g *__gameStateReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "Phase":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such Enum prop: " + name)

}

func (g *__gameStateReader) MutableEnumProp(name string) (enum.MutableVal, error) {

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

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (g *__gameStateReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (g *__gameStateReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (g *__gameStateReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

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

func (g *__gameStateReader) ConfigureMutableStackProp(name string, value boardgame.MutableStack) error {

	switch name {
	case "Spaces":
		g.data.Spaces = value
		return nil
	case "UnusedTokens":
		g.data.UnusedTokens = value
		return nil

	}

	return errors.New("No such MutableStack prop: " + name)

}

func (g *__gameStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "Spaces":
		return boardgame.ErrPropertyImmutable
	case "UnusedTokens":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such Stack prop: " + name)

}

func (g *__gameStateReader) MutableStackProp(name string) (boardgame.MutableStack, error) {

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

func (g *__gameStateReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (g *__gameStateReader) ConfigureMutableTimerProp(name string, value boardgame.MutableTimer) error {

	return errors.New("No such MutableTimer prop: " + name)

}

func (g *__gameStateReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (g *__gameStateReader) MutableTimerProp(name string) (boardgame.MutableTimer, error) {

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
	case boardgame.TypeBool:
		return p.BoolProp(name)
	case boardgame.TypeBoolSlice:
		return p.BoolSliceProp(name)
	case boardgame.TypeEnum:
		return p.EnumProp(name)
	case boardgame.TypeInt:
		return p.IntProp(name)
	case boardgame.TypeIntSlice:
		return p.IntSliceProp(name)
	case boardgame.TypePlayerIndex:
		return p.PlayerIndexProp(name)
	case boardgame.TypePlayerIndexSlice:
		return p.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return p.StackProp(name)
	case boardgame.TypeString:
		return p.StringProp(name)
	case boardgame.TypeStringSlice:
		return p.StringSliceProp(name)
	case boardgame.TypeTimer:
		return p.TimerProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (p *__playerStateReader) PropMutable(name string) bool {
	switch name {
	case "CapturedTokens":
		return true
	case "Color":
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
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeStack:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
	case boardgame.TypeTimer:
		return errors.New("SetProp does not allow setting mutable types. Use ConfigureProp instead.")
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
	case boardgame.TypeEnum:
		if p.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.MutableVal)
			if !ok {
				return errors.New("Provided value was not of type enum.MutableVal")
			}
			return p.ConfigureMutableEnumProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return p.ConfigureEnumProp(name, val)
		}
	case boardgame.TypeStack:
		if p.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.MutableStack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.MutableStack")
			}
			return p.ConfigureMutableStackProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.Stack)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Stack")
			}
			return p.ConfigureStackProp(name, val)
		}
	case boardgame.TypeTimer:
		if p.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.MutableTimer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.MutableTimer")
			}
			return p.ConfigureMutableTimerProp(name, val)
		} else {
			//Immutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return p.ConfigureTimerProp(name, val)
		}
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

	}

	return errors.New("Unexpected property type: " + propType.String())
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

func (p *__playerStateReader) EnumProp(name string) (enum.Val, error) {

	switch name {
	case "Color":
		return p.data.Color, nil

	}

	return nil, errors.New("No such Enum prop: " + name)

}

func (p *__playerStateReader) ConfigureMutableEnumProp(name string, value enum.MutableVal) error {

	switch name {
	case "Color":
		p.data.Color = value
		return nil

	}

	return errors.New("No such MutableEnum prop: " + name)

}

func (p *__playerStateReader) ConfigureEnumProp(name string, value enum.Val) error {

	switch name {
	case "Color":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such Enum prop: " + name)

}

func (p *__playerStateReader) MutableEnumProp(name string) (enum.MutableVal, error) {

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

func (p *__playerStateReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "CapturedTokens":
		return p.data.CapturedTokens, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (p *__playerStateReader) ConfigureMutableStackProp(name string, value boardgame.MutableStack) error {

	switch name {
	case "CapturedTokens":
		p.data.CapturedTokens = value
		return nil

	}

	return errors.New("No such MutableStack prop: " + name)

}

func (p *__playerStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "CapturedTokens":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such Stack prop: " + name)

}

func (p *__playerStateReader) MutableStackProp(name string) (boardgame.MutableStack, error) {

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

func (p *__playerStateReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (p *__playerStateReader) ConfigureMutableTimerProp(name string, value boardgame.MutableTimer) error {

	return errors.New("No such MutableTimer prop: " + name)

}

func (p *__playerStateReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (p *__playerStateReader) MutableTimerProp(name string) (boardgame.MutableTimer, error) {

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
