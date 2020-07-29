package boardgame

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
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

/*

StructInflater is an object that inspects structs for tags with instructions
using reflection. Later, it can use what it learned to auto-inflate nil
interface properties (e.g. Timers, Stacks, and Enums) on those structs using
the instructions encoded in the tag. After the creation of a StructInflater,
reflection is no longer necessary, so operations are fast. In addition,
StructInflaters can inspect a struct for whether it's valid (has no nil
properties, has no illegal property types in this context), as well as
reporting what the configuration of sanitization policies was based on the
struct tags. Get a new one from NewStructInflater. For more about the precise
configuration of struct tags that StructInlater understands, see the
documenation for the methods on StructInflater that make use of it, including
Inflate() and PropertySanitizationPolicy().

*/
type StructInflater struct {
	autoEnumFields        map[string]enum.Enum
	autoMutableEnumFields map[string]enum.Enum
	autoStackFields       map[string]*autoStackConfig
	autoMergedStackFields map[string]*autoMergedStackConfig
	sanitizationPolicy    map[string]map[string]Policy
	illegalTypes          map[PropertyType]bool
}

//NewStructInflater returns a new StructInflater configured for use based on
//the given object. NewStructInflater does all of the reflection necessary to
//do auto-inflation later, meaning that although this is a bit slow, later
//calls on the StructInflater don't need to use reflection again. Chest must
//be non-nil, so that we can validate that the tag-based configuration denotes
//valid properties. If illegalTypes is non-nil, then this constructor, and
//calls to this StructInflater's Valid() method, will error if the struct has
//any of those fields defined.
//
//NewStructInflater checks for any number of illegal or nonsensical
//conditions, including checking Valid() on the return value, as well as
//verifying that if the exampleObj also has a ReadSetter that things like
//MergedStacksa are not accesible from mutable reader accessors, retuning an
//error and a nil StructInflater if anything is invalid.
//
//You typically do not use this directly; the base library will automatically
//create ones for you for its own use to infalte your gameStates,
//playerStates, dynamicComponentValueStates, and Moves, and which you can get
//access to via manager.Internals().StructInflater().
func NewStructInflater(exampleObj Reader, illegalTypes map[PropertyType]bool, chest *ComponentChest) (*StructInflater, error) {

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
	sanitizationPolicy := make(map[string]map[string]Policy)

	//NewGameManger already verified that we could rely on the groupEnum to have
	//GroupAll, groupOther.
	defaultGroup := SanitizationDefaultGroup
	//If the object apeparst to be a playerState, then the default group is "other", not "all".
	if subState, ok := exampleObj.(SubState); ok {
		if subState.StatePropertyRef().Group == StateGroupPlayer {
			defaultGroup = SanitizationDefaultPlayerGroup
		}
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
					return nil, errors.New("[rovided a board tag with a sizedstack, which is invalid")
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
						return nil, errors.New("board stuct tag provided, without a corresponding stack struct tag")
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

	result := &StructInflater{
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

func policyFromStructTag(tag string, defaultGroup string) map[string]Policy {
	if tag == "" {
		tag = "visible"
	}

	errorMap := make(map[string]Policy)
	errorMap[SanitizationDefaultGroup] = PolicyInvalid

	result := make(map[string]Policy)

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

		policy := policyFromString(policyString)

		result[groupString] = policy

	}

	return result

}

//sanitizationPolicyGroupNames returns a map of all sanitization group names
//used in this inflater.
func (s *StructInflater) sanitizationPolicyGroupNames() map[string]bool {
	result := make(map[string]bool)
	for _, policyMap := range s.sanitizationPolicy {
		for key := range policyMap {
			result[key] = true
		}
	}
	return result
}

/*

PropertySanitizationPolicy returns the policy (map[GroupIndex]Policy) based on
the struct tags from the example struct given originally to NewStructInflater.
It does not use reflection, relying on reflection at the time of creation of the
StructInflater. In particular, it interprets policy tags in the following way:

It looks for struct tag configuration with the `sanitize` keyword.

Keywords are interpreted by splitting at "," into a series of configurations.
For each configuration, a group name ("all", "other", "self", or one of the
other valid values described below) is followed by a ":" and then a policy, one
of "visible", "order", "len", "nonempty", and "hidden".

If the group name is omitted for a config item, it is assumed to be "all" (for
non-playerState structs), or "other" for playerState structs. We decide if a
struct is a playerState if it can be cast to a boardgame.PlayerState. The
constants for "all" and "other" are available as SanitizationDefaultGroup and
SanitizationDefaultPlayerGroup.

Any string key that is a member of enum returned from delegate.GroupEnum() may
be used, not just 'all', 'other', or 'self'. In addition, any group name that is
handled by your delegate's ComputedPlayerGroupMembership() may be used.

Group name 'all' will always be passed to delegate.SanitizationPolicy. Group
name 'self' will be passed for PlayerStates where that player is also the player
the state is being sanitized for. Group name 'other' is the opposite behavior of
'self'.

This means all of the following are valid:

    type myPlayerState struct {
        base.SubState
        playerIndex boardgame.PlayerIndex
        VisibleHand boardgame.Stack //Equivalent to `sanitize:"all:visible"`
        HiddenHand boardgame.Stack `sanitize:"len"` // Equivalent to `sanitize:"other:len"`, since this is a player state.
        OtherStack boardgame.Stack `sanitize:"nonempty,self:len"` //Eqiuvalent to `sanitize:"other:nonempty,self:len"`
    }

Missing policy configuration is interpreted for that property as though it said
`sanitize:"all:visible"`

*/
func (s *StructInflater) PropertySanitizationPolicy(propName string) map[string]Policy {
	return s.sanitizationPolicy[propName]
}

/*

Inflate uses tag-based configuration it detected when this StructInflater
was created in order to fill in instantiated values for nil Interface
properties (e.g. Stacks, Timers, and Enums). It skips any properties that
didn't have configuration provided via struct tags, or that were already
non-nil. It assumes that the object is the same underlying shape as the one
that was passed when this StructInflater was created.

The struct tag based configuration is interpreted as follows:

For Timers, no struct-based configuration is necessary, any property of type
Timer is simply replaced with a timer object (since timers don't take any
configuration to set-up).

For default (growable) Stacks, the struct tag is `stack`. The contents of the
tag is the name of the deck this stack is affiliated with. If the given deck
name does not exist in the ComponentChest in use, the StructInflater would
have failed to be created. You can also optionally provide a max-size by
appending ",SIZE" into your struct tag.

For sized Stacks, you provide the same types of configuration as for stacks,
but with the struct tag of "sizedstack" instead. Note that whereas a max-size
of a growable stack is the default, a size of 0 for a sizedstack is
effectively useless, so generally for sized stacks you provide a size.

For boards, you provide a stack tag, and also a "board" tag which denotes how
many slots the board should have.

For merged stacks, you provide a tag of either "concatenate" or "overlap" and
then a comma-separated list of the names of stacks on this object to combine
in that way. If any of those property names are not defined on this object,
the StructInflater would have failed to have been crated.

For enums, you provide the name of the enum this val is associated with by
providing the struct-tag "enum". If that named enum does not exist in the
ComponentChest in use, the StructInflater would have failed to be created.

For every integer-based property described above, you can replace the int in
the struct value with the name of a constant value that is defined on this
ComponentChest. We'll fetch that constant and use that for the int (erroring
if it's not an int).

*/
func (s *StructInflater) Inflate(obj ReadSetConfigurer, st ImmutableState) error {

	readSetConfigurer := obj.ReadSetConfigurer()

	if readSetConfigurer == nil {
		return errors.New("Object's ReadSetConfigurer returned nil")
	}

	for propName, config := range s.autoStackFields {

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

	for propName, config := range s.autoMergedStackFields {
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

	for propName, enum := range s.autoEnumFields {
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

	for propName, enum := range s.autoMutableEnumFields {
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

func (s *StructInflater) verifyNoIllegalProps(reader PropertyReader) error {

	for propName, propType := range reader.Props() {
		if propType == TypeIllegal {
			return errors.New(propName + " was TypeIllegal, which is always illegal")
		}
		if _, illegal := s.illegalTypes[propType]; illegal {
			return errors.New(propName + " was the type " + propType.String() + ", which is illegal in this context")
		}
	}
	return nil
}

//Valid will return an error if the object has any properties defined of a
//type that is part of the illegalTypes passed to the StructInflater
//constructor, or if any Interface property (e.g. Stack, Timer, Enum) is
//currently nil. Valid can help ensure that a given object has been fully
//inflated.
func (s *StructInflater) Valid(obj Reader) error {

	reader := obj.Reader()

	if reader == nil {
		return errors.New("Object's Reader returned nil")
	}

	if err := s.verifyNoIllegalProps(reader); err != nil {
		return err
	}
	for propName, propType := range reader.Props() {

		policyMap := s.sanitizationPolicy[propName]

		if policyMap == nil {
			return errors.New(propName + " had no sanitization policy")
		}

		for group, policy := range policyMap {
			if policy == PolicyInvalid {
				return errors.New(propName + " had invalid policy for group " + group)
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
		return nil, errors.New("there were fewer properties than we expected")
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
