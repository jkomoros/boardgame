package api

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/jkomoros/boardgame/server/api/users"
	"github.com/jkomoros/boardgame/storage/mysql"
)

//StorageManager extends the base boardgame.StorageManager with a few more
//methods necessary to make server work. When creating a new Server, you need
//to pass in a ServerStorageManager, which wraps one of these objects and thus
//implements these methods, too.
type StorageManager interface {

	//StorageManager extends the boardgame.StorageManager interface. Those
	//methods have two additional semantic expectations, however:
	//SaveGameAndCurrentState should create an ExtendedGameStorageRecord on
	//the first save of a game.
	boardgame.StorageManager

	//Name returns the name of the storage manager type, for example "memory", "bolt", or "mysql"
	Name() string

	//WithManagers is called during set up with references to all of the
	//managers. Will be called before Connect() is called.
	WithManagers(managers []*boardgame.GameManager)

	//Connect will be called before issuing any other substantive calls. The
	//config string is specific to the type of storage layer, which can be
	//interrogated with Nmae().
	Connect(config string) error

	//ExtendedGame is like Game(), but it returns an extended storage record
	//with additional fields necessary for Server.
	ExtendedGame(id string) (*extendedgame.StorageRecord, error)

	CombinedGame(id string) (*extendedgame.CombinedStorageRecord, error)

	//UpdateExtendedGame updates the extended game with the given Id.
	UpdateExtendedGame(id string, eGame *extendedgame.StorageRecord) error

	//Close should be called before the server is shut down.
	Close()

	//ListGames should list up to max games, in descending order based on the
	//LastActivity. If gameType is not "", only returns games that are that
	//gameType. If gameType is "", all gametypes are fine.
	ListGames(max int, list listing.Type, userId string, gameType string) []*extendedgame.CombinedStorageRecord

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

func (s *ServerStorageManager) PlayerMoveApplied(game *boardgame.GameStorageRecord) error {

	//Do the wrapped manager's PlayerMoveApplied in case it has one.
	if err := s.StorageManager.PlayerMoveApplied(game); err != nil {
		return err
	}

	server := s.server

	if server == nil {
		return errors.New("No server configured. The storage manager should be added to a Server before it's used.")
	}

	//Notify the web sockets that the game was changed
	server.notifier.gameChanged(game)

	return nil

}
