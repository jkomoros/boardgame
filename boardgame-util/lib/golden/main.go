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
	"strconv"
	"strings"

	"github.com/go-test/deep"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/storage/filesystem"
	"github.com/jkomoros/boardgame/storage/filesystem/record"
	"github.com/sirupsen/logrus"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

//Note: these are also duplicated in moves/seat_player.go and server/api/storage.go
const playerToSeatRendevousDataType = "github.com/jkomoros/boardgame/server/api.PlayerToSeat"
const willSeatPlayerRendevousDataType = "github.com/jkomoros/boardgame/server/api.WillSeatPlayer"

//by defining the variable type, we verify we actually do implement the
//interface. Since it flows via FetchInejctedData, there's no type
//checking otherwise.
var testPlayerSeat interfaces.SeatPlayerSignaler = &player{}

type player struct {
	index boardgame.PlayerIndex
	s     *storageManager
}

func (p *player) SeatIndex() boardgame.PlayerIndex {
	return p.index
}

func (p *player) Committed() {
	p.s.playerToSeat = nil
}

type storageManager struct {
	*filesystem.StorageManager
	manager      *boardgame.GameManager
	playerToSeat *player
	//A cache of whether a given gameID will call seatPlayer.
	memoizedGameWillSeatPlayer map[string]bool
	gameRecords                map[string]*record.Record
}

func (s *storageManager) gameWillSeatPlayer(gameID string) bool {
	if result, ok := s.memoizedGameWillSeatPlayer[gameID]; ok {
		return result
	}
	if s.gameRecords[gameID] == nil {
		panic("Game record didn't exist for " + gameID)
	}

	rec := s.gameRecords[gameID]

	foundSeatPlayerMove := false

	for i := 1; i <= rec.Game().Version; i++ {
		moveRec, err := rec.Move(i)
		if err != nil {
			panic("Couldn't get move " + strconv.Itoa(i) + ": " + err.Error())
		}
		exampleMove := s.manager.ExampleMoveByName(moveRec.Name)
		if seatPlayerer, ok := exampleMove.(interfaces.SeatPlayerMover); ok {
			if seatPlayerer.IsSeatPlayerMove() {
				foundSeatPlayerMove = true
				break
			}
		}
	}

	s.memoizedGameWillSeatPlayer[gameID] = foundSeatPlayerMove
	return foundSeatPlayerMove

}

func (s *storageManager) FetchInjectedDataForGame(gameID string, dataType string) interface{} {
	if dataType == willSeatPlayerRendevousDataType {
		//This data type should return anything non-nil to signal, yes, I am a
		//context that will pass you SeatPlayers when there's a player to seat.

		//Only games that do have a SeatPlayer in their golden should return
		//true.
		return s.gameWillSeatPlayer(gameID)
	}
	if dataType == playerToSeatRendevousDataType {
		if s.playerToSeat == nil {
			//Return an untyped nil
			return nil
		}
		return s.playerToSeat
	}
	return s.StorageManager.FetchInjectedDataForGame(gameID, dataType)
}

//injectPlayerToSeat is how you make StorageManager tell the SeatPlayer move to
//seat the player at the given index. You also need to call ForceFixUp after
//caling this.
func (s *storageManager) injectPlayerToSeat(index boardgame.PlayerIndex) {
	s.playerToSeat = &player{
		index,
		s,
	}
}

