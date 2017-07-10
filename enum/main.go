/*

In a number of cases you have a property that can only have a handful of
possible values. You want to verify that the value is always one of those
legal values, and make sure that you can compare it to a known constant so you
can make sure you don't have a typo at compile time instead of run time. It's
also nice to have them have an order in many cases, and to be serialized with
the string value so it's easier to read.

Enums are useful for this case. An EnumSet contains multiple enums, and you
can create an EnumValue which can be used as a property on a PropertyReader
object.

The idiomatic way to create an enum is the following:
	const (
		//The first enum can start at 0
		ColorRed = iota
		ColorBlue
		ColorGreen
	)

	const (
		//The second enum should start at 1 plus the last item in the
		//previous, because all int vals in an EnumSet must be unique.
		CardSpade = ColorGreen + 1 + iota
		CardHeart
		CardDiamond
		CardClub
	)

	var Enums = enum.NewSet()

	var ColorEnum = Enums.MustAdd("Color", map[int]string{
		ColorRed: "Red",
		ColorBlue: "Blue",
		ColorGreen: "Green",
	})

	var CardEnum = Enums.MustAdd("Card", map[int]string{
		CardSpade: "Spade",
		CardHeart: "Heart",
		CardDiamond: "Diamond",
		CardClub: "Club",
	})

	//...

	func (g *GameDelegate) EmptyGameState() boardgame.SubState {
		return &gameState{
			MyIntProp: 0,
			MyColorEnumProp: ColorEnum.NewEnumValue(),
		}
	}

	//...

	func NewManager() *boardgame.GameManager {
		//...

		//NewComponentChest will call Finish() on our Enums
		chest := boardgame.NewComponentChest(Enums)

		//...
	}

*/
package enum

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
)

//EnumSet is a set of enums where each Enum's values are unique. Normally you
//will create one in your package, add enums to it during initalization, and
//then use it for all managers you create.
type Set struct {
	finished bool
	enums    map[string]*Enum
	//A map of which int goes to which Enum
	values map[int]*Enum
}

//Enum is a named set of values within a set. Get a new one with
//enumSet.Add().
type Enum struct {
	name         string
	values       map[int]string
	defaultValue int
}

//variable is the underlying type we'll return for both Value and Constant.
type variable struct {
	enum *Enum
	val  int
}

//Const is an instantiation of an Enum that cannot be changed. You retrieve it
//from enum.NewConst(val).
type Const interface {
	Enum() *Enum
	Value() int
	String() string
	Copy() Const
}

//Var is an instantiation of a value that must be set to a value in
//the given enum. You retrieve one from enum.NewVar().
type Var interface {
	Const
	SetValue(int) error
	CopyVar() Var
}

//NewSet returns a new Set. Generally you'll call this once in a
//package and create the set during initalization.
func NewSet() *Set {
	return &Set{
		false,
		make(map[string]*Enum),
		make(map[int]*Enum),
	}
}

//MustCombineSets wraps CombineEnumSets, but instead of erroring will
//panic. Useful for package-level declarations outside of init().
func MustCombineSets(sets ...*Set) *Set {
	result, err := CombineSets(sets...)
	if err != nil {
		panic("Couldn't combine sets: " + err.Error())
	}
	return result
}

//CombineSets returns a new EnumSet that contains all of the EnumSets
//combined into one. The individual enums will literally be the same as the
//enums from the provided sets, so enum equality will work. Generally the sets
//have to know about each other, otherwise they are liable to overlap, which
//will error.
func CombineSets(sets ...*Set) (*Set, error) {
	result := NewSet()
	for i, set := range sets {
		for _, enumName := range set.EnumNames() {
			enum := set.Enum(enumName)
			if err := result.addEnum(enumName, enum); err != nil {
				return nil, errors.New("Couldn't add the " + strconv.Itoa(i) + " enumset because " + enumName + " had error: " + err.Error())
			}
		}
	}
	return result, nil
}

//Finish finalizes an EnumSet so that no more enums may be added. After this
//is called it is safe to use this in a multi-threaded environment. Repeated
//calls do nothing. ComponenChest automatically calls Finish() on the set you
//pass it.
func (e *Set) Finish() {
	e.finished = true
}

//EnumNames returns a list of all names in the Enum.
func (e *Set) EnumNames() []string {
	var result []string
	for key, _ := range e.enums {
		result = append(result, key)
	}
	return result
}

//Returns the Enum with the given name. In general you keep a reference to the
//enum yourself, but this is useful for programatically enumerating the enums.
func (e *Set) Enum(name string) *Enum {
	return e.enums[name]
}

//Membership returns the enum that the given val is a member of.
func (e *Set) Membership(val int) *Enum {
	return e.values[val]
}

