package debuganimations

import (
	"github.com/jkomoros/boardgame"
)

func concreteStates(state boardgame.State) (*gameState, []*playerState) {
	game := state.Game().(*gameState)

	players := make([]*playerState, len(state.Players()))

	for i, player := range state.Players() {
		players[i] = player.(*playerState)
	}

	return game, players
}

//+autoreader
type gameState struct {
	DiscardStack     *boardgame.GrowableStack
	DrawStack        *boardgame.GrowableStack
	FirstShortStack  *boardgame.GrowableStack
	SecondShortStack *boardgame.GrowableStack
	HiddenCard       *boardgame.SizedStack
	RevealedCard     *boardgame.SizedStack
	FanStack         *boardgame.GrowableStack
	FanDiscard       *boardgame.GrowableStack
	VisibleStack     *boardgame.GrowableStack
	HiddenStack      *boardgame.GrowableStack
	CurrentPlayer    boardgame.PlayerIndex
}

//+autoreader
type playerState struct {
	playerIndex boardgame.PlayerIndex
	Hand        *boardgame.GrowableStack
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}
