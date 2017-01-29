package boardgame

import (
	"reflect"
)

type State struct {
	//The version number of the state. Increments by one each time a Move is
	//applied.
	Version int
	//The schema version that this state object uses. This number will not
	//change often, but is useful to detect if the state was saved back when a
	//diferent schema was in use and needs to be migrated.
	Schema  int
	Payload StatePayload
}

//StatePayload is where the "meat" of the state goes. It is one object so that
//client games can cast it quickly to the concrete struct for their game, so
//that they can get to a type-checked world with minimal fuss inside of
//Move.Legal and move.Apply.
type StatePayload interface {
	//Game includes the non-user state for the game.
	Game() GameState
	//Users contains a UserState object for each user in the game.
	Users() []UserState
	//Copy returns a copy of the Payload.
	Copy() StatePayload
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
func (s *State) Copy() *State {
	//TODO: test this
	return &State{
		Version: s.Version,
		Schema:  s.Schema,
		Payload: s.Payload.Copy(),
	}

}

//JSON returns the JSONObject representing the State's full state.
func (s *State) JSON() JSONObject {

	return JSONMap{
		"Version": s.Version,
		"Schema":  s.Schema,
		"Payload": s.Payload.JSON(),
	}

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

	result := make([]string, s.NumField())

	for i := 0; i < s.NumField(); i++ {
		result[i] = typeOfObj.Field(i).Name
	}

	return result
}

//PropertyReaderPropImpl is a helper method useful for satisfying the
//PropertyReader interface without writing finicky, bespoke code. It uses
//reflection to return the value of the named field or nil. You'd use it as
//the single line of implementation in your struct's Prop() implementation,
//passing in self, where self is the pointer receiver.
func PropertyReaderPropImpl(obj interface{}, name string) interface{} {
	//TODO: skip fields that have a propertyreader:omit
	s := reflect.ValueOf(obj).Elem()
	return s.FieldByName(name).Interface()
}
