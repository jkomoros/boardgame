package buildtag

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/examples/pig"
)

func NewDelegate() boardgame.GameDelegate { return pig.NewDelegate() }
