package tictactoe

import (
	"github.com/jkomoros/boardgame"
)

func concreteStates(state *boardgame.State) (*gameState, []*playerState) {
	game := state.Game().(*gameState)

	players := make([]*playerState, len(state.Players()))

	for i, player := range state.Players() {
		players[i] = player.(*playerState)
	}

	return game, players
}

type gameState struct {
	CurrentPlayer int
	Slots         *boardgame.SizedStack
}

func (g *gameState) tokenValue(row, col int) string {
	return g.tokenValueAtIndex(rowColToIndex(row, col))
}

func (g *gameState) tokenValueAtIndex(index int) string {
	c := g.Slots.ComponentAt(index)
	if c == nil {
		return ""
	}
	return c.Values.(*playerToken).Value
}

func rowColToIndex(row, col int) int {
	return row*DIM + col
}

func (g *gameState) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(g)
}

func (g *gameState) Reader() boardgame.PropertyReader {
	return boardgame.DefaultReader(g)
}

func (g *gameState) MutableCopy() boardgame.MutableGameState {
	var result gameState
	result = *g
	result.Slots = g.Slots.Copy()
	return &result
}

func (g *gameState) Copy() boardgame.GameState {
	return g.MutableCopy()
}

type playerState struct {
	playerIndex  int
	TokenValue   string
	UnusedTokens *boardgame.GrowableStack
	//How many tokens they have left to place this turn.
	TokensToPlaceThisTurn int
}

func (p *playerState) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(p)
}

func (p *playerState) Reader() boardgame.PropertyReader {
	return boardgame.DefaultReader(p)
}

func (p *playerState) MutableCopy() boardgame.MutablePlayerState {
	var result playerState
	result = *p
	result.UnusedTokens = p.UnusedTokens.Copy()
	return &result
}

func (p *playerState) Copy() boardgame.PlayerState {
	return p.MutableCopy()
}

func (p *playerState) PlayerIndex() int {
	return p.playerIndex
}
