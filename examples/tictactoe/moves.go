package tictactoe

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

//TODO: test this!!

//+autoreader readsetter
type MovePlaceToken struct {
	//Which token to place the token
	Slot int
	//Which player we THINK is making the move.
	TargetPlayerIndex boardgame.PlayerIndex
}

func MovePlaceTokenFactory(state boardgame.State) boardgame.Move {
	result := &MovePlaceToken{}

	if state != nil {
		game, _ := concreteStates(state)

		result.TargetPlayerIndex = game.CurrentPlayer

		//Default to setting a slot that's empty.
		for i, token := range game.Slots.ComponentValues() {
			if token == nil {
				result.Slot = i
				break
			}
		}
	}

	return result
}

func (m *MovePlaceToken) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	game, players := concreteStates(state)

	if game.CurrentPlayer != m.TargetPlayerIndex {
		return errors.New("The specified player is not the current player.")
	}

	if !m.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("The proposing player is not the target player.")
	}

	if players[m.TargetPlayerIndex].UnusedTokens.Len() < 1 {
		return errors.New("There aren't any remaining tokens for the current player to place.")
	}

	if m.Slot < 0 || m.Slot >= game.Slots.Len() {
		return errors.New("The specified slot is not legal.")
	}

	if game.Slots.ComponentAt(m.Slot) != nil {
		return errors.New("The specified slot is already taken.")
	}

	return nil

}

func (m *MovePlaceToken) Apply(state boardgame.MutableState) error {

	game, players := concreteStates(state)

	u := players[m.TargetPlayerIndex]

	if err := u.UnusedTokens.MoveComponent(boardgame.FirstComponentIndex, game.Slots, m.Slot); err != nil {
		return err
	}

	u.TokensToPlaceThisTurn--

	return nil
}

func (m *MovePlaceToken) Name() string {
	return "Place Token"
}

func (m *MovePlaceToken) HelpText() string {
	return "Place a player's token in a specific space."
}

func (t *MovePlaceToken) ImmediateFixUp(state boardgame.State) boardgame.Move {
	return nil
}

//+autoreader readsetter
type MoveAdvancePlayer struct{}

func MoveAdvancePlayerFactory(state boardgame.State) boardgame.Move {
	return &MoveAdvancePlayer{}
}

func (m *MoveAdvancePlayer) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	game, players := concreteStates(state)

	user := players[game.CurrentPlayer]

	if user.TokensToPlaceThisTurn > 0 {
		return errors.New("The current player still has tokens left to place this turn.")
	}

	return nil
}

func (m *MoveAdvancePlayer) Apply(state boardgame.MutableState) error {

	game, players := concreteStates(state)

	game.CurrentPlayer = game.CurrentPlayer.Next(state)

	newUser := players[game.CurrentPlayer]

	newUser.TokensToPlaceThisTurn = 1

	return nil

}

func (m *MoveAdvancePlayer) Name() string {
	//TODO: these should be package constants
	return "Advance Player"
}

func (m *MoveAdvancePlayer) HelpText() string {
	return "After the current player has made all of their moves, this fix-up move advances to the next player."
}

func (t *MoveAdvancePlayer) ImmediateFixUp(state boardgame.State) boardgame.Move {
	return nil
}
