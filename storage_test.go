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

func (s *testStorageManager) String() string {
	var results []string

	results = append(results, "States")

	for key, states := range s.states {
		results = append(results, key)
		for version, state := range states {
			results = append(results, strconv.Itoa(version)+": "+string(state))
		}
	}

	results = append(results, "Games")

	for key, game := range s.games {
		results = append(results, key, game.Name, game.Id, strconv.Itoa(game.Version))
	}

	return strings.Join(results, "\n")
}

func (i *testStorageManager) State(gameId string, version int) (StateStorageRecord, error) {
	if gameId == "" {
		return nil, errors.New("No game provided")
	}

	if version < 0 {
		return nil, errors.New("Illegal version")
	}

	versionMap, ok := i.states[gameId]

	if !ok {
		return nil, errors.New("That game does not exist")
	}

	record, ok := versionMap[version]

	if !ok {
		return nil, errors.New("That version of that game doesn't exist")
	}

	return record, nil
}

func (i *testStorageManager) Move(gameId string, version int) (*MoveStorageRecord, error) {
	if gameId == "" {
		return nil, errors.New("No game provided")
	}

	if version < 0 {
		return nil, errors.New("Illegal version")
	}

	versionMap, ok := i.moves[gameId]

	if !ok {
		return nil, errors.New("That game does not exist")
	}

	record, ok := versionMap[version]

	if !ok {
		return nil, errors.New("That version of that game doesn't exist")
	}

	return record, nil
}

func (i *testStorageManager) Game(id string) (*GameStorageRecord, error) {
	record := i.games[id]

	if record == nil {
		return nil, errors.New("That game does not exist")
	}

	return record, nil

}

func (i *testStorageManager) SaveGameAndCurrentState(game *GameStorageRecord, state StateStorageRecord, move *MoveStorageRecord) error {
	if game == nil {
		return errors.New("No game provided")
	}

	//TODO: validate that state.Version is reasonable.

	if _, ok := i.states[game.Id]; !ok {
		i.states[game.Id] = make(map[int]StateStorageRecord)
	}

	if _, ok := i.moves[game.Id]; !ok {
		i.moves[game.Id] = make(map[int]*MoveStorageRecord)
	}

	version := game.Version

	versionMap := i.states[game.Id]
	moveMap := i.moves[game.Id]

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

	i.games[game.Id] = game

	return nil
}

func (i *testStorageManager) PlayerMoveApplied(game *GameStorageRecord) error {
	//Pass
	return nil
}

func (i *testStorageManager) AgentState(gameId string, player PlayerIndex) ([]byte, error) {
	//TODO: implement
	return nil, nil
}

func (i *testStorageManager) SaveAgentState(gameId string, player PlayerIndex, state []byte) error {
	//TODO: implement
	return nil
}
