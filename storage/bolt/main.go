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

//gameRecord is suitable for being marshaled as JSON
type gameRecord struct {
	Name     string
	Id       string
	Version  int
	Finished bool
	Winners  []int
}

//stateRecord is suitable for being marshaled as JSON
type stateRecord struct {
	Version int
	State   []byte
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

func keyForState(game *boardgame.Game, version int) []byte {
	return []byte(game.Id() + "_" + strconv.Itoa(version))
}

func keyForGame(id string) []byte {
	return []byte(strings.ToUpper(id))
}

func (s *StorageManager) State(game *boardgame.Game, version int) boardgame.State {
	if game == nil {
		return nil
	}

	if version < 0 || version > game.Version() {
		return nil
	}

	var record []byte

	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(statesBucket)

		if b == nil {
			return errors.New("Couldn't get bucket")
		}

		record = b.Get(keyForState(game, version))
		return nil
	})

	if record == nil {
		return nil
	}

	//Let's try to deserialize!

	var decodedRecord stateRecord

	if err := json.Unmarshal(record, &decodedRecord); err != nil {
		return nil
	}

	state, err := game.Manager().Delegate().StateFromBlob(decodedRecord.State)

	if err != nil {
		return nil
	}

	return state
}

func (s *StorageManager) Game(manager *boardgame.GameManager, id string) *boardgame.Game {

	if manager == nil {
		return nil
	}

	var rawRecord []byte

	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(gamesBucket)
		if b == nil {
			return errors.New("Couldn't open bucket")
		}
		rawRecord = b.Get(keyForGame(id))
		return nil
	})

	if rawRecord == nil {
		return nil
	}

	var record gameRecord

	if err := json.Unmarshal(rawRecord, &record); err != nil {
		return nil
	}

	return manager.LoadGame(record.Name, record.Id, record.Version, record.Finished, record.Winners)

}

func (s *StorageManager) SaveGameAndState(game *boardgame.Game, version int, state boardgame.State) error {

	gameRec := &gameRecord{
		Name:     game.Name(),
		Id:       game.Id(),
		Version:  game.Version(),
		Finished: game.Finished(),
		Winners:  game.Winners(),
	}

	serializedState, err := json.Marshal(state)

	if err != nil {
		return errors.New("Could not serialize state: " + err.Error())
	}

	stateRec := &stateRecord{
		Version: version,
		State:   serializedState,
	}

	serializedGameRecord, err := json.Marshal(gameRec)

	if err != nil {
		return errors.New("Couldn't serialize the internal game record: " + err.Error())
	}

	serializedStateRecord, err := json.Marshal(stateRec)

	if err != nil {
		return errors.New("Couldn't serialize the internal state record: " + err.Error())
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

		if err := gBucket.Put(keyForGame(game.Id()), serializedGameRecord); err != nil {
			return err
		}

		if err := sBucket.Put(keyForState(game, version), serializedStateRecord); err != nil {
			return err
		}

		return nil

	})

}

func (s *StorageManager) ListGames(manager *boardgame.GameManager, max int) []*boardgame.Game {

	var result []*boardgame.Game

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

			var record gameRecord

			if err := json.Unmarshal(v, &record); err != nil {
				return errors.New("Couldn't deserialize a game: " + err.Error())
			}

			game := manager.LoadGame(record.Name, record.Id, record.Version, record.Finished, record.Winners)

			if game != nil {
				result = append(result, game)
			}
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
