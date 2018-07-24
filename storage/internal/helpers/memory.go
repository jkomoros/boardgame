package helpers

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/users"
	"sync"
)

type GameChecker interface {
	Game(gameId string) (*boardgame.GameStorageRecord, error)
}

//ExtendedMemoryStorageManager implements the ExtendedGame methods (i.e the
//methods in server.StorageManager) in memory. Designed to be embedded
//anonymously in the containing item. Get a new one from
//NewExtendedMemoryStorageManager.
type ExtendedMemoryStorageManager struct {
	agentStates   map[string][]byte
	extendedGames map[string]*extendedgame.StorageRecord
	usersById     map[string]*users.StorageRecord
	usersByCookie map[string]*users.StorageRecord
	usersForGames map[string][]string

	agentStatesLock   sync.RWMutex
	extendedGamesLock sync.RWMutex
	usersLock         sync.RWMutex
	usersForGamesLock sync.RWMutex

	gameChecker GameChecker
}

//Returns a new extended memory storage manager. Checker is generally the
//storage manager you're embedded in.
func NewExtendedMemoryStorageManager(checker GameChecker) *ExtendedMemoryStorageManager {

	if checker == nil {
		return nil
	}

	return &ExtendedMemoryStorageManager{
		extendedGames: make(map[string]*extendedgame.StorageRecord),
		usersById:     make(map[string]*users.StorageRecord),
		usersByCookie: make(map[string]*users.StorageRecord),
		usersForGames: make(map[string][]string),
		agentStates:   make(map[string][]byte),
		gameChecker:   checker,
	}
}

func keyForAgent(gameId string, player boardgame.PlayerIndex) string {
	return gameId + "-" + player.String()
}

func (s *ExtendedMemoryStorageManager) AgentState(gameId string, player boardgame.PlayerIndex) ([]byte, error) {

	key := keyForAgent(gameId, player)

	s.agentStatesLock.RLock()
	result := s.agentStates[key]
	s.agentStatesLock.RUnlock()

	return result, nil
}

func (s *ExtendedMemoryStorageManager) SaveAgentState(gameId string, player boardgame.PlayerIndex, state []byte) error {
	key := keyForAgent(gameId, player)

	s.agentStatesLock.Lock()
	s.agentStates[key] = state
	s.agentStatesLock.Unlock()

	return nil
}

func (s *ExtendedMemoryStorageManager) CombinedGame(id string) (*extendedgame.CombinedStorageRecord, error) {
	s.extendedGamesLock.RLock()
	eGame := s.extendedGames[id]
	s.extendedGamesLock.RUnlock()
	if eGame == nil {
		return nil, errors.New("No such extended game")
	}

	game, err := s.gameChecker.Game(id)

	if err != nil {
		return nil, err
	}

	result := &extendedgame.CombinedStorageRecord{
		GameStorageRecord: *game,
		StorageRecord:     *eGame,
	}

	return result, nil
}

//ExtendedGame will return extendedgame.DefaultStorageRecord() if the
//associated game exists.
func (s *ExtendedMemoryStorageManager) ExtendedGame(id string) (*extendedgame.StorageRecord, error) {
	s.extendedGamesLock.RLock()
	eGame := s.extendedGames[id]
	s.extendedGamesLock.RUnlock()
	if eGame == nil {

		if game, _ := s.gameChecker.Game(id); game != nil {
			//If there's supposed to be a game, return a default.
			return extendedgame.DefaultStorageRecord(), nil
		}

		return nil, errors.New("No such extended game")
	}

	return eGame, nil
}

func (s *ExtendedMemoryStorageManager) UpdateExtendedGame(id string, eGame *extendedgame.StorageRecord) error {
	s.extendedGamesLock.Lock()
	s.extendedGames[id] = eGame
	s.extendedGamesLock.Unlock()
	return nil
}

func (s *ExtendedMemoryStorageManager) UserIdsForGame(gameId string) []string {
	s.usersForGamesLock.RLock()
	ids := s.usersForGames[gameId]
	s.usersForGamesLock.RUnlock()

	if ids == nil {
		game, _ := s.gameChecker.Game(gameId)
		if game == nil {
			return nil
		}
		return make([]string, game.NumPlayers)
	}

	return ids
}

func (s *ExtendedMemoryStorageManager) SetPlayerForGame(gameId string, playerIndex boardgame.PlayerIndex, userId string) error {
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
func (s *ExtendedMemoryStorageManager) UpdateUser(user *users.StorageRecord) error {

	s.usersLock.Lock()
	s.usersById[user.Id] = user
	s.usersLock.Unlock()

	return nil

}

func (s *ExtendedMemoryStorageManager) GetUserById(uid string) *users.StorageRecord {
	s.usersLock.RLock()
	user := s.usersById[uid]
	s.usersLock.RUnlock()

	return user
}

func (s *ExtendedMemoryStorageManager) GetUserByCookie(cookie string) *users.StorageRecord {
	s.usersLock.RLock()
	user := s.usersByCookie[cookie]
	s.usersLock.RUnlock()

	return user
}

//If user is nil, the cookie should be deleted if it exists. If the user
//does not yet exist, it should be added to the database.
func (s *ExtendedMemoryStorageManager) ConnectCookieToUser(cookie string, user *users.StorageRecord) error {
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

//Provide defaults for all of these that are no op

//No op
func (s *ExtendedMemoryStorageManager) Connect(config string) error {
	return nil
}

//No op
func (s *ExtendedMemoryStorageManager) Close() {
	//Don't need to do anything
}

//No op
func (s *ExtendedMemoryStorageManager) CleanUp() {
	//Don't need to do
}

//No op
func (s *ExtendedMemoryStorageManager) PlayerMoveApplied(game *boardgame.GameStorageRecord) error {
	//Don't need to do anything
	return nil
}

func (s *ExtendedMemoryStorageManager) WithManagers(managers []*boardgame.GameManager) {
	//Do nothing
}
