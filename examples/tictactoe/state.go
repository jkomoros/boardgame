package tictactoe

import (
	"github.com/jkomoros/boardgame"
)

func concreteStates(state boardgame.State) (*gameState, []*playerState) {
	game := state.GameState().(*gameState)

	players := make([]*playerState, len(state.PlayerStates()))

	for i, player := range state.PlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}

//+autoreader
type gameState struct {
	CurrentPlayer boardgame.PlayerIndex
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

//+autoreader
type playerState struct {
	playerIndex  boardgame.PlayerIndex
	TokenValue   string
	UnusedTokens *boardgame.GrowableStack
	//How many tokens they have left to place this turn.
	TokensToPlaceThisTurn int
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}
