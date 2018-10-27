package boardgame

import (
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
	"reflect"
	"strconv"
	"strings"
)

const enumStructTag = "enum"
const stackStructTag = "stack"
const concatenateStructTag = "concatenate"
const overlapStructTag = "overlap"
const fixedStackStructTag = "sizedstack"
const sanitizationStructTag = "sanitize"
const boardStructTag = "board"

type autoStackConfig struct {
	deck      *Deck
	size      int
	fixedSize bool
	//If more than 0, then a board config.
	boardSize int
}

type autoMergedStackConfig struct {
	props   []string
	overlap bool
}

type readerValidator struct {
	autoEnumFields        map[string]enum.Enum
	autoMutableEnumFields map[string]enum.Enum
	autoStackFields       map[string]*autoStackConfig
	autoMergedStackFields map[string]*autoMergedStackConfig
	sanitizationPolicy    map[string]map[int]Policy
	illegalTypes          map[PropertyType]bool
}

//newReaderValidator returns a new readerValidator configured to disallow the
//given types. It will also do an expensive processing for any nil pointer-
//properties to see if they have struct tags that tell us how to inflate them.
//This processing uses reflection, but afterwards AutoInflate can run quickly.
//If exampleObj also implements ReadSetter, the resulting ReadSetter is used
//to do a PropMutable check for those properties--and if not provided, we
//assume all of the interface props are not mutable.
func newReaderValidator(exampleObj Reader, illegalTypes map[PropertyType]bool, chest *ComponentChest) (*readerValidator, error) {

	if chest == nil {
		return nil, errors.New("Passed nil chest")
	}

	if illegalTypes == nil {
		illegalTypes = make(map[PropertyType]bool)
	}

	exampleReader := exampleObj.Reader()

	if exampleReader == nil {
		return nil, errors.New("Example object's Reader() returned nil")
	}

	var exampleReadSetter PropertyReadSetter
	if readSetter, ok := exampleObj.(ReadSetter); ok {
		exampleReadSetter = readSetter.ReadSetter()
	}

	autoEnumFields := make(map[string]enum.Enum)
	autoMutableEnumFields := make(map[string]enum.Enum)
	autoStackFields := make(map[string]*autoStackConfig)
	autoMergedStackFields := make(map[string]*autoMergedStackConfig)
	sanitizationPolicy := make(map[string]map[int]Policy)

	defaultGroup := "all"
	//If the object apeparst to be a playerState, then the default group is "other", not "all".
	if _, ok := exampleObj.(PlayerState); ok {
		defaultGroup = "other"
	}

	for propName, propType := range exampleReader.Props() {

		sanitizationPolicy[propName] = policyFromStructTag(structTagForField(exampleObj, propName, sanitizationStructTag), defaultGroup)

		switch propType {
		case TypeStack, TypeBoard:

			if propType == TypeStack {
				stack, err := exampleReader.ImmutableStackProp(propName)
				if err != nil {
					return nil, errors.New("Couldn't fetch stack prop: " + propName)
				}
				if stack != nil {
					//This stack prop is already non-nil, so we don't need to do
					//any processing to tell how to inflate it.
					continue
				}
			} else {
				board, err := exampleReader.ImmutableBoardProp(propName)
				if err != nil {
					return nil, errors.New("Couldn't fetch board prop: " + propName)
				}
				if board != nil {
					//This board prop is already non-nil, so we don't need to do
					//any processing to tell how to inflate it.
					continue
				}
			}

			var tag string

			structTags := structTagsForField(exampleObj, propName, []string{
				stackStructTag,
				fixedStackStructTag,
				concatenateStructTag,
				overlapStructTag,
				boardStructTag,
			})

			if exampleReadSetter != nil && exampleReadSetter.PropMutable(propName) {

				if structTags[concatenateStructTag] != "" {
					return nil, errors.New(propName + " included a concatenate struct tag on a mutable stack property")
				}

				if structTags[overlapStructTag] != "" {
					return nil, errors.New(propName + " included a overlap struct tag on a mutable stack property")
				}

				boardSize, err := unpackBoardStructTag(structTags[boardStructTag], chest)

				if err != nil {
					return nil, errors.New("Invalid board struct tag: " + err.Error())
				}

				isFixed := false

				tag = structTags[stackStructTag]

				if tag == "" {
					tag = structTags[fixedStackStructTag]
					if tag != "" {
						isFixed = true
					}
				}

				if isFixed && boardSize > 0 {
					return nil, errors.New("Provided a board tag with a sizedstack, which is invalid.")
				}

				if tag != "" {

					deck, size, err := unpackStackStructTag(tag, chest)

					if err != nil {
						return nil, errors.New(propName + " was a nil SizedStack and its struct tag was not valid: " + err.Error())
					}

					if isFixed && size == 0 {
						//Size for sizedstacks defaults to 1 (which can be grown
						//easily to any other size).
						size = 1
					}

					autoStackFields[propName] = &autoStackConfig{
						deck,
						size,
						isFixed,
						boardSize,
					}
				} else {
					if boardSize > 0 {
						return nil, errors.New("board stuct tag provided, without a corresponding stack struct tag.")
					}
				}
			}

			//If the read setter isn't provided we assume that the stack
			//properties are all immutable.
			if exampleReadSetter == nil || (exampleReadSetter != nil && !exampleReadSetter.PropMutable(propName)) {

				if structTags[stackStructTag] != "" {
					return nil, errors.New(propName + " included a stack struct tag on an immutable stack property")
				}

				if structTags[fixedStackStructTag] != "" {
					return nil, errors.New(propName + " included a sizedstack struct tag on an immutable stack property")
				}

				overlap := false

				tag = structTags[concatenateStructTag]

				if tag == "" {
					tag = structTags[overlapStructTag]
					if tag != "" {
						overlap = true
					}
				}

				if tag != "" {
					props, err := unpackMergedStackStructTag(tag, exampleReader)
					if err != nil {
						return nil, errors.New(propName + " was a nil stack and its struct tag was not valid: " + err.Error())
					}

					//unpackMergedStackStructTag already checked that the
					//props pointed to are valid on this reader.

					autoMergedStackFields[propName] = &autoMergedStackConfig{
						props,
						overlap,
					}
				}
			}

		case TypeEnum:
			enumConst, err := exampleReader.ImmutableEnumProp(propName)
			if err != nil {
				return nil, errors.New("Couldn't fetch enum  prop: " + propName)
			}
			if enumConst != nil {
				//This enum prop is already non-nil, so we don't need to do
				//any processing to tell how to inflate it.
				continue
			}
			if enumName := structTagForField(exampleObj, propName, enumStructTag); enumName != "" {
				theEnum := chest.Enums().Enum(enumName)
				if theEnum == nil {
					return nil, errors.New(propName + " was a nil enum.Val and the struct tag named " + enumName + " was not a valid enum.")
				}
				//Found one!
				if exampleReadSetter != nil {
					if exampleReadSetter.PropMutable(propName) {
						autoMutableEnumFields[propName] = theEnum
					} else {
						autoEnumFields[propName] = theEnum
					}

				} else {
					//Just assume they're immutable
					autoEnumFields[propName] = theEnum
				}
			}
		}

	}

	result := &readerValidator{
		autoEnumFields,
		autoMutableEnumFields,
		autoStackFields,
		autoMergedStackFields,
		sanitizationPolicy,
		illegalTypes,
	}

	if err := result.verifyNoIllegalProps(exampleReader); err != nil {
		return nil, errors.New("Example had illegal prop types: " + err.Error())
	}

	return result, nil
}

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

