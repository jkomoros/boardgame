/*

mysql provides a bolt-backed database that implements both
boardgame.StorageManager and boardgame/server.StorageManager.

*/
package mysql

import (
	"database/sql"
	"errors"
	"github.com/go-gorp/gorp"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/jkomoros/boardgame/server/api/users"
	"github.com/jkomoros/boardgame/storage/mysql/connect"
	"log"
	"time"
)

const (
	TableGames         = "games"
	TableExtendedGames = "extendedgames"
	TableMoves         = "moves"
	TableUsers         = "users"
	TableStates        = "states"
	TableCookies       = "cookies"
	TablePlayers       = "players"
	TableAgentStates   = "agentstates"
)

const baseCombinedSelectQuery = "select g.Name, g.Id, g.SecretSalt, g.Version, g.Winners, g.Finished, g.NumPlayers, g.Agents, " +
	"g.Created, e.LastActivity, e.Open, e.Visible, e.Owner"

const baseCombinedFromQuery = "from " + TableGames + " g, " + TableExtendedGames + " e"

const baseCombinedWhereQuery = "where g.Id = e.Id"

const combinedPlayerFilterQuery = baseCombinedSelectQuery + " " + baseCombinedFromQuery + ", players p " + baseCombinedWhereQuery +
	" and p.GameId = g.Id and p.UserId = ?"

const combinedGameStorageRecordQuery = baseCombinedSelectQuery + " " + baseCombinedFromQuery + " " + baseCombinedWhereQuery

const userNotInQuery = "not exists (select * from players where GameId = g.Id and UserId = ?)"

const emptySlotsQuery = "(g.NumPlayers > coalesce(c.NumActivePlayers, 0) + g.NumAgents)"

const combinedHasSlots = baseCombinedSelectQuery + ` from games as g
left join extendedgames as e
	left join (select GameId as Id, count(*) as NumActivePlayers from players group by GameId) as c
	on e.Id = c.Id
on g.Id = e.Id
where`

const combinedNotPlayerFilterQuery = combinedHasSlots + " " + userNotInQuery

const combinedNotPlayerOpenSlotsQuery = combinedNotPlayerFilterQuery + " and " + emptySlotsQuery

const combinedNotPlayerNoOpenSlotsQuery = combinedNotPlayerFilterQuery + " and (not " + emptySlotsQuery + " or e.Open = 0)"

type StorageManager struct {
	db       *sql.DB
	dbMap    *gorp.DbMap
	testMode bool
	//The config string that we were provided in connect.
	config string
}

func NewStorageManager(testMode bool) *StorageManager {
	//We actually don't do much; we do more of our work in Connect()
	return &StorageManager{
		testMode: testMode,
	}

}

func (s *StorageManager) Connect(config string) error {

	db, err := connect.Db(config, s.testMode, s.testMode)

	if err != nil {
		return errors.New("Couldn't connect to db: " + err.Error())
	}

	s.config = config

	s.db = db

	s.dbMap = &gorp.DbMap{
		Db: db,
		Dialect: gorp.MySQLDialect{
			Engine: "InnoDB",
			//the mb4 is necessary to support e.g. emojis
			Encoding: "utf8mb4",
		},
	}

	s.dbMap.AddTableWithName(UserStorageRecord{}, TableUsers).SetKeys(false, "Id")
	s.dbMap.AddTableWithName(GameStorageRecord{}, TableGames).SetKeys(false, "Id")
	s.dbMap.AddTableWithName(ExtendedGameStorageRecord{}, TableExtendedGames).SetKeys(false, "Id")
	s.dbMap.AddTableWithName(StateStorageRecord{}, TableStates).SetKeys(true, "Id")
	s.dbMap.AddTableWithName(CookieStorageRecord{}, TableCookies).SetKeys(false, "Cookie")
	s.dbMap.AddTableWithName(PlayerStorageRecord{}, TablePlayers).SetKeys(true, "Id")
	s.dbMap.AddTableWithName(AgentStateStorageRecord{}, TableAgentStates).SetKeys(true, "Id")
	s.dbMap.AddTableWithName(MoveStorageRecord{}, TableMoves).SetKeys(true, "Id")

	_, err = s.dbMap.SelectInt("select count(*) from " + TableGames)

	if err != nil {
		return errors.New("Sanity check failed for db. Have you used the admin tool to migrate it up? " + err.Error())
	}

	return nil

}

func (s *StorageManager) Close() {
	if s.db == nil {
		return
	}
	s.db.Close()
	s.db = nil
}

