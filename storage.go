package boardgame

//StateStorageRecord is a record representing a state that can be written to
//storage and later returned. It is an opaque blob, so in most cases storage
//managers can just write straight to disk with no transformations.
type StateStorageRecord []byte

//StorageManager is an interface that anything can implement to handle the
//persistence of Games and States.
type StorageManager interface {
	//State returns the StateWrapper for the game at the given version, or
	//nil.
	State(gameId string, version int) (StateStorageRecord, error)

	//Game fetches the game with the given ID from the store, if it exists.
	//The implementation should use manager.LoadGame to get a real game object
	//that is ready for use. The returned game will always have Modifiable()
	//as false. If you want a modifiable version, use
	//GameManager.ModifiableGame(id).
	Game(manager *GameManager, id string) (*Game, error)

	//SaveGameAndCurrentState stores the game and the current state (at
	//game.Version()) into the store at the same time in a transaction. If
	//Game.Modifiable() is false, storage should fail.
	SaveGameAndCurrentState(game *Game, state StateStorageRecord) error
}
