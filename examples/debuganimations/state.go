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
	DiscardStack     boardgame.Stack `stack:"cards"`
	DrawStack        boardgame.Stack `stack:"cards" sanitize:"order"`
	FirstShortStack  boardgame.Stack `stack:"cards" sanitize:"order"`
	SecondShortStack boardgame.Stack `stack:"cards" sanitize:"order"`
	HiddenCard       boardgame.Stack `sizedstack:"cards,1" sanitize:"hidden"`
	RevealedCard     boardgame.Stack `sizedstack:"cards,1"`
	FanStack         boardgame.Stack `stack:"cards"`
	FanDiscard       boardgame.Stack `stack:"cards" sanitize:"order"`
	VisibleStack     boardgame.Stack `stack:"cards"`
	HiddenStack      boardgame.Stack `stack:"cards" sanitize:"nonempty"`
	CurrentPlayer    boardgame.PlayerIndex
}

//+autoreader
type playerState struct {
	boardgame.BaseSubState
	playerIndex boardgame.PlayerIndex
	Hand        boardgame.Stack `stack:"cards,1"`
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}
