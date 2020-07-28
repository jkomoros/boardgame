/*

Package golden is a package designed to make it possible to compare a game to a
golden run for testing purposes. It takes a record saved in storage/filesystem
format and compares it.

Typical Use

Typically you generate a new set of goldens by cd'ing into the package
containing the game you want to test. Then you run `boardgame-util
create-golden`, which generates a stub server for that game. As you play and
create games, it will save the records as filesystem.record.Record, a format
that is designed to be easy to read and modify as JSON. `create-golden` also
will save a `golden_test.go` which will, when tests are run, compare the current
GameManager's operation to the states and moves saved in the golden json blobs.
These tests are a great way to verify that the behavior of your game does not
accidentlaly change.

The format of filesystem.record.Record is optimized to be able to be
hand-edited. It does a number of tricks to make sure that states and moves can
be spliced in easily. First and foremost, it typically stores subsequent states
not as full blobs but as diffs from the state before. This means that changing
one state doesn't require also modifying all subsequent states to have the same
values. The format also "relativizes" moves, setting their Version to -1,
signifying that when fetched via record.Move(version), it should just use the
version number it said it was. In addition, the Initiator field is stored as a
relative number for how many moves back in history to go to. Finally, the
Timestamp field is stored in a format that is as many seconds past the Unix
epoch as the move is from the Game's Created timestamp. All of these properties
mean that the format is (relatively) easy to tweak manually to add or remove
moves.

You can also add a "Description" top-level field in the json to describe what
the game is testing, which is useful to keep track of goldens that test various
edge cases.

Remastering Goldens

Typically you record a golden, and then every time you test the game package, it
will just verify the game logic still applies the same moves in the right order,
with the right state modifications. But every so often you want to 'remaster'
your golden. For example, perhaps the game logic has changed to have slightly
different behavior, or you want to update the format of the golden to ensure
it's canonical and up-to-date, to match changes in the underlying library.

It's possible to 'remaster' goldens, which means to re-record them and overwrite
the original. You do this by passing true as the last value to Compare or
CompareFolder. The `golden_test.go` that is generated for you will also
automatically add a flag that will be passed in, so the canonical way to
remaster is to run `go test -update-golden`, which instead of comparing the
golden, will remaster it, overwriting the original.

After you remaster, it's a good idea to sanity check by doing a normal test (`go
test`) to verify the game logic matches the new golden. You should also visually
inspect the diff of the golden before commiting to make sure there aren't any
unintended changes. The remastering process is designed to ensure that wherever
possible content doesn't change. The design of filesystem.record.Record
(described above) helps, but the remastering process also seeks to, for example,
use existing timestamps wherever possible and generate reasonable intermediate
timestamps for new moves that have been added.

*/
package golden

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-test/deep"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/filesystem/record"
	"github.com/sirupsen/logrus"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

//Compare is the primary method in the package. It takes a game delegate and a
//filename denoting a record to compare against. delegate shiould be a fresh
//delegate not yet affiliated with a manager. It compares every version and move
//in the history (ignoring things that shouldn't be the same, like timestamps)
//and reports the first place they divrge. Any time it finds a move not proposed
//by AdminPlayerIndex it will propose that move. As long as your game uses
//state.Rand() for all randomness and is otherwise deterministic then everything
//should work. If updateOnDifferent is true, instead of erroring, it will
//instead overwrite the existing golden with a new one. The boardgame-util
//create-goldens tool will output a test that will look for a `-update-golden`
//flag and pass in that variable here.
func Compare(delegate boardgame.GameDelegate, recFilename string, updateOnDifferent bool) error {

	storage := newStorageManager()

	manager, err := boardgame.NewGameManager(delegate, storage)

	if err != nil {
		return errors.New("Couldn't create new manager: " + err.Error())
	}

	storage.manager = manager

	rec, err := record.New(recFilename)

	if err != nil {
		return errors.New("Couldn't create record: " + err.Error())
	}

	return compare(manager, rec, storage, updateOnDifferent)

}

