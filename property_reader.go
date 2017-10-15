package boardgame

import (
	"fmt"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

//Property reader is a way to read out properties on an object with unknown
//shape.
type PropertyReader interface {
	//Props returns a list of all property names that are defined for this
	//object.
	Props() map[string]PropertyType
	//IntProp fetches the int property with that name, returning an error if
	//that property doese not exist.
	IntProp(name string) (int, error)
	BoolProp(name string) (bool, error)
	StringProp(name string) (string, error)
	EnumProp(name string) (enum.Val, error)
	IntSliceProp(name string) ([]int, error)
	BoolSliceProp(name string) ([]bool, error)
	StringSliceProp(name string) ([]string, error)
	PlayerIndexSliceProp(name string) ([]PlayerIndex, error)
	PlayerIndexProp(name string) (PlayerIndex, error)
	StackProp(name string) (Stack, error)
	TimerProp(name string) (Timer, error)
	//Prop fetches the given property generically. If you already know the
	//type, it's better to use the typed methods.
	Prop(name string) (interface{}, error)
}

//PropertyType is an enumeration of the types that are legal to have on an
//underyling object that can return a Reader. This ensures that State objects
//are not overly complex and can be reasoned about clearnly.
type PropertyType int

const (
	TypeIllegal PropertyType = iota
	TypeInt
	TypeBool
	TypeString
	TypePlayerIndex
	TypeEnum
	TypeIntSlice
	TypeBoolSlice
	TypeStringSlice
	TypePlayerIndexSlice
	TypeStack
	TypeTimer
)

//Property read setter is a way to enumerate and set properties on an object with an
//unknown shape.
type PropertyReadSetter interface {
	//All PropertyReadSetters have read interfaces
	PropertyReader

	//SetTYPEProp sets the given property name to the given type.
	SetIntProp(name string, value int) error
	SetBoolProp(name string, value bool) error
	SetStringProp(name string, value string) error
	SetPlayerIndexProp(name string, value PlayerIndex) error
	SetMutableEnumProp(name string, value enum.MutableVal) error
	SetIntSliceProp(name string, value []int) error
	SetBoolSliceProp(name string, value []bool) error
	SetStringSliceProp(name string, value []string) error
	SetPlayerIndexSliceProp(name string, value []PlayerIndex) error
	SetStackProp(name string, value Stack) error
	SetTimerProp(name string, value Timer) error

	//For interface types the setter also wants to give access to the mutable
	//underlying value so it can be mutated in place.
	MutableEnumProp(name string) (enum.MutableVal, error)

	//SetProp sets the property with the given name. If the value does not
	//match the underlying slot type, it should return an error. If you know
	//the underlying type it's always better to use the typed accessors.
	SetProp(name string, value interface{}) error
}

func (t PropertyType) String() string {
	switch t {
	case TypeIllegal:
		return "TypeIllegal"
	case TypeInt:
		return "TypeInt"
	case TypeBool:
		return "TypeBool"
	case TypeString:
		return "TypeString"
	case TypePlayerIndex:
		return "TypePlayerIndex"
	case TypeEnum:
		return "TypeEnum"
	case TypeIntSlice:
		return "TypeIntSlice"
	case TypeBoolSlice:
		return "TypeBoolSlice"
	case TypeStringSlice:
		return "TypeStringSlice"
	case TypePlayerIndexSlice:
		return "TypePlayerIndexSlice"
	case TypeStack:
		return "TypeStack"
	case TypeTimer:
		return "TypeTimer"
	default:
		return "TypeIllegal"
	}
}

//TODO: protect access to this with a mutex.
var defaultReaderCacheLock sync.RWMutex
var defaultReaderCache map[interface{}]*defaultReader

func init() {
	defaultReaderCacheLock.Lock()
	defaultReaderCache = make(map[interface{}]*defaultReader)
	defaultReaderCacheLock.Unlock()
}

//genericReader is a generic PropertyReadSetter that allows users to Set
//various properties and have them configed.
type genericReader struct {
	types  map[string]PropertyType
	values map[string]interface{}
}

type defaultReader struct {
	i     interface{}
	props map[string]PropertyType
}

//DefaultReader returns an object that satisfies the PropertyReader interface
//for the given concrete object, using reflection. Make it easy to implement
//the Reader method in a line. It will return an existing wrapper or create a
//new one if necessary. This used to be public, but it never really made sense
//to expose and doesn't understand embedded types, so it's now just used for
//testing within the package.
func getDefaultReader(i interface{}) PropertyReader {
	return getDefaultReadSetter(i)
}

//DefaultReadSetter returns an object that satisfies the PropertyReadSetter
//interface for the given concrete object, using reflection. Make it easy to
//implement the Reader method in a line. It will return an existing wrapper or
//create a new one if necessary. This used to be public, but it never really
//made sense to expose and doesn't understand embedded types, so it's now just
//used for testing within the package.
func getDefaultReadSetter(i interface{}) PropertyReadSetter {

	defaultReaderCacheLock.RLock()
	reader := defaultReaderCache[i]
	defaultReaderCacheLock.RUnlock()

	if reader != nil {
		return reader
	}

	result := &defaultReader{
		i: i,
	}

	//Sanity check right now if the object we were just passed will have
	//serious errors trying to parse it.
	if _, err := result.propsImpl(); err != nil {
		//It's OK to panic here because defaultreader is only used in tests
		//and its' not exported.
		panic("Got error from default propsImpl: " + err.Error())
		return nil
	}

	defaultReaderCacheLock.Lock()
	defaultReaderCache[i] = result
	defaultReaderCacheLock.Unlock()
	return result
}

func propertyReaderImplNameShouldBeIncluded(name string) bool {

	if len(name) < 1 {
		return false
	}

	firstChar := []rune(name)[0]

	if firstChar != unicode.ToUpper(firstChar) {
		//It was not upper case, thus private, thus should not be included.
		return false
	}

	//TODO: check if the struct says propertyreader:omit

	return true
}

func (d *defaultReader) Props() map[string]PropertyType {
	result, err := d.propsImpl()
	if err != nil {
		//OK to panic here because we only use defaultReader for tests in this
		//package and it's not exported.
		panic("Default Reader got error: " + err.Error())
	}
	return result
}

func (d *defaultReader) propsImpl() (map[string]PropertyType, error) {

	//TODO: skip fields that have a propertyreader:omit

	if d.props == nil {

		obj := d.i

		result := make(map[string]PropertyType)

		s := reflect.ValueOf(obj).Elem()
		typeOfObj := s.Type()

		for i := 0; i < s.NumField(); i++ {
			name := typeOfObj.Field(i).Name

			if typeOfObj.Field(i).Anonymous {
				//Anonymous fields are likely BaseSubState.
				continue
			}

			field := s.Field(i)

			if !propertyReaderImplNameShouldBeIncluded(name) {
				continue
			}

			var pType PropertyType

			switch field.Type().Kind() {
			case reflect.Bool:
				pType = TypeBool
			case reflect.Int:

				//Both Int and PlayerIndex have Kind() int, so we need to
				//disambiguate.
				intType := field.Type().String()

				if strings.Contains(intType, "PlayerIndex") {
					pType = TypePlayerIndex
				} else {
					pType = TypeInt
				}
			case reflect.String:
				pType = TypeString
			case reflect.Slice:
				sliceType := field.Type().String()

				if strings.Contains(sliceType, "string") {
					pType = TypeStringSlice
				} else if strings.Contains(sliceType, "int") {
					pType = TypeIntSlice
				} else if strings.Contains(sliceType, "bool") {
					pType = TypeBoolSlice
				} else if strings.Contains(sliceType, "PlayerIndex") {
					pType = TypePlayerIndexSlice
				}
			case reflect.Interface:
				interfaceType := field.Type().String()
				if strings.Contains(interfaceType, "enum.MutableVal") {
					pType = TypeEnum
				} else if strings.Contains(interfaceType, "enum.Val") {
					pType = TypeEnum
				} else if strings.Contains(interfaceType, "Stack") {
					pType = TypeStack
				} else if strings.Contains(interfaceType, "Timer") {
					pType = TypeTimer
				}
			default:
				return nil, errors.New("Unsupported field in underlying type" + strconv.Itoa(int(field.Type().Kind())))
			}

			result[name] = pType

		}
		d.props = result
	}

	return d.props, nil
}

func (d *defaultReader) IntProp(name string) (int, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypeInt {
		return 0, errors.New("That property is not an int: " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	return int(s.FieldByName(name).Int()), nil
}

func (d *defaultReader) BoolProp(name string) (bool, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypeBool {
		return false, errors.New("That property is not a bool: " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	return s.FieldByName(name).Bool(), nil
}

func (d *defaultReader) StringProp(name string) (string, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypeString {
		return "", errors.New("That property is not a string: " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	return s.FieldByName(name).String(), nil
}

func (d *defaultReader) PlayerIndexProp(name string) (PlayerIndex, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypePlayerIndex {
		return 0, errors.New("That property is not a PlayerIndex: " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	return PlayerIndex(s.FieldByName(name).Int()), nil
}

func (d *defaultReader) EnumProp(name string) (enum.Val, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypeEnum {
		return nil, errors.New("That property is not a Enum: " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	field := s.FieldByName(name)
	if field.IsNil() {
		//This isn't an error; it's just that we shouldn't dereference it.
		return nil, nil
	}
	result := field.Interface().(enum.Val)
	return result, nil
}

func (d *defaultReader) IntSliceProp(name string) ([]int, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypeIntSlice {
		return nil, errors.New("That property is not an int slice: " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	field := s.FieldByName(name)

	result := make([]int, field.Len())

	for i := 0; i < field.Len(); i++ {
		result[i] = int(field.Index(i).Int())
	}

	return result, nil
}

func (d *defaultReader) BoolSliceProp(name string) ([]bool, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypeBoolSlice {
		return nil, errors.New("That property is not a bool slice: " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	field := s.FieldByName(name)

	result := make([]bool, field.Len())

	for i := 0; i < field.Len(); i++ {
		result[i] = field.Index(i).Bool()
	}

	return result, nil
}

func (d *defaultReader) StringSliceProp(name string) ([]string, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypeStringSlice {
		return nil, errors.New("That property is not a string slice: " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	field := s.FieldByName(name)

	result := make([]string, field.Len())

	for i := 0; i < field.Len(); i++ {
		result[i] = field.Index(i).String()
	}

	return result, nil
}

func (d *defaultReader) PlayerIndexSliceProp(name string) ([]PlayerIndex, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypePlayerIndexSlice {
		return nil, errors.New("That property is not a player index slice: " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	field := s.FieldByName(name)

	result := make([]PlayerIndex, field.Len())

	for i := 0; i < field.Len(); i++ {
		result[i] = PlayerIndex(field.Index(i).Int())
	}

	return result, nil
}

func (d *defaultReader) StackProp(name string) (Stack, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypeStack {
		return nil, errors.New("That property is not a Stack: " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	field := s.FieldByName(name)
	if field.IsNil() {
		//This isn't an error; it's just that we shouldn't dereference it.
		return nil, nil
	}
	result := field.Interface().(Stack)
	return result, nil
}

func (d *defaultReader) TimerProp(name string) (Timer, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypeTimer {
		return nil, errors.New("That property is not a Timer: " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	field := s.FieldByName(name)
	if field.IsNil() {
		//This isn't an error; it's just that we shouldn't dereference it.
		return nil, nil
	}
	result := field.Interface().(Timer)
	return result, nil
}

func (d *defaultReader) Prop(name string) (interface{}, error) {

	props := d.Props()

	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No such property")
	}

	if propType == TypeIllegal {
		return nil, errors.New("That property is not a supported type")
	}

	s := reflect.ValueOf(d.i).Elem()
	return s.FieldByName(name).Interface(), nil
}

func (d *defaultReader) SetIntProp(name string, val int) (err error) {
	props := d.Props()

	if props[name] != TypeInt {
		return errors.New("That property is not a settable int")
	}

	s := reflect.ValueOf(d.i).Elem()

	f := s.FieldByName(name)

	if !f.IsValid() {
		return errors.New("that name was not available on the struct")
	}

	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	f.SetInt(int64(val))

	return nil

}

func (d *defaultReader) SetBoolProp(name string, val bool) (err error) {
	props := d.Props()

	if props[name] != TypeBool {
		return errors.New("That property is not a settable bool")
	}

	s := reflect.ValueOf(d.i).Elem()

	f := s.FieldByName(name)

	if !f.IsValid() {
		return errors.New("that name was not available on the struct")
	}

	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	f.SetBool(val)

	return nil

}

func (d *defaultReader) SetStringProp(name string, val string) (err error) {
	props := d.Props()

	if props[name] != TypeString {
		return errors.New("That property is not a settable string")
	}

	s := reflect.ValueOf(d.i).Elem()

	f := s.FieldByName(name)

	if !f.IsValid() {
		return errors.New("that name was not available on the struct")
	}

	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	f.SetString(val)

	return nil

}

func (d *defaultReader) SetPlayerIndexProp(name string, val PlayerIndex) (err error) {
	props := d.Props()

	if props[name] != TypePlayerIndex {
		return errors.New("That property is not a settable player index")
	}

	s := reflect.ValueOf(d.i).Elem()

	f := s.FieldByName(name)

	if !f.IsValid() {
		return errors.New("that name was not available on the struct")
	}

	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	f.SetInt(int64(val))

	return nil

}

func (d *defaultReader) MutableEnumProp(name string) (enum.MutableVal, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypeEnum {
		return nil, errors.New("That property is not a Enum: " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	field := s.FieldByName(name)
	if field.IsNil() {
		//This isn't an error; it's just that we shouldn't dereference it.
		return nil, nil
	}
	result := field.Interface().(enum.MutableVal)
	return result, nil
}

func (d *defaultReader) SetMutableEnumProp(name string, val enum.MutableVal) (err error) {
	props := d.Props()

	if props[name] != TypeEnum {
		return errors.New("That property is not a settable enum var")
	}

	s := reflect.ValueOf(d.i).Elem()

	f := s.FieldByName(name)

	if !f.IsValid() {
		return errors.New("that name was not available on the struct")
	}

	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	f.Set(reflect.ValueOf(val))

	return nil

}

func (d *defaultReader) SetIntSliceProp(name string, val []int) (err error) {
	props := d.Props()

	if props[name] != TypeIntSlice {
		return errors.New("That property is not a settable int slice")
	}

	s := reflect.ValueOf(d.i).Elem()

	f := s.FieldByName(name)

	if !f.IsValid() {
		return errors.New("that name was not available on the struct")
	}

	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	f.Set(reflect.ValueOf(val))

	return nil

}

func (d *defaultReader) SetBoolSliceProp(name string, val []bool) (err error) {
	props := d.Props()

	if props[name] != TypeBoolSlice {
		return errors.New("That property is not a settable int slice")
	}

	s := reflect.ValueOf(d.i).Elem()

	f := s.FieldByName(name)

	if !f.IsValid() {
		return errors.New("that name was not available on the struct")
	}

	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	f.Set(reflect.ValueOf(val))

	return nil

}

func (d *defaultReader) SetStringSliceProp(name string, val []string) (err error) {
	props := d.Props()

	if props[name] != TypeStringSlice {
		return errors.New("That property is not a settable string slice")
	}

	s := reflect.ValueOf(d.i).Elem()

	f := s.FieldByName(name)

	if !f.IsValid() {
		return errors.New("that name was not available on the struct")
	}

	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	f.Set(reflect.ValueOf(val))

	return nil

}

func (d *defaultReader) SetPlayerIndexSliceProp(name string, val []PlayerIndex) (err error) {
	props := d.Props()

	if props[name] != TypePlayerIndexSlice {
		return errors.New("That property is not a settable player index slice")
	}

	s := reflect.ValueOf(d.i).Elem()

	f := s.FieldByName(name)

	if !f.IsValid() {
		return errors.New("that name was not available on the struct")
	}

	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	f.Set(reflect.ValueOf(val))

	return nil

}

func (d *defaultReader) SetStackProp(name string, val Stack) (err error) {
	props := d.Props()

	if props[name] != TypeStack {
		return errors.New("That property is not a settable stack")
	}

	s := reflect.ValueOf(d.i).Elem()

	f := s.FieldByName(name)

	if !f.IsValid() {
		return errors.New("that name was not available on the struct")
	}

	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	f.Set(reflect.ValueOf(val))

	return nil

}

func (d *defaultReader) SetTimerProp(name string, val Timer) (err error) {
	props := d.Props()

	if props[name] != TypeTimer {
		return errors.New("That property is not a settable Timer")
	}

	s := reflect.ValueOf(d.i).Elem()

	f := s.FieldByName(name)

	if !f.IsValid() {
		return errors.New("that name was not available on the struct")
	}

	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	f.Set(reflect.ValueOf(val))

	return nil

}

func (d *defaultReader) SetProp(name string, val interface{}) (err error) {

	obj := d.i

	props := d.Props()

	propType, ok := props[name]

	if !ok {
		return errors.New("Not a settable name")
	}

	if propType == TypeIllegal {
		return errors.New("Unsupported type of prop")
	}

	if !propertyReaderImplNameShouldBeIncluded(name) {
		return errors.New("That name is not valid to set.")
	}

	s := reflect.ValueOf(obj).Elem()

	f := s.FieldByName(name)

	if !f.IsValid() {
		return errors.New("That name was not available on the struct")
	}

	//f.Set will panic if it's not possible to set the field to the given
	//value kind.
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	f.Set(reflect.ValueOf(val))

	return

}

func newGenericReader() *genericReader {
	return &genericReader{
		make(map[string]PropertyType),
		make(map[string]interface{}),
	}
}

//Implement reader so we can be used directly or in e.g.
//ComputedPropertyCollection.
func (g *genericReader) ReadSetter() PropertyReadSetter {
	return g
}

func (g *genericReader) Reader() PropertyReader {
	return g
}

func (g *genericReader) Props() map[string]PropertyType {
	return g.types
}

func (g *genericReader) Prop(name string) (interface{}, error) {
	val, ok := g.values[name]

	if !ok {
		return nil, errors.New("No such property: " + name)
	}

	return val, nil
}

func (g *genericReader) IntProp(name string) (int, error) {
	val, err := g.Prop(name)

	if err != nil {
		return 0, err
	}

	propType, ok := g.types[name]

	if !ok {
		return 0, errors.New("Unexpected error: Missing Prop type for " + name)
	}

	if propType != TypeInt {
		return 0, errors.New(name + "was expected to be TypeInt but was not")
	}

	return val.(int), nil
}

func (g *genericReader) BoolProp(name string) (bool, error) {
	val, err := g.Prop(name)

	if err != nil {
		return false, err
	}

	propType, ok := g.types[name]

	if !ok {
		return false, errors.New("Unexpected error: Missing Prop type for " + name)
	}

	if propType != TypeBool {
		return false, errors.New(name + "was expected to be TypeBool but was not")
	}

	return val.(bool), nil
}

func (g *genericReader) StringProp(name string) (string, error) {
	val, err := g.Prop(name)

	if err != nil {
		return "", err
	}

	propType, ok := g.types[name]

	if !ok {
		return "", errors.New("Unexpected error: Missing Prop type for " + name)
	}

	if propType != TypeString {
		return "", errors.New(name + "was expected to be TypeString but was not")
	}

	return val.(string), nil
}

func (g *genericReader) PlayerIndexProp(name string) (PlayerIndex, error) {
	val, err := g.Prop(name)

	if err != nil {
		return 0, err
	}

	propType, ok := g.types[name]

	if !ok {
		return 0, errors.New("Unexpected error: Missing Prop type for " + name)
	}

	if propType != TypePlayerIndex {
		return 0, errors.New(name + "was expected to be TypePlayerIndex but was not")
	}

	return val.(PlayerIndex), nil
}

func (g *genericReader) EnumProp(name string) (enum.Val, error) {
	val, err := g.Prop(name)

	if err != nil {
		return nil, err
	}

	propType, ok := g.types[name]

	if !ok {
		return nil, errors.New("Unexpected error: Missing Prop type for " + name)
	}

	if propType != TypeEnum {
		return nil, errors.New(name + "was expected to be TypeEnum but was not")
	}

	return val.(enum.Val), nil
}

func (g *genericReader) IntSliceProp(name string) ([]int, error) {
	val, err := g.Prop(name)

	if err != nil {
		return nil, err
	}

	propType, ok := g.types[name]

	if !ok {
		return nil, errors.New("Unexpected error: Missing Prop type for " + name)
	}

	if propType != TypeIntSlice {
		return nil, errors.New(name + "was expected to be TypeIntSlice but was not")
	}

	return val.([]int), nil
}

func (g *genericReader) BoolSliceProp(name string) ([]bool, error) {
	val, err := g.Prop(name)

	if err != nil {
		return nil, err
	}

	propType, ok := g.types[name]

	if !ok {
		return nil, errors.New("Unexpected error: Missing Prop type for " + name)
	}

	if propType != TypeBoolSlice {
		return nil, errors.New(name + "was expected to be TypeBoolSlice but was not")
	}

	return val.([]bool), nil
}

func (g *genericReader) StringSliceProp(name string) ([]string, error) {
	val, err := g.Prop(name)

	if err != nil {
		return nil, err
	}

	propType, ok := g.types[name]

	if !ok {
		return nil, errors.New("Unexpected error: Missing Prop type for " + name)
	}

	if propType != TypeStringSlice {
		return nil, errors.New(name + "was expected to be TypeStringSlice but was not")
	}

	return val.([]string), nil
}

func (g *genericReader) PlayerIndexSliceProp(name string) ([]PlayerIndex, error) {
	val, err := g.Prop(name)

	if err != nil {
		return nil, err
	}

	propType, ok := g.types[name]

	if !ok {
		return nil, errors.New("Unexpected error: Missing Prop type for " + name)
	}

	if propType != TypePlayerIndexSlice {
		return nil, errors.New(name + "was expected to be TypePlayerIndexSlice but was not")
	}

	return val.([]PlayerIndex), nil
}

func (g *genericReader) StackProp(name string) (Stack, error) {
	val, err := g.Prop(name)

	if err != nil {
		return nil, err
	}

	propType, ok := g.types[name]

	if !ok {
		return nil, errors.New("Unexpected error: Missing Prop type for " + name)
	}

	if propType != TypeStack {
		return nil, errors.New(name + "was expected to be TypeStack but was not")
	}

	return val.(Stack), nil
}

func (g *genericReader) TimerProp(name string) (Timer, error) {
	val, err := g.Prop(name)

	if err != nil {
		return nil, err
	}

	propType, ok := g.types[name]

	if !ok {
		return nil, errors.New("Unexpected error: Missing Prop type for " + name)
	}

	if propType != TypeTimer {
		return nil, errors.New(name + "was expected to be TypeTimer but was not")
	}

	return val.(Timer), nil
}

func (g *genericReader) SetProp(name string, val interface{}) error {

	propType, ok := g.types[name]

	if ok && propType != TypeIllegal {
		return errors.New("That property was already set but was a different type")
	}

	g.types[name] = TypeIllegal
	g.values[name] = val

	return nil
}

func (g *genericReader) SetIntProp(name string, val int) error {
	propType, ok := g.types[name]

	if ok && propType != TypeInt {
		return errors.New("That property was already set but was not an int")
	}

	g.types[name] = TypeInt
	g.values[name] = val

	return nil
}

func (g *genericReader) SetBoolProp(name string, val bool) error {
	propType, ok := g.types[name]

	if ok && propType != TypeBool {
		return errors.New("That property was already set but was not an bool")
	}

	g.types[name] = TypeBool
	g.values[name] = val

	return nil
}

func (g *genericReader) SetStringProp(name string, val string) error {
	propType, ok := g.types[name]

	if ok && propType != TypeString {
		return errors.New("That property was already set but was not an string")
	}

	g.types[name] = TypeString
	g.values[name] = val

	return nil
}

func (g *genericReader) SetPlayerIndexProp(name string, val PlayerIndex) error {
	propType, ok := g.types[name]

	if ok && propType != TypePlayerIndex {
		return errors.New("That property was already set but was not an PlayerIndex")
	}

	g.types[name] = TypePlayerIndex
	g.values[name] = val

	return nil
}

func (g *genericReader) MutableEnumProp(name string) (enum.MutableVal, error) {
	val, err := g.Prop(name)

	if err != nil {
		return nil, err
	}

	propType, ok := g.types[name]

	if !ok {
		return nil, errors.New("Unexpected error: Missing Prop type for " + name)
	}

	if propType != TypeEnum {
		return nil, errors.New(name + "was expected to be TypeEnum but was not")
	}

	return val.(enum.MutableVal), nil
}

func (g *genericReader) SetMutableEnumProp(name string, val enum.MutableVal) error {
	propType, ok := g.types[name]

	if ok && propType != TypeEnum {
		return errors.New("That property was already set but was not an enum mutable val")
	}

	g.types[name] = TypeEnum
	g.values[name] = val

	return nil
}

func (g *genericReader) SetIntSliceProp(name string, val []int) error {
	propType, ok := g.types[name]

	if ok && propType != TypeIntSlice {
		return errors.New("That property was already set but was not an int slice")
	}

	g.types[name] = TypeIntSlice
	g.values[name] = val

	return nil
}

func (g *genericReader) SetBoolSliceProp(name string, val []bool) error {
	propType, ok := g.types[name]

	if ok && propType != TypeBoolSlice {
		return errors.New("That property was already set but was not an bool slice")
	}

	g.types[name] = TypeBoolSlice
	g.values[name] = val

	return nil
}

func (g *genericReader) SetStringSliceProp(name string, val []string) error {
	propType, ok := g.types[name]

	if ok && propType != TypeStringSlice {
		return errors.New("That property was already set but was not an string slice")
	}

	g.types[name] = TypeStringSlice
	g.values[name] = val

	return nil
}

func (g *genericReader) SetPlayerIndexSliceProp(name string, val []PlayerIndex) error {
	propType, ok := g.types[name]

	if ok && propType != TypePlayerIndexSlice {
		return errors.New("That property was already set but was not an PlayerIndex slice")
	}

	g.types[name] = TypePlayerIndexSlice
	g.values[name] = val

	return nil
}

func (g *genericReader) SetStackProp(name string, val Stack) error {
	propType, ok := g.types[name]

	if ok && propType != TypeStack {
		return errors.New("That property was already set but was not a stack")
	}

	g.types[name] = TypeStack
	g.values[name] = val

	return nil
}

func (g *genericReader) SetTimerProp(name string, val Timer) error {
	propType, ok := g.types[name]

	if ok && propType != TypeTimer {
		return errors.New("That property was already set but was not a timer")
	}

	g.types[name] = TypeTimer
	g.values[name] = val

	return nil
}
