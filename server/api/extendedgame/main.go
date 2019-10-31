/*

Package extendedgame is the definition of a StorageRecord for ExtendedGame. In a
separate package to avoid dependency cycles.

*/
package extendedgame

import (
	"encoding/json"

	"github.com/jkomoros/boardgame"
)

//StorageRecord is the extra information the server wants stored along with the
//game.
type StorageRecord struct {
	Open    bool
	Visible bool
	Owner   string
}

//CombinedStorageRecord combines the base GameStorageRecord and StorageRecord
//into one struct.
type CombinedStorageRecord struct {
	boardgame.GameStorageRecord
	StorageRecord
}

//DefaultStorageRecord returns a StorageRecord with all defaults set to default
//values.
func DefaultStorageRecord() *StorageRecord {
	return &StorageRecord{
		Open:    true,
		Visible: true,
		Owner:   "",
	}
}

func (c *CombinedStorageRecord) String() string {
	blob, _ := json.Marshal(c)
	return string(blob) + "\n"
}
