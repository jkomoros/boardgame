/*

Package mysql provides a mysql-backed database that implements both
boardgame.StorageManager and boardgame/server.StorageManager. See the README.md
for more information on how to configure and use it.

*/
package mysql

import (
	"database/sql"
	"errors"
	"log"

	"github.com/go-gorp/gorp"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/jkomoros/boardgame/server/api/users"
	"github.com/jkomoros/boardgame/storage/mysql/connect"
)

const (
	tableGames         = "games"
	tableExtendedGames = "extendedgames"
	tableMoves         = "moves"
	tableUsers         = "users"
	tableStates        = "states"
	tableCookies       = "cookies"
	tablePlayers       = "players"
	tableAgentStates   = "agentstates"
)

const baseCombinedSelectQuery = "select g.Name, g.Id, g.SecretSalt, g.Version, g.Winners, g.Finished, g.NumPlayers, g.Agents, " +
	"g.Created, g.Modified, e.Open, e.Visible, e.Owner"

const baseCombinedFromQuery = "from " + tableGames + " g, " + tableExtendedGames + " e"

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

//StorageManager is the primary type in this package.
type StorageManager struct {
	db       *sql.DB
	dbMap    *gorp.DbMap
	testMode bool
	//The config string that we were provided in connect.
	config    string
	connected bool
}

//NewStorageManager returns a new storage manager. Does most of its set-up
//work in Connect(), which is when the database configuration information is
//passed. testMode is whether or not the storage manager is being run in the
//context of a test; if false, then calls to CleanUp (which drops the entire
//database) won't do anything.
func NewStorageManager(testMode bool) *StorageManager {
	//We actually don't do much; we do more of our work in Connect()
	return &StorageManager{
		testMode: testMode,
	}

}

//Connect connects to the database using the given DSN config string.
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

	s.dbMap.AddTableWithName(userStorageRecord{}, tableUsers).SetKeys(false, "Id")
	s.dbMap.AddTableWithName(gameStorageRecord{}, tableGames).SetKeys(false, "Id")
	s.dbMap.AddTableWithName(extendedGameStorageRecord{}, tableExtendedGames).SetKeys(false, "Id")
	s.dbMap.AddTableWithName(stateStorageRecord{}, tableStates).SetKeys(true, "Id")
	s.dbMap.AddTableWithName(cookieStorageRecord{}, tableCookies).SetKeys(false, "Cookie")
	s.dbMap.AddTableWithName(playerStorageRecord{}, tablePlayers).SetKeys(true, "Id")
	s.dbMap.AddTableWithName(agentStateStorageRecord{}, tableAgentStates).SetKeys(true, "Id")
	s.dbMap.AddTableWithName(moveStorageRecord{}, tableMoves).SetKeys(true, "Id")

	_, err = s.dbMap.SelectInt("select count(*) from " + tableGames)

	if err != nil {
		return errors.New("Sanity check failed for db. Have you used the admin tool to migrate it up? " + err.Error())
	}

	s.connected = true

	return nil

}

//Close closes out the connection to the database.
func (s *StorageManager) Close() {
	if s.db == nil {
		return
	}
	s.db.Close()
	s.db = nil
	s.dbMap = nil
	s.connected = false
}

//CleanUp drops the test DB, but only if it was created in TestMode.
func (s *StorageManager) CleanUp() {
	if !s.testMode {
		return
	}
	//connect will refuse to drop the db if it's not the test db name.
	connect.DropTestDb(s.config)
}

//Name returns 'mysql'
func (s *StorageManager) Name() string {
	return "mysql"
}

//State returns the given state
func (s *StorageManager) State(gameID string, version int) (boardgame.StateStorageRecord, error) {

	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}

	var state stateStorageRecord

	err := s.dbMap.SelectOne(&state, "select * from "+tableStates+" where GameId=? and Version=?", gameID, version)

	if err == sql.ErrNoRows {
		return nil, errors.New("No such state")
	}

	if err != nil {
		return nil, errors.New("Unexpected error: " + err.Error())
	}

	return (&state).ToStorageRecord(), nil
}

//Moves returns the given moves
func (s *StorageManager) Moves(gameID string, fromVersion, toVersion int) ([]*boardgame.MoveStorageRecord, error) {

	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}

	var moves []*moveStorageRecord

	if fromVersion == toVersion {
		fromVersion = fromVersion - 1
	}

	_, err := s.dbMap.Select(&moves, "select * from "+tableMoves+" where GameId=? and Version>? and Version<=? order by Version", gameID, fromVersion, toVersion)

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

