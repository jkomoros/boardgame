package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/moveinterfaces"
	"strconv"
)

//ApplyUntil is a simple move that is legal to apply in succession until its
//ConditionMet returns nil. You need to implement
//moveinterfaces.ConditionMetter by implementing a ConditionMet method.
type ApplyUntil struct {
	Base
}

//AllowMultipleInProgression returns true because the move is applied until
//ConditionMet returns nil.
func (a *ApplyUntil) AllowMultipleInProgression() bool {
	return true
}

func (a *ApplyUntil) ValidConfiguration(exampleState boardgame.MutableState) error {
	if _, ok := a.TopLevelStruct().(moveinterfaces.ConditionMetter); !ok {
		return errors.New("Embedding Move doesn't have ConditionMet")
	}
	return nil
}

//Legal returns an error until ConditionMet returns nil.
func (a *ApplyUntil) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	if err := a.Base.Legal(state, proposer); err != nil {
		return err
	}

	conditionMet, ok := a.TopLevelStruct().(moveinterfaces.ConditionMetter)

	if !ok {
		//This should be extremely rare since we ourselves have the right method.
		return errors.New("ApplyUntil top level struct unexpectedly did not have ConditionMet method")
	}

	if err := conditionMet.ConditionMet(state); err != nil {
		//The condition is not yet met, which means it's legal.
		return nil
	}

	return errors.New("The condition was met, so the move is no longer legal.")

}

func (a *ApplyUntil) MoveTypeName(manager *boardgame.GameManager) string {
	return "Apply Until"
}

func (a *ApplyUntil) MoveTypeHelpText(manager *boardgame.GameManager) string {
	return "Applies the move until a condition is met."
}

func (a *ApplyUntil) MoveTypeIsFixUp(manager *boardgame.GameManager) bool {
	return true
}

type counter interface {
	Count(state boardgame.State) int
	moveinterfaces.TargetCounter
	CountDown(state boardgame.State) bool
}

//ApplyUntilCount is a subclass of ApplyUntil that is legal until Count() is
//one past TargetCount()'s value (which direction "past" is determined by the
//result of CountDown()). At the minimum you'll want to provide your own
//Count() and Apply() methods, or use the moves that subclass from this, like
//MoveComponentsUntilCountReached.
type ApplyUntilCount struct {
	ApplyUntil
}

func (a *ApplyUntilCount) ValidConfiguration(exampleState boardgame.MutableState) error {
	if err := a.ApplyUntil.ValidConfiguration(exampleState); err != nil {
		return err
	}

	if _, ok := a.TopLevelStruct().(counter); !ok {
		return errors.New("EmeddingMove doesn't have Count/TargetCount")
	}

	return nil
}

//Count is consulted in ConditionMet to see what the current count is. Simply
//returns 1 by default. You almost certainly want to override this.
func (a *ApplyUntilCount) Count(state boardgame.State) int {
	return 1
}

//TargetCount should return the count that you want to target. Note that it's
//also important to override CountDown() if you're counting down, not up. By
//default returns 1.
func (a *ApplyUntilCount) TargetCount() int {
	return 1
}

//CountDown should return true if we're counting downward, or false if we're
//counting up. ConditionMet() needs to know if we're counting down or we're
//counting up because it can't tell that by itself, and needs to stop one
//after the target is reached. Defaults to false.
func (a *ApplyUntilCount) CountDown(state boardgame.State) bool {
	return false
}

//ConditionMet returns nil once TargetCount() is one past Count() (which
//direction is picked based on CountDown()). In general you override Count()
//and TargetCount() to customize behavior instead of overriding this.
func (a *ApplyUntilCount) ConditionMet(state boardgame.State) error {

	embeddingMove := a.TopLevelStruct()

	moveCounter, ok := embeddingMove.(counter)

	if !ok {
		return errors.New("Embedding move unexpectedly did not have Count/TargetCount")
	}

	count := moveCounter.Count(state)
	targetCount := moveCounter.TargetCount()
	countDown := moveCounter.CountDown(state)

	if targetCount == count {
		return errors.New("Count is equal to TargetCount. This will be our last move")
	}

	if countDown {
		if count > targetCount {
			return errors.New("Counting down, and still greater than TargetCount")
		}
	} else {
		if count < targetCount {
			return errors.New("Counting up, and still less than target count")
		}
	}

	return nil

}

func (a *ApplyUntilCount) targetCountString() string {
	moveCounter, ok := a.TopLevelStruct().(counter)

	if !ok {
		return "unknown"
	}

	targetCount := moveCounter.TargetCount()

	return strconv.Itoa(targetCount)

}

func (a *ApplyUntilCount) MoveTypeName(manager *boardgame.GameManager) string {

	return "Apply Until Count of " + a.targetCountString()
}

func (a *ApplyUntilCount) MoveTypeHelpText(manager *boardgame.GameManager) string {
	return "Applies the move until a target count of " + a.targetCountString() + " is met."
}

//countMovesApplied is where the majority of logic for the count method of
//ApplyUntilCount goes. It makes it easy to plug in the logic in multiple
//types of moves that have the same type of behavior for Count() but can't
//directly subclass one another.
func countMovesApplied(topLevelStruct boardgame.Move, state boardgame.State) int {

	records := state.Game().MoveRecords(state.Version())

	if len(records) == 0 {
		return 0
	}

	targetName := topLevelStruct.Info().Type().Name()
	targetPhase := state.Game().Manager().Delegate().CurrentPhase(state)

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

//ApplyCountTimes subclasses ApplyUntilCount. It applies the move until
//TargetCount() number of this move have been applied in a row within the
//current phase. Override TargetCount() to return the number of moves you
//actually want to apply. You'll need to provide your own Apply() method.
type ApplyCountTimes struct {
	ApplyUntilCount
}

//Count returns the number of times this move has been applied in a row in the
//immediate past in the current phase.
func (a *ApplyCountTimes) Count(state boardgame.State) int {
	return countMovesApplied(a.TopLevelStruct(), state)
}

func (a *ApplyCountTimes) MoveTypeName(manager *boardgame.GameManager) string {

	return "Apply " + a.targetCountString() + " Times"
}

func (a *ApplyCountTimes) MoveTypeHelpText(manager *boardgame.GameManager) string {
	return "Applies the move " + a.targetCountString() + " times in a row."
}
