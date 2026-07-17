/*
unsaferanduse is a simple fake game pacakge that imports math.rand unsafely
*/
package unsaferanduse

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	_ "math/rand"
)

type gameDelegate struct {
	base.GameDelegate
}

func (g *gameDelegate) Name() string {
	return "unsaferanduse"
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return nil
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return nil
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig { return nil }

func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
