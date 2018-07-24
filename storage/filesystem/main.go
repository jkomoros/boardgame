/*

	filesystem is a storage layer that stores information about games as JSON
	files within a given folder, one per game. It's extremely inefficient and
	doesn't even persist extended game information to disk. It's most useful
	for cases where having an easy-to-read, diffable representation for games
	makes sense, for example to create golden tester games for use in testing.

	filesystem stores files according to their gametype in the given base
	folder, for example 'checkers/a22ffcdef.json'. If the sub-folders don't
	exist, they will be created. Folders may be soft-linked from within the
	base folder; often when using the filesystem storage layer to help
	generate test cases you set up soft-links from a central location to a
	folder for test files in each game's sub-directory, so the test files can
	be in the same place.

*/
package filesystem

import (
	"bytes"
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/jkomoros/boardgame/storage/filesystem/record"
	"github.com/jkomoros/boardgame/storage/internal/helpers"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type StorageManager struct {
	//Fall back on those methods
	*helpers.ExtendedMemoryStorageManager
	basePath         string
	goldenFolderName string
	managers         []*boardgame.GameManager
}

//Store seen ids and remember where the path was
var idToPath map[string]string

func init() {
	idToPath = make(map[string]string)
}

//NewStorageManager returns a new filesystem storage manager. basePath is the
//folder, relative to this executable, to have as the root of the storage
//pool. If goldenFolderName is not "", then we will use reflection to find the
//package path for each delegate, ensure a folder exists within it with that
//name, create a soft-link from basePath to that folder, and create a
//`golden_test.go` file that automatically tests all of those golden files
//(and assumes that your package defines a `NewDelegate()
//boardgame.GameDelegate` method). The result is that the underlying files
//will be stored in folders adjacent to the games they are relative to, which
//is convenient if you're adding new golden games to the test set.
func NewStorageManager(basePath string, goldenFolderName string) *StorageManager {

	result := &StorageManager{
		basePath:         basePath,
		goldenFolderName: goldenFolderName,
	}

	result.ExtendedMemoryStorageManager = helpers.NewExtendedMemoryStorageManager(result)

	return result
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

	if s.goldenFolderName != "" {

		log.Println("Creating and linking golden folders")
		for _, manager := range s.managers {
			pkgPath := reflect.ValueOf(manager.Delegate()).Elem().Type().PkgPath()
			if err := s.LinkGoldenFolder(manager.Delegate().Name(), pkgPath); err != nil {
				return errors.New("Couldn't link golden folder for " + manager.Delegate().Name() + ": " + err.Error())
			}
		}
	}

	return nil
}

func (s *StorageManager) LinkGoldenFolder(gameType, pkgPath string) error {
	//TODO: does this handle vendoring correctly?

	if s.goldenFolderName == "" {
		return nil
	}

	goPath := os.Getenv("GOPATH")

	if goPath == "" {
		return errors.New("Gopath wasn't set")
	}

	fullPkgPath := filepath.Join(goPath, "src", pkgPath)

	fullPath := filepath.Join(fullPkgPath, s.goldenFolderName)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		log.Println("Creating " + fullPath)
		if err := os.Mkdir(fullPath, 0700); err != nil {
			return errors.New("Could not make golden path: " + err.Error())
		}
	}

	gamePath := filepath.Join(s.basePath, gameType)

	if _, err := os.Stat(gamePath); os.IsNotExist(err) {
		log.Println("Linking " + gamePath + " to " + fullPath)
		//Soft link from basePath.
		if err := os.Symlink(fullPath, gamePath); err != nil {
			return errors.New("Couldn't create symlink: " + err.Error())
		}
	}

	//TODO: ok, this is getting weird that the filesystem storage layer is now
	//a full-fledged util.

	//TODO: allow this to be skipped as an otpion.
	if err := s.SaveGoldenTest(fullPkgPath); err != nil {
		return errors.New("Couldn't store golden test: " + err.Error())
	}

	return nil

}

func (s *StorageManager) SaveGoldenTest(fullPkgPath string) error {

	pkgName, err := verifyPkgForGolden(fullPkgPath)

	if err != nil {
		return errors.New("Package didn't validate: " + err.Error())
	}

	buf := new(bytes.Buffer)

	err = goldenTestTemplate.Execute(buf, map[string]string{
		"gametype": pkgName,
		"folder":   s.goldenFolderName,
	})

	if err != nil {
		return errors.New("Couldn't generate blob from template: " + err.Error())
	}

	return ioutil.WriteFile(filepath.Join(fullPkgPath, "golden_test.go"), buf.Bytes(), 0644)

}

