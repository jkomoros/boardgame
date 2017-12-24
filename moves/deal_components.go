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

func dealComponentsPlayerConditionMetHelper(topLevelStruct boardgame.Move, playerState boardgame.PlayerState) (playerCount, targetCount int, err error) {
	playerStacker, ok := topLevelStruct.(moveinterfaces.PlayerStacker)

	if !ok {
		return 0, 0, errors.New("Didn't implement playerStacker")
	}

	//Ugly hack. :-/
	mutablePState := playerState.(boardgame.MutablePlayerState)

	playerStack := playerStacker.PlayerStack(mutablePState)

	targetCounter, ok := topLevelStruct.(moveinterfaces.TargetCounter)

	if !ok {
		return 0, 0, errors.New("Didn't implement target counter")
	}

	return playerStack.NumComponents(), targetCounter.TargetCount(), nil
}

func dealComponentsConditionMetHelper(topLevelStruct boardgame.Move, state boardgame.State) (gameCount, targetCount int, err error) {
	gameStacker, ok := topLevelStruct.(moveinterfaces.GameStacker)

	if !ok {
		return 0, 0, errors.New("Unexpectedly didn't implement gameStacker")
	}

	//Total hack :-/
	mutableState := state.(boardgame.MutableState)

	gameStack := gameStacker.GameStack(mutableState.MutableGameState())

	if gameStack == nil {
		return 0, 0, errors.New("GameStack gave a nil stack")
	}

	targetCounter, ok := topLevelStruct.(moveinterfaces.TargetCounter)

	if !ok {
		return 0, 0, errors.New("Unexpectedly did not implement TargetCount")
	}

	return gameStack.NumComponents(), targetCounter.TargetCount(), nil

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

//TargetCount should return the count that you want to target. Will return the
//configuration option passed via WithTargetCount in DefaultConfig, or 1 if
//that wasn't provided.
func (d *DealCountComponents) TargetCount() int {

	config := d.Info().Type().CustomConfiguration()

	val, ok := config[configNameTargetCount]

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
	targetCounter, ok := d.TopLevelStruct().(moveinterfaces.TargetCounter)

	if !ok {
		return 1
	}

	return targetCounter.TargetCount()
}

//PlayerStack by default just returns the property on GameState with the name
//passed to DefaultConfig by WithPlayerStack. If that is not sufficient,
//override this in your embedding struct.
func (d *DealCountComponents) PlayerStack(playerState boardgame.MutablePlayerState) boardgame.MutableStack {
	config := d.Info().Type().CustomConfiguration()

	stackName, ok := config[configNamePlayerStack]

	if !ok {
		return nil
	}

	strStackName, ok := stackName.(string)

	if !ok {
		return nil
	}

	stack, err := playerState.ReadSetter().MutableStackProp(strStackName)

	if err != nil {
		return nil
	}

	return stack
}

//GameStack by default just returns the property on GameState with the name
//passed to DefaultConfig by WithGameStack. If that is not sufficient,
//override this in your embedding struct.
func (d *DealCountComponents) GameStack(gameState boardgame.MutableSubState) boardgame.MutableStack {
	config := d.Info().Type().CustomConfiguration()

	stackName, ok := config[configNameGameStack]

	if !ok {
		return nil
	}

	strStackName, ok := stackName.(string)

	if !ok {
		return nil
	}

	stack, err := gameState.ReadSetter().MutableStackProp(strStackName)

	if err != nil {
		return nil
	}

	return stack
}

func (d *DealCountComponents) ValidConfiguration(exampleState boardgame.MutableState) error {

	playerStacker, ok := d.TopLevelStruct().(moveinterfaces.PlayerStacker)

	if !ok {
		return errors.New("Embedding move doesn't implement PlayerStacker")
	}

	if playerStacker.PlayerStack(exampleState.MutablePlayerStates()[0]) == nil {
		return errors.New("PlayerStack returned a nil stack")
	}

	gameStacker, ok := d.TopLevelStruct().(moveinterfaces.GameStacker)

	if !ok {
		return errors.New("Embedding move doesn't implement GameStacker")
	}

	if gameStacker.GameStack(exampleState.MutableGameState()) == nil {
		return errors.New("GameStack returned a nil stack")
	}

	targetCounter, ok := d.TopLevelStruct().(moveinterfaces.TargetCounter)

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
func (d *DealCountComponents) RoundRobinAction(playerState boardgame.MutablePlayerState) error {

	playerStack, gameStack, err := dealActionHelper(d.TopLevelStruct(), playerState)

	if err != nil {
		return err
	}

	return gameStack.MoveComponent(boardgame.FirstComponentIndex, playerStack, boardgame.NextSlotIndex)
}

//moveTypeInfo is used as a helper to generate sttrings for all of the MoveType getters.
func (d *DealCountComponents) moveTypeInfo(manager *boardgame.GameManager) (player, game, count string) {
	exampleState := manager.ExampleState()

	playerName := "unknown stack"
	gameName := "unknown stack"

	if playerStack, gameStack, err := dealActionHelper(d.TopLevelStruct(), exampleState.MutablePlayerStates()[0]); err == nil {

		if name := stackPropName(playerStack, exampleState); name != "" {
			playerName = name
		}
		if name := stackPropName(gameStack, exampleState); name != "" {
			gameName = name
		}
	}

	return playerName, gameName, targetCountString(d.TopLevelStruct())
}

func (d *DealCountComponents) MoveTypeFallbackName(manager *boardgame.GameManager) string {

	player, game, count := d.moveTypeInfo(manager)

	return "Deal Components From Game Stack " + game + " To Player Stack " + player + " To Each Player " + count + " Times"
}

func (d *DealCountComponents) MoveTypeHelpText(manager *boardgame.GameManager) string {
	player, game, count := d.moveTypeInfo(manager)

	return "Deals " + count + " components from game stack " + game + " to each player stack " + player
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
	playerCount, targetCount, err := dealComponentsPlayerConditionMetHelper(d.TopLevelStruct(), pState)

	if err != nil {
		return false
	}

	return playerCount >= targetCount
}

//ConditionMet simply returns the RoundRobin.ConditionMet (throwing out the
//RoundCount alternate of ConditionMet we get via sub-classing), since our
//PlayerConditionMet handles the end condtion.
func (d *DealComponentsUntilPlayerCountReached) ConditionMet(state boardgame.State) error {
	return d.RoundRobin.ConditionMet(state)
}

func (d *DealComponentsUntilPlayerCountReached) MoveTypeFallbackName(manager *boardgame.GameManager) string {

	player, game, count := d.moveTypeInfo(manager)

	return "Deal Components From Game Stack " + game + " To Player Stack " + player + " Until Each Player Has " + count
}

func (d *DealComponentsUntilPlayerCountReached) MoveTypeHelpText(manager *boardgame.GameManager) string {
	player, game, count := d.moveTypeInfo(manager)

	return "Deals components from game stack " + game + " to each player's " + player + " until each player has " + count
}

//DealComponentsUntilGameCountLeft goes around and deals components to each
//player until the GameStack() has TargetCount() or fewer components left.
type DealComponentsUntilGameCountLeft struct {
	DealCountComponents
}

//ConditionMet returns nil if GameStack's NumComponents is TargetCount or
//less, and otherwise defaults to RoundRobin's ConditionMet.
func (d *DealComponentsUntilGameCountLeft) ConditionMet(state boardgame.State) error {

	gameCount, targetCount, err := dealComponentsConditionMetHelper(d.TopLevelStruct(), state)

	if err != nil {
		return nil
	}

	if gameCount <= targetCount {
		return nil
	}

	return d.RoundRobin.ConditionMet(state)

}

func (d *DealComponentsUntilGameCountLeft) MoveTypeFallbackName(manager *boardgame.GameManager) string {

	player, game, count := d.moveTypeInfo(manager)

	return "Deal Components From Game Stack " + game + " To Player Stack " + player + " Until Game Stack Has " + count + " Total"
}

func (d *DealComponentsUntilGameCountLeft) MoveTypeHelpText(manager *boardgame.GameManager) string {
	player, game, count := d.moveTypeInfo(manager)

	return "Deals components from game stack " + game + " to each player's " + player + " until the game stack has " + count + " left"
}
