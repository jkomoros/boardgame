package api

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/users"
	"github.com/jkomoros/boardgame/storage/bolt"
)

//StorageManager extends the base boardgame.StorageManager with a few more
//methods necessary to make server work.
type StorageManager interface {
	boardgame.StorageManager

	//Name returns the name of the storage manager type, for example "memory", "bolt", or "mysql"
	Name() string

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

//DefaultStorageManager currently uses bolt. It will create the database file
//in the same directory the server is run.
type DefaultStorageManager struct {
	*bolt.StorageManager
}

func NewDefaultStorageManager() *DefaultStorageManager {
	return &DefaultStorageManager{
		bolt.NewStorageManager(".database"),
	}
}
