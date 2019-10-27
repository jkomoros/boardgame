package boardgame

import (
	"errors"
	"strconv"
	"strings"
)

//This file is actually used to just implement a shim StorageManager so our
//own package tests can run. The shim is basically storage/memory/StorageManager.

type testStorageManager struct {
	states map[string]map[int]StateStorageRecord
	moves  map[string]map[int]*MoveStorageRecord
	games  map[string]*GameStorageRecord
}

func newTestStorageManager() *testStorageManager {
	//InMemoryStorageManager is an extremely simple StorageManager that just keeps
	//track of the objects in memory.
	return &testStorageManager{
		states: make(map[string]map[int]StateStorageRecord),
		moves:  make(map[string]map[int]*MoveStorageRecord),
		games:  make(map[string]*GameStorageRecord),
	}
}

func (t *testStorageManager) String() string {
	var results []string

	results = append(results, "States")

	for key, states := range t.states {
		results = append(results, key)
		for version, state := range states {
			results = append(results, strconv.Itoa(version)+": "+string(state))
		}
	}

	results = append(results, "Games")

	for key, game := range t.games {
		results = append(results, key, game.Name, game.ID, strconv.Itoa(game.Version))
	}

	return strings.Join(results, "\n")
}

func (t *testStorageManager) State(gameID string, version int) (StateStorageRecord, error) {
	if gameID == "" {
		return nil, errors.New("No game provided")
	}

	if version < 0 {
		return nil, errors.New("Illegal version")
	}

	versionMap, ok := t.states[gameID]

	if !ok {
		return nil, errors.New("That game does not exist")
	}

	record, ok := versionMap[version]

	if !ok {
		return nil, errors.New("That version of that game doesn't exist")
	}

	return record, nil
}

func (t *testStorageManager) Moves(gameID string, fromVersion, toVersion int) ([]*MoveStorageRecord, error) {
	result := make([]*MoveStorageRecord, toVersion-fromVersion+1)

	index := 0
	for i := fromVersion; i <= toVersion; i++ {
		move, err := t.Move(gameID, i)
		if err != nil {
			return nil, err
		}
		result[index] = move
		index++
	}
	return result, nil
}

func (t *testStorageManager) Move(gameID string, version int) (*MoveStorageRecord, error) {
	if gameID == "" {
		return nil, errors.New("No game provided")
	}

	if version < 0 {
		return nil, errors.New("Illegal version")
	}

	versionMap, ok := t.moves[gameID]

	if !ok {
		return nil, errors.New("That game does not exist")
	}

	record, ok := versionMap[version]

	if !ok {
		return nil, errors.New("That version of that game doesn't exist")
	}

	return record, nil
}

func (t *testStorageManager) Game(id string) (*GameStorageRecord, error) {
	record := t.games[id]

	if record == nil {
		return nil, errors.New("That game does not exist")
	}

	return record, nil

}

func (t *testStorageManager) SaveGameAndCurrentState(game *GameStorageRecord, state StateStorageRecord, move *MoveStorageRecord) error {
	if game == nil {
		return errors.New("No game provided")
	}

	//TODO: validate that state.Version is reasonable.

	if _, ok := t.states[game.ID]; !ok {
		t.states[game.ID] = make(map[int]StateStorageRecord)
	}

	if _, ok := t.moves[game.ID]; !ok {
		t.moves[game.ID] = make(map[int]*MoveStorageRecord)
	}

	version := game.Version

	versionMap := t.states[game.ID]
	moveMap := t.moves[game.ID]

	if _, ok := versionMap[version]; ok {
		//Wait, there was already a version stored there?
		return errors.New("There was already a version for that game stored")
	}

	if _, ok := moveMap[version]; ok {
		//Wait, there was already a version stored there?
		return errors.New("There was already a version for that game stored")
	}

	versionMap[version] = state
	moveMap[version] = move

	t.games[game.ID] = game

	return nil
}

func (t *testStorageManager) PlayerMoveApplied(game *GameStorageRecord) error {
	//Pass
	return nil
}

func (t *testStorageManager) AgentState(gameID string, player PlayerIndex) ([]byte, error) {
	//TODO: implement
	return nil, nil
}

func (t *testStorageManager) SaveAgentState(gameID string, player PlayerIndex, state []byte) error {
	//TODO: implement
	return nil
}
