package moves

import (
	"errors"
	"strconv"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

// targetCountString is a simple helper that returns the string of the target count.
func targetCountString(concreteMove boardgame.Move) string {
	moveCounter, ok := concreteMove.(interfaces.TargetCounter)

	if !ok {
		return "unknown"
	}

	//Technically it's possible that the embedding move could need to do
	//something with the state, but we don't have a reference to one so :shrug:
	targetCount := moveCounter.TargetCount(nil)

	return strconv.Itoa(targetCount)

}

// ApplyUntil is a simple move that is legal to apply in succession until its
// ConditionMet returns nil. You need to implement
// interfaces.ConditionMetter by implementing a ConditionMet method.
//
//boardgame:codegen
type ApplyUntil struct {
	FixUpMulti
}

// ValidConfiguration verifies the Move this is embedded in implements
// interfaces.ConditionMetter.
func (a *ApplyUntil) ValidConfiguration(exampleState boardgame.State) error {

	if _, ok := a.Info().ConcreteMove().(interfaces.ConditionMetter); !ok {
		return errors.New("Embedding Move doesn't have ConditionMet")
	}

	return a.FixUpMulti.ValidConfiguration(exampleState)
}

// Legal returns an error until ConditionMet returns nil.
func (a *ApplyUntil) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	return a.legalWithBase(state, proposer, a.Default.Legal)
}

func (a *ApplyUntil) legalWithBase(state boardgame.ImmutableState, proposer boardgame.PlayerIndex, baseLegal func(boardgame.ImmutableState, boardgame.PlayerIndex) error) error {
	if err := baseLegal(state, proposer); err != nil {
		return err
	}

	conditionMet, ok := a.Info().ConcreteMove().(interfaces.ConditionMetter)

	if !ok {
		//This should be extremely rare since we ourselves have the right method.
		return errors.New("ApplyUntil concrete move unexpectedly did not have ConditionMet method")
	}

	if err := conditionMet.ConditionMet(state); err != nil {
		//The condition is not yet met, which means it's legal.
		return nil
	}

	return errors.New("the condition was met, so the move is no longer legal")
}

// FallbackName simply returns "Apply Until"
func (a *ApplyUntil) FallbackName(m *boardgame.GameManager) string {
	return "Apply Until"
}

// FallbackHelpText simply returns "Applies the move until a condition is met."
func (a *ApplyUntil) FallbackHelpText() string {
	return "Applies the move until a condition is met."
}

type counter interface {
	Count(state boardgame.ImmutableState) int
	interfaces.TargetCounter
}

// ApplyUntilCount is a subclass of ApplyUntil that is legal until Count() is
// equal to TargetCount(). (This presumes that each time the move is applied it
// gets the TargetCount one closer to Count and never overshoots). At the
// minimum you'll want to provide your own Count() and Apply() methods, or use
// the moves that subclass from this, like MoveComponentsUntilCountReached.
//
//boardgame:codegen
type ApplyUntilCount struct {
	ApplyUntil
}

// ValidConfiguration verifies the concrete move implements Count() and
// interfaces.TargetCounter, and that TargetCount doesn't return below 0.
func (a *ApplyUntilCount) ValidConfiguration(exampleState boardgame.State) error {
	if err := a.ApplyUntil.ValidConfiguration(exampleState); err != nil {
		return err
	}

	theCounter, ok := a.Info().ConcreteMove().(counter)

	if !ok {
		return errors.New("EmeddingMove doesn't have Count/TargetCount")
	}

	if theCounter.Count(exampleState) < 0 {
		return errors.New("Count returned a value below 0, which signals an error")
	}

	if theCounter.TargetCount(exampleState) < 0 {
		return errors.New("TargetCount returned a value below 0, which signals an error")
	}

	return nil
}

// Count is consulted in ConditionMet to see what the current count is. Simply
// returns 1 by default. You almost certainly want to override this.
func (a *ApplyUntilCount) Count(state boardgame.ImmutableState) int {
	return 1
}

// TargetCount should return the count that you want to target. Will return the
// configuration option passed via WithTargetCount in DefaultConfig, or 1 if
// that wasn't provided.
func (a *ApplyUntilCount) TargetCount(state boardgame.ImmutableState) int {

	config := a.CustomConfiguration()

	val, ok := config[configPropTargetCount]

	if !ok {
		//No configuration provided, just return default
		return 1
	}

	intVal, ok := val.(int)

	if !ok {
		//signal error
		return -1
	}

	return intVal

}

// ConditionMet returns nil once Count() is equal to TargetCount(). Note this
// presumes that repeated applciations of this move move Count one closer to
// TargetCount, and that it never overshoots, otherwise this could never
// terminate. In general you override Count() and TargetCount() to customize
// behavior instead of overriding this.
func (a *ApplyUntilCount) ConditionMet(state boardgame.ImmutableState) error {

	embeddingMove := a.Info().ConcreteMove()

	moveCounter, ok := embeddingMove.(counter)

	if !ok {
		return errors.New("Embedding move unexpectedly did not have Count/TargetCount")
	}

	count := moveCounter.Count(state)
	targetCount := moveCounter.TargetCount(state)

	if targetCount == count {
		//We're at the goal!
		return nil
	}

	return errors.New("Not yet at the count goal (" + strconv.Itoa(targetCount) + " / " + strconv.Itoa(count) + ")")

}

// FallbackName returns "Apply Until Count of INT", where INT is the
// target count.
func (a *ApplyUntilCount) FallbackName(m *boardgame.GameManager) string {

	return "Apply Until Count of " + targetCountString(a.Info().ConcreteMove())
}

// FallbackHelpText returns "Applies the move until a target count of
// INT is met.", where INT is the target count.
func (a *ApplyUntilCount) FallbackHelpText() string {
	return "Applies the move until a target count of " + targetCountString(a.Info().ConcreteMove()) + " is met."
}

// countMovesApplied is where the majority of logic for the count method of
// ApplyUntilCount goes. It makes it easy to plug in the logic in multiple
// types of moves that have the same type of behavior for Count() but can't
// directly subclass one another.
func countMovesApplied(concreteMove boardgame.Move, state boardgame.ImmutableState) int {

	records := state.Game().MoveRecords(state.Version())

	if len(records) == 0 {
		return 0
	}

	targetName := concreteMove.Info().Name()
	_, targetPhase := currentPhaseInfo(state)

	count := 0

	for i := len(records) - 1; i >= 0; i-- {
		record := records[i]

		if record.Phase != targetPhase {
			break
		}

		if record.Name != targetName {
			break
		}

		count++
	}

	return count
}

// ApplyCountTimes subclasses ApplyUntilCount. It applies the move until
// TargetCount() number of this move have been applied in a row within the
// current phase. Override TargetCount() to return the number of moves you
// actually want to apply. You'll need to provide your own Apply() method.
//
//boardgame:codegen
type ApplyCountTimes struct {
	ApplyUntilCount
}

// Count returns the number of times this move has been applied in a row in the
// immediate past in the current phase.
func (a *ApplyCountTimes) Count(state boardgame.ImmutableState) int {
	return countMovesApplied(a.Info().ConcreteMove(), state)
}

// FallbackName returns "Apply INT Times", where INT is the target
// count.
func (a *ApplyCountTimes) FallbackName(m *boardgame.GameManager) string {

	return "Apply " + targetCountString(a.Info().ConcreteMove()) + " Times"
}

// FallbackHelpText returns "Applies the move INT times in a row.",
// where INT is the target count.
func (a *ApplyCountTimes) FallbackHelpText() string {
	return "Applies the move " + targetCountString(a.Info().ConcreteMove()) + " times in a row."
}
