package memory

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
	"time"
)

//+autoreader readsetter
type MoveFinishTurn struct {
	moves.FinishTurn
}

//+autoreader readsetter
type MoveRevealCard struct {
	moves.CurrentPlayer
	CardIndex int
}

//+autoreader readsetter
type MoveStartHideCardsTimer struct {
	moves.Base
}

//+autoreader readsetter
type MoveCaptureCards struct {
	moves.Base
}

//+autoreader readsetter
type MoveHideCards struct {
	moves.CurrentPlayer
}

const HideCardsDuration = 4 * time.Second

/**************************************************
 *
 * MoveFinishTurn Implementation
 *
 **************************************************/

var moveFinishTurnConfig = boardgame.MoveTypeConfig{
	Name:     "Finish Turn",
	HelpText: "Advances to the next player when the current player has no more legal moves.",
	MoveConstructor: func() boardgame.Move {
		return new(MoveFinishTurn)
	},
	IsFixUp: true,
}

/**************************************************
 *
 * MoveRevealCard Implementation
 *
 **************************************************/

var moveRevealCardConfig = boardgame.MoveTypeConfig{
	Name:     "Reveal Card",
	HelpText: "Reveals the card at the specified location",
	MoveConstructor: func() boardgame.Move {
		return new(MoveRevealCard)
	},
}

func (m *MoveRevealCard) DefaultsForState(state boardgame.State) {

	m.CurrentPlayer.DefaultsForState(state)

	game, _ := concreteStates(state)

	for i, c := range game.HiddenCards.Components() {
		if c != nil {
			m.CardIndex = i
			break
		}
	}
}

func (m *MoveRevealCard) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

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

/**************************************************
 *
 * MoveStartHideCardsTimer Implementation
 *
 **************************************************/

var moveStartHideCardsTimerConfig = boardgame.MoveTypeConfig{
	Name:     "Start Hide Cards Timer",
	HelpText: "If two cards are showing and they are not the same type and the timer is not active, start a timer to automatically hide them.",
	MoveConstructor: func() boardgame.Move {
		return new(MoveStartHideCardsTimer)
	},
	IsFixUp: true,
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

	moveType := state.Game().Manager().PlayerMoveTypeByName(moveHideCardsConfig.Name)

	move := moveType.NewMove(state)

	game.HideCardsTimer.Start(HideCardsDuration, move)

	return nil
}

/**************************************************
 *
 * MoveCaptureCards Implementation
 *
 **************************************************/

var moveCaptureCardsConfig = boardgame.MoveTypeConfig{
	Name:     "Capture Cards",
	HelpText: "If two cards are showing and they are the same type, capture them to the current player's hand.",
	MoveConstructor: func() boardgame.Move {
		return new(MoveCaptureCards)
	},
	IsFixUp: true,
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

/**************************************************
 *
 * MoveHideCards Implementation
 *
 **************************************************/

var moveHideCardsConfig = boardgame.MoveTypeConfig{
	Name:     "Hide Cards",
	HelpText: "After the current player has revealed both cards and tried to memorize them, this move hides the cards so that play can continue to next player.",
	MoveConstructor: func() boardgame.Move {
		return new(MoveHideCards)
	},
}

func (m *MoveHideCards) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

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
