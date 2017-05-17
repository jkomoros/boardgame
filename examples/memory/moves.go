package memory

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"time"
)

//+autoreader readsetter
type MoveAdvanceNextPlayer struct{}

//+autoreader readsetter
type MoveRevealCard struct {
	TargetPlayerIndex boardgame.PlayerIndex
	CardIndex         int
}

//+autoreader readsetter
type MoveStartHideCardsTimer struct{}

//+autoreader readsetter
type MoveCaptureCards struct{}

//+autoreader readsetter
type MoveHideCards struct {
	TargetPlayerIndex boardgame.PlayerIndex
}

const HideCardsDuration = 4 * time.Second

/**************************************************
 *
 * MoveAdvanceNextPlayer Implementation
 *
 **************************************************/

func MoveAdvanceNextPlayerFactory(state boardgame.State) boardgame.Move {
	return &MoveAdvanceNextPlayer{}
}

func (m *MoveAdvanceNextPlayer) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	if p.CardsLeftToReveal > 0 {
		return errors.New("The current player still has cards left to reveal")
	}

	if game.RevealedCards.NumComponents() > 0 {
		return errors.New("There are still some cards revealed. The current player must play MoveHideCards")
	}

	return nil
}

func (m *MoveAdvanceNextPlayer) Apply(state boardgame.MutableState) error {
	game, players := concreteStates(state)

	game.CurrentPlayer = game.CurrentPlayer.Next(state)

	p := players[game.CurrentPlayer]

	p.CardsLeftToReveal = 2

	return nil
}

func (m *MoveAdvanceNextPlayer) Name() string {
	return "Advance To Next Player"
}

func (m *MoveAdvanceNextPlayer) Description() string {
	return "Advances to the next player when the current player has no more legal moves."
}

func (t *MoveAdvanceNextPlayer) ImmediateFixUp(state boardgame.State) boardgame.Move {
	return nil
}

/**************************************************
 *
 * MoveRevealCard Implementation
 *
 **************************************************/

func MoveRevealCardFactory(state boardgame.State) boardgame.Move {
	result := &MoveRevealCard{}

	if state != nil {

		result.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()

		game, _ := concreteStates(state)

		for i, c := range game.HiddenCards.Components() {
			if c != nil {
				result.CardIndex = i
				break
			}
		}
	}

	return result
}

func (m *MoveRevealCard) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	game, players := concreteStates(state)

	if !m.TargetPlayerIndex.Equivalent(game.CurrentPlayer) {
		return errors.New("It isn't your turn")
	}

	//TargetPlayer is set automatically, so this is the message that will be
	//shown if the user tries to move when it's not their turn.
	if !m.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("It isn't your turn")
	}

	p := players[game.CurrentPlayer]

	if p.CardsLeftToReveal < 1 {
		return errors.New("You have no cards left to reveal this turn")
	}

	if m.CardIndex < 0 || m.CardIndex >= game.HiddenCards.Len() {
		return errors.New("Illegal card index.")
	}

	if game.HiddenCards.ComponentAt(m.CardIndex) == nil {
		if game.RevealedCards.ComponentAt(m.CardIndex) == nil {
			return errors.New("There is no card at that index.")
		} else {
			return errors.New("That card has already been revealed.")
		}
	}

	return nil
}

func (m *MoveRevealCard) Apply(state boardgame.MutableState) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	p.CardsLeftToReveal--
	game.HiddenCards.MoveComponent(m.CardIndex, game.RevealedCards, m.CardIndex)

	//If the cards are the same, the FixUpMove CaptureCards will fire after this.

	return nil
}

func (m *MoveRevealCard) Name() string {
	return "Reveal Card"
}

func (m *MoveRevealCard) Description() string {
	return "Reveals the card at the specified location"
}

func (t *MoveRevealCard) ImmediateFixUp(state boardgame.State) boardgame.Move {
	return nil
}

/**************************************************
 *
 * MoveStartHideCardsTimer Implementation
 *
 **************************************************/

func MoveStartHideCardsTimerFactory(state boardgame.State) boardgame.Move {
	return &MoveStartHideCardsTimer{}
}

