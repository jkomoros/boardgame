/*

	filesystem is a storage layer that stores information about games as JSON
	files within a given folder, one per game. It's extremely inefficient and
	doesn't even persist extended game information to disk. It's most useful
	for cases where having an easy-to-read, diffable representation for games
	makes sense, for example to create golden tester games for use in testing.

*/
package filesystem

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
	"os"
)

type StorageManager struct {
	//Fall back on those methods
	*memory.StorageManager
	basePath string
}

func NewStorageManager() *StorageManager {

	panic("This is not yet implemented")

	return &StorageManager{
		memory.NewStorageManager(),
		"",
	}
}

func (s *StorageManager) Name() string {
	return "filesystem"
}

func (s *StorageManager) Connect(config string) error {

	if _, err := os.Stat(config); os.IsNotExist(err) {
		return errors.New("BasePath of " + config + " does not exist.")
	}

	s.basePath = config

	return nil
}

func (s *StorageManager) State(gameId string, version int) (boardgame.StateStorageRecord, error) {
	return nil, nil
}

func (s *StorageManager) Move(gameId string, version int) (*boardgame.MoveStorageRecord, error) {
	return nil, nil
}

func (s *StorageManager) Moves(gameId string, fromVersion, toVersion int) ([]*boardgame.MoveStorageRecord, error) {
	return nil, nil
}

func (s *StorageManager) Game(id string) (*boardgame.GameStorageRecord, error) {
	return nil, nil
}

func (s *StorageManager) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {
	return nil
}
