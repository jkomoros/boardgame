package pig

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/components/dice"
	"github.com/jkomoros/boardgame/moves"
)

//+autoreader
type moveRollDice struct {
	moves.CurrentPlayer
}

//+autoreader
type moveDoneTurn struct {
	moves.CurrentPlayer
}

//+autoreader
type moveCountDie struct {
	moves.CurrentPlayer
}

//+autoreader
type moveFinishTurn struct {
	moves.FinishTurn
}

/**************************************************
 *
 * MoveRollDice Implementation
 *
 **************************************************/

var moveRollDiceConfig = boardgame.MoveTypeConfig{
	Name:     "Roll Dice",
	HelpText: "Rolls the dice for the current player",
	MoveConstructor: func() boardgame.Move {
		return new(moveRollDice)
	},
}

func (m *moveRollDice) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

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

func (m *moveRollDice) Apply(state boardgame.MutableState) error {
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

var moveDoneTurnConfig = boardgame.MoveTypeConfig{
	Name:     "Done Turn",
	HelpText: "Played when a player is done with their turn and wants to keep their score.",
	MoveConstructor: func() boardgame.Move {
		return new(moveDoneTurn)
	},
}

func (m *moveDoneTurn) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

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

func (m *moveDoneTurn) Apply(state boardgame.MutableState) error {
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

var moveCountDieConfig = boardgame.MoveTypeConfig{
	Name:     "Count Die",
	HelpText: "After a die has been rolled, tabulating its impact",
	MoveConstructor: func() boardgame.Move {
		return new(moveCountDie)
	},
	IsFixUp: true,
}

func (m *moveCountDie) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

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

func (m *moveCountDie) Apply(state boardgame.MutableState) error {
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

/**************************************************
 *
 * MoveAdvanceNextPlayer Implementation
 *
 **************************************************/

var moveFinishTurnConfig = boardgame.MoveTypeConfig{
	Name:     "Finish Turn",
	HelpText: "Advance to the next player when the current player has busted or said they are done.",
	MoveConstructor: func() boardgame.Move {
		return new(moveFinishTurn)
	},
	IsFixUp: true,
}