func (s *StorageManager) CleanUp() {
	if !s.testMode {
		return
	}
	//connect will refuse to drop the db if it's not the test db name.
	connect.DropTestDb(s.config)
}

func (s *StorageManager) Name() string {
	return "mysql"
}

func (s *StorageManager) State(gameId string, version int) (boardgame.StateStorageRecord, error) {
	var state StateStorageRecord

	err := s.dbMap.SelectOne(&state, "select * from "+TableStates+" where GameId=? and Version=?", gameId, version)

	if err == sql.ErrNoRows {
		return nil, errors.New("No such state")
	}

	if err != nil {
		return nil, errors.New("Unexpected error: " + err.Error())
	}

	return (&state).ToStorageRecord(), nil
}

func (s *StorageManager) Moves(gameId string, fromVersion, toVersion int) ([]*boardgame.MoveStorageRecord, error) {

	var moves []*MoveStorageRecord

	if fromVersion == toVersion {
		fromVersion = fromVersion - 1
	}

	_, err := s.dbMap.Select(&moves, "select * from "+TableMoves+" where GameId=? and Version>? and Version<=? order by Version", gameId, fromVersion, toVersion)

	if err == sql.ErrNoRows {
		return nil, errors.New("No moves returned")
	}

	if err != nil {
		return nil, errors.New("Unexpected error: " + err.Error())
	}

	result := make([]*boardgame.MoveStorageRecord, len(moves))

	for i, move := range moves {
		result[i] = move.ToStorageRecord()
	}

	return result, nil

}

func (s *StorageManager) Move(gameId string, version int) (*boardgame.MoveStorageRecord, error) {
	var move MoveStorageRecord

	err := s.dbMap.SelectOne(&move, "select * from "+TableMoves+" where GameId=? and Version=?", gameId, version)

	if err == sql.ErrNoRows {
		return nil, errors.New("No such state")
	}

	if err != nil {
		return nil, errors.New("Unexpected error: " + err.Error())
	}

	return (&move).ToStorageRecord(), nil
}

func (s *StorageManager) Game(id string) (*boardgame.GameStorageRecord, error) {
	var game GameStorageRecord

	err := s.dbMap.SelectOne(&game, "select * from "+TableGames+" where Id=?", id)

	if err == sql.ErrNoRows {
		return nil, errors.New("No such game")
	}

	if err != nil {
		return nil, errors.New("Unexpected error: " + err.Error())
	}

	return (&game).ToStorageRecord(), nil
}

func (s *StorageManager) ExtendedGame(id string) (*extendedgame.StorageRecord, error) {
	var record ExtendedGameStorageRecord

	err := s.dbMap.SelectOne(&record, "select * from "+TableExtendedGames+" where Id=?", id)

	if err != nil {
		return nil, err
	}

	return (&record).ToStorageRecord(), nil
}

func (s *StorageManager) CombinedGame(id string) (*extendedgame.CombinedStorageRecord, error) {
	var record CombinedGameStorageRecord

	err := s.dbMap.SelectOne(&record, combinedGameStorageRecordQuery+" and g.Id = ?", id)

	if err != nil {
		return nil, err
	}

	return (&record).ToStorageRecord(), nil
}

func (s *StorageManager) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {

	version := game.Version

	gameRecord := NewGameStorageRecord(game)
	stateRecord := NewStateStorageRecord(game.Id, version, state)

	var moveRecord *MoveStorageRecord

	if move != nil {
		moveRecord = NewMoveStorageRecord(game.Id, version, move)
	}

	count, _ := s.dbMap.SelectInt("select count(*) from "+TableGames+" where Id=?", game.Id)

	if count < 1 {
		//Need to insert
		err := s.dbMap.Insert(gameRecord)

		if err != nil {
			return errors.New("Couldn't update game: " + err.Error())
		}

		extendedRecord := NewExtendedGameStorageRecord(extendedgame.DefaultStorageRecord())

		extendedRecord.Id = game.Id

		err = s.dbMap.Insert(extendedRecord)

		if err != nil {
			return errors.New("Couldn't insert the extended game info: " + err.Error())
		}

	} else {
		//Need to update
		_, err := s.dbMap.Update(gameRecord)

		if err != nil {
			return errors.New("Couldn't insert game: " + err.Error())
		}

		err = s.touchExtendedGameLastActivity(game.Id)

		if err != nil {
			return errors.New("Couldn't update LastActivty on game: " + err.Error())
		}

	}

	err := s.dbMap.Insert(stateRecord)

	if err != nil {
		return errors.New("Couldn't insert state: " + err.Error())
	}

	if moveRecord != nil {
		err = s.dbMap.Insert(moveRecord)

		if err != nil {
			return errors.New("couldn't insert move: " + err.Error())
		}
	}

	return nil
}

