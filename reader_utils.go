package boardgame

import (
	"github.com/jkomoros/boardgame/errors"
	"strconv"
)

//stacksForReader returns all stacks in reader, inclduing all StackProps, and all stacks within Boards.
func stacksForReader(reader PropertyReader) []ImmutableStack {
	var result []ImmutableStack

	for propName, propType := range reader.Props() {
		if propType == TypeStack {
			stack, err := reader.ImmutableStackProp(propName)
			if err != nil {
				continue
			}
			result = append(result, stack)
		} else if propType == TypeBoard {
			board, err := reader.ImmutableBoardProp(propName)
			if err != nil {
				continue
			}
			for _, stack := range board.ImmutableSpaces() {
				result = append(result, stack)
			}
		}
	}

	return result
}

func setReaderStatePtr(reader PropertyReader, st ImmutableState) error {

	statePtr, ok := st.(*state)
	if !ok {
		return errors.New("The provided non-nil State could not be conveted to a state ptr")
	}

	for propName, propType := range reader.Props() {
		switch propType {
		case TypeStack:
			val, err := reader.ImmutableStackProp(propName)
			if val == nil {
				return errors.New("Stack Prop " + propName + " was nil")
			}
			if err != nil {
				return errors.New("Stack prop " + propName + " had unexpected error: " + err.Error())
			}
			val.setState(statePtr)
		case TypeBoard:
			val, err := reader.ImmutableBoardProp(propName)
			if val == nil {
				return errors.New("Board Prop " + propName + " was nil")
			}
			if err != nil {
				return errors.New("Board prop " + propName + " had unexpected error: " + err.Error())
			}
			val.setState(statePtr)
		case TypeTimer:
			val, err := reader.ImmutableTimerProp(propName)
			if val == nil {
				return errors.New("TimerProp " + propName + " was nil")
			}
			if err != nil {
				return errors.New("TimerProp " + propName + " had unexpected error: " + err.Error())
			}
			val.setState(statePtr)
		}
	}
	return nil
}

