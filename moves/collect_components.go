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

boardgame:codegen
*/
type CollectCountComponents struct {
	DealCountComponents
}

//RoundRobinAction moves a component from the PlayerStack to the GameStack, as
//configured by the PlayerStacker and GameStacker interfaces.
func (d *CollectCountComponents) RoundRobinAction(playerState boardgame.PlayerState) error {

	playerStack, gameStack, err := dealActionHelper(d.TopLevelStruct(), playerState)

	if err != nil {
		return err
	}

	return playerStack.First().MoveToNextSlot(gameStack)

}

//FallbackName returns a string based on the names of the player
//stack name, game stack name, and target count.
func (d *CollectCountComponents) FallbackName(m *boardgame.GameManager) string {

	player, game, count := d.moveTypeInfo(m.ExampleState())

	return "Collect Components From " + player + " in each PlayerState To " + game + " in GameState " + count + " Times Per Player"
}

//FallbackHelpText returns a string based on the names of the player
//stack name, game stack name, and target count.
func (d *CollectCountComponents) FallbackHelpText() string {
	player, game, count := d.moveTypeInfo(nil)

	return "Collects " + count + " components from " + player + " in each PlayerState to " + game + " in GameState"
}

//CollectComponentsUntilPlayerCountLeft goes around and collects components
//from each player until each player has TargetCount() or fewer components in
//their PlayerStack(). It's the same as DealComponentsUntilPlayerCountReached,
//just with the action reversed and the size check flipped.
//
//boardgame:codegen
type CollectComponentsUntilPlayerCountLeft struct {
	DealComponentsUntilPlayerCountReached
}

//RoundRobinAction moves a component from the PlayerStack to the GameStack, as
//configured by the PlayerStacker and GameStacker interfaces.
func (d *CollectComponentsUntilPlayerCountLeft) RoundRobinAction(playerState boardgame.PlayerState) error {

	playerStack, gameStack, err := dealActionHelper(d.TopLevelStruct(), playerState)

	if err != nil {
		return err
	}

	return playerStack.First().MoveToNextSlot(gameStack)

}

//PlayerConditionMet is true if the NumComponents in the given player's
//PlayerStack() is TargetCount or less.
func (d *CollectComponentsUntilPlayerCountLeft) PlayerConditionMet(pState boardgame.ImmutablePlayerState) bool {
	playerCount, targetCount, err := dealComponentsPlayerConditionMetHelper(d.TopLevelStruct(), pState)

	if err != nil {
		return false
	}

	return playerCount <= targetCount
}

//FallbackName returns a string based on the names of the player
//stack name, game stack name, and target count.
func (d *CollectComponentsUntilPlayerCountLeft) FallbackName(m *boardgame.GameManager) string {

	player, game, count := d.moveTypeInfo(m.ExampleState())

	return "Collect Components From " + player + " in each PlayerState To " + game + " In GameState Until Each Player Is Down To " + count
}

//FallbackHelpText returns a string based on the names of the player
//stack name, game stack name, and target count.
func (d *CollectComponentsUntilPlayerCountLeft) FallbackHelpText() string {
	player, game, count := d.moveTypeInfo(nil)

	return "Collects components from " + player + " in each PlayerState to " + game + " in GameState until each player has " + count + " left"
}

//CollectComponentsUntilGameCountReached goes around and collects components from
//each player until GameStack() NumComponents() is TargetCount or greater. It's
//the same as DealComponentsUntilGameCountLeft, just with the action
//reversed and the size check flipped.
//
//boardgame:codegen
type CollectComponentsUntilGameCountReached struct {
	DealComponentsUntilGameCountLeft
}

//RoundRobinAction moves a component from the PlayerStack to the GameStack, as
//configured by the PlayerStacker and GameStacker interfaces.
func (d *CollectComponentsUntilGameCountReached) RoundRobinAction(playerState boardgame.PlayerState) error {

	playerStack, gameStack, err := dealActionHelper(d.TopLevelStruct(), playerState)

	if err != nil {
		return err
	}

	return playerStack.First().MoveToNextSlot(gameStack)

}

//ConditionMet returns nil if GameStack's NumComponents is TargetCount or
//greater, and otherwise defaults to RoundRobin's ConditionMet.
func (d *CollectComponentsUntilGameCountReached) ConditionMet(state boardgame.ImmutableState) error {

	gameCount, targetCount, err := dealComponentsConditionMetHelper(d.TopLevelStruct(), state)

	if err != nil {
		return nil
	}

	if gameCount >= targetCount {
		return nil
	}

	return d.RoundRobin.ConditionMet(state)

}

//FallbackName returns a string based on the names of the player
//stack name, game stack name, and target count.
func (d *CollectComponentsUntilGameCountReached) FallbackName(m *boardgame.GameManager) string {

	player, game, count := d.moveTypeInfo(m.ExampleState())

	return "Collect Components From " + player + " in each PlayerState To " + game + " In GameState Until The Game Has " + count + " Total"
}

//FallbackHelpText returns a string based on the names of the player
//stack name, game stack name, and target count.
func (d *CollectComponentsUntilGameCountReached) FallbackHelpText() string {
	player, game, count := d.moveTypeInfo(nil)

	return "Collects components from " + player + " in each PlayerState to " + game + " in GameState until the game has " + count + " total"
}
