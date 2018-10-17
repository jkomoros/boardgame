package debuganimations

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

//boardgame:codegen
type moveMoveCardBetweenShortStacks struct {
	moves.Base
	FromFirst bool
}

//boardgame:codegen
type moveMoveCardBetweenDrawAndDiscardStacks struct {
	moves.Base
	FromDraw bool
}

//boardgame:codegen
type moveFlipHiddenCard struct {
	moves.Base
}

//boardgame:codegen
type moveMoveCardBetweenFanStacks struct {
	moves.Base
}

//boardgame:codegen
type moveVisibleShuffleCards struct {
	moves.Base
}

//boardgame:codegen
type moveShuffleCards struct {
	moves.Base
}

//boardgame:codegen
type moveMoveBetweenHidden struct {
	moves.Base
}

//boardgame:codegen
type moveMoveToken struct {
	moves.Base
}

//boardgame:codegen
type moveMoveTokenSanitized struct {
	moves.Base
}

//boardgame:codegen
type moveStartMoveAllComponentsToHidden struct {
	moves.Done
}

//boardgame:codegen
type moveStartMoveAllComponentsToVisible struct {
	moves.Done
}

/**************************************************
 *
 * moveMoveCardBetweenShortStacks Implementation
 *
 **************************************************/

func (m *moveMoveCardBetweenShortStacks) HelpText() string {
	return "Moves a card between two short stacks"
}

func (m *moveMoveCardBetweenShortStacks) DefaultsForState(state boardgame.ImmutableState) {
	gameState, _ := concreteStates(state)

	if gameState.FirstShortStack.NumComponents() < 1 {
		m.FromFirst = false
	} else {
		m.FromFirst = true
	}
}

func (m *moveMoveCardBetweenShortStacks) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Base.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.FirstShortStack.NumComponents() < 1 && m.FromFirst {
		return errors.New("First short stack has no cards to move")
	}

	if game.SecondShortStack.NumComponents() < 1 && !m.FromFirst {
		return errors.New("Second short stack has no cards to move")
	}

	return nil
}

func (m *moveMoveCardBetweenShortStacks) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	from := game.SecondShortStack
	to := game.FirstShortStack

	if m.FromFirst {
		from = game.FirstShortStack
		to = game.SecondShortStack
	}

	if err := from.First().MoveToFirstSlot(to); err != nil {
		return err
	}

	return nil
}

/**************************************************
 *
 * moveMoveCardBetweenDrawAndDiscardStacks Implementation
 *
 **************************************************/

func (m *moveMoveCardBetweenDrawAndDiscardStacks) HelpText() string {
	return "Moves a card between draw and discard stacks"
}

func (m *moveMoveCardBetweenDrawAndDiscardStacks) DefaultsForState(state boardgame.ImmutableState) {
	gameState, _ := concreteStates(state)

	if gameState.DiscardStack.NumComponents() < 3 {
		m.FromDraw = true
	} else {
		m.FromDraw = false
	}
}

func (m *moveMoveCardBetweenDrawAndDiscardStacks) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Base.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.DrawStack.NumComponents() < 1 && m.FromDraw {
		return errors.New("Draw stack has no cards to move")
	}

	if game.DiscardStack.NumComponents() < 1 && !m.FromDraw {
		return errors.New("Discard stack has no cards to move")
	}

	return nil
}

func (m *moveMoveCardBetweenDrawAndDiscardStacks) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	from := game.DiscardStack
	to := game.DrawStack

	if m.FromDraw {
		from = game.DrawStack
		to = game.DiscardStack
	}

	if err := from.First().MoveToFirstSlot(to); err != nil {
		return err
	}

	return nil
}

/**************************************************
 *
 * moveFlipHiddenCard Implementation
 *
 **************************************************/

func (m *moveFlipHiddenCard) HelpText() string {
	return "Flips the card between hidden and revealed"
}

func (m *moveFlipHiddenCard) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Base.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.HiddenCard.NumComponents() < 1 && game.VisibleCard.NumComponents() < 1 {
		return errors.New("Neither the HiddenCard nor RevealedCard is set")
	}

	if game.HiddenCard.NumComponents() > 0 && game.VisibleCard.NumComponents() > 0 {
		return errors.New("Both hidden and revealed are full!")
	}

	return nil
}

func (m *moveFlipHiddenCard) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	from := game.VisibleCard
	to := game.HiddenCard

	if game.HiddenCard.NumComponents() > 0 {
		from = game.HiddenCard
		to = game.VisibleCard
	}

	if err := from.First().MoveToFirstSlot(to); err != nil {
		return err
	}

	return nil
}

/**************************************************
 *
 * moveMoveCardBetweenFanStacks Implementation
 *
 **************************************************/

func (m *moveMoveCardBetweenFanStacks) HelpText() string {
	return "Moves a card from or to Fan and Fan Discard"
}

func (m *moveMoveCardBetweenFanStacks) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Base.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.FanStack.NumComponents() == 6 && game.FanDiscard.NumComponents() == 3 {
		return nil
	}

	if game.FanStack.NumComponents() == 5 && game.FanDiscard.NumComponents() == 4 {
		return nil
	}

	return errors.New("Fan stacks aren't in known toggle state")
}

