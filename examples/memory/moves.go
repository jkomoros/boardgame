package memory

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"time"
)

type MoveAdvanceNextPlayer struct{}

type MoveRevealCard struct {
	TargetPlayerIndex int
	CardIndex         int
}

type MoveStartHideCardsTimer struct{}

type MoveCaptureCards struct{}

type MoveHideCards struct{}

const HideCardsDuration = 4 * time.Second

/**************************************************
 *
 * MoveAdvanceNextPlayer Implementation
 *
 **************************************************/

func (m *MoveAdvanceNextPlayer) Legal(state boardgame.State) error {
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

	game.CurrentPlayer++

	if game.CurrentPlayer >= len(players) {
		game.CurrentPlayer = 0
	}

	p := players[game.CurrentPlayer]

	p.CardsLeftToReveal = 2

	return nil
}

func (m *MoveAdvanceNextPlayer) Copy() boardgame.Move {
	var result MoveAdvanceNextPlayer
	result = *m
	return &result
}

func (m *MoveAdvanceNextPlayer) DefaultsForState(state boardgame.State) {
	//Nothing to do
}

func (m *MoveAdvanceNextPlayer) Name() string {
	return "Advance To Next Player"
}

func (m *MoveAdvanceNextPlayer) Description() string {
	return "Advances to the next player when the current player has no more legal moves."
}

func (m *MoveAdvanceNextPlayer) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}

/**************************************************
 *
 * MoveRevealCard Implementation
 *
 **************************************************/

func (m *MoveRevealCard) Legal(state boardgame.State) error {
	game, players := concreteStates(state)

	if game.CurrentPlayer != m.TargetPlayerIndex {
		return errors.New("The target player is not the current player")
	}

	p := players[game.CurrentPlayer]

	if p.CardsLeftToReveal < 1 {
		return errors.New("The current player has no cards left to reveal")
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

func (m *MoveRevealCard) Copy() boardgame.Move {
	var result MoveRevealCard
	result = *m
	return &result
}

func (m *MoveRevealCard) DefaultsForState(state boardgame.State) {
	game, _ := concreteStates(state)
	m.TargetPlayerIndex = game.CurrentPlayer

	for i, c := range game.HiddenCards.Components() {
		if c != nil {
			m.CardIndex = i
			return
		}
	}

}

func (m *MoveRevealCard) Name() string {
	return "Reveal Card"
}

func (m *MoveRevealCard) Description() string {
	return "Reveals the card at the specified location"
}

func (m *MoveRevealCard) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}

/**************************************************
 *
 * MoveStartHideCardsTimer Implementation
 *
 **************************************************/

func (m *MoveStartHideCardsTimer) Legal(state boardgame.State) error {
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

	game.HideCardsTimer.Start(HideCardsDuration, &MoveHideCards{})

	return nil
}

func (m *MoveStartHideCardsTimer) Copy() boardgame.Move {
	var result MoveStartHideCardsTimer
	result = *m
	return &result
}

func (m *MoveStartHideCardsTimer) DefaultsForState(state boardgame.State) {
	//Nothing to do
}

func (m *MoveStartHideCardsTimer) Name() string {
	return "Start Hide Cards Timer"
}

func (m *MoveStartHideCardsTimer) Description() string {
	return "If two cards are showing and they are not the same type and the timer is not active, start a timer to automatically hide them."
}

func (m *MoveStartHideCardsTimer) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}

/**************************************************
 *
 * MoveCaptureCards Implementation
 *
 **************************************************/

func (m *MoveCaptureCards) Legal(state boardgame.State) error {
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

func (m *MoveCaptureCards) Copy() boardgame.Move {
	var result MoveCaptureCards
	result = *m
	return &result
}

func (m *MoveCaptureCards) DefaultsForState(state boardgame.State) {
	//Nothing to do
}

func (m *MoveCaptureCards) Name() string {
	return "Capture Cards"
}

func (m *MoveCaptureCards) Description() string {
	return "If two cards are showing and they are the same type, capture them to the current player's hand."
}

func (m *MoveCaptureCards) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}

/**************************************************
 *
 * MoveHideCards Implementation
 *
 **************************************************/

func (m *MoveHideCards) Legal(state boardgame.State) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	if p.CardsLeftToReveal > 0 {
		return errors.New("The current player still has cards left to reveal")
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

func (m *MoveHideCards) Copy() boardgame.Move {
	var result MoveHideCards
	result = *m
	return &result
}

func (m *MoveHideCards) DefaultsForState(state boardgame.State) {
	//Nothing to do
}

func (m *MoveHideCards) Name() string {
	return "Hide Cards"
}

func (m *MoveHideCards) Description() string {
	return "After the current player has revealed both cards and tried to memorize them, this move hides the cards so that play can continue to next player."
}

func (m *MoveHideCards) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}
