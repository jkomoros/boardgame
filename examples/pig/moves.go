package pig

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

//+autoreader readsetter
type moveRollDice struct {
	boardgame.DefaultMove
	TargetPlayerIndex boardgame.PlayerIndex
}

//+autoreader readsetter
type moveDoneTurn struct {
	boardgame.DefaultMove
	TargetPlayerIndex boardgame.PlayerIndex
}

/**************************************************
 *
 * MoveRollDice Implementation
 *
 **************************************************/

func MoveRollDiceFactory(state boardgame.State) boardgame.Move {
	result := &moveRollDice{
		boardgame.DefaultMove{
			"Roll Dice",
			"Rolls the dice for the current player",
		},
		0,
	}

	if state != nil {
		result.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
	}

	return result
}

func (m *moveRollDice) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	game, players := concreteStates(state)

	if !proposer.Equivalent(game.CurrentPlayer) {
		return errors.New("You are not the current player!")
	}

	if !m.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("You are not the current player!")
	}

	p := players[game.CurrentPlayer]

	if !p.DieCounted {
		return errors.New("Your most recent roll has not yet been counted")
	}

	return nil
}

func (m *moveRollDice) Apply(state boardgame.MutableState) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	die := game.Die.ComponentAt(0)

	if err := die.DynamicValues(state).(*dieDynamicValue).Roll(die); err != nil {
		return errors.New("Couldn't roll die: " + err.Error())
	}

	p.DieCounted = false

	return nil
}

/**************************************************
 *
 * MoveDoneTurn Implementation
 *
 **************************************************/

func MoveDoneTurnFactory(state boardgame.State) boardgame.Move {
	result := &moveDoneTurn{
		boardgame.DefaultMove{
			"Done Turn",
			"Played when a player is done with their turn and wants to keep their score.",
		},
		0,
	}

	if state != nil {
		result.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
	}

	return result
}

func (m *moveDoneTurn) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	game, players := concreteStates(state)

	if !proposer.Equivalent(game.CurrentPlayer) {
		return errors.New("You are not the current player!")
	}

	if !m.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("You are not the current player!")
	}

	p := players[game.CurrentPlayer]

	if !p.DieCounted {
		return errors.New("Your most recent roll has not yet been counted")
	}

	if p.Done {
		return errors.New("You already signaled that you are done!")
	}

	return nil
}

func (m *moveDoneTurn) Apply(state boardgame.MutableState) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	p.Done = true

	return nil
}
