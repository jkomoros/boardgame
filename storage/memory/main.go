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
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/jkomoros/boardgame/server/api/users"
	"sort"
	"sync"
	"time"
)

type StorageManager struct {
	states            map[string]map[int]boardgame.StateStorageRecord
	moves             map[string]map[int]*boardgame.MoveStorageRecord
	games             map[string]*boardgame.GameStorageRecord
	extendedGames     map[string]*extendedgame.StorageRecord
	usersById         map[string]*users.StorageRecord
	usersByCookie     map[string]*users.StorageRecord
	usersForGames     map[string][]string
	agentStates       map[string][]byte
	statesLock        sync.RWMutex
	movesLock         sync.RWMutex
	gamesLock         sync.RWMutex
	extendedGamesLock sync.RWMutex
	usersLock         sync.RWMutex
	usersForGamesLock sync.RWMutex
	agentStatesLock   sync.RWMutex
}

func NewStorageManager() *StorageManager {
	//InMemoryStorageManager is an extremely simple StorageManager that just keeps
	//track of the objects in memory.
	return &StorageManager{
		states:        make(map[string]map[int]boardgame.StateStorageRecord),
		moves:         make(map[string]map[int]*boardgame.MoveStorageRecord),
		games:         make(map[string]*boardgame.GameStorageRecord),
		extendedGames: make(map[string]*extendedgame.StorageRecord),
		usersById:     make(map[string]*users.StorageRecord),
		usersByCookie: make(map[string]*users.StorageRecord),
		usersForGames: make(map[string][]string),
		agentStates:   make(map[string][]byte),
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

func (s *StorageManager) Move(gameId string, version int) (*boardgame.MoveStorageRecord, error) {
	if gameId == "" {
		return nil, errors.New("No game provided")
	}

	if version < 0 {
		return nil, errors.New("Invalid version")
	}

	s.movesLock.RLock()

	versionMap, ok := s.moves[gameId]

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

func (s *StorageManager) Game(id string) (*boardgame.GameStorageRecord, error) {

	s.gamesLock.RLock()
	record := s.games[id]
	s.gamesLock.RUnlock()

	if record == nil {
		return nil, errors.New("No such game")
	}

	return record, nil
}

func (s *StorageManager) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {
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

	s.movesLock.RLock()
	_, ok = s.moves[game.Id]
	s.movesLock.RUnlock()
	if !ok {
		s.movesLock.Lock()
		s.moves[game.Id] = make(map[int]*boardgame.MoveStorageRecord)
		s.movesLock.Unlock()
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

	s.movesLock.RLock()
	moveMap := s.moves[game.Id]
	_, ok = moveMap[version]
	s.movesLock.RUnlock()

	if ok {
		//Wait, there was already a version stored there?
		return errors.New("There was already a version for that game stored")
	}

	s.extendedGamesLock.RLock()
	eGame, ok := s.extendedGames[game.Id]
	s.extendedGamesLock.RUnlock()
	if !ok {
		s.extendedGamesLock.Lock()
		s.extendedGames[game.Id] = extendedgame.DefaultStorageRecord()
		s.extendedGamesLock.Unlock()
	} else {
		eGame.LastActivity = time.Now().UnixNano()
	}

	s.statesLock.Lock()
	versionMap[version] = state
	s.statesLock.Unlock()

	s.movesLock.Lock()
	moveMap[version] = move
	s.movesLock.Unlock()

	s.gamesLock.Lock()
	s.games[game.Id] = game
	s.gamesLock.Unlock()

	return nil
}

func keyForAgent(gameId string, player boardgame.PlayerIndex) string {
	return gameId + "-" + player.String()
}

func (s *StorageManager) AgentState(gameId string, player boardgame.PlayerIndex) ([]byte, error) {

	key := keyForAgent(gameId, player)

	s.agentStatesLock.RLock()
	result := s.agentStates[key]
	s.agentStatesLock.RUnlock()

	return result, nil
}

func (s *StorageManager) SaveAgentState(gameId string, player boardgame.PlayerIndex, state []byte) error {
	key := keyForAgent(gameId, player)

	s.agentStatesLock.Lock()
	s.agentStates[key] = state
	s.agentStatesLock.Unlock()

	return nil
}

//ListGames will return game objects for up to max number of games
func (s *StorageManager) ListGames(max int, list listing.Type, userId string) []*extendedgame.CombinedStorageRecord {

	var result []*extendedgame.CombinedStorageRecord

	for _, game := range s.games {

		eGame := s.extendedGames[game.Id]

		result = append(result, &extendedgame.CombinedStorageRecord{
			*game,
			*eGame,
		})

		if len(result) >= max {
			break
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].LastActivity > result[j].LastActivity
	})

	return result

}

func (s *StorageManager) ExtendedGame(id string) (*extendedgame.StorageRecord, error) {
	s.extendedGamesLock.RLock()
	eGame := s.extendedGames[id]
	s.extendedGamesLock.RUnlock()
	if eGame == nil {
		return nil, errors.New("No such extended game")
	}

	return eGame, nil
}

func (s *StorageManager) CombinedGame(id string) (*extendedgame.CombinedStorageRecord, error) {
	s.extendedGamesLock.RLock()
	eGame := s.extendedGames[id]
	s.extendedGamesLock.RUnlock()
	if eGame == nil {
		return nil, errors.New("No such extended game")
	}

	game, err := s.Game(id)

	if err != nil {
		return nil, err
	}

	result := &extendedgame.CombinedStorageRecord{
		*game,
		*eGame,
	}

	return result, nil
}

func (s *StorageManager) UpdateExtendedGame(id string, eGame *extendedgame.StorageRecord) error {
	s.extendedGamesLock.Lock()
	s.extendedGames[id] = eGame
	s.extendedGamesLock.Unlock()
	return nil
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

func (s *StorageManager) Connect(config string) error {
	return nil
}

func (s *StorageManager) Close() {
	//Don't need to do anything
}

func (s *StorageManager) CleanUp() {
	//Don't need to do
}

func (s *StorageManager) PlayerMoveApplied(game *boardgame.GameStorageRecord) error {
	//Don't need to do anything
	return nil
}
