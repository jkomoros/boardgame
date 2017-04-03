package memory

import (
	"github.com/jkomoros/boardgame"
)

var cardNames []string = []string{
	"Apple",
	"Orange",
	"Pear",
	"Ball",
	"Hammer",
	"Star",
	"Dog",
	"Cat",
}

const cardsDeckName = "cards"

type cardValue struct {
	Type string
}

func (c *cardValue) Reader() boardgame.PropertyReader {
	return boardgame.DefaultReader(c)
}
