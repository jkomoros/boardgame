package server

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/bolt"
)

//StorageManager extends the base boardgame.StorageManager with a few more
//methods necessary to make server work.
type StorageManager interface {
	boardgame.StorageManager
	//Close should be called before the server is shut down.
	Close()
	ListGames(manager *boardgame.GameManager, max int) []*boardgame.Game
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
