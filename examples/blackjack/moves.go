package blackjack

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

//boardgame:codegen
type MoveShuffleDiscardToDraw struct {
	moves.FixUp
}

//boardgame:codegen
type MoveFinishTurn struct {
	moves.FinishTurn
}

//boardgame:codegen
type MoveRevealHiddenCard struct {
	moves.CurrentPlayer
}

//boardgame:codegen
type MoveCurrentPlayerHit struct {
	moves.CurrentPlayer
}

//boardgame:codegen
type MoveCurrentPlayerStand struct {
	moves.CurrentPlayer
}

/**************************************************
 *
 * MoveShuffleDiscardToDraw Implementation
 *
 **************************************************/

func (m *MoveShuffleDiscardToDraw) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.DrawStack.Len() > 0 {
		return errors.New("The draw stack is not yet empty")
	}

	return nil
}

func (m *MoveShuffleDiscardToDraw) Apply(state boardgame.State) error {
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

func (m *MoveCurrentPlayerHit) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

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

func (m *MoveCurrentPlayerHit) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer]

	game.DrawStack.First().MoveToFirstSlot(currentPlayer.VisibleHand)

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

func (m *MoveCurrentPlayerStand) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

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

func (m *MoveCurrentPlayerStand) Apply(state boardgame.State) error {

	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer]

	currentPlayer.Stood = true

	return nil
}

/**************************************************
 *
 * MoveRevealHiddenCard Implementation
 *
 **************************************************/

func (m *MoveRevealHiddenCard) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

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

func (m *MoveRevealHiddenCard) Apply(state boardgame.State) error {
	_, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	p.HiddenHand.First().MoveToFirstSlot(p.VisibleHand)

	return nil
}
