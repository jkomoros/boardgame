package checkers

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/enum/graph"
)

//boardgame:codegen
const (
	PhaseSetup = iota
	PhasePlaying
)

//boardgame:codegen
const (
	ColorBlack = iota
	ColorRed
)

//boardgame:codegen reader
type token struct {
	boardgame.BaseComponentValues
	Color enum.Val
}

//boardgame:codegen
type tokenDynamic struct {
	boardgame.BaseSubState
	Crowned bool
}

const numTokens = 12

const boardWidth = 8
const boardSize = boardWidth * boardWidth

var SpacesEnum = Enums.MustAddRange("Spaces", boardWidth, boardWidth)

var GraphDownward = graph.MustNewGridConnectedness(SpacesEnum, graph.DirectionDown, graph.DirectionDiagonal)
var GraphUpward = graph.MustNewGridConnectedness(SpacesEnum, graph.DirectionUp, graph.DirectionDiagonal)

const tokenDeckName = "Tokens"

//The first space in the upper left is black, and it alternates from there.
//The red tokens start at the top, and the black tokens are arrayed from the
//bottom.
func spaceIsBlack(spaceIndex int) bool {
	return spaceIndex%2 == 0
}

func newTokenDeck() *boardgame.Deck {

	deck := boardgame.NewDeck()

	for i := 0; i < numTokens; i++ {
		deck.AddComponent(&token{
			Color: ColorEnum.MustNewVal(ColorBlack),
		})
	}

	for i := 0; i < numTokens; i++ {
		deck.AddComponent(&token{
			Color: ColorEnum.MustNewVal(ColorRed),
		})
	}

	return deck
}

func (t *token) Dynamic(state boardgame.ImmutableState) *tokenDynamic {
	return t.ContainingComponent().ImmutableInstance(state).ImmutableDynamicValues().(*tokenDynamic)
}

func (t *token) Legal(state boardgame.ImmutableState, legalType int) error {
	//Red starts at top, moves towards bottom
	targetRow := boardWidth - 1

	if t.Color.Value() == ColorBlack {
		//Black starts at top, moves towards bottom
		targetRow = 0
	}

	_, slotIndex, err := t.ContainingComponent().ImmutableInstance(state).ContainingImmutableStack()

	if err != nil {
		return errors.New("Component's position could not be found: " + err.Error())
	}

	indexes := SpacesEnum.ValueToRange(slotIndex)

	if indexes[0] != targetRow {
		//Not in the target row
		return errors.New("Not in the target row")
	}

	d := t.Dynamic(state)

	if d.Crowned {
		//Already crowned
		return errors.New("Already crowned")
	}

	return nil
}

//FreeNextSpaces is like AllNextSpaces, but spaces taht are occupied won't be returned.
func (t *token) FreeNextSpaces(state boardgame.ImmutableState, componentIndex int) []int {

	spaces := state.ImmutableGameState().(*gameState).Spaces

	var result []int
	for _, space := range t.FreeNextSpaces(state, componentIndex) {
		if spaces.ComponentAt(space) == nil {
			result = append(result, space)
		}
	}

	return result
}

//AllNextSpaces returns all the spaces that t could move to, if the rest of
//the board were empty.
func (t *token) AllNextSpaces(state boardgame.ImmutableState, componentIndex int) []int {

	//Red starts from top
	fromBottom := false

	if t.Color.Value() == ColorBlack {
		fromBottom = true
	}

	var nextSpaces []int

	dyn := t.Dynamic(state)

	crowned := dyn.Crowned

	g := GraphUpward
	oppositeG := GraphDownward

	if fromBottom {
		g = GraphDownward
		oppositeG = GraphUpward
	}

	for _, val := range g.Neighbors(componentIndex) {
		nextSpaces = append(nextSpaces, val)
	}

	if crowned {
		for _, val := range oppositeG.Neighbors(componentIndex) {
			nextSpaces = append(nextSpaces, val)
		}
	}

	return nextSpaces
}

//LegalCaptureSpaces returns cells that are legal for this cell to capture from there.
func (t *token) LegalCaptureSpaces(state boardgame.ImmutableState, componentIndex int) []int {

	spaces := state.ImmutableGameState().(*gameState).Spaces

	nextSpaces := t.AllNextSpaces(state, componentIndex)

	var result []int

	for _, space := range nextSpaces {
		c := spaces.ComponentAt(space)
		if c == nil {
			continue
		}
		if c.Values() == nil {
			continue
		}
		v := c.Values().(*token)
		if v.Color.Equals(t.Color) {
			//One of our own.
			continue
		}
		//The item at space is a legal capture. What's the spot one beyond it,
		//and is it taken?

		startIndexes := SpacesEnum.ValueToRange(componentIndex)
		endIndexes := SpacesEnum.ValueToRange(space)

		diff := []int{
			endIndexes[0] - startIndexes[0],
			endIndexes[1] - startIndexes[1],
		}

		finalIndexes := []int{
			endIndexes[0] + diff[0],
			endIndexes[1] + diff[1],
		}

		finalSpace := SpacesEnum.RangeToValue(finalIndexes...)

		if finalSpace == enum.IllegalValue {
			//A space beyond the bounds
			continue
		}

		if spaces.ComponentAt(finalSpace) == nil {
			//An empty, real space!
			result = append(result, finalSpace)
		}

	}

	return result
}
