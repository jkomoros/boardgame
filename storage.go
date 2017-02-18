package boardgame

//StorageManager is an interface that anything can implement to handle the
//persistence of Games and States.
type StorageManager interface {
	//State returns the StateWrapper for the game at the given version, or
	//nil.
	State(game *Game, version int) State

	//Game fetches the game with the given ID from the store, if it exists.
	//The implementation should use manager.LoadGame to get a real game object
	//that is ready for use. The returned game will always have Modifiable()
	//as false. If you want a modifiable version, use
	//GameManager.ModifiableGame(id).
	Game(manager *GameManager, id string) *Game

	//SaveGameAndCurrentState stores the game and the current state (at
	//game.Version()) into the store at the same time in a transaction. If
	//Game.Modifiable() is false, storage should fail.
	SaveGameAndCurrentState(game *Game, state State) error
}
