package blackjack

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

//boardgame:codegen
type moveShuffleDiscardToDraw struct {
	moves.FixUp
}

//boardgame:codegen
type moveFinishTurn struct {
	moves.FinishTurn
}

//boardgame:codegen
type moveRevealHiddenCard struct {
	moves.CurrentPlayer
}

//boardgame:codegen
type moveCurrentPlayerHit struct {
	moves.CurrentPlayer
}

//boardgame:codegen
type moveCurrentPlayerStand struct {
	moves.CurrentPlayer
}

/**************************************************
 *
 * moveShuffleDiscardToDraw Implementation
 *
 **************************************************/

func (m *moveShuffleDiscardToDraw) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.FixUp.Legal(state, proposer); err != nil {
		return err
	}

	game, _ := concreteStates(state)

	if game.DrawStack.Len() > 0 {
		return errors.New("The draw stack is not yet empty")
	}

	return nil
}

func (m *moveShuffleDiscardToDraw) Apply(state boardgame.State) error {
	game, _ := concreteStates(state)

	game.DiscardStack.MoveAllTo(game.DrawStack)
	game.DrawStack.Shuffle()

	return nil
}

/**************************************************
 *
 * moveCurrentPlayerHit Implementation
 *
 **************************************************/

func (m *moveCurrentPlayerHit) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

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

func (m *moveCurrentPlayerHit) Apply(state boardgame.State) error {
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
 * moveCurrentPlayerStand Implementation
 *
 **************************************************/

func (m *moveCurrentPlayerStand) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer]

	if currentPlayer.Busted {
		return errors.New("the current player has already busted")
	}

	if currentPlayer.Stood {
		return errors.New("the current player already stood")
	}

	return nil

}

func (m *moveCurrentPlayerStand) Apply(state boardgame.State) error {

	game, players := concreteStates(state)

	currentPlayer := players[game.CurrentPlayer]

	currentPlayer.Stood = true

	return nil
}

/**************************************************
 *
 * moveRevealHiddenCard Implementation
 *
 **************************************************/

func (m *moveRevealHiddenCard) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

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

func (m *moveRevealHiddenCard) Apply(state boardgame.State) error {
	_, players := concreteStates(state)

	p := players[m.TargetPlayerIndex]

	p.HiddenHand.First().MoveToFirstSlot(p.VisibleHand)

	return nil
}
