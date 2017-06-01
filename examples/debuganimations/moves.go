package debuganimations

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

//+autoreader readsetter
type moveMoveCardBetweenShortStacks struct {
	boardgame.DefaultMove
	FromFirst bool
}

//+autoreader readsetter
type moveMoveCardBetweenDrawAndDiscardStacks struct {
	boardgame.DefaultMove
	FromDraw bool
}

//+autoreader readsetter
type moveFlipHiddenCard struct {
	boardgame.DefaultMove
}

//+autoreader readsetter
type moveMoveCardBetweenFanStacks struct {
	boardgame.DefaultMove
}

/**************************************************
 *
 * moveMoveCardBetweenShortStacks Implementation
 *
 **************************************************/

func MoveMoveCardBetweenShortStacksFactory(state boardgame.State) boardgame.Move {
	result := &moveMoveCardBetweenShortStacks{
		boardgame.DefaultMove{
			"Move Card Between Short Stacks",
			"Moves a card between two short stacks",
		},
		true,
	}

	if state != nil {
		gameState, _ := concreteStates(state)

		if gameState.FirstShortStack.NumComponents() < 1 {
			result.FromFirst = false
		}
	}

	return result
}

func (m *moveMoveCardBetweenShortStacks) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	game, _ := concreteStates(state)

	if game.FirstShortStack.NumComponents() < 1 && m.FromFirst {
		return errors.New("First short stack has no cards to move")
	}

	if game.SecondShortStack.NumComponents() < 1 && !m.FromFirst {
		return errors.New("Second short stack has no cards to move")
	}

	return nil
}

func (m *moveMoveCardBetweenShortStacks) Apply(state boardgame.MutableState) error {

	game, _ := concreteStates(state)

	from := game.SecondShortStack
	to := game.FirstShortStack

	if m.FromFirst {
		from = game.FirstShortStack
		to = game.SecondShortStack
	}

	if err := from.MoveComponent(boardgame.FirstComponentIndex, to, boardgame.FirstSlotIndex); err != nil {
		return err
	}

	return nil
}

/**************************************************
 *
 * moveMoveCardBetweenDrawAndDiscardStacks Implementation
 *
 **************************************************/

func MoveMoveCardBetweenDrawAndDiscardStacksFactory(state boardgame.State) boardgame.Move {
	result := &moveMoveCardBetweenDrawAndDiscardStacks{
		boardgame.DefaultMove{
			"Move Card Between Draw and Discard Stacks",
			"Moves a card between draw and discard stacks",
		},
		true,
	}

	if state != nil {
		gameState, _ := concreteStates(state)

		if gameState.DiscardStack.NumComponents() < 3 {
			result.FromDraw = true
		} else {
			result.FromDraw = false
		}
	}

	return result
}

func (m *moveMoveCardBetweenDrawAndDiscardStacks) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	game, _ := concreteStates(state)

	if game.DrawStack.NumComponents() < 1 && m.FromDraw {
		return errors.New("Draw stack has no cards to move")
	}

	if game.DiscardStack.NumComponents() < 1 && !m.FromDraw {
		return errors.New("Discard stack has no cards to move")
	}

	return nil
}

func (m *moveMoveCardBetweenDrawAndDiscardStacks) Apply(state boardgame.MutableState) error {

	game, _ := concreteStates(state)

	from := game.DiscardStack
	to := game.DrawStack

	if m.FromDraw {
		from = game.DrawStack
		to = game.DiscardStack
	}

	if err := from.MoveComponent(boardgame.FirstComponentIndex, to, boardgame.FirstSlotIndex); err != nil {
		return err
	}

	return nil
}

/**************************************************
 *
 * moveFlipHiddenCard Implementation
 *
 **************************************************/

func MoveFlipHiddenCardFactory(state boardgame.State) boardgame.Move {
	result := &moveFlipHiddenCard{
		boardgame.DefaultMove{
			"Flip Card Between Hidden and Revealed",
			"Flips the card between hidden and revealed",
		},
	}

	return result
}

func (m *moveFlipHiddenCard) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	game, _ := concreteStates(state)

	if game.HiddenCard.NumComponents() < 1 && game.RevealedCard.NumComponents() < 1 {
		return errors.New("Neither the HiddenCard nor RevealedCard is set")
	}

	if game.HiddenCard.NumComponents() > 0 && game.RevealedCard.NumComponents() > 0 {
		return errors.New("Both hidden and revealed are full!")
	}

	return nil
}

func (m *moveFlipHiddenCard) Apply(state boardgame.MutableState) error {

	game, _ := concreteStates(state)

	from := game.RevealedCard
	to := game.HiddenCard

	if game.HiddenCard.NumComponents() > 0 {
		from = game.HiddenCard
		to = game.RevealedCard
	}

	if err := from.MoveComponent(boardgame.FirstComponentIndex, to, boardgame.FirstSlotIndex); err != nil {
		return err
	}

	return nil
}

/**************************************************
 *
 * moveMoveCardBetweenFanStacks Implementation
 *
 **************************************************/

func MoveMoveCardBetweenFanStacksFactory(state boardgame.State) boardgame.Move {
	result := &moveMoveCardBetweenFanStacks{
		boardgame.DefaultMove{
			"Move Fan Card",
			"Moves a card from or to Fan and Fan Discard",
		},
	}

	return result
}

func (m *moveMoveCardBetweenFanStacks) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	game, _ := concreteStates(state)

	if game.FanStack.NumComponents() == 6 && game.FanDiscard.NumComponents() == 3 {
		return nil
	}

	if game.FanStack.NumComponents() == 5 && game.FanDiscard.NumComponents() == 4 {
		return nil
	}

	return errors.New("Fan stacks aren't in known toggle state")
}

func (m *moveMoveCardBetweenFanStacks) Apply(state boardgame.MutableState) error {

	game, _ := concreteStates(state)

	from := game.FanStack
	to := game.FanDiscard
	fromIndex := 2
	toIndex := boardgame.FirstSlotIndex

	if game.FanStack.NumComponents() < 6 {
		from = game.FanDiscard
		to = game.FanStack
		fromIndex = boardgame.FirstComponentIndex
		toIndex = 2
	}

	if err := from.MoveComponent(fromIndex, to, toIndex); err != nil {
		return err
	}

	return nil
}
