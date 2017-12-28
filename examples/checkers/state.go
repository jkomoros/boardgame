package checkers

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

//+autoreader
type gameState struct {
	boardgame.BaseSubState
	Phase enum.MutableVal `enum:"Phase"`
	//Note: the struct tag here implicitly depends on the value of boardWidth.
	Spaces       boardgame.MutableStack `sizedstack:"Tokens,64"`
	UnusedTokens boardgame.MutableStack `stack:"Tokens"`
}

//+autoreader
type playerState struct {
	boardgame.BaseSubState
	playerIndex boardgame.PlayerIndex
	Color       enum.MutableVal `enum:"Color"`
	//The tokens of the OTHER player that we've captured.
	CapturedTokens boardgame.MutableStack `stack:"Tokens"`
}

func concreteStates(state boardgame.State) (*gameState, []*playerState) {
	game := state.GameState().(*gameState)

	players := make([]*playerState, len(state.PlayerStates()))

	for i, player := range state.PlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}
