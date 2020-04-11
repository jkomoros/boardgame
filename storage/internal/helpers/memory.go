package helpers

import (
	"errors"
	"sync"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/users"
)

//GameChecker is just one meethod to fetch a Game given an ID. It's typically
//just the storage manager you're embedded in but can be other objects in other
//cases, thus why it's split out separately.
type GameChecker interface {
	Game(gameID string) (*boardgame.GameStorageRecord, error)
}

//ExtendedMemoryStorageManager implements the ExtendedGame methods (i.e the
//methods in server.StorageManager) in memory. Designed to be embedded
//anonymously in the containing item. Get a new one from
//NewExtendedMemoryStorageManager.
type ExtendedMemoryStorageManager struct {
	agentStates   map[string][]byte
	extendedGames map[string]*extendedgame.StorageRecord
	usersByID     map[string]*users.StorageRecord
	usersByCookie map[string]*users.StorageRecord
	usersForGames map[string][]string

	agentStatesLock   sync.RWMutex
	extendedGamesLock sync.RWMutex
	usersLock         sync.RWMutex
	usersForGamesLock sync.RWMutex

	gameChecker GameChecker
}

//NewExtendedMemoryStorageManager returns a new extended memory storage manager.
//Checker is generally the storage manager you're embedded in.
func NewExtendedMemoryStorageManager(checker GameChecker) *ExtendedMemoryStorageManager {

	if checker == nil {
		return nil
	}

	return &ExtendedMemoryStorageManager{
		extendedGames: make(map[string]*extendedgame.StorageRecord),
		usersByID:     make(map[string]*users.StorageRecord),
		usersByCookie: make(map[string]*users.StorageRecord),
		usersForGames: make(map[string][]string),
		agentStates:   make(map[string][]byte),
		gameChecker:   checker,
	}
}

func keyForAgent(gameID string, player boardgame.PlayerIndex) string {
	return gameID + "-" + player.String()
}

//AgentState implements the AgentState part of the interface.
func (s *ExtendedMemoryStorageManager) AgentState(gameID string, player boardgame.PlayerIndex) ([]byte, error) {

	key := keyForAgent(gameID, player)

	s.agentStatesLock.RLock()
	result := s.agentStates[key]
	s.agentStatesLock.RUnlock()

	return result, nil
}

//SaveAgentState implements that part of the StorageManager interface.
func (s *ExtendedMemoryStorageManager) SaveAgentState(gameID string, player boardgame.PlayerIndex, state []byte) error {
	key := keyForAgent(gameID, player)

	s.agentStatesLock.Lock()
	s.agentStates[key] = state
	s.agentStatesLock.Unlock()

	return nil
}

//CombinedGame implements that part of the server interface.
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

//UpdateExtendedGame implements that part of the server storage interface
func (s *ExtendedMemoryStorageManager) UpdateExtendedGame(id string, eGame *extendedgame.StorageRecord) error {
	s.extendedGamesLock.Lock()
	s.extendedGames[id] = eGame
	s.extendedGamesLock.Unlock()
	return nil
}

//UserIDsForGame implements that part of the server storage interface.
func (s *ExtendedMemoryStorageManager) UserIDsForGame(gameID string) []string {
	s.usersForGamesLock.RLock()
	ids := s.usersForGames[gameID]
	s.usersForGamesLock.RUnlock()

	if ids == nil {
		game, _ := s.gameChecker.Game(gameID)
		if game == nil {
			return nil
		}
		return make([]string, game.NumPlayers)
	}

	return ids
}

//SetPlayerForGame implemnts that part of the server storage interface.
func (s *ExtendedMemoryStorageManager) SetPlayerForGame(gameID string, playerIndex boardgame.PlayerIndex, userID string) error {
	ids := s.UserIDsForGame(gameID)

	if int(playerIndex) < 0 || int(playerIndex) >= len(ids) {
		return errors.New("PlayerIndex " + playerIndex.String() + " is not valid for this game.")
	}

	if ids[playerIndex] != "" {
		return errors.New("PlayerIndex " + playerIndex.String() + " is already taken.")
	}

	user := s.GetUserByID(userID)

	if user == nil {
		return errors.New("That uid does not describe an existing user")
	}

	ids[playerIndex] = userID

	s.usersForGamesLock.Lock()
	s.usersForGames[gameID] = ids
	s.usersForGamesLock.Unlock()

	return nil
}

//UpdateUser stores or update all fields
func (s *ExtendedMemoryStorageManager) UpdateUser(user *users.StorageRecord) error {

	s.usersLock.Lock()
	s.usersByID[user.ID] = user
	s.usersLock.Unlock()

	return nil

}

//GetUserByID implements that part of the server storage interface.
func (s *ExtendedMemoryStorageManager) GetUserByID(uid string) *users.StorageRecord {
	s.usersLock.RLock()
	user := s.usersByID[uid]
	s.usersLock.RUnlock()

	return user
}

//GetUserByCookie implements that part of the server storage interface.
func (s *ExtendedMemoryStorageManager) GetUserByCookie(cookie string) *users.StorageRecord {
	s.usersLock.RLock()
	user := s.usersByCookie[cookie]
	s.usersLock.RUnlock()

	return user
}

//ConnectCookieToUser implements that part of the server storage interface. If
//user is nil, the cookie should be deleted if it exists. If the user does not
//yet exist, it should be added to the database.
func (s *ExtendedMemoryStorageManager) ConnectCookieToUser(cookie string, user *users.StorageRecord) error {
	if user == nil {
		s.usersLock.Lock()
		delete(s.usersByCookie, cookie)
		s.usersLock.Unlock()
		return nil
	}

	otherUser := s.GetUserByID(user.ID)

	if otherUser == nil {
		s.UpdateUser(user)
	}

	s.usersLock.Lock()
	s.usersByCookie[cookie] = user
	s.usersLock.Unlock()

	return nil
}

//Provide defaults for all of these that are no op

//Connect is a no op
func (s *ExtendedMemoryStorageManager) Connect(config string) error {
	return nil
}

//Close is a no op
func (s *ExtendedMemoryStorageManager) Close() {
	//Don't need to do anything
}

//CleanUp is a no op
func (s *ExtendedMemoryStorageManager) CleanUp() {
	//Don't need to do
}

//PlayerMoveApplied is a no op
func (s *ExtendedMemoryStorageManager) PlayerMoveApplied(game *boardgame.GameStorageRecord) error {
	//Don't need to do anything
	return nil
}

//FetchInjectedDataForGame can just return nil
func (s *ExtendedMemoryStorageManager) FetchInjectedDataForGame(gameID string, dataType string) interface{} {
	//Don't need to do anything
	return nil
}

//WithManagers is a no op
func (s *ExtendedMemoryStorageManager) WithManagers(managers []*boardgame.GameManager) {
	//Do nothing
}
