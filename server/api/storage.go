package api

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/users"
	"github.com/jkomoros/boardgame/storage/mysql"
)

//StorageManager extends the base boardgame.StorageManager with a few more
//methods necessary to make server work. When creating a new Server, you need
//to pass in a ServerStorageManager, which wraps one of these objects and thus
//implements these methods, too.
type StorageManager interface {
	boardgame.StorageManager

	//Name returns the name of the storage manager type, for example "memory", "bolt", or "mysql"
	Name() string

	//Connect will be called before issuing any other substantive calls. The
	//config string is specific to the type of storage layer, which can be
	//interrogated with Nmae().
	Connect(config string) error

	//Close should be called before the server is shut down.
	Close()
	ListGames(max int) []*boardgame.GameStorageRecord

	//UserIdsForGame returns an array whose length equals game.NumPlayers.
	//Each one is either empty if there is no user in that slot yet, or the
	//uid representing the user.
	UserIdsForGame(gameId string) []string

	SetPlayerForGame(gameId string, playerIndex boardgame.PlayerIndex, userId string) error

	//Store or update all fields
	UpdateUser(user *users.StorageRecord) error

	GetUserById(uid string) *users.StorageRecord

	GetUserByCookie(cookie string) *users.StorageRecord

	//If user is nil, the cookie should be deleted if it exists. If the user
	//does not yet exist, it should be added to the database.
	ConnectCookieToUser(cookie string, user *users.StorageRecord) error

	//Note: whenever you add methods here, also add them to boardgame/storage/test/StorageManager
}

//ServerStorageManager implements the ServerStorage interface by wrapping an
//object that supports StorageManager.
type ServerStorageManager struct {
	StorageManager
	server *Server
}

//NewServerStorageManager takes an object that implements StorageManager and
//wraps it.
func NewServerStorageManager(manager StorageManager) *ServerStorageManager {
	return &ServerStorageManager{
		manager,
		nil,
	}
}

//NewDefaultStorageManager currently uses mysql. See the README in
//github.com/jkomoros/boardgame/storage/mysql for how to set up and configure it.
func NewDefaultStorageManager() *ServerStorageManager {
	return NewServerStorageManager(mysql.NewStorageManager(false))
}

//We wrap SaveGameandCurrentState so we can update our game version cache
func (s *ServerStorageManager) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord) error {

	result := s.StorageManager.SaveGameAndCurrentState(game, state)

	if result != nil {
		return result
	}

	server := s.server

	if server == nil {
		return errors.New("No server configured. The storage manager should be added to a Server before it's used.")
	}

	server.gameVersionCacheLock.Lock()

	server.gameVersionCache[game.Id] = game.Version

	server.gameVersionCacheLock.Unlock()

	return nil

}
