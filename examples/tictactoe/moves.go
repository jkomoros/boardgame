package tictactoe

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

//TODO: test this!!

//boardgame:codegen
type movePlaceToken struct {
	moves.CurrentPlayer
	//Which token to place the token
	Slot int
}

func (m *movePlaceToken) DefaultsForState(state boardgame.ImmutableState) {
	game, _ := concreteStates(state)

	m.CurrentPlayer.DefaultsForState(state)

	//Default to setting a slot that's empty.
	for i, c := range game.Slots.Components() {
		if c == nil {
			m.Slot = i
			break
		}
	}
}

func (m *movePlaceToken) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	if players[m.TargetPlayerIndex].UnusedTokens.Len() < 1 {
		return errors.New("there aren't any remaining tokens for the current player to place")
	}

	if m.Slot < 0 || m.Slot >= game.Slots.Len() {
		return errors.New("the specified slot is not legal")
	}

	if game.Slots.ComponentAt(m.Slot) != nil {
		return errors.New("the specified slot is already taken")
	}

	return nil

}

func (m *movePlaceToken) Apply(state boardgame.State) error {

	game, players := concreteStates(state)

	u := players[m.TargetPlayerIndex]

	if err := u.UnusedTokens.First().MoveTo(game.Slots, m.Slot); err != nil {
		return err
	}

	u.TokensToPlaceThisTurn--

	game.Phase.SetValue(phaseAfterFirstMove)

	return nil
}
