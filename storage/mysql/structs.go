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
	Id          string `db:",size:128"`
	Created     int64
	LastSeen    int64
	DisplayName string `db:",size:64"`
	PhotoUrl    string `db:",size:1024"`
	Email       string `db:",size:128"`
}

type CookieStorageRecord struct {
	Cookie string `db:",size:64"`
	UserId string `db:",size:128"`
}

type GameStorageRecord struct {
	Name       string `db:",size:64"`
	Id         string `db:",size:16"`
	SecretSalt string `db:",size:16"`
	Version    int64
	Winners    string `db:",size:128"`
	Finished   bool
	//NumPlayers is the reported number of players when it was created.
	//Primarily for convenience to storage layer so they know how many players
	//are in the game.
	NumPlayers int64
	Agents     string `db:",size:1024"`
}

type StateStorageRecord struct {
	Id      int64
	GameId  string `db:",size:16"`
	Version int64
	Blob    string `db:",size:10000000"`
}

type MoveStorageRecord struct {
	Id      int64
	GameId  string `db:",size:16"`
	Version int64
	Name    string `db:":size:128"`
	Blob    string `db:",size:100000"`
}

type PlayerStorageRecord struct {
	Id          int64
	GameId      string `db:",size:16"`
	PlayerIndex int64
	UserId      string `db:",size:128"`
}

type AgentStateStorageRecord struct {
	Id          int64
	GameId      string `db:",size:16"`
	PlayerIndex int64
	Blob        string `db:",size:1000000"`
}

func agentsToString(agents []string) string {
	if agents == nil {
		return ""
	}
	return strings.Join(agents, ",")
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

func stringToAgents(agents string) []string {
	if agents == "" {
		return nil
	}

	return strings.Split(agents, ",")
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
		SecretSalt: g.SecretSalt,
		Version:    int(g.Version),
		Winners:    winners,
		Finished:   g.Finished,
		NumPlayers: int(g.NumPlayers),
		Agents:     stringToAgents(g.Agents),
	}
}

func NewGameStorageRecord(game *boardgame.GameStorageRecord) *GameStorageRecord {
	if game == nil {
		return nil
	}

	return &GameStorageRecord{
		Name:       game.Name,
		Id:         game.Id,
		SecretSalt: game.SecretSalt,
		Version:    int64(game.Version),
		Winners:    winnersToString(game.Winners),
		NumPlayers: int64(game.NumPlayers),
		Finished:   game.Finished,
		Agents:     agentsToString(game.Agents),
	}
}

func (m *MoveStorageRecord) ToStorageRecord() *boardgame.MoveStorageRecord {
	return &boardgame.MoveStorageRecord{
		Name: m.Name,
		Blob: []byte(m.Blob),
	}
}

func NewMoveStorageRecord(gameId string, version int, record *boardgame.MoveStorageRecord) *MoveStorageRecord {
	return &MoveStorageRecord{
		GameId:  gameId,
		Version: int64(version),
		Name:    record.Name,
		Blob:    string(record.Blob),
	}
}

func (s *UserStorageRecord) ToStorageRecord() *users.StorageRecord {
	return &users.StorageRecord{
		Id:          s.Id,
		DisplayName: s.DisplayName,
		Created:     s.Created,
		LastSeen:    s.LastSeen,
		PhotoUrl:    s.PhotoUrl,
		Email:       s.Email,
	}
}

func NewUserStorageRecord(user *users.StorageRecord) *UserStorageRecord {
	return &UserStorageRecord{
		Id:          user.Id,
		DisplayName: user.DisplayName,
		Created:     user.Created,
		LastSeen:    user.LastSeen,
		PhotoUrl:    user.PhotoUrl,
		Email:       user.Email,
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

func (a *AgentStateStorageRecord) ToStorageRecord() []byte {
	if a == nil {
		return nil
	}
	return []byte(a.Blob)
}

func NewAgentStateStorageRecord(gameId string, player boardgame.PlayerIndex, state []byte) *AgentStateStorageRecord {
	return &AgentStateStorageRecord{
		GameId:      gameId,
		PlayerIndex: int64(player),
		Blob:        string(state),
	}
}
