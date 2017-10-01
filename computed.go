package boardgame

import (
	"encoding/json"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
	"strconv"
)

//ComputedProperties represents a collection of computed properties for a
//given state. An object conforming to this interface will be returned from
//state.Computed(). Its values will be set based on what
//Delegate.ComputedPropertiesConfig returns.
type ComputedProperties interface {
	//The primary property reader is where top-level computed properties can
	//be accessed.
	Global() MutableSubState
	//To get the ComputedPlayerProperties, pass in the player index.
	Player(index PlayerIndex) MutableSubState
}

//ComputedPropertiesConfig is the struct that contains configuration for which
//properties to compute and how to compute them. See the package documentation
//on Computed Properties for more information.
type ComputedPropertiesConfig struct {
	//The top-level computed properties.
	Global map[string]ComputedGlobalPropertyDefinition
	//The properties that are computed for each PlayerState individually.
	Player map[string]ComputedPlayerPropertyDefinition
}

//ComputedGlobalPropertyDefinition defines how to calculate a given top-level
//computed property.
type ComputedGlobalPropertyDefinition struct {
	//Dependencies exhaustively enumerates all of the properties that need to
	//be populated on the ShadowState to calculate this value. Defining your
	//dependencies allows us to only recalculate computed properties when
	//necessary, and other kewl tricks.
	Dependencies []StatePropertyRef
	//The thing we expect to be able to cast the result of Compute to (since
	//the method necessarily has to be general).
	PropType PropertyType
	//Where the actual logic of the computed property goes. sanitizedState
	//will be a Sanitized() State populated with all of the properties
	//enumerated in Dependencies, with the other properties obscured with
	//PolicyRandom  (For PlayerState properties, we will include that property
	//on each ShadowPlayerState object). Since it's just a sanitized State,
	//you may cast the state to the concrete types for your game to more
	//easily retrieve values. The return value will be casted to PropType
	//afterward. Return an error if any state is configured in an unexpected
	//way. Note: your compute function should be resilient to values that are
	//sanitized. In many cases it makes sense to factor your compute
	//computation out into a shim that fetches the relevant properties from
	//the ShadowState and then passes them to the core computation function,
	//so that other methods can reuse the same logic.
	Compute func(sanitizedState State) (interface{}, error)
}

//ComputedPlayerPropertyDefinition is the analogue for
//ComputedPropertyDefintion, but operates on a single PlayerState at a time
//and returns properties for that particular PlayerState.
type ComputedPlayerPropertyDefinition struct {
	//Dependencies exhaustively enumerates all of the properties that need to
	//be populated on the ShadowState to calculate this value. Defining your
	//dependencies allows us to only recalculate computed properties when
	//necessary, and other kewl tricks. All Dependencies must have Group
	//StateGroupPlayer, otherwise the computation will error.
	Dependencies []StatePropertyRef
	//The thing we expect to be able to cast the result of Compute to (since
	//the method necessarily has to be general).
	PropType PropertyType
	//Where the actual logic of the computed property goes. playerState will
	//be a PlayerState from a Sanitized() state, populated with all of the
	//properties enumerated in Dependencies, with other properties obscured by
	//PolicyRandom. Since it's just a PlayerState from a sanitized State, it
	//is safe to cast to the underlying PlayerState type you know it is for
	//your package for convenience. This method will be called once per
	//PlayerState in turn. The return value will be casted to PropType
	//afterward. Return an error if any state is configured in an unexpected
	//way. Note: your compute function should be resilient to values that are
	//sanitized. In many cases it makes sense to factor your compute
	//computation out into a shim that fetches the relevant properties from
	//the ShadowState and then passes them to the core computation function,
	//so that other methods can reuse the same logic. If you need more state
	//than what is available on just the playerState, consider defining
	//GlobalCompute instead.
	Compute func(playerState PlayerState) (interface{}, error)

	//If Compute is nil but GlobalCompute is non-nil, then GlobalCompute will
	//be called instead. Instead of passing in just the single playerState for
	//the player in question, it passes the fullState, as well as the
	//PlayerIndex currently being prepared for, and expects the caller to
	//return the value for the specific playerIndex.
	GlobalCompute func(state State, player PlayerIndex) (interface{}, error)
}

//StateGroupType is the top-level grouping object used in a StatePropertyRef.
type StateGroupType int

const (
	StateGroupGame StateGroupType = iota
	StateGroupPlayer
	StateGroupDynamicComponentValues
)

//A StatePropertyRef is a reference to a particular property in a State, in a
//structured way. Currently used when defining your dependencies for computed
//properties.
type StatePropertyRef struct {
	Group StateGroupType
	//DeckName is only used when Group is StateGroupDynamicComponentValues
	DeckName string
	//PropName is the specific property on the given SubStateObject specified
	//by the rest of the StatePropertyRef.
	PropName string
}

