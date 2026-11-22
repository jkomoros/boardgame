/************************************
 *
 * This file contains auto-generated methods to help certain structs
 * implement boardgame.PropertyReader and friends. It was generated
 * by the codegen package via 'boardgame-util codegen'.
 *
 * DO NOT EDIT by hand.
 *
 ************************************/

package emitgame

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// Implementation for gameState

var ȧutoGeneratedGameStateReaderProps = map[string]boardgame.PropertyType{
	"Board":         boardgame.TypeStack,
	"Cards":         boardgame.TypeStack,
	"Count":         boardgame.TypeInt,
	"CurrentPlayer": boardgame.TypePlayerIndex,
	"Discards":      boardgame.TypeStack,
	"Label":         boardgame.TypeString,
	"Phase":         boardgame.TypeEnum,
	"Scores":        boardgame.TypeIntSlice,
	"Timer":         boardgame.TypeTimer,
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
	case "Board":
		return true
	case "Cards":
		return true
	case "Count":
		return true
	case "CurrentPlayer":
		return true
	case "Discards":
		return true
	case "Label":
		return true
	case "Phase":
		return true
	case "Scores":
		return true
	case "Timer":
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
	case "Count":
		return g.data.Count, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetIntProp(name string, value int) error {

	switch name {
	case "Count":
		g.data.Count = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) StringProp(name string) (string, error) {

	switch name {
	case "Label":
		return g.data.Label, nil

	}

	return "", errors.New("No such String prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetStringProp(name string, value string) error {

	switch name {
	case "Label":
		g.data.Label = value
		return nil

	}

	return errors.New("No such String prop: " + name)

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

	switch name {
	case "Scores":
		return g.data.Scores, nil

	}

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) SetIntSliceProp(name string, value []int) error {

	switch name {
	case "Scores":
		g.data.Scores = value
		return nil

	}

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
	case "Board":
		return g.data.Board, nil
	case "Cards":
		return g.data.Cards, nil
	case "Discards":
		return g.data.Discards, nil

	}

	return nil, errors.New("No such Stack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	switch name {
	case "Board":
		slotValue := value.SizedStack()
		if slotValue == nil {
			return errors.New("Board couldn't be upconverted, returned nil")
		}
		g.data.Board = slotValue
		return nil
	case "Cards":
		g.data.Cards = value
		return nil
	case "Discards":
		g.data.Discards = value
		return nil

	}

	return errors.New("No such Stack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	switch name {
	case "Board":
		return boardgame.ErrPropertyImmutable
	case "Cards":
		return boardgame.ErrPropertyImmutable
	case "Discards":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableStack prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) StackProp(name string) (boardgame.Stack, error) {

	switch name {
	case "Board":
		return g.data.Board, nil
	case "Cards":
		return g.data.Cards, nil
	case "Discards":
		return g.data.Discards, nil

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

	switch name {
	case "Timer":
		return g.data.Timer, nil

	}

	return nil, errors.New("No such Timer prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	switch name {
	case "Timer":
		g.data.Timer = value
		return nil

	}

	return errors.New("No such Timer prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	switch name {
	case "Timer":
		return boardgame.ErrPropertyImmutable

	}

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (g *ȧutoGeneratedGameStateReader) TimerProp(name string) (boardgame.Timer, error) {

	switch name {
	case "Timer":
		return g.data.Timer, nil

	}

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
	"Done":           boardgame.TypeBool,
	"Eliminated":     boardgame.TypeBool,
	"Hand":           boardgame.TypeStack,
	"PlayerInactive": boardgame.TypeBool,
	"Score":          boardgame.TypeInt,
	"SeatClosed":     boardgame.TypeBool,
	"SeatFilled":     boardgame.TypeBool,
	"Votes":          boardgame.TypePlayerIndexSlice,
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
	case "Done":
		return true
	case "Eliminated":
		return true
	case "Hand":
		return true
	case "PlayerInactive":
		return true
	case "Score":
		return true
	case "SeatClosed":
		return true
	case "SeatFilled":
		return true
	case "Votes":
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
	case "Done":
		return p.data.Done, nil
	case "Eliminated":
		return p.data.Eliminated, nil
	case "PlayerInactive":
		return p.data.PlayerInactive, nil
	case "SeatClosed":
		return p.data.SeatClosed, nil
	case "SeatFilled":
		return p.data.SeatFilled, nil

	}

	return false, errors.New("No such Bool prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetBoolProp(name string, value bool) error {

	switch name {
	case "Done":
		p.data.Done = value
		return nil
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

	switch name {
	case "Votes":
		return p.data.Votes, nil

	}

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (p *ȧutoGeneratedPlayerStateReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	switch name {
	case "Votes":
		p.data.Votes = value
		return nil

	}

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

// Implementation for cardValue

var ȧutoGeneratedCardValueReaderProps = map[string]boardgame.PropertyType{
	"Name":  boardgame.TypeString,
	"Value": boardgame.TypeInt,
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
	case boardgame.TypeInt:
		return c.IntProp(name)
	case boardgame.TypeBool:
		return c.BoolProp(name)
	case boardgame.TypeString:
		return c.StringProp(name)
	case boardgame.TypePlayerIndex:
		return c.PlayerIndexProp(name)
	case boardgame.TypeEnum:
		return c.ImmutableEnumProp(name)
	case boardgame.TypeIntSlice:
		return c.IntSliceProp(name)
	case boardgame.TypeBoolSlice:
		return c.BoolSliceProp(name)
	case boardgame.TypeStringSlice:
		return c.StringSliceProp(name)
	case boardgame.TypePlayerIndexSlice:
		return c.PlayerIndexSliceProp(name)
	case boardgame.TypeStack:
		return c.ImmutableStackProp(name)
	case boardgame.TypeBoard:
		return c.ImmutableBoardProp(name)
	case boardgame.TypeTimer:
		return c.ImmutableTimerProp(name)
	case boardgame.TypeEnumSlice:
		return c.ImmutableEnumSliceProp(name)

	}

	return nil, errors.New("Unexpected property type: " + propType.String())
}

func (c *ȧutoGeneratedCardValueReader) PropMutable(name string) bool {
	switch name {
	case "Name":
		return true
	case "Value":
		return true
	}

	return false
}

func (c *ȧutoGeneratedCardValueReader) SetProp(name string, value interface{}) error {
	props := c.Props()
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
		return c.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return c.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return c.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		return errors.New("SetProp does not allow setting mutable types; use ConfigureProp instead")
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return c.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return c.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return c.SetStringSliceProp(name, val)
	case boardgame.TypePlayerIndexSlice:
		val, ok := value.([]boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type []boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexSliceProp(name, val)
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

func (c *ȧutoGeneratedCardValueReader) ConfigureProp(name string, value interface{}) error {
	props := c.Props()
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
		return c.SetIntProp(name, val)
	case boardgame.TypeBool:
		val, ok := value.(bool)
		if !ok {
			return errors.New("Provided value was not of type bool")
		}
		return c.SetBoolProp(name, val)
	case boardgame.TypeString:
		val, ok := value.(string)
		if !ok {
			return errors.New("Provided value was not of type string")
		}
		return c.SetStringProp(name, val)
	case boardgame.TypePlayerIndex:
		val, ok := value.(boardgame.PlayerIndex)
		if !ok {
			return errors.New("Provided value was not of type boardgame.PlayerIndex")
		}
		return c.SetPlayerIndexProp(name, val)
	case boardgame.TypeEnum:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.Val)
			if !ok {
				return errors.New("Provided value was not of type enum.Val")
			}
			return c.ConfigureEnumProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableVal)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableVal")
		}
		return c.ConfigureImmutableEnumProp(name, val)
	case boardgame.TypeIntSlice:
		val, ok := value.([]int)
		if !ok {
			return errors.New("Provided value was not of type []int")
		}
		return c.SetIntSliceProp(name, val)
	case boardgame.TypeBoolSlice:
		val, ok := value.([]bool)
		if !ok {
			return errors.New("Provided value was not of type []bool")
		}
		return c.SetBoolSliceProp(name, val)
	case boardgame.TypeStringSlice:
		val, ok := value.([]string)
		if !ok {
			return errors.New("Provided value was not of type []string")
		}
		return c.SetStringSliceProp(name, val)
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
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableStack)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableStack")
		}
		return c.ConfigureImmutableStackProp(name, val)
	case boardgame.TypeBoard:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Board)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Board")
			}
			return c.ConfigureBoardProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableBoard)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableBoard")
		}
		return c.ConfigureImmutableBoardProp(name, val)
	case boardgame.TypeTimer:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(boardgame.Timer)
			if !ok {
				return errors.New("Provided value was not of type boardgame.Timer")
			}
			return c.ConfigureTimerProp(name, val)
		}
		//Immutable variant
		val, ok := value.(boardgame.ImmutableTimer)
		if !ok {
			return errors.New("Provided value was not of type boardgame.ImmutableTimer")
		}
		return c.ConfigureImmutableTimerProp(name, val)
	case boardgame.TypeEnumSlice:
		if c.PropMutable(name) {
			//Mutable variant
			val, ok := value.(enum.EnumSlice)
			if !ok {
				return errors.New("Provided value was not of type enum.EnumSlice")
			}
			return c.ConfigureEnumSliceProp(name, val)
		}
		//Immutable variant
		val, ok := value.(enum.ImmutableEnumSlice)
		if !ok {
			return errors.New("Provided value was not of type enum.ImmutableEnumSlice")
		}
		return c.ConfigureImmutableEnumSliceProp(name, val)

	}

	return errors.New("Unexpected property type: " + propType.String())
}

