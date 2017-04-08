/*

bolt provides a bolt-backed database that implements both
boardgame.StorageManager and boardgame/server.StorageManager.

*/
package bolt

import (
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
	"github.com/jkomoros/boardgame"
	"os"
	"strconv"
	"strings"
)

//TODO: test this package

type StorageManager struct {
	db       *bolt.DB
	filename string
}

var (
	statesBucket = []byte("States")
	gamesBucket  = []byte("Games")
)

func NewStorageManager(fileName string) *StorageManager {
	db, err := bolt.Open(fileName, 0600, nil)

	if err != nil {
		panic("Couldn't open db")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(gamesBucket); err != nil {
			return errors.New("Cannot create games bucket" + err.Error())
		}
		if _, err := tx.CreateBucketIfNotExists(statesBucket); err != nil {
			return errors.New("Cannot create states bucket" + err.Error())
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

func keyForState(gameId string, version int) []byte {
	return []byte(gameId + "_" + strconv.Itoa(version))
}

func keyForGame(id string) []byte {
	return []byte(strings.ToUpper(id))
}

func (s *StorageManager) State(gameId string, version int) (boardgame.StateStorageRecord, error) {
	if gameId == "" {
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

		record = b.Get(keyForState(gameId, version))
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

func (s *StorageManager) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord) error {

	version := game.Version

	serializedGameRecord, err := json.Marshal(game)

	if err != nil {
		return errors.New("Couldn't serialize the internal game record: " + err.Error())
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		gBucket := tx.Bucket(gamesBucket)

		if gBucket == nil {
			return errors.New("Couldn't open games bucket")
		}

		sBucket := tx.Bucket(statesBucket)

		if sBucket == nil {
			return errors.New("Could open states bucket")
		}

		if err := gBucket.Put(keyForGame(game.Id), serializedGameRecord); err != nil {
			return err
		}

		if err := sBucket.Put(keyForState(game.Id, version), state); err != nil {
			return err
		}

		return nil

	})

}

func (s *StorageManager) ListGames(managers boardgame.ManagerCollection, max int) []*boardgame.GameStorageRecord {

	var result []*boardgame.GameStorageRecord

	err := s.db.View(func(tx *bolt.Tx) error {

		gBucket := tx.Bucket(gamesBucket)

		if gBucket == nil {
			return errors.New("couldn't open games bucket")
		}

		c := gBucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if len(result) >= max {
				break
			}

			var record boardgame.GameStorageRecord

			if err := json.Unmarshal(v, &record); err != nil {
				return errors.New("Couldn't deserialize a game: " + err.Error())
			}

			manager := managers.Get(record.Name)

			if manager == nil {

				//Hmm, I guess we didn't know about this type of manager...

				//TODO: log an error
				continue
			}

			result = append(result, &record)
		}

		return nil

	})

	if err != nil {
		return nil
	}

	return result

}

func (s *StorageManager) Close() {
	s.db.Close()
}

func (s *StorageManager) CleanUp() {
	os.Remove(s.filename)
}
