package checkers

import (
	"errors"
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

func (t *token) Legal(state boardgame.State, legalType int, componentIndex int) error {
	//Red starts at top, moves towards bottom
	targetRow := boardWidth - 1

	if t.Color.Value() == ColorBlack {
		//Black starts at top, moves towards bottom
		targetRow = 0
	}

	indexes := SpacesEnum.ValueToRange(componentIndex)

	if indexes[0] != targetRow {
		//Not in the target row
		return errors.New("Not in the target row")
	}

	d := t.ContainingComponent().DynamicValues(state).(*tokenDynamic)

	if d.Crowned {
		//Already crowned
		return errors.New("Already crowned")
	}

	return nil
}
