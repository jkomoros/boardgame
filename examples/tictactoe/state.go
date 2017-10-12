package tictactoe

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

func init() {

	//Make sure that we get compile-time errors if our player and game state
	//don't adhere to the interfaces that moves.FinishTurn expects
	moves.VerifyFinishTurnStates(&gameState{}, &playerState{})
}

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

func (g *gameState) SetCurrentPlayer(currentPlayer boardgame.PlayerIndex) {
	g.CurrentPlayer = currentPlayer
}

//+autoreader
type playerState struct {
	boardgame.BaseSubState
	playerIndex  boardgame.PlayerIndex
	TokenValue   string
	UnusedTokens *boardgame.GrowableStack `stack:"tokens"`
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
