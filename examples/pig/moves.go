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

//+autoreader readsetter
type moveCountDie struct {
	boardgame.DefaultMove
	TargetPlayerIndex boardgame.PlayerIndex
}

//+autoreader readsetter
type moveAdvanceNextPlayer struct {
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

/**************************************************
 *
 * MoveCountDie Implementation
 *
 **************************************************/

func MoveCountDieFactory(state boardgame.State) boardgame.Move {
	result := &moveCountDie{
		boardgame.DefaultMove{
			"Count Die",
			"After a die has been rolled, tabulating its impact",
		},
		0,
	}

	if state != nil {
		result.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
	}

	return result
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

	value := game.Die.ComponentAt(0).DynamicValues(state).(*dieDynamicValue).Value

	if value == 1 {
		//Bust!
		p.Busted = true
	} else {
		p.RoundScore += value
	}

	return nil
}

/**************************************************
 *
 * MoveAdvanceNextPlayer Implementation
 *
 **************************************************/

func MoveAdvanceNextPlayerFactory(state boardgame.State) boardgame.Move {
	result := &moveAdvanceNextPlayer{
		boardgame.DefaultMove{
			"Advance Next Player",
			"Advance to the next player when the current player has busted or said they are done.",
		},
		0,
	}

	if state != nil {
		result.TargetPlayerIndex = state.CurrentPlayer().PlayerIndex()
	}

	return result
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

	p.TotalScore += p.RoundScore
	p.ResetForTurn()

	game.CurrentPlayer = game.CurrentPlayer.Next(state)

	p = players[game.CurrentPlayer]

	p.ResetForTurn()

	return nil
}
