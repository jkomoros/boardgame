package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

type sourceDestinationStacker interface {
	interfaces.SourceStacker
	interfaces.DestinationStacker
}

// MoveCountComponents is a move that will move components, one at a time, from
// SourceStack() to DestinationStack() until TargetCount() components have been
// moved. It is like DealComponents or CollectComponnets, except instead of
// working on a certain stack for each player, it operates on two fixed stacks.
// Other MoveComponents-style moves derive from this. When using it you must
// implement interfaces.SourceStacker and interfaces.DestinationStacker
// to encode which stacks to use. You may also want to override TargetCount()
// if you want to move more than one component.
//
// In practice it is most common to just use this move (and its subclasses)
// directly, and pass configuration for SourceStack, DestinationStack, and
// TargetCount with WithSourceProperty, WithDestinationProperty, and
// WithTargetCount to auto.Config.
//
//boardgame:codegen
type MoveCountComponents struct {
	ApplyCountTimes
}

// ValidConfiguration checks to make sure that SourceStack and DestinationStack
// both exist and return non-nil stacks.
func (m *MoveCountComponents) ValidConfiguration(exampleState boardgame.State) error {
	if err := m.ApplyCountTimes.ValidConfiguration(exampleState); err != nil {
		return err
	}

	theSourceDestinationStacker, ok := m.Info().ConcreteMove().(sourceDestinationStacker)

	if !ok {
		return errors.New("embeddingMove doesn't have Source/Destination stacker")
	}

	if theSourceDestinationStacker.DestinationStack(exampleState) == nil {
		return errors.New("DestinationStack returned nil")
	}

	if theSourceDestinationStacker.SourceStack(exampleState) == nil {
		return errors.New("SourceStack returned nil")
	}

	return nil
}

