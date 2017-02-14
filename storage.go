package boardgame

import (
	"errors"
)

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

type memoryStateRecord struct {
	Schema          int
	Version         int
	SerializedState []byte
}

type inMemoryStorageManager struct {
	states map[string]map[int]*memoryStateRecord
}

func NewInMemoryStorageManager() StorageManager {
	//InMemoryStorageManager is an extremely simple StorageManager that just keeps
	//track of the objects in memory.
	return &inMemoryStorageManager{
		states: make(map[string]map[int]*memoryStateRecord),
	}
}

func (i *inMemoryStorageManager) State(game *Game, version int) *StateWrapper {
	if game == nil {
		return nil
	}

	if version < 0 || version > game.StateWrapper.Version {
		return nil
	}

	versionMap, ok := i.states[game.Id()]

	if !ok {
		return nil
	}

	record, ok := versionMap[version]

	if !ok {
		return nil
	}

	state, err := game.Delegate.StateFromBlob(record.SerializedState, record.Schema)

	if err != nil {
		return nil
	}

	return &StateWrapper{
		Version: record.Version,
		Schema:  record.Schema,
		State:   state,
	}
}

func (i *inMemoryStorageManager) SaveState(game *Game, state *StateWrapper) error {
	//TODO: do something
	return errors.New("That method is not yet implemented")
}