func newStorageManager() *storageManager {
	fsStorage := filesystem.NewStorageManager("")
	fsStorage.DebugNoDisk = true
	return &storageManager{
		fsStorage,
		nil,
		nil,
		make(map[string]bool),
		make(map[string]*record.Record),
	}
}

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

	return compare(manager, rec, storage)

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

		if err := compare(manager, rec, storage); err != nil {
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

//comparer is a struct that manages the process of comparing a given golden to a
//new game. Typically you create one wiht newComparer, and then call Compare.
type comparer struct {
	manager             *boardgame.GameManager
	golden              *record.Record
	storage             *storageManager
	game                *boardgame.Game
	buf                 *bytes.Buffer
	lastBuf             *bytes.Buffer
	lastVerifiedVersion int
}

func newComparer(manager *boardgame.GameManager, rec *record.Record, storage *storageManager) (*comparer, error) {

	game, err := manager.Internals().RecreateGame(rec.Game())

	if err != nil {
		return nil, errors.New("Couldn't create game: " + err.Error())
	}

	result := &comparer{
		manager,
		rec,
		storage,
		game,
		nil,
		nil,
		0,
	}

	//FetchInjectedDataForGame requires a reference to the game
	storage.gameRecords[rec.Game().ID] = rec

	return result, nil
}

func (c *comparer) PrintDebug() {

	if c.lastBuf == nil {
		return
	}

	if c.lastBuf.Len() == 0 {
		return
	}
	fmt.Println(c.lastBuf)

	fmt.Println("Last state: ")
	state := c.game.State(c.lastVerifiedVersion)
	jsonBuf, _ := json.MarshalIndent(state, "", "\t")
	fmt.Println(string(jsonBuf))
}

//ResetDebugLog should be called at the top of the loop, so that PrintDebug()
//will only print the last turn through the game's worth of logs.
func (c *comparer) ResetDebugLog() {
	c.lastBuf = c.buf
	logger, buf := newLogger()
	c.manager.SetLogger(logger)
	c.buf = buf
}

func (c *comparer) LastVerifiedVersion() int {
	return c.lastVerifiedVersion
}

//VerifyUnverifiedMoves compares moves and states for moves that have been
//applied but not yet verified. Even if it errors, it may have incremented
//LastVerifiedVersion(). If skipComparingStates is true, then the state won't be
//compared, only the move. That's useful in cases where you aren't trying to
//actually verify the golden, just skip moves.
func (c *comparer) VerifyUnverifiedMoves(skipComparingStates bool) error {
	for c.lastVerifiedVersion < c.game.Version() {
		stateToCompare, err := c.golden.State(c.lastVerifiedVersion)

		if err != nil {
			return errors.New("Couldn't get " + strconv.Itoa(c.lastVerifiedVersion) + " state: " + err.Error())
		}

		//We used to just do
		//game.State(lastVerifiedVersion).StorageRecord(), but that
		//doesn't guarantee that it's the same as
		//manager.Storage().State() because it relies on state in
		//timerManager. This is unexpected and needs its own issue.
		storageRec, err := c.manager.Storage().State(c.game.ID(), c.lastVerifiedVersion)

		if err != nil {
			return errors.New("Couldn't get state storage rec from game: " + err.Error())
		}

		//Compare move first, because if the state doesn't match, it's
		//important to know first if the wrong move was applied or if the
		//state was wrong.

		if c.lastVerifiedVersion > 0 {

			//Version 0 has no associated move

			recMove, err := c.golden.Move(c.lastVerifiedVersion)

			if err != nil {
				return errors.New("Couldn't get move " + strconv.Itoa(c.lastVerifiedVersion) + " from record")
			}

			moves := c.game.MoveRecords(c.lastVerifiedVersion)

			if len(moves) < 1 {
				return errors.New("Didn't fetch historical move records for " + strconv.Itoa(c.lastVerifiedVersion))
			}

			//Warning: records are modified by this method
			if err := compareMoveStorageRecords(moves[len(moves)-1], recMove); err != nil {
				return errors.New("Move " + strconv.Itoa(c.lastVerifiedVersion) + " compared differently: " + err.Error())
			}
		}

		if !skipComparingStates {
			if err := compareJSONBlobs(storageRec, stateToCompare); err != nil {
				return errors.New("State " + strconv.Itoa(c.lastVerifiedVersion) + " compared differently: " + err.Error())
			}
		}

		c.lastVerifiedVersion++
	}
	return nil
}

//ApplyNextMove will apply the next move, whether it's a player move or a
//special admin move.
func (c *comparer) ApplyNextMove() (bool, error) {
	applied, err := c.ApplyNextSpecialAdminMove()
	if err != nil {
		return false, errors.New("Couldn't apply next special admin move: " + err.Error())
	}
	if applied {
		return true, nil
	}
	applied, err = c.ApplyNextPlayerMove()
	if err != nil {
		return false, errors.New("Couldn't apply next player move: " + err.Error())
	}
	return applied, nil
}

//ApplyNextPlayerMove applies the next player move in the sequence.
func (c *comparer) ApplyNextPlayerMove() (bool, error) {

	nextMoveRec, err := c.golden.Move(c.lastVerifiedVersion + 1)

	if err != nil {
		//We'll assume that menas that's all of the moves there are to make.
		return false, nil
	}

	if nextMoveRec.Proposer < 0 {
		//The next move was not a PlayerMove
		return false, nil
	}

	nextMove, err := c.manager.Internals().InflateMoveStorageRecord(nextMoveRec, c.game)

	if err != nil {
		return false, errors.New("Couldn't inflate move: " + err.Error())
	}

	if err := <-c.game.ProposeMove(nextMove, nextMoveRec.Proposer); err != nil {
		return false, errors.New("Couldn't propose next move in chain: " + err.Error())
	}

	return true, nil
}

//ApplyNextSpecialAdminMove will return the next special admin move, returning
//true if one was applied. Typically Admin (FixUp) moves are applied
//automatically by the engine after a PlayerMove is applied. But two special
//kinds of moves have to be handeled specially: 1) timers that are queued up but
//have not fired yet, and 2) SeatPlayer moves.
func (c *comparer) ApplyNextSpecialAdminMove() (bool, error) {
	nextMoveRec, err := c.golden.Move(c.lastVerifiedVersion + 1)

	if err != nil {
		//We'll assume that menas that's all of the moves there are to make.
		return false, nil
	}

	//There was no special move to apply
	if nextMoveRec.Proposer >= 0 {
		return false, nil
	}

	//The next move was applied by admin, but wasn't already applied.
	//That means it's either a SeatPlayer move, or a timer that fired.

	//First, check if the nextMoveRec is a type of move that is a Seat
	//Player move.

	exampleMove := c.manager.ExampleMoveByName(nextMoveRec.Name)

	if exampleMove == nil {
		return false, errors.New("Couldn't fetch move named: " + nextMoveRec.Name)
	}

	if isSeatPlayer, ok := exampleMove.(interfaces.SeatPlayerMover); ok && isSeatPlayer.IsSeatPlayerMove() {
		//It does seem to be a seat Player mover.
		index, err := exampleMove.Reader().PlayerIndexProp("TargetPlayerIndex")
		if err != nil {
			return false, errors.New("Couldn't get expected TargetPlayerIndex from next SeatPlayer: " + err.Error())
		}
		c.storage.injectPlayerToSeat(index)
		if err := <-c.manager.Internals().ForceFixUp(c.game); err != nil {
			return false, errors.New("Couldn't force inject a SeatPlayer move: " + err.Error())
		}
		return true, nil
	}

	//We could be waiting for a timer to fire.
	//If there was a timer, try to force it to fire early.
	if c.manager.Internals().ForceNextTimer() {
		return true, nil
	}

	//If we get to here, there's another admin move to apply that wasn't for
	//some reason. There must be a problem in the game logic.

	//There wasn't a timer pending, so it's just an error
	return false, errors.New("At version " + strconv.Itoa(c.lastVerifiedVersion) + " the next player move to apply was not applied by a player. This implies that the fixUp move named " + nextMoveRec.Name + " is erroneously returning an error from its Legal method.")
}

//RegenerateGolden returns a new golden that has applied the non-fixup moves
//like the golden that the comparer is based off of.
func (c *comparer) RegenerateGolden() (*record.Record, error) {
	return nil, errors.New("Not yet implemented")
}

func (c *comparer) Compare() error {
	for !c.game.Finished() {

		c.ResetDebugLog()

		//Verify all new moves that have happened since the last time we
		//checked (often, fix-up moves).
		if err := c.VerifyUnverifiedMoves(false); err != nil {
			return errors.New("VerifyUnverifiedMoves failed: " + err.Error())
		}

		applied, err := c.ApplyNextMove()
		if err != nil {
			return errors.New("Couldn't apply move: " + err.Error())
		}
		if !applied {
			break
		}
	}
	return c.CompareFinished()
}

func (c *comparer) CompareFinished() error {

	if c.game.Finished() != c.golden.Game().Finished {
		return errors.New("Game finished did not match rec")
	}

	if !reflect.DeepEqual(c.game.Winners(), c.golden.Game().Winners) {
		return errors.New("Game winners did not match")
	}

	return nil
}

func compare(manager *boardgame.GameManager, rec *record.Record, storage *storageManager) error {

	//TODO: get rid of this function once refactored
	comparer, err := newComparer(manager, rec, storage)

	if err != nil {
		return errors.New("Couldn't create comparer: " + err.Error())
	}

	if err := comparer.Compare(); err != nil {
		comparer.PrintDebug()
		return err
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
		return errors.New("Move storage records differed in base fields: " + strings.Join(deep.Equal(one, two), ", "))
	}

	return compareJSONBlobs(oneBlob, twoBlob)

}
