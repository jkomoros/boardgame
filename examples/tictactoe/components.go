package tictactoe

import (
	"github.com/jkomoros/boardgame"
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
	boardgame.BaseComponentValues
	Value string
}
