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
	"strconv"
	"time"

	"github.com/go-gorp/gorp"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/jkomoros/boardgame/server/api/seatpresentation"
	"github.com/jkomoros/boardgame/server/api/tablelease"
	"github.com/jkomoros/boardgame/server/api/users"
	"github.com/jkomoros/boardgame/storage/mysql/connect"
)

const (
	tableGames             = "games"
	tableExtendedGames     = "extendedgames"
	tableMoves             = "moves"
	tableUsers             = "users"
	tableStates            = "states"
	tableCookies           = "cookies"
	tablePlayers           = "players"
	tableAgentStates       = "agentstates"
	tableSeatPresentations = "seatpresentations"
	tableChatMessages      = "chatmessages"
	tableTableLeases       = "companiontableleases"
)

const baseCombinedSelectQuery = "select g.Name, g.ID, g.SecretSalt, g.Version, g.Winners, g.Finished, g.NumPlayers, g.Agents, " +
	"g.Created, g.Modified, e.Open, e.Visible, e.Owner, e.CompanionRoomCode, e.CompanionLocked"

const baseCombinedFromQuery = "from " + tableGames + " g, " + tableExtendedGames + " e"

const baseCombinedWhereQuery = "where g.ID = e.ID"

const combinedPlayerFilterQuery = baseCombinedSelectQuery + " " + baseCombinedFromQuery + ", players p " + baseCombinedWhereQuery +
	" and p.GameID = g.ID and p.UserID = ?"

const combinedGameStorageRecordQuery = baseCombinedSelectQuery + " " + baseCombinedFromQuery + " " + baseCombinedWhereQuery

const userNotInQuery = "not exists (select * from players where GameID = g.ID and UserID = ?)"

const tableLeaseBaseColumns = "GameID, Generation, DeviceID, SecretDigest, HolderUserID, Expires"
const tableLeaseAllColumns = tableLeaseBaseColumns + ", TransferID, TransferTokenDigest, TransferCodeDigest, TransferExpires, TransferTargetDeviceID, PreviousDeviceID, TransitionKind"

const emptySlotsQuery = "(g.NumPlayers > coalesce(c.NumActivePlayers, 0) + g.NumAgents)"

const combinedHasSlots = baseCombinedSelectQuery + ` from games as g
left join extendedgames as e
	left join (select GameID as ID, count(*) as NumActivePlayers from players group by GameID) as c
	on e.Id = c.Id
on g.Id = e.Id
where`

const combinedNotPlayerFilterQuery = combinedHasSlots + " " + userNotInQuery

const combinedNotPlayerOpenSlotsQuery = combinedNotPlayerFilterQuery + " and " + emptySlotsQuery

const combinedNotPlayerNoOpenSlotsQuery = combinedNotPlayerFilterQuery + " and (not " + emptySlotsQuery + " or e.Open = 0)"

// StorageManager is the primary type in this package.
type StorageManager struct {
	db       *sql.DB
	dbMap    *gorp.DbMap
	testMode bool
	//The config string that we were provided in connect.
	config    string
	connected bool
}

// NewStorageManager returns a new storage manager. Does most of its set-up
// work in Connect(), which is when the database configuration information is
// passed. testMode is whether or not the storage manager is being run in the
// context of a test; if false, then calls to CleanUp (which drops the entire
// database) won't do anything.
func NewStorageManager(testMode bool) *StorageManager {
	//We actually don't do much; we do more of our work in Connect()
	return &StorageManager{
		testMode: testMode,
	}

}

// Connect connects to the database using the given DSN config string.
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

	s.dbMap.AddTableWithName(userStorageRecord{}, tableUsers).SetKeys(false, "ID")
	s.dbMap.AddTableWithName(gameStorageRecord{}, tableGames).SetKeys(false, "ID")
	s.dbMap.AddTableWithName(extendedGameStorageRecord{}, tableExtendedGames).SetKeys(false, "ID")
	s.dbMap.AddTableWithName(stateStorageRecord{}, tableStates).SetKeys(true, "ID")
	s.dbMap.AddTableWithName(cookieStorageRecord{}, tableCookies).SetKeys(false, "Cookie")
	s.dbMap.AddTableWithName(playerStorageRecord{}, tablePlayers).SetKeys(true, "ID")
	s.dbMap.AddTableWithName(agentStateStorageRecord{}, tableAgentStates).SetKeys(true, "ID")
	s.dbMap.AddTableWithName(moveStorageRecord{}, tableMoves).SetKeys(true, "ID")
	s.dbMap.AddTableWithName(chatStorageRecord{}, tableChatMessages).SetKeys(true, "ID")
	s.dbMap.AddTableWithName(seatPresentationStorageRecord{}, tableSeatPresentations).SetKeys(true, "ID")
	s.dbMap.AddTableWithName(tableLeaseStorageRecord{}, tableTableLeases).SetKeys(false, "GameID")

	// Create chat table if it doesn't exist (auto-migration for chat)
	s.dbMap.CreateTablesIfNotExists()

	_, err = s.dbMap.SelectInt("select count(*) from " + tableGames)

	if err != nil {
		return errors.New("Sanity check failed for db. Have you used the admin tool to migrate it up? " + err.Error())
	}

	s.connected = true

	return nil

}

