/*

Package memory is a storage manager that just keeps the games and storage in
memory, which means that when the program exits the storage evaporates.
Useful in cases where you don't want a persistent store (e.g. testing or
fast iteration). Implements both boardgame.StorageManager and
boardgame/server.StorageManager.

*/
package memory

import (
	"errors"
	"sync"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/jkomoros/boardgame/storage/internal/helpers"
)

//StorageManager is the primary type of this package. Get a new one with
//NewStorageManager.
type StorageManager struct {
	states map[string]map[int]boardgame.StateStorageRecord
	moves  map[string]map[int]*boardgame.MoveStorageRecord
	games  map[string]*boardgame.GameStorageRecord

	statesLock sync.RWMutex
	movesLock  sync.RWMutex
	gamesLock  sync.RWMutex

	*helpers.ExtendedMemoryStorageManager
}

//NewStorageManager is the way to get a new StorageManager.
func NewStorageManager() *StorageManager {
	//InMemoryStorageManager is an extremely simple StorageManager that just keeps
	//track of the objects in memory.
	result := &StorageManager{
		states: make(map[string]map[int]boardgame.StateStorageRecord),
		moves:  make(map[string]map[int]*boardgame.MoveStorageRecord),
		games:  make(map[string]*boardgame.GameStorageRecord),
	}
	result.ExtendedMemoryStorageManager = helpers.NewExtendedMemoryStorageManager(result)
	return result
}

//Name returns 'memory'
func (s *StorageManager) Name() string {
	return "memory"
}

//State implements that part of the core storage interface
func (s *StorageManager) State(gameID string, version int) (boardgame.StateStorageRecord, error) {
	if gameID == "" {
		return nil, errors.New("No game provided")
	}

	if version < 0 {
		return nil, errors.New("Invalid version")
	}

	s.statesLock.RLock()

	versionMap, ok := s.states[gameID]

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

//Moves implements that part of the core storage interface
func (s *StorageManager) Moves(gameID string, fromVersion, toVersion int) ([]*boardgame.MoveStorageRecord, error) {
	return helpers.MovesHelper(s, gameID, fromVersion, toVersion)
}

//Move implements that part of the core storage interface
func (s *StorageManager) Move(gameID string, version int) (*boardgame.MoveStorageRecord, error) {
	if gameID == "" {
		return nil, errors.New("No game provided")
	}

	if version < 0 {
		return nil, errors.New("Invalid version")
	}

	s.movesLock.RLock()

	versionMap, ok := s.moves[gameID]

	s.movesLock.RUnlock()

	if !ok {
		return nil, errors.New("No such game")
	}
	s.movesLock.RLock()
	record, ok := versionMap[version]
	s.movesLock.RUnlock()

	if !ok {
		return nil, errors.New("No such version for that game")
	}

	return record, nil

}

//Game implements that part of the core storage interface
func (s *StorageManager) Game(id string) (*boardgame.GameStorageRecord, error) {

	s.gamesLock.RLock()
	record := s.games[id]
	s.gamesLock.RUnlock()

	if record == nil {
		return nil, errors.New("No such game")
	}

	return record, nil
}

//SaveGameAndCurrentState implements that part of the core storage interface
func (s *StorageManager) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {
	if game == nil {
		return errors.New("No game provided")
	}

	s.statesLock.RLock()
	_, ok := s.states[game.ID]
	s.statesLock.RUnlock()
	if !ok {
		s.statesLock.Lock()
		s.states[game.ID] = make(map[int]boardgame.StateStorageRecord)
		s.statesLock.Unlock()
	}

	s.movesLock.RLock()
	_, ok = s.moves[game.ID]
	s.movesLock.RUnlock()
	if !ok {
		s.movesLock.Lock()
		s.moves[game.ID] = make(map[int]*boardgame.MoveStorageRecord)
		s.movesLock.Unlock()
	}

	version := game.Version

	s.statesLock.RLock()
	versionMap := s.states[game.ID]
	_, ok = versionMap[version]
	s.statesLock.RUnlock()

	if ok {
		//Wait, there was already a version stored there?
		return errors.New("There was already a version for that game stored")
	}

	s.movesLock.RLock()
	moveMap := s.moves[game.ID]
	_, ok = moveMap[version]
	s.movesLock.RUnlock()

	if ok {
		//Wait, there was already a version stored there?
		return errors.New("There was already a version for that game stored")
	}

	s.statesLock.Lock()
	versionMap[version] = state
	s.statesLock.Unlock()

	s.movesLock.Lock()
	if move != nil {
		moveMap[version] = move
	}
	s.movesLock.Unlock()

	s.gamesLock.Lock()
	s.games[game.ID] = game
	s.gamesLock.Unlock()

	return nil
}

//AllGames implements the extra method that storage/internal/helpers needs.
func (s *StorageManager) AllGames() []*boardgame.GameStorageRecord {
	var result []*boardgame.GameStorageRecord

	s.gamesLock.RLock()
	for _, game := range s.games {
		result = append(result, game)
	}
	s.gamesLock.RUnlock()

	return result
}

//ListGames will return game objects for up to max number of games
func (s *StorageManager) ListGames(max int, list listing.Type, userID string, gameType string) []*extendedgame.CombinedStorageRecord {
	return helpers.ListGamesHelper(s, max, list, userID, gameType)
}
