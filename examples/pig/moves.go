package pig

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/components/dice"
	"github.com/jkomoros/boardgame/moves"
)

//+autoreader
type MoveRollDice struct {
	moves.CurrentPlayer
}

//+autoreader
type MoveDoneTurn struct {
	moves.CurrentPlayer
}

//+autoreader
type MoveCountDie struct {
	moves.CurrentPlayer
}

/**************************************************
 *
 * MoveRollDice Implementation
 *
 **************************************************/

func (m *MoveRollDice) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return nil
	}

	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	if !p.DieCounted {
		return errors.New("Your most recent roll has not yet been counted")
	}

	return nil
}

func (m *MoveRollDice) Apply(state boardgame.MutableState) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	die := game.Die.ComponentAt(0)

	if err := die.DynamicValues(state).(*dice.DynamicValue).Roll(die); err != nil {
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

func (m *MoveDoneTurn) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	if !p.DieCounted {
		return errors.New("Your most recent roll has not yet been counted")
	}

	if p.Done {
		return errors.New("You already signaled that you are done!")
	}

	return nil
}

func (m *MoveDoneTurn) Apply(state boardgame.MutableState) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	p.Done = true

	return nil
}

/**************************************************
 *
 * MoveCountDie Implementation
 *
 **************************************************/

func (m *MoveCountDie) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	if p.DieCounted {
		return errors.New("The most recent die roll has already been counted.")
	}

	return nil
}

func (m *MoveCountDie) Apply(state boardgame.MutableState) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	value := game.Die.ComponentAt(0).DynamicValues(state).(*dice.DynamicValue).Value

	if value == 1 {
		//Bust!
		p.Busted = true
	} else {
		p.RoundScore += value
	}

	p.DieCounted = true

	return nil
}
