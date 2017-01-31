package tictactoe

import (
	"github.com/jkomoros/boardgame"
)

const (
	X = "X"
	O = "O"
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
