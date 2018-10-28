/*

saferanduse is a simple fake game pacakge that imports math.rand but with an override to assert it's safe

*/
package saferanduse

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	//boardgame:assert(rand_use_deterministic)
	"math/rand"
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

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurablePlayerState {
	return nil
}

func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