// SourceStack by default just returns the property on GameState with the name
// passed to DefaultConfig by WithSourceProperty. If that is not sufficient,
// override this in your embedding struct.
func (m *MoveCountComponents) SourceStack(state boardgame.State) boardgame.Stack {
	config := m.CustomConfiguration()

	stackName, ok := config[configPropSourceProperty]

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

// DestinationStack by default just returns the property on GameState with the
// name passed to DefaultConfig by WithDestinationProperty. If that is not sufficient,
// override this in your embedding struct.
func (m *MoveCountComponents) DestinationStack(state boardgame.State) boardgame.Stack {
	config := m.CustomConfiguration()

	stackName, ok := config[configPropDestinationProperty]

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

// stacks returns the source and desitnation so you don't have to do the cast.
func (m *MoveCountComponents) stacks(state boardgame.ImmutableState) (source, destination boardgame.Stack) {

	//TODO: this is a total hack
	mState := state.(boardgame.State)

	stacker, ok := m.Info().ConcreteMove().(sourceDestinationStacker)

	if !ok {
		return nil, nil
	}

	return stacker.SourceStack(mState), stacker.DestinationStack(mState)

}

func (m *MoveCountComponents) stackNames(state boardgame.ImmutableState) (starter, destination string) {

	var sourceStack boardgame.ImmutableStack
	var destinationStack boardgame.ImmutableStack

	if state != nil {
		sourceStack, destinationStack = m.stacks(state)
	}

	return stackName(m, configPropSourceProperty, sourceStack, state), stackName(m, configPropDestinationProperty, destinationStack, state)
}

// Legal checks that source and destiantion stacks exist, that enough components
// to move exist.
func (m *MoveCountComponents) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.ApplyCountTimes.Legal(state, proposer); err != nil {
		return err
	}

	source, destination := m.stacks(state)

	if source == nil {
		return errors.New("Source was nil")
	}

	if destination == nil {
		return errors.New("Destination was nil")
	}

	moveCounter, ok := m.Info().ConcreteMove().(counter)
	if !ok {
		return errors.New("concrete move unexpectedly did not implement Count/TargetCount")
	}
	count := moveCounter.Count(state)
	target := moveCounter.TargetCount(state)
	if count < 0 {
		return errors.New("Count returned a negative value")
	}
	if target < 0 {
		return errors.New("TargetCount returned a negative value")
	}
	remaining := target - count
	if remaining < 0 {
		remaining = -remaining
	}

	return source.MayMoveCountTo(destination, remaining)

}

// Apply by default moves one component from SourceStack() to
// DestinationStack(). You likely do not need to override this method.
func (m *MoveCountComponents) Apply(state boardgame.State) error {

	source, destination := m.stacks(state)

	if source == nil {
		return errors.New("Source was nil")
	}

	if destination == nil {
		return errors.New("Destination was nil")
	}

	return source.MoveCountTo(destination, 1)

}

// FallbackName returns a string based on the names of the player
// stack name, game stack name, and target count.
func (m *MoveCountComponents) FallbackName(g *boardgame.GameManager) string {

	source, destination := m.stackNames(g.ExampleState())

	return "Move " + targetCountString(m.Info().ConcreteMove()) + " Components From " + source + " To " + destination
}

// FallbackHelpText returns a string based on the names of the player
// stack name, game stack name, and target count.
func (m *MoveCountComponents) FallbackHelpText() string {
	source, destination := m.stackNames(nil)

	return "Moves " + targetCountString(m.Info().ConcreteMove()) + " components from " + source + " to " + destination
}

// MoveComponentsUntilCountReached is a move that will move components, one at
// a time, from SourceStack() to DestinationStack() until the target stack is
// up to having TargetCount components in it. See also
// MoveComponentsUntilCountLeft for a slightly different end condition.
//
//boardgame:codegen
type MoveComponentsUntilCountReached struct {
	MoveCountComponents
}

// ConditionMet returns nil once DestinationStack has at least TargetCount
// components. Using a threshold rather than exact equality makes an already
// overfilled destination a completed operation instead of an endless sequence
// that moves farther away from its goal.
func (m *MoveComponentsUntilCountReached) ConditionMet(state boardgame.ImmutableState) error {
	moveCounter, ok := m.Info().ConcreteMove().(counter)
	if !ok {
		return errors.New("concrete move unexpectedly did not implement Count/TargetCount")
	}
	count, target := moveCounter.Count(state), moveCounter.TargetCount(state)
	if count < 0 || target < 0 {
		return errors.New("count or target count is invalid")
	}
	if count >= target {
		return nil
	}
	return errors.New("destination has not reached its target count")
}

// Count returns the number of components in DestinationStack().
func (m *MoveComponentsUntilCountReached) Count(state boardgame.ImmutableState) int {

	_, targetStack := m.stacks(state)

	if targetStack == nil {
		return 0
	}

	return targetStack.NumComponents()
}

// FallbackName returns a string based on the names of the player
// stack name, game stack name, and target count.
func (m *MoveComponentsUntilCountReached) FallbackName(g *boardgame.GameManager) string {

	source, destination := m.stackNames(g.ExampleState())

	return "Move Components From " + source + " Until " + destination + " Has " + targetCountString(m.Info().ConcreteMove())
}

// FallbackHelpText returns a string based on the names of the player
// stack name, game stack name, and target count.
func (m *MoveComponentsUntilCountReached) FallbackHelpText() string {
	source, destination := m.stackNames(nil)

	return "Moves components from " + source + " to " + destination + " until " + destination + " has " + targetCountString(m.Info().ConcreteMove())
}

// MoveComponentsUntilCountLeft is a move that will move components, one at a
// time, from SourceStack() to DestinationStack() until the source stack is
// down to having  TargetCount components in it. Its primary difference from
// MoveComponentsUntilCountReached is that its target is based on reducing the
// size of SourceStack to a target size.
//
//boardgame:codegen
type MoveComponentsUntilCountLeft struct {
	MoveCountComponents
}

// ConditionMet returns nil once SourceStack has at most TargetCount
// components. Using a threshold rather than exact equality makes an already
// undersized source a completed operation instead of moving farther away from
// its goal.
func (m *MoveComponentsUntilCountLeft) ConditionMet(state boardgame.ImmutableState) error {
	moveCounter, ok := m.Info().ConcreteMove().(counter)
	if !ok {
		return errors.New("concrete move unexpectedly did not implement Count/TargetCount")
	}
	count, target := moveCounter.Count(state), moveCounter.TargetCount(state)
	if count < 0 || target < 0 {
		return errors.New("count or target count is invalid")
	}
	if count <= target {
		return nil
	}
	return errors.New("source has not reached its target count")
}

// Count returns the number of components in the SourceStack().
func (m *MoveComponentsUntilCountLeft) Count(state boardgame.ImmutableState) int {
	targetStack, _ := m.stacks(state)

	if targetStack == nil {
		return 0
	}

	return targetStack.NumComponents()
}

// FallbackName returns a string based on the names of the player
// stack name, game stack name, and target count.
func (m *MoveComponentsUntilCountLeft) FallbackName(g *boardgame.GameManager) string {

	source, destination := m.stackNames(g.ExampleState())

	return "Move Components To " + destination + " Until " + source + " Has " + targetCountString(m.Info().ConcreteMove())
}

// FallbackHelpText returns a string based on the names of the player
// stack name, game stack name, and target count.
func (m *MoveComponentsUntilCountLeft) FallbackHelpText() string {
	source, destination := m.stackNames(nil)

	return "Moves components from " + source + " to " + destination + " until " + source + " has " + targetCountString(m.Info().ConcreteMove()) + " left"
}

// MoveAllComponents is simply a MoveComponentsUntilCountLeft that overrides
// TargetCount() to return 0. It's effectively the equivalent of
// stack.MoveAllTo, just broken into individual moves. A simple convenience
// since that combination is common.
//
//boardgame:codegen
type MoveAllComponents struct {
	MoveComponentsUntilCountLeft
}

// TargetCount returns 0, no matter what was passed with WithTargetCount. This
// is the primary behavior of this move, compared to
// MoveComponentsUntilCountLeft.
func (m *MoveAllComponents) TargetCount(state boardgame.ImmutableState) int {
	return 0
}

// FallbackName returns "Move All Components From SOURCESTACKNAME To
// DESTINATIONSTACKNAME"
func (m *MoveAllComponents) FallbackName(g *boardgame.GameManager) string {

	source, destination := m.stackNames(g.ExampleState())

	return "Move All Components From " + source + " To " + destination
}

// FallbackHelpText returns "Moves all components from SOURCESTACKNAME to
// DESTINATIONSTACKNAME"
func (m *MoveAllComponents) FallbackHelpText() string {
	source, destination := m.stackNames(nil)

	return "Moves all components from " + source + " to " + destination
}
