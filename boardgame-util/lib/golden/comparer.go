package golden

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/storage/filesystem/record"
)

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

	if existingGameRec, _ := storage.Game(rec.Game().ID); existingGameRec != nil {
		return nil, errors.New("The storage layer already has a game with that ID. Use a fresh storage manager")
	}

	//FetchInjectedDataForGame requires a reference to the game, and it has to
	//be there before any FixUp moves are applied, so before the game is
	//created.
	storage.gameRecords[rec.Game().ID] = rec

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

//GoldenHasRemainingMoves returns whether there are moves beyond what golden
//contains.
func (c *comparer) GoldenHasRemainingMoves() bool {
	return c.golden.Game().Version > c.lastVerifiedVersion
}

func (c *comparer) LastVerifiedVersion() int {
	return c.lastVerifiedVersion
}

//AdvanceToNextInitiatorMove is like VerifyUnverifiedMoves, but advances
//LastVerifiedVersion automatically up to the next move whose Initator is
//itself: that is, that is a Player move, a Timer move, or a SeatPlayer move.
//This leaves the golden in the state where the next move to make in the new
//version is the next unverified move.
func (c *comparer) AdvanceToNextInitiatorMove() {

	if c.game.Version() > 0 {
		//Always advance by at least one, unless we're the first version--if we
		//are, and the very first move is a player move, we'd skip it
		//erroneously. We check the game.Version() otherwise if we checked
		//lastVerifiedVersion it would never advance.
		c.lastVerifiedVersion++
	}

	//We can't use golden.Game().Version, because in the case of a record with
	//manual surgery to introduce another move, the Version won't have been
	//upped. Why +2? :shrug: empiricially that's what worked.
	for c.lastVerifiedVersion <= len(c.golden.RawMoves())+2 {
		//We want to return one BEFORE the next initator move.
		nextMoveRec, err := c.golden.Move(c.lastVerifiedVersion + 1)

		if err != nil {
			//Assume we're at the end of the game
			return
		}
		if nextMoveRec.Initiator == nextMoveRec.Version {
			//We found the next move that starts a chain
			return
		}
		c.lastVerifiedVersion++
	}
}

//VerifyUnverifiedMoves compares moves and states for moves that have been
//applied but not yet verified. Even if it errors, it may have incremented
//LastVerifiedVersion().
func (c *comparer) VerifyUnverifiedMoves() error {
	verifiedAtLeastOne := false
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
			if err := compareMoveStorageRecords(*moves[len(moves)-1], *recMove, false); err != nil {
				return errors.New("Move " + strconv.Itoa(c.lastVerifiedVersion) + " compared differently: " + err.Error())
			}
		}

		if err := compareJSONBlobs(storageRec, stateToCompare); err != nil {
			return errors.New("State " + strconv.Itoa(c.lastVerifiedVersion) + " compared differently: " + err.Error())
		}

		c.lastVerifiedVersion++
		verifiedAtLeastOne = true
	}
	if !verifiedAtLeastOne {
		return errors.New("VerifyUnverifiedMoves didn't verify any new moves; this implies that ApplyNextMove isn't actually applying the next move")
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
		return false, errors.New("Couldn't propose next move in chain: " + strconv.Itoa(nextMoveRec.Version) + ": " + err.Error())
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

	c.AdvanceToNextInitiatorMove()
	for c.GoldenHasRemainingMoves() {

		c.ResetDebugLog()

		applied, err := c.ApplyNextMove()
		if err != nil {
			return nil, errors.New("Couldnt apply next move: " + err.Error())
		}
		if !applied {
			return nil, errors.New("There were still moves left in golden to apply but they hadn't been triggered: " + strconv.Itoa(c.lastVerifiedVersion))
		}
		//We assume that the chain of fixups that are all applied--even if it
		//now has fewer or more items--is fine in this new world. We want to
		//fast-forward to the next non-initatied move in the golden to apply,
		//and then apply that.
		c.AdvanceToNextInitiatorMove()
	}

	newRecord, err := c.storage.RecordForID(c.golden.Game().ID)

	if err != nil {
		return nil, errors.New("Couldn't get RecordForID: " + err.Error())
	}

	newRecord.SetDescription(c.golden.Description())

	if err := alignTimes(newRecord, c.golden); err != nil {
		return nil, errors.New("Couldn't align times: " + err.Error())
	}

	return newRecord, nil

}

