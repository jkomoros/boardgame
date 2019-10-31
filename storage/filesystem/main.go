/*

Package filesystem is a storage layer that stores information about games as
JSON files within a given folder, (or somewhere nested in a folder within base
folder) one per game. It's extremely inefficient and doesn't even persist
extended game information to disk. It's most useful for cases where having an
easy-to-read, diffable representation for games makes sense, for example to
create golden tester games for use in testing.

filesystem stores files in the given base folder. If a sub-folder exists with
the name of the gameType, then the game will be stored in that folder instead.
For example if the gametype is "checkers" and the checkers subdir exists, will
store at  'checkers/a22ffcdef.json'. Folders may be soft- linked from within the
base folder; often when using the filesystem storage layer to help generate test
cases you set up soft- links from a central location to a folder for test files
in each game's sub-directory, so the test files can be in the same place.

*/
package filesystem

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/jkomoros/boardgame/storage/filesystem/record"
	"github.com/jkomoros/boardgame/storage/internal/helpers"
)

//StorageManager is the primary type for this package.
type StorageManager struct {
	//Fall back on those methods
	*helpers.ExtendedMemoryStorageManager
	basePath string
	managers []*boardgame.GameManager
	//Only shoiuld be on in testing scenarios
	forceFullEncoding bool
}

//Store seen ids and remember where the path was
var idToPath map[string]string

func init() {
	idToPath = make(map[string]string)
}

//NewStorageManager returns a new filesystem storage manager. basePath is the
//folder, relative to this executable, to have as the root of the storage
//pool.
func NewStorageManager(basePath string) *StorageManager {

	result := &StorageManager{
		basePath: basePath,
	}

	result.ExtendedMemoryStorageManager = helpers.NewExtendedMemoryStorageManager(result)

	return result
}

//Name returns 'filesystem'
func (s *StorageManager) Name() string {
	return "filesystem"
}

//Connect verifies the given basePath exists.
func (s *StorageManager) Connect(config string) error {

	if _, err := os.Stat(s.basePath); os.IsNotExist(err) {
		if err := os.Mkdir(s.basePath, 0700); err != nil {
			return errors.New("Base path didn't exist and couldn't create it: " + err.Error())
		}
	}

	return nil
}

//WithManagers sets the managers
func (s *StorageManager) WithManagers(managers []*boardgame.GameManager) {
	s.managers = managers
}

//CleanUp cleans up evertyhing in basePath.
func (s *StorageManager) CleanUp() {
	os.RemoveAll(s.basePath)
}

//pathForID will look through each sub-folder and look for a file named
//gameId.json, returning its relative path if it is found, "" otherwise.
func pathForID(basePath, gameID string) string {

	if path, ok := idToPath[gameID]; ok {
		return path
	}

	items, err := ioutil.ReadDir(basePath)
	if err != nil {
		return ""
	}
	for _, item := range items {
		if item.IsDir() {
			if recursiveResult := pathForID(filepath.Join(basePath, item.Name()), gameID); recursiveResult != "" {
				return recursiveResult
			}
			continue
		}

		if item.Name() == gameID+".json" {
			result := filepath.Join(basePath, item.Name())
			idToPath[gameID] = result
			return result
		}
	}
	return ""
}

func (s *StorageManager) recordForID(gameID string) (*record.Record, error) {
	if s.basePath == "" {
		return nil, errors.New("No base path provided")
	}

	gameID = strings.ToLower(gameID)

	path := pathForID(s.basePath, gameID)

	if path == "" {
		return nil, errors.New("Couldn't find file matching: " + gameID)
	}

	return record.New(path)
}