// Close closes out the connection to the database.
func (s *StorageManager) Close() {
	if s.db == nil {
		return
	}
	s.db.Close()
	s.db = nil
	s.dbMap = nil
	s.connected = false
}

// CleanUp drops the test DB, but only if it was created in TestMode.
func (s *StorageManager) CleanUp() {
	if !s.testMode {
		return
	}
	//connect will refuse to drop the db if it's not the test db name.
	connect.DropTestDb(s.config)
}

// Name returns 'mysql'
func (s *StorageManager) Name() string {
	return "mysql"
}

// State returns the given state
func (s *StorageManager) State(gameID string, version int) (boardgame.StateStorageRecord, error) {

	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}

	var state stateStorageRecord

	err := s.dbMap.SelectOne(&state, "select * from "+tableStates+" where GameID=? and Version=?", gameID, version)

	if err == sql.ErrNoRows {
		return nil, errors.New("No such state")
	}

	if err != nil {
		return nil, errors.New("Unexpected error: " + err.Error())
	}

	return (&state).ToStorageRecord(), nil
}

// Moves returns the given moves
func (s *StorageManager) Moves(gameID string, fromVersion, toVersion int) ([]*boardgame.MoveStorageRecord, error) {

	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}

	var moves []*moveStorageRecord

	if fromVersion == toVersion {
		fromVersion = fromVersion - 1
	}

	_, err := s.dbMap.Select(&moves, "select * from "+tableMoves+" where GameID=? and Version>? and Version<=? order by Version", gameID, fromVersion, toVersion)

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

// Move returns the given Move
func (s *StorageManager) Move(gameID string, version int) (*boardgame.MoveStorageRecord, error) {
	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}

	var move moveStorageRecord

	err := s.dbMap.SelectOne(&move, "select * from "+tableMoves+" where GameID=? and Version=?", gameID, version)

	if err == sql.ErrNoRows {
		return nil, errors.New("No such state")
	}

	if err != nil {
		return nil, errors.New("Unexpected error: " + err.Error())
	}

	return (&move).ToStorageRecord(), nil
}

// Game returns the given Game
func (s *StorageManager) Game(id string) (*boardgame.GameStorageRecord, error) {

	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}

	var game gameStorageRecord

	err := s.dbMap.SelectOne(&game, "select * from "+tableGames+" where ID=?", id)

	if err == sql.ErrNoRows {
		return nil, errors.New("No such game")
	}

	if err != nil {
		return nil, errors.New("Unexpected error: " + err.Error())
	}

	return (&game).ToStorageRecord(), nil
}

// ExtendedGame returns the given ExtendedGame
func (s *StorageManager) ExtendedGame(id string) (*extendedgame.StorageRecord, error) {
	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}

	var record extendedGameStorageRecord

	err := s.dbMap.SelectOne(&record, "select * from "+tableExtendedGames+" where ID=?", id)

	if err != nil {
		return nil, err
	}

	return (&record).ToStorageRecord(), nil
}

// CombinedGame returns the given CombinedGame
func (s *StorageManager) CombinedGame(id string) (*extendedgame.CombinedStorageRecord, error) {

	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}

	var record combinedGameStorageRecord

	err := s.dbMap.SelectOne(&record, combinedGameStorageRecordQuery+" and g.ID = ?", id)

	if err != nil {
		return nil, err
	}

	return (&record).ToStorageRecord(), nil
}

// SeatPresentation looks up the per-(gameID, playerIndex) presentation row.
func (s *StorageManager) SeatPresentation(gameID string, playerIndex boardgame.PlayerIndex) (*seatpresentation.StorageRecord, error) {
	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}
	var rec seatPresentationStorageRecord
	err := s.dbMap.SelectOne(&rec, "select * from "+tableSeatPresentations+" where GameID = ? and PlayerIndex = ?", gameID, int64(playerIndex))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return rec.ToStorageRecord(), nil
}

