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
	"github.com/jkomoros/boardgame/server/api/users"
	"sync"
)

type StorageManager struct {
	states            map[string]map[int]boardgame.StateStorageRecord
	games             map[string]*boardgame.GameStorageRecord
	usersById         map[string]*users.StorageRecord
	usersByCookie     map[string]*users.StorageRecord
	usersForGames     map[string][]string
	statesLock        sync.RWMutex
	gamesLock         sync.RWMutex
	usersLock         sync.RWMutex
	usersForGamesLock sync.RWMutex
}

func NewStorageManager() *StorageManager {
	//InMemoryStorageManager is an extremely simple StorageManager that just keeps
	//track of the objects in memory.
	return &StorageManager{
		states:        make(map[string]map[int]boardgame.StateStorageRecord),
		games:         make(map[string]*boardgame.GameStorageRecord),
		usersById:     make(map[string]*users.StorageRecord),
		usersByCookie: make(map[string]*users.StorageRecord),
		usersForGames: make(map[string][]string),
	}
}

func (s *StorageManager) Name() string {
	return "memory"
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
func (s *StorageManager) ListGames(max int) []*boardgame.GameStorageRecord {

	var result []*boardgame.GameStorageRecord

	for _, game := range s.games {

		result = append(result, game)
		if len(result) >= max {
			break
		}
	}

	return result

}

func (s *StorageManager) UserIdsForGame(gameId string) []string {
	s.usersForGamesLock.RLock()
	ids := s.usersForGames[gameId]
	s.usersForGamesLock.RUnlock()

	if ids == nil {
		game, _ := s.Game(gameId)
		if game == nil {
			return nil
		}
		return make([]string, game.NumPlayers)
	}

	return ids
}

func (s *StorageManager) SetPlayerForGame(gameId string, playerIndex boardgame.PlayerIndex, userId string) error {
	ids := s.UserIdsForGame(gameId)

	if int(playerIndex) < 0 || int(playerIndex) >= len(ids) {
		return errors.New("PlayerIndex " + playerIndex.String() + " is not valid for this game.")
	}

	if ids[playerIndex] != "" {
		return errors.New("PlayerIndex " + playerIndex.String() + " is already taken.")
	}

	user := s.GetUserById(userId)

	if user == nil {
		return errors.New("That uid does not describe an existing user")
	}

	ids[playerIndex] = userId

	s.usersForGamesLock.Lock()
	s.usersForGames[gameId] = ids
	s.usersForGamesLock.Unlock()

	return nil
}

//Store or update all fields
func (s *StorageManager) UpdateUser(user *users.StorageRecord) error {

	s.usersLock.Lock()
	s.usersById[user.Id] = user
	s.usersLock.Unlock()

	return nil

}

func (s *StorageManager) GetUserById(uid string) *users.StorageRecord {
	s.usersLock.RLock()
	user := s.usersById[uid]
	s.usersLock.RUnlock()

	return user
}

func (s *StorageManager) GetUserByCookie(cookie string) *users.StorageRecord {
	s.usersLock.RLock()
	user := s.usersByCookie[cookie]
	s.usersLock.RUnlock()

	return user
}

//If user is nil, the cookie should be deleted if it exists. If the user
//does not yet exist, it should be added to the database.
func (s *StorageManager) ConnectCookieToUser(cookie string, user *users.StorageRecord) error {
	if user == nil {
		s.usersLock.Lock()
		delete(s.usersByCookie, cookie)
		s.usersLock.Unlock()
		return nil
	}

	otherUser := s.GetUserById(user.Id)

	if otherUser == nil {
		s.UpdateUser(user)
	}

	s.usersLock.Lock()
	s.usersByCookie[cookie] = user
	s.usersLock.Unlock()

	return nil
}

func (s *StorageManager) Close() {
	//Don't need to do anything
}

func (s *StorageManager) CleanUp() {
	//Don't need to do
}