func policyFromStructTag(tag string, defaultGroup string) map[int]Policy {
	if tag == "" {
		tag = "visible"
	}

	errorMap := map[int]Policy{
		GroupAll: PolicyInvalid,
	}

	result := make(map[int]Policy)

	pieces := strings.Split(tag, ",")
	for _, piece := range pieces {
		splitPiece := strings.Split(piece, ":")
		var groupString string
		var policyString string
		if len(splitPiece) > 2 {
			return errorMap
		}
		if len(splitPiece) == 1 {
			groupString = defaultGroup
			policyString = splitPiece[0]
		} else {
			groupString = splitPiece[0]
			policyString = splitPiece[1]
		}

		group := groupFromString(groupString)
		policy := policyFromString(policyString)

		result[group] = policy

	}

	return result

}

//PropertySanitizationPolicy returns the policy (map[GroupIndex]Policy) based
//on the struct tags from the example struct given to NewStructInflater.
func (r *readerValidator) PropertySanitizationPolicy(propName string) map[int]Policy {
	return r.sanitizationPolicy[propName]
}

//Inflate will go through and inflate fields that are nil that it knows
//how to inflate due to comments in structs detected in the constructor for
//this validator.
func (r *readerValidator) Inflate(obj ReadSetConfigurer, st ImmutableState) error {

	readSetConfigurer := obj.ReadSetConfigurer()

	if readSetConfigurer == nil {
		return errors.New("Object's ReadSetConfigurer returned nil")
	}

	for propName, config := range r.autoStackFields {

		if config.boardSize > 0 {

			board, err := readSetConfigurer.BoardProp(propName)
			if board != nil {
				//Guess it was already set!
				continue
			}
			if err != nil {
				return errors.New(propName + " had error fetching board: " + err.Error())
			}
			if config == nil {
				return errors.New("The config for " + propName + " was unexpectedly nil")
			}
			if config.deck == nil {
				return errors.New("The deck for " + propName + " was unexpectedly nil")
			}

			board = config.deck.NewBoard(config.boardSize, config.size)

			if err := readSetConfigurer.ConfigureBoardProp(propName, board); err != nil {
				return errors.New("Couldn't set " + propName + " to board: " + err.Error())
			}

		} else {

			stack, err := readSetConfigurer.StackProp(propName)
			if stack != nil {
				//Guess it was already set!
				continue
			}
			if err != nil {
				return errors.New(propName + " had error fetching stack: " + err.Error())
			}
			if config == nil {
				return errors.New("The config for " + propName + " was unexpectedly nil")
			}
			if config.deck == nil {
				return errors.New("The deck for " + propName + " was unexpectedly nil")
			}

			if config.fixedSize {
				stack = config.deck.NewSizedStack(config.size)
			} else {
				stack = config.deck.NewStack(config.size)
			}

			if err := readSetConfigurer.ConfigureStackProp(propName, stack); err != nil {
				return errors.New("Couldn't set " + propName + " to stack: " + err.Error())
			}
		}
	}

	for propName, config := range r.autoMergedStackFields {
		stack, err := readSetConfigurer.ImmutableStackProp(propName)
		if stack != nil {
			//Guess it was already set!
			continue
		}
		if err != nil {
			return errors.New(propName + " had error fetching stack: " + err.Error())
		}
		if config == nil {
			return errors.New("The config for " + propName + " was unexpectedly nil")
		}
		stacks := make([]ImmutableStack, len(config.props))
		for i, prop := range config.props {
			stack, err := readSetConfigurer.ImmutableStackProp(prop)
			if err != nil {
				return errors.New(propName + " Couldn't fetch the " + strconv.Itoa(i) + " stack to merge: " + prop + ": " + err.Error())
			}
			if stack == nil {
				return errors.New(propName + " had a nil " + strconv.Itoa(i) + " stack")
			}
			stacks[i] = stack
		}

		if config.overlap {
			stack = NewOverlappedStack(stacks...)
		} else {
			stack = NewConcatenatedStack(stacks...)
		}

		if err := readSetConfigurer.ConfigureImmutableStackProp(propName, stack); err != nil {
			return errors.New("Couldn't set " + propName + " to stack: " + err.Error())
		}

	}

	for propName, enum := range r.autoEnumFields {
		enumConst, err := readSetConfigurer.ImmutableEnumProp(propName)
		if enumConst != nil {
			//Guess it was already set!
			continue
		}
		if err != nil {
			return errors.New(propName + " had error fetching Enum: " + err.Error())
		}
		if enum == nil {
			return errors.New("The enum for " + propName + " was unexpectedly nil")
		}
		if err := readSetConfigurer.ConfigureImmutableEnumProp(propName, enum.NewDefaultVal()); err != nil {
			return errors.New("Couldn't set " + propName + " to NewDefaultVal: " + err.Error())
		}
	}

	for propName, enum := range r.autoMutableEnumFields {
		enumConst, err := readSetConfigurer.EnumProp(propName)
		if enumConst != nil {
			//Guess it was already set!
			continue
		}
		if err != nil {
			return errors.New(propName + " had error fetching Enum: " + err.Error())
		}
		if enum == nil {
			return errors.New("The enum for " + propName + " was unexpectedly nil")
		}
		if err := readSetConfigurer.ConfigureEnumProp(propName, enum.NewVal()); err != nil {
			return errors.New("Couldn't set " + propName + " to NewDefaultVal: " + err.Error())
		}
	}

	for propName, propType := range readSetConfigurer.Props() {
		switch propType {
		case TypeTimer:
			timer := NewTimer()
			if readSetConfigurer.PropMutable(propName) {
				if err := readSetConfigurer.ConfigureTimerProp(propName, timer); err != nil {
					return errors.New("Couldn't set " + propName + " to a new timer: " + err.Error())
				}
			} else {
				if err := readSetConfigurer.ConfigureImmutableTimerProp(propName, timer); err != nil {
					return errors.New("Couldn't set " + propName + " to a new timer: " + err.Error())
				}
			}
		}
	}

	if st != nil {
		if err := setReaderStatePtr(readSetConfigurer, st); err != nil {
			return errors.New("Couldn't set state ptrs: " + err.Error())
		}
	}

	//TODO: process Stack, Timer fields (convert to state pointer if non-nil)
	return nil
}

