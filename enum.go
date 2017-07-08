package boardgame

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
)

//An EnumValue is an instantiation of a value that must be set to a value in
//the given enum. Inflate() must be called passing the EnumManager it should
//use. If you use an EnumValue in your gameState, playerState, or
//dynamicComponentState it will automatically be inflated for you.
type EnumValue struct {
	locked          bool
	enumName        string
	val             int
	stringToInflate string
	manager         *EnumManager
}

func NewEnumValue(enumName string) *EnumValue {
	return &EnumValue{
		false,
		enumName,
		0,
		"",
		nil,
	}
}

func (e *EnumValue) copy() *EnumValue {
	return &EnumValue{
		e.locked,
		e.enumName,
		e.val,
		e.stringToInflate,
		e.manager,
	}
}

//Valid will return true if the enumName exists for this game's enum manager,
//and the val is a member of that enum.
func (e *EnumValue) Valid() bool {
	if e.manager == nil {
		return false
	}

	if e.manager.DefaultValue(e.enumName) == -1 {
		return false
	}

	if e.manager.Membership(e.val) != e.enumName {
		return false
	}

	return true
}

func (e *EnumValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *EnumValue) UnmarshalJSON(blob []byte) error {
	var str string
	if err := json.Unmarshal(blob, &str); err != nil {
		return err
	}
	e.stringToInflate = str
	return nil
}

func (e *EnumValue) Value() int {
	return e.val
}

func (e *EnumValue) String() string {
	return e.manager.String(e.val)
}

//SetValue changes the value. Returns true if successful. Will fail if the
//value is locked or the val you want to set is not a valid number for the
//enum this value is associated with.
func (e *EnumValue) SetValue(val int) bool {

	if !e.Inflated() {
		return false
	}

	if e.locked {
		return false
	}
	if e.manager.Membership(val) != e.enumName {
		return false
	}
	e.val = val
	return true
}

//Lock locks in the value of the EnumValue, so that in the future all calls to
//SetValue will fail.
func (e *EnumValue) Lock() {
	e.locked = true
}

func (e *EnumValue) Inflated() bool {
	return e.manager != nil
}

//Inflate associates this Value with the EnumManager it is part of, so it can
//check whether its value is legal and retrieve other information about it.
//Inflate must be called to do most actions. If you use an EnumValue in your
//gameState, playerState, or dynamicComponentState it will automatically be
//inflated for you.
func (e *EnumValue) Inflate(manager *EnumManager) {
	e.manager = manager
	if e.manager.DefaultValue(e.enumName) == -1 {
		return
	}
	if e.stringToInflate != "" {
		e.val = e.manager.ValueFromString(e.enumName, e.stringToInflate)
		e.stringToInflate = ""
	}
	if e.manager.Membership(e.val) != e.enumName {
		e.val = e.manager.DefaultValue(e.enumName)
	}
}

type valueRecord struct {
	enumName string
	str      string
}

type enumRecord struct {
	min          int
	strsToValues map[string]int
}

//EnumManager manages all of the enums for a given Game. Enums are useful for
//sanity checking that certain properties are always set in a known way and
//also have convenient String values.
type EnumManager struct {
	frozen bool
	//enums encodes the set of named enums in the manager, along with the
	//lowest-observed value for that set.
	enums  map[string]enumRecord
	values map[int]valueRecord
}

//NewEnumManager returns a new, initialized EnumManager ready for use. In
//general you can just use the one that is automatically available on
//ComponentChest.
func NewEnumManager() *EnumManager {
	return &EnumManager{
		false,
		make(map[string]enumRecord),
		make(map[int]valueRecord),
	}
}

/*
Add ads an enum with the given name and values to the enum manager. Will
error if that name has already been added, or any of the int values has been
used for any other enum item already. This means that enums must be unique
within a manager. The idiomatic way to do this is using chained iota's, like so:
	const (
		//The first enum can start at 0
		ColorRed = iota
		ColorBlue
		ColorGreen
	)

	const (
		//The second enum should start at 1 plus the last item in the previous
		//enum.
		CardSpade = ColorGreen + 1 + iota
		CardHeart
		CardDiamond
		CardClub
	)

	func addEnumsToManager(e *EnumManager) {
		e.Add("Color", map[int]string{
			ColorRed:   "Red",
			ColorBlue:  "Blue",
			ColorGreen: "Green",
		})

		e.Add("Card", map[int]string{
			CardSpade:   "Spade",
			CardHeart:   "Heart",
			CardDiamond: "Diamond",
			CardClub:    "Club",
		})
	}
*/
func (e *EnumManager) Add(name string, values map[int]string) error {

	if e.frozen {
		return errors.New("The enum has been frozen")
	}

	if len(values) == 0 {
		return errors.New("No values provided")
	}

	if _, ok := e.enums[name]; ok {
		return errors.New("That enum name has already been provided")
	}

	eRecord := enumRecord{
		min:          math.MaxInt32,
		strsToValues: make(map[string]int),
	}

	for v, s := range values {
		if _, ok := e.values[v]; ok {
			//Already registered
			return errors.New("Value " + strconv.Itoa(v) + " was registered twice")
		}

		if _, ok := eRecord.strsToValues[s]; ok {
			return errors.New("String " + s + " was not unique within enum " + name)
		}

		e.values[v] = valueRecord{name, s}
		eRecord.strsToValues[s] = v

		if v < eRecord.min {
			eRecord.min = v
		}

		e.enums[name] = eRecord
	}
	return nil
}

//String returns the string for the given value that was configured.
func (e *EnumManager) String(value int) string {
	return e.values[value].str
}

//Membership returns the string name of the enum that that value is part of,
//or "" if not part of an enum.
func (e *EnumManager) Membership(value int) string {
	return e.values[value].enumName
}

//ValueFromString returns the underlying constant value associtaed with that
//str within enumName, or -1 if the enum doesn't exist.
func (e *EnumManager) ValueFromString(enumName string, str string) int {

	rec, ok := e.enums[enumName]

	if !ok {
		return -1
	}

	return rec.strsToValues[str]
}

//DefaultValue returns the lowest value in that enum, or -1 if that enum
//doesn't exist.
func (e *EnumManager) DefaultValue(enumName string) int {
	rec, ok := e.enums[enumName]
	if !ok {
		return -1
	}
	return rec.min
}

//Finish makes it so future calls to Add() will fail.
func (e *EnumManager) Finish() {
	e.frozen = true
}
