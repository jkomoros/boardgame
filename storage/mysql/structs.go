package mysql

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/users"
)

//We define our own records, primarily to decorate with tags for Gorp, but
//also because e.g boardgame.storage.users.StorageRecord isn't structured the
//way we actually want to store in DB.

type userStorageRecord struct {
	ID          string `db:",size:128"`
	Created     int64
	LastSeen    int64
	DisplayName string `db:",size:64"`
	PhotoURL    string `db:",size:1024"`
	Email       string `db:",size:128"`
}

type cookieStorageRecord struct {
	Cookie string `db:",size:64"`
	UserID string `db:",size:128"`
}

type gameStorageRecord struct {
	Name       string `db:",size:64"`
	ID         string `db:",size:16"`
	SecretSalt string `db:",size:16"`
	Version    int64
	Winners    string `db:",size:128"`
	Finished   bool
	Created    int64
	Modified   int64
	//NumPlayers is the reported number of players when it was created.
	//Primarily for convenience to storage layer so they know how many players
	//are in the game.
	NumPlayers int64
	Agents     string `db:",size:1024"`
	//Derived field to enable HasEmptySlots SQL query
	NumAgents int64
	Variant   string `db:",size:65536"`
}

type extendedGameStorageRecord struct {
	ID      string `db:",size:16"`
	Open    bool
	Visible bool
	Owner   string `db:",size:128"`
}

//Used for pulling out of a db with a join
type combinedGameStorageRecord struct {
	Name       string
	ID         string
	SecretSalt string
	Version    int64
	Winners    string
	Finished   bool
	NumPlayers int64
	Agents     string
	Created    int64
	Modified   int64
	Open       bool
	Visible    bool
	Owner      string
}

type stateStorageRecord struct {
	ID      int64
	GameID  string `db:",size:16"`
	Version int64
	Blob    string `db:",size:10000000"`
}

type moveStorageRecord struct {
	ID        int64
	GameID    string `db:",size:16"`
	Version   int64
	Initiator int64
	Timestamp int64
	Phase     int64
	Proposer  int64
	Name      string `db:",size:128"`
	Blob      string `db:",size:100000"`
}

type playerStorageRecord struct {
	ID          int64
	GameID      string `db:",size:16"`
	PlayerIndex int64
	UserID      string `db:",size:128"`
}

