package golden

import (
	"strconv"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/storage/filesystem"
	"github.com/jkomoros/boardgame/storage/filesystem/record"
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
