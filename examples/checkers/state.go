package checkers

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/enum"
)

//boardgame:codegen
type gameState struct {
	base.SubState
	Phase         enum.Val `enum:"Phase"`
	CurrentPlayer boardgame.PlayerIndex
	Spaces        boardgame.SizedStack `sizedstack:"Tokens,BOARD_SIZE"`
	UnusedTokens  boardgame.Stack      `stack:"Tokens"`
}

//boardgame:codegen
type playerState struct {
	base.SubState
	playerIndex boardgame.PlayerIndex
	Color       enum.Val `enum:"Color"`
	//The tokens of the OTHER player that we've captured.
	CapturedTokens boardgame.Stack `stack:"Tokens"`
	FinishedTurn   bool
}

func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
	game := state.ImmutableGameState().(*gameState)

	players := make([]*playerState, len(state.ImmutablePlayerStates()))

	for i, player := range state.ImmutablePlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}

func (g *gameState) SetCurrentPhase(phase int) {
	g.Phase.SetValue(phase)
}

func (g *gameState) SetCurrentPlayer(player boardgame.PlayerIndex) {
	g.CurrentPlayer = player
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}

func (p *playerState) TurnDone() error {
	if !p.FinishedTurn {
		return errors.New("The player has not yet finished their turn.")
	}
	return nil
}

func (p *playerState) ResetForTurnStart() error {
	p.FinishedTurn = false
	return nil
}

func (p *playerState) ResetForTurnEnd() error {
	//Pass
	return nil
}
