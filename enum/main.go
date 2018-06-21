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

The idiomatic way to create an enum is the following.

In components.go:
	const (
		ColorRed = iota
		ColorBlue
		ColorGreen
	)

	const (
		CardSpade = iota
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

And then in your main.go:

	func (g *GameDelegate) EmptyGameState() boardgame.ConfigurableSubState {

		//You could also just return a zero-valued struct if you used struct
		//tags for the enum. See the Constructors section of boardgame package
		//doc for more.
		return &gameState{
			MyIntProp: 0,
			MyColorEnumProp: ColorEnum.NewVal(),
		}
	}

	func (g *GameDelegate) ConfigureEnums() *enum.Set {
		return Enums
	}

This is a fair bit of boilerplate to inlude in your components.go. You can use
the autoreader package to generate the repetitive boilerplate for you.

Instead of the above code for components.go, you'd instead only include the
following:

	//+autoreader
	const (
		ColorRed = iota
		ColorBlue
		ColorGreen
	)

	//+autoreader
	const (
		CardSpade = iota
		CardHeart
		CardDiamond
		CardClub
	)

Then, the rest of the example code shown above in components.go would be
automatically generated, including the ConfigureEnums definition on the
structs in your package that appear to be your game delegate, if you don't
have your own definition of that method. The longest common prefix for each
name in the constant block would be used as the name of the enum. autoreader
has more options for controlling the precise way the enums are created; see
autoreader's package doc for more information.

Ranged Enums

You can also create RangeEnums. These are just normal enums, but with a
multi-index enumeration translated to and from string values in a known way.

When you create RangeEnums, you don't use AutoReader, because the values are
created for you with only minimal configuration.

set.AddRange returns a RangeEnum directly. If you have a normal Enum, you can
use RangeEnum to get access to the RangeEnum, if valid, or nil otherwise.
Generally you should store RangeEnumVal directly in your structs if it's a
range value, so you don't have to up-convert.

Tree Enums

You can also create TreeEnums. These are just normal enums, but that
additionally encode a tree structure on top of the values. A tree enum always
has value 0 as the root node, with string value "", and that all other nodes
have at the top of their ancestor chain.

A common use for Tree Enums is for the Phase enum in your game state, allowing
there to be sets of sub-phases that the game is in.

Tree Enum values can be either branches (has 1 or more children) or leaves
(have no children). Typically a branch node means "all of the values below
me," while a leaf node means "precisely this value". In some contexts it only
makes sense for the value to be set to a leaf node. For example, if the
PhaseEnum in your game is a TreeEnum, then the state will refuse to be saved
if the value is not a leaf value. BranchDefaulValue is the best way to get the
default leaf value within a sub-tree.

Creating Tree Enums with autoreader

autoreader is able to make TreeEnums for you automatically.

The signal to create a TreeEnum is that you have an item in your enum whose string value evaluates to "". (Theoretically the int value of that "" node should also be 0, but the actual constant value for constants in Enums is currently ignored due to #631).

	//+autoreader
	const (
	  //Because the next item's string value is "" (there is no text beyond the shared prefix), this will be a tree enum
	  Phase = iota
	  PhaseRed
	  PhaseBlue
	)

Creates a TreeEnum shaped like:

	""
	  Red
	  Blue

*/
package enum

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"strconv"
)

//IllegalValue is the senitnel value that will be returned for illegal values.
const IllegalValue = math.MaxInt64

const rangedValueSeparator = "_"

//EnumSet is a set of enums where each Enum's values are unique. Normally you
//will create one in your package, add enums to it during initalization, and
//then use it for all managers you create.
type Set struct {
	finished bool
	enums    map[string]Enum
}

//Enum is a named set of values within a set. Get a new one with
//enumSet.Add(). It's an interface to better support anonymous-embedding
//scenarios.
type Enum interface {
	//Values returns all values that are in this enum--all values for which
	//enum.Valid(val) would return true.
	Values() []int
	//DefaultValue returns the default value for this enum (the lowest valid
	//value in it). TreeEnums will instead return BranchDefaultValue for 0, to
	//ensure that the DefaultValue is a leaf.z
	DefaultValue() int
	//RandomValue returns a random value that is Valid() for this enum.
	RandomValue() int
	//Valid returns whether the given value is a valid member of this enum.
	Valid(val int) bool
	//String returns the string value associated with the given value.
	String(val int) string
	//Name returns the name of this enum; if set is the set this enum is part of,
	//set.Enum(enum.Name()) == enum will be true.
	Name() string
	//ValueFromString returns the enum value that corresponds to the given string,
	//or IllegalValue if no value has that string.
	ValueFromString(in string) int
	//NewVal returns an enum.Val that is permanently set to the provided
	//val. If that value is not valid for this enum, it will error.
	NewImmutableVal(val int) (ImmutableVal, error)
	NewVal() Val
	//NewMutableVal returns a new EnumValue associated with this enum, set to the
	//Enum's DefaultValue to start.
	//NewDefaultVal is a convenience shortcut for creating a new const that is
	//set to the default value, which is moderately common enough that it makes
	//sense to do it without the possibility of errors.
	NewDefaultVal() ImmutableVal

	//MustNewVal is like NewVal, but if it would have errored it panics
	//instead. It's convenient for initial set up where the whole app should fail
	//to startup if it can't be configured anyway, and dealing with errors would
	//be a lot of boilerplate.
	MustNewImmutableVal(val int) ImmutableVal
	MustNewVal(val int) Val

	//RangeEnum will return a version of this Enum that satisifies the
	//RangeEnum interface, if possible, nil otherwise.
	RangeEnum() RangeEnum

	//TreeEnum will return a version of this Enum that satisfies the TreeEnum
	//interface, if possible, nil otherwise.
	TreeEnum() TreeEnum
}

//enum is the underlying type we use to implement Enum.
type enum struct {
	name         string
	values       map[int]string
	defaultValue int
	dimensions   []int
	//parents is the direct map passed to us to start.
	parents map[int]int
	//children is created based on the parents map we got.
	children map[int][]int
}

//variable is the underlying type we'll return for both Value and Constant.
type variable struct {
	enum Enum
	val  int
}

//ImmutableVal is an instantiation of an Enum that cannot be changed. You retrieve it
//from enum.NewImmutableVal(val).
type ImmutableVal interface {
	Enum() Enum
	Value() int
	String() string
	ImmutableCopy() ImmutableVal
	Copy() Val
	Equals(other ImmutableVal) bool

	//ImmutableRangeVal will return a version of this Val that implements RangeVal(),
	//if that's possible, nil otherwise.≈
	ImmutableRangeVal() ImmutableRangeVal

	//ImmutableTreeVAl will return a version of this Val that implements
	//TreeVal, if that's possible, nil otherwise.
	ImmutableTreeVal() ImmutableTreeVal
}

//MutableVal is an instantiation of a value that must be set to a value in the
//given enum. You retrieve one from enum.NewMutableVal().
type Val interface {
	ImmutableVal
	//SetValue changes the value. Returns true if successful. Will fail if the
	//value is locked or the val you want to set is not a valid number for the
	//enum this value is associated with.
	SetValue(val int) error
	//SetStringValue sets the value to the value associated with that string.
	SetStringValue(str string) error

	//RangeVal returns a version of this Val that implements RangeVal, if
	//that's posisble, nil otherwise.
	RangeVal() RangeVal

	//TreeVal returns a version of this Val that implements TreeVal, if that's
	//possible, nil otherwise.
	TreeVal() TreeVal
}

//NewSet returns a new Set. Generally you'll call this once in a
//package and create the set during initalization.
func NewSet() *Set {
	return &Set{
		false,
		make(map[string]Enum),
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
//enums from the provided sets, so enum equality will work.
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

func (e *Set) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.enums)
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
func (e *Set) Enum(name string) Enum {
	return e.enums[name]
}

//MustAdd is like Add, but instead of an error it will panic if the enum
//cannot be added. This is useful for defining your enums at the package level
//outside of an init().
func (e *Set) MustAdd(enumName string, values map[int]string) Enum {
	result, err := e.Add(enumName, values)

	if err != nil {
		panic("Couldn't add to enumset: " + err.Error())
	}

	return result
}

/*
Add ads an enum with the given name and values to the enum manager. Will error
if that name has already been added or if the config you provide has more than
one string with the same value.
*/
func (e *Set) Add(enumName string, values map[int]string) (Enum, error) {
	return e.addEnumImpl(enumName, values)
}

func (e *Set) addEnumImpl(enumName string, values map[int]string) (*enum, error) {
	if len(values) == 0 {
		return nil, errors.New("No values provided")
	}

	enum := &enum{
		enumName,
		make(map[int]string),
		math.MaxInt64,
		nil,
		nil,
		nil,
	}

	seenValues := make(map[string]bool)

	for v, s := range values {

		numString := strconv.Itoa(v)

		if seenValues[numString] {
			return nil, errors.New("The string value of " + numString + " was already in the enum")
		}

		if seenValues[s] {
			return nil, errors.New("String " + s + " was not unique within enum " + enumName)
		}

		//We put in both values into seenValues here because it's legal for a
		//const to have its string-value be the stringification of its own
		//int.
		seenValues[numString] = true
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

func (e *Set) addEnum(enumName string, enum Enum) error {

	if e.finished {
		return errors.New("The set has been finished so no more enums can be added")
	}

	if _, ok := e.enums[enumName]; ok {
		return errors.New("That enum name has already been provided")
	}

	e.enums[enumName] = enum

	return nil
}

func (e *enum) TreeEnum() TreeEnum {
	if e.parents != nil {
		return e
	}

	return nil
}

func (e *enum) RangeEnum() RangeEnum {
	if len(e.dimensions) > 0 {
		return e
	}
	return nil
}

func (e *enum) Values() []int {
	result := make([]int, len(e.values))
	counter := 0
	for key, _ := range e.values {
		result[counter] = key
		counter++
	}
	return result
}

func (e *enum) DefaultValue() int {
	return e.defaultValue
}

func (e *enum) RandomValue() int {
	keys := make([]int, len(e.values))

	i := 0
	for key, _ := range e.values {
		keys[i] = key
		i++
	}
	return keys[rand.Intn(len(keys))]
}

func (e *enum) Valid(val int) bool {
	_, ok := e.values[val]
	return ok
}

func (e *enum) String(val int) string {
	return e.values[val]
}

func (e *enum) Name() string {
	return e.name
}

func (e *enum) ValueFromString(in string) int {
	for v, str := range e.values {
		if str == in {
			return v
		}
	}
	//Hmmm... see if they gave us the string equivalent of the int?
	num, err := strconv.Atoi(in)
	if err == nil {
		if _, ok := e.values[num]; ok {
			return num
		}
	}
	return IllegalValue
}

//Copy returns a copy of the Value, that is equivalent, but will not be
//locked.
func (e *variable) ImmutableCopy() ImmutableVal {
	return &variable{
		e.enum,
		e.val,
	}
}

func (e *variable) Copy() Val {
	return &variable{
		e.enum,
		e.val,
	}
}

func (e *enum) NewVal() Val {
	return e.NewRangeVal()
}

func (e *enum) MustNewVal(val int) Val {
	enu := e.NewVal()
	if err := enu.SetValue(val); err != nil {
		panic("Couldn't create mutable val: " + err.Error())
	}
	return enu
}

func (e *enum) MustNewImmutableVal(val int) ImmutableVal {
	result, err := e.NewImmutableVal(val)
	if err != nil {
		panic("Couldn't create constant: " + err.Error())
	}
	return result
}

func (e *enum) NewDefaultVal() ImmutableVal {
	c, err := e.NewImmutableVal(e.DefaultValue())
	if err != nil {
		panic("Unexpected error in NewDefaultConst: " + err.Error())
	}
	return c
}

func (e *enum) NewImmutableVal(val int) (ImmutableVal, error) {
	variable := e.NewVal()
	if err := variable.SetValue(val); err != nil {
		return nil, err
	}
	return variable, nil
}

func (e *enum) MarshalJSON() ([]byte, error) {
	obj := map[string]interface{}{
		"Values":       e.values,
		"DefaultValue": e.defaultValue,
	}

	if e.RangeEnum() != nil {
		obj["Dimensions"] = e.dimensions
	}

	return json.Marshal(obj)
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
	if val == IllegalValue {
		return errors.New("That string value had no enum in the value")
	}
	return e.SetValue(val)
}

func (e *variable) Enum() Enum {
	return e.enum
}

func (e *variable) Value() int {
	return e.val
}

func (e *variable) String() string {
	return e.enum.String(e.val)
}

func (e *variable) SetValue(val int) error {
	if !e.enum.Valid(val) {
		return errors.New("That value is not valid for this enum")
	}
	e.val = val
	return nil
}

func (e *variable) SetStringValue(str string) error {
	val := e.Enum().ValueFromString(str)
	return e.SetValue(val)
}

//Equals returns true if the two Consts are equivalent.
func (e *variable) Equals(other ImmutableVal) bool {
	if other == nil {
		return false
	}
	if e.Enum() != other.Enum() {
		return false
	}
	if e.Value() != other.Value() {
		return false
	}
	return true
}
