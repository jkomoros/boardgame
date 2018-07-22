/*

	filesystem is a storage layer that stores information about games as JSON
	files within a given folder, one per game. It's extremely inefficient and
	doesn't even persist extended game information to disk. It's most useful
	for cases where having an easy-to-read, diffable representation for games
	makes sense, for example to create golden tester games for use in testing.

*/
package filesystem

import (
	"encoding/json"
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func (s *StorageManager) recordForId(gameId string) (*record, error) {
	if s.basePath == "" {
		return nil, errors.New("No base path provided")
	}

	gameId = strings.ToLower(gameId)

	path := filepath.Join(s.basePath, gameId)

	var result record

	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, errors.New("Couldn't read file: " + err.Error())
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, errors.New("Couldn't decode json: " + err.Error())
	}

	return &result, nil
}

func (s *StorageManager) saveRecordForId(gameId string, rec *record) error {
	if s.basePath == "" {
		return errors.New("Invalid base path")
	}

	gameId = strings.ToLower(gameId)

	path := filepath.Join(s.basePath, gameId)

	blob, err := json.Marshal(rec)

	if err != nil {
		return errors.New("Couldn't marshal blob: " + err.Error())
	}

	return ioutil.WriteFile(path, blob, 0644)
}

func (s *StorageManager) State(gameId string, version int) (boardgame.StateStorageRecord, error) {
	rec, err := s.recordForId(gameId)

	if err != nil {
		return nil, err
	}

	if len(rec.States) < version {
		return nil, errors.New("Not enough states to return: " + strconv.Itoa(len(rec.States)))
	}

	return rec.States[version], nil
}

func (s *StorageManager) Move(gameId string, version int) (*boardgame.MoveStorageRecord, error) {
	rec, err := s.recordForId(gameId)

	if err != nil {
		return nil, err
	}

	if len(rec.Moves) < version {
		return nil, errors.New("Not enough moves to return: " + strconv.Itoa(len(rec.Moves)))
	}

	return rec.Moves[version], nil
}

func (s *StorageManager) Moves(gameId string, fromVersion, toVersion int) ([]*boardgame.MoveStorageRecord, error) {
	var result []*boardgame.MoveStorageRecord

	if fromVersion == toVersion {
		move, err := s.Move(gameId, toVersion)
		if err != nil {
			return nil, err
		}
		return []*boardgame.MoveStorageRecord{
			move,
		}, nil
	}

	for i := fromVersion; i < toVersion; i++ {
		move, err := s.Move(gameId, i)
		if err != nil {
			return nil, err
		}
		result = append(result, move)
	}

	return result, nil
}

func (s *StorageManager) Game(id string) (*boardgame.GameStorageRecord, error) {

	rec, err := s.recordForId(id)

	if err != nil {
		return nil, err
	}

	return rec.Game, nil
}

func (s *StorageManager) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {
	rec, err := s.recordForId(game.Id)

	if err != nil {
		//Must be the first save.
		rec = &record{}
	}

	rec.Game = game

	rec.States = append(rec.States, state)

	if move != nil {
		rec.Moves = append(rec.Moves, move)
	}

	return s.saveRecordForId(game.Id, rec)

}
