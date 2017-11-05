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
conditions, override RoundRobinFinished() on your move.

For example, if you want to deal two cards to each player, set your
RoundRobinFinished like so:

	func (m *MyMove) RoundRobinFinished(state boardgame.State) error {
		return m.RoundRobinFinishedMultiCircuit(2, state)
	}

If you wanted to draw cards to players until each player had two cards, but
players might start with different number of cards, you'd configure it like
so:

	func (m *MyMove) RoundRobinFinished(state boardgame.State) error {
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
conditions, override RoundRobinFinished() on your move.

For example, if you want to collect two cards from each player, set your
RoundRobinFinished like so:

	func (m *MyMove) RoundRobinFinished(state boardgame.State) error {
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