func (s *StorageManager) touchExtendedGameLastActivity(id string) error {
	var rec ExtendedGameStorageRecord

	if err := s.dbMap.SelectOne(&rec, "select * from "+TableExtendedGames+" where Id=? limit 1", id); err != nil {
		return errors.New("Couldn't fetch lastActivity: " + err.Error())
	}

	rec.LastActivity = time.Now().UnixNano()

	if _, err := s.dbMap.Update(&rec); err != nil {
		return errors.New("Couldn't update lastActivity: " + err.Error())
	}

	return nil
}

func (s *StorageManager) AgentState(gameId string, player boardgame.PlayerIndex) ([]byte, error) {

	var agent AgentStateStorageRecord

	err := s.dbMap.SelectOne(&agent, "select * from "+TableAgentStates+" where GameId=? and PlayerIndex=? order by Id desc limit 1", gameId, int64(player))

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return agent.ToStorageRecord(), nil

}

func (s *StorageManager) SaveAgentState(gameId string, player boardgame.PlayerIndex, state []byte) error {
	record := NewAgentStateStorageRecord(gameId, player, state)

	err := s.dbMap.Insert(record)

	if err != nil {
		return errors.New("Couldn't save record: " + err.Error())
	}

	return nil
}

func (s *StorageManager) UpdateExtendedGame(id string, eGame *extendedgame.StorageRecord) error {
	record := NewExtendedGameStorageRecord(eGame)
	record.Id = id

	_, err := s.dbMap.Update(record)

	return err
}

func (s *StorageManager) ListGames(max int, list listing.Type, userId string, gameType string) []*extendedgame.CombinedStorageRecord {
	var games []CombinedGameStorageRecord

	if max < 1 {
		max = 100
	}

	if (list == listing.ParticipatingActive || list == listing.ParticipatingFinished) && userId == "" {
		//If we're filtering to only participating games and there's no userId, then there can't be any games,
		//because the non-user can't be participating in any games.
		return nil
	}

	query := combinedGameStorageRecordQuery

	var args []interface{}

	if list != listing.All {

		switch list {
		case listing.VisibleActive:
			query = combinedNotPlayerNoOpenSlotsQuery
		case listing.VisibleJoinableActive:
			query = combinedNotPlayerOpenSlotsQuery
		default:
			query = combinedPlayerFilterQuery
		}
		args = append(args, userId)
	}

	switch list {
	case listing.ParticipatingActive:
		query += " and g.Finished = 0"
	case listing.ParticipatingFinished:
		query += " and g.Finished = 1"
	case listing.VisibleJoinableActive:
		query += " and g.Finished = 0 and e.Visible = 1 and e.Open = 1"
	case listing.VisibleActive:
		query += " and g.Finished = 0 and e.Visible = 1"
	}

	if gameType != "" {
		query += " and g.Name = ?"
		args = append(args, gameType)
	}

	query += " order by e.LastActivity desc limit ?"

	args = append(args, max)

	if _, err := s.dbMap.Select(&games, query, args...); err != nil {
		log.Println("List games failed: " + err.Error())
		return nil
	}

	result := make([]*extendedgame.CombinedStorageRecord, len(games))

	for i, record := range games {
		result[i] = (&record).ToStorageRecord()
	}

	return result
}

func (s *StorageManager) SetPlayerForGame(gameId string, playerIndex boardgame.PlayerIndex, userId string) error {

	game, err := s.Game(gameId)

	if err != nil {
		return errors.New("Couldn't get game: " + err.Error())
	}

	if game == nil {
		return errors.New("No game returned")
	}

	if playerIndex < 0 || int(playerIndex) >= int(game.NumPlayers) {
		return errors.New("Invalid player index")
	}

	//TODO: should we validate that this is a real userId?

	var player PlayerStorageRecord

	err = s.dbMap.SelectOne(&player, "select * from "+TablePlayers+" where GameId=? and PlayerIndex=?", game.Id, int(playerIndex))

	if err == sql.ErrNoRows {
		// Insert the row

		player = PlayerStorageRecord{
			GameId:      game.Id,
			PlayerIndex: int64(playerIndex),
			UserId:      userId,
		}

		err = s.dbMap.Insert(&player)

		if err != nil {
			return errors.New("Couldn't insert new player line: " + err.Error())
		}

		return nil
	}

	//Update the row, if it wasn't an error.

	if err != nil {
		return errors.New("Failed to retrieve existing Player line: " + err.Error())
	}

	player.UserId = userId

	_, err = s.dbMap.Update(player)

	if err != nil {
		return errors.New("Couldn't update player line: " + err.Error())
	}

	return nil

}

