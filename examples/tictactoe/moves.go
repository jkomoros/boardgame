package tictactoe

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

//TODO: test this!!

//+autoreader readsetter
type MovePlaceToken struct {
	moves.CurrentPlayer
	//Which token to place the token
	Slot int
}

var movePlayTokenConfig = boardgame.MoveTypeConfig{
	Name:     "Place Token",
	HelpText: "Place a player's token in a specific space.",
	MoveConstructor: func() boardgame.Move {
		return new(MovePlaceToken)
	},
}

func (m *MovePlaceToken) DefaultsForState(state boardgame.State) {
	game, _ := concreteStates(state)

	m.CurrentPlayer.DefaultsForState(state)

	//Default to setting a slot that's empty.
	for i, token := range game.Slots.ComponentValues() {
		if token == nil {
			m.Slot = i
			break
		}
	}
}

func (m *MovePlaceToken) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

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

//+autoreader readsetter
type MoveFinishTurn struct {
	moves.FinishTurn
}

var moveFinishTurnConfig = boardgame.MoveTypeConfig{
	Name:     "Finish Turn",
	HelpText: "After the current player has made all of their moves, this fix-up move advances to the next player.",
	MoveConstructor: func() boardgame.Move {
		return new(MoveFinishTurn)
	},
	IsFixUp: true,
}
