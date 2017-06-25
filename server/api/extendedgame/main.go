/*

	extendedgame is the definition of a StorageRecord for ExtendedGame. In a
	separate package to avoid dependency cycles.

*/
package extendedgame

import (
	"github.com/jkomoros/boardgame"
	"time"
)

type StorageRecord struct {
	Created      int64
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
		Created:      time.Now().UnixNano(),
		LastActivity: time.Now().UnixNano(),
		Open:         true,
		Visible:      true,
		Owner:        "",
	}
}
