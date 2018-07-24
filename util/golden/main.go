/*

	golden is a package designed to make it possible to compare a game to a
	golden run for testing purposes. It takes a record saved in
	storage/filesystem format and compares it.

*/
package golden

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/filesystem/record"
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
)

//Compare is the primary method in the package. It takes a game delegate and a
//filename denoting a record to compare against. delegate shiould be a fresh
//delegate not yet affiliated with a manager.
func Compare(delegate boardgame.GameDelegate, recFilename string) error {

	manager, err := boardgame.NewGameManager(delegate, memory.NewStorageManager())

	if err != nil {
		return errors.New("Couldn't create new manager: " + err.Error())
	}

	rec, err := record.New(recFilename)

	if err != nil {
		return errors.New("Couldn't create record: " + err.Error())
	}

	return compare(manager, rec)

}

//CompareFolder is like Compare, except it will iterate through any file in
//recFolder that ends in .json. Errors if any of those files cannot be parsed
//into recs, or if no files match.
func CompareFolder(delegate boardgame.GameDelegate, recFolder string) error {
	manager, err := boardgame.NewGameManager(delegate, memory.NewStorageManager())

	if err != nil {
		return errors.New("Couldn't create new manager: " + err.Error())
	}

	infos, err := ioutil.ReadDir(recFolder)

	if err != nil {
		return errors.New("Couldn't read folder: " + err.Error())
	}

	processedRecs := 0

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

		if err := compare(manager, rec); err != nil {
			return errors.New("File named " + info.Name() + " had compare error: " + err.Error())
		}

		processedRecs++
	}

	if processedRecs < 1 {
		return errors.New("Processed 0 recs in folder")
	}

	return nil
}

func compare(manager *boardgame.GameManager, rec *record.Record) error {
	game, err := manager.RecreateGame(rec.Game())

	if err != nil {
		return errors.New("Couldn't create game: " + err.Error())
	}

	lastVerifiedVersion := 0

	for !game.Finished() {
		//Verify all new moves that have happened since the last time we
		//checked (often, fix-up moves).
		for lastVerifiedVersion < game.Version() {
			stateToCompare, err := rec.State(lastVerifiedVersion)

			if err != nil {
				return errors.New("Couldn't get " + strconv.Itoa(lastVerifiedVersion) + " state: " + err.Error())
			}

			if err := compareJsonBlobs(game.State(lastVerifiedVersion).StorageRecord(), stateToCompare); err != nil {
				return errors.New("State " + strconv.Itoa(lastVerifiedVersion) + " compared differently: " + err.Error())
			}

			if lastVerifiedVersion > 0 {

				//Version 0 has no associated move

				recMove, err := rec.Move(lastVerifiedVersion)

				if err != nil {
					return errors.New("Couldn't get move " + strconv.Itoa(lastVerifiedVersion) + " from record")
				}

				moves := game.MoveRecords(lastVerifiedVersion)

				if len(moves) < 1 {
					return errors.New("Didn't fetch historical move records for " + strconv.Itoa(lastVerifiedVersion))
				}

				//Warning: records are modified by this method
				if err := compareMoveStorageRecords(moves[len(moves)-1], recMove); err != nil {
					return errors.New("Move " + strconv.Itoa(lastVerifiedVersion) + " compared differently: " + err.Error())
				}
			}

			lastVerifiedVersion++
		}

		nextMoveRec, err := rec.Move(lastVerifiedVersion + 1)

		if err != nil {
			//We'll assume that menas that's all of the moves there are to make.
			break
		}

		if nextMoveRec.Proposer < 0 {
			return errors.New("At version " + strconv.Itoa(lastVerifiedVersion) + " the next player move to apply was not applied by a player")
		}

		nextMove, err := nextMoveRec.Inflate(game)

		if err != nil {
			return errors.New("Couldn't inflate move: " + err.Error())
		}

		if err := <-game.ProposeMove(nextMove, nextMoveRec.Proposer); err != nil {
			return errors.New("Couldn't propose next move in chain: " + err.Error())
		}

	}

	if game.Finished() != rec.Game().Finished {
		return errors.New("Game finished did not match rec")
	}

	if !reflect.DeepEqual(game.Winners(), rec.Game().Winners) {
		return errors.New("Game winners did not match")
	}

	return nil
}

var differ = gojsondiff.New()

var diffformatter = formatter.NewDeltaFormatter()

func compareJsonBlobs(one, two []byte) error {

	diff, err := differ.Compare(one, two)

	if err != nil {
		return errors.New("Couldn't diff: " + err.Error())
	}

	if diff.Modified() {

		str, err := diffformatter.Format(diff)

		if err != nil {
			return errors.New("Couldn't format diff: " + err.Error())
		}

		return errors.New("Diff: " + str)
	}

	return nil

}

//warning: modifies the records
func compareMoveStorageRecords(one, two *boardgame.MoveStorageRecord) error {

	if one == nil {
		return errors.New("One was nil")
	}

	if two == nil {
		return errors.New("Two was nil")
	}

	oneBlob := one.Blob
	twoBlob := two.Blob

	//Set the fields we know might differ to known values
	one.Blob = nil
	two.Blob = nil

	two.Timestamp = one.Timestamp

	if !reflect.DeepEqual(one, two) {
		return errors.New("Move storage records differed in base fields")
	}

	return compareJsonBlobs(oneBlob, twoBlob)

}
