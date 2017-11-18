package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/moveinterfaces"
)

func dealActionHelper(topLevelStruct boardgame.Move, playerState boardgame.MutablePlayerState) (playerStack boardgame.MutableStack, gameStack boardgame.MutableStack, err error) {
	playerStacker, ok := topLevelStruct.(moveinterfaces.PlayerStacker)

	if !ok {
		return nil, nil, errors.New("Embedding move unexpectedly doesn't implement PlayerStacker")
	}

	targetStack := playerStacker.PlayerStack(playerState)

	if targetStack == nil {
		return nil, nil, errors.New("PlayerStacker didn't return a valid stack")
	}

	gameStacker, ok := topLevelStruct.(moveinterfaces.GameStacker)

	if !ok {
		return nil, nil, errors.New("Embedding move unexpectedly doesn't implement GameStacker")
	}

	sourceStack := gameStacker.GameStack(playerState.MutableState().MutableGameState())

	if sourceStack == nil {
		return nil, nil, errors.New("GameStacker didn't return a valid stack")
	}

	return targetStack, sourceStack, nil
}

/*

DealCountComponents is a type of RoundRobin move that deals components from
gameState's GameStack() to each PlayerState's PlayerStack(). It goes around
TargetCount() times. TargetCount() defaults to 1; override if you want to deal
out a different number of components.

*/
type DealCountComponents struct {
	RoundRobinNumRounds
}

//TargetCount by default returns 1. Override if you want to deal more
//components.
func (d *DealCountComponents) TargetCount() int {
	return 1
}

//NumRounds simply returns TargetCount. NumRounds is what RoundRobinNumRounds
//expects, but TargetCount() is the terminology used for all of the similar
//Deal/Collect/MoveComponents methods.
func (d *DealCountComponents) NumRounds() int {
	targetCounter, ok := d.TopLevelStruct().(moveinterfaces.TargetCounter)

	if !ok {
		return 1
	}

	return targetCounter.TargetCount()
}

func (d *DealCountComponents) ValidConfiguration(exampleState boardgame.MutableState) error {
	if _, ok := d.TopLevelStruct().(moveinterfaces.PlayerStacker); !ok {
		return errors.New("Embedding move doesn't implement PlayerStacker")
	}

	if _, ok := d.TopLevelStruct().(moveinterfaces.GameStacker); !ok {
		return errors.New("Embedding move doesn't implement GameStacker")
	}
	if _, ok := d.TopLevelStruct().(moveinterfaces.TargetCounter); !ok {
		return errors.New("Embedding move doesn't implement TargetCounter")
	}

	return d.RoundRobinNumRounds.ValidConfiguration(exampleState)
}

//RoundRobinAction moves a component from the GameStack to the PlayerStack, as
//configured by the PlayerStacker and GameStacker interfaces.
func (d *DealCountComponents) RoundRobinAction(playerState boardgame.MutablePlayerState) error {

	playerStack, gameStack, err := dealActionHelper(d.TopLevelStruct(), playerState)

	if err != nil {
		return err
	}

	return gameStack.MoveComponent(boardgame.FirstComponentIndex, playerStack, boardgame.NextSlotIndex)
}

//DealComponentsUntilPlayerCountReached goes around and deals components to
//each player until each player has TargetCount() or greater components in
//their PlayerStack().
type DealComponentsUntilPlayerCountReached struct {
	DealCountComponents
}

//PlayerConditionMet is true if the NumComponents in the given player's
//PlayerStack() is TargetCount or greater.
func (d *DealComponentsUntilPlayerCountReached) PlayerConditionMet(pState boardgame.PlayerState) bool {
	playerStacker, ok := d.TopLevelStruct().(moveinterfaces.PlayerStacker)

	if !ok {
		return false
	}

	//Ugly hack. :-/
	mutablePState := pState.(boardgame.MutablePlayerState)

	playerStack := playerStacker.PlayerStack(mutablePState)

	targetCounter, ok := d.TopLevelStruct().(moveinterfaces.TargetCounter)

	if !ok {
		return false
	}

	return playerStack.NumComponents() >= targetCounter.TargetCount()
}

//ConditionMet simply returns the RoundRobin.ConditionMet (throwing out the
//RoundCount alternate of ConditionMet we get via sub-classing), since our
//PlayerConditionMet handles the end condtion.
func (d *DealComponentsUntilPlayerCountReached) ConditionMet(state boardgame.State) error {
	return d.RoundRobin.ConditionMet(state)
}

//DealComponentsUntilGameCountLeft goes around and deals components to each
//player until the GameStack() has TargetCount() or fewer components left.
type DealComponentsUntilGameCountLeft struct {
	DealCountComponents
}

//ConditionMet returns nil if GameStack's NumComponents is TargetCount or
//less, and otherwise defaults to RoundRobin's ConditionMet.
func (d *DealComponentsUntilGameCountLeft) ConditionMet(state boardgame.State) error {
	gameStacker, ok := d.TopLevelStruct().(moveinterfaces.GameStacker)

	if !ok {
		return errors.New("Unexpectedly didn't implement gameStacker")
	}

	//Total hack :-/
	mutableState := state.(boardgame.MutableState)

	gameStack := gameStacker.GameStack(mutableState.MutableGameState())

	if gameStack == nil {
		return errors.New("GameStack gave a nil stack")
	}

	targetCounter, ok := d.TopLevelStruct().(moveinterfaces.TargetCounter)

	if !ok {
		return errors.New("Unexpectedly did not implement TargetCount")
	}

	if gameStack.NumComponents() <= targetCounter.TargetCount() {
		return nil
	}

	return d.RoundRobin.ConditionMet(state)
}
