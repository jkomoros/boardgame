/*

	memory is a storage manager that just keeps the games and storage in
	memory, which means that when the program exits the storage evaporates.
	Useful in cases where you don't want a persistent store (e.g. testing or
	fast iteration). Implements both boardgame.StorageManager and
	boardgame/server.StorageManager.

*/
package memory

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"sync"
)

type StorageManager struct {
	states     map[string]map[int]boardgame.StateStorageRecord
	games      map[string]*boardgame.GameStorageRecord
	statesLock sync.RWMutex
	gamesLock  sync.RWMutex
}

func NewStorageManager() *StorageManager {
	//InMemoryStorageManager is an extremely simple StorageManager that just keeps
	//track of the objects in memory.
	return &StorageManager{
		states: make(map[string]map[int]boardgame.StateStorageRecord),
		games:  make(map[string]*boardgame.GameStorageRecord),
	}
}

func (s *StorageManager) State(gameId string, version int) (boardgame.StateStorageRecord, error) {
	if gameId == "" {
		return nil, errors.New("No game provided")
	}

	if version < 0 {
		return nil, errors.New("Invalid version")
	}

	s.statesLock.RLock()

	versionMap, ok := s.states[gameId]

	s.statesLock.RUnlock()

	if !ok {
		return nil, errors.New("No such game")
	}
	s.statesLock.RLock()
	record, ok := versionMap[version]
	s.statesLock.RUnlock()

	if !ok {
		return nil, errors.New("No such version for that game")
	}

	return record, nil

}

func (s *StorageManager) Game(id string) (*boardgame.GameStorageRecord, error) {

	s.gamesLock.RLock()
	record := s.games[id]
	s.gamesLock.RUnlock()

	if record == nil {
		return nil, errors.New("No such game")
	}

	return record, nil
}

func (s *StorageManager) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord) error {
	if game == nil {
		return errors.New("No game provided")
	}

	s.statesLock.RLock()
	_, ok := s.states[game.Id]
	s.statesLock.RUnlock()
	if !ok {
		s.statesLock.Lock()
		s.states[game.Id] = make(map[int]boardgame.StateStorageRecord)
		s.statesLock.Unlock()
	}

	version := game.Version

	s.statesLock.RLock()
	versionMap := s.states[game.Id]
	_, ok = versionMap[version]
	s.statesLock.RUnlock()

	if ok {
		//Wait, there was already a version stored there?
		return errors.New("There was already a version for that game stored")
	}

	s.statesLock.Lock()
	versionMap[version] = state
	s.statesLock.Unlock()

	s.gamesLock.Lock()
	s.games[game.Id] = game
	s.gamesLock.Unlock()

	return nil
}

//ListGames will return game objects for up to max number of games
func (s *StorageManager) ListGames(managers boardgame.ManagerCollection, max int) []*boardgame.GameStorageRecord {

	var result []*boardgame.GameStorageRecord

	for _, game := range s.games {

		manager := managers.Get(game.Name)

		if manager == nil {
			//Hmm, guess it wasn't a manager we knew about.

			//TODO: log an error
			continue
		}

		result = append(result, game)
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