// SetSeatPresentation upserts the seat presentation row.
func (s *StorageManager) SetSeatPresentation(rec *seatpresentation.StorageRecord) error {
	if !s.connected {
		return errors.New("Database not connected yet")
	}
	if rec == nil {
		return errors.New("nil seat presentation record")
	}
	// Look for an existing row to update; insert if none.
	var existing seatPresentationStorageRecord
	err := s.dbMap.SelectOne(&existing, "select * from "+tableSeatPresentations+" where GameID = ? and PlayerIndex = ?", rec.GameID, int64(rec.PlayerIndex))
	if err == sql.ErrNoRows {
		newRow := newSeatPresentationStorageRecord(rec)
		return s.dbMap.Insert(newRow)
	}
	if err != nil {
		return err
	}
	existing.DisplayName = rec.DisplayName
	existing.AvatarSlug = rec.AvatarSlug
	_, err = s.dbMap.Update(&existing)
	return err
}

// ClearSeatPresentation removes the row for (gameID, playerIndex). No error
// if no row exists.
func (s *StorageManager) ClearSeatPresentation(gameID string, playerIndex boardgame.PlayerIndex) error {
	if !s.connected {
		return errors.New("Database not connected yet")
	}
	_, err := s.dbMap.Exec("delete from "+tableSeatPresentations+" where GameID = ? and PlayerIndex = ?", gameID, int64(playerIndex))
	return err
}

// CompanionTableLease implements the server storage interface.
func (s *StorageManager) CompanionTableLease(gameID string) (*tablelease.StorageRecord, error) {
	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}
	transferColumns, err := s.companionTableTransferColumnsAvailable()
	if err != nil {
		return nil, err
	}
	var record tableLeaseStorageRecord
	columns := tableLeaseBaseColumns
	if transferColumns {
		columns = tableLeaseAllColumns
	}
	if err := s.dbMap.SelectOne(&record, "select "+columns+" from "+tableTableLeases+" where GameID = ?", gameID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return record.ToStorageRecord(), nil
}

func (s *StorageManager) companionTableTransferColumnsAvailable() (bool, error) {
	var count int
	err := s.db.QueryRow("select count(*) from information_schema.columns where table_schema = database() and table_name = ? and column_name = 'TransferID'", tableTableLeases).Scan(&count)
	return count == 1, err
}

func tableLeaseUsesTransferColumns(record *tablelease.StorageRecord) bool {
	return record != nil && (record.TransferID != "" || record.TransferTokenDigest != "" || record.TransferCodeDigest != "" ||
		record.TransferExpires != 0 || record.TransferTargetDeviceID != "" || record.PreviousDeviceID != "" || record.TransitionKind != "")
}

// CompareAndSwapCompanionTableLease implements the server storage interface.
func (s *StorageManager) CompareAndSwapCompanionTableLease(gameID string, expectedGeneration uint64, replacement *tablelease.StorageRecord) (*tablelease.StorageRecord, bool, error) {
	if !s.connected {
		return nil, false, errors.New("Database not connected yet")
	}
	if gameID == "" {
		return nil, false, errors.New("empty game ID")
	}
	if replacement == nil {
		return nil, false, errors.New("nil companion Table lease replacement")
	}
	if replacement.GameID != "" && replacement.GameID != gameID {
		return nil, false, errors.New("companion Table lease game ID mismatch")
	}
	if err := replacement.ValidateTransfer(); err != nil {
		return nil, false, err
	}
	if expectedGeneration == ^uint64(0) {
		return nil, false, errors.New("companion Table lease generation exhausted")
	}

	next := replacement.Clone()
	next.GameID = gameID
	next.Generation = expectedGeneration + 1
	stored := newTableLeaseStorageRecord(next)
	transferColumns, err := s.companionTableTransferColumnsAvailable()
	if err != nil {
		return nil, false, err
	}
	if !transferColumns && tableLeaseUsesTransferColumns(next) {
		return nil, false, errors.New("companion Table transfer migration 0023 is not applied")
	}

	if expectedGeneration == 0 {
		var insertErr error
		if transferColumns {
			insertErr = s.dbMap.Insert(stored)
		} else {
			_, insertErr = s.db.Exec("insert into "+tableTableLeases+" ("+tableLeaseBaseColumns+") values (?, ?, ?, ?, ?, ?)",
				next.GameID, next.Generation, next.DeviceID, next.SecretDigest, next.HolderUserID, next.Expires)
		}
		if insertErr == nil {
			return next.Clone(), true, nil
		} else {
			// A concurrent insert is the expected losing path. Distinguish it
			// from infrastructure errors by requiring the winning row to exist.
			current, readErr := s.CompanionTableLease(gameID)
			if readErr != nil || current == nil {
				return nil, false, insertErr
			}
			return current, false, nil
		}
	}

	var result sql.Result
	if transferColumns {
		result, err = s.dbMap.Exec("update "+tableTableLeases+" set Generation = ?, DeviceID = ?, SecretDigest = ?, HolderUserID = ?, Expires = ?, TransferID = ?, TransferTokenDigest = ?, TransferCodeDigest = ?, TransferExpires = ?, TransferTargetDeviceID = ?, PreviousDeviceID = ?, TransitionKind = ? where GameID = ? and Generation = ?",
			next.Generation, next.DeviceID, next.SecretDigest, next.HolderUserID, next.Expires,
			next.TransferID, next.TransferTokenDigest, next.TransferCodeDigest, next.TransferExpires,
			next.TransferTargetDeviceID, next.PreviousDeviceID, next.TransitionKind,
			gameID, expectedGeneration)
	} else {
		result, err = s.dbMap.Exec("update "+tableTableLeases+" set Generation = ?, DeviceID = ?, SecretDigest = ?, HolderUserID = ?, Expires = ? where GameID = ? and Generation = ?",
			next.Generation, next.DeviceID, next.SecretDigest, next.HolderUserID, next.Expires, gameID, expectedGeneration)
	}
	if err != nil {
		return nil, false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, false, err
	}
	if rows == 1 {
		return next.Clone(), true, nil
	}
	current, err := s.CompanionTableLease(gameID)
	return current, false, err
}