//verifyPkgForGolden looks at the given package, returns the package name, and
//verifies that it has a NewDelegate method.
func verifyPkgForGolden(fullPkgName string) (string, error) {
	pkgs, err := parser.ParseDir(token.NewFileSet(), fullPkgName, nil, 0)

	if err != nil {
		return "", errors.New("Couldn't parse folder: " + err.Error())
	}

	if len(pkgs) < 1 {
		return "", errors.New("No packages in that directory")
	}

	if len(pkgs) > 1 {
		return "", errors.New("More than one package in that directory")
	}

	var pkg *ast.Package
	pkgName := ""

	for key, p := range pkgs {
		pkgName = key
		pkg = p
	}

	foundNewDelegate := false

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			fun, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if fun.Name.String() != "NewDelegate" {
				continue
			}

			//OK, it might be the function. Does it have the right signature?

			if fun.Recv != nil {
				return "", errors.New("NewDelegate had a receiver")
			}

			if fun.Type.Params.NumFields() > 0 {
				return "", errors.New("NewDelegate took more than 0 items")
			}

			if fun.Type.Results.NumFields() != 1 {
				return "", errors.New("NewDelegate didn't return exactly one item")
			}

			//TODO: check that the returned item implements
			//boardgame.GameDelegate.

			foundNewDelegate = true
			break

		}

		if foundNewDelegate {
			break
		}
	}

	if !foundNewDelegate {
		return "", errors.New("Couldn't find NewDelegate")
	}

	return pkgName, nil

}

func (s *StorageManager) WithManagers(managers []*boardgame.GameManager) {
	s.managers = managers
}

func (s *StorageManager) CleanUp() {
	os.RemoveAll(s.basePath)
}

//pathForId will look through each sub-folder and look for a file named
//gameId.json, returning its relative path if it is found, "" otherwise.
func pathForId(basePath, gameId string) string {

	if path, ok := idToPath[gameId]; ok {
		return path
	}

	items, err := ioutil.ReadDir(basePath)
	if err != nil {
		return ""
	}
	for _, item := range items {
		if item.IsDir() {
			if recursiveResult := pathForId(filepath.Join(basePath, item.Name()), gameId); recursiveResult != "" {
				return recursiveResult
			}
			continue
		}

		if item.Name() == gameId+".json" {
			result := filepath.Join(basePath, item.Name())
			idToPath[gameId] = result
			return result
		}
	}
	return ""
}

func (s *StorageManager) recordForId(gameId string) (*record.Record, error) {
	if s.basePath == "" {
		return nil, errors.New("No base path provided")
	}

	gameId = strings.ToLower(gameId)

	path := pathForId(s.basePath, gameId)

	if path == "" {
		return nil, errors.New("Couldn't find file matching: " + gameId)
	}

	return record.New(path)
}

func (s *StorageManager) saveRecordForId(gameId string, rec *record.Record) error {
	if s.basePath == "" {
		return errors.New("Invalid base path")
	}

	if rec.Game() == nil {
		return errors.New("Game record in rec was nil")
	}

	gameId = strings.ToLower(gameId)

	path := filepath.Join(s.basePath, rec.Game().Name, gameId+".json")

	dir, _ := filepath.Split(path)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.New("Couldn't create all necessary sub-paths: " + err.Error())
	}

	if err := rec.Save(path); err != nil {
		return err
	}

	idToPath[gameId] = path

	return nil
}

func (s *StorageManager) State(gameId string, version int) (boardgame.StateStorageRecord, error) {
	rec, err := s.recordForId(gameId)

	if err != nil {
		return nil, err
	}

	result, err := rec.State(version)

	if err != nil {
		return nil, err
	}

	return boardgame.StateStorageRecord(result), nil

}

func (s *StorageManager) Move(gameId string, version int) (*boardgame.MoveStorageRecord, error) {
	rec, err := s.recordForId(gameId)

	if err != nil {
		return nil, err
	}

	return rec.Move(version)
}

func (s *StorageManager) Moves(gameId string, fromVersion, toVersion int) ([]*boardgame.MoveStorageRecord, error) {
	return helpers.MovesHelper(s, gameId, fromVersion, toVersion)
}

func (s *StorageManager) Game(id string) (*boardgame.GameStorageRecord, error) {

	rec, err := s.recordForId(id)

	if err != nil {
		return nil, err
	}

	return rec.Game(), nil
}

func (s *StorageManager) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {
	rec, err := s.recordForId(game.Id)

	if err != nil {
		//Must be the first save.
		rec = &record.Record{}
	}

	if err := rec.AddGameAndCurrentState(game, state, move); err != nil {
		return errors.New("Couldn't add state: " + err.Error())
	}

	return s.saveRecordForId(game.Id, rec)

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
		*rec.Game(),
		*eGame,
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
		rec, err := s.recordForId(idFromPath(file.Name()))
		if err != nil {
			return nil
		}
		result = append(result, rec.Game())
	}
	return result
}

func (s *StorageManager) AllGames() []*boardgame.GameStorageRecord {
	return s.recursiveAllGames(s.basePath)
}

func (s *StorageManager) ListGames(max int, list listing.Type, userId string, gameType string) []*extendedgame.CombinedStorageRecord {
	return helpers.ListGamesHelper(s, max, list, userId, gameType)
}