type agentStateStorageRecord struct {
	ID          int64
	GameID      string `db:",size:16"`
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

func stringToConfig(config string) (boardgame.Variant, error) {
	if config == "" {
		return nil, nil
	}

	var result boardgame.Variant

	if err := json.Unmarshal([]byte(config), &result); err != nil {
		return nil, errors.New("Couldn't unmarshal value: " + err.Error())
	}

	return result, nil
}

func configToString(config boardgame.Variant) string {
	if config == nil {
		return ""
	}

	blob, _ := json.Marshal(config)

	return string(blob)
}

func (g *gameStorageRecord) ToStorageRecord() *boardgame.GameStorageRecord {

	if g == nil {
		return nil
	}

	winners, err := stringToWinners(g.Winners)

	if err != nil {
		return nil
	}

	variant, err := stringToConfig(g.Variant)

	if err != nil {
		return nil
	}

	return &boardgame.GameStorageRecord{
		Name:       g.Name,
		ID:         g.ID,
		SecretSalt: g.SecretSalt,
		Version:    int(g.Version),
		Winners:    winners,
		Created:    time.Unix(0, g.Created),
		Modified:   time.Unix(0, g.Modified),
		Finished:   g.Finished,
		NumPlayers: int(g.NumPlayers),
		Agents:     stringToAgents(g.Agents),
		Variant:    variant,
	}
}

func NewGameStorageRecord(game *boardgame.GameStorageRecord) *gameStorageRecord {
	if game == nil {
		return nil
	}

	numAgents := 0

	for _, agent := range game.Agents {
		if agent != "" {
			numAgents++
		}
	}

	return &gameStorageRecord{
		Name:       game.Name,
		ID:         game.ID,
		SecretSalt: game.SecretSalt,
		Version:    int64(game.Version),
		Winners:    winnersToString(game.Winners),
		NumPlayers: int64(game.NumPlayers),
		Finished:   game.Finished,
		Created:    game.Created.UnixNano(),
		Modified:   game.Modified.UnixNano(),
		Agents:     agentsToString(game.Agents),
		NumAgents:  int64(numAgents),
		Variant:    configToString(game.Variant),
	}
}

func (c *combinedGameStorageRecord) ToStorageRecord() *extendedgame.CombinedStorageRecord {
	if c == nil {
		return nil
	}

	winners, err := stringToWinners(c.Winners)

	if err != nil {
		return nil
	}

	return &extendedgame.CombinedStorageRecord{
		GameStorageRecord: boardgame.GameStorageRecord{
			Name:       c.Name,
			ID:         c.ID,
			SecretSalt: c.SecretSalt,
			Version:    int(c.Version),
			Winners:    winners,
			Finished:   c.Finished,
			NumPlayers: int(c.NumPlayers),
			Agents:     stringToAgents(c.Agents),
			Created:    time.Unix(0, c.Created),
			Modified:   time.Unix(0, c.Modified),
		},
		StorageRecord: extendedgame.StorageRecord{
			Open:    c.Open,
			Visible: c.Visible,
			Owner:   c.Owner,
		},
	}

}

func NewCombinedGameStorageRecord(combined *extendedgame.CombinedStorageRecord) *combinedGameStorageRecord {

	if combined == nil {
		return nil
	}

	return &combinedGameStorageRecord{
		Name:       combined.Name,
		ID:         combined.ID,
		SecretSalt: combined.SecretSalt,
		Version:    int64(combined.Version),
		Winners:    winnersToString(combined.Winners),
		NumPlayers: int64(combined.NumPlayers),
		Finished:   combined.Finished,
		Agents:     agentsToString(combined.Agents),
		Created:    combined.Created.UnixNano(),
		Modified:   combined.Modified.UnixNano(),
		Open:       combined.Open,
		Visible:    combined.Visible,
		Owner:      combined.Owner,
	}

}

func (e *extendedGameStorageRecord) ToStorageRecord() *extendedgame.StorageRecord {
	if e == nil {
		return nil
	}

	return &extendedgame.StorageRecord{
		Open:    e.Open,
		Visible: e.Visible,
		Owner:   e.Owner,
	}
}

func NewExtendedGameStorageRecord(eGame *extendedgame.StorageRecord) *extendedGameStorageRecord {
	if eGame == nil {
		return nil
	}

	return &extendedGameStorageRecord{
		Open:    eGame.Open,
		Visible: eGame.Visible,
		Owner:   eGame.Owner,
	}
}

func (m *moveStorageRecord) ToStorageRecord() *boardgame.MoveStorageRecord {
	return &boardgame.MoveStorageRecord{
		Name:      m.Name,
		Version:   int(m.Version),
		Initiator: int(m.Initiator),
		Phase:     int(m.Phase),
		Proposer:  boardgame.PlayerIndex(m.Proposer),
		Timestamp: time.Unix(0, m.Timestamp),
		Blob:      []byte(m.Blob),
	}
}

func NewMoveStorageRecord(gameId string, version int, record *boardgame.MoveStorageRecord) *moveStorageRecord {
	return &moveStorageRecord{
		GameID:    gameId,
		Version:   int64(version),
		Initiator: int64(record.Initiator),
		Timestamp: record.Timestamp.UnixNano(),
		Name:      record.Name,
		Phase:     int64(record.Phase),
		Proposer:  int64(record.Proposer),
		Blob:      string(record.Blob),
	}
}

func (s *userStorageRecord) ToStorageRecord() *users.StorageRecord {
	return &users.StorageRecord{
		ID:          s.ID,
		DisplayName: s.DisplayName,
		Created:     s.Created,
		LastSeen:    s.LastSeen,
		PhotoURL:    s.PhotoURL,
		Email:       s.Email,
	}
}

func NewUserStorageRecord(user *users.StorageRecord) *userStorageRecord {
	return &userStorageRecord{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		Created:     user.Created,
		LastSeen:    user.LastSeen,
		PhotoURL:    user.PhotoURL,
		Email:       user.Email,
	}
}

func (s *stateStorageRecord) ToStorageRecord() boardgame.StateStorageRecord {
	if s == nil {
		return nil
	}
	return []byte(s.Blob)
}

func NewStateStorageRecord(gameId string, version int, record boardgame.StateStorageRecord) *stateStorageRecord {
	return &stateStorageRecord{
		GameID:  gameId,
		Version: int64(version),
		Blob:    string(record),
	}
}

func (a *agentStateStorageRecord) ToStorageRecord() []byte {
	if a == nil {
		return nil
	}
	return []byte(a.Blob)
}

func NewAgentStateStorageRecord(gameId string, player boardgame.PlayerIndex, state []byte) *agentStateStorageRecord {
	return &agentStateStorageRecord{
		GameID:      gameId,
		PlayerIndex: int64(player),
		Blob:        string(state),
	}
}