type computedPropertiesImpl struct {
	global  MutableSubState
	players []MutableSubState
}

func policyForDependencies(dependencies []StatePropertyRef) *StatePolicy {
	result := &StatePolicy{
		Game:                   make(SubStatePolicy),
		Player:                 make(SubStatePolicy),
		DynamicComponentValues: make(map[string]SubStatePolicy),
	}
	for _, dependency := range dependencies {
		if dependency.Group == StateGroupGame {
			result.Game[dependency.PropName] = GroupPolicy{
				GroupAll: PolicyVisible,
			}
		} else if dependency.Group == StateGroupPlayer {
			result.Player[dependency.PropName] = GroupPolicy{
				GroupAll: PolicyVisible,
			}
		} else if dependency.Group == StateGroupDynamicComponentValues {
			if _, ok := result.DynamicComponentValues[dependency.DeckName]; !ok {
				result.DynamicComponentValues[dependency.DeckName] = make(SubStatePolicy)
			}
			policy := result.DynamicComponentValues[dependency.DeckName]
			policy[dependency.PropName] = GroupPolicy{
				GroupAll: PolicyVisible,
			}
		}
	}

	return result
}

func newComputedPropertiesImpl(config *ComputedPropertiesConfig, state *state) (*computedPropertiesImpl, error) {

	if !state.calculatingComputed {
		return nil, errors.New("State didn't think it was calculatingComputed when it was")
	}

	if config == nil {
		//It's fine if no config is provided--that just means no computed
		//properties.
		return nil, nil
	}

	playerBags := make([]MutableSubState, len(state.PlayerStates()))

	//TODO: calculate all properties.
	for i, _ := range state.PlayerStates() {
		collection := state.game.manager.delegate.ComputedPlayerPropertyCollectionConstructor()
		if collection == nil {
			collection = newGenericReader()
		} else if collection.ReadSetter() == nil {
			return nil, errors.New("Player State readsetter returned nil")
		}

		playerBags[i] = collection

		if config.Player == nil {
			continue
		}

		reader := collection.ReadSetter()

		for name, propConfig := range config.Player {
			if err := propConfig.calculate(name, PlayerIndex(i), state, reader); err != nil {
				//TODO: do something better here.
				return nil, errors.Extend(err, "Player failed")
			}
		}

	}

	globalBag := state.game.manager.delegate.ComputedGlobalPropertyCollectionConstructor()

	if globalBag == nil {
		globalBag = newGenericReader()
	} else if globalBag.ReadSetter() == nil {
		return nil, errors.New("Global bag readSetter returned nil")
	}

	if config.Global != nil {
		for name, propConfig := range config.Global {
			if err := propConfig.calculate(name, state, globalBag.ReadSetter()); err != nil {
				//TODO: do something better here.
				return nil, errors.Extend(err, "global failed")
			}
		}
	}

	return &computedPropertiesImpl{
		global:  globalBag,
		players: playerBags,
	}, nil
}

func (c *ComputedGlobalPropertyDefinition) calculate(propName string, state *state, output PropertyReadSetter) error {

	result, err := c.compute(state)

	if err != nil {
		return errors.New("Error computing calculated prop: " + err.Error())
	}

	switch c.PropType {
	case TypeBool:
		boolVal, ok := result.(bool)
		if !ok {
			return errors.New("Property did not return bool as expected")
		}
		output.SetBoolProp(propName, boolVal)
	case TypeInt:
		intVal, ok := result.(int)
		if !ok {
			return errors.New("Property did not return int as expected")
		}
		output.SetIntProp(propName, intVal)
	case TypeString:
		stringVal, ok := result.(string)
		if !ok {
			return errors.New("Property did not return string as expected")
		}
		output.SetStringProp(propName, stringVal)
	case TypePlayerIndex:
		playerIndexVal, ok := result.(PlayerIndex)
		if !ok {
			return errors.New("Property did not return PlayerIndex as expected")
		}
		output.SetPlayerIndexProp(propName, playerIndexVal)
	case TypeGrowableStack:
		growableStackVal, ok := result.(*GrowableStack)
		if !ok {
			return errors.New("Property did not return growable stack as expected")
		}
		output.SetGrowableStackProp(propName, growableStackVal)
	case TypeSizedStack:
		sizedStackVal, ok := result.(*SizedStack)
		if !ok {
			return errors.New("Property did not return sized stack as expected")
		}
		output.SetSizedStackProp(propName, sizedStackVal)
	case TypeEnumConst:
		enumConstVal, ok := result.(enum.Const)
		if !ok {
			return errors.New("Property did not return enum const as expected")
		}
		output.SetEnumConstProp(propName, enumConstVal)
	case TypeEnumVar:
		enumVarVal, ok := result.(enum.Var)
		if !ok {
			return errors.New("Property did not return enum var as expected")
		}
		output.SetEnumVarProp(propName, enumVarVal)
	default:
		return errors.New("That property type, " + c.PropType.String() + " is not currently supported")
	}

	return nil

}

