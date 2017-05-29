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

	if err := from.MoveComponent(boardgame.FirstComponentIndex, to, boardgame.LastSlotIndex); err != nil {
		return err
	}

	return nil
}
