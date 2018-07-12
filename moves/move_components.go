package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/moves/internal/privateconstants"
)

type sourceDestinationStacker interface {
	interfaces.SourceStacker
	interfaces.DestinationStacker
}

//MoveCountComponents is a move that will move components, one at a time, from
//SourceStack() to DestinationStack() until TargetCount() components have been
//moved. It is like DealComponents or CollectComponnets, except instead of
//working on a certain stack for each player, it operates on two fixed stacks.
//Other MoveComponents-style moves derive from this. When using it you must
//implement interfaces.SourceStacker and interfaces.DestinationStacker
//to encode which stacks to use. You may also want to override TargetCount()
//if you want to move more than one component.
//
//In practice it is most common to just use this move (and its subclasses)
//directly, and pass configuration for SourceStack, DestinationStack, and
//TargetCount with WithSourceStack, WithDestinationStack, and WithTargetCount
//to auto.Config.
//
//+autoreader
type MoveCountComponents struct {
	ApplyCountTimes
}

func (m *MoveCountComponents) ValidConfiguration(exampleState boardgame.State) error {
	if err := m.ApplyCountTimes.ValidConfiguration(exampleState); err != nil {
		return err
	}

	theSourceDestinationStacker, ok := m.TopLevelStruct().(sourceDestinationStacker)

	if !ok {
		return errors.New("EmbeddingMove doesn't have Source/Destination stacker.")
	}

	if theSourceDestinationStacker.DestinationStack(exampleState) == nil {
		return errors.New("DestinationStack returned nil")
	}

	if theSourceDestinationStacker.SourceStack(exampleState) == nil {
		return errors.New("SourceStack returned nil")
	}

	return nil
}

//SourceStack by default just returns the property on GameState with the name
//passed to DefaultConfig by WithSourceStack. If that is not sufficient,
//override this in your embedding struct.
func (m *MoveCountComponents) SourceStack(state boardgame.State) boardgame.Stack {
	config := m.CustomConfiguration()

	stackName, ok := config[privateconstants.SourceStack]

	if !ok {
		return nil
	}

	strStackName, ok := stackName.(string)

	if !ok {
		return nil
	}

	stack, err := state.GameState().ReadSetter().StackProp(strStackName)

	if err != nil {
		return nil
	}

	return stack
}

//DestinationStack by default just returns the property on GameState with the
//name passed to DefaultConfig by WithDestinationStack. If that is not sufficient,
//override this in your embedding struct.
func (m *MoveCountComponents) DestinationStack(state boardgame.State) boardgame.Stack {
	config := m.CustomConfiguration()

	stackName, ok := config[privateconstants.DestinationStack]

	if !ok {
		return nil
	}

	strStackName, ok := stackName.(string)

	if !ok {
		return nil
	}

	stack, err := state.GameState().ReadSetter().StackProp(strStackName)

	if err != nil {
		return nil
	}

	return stack
}

//stacks returns the source and desitnation so you don't have to do the cast.
func (m *MoveCountComponents) stacks(state boardgame.ImmutableState) (source, destination boardgame.Stack) {

	//TODO: this is a total hack
	mState := state.(boardgame.State)

	stacker, ok := m.TopLevelStruct().(sourceDestinationStacker)

	if !ok {
		return nil, nil
	}

	return stacker.SourceStack(mState), stacker.DestinationStack(mState)

}

func (m *MoveCountComponents) stackNames() (starter, destination string) {

	return stackName(m, privateconstants.SourceStack), stackName(m, privateconstants.DestinationStack)
}

//Apply by default moves one component from SourceStack() to
//DestinationStack(). You likely do not need to override this method.
func (m *MoveCountComponents) Apply(state boardgame.State) error {

	source, destination := m.stacks(state)

	if source == nil {
		return errors.New("Source was nil")
	}

	if destination == nil {
		return errors.New("Destination was nil")
	}

	return source.First().MoveToNextSlot(destination)

}

//FallbackName returns a string based on the names of the player
//stack name, game stack name, and target count.
func (m *MoveCountComponents) FallbackName(g *boardgame.GameManager) string {

	source, destination := m.stackNames()

	return "Move " + targetCountString(m.TopLevelStruct()) + " Components From " + source + " To " + destination
}

//FallbackHelpText returns a string based on the names of the player
//stack name, game stack name, and target count.
func (m *MoveCountComponents) FallbackHelpText() string {
	source, destination := m.stackNames()

	return "Moves " + targetCountString(m.TopLevelStruct()) + " components from " + source + " to " + destination
}

//MoveComponentsUntilCountReached is a move that will move components, one at
//a time, from SourceStack() to DestinationStack() until the target stack is
//up to having TargetCount components in it. See also
//MoveComponentsUntilCountLeft for a slightly different end condition.
//
//+autoreader
type MoveComponentsUntilCountReached struct {
	MoveCountComponents
}

//CountDown returns false, because as we move components from source to
//destination, destination will be getting larger.
func (m *MoveComponentsUntilCountReached) CountDown(state boardgame.ImmutableState) bool {
	return false
}

//Count returns the number of components in DestinationStack().
func (m *MoveComponentsUntilCountReached) Count(state boardgame.ImmutableState) int {

	_, targetStack := m.stacks(state)

	if targetStack == nil {
		return 0
	}

	return targetStack.NumComponents()
}

//FallbackName returns a string based on the names of the player
//stack name, game stack name, and target count.
func (m *MoveComponentsUntilCountReached) FallbackName(g *boardgame.GameManager) string {

	source, destination := m.stackNames()

	return "Move Components From " + source + " Until " + destination + " Has " + targetCountString(m.TopLevelStruct())
}

//FallbackHelpText returns a string based on the names of the player
//stack name, game stack name, and target count.
func (m *MoveComponentsUntilCountReached) FallbackHelpText() string {
	source, destination := m.stackNames()

	return "Moves components from " + source + " to " + destination + " until " + destination + " has " + targetCountString(m.TopLevelStruct())
}

//MoveComponentsUntilCountLeft is a move that will move components, one at a
//time, from SourceStack() to DestinationStack() until the source stack is
//down to having  TargetCount components in it. Its primary difference from
//MoveComponentsUntilCountReached is that its target is based on reducing the
//size of SourceStack to a target size.
//
//+autoreader
type MoveComponentsUntilCountLeft struct {
	MoveCountComponents
}

//Count returns the number of components in the SourceStack().
func (m *MoveComponentsUntilCountLeft) Count(state boardgame.ImmutableState) int {
	targetStack, _ := m.stacks(state)

	if targetStack == nil {
		return 0
	}

	return targetStack.NumComponents()
}

//CountDown returns true, because as we move components from source to
//destination, source will be getting smaller and smaller.
func (m *MoveComponentsUntilCountLeft) CountDown(state boardgame.ImmutableState) bool {
	return true
}

//FallbackName returns a string based on the names of the player
//stack name, game stack name, and target count.
func (m *MoveComponentsUntilCountLeft) FallbackName(g *boardgame.GameManager) string {

	source, destination := m.stackNames()

	return "Move Components To " + destination + " Until " + source + " Has " + targetCountString(m.TopLevelStruct())
}

//FallbackHelpText returns a string based on the names of the player
//stack name, game stack name, and target count.
func (m *MoveComponentsUntilCountLeft) FallbackHelpText() string {
	source, destination := m.stackNames()

	return "Moves components from " + source + " to " + destination + " until " + source + " has " + targetCountString(m.TopLevelStruct()) + " left"
}
