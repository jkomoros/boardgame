package boardgame

//StorageManager is an interface that anything can implement to handle the
//persistence of Games and States.
type StorageManager interface {
	//State returns the StateWrapper for the game at the given version, or
	//nil.
	State(game *Game, version int) *StateWrapper
	//SaveState puts the given stateWrapper (at the specified version and
	//schema) into storage.
	SaveState(game *Game, state *StateWrapper) error
}
