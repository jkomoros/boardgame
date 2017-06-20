package debuganimations

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

//+autoreader readsetter
type moveMoveCardBetweenShortStacks struct {
	boardgame.BaseMove
	FromFirst bool
}

//+autoreader readsetter
type moveMoveCardBetweenDrawAndDiscardStacks struct {
	boardgame.BaseMove
	FromDraw bool
}

//+autoreader readsetter
type moveFlipHiddenCard struct {
	boardgame.BaseMove
}

//+autoreader readsetter
type moveMoveCardBetweenFanStacks struct {
	boardgame.BaseMove
}

//+autoreader readsetter
type moveVisibleShuffleCards struct {
	boardgame.BaseMove
}

//+autoreader readsetter
type moveShuffleCards struct {
	boardgame.BaseMove
}

//+autoreader readsetter
type moveMoveBetweenHidden struct {
	boardgame.BaseMove
}

/**************************************************
 *
 * moveMoveCardBetweenShortStacks Implementation
 *
 **************************************************/

var moveMoveCardBetweenShortStacksConfig = boardgame.MoveTypeConfig{
	Name:     "Move Card Between Short Stacks",
	HelpText: "Moves a card between two short stacks",
	MoveConstructor: func(mType *boardgame.MoveType) boardgame.Move {
		return &moveMoveCardBetweenShortStacks{
			BaseMove: boardgame.BaseMove{mType},
		}
	},
}

func (m *moveMoveCardBetweenShortStacks) DefaultsForState(state boardgame.State) {
	gameState, _ := concreteStates(state)

	if gameState.FirstShortStack.NumComponents() < 1 {
		m.FromFirst = false
	} else {
		m.FromFirst = true
	}
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

var moveMoveCardBetweenDrawAndDiscardStacksConfig = boardgame.MoveTypeConfig{
	Name:     "Move Card Between Draw and Discard Stacks",
	HelpText: "Moves a card between draw and discard stacks",
	MoveConstructor: func(mType *boardgame.MoveType) boardgame.Move {
		return &moveMoveCardBetweenDrawAndDiscardStacks{
			BaseMove: boardgame.BaseMove{mType},
		}
	},
}

func (m *moveMoveCardBetweenDrawAndDiscardStacks) DefaultsForState(state boardgame.State) {
	gameState, _ := concreteStates(state)

	if gameState.DiscardStack.NumComponents() < 3 {
		m.FromDraw = true
	} else {
		m.FromDraw = false
	}
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

var moveFlipHiddenCardConfig = boardgame.MoveTypeConfig{
	Name:     "Flip Card Between Hidden and Revealed",
	HelpText: "Flips the card between hidden and revealed",
	MoveConstructor: func(mType *boardgame.MoveType) boardgame.Move {
		return &moveFlipHiddenCard{
			BaseMove: boardgame.BaseMove{mType},
		}
	},
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

var moveMoveCardBetweenFanStacksConfig = boardgame.MoveTypeConfig{
	Name:     "Move Fan Card",
	HelpText: "Moves a card from or to Fan and Fan Discard",
	MoveConstructor: func(mType *boardgame.MoveType) boardgame.Move {
		return &moveMoveCardBetweenFanStacks{
			BaseMove: boardgame.BaseMove{mType},
		}
	},
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

/**************************************************
 *
 * moveVisibleShuffleCards Implementation
 *
 **************************************************/

var moveVisibleShuffleCardsConfig = boardgame.MoveTypeConfig{
	Name:     "Visible Shuffle",
	HelpText: "Performs a visible shuffle",
	MoveConstructor: func(mType *boardgame.MoveType) boardgame.Move {
		return &moveVisibleShuffleCards{
			BaseMove: boardgame.BaseMove{mType},
		}
	},
}

func (m *moveVisibleShuffleCards) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	game, _ := concreteStates(state)

	if game.FanStack.NumComponents() > 1 {
		return nil
	}

	return errors.New("Aren't enough cards to shuffle")
}

func (m *moveVisibleShuffleCards) Apply(state boardgame.MutableState) error {

	game, _ := concreteStates(state)

	return game.FanStack.PublicShuffle()

}

/**************************************************
 *
 * moveShuffleCards Implementation
 *
 **************************************************/

var moveShuffleCardsConfig = boardgame.MoveTypeConfig{
	Name:     "Shuffle",
	HelpText: "Performs a secret shuffle",
	MoveConstructor: func(mType *boardgame.MoveType) boardgame.Move {
		return &moveShuffleCards{
			BaseMove: boardgame.BaseMove{mType},
		}
	},
}

func (m *moveShuffleCards) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	game, _ := concreteStates(state)

	if game.FanStack.NumComponents() > 1 {
		return nil
	}

	return errors.New("Aren't enough cards to shuffle")
}

func (m *moveShuffleCards) Apply(state boardgame.MutableState) error {

	game, _ := concreteStates(state)

	return game.FanStack.Shuffle()

}

/**************************************************
 *
 * moveMoveBetweenHidden Implementation
 *
 **************************************************/

var moveMoveBetweenHiddenConfig = boardgame.MoveTypeConfig{
	Name:     "Move Between Hidden",
	HelpText: "Moves between hidden and visible stacks",
	MoveConstructor: func(mType *boardgame.MoveType) boardgame.Move {
		return &moveMoveBetweenHidden{
			BaseMove: boardgame.BaseMove{mType},
		}
	},
}

func (m *moveMoveBetweenHidden) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	game, _ := concreteStates(state)

	if game.VisibleStack.NumComponents() == 5 && game.HiddenStack.NumComponents() == 4 {
		return nil
	}

	if game.VisibleStack.NumComponents() == 4 && game.HiddenStack.NumComponents() == 5 {
		return nil
	}

	return errors.New("Cards aren't in known position")
}

func (m *moveMoveBetweenHidden) Apply(state boardgame.MutableState) error {

	game, _ := concreteStates(state)

	from := game.VisibleStack
	to := game.HiddenStack
	fromIndex := 2
	toIndex := boardgame.FirstSlotIndex

	if game.VisibleStack.NumComponents() < 5 {
		from = game.HiddenStack
		to = game.VisibleStack
		fromIndex = boardgame.FirstComponentIndex
		toIndex = 2
	}

	if err := from.MoveComponent(fromIndex, to, toIndex); err != nil {
		return err
	}

	return nil

}