//CompareFolder is like Compare, except it will iterate through any file in
//recFolder that ends in .json. Errors if any of those files cannot be parsed
//into recs. See Compare for more documentation.
func CompareFolder(delegate boardgame.GameDelegate, recFolder string, updateOnDifferent bool) error {

	storage := newStorageManager()

	manager, err := boardgame.NewGameManager(delegate, storage)

	if err != nil {
		return errors.New("Couldn't create new manager: " + err.Error())
	}

	storage.manager = manager

	infos, err := ioutil.ReadDir(recFolder)

	if err != nil {
		return errors.New("Couldn't read folder: " + err.Error())
	}

	for _, info := range infos {
		if info.IsDir() {
			continue
		}

		if filepath.Ext(info.Name()) != ".json" {
			continue
		}

		rec, err := record.New(filepath.Join(recFolder, info.Name()))

		if err != nil {
			return errors.New("File with name " + info.Name() + " couldn't be loaded into rec: " + err.Error())
		}

		if err := compare(manager, rec, storage, updateOnDifferent); err != nil {
			return errors.New("File named " + info.Name() + " had compare error: " + err.Error())
		}

	}

	return nil
}

func newLogger() (*logrus.Logger, *bytes.Buffer) {
	result := logrus.New()
	buf := &bytes.Buffer{}
	result.Out = buf
	result.SetLevel(logrus.DebugLevel)
	return result, buf
}

func compare(manager *boardgame.GameManager, rec *record.Record, storage *storageManager, updateOnDifferent bool) error {

	//TODO: get rid of this function once refactored
	comparer, err := newComparer(manager, rec, storage)

	if err != nil {
		return errors.New("Couldn't create comparer: " + err.Error())
	}

	if updateOnDifferent {
		fmt.Println("WARNING: overwriting old goldens, verify the diff looks sane before committing!")
		newGolden, err := comparer.RegenerateGolden()
		if err != nil {
			comparer.PrintDebug()
			return err
		}
		if err := newGolden.Save(rec.Path(), false); err != nil {
			return errors.New("Could not overwrite " + rec.Path() + ": " + err.Error())
		}
	} else {

		if err := comparer.Compare(); err != nil {
			comparer.PrintDebug()
			return err
		}
	}

	return nil
}

var differ = gojsondiff.New()

func compareJSONBlobs(one, two []byte) error {

	diff, err := differ.Compare(one, two)

	if err != nil {
		return errors.New("Couldn't diff: " + err.Error())
	}

	if diff.Modified() {

		var oneJSON map[string]interface{}

		if err := json.Unmarshal(one, &oneJSON); err != nil {
			return errors.New("Couldn't unmarshal left")
		}

		diffformatter := formatter.NewAsciiFormatter(oneJSON, formatter.AsciiFormatterConfig{
			Coloring: true,
		})

		str, err := diffformatter.Format(diff)

		if err != nil {
			return errors.New("Couldn't format diff: " + err.Error())
		}

		return errors.New("Diff: " + str)
	}

	return nil

}

func compareMoveStorageRecords(one, two boardgame.MoveStorageRecord, skipAbsoluteVersions bool) error {

	oneBlob := one.Blob
	twoBlob := two.Blob

	//Set the fields we know might differ to known values
	one.Blob = nil
	two.Blob = nil

	two.Timestamp = one.Timestamp

	if skipAbsoluteVersions {
		if one.Version >= 0 || two.Version >= 0 {
			two.Version = one.Version
			two.Initiator = one.Initiator
		}
	}

	if !reflect.DeepEqual(one, two) {
		return errors.New("Move storage records differed in base fields: " + strings.Join(deep.Equal(one, two), ", "))
	}

	return compareJSONBlobs(oneBlob, twoBlob)

}