//Move returns the given Move
func (s *StorageManager) Move(gameID string, version int) (*boardgame.MoveStorageRecord, error) {
	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}

	var move moveStorageRecord

	err := s.dbMap.SelectOne(&move, "select * from "+tableMoves+" where GameId=? and Version=?", gameID, version)

	if err == sql.ErrNoRows {
		return nil, errors.New("No such state")
	}

	if err != nil {
		return nil, errors.New("Unexpected error: " + err.Error())
	}

	return (&move).ToStorageRecord(), nil
}

//Game returns the given Game
func (s *StorageManager) Game(id string) (*boardgame.GameStorageRecord, error) {

	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}

	var game gameStorageRecord

	err := s.dbMap.SelectOne(&game, "select * from "+tableGames+" where Id=?", id)

	if err == sql.ErrNoRows {
		return nil, errors.New("No such game")
	}

	if err != nil {
		return nil, errors.New("Unexpected error: " + err.Error())
	}

	return (&game).ToStorageRecord(), nil
}

//ExtendedGame returns the given ExtendedGame
func (s *StorageManager) ExtendedGame(id string) (*extendedgame.StorageRecord, error) {
	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}

	var record extendedGameStorageRecord

	err := s.dbMap.SelectOne(&record, "select * from "+tableExtendedGames+" where Id=?", id)

	if err != nil {
		return nil, err
	}

	return (&record).ToStorageRecord(), nil
}

//CombinedGame returns the given CombinedGame
func (s *StorageManager) CombinedGame(id string) (*extendedgame.CombinedStorageRecord, error) {

	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}

	var record combinedGameStorageRecord

	err := s.dbMap.SelectOne(&record, combinedGameStorageRecordQuery+" and g.Id = ?", id)

	if err != nil {
		return nil, err
	}

	return (&record).ToStorageRecord(), nil
}

