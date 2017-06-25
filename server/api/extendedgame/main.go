/*

	extendedgame is the definition of a StorageRecord for ExtendedGame. In a
	separate package to avoid dependency cycles.

*/
package extendedgame

import (
	"github.com/jkomoros/boardgame"
)

type StorageRecord struct {
	boardgame.GameStorageRecord
	Created      int64
	LastActivity int64
	Open         bool
	Visible      bool
	Owner        string
}
