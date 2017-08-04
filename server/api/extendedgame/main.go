/*

	extendedgame is the definition of a StorageRecord for ExtendedGame. In a
	separate package to avoid dependency cycles.

*/
package extendedgame

import (
	"encoding/json"
	"github.com/jkomoros/boardgame"
	"time"
)

type StorageRecord struct {
	LastActivity int64
	Open         bool
	Visible      bool
	Owner        string
}

type CombinedStorageRecord struct {
	boardgame.GameStorageRecord
	StorageRecord
}

func DefaultStorageRecord() *StorageRecord {
	return &StorageRecord{
		LastActivity: time.Now().UnixNano(),
		Open:         true,
		Visible:      true,
		Owner:        "",
	}
}

func (c *CombinedStorageRecord) String() string {
	blob, _ := json.Marshal(c)
	return string(blob) + "\n"
}
