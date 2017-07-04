package pig

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/components/dice"
)

//+autoreader readsetter
type moveRollDice struct {
	boardgame.BaseMove
	TargetPlayerIndex boardgame.PlayerIndex
}

//+autoreader readsetter
type moveDoneTurn struct {
	boardgame.BaseMove
	TargetPlayerIndex boardgame.PlayerIndex
}

//+autoreader readsetter
type moveCountDie struct {
	boardgame.BaseMove
	TargetPlayerIndex boardgame.PlayerIndex
}

//+autoreader readsetter
type moveAdvanceNextPlayer struct {
	boardgame.BaseMove
	TargetPlayerIndex boardgame.PlayerIndex
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

func (m *moveRollDice) DefaultsForState(state boardgame.State) {
	m.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
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

func (m *moveDoneTurn) DefaultsForState(state boardgame.State) {
	m.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
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

func (m *moveCountDie) DefaultsForState(state boardgame.State) {
	m.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
}

func (m *moveCountDie) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	game, players := concreteStates(state)

	if !proposer.Equivalent(game.CurrentPlayer) {
		return errors.New("You are not the current player!")
	}

	if !m.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("You are not the current player!")
	}

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

var moveAdvanceNextPlayerConfig = boardgame.MoveTypeConfig{
	Name:     "Advance Next Player",
	HelpText: "Advance to the next player when the current player has busted or said they are done.",
	MoveConstructor: func() boardgame.Move {
		return new(moveAdvanceNextPlayer)
	},
	IsFixUp: true,
}

func (m *moveAdvanceNextPlayer) DefaultsForState(state boardgame.State) {
	m.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
}

func (m *moveAdvanceNextPlayer) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	game, players := concreteStates(state)

	if !proposer.Equivalent(game.CurrentPlayer) {
		return errors.New("You are not the current player!")
	}

	if !m.TargetPlayerIndex.Equivalent(proposer) {
		return errors.New("You are not the current player!")
	}

	p := players[game.CurrentPlayer]

	if !p.DieCounted {
		return errors.New("The most recent die roll has not been counted!")
	}

	if !p.Busted && !p.Done {
		return errors.New("The player has not either busted or signaled that they are done.")
	}

	return nil
}

func (m *moveAdvanceNextPlayer) Apply(state boardgame.MutableState) error {
	game, players := concreteStates(state)

	p := players[game.CurrentPlayer]

	if p.Done {
		p.TotalScore += p.RoundScore
	}
	p.ResetForTurn()

	game.CurrentPlayer = game.CurrentPlayer.Next(state)

	p = players[game.CurrentPlayer]

	p.ResetForTurn()

	return nil
}
