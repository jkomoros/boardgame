package checkers

import (
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
)

//boardgame:codegen
type playerState struct {
	base.SubState
	behaviors.Seat
	behaviors.InactivePlayer
}
