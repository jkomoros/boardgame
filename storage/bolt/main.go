/*

Package bolt provides a bolt-backed database that implements both
boardgame.StorageManager and boardgame/server.StorageManager.

*/
package bolt

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/listing"
	"github.com/jkomoros/boardgame/server/api/users"
	"github.com/jkomoros/boardgame/storage/internal/helpers"
)

//TODO: test this package

//StorageManager is the main object that implements StorageManager and
//api.StorageManager. Get a new one from NewStorageManager.
type StorageManager struct {
	db       *bolt.DB
	filename string
}

var (
	statesBucket        = []byte("States")
	movesBucket         = []byte("Moves")
	gamesBucket         = []byte("Games")
	extendedGamesBucket = []byte("ExtendedGames")
	usersBucket         = []byte("Users")
	cookiesBucket       = []byte("Cookies")
	gameUsersBucket     = []byte("GameUsers")
	agentStatesBucket   = []byte("AgentStates")
)

//NewStorageManager returns a new StorageManager ready for use, backed by the
//boltDB that exists at that filename, relative to current directory. Will
//create that file if it doesn't exist.
func NewStorageManager(fileName string) *StorageManager {
	db, err := bolt.Open(fileName, 0600, nil)

	if err != nil {
		panic("Couldn't open db")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(gamesBucket); err != nil {
			return errors.New("Cannot create games bucket" + err.Error())
		}
		if _, err := tx.CreateBucketIfNotExists(extendedGamesBucket); err != nil {
			return errors.New("Could not create extended games bucket" + err.Error())
		}
		if _, err := tx.CreateBucketIfNotExists(statesBucket); err != nil {
			return errors.New("Cannot create states bucket" + err.Error())
		}
		if _, err := tx.CreateBucketIfNotExists(movesBucket); err != nil {
			return errors.New("Cannot create moves bucket" + err.Error())
		}
		if _, err := tx.CreateBucketIfNotExists(usersBucket); err != nil {
			return errors.New("Cannot create users bucket" + err.Error())
		}
		if _, err := tx.CreateBucketIfNotExists(cookiesBucket); err != nil {
			return errors.New("Cannot create cookies bucket" + err.Error())
		}
		if _, err := tx.CreateBucketIfNotExists(gameUsersBucket); err != nil {
			return errors.New("Cannot create game users bucket" + err.Error())
		}
		if _, err := tx.CreateBucketIfNotExists(agentStatesBucket); err != nil {
			return errors.New("Cannot create agent states bucket" + err.Error())
		}
		return nil
	})

	if err != nil {
		//Not able to initalize DB
		return nil
	}
	//We don't defer DB close; our users need to.
	return &StorageManager{
		db:       db,
		filename: fileName,
	}

}

func keyForState(gameID string, version int) []byte {
	return []byte(gameID + "_" + strconv.Itoa(version))
}

func keyForMove(gameID string, version int) []byte {
	return []byte(gameID + "_" + strconv.Itoa(version))
}

func keyForGame(id string) []byte {
	return []byte(strings.ToUpper(id))
}

func keyForUser(uid string) []byte {
	return []byte(uid)
}

func keyForCookie(cookie string) []byte {
	return []byte(cookie)
}

func keyForAgentState(gameID string, player boardgame.PlayerIndex) []byte {
	return []byte(gameID + "-" + player.String())
}

//Name returns 'bolt'
func (s *StorageManager) Name() string {
	return "bolt"
}

//State implements that method from the main storagemanager interface
func (s *StorageManager) State(gameID string, version int) (boardgame.StateStorageRecord, error) {
	if gameID == "" {
		return nil, errors.New("No game provided")
	}

	if version < 0 {
		return nil, errors.New("Invalid version")
	}

	var record []byte

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(statesBucket)

		if b == nil {
			return errors.New("Couldn't get bucket")
		}

		record = b.Get(keyForState(gameID, version))
		return nil
	})

	if err != nil {
		return nil, err
	}

	if record == nil {
		return nil, errors.New("No such version for game")
	}

	return record, nil
}

//Moves implements that method from the main storagemanager interface
func (s *StorageManager) Moves(gameID string, fromVersion, toVersion int) ([]*boardgame.MoveStorageRecord, error) {
	return helpers.MovesHelper(s, gameID, fromVersion, toVersion)
}

