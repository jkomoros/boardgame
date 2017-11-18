package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/moveinterfaces"
)

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