func (m *moveMoveCardBetweenFanStacks) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	if game.FanStack.NumComponents() < 6 {
		return game.FanDiscard.First().MoveTo(game.FanStack, 2)
	}

	return game.FanStack.ComponentAt(2).MoveToFirstSlot(game.FanDiscard)
}

/**************************************************
 *
 * moveVisibleShuffleCards Implementation
 *
 **************************************************/

func (m *moveVisibleShuffleCards) HelpText() string {
	return "Performs a visible shuffle"
}

func (m *moveVisibleShuffleCards) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Base.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.FanStack.NumComponents() > 1 {
		return nil
	}

	return errors.New("Aren't enough cards to shuffle")
}

func (m *moveVisibleShuffleCards) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	return game.FanStack.PublicShuffle()

}

/**************************************************
 *
 * moveShuffleCards Implementation
 *
 **************************************************/

func (m *moveShuffleCards) HelpText() string {
	return "Performs a secret shuffle"
}

func (m *moveShuffleCards) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Base.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.FanStack.NumComponents() > 1 {
		return nil
	}

	return errors.New("Aren't enough cards to shuffle")
}

func (m *moveShuffleCards) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	return game.FanStack.Shuffle()

}

/**************************************************
 *
 * moveMoveBetweenHidden Implementation
 *
 **************************************************/

func (m *moveMoveBetweenHidden) HelpText() string {
	return "Moves between hidden and visible stacks"
}

func (m *moveMoveBetweenHidden) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Base.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.VisibleStack.NumComponents() == 5 && game.HiddenStack.NumComponents() == 4 {
		return nil
	}

	if game.VisibleStack.NumComponents() == 4 && game.HiddenStack.NumComponents() == 5 {
		return nil
	}

	return errors.New("Cards aren't in known position")
}

func (m *moveMoveBetweenHidden) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	if game.VisibleStack.NumComponents() < 5 {
		return game.HiddenStack.First().MoveTo(game.VisibleStack, 2)
	}

	return game.VisibleStack.ComponentAt(2).MoveToFirstSlot(game.HiddenStack)

}

/**************************************************
 *
 * moveMoveToken Implementation
 *
 **************************************************/

func (m *moveMoveToken) HelpText() string {
	return "Moves tokens"
}

func (m *moveMoveToken) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Base.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.TokensFrom.NumComponents() == 10 && game.TokensTo.NumComponents() == 9 {
		return nil
	}

	if game.TokensFrom.NumComponents() == 9 && game.TokensTo.NumComponents() == 10 {
		return nil
	}

	return errors.New("tokens aren't in known position")
}

func (m *moveMoveToken) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	if game.TokensFrom.NumComponents() < 10 {
		return game.TokensTo.First().MoveTo(game.TokensFrom, 2)
	}

	return game.TokensFrom.ComponentAt(2).MoveToFirstSlot(game.TokensTo)

}

/**************************************************
 *
 * moveMoveTokenSanitized Implementation
 *
 **************************************************/

func (m *moveMoveTokenSanitized) HelpText() string {
	return "Moves tokens"
}

func (m *moveMoveTokenSanitized) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.Base.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.SanitizedTokensFrom.NumComponents() == 10 && game.SanitizedTokensTo.NumComponents() == 9 {
		return nil
	}

	if game.SanitizedTokensFrom.NumComponents() == 9 && game.SanitizedTokensTo.NumComponents() == 10 {
		return nil
	}

	return errors.New("tokens aren't in known position")
}

func (m *moveMoveTokenSanitized) Apply(state boardgame.State) error {

	game, _ := concreteStates(state)

	if game.SanitizedTokensFrom.NumComponents() < 10 {
		return game.SanitizedTokensTo.First().MoveTo(game.SanitizedTokensFrom, 2)
	}

	return game.SanitizedTokensFrom.ComponentAt(2).MoveToFirstSlot(game.SanitizedTokensTo)

}

/**************************************************
 *
 * moveStartMoveAllComponentsToHidden Implementation
 *
 **************************************************/

func (m *moveStartMoveAllComponentsToHidden) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	//TODO: it's kind of weird that this has to be above the m.Done.Legal
	//check, right? Does that imply a bug in group matching logic?

	game := state.ImmutableGameState().(*gameState)

	if game.AllHiddenStack.NumComponents() > 0 {
		return errors.New("The hidden stack already has items. Use the 'To Visible' move")
	}

	if err := m.Done.Legal(state, proposer); err != nil {
		return err
	}

	return nil
}

/**************************************************
 *
 * moveStartMoveAllComponentsToVisible Implementation
 *
 **************************************************/

func (m *moveStartMoveAllComponentsToVisible) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	//TODO: it's kind of weird that this has to be above the m.Done.Legal check, right?

	game := state.ImmutableGameState().(*gameState)

	if game.AllVisibleStack.NumComponents() > 0 {
		return errors.New("The visible stack already has items. Use the 'To Hidden' move")
	}

	if err := m.Done.Legal(state, proposer); err != nil {
		return err
	}

	return nil
}
