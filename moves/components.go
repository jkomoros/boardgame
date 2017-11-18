package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/moveinterfaces"
)

/*

DealComponents is a type of RoundRobin move that deals components from
gameState's GameStack() to each PlayerState's PlayerStack(). By default it
goes around once and deals a single component. If you want different end
conditions, override ConditionMet() on your move.

For example, if you want to deal two cards to each player, set your
ConditionMet like so:

	func (m *MyMove) ConditionMet(state boardgame.State) error {
		return m.RoundRobinFinishedMultiCircuit(2, state)
	}

If you wanted to draw cards to players until each player had two cards, but
players might start with different number of cards, you'd configure it like
so:

	func (m *MyMove) ConditionMet(state boardgame.State) error {
		//Configure that the finished function should be when all players have
		//their conditions met.
		return m.RoundRobinFinishedPlayerConditionsMet(state)
	}

	func (m *MyMove) RoundRobinPlayerConditionMet(playerState boardgame.PlayerState) bool {
		//Configure that the player condition is met when the PlayerStack is size 2
		return m.RoundRobinPlayerConditionStackTargetSizeMet(2, playerState)
	}

	func (m *MyMove) PlayerStack(playerState boardgame.MutablePlayerState) boardgame.MutableStack {
		//Configure the stack whose size we want to be 2 is the player's hand
		return playerState.(*playerState).Hand
	}

*/
type DealComponents struct {
	RoundRobin
}

func (d *DealComponents) ValidConfiguration(exampleState boardgame.MutableState) error {
	if _, ok := d.TopLevelStruct().(moveinterfaces.PlayerStacker); !ok {
		return errors.New("Embedding move doesn't implement PlayerStacker")
	}

	if _, ok := d.TopLevelStruct().(moveinterfaces.GameStacker); !ok {
		return errors.New("Embedding move doesn't implement GameStacker")
	}

	return d.RoundRobin.ValidConfiguration(exampleState)
}

//RoundRobinAction moves a component from the GameStack to the PlayerStack, as
//configured by the PlayerStacker and GameStacke interfaces.
func (d *DealComponents) RoundRobinAction(playerState boardgame.MutablePlayerState) error {

	playerStacker, ok := d.TopLevelStruct().(moveinterfaces.PlayerStacker)

	if !ok {
		return errors.New("Embedding move unexpectedly doesn't implement PlayerStacker")
	}

	targetStack := playerStacker.PlayerStack(playerState)

	if targetStack == nil {
		return errors.New("PlayerStacker didn't return a valid stack")
	}

	gameStacker, ok := d.TopLevelStruct().(moveinterfaces.GameStacker)

	if !ok {
		return errors.New("Embedding move unexpectedly doesn't implement GameStacker")
	}

	sourceStack := gameStacker.GameStack(playerState.MutableState().MutableGameState())

	if sourceStack == nil {
		return errors.New("GameStacker didn't return a valid stack")
	}

	return sourceStack.MoveComponent(boardgame.FirstComponentIndex, targetStack, boardgame.NextSlotIndex)

}

/*

CollectComponents is a type of RoundRobin move that collects components from
each PlayerState's PlayerStack() to gameState's GameStack(). By default it
goes around once and collects a component from each. If you want different end
conditions, override ConditionMet() on your move.

For example, if you want to collect two cards from each player, set your
ConditionMet like so:

	func (m *MyMove) ConditionMet(state boardgame.State) error {
		return m.RoundRobinFinishedMultiCircuit(2, state)
	}

*/
type CollectComponents struct {
	RoundRobin
}

func (d *CollectComponents) ValidConfiguration(exampleState boardgame.MutableState) error {
	if _, ok := d.TopLevelStruct().(moveinterfaces.PlayerStacker); !ok {
		return errors.New("Embedding move doesn't implement PlayerStacker")
	}

	if _, ok := d.TopLevelStruct().(moveinterfaces.GameStacker); !ok {
		return errors.New("Embedding move doesn't implement GameStacker")
	}

	return d.RoundRobin.ValidConfiguration(exampleState)
}

//RoundRobinAction moves a component from the PlayerStack to the GameStack, as
//configured by the PlayerStacker and GameStacke interfaces.
func (d *CollectComponents) RoundRobinAction(playerState boardgame.MutablePlayerState) error {

	playerStacker, ok := d.TopLevelStruct().(moveinterfaces.PlayerStacker)

	if !ok {
		return errors.New("Embedding move unexpectedly doesn't implement PlayerStacker")
	}

	playerStack := playerStacker.PlayerStack(playerState)

	if playerStack == nil {
		return errors.New("PlayerStacker didn't return a valid stack")
	}

	gameStacker, ok := d.TopLevelStruct().(moveinterfaces.GameStacker)

	if !ok {
		return errors.New("Embedding move unexpectedly doesn't implement GameStacker")
	}

	targetStack := gameStacker.GameStack(playerState.MutableState().MutableGameState())

	if targetStack == nil {
		return errors.New("GameStacker didn't return a valid stack")
	}

	return playerStack.MoveComponent(boardgame.FirstComponentIndex, targetStack, boardgame.NextSlotIndex)

}

type sourceDestinationStacker interface {
	moveinterfaces.SourceStacker
	moveinterfaces.DestinationStacker
}

//MoveComponentsUntilCountReached is a move that will move components, one at
//a time, from SourceStack() to DestinationStack() until the target stack is
//up to having TargetCount components in it. Other MoveComponents-style moves
//derive from this. When using it you most likely want to override
//SourceStack(), DestinationStack(), and possibly TargetCount() if you want to
//move more than one. MoveComponentsUntilCountLeft and MoveCountComponents
//both subclass this move.
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

//SourceStack is by default called in Apply() to get the stack to move from.
//The default simply returns nil; if you want to have ApplyUntilCount do its
//default move-a-component action, override this.
func (m *MoveComponentsUntilCountReached) SourceStack(state boardgame.MutableState) boardgame.MutableStack {
	return nil
}

//DesitnationStack is by default called in Count(), TargetCount(), and
//Apply(). The default simply returns nil; if you want to have ApplyUntilCount
//do its default move-a-component action, override this.
func (m *MoveComponentsUntilCountReached) DestinationStack(state boardgame.MutableState) boardgame.MutableStack {
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
