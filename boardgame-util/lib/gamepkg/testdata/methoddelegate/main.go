package methoddelegate

import "github.com/jkomoros/boardgame"

type factory struct{}

func (factory) NewDelegate() boardgame.GameDelegate { return nil }
