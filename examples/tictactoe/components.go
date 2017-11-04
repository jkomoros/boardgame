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
	Value string
}

//Designed to be used with stack.ComponentValues()
func playerTokenValues(in []boardgame.Reader) []*playerToken {
	result := make([]*playerToken, len(in))
	for i := 0; i < len(in); i++ {
		c := in[i]
		if c == nil {
			result[i] = nil
			continue
		}
		result[i] = c.(*playerToken)
	}
	return result
}
