/*

	record is a package to open, read, and save game records stored in
	filesystem's format.

*/
package record

import (
	"encoding/json"
	"errors"
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

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

//Record is a record of moves, states, and game. Get a new one based on the
//contents of a file with New(). Instantiate directly if you'll set the
//starter state yourself.
type Record struct {
	Game  *boardgame.GameStorageRecord
	Moves []*boardgame.MoveStorageRecord
	//StatePatches are diffs from the state before. Get the actual state for a
	//version with State().
	StatePatches   []json.RawMessage
	states         []json.RawMessage
	expandedStates []map[string]interface{}
}

func New(filename string) (*Record, error) {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, errors.New("Couldn't read file: " + err.Error())
	}

	var result Record

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, errors.New("Couldn't decode json: " + err.Error())
	}

	return &result, nil
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

	blob, err := json.MarshalIndent(r, "", "\t")

	if err != nil {
		return errors.New("Couldn't marshal blob: " + err.Error())
	}

	return safeOvewritefile(filename, blob)
}

//AddGameAndCurrentState adds the game, state, and move (if non-nil), ready
//for saving. Designed to be used in a SaveGameAndCurrentState method.
func (r *Record) AddGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {

	r.Game = game

	if move != nil {
		r.Moves = append(r.Moves, move)
	}

	lastState, err := r.State(len(r.StatePatches) - 1)

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

	r.states = append(r.states, json.RawMessage(state))
	r.StatePatches = append(r.StatePatches, formattedPatch)

	return nil

}

//State is a method, not just a normal array, because we actually serialize to
//disk state patches, given that states are big objects and they rarely
//change. So when you request a State we have to derive what it is by applying
//all of the patches up in order.
func (r *Record) State(version int) (json.RawMessage, error) {

	if version < 0 {
		//The base object that version 0 is diffed against is the empty object
		return json.RawMessage(`{}`), nil
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

	patch, err := unmarshaller.UnmarshalBytes(r.StatePatches[version])

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
