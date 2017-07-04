package tictactoe

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

//TODO: test this!!

//+autoreader readsetter
type MovePlaceToken struct {
	boardgame.BaseMove
	//Which token to place the token
	Slot int
	//Which player we THINK is making the move.
	TargetPlayerIndex boardgame.PlayerIndex
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

	m.TargetPlayerIndex = game.CurrentPlayer

	//Default to setting a slot that's empty.
	for i, token := range game.Slots.ComponentValues() {
		if token == nil {
			m.Slot = i
			break
		}
	}
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

//+autoreader readsetter
type MoveAdvancePlayer struct {
	boardgame.BaseMove
}

var moveAdvancePlayerConfig = boardgame.MoveTypeConfig{
	Name:     "Advance Player",
	HelpText: "After the current player has made all of their moves, this fix-up move advances to the next player.",
	MoveConstructor: func() boardgame.Move {
		return new(MoveAdvancePlayer)
	},
	IsFixUp: true,
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