//copyReader assumes input and output container are the same "shape" (that is,
//outputContainer can have all of input's properties set). It also assumes
//that the output has all interface types initalized to the right shape.
func copyReader(input PropertyReadSetter, outputContainer PropertyReadSetter) error {

	for propName, propType := range input.Props() {
		switch propType {
		case TypeBool:
			boolVal, err := input.BoolProp(propName)
			if err != nil {
				return errors.New(propName + " did not return a bool as expected: " + err.Error())
			}
			err = outputContainer.SetBoolProp(propName, boolVal)
			if err != nil {
				return errors.New(propName + " could not be set on output: " + err.Error())
			}
		case TypeInt:
			intVal, err := input.IntProp(propName)
			if err != nil {
				return errors.New(propName + " did not return an int as expected: " + err.Error())
			}
			err = outputContainer.SetIntProp(propName, intVal)
			if err != nil {
				return errors.New(propName + " could not be set on output: " + err.Error())
			}
		case TypeString:
			stringVal, err := input.StringProp(propName)
			if err != nil {
				return errors.New(propName + " did not return a string as expected: " + err.Error())
			}
			err = outputContainer.SetStringProp(propName, stringVal)
			if err != nil {
				return errors.New(propName + " could not be set on output: " + err.Error())
			}
		case TypePlayerIndex:
			playerIndexVal, err := input.PlayerIndexProp(propName)
			if err != nil {
				return errors.New(propName + " did not return a player index as expected: " + err.Error())
			}
			err = outputContainer.SetPlayerIndexProp(propName, playerIndexVal)
			if err != nil {
				return errors.New(propName + " could not be set on output: " + err.Error())
			}
		case TypeIntSlice:
			intSliceVal, err := input.IntSliceProp(propName)
			if err != nil {
				return errors.New(propName + " did not return an int slice as expected: " + err.Error())
			}
			err = outputContainer.SetIntSliceProp(propName, intSliceVal)
			if err != nil {
				return errors.New(propName + " could not be set on output: " + err.Error())
			}
		case TypeBoolSlice:
			boolSliceVal, err := input.BoolSliceProp(propName)
			if err != nil {
				return errors.New(propName + " did not return an bool slice as expected: " + err.Error())
			}
			err = outputContainer.SetBoolSliceProp(propName, boolSliceVal)
			if err != nil {
				return errors.New(propName + " could not be set on output: " + err.Error())
			}
		case TypeStringSlice:
			stringSliceVal, err := input.StringSliceProp(propName)
			if err != nil {
				return errors.New(propName + " did not return an string slice as expected: " + err.Error())
			}
			err = outputContainer.SetStringSliceProp(propName, stringSliceVal)
			if err != nil {
				return errors.New(propName + " could not be set on output: " + err.Error())
			}
		case TypePlayerIndexSlice:
			playerIndexSliceVal, err := input.PlayerIndexSliceProp(propName)
			if err != nil {
				return errors.New(propName + " did not return a player index slice as expected: " + err.Error())
			}
			err = outputContainer.SetPlayerIndexSliceProp(propName, playerIndexSliceVal)
			if err != nil {
				return errors.New(propName + " could not be set on output: " + err.Error())
			}
		case TypeEnum:
			enumConst, err := input.EnumProp(propName)
			if err != nil {
				//if the err is ErrPropertyImmutable, that's OK, just skip
				if err == ErrPropertyImmutable {
					continue
				}
				return errors.New(propName + " did not return an EnumVal as expected: " + err.Error())
			}
			outputEnum, err := outputContainer.EnumProp(propName)
			if err != nil {
				//if the err is ErrPropertyImmutable, that's OK, just skip
				if err == ErrPropertyImmutable {
					continue
				}
				return errors.New(propName + " could not get mutable enum on output: " + err.Error())
			}
			outputEnum.SetValue(enumConst.Value())
		case TypeStack:
			stackVal, err := input.StackProp(propName)
			if err != nil {
				//if the err is ErrPropertyImmutable, that's OK, just skip
				if err == ErrPropertyImmutable {
					continue
				}
				return errors.New(propName + " did not return a stack as expected: " + err.Error())
			}
			outputStack, err := outputContainer.StackProp(propName)
			if err != nil {
				//if the err is ErrPropertyImmutable, that's OK, just skip
				if err == ErrPropertyImmutable {
					continue
				}
				return errors.New(propName + " could not get mutable stack on output: " + err.Error())
			}
			if err := outputStack.importFrom(stackVal); err != nil {
				return errors.New(propName + " could not import from input: " + err.Error())
			}
		case TypeBoard:
			boardVal, err := input.BoardProp(propName)
			if err != nil {
				//if the err is ErrPropertyImmutable, that's OK, just skip
				if err == ErrPropertyImmutable {
					continue
				}
				return errors.New(propName + " did not return a board as expected: " + err.Error())
			}
			outputBoard, err := outputContainer.BoardProp(propName)
			if err != nil {
				//if the err is ErrPropertyImmutable, that's OK, just skip
				if err == ErrPropertyImmutable {
					continue
				}
				return errors.New(propName + " could not get mutable board on output: " + err.Error())
			}
			if err := outputBoard.importFrom(boardVal); err != nil {
				return errors.New(propName + " could not import from input: " + err.Error())
			}
		case TypeTimer:
			timerVal, err := input.TimerProp(propName)
			if err != nil {
				//if the err is ErrPropertyImmutable, that's OK, just skip
				if err == ErrPropertyImmutable {
					continue
				}
				return errors.New(propName + " did not return a timer as expected: " + err.Error())
			}
			outputTimer, err := outputContainer.TimerProp(propName)
			if err != nil {
				//if the err is ErrPropertyImmutable, that's OK, just skip
				if err == ErrPropertyImmutable {
					continue
				}
				return errors.New(propName + " could not get mutable timer on output: " + err.Error())
			}
			if err := outputTimer.importFrom(timerVal); err != nil {
				return errors.New(propName + " could not import from input: " + err.Error())
			}
		default:
			return errors.New(propName + " was an unsupported property type: " + strconv.Itoa(int(propType)))
		}
	}

	return nil

}
