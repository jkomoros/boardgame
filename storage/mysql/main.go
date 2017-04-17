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

	s.db = db

	s.dbMap = &gorp.DbMap{
		Db: db,
		Dialect: gorp.MySQLDialect{
			Engine:   "MyISAM",
			Encoding: "utf8",
		},
	}

	s.dbMap.AddTableWithName(UserStorageRecord{}, TableUsers)
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
}

func (s *StorageManager) Name() string {
	return "mysql"
}

func (s *StorageManager) State(gameId string, version int) (boardgame.StateStorageRecord, error) {
	return nil, errors.New("Not yet implemented")
}

func (s *StorageManager) Game(id string) (*boardgame.GameStorageRecord, error) {
	return nil, errors.New("Not yet implemented")
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
	return nil
}

func (s *StorageManager) GetUserByCookie(cookie string) *users.StorageRecord {

	return nil

}

func (s *StorageManager) ConnectCookieToUser(cookie string, user *users.StorageRecord) error {
	return errors.New("Not yet implemented")
}

func (s *StorageManager) CleanUp() {
	//Pass
}
