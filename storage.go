package boardgame

import (
	"encoding/json"
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
	if game == nil {
		return errors.New("No game provided")
	}

	//TODO: validate that state.Version is reasonable.

	if _, ok := i.states[game.Id()]; !ok {
		i.states[game.Id()] = make(map[int]*memoryStateRecord)
	}

	versionMap := i.states[game.Id()]

	if _, ok := versionMap[state.Version]; ok {
		//Wait, there was already a version stored there?
		return errors.New("There was already a version for that game stored")
	}

	blob, err := json.Marshal(state.State)

	if err != nil {
		return errors.New("Error marshalling State: " + err.Error())
	}

	versionMap[state.Version] = &memoryStateRecord{
		Version:         state.Version,
		Schema:          state.Schema,
		SerializedState: blob,
	}

	return nil
}
