package tictactoe

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

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
