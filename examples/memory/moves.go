package memory

import (
	"errors"
	"time"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

const hideCardsDuration = 4 * time.Second

/**************************************************
 *
 * moveRevealCard Implementation
 *
 **************************************************/

//boardgame:codegen
type moveRevealCard struct {
	moves.CurrentPlayer
	CardIndex int
}

func (m *moveRevealCard) DefaultsForState(state boardgame.ImmutableState) {

	m.CurrentPlayer.DefaultsForState(state)

	game, _ := concreteStates(state)

	for i, c := range game.HiddenCards.Components() {
		if c != nil {
			m.CardIndex = i
			break
		}
	}
}

func (m *moveRevealCard) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	if p.CardsLeftToReveal < 1 {
		return errors.New("You have no cards left to reveal this turn")
	}

	if m.CardIndex < 0 || m.CardIndex >= game.HiddenCards.Len() {
		return errors.New("illegal card index")
	}

	if game.HiddenCards.ComponentAt(m.CardIndex) == nil {
		if game.VisibleCards.ComponentAt(m.CardIndex) == nil {
			return errors.New("there is no card at that index")
		}
		return errors.New("that card has already been revealed")
	}

	return nil
}

func (m *moveRevealCard) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	p.CardsLeftToReveal--
	game.HiddenCards.ComponentAt(m.CardIndex).MoveTo(game.VisibleCards, m.CardIndex)

	//If the cards are the same, the FixUpMove CaptureCards will fire after this.

	return nil
}

/**************************************************
 *
 * moveStartHideCardsTimer Implementation
 *
 **************************************************/

//boardgame:codegen
type moveStartHideCardsTimer struct {
	moves.FixUp
}

func (m *moveStartHideCardsTimer) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.VisibleCards.NumComponents() != 2 {
		return errors.New("there aren't two cards showing")
	}

	if game.HideCardsTimer.Active() {
		return errors.New("the timer is already active")
	}

	var revealedCards []boardgame.Component

	for _, c := range game.VisibleCards.Components() {
		if c != nil {
			revealedCards = append(revealedCards, c)
		}
	}

	cardOneType := revealedCards[0].Values().(*cardValue).Type
	cardTwoType := revealedCards[1].Values().(*cardValue).Type

	if cardOneType == cardTwoType {
		return errors.New("The two revealed cards are of the same type")
	}

	return nil
}

func (m *moveStartHideCardsTimer) Apply(state boardgame.State) error {
	game, _ := concreteStates(state)

	move := state.Game().MoveByName(hideCardMoveName)

	game.HideCardsTimer.Start(hideCardsDuration, move)

	return nil
}

/**************************************************
 *
 * moveCaptureCards Implementation
 *
 **************************************************/

//boardgame:codegen
type moveCaptureCards struct {
	moves.FixUp
}

func (m *moveCaptureCards) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.VisibleCards.NumComponents() != 2 {
		return errors.New("there aren't two cards showing")
	}

	var revealedCards []boardgame.Component

	for _, c := range game.VisibleCards.Components() {
		if c != nil {
			revealedCards = append(revealedCards, c)
		}
	}

	cardOneType := revealedCards[0].Values().(*cardValue).Type
	cardTwoType := revealedCards[1].Values().(*cardValue).Type

	if cardOneType != cardTwoType {
		return errors.New("The two revealed cards are not of the same type")
	}

	return nil
}

func (m *moveCaptureCards) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	for i, c := range game.VisibleCards.Components() {
		if c != nil {
			game.VisibleCards.ComponentAt(i).MoveToNextSlot(p.WonCards)
		}
	}

	return nil
}

/**************************************************
 *
 * moveHideCards Implementation
 *
 **************************************************/

//boardgame:codegen
type moveHideCards struct {
	moves.CurrentPlayer
}

func (m *moveHideCards) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	if p.CardsLeftToReveal > 0 {
		return errors.New("You still have to reveal more cards before your turn is over")
	}

	if game.VisibleCards.NumComponents() < 1 {
		return errors.New("no cards left to hide")
	}

	return nil
}

func (m *moveHideCards) Apply(state boardgame.State) error {
	game, _ := concreteStates(state)

	//Cancel a timer in case it was still going.
	game.HideCardsTimer.Cancel()

	for i, c := range game.VisibleCards.Components() {
		if c != nil {
			if err := c.MoveTo(game.HiddenCards, i); err != nil {
				return errors.New("Couldn't move component: " + err.Error())
			}
		}
	}

	return nil
}
