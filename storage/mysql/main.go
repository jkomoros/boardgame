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
	TableGames  = "games"
	TableUsers  = "users"
	TableStates = "states"
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
	//TODO: Add other to DBMap

	if err := s.dbMap.CreateTablesIfNotExists(); err != nil {
		return errors.New("Couldn't create tables: " + err.Error())
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
	return nil, errors.New("Not yet implemented")
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

	return errors.New("Not yet implemented")

}

func (s *StorageManager) ListGames(max int) []*boardgame.GameStorageRecord {
	return nil
}

func (s *StorageManager) SetPlayerForGame(gameId string, playerIndex boardgame.PlayerIndex, userId string) error {

	return errors.New("Not yet implemented")

}

func (s *StorageManager) UserIdsForGame(gameId string) []string {
	return nil
}

func (s *StorageManager) UpdateUser(user *users.StorageRecord) error {
	return errors.New("not yet implemented")
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

	return nil

}

func (s *StorageManager) ConnectCookieToUser(cookie string, user *users.StorageRecord) error {
	return errors.New("Not yet implemented")
}
