/*

	record is a package to open, read, and save game records stored in
	filesystem's format.

	We encode states as diffs by default, but if that's not possible (for
	example, the diff the diffing library gave did not give us a valid diff
	because applying it to the left input does not provide the right input)
	then we convert to a full encoding mode, which encodes the entirerty of
	the blobs. Every time we save we try to revert to a diffed encoding if
	possible. This allows these files to be relatively resilient to errors in
	the undelrying diff library and heal as that library improves.

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
	"github.com/go-test/deep"
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

//Record is a record of moves, states, and game. Get a new one based on the
//contents of a file with New(). If you want a new blank one you can just use
//a zero value of this.
type Record struct {
	data              *storageRecord
	states            []boardgame.StateStorageRecord
	fullStateEncoding bool
	//If true, we will never try to turn off full-state encoding
	preferFullStateEncoding bool
}

type storageRecord struct {
	Game  *boardgame.GameStorageRecord
	Moves []*boardgame.MoveStorageRecord
	//StatePatches are diffs from the state before. Get the actual state for a
	//version with State().
	StatePatches []json.RawMessage
	Description  string `json:",omitempty"`
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

//EmptyWithFullStateEncoding returns a record that will default to full state
//encoding and will never automatically try to reduce down to
//fullstateencoding. Primarily useful for testing.
func EmptyWithFullStateEncoding() *Record {
	return &Record{
		fullStateEncoding:       true,
		preferFullStateEncoding: true,
	}
}

//New returns a new record with the data encoded in the file. If you want one
//that does not yet have a file backing it, you can just use an empty value of
//Record. If a record with that filename has already been saved, it will
//return that record.
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

	result := &Record{
		data: &storageRec,
	}

	if len(storageRec.StatePatches) > 0 {
		if err := fullEncoder.Matches(storageRec.StatePatches[0]); err == nil {
			result.fullStateEncoding = true
		}
	}

	return result, nil
}

func (r *Record) encoder() encoder {
	if r.fullStateEncoding {
		return fullEncoder
	}
	return diffEncoder
}

//FullStateEncoding returns whether the record is using full state encoding
//instead of the default diff.
func (r *Record) FullStateEncoding() bool {
	return r.fullStateEncoding
}

//Compress converts from full state encoding to diff encoding, if possible.
//Noop if already diff encoded.
func (r *Record) Compress() error {

	if !r.FullStateEncoding() {
		return nil
	}

	if err := r.reencode(diffEncoder); err != nil {
		return err
	}

	r.fullStateEncoding = false

	return nil

}

//Expand converts from diff state encoding to full encoding, if possible.
//Noop if already full encoded.
func (r *Record) Expand() error {
	if r.FullStateEncoding() {
		return nil
	}

	if err := r.reencode(fullEncoder); err != nil {
		return err
	}

	r.fullStateEncoding = true
	return nil
}

//reencode converts the given contents to the new encoding, returning
//an error if that's not possible. You still need to re-save to disk if you
//want to save the new contents. If error is non nil, the contents of the
//record won't have been modified. If newEncoding is the same as current encoding,
//will be a no op.
func (r *Record) reencode(targetEncoder encoder) error {

	if r.data == nil {
		return errors.New("No data!")
	}

	if targetEncoder == nil {
		return errors.New("The target encoder doesn't exist")
	}

	var newStatePatches []json.RawMessage

	lastState := boardgame.StateStorageRecord(`{}`)

	for i := 0; i < len(r.data.StatePatches); i++ {
		state, err := r.State(i)

		if err != nil {
			return errors.New("Couldn't fetch state " + strconv.Itoa(i) + ": " + err.Error())
		}

		patch, err := targetEncoder.CreatePatch(lastState, state)

		if err != nil {
			return errors.New("Couldn't create patch for state " + strconv.Itoa(i) + ": " + err.Error())
		}

		if err := targetEncoder.ConfirmPatch(lastState, state, patch); err != nil {
			return errors.New("Created patch did not confirm for state " + strconv.Itoa(i) + ": " + err.Error())
		}

		newStatePatches = append(newStatePatches, patch)

		lastState = state

	}

	if len(newStatePatches) != len(r.data.StatePatches) {
		return errors.New("Unexpected error: after converting didn't have enough state patches")
	}

	r.data.StatePatches = newStatePatches

	return nil

}

//compare ensures that this and the other contain the same information
func (r *Record) compare(other *Record) error {

	if r.data == nil {
		return errors.New("Nil for us")
	}

	if other.data == nil {
		return errors.New("nil for them")
	}

	if diff := deep.Equal(r.data.Game, other.data.Game); len(diff) != 0 {
		return errors.New("Game was not the same: " + strings.Join(diff, "\n"))
	}

	if len(r.data.Moves) != len(other.data.Moves) {
		return errors.New("Length of moves doesn't match")
	}

	for i := 0; i < len(r.data.Moves); i++ {
		if diff := deep.Equal(r.data.Moves[i], other.data.Moves[i]); len(diff) != 0 {
			return errors.New("Move " + strconv.Itoa(i) + " was not the same: " + strings.Join(diff, "\n"))
		}
	}

	for i := 0; i < len(r.data.StatePatches); i++ {

		left, err := r.State(i)
		if err != nil {
			return errors.New("Couldn't get left state for version " + strconv.Itoa(i) + ": " + err.Error())
		}
		var leftContents map[string]interface{}
		if err := json.Unmarshal(left, &leftContents); err != nil {
			return errors.New("Couldn't inflate left contents for version " + strconv.Itoa(i) + ": " + err.Error())
		}

		right, err := other.State(i)
		if err != nil {
			return errors.New("Couldn't get right state for version " + strconv.Itoa(i) + ": " + err.Error())
		}
		var rightContents map[string]interface{}
		if err := json.Unmarshal(right, &rightContents); err != nil {
			return errors.New("Couldn't inflate right contents for version " + strconv.Itoa(i) + ": " + err.Error())
		}

		if diff := deep.Equal(leftContents, rightContents); len(diff) != 0 {
			return errors.New("Diff of version " + strconv.Itoa(i) + " was not nil: " + strings.Join(diff, "\n"))
		}

	}

	return nil

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

	version -= 1
	//version is effectively 1-indexed, since we don't store a move for the
	//first version, but we store them in 0-indexed since we use the array
	//index. So convert to that.

	if version < 0 {
		return nil, errors.New("Version too low")
	}

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

//Save saves to the given path. Always try to Compress() if possible. If
//fullEncodingErrors is true then we'll error if we can't compress, otherwise
//we'll be OK with saving a full encoded version.
func (r *Record) Save(filename string, fullEncodingErrors bool) error {

	if err := r.Compress(); err != nil {
		if fullEncodingErrors {
			return errors.New("The data would have been saved in expanded form but that would be an error: " + err.Error())
		}
		//If we get to here then it's OK that we failed to compress.
	}

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
//for saving. Designed to be used in a SaveGameAndCurrentState method. If the
//state cannot be succcesfully encoded as a diffed encoding (due to an
//underlying issue in the diffing library, for example, that gives an invalid
//diff) then this will automatically expand the record into a
//FullStateEncoding mode.
func (r *Record) AddGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {

	if r.data == nil {
		r.data = &storageRecord{}
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

		fmt.Println("UNEXPECTED ERROR IN UNDERLYING LIBRARY")
		fmt.Println("LastState:")
		fmt.Println(string(lastState))
		fmt.Println("\nState:")
		fmt.Println(string(state))
		fmt.Println("\nFormatted Patch:")
		fmt.Println(string(patch))
		fmt.Println("Trying to auto expand...")

		if r.FullStateEncoding() {
			//We're already fully encoded, this really shouldn't happen
			return errors.New("Sanity check failed: patch did not do what it should: " + err.Error())
		}

		//OK, we found an error in underlying data. Exapnd and try again.
		if err := r.Expand(); err != nil {
			return errors.New("Saving failed in compressed mode, but expanding didn't work: " + err.Error())
		}

		//Try again now that we're expanded
		return r.AddGameAndCurrentState(game, state, move)

	}

	//Now that we've failed and expanded, actually modify the various
	//datastrutures in ourself (otherwise we could have, for example, double
	//moves)
	r.data.Game = game

	if move != nil {
		r.data.Moves = append(r.data.Moves, move)
	}

	r.states = append(r.states, state)
	r.data.StatePatches = append(r.data.StatePatches, patch)

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

	//Sanity check the patch is a format we expect
	if err := enc.Matches(patch); err != nil {
		return nil, errors.New("Unexpected error: Sanity check failed: the stored patch does not appear to be in the format this encoder expects: " + err.Error())
	}

	blob, err := enc.ApplyPatch(lastStateBlob, patch)

	if err != nil {
		return nil, errors.New("Couldn't apply patch: " + err.Error())
	}

	r.states = append(r.states, blob)

	return blob, nil

}
