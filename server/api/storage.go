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
	//Close should be called before the server is shut down.
	Close()
	ListGames(max int) []*boardgame.GameStorageRecord

	//Store or update all fields
	UpdateUser(user *users.StorageRecord) error

	GetUserByCookie(cookie string) *users.StorageRecord

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
