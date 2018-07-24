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
	"reflect"
	"strconv"
)

//Compare is the primary method in the package. It takes a fresh game delegate
//and a record to compare against. delegate shiould be a fresh delegate not
//yet affiliated with a manager.s
func Compare(delegate boardgame.GameDelegate, rec *record.Record) error {

	manager, err := boardgame.NewGameManager(delegate, memory.NewStorageManager())

	if err != nil {
		return errors.New("Couldn't create new manager: " + err.Error())
	}

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

			//TODO: use go-test/deep (if vendored) for a more descriptive error.
			if err := compareStorageRecords(game.State(lastVerifiedVersion).StorageRecord(), stateToCompare); err != nil {
				return errors.New("State " + strconv.Itoa(lastVerifiedVersion) + " compared differently: " + err.Error())
			}

			//TODO: compare the move storage records too.

			lastVerifiedVersion++
		}

		nextMoveRec, err := rec.Move(lastVerifiedVersion + 1)

		if err != nil {
			//We'll assume that menas that's all of the moves there are to make.
			continue
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

func compareStorageRecords(one, two boardgame.StateStorageRecord) error {

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
