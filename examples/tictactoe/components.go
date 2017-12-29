package tictactoe

import (
	"github.com/jkomoros/boardgame"
)

//+autoreader
const (
	PhaseBeforeFirstMove = iota
	PhaseAfterFirstMove
)

const (
	X     = "X"
	O     = "O"
	Empty = ""
)

//+autoreader reader
type playerToken struct {
	boardgame.BaseComponentValues
	Value string
}
