package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

func dealActionHelper(topLevelStruct boardgame.Move, playerState boardgame.PlayerState) (playerStack boardgame.Stack, gameStack boardgame.Stack, err error) {
	playerStacker, ok := topLevelStruct.(interfaces.PlayerStacker)

	if !ok {
		return nil, nil, errors.New("Embedding move unexpectedly doesn't implement PlayerStacker")
	}

	targetStack := playerStacker.PlayerStack(playerState)

	if targetStack == nil {
		return nil, nil, errors.New("PlayerStacker didn't return a valid stack")
	}

	gameStacker, ok := topLevelStruct.(interfaces.GameStacker)

	if !ok {
		return nil, nil, errors.New("Embedding move unexpectedly doesn't implement GameStacker")
	}

	sourceStack := gameStacker.GameStack(playerState.State().GameState())

	if sourceStack == nil {
		return nil, nil, errors.New("GameStacker didn't return a valid stack")
	}

	return targetStack, sourceStack, nil
}

func dealComponentsPlayerConditionMetHelper(topLevelStruct boardgame.Move, playerState boardgame.ImmutablePlayerState) (playerCount, targetCount int, err error) {
	playerStacker, ok := topLevelStruct.(interfaces.PlayerStacker)

	if !ok {
		return 0, 0, errors.New("Didn't implement playerStacker")
	}

	//Ugly hack. :-/
	mutablePState := playerState.(boardgame.PlayerState)

	playerStack := playerStacker.PlayerStack(mutablePState)

	targetCounter, ok := topLevelStruct.(interfaces.TargetCounter)

	if !ok {
		return 0, 0, errors.New("Didn't implement target counter")
	}

	return playerStack.NumComponents(), targetCounter.TargetCount(), nil
}

func dealComponentsConditionMetHelper(topLevelStruct boardgame.Move, state boardgame.ImmutableState) (gameCount, targetCount int, err error) {
	gameStacker, ok := topLevelStruct.(interfaces.GameStacker)

	if !ok {
		return 0, 0, errors.New("Unexpectedly didn't implement gameStacker")
	}

	//Total hack :-/
	mutableState := state.(boardgame.State)

	gameStack := gameStacker.GameStack(mutableState.GameState())

	if gameStack == nil {
		return 0, 0, errors.New("GameStack gave a nil stack")
	}

	targetCounter, ok := topLevelStruct.(interfaces.TargetCounter)

	if !ok {
		return 0, 0, errors.New("Unexpectedly did not implement TargetCount")
	}

	return gameStack.NumComponents(), targetCounter.TargetCount(), nil

}

/*

DealCountComponents is a type of RoundRobin move that deals components from
gameState's GameStack() to each PlayerState's PlayerStack(). It goes around
TargetCount() times. TargetCount() defaults to 1; override if you want to deal
out a different number of components. In practice it is more common to use
this move (and its subclasses) directly, and pass configuration for GameStack,
PlayerStack, and TargetCount via WithGameStack, WithPlayerStack, and
WithTargetCount into auto.Config.

boardgame:codegen
*/
type DealCountComponents struct {
	RoundRobinNumRounds
}

//TargetCount should return the count that you want to target. Will return the
//configuration option passed via WithTargetCount in auto.Config, or 1 if
//that wasn't provided.
func (d *DealCountComponents) TargetCount() int {

	config := d.CustomConfiguration()

	val, ok := config[configPropTargetCount]

	if !ok {
		//No configuration provided, just return default
		return 1
	}

	intVal, ok := val.(int)

	if !ok {
		//signal error
		return -1
	}

	return intVal
}

//NumRounds simply returns TargetCount. NumRounds is what RoundRobinNumRounds
//expects, but TargetCount() is the terminology used for all of the similar
//Deal/Collect/MoveComponents methods.
func (d *DealCountComponents) NumRounds() int {
	targetCounter, ok := d.TopLevelStruct().(interfaces.TargetCounter)

	if !ok {
		return 1
	}

	return targetCounter.TargetCount()
}

//PlayerStack by default just returns the property on GameState with the name
//passed to auto.Config by WithPlayerStack. If that is not sufficient,
//override this in your embedding struct.
func (d *DealCountComponents) PlayerStack(playerState boardgame.PlayerState) boardgame.Stack {
	config := d.CustomConfiguration()

	stackName, ok := config[configPropPlayerStack]

	if !ok {
		return nil
	}

	strStackName, ok := stackName.(string)

	if !ok {
		return nil
	}

	stack, err := playerState.ReadSetter().StackProp(strStackName)

	if err != nil {
		return nil
	}

	return stack
}

//GameStack by default just returns the property on GameState with the name
//passed to auto.Config by WithGameStack. If that is not sufficient,
//override this in your embedding struct.
func (d *DealCountComponents) GameStack(gameState boardgame.SubState) boardgame.Stack {
	config := d.CustomConfiguration()

	stackName, ok := config[configPropGameStack]

	if !ok {
		return nil
	}

	strStackName, ok := stackName.(string)

	if !ok {
		return nil
	}

	stack, err := gameState.ReadSetter().StackProp(strStackName)

	if err != nil {
		return nil
	}

	return stack
}

