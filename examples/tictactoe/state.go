package tictactoe

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
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
	boardgame.BaseSubState
	CurrentPlayer boardgame.PlayerIndex
	Slots         boardgame.MutableSizedStack `sizedstack:"tokens,TOTAL_DIM"`
	//We don't actually need this; we mainly do it because the storage manager
	//tests use tictactoe as an example and need to test a phase transition.
	Phase enum.MutableVal `enum:"Phase"`
}

func (g *gameState) tokenValue(row, col int) string {
	return g.tokenValueAtIndex(rowColToIndex(row, col))
}

func (g *gameState) tokenValueAtIndex(index int) string {
	c := g.Slots.ComponentAt(index)
	if c == nil {
		return ""
	}
	return c.Values().(*playerToken).Value
}

func rowColToIndex(row, col int) int {
	return row*DIM + col
}

func (g *gameState) SetCurrentPlayer(currentPlayer boardgame.PlayerIndex) {
	g.CurrentPlayer = currentPlayer
}

//+autoreader
type playerState struct {
	boardgame.BaseSubState
	playerIndex  boardgame.PlayerIndex
	TokenValue   string
	UnusedTokens boardgame.MutableStack `stack:"tokens"`
	//How many tokens they have left to place this turn.
	TokensToPlaceThisTurn int
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}

func (p *playerState) ResetForTurnStart() error {
	p.TokensToPlaceThisTurn = 1
	return nil
}

func (p *playerState) ResetForTurnEnd() error {
	return nil
}

func (p *playerState) TurnDone() error {
	if p.TokensToPlaceThisTurn > 0 {
		return errors.New("they still have tokens left to place this turn")
	}
	return nil
}
