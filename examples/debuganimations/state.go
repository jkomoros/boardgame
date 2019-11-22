package debuganimations

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
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
	behaviors.CurrentPlayerBehavior
	behaviors.PhaseBehavior
	DiscardStack        boardgame.Stack       `stack:"cards"`
	DrawStack           boardgame.Stack       `stack:"cards" sanitize:"order"`
	FirstShortStack     boardgame.Stack       `stack:"cards" sanitize:"order"`
	SecondShortStack    boardgame.Stack       `stack:"cards" sanitize:"order"`
	HiddenCard          boardgame.SizedStack  `sizedstack:"cards,1" sanitize:"order"`
	VisibleCard         boardgame.SizedStack  `sizedstack:"cards,1"`
	Card                boardgame.MergedStack `overlap:"VisibleCard,HiddenCard"`
	FanStack            boardgame.Stack       `stack:"cards"`
	FanDiscard          boardgame.Stack       `stack:"cards" sanitize:"order"`
	FanShuffleCount     int
	VisibleStack        boardgame.Stack `stack:"cards"`
	HiddenStack         boardgame.Stack `stack:"cards" sanitize:"nonempty"`
	AllVisibleStack     boardgame.Stack `stack:"cards"`
	AllHiddenStack      boardgame.Stack `stack:"cards" sanitize:"order"`
	TokensFrom          boardgame.Stack `stack:"tokens"`
	TokensTo            boardgame.Stack `stack:"tokens"`
	SanitizedTokensFrom boardgame.Stack `stack:"tokens"`
	SanitizedTokensTo   boardgame.Stack `stack:"tokens" sanitize:"nonempty"`
}

//boardgame:codegen
type playerState struct {
	base.SubState
	Hand boardgame.Stack `stack:"cards,1"`
}