func (c *comparer) Compare() error {
	for !c.game.Finished() {

		c.ResetDebugLog()

		//Verify all new moves that have happened since the last time we
		//checked (often, fix-up moves).
		if err := c.VerifyUnverifiedMoves(); err != nil {
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

//alignTimes MODIFIES the given new storage record, specifically trying to
//minimize diffs between the two, at least due to timeStamps, as much as
//possible.
func alignTimes(new, golden *record.Record) error {
	new.Game().Created = golden.Game().Created
	new.Game().Modified = golden.Game().Modified
	//TODO: also align move times as much as possible!
	return alignMoveTimes(new.RawMoves(), golden.RawMoves())
}

func alignMoveTimes(new, golden []*boardgame.MoveStorageRecord) error {
	goldenIndex := 0

	//Guard against the case where there is no valid index into golden
	if len(golden) == 0 {
		return nil
	}

	//The first part is simple: look for analoges between new and golden and
	//copy the timestamp over from golden.

	//goldenIndexes keeps track of which index into golden it was copying from.
	//-1 signifies one that wasn't copied.
	goldenIndexes := make([]int, len(new))

	for newIndex := 0; newIndex < len(new); newIndex++ {
		//If the goldenIndex is valid (that is, hasn't advanced past the end of
		//golden) then see if the moves are equivalent, and if so copy the
		//timestamps.
		if goldenIndex < len(golden) {
			if err := compareMoveStorageRecords(*new[newIndex], *golden[goldenIndex], true); err == nil {
				//Match!
				new[newIndex].Timestamp = golden[goldenIndex].Timestamp
				goldenIndexes[newIndex] = goldenIndex
				goldenIndex++
				continue
			}
			doContinue := false
			//They didn't match. But scan ahead and see if there's a match--if
			//there is, that implies that those items were deleted.
			for tempGoldenIndex := goldenIndex; tempGoldenIndex < len(golden); tempGoldenIndex++ {
				if err := compareMoveStorageRecords(*new[newIndex], *golden[tempGoldenIndex], true); err == nil {
					//We found it by skipping ahead in our new index. TODO: this
					//logic could get confused by a singular move that appears
					//to match; ideally we'd scan forward and pick the match
					//that has the longest matching subset.
					goldenIndex = tempGoldenIndex
					new[newIndex].Timestamp = golden[goldenIndex].Timestamp
					goldenIndexes[newIndex] = goldenIndex
					goldenIndex++
					//We want to stop iterating in this for loop, and continue in the one above.
					doContinue = true
					break
				}
			}
			if doContinue {
				continue
			}
		}
		//Signal that we didn't have a goldenIndex to copy from--that is, the
		//item appears to be new in new and not present in golden.
		goldenIndexes[newIndex] = -1
	}

	//Now we need to "smear" the times into the "holes" (that is, the moves that
	//were added in new but have no analog in golden) so that the timestamp
	//differences smoothly spread across the hole.

	//"holes" are the indexes in goldenIndexes that are -1... that is, that
	//don't have a corresponding pair in golden. holeLeftIndexes and
	//holeRightIndexes encode the start and end index of the indexes on either
	//end of the "hole" that are not themselves part of the hole.
	holeLeftIndexes := make([]int, len(new))
	holeRightIndexes := make([]int, len(new))

	nonHoleIndex := -1
	for i := 0; i < len(goldenIndexes); i++ {
		goldenIndex := goldenIndexes[i]
		if goldenIndex == -1 {
			holeLeftIndexes[i] = nonHoleIndex
			continue
		}
		nonHoleIndex = i
	}

	nonHoleIndex = -1
	for i := len(goldenIndexes) - 1; i >= 0; i-- {
		goldenIndex := goldenIndexes[i]
		if goldenIndex == -1 {
			holeRightIndexes[i] = nonHoleIndex
			continue
		}
		nonHoleIndex = i
	}

	for i, goldenIndex := range goldenIndexes {
		if goldenIndex != -1 {
			continue
		}
		leftHoleIndex := holeLeftIndexes[i]
		rightHoleIndex := holeRightIndexes[i]
		if leftHoleIndex == -1 && rightHoleIndex == -1 {
			//Hmmm, guess there's no matches whatsoever, just leave it alone
			continue
		}
		//If either the left or right hole index is the sentinel, then the new
		//moves are either at the very beginning or very end. Just copy over the
		//values of the closest non-hole one.
		if leftHoleIndex == -1 {
			new[i].Timestamp = new[rightHoleIndex].Timestamp
			continue
		}
		if rightHoleIndex == -1 {
			new[i].Timestamp = new[leftHoleIndex].Timestamp
			continue
		}
		holeSize := rightHoleIndex - leftHoleIndex
		duration := new[rightHoleIndex].Timestamp.Sub(new[leftHoleIndex].Timestamp)
		new[i].Timestamp = new[leftHoleIndex].Timestamp.Add(duration / time.Duration(holeSize) * time.Duration(i-leftHoleIndex))
	}

	return nil
}
