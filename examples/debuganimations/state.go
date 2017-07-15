package debuganimations

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
	DiscardStack     *boardgame.GrowableStack `stack:"cards"`
	DrawStack        *boardgame.GrowableStack `stack:"cards"`
	FirstShortStack  *boardgame.GrowableStack `stack:"cards"`
	SecondShortStack *boardgame.GrowableStack `stack:"cards"`
	HiddenCard       *boardgame.SizedStack    `stack:"cards,1"`
	RevealedCard     *boardgame.SizedStack    `stack:"cards,1"`
	FanStack         *boardgame.GrowableStack `stack:"cards"`
	FanDiscard       *boardgame.GrowableStack `stack:"cards"`
	VisibleStack     *boardgame.GrowableStack `stack:"cards"`
	HiddenStack      *boardgame.GrowableStack `stack:"cards"`
	CurrentPlayer    boardgame.PlayerIndex
}

//+autoreader
type playerState struct {
	playerIndex boardgame.PlayerIndex
	Hand        *boardgame.GrowableStack `stack:"cards,1"`
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}