//Move implements that method from the main storagemanager interface
func (s *StorageManager) Move(gameID string, version int) (*boardgame.MoveStorageRecord, error) {
	if gameID == "" {
		return nil, errors.New("No game provided")
	}

	if version < 0 {
		return nil, errors.New("Invalid version")
	}

	var record []byte

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(movesBucket)

		if b == nil {
			return errors.New("Couldn't get bucket")
		}

		record = b.Get(keyForMove(gameID, version))
		return nil
	})

	if err != nil {
		return nil, err
	}

	if record == nil {
		return nil, errors.New("No such version (" + strconv.Itoa(version) + ") for game " + gameID)
	}

	var result boardgame.MoveStorageRecord

	if err := json.Unmarshal(record, &result); err != nil {
		return nil, errors.New("Couldn't unmarshal internal blob: " + err.Error())
	}

	return &result, nil
}

//Game implements that method from the main storagemanager interface
func (s *StorageManager) Game(id string) (*boardgame.GameStorageRecord, error) {

	var rawRecord []byte

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(gamesBucket)
		if b == nil {
			return errors.New("Couldn't open bucket")
		}
		rawRecord = b.Get(keyForGame(id))
		return nil
	})

	if err != nil {
		return nil, errors.New("Transacation error " + err.Error())
	}

	if rawRecord == nil {
		return nil, errors.New("No such game found")
	}

	var record boardgame.GameStorageRecord

	if err := json.Unmarshal(rawRecord, &record); err != nil {
		return nil, errors.New("Unmarshal error " + err.Error())
	}

	return &record, nil

}

//SaveGameAndCurrentState implements that method from the main storagemanager interface
func (s *StorageManager) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {

	version := game.Version

	serializedGameRecord, err := json.Marshal(game)

	if err != nil {
		return errors.New("Couldn't serialize the internal game record: " + err.Error())
	}

	previousGame, _ := s.Game(game.ID)

	eGame := extendedgame.DefaultStorageRecord()

	if previousGame != nil {
		//This is not a new game!
		eGame, err = s.ExtendedGame(game.ID)
		if err != nil {
			return errors.New("Couldnt' find extended game for an already created game: " + err.Error())
		}
	}

	serializedExtendedGameRecord, err := json.Marshal(eGame)

	if err != nil {
		return errors.New("Couldn't serialize the internal extended game record: " + err.Error())
	}

	var serializedMoveRecord []byte

	if move != nil {
		serializedMoveRecord, err = json.Marshal(move)
	}

	if err != nil {
		return errors.New("Couldn't serialize the internal move record: " + err.Error())
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		gBucket := tx.Bucket(gamesBucket)

		if gBucket == nil {
			return errors.New("Couldn't open games bucket")
		}

		mBucket := tx.Bucket(movesBucket)

		if mBucket == nil {
			return errors.New("Couldn't open moves bucket")
		}

		sBucket := tx.Bucket(statesBucket)

		if sBucket == nil {
			return errors.New("Could open states bucket")
		}

		eBucket := tx.Bucket(extendedGamesBucket)

		if eBucket == nil {
			return errors.New("Couldn't open extended games bucket")
		}

		if err := gBucket.Put(keyForGame(game.ID), serializedGameRecord); err != nil {
			return err
		}

		if err := sBucket.Put(keyForState(game.ID, version), state); err != nil {
			return err
		}

		if err := eBucket.Put(keyForGame(game.ID), serializedExtendedGameRecord); err != nil {
			return err
		}

		if serializedMoveRecord != nil {
			if err := mBucket.Put(keyForMove(game.ID, version), serializedMoveRecord); err != nil {
				return err
			}

		}

		return nil

	})

}

//AgentState implements that method from the main storagemanager interface
func (s *StorageManager) AgentState(gameID string, player boardgame.PlayerIndex) ([]byte, error) {

	var result []byte

	err := s.db.View(func(tx *bolt.Tx) error {

		aBucket := tx.Bucket(agentStatesBucket)

		if aBucket == nil {
			return errors.New("Couldn't open agent states bucket")
		}

		result = aBucket.Get(keyForAgentState(gameID, player))
		return nil

	})

	if err != nil {
		return nil, err
	}

	return result, nil

}

//SaveAgentState implements that method from the main storagemanager interface
func (s *StorageManager) SaveAgentState(gameID string, player boardgame.PlayerIndex, state []byte) error {

	return s.db.Update(func(tx *bolt.Tx) error {
		aBucket := tx.Bucket(agentStatesBucket)

		if aBucket == nil {
			return errors.New("Couldn't open agent states bucket")
		}

		if err := aBucket.Put(keyForAgentState(gameID, player), state); err != nil {
			return err
		}
		return nil
	})

}

