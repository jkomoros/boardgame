package boardgame

import (
	"errors"
	"math"
	"strconv"
)

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
		ColorRed = iota
		ColorBlue
		ColorGreen
	)

	const (
		CardSpade = ColorGreen + 1 + iota
		CardHeart
		CardDiamond
		CardClub
	)
*/
func (e *EnumManager) Add(name string, values map[int]string) error {

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

//DefaultValue returns the lowest value in that enum, or 0 if that enum
//doesn't exist.
func (e *EnumManager) DefaultValue(enumName string) int {
	return e.enums[enumName].min
}
