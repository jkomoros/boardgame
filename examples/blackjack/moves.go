package blackjack

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

type MoveShuffleDiscardToDraw struct{}

type MoveAdvanceNextPlayer struct{}

type MoveCurrentPlayerHit struct {
	TargetPlayerIndex int
}

type MoveCurrentPlayerStand struct {
	TargetPlayerIndex int
}

/**************************************************
 *
 * MoveShuffleDiscardToDraw Implementation
 *
 **************************************************/

func (m *MoveShuffleDiscardToDraw) Legal(state boardgame.State) error {
	s := state.(*mainState)

	if s.Game.DrawStack.Len() > 0 {
		return errors.New("The draw stack is not yet empty")
	}

	return nil
}

func (m *MoveShuffleDiscardToDraw) Apply(state boardgame.State) error {
	s := state.(*mainState)

	s.Game.DiscardStack.MoveAllTo(s.Game.DrawStack)
	s.Game.DrawStack.Shuffle()

	return nil
}

func (m *MoveShuffleDiscardToDraw) Copy() boardgame.Move {
	var result MoveShuffleDiscardToDraw
	result = *m
	return &result
}

func (m *MoveShuffleDiscardToDraw) DefaultsForState(state boardgame.State) {
	//Nothing to do
}

func (m *MoveShuffleDiscardToDraw) Name() string {
	return "Shuffle Discard To Draw"
}

func (m *MoveShuffleDiscardToDraw) Description() string {
	return "When the draw deck is empty, shuffles the discard deck into draw deck."
}

func (m *MoveShuffleDiscardToDraw) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}

/**************************************************
 *
 * MoveCurrentPlayerHit Implementation
 *
 **************************************************/

func (m *MoveCurrentPlayerHit) Legal(state boardgame.State) error {
	s := state.(*mainState)

	if s.Game.CurrentPlayer != m.TargetPlayerIndex {
		return errors.New("The specified player is not the current player.")
	}

	currentPlayer := s.Players[s.Game.CurrentPlayer]

	if currentPlayer.Busted {
		return errors.New("Current player is busted")
	}

	handValue := currentPlayer.HandValue()

	if handValue >= targetScore {
		return errors.New("Current player is already at target scores")
	}

	return nil
}

func (m *MoveCurrentPlayerHit) Apply(state boardgame.State) error {
	s := state.(*mainState)

	currentPlayer := s.Players[s.Game.CurrentPlayer]

	currentPlayer.Hand.InsertFront(s.Game.DrawStack.RemoveFirst())

	handValue := currentPlayer.HandValue()

	if handValue > targetScore {
		currentPlayer.Busted = true
	}

	if handValue == targetScore {
		currentPlayer.Stood = true
	}

	return nil
}

func (m *MoveCurrentPlayerHit) Copy() boardgame.Move {
	var result MoveCurrentPlayerHit
	result = *m
	return &result
}

func (m *MoveCurrentPlayerHit) DefaultsForState(state boardgame.State) {
	s := state.(*mainState)

	m.TargetPlayerIndex = s.Game.CurrentPlayer
}

func (m *MoveCurrentPlayerHit) Name() string {
	return "Current Player Hit"
}

func (m *MoveCurrentPlayerHit) Description() string {
	return "The current player hits, drawing a card."
}

func (m *MoveCurrentPlayerHit) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}

/**************************************************
 *
 * MoveCurrentPlayerStand Implementation
 *
 **************************************************/

func (m *MoveCurrentPlayerStand) Legal(state boardgame.State) error {

	s := state.(*mainState)

	if s.Game.CurrentPlayer != m.TargetPlayerIndex {
		return errors.New("The specified player is not the current player.")
	}

	currentPlayer := s.Players[s.Game.CurrentPlayer]

	if currentPlayer.Busted {
		return errors.New("The current player has already busted.")
	}

	if currentPlayer.Stood {
		return errors.New("The current player already stood.")
	}

	return nil

}

func (m *MoveCurrentPlayerStand) Apply(state boardgame.State) error {
	s := state.(*mainState)

	currentPlayer := s.Players[s.Game.CurrentPlayer]

	currentPlayer.Stood = true

	return nil
}

func (m *MoveCurrentPlayerStand) Copy() boardgame.Move {
	var result MoveCurrentPlayerStand
	result = *m
	return &result
}

func (m *MoveCurrentPlayerStand) DefaultsForState(state boardgame.State) {
	s := state.(*mainState)
	m.TargetPlayerIndex = s.Game.CurrentPlayer
}

func (m *MoveCurrentPlayerStand) Name() string {
	return "Current Player Stand"
}

func (m *MoveCurrentPlayerStand) Description() string {
	return "If the current player no longer wants to draw cards, they can stand."
}

func (m *MoveCurrentPlayerStand) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}

/**************************************************
 *
 * MoveAdvanceNextPlayer Implementation
 *
 **************************************************/

func (m *MoveAdvanceNextPlayer) Legal(state boardgame.State) error {
	s := state.(*mainState)

	currentPlayer := s.Players[s.Game.CurrentPlayer]

	if currentPlayer.Busted || currentPlayer.Stood {
		return nil
	}

	return errors.New("The current player has neither busted nor decided to stand.")
}

func (m *MoveAdvanceNextPlayer) Apply(state boardgame.State) error {
	s := state.(*mainState)

	s.Game.CurrentPlayer++
	if s.Game.CurrentPlayer >= len(s.Players) {
		s.Game.CurrentPlayer = 0
	}

	currentPlayer := s.Players[s.Game.CurrentPlayer]

	currentPlayer.Stood = false

	return nil

}

func (m *MoveAdvanceNextPlayer) Copy() boardgame.Move {
	var result MoveAdvanceNextPlayer
	result = *m
	return &result
}

func (m *MoveAdvanceNextPlayer) DefaultsForState(state boardgame.State) {
	//TODO: implement
}

func (m *MoveAdvanceNextPlayer) Name() string {
	return "Advance Next Player"
}

func (m *MoveAdvanceNextPlayer) Description() string {
	return "When the current player has either busted or decided to stand, we advance to next player."
}

func (m *MoveAdvanceNextPlayer) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}
