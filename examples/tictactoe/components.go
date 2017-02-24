package tictactoe

import (
	"github.com/jkomoros/boardgame"
)

const (
	X     = "X"
	O     = "O"
	Empty = ""
)

type playerToken struct {
	Value string
}

func (p *playerToken) Reader() boardgame.PropertyReader {
	return boardgame.NewDefaultReader(p)
}

//Designed to be used with stack.ComponentValues()
func playerTokenValues(in []boardgame.ComponentValues) []*playerToken {
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
