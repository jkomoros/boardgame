/*

	memory is a storage manager that just keeps the games and storage in
	memory, which means that when the program exits the storage evaporates.
	Useful in cases where you don't want a persistent store (e.g. testing or
	fast iteration). Implements both boardgame.StorageManager and
	boardgame/server.StorageManager.

*/
package memory

import (
	"encoding/json"
	"errors"
	"github.com/jkomoros/boardgame"
	"strconv"
	"strings"
)

type memoryStateRecord struct {
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

type StorageManager struct {
	states map[string]map[int]*memoryStateRecord
	games  map[string]*memoryGameRecord
}

func NewStorageManager() *StorageManager {
	//InMemoryStorageManager is an extremely simple StorageManager that just keeps
	//track of the objects in memory.
	return &StorageManager{
		states: make(map[string]map[int]*memoryStateRecord),
		games:  make(map[string]*memoryGameRecord),
	}
}

func (s *StorageManager) State(game *boardgame.Game, version int) (*boardgame.State, error) {
	if game == nil {
		return nil, errors.New("No game provided")
	}

	if version < 0 || version > game.Version() {
		return nil, errors.New("Invalid version")
	}

	versionMap, ok := s.states[game.Id()]

	if !ok {
		return nil, errors.New("No such game")
	}

	record, ok := versionMap[version]

	if !ok {
		return nil, errors.New("No such version for that game")
	}

	state, err := game.Manager().StateFromBlob(record.SerializedState)

	if err != nil {
		return nil, errors.New("StateForBlob failed" + err.Error())
	}

	return state, nil
}

func (s *StorageManager) Game(manager *boardgame.GameManager, id string) (*boardgame.Game, error) {
	record := s.games[id]

	if record == nil {
		return nil, errors.New("No such game")
	}

	if manager == nil {
		return nil, errors.New("No manager provided")
	}

	return manager.LoadGame(record.Name, id, record.Version, record.Finished, s.winnersFromStorage(record.Winners)), nil
}

func (s *StorageManager) winnersForStorage(winners []int) string {

	if winners == nil {
		return ""
	}

	result := make([]string, len(winners))

	for i, num := range winners {
		result[i] = strconv.Itoa(num)
	}

	return strings.Join(result, ",")
}

func (s *StorageManager) winnersFromStorage(winners string) []int {

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

func (s *StorageManager) SaveGameAndCurrentState(game *boardgame.Game, state *boardgame.State) error {
	if game == nil {
		return errors.New("No game provided")
	}

	if !game.Modifiable() {
		return errors.New("Game is not modifiable")
	}

	//TODO: validate that state.Version is reasonable.

	if _, ok := s.states[game.Id()]; !ok {
		s.states[game.Id()] = make(map[int]*memoryStateRecord)
	}

	version := game.Version()

	versionMap := s.states[game.Id()]

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
		SerializedState: blob,
	}

	s.games[game.Id()] = &memoryGameRecord{
		Version:  version,
		Winners:  s.winnersForStorage(game.Winners()),
		Finished: game.Finished(),
		Id:       game.Id(),
		Name:     game.Name(),
	}

	return nil
}

//ListGames will return game objects for up to max number of games
func (s *StorageManager) ListGames(managers boardgame.ManagerCollection, max int) []*boardgame.Game {

	var result []*boardgame.Game

	for _, game := range s.games {

		manager := managers.Get(game.Name)

		if manager == nil {
			//Hmm, guess it wasn't a manager we knew about.

			//TODO: log an error
			continue
		}

		result = append(result, manager.Game(game.Id))
		if len(result) >= max {
			break
		}
	}

	return result

}

func (s *StorageManager) Close() {
	//Don't need to do anything
}

func (s *StorageManager) CleanUp() {
	//Don't need to do
}
