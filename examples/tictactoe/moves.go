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

func (m *MovePlaceToken) Legal(payload boardgame.StatePayload) error {
	p := payload.(*statePayload)

	if p.game.CurrentPlayer != m.TargetPlayerIndex {
		return errors.New("The specified player is not the current player.")
	}

	if p.users[m.TargetPlayerIndex].UnusedTokens.Len() < 1 {
		return errors.New("There aren't any remaining tokens for the current player to place.")
	}

	if p.game.Slots.ComponentAt(m.Slot) != nil {
		return errors.New("The specified slot is already taken.")
	}

	return nil

}

func (m *MovePlaceToken) Apply(payload boardgame.StatePayload) boardgame.StatePayload {

	result := payload.Copy()

	p := result.(*statePayload)

	u := p.users[m.TargetPlayerIndex]

	c := u.UnusedTokens.RemoveFirst()

	p.game.Slots.InsertAtSlot(c, m.Slot)

	u.TokensToPlaceThisTurn--

	return result
}

func (m *MovePlaceToken) DefaultsForState(state boardgame.StatePayload) {
	s := state.(*statePayload)

	m.TargetPlayerIndex = s.game.CurrentPlayer

	//Default to setting a slot that's empty.
	for i, token := range s.game.Slots.ComponentValues() {
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

func (m *MovePlaceToken) Props() []string {
	return boardgame.PropertyReaderPropsImpl(m)
}

func (m *MovePlaceToken) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(m, name)
}

func (m *MovePlaceToken) SetProp(name string, val interface{}) error {
	return boardgame.PropertySetImpl(m, name, val)
}

func (m *MovePlaceToken) JSON() boardgame.JSONObject {
	return m
}

type MoveAdvancePlayer struct{}

func (m *MoveAdvancePlayer) Legal(payload boardgame.StatePayload) error {
	p := payload.(*statePayload)

	user := p.users[p.game.CurrentPlayer]

	if user.TokensToPlaceThisTurn > 0 {
		return errors.New("The current player still has tokens left to place this turn.")
	}

	return nil
}

func (m *MoveAdvancePlayer) Apply(payload boardgame.StatePayload) boardgame.StatePayload {
	result := payload.Copy()

	p := result.(*statePayload)

	p.game.CurrentPlayer++

	if p.game.CurrentPlayer >= len(p.users) {
		p.game.CurrentPlayer = 0
	}

	newUser := p.users[p.game.CurrentPlayer]

	newUser.TokensToPlaceThisTurn = 1

	return result

}

func (m *MoveAdvancePlayer) DefaultsForState(state boardgame.StatePayload) {
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

func (m *MoveAdvancePlayer) JSON() boardgame.JSONObject {
	return m
}

func (m *MoveAdvancePlayer) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(m, name)
}

func (m *MoveAdvancePlayer) Props() []string {
	return boardgame.PropertyReaderPropsImpl(m)
}

func (m *MoveAdvancePlayer) SetProp(name string, val interface{}) error {
	return boardgame.PropertySetImpl(m, name, val)
}
