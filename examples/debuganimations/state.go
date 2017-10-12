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
	boardgame.BaseSubState
	DiscardStack     *boardgame.GrowableStack `stack:"cards"`
	DrawStack        *boardgame.GrowableStack `stack:"cards" sanitize:"order"`
	FirstShortStack  *boardgame.GrowableStack `stack:"cards" sanitize:"order"`
	SecondShortStack *boardgame.GrowableStack `stack:"cards" sanitize:"order"`
	HiddenCard       *boardgame.SizedStack    `stack:"cards,1" sanitize:"hidden"`
	RevealedCard     *boardgame.SizedStack    `stack:"cards,1"`
	FanStack         *boardgame.GrowableStack `stack:"cards"`
	FanDiscard       *boardgame.GrowableStack `stack:"cards" sanitize:"order"`
	VisibleStack     *boardgame.GrowableStack `stack:"cards"`
	HiddenStack      *boardgame.GrowableStack `stack:"cards" sanitize:"nonempty"`
	CurrentPlayer    boardgame.PlayerIndex
}

//+autoreader
type playerState struct {
	boardgame.BaseSubState
	playerIndex boardgame.PlayerIndex
	Hand        *boardgame.GrowableStack `stack:"cards,1"`
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}
