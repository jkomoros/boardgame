package compileerror

import "github.com/jkomoros/boardgame"

var _ = deliberatelyMissing

func NewDelegate() boardgame.GameDelegate { return nil }
