package helper

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
)

type Delegate struct{ base.GameDelegate }

func New() *Delegate { return new(Delegate) }

func (*Delegate) Name() string { return "transitivereturn" }

func (*Delegate) GameStateConstructor() boardgame.ConfigurableSubState { return nil }

func (*Delegate) PlayerStateConstructor(boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return nil
}

func (*Delegate) ConfigureMoves() []boardgame.MoveConfig { return nil }
