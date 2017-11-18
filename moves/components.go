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

type targetSourceSize interface {
	TargetSourceSize() bool
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
	if _, ok := m.TopLevelStruct().(targetSourceSize); !ok {
		return errors.New("EmbeddingMove doesn't have TargetSourceSize")
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

//TargetSourceSize should return whether Count() and TargetCount() are based
//on increasing destination's size to target (default), or declining source's
//size to target. This is used primarily to help the default Count(),
//TargetCount() do the right thing without being overriden. Defaults to false,
//which denotes that the target we're trying to hit is based on destination's
//size. If you want the opposite behavior, just use
//MoveComponentsUntilCountLeft, which basically just overrides this method,
//because your intent will be more clear in your move structures.
func (m *MoveComponentsUntilCountReached) TargetSourceSize() bool {
	return false
}

//targetSourceSizeImpl is a convenience method that does the interface cast to
//get TargetSourceSize.z
func (m *MoveComponentsUntilCountReached) targetSourceSizeImpl() bool {
	targetSourcer, ok := m.TopLevelStruct().(targetSourceSize)

	if !ok {
		return false
	}
	return targetSourcer.TargetSourceSize()
}

//CountDown by default returns TargetSourceSize(). That is, if you're moving
//from source to destination until a count is reached, it will return false,
//otherwise will return true. You normally don't need to override this and can
//instead override TargetSourceSize().
func (m *MoveComponentsUntilCountReached) CountDown(state boardgame.State) bool {
	return m.targetSourceSizeImpl()
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

//Count is consulted in ConditionMet to see what the current count is. By
//default it's the destination Stack's NumComponents, but if
//TargetSourceSize() returns true, it will instead be the destination stack's
//size. Generally you don't override this directly and instead override
//TargetSourceSize().
func (m *MoveComponentsUntilCountReached) Count(state boardgame.State) int {

	var targetStack boardgame.MutableStack

	if m.targetSourceSizeImpl() {
		targetStack, _ = m.stacks(state)
	} else {
		_, targetStack = m.stacks(state)
	}

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
//down to having  TargetCount components in it. It subclasses from
//MoveComponentsUntilCountReached, and its primary difference is a different
//return value for TargetSourceSize(). However, using this move class directly
//instead of override MoveComponentsUntilCountReached.TargetSourceSize() is
//recommended for clarity of intent in your codebase.
type MoveComponentsUntilCountLeft struct {
	MoveComponentsUntilCountReached
}

//TargetSourceSize returns true, denoting to ApplyUntilCount that we are
//counting down until our source meets TargetCount(). This is sufficient to
//change the behavior of CountDown() and Count() to the right behavior.
func (m *MoveComponentsUntilCountLeft) TargetSourceSize() bool {
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
