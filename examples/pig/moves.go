package pig

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/components/dice"
	"github.com/jkomoros/boardgame/moves"
)

//boardgame:codegen
type moveRollDice struct {
	moves.CurrentPlayer
}

//boardgame:codegen
type moveDoneTurn struct {
	moves.CurrentPlayer
}

//boardgame:codegen
type moveCountDie struct {
	moves.CurrentPlayer
}

/**************************************************
 *
 * moveRollDice Implementation
 *
 **************************************************/

func (m *moveRollDice) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return nil
	}

	game, players := concreteStates(state)

	p := players[game.CurrentPlayer.EnsureValid(state)]

	if !p.DieCounted {
		return errors.New("Your most recent roll has not yet been counted")
	}

	return nil
}

func (m *moveRollDice) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer.EnsureValid(state)]

	die := game.Die.ComponentAt(0)

	die.DynamicValues().(*dice.DynamicValue).Roll(state.Rand())

	p.DieCounted = false

	return nil
}

/**************************************************
 *
 * moveDoneTurn Implementation
 *
 **************************************************/

func (m *moveDoneTurn) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	p := players[game.CurrentPlayer.EnsureValid(state)]

	if !p.DieCounted {
		return errors.New("your most recent roll has not yet been counted")
	}

	if p.Done {
		return errors.New("you already signaled that you are done")
	}

	return nil
}

func (m *moveDoneTurn) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer.EnsureValid(state)]

	p.Done = true

	return nil
}

/**************************************************
 *
 * moveCountDie Implementation
 *
 **************************************************/

func (m *moveCountDie) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game, players := concreteStates(state)

	p := players[game.CurrentPlayer.EnsureValid(state)]

	if p.DieCounted {
		return errors.New("the most recent die roll has already been counted")
	}

	return nil
}

func (m *moveCountDie) Apply(state boardgame.State) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer.EnsureValid(state)]

	value := game.Die.ComponentAt(0).DynamicValues().(*dice.DynamicValue).Value

	if value == 1 {
		//Bust!
		p.Busted = true
	} else {
		p.RoundScore += value
	}

	p.DieCounted = true

	return nil
}
