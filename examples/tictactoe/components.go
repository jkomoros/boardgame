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
	//X represents the x player
	X = "X"
	//O represents the o player
	O = "O"
	//Empty represents an unfilled cell
	Empty = ""
)

//boardgame:codegen reader
type playerToken struct {
	base.ComponentValues
	Value string
}
