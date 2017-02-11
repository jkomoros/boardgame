package boardgame

import (
	"errors"
	"fmt"
	"reflect"
	"unicode"
)

type StateWrapper struct {
	//The version number of the state. Increments by one each time a Move is
	//applied.
	Version int
	//The schema version that this state object uses. This number will not
	//change often, but is useful to detect if the state was saved back when a
	//diferent schema was in use and needs to be migrated.
	Schema int
	State  State
}

//newStarterState creates a new state initialized for the first move.
func newStarterStateWrapper(state State) *StateWrapper {
	return &StateWrapper{
		Version: 0,
		Schema:  0,
		State:   state,
	}
}

//StatePayload is where the "meat" of the state goes. It is one object so that
//client games can cast it quickly to the concrete struct for their game, so
//that they can get to a type-checked world with minimal fuss inside of
//Move.Legal and move.Apply.
type State interface {
	//Game includes the non-user state for the game.
	Game() GameState
	//Users contains a UserState object for each user in the game.
	Users() []UserState
	//Copy returns a copy of the Payload.
	Copy() State
	//Diagram should return a basic debug rendering of state in multi-line
	//ascii art. Useful for debugging.
	Diagram() string
	JSONer
	//TODO: it's annoying that we have to reimplement JSON() for every struct
	//even though there should just be generic. Move to a top-level Method.
}

//Property reader is a way to read out properties on an object with unknown
//shape.
type PropertyReader interface {
	//Props returns a list of all property names that are defined for this
	//object.
	Props() []string
	//Prop returns the value for that property.
	Prop(name string) interface{}
}

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
	JSONer
	PropertyReader
}

//UserState represents the state of a game associated with a specific user.
type UserState interface {
	//PlayerIndex encodes the index this user's state is in the containing
	//state object.
	PlayerIndex() int
	//Copy produces a copy of our current state
	Copy() UserState
	BaseState
}

//GameState represents the state of a game that is not associated with a
//particular user. For example, the draw stack of cards, who the current
//player is, and other properites.
type GameState interface {
	//Copy returns a copy of our current state
	Copy() GameState
	BaseState
}

//Copy prepares another version of State that is set exactly the same. This is
//done before a modification is made.
func (s *StateWrapper) Copy() *StateWrapper {
	//TODO: test this
	return &StateWrapper{
		Version: s.Version,
		Schema:  s.Schema,
		State:   s.State.Copy(),
	}

}

//JSON returns the JSONObject representing the State's full state.
func (s *StateWrapper) JSON() JSONObject {

	return JSONMap{
		"Version": s.Version,
		"Schema":  s.Schema,
		"State":   s.State.JSON(),
	}

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

//PropertyReaderPropsImpl is a helper method useful for satisfying the
//PropertyReader interface without writing finicky, bespoke code. It uses
//reflection to enumerate all of the properties. You'd use it as the single
//line of implementation in your struct's Props() implementation, passing in
//self, where self is the pointer receiver.
func PropertyReaderPropsImpl(obj interface{}) []string {

	//TODO: skip fields that have a propertyreader:omit

	s := reflect.ValueOf(obj).Elem()
	typeOfObj := s.Type()

	var result []string

	for i := 0; i < s.NumField(); i++ {
		name := typeOfObj.Field(i).Name

		if propertyReaderImplNameShouldBeIncluded(name) {
			result = append(result, name)
		}

	}

	return result
}

//PropertyReaderPropImpl is a helper method useful for satisfying the
//PropertyReader interface without writing finicky, bespoke code. It uses
//reflection to return the value of the named field or nil. You'd use it as
//the single line of implementation in your struct's Prop() implementation,
//passing in self, where self is the pointer receiver.
func PropertyReaderPropImpl(obj interface{}, name string) interface{} {

	if !propertyReaderImplNameShouldBeIncluded(name) {
		return nil
	}

	s := reflect.ValueOf(obj).Elem()
	return s.FieldByName(name).Interface()
}

//PropertySetImpl is a helper method useful for satisfying the
//PropertyReadSetter interface without writing finicky, bespoke code. It uses
//reflection to set the value. You'd use it as the single-line implementation
//of your struct's SetProp() implementation, passing in self, where self is
//the pointer receiver.
func PropertySetImpl(obj interface{}, name string, val interface{}) (err error) {

	//TODO: name this consistently with the other PropertyReader helpers.

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