func (r *readerValidator) verifyNoIllegalProps(reader PropertyReader) error {

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
func (r *readerValidator) Valid(obj Reader) error {

	reader := obj.Reader()

	if reader == nil {
		return errors.New("Object's Reader returned nil")
	}

	if err := r.verifyNoIllegalProps(reader); err != nil {
		return err
	}
	for propName, propType := range reader.Props() {

		policyMap := r.sanitizationPolicy[propName]

		if policyMap == nil {
			return errors.New(propName + " had no sanitization policy")
		}

		for group, policy := range policyMap {
			if policy == PolicyInvalid {
				return errors.New(propName + " had invalid policy for group " + strconv.Itoa(group))
			}
		}

		//TODO: verifyReader should be gotten rid of in favor of this
		switch propType {
		case TypeStack:
			val, err := reader.ImmutableStackProp(propName)
			if val == nil {
				return errors.New("Stack Prop " + propName + " was nil")
			}
			if err != nil {
				return errors.New("Stack prop " + propName + " had unexpected error: " + err.Error())
			}
			if val.state() == nil {
				return errors.New("Stack prop " + propName + " didn't have its state set")
			}
		case TypeBoard:
			val, err := reader.ImmutableBoardProp(propName)
			if val == nil {
				return errors.New("Board Prop " + propName + " was nil")
			}
			if err != nil {
				return errors.New("Board prop " + propName + " had unexpected error: " + err.Error())
			}
			if val.state() == nil {
				return errors.New("Stack prop " + propName + " didn't have its state set")
			}
		case TypeTimer:
			val, err := reader.ImmutableTimerProp(propName)
			if val == nil {
				return errors.New("TimerProp " + propName + " was nil")
			}
			if err != nil {
				return errors.New("TimerProp " + propName + " had unexpected error: " + err.Error())
			}
			if val.state() == nil {
				return errors.New("TimerProp " + propName + " didn't have its statePtr set")
			}
		case TypeEnum:
			val, err := reader.ImmutableEnumProp(propName)
			if val == nil {
				return errors.New("EnumProp " + propName + " was nil")
			}
			if err != nil {
				return errors.New("EnumProp " + propName + " had unexpected error: " + err.Error())
			}
		}

	}
	return nil
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

func unpackBoardStructTag(tag string, chest *ComponentChest) (length int, err error) {
	if tag == "" {
		return 0, nil
	}

	val, err := intEffectiveValue(tag, chest)

	if err != nil {
		return 0, err
	}

	if val < 0 {
		return 0, nil
	}

	return val, nil
}

func unpackMergedStackStructTag(tag string, reader PropertyReader) (stackNames []string, err error) {
	pieces := strings.Split(tag, ",")

	if len(pieces) < 2 {
		return nil, errors.New("There were fewer properties than we expected.")
	}

	result := make([]string, len(pieces))

	for i, piece := range pieces {
		prop := strings.TrimSpace(piece)

		if _, err := reader.ImmutableStackProp(prop); err != nil {
			return nil, errors.New(prop + " does not denote a valid stack property on that object")
		}
		result[i] = prop
	}

	return result, nil

}

func unpackStackStructTag(tag string, chest *ComponentChest) (*Deck, int, error) {
	pieces := strings.Split(tag, ",")

	if len(pieces) > 2 {
		return nil, 0, errors.New("There were more fields in the struct tag than expected")
	}

	deckName := strings.TrimSpace(pieces[0])

	deck := chest.Deck(deckName)

	if deck == nil {
		return nil, 0, errors.New("The deck name " + deckName + " was not a valid deck")
	}

	size := 0

	if len(pieces) > 1 {
		var err error
		size, err = intEffectiveValue(pieces[1], chest)
		if err != nil {
			return nil, 0, errors.New("The size in the struct tag was not a valid int: " + err.Error())
		}
	}

	return deck, size, nil

}

//intEffectiveValue either returns the integer encoded by the string, or if
//the string encodes the name of a constant in chest that is an int, that.
func intEffectiveValue(str string, chest *ComponentChest) (int, error) {
	str = strings.TrimSpace(str)

	val := chest.Constant(str)

	if val != nil {

		i, ok := val.(int)

		if !ok {
			return 0, errors.New(str + "was a cosntant, but not of type int as required")
		}
		return i, nil
	}

	intVal, err := strconv.Atoi(str)

	if err != nil {
		return 0, errors.New(str + " is not convertable to int: " + err.Error())
	}

	return intVal, nil
}

//structTagForField will use reflection to fetch the named field from the
//object and return the value of its `enum` field. Works even if fieldName is
//in an embedded struct.
func structTagForField(obj interface{}, fieldName string, structTag string) string {
	result := structTagsForField(obj, fieldName, []string{structTag})
	return result[structTag]
}

func structTagsForField(obj interface{}, fieldName string, structTags []string) map[string]string {

	result := make(map[string]string, len(structTags))

	v := reflect.Indirect(reflect.ValueOf(obj))

	t := reflect.TypeOf(v.Interface())

	field, ok := t.FieldByNameFunc(func(str string) bool {
		return str == fieldName
	})

	if !ok {
		return result
	}

	theTag := field.Tag

	for _, structTag := range structTags {
		result[structTag] = theTag.Get(structTag)
	}

	return result
}
