package checkers

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

//+autoreader
type gameState struct {
	boardgame.BaseSubState
	Phase         enum.MutableVal `enum:"Phase"`
	CurrentPlayer boardgame.PlayerIndex
	//Note: the struct tag here implicitly depends on the value of boardWidth.
	Spaces       boardgame.MutableSizedStack `sizedstack:"Tokens,64"`
	UnusedTokens boardgame.MutableStack      `stack:"Tokens"`
}

//+autoreader
type playerState struct {
	boardgame.BaseSubState
	playerIndex boardgame.PlayerIndex
	Color       enum.MutableVal `enum:"Color"`
	//The tokens of the OTHER player that we've captured.
	CapturedTokens boardgame.MutableStack `stack:"Tokens"`
	FinishedTurn   bool
}

func concreteStates(state boardgame.State) (*gameState, []*playerState) {
	game := state.GameState().(*gameState)

	players := make([]*playerState, len(state.PlayerStates()))

	for i, player := range state.PlayerStates() {
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
