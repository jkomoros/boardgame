package blackjack

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

//+autoreader
type MoveShuffleDiscardToDraw struct {
	moves.Base
}

//+autoreader
type MoveFinishTurn struct {
	moves.FinishTurn
}

//+autoreader
type MoveDealInitialHiddenCard struct {
	moves.DealComponents
}

//+autoreader
type MoveDealInitialVisibleCard struct {
	moves.DealComponents
}

//+autoreader
type MoveRevealHiddenCard struct {
	moves.CurrentPlayer
}

//+autoreader
type MoveCurrentPlayerHit struct {
	moves.CurrentPlayer
}

//+autoreader
type MoveCurrentPlayerStand struct {
	moves.CurrentPlayer
}

/**************************************************
 *
 * MoveShuffleDiscardToDraw Implementation
 *
 **************************************************/

var moveShuffleDiscardToDrawConfig = boardgame.MoveTypeConfig{
	Name:     "Shuffle Discard To Draw",
	HelpText: "When the draw deck is empty, shuffles the discard deck into draw deck.",
	MoveConstructor: func() boardgame.Move {
		return new(MoveShuffleDiscardToDraw)
	},
	IsFixUp: true,
}

func (m *MoveShuffleDiscardToDraw) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := m.Base.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.DrawStack.Len() > 0 {
		return errors.New("The draw stack is not yet empty")
	}

	return nil
}

func (m *MoveShuffleDiscardToDraw) Apply(state boardgame.MutableState) error {
	game, _ := concreteStates(state)

	game.DiscardStack.MoveAllTo(game.DrawStack)
	game.DrawStack.Shuffle()

	return nil
}

/**************************************************
 *
 * MoveCurrentPlayerHit Implementation
 *
 **************************************************/

var moveCurrentPlayerHitConfig = boardgame.MoveTypeConfig{
	Name:     "Current Player Hit",
	HelpText: "The current player hits, drawing a card.",
	MoveConstructor: func() boardgame.Move {
		return new(MoveCurrentPlayerHit)
	},
	LegalPhases: []int{
		PhaseNormalPlay,
	},
}

func (m *MoveCurrentPlayerHit) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer]

	if currentPlayer.Busted {
		return errors.New("Current player is busted")
	}

	if currentPlayer.HandValue() >= targetScore {
		return errors.New("Current player is already at target scores")
	}

	return nil
}

func (m *MoveCurrentPlayerHit) Apply(state boardgame.MutableState) error {
	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer]

	game.DrawStack.MoveComponent(boardgame.FirstComponentIndex, currentPlayer.VisibleHand, boardgame.FirstSlotIndex)

	handValue := currentPlayer.HandValue()

	if handValue > targetScore {
		currentPlayer.Busted = true
	}

	if handValue == targetScore {
		currentPlayer.Stood = true
	}

	return nil
}

/**************************************************
 *
 * MoveCurrentPlayerStand Implementation
 *
 **************************************************/

var moveCurrentPlayerStandConfig = boardgame.MoveTypeConfig{
	Name:     "Current Player Stand",
	HelpText: "If the current player no longer wants to draw cards, they can stand.",
	MoveConstructor: func() boardgame.Move {
		return new(MoveCurrentPlayerStand)
	},
	LegalPhases: []int{
		PhaseNormalPlay,
	},
}

func (m *MoveCurrentPlayerStand) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer]

	if currentPlayer.Busted {
		return errors.New("The current player has already busted.")
	}

	if currentPlayer.Stood {
		return errors.New("The current player already stood.")
	}

	return nil

}

func (m *MoveCurrentPlayerStand) Apply(state boardgame.MutableState) error {

	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer]

	currentPlayer.Stood = true

	return nil
}

/**************************************************
 *
 * MoveAdvanceNextPlayer Implementation
 *
 **************************************************/

var moveFinishTurnConfig = boardgame.MoveTypeConfig{
	Name:     "Finish Turn",
	HelpText: "When the current player has either busted or decided to stand, we advance to next player.",
	MoveConstructor: func() boardgame.Move {
		return new(MoveFinishTurn)
	},
	IsFixUp: true,
	LegalPhases: []int{
		PhaseNormalPlay,
	},
}

/**************************************************
 *
 * MoveRevealHiddenCard Implementation
 *
 **************************************************/

var moveRevealHiddenCardConfig = boardgame.MoveTypeConfig{
	Name:     "Reveal Hidden Card",
	HelpText: "Reveals the hidden card in the user's hand",
	MoveConstructor: func() boardgame.Move {
		return new(MoveRevealHiddenCard)
	},
	IsFixUp: true,
	LegalPhases: []int{
		PhaseNormalPlay,
	},
}

func (m *MoveRevealHiddenCard) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	_, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	if p.HiddenHand.NumComponents() < 1 {
		return errors.New("Target player has no cards to reveal")
	}

	return nil
}

func (m *MoveRevealHiddenCard) Apply(state boardgame.MutableState) error {
	_, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	p.HiddenHand.MoveComponent(boardgame.FirstComponentIndex, p.VisibleHand, boardgame.FirstSlotIndex)

	return nil
}

/**************************************************
 *
 * MoveDealInitialHiddenCard Implementation
 *
 **************************************************/

var moveDealInitialHiddenCardConfig = boardgame.MoveTypeConfig{
	Name:     "Deal Initial Hidden Card",
	HelpText: "Deals a hidden card to each player",
	MoveConstructor: func() boardgame.Move {
		return new(MoveDealInitialHiddenCard)
	},
	LegalPhases: []int{
		PhaseInitialDeal,
	},
	IsFixUp: true,
}

func (m *MoveDealInitialHiddenCard) GameStack(gState boardgame.MutableSubState) boardgame.MutableStack {
	return gState.(*gameState).DrawStack
}

func (m *MoveDealInitialHiddenCard) PlayerStack(pState boardgame.MutablePlayerState) boardgame.MutableStack {
	return pState.(*playerState).HiddenHand
}

/**************************************************
 *
 * MoveDealInitialVisbibleCard Implementation
 *
 **************************************************/

var moveDealInitialVisibleCardConfig = boardgame.MoveTypeConfig{
	Name:     "Deal Initial Visible Card",
	HelpText: "Deals a visible card to each player",
	MoveConstructor: func() boardgame.Move {
		return new(MoveDealInitialVisibleCard)
	},
	LegalPhases: []int{
		PhaseInitialDeal,
	},
	IsFixUp: true,
}

func (m *MoveDealInitialVisibleCard) GameStack(gState boardgame.MutableSubState) boardgame.MutableStack {
	return gState.(*gameState).DrawStack
}

func (m *MoveDealInitialVisibleCard) PlayerStack(pState boardgame.MutablePlayerState) boardgame.MutableStack {
	return pState.(*playerState).VisibleHand
}