// GameByRoomCode looks up a gameID by CompanionRoomCode. Returns "" with
// a nil error if no match (caller treats as 404). Empty code short-circuits
// to "" with no DB query.
func (s *StorageManager) GameByRoomCode(code string) (string, error) {
	if !s.connected {
		return "", errors.New("Database not connected yet")
	}
	if code == "" {
		return "", nil
	}

	var id string
	err := s.dbMap.SelectOne(&id, "select ID from "+tableExtendedGames+" where CompanionRoomCode = ?", code)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

// SaveGameAndCurrentState saves the given game and current state.
func (s *StorageManager) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {

	if !s.connected {
		return errors.New("Database not connected yet")
	}

	version := game.Version

	gameRecord := newGameStorageRecord(game)
	stateRecord := newStateStorageRecord(game.ID, version, state)

	var moveRecord *moveStorageRecord

	if move != nil {
		moveRecord = newMoveStorageRecord(game.ID, version, move)
	}

	count, _ := s.dbMap.SelectInt("select count(*) from "+tableGames+" where ID=?", game.ID)

	if count < 1 {
		//Need to insert
		err := s.dbMap.Insert(gameRecord)

		if err != nil {
			return errors.New("Couldn't update game: " + err.Error())
		}

		extendedRecord := newExtendedGameStorageRecord(extendedgame.DefaultStorageRecord())

		extendedRecord.ID = game.ID

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

// AgentState returns the given AgentState
func (s *StorageManager) AgentState(gameID string, player boardgame.PlayerIndex) ([]byte, error) {

	if !s.connected {
		return nil, errors.New("Database not connected yet")
	}

	var agent agentStateStorageRecord

	err := s.dbMap.SelectOne(&agent, "select * from "+tableAgentStates+" where GameID=? and PlayerIndex=? order by ID desc limit 1", gameID, int64(player))

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return agent.ToStorageRecord(), nil

}

// SaveAgentState saves the given agent state
func (s *StorageManager) SaveAgentState(gameID string, player boardgame.PlayerIndex, state []byte) error {
	if !s.connected {
		return errors.New("Database not connected yet")
	}

	record := newAgentStateStorageRecord(gameID, player, state)

	err := s.dbMap.Insert(record)

	if err != nil {
		return errors.New("Couldn't save record: " + err.Error())
	}

	return nil
}

// UpdateExtendedGame updates the given extended game properties
func (s *StorageManager) UpdateExtendedGame(id string, eGame *extendedgame.StorageRecord) error {

	if !s.connected {
		return errors.New("Database not connected yet")
	}

	record := newExtendedGameStorageRecord(eGame)
	record.ID = id

	_, err := s.dbMap.Update(record)

	return err
}

// ListGames lists the given games
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

// SetPlayerForGame affiliates the given user in the given game to the given player
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
	err = s.dbMap.SelectOne(&player, "select * from "+tablePlayers+" where GameID=? and UserID=?", game.ID, userID)
	if err == nil {
		if player.PlayerIndex == int64(playerIndex) {
			return nil
		}
		return errors.New("That user is already assigned to another seat in this game")
	}
	if err != sql.ErrNoRows {
		return errors.New("Failed to check existing user assignment: " + err.Error())
	}

	err = s.dbMap.SelectOne(&player, "select * from "+tablePlayers+" where GameID=? and PlayerIndex=?", game.ID, int(playerIndex))

	if err == sql.ErrNoRows {
		// Insert the row

		player = playerStorageRecord{
			GameID:      game.ID,
			PlayerIndex: int64(playerIndex),
			UserID:      userID,
		}

		err = s.dbMap.Insert(&player)

		if err != nil {
			return errors.New("Couldn't insert new player line: " + err.Error())
		}

		return nil
	}

	if err != nil {
		return errors.New("Failed to retrieve existing Player line: " + err.Error())
	}
	if player.UserID == userID {
		return nil
	}
	return errors.New("PlayerIndex " + playerIndex.String() + " is already taken")

}

// UserIDsForGame returns the given UserIds
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

	_, err = s.dbMap.Select(&players, "select * from "+tablePlayers+" where GameID=? order by PlayerIndex desc", game.ID)

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

		result[index] = rec.UserID
	}

	return result

}

