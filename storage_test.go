package boardgame

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

//This file is actually used to just implement a shim StorageManager so our
//own package tests can run. The shim is basically storage/memory/StorageManager.

type memoryStateRecord struct {
	Schema          int
	Version         int
	SerializedState []byte
}

type memoryGameRecord struct {
	Name     string
	Id       string
	Version  int
	Finished bool
	//We'll serialize as a string and then back out to simulate what a real DB
	//would do, and make sure we don't hand out the same string all of the
	//time.
	Winners string
}

type testStorageManager struct {
	states map[string]map[int]*memoryStateRecord
	games  map[string]*memoryGameRecord
}

func newTestStorageManager() *testStorageManager {
	//InMemoryStorageManager is an extremely simple StorageManager that just keeps
	//track of the objects in memory.
	return &testStorageManager{
		states: make(map[string]map[int]*memoryStateRecord),
		games:  make(map[string]*memoryGameRecord),
	}
}

func (i *testStorageManager) State(game *Game, version int) State {
	if game == nil {
		return nil
	}

	if version < 0 || version > game.version {
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

	state, err := game.manager.Delegate().StateFromBlob(record.SerializedState, record.Schema)

	if err != nil {
		return nil
	}

	return state
}

func (i *testStorageManager) Game(manager *GameManager, id string) *Game {
	record := i.games[id]

	if record == nil {
		return nil
	}

	if manager == nil {
		return nil
	}

	return manager.LoadGame(record.Name, id, record.Version, record.Finished, i.winnersFromStorage(record.Winners))
}

func (i *testStorageManager) winnersForStorage(winners []int) string {

	if winners == nil {
		return ""
	}

	result := make([]string, len(winners))

	for i, num := range winners {
		result[i] = strconv.Itoa(num)
	}

	return strings.Join(result, ",")
}

func (i *testStorageManager) winnersFromStorage(winners string) []int {

	if winners == "" {
		return nil
	}

	pieces := strings.Split(winners, ",")

	result := make([]int, len(pieces))

	for i, piece := range pieces {
		num, err := strconv.Atoi(piece)

		if err != nil {
			panic("Unexpected number stored:" + err.Error())
		}

		result[i] = num
	}
	return result
}

func (i *testStorageManager) SaveGameAndState(game *Game, version int, schema int, state State) error {
	if game == nil {
		return errors.New("No game provided")
	}

	if !game.Modifiable() {
		return errors.New("Game is not modifiable")
	}

	//TODO: validate that state.Version is reasonable.

	if _, ok := i.states[game.Id()]; !ok {
		i.states[game.Id()] = make(map[int]*memoryStateRecord)
	}

	versionMap := i.states[game.Id()]

	if _, ok := versionMap[version]; ok {
		//Wait, there was already a version stored there?
		return errors.New("There was already a version for that game stored")
	}

	blob, err := json.Marshal(state)

	if err != nil {
		return errors.New("Error marshalling State: " + err.Error())
	}

	versionMap[version] = &memoryStateRecord{
		Version:         version,
		Schema:          schema,
		SerializedState: blob,
	}

	i.games[game.Id()] = &memoryGameRecord{
		Version:  version,
		Winners:  i.winnersForStorage(game.Winners()),
		Finished: game.Finished(),
		Id:       game.Id(),
		Name:     game.Name(),
	}

	return nil
}
