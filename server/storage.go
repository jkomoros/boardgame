package server

import (
	"github.com/jkomoros/boardgame"
)

//StorageManager extends the base boardgame.StorageManager with a few more
//methods necessary to make server work.
type StorageManager interface {
	boardgame.StorageManager
	ListGames(manager *boardgame.GameManager, max int) []*boardgame.Game
}

type DefaultStorageManager struct {
	*boardgame.InMemoryStorageManager
}

func NewDefaultStorageManager() *DefaultStorageManager {
	return &DefaultStorageManager{
		boardgame.NewInMemoryStorageManager(),
	}
}