func (m *MoveStartHideCardsTimer) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	game, _ := concreteStates(state)

	if game.RevealedCards.NumComponents() != 2 {
		return errors.New("There aren't two cards showing!")
	}

	if game.HideCardsTimer.Active() {
		return errors.New("The timer is already active.")
	}

	var revealedCards []*boardgame.Component

	for _, c := range game.RevealedCards.Components() {
		if c != nil {
			revealedCards = append(revealedCards, c)
		}
	}

	cardOneType := revealedCards[0].Values.(*cardValue).Type
	cardTwoType := revealedCards[1].Values.(*cardValue).Type

	if cardOneType == cardTwoType {
		return errors.New("The two revealed cards are of the same type")
	}

	return nil
}

func (m *MoveStartHideCardsTimer) Apply(state boardgame.MutableState) error {
	game, _ := concreteStates(state)

	game.HideCardsTimer.Start(HideCardsDuration, MoveHideCardsFactory(state))

	return nil
}

func (m *MoveStartHideCardsTimer) Name() string {
	return "Start Hide Cards Timer"
}

func (m *MoveStartHideCardsTimer) Description() string {
	return "If two cards are showing and they are not the same type and the timer is not active, start a timer to automatically hide them."
}

func (t *MoveStartHideCardsTimer) ImmediateFixUp(state boardgame.State) boardgame.Move {
	return nil
}

/**************************************************
 *
 * MoveCaptureCards Implementation
 *
 **************************************************/

func MoveCaptureCardsFactory(state boardgame.State) boardgame.Move {
	return &MoveCaptureCards{}
}

func (m *MoveCaptureCards) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	game, _ := concreteStates(state)

	if game.RevealedCards.NumComponents() != 2 {
		return errors.New("There aren't two cards showing!")
	}

	var revealedCards []*boardgame.Component

	for _, c := range game.RevealedCards.Components() {
		if c != nil {
			revealedCards = append(revealedCards, c)
		}
	}

	cardOneType := revealedCards[0].Values.(*cardValue).Type
	cardTwoType := revealedCards[1].Values.(*cardValue).Type

	if cardOneType != cardTwoType {
		return errors.New("The two revealed cards are not of the same type")
	}

	return nil
}

func (m *MoveCaptureCards) Apply(state boardgame.MutableState) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	for i, c := range game.RevealedCards.Components() {
		if c != nil {
			game.RevealedCards.MoveComponent(i, p.WonCards, boardgame.NextSlotIndex)
		}
	}

	return nil
}

func (m *MoveCaptureCards) Name() string {
	return "Capture Cards"
}

func (m *MoveCaptureCards) Description() string {
	return "If two cards are showing and they are the same type, capture them to the current player's hand."
}

func (t *MoveCaptureCards) ImmediateFixUp(state boardgame.State) boardgame.Move {
	return nil
}

/**************************************************
 *
 * MoveHideCards Implementation
 *
 **************************************************/

func MoveHideCardsFactory(state boardgame.State) boardgame.Move {
	result := &MoveHideCards{}

	if state != nil {
		result.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
	}

	return result
}

func (m *MoveHideCards) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	if !m.TargetPlayerIndex.Equivalent(game.CurrentPlayer) {
		return errors.New("It's not your turn")
	}

	//TargetPlayer is set automatically, so this is the message that will be
	//shown if the user tries to move when it's not their turn.
	if !m.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("It's not your turn")
	}

	if p.CardsLeftToReveal > 0 {
		return errors.New("You still have to reveal more cards before your turn is over")
	}

	if game.RevealedCards.NumComponents() < 1 {
		return errors.New("No cards left to hide!")
	}

	return nil
}

func (m *MoveHideCards) Apply(state boardgame.MutableState) error {
	game, _ := concreteStates(state)

	//Cancel a timer in case it was still going.
	game.HideCardsTimer.Cancel()

	for i, c := range game.RevealedCards.Components() {
		if c != nil {
			game.RevealedCards.MoveComponent(i, game.HiddenCards, i)
		}
	}

	return nil
}

func (m *MoveHideCards) Name() string {
	return "Hide Cards"
}

func (m *MoveHideCards) Description() string {
	return "After the current player has revealed both cards and tried to memorize them, this move hides the cards so that play can continue to next player."
}

func (t *MoveHideCards) ImmediateFixUp(state boardgame.State) boardgame.Move {
	return nil
}
