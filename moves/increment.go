package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"strconv"
)

//Increment is a simple move that modifies the specified int property by
//adding Amount() to it. It's often useful to run in a move progression, for
//example to increment a round count. Its Legal() is just the Legal for FixUp,
//so this should only be used within a MoveProgression, unless you provide
//your own Legal method.
//
//boardgame:codegen
type Increment struct {
	FixUp
}

//ValidConfiguration checks to ensure that the specified property (via
//GameProperty and PlayerProperty) denotes an int property on the given sub-
//state object.
func (i *Increment) ValidConfiguration(exampleState boardgame.State) error {

	if err := i.FixUp.ValidConfiguration(exampleState); err != nil {
		return err
	}

	_, _, _, err := i.intProp(exampleState)

	return err

}

type gamePlayerProperty interface {
	GameProperty() string
	PlayerProperty() string
	Amount() int
}

//intProp returns the actual prop to use. If state is nil, skips the check of
//whether that's a valid int prop and assumes it is.
func (i *Increment) intProp(state boardgame.ImmutableState) (isPlayer bool, propName string, amount int, err error) {

	m := i.TopLevelStruct()

	proper, ok := m.(gamePlayerProperty)

	if !ok {
		return false, "", 0, errors.New("Top level struct did not implement GameProperty and PlayerProperty and Amount")
	}

	amount = proper.Amount()

	gamePropName := proper.GameProperty()

	if gamePropName != "" {

		if state != nil {
			_, err := state.ImmutableGameState().Reader().IntProp(gamePropName)

			if err != nil {
				return false, "", 0, errors.New("The provided game property was not an int prop on game: " + err.Error())
			}
		}

		return false, gamePropName, amount, nil
	}

	playerPropName := proper.PlayerProperty()

	if playerPropName != "" {

		if state != nil {

			_, err := state.ImmutablePlayerStates()[0].Reader().IntProp(playerPropName)

			if err != nil {
				return false, "", 0, errors.New("The provided player property was not an int prop on player: " + err.Error())
			}
		}

		return true, playerPropName, amount, nil
	}

	return false, "", 0, errors.New("Neither GameProperty or PlayerProperty specified a property")

}

//GameProperty returns the name of the GameProperty provided by
//WithGameProperty, or "" if that wasn't called.
func (i *Increment) GameProperty() string {
	config := i.CustomConfiguration()
	val, ok := config[configPropGameProperty]
	if !ok {
		return ""
	}
	strVal, ok := val.(string)
	if !ok {
		return ""
	}
	return strVal
}

//PlayerProperty returns the name of the property on PlayerState provided by
//WithPlayerProperty, or "" if that wasn't called.
func (i *Increment) PlayerProperty() string {
	config := i.CustomConfiguration()
	val, ok := config[configPropPlayerProperty]
	if !ok {
		return ""
	}
	strVal, ok := val.(string)
	if !ok {
		return ""
	}
	return strVal
}

//Amount returns the amount to increment the given property by. Will use the
//value passed to WithAmount, or 1 if that wasn't used.
func (i *Increment) Amount() int {
	config := i.CustomConfiguration()
	val, ok := config[configPropAmount]
	if !ok {
		return 1
	}
	intVal, ok := val.(int)
	if !ok {
		return 1
	}
	return intVal
}

//Apply will increment the given property by Amount(). If GameProperty returns
//a valid int property, will increment that. If GameProperty is not provided
//but PlayerProperty is, will increment that property on the CurrentPlayer. If
//the CurrentPlayer is not valid, will error.
func (i *Increment) Apply(state boardgame.State) error {

	isPlayer, propName, amount, err := i.intProp(state)

	if err != nil {
		return err
	}

	if isPlayer {

		playerState := state.CurrentPlayer()

		if playerState == nil {
			return errors.New("There is no current player at the moment")
		}

		currentVal, err := playerState.ReadSetter().IntProp(propName)

		if err != nil {
			return errors.New("Couldn't fetch player property off player state")
		}

		if err := playerState.ReadSetter().SetIntProp(propName, currentVal+amount); err != nil {
			return errors.New("Couldn't set the player property")
		}

	} else {

		currentVal, err := state.GameState().ReadSetter().IntProp(propName)

		if err != nil {
			return errors.New("Couldn't fetch game property off game state")
		}

		if err := state.GameState().ReadSetter().SetIntProp(propName, currentVal+amount); err != nil {
			return errors.New("Couldn't set the game property")
		}
	}

	return nil
}

//FallbackName returns "Increment {Game|Player} Property PROPNAME By Amount"
func (i *Increment) FallbackName(m *boardgame.GameManager) string {
	return "Increment " + i.helpTextSuffix()
}

func (i *Increment) helpTextSuffix() string {
	isPlayer, propName, amount, _ := i.intProp(nil)

	propType := "Game"

	if isPlayer {
		propType = "Player"
	}

	return propType + " Property " + propName + " By " + strconv.Itoa(amount)
}

//FallbackName returns "Increments {Game|Player} Property PROPNAME By Amount"
func (i *Increment) FallbackHelpText() string {
	return "Increments " + i.helpTextSuffix()
}
