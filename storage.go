package boardgame

import (
	"time"
)

//StateStorageRecord is a record representing a state that can be written to
//storage and later returned. It is an opaque blob, so in most cases storage
//managers can just write straight to disk with no transformations.
type StateStorageRecord []byte

//MoveStorageRecord is a record representing the Move that was made to get the
//game to its most recent version.
type MoveStorageRecord struct {
	Name      string
	Version   int
	Initiator int
	Timestamp time.Time
	Blob      []byte
}

//GameStorageRecord is a simple struct with public fields representing the
//important aspects of a game that should be serialized to storage. The fields
//are broken out specifically so that the storage layer can understand these
//properties in queries.
type GameStorageRecord struct {
	//Name is the type of the game, from its manager. Used for sanity
	//checking.
	Name string
	Id   string
	//SecretSalt for this game for things like component Ids. Should never be
	//transmitted to an insecure or untrusted environment.
	SecretSalt string `json:",omitempty"`
	Version    int
	Winners    []PlayerIndex
	Finished   bool
	Created    time.Time
	//NumPlayers is the reported number of players when it was created.
	//Primarily for convenience to storage layer so they know how many players
	//are in the game.
	NumPlayers int
	Agents     []string
}

//StorageManager is an interface that anything can implement to handle the
//persistence of Games and States.
type StorageManager interface {
	//State returns the StateWrapper for the game at the given version, or
	//nil.
	State(gameId string, version int) (StateStorageRecord, error)

	//Move returns the MoveStorageRecord for the game at the given version, or
	//nil.
	Move(gameId string, version int) (*MoveStorageRecord, error)

	//Moves is like Move but returns all moves from fromVersion to toVersion,
	//inclusive. In many storage subsystems this is cheaper than repeated
	//calls to Move.
	Moves(gameId string, fromVersion, toVersion int) ([]*MoveStorageRecord, error)

	//Game fetches the game with the given ID from the store, if it exists.
	Game(id string) (*GameStorageRecord, error)

	//AgentState retrieves the most recent state for the given agent
	AgentState(gameId string, player PlayerIndex) ([]byte, error)

	//SaveGameAndCurrentState stores the game and the current state (at
	//game.Version()) into the store at the same time in a transaction. If
	//Game.Modifiable() is false, storage should fail. Move can be nil (if game.Version() is 0)
	SaveGameAndCurrentState(game *GameStorageRecord, state StateStorageRecord, move *MoveStorageRecord) error

	//SaveAgentState saves the agent state for the given player
	SaveAgentState(gameId string, player PlayerIndex, state []byte) error

	//PlayerMoveApplied is called after a PlayerMove and all of its resulting
	//FixUp moves have been applied. Most StorageManagers don't need to do
	//anything here; it's primarily useful for signaling that a run of moves
	//has been applied, e.g. in the server.
	PlayerMoveApplied(game *GameStorageRecord) error
}