func (s *StorageManager) UserIdsForGame(gameId string) []string {

	game, err := s.Game(gameId)

	if err != nil {
		log.Println("Couldn't get game: " + err.Error())
		return nil
	}

	if game == nil {
		log.Println("No game returned.")
		return nil
	}

	var players []PlayerStorageRecord

	_, err = s.dbMap.Select(&players, "select * from "+TablePlayers+" where GameId=? order by PlayerIndex desc", game.Id)

	result := make([]string, game.NumPlayers)

	if err == sql.ErrNoRows {
		return result
	}

	if err != nil {
		log.Println("Couldn't get rows: ", err.Error())
		return result
	}

	for _, rec := range players {
		index := int(rec.PlayerIndex)

		if index < 0 || index >= len(result) {
			log.Println("Invalid index", rec)
			continue
		}

		result[index] = rec.UserId
	}

	return result

}

func (s *StorageManager) UpdateUser(user *users.StorageRecord) error {
	userRecord := NewUserStorageRecord(user)

	existingRecord, _ := s.dbMap.SelectInt("select count(*) from "+TableUsers+" where Id=?", user.Id)

	if existingRecord < 1 {
		//Need to insert
		err := s.dbMap.Insert(userRecord)

		if err != nil {
			return errors.New("Couldn't insert user: " + err.Error())
		}
	} else {
		//Need to update
		//TODO: I wonder if this will fail if the user is not yet in the database.
		count, err := s.dbMap.Update(userRecord)
		if err != nil {
			return errors.New("Couldn't update user: " + err.Error())
		}

		if count < 1 {
			return errors.New("Row could not be updated.")
		}
	}

	return nil
}

func (s *StorageManager) GetUserById(uid string) *users.StorageRecord {
	var user UserStorageRecord

	err := s.dbMap.SelectOne(&user, "select * from "+TableUsers+" where Id=?", uid)

	if err == sql.ErrNoRows {
		//Normal
		return nil
	}

	if err != nil {
		log.Println("Unexpected error getting user:", err)
		return nil
	}

	return (&user).ToStorageRecord()
}

func (s *StorageManager) GetUserByCookie(cookie string) *users.StorageRecord {

	var cookieRecord CookieStorageRecord

	err := s.dbMap.SelectOne(&cookieRecord, "select * from "+TableCookies+" where Cookie=?", cookie)

	if err == sql.ErrNoRows {
		//No user
		return nil
	}

	if err != nil {
		log.Println("Unexpected error getting user by cookie: " + err.Error())
		return nil
	}

	return s.GetUserById(cookieRecord.UserId)

}

func (s *StorageManager) ConnectCookieToUser(cookie string, user *users.StorageRecord) error {
	//If user is nil, then delete any records with that cookie.
	if user == nil {

		var cookieRecord CookieStorageRecord

		err := s.dbMap.SelectOne(&cookieRecord, "select * from "+TableCookies+" where Cookie=?", cookie)

		if err == sql.ErrNoRows {
			//We're fine, because it wasn't in the table any way!
			return nil
		}

		if err != nil {
			return errors.New("Unexpected error: " + err.Error())
		}

		//It was there, so we need to delete it.

		count, err := s.dbMap.Delete(&cookieRecord)

		if count < 1 && err != nil {
			return errors.New("Couldnt' delete cookie record when instructed to: " + err.Error())
		}

		return nil
	}

	//If user does not yet exist in database, put them in.
	otherUser := s.GetUserById(user.Id)

	if otherUser == nil {

		//Have to save the user for the first time
		if err := s.UpdateUser(user); err != nil {
			return errors.New("Couldn't add a new user to the database when connecting to cookie: " + err.Error())
		}

		return nil
	}

	record := &CookieStorageRecord{
		Cookie: cookie,
		UserId: user.Id,
	}

	if err := s.dbMap.Insert(record); err != nil {
		return errors.New("Failed to insert cookie pointer record: " + err.Error())
	}
	return nil
}

func (s *StorageManager) PlayerMoveApplied(game *boardgame.GameStorageRecord) error {
	//Don't need to do anything
	return nil
}
