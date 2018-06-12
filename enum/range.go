package enum

import (
	"errors"
	"strconv"
	"strings"
)

//RangeEnum is a special type of Enum that also allows indexing via numbers.
type RangeEnum interface {
	Enum

	NewImmutableRangeVal(indexes ...int) (ImmutableRangeVal, error)
	NewRangeVal() RangeVal

	MustNewImmutableRangeVal(indexes ...int) ImmutableRangeVal
	MustNewRangeVal(indexes ...int) RangeVal

	//RangeDimensions will return the number of dimensions used to create this
	//enum with AddRange.
	RangeDimensions() []int

	//ValueToRange will return the multi-dimensional value associated with the
	//given value. A simple convenience wrapper around
	//enum.MutableNewVal(val).RangeValues(), except it won't panic if that
	//value isn't legal.
	ValueToRange(val int) []int

	//RangeToValue takes multi-dimensional indexes and returns the int value
	//associated with those indexes. Will return IllegalValue if it wasn't
	//legal.
	RangeToValue(indexes ...int) int
}

//ImmutableRangeVal is a Val that comes from a RangeEnum.
type ImmutableRangeVal interface {
	ImmutableVal

	//RangeValue will return an array of indexes that this value represents.
	RangeValue() []int
}

//RangeVal is a MutableVal that comes from a RangeEnum.
type RangeVal interface {
	Val

	//RangeValue will return an array of indexes that this value represents.
	RangeValue() []int

	//SetRangeValue can be used to set Range values via the indexes directly.
	SetRangeValue(indexes ...int) error
}

//MustAddRange is like AddRange, but instead of an error it will panic if the
//enum cannot be added. This is useful for defining your enums at the package
//level outside of an init().
func (e *Set) MustAddRange(enumName string, dimensionSize ...int) RangeEnum {
	result, err := e.AddRange(enumName, dimensionSize...)

	if err != nil {
		panic("Couldn't add to enumset: " + err.Error())
	}

	return result
}

func keyForIndexes(values ...int) string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = strconv.Itoa(value)
	}
	return strings.Join(result, rangedValueSeparator)
}

func indexesForKey(key string) []int {
	strs := strings.Split(key, rangedValueSeparator)
	result := make([]int, len(strs))
	for i, str := range strs {
		theInt, err := strconv.Atoi(str)
		if err != nil {
			return nil
		}
		result[i] = theInt
	}
	return result
}

/*
AddRange creates a new Enum that automatically enumerates all indexes in
the multi-dimensional space provided. Each dimensionSize must be greater than
0 or AddRange will error. At its core a RangeEnum is just an enum with a known
mapping of multiple dimensions into string values in a known, stable way, and
with additional convenience methods to automatically convert between that
mapping.

	//Returns an enum like:
	//0 -> '0'
	//1 -> '1'
	AddRange("single", 2)

	// Returns an enum like:
	// 0 -> '0_0'
	// 1 -> '0_1'
	// 2 -> '0_2'
	// 3 -> '1_0'
	// 4 -> '1_1'
	// 5 -> '1_2'
	AddRange("double", 2,3)
*/
func (e *Set) AddRange(enumName string, dimensionSize ...int) (RangeEnum, error) {
	if len(dimensionSize) == 0 {
		return nil, errors.New("No dimensions passed")
	}
	numValues := 1
	for i, dimension := range dimensionSize {
		if dimension <= 0 {
			return nil, errors.New("Dimension " + strconv.Itoa(i) + " is less than or equal to 0, which is illegal")
		}
		numValues *= dimension
	}

	values := make(map[int]string, numValues)
	indexes := make([]int, len(dimensionSize))

	for i := 0; i < numValues; i++ {
		values[i] = keyForIndexes(indexes...)

		//Now, increment indexes.

		//Start at the back, and try to increment one
		for j := len(indexes) - 1; j >= 0; j-- {
			indexes[j]++
			if indexes[j] < dimensionSize[j] {
				break
			}
			//Uh oh, wrapped around.
			indexes[j] = 0
			//The for loop will go back and increment the next thing
		}
	}

	enum, err := e.addEnumImpl(enumName, values)

	if err != nil {
		return nil, err
	}

	enum.dimensions = dimensionSize
	return enum, nil

}

func (e *enum) RangeDimensions() []int {
	return e.dimensions
}

func (e *enum) ValueToRange(val int) []int {
	return indexesForKey(e.String(val))
}

func (e *enum) RangeToValue(indexes ...int) int {
	return e.ValueFromString(keyForIndexes(indexes...))
}

func (e *variable) ImmutableRangeVal() ImmutableRangeVal {
	if e.enum.RangeEnum() == nil {
		return nil
	}
	return e
}

func (e *variable) RangeVal() RangeVal {
	if e.enum.RangeEnum() == nil {
		return nil
	}
	return e
}

func (e *enum) NewImmutableRangeVal(indexes ...int) (ImmutableRangeVal, error) {
	val := e.NewRangeVal()
	if err := val.SetRangeValue(indexes...); err != nil {
		return nil, err
	}
	return val, nil
}

func (e *enum) NewRangeVal() RangeVal {
	return &variable{
		e,
		e.DefaultValue(),
	}
}

func (e *enum) MustNewImmutableRangeVal(indexes ...int) ImmutableRangeVal {
	val, err := e.NewImmutableRangeVal(indexes...)
	if err != nil {
		panic("Couldn't create Range val: " + err.Error())
	}
	return val
}

func (e *enum) MustNewRangeVal(indexes ...int) RangeVal {
	val := e.NewRangeVal()
	if err := val.SetRangeValue(indexes...); err != nil {
		panic("Couldn't create Range val: " + err.Error())
	}
	return val
}

func (e *variable) RangeValue() []int {
	return indexesForKey(e.String())
}

func (e *variable) SetRangeValue(indexes ...int) error {
	return e.SetStringValue(keyForIndexes(indexes...))
}
