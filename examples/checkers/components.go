package checkers

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

//+autoreader
const (
	PhaseSetup = iota
	PhasePlaying
)

//+autoreader
const (
	ColorBlack = iota
	ColorRed
)

//+autoreader reader
type token struct {
	Color enum.Val
}

//+autoreader
type tokenDynamic struct {
	boardgame.BaseSubState
	Crowned bool
}

const numTokens = 12

func newTokenDeck(color int) *boardgame.Deck {

	deck := boardgame.NewDeck()

	deck.AddComponentMulti(&token{
		Color: ColorEnum.MustNewVal(color),
	}, numTokens)

	return deck
}
