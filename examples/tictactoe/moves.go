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

	first := players[m.TargetPlayerIndex.EnsureValid(state)].UnusedTokens.ImmutableFirst()
	if first == nil {
		return errors.New("there aren't any remaining tokens for the current player to place")
	}

	return first.MayMoveToSlot(game.Slots, m.Slot)

}

func (m *movePlaceToken) Apply(state boardgame.State) error {

	game, players := concreteStates(state)

	u := players[m.TargetPlayerIndex.EnsureValid(state)]

	if err := u.UnusedTokens.First().MoveTo(game.Slots, m.Slot); err != nil {
		return err
	}

	u.TokensToPlaceThisTurn--

	game.Phase.SetValue(phaseAfterFirstMove)

	return nil
}
