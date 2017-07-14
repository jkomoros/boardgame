package boardgame

import (
	"errors"
	"github.com/jkomoros/boardgame/enum"
	"reflect"
	"strconv"
	"strings"
)

const enumStructTag = "enum"
const stackStructTag = "stack"

type autoStackConfig struct {
	deck *Deck
	size int
}

type readerValidator struct {
	autoEnumVarFields       map[string]*enum.Enum
	autoEnumConstFields     map[string]*enum.Enum
	autoGrowableStackFields map[string]*autoStackConfig
	autoSizedStackFields    map[string]*autoStackConfig
	illegalTypes            map[PropertyType]bool
}

//newReaderValidator returns a new readerValidator configured to disallow the
//given types. It will also do an expensive processing for any nil pointer-
//properties to see if they have struct tags that tell us how to inflate them.
//This processing uses reflection, but afterwards AutoInflate can run quickly.
func newReaderValidator(exampleReader PropertyReader, exampleObj interface{}, illegalTypes map[PropertyType]bool, chest *ComponentChest) (*readerValidator, error) {
	//TODO: there's got to be a way to not need both exampleReader and exampleObj, but only one.

	if illegalTypes == nil {
		illegalTypes = make(map[PropertyType]bool)
	}

	autoEnumVarFields := make(map[string]*enum.Enum)
	autoEnumConstFields := make(map[string]*enum.Enum)
	autoSizedStackFields := make(map[string]*autoStackConfig)
	autoGrowableStackFields := make(map[string]*autoStackConfig)

	for propName, propType := range exampleReader.Props() {
		switch propType {
		case TypeSizedStack:
			sizedStack, err := exampleReader.SizedStackProp(propName)
			if err != nil {
				return nil, errors.New("Couldn't fetch sized stack prop: " + propName)
			}
			if sizedStack != nil {
				//This stack prop is already non-nil, so we don't need to do
				//any processing to tell how to inflate it.
				continue
			}
			if tag := structTagForField(exampleObj, propName, stackStructTag); tag != "" {

				deck, size, err := unpackStackStructTag(tag, chest)

				if err != nil {
					return nil, errors.New(propName + " was a nil SizedStack and its struct tag was not valid: " + err.Error())
				}

				autoSizedStackFields[propName] = &autoStackConfig{
					deck,
					size,
				}
			}
		case TypeGrowableStack:
			growableStack, err := exampleReader.GrowableStackProp(propName)
			if err != nil {
				return nil, errors.New("Couldn't fetch sized growable prop: " + propName)
			}
			if growableStack != nil {
				//This stack prop is already non-nil, so we don't need to do
				//any processing to tell how to inflate it.
				continue
			}
			if tag := structTagForField(exampleObj, propName, stackStructTag); tag != "" {

				deck, size, err := unpackStackStructTag(tag, chest)

				if err != nil {
					return nil, errors.New(propName + " was a nil growable stack and its struct tag was not valid: " + err.Error())
				}

				autoGrowableStackFields[propName] = &autoStackConfig{
					deck,
					size,
				}
			}
		case TypeEnumConst:
			enumConst, err := exampleReader.EnumConstProp(propName)
			if err != nil {
				return nil, errors.New("Couldn't fetch enum const prop: " + propName)
			}
			if enumConst != nil {
				//This enum prop is already non-nil, so we don't need to do
				//any processing to tell how to inflate it.
				continue
			}
			if enumName := structTagForField(exampleObj, propName, enumStructTag); enumName != "" {
				theEnum := chest.Enums().Enum(enumName)
				if theEnum == nil {
					return nil, errors.New(propName + " was a nil enum.Const and the struct tag named " + enumName + " was not a valid enum.")
				}
				//Found one!
				autoEnumConstFields[propName] = theEnum
			}
		case TypeEnumVar:
			enumVar, err := exampleReader.EnumVarProp(propName)
			if err != nil {
				return nil, errors.New("Couldn't fetch enum var prop: " + propName)
			}
			if enumVar != nil {
				//This enum prop is already non-nil, so we don't need to do
				//any processing to tell how to inflate it.
				continue
			}
			if enumName := structTagForField(exampleObj, propName, enumStructTag); enumName != "" {
				theEnum := chest.Enums().Enum(enumName)
				if theEnum == nil {
					return nil, errors.New(propName + " was a nil enum.Var and the struct tag named " + enumName + " was not a valid enum.")
				}
				//Found one!
				autoEnumVarFields[propName] = theEnum
			}
		}
	}

	result := &readerValidator{
		autoEnumVarFields,
		autoEnumConstFields,
		autoGrowableStackFields,
		autoSizedStackFields,
		illegalTypes,
	}

	if err := result.VerifyNoIllegalProps(exampleReader); err != nil {
		return nil, errors.New("Example had illegal prop types: " + err.Error())
	}

	return result, nil
}