//AllGames implements the extra method necessary for storage/internal/helpers
func (s *StorageManager) AllGames() []*boardgame.GameStorageRecord {
	var results []*boardgame.GameStorageRecord

	err := s.db.View(func(tx *bolt.Tx) error {

		gBucket := tx.Bucket(gamesBucket)

		if gBucket == nil {
			return errors.New("couldn't open games bucket")
		}
		c := gBucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {

			var record boardgame.GameStorageRecord

			if err := json.Unmarshal(v, &record); err != nil {
				return errors.New("Couldn't deserialize a game: " + err.Error())
			}

			results = append(results, &record)
		}

		return nil

	})

	if err != nil {
		return nil
	}

	return results
}

//ListGames implements that method from the server api storagemanager interface
func (s *StorageManager) ListGames(max int, list listing.Type, userID string, gameType string) []*extendedgame.CombinedStorageRecord {
	return helpers.ListGamesHelper(s, max, list, userID, gameType)
}

//ExtendedGame implements that method from the server api storagemanager interface
func (s *StorageManager) ExtendedGame(id string) (*extendedgame.StorageRecord, error) {

	var rawRecord []byte

	err := s.db.View(func(tx *bolt.Tx) error {
		eBucket := tx.Bucket(extendedGamesBucket)

		if eBucket == nil {
			return errors.New("Couldn't open extended games bucket")
		}

		rawRecord = eBucket.Get(keyForGame(id))

		return nil
	})

	if err != nil {
		return nil, errors.New("Couldn't get raw record: " + err.Error())
	}

	if rawRecord == nil {
		return nil, errors.New("No such extended game found")
	}

	var eGame *extendedgame.StorageRecord

	if err = json.Unmarshal(rawRecord, &eGame); err != nil {
		return nil, errors.New("Couldn't unmarshal record: " + err.Error())
	}

	return eGame, nil
}

//CombinedGame implements that method from the server api storagemanager interface
func (s *StorageManager) CombinedGame(id string) (*extendedgame.CombinedStorageRecord, error) {

	game, err := s.Game(id)

	if err != nil {
		return nil, errors.New("Couldn't get game: " + err.Error())
	}

	var rawRecord []byte

	err = s.db.View(func(tx *bolt.Tx) error {
		eBucket := tx.Bucket(extendedGamesBucket)

		if eBucket == nil {
			return errors.New("Couldn't open extended games bucket")
		}

		rawRecord = eBucket.Get(keyForGame(id))

		return nil
	})

	if err != nil {
		return nil, errors.New("Couldn't get raw record: " + err.Error())
	}

	if rawRecord == nil {
		return nil, errors.New("No such extended game found")
	}

	var eGame *extendedgame.StorageRecord

	if err = json.Unmarshal(rawRecord, &eGame); err != nil {
		return nil, errors.New("Couldn't unmarshal record: " + err.Error())
	}

	return &extendedgame.CombinedStorageRecord{
		GameStorageRecord: *game,
		StorageRecord:     *eGame,
	}, nil
}

//UpdateExtendedGame implements that method from the server api storagemanager interface
func (s *StorageManager) UpdateExtendedGame(id string, eGame *extendedgame.StorageRecord) error {

	serializedExtendedGameRecord, err := json.Marshal(eGame)

	if err != nil {
		return errors.New("couldn't serialize record: " + err.Error())
	}

	err = s.db.Update(func(tx *bolt.Tx) error {

		eBucket := tx.Bucket(extendedGamesBucket)

		if eBucket == nil {
			return errors.New("Couldn't open extended games bucket")
		}

		return eBucket.Put(keyForGame(id), serializedExtendedGameRecord)
	})

	if err != nil {
		return errors.New("Couldn't save extended game: " + err.Error())
	}

	return nil

}

//SetPlayerForGame implements that method from the server api storagemanager interface
func (s *StorageManager) SetPlayerForGame(gameID string, playerIndex boardgame.PlayerIndex, userID string) error {

	ids := s.UserIDsForGame(gameID)

	if ids == nil {
		return errors.New("Couldn't fetch original player indexes for that game")
	}

	if int(playerIndex) < 0 || int(playerIndex) >= len(ids) {
		return errors.New("PlayerIndex " + playerIndex.String() + " is not valid for this game")
	}

	if ids[playerIndex] != "" {
		return errors.New("PlayerIndex " + playerIndex.String() + " is already taken")
	}

	user := s.GetUserByID(userID)

	if user == nil {
		return errors.New("That userId does not describe an existing user")
	}

	ids[playerIndex] = userID

	err := s.db.Update(func(tx *bolt.Tx) error {
		gUBucket := tx.Bucket(gameUsersBucket)

		if gUBucket == nil {
			return errors.New("Couldn't open game useres bucket")
		}

		blob, err := json.Marshal(ids)

		if err != nil {
			return errors.New("Unable to marshal ids blob: " + err.Error())
		}

		return gUBucket.Put(keyForGame(gameID), blob)
	})

	if err != nil {
		return errors.New("Unable to form association: " + err.Error())
	}

	return nil

}

