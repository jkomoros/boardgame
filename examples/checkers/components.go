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

//note: the struct tag for Spaces in gameState implicitly depends on this
//value.
const boardWidth = 8

var SpacesEnum = Enums.MustAddRange("Spaces", boardWidth, boardWidth)

const tokenDeckName = "tokens"

//The first space in the upper left is black, and it alternates from there.
//The black tokens start at the top, and the red tokens are arrayed from the
//bottom.
func spaceIsBlack(spaceIndex int) bool {
	return spaceIndex%2 == 0
}

func newTokenDeck() *boardgame.Deck {

	deck := boardgame.NewDeck()

	deck.AddComponentMulti(&token{
		Color: ColorEnum.MustNewVal(ColorBlack),
	}, numTokens)

	deck.AddComponentMulti(&token{
		Color: ColorEnum.MustNewVal(ColorRed),
	}, numTokens)

	return deck
}
