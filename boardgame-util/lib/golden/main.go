/*

Package golden is a package designed to make it possible to compare a game to a
golden run for testing purposes. It takes a record saved in storage/filesystem
format and compares it.

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