func (d *DealCountComponents) ValidConfiguration(exampleState boardgame.State) error {

	playerStacker, ok := d.TopLevelStruct().(interfaces.PlayerStacker)

	if !ok {
		return errors.New("Embedding move doesn't implement PlayerStacker")
	}

	if playerStacker.PlayerStack(exampleState.PlayerStates()[0]) == nil {
		return errors.New("PlayerStack returned a nil stack")
	}

	gameStacker, ok := d.TopLevelStruct().(interfaces.GameStacker)

	if !ok {
		return errors.New("Embedding move doesn't implement GameStacker")
	}

	if gameStacker.GameStack(exampleState.GameState()) == nil {
		return errors.New("GameStack returned a nil stack")
	}

	targetCounter, ok := d.TopLevelStruct().(interfaces.TargetCounter)

	if !ok {
		return errors.New("Embedding move doesn't implement TargetCounter")
	}

	if targetCounter.TargetCount() < 0 {
		return errors.New("TargetCount returned a number below 0, which signals an error")
	}

	return d.RoundRobinNumRounds.ValidConfiguration(exampleState)
}

//RoundRobinAction moves a component from the GameStack to the PlayerStack, as
//configured by the PlayerStacker and GameStacker interfaces.
func (d *DealCountComponents) RoundRobinAction(playerState boardgame.PlayerState) error {

	playerStack, gameStack, err := dealActionHelper(d.TopLevelStruct(), playerState)

	if err != nil {
		return err
	}

	return gameStack.First().MoveToNextSlot(playerStack)
}

//moveTypeInfo is used as a helper to generate sttrings for all of the MoveType getters.
func (d *DealCountComponents) moveTypeInfo(exampleState boardgame.ImmutableState) (player, game, count string) {

	var playerStack boardgame.Stack
	var gameStack boardgame.Stack

	if exampleState != nil {

		//Ugly hack to cast these to mutable state :-/
		playerState := exampleState.ImmutablePlayerStates()[0].(boardgame.PlayerState)

		playerStack, gameStack, _ = dealActionHelper(d.TopLevelStruct(), playerState)

	}

	return stackName(d, configPropPlayerStack, playerStack, exampleState), stackName(d, configPropGameStack, gameStack, exampleState), targetCountString(d.TopLevelStruct())
}

//FallbackName returns a string based on the names of the player
//stack name, game stack name, and target count.
func (d *DealCountComponents) FallbackName(m *boardgame.GameManager) string {

	player, game, count := d.moveTypeInfo(m.ExampleState())

	return "Deal Components From " + game + " In GameState To " + player + " In PlayerState To Each Player " + count + " Times"
}

//FallbackHelpText returns a string based on the names of the player
//stack name, game stack name, and target count.
func (d *DealCountComponents) FallbackHelpText() string {
	player, game, count := d.moveTypeInfo(nil)

	return "Deals " + count + " components from " + game + " in GameState to " + player + " in each PlayerState"
}

//DealComponentsUntilPlayerCountReached goes around and deals components to
//each player until each player has TargetCount() or greater components in
//their PlayerStack().
//
//boardgame:codegen
type DealComponentsUntilPlayerCountReached struct {
	DealCountComponents
}

//PlayerConditionMet is true if the NumComponents in the given player's
//PlayerStack() is TargetCount or greater.
func (d *DealComponentsUntilPlayerCountReached) PlayerConditionMet(pState boardgame.ImmutablePlayerState) bool {
	playerCount, targetCount, err := dealComponentsPlayerConditionMetHelper(d.TopLevelStruct(), pState)

	if err != nil {
		return false
	}

	return playerCount >= targetCount
}

//ConditionMet simply returns the RoundRobin.ConditionMet (throwing out the
//RoundCount alternate of ConditionMet we get via sub-classing), since our
//PlayerConditionMet handles the end condtion.
func (d *DealComponentsUntilPlayerCountReached) ConditionMet(state boardgame.ImmutableState) error {
	return d.RoundRobin.ConditionMet(state)
}

//FallbackName returns a string based on the names of the player
//stack name, game stack name, and target count.
func (d *DealComponentsUntilPlayerCountReached) FallbackName(m *boardgame.GameManager) string {

	player, game, count := d.moveTypeInfo(m.ExampleState())

	return "Deal Components From " + game + " In GameState To " + player + " In Each PlayerState Until Each Player Has " + count
}

//allbackHelpText returns a string based on the names of the player
//stack name, game stack name, and target count.
func (d *DealComponentsUntilPlayerCountReached) FallbackHelpText() string {
	player, game, count := d.moveTypeInfo(nil)

	return "Deals components from " + game + " in GameState to " + player + " in each PlayerState until each player has " + count
}

//DealComponentsUntilGameCountLeft goes around and deals components to each
//player until the GameStack() has TargetCount() or fewer components left.
//
//boardgame:codegen
type DealComponentsUntilGameCountLeft struct {
	DealCountComponents
}

//ConditionMet returns nil if GameStack's NumComponents is TargetCount or
//less, and otherwise defaults to RoundRobin's ConditionMet.
func (d *DealComponentsUntilGameCountLeft) ConditionMet(state boardgame.ImmutableState) error {

	gameCount, targetCount, err := dealComponentsConditionMetHelper(d.TopLevelStruct(), state)

	if err != nil {
		return nil
	}

	if gameCount <= targetCount {
		return nil
	}

	return d.RoundRobin.ConditionMet(state)

}

//FallbackName returns a string based on the names of the player
//stack name, game stack name, and target count.
func (d *DealComponentsUntilGameCountLeft) FallbackName(m *boardgame.GameManager) string {

	player, game, count := d.moveTypeInfo(m.ExampleState())

	return "Deal Components From " + game + " In GameState To " + player + " In Each PlayerState Until Game Stack Has " + count + " Total"
}

//FallbackHelpText returns a string based on the names of the player
//stack name, game stack name, and target count.
func (d *DealComponentsUntilGameCountLeft) FallbackHelpText() string {
	player, game, count := d.moveTypeInfo(nil)

	return "Deals components from " + game + " in GameState to " + player + " in each PlayerState until the game stack has " + count + " left"
}
