/*

bolt provides a bolt-backed database that implements both
boardgame.StorageManager and boardgame/server.StorageManager.

*/
package mysql

import (
	"database/sql"
	"errors"
	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/users"
	"log"
	"strings"
)

const (
	TableGames   = "games"
	TableUsers   = "users"
	TableStates  = "states"
	TableCookies = "cookies"
	TablePlayers = "players"
)

type StorageManager struct {
	db    *sql.DB
	dbMap *gorp.DbMap
	//If in test mode we'll... mock stuff, I guess?
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

	db, err := sql.Open("mysql", config)
	if err != nil {
		return errors.New("Unable to open database: " + err.Error())
	}

	s.config = config

	s.db = db

	s.dbMap = &gorp.DbMap{
		Db: db,
		Dialect: gorp.MySQLDialect{
			Engine:   "MyISAM",
			Encoding: "utf8",
		},
	}

	s.dbMap.AddTableWithName(UserStorageRecord{}, TableUsers).SetKeys(false, "Id")
	s.dbMap.AddTableWithName(GameStorageRecord{}, TableGames).SetKeys(false, "Id")
	s.dbMap.AddTableWithName(StateStorageRecord{}, TableStates).SetKeys(true, "Id")
	s.dbMap.AddTableWithName(CookieStorageRecord{}, TableCookies).SetKeys(false, "Cookie")
	s.dbMap.AddTableWithName(PlayerStorageRecord{}, TablePlayers).SetKeys(true, "Id")

	if s.testMode {

		if err := s.dbMap.CreateTablesIfNotExists(); err != nil {
			return errors.New("Couldn't create tables: " + err.Error())
		}

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

//Cleanup will only drop tables if we're in test mode, and the config string
//used to open the database talked about a test database on localhost (as
//sanity check).
func (s *StorageManager) CleanUp() {

	if !s.testMode {
		return
	}

	if !strings.Contains(s.config, "_test") {
		log.Println("Sanity check on boardgame config before cleanup didn't find _test")
		return
	}

	if !strings.Contains(s.config, "localhost") {
		log.Println("Sanity check on boardgame config before cleanup didn't find localhost")
	}

	if s.db == nil {
		log.Println("Couldn't clean up; db already closed")
		return
	}

	log.Println("Sanity checks passed. Dropping tables to cleanup...")

	if err := s.dbMap.DropTables(); err != nil {
		log.Println("Error dropping tables:", err)
		return
	}
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

func (s *StorageManager) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord) error {

	version := game.Version

	gameRecord := NewGameStorageRecord(game)
	stateRecord := NewStateStorageRecord(game.Id, version, state)

	count, _ := s.dbMap.SelectInt("select count(*) from "+TableGames+" where Id=?", game.Id)

	if count < 1 {
		//Need to insert
		err := s.dbMap.Insert(gameRecord)

		if err != nil {

			return errors.New("Couldn't update game: " + err.Error())

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

	return nil
}

func (s *StorageManager) ListGames(max int) []*boardgame.GameStorageRecord {
	var games []GameStorageRecord

	if max < 1 {
		max = 100
	}

	if _, err := s.dbMap.Select(&games, "select * from "+TableGames+" limit ?", max); err != nil {
		return nil
	}

	result := make([]*boardgame.GameStorageRecord, len(games))

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
