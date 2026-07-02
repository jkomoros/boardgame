/*
Package extendedgame is the definition of a StorageRecord for ExtendedGame. In a
separate package to avoid dependency cycles.
*/
package extendedgame

import (
	"encoding/json"

	"github.com/jkomoros/boardgame"
)

// StorageRecord is the extra information the server wants stored along with the
// game.
type StorageRecord struct {
	Open    bool
	Visible bool
	Owner   string
	// CompanionRoomCode is the 4-letter (or fallback 5-letter) join code for
	// games created in Table+Hand companion mode. Empty for solo-mode games.
	// See docs/superpowers/specs/2026-05-23-per-person-mobile-ui-design.md §6.1.
	CompanionRoomCode string
	// CompanionLocked, when true, prevents new phones from joining the room
	// even if the code is known. Host-controlled. Always false for solo-mode.
	CompanionLocked bool
	// CompanionHostOverride is the userID of a player who has claimed host
	// privileges after the original Owner went stale (spec §9.4). When set,
	// IsHost recognizes EITHER eGame.Owner OR this override as host.
	// Empty for new games — the original Owner is host by default. Cleared
	// when the game reaches Finished.
	CompanionHostOverride string
}

// CombinedStorageRecord combines the base GameStorageRecord and StorageRecord
// into one struct.
type CombinedStorageRecord struct {
	boardgame.GameStorageRecord
	StorageRecord
}

// DefaultStorageRecord returns a StorageRecord with all defaults set to default
// values.
func DefaultStorageRecord() *StorageRecord {
	return &StorageRecord{
		Open:    true,
		Visible: true,
		Owner:   "",
		// CompanionRoomCode and CompanionLocked default to "" / false — solo-mode.
	}
}

func (c *CombinedStorageRecord) String() string {
	blob, _ := json.Marshal(c)
	return string(blob) + "\n"
}