// UpdateUser updates the given user
func (s *StorageManager) UpdateUser(user *users.StorageRecord) error {
	userRecord := newUserStorageRecord(user)

	existingRecord, _ := s.dbMap.SelectInt("select count(*) from "+tableUsers+" where ID=?", user.ID)

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

// GetUserByID gets the given user
func (s *StorageManager) GetUserByID(uid string) *users.StorageRecord {
	if !s.connected {
		return nil
	}

	var user userStorageRecord

	err := s.dbMap.SelectOne(&user, "select * from "+tableUsers+" where ID=?", uid)

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

// GetUserByCookie gets the given user
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

	return s.GetUserByID(cookieRecord.UserID)

}

// ConnectCookieToUser affiliates the given cookie to the given user
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
		UserID: user.ID,
	}

	if err := s.dbMap.Insert(record); err != nil {
		return errors.New("Failed to insert cookie pointer record: " + err.Error())
	}
	return nil
}

// PlayerMoveApplied does nothing
func (s *StorageManager) PlayerMoveApplied(game *boardgame.GameStorageRecord) error {
	//Don't need to do anything
	return nil
}

// FetchInjectedDataForGame can just return nil
func (s *StorageManager) FetchInjectedDataForGame(gameID string, dataType string) interface{} {
	//Don't need to do anything
	return nil
}

// WithManagers does nothing
func (s *StorageManager) WithManagers(managers []*boardgame.GameManager) {
	//Do nothing
}

// SaveChatMessage implements boardgame.ChatStorageManager.
func (s *StorageManager) SaveChatMessage(msg *boardgame.ChatMessage) error {
	rec := &chatStorageRecord{
		GameID:    msg.GameID,
		Version:   int64(msg.Version),
		Sender:    int64(msg.Sender),
		Channel:   msg.Channel,
		Body:      msg.Body,
		Timestamp: msg.Timestamp.UnixMilli(),
	}
	if err := s.dbMap.Insert(rec); err != nil {
		return err
	}
	msg.ID = strconv.FormatInt(rec.ID, 10)
	return nil
}

// ChatMessages implements boardgame.ChatStorageManager.
func (s *StorageManager) ChatMessages(gameID string, channel string, sinceID string, limit int) ([]*boardgame.ChatMessage, error) {
	query := "select * from " + tableChatMessages + " where GameID = ?"
	args := []interface{}{gameID}

	if channel != "" {
		query += " and Channel = ?"
		args = append(args, channel)
	}

	if sinceID != "" {
		sinceIDInt, err := strconv.ParseInt(sinceID, 10, 64)
		if err == nil {
			query += " and ID > ?"
			args = append(args, sinceIDInt)
		}
	}

	query += " order by ID asc"

	if limit > 0 {
		query += " limit ?"
		args = append(args, limit)
	}

	var records []chatStorageRecord
	if _, err := s.dbMap.Select(&records, query, args...); err != nil {
		return nil, err
	}

	var result []*boardgame.ChatMessage
	for _, rec := range records {
		result = append(result, &boardgame.ChatMessage{
			ID:        strconv.FormatInt(rec.ID, 10),
			GameID:    rec.GameID,
			Version:   int(rec.Version),
			Sender:    boardgame.PlayerIndex(rec.Sender),
			Channel:   rec.Channel,
			Body:      rec.Body,
			Timestamp: time.UnixMilli(rec.Timestamp),
		})
	}

	return result, nil
}
