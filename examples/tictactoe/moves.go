package tictactoe

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

//TODO: test this!!

type MovePlaceToken struct {
	//Which token to place the token
	Slot int
	//Which player we THINK is making the move.
	TargetPlayerIndex int
}

func (m *MovePlaceToken) Legal(payload boardgame.State) error {
	p := payload.(*mainState)

	if p.Game.CurrentPlayer != m.TargetPlayerIndex {
		return errors.New("The specified player is not the current player.")
	}

	if p.Players[m.TargetPlayerIndex].UnusedTokens.Len() < 1 {
		return errors.New("There aren't any remaining tokens for the current player to place.")
	}

	if m.Slot < 0 || m.Slot >= p.Game.Slots.Len() {
		return errors.New("The specified slot is not legal.")
	}

	if p.Game.Slots.ComponentAt(m.Slot) != nil {
		return errors.New("The specified slot is already taken.")
	}

	return nil

}

func (m *MovePlaceToken) Apply(payload boardgame.State) error {

	p := payload.(*mainState)

	u := p.Players[m.TargetPlayerIndex]

	c := u.UnusedTokens.RemoveFirst()

	p.Game.Slots.InsertAtSlot(c, m.Slot)

	u.TokensToPlaceThisTurn--

	return nil
}

func (m *MovePlaceToken) DefaultsForState(state boardgame.State) {
	s := state.(*mainState)

	m.TargetPlayerIndex = s.Game.CurrentPlayer

	//Default to setting a slot that's empty.
	for i, token := range s.Game.Slots.ComponentValues() {
		if token == nil {
			m.Slot = i
			break
		}
	}

}

func (m *MovePlaceToken) Name() string {
	return "Place Token"
}

func (m *MovePlaceToken) Description() string {
	return "Place a player's token in a specific space."
}

func (m *MovePlaceToken) Copy() boardgame.Move {
	var result MovePlaceToken
	result = *m
	return &result
}

func (m *MovePlaceToken) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}

type MoveAdvancePlayer struct{}

func (m *MoveAdvancePlayer) Legal(payload boardgame.State) error {
	p := payload.(*mainState)

	user := p.Players[p.Game.CurrentPlayer]

	if user.TokensToPlaceThisTurn > 0 {
		return errors.New("The current player still has tokens left to place this turn.")
	}

	return nil
}

func (m *MoveAdvancePlayer) Apply(payload boardgame.State) error {

	p := payload.(*mainState)

	p.Game.CurrentPlayer++

	if p.Game.CurrentPlayer >= len(p.Players) {
		p.Game.CurrentPlayer = 0
	}

	newUser := p.Players[p.Game.CurrentPlayer]

	newUser.TokensToPlaceThisTurn = 1

	return nil

}

func (m *MoveAdvancePlayer) DefaultsForState(state boardgame.State) {
	//Nothing to set.
}

func (m *MoveAdvancePlayer) Name() string {
	//TODO: these should be package constants
	return "Advance Player"
}

func (m *MoveAdvancePlayer) Description() string {
	return "After the current player has made all of their moves, this fix-up move advances to the next player."
}

func (m *MoveAdvancePlayer) Copy() boardgame.Move {
	var result MoveAdvancePlayer
	result = *m
	return &result
}

func (m *MoveAdvancePlayer) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(m)
}
