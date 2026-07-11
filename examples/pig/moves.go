package pig

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/components/dice"
	"github.com/jkomoros/boardgame/legal"
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

// Legal() is deliberately absent: this move opted into declarative legality
// (design spec §8 survey, Task 12) via the moves.WithPreconditions call in
// main.go's ConfigureMoves. The original imperative body (kept only as
// legacyLegalMoveRollDice, a private copy in legal_golden_test.go, for
// golden-equivalence testing) read:
//
//	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
//		return nil
//	}
//	game, players := concreteStates(state)
//	p := players[game.CurrentPlayer.EnsureValid(state)]
//	if !p.DieCounted {
//		return errors.New("Your most recent roll has not yet been counted")
//	}
//	return nil
//
// Pre-existing bug fixed by this migration (flagged in the Task 12 report,
// not hidden): the super-call's error was discarded ("return nil" instead
// of "return err"), so a proposer who was NOT the current player always
// passed this check -- only the DieCounted gate below actually mattered.
// The migrated plan's contributed proposer atom (legal.ProposerIsCurrentPlayer,
// added base-first by moves.CurrentPlayer.ContributedPreconditions) has no
// such bug, so a wrong-proposer roll is now correctly rejected. This is a
// deliberate, documented behavior IMPROVEMENT that "declaring is
// implementing" sanctions (the frozen-chain guarantee only protects moves
// that do NOT opt in) -- see legal_golden_test.go's
// knownBugFixNilnessDivergence for the golden-equivalence handling.
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

// Legal() is deliberately absent: this move opted into declarative legality
// (design spec §8 survey, Task 12) PARTIALLY -- only the DieCounted gate has
// a catalog primitive (legal.PlayerBool); the "already done" check has none
// (v1's catalog has no negated-boolean predicate -- see LegalCustom below
// and the Task 12 report's design-feedback note), so it survives as
// imperative residue. The original imperative body (kept only as
// legacyLegalMoveDoneTurn, a private copy in legal_golden_test.go, for
// golden-equivalence testing) read:
//
//	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
//		return err
//	}
//	game, players := concreteStates(state)
//	p := players[game.CurrentPlayer.EnsureValid(state)]
//	if !p.DieCounted {
//		return errors.New("your most recent roll has not yet been counted")
//	}
//	if p.Done {
//		return errors.New("you already signaled that you are done")
//	}
//	return nil
func (m *moveDoneTurn) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	_, players := concreteStates(state)

	p := players[state.CurrentPlayerIndex().EnsureValid(state)]

	if p.Done {
		return legal.Errorf("pig.already_done", nil)
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

// moveCountDie stays fully imperative (spec §8 survey, Task 12): its ONLY
// gate is "the most recent die roll has already been counted" -- a NEGATED
// boolean (DieCounted must be false). legal.PlayerBool only expresses "prop
// is true"; v1's catalog has no negation wrapper (the only compositor is
// legal.Any, an OR), so there is no declarative spec this move could author
// at all. With zero natural WithPreconditions candidates there is nothing
// to opt in with (an empty authored list is treated as not-opted-in, so
// LegalCustom would never be consulted even if implemented -- see
// examples/memory/moves.go's moveStartHideCardsTimer for the same Task 11
// precedent). Left byte-for-byte unchanged from pre-Task-12.
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
		p.Eliminated = true
	} else {
		p.RoundScore += value
	}

	p.DieCounted = true

	return nil
}
