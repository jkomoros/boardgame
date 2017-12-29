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
	boardgame.BaseComponentValues
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

const tokenDeckName = "Tokens"

//The first space in the upper left is black, and it alternates from there.
//The red tokens start at the top, and the black tokens are arrayed from the
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

func (t *token) ShouldBeCrowned(state boardgame.State, spaceIndex int, c *boardgame.Component) bool {
	//Red starts at top, moves towards bottom
	targetRow := boardWidth - 1

	if t.Color.Value() == ColorBlack {
		//Black starts at top, moves towards bottom
		targetRow = 0
	}

	indexes := SpacesEnum.ValueToRange(spaceIndex)

	if indexes[0] != targetRow {
		//Not in the target row
		return false
	}

	d := c.DynamicValues(state).(*tokenDynamic)

	if d.Crowned {
		//Already crowned
		return false
	}

	return true
}