//AutoInflate will go through and inflate fields that are nil that it knows
//how to inflate due to comments in structs detected in the constructor for
//this validator.
func (r *readerValidator) AutoInflate(readSetter PropertyReadSetter, st State) error {

	var statePtr *state

	if st != nil {
		var ok bool
		statePtr, ok = st.(*state)
		if !ok {
			return errors.New("The provided non-nil State could not be conveted to a state ptr")
		}
	}

	for propName, config := range r.autoGrowableStackFields {

		if statePtr == nil {
			return errors.New("Provided state was nil but there are growable stacks to expand")
		}

		growableStack, err := readSetter.GrowableStackProp(propName)
		if growableStack != nil {
			//Guess it was already set!
			continue
		}
		if err != nil {
			return errors.New(propName + " had error fetching growable stack: " + err.Error())
		}
		if config == nil {
			return errors.New("The config for " + propName + " was unexpectedly nil")
		}
		if config.deck == nil {
			return errors.New("The deck for " + propName + " was unexpectedly nil")
		}

		stack := NewGrowableStack(config.deck, config.size)
		stack.statePtr = statePtr

		if err := readSetter.SetGrowableStackProp(propName, stack); err != nil {
			return errors.New("Couldn't set " + propName + " to growable stack: " + err.Error())
		}
	}

	for propName, config := range r.autoSizedStackFields {

		if statePtr == nil {
			return errors.New("Provided state was nil but there are sized stacks to expand")
		}

		sizedStack, err := readSetter.SizedStackProp(propName)
		if sizedStack != nil {
			//Guess it was already set!
			continue
		}
		if err != nil {
			return errors.New(propName + " had error fetching sized stack: " + err.Error())
		}
		if config == nil {
			return errors.New("The config for " + propName + " was unexpectedly nil")
		}
		if config.deck == nil {
			return errors.New("The deck for " + propName + " was unexpectedly nil")
		}

		stack := NewSizedStack(config.deck, config.size)
		stack.statePtr = statePtr

		if err := readSetter.SetSizedStackProp(propName, stack); err != nil {
			return errors.New("Couldn't set " + propName + " to sized stack: " + err.Error())
		}
	}

	for propName, enum := range r.autoEnumVarFields {
		enumVar, err := readSetter.EnumVarProp(propName)
		if enumVar != nil {
			//Guess it was already set!
			continue
		}
		if err != nil {
			return errors.New(propName + " had error fetching EnumVar: " + err.Error())
		}
		if enum == nil {
			return errors.New("The enum for " + propName + " was unexpectedly nil")
		}
		if err := readSetter.SetEnumVarProp(propName, enum.NewVar()); err != nil {
			return errors.New("Couldn't set " + propName + " to NewVar: " + err.Error())
		}
	}

	for propName, enum := range r.autoEnumConstFields {
		enumConst, err := readSetter.EnumConstProp(propName)
		if enumConst != nil {
			//Guess it was already set!
			continue
		}
		if err != nil {
			return errors.New(propName + " had error fetching EnumConst: " + err.Error())
		}
		if enum == nil {
			return errors.New("The enum for " + propName + " was unexpectedly nil")
		}
		if err := readSetter.SetEnumConstProp(propName, enum.NewDefaultConst()); err != nil {
			return errors.New("Couldn't set " + propName + " to NewDefaultConst: " + err.Error())
		}
	}

	for propName, propType := range readSetter.Props() {
		switch propType {
		case TypeTimer:
			if statePtr == nil {
				return errors.New("Provided state was nil but there are timers to expand")
			}
			timer := NewTimer()
			timer.statePtr = statePtr
			if err := readSetter.SetTimerProp(propName, timer); err != nil {
				return errors.New("Couldn't set " + propName + " to a new timer: " + err.Error())
			}
		}
	}

	//TODO: process Stack, Timer fields (convert to state pointer if non-nil)
	return nil
}