func (s *StorageManager) saveRecordForID(gameID string, rec *record.Record) error {
	if s.basePath == "" {
		return errors.New("Invalid base path")
	}

	if rec.Game() == nil {
		return errors.New("Game record in rec was nil")
	}

	gameID = strings.ToLower(gameID)

	//If a sub directory for that game type exists, save there. If not, save in the root of basePath.
	gameTypeSubDir := filepath.Join(s.basePath, rec.Game().Name)

	var path string

	if _, err := os.Stat(gameTypeSubDir); err == nil {
		path = filepath.Join(gameTypeSubDir, gameID+".json")
	} else {
		path = filepath.Join(s.basePath, gameID+".json")
	}

	if err := rec.Save(path, false); err != nil {
		return err
	}

	idToPath[gameID] = path

	return nil
}

//State returns the state for that gameID and version.
func (s *StorageManager) State(gameID string, version int) (boardgame.StateStorageRecord, error) {
	rec, err := s.recordForID(gameID)

	if err != nil {
		return nil, err
	}

	result, err := rec.State(version)

	if err != nil {
		return nil, err
	}

	return boardgame.StateStorageRecord(result), nil

}

//Move returns the move for that gameID and version
func (s *StorageManager) Move(gameID string, version int) (*boardgame.MoveStorageRecord, error) {
	rec, err := s.recordForID(gameID)

	if err != nil {
		return nil, err
	}

	return rec.Move(version)
}

//Moves returns all of the moves
func (s *StorageManager) Moves(gameID string, fromVersion, toVersion int) ([]*boardgame.MoveStorageRecord, error) {
	return helpers.MovesHelper(s, gameID, fromVersion, toVersion)
}

//Game returns the game storage record for that game.
func (s *StorageManager) Game(id string) (*boardgame.GameStorageRecord, error) {

	rec, err := s.recordForID(id)

	if err != nil {
		return nil, err
	}

	return rec.Game(), nil
}

//SaveGameAndCurrentState saves the game and current state.
func (s *StorageManager) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {
	rec, err := s.recordForID(game.ID)

	if err != nil {
		//Must be the first save.
		if s.forceFullEncoding {
			rec = record.EmptyWithFullStateEncoding()
		} else {
			rec = &record.Record{}
		}
	}

	if err := rec.AddGameAndCurrentState(game, state, move); err != nil {
		return errors.New("Couldn't add state: " + err.Error())
	}

	return s.saveRecordForID(game.ID, rec)

}

//CombinedGame returns the combined game
func (s *StorageManager) CombinedGame(id string) (*extendedgame.CombinedStorageRecord, error) {
	rec, err := s.recordForID(id)

	if err != nil {
		return nil, err
	}

	eGame, err := s.ExtendedGame(id)

	if err != nil {
		return nil, err
	}

	return &extendedgame.CombinedStorageRecord{
		GameStorageRecord: *rec.Game(),
		StorageRecord:     *eGame,
	}, nil
}

func idFromPath(path string) string {
	_, filename := filepath.Split(path)
	return strings.TrimSuffix(filename, ".json")
}

func (s *StorageManager) recursiveAllGames(basePath string) []*boardgame.GameStorageRecord {

	files, err := ioutil.ReadDir(basePath)

	if err != nil {
		return nil
	}

	var result []*boardgame.GameStorageRecord

	for _, file := range files {

		if file.IsDir() {
			result = append(result, s.recursiveAllGames(filepath.Join(basePath, file.Name()))...)
			continue
		}
		ext := filepath.Ext(file.Name())
		if ext != ".json" {
			continue
		}
		rec, err := s.recordForID(idFromPath(file.Name()))
		if err != nil {
			return nil
		}
		result = append(result, rec.Game())
	}
	return result
}

//AllGames returns all games
func (s *StorageManager) AllGames() []*boardgame.GameStorageRecord {
	return s.recursiveAllGames(s.basePath)
}

//ListGames returns all of the games
func (s *StorageManager) ListGames(max int, list listing.Type, userID string, gameType string) []*extendedgame.CombinedStorageRecord {
	return helpers.ListGamesHelper(s, max, list, userID, gameType)
}
