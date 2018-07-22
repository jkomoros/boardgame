package filesystem

import (
	"github.com/jkomoros/boardgame"
)

type record struct {
	Game   *boardgame.GameStorageRecord
	States []boardgame.StateStorageRecord
	Moves  []*boardgame.MoveStorageRecord
}