//SaveGameAndCurrentState saves the given game and current state.
func (s *StorageManager) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {

	if !s.connected {
		return errors.New("Database not connected yet")
	}

	version := game.Version

	gameRecord := NewGameStorageRecord(game)
	stateRecord := NewStateStorageRecord(game.ID, version, state)

	var moveRecord *moveStorageRecord

	if move != nil {
		moveRecord = NewMoveStorageRecord(game.ID, version, move)
	}

	count, _ := s.dbMap.SelectInt("select count(*) from "+tableGames+" where Id=?", game.ID)

	if count < 1 {
		//Need to insert
		err := s.dbMap.Insert(gameRecord)

		if err != nil {
			return errors.New("Couldn't update game: " + err.Error())
		}

		extendedRecord := NewExtendedGameStorageRecord(extendedgame.DefaultStorageRecord())

		extendedRecord.Id = game.ID

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

//AgentState returns the given AgentState
func (s *StorageManager) AgentState(gameID string, player boardgame.PlayerIndex) ([]byte, error) {

	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}

	var agent agentStateStorageRecord

	err := s.dbMap.SelectOne(&agent, "select * from "+tableAgentStates+" where GameId=? and PlayerIndex=? order by Id desc limit 1", gameID, int64(player))

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return agent.ToStorageRecord(), nil

}

//SaveAgentState saves the given agent state
func (s *StorageManager) SaveAgentState(gameID string, player boardgame.PlayerIndex, state []byte) error {
	if !s.connected {
		return errors.New("Database not connected yet")
	}

	record := NewAgentStateStorageRecord(gameID, player, state)

	err := s.dbMap.Insert(record)

	if err != nil {
		return errors.New("Couldn't save record: " + err.Error())
	}

	return nil
}

//UpdateExtendedGame updates the given extended game properties
func (s *StorageManager) UpdateExtendedGame(id string, eGame *extendedgame.StorageRecord) error {

	if !s.connected {
		return errors.New("Database not connected yet")
	}

	record := NewExtendedGameStorageRecord(eGame)
	record.Id = id

	_, err := s.dbMap.Update(record)

	return err
}

//ListGames lists the given games
func (s *StorageManager) ListGames(max int, list listing.Type, userID string, gameType string) []*extendedgame.CombinedStorageRecord {

	if !s.connected {
		return nil
	}

	var games []combinedGameStorageRecord

	if max < 1 {
		max = 100
	}

	if (list == listing.ParticipatingActive || list == listing.ParticipatingFinished) && userID == "" {
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
		args = append(args, userID)
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

	query += " order by g.Modified desc limit ?"

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

//SetPlayerForGame affiliates the given user in the given game to the given player
func (s *StorageManager) SetPlayerForGame(gameID string, playerIndex boardgame.PlayerIndex, userID string) error {

	if !s.connected {
		return errors.New("Database not connected yet")
	}

	game, err := s.Game(gameID)

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

	var player playerStorageRecord

	err = s.dbMap.SelectOne(&player, "select * from "+tablePlayers+" where GameId=? and PlayerIndex=?", game.ID, int(playerIndex))

	if err == sql.ErrNoRows {
		// Insert the row

		player = playerStorageRecord{
			GameId:      game.ID,
			PlayerIndex: int64(playerIndex),
			UserId:      userID,
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

	player.UserId = userID

	_, err = s.dbMap.Update(player)

	if err != nil {
		return errors.New("Couldn't update player line: " + err.Error())
	}

	return nil

}

//UserIDsForGame returns the given UserIds
func (s *StorageManager) UserIDsForGame(gameID string) []string {

	if !s.connected {
		return nil
	}

	game, err := s.Game(gameID)

	if err != nil {
		log.Println("Couldn't get game: " + err.Error())
		return nil
	}

	if game == nil {
		log.Println("No game returned.")
		return nil
	}

	var players []playerStorageRecord

	_, err = s.dbMap.Select(&players, "select * from "+tablePlayers+" where GameId=? order by PlayerIndex desc", game.ID)

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

//UpdateUser updates the given user
func (s *StorageManager) UpdateUser(user *users.StorageRecord) error {
	userRecord := NewUserStorageRecord(user)

	existingRecord, _ := s.dbMap.SelectInt("select count(*) from "+tableUsers+" where Id=?", user.ID)

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
			return errors.New("row could not be updated")
		}
	}

	return nil
}

//GetUserByID gets the given user
func (s *StorageManager) GetUserByID(uid string) *users.StorageRecord {
	if !s.connected {
		return nil
	}

	var user userStorageRecord

	err := s.dbMap.SelectOne(&user, "select * from "+tableUsers+" where Id=?", uid)

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

//GetUserByCookie gets the given user
func (s *StorageManager) GetUserByCookie(cookie string) *users.StorageRecord {

	if !s.connected {
		return nil
	}

	var cookieRecord cookieStorageRecord

	err := s.dbMap.SelectOne(&cookieRecord, "select * from "+tableCookies+" where Cookie=?", cookie)

	if err == sql.ErrNoRows {
		//No user
		return nil
	}

	if err != nil {
		log.Println("Unexpected error getting user by cookie: " + err.Error())
		return nil
	}

	return s.GetUserByID(cookieRecord.UserId)

}

//ConnectCookieToUser affiliates the given cookie to the given user
func (s *StorageManager) ConnectCookieToUser(cookie string, user *users.StorageRecord) error {

	if !s.connected {
		return errors.New("Database not connected yet")
	}

	//If user is nil, then delete any records with that cookie.
	if user == nil {

		var cookieRecord cookieStorageRecord

		err := s.dbMap.SelectOne(&cookieRecord, "select * from "+tableCookies+" where Cookie=?", cookie)

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
	otherUser := s.GetUserByID(user.ID)

	if otherUser == nil {

		//Have to save the user for the first time
		if err := s.UpdateUser(user); err != nil {
			return errors.New("Couldn't add a new user to the database when connecting to cookie: " + err.Error())
		}

		return nil
	}

	record := &cookieStorageRecord{
		Cookie: cookie,
		UserId: user.ID,
	}

	if err := s.dbMap.Insert(record); err != nil {
		return errors.New("Failed to insert cookie pointer record: " + err.Error())
	}
	return nil
}

//PlayerMoveApplied does nothing
func (s *StorageManager) PlayerMoveApplied(game *boardgame.GameStorageRecord) error {
	//Don't need to do anything
	return nil
}

//WithManagers does nothing
func (s *StorageManager) WithManagers(managers []*boardgame.GameManager) {
	//Do nothing
}