func (c *ComputedPlayerPropertyDefinition) calculate(propName string, playerIndex PlayerIndex, state *state, output PropertyReadSetter) error {

	result, err := c.compute(state, playerIndex)

	if err != nil {
		return errors.New("Error computing calculated prop: " + err.Error())
	}

	switch c.PropType {
	case TypeBool:
		boolVal, ok := result.(bool)
		if !ok {
			return errors.New("Property did not return bool as expected")
		}
		output.SetBoolProp(propName, boolVal)
	case TypeInt:
		intVal, ok := result.(int)
		if !ok {
			return errors.New("Property did not return int as expected")
		}
		output.SetIntProp(propName, intVal)
	case TypeString:
		stringVal, ok := result.(string)
		if !ok {
			return errors.New("Property did not return string as expected")
		}
		output.SetStringProp(propName, stringVal)
	case TypePlayerIndex:
		playerIndexVal, ok := result.(PlayerIndex)
		if !ok {
			return errors.New("Property did not return PlayerIndex as expected")
		}
		output.SetPlayerIndexProp(propName, playerIndexVal)
	case TypeGrowableStack:
		growableStackVal, ok := result.(*GrowableStack)
		if !ok {
			return errors.New("Property did not return growable stack as expected")
		}
		output.SetGrowableStackProp(propName, growableStackVal)
	case TypeSizedStack:
		sizedStackVal, ok := result.(*SizedStack)
		if !ok {
			return errors.New("Property did not return sized stack as expected")
		}
		output.SetSizedStackProp(propName, sizedStackVal)
	case TypeEnumConst:
		enumConstVal, ok := result.(enum.Const)
		if !ok {
			return errors.New("Property did not return enum const as expected")
		}
		output.SetEnumConstProp(propName, enumConstVal)
	case TypeEnumVar:
		enumVarVal, ok := result.(enum.Var)
		if !ok {
			return errors.New("Property did not return enum var as expected")
		}
		output.SetEnumVarProp(propName, enumVarVal)
	default:
		return errors.New("That property type, " + c.PropType.String() + " is not currently supported")
	}

	return nil

}

func (c *ComputedGlobalPropertyDefinition) compute(state *state) (interface{}, error) {

	//First, prepare a shadow state with all of the dependencies.

	policy := policyForDependencies(c.Dependencies)

	sanitized, err := state.deprecatedSanitizedWithDefault(policy, -1, PolicyRandom)

	if err != nil {
		return nil, errors.Extend(err, "Couldn't create randomized state for globals")
	}

	return c.Compute(sanitized)

}

func (c *ComputedPlayerPropertyDefinition) compute(state *state, playerIndex PlayerIndex) (interface{}, error) {

	policy := policyForDependencies(c.Dependencies)

	sanitized, err := state.deprecatedSanitizedWithDefault(policy, -1, PolicyRandom)

	if err != nil {
		return nil, errors.Extend(err, "Couldn't create randomized state for players")
	}

	if c.Compute != nil {
		return c.Compute(sanitized.PlayerStates()[playerIndex])
	}

	if c.GlobalCompute != nil {
		return c.GlobalCompute(sanitized, playerIndex)
	}

	return nil, errors.New("Neither Compute nor GlobalCompute were defined. One of them must be.")
}

func (c *computedPropertiesImpl) Global() MutableSubState {
	return c.global
}

func (c *computedPropertiesImpl) Player(index PlayerIndex) MutableSubState {
	return c.players[int(index)]
}

func (c *computedPropertiesImpl) MarshalJSON() ([]byte, error) {

	result := make(map[string]interface{})

	playerProperties := make([]map[string]interface{}, len(c.players))

	for i, player := range c.players {
		playerProperties[i] = make(map[string]interface{})
		for propName, _ := range player.Reader().Props() {
			val, err := player.Reader().Prop(propName)

			if err != nil {
				return nil, errors.New("Player computed prop " + propName + " for player " + strconv.Itoa(i) + " returned an error: " + err.Error())
			}
			playerProperties[i][propName] = val
		}
	}

	props := c.Global()

	globalProperties := make(map[string]interface{})

	for propName, _ := range props.Reader().Props() {
		val, err := props.Reader().Prop(propName)

		if err != nil {
			return nil, errors.New("Computed Prop " + propName + " returned an error: " + err.Error())
		}

		globalProperties[propName] = val
	}

	//TODO: can't I just have this have a default marshal JSON and then move
	//these sub-impls to the global and player group?

	result["Global"] = globalProperties

	result["Players"] = playerProperties

	return json.Marshal(result)
}