func (c *ȧutoGeneratedCardValueReader) IntProp(name string) (int, error) {

	switch name {
	case "Value":
		return c.data.Value, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) SetIntProp(name string, value int) error {

	switch name {
	case "Value":
		c.data.Value = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) StringProp(name string) (string, error) {

	switch name {
	case "Name":
		return c.data.Name, nil

	}

	return "", errors.New("No such String prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) SetStringProp(name string, value string) error {

	switch name {
	case "Name":
		c.data.Name = value
		return nil

	}

	return errors.New("No such String prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (c *ȧutoGeneratedCardValueReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for cardValue
func (c *cardValue) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedCardValueReader{c}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for cardValue
func (c *cardValue) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedCardValueReader{c}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for cardValue
func (c *cardValue) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedCardValueReader{c}
}

// Implementation for moveAll

var ȧutoGeneratedMoveAllReaderProps = map[string]boardgame.PropertyType{
	"Amount":            boardgame.TypeInt,
	"TargetPlayerIndex": boardgame.TypePlayerIndex,
}

type ȧutoGeneratedMoveAllReader struct {
	data *moveAll
}

func (m *ȧutoGeneratedMoveAllReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveAllReaderProps
}

func (m *ȧutoGeneratedMoveAllReader) Prop(name string) (interface{}, error) {
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

func (m *ȧutoGeneratedMoveAllReader) PropMutable(name string) bool {
	switch name {
	case "Amount":
		return true
	case "TargetPlayerIndex":
		return true
	}

	return false
}

func (m *ȧutoGeneratedMoveAllReader) SetProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveAllReader) ConfigureProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveAllReader) IntProp(name string) (int, error) {

	switch name {
	case "Amount":
		return m.data.Amount, nil

	}

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) SetIntProp(name string, value int) error {

	switch name {
	case "Amount":
		m.data.Amount = value
		return nil

	}

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	switch name {
	case "TargetPlayerIndex":
		return m.data.TargetPlayerIndex, nil

	}

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	switch name {
	case "TargetPlayerIndex":
		m.data.TargetPlayerIndex = value
		return nil

	}

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveAllReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for moveAll
func (m *moveAll) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveAllReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveAll
func (m *moveAll) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveAllReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveAll
func (m *moveAll) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveAllReader{m}
}

// Implementation for moveFinish

var ȧutoGeneratedMoveFinishReaderProps = map[string]boardgame.PropertyType{}

type ȧutoGeneratedMoveFinishReader struct {
	data *moveFinish
}

func (m *ȧutoGeneratedMoveFinishReader) Props() map[string]boardgame.PropertyType {
	return ȧutoGeneratedMoveFinishReaderProps
}

func (m *ȧutoGeneratedMoveFinishReader) Prop(name string) (interface{}, error) {
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

func (m *ȧutoGeneratedMoveFinishReader) PropMutable(name string) bool {
	switch name {
	}

	return false
}

func (m *ȧutoGeneratedMoveFinishReader) SetProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveFinishReader) ConfigureProp(name string, value interface{}) error {
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

func (m *ȧutoGeneratedMoveFinishReader) IntProp(name string) (int, error) {

	return 0, errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) SetIntProp(name string, value int) error {

	return errors.New("No such Int prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) BoolProp(name string) (bool, error) {

	return false, errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) SetBoolProp(name string, value bool) error {

	return errors.New("No such Bool prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) StringProp(name string) (string, error) {

	return "", errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) SetStringProp(name string, value string) error {

	return errors.New("No such String prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) PlayerIndexProp(name string) (boardgame.PlayerIndex, error) {

	return 0, errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) SetPlayerIndexProp(name string, value boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndex prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ImmutableEnumProp(name string) (enum.ImmutableVal, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ConfigureEnumProp(name string, value enum.Val) error {

	return errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ConfigureImmutableEnumProp(name string, value enum.ImmutableVal) error {

	return errors.New("No such ImmutableEnum prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) EnumProp(name string) (enum.Val, error) {

	return nil, errors.New("No such Enum prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) IntSliceProp(name string) ([]int, error) {

	return []int{}, errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) SetIntSliceProp(name string, value []int) error {

	return errors.New("No such IntSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) BoolSliceProp(name string) ([]bool, error) {

	return []bool{}, errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) SetBoolSliceProp(name string, value []bool) error {

	return errors.New("No such BoolSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) StringSliceProp(name string) ([]string, error) {

	return []string{}, errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) SetStringSliceProp(name string, value []string) error {

	return errors.New("No such StringSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) PlayerIndexSliceProp(name string) ([]boardgame.PlayerIndex, error) {

	return []boardgame.PlayerIndex{}, errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) SetPlayerIndexSliceProp(name string, value []boardgame.PlayerIndex) error {

	return errors.New("No such PlayerIndexSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ImmutableStackProp(name string) (boardgame.ImmutableStack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ConfigureStackProp(name string, value boardgame.Stack) error {

	return errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ConfigureImmutableStackProp(name string, value boardgame.ImmutableStack) error {

	return errors.New("No such ImmutableStack prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) StackProp(name string) (boardgame.Stack, error) {

	return nil, errors.New("No such Stack prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ImmutableBoardProp(name string) (boardgame.ImmutableBoard, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ConfigureBoardProp(name string, value boardgame.Board) error {

	return errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ConfigureImmutableBoardProp(name string, value boardgame.ImmutableBoard) error {

	return errors.New("No such ImmutableBoard prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) BoardProp(name string) (boardgame.Board, error) {

	return nil, errors.New("No such Board prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ImmutableTimerProp(name string) (boardgame.ImmutableTimer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ConfigureTimerProp(name string, value boardgame.Timer) error {

	return errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ConfigureImmutableTimerProp(name string, value boardgame.ImmutableTimer) error {

	return errors.New("No such ImmutableTimer prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) TimerProp(name string) (boardgame.Timer, error) {

	return nil, errors.New("No such Timer prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ImmutableEnumSliceProp(name string) (enum.ImmutableEnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ConfigureEnumSliceProp(name string, value enum.EnumSlice) error {

	return errors.New("No such EnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) ConfigureImmutableEnumSliceProp(name string, value enum.ImmutableEnumSlice) error {

	return errors.New("No such ImmutableEnumSlice prop: " + name)

}

func (m *ȧutoGeneratedMoveFinishReader) EnumSliceProp(name string) (enum.EnumSlice, error) {

	return nil, errors.New("No such EnumSlice prop: " + name)

}

// Reader returns an autp-generated boardgame.PropertyReader for moveFinish
func (m *moveFinish) Reader() boardgame.PropertyReader {
	return &ȧutoGeneratedMoveFinishReader{m}
}

// ReadSetter returns an autp-generated boardgame.PropertyReadSetter for moveFinish
func (m *moveFinish) ReadSetter() boardgame.PropertyReadSetter {
	return &ȧutoGeneratedMoveFinishReader{m}
}

// ReadSetConfigurer returns an autp-generated boardgame.PropertyReadSetConfigurer for moveFinish
func (m *moveFinish) ReadSetConfigurer() boardgame.PropertyReadSetConfigurer {
	return &ȧutoGeneratedMoveFinishReader{m}
}
