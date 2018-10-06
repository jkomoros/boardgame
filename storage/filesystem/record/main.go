/*

	record is a package to open, read, and save game records stored in
	filesystem's format.

	Note that because reading the files from disk is expensive, this library
	maintains a cache of records by filename that it returns, for a
	considerable performance boost. This means that changes in the filesystem
	while the storage layer is running that aren't mediated by this controller
	will cause undefined behavior.

*/
package record

import (
	"encoding/json"
	"errors"
	"github.com/go-test/deep"
	"github.com/jkomoros/boardgame"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const randomStringChars = "ABCDEF0123456789"

var recCache map[string]*Record

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	recCache = make(map[string]*Record, 16)
}

//Record is a record of moves, states, and game. Get a new one based on the
//contents of a file with New(). Instantiate directly for a blank one.
type Record struct {
	data   *storageRecord
	states []boardgame.StateStorageRecord
}

type storageRecord struct {
	Game  *boardgame.GameStorageRecord
	Moves []*boardgame.MoveStorageRecord
	//StatePatches are diffs from the state before. Get the actual state for a
	//version with State().
	StatePatches []json.RawMessage
	Description  string `json:"omitempty"`
}

//New returns a new record with the data encoded in the file. If you want an
//empty record, just instantiate a blank struct. If a record with that
//filename has already been saved, it will return that record.
func New(filename string) (*Record, error) {

	if cachedRec := recCache[filename]; cachedRec != nil {
		return cachedRec, nil
	}

	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, errors.New("Couldn't read file: " + err.Error())
	}

	var storageRec storageRecord

	if err := json.Unmarshal(data, &storageRec); err != nil {
		return nil, errors.New("Couldn't decode json: " + err.Error())
	}

	return &Record{
		data: &storageRec,
	}, nil
}

func (r *Record) Game() *boardgame.GameStorageRecord {
	if r.data == nil {
		return nil
	}
	return r.data.Game
}

//Description returns the top-level description string set in the json file.
//There's no way to set this except by modifying the JSON serialization
//directly, but it can be read from record.
func (r *Record) Description() string {
	if r.data == nil {
		return ""
	}
	return r.data.Description
}

func (r *Record) Move(version int) (*boardgame.MoveStorageRecord, error) {
	if r.data == nil {
		return nil, errors.New("No data")
	}
	if version < 0 {
		return nil, errors.New("Version too low")
	}

	version -= 1
	//version is effectively 1-indexed, since we don't store a move for the
	//first version, but we store them in 0-indexed since we use the array
	//index. So convert to that.
	if len(r.data.Moves) <= version {
		return nil, errors.New("Not enough moves")
	}

	return r.data.Moves[version], nil
}

//randomString returns a random string of the given length.
func randomString(length int) string {
	var result = ""

	for len(result) < length {
		result += string(randomStringChars[rand.Intn(len(randomStringChars))])
	}

	return result
}

func safeOvewritefile(path string, blob []byte) error {

	//Check for the easy case where the file doesn't exist yet
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ioutil.WriteFile(path, blob, 0644)
	}

	dir, name := filepath.Split(path)

	ext := filepath.Ext(name)

	nameWithoutExt := strings.TrimSuffix(name, ext)

	tempFileName := filepath.Join(dir, nameWithoutExt+".TEMP."+randomString(6)+ext)

	if err := ioutil.WriteFile(tempFileName, blob, 0644); err != nil {
		return errors.New("Couldn't write temp file: " + err.Error())
	}

	if err := os.Remove(path); err != nil {
		return errors.New("Couldn't delete the original file: " + err.Error())
	}

	if err := os.Rename(tempFileName, path); err != nil {
		return errors.New("Couldn't rename the new file: " + err.Error())
	}

	return nil

}

//Save saves to the given path.
func (r *Record) Save(filename string) error {

	blob, err := json.MarshalIndent(r.data, "", "\t")

	if err != nil {
		return errors.New("Couldn't marshal blob: " + err.Error())
	}

	if err := safeOvewritefile(filename, blob); err != nil {
		return err
	}

	recCache[filename] = r

	return nil
}

