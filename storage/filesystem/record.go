package filesystem

import (
	"encoding/json"
	"github.com/jkomoros/boardgame"
)

type record struct {
	Game   *boardgame.GameStorageRecord
	States []json.RawMessage
	Moves  []*boardgame.MoveStorageRecord
}
