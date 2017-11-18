package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/moveinterfaces"
)

type sourceDestinationStacker interface {
	moveinterfaces.SourceStacker
	moveinterfaces.DestinationStacker
}

//MoveComponentsUntilCountReached is a move that will move components, one at
//a time, from SourceStack() to DestinationStack() until the target stack is
//up to having TargetCount components in it. Other MoveComponents-style moves
//derive from this. When using it you must implement
//moveinterfaces.SourceStacker and moveinterfaces.DestinationStacker to encode
//which stacks to use. You may also want to override TargetCount() if you want
//to move more than one component.
type MoveComponentsUntilCountReached struct {
	ApplyUntilCount
}

func (m *MoveComponentsUntilCountReached) ValidConfiguration(exampleState boardgame.MutableState) error {
	if err := m.ApplyUntilCount.ValidConfiguration(exampleState); err != nil {
		return err
	}

	if _, ok := m.TopLevelStruct().(sourceDestinationStacker); !ok {
		return errors.New("EmbeddingMove doesn't have Source/Destination stacker.")
	}

	return nil
}

//CountDown returns false, because as we move components from source to
//destination, destination will be getting larger.
func (m *MoveComponentsUntilCountReached) CountDown(state boardgame.State) bool {
	return false
}

//stacks returns the source and desitnation so you don't have to do the cast.
func (m *MoveComponentsUntilCountReached) stacks(state boardgame.State) (source, destination boardgame.MutableStack) {

	//TODO: this is a total hack
	mState := state.(boardgame.MutableState)

	stacker, ok := m.TopLevelStruct().(sourceDestinationStacker)

	if !ok {
		return nil, nil
	}

	return stacker.SourceStack(mState), stacker.DestinationStack(mState)

}

//Count returns the number of components in DestinationStack().
func (m *MoveComponentsUntilCountReached) Count(state boardgame.State) int {

	_, targetStack := m.stacks(state)

	if targetStack == nil {
		return 0
	}

	return targetStack.NumComponents()
}

//Apply by default moves one component from SourceStack() to
//DestinationStack(). You likely do not need to override this method.
func (m *MoveComponentsUntilCountReached) Apply(state boardgame.MutableState) error {

	source, destination := m.stacks(state)

	if source == nil {
		return errors.New("Source was nil")
	}

	if destination == nil {
		return errors.New("Destination was nil")
	}

	return source.MoveComponent(boardgame.FirstComponentIndex, destination, boardgame.NextSlotIndex)

}

//MoveComponentsUntilCountLeft is a move that will move components, one at a
//time, from SourceStack() to DestinationStack() until the source stack is
//down to having  TargetCount components in it. It subclasses
//MoveComponentsUntilCountReached, and its primary difference is that its
//target is based on reducing the size of SourceStack to a target size.
type MoveComponentsUntilCountLeft struct {
	MoveComponentsUntilCountReached
}

//Count returns the number of components in the SourceStack().
func (m *MoveComponentsUntilCountLeft) Count(state boardgame.State) int {
	targetStack, _ := m.stacks(state)

	if targetStack == nil {
		return 0
	}

	return targetStack.NumComponents()
}

//CountDown returns true, because as we move components from source to
//destination, source will be getting smaller and smaller.
func (m *MoveComponentsUntilCountLeft) CountDown(state boardgame.State) bool {
	return true
}

//MoveComponents is a move that will move components, one at a time, from
//SourceStack() to DestinationStack() until TargetCount() components have been
//moved. It is like DealComponents or CollectComponnets, except instead of
//working on a certain stack for each player, it operates on two fixed stacks.
type MoveCountComponents struct {
	MoveComponentsUntilCountReached
}

//Count counts the number of times this move has been applied in a row.
func (m *MoveCountComponents) Count(state boardgame.State) int {
	return countMovesApplied(m.TopLevelStruct(), state)
}
