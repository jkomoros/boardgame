package tictactoe

import (
	"github.com/jkomoros/boardgame/base"
)

//boardgame:codegen
const (
	phaseBeforeFirstMove = iota
	phaseAfterFirstMove
)

//TODO: this should use enums, and then PlayerStateConstructor should just use
//new(playerState).
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
