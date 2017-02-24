package boardgame

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

//StatePayload is where the "meat" of the state goes. It is one object so that
//client games can cast it quickly to the concrete struct for their game, so
//that they can get to a type-checked world with minimal fuss inside of
//Move.Legal and move.Apply. Your underlying struct should have a Game and
//Players property, so they serialize properly to JSON. Most importantly,
//json.Marshal() should round trip through your GameDelegate.StateFromBlob()
//without modifications in order for persistence to work. Each PlayerState you
//return should be the same underlying type. This means that if different
//players have very different roles in a game, there might be many properties
//that are not in use for any given player.
type State interface {
	//Game includes the non-user state for the game.
	GameState() GameState
	//Users contains a PlayerState object for each user in the game.
	PlayerStates() []PlayerState
	//Copy returns a copy of the Payload. Be sure it's a deep copy that makes
	//a copy of any pointer arguments.
	Copy() State
	//Diagram should return a basic debug rendering of state in multi-line
	//ascii art. Useful for debugging.
	Diagram() string
	//TODO: it's annoying that we have to reimplement JSON() for every struct
	//even though there should just be generic. Move to a top-level Method.
}

//Property reader is a way to read out properties on an object with unknown
//shape.
type PropertyReader interface {
	//Props returns a list of all property names that are defined for this
	//object.
	Props() map[string]PropertyType
	//Prop returns the value for that property.
	Prop(name string) interface{}
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
	TypeGrowableStack
	TypeSizedStack
)

//Property read setter is a way to enumerate and set properties on an object with an
//unknown shape.
type PropertyReadSetter interface {
	//All PropertyReadSetters have read interfaces
	PropertyReader
	//SetProp sets the property with the given name. If the value does not
	//match the underlying slot type, it should return an error.
	SetProp(name string, value interface{}) error
}

//BaseState is the interface that all state objects--UserStates and GameStates
//--implement.
type BaseState interface {
	Reader() PropertyReader
}

//PlayerState represents the state of a game associated with a specific user.
type PlayerState interface {
	//PlayerIndex encodes the index this user's state is in the containing
	//state object.
	PlayerIndex() int
	//Copy produces a copy of our current state. Be sure it's a deep copy that
	//makes a copy of any pointer arguments.
	Copy() PlayerState
	BaseState
}

//GameState represents the state of a game that is not associated with a
//particular user. For example, the draw stack of cards, who the current
//player is, and other properites.
type GameState interface {
	//Copy returns a copy of our current state. Be sure it's a deep copy that
	//makes a copy of any pointer arguments.
	Copy() GameState
	BaseState
}

//TODO: protect access to this with a mutex.
var defaultReaderCache map[interface{}]*defaultReader

func init() {
	defaultReaderCache = make(map[interface{}]*defaultReader)
}

type defaultReader struct {
	i     interface{}
	props map[string]PropertyType
}

//NewDefaultReader returns an object that satisfies the PropertyReader
//interface for the given concrete object, using reflection. Make it easy to
//implement the Reader method in a line. It will return an existing wrapper or
//create a new one if necessary.
func NewDefaultReader(i interface{}) PropertyReader {
	return NewDefaultReadSetter(i)
}

//NewDefaultReadSetter returns an object that satisfies the PropertyReadSetter
//interface for the given concrete object, using reflection. Make it easy to
//implement the Reader method in a line. It will return an existing wrapper or
//create a new one if necessary.
func NewDefaultReadSetter(i interface{}) PropertyReadSetter {
	if reader := defaultReaderCache[i]; reader != nil {
		return reader
	}
	result := &defaultReader{
		i: i,
	}
	defaultReaderCache[i] = result
	return result
}

//DefaultMarshalJSON is a simple wrapper around json.MarshalIndent, with the
//right defaults set. If your structs need to implement MarshaLJSON to output
//JSON, use this to encode it.
func DefaultMarshalJSON(obj interface{}) ([]byte, error) {
	return json.MarshalIndent(obj, "", "  ")
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
				pType = TypeInt
			case reflect.String:
				pType = TypeString
			case reflect.Ptr:
				//Is it a growable stack or a sizedStack?
				ptrType := field.Elem().Type().String()

				if strings.Contains(ptrType, "GrowableStack") {
					pType = TypeGrowableStack
				} else if strings.Contains(ptrType, "SizedStack") {
					pType = TypeSizedStack
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

func (d *defaultReader) Prop(name string) interface{} {

	if !propertyReaderImplNameShouldBeIncluded(name) {
		return nil
	}

	obj := d.i

	s := reflect.ValueOf(obj).Elem()
	return s.FieldByName(name).Interface()
}

func (d *defaultReader) SetProp(name string, val interface{}) (err error) {

	obj := d.i

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