//AddGameAndCurrentState adds the game, state, and move (if non-nil), ready
//for saving. Designed to be used in a SaveGameAndCurrentState method.
func (r *Record) AddGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {

	if r.data == nil {
		r.data = &storageRecord{}
	}

	r.data.Game = game

	if move != nil {
		r.data.Moves = append(r.data.Moves, move)
	}

	lastState, err := r.State(len(r.data.StatePatches) - 1)

	if err != nil {
		return errors.New("Couldn't fetch last state: " + err.Error())
	}

	differ := gojsondiff.New()

	patch, err := differ.Compare(lastState, state)

	if err != nil {
		return err
	}

	f := formatter.NewDeltaFormatter()

	js, err := f.FormatAsJson(patch)

	if err != nil {
		return errors.New("Couldn't format patch as json: " + err.Error())
	}

	formattedPatch, err := json.Marshal(js)

	if err != nil {
		return errors.New("Couldn't format patch json to byte: " + err.Error())
	}

	if err := confirmPatch(lastState, state, formattedPatch); err != nil {
		return errors.New("Sanity check failed: patch did not do what it should: " + err.Error())
	}

	r.states = append(r.states, state)
	r.data.StatePatches = append(r.data.StatePatches, formattedPatch)

	return nil

}

//confirmPatch does a sanity check by ensuring that applying the formatted
//patch to before would give you after. Helps ensure that there aren't
//unexpected bugs in the diffing library (which has been known to happen with
//these kinds of things).
func confirmPatch(before, after, formattedPatch []byte) error {

	var inflatedBefore map[string]interface{}
	if err := json.Unmarshal(before, &inflatedBefore); err != nil {
		return errors.New("Couldn't unmarshal before blob: " + err.Error())
	}

	var inflatedAfter map[string]interface{}
	if err := json.Unmarshal(after, &inflatedAfter); err != nil {
		return errors.New("Couldn't unmarshal before blob: " + err.Error())
	}

	unmarshaller := gojsondiff.NewUnmarshaller()

	reinflatedPatch, err := unmarshaller.UnmarshalBytes(formattedPatch)
	if err != nil {
		return errors.New("Couldn't reinflate patch: " + err.Error())
	}

	differ := gojsondiff.New()
	differ.ApplyPatch(inflatedBefore, reinflatedPatch)

	if diff := deep.Equal(inflatedBefore, inflatedAfter); len(diff) > 0 {
		return errors.New("Patched before did not equal after: " + strings.Join(diff, "\n"))
	}

	return nil
}

//State fetches the State object at that version. It can return an error
//because under the covers it has to apply serialized patches.
func (r *Record) State(version int) (boardgame.StateStorageRecord, error) {

	if r.data == nil {
		return nil, errors.New("No data")
	}

	if version < 0 {
		//The base object that version 0 is diffed against is the empty object
		return boardgame.StateStorageRecord(`{}`), nil
	}

	if len(r.states) > version {
		return r.states[version], nil
	}

	//Otherwise, derive forward, recursively.

	lastStateBlob, err := r.State(version - 1)

	if err != nil {
		//Don't decorate the error because it will likely stack
		return nil, err
	}

	unmarshaller := gojsondiff.NewUnmarshaller()

	patch, err := unmarshaller.UnmarshalBytes(r.data.StatePatches[version])

	if err != nil {
		return nil, err
	}

	differ := gojsondiff.New()

	var state map[string]interface{}

	if err := json.Unmarshal(lastStateBlob, &state); err != nil {
		return nil, errors.New("Couldn't unmarshal last blob: " + err.Error())
	}

	differ.ApplyPatch(state, patch)

	blob, err := json.MarshalIndent(state, "", "\t")

	if err != nil {
		return nil, errors.New("Couldn't marshal modified blob: " + err.Error())
	}

	r.states = append(r.states, blob)

	return blob, nil

}
