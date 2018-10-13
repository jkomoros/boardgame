package checkers

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves"
)

//boardgame:codegen
type MovePlaceToken struct {
	moves.FixUpMulti
	TargetIndex enum.RangeVal `enum:"Spaces"`
}

func (m *MovePlaceToken) DefaultsForState(state boardgame.ImmutableState) {

	game := state.ImmutableGameState().(*gameState)

	if game.UnusedTokens.NumComponents() <= 0 {
		return
	}

	nextToken := game.UnusedTokens.ComponentAt(0)

	nextTokenVals := nextToken.Values().(*token)

	//Red starts at top
	fromBottom := false

	if nextTokenVals.Color.Value() == ColorBlack {
		fromBottom = true
	}

	startIndex := 0
	increment := 1
	endCondition := game.Spaces.Len()

	if fromBottom {
		startIndex = game.Spaces.Len() - 1
		increment = -1
		endCondition = 0
	}

	for i := startIndex; i != endCondition; i += increment {
		//We're only allowed to put tokens on black spaces
		if !spaceIsBlack(i) {
			continue
		}
		if game.Spaces.ComponentAt(i) == nil {
			m.TargetIndex.SetValue(i)
			return
		}
	}

}

func (m *MovePlaceToken) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUpMulti.Legal(state, proposer); err != nil {
		return err
	}

	game := state.ImmutableGameState().(*gameState)

	if game.UnusedTokens.NumComponents() == 0 {
		return errors.New("No more components to place")
	}

	if game.Spaces.ComponentAt(m.TargetIndex.Value()) != nil {
		return errors.New("That space is already filled")
	}

	if !spaceIsBlack(m.TargetIndex.Value()) {
		return errors.New("The proposed space is not black")
	}

	return nil
}

func (m *MovePlaceToken) Apply(state boardgame.State) error {
	game := state.GameState().(*gameState)
	return game.UnusedTokens.First().MoveTo(game.Spaces, m.TargetIndex.Value())
}

//boardgame:codegen
type MoveMoveToken struct {
	moves.CurrentPlayer
	TokenIndexToMove enum.RangeVal `enum:"Spaces"`
	SpaceIndex       enum.RangeVal `enum:"Spaces"`
}

func (m *MoveMoveToken) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	p := state.ImmutableCurrentPlayer().(*playerState)

	g := state.ImmutableGameState().(*gameState)

	c := g.Spaces.ComponentAt(m.TokenIndexToMove.Value())

	if c == nil {
		return errors.New("That space does not have a component in it")
	}

	t := c.Values().(*token)

	if !p.Color.Equals(t.Color) {
		return errors.New("That token isn't your token to move!")
	}

	if !spaceIsBlack(m.SpaceIndex.Value()) {
		return errors.New("You can only move to spaces that are black.")
	}

	if g.Spaces.ComponentAt(m.SpaceIndex.Value()) != nil {
		return errors.New("The space you're trying to move to is occupied.")
	}

	//If it's one of the legal spaces, great.
	for _, space := range t.FreeNextSpaces(state, m.TokenIndexToMove.Value()) {
		if m.SpaceIndex.Value() == space {
			return nil
		}
	}

	for _, space := range t.LegalCaptureSpaces(state, m.TokenIndexToMove.Value()) {
		if m.SpaceIndex.Value() == space {
			return nil
		}
	}

	return errors.New("SpaceIndex does not represent a legal space for that token to move to.")

}

func (m *MoveMoveToken) Apply(state boardgame.State) error {

	g := state.GameState().(*gameState)

	p := state.CurrentPlayer().(*playerState)

	if err := g.Spaces.SwapComponents(m.TokenIndexToMove.Value(), m.SpaceIndex.Value()); err != nil {
		return errors.New("Couldn't move token: " + err.Error())
	}

	startIndexes := m.TokenIndexToMove.RangeValue()

	if startIndexes == nil || len(startIndexes) != 2 {
		return errors.New("Couldn't get indexes for token space")
	}

	finishIndexes := m.SpaceIndex.RangeValue()

	if finishIndexes == nil || len(finishIndexes) != 2 {
		return errors.New("Couldn't get indexes for finish space")
	}

	middleIndexes := []int{
		finishIndexes[0] - startIndexes[0],
		finishIndexes[1] - startIndexes[1],
	}

	middleSpace := SpacesEnum.RangeToValue(middleIndexes...)

	if middleSpace < 0 {
		return errors.New("Invalid resule from range to value")
	}

	c := g.Spaces.ComponentAt(middleSpace)

	tokenCaptured := false

	if c != nil {

		tokenValues := c.Values().(*token)

		if !tokenValues.Color.Equals(p.Color) {
			tokenCaptured = true
			if err := g.Spaces.ComponentAt(middleSpace).MoveToLastSlot(p.CapturedTokens); err != nil {
				return errors.New("Couldn't capture token: " + err.Error())
			}
		}

	}

	//The turn is over if a token wasn't captured
	if !tokenCaptured {
		p.FinishedTurn = true
	} else {
		//The turn is also over if there isn't another cpature space to move
		//to.
		t := g.Spaces.ComponentAt(m.SpaceIndex.Value()).Values().(*token)
		if len(t.LegalCaptureSpaces(state, m.SpaceIndex.Value())) == 0 {
			p.FinishedTurn = true
		}
	}

	return nil

}

//boardgame:codegen
type MoveCrownToken struct {
	moves.DefaultComponent
}

func (m *MoveCrownToken) Apply(state boardgame.State) error {
	g := state.GameState().(*gameState)

	c := g.Spaces.ComponentAt(m.ComponentIndex)

	if c == nil {
		return errors.New("No token at that space")
	}

	d := c.DynamicValues().(*tokenDynamic)

	d.Crowned = true

	return nil
}
