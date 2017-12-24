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

//MoveCountComponents is a move that will move components, one at a time, from
//SourceStack() to DestinationStack() until TargetCount() components have been
//moved. It is like DealComponents or CollectComponnets, except instead of
//working on a certain stack for each player, it operates on two fixed stacks.
//Other MoveComponents-style moves derive from this. When using it you must
//implement moveinterfaces.SourceStacker and moveinterfaces.DestinationStacker
//to encode which stacks to use. You may also want to override TargetCount()
//if you want to move more than one component.
type MoveCountComponents struct {
	ApplyCountTimes
}

func (m *MoveCountComponents) ValidConfiguration(exampleState boardgame.MutableState) error {
	if err := m.ApplyUntilCount.ValidConfiguration(exampleState); err != nil {
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
func (m *MoveCountComponents) SourceStack(state boardgame.MutableState) boardgame.MutableStack {
	config := m.Info().Type().CustomConfiguration()

	stackName, ok := config[configNameSourceStack]

	if !ok {
		return nil
	}

	strStackName, ok := stackName.(string)

	if !ok {
		return nil
	}

	stack, err := state.MutableGameState().ReadSetter().MutableStackProp(strStackName)

	if err != nil {
		return nil
	}

	return stack
}

//DestinationStack by default just returns the property on GameState with the
//name passed to DefaultConfig by WithDestinationStack. If that is not sufficient,
//override this in your embedding struct.
func (m *MoveCountComponents) DestinationStack(state boardgame.MutableState) boardgame.MutableStack {
	config := m.Info().Type().CustomConfiguration()

	stackName, ok := config[configNameDestinationStack]

	if !ok {
		return nil
	}

	strStackName, ok := stackName.(string)

	if !ok {
		return nil
	}

	stack, err := state.MutableGameState().ReadSetter().MutableStackProp(strStackName)

	if err != nil {
		return nil
	}

	return stack
}

//stacks returns the source and desitnation so you don't have to do the cast.
func (m *MoveCountComponents) stacks(state boardgame.State) (source, destination boardgame.MutableStack) {

	//TODO: this is a total hack
	mState := state.(boardgame.MutableState)

	stacker, ok := m.TopLevelStruct().(sourceDestinationStacker)

	if !ok {
		return nil, nil
	}

	return stacker.SourceStack(mState), stacker.DestinationStack(mState)

}

func (m *MoveCountComponents) stackNames(state boardgame.MutableState) (starter, destination string) {
	starterStack, destinationStack := m.stacks(state)

	starter = "unknown stack"
	destination = "unknown stack"

	if name := stackPropName(starterStack, state); name != "" {
		starter = name
	}

	if name := stackPropName(destinationStack, state); name != "" {
		destination = name
	}

	return starter, destination
}

//Apply by default moves one component from SourceStack() to
//DestinationStack(). You likely do not need to override this method.
func (m *MoveCountComponents) Apply(state boardgame.MutableState) error {

	source, destination := m.stacks(state)

	if source == nil {
		return errors.New("Source was nil")
	}

	if destination == nil {
		return errors.New("Destination was nil")
	}

	return source.MoveComponent(boardgame.FirstComponentIndex, destination, boardgame.NextSlotIndex)

}

func (m *MoveCountComponents) MoveTypeFallbackName(manager *boardgame.GameManager) string {

	source, destination := m.stackNames(manager.ExampleState())

	return "Move " + targetCountString(m.TopLevelStruct()) + " Components From " + source + " To " + destination
}

func (m *MoveCountComponents) MoveTypeFallbackHelpText(manager *boardgame.GameManager) string {
	source, destination := m.stackNames(manager.ExampleState())

	return "Moves " + targetCountString(m.TopLevelStruct()) + " components from " + source + " to " + destination
}

//MoveComponentsUntilCountReached is a move that will move components, one at
//a time, from SourceStack() to DestinationStack() until the target stack is
//up to having TargetCount components in it. See also
//MoveComponentsUntilCountLeft for a slightly different end condition.
type MoveComponentsUntilCountReached struct {
	MoveCountComponents
}

//CountDown returns false, because as we move components from source to
//destination, destination will be getting larger.
func (m *MoveComponentsUntilCountReached) CountDown(state boardgame.State) bool {
	return false
}

//Count returns the number of components in DestinationStack().
func (m *MoveComponentsUntilCountReached) Count(state boardgame.State) int {

	_, targetStack := m.stacks(state)

	if targetStack == nil {
		return 0
	}

	return targetStack.NumComponents()
}

func (m *MoveComponentsUntilCountReached) MoveTypeFallbackName(manager *boardgame.GameManager) string {

	source, destination := m.stackNames(manager.ExampleState())

	return "Move Components From " + source + " Until " + destination + " Has " + targetCountString(m.TopLevelStruct())
}

func (m *MoveComponentsUntilCountReached) MoveTypeFallbackHelpText(manager *boardgame.GameManager) string {
	source, destination := m.stackNames(manager.ExampleState())

	return "Moves components from " + source + " to " + destination + " until " + destination + " has " + targetCountString(m.TopLevelStruct())
}

//MoveComponentsUntilCountLeft is a move that will move components, one at a
//time, from SourceStack() to DestinationStack() until the source stack is
//down to having  TargetCount components in it. Its primary difference from
//MoveComponentsUntilCountReached is that its target is based on reducing the
//size of SourceStack to a target size.
type MoveComponentsUntilCountLeft struct {
	MoveCountComponents
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

func (m *MoveComponentsUntilCountLeft) MoveTypeFallbackName(manager *boardgame.GameManager) string {

	source, destination := m.stackNames(manager.ExampleState())

	return "Move Components To " + destination + " Until " + source + " Has " + targetCountString(m.TopLevelStruct())
}

func (m *MoveComponentsUntilCountLeft) MoveTypeFallbackHelpText(manager *boardgame.GameManager) string {
	source, destination := m.stackNames(manager.ExampleState())

	return "Moves components from " + source + " to " + destination + " until " + source + " has " + targetCountString(m.TopLevelStruct()) + " left"
}
