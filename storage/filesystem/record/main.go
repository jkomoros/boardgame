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
	"fmt"
	"github.com/jkomoros/boardgame"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const randomStringChars = "ABCDEF0123456789"

var recCache map[string]*Record

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	recCache = make(map[string]*Record, 16)
}

//FullPatchSanityCheck is whether we ensure that each state patch not only
//matches our current encoder, but DOESN'T match other encoders. More
//expensive, should only be set in testing scenarios.
var FullPatchSanityCheck bool

//StateEncoding captures how the states are encoded within the file.
type StateEncoding int

const (
	//The entirety of each state is encoded in StatePatch, with no diffing.
	//The most robust, but least efficient encoding. Also makes it hard to
	//eyeball the state patches to see what changed from state to state. This
	//is the encoding we return when there isn't yet enough information
	//encoded to determine the encoding (e.g. no second state encoded)
	StateEncodingFull StateEncoding = iota
	//Patches are encoded with delta format from github.com/yudai/gojsondiff.
	//This format is easy to represent in json, and is compatible with the
	//popular github.com/benjamine/jsondiffpatch (in javascript). However, it
	//handles at least some cases incorrectly.
	StateEncodingYudai
	//Patches are encoded using the jd format transformed to json, where each
	//line is represented as an item in an array of JSON strings.
	StateEncodingJosephBurnett

	//When adding a new one here, change the loop condition in
	//fullSanityCheckPatchEncoding.
)

func (s StateEncoding) encoder() encoder {
	switch s {
	case StateEncodingYudai:
		return &yudaiEncoder{}
	case StateEncodingFull:
		return &fullEncoder{}
	default:
		return nil
	}
}

//The encoding that new records should have.
var DefaultStateEncoding StateEncoding = StateEncodingYudai

//Record is a record of moves, states, and game. Get a new one based on the
//contents of a file with New(). If you want a new blank one using the default
//encoding use Empty().
type Record struct {
	data   *storageRecord
	states []boardgame.StateStorageRecord
	//The StateEncoding we've figure out our patches are represented as
	detectedEncoding StateEncoding
	//Whether or not we have figured out our encoding. Helps detect if the
	//zero value of detectedEncoding means that we've affirmatively detected
	//that encoding or haven'et yet.
	encodingDetected bool
}

type storageRecord struct {
	Game  *boardgame.GameStorageRecord
	Moves []*boardgame.MoveStorageRecord
	//StatePatches are diffs from the state before. Get the actual state for a
	//version with State().
	StatePatches []json.RawMessage
	Description  string `json:"omitempty"`
}

//encoder is the thing that actually does the encoding
type encoder interface {
	//CreatePatch returns the patch object to save. Doesn't have to confirm;
	//we'll call that automatically.
	CreatePatch(lastState, state boardgame.StateStorageRecord) ([]byte, error)
	//ConfirmPatch verifies the patch it returned will create state given
	//lastState + patch. confirmPatch does a sanity check by ensuring that
	//applying the formatted patch to before would give you after. Helps
	//ensure that there aren't unexpected bugs in the diffing library (which
	//has been known to happen with these kinds of things).
	ConfirmPatch(lastState, state, patch []byte) error
	//ApplyPatch takes a previous state and the patch and returns the new state.
	ApplyPatch(lastState, patch []byte) (boardgame.StateStorageRecord, error)
	//Matches should return nil if the patch is not in a format we accept, or
	//a descriptive error otherwise.
	Matches(examplePatch []byte) error
}

//Empty returns an empty record initialized to use the DefaultStateEncoding
//provided.
func Empty() *Record {
	return &Record{
		detectedEncoding: DefaultStateEncoding,
		encodingDetected: true,
	}
}

//New returns a new record with the data encoded in the file. If you want an
//empty record, use Empty(). If a record with that
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

func (r *Record) encoder() encoder {
	//TODO: figure out our encoding based on contents if not set.

	if !r.encodingDetected {
		r.detectedEncoding = StateEncodingYudai
		r.encodingDetected = true
	}

	enc := r.detectedEncoding.encoder()

	if enc == nil {
		panic("Unsupported encoder")
	}

	return enc
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

	enc := r.encoder()

	patch, err := enc.CreatePatch(lastState, state)

	if err != nil {
		return errors.New("Couldn't create patch: " + err.Error())
	}

	if err := enc.ConfirmPatch(lastState, state, patch); err != nil {

		fmt.Println("UNEXPECTED ERROR IN UNDERLYING LIBRARY:")
		fmt.Println("LastState:")
		fmt.Println(string(lastState))
		fmt.Println("\nState:")
		fmt.Println(string(state))
		fmt.Println("\nFormatted Patch:")
		fmt.Println(string(patch))

		return errors.New("Sanity check failed: patch did not do what it should: " + err.Error())
	}

	r.states = append(r.states, state)
	r.data.StatePatches = append(r.data.StatePatches, patch)

	return nil

}

func (r *Record) fullSanityCheckPatchEncoding(patch []byte) error {

	//This should set detectedEncoding
	enc := r.encoder()

	var i StateEncoding

	for i = 0; i <= StateEncodingJosephBurnett; i++ {
		//we want to make sure we try all of the other ones and that they fail first.
		if i == r.detectedEncoding {
			continue
		}

		testEncoding := i.encoder()

		if testEncoding == nil {
			continue
		}

		if err := testEncoding.Matches(patch); err == nil {
			return errors.New("Encoder that was not us matched when we expected it not to:" + strconv.Itoa(int(i)))
		}

	}

	if err := enc.Matches(patch); err != nil {
		return errors.New("The encoder we thought we were didn't mtach: " + err.Error())
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

	enc := r.encoder()

	patch := r.data.StatePatches[version]

	if FullPatchSanityCheck {
		if err := r.fullSanityCheckPatchEncoding(patch); err != nil {
			return nil, errors.New("Full patch sanity check failed: " + err.Error())
		}
	} else {
		//Sanity check the patch is a format we expect
		if err := enc.Matches(patch); err != nil {
			return nil, errors.New("Unexpected error: Sanity check failed: the stored patch does not appear to be in the format this encoder expects: " + err.Error())
		}
	}

	blob, err := enc.ApplyPatch(lastStateBlob, patch)

	if err != nil {
		return nil, errors.New("Couldn't apply patch: " + err.Error())
	}

	r.states = append(r.states, blob)

	return blob, nil

}
