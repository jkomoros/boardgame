package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

//ApplyUntil is a simple move that is legal to apply in succession until its
//ConditionMet returns nil. You need to override ConditionMet as well as
//provide an Apply method.
type ApplyUntil struct {
	Base
}

type conditionMetter interface {
	ConditionMet(state boardgame.State) error
}

//AllowMultipleInProgression returns true because the move is applied until
//ConditionMet returns nil.
func (a *ApplyUntil) AllowMultipleInProgression() bool {
	return true
}

//ConditionMet is called in ApplyUntil's Legal method. If the condition has
//been met, return nil. If it has not been met, return an error describing why
//it is not yet met. The default ConditionMet returns nil always; you almost
//certainly want to override it.
func (a *ApplyUntil) ConditionMet(state boardgame.State) error {
	return nil
}

func (a *ApplyUntil) ValidConfiguration(exampleState boardgame.MutableState) error {
	if _, ok := a.TopLevelStruct().(conditionMetter); !ok {
		return errors.New("Embedding Move doesn't have ConditionMet")
	}
	return nil
}

//Legal returns an error until ConditionMet returns nil.
func (a *ApplyUntil) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
	if err := a.Base.Legal(state, proposer); err != nil {
		return err
	}

	conditionMet, ok := a.TopLevelStruct().(conditionMetter)

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
