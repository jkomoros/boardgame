package tictactoe

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/enum"
)

func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
	game := state.ImmutableGameState().(*gameState)

	players := make([]*playerState, len(state.ImmutablePlayerStates()))

	for i, player := range state.ImmutablePlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}

//boardgame:codegen
type gameState struct {
	base.SubState
	CurrentPlayer boardgame.PlayerIndex
	Slots         boardgame.SizedStack `sizedstack:"tokens,TOTAL_DIM"`
	//We don't actually need this; we mainly do it because the storage manager
	//tests use tictactoe as an example and need to test a phase transition.
	Phase enum.Val `enum:"phase"`
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
	return row*dim + col
}

func (g *gameState) SetCurrentPlayer(currentPlayer boardgame.PlayerIndex) {
	g.CurrentPlayer = currentPlayer
}

//boardgame:codegen
type playerState struct {
	base.SubState
	TokenValue   string
	UnusedTokens boardgame.Stack `stack:"tokens"`
	//How many tokens they have left to place this turn.
	TokensToPlaceThisTurn int
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
