package checkers

import (
	"github.com/jkomoros/boardgame"
)

//boardgame:codegen
type playerState struct {
	boardgame.BaseSubState
	playerIndex boardgame.PlayerIndex
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}
