package tictactoe

import (
	"github.com/jkomoros/boardgame/base"
)

//boardgame:codegen
const (
	PhaseBeforeFirstMove = iota
	PhaseAfterFirstMove
)

const (
	X     = "X"
	O     = "O"
	Empty = ""
)

//boardgame:codegen reader
type playerToken struct {
	base.ComponentValues
	Value string
}
