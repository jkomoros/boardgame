package moves

import (
	"github.com/jkomoros/boardgame"
)

/*

CollectCountComponents is a type of RoundRobin move that collects components from
each PlayerState's PlayerStack() to gameState's GameStack(). By default it
goes around once and collects a component from each. If you want a different
number of rounds, override TargetCount(). It subclasses DealCountComponents,
but simply reverses the action to make.

*/
type CollectCountComponents struct {
	DealCountComponents
}

//RoundRobinAction moves a component from the PlayerStack to the GameStack, as
//configured by the PlayerStacker and GameStacker interfaces.
func (d *CollectCountComponents) RoundRobinAction(playerState boardgame.MutablePlayerState) error {

	playerStack, gameStack, err := dealActionHelper(d.TopLevelStruct(), playerState)

	if err != nil {
		return err
	}

	return playerStack.MoveComponent(boardgame.FirstComponentIndex, gameStack, boardgame.NextSlotIndex)

}

//CollectComponentsUntilPlayerCountLeft goes around and collects components
//from each player until each player has TargetCount() or fewer components in
//their PlayerStack(). It's the same as DealComponentsUntilPlayerCountReached,
//just with the action reversed and the size check flipped.
type CollectComponentsUntilPlayerCountLeft struct {
	DealComponentsUntilPlayerCountReached
}

//RoundRobinAction moves a component from the PlayerStack to the GameStack, as
//configured by the PlayerStacker and GameStacker interfaces.
func (d *CollectComponentsUntilPlayerCountLeft) RoundRobinAction(playerState boardgame.MutablePlayerState) error {

	playerStack, gameStack, err := dealActionHelper(d.TopLevelStruct(), playerState)

	if err != nil {
		return err
	}

	return playerStack.MoveComponent(boardgame.FirstComponentIndex, gameStack, boardgame.NextSlotIndex)

}

//PlayerConditionMet is true if the NumComponents in the given player's
//PlayerStack() is TargetCount or less.
func (d *CollectComponentsUntilPlayerCountLeft) PlayerConditionMet(pState boardgame.PlayerState) bool {
	playerCount, targetCount, err := dealComponentsPlayerConditionMetHelper(d.TopLevelStruct(), pState)

	if err != nil {
		return false
	}

	return playerCount <= targetCount
}

//CollectComponentsUntilGameCountReached goes around and collects components from
//each player until GameStack() NumComponents() is TargetCount or greater. It's
//the same as DealComponentsUntilGameCountLeft, just with the action
//reversed and the size check flipped.
type CollectComponentsUntilGameCountReached struct {
	DealComponentsUntilGameCountLeft
}

//RoundRobinAction moves a component from the PlayerStack to the GameStack, as
//configured by the PlayerStacker and GameStacker interfaces.
func (d *CollectComponentsUntilGameCountReached) RoundRobinAction(playerState boardgame.MutablePlayerState) error {

	playerStack, gameStack, err := dealActionHelper(d.TopLevelStruct(), playerState)

	if err != nil {
		return err
	}

	return playerStack.MoveComponent(boardgame.FirstComponentIndex, gameStack, boardgame.NextSlotIndex)

}

//ConditionMet returns nil if GameStack's NumComponents is TargetCount or
//greater, and otherwise defaults to RoundRobin's ConditionMet.
func (d *CollectComponentsUntilGameCountReached) ConditionMet(state boardgame.State) error {

	gameCount, targetCount, err := dealComponentsConditionMetHelper(d.TopLevelStruct(), state)

	if err != nil {
		return nil
	}

	if gameCount >= targetCount {
		return nil
	}

	return d.RoundRobin.ConditionMet(state)

}
