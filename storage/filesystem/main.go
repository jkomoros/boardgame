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
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/jkomoros/boardgame/storage/internal/helpers"
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

func NewStorageManager(basePath string) *StorageManager {

	return &StorageManager{
		memory.NewStorageManager(),
		basePath,
	}
}

func (s *StorageManager) Name() string {
	return "filesystem"
}

func (s *StorageManager) Connect(config string) error {

	if _, err := os.Stat(s.basePath); os.IsNotExist(err) {
		if err := os.Mkdir(s.basePath, 0700); err != nil {
			return errors.New("Base path didn't exist and couldn't create it: " + err.Error())
		}
	}

	return nil
}

func (s *StorageManager) CleanUp() {
	os.RemoveAll(s.basePath)
}

func (s *StorageManager) recordForId(gameId string) (*record, error) {
	if s.basePath == "" {
		return nil, errors.New("No base path provided")
	}

	gameId = strings.ToLower(gameId)

	path := filepath.Join(s.basePath, gameId+".json")

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

	path := filepath.Join(s.basePath, gameId+".json")

	blob, err := json.MarshalIndent(rec, "", "\t")

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

	//version is effectively 1-indexed, since we don't store a move for the
	//first version, but we store them in 0-indexed since we use the array
	//index. So convert to that.

	version -= 1

	if len(rec.Moves) < version {
		return nil, errors.New("Not enough moves to return: " + strconv.Itoa(len(rec.Moves)))
	}

	return rec.Moves[version], nil
}

func (s *StorageManager) Moves(gameId string, fromVersion, toVersion int) ([]*boardgame.MoveStorageRecord, error) {
	return helpers.MovesHelper(s, gameId, fromVersion, toVersion)
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

	if err := s.saveRecordForId(game.Id, rec); err != nil {
		return errors.New("Couldn't save primary game: " + err.Error())
	}

	//Also pass down into the memory so that other things like ExtendedGame
	//work as expected. Note that this won't work for games that exist in
	//filesystem when the storage maanager is booted; but this is primarily
	//just to pass the server.StorageManager test suite.
	return s.StorageManager.SaveGameAndCurrentState(game, state, move)

}

func (s *StorageManager) CombinedGame(id string) (*extendedgame.CombinedStorageRecord, error) {
	rec, err := s.recordForId(id)

	if err != nil {
		return nil, err
	}

	eGame, err := s.ExtendedGame(id)

	if err != nil {
		return nil, err
	}

	return &extendedgame.CombinedStorageRecord{
		*rec.Game,
		*eGame,
	}, nil
}

func (s *StorageManager) ListGames(max int, list listing.Type, userId string, gameType string) []*extendedgame.CombinedStorageRecord {
	return nil
}