func (r *readerValidator) VerifyNoIllegalProps(reader PropertyReader) error {
	for propName, propType := range reader.Props() {
		if propType == TypeIllegal {
			return errors.New(propName + " was TypeIllegal, which is always illegal")
		}
		if _, illegal := r.illegalTypes[propType]; illegal {
			return errors.New(propName + " was the type " + propType.String() + ", which is illegal in this context")
		}
	}
	return nil
}

//Valid will return an error if the reader is not valid according to the
//configuration of this validator.
func (r *readerValidator) Valid(reader PropertyReader) error {
	if err := r.VerifyNoIllegalProps(reader); err != nil {
		return err
	}
	for propName, propType := range reader.Props() {
		//TODO: verifyReader should be gotten rid of in favor of this
		switch propType {
		case TypeGrowableStack:
			val, err := reader.GrowableStackProp(propName)
			if val == nil {
				return errors.New("GrowableStack Prop " + propName + " was nil")
			}
			if err != nil {
				return errors.New("GrowableStack prop " + propName + " had unexpected error: " + err.Error())
			}
			if val.state() == nil {
				return errors.New("GrowableStack prop " + propName + " didn't have its state set")
			}
		case TypeSizedStack:
			val, err := reader.SizedStackProp(propName)
			if val == nil {
				return errors.New("SizedStackProp " + propName + " was nil")
			}
			if err != nil {
				return errors.New("SizedStack prop " + propName + " had unexpected error: " + err.Error())
			}

			if val.state() == nil {
				return errors.New("SizedStack prop " + propName + " didn't have its state set")
			}
		case TypeTimer:
			val, err := reader.TimerProp(propName)
			if val == nil {
				return errors.New("TimerProp " + propName + " was nil")
			}
			if err != nil {
				return errors.New("TimerProp " + propName + " had unexpected error: " + err.Error())
			}
			if val.statePtr == nil {
				return errors.New("TimerProp " + propName + " didn't have its statePtr set")
			}
		case TypeEnumVar:
			val, err := reader.EnumVarProp(propName)
			if val == nil {
				return errors.New("EnumVarProp " + propName + " was nil")
			}
			if err != nil {
				return errors.New("EnumVarProp " + propName + " had unexpected error: " + err.Error())
			}
		case TypeEnumConst:
			val, err := reader.EnumConstProp(propName)
			if val == nil {
				return errors.New("EnumConstProp " + propName + " was nil")
			}
			if err != nil {
				return errors.New("EnumConstProp " + propName + " had unexpected error: " + err.Error())
			}
		}
	}
	return nil
}

func unpackStackStructTag(tag string, chest *ComponentChest) (*Deck, int, error) {
	pieces := strings.Split(tag, ",")

	if len(pieces) > 2 {
		return nil, 0, errors.New("There were more fields in the struct tag than expected")
	}

	deckName := pieces[0]

	deck := chest.Deck(deckName)

	if deck == nil {
		return nil, 0, errors.New("The deck name " + deckName + " was not a valid deck")
	}

	size := 0

	if len(pieces) > 1 {
		var err error
		size, err = strconv.Atoi(pieces[1])
		if err != nil {
			return nil, 0, errors.New("The size in the struct tag was not a valid int: " + err.Error())
		}
	}

	return deck, size, nil

}

//structTagForField will use reflection to fetch the named field from the
//object and return the value of its `enum` field. Works even if fieldName is
//in an embedded struct.
func structTagForField(obj interface{}, fieldName string, structTag string) string {

	v := reflect.Indirect(reflect.ValueOf(obj))

	t := reflect.TypeOf(v.Interface())

	field, ok := t.FieldByNameFunc(func(str string) bool {
		return str == fieldName
	})

	if !ok {
		return ""
	}

	return field.Tag.Get(structTag)

}
