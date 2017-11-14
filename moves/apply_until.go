package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/moveinterfaces"
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

type sourceDestinationStacker interface {
	moveinterfaces.SourceStacker
	moveinterfaces.DestinationStacker
}

type counter interface {
	Count(state boardgame.State) int
	TargetCount(state boardgame.State) (int, bool)
}

//ApplyUntilCount is a subclass of ApplyUntil that is legal until Count() is
//one past TargetCount()'s value. By default it moves a component from
//SourceStack() to DestinationStack().
type ApplyUntilCount struct {
	ApplyUntil
}

//SourceStack is by default called in Apply() to get the stack to move from.
//The default simply returns nil; if you want to have ApplyUntilCount do its
//default move-a-component action, override this.
func (a *ApplyUntilCount) SourceStack(state boardgame.MutableState) boardgame.MutableStack {
	return nil
}

//DesitnationStack is by default called in Count(), TargetCount(), and
//Apply(). The default simply returns nil; if you want to have ApplyUntilCount
//do its default move-a-component action, override this.
func (a *ApplyUntilCount) DestinationStack(state boardgame.MutableState) boardgame.MutableStack {
	return nil
}

func (a *ApplyUntilCount) ValidConfiguration(exampleState boardgame.MutableState) error {
	if err := a.ApplyUntil.ValidConfiguration(exampleState); err != nil {
		return err
	}
	if _, ok := a.TopLevelStruct().(sourceDestinationStacker); !ok {
		return errors.New("EmbeddingMove doesn't have Source/Destination stacker.")
	}
	if _, ok := a.TopLevelStruct().(counter); !ok {
		return errors.New("EmeddingMove doesn't have Count/TargetCount")
	}
	return nil
}

//stacks returns the source and desitnation so you don't have to do the cast.
func (a *ApplyUntilCount) stacks(state boardgame.State) (source, destination boardgame.MutableStack) {

	//TODO: this is a total hack
	mState := state.(boardgame.MutableState)

	stacker, ok := a.TopLevelStruct().(sourceDestinationStacker)

	if !ok {
		return nil, nil
	}

	return stacker.SourceStack(mState), stacker.DestinationStack(mState)

}

//Count is consulted in ConditionMet to see what the current count is. By
//default it's the destination Stack's NumComponents.
func (a *ApplyUntilCount) Count(state boardgame.State) int {
	_, destination := a.stacks(state)

	if destination == nil {
		return 0
	}

	return destination.NumComponents()
}

//TargetCount should return the count that we want to apply moves until Count
//equals. After Count has been met, Legal() will start failing. countDown
//should return if we're counting down or up from Count() to TargetCount(), as
//the move itself can't detect that, and the end condition is either Count()
//is 1 greater than TargetCount() or 1 less than TargetCount().
func (a *ApplyUntilCount) TargetCount(state boardgame.State) (count int, countDown bool) {
	return 1, true
}

//Apply by default moves one component from SourceStack() to
//DestinationStack(). Override if you want different behavior.
func (a *ApplyUntilCount) Apply(state boardgame.MutableState) error {

	source, destination := a.stacks(state)

	if source == nil {
		return errors.New("Source was nil")
	}

	if destination == nil {
		return errors.New("Destination was nil")
	}

	return source.MoveComponent(boardgame.FirstComponentIndex, destination, boardgame.NextSlotIndex)

}

//ConditionMet returns nil once TargetCount() is one past Count(). In general
//you override Count() and TargetCount() to customize behavior.
func (a *ApplyUntilCount) ConditionMet(state boardgame.State) error {

	embeddingMove := a.TopLevelStruct()

	moveCounter, ok := embeddingMove.(counter)

	if !ok {
		return errors.New("Embedding move unexpectedly did not have Count/TargetCount")
	}

	count := moveCounter.Count(state)
	targetCount, countDown := moveCounter.TargetCount(state)

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
