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
	TargetCount(state boardgame.State) int
	CountDown(state boardgame.State) bool
}

type targetSourceSize interface {
	TargetSourceSize() bool
}

//ApplyUntilCount is a subclass of ApplyUntil that is legal until Count() is
//one past TargetCount()'s value (which direction "past" is determined by the
//result of CountDown()). By default it moves a component from SourceStack()
//to DestinationStack(). If you use this move type directly (as opposed to in
//other moves in this package that embed it anonymously), you generally want
//to override SourceStack(), DestinationStack(), and possibly
//TargetSourceSize() and TargetCount() and leave all other methods untouched.
//The default methods in this move mean that it is effectively equivalent to
//MoveComponentsUntilNumReached, but generally you should use that move in
//your code for clarity when you just want the default behavior.
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

//TargetSourceSize should return whether Count() and TargetCount() are based
//on increasing destination's size to target (default), or declining source's
//size to target. This is used primarily to help the default Count(),
//TargetCount() do the right thing without being overriden. Defaults to false,
//which denotes that the target we're trying to hit is based on destination's
//size.
func (a *ApplyUntilCount) TargetSourceSize() bool {
	return false
}

//targetSourceSizeImpl is a convenience method that does the interface cast to
//get TargetSourceSize.z
func (a *ApplyUntilCount) targetSourceSizeImpl() bool {
	targetSourcer, ok := a.TopLevelStruct().(targetSourceSize)

	if !ok {
		return false
	}
	return targetSourcer.TargetSourceSize()
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
	if _, ok := a.TopLevelStruct().(targetSourceSize); !ok {
		return errors.New("EmbeddingMove doesn't have TargetSourceSize")
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
//default it's the destination Stack's NumComponents, but if
//TargetSourceSize() returns true, it will instead be the destination stack's
//size. Generally you don't override this directly and instead override
//TargetSourceSize().
func (a *ApplyUntilCount) Count(state boardgame.State) int {

	var targetStack boardgame.MutableStack

	if a.targetSourceSizeImpl() {
		targetStack, _ = a.stacks(state)
	} else {
		_, targetStack = a.stacks(state)
	}

	if targetStack == nil {
		return 0
	}

	return targetStack.NumComponents()
}

//TargetCount should return the count that you want to target. Note that it's
//also important to override CountDown() if we
func (a *ApplyUntilCount) TargetCount(state boardgame.State) int {
	return 1
}

//CountDown should return true if we're counting downward, or false (or remain
//unimplmented) if we're counting up. ConditionMet() needs to know if we're
//counting down or we're counting up because it can't tell that by itself, and
//needs to stop one after the target is reached. The default CountDown()
//returns the result of TargetSourceSize().
func (a *ApplyUntilCount) CountDown(state boardgame.State) bool {
	return a.targetSourceSizeImpl()
}

//TargetCountAndDirection should return the count that we want to apply moves
//until Count equals, plus if we're counting down or up. After Count has been
//met, Legal() will start failing. countDown should be if we're counting down
//or up from Count() to TargetCount(), as the move itself can't detect that,
//and the end condition is either Count() is 1 greater than TargetCount() or 1
//less than TargetCount(). The default implementation uses TargetCount() for
//the number. For the countDown, if TargetSourceSize() is false (default), we
//return (1, false); if it is true, we return (1, true) (since moving items
//from Destination to Stack will decline Source's size). Generally you don't
//override this directly and instead override TargetSourceSize() and/or
//TargetCount().

//Apply by default moves one component from SourceStack() to
//DestinationStack(). If you want different behavior, you should override this
//--but then will also want to override Count() and TargetCount() as well.
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
//you override Count() and TargetCount() to customize behavior instead of
//overriding this.
func (a *ApplyUntilCount) ConditionMet(state boardgame.State) error {

	embeddingMove := a.TopLevelStruct()

	moveCounter, ok := embeddingMove.(counter)

	if !ok {
		return errors.New("Embedding move unexpectedly did not have Count/TargetCount")
	}

	count := moveCounter.Count(state)
	targetCount := moveCounter.TargetCount(state)
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

//ApplyNTimes subclasses ApplyUntilCount. It applies the move until
//TargetCount() number of this move have been applied in a row within the
//current phase. Override TargetCount() to return the number of moves you
//actually want to apply.
type ApplyNTimes struct {
	ApplyUntilCount
}

//TargetCount by default returns 1. Override it if you want to apply more
//moves.
func (a *ApplyNTimes) TargetCount(state boardgame.State) (count int) {
	return 1
}

//Count returns the number of times this move has been applied in a row in the
//immediate past in the current phase.
func (a *ApplyNTimes) Count(state boardgame.State) int {

	records := state.Game().MoveRecords(state.Version())

	if len(records) == 0 {
		return 0
	}

	targetName := a.TopLevelStruct().Info().Type().Name()
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
