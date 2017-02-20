package blackjack

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

type MoveShuffleDiscardToDraw struct{}

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

func (m *MoveShuffleDiscardToDraw) Props() []string {
	return boardgame.PropertyReaderPropsImpl(m)
}

func (m *MoveShuffleDiscardToDraw) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(m, name)
}

func (m *MoveShuffleDiscardToDraw) SetProp(name string, val interface{}) error {
	return boardgame.PropertySetImpl(m, name, val)
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

func (m *MoveCurrentPlayerHit) Props() []string {
	return boardgame.PropertyReaderPropsImpl(m)
}

func (m *MoveCurrentPlayerHit) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(m, name)
}

func (m *MoveCurrentPlayerHit) SetProp(name string, val interface{}) error {
	return boardgame.PropertySetImpl(m, name, val)
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

func (m *MoveCurrentPlayerStand) Props() []string {
	return boardgame.PropertyReaderPropsImpl(m)
}

func (m *MoveCurrentPlayerStand) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(m, name)
}

func (m *MoveCurrentPlayerStand) SetProp(name string, val interface{}) error {
	return boardgame.PropertySetImpl(m, name, val)
}

/*

//For ease of copying for the next move to create. :-)

type MoveShuffleDiscardToDraw struct{}

func (m *MoveShuffleDiscardToDraw) Legal(state boardgame.State) error {
	return nil
}

func (m *MoveShuffleDiscardToDraw) Apply(state boardgame.State) error {
	return nil
}

func (m *MoveShuffleDiscardToDraw) Copy() boardgame.Move {
	var result MoveShuffleDiscardToDraw
	result = *m
	return &result
}

func (m *MoveShuffleDiscardToDraw) DefaultsForState(state boardgame.State) {
	//TODO: implement
}

func (m *MoveShuffleDiscardToDraw) Name() string {
	return "Shuffle Discard To Draw"
}

func (m *MoveShuffleDiscardToDraw) Description() string {
	return "When the draw deck is empty, shuffles the discard deck into draw deck."
}

func (m *MoveShuffleDiscardToDraw) Props() []string {
	return boardgame.PropertyReaderPropsImpl(m)
}

func (m *MoveShuffleDiscardToDraw) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(m, name)
}

func (m *MoveShuffleDiscardToDraw) SetProp(name string, val interface{}) error {
	return boardgame.PropertySetImpl(m, name, val)
}

*/
