package boardgame

//StateStorageRecord is a record representing a state that can be written to
//storage and later returned. It is an opaque blob, so in most cases storage
//managers can just write straight to disk with no transformations.
type StateStorageRecord []byte

//GameStorageRecord is a simple struct with public fields representing the
//important aspects of a game that should be serialized to storage. The fields
//are broken out specifically so that the storage layer can understand these
//properties in queries.
type GameStorageRecord struct {
	//Name is the type of the game, from its manager. Used for sanity
	//checking.
	Name     string
	Id       string
	Version  int
	Winners  []int
	Finished bool
}

//StorageManager is an interface that anything can implement to handle the
//persistence of Games and States.
type StorageManager interface {
	//State returns the StateWrapper for the game at the given version, or
	//nil.
	State(gameId string, version int) (StateStorageRecord, error)

	//Game fetches the game with the given ID from the store, if it exists.
	Game(id string) (*GameStorageRecord, error)

	//SaveGameAndCurrentState stores the game and the current state (at
	//game.Version()) into the store at the same time in a transaction. If
	//Game.Modifiable() is false, storage should fail.
	SaveGameAndCurrentState(game *GameStorageRecord, state StateStorageRecord) error
}
