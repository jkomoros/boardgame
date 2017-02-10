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

func (p *playerToken) Props() []string {
	return boardgame.PropertyReaderPropsImpl(p)
}

func (p *playerToken) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(p, name)
}

//Designed to be used with stack.ComponentValues()
func playerTokenValues(in []boardgame.PropertyReader) []*playerToken {
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
