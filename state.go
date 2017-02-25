package boardgame

import (
	"encoding/json"
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

//DefaultMarshalJSON is a simple wrapper around json.MarshalIndent, with the
//right defaults set. If your structs need to implement MarshaLJSON to output
//JSON, use this to encode it.
func DefaultMarshalJSON(obj interface{}) ([]byte, error) {
	return json.MarshalIndent(obj, "", "  ")
}