//UserIDsForGame implements that method from the server api storagemanager
//interface
func (s *StorageManager) UserIDsForGame(gameID string) []string {

	noRecordErr := errors.New("No such record")

	var result []string

	err := s.db.View(func(tx *bolt.Tx) error {
		gUBucket := tx.Bucket(gameUsersBucket)

		if gUBucket == nil {
			return errors.New("Couldn't open game users bucket")
		}

		blob := gUBucket.Get(keyForGame(gameID))

		if blob == nil {
			//NO such game info.
			return noRecordErr
		}

		return json.Unmarshal(blob, &result)
	})

	if err == noRecordErr {
		//It's possible that we just haven't stored anything for this user before.

		gameRecord, err := s.Game(gameID)

		if err != nil {
			log.Println("Couldn fetch game: " + err.Error())
			return nil
		}

		if gameRecord == nil {
			log.Println("No such game")
			return nil
		}

		return make([]string, gameRecord.NumPlayers)
	}

	if err != nil {
		log.Println("Error in UserIdsForGame: ", err)
		return nil
	}

	return result
}

//UpdateUser implements that method from the server api storagemanager interface
func (s *StorageManager) UpdateUser(user *users.StorageRecord) error {
	err := s.db.Update(func(tx *bolt.Tx) error {

		uBucket := tx.Bucket(usersBucket)

		if uBucket == nil {
			return errors.New("couldn't open users bucket")
		}

		blob, err := json.Marshal(user)

		if err != nil {
			return errors.New("Couldn't marshal user: " + err.Error())
		}

		return uBucket.Put(keyForUser(user.ID), blob)

	})

	return err
}

//GetUserByID implements that method from the server api storagemanager interface
func (s *StorageManager) GetUserByID(uid string) *users.StorageRecord {
	var result users.StorageRecord

	err := s.db.View(func(tx *bolt.Tx) error {
		uBucket := tx.Bucket(usersBucket)

		if uBucket == nil {
			return errors.New("Couldn't open users bucket")
		}

		uBlob := uBucket.Get(keyForUser(uid))

		if uBlob == nil {
			return errors.New("No such user")
		}

		return json.Unmarshal(uBlob, &result)
	})

	if err != nil {
		log.Println("Failure in GetUserById: ", err)
		return nil
	}

	return &result
}

//GetUserByCookie implements that method from the server api storagemanager
//interface
func (s *StorageManager) GetUserByCookie(cookie string) *users.StorageRecord {

	var result users.StorageRecord

	err := s.db.View(func(tx *bolt.Tx) error {

		cBucket := tx.Bucket(cookiesBucket)

		if cBucket == nil {
			return errors.New("Couldn't open cookies bucket")
		}

		c := cBucket.Get(keyForCookie(cookie))

		if c == nil {
			return errors.New("No such cookie: " + cookie)
		}

		uBucket := tx.Bucket(usersBucket)

		if uBucket == nil {
			return errors.New("couldn't open users bucket")
		}

		uBlob := uBucket.Get(keyForUser(string(c)))

		if uBlob == nil {
			return errors.New("The user specified by cookie was not found")
		}

		if err := json.Unmarshal(uBlob, &result); err != nil {
			return errors.New("Unable to unmarshal user objet: " + err.Error())
		}

		return nil

	})

	if err != nil {
		log.Println("Failure in GetUserByCookie", err)
		return nil
	}

	return &result

}

//ConnectCookieToUser implements that method from the server api storagemanager
//interface
func (s *StorageManager) ConnectCookieToUser(cookie string, user *users.StorageRecord) error {
	err := s.db.Update(func(tx *bolt.Tx) error {

		cBucket := tx.Bucket(cookiesBucket)

		if cBucket == nil {
			return errors.New("couldn't open cookies bucket")
		}

		if user == nil {
			//Delete the cookie.
			return cBucket.Delete(keyForCookie(cookie))
		}

		return cBucket.Put(keyForCookie(cookie), keyForUser(user.ID))

	})

	return err
}

//Connect is a no op
func (s *StorageManager) Connect(config string) error {
	return nil
}

//Close closes the handle to the bolt db
func (s *StorageManager) Close() {
	s.db.Close()
}

//CleanUp removes the backing file
func (s *StorageManager) CleanUp() {
	os.Remove(s.filename)
}

//PlayerMoveApplied is a noop
func (s *StorageManager) PlayerMoveApplied(game *boardgame.GameStorageRecord) error {
	//Don't need to do anything
	return nil
}

//WithManagers  does nothing.
func (s *StorageManager) WithManagers(managers []*boardgame.GameManager) {
	//Do nothing
}
