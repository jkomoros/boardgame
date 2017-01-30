package tictactoe

import (
	"github.com/jkomoros/boardgame"
)

const (
	X rune = 'X'
	O rune = 'O'
)

type playerToken struct {
	Value rune
}

func (p *playerToken) Props() []string {
	return boardgame.PropertyReaderPropsImpl(p)
}

func (p *playerToken) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(p, name)
}
