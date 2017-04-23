package mysql

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/users"
	"strconv"
	"strings"
)

//We define our own records, primarily to decorate with tags for Gorp, but
//also because e.g boardgame.storage.users.StorageRecord isn't structured the
//way we actually want to store in DB.

type UserStorageRecord struct {
	Id string `db:",size:128"`
}

type CookieStorageRecord struct {
	Cookie string `db:",size:64"`
	UserId string `db:",size:128"`
}

type GameStorageRecord struct {
	Name     string `db:",size:64"`
	Id       string `db:",size:16"`
	Version  int64
	Winners  string `db:",size:128"`
	Finished bool
	//NumPlayers is the reported number of players when it was created.
	//Primarily for convenience to storage layer so they know how many players
	//are in the game.
	NumPlayers int64
}

type StateStorageRecord struct {
	Id      int64
	GameId  string `db:",size:16"`
	Version int64
	Blob    string `db:",size:10000000"`
}

type PlayerStorageRecord struct {
	Id          int64
	GameId      string `db:",size:16"`
	PlayerIndex int64
	UserId      string `db:",size:128"`
}

func winnersToString(winners []boardgame.PlayerIndex) string {
	if winners == nil {
		return ""
	}
	strs := make([]string, len(winners))
	for i, player := range winners {
		strs[i] = player.String()
	}
	return strings.Join(strs, ",")
}

func stringToWinners(winners string) ([]boardgame.PlayerIndex, error) {

	if winners == "" {
		return nil, nil
	}

	strs := strings.Split(winners, ",")

	result := make([]boardgame.PlayerIndex, len(strs))

	for i, str := range strs {
		intIndex, err := strconv.Atoi(str)
		if err != nil {
			return nil, errors.New("couldn't decode " + strconv.Itoa(i) + " player index: " + err.Error())
		}
		result[i] = boardgame.PlayerIndex(intIndex)
	}

	return result, nil

}

func (g *GameStorageRecord) ToStorageRecord() *boardgame.GameStorageRecord {

	if g == nil {
		return nil
	}

	winners, err := stringToWinners(g.Winners)

	if err != nil {
		return nil
	}

	return &boardgame.GameStorageRecord{
		Name:       g.Name,
		Id:         g.Id,
		Version:    int(g.Version),
		Winners:    winners,
		NumPlayers: int(g.NumPlayers),
	}
}

func NewGameStorageRecord(game *boardgame.GameStorageRecord) *GameStorageRecord {
	if game == nil {
		return nil
	}

	return &GameStorageRecord{
		Name:       game.Name,
		Id:         game.Id,
		Version:    int64(game.Version),
		Winners:    winnersToString(game.Winners),
		NumPlayers: int64(game.NumPlayers),
	}
}

func (s *UserStorageRecord) ToStorageRecord() *users.StorageRecord {
	return &users.StorageRecord{
		Id: s.Id,
	}
}

func NewUserStorageRecord(user *users.StorageRecord) *UserStorageRecord {
	return &UserStorageRecord{
		Id: user.Id,
	}
}

func (s *StateStorageRecord) ToStorageRecord() boardgame.StateStorageRecord {
	if s == nil {
		return nil
	}
	return []byte(s.Blob)
}

func NewStateStorageRecord(gameId string, version int, record boardgame.StateStorageRecord) *StateStorageRecord {
	return &StateStorageRecord{
		GameId:  gameId,
		Version: int64(version),
		Blob:    string(record),
	}
}
