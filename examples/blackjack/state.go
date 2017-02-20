package blackjack

import (
	"github.com/jkomoros/boardgame"
)

type mainState struct {
	Game    *gameState
	Players []*playerState
}

type gameState struct {
	DiscardStack *boardgame.GrowableStack
	DrawStack    *boardgame.GrowableStack
}

type playerState struct {
	playerIndex int
	Hand        *boardgame.GrowableStack
}

func (g *gameState) Props() []string {
	return boardgame.PropertyReaderPropsImpl(g)
}

func (g *gameState) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(g, name)
}

func (g *gameState) Copy() boardgame.GameState {
	var result gameState
	result = *g
	return &result
}

func (p *playerState) Props() []string {
	return boardgame.PropertyReaderPropsImpl(p)
}

func (p *playerState) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(p, name)
}

func (p *playerState) Copy() boardgame.PlayerState {
	var result playerState
	result = *p
	return &result
}

func (p *playerState) PlayerIndex() int {
	return p.playerIndex
}

func (m *mainState) Diagram() string {
	return "TODO: implement this"
}

func (s *mainState) GameState() boardgame.GameState {
	return s.Game
}

func (s *mainState) PlayerStates() []boardgame.PlayerState {
	array := make([]boardgame.PlayerState, len(s.Players))

	for i := 0; i < len(s.Players); i++ {
		array[i] = s.Players[i]
	}

	return array
}

func (s *mainState) Copy() boardgame.State {
	array := make([]*playerState, len(s.Players))

	for i := 0; i < len(s.Players); i++ {
		array[i] = s.Players[i].Copy().(*playerState)
	}

	return &mainState{
		Game:    s.Game.Copy().(*gameState),
		Players: array,
	}
}
