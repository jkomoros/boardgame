package boardgame

import (
	"errors"
	"fmt"
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
	PlayerIndexProp(name string) (PlayerIndex, error)
	GrowableStackProp(name string) (*GrowableStack, error)
	SizedStackProp(name string) (*SizedStack, error)
	TimerProp(name string) (*Timer, error)
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
	TypeGrowableStack
	TypeSizedStack
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
	SetGrowableStackProp(name string, value *GrowableStack) error
	SetSizedStackProp(name string, value *SizedStack) error
	SetTimerProp(name string, value *Timer) error
	//SetProp sets the property with the given name. If the value does not
	//match the underlying slot type, it should return an error. If you know
	//the underlying type it's always better to use the typed accessors.
	SetProp(name string, value interface{}) error
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

//DefaultReader returns an object that satisfies the PropertyReader
//interface for the given concrete object, using reflection. Make it easy to
//implement the Reader method in a line. It will return an existing wrapper or
//create a new one if necessary.
func DefaultReader(i interface{}) PropertyReader {
	return DefaultReadSetter(i)
}

//DefaultReadSetter returns an object that satisfies the PropertyReadSetter
//interface for the given concrete object, using reflection. Make it easy to
//implement the Reader method in a line. It will return an existing wrapper or
//create a new one if necessary.
func DefaultReadSetter(i interface{}) PropertyReadSetter {

	defaultReaderCacheLock.RLock()
	reader := defaultReaderCache[i]
	defaultReaderCacheLock.RUnlock()

	if reader != nil {
		return reader
	}

	result := &defaultReader{
		i: i,
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

	//TODO: skip fields that have a propertyreader:omit

	if d.props == nil {

		obj := d.i

		result := make(map[string]PropertyType)

		s := reflect.ValueOf(obj).Elem()
		typeOfObj := s.Type()

		for i := 0; i < s.NumField(); i++ {
			name := typeOfObj.Field(i).Name

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
			case reflect.Ptr:
				//Is it a growable stack or a sizedStack?
				ptrType := field.Type().String()

				if strings.Contains(ptrType, "GrowableStack") {
					pType = TypeGrowableStack
				} else if strings.Contains(ptrType, "SizedStack") {
					pType = TypeSizedStack
				} else if strings.Contains(ptrType, "Timer") {
					pType = TypeTimer
				} else {
					panic("Unknown ptr type:" + ptrType)
				}
			default:
				panic("Unsupported field in underlying type" + strconv.Itoa(int(field.Type().Kind())))
			}

			result[name] = pType

		}
		d.props = result
	}

	return d.props
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

func (d *defaultReader) GrowableStackProp(name string) (*GrowableStack, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypeGrowableStack {
		return nil, errors.New("That property is not a growable stack: " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	result := s.FieldByName(name).Interface().(*GrowableStack)

	return result, nil
}

func (d *defaultReader) SizedStackProp(name string) (*SizedStack, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypeSizedStack {
		return nil, errors.New("That property is not a sized stack: " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	result := s.FieldByName(name).Interface().(*SizedStack)

	return result, nil
}

func (d *defaultReader) TimerProp(name string) (*Timer, error) {
	//Verify that this seems legal.
	props := d.Props()

	if props[name] != TypeTimer {
		return nil, errors.New("That property is not a timer " + name)
	}

	s := reflect.ValueOf(d.i).Elem()
	result := s.FieldByName(name).Interface().(*Timer)

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

func (d *defaultReader) SetGrowableStackProp(name string, val *GrowableStack) (err error) {
	props := d.Props()

	if props[name] != TypeGrowableStack {
		return errors.New("That property is not a settable growable stack")
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

func (d *defaultReader) SetSizedStackProp(name string, val *SizedStack) (err error) {
	props := d.Props()

	if props[name] != TypeSizedStack {
		return errors.New("That property is not a settable sized stack")
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

func (d *defaultReader) SetTimerProp(name string, val *Timer) (err error) {
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

func (g *genericReader) GrowableStackProp(name string) (*GrowableStack, error) {
	val, err := g.Prop(name)

	if err != nil {
		return nil, err
	}

	propType, ok := g.types[name]

	if !ok {
		return nil, errors.New("Unexpected error: Missing Prop type for " + name)
	}

	if propType != TypeGrowableStack {
		return nil, errors.New(name + "was expected to be TypeGrowableStack but was not")
	}

	return val.(*GrowableStack), nil
}

func (g *genericReader) SizedStackProp(name string) (*SizedStack, error) {
	val, err := g.Prop(name)

	if err != nil {
		return nil, err
	}

	propType, ok := g.types[name]

	if !ok {
		return nil, errors.New("Unexpected error: Missing Prop type for " + name)
	}

	if propType != TypeSizedStack {
		return nil, errors.New(name + "was expected to be TypeSizedStack but was not")
	}

	return val.(*SizedStack), nil
}

func (g *genericReader) TimerProp(name string) (*Timer, error) {
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

	return val.(*Timer), nil
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

func (g *genericReader) SetGrowableStackProp(name string, val *GrowableStack) error {
	propType, ok := g.types[name]

	if ok && propType != TypeGrowableStack {
		return errors.New("That property was already set but was not an growable stack")
	}

	g.types[name] = TypeGrowableStack
	g.values[name] = val

	return nil
}

func (g *genericReader) SetSizedStackProp(name string, val *SizedStack) error {
	propType, ok := g.types[name]

	if ok && propType != TypeSizedStack {
		return errors.New("That property was already set but was not an sized stack")
	}

	g.types[name] = TypeSizedStack
	g.values[name] = val

	return nil
}

func (g *genericReader) SetTimerProp(name string, val *Timer) error {
	propType, ok := g.types[name]

	if ok && propType != TypeTimer {
		return errors.New("That property was already set but was not a timer")
	}

	g.types[name] = TypeTimer
	g.values[name] = val

	return nil
}