//MustAdd is like Add, but instead of an error it will panic if the enum
//cannot be added. This is useful for defining your enums at the package level
//outside of an init().
func (e *Set) MustAdd(enumName string, values map[int]string) *Enum {
	result, err := e.Add(enumName, values)

	if err != nil {
		panic("Couldn't add to enumset: " + err.Error())
	}

	return result
}

/*
Add ads an enum with the given name and values to the enum manager. Will error
if that name has already been added, or any of the int values has been used
for any other enum item already. This means that enums must be unique within a
manager. Check out the package doc for the idiomatic way to initalize enums.
*/
func (e *Set) Add(enumName string, values map[int]string) (*Enum, error) {

	if len(values) == 0 {
		return nil, errors.New("No values provided")
	}

	enum := &Enum{
		enumName,
		make(map[int]string),
		math.MaxInt64,
	}

	seenValues := make(map[string]bool)

	for v, s := range values {

		if seenValues[s] {
			return nil, errors.New("String " + s + " was not unique within enum " + enumName)
		}

		seenValues[s] = true

		enum.values[v] = s

		if v < enum.defaultValue {
			enum.defaultValue = v
		}

	}
	if err := e.addEnum(enumName, enum); err != nil {
		return nil, err
	}
	return enum, nil
}

func (e *Set) addEnum(enumName string, enum *Enum) error {

	if e.finished {
		return errors.New("The set has been finished so no more enums can be added")
	}

	if _, ok := e.enums[enumName]; ok {
		return errors.New("That enum name has already been provided")
	}

	for v, _ := range enum.values {
		if _, ok := e.values[v]; ok {
			//Already registered
			return errors.New("Value " + strconv.Itoa(v) + " was registered twice")
		}

		e.values[v] = enum
	}

	e.enums[enumName] = enum

	return nil
}

//DefaultValue returns the default value for this enum (the lowest valid value
//in it).
func (e *Enum) DefaultValue() int {
	return e.defaultValue
}

//Valid returns whether the given value is a valid member of this enum.
func (e *Enum) Valid(val int) bool {
	_, ok := e.values[val]
	return ok
}

//String returns the string value associated with the given value.
func (e *Enum) String(val int) string {
	return e.values[val]
}

//Name returns the name of this enum; if set is the set this enum is part of,
//set.Enum(enum.Name()) == enum will be true.
func (e *Enum) Name() string {
	return e.name
}

//ValueFromString returns the enum value that corresponds to the given string,
//or -1 if no value has that string.
func (e *Enum) ValueFromString(in string) int {
	for v, str := range e.values {
		if str == in {
			return v
		}
	}
	return -1
}

//Copy returns a copy of the Value, that is equivalent, but will not be
//locked.
func (e *variable) Copy() Const {
	return &variable{
		e.enum,
		e.val,
	}
}

func (e *variable) CopyVar() Var {
	return &variable{
		e.enum,
		e.val,
	}
}

//NewEnumValue returns a new EnumValue associated with this enum, set to the
//Enum's DefaultValue to start.
func (e *Enum) NewVar() Var {
	return &variable{
		e,
		e.DefaultValue(),
	}
}

//MustNewConst is like NewConst, but if it would have errored it panics
//instead. It's convenient for initial set up where the whole app should fail
//to startup if it can't be configured anyway, and dealing with errors would
//be a lot of boilerplate.
func (e *Enum) MustNewConst(val int) Const {
	result, err := e.NewConst(val)
	if err != nil {
		panic("Couldn't create constant: " + err.Error())
	}
	return result
}

//NewConstant returns an enum.Constant that is permanently set to the provided
//val. If that value is not valid for this enum, it will error.
func (e *Enum) NewConst(val int) (Const, error) {
	variable := e.NewVar()
	if err := variable.SetValue(val); err != nil {
		return nil, err
	}
	return variable, nil
}

//The enum marshals as the string value of the enum so it's more readable.
func (e *variable) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

//UnmarshalJSON expects the blob to be the string value. Will error if that
//doesn't correspond to a valid value for this enum.
func (e *variable) UnmarshalJSON(blob []byte) error {
	var str string
	if err := json.Unmarshal(blob, &str); err != nil {
		return err
	}
	val := e.enum.ValueFromString(str)
	if val == -1 {
		return errors.New("That string value had no enum in the value")
	}
	return e.SetValue(val)
}

func (e *variable) Enum() *Enum {
	return e.enum
}

func (e *variable) Value() int {
	return e.val
}

func (e *variable) String() string {
	return e.enum.String(e.val)
}

//SetValue changes the value. Returns true if successful. Will fail if the
//value is locked or the val you want to set is not a valid number for the
//enum this value is associated with.
func (e *variable) SetValue(val int) error {
	if !e.enum.Valid(val) {
		return errors.New("That value is not valid for this enum")
	}
	e.val = val
	return nil
}
