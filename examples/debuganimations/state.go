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
	DiscardStack        boardgame.MutableStack `stack:"cards"`
	DrawStack           boardgame.MutableStack `stack:"cards" sanitize:"order"`
	FirstShortStack     boardgame.MutableStack `stack:"cards" sanitize:"order"`
	SecondShortStack    boardgame.MutableStack `stack:"cards" sanitize:"order"`
	HiddenCard          boardgame.MutableStack `sizedstack:"cards,1" sanitize:"order"`
	RevealedCard        boardgame.MutableStack `sizedstack:"cards,1"`
	Card                boardgame.Stack        `overlap:"RevealedCard,HiddenCard"`
	FanStack            boardgame.MutableStack `stack:"cards"`
	FanDiscard          boardgame.MutableStack `stack:"cards" sanitize:"order"`
	VisibleStack        boardgame.MutableStack `stack:"cards"`
	HiddenStack         boardgame.MutableStack `stack:"cards" sanitize:"nonempty"`
	TokensFrom          boardgame.MutableStack `stack:"tokens"`
	TokensTo            boardgame.MutableStack `stack:"tokens"`
	SanitizedTokensFrom boardgame.MutableStack `stack:"tokens"`
	SanitizedTokensTo   boardgame.MutableStack `stack:"tokens" sanitize:"nonempty"`
	CurrentPlayer       boardgame.PlayerIndex
}

//+autoreader
type playerState struct {
	boardgame.BaseSubState
	playerIndex boardgame.PlayerIndex
	Hand        boardgame.MutableStack `stack:"cards,1"`
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}
