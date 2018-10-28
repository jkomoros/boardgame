package checkers

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
)

//boardgame:codegen
type playerState struct {
	base.SubState
	playerIndex boardgame.PlayerIndex
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}
