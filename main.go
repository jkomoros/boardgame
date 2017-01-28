package boardgame

type State struct {
	//The version number of the state. Increments by one each time a Move is
	//applied.
	Version int
	//The schema version that this state object uses. This number will not
	//change often, but is useful to detect if the state was saved back when a
	//diferent schema was in use and needs to be migrated.
	Schema int
	//Game includes the non-user state for the game.
	Game GameState
	//Users contains a UserState object for each user in the game.
	Users []UserState
}

type JSONer interface {
	//Returns the canonical JSON representation of this object, suitable to
	//being communicated across the wire or saved in a DB.
	JSON() []byte
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

//UserState represents the state of a game associated with a specific user.
type UserState interface {
	//PlayerIndex encodes the index this user's state is in the containing
	//state object.
	PlayerIndex() int
	JSONer
	PropertyReader
}

//GameState represents the state of a game that is not associated with a
//particular user. For example, the draw stack of cards, who the current
//player is, and other properites.
type GameState interface {
	JSONer
	PropertyReader
}
