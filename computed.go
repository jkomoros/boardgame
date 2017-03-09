package boardgame

import (
	"encoding/json"
	"errors"
	"strconv"
)

//ComputedProperties represents a collection of computed properties for a
//given state. An object conforming to this interface will be returned from
//state.Computed(). Its values will be set based on what
//Delegate.ComputedPropertiesConfig returns.
type ComputedProperties interface {
	//The primary property reader is where top-level computed properties can
	//be accessed.
	PropertyReader
	//To get the ComputedPlayerProperties, pass in the player index.
	Player(index int) PropertyReader
}

//ComputedPropertiesConfig is the struct that contains configuration for which
//properties to compute and how to compute them. See the package documentation
//on Computed Properties for more information.
type ComputedPropertiesConfig struct {
	//The top-level computed properties.
	Properties map[string]ComputedPropertyDefinition
	//The properties that are computed for each PlayerState individually.
	PlayerProperties map[string]ComputedPlayerPropertyDefinition
}

//ComputedPropertyDefinition defines how to calculate a given top-level
//computed property.
type ComputedPropertyDefinition struct {
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
	//on each ShadowPlayerState object). The return value will be casted to
	//PropType afterward. Return an error if any state is configured in an
	//unexpected way. Note: your compute function should be resilient to
	//values that are sanitized. In many cases it makes sense to factor your
	//compute computation out into a shim that fetches the relevant properties
	//from the ShadowState and then passes them to the core computation
	//function, so that other methods can reuse the same logic.
	Compute func(sanitizedState *State) (interface{}, error)
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
	//PolicyRandom. This method will be called once per PlayerState in turn.
	//The return value will be casted to PropType afterward. Return an error
	//if any state is configured in an unexpected way. Note: your compute
	//function should be resilient to values that are sanitized. In many cases
	//it makes sense to factor your compute computation out into a shim that
	//fetches the relevant properties from the ShadowState and then passes
	//them to the core computation function, so that other methods can reuse
	//the same logic.
	Compute func(playerState PlayerState) (interface{}, error)
}

//StateGroupType is the top-level grouping object used in a StatePropertyRef.
type StateGroupType int

const (
	StateGroupGame StateGroupType = iota
	StateGroupPlayer
)

//A StatePropertyRef is a reference to a particular property in a State, in a
//structured way. Currently used when defining your dependencies for computed
//properties.
type StatePropertyRef struct {
	Group    StateGroupType
	PropName string
}

//The private impl for ComputedProperties
type computedPropertiesImpl struct {
	bag     *computedPropertiesBag
	players []*computedPlayerPropertiesImpl
	state   *State
	config  *ComputedPropertiesConfig
}

type computedPlayerPropertiesImpl struct {
	bag    *computedPropertiesBag
	config map[string]ComputedPlayerPropertyDefinition
	state  *State
	i      int
}

type computedPropertiesBag struct {
	unknownProps       map[string]interface{}
	intProps           map[string]int
	boolProps          map[string]bool
	stringProps        map[string]string
	growableStackProps map[string]*GrowableStack
	sizedStackProps    map[string]*SizedStack
}

func policyForDependencies(dependencies []StatePropertyRef) *StatePolicy {
	result := &StatePolicy{
		Game:   make(map[string]GroupPolicy),
		Player: make(map[string]GroupPolicy),
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
		}
	}
	return result
}

func (c *ComputedPropertyDefinition) compute(state *State) (interface{}, error) {

	//First, prepare a shadow state with all of the dependencies.

	policy := policyForDependencies(c.Dependencies)

	sanitized := state.sanitizedWithExceptions(policy)

	return c.Compute(sanitized)

}

func (c *ComputedPlayerPropertyDefinition) compute(state *State, playerIndex int) (interface{}, error) {

	policy := policyForDependencies(c.Dependencies)

	sanitized := state.sanitizedWithExceptions(policy)

	return c.Compute(sanitized.Players[playerIndex])
}

func (c *computedPropertiesImpl) MarshalJSON() ([]byte, error) {

	result := make(map[string]interface{})

	playerProperties := make([]map[string]interface{}, len(c.players))

	for i, player := range c.players {
		playerProperties[i] = make(map[string]interface{})
		for propName, _ := range player.Props() {
			val, err := player.Prop(propName)

			if err != nil {
				return nil, errors.New("Player computed prop " + propName + " for player " + strconv.Itoa(i) + " returned an error: " + err.Error())
			}
			playerProperties[i][propName] = val
		}
	}

	for propName, _ := range c.Props() {
		val, err := c.Prop(propName)

		if err != nil {
			return nil, errors.New("Computed Prop " + propName + " returned an error: " + err.Error())
		}

		result[propName] = val
	}

	result["Players"] = playerProperties

	return json.Marshal(result)
}

func (c *computedPropertiesImpl) Player(index int) PropertyReader {
	return c.players[index]
}

func (c *computedPlayerPropertiesImpl) Props() map[string]PropertyType {
	result := make(map[string]PropertyType)

	if c.config == nil {
		return result
	}

	for name, config := range c.config {
		result[name] = config.PropType
	}

	return result
}

func (c *computedPropertiesImpl) Props() map[string]PropertyType {

	result := make(map[string]PropertyType)

	if c.config == nil {
		return result
	}

	for name, config := range c.config.Properties {
		result[name] = config.PropType
	}

	return result
}

func (c *computedPlayerPropertiesImpl) IntProp(name string) (int, error) {
	if val, err := c.bag.IntProp(name); err == nil {
		return val, nil
	}

	definition, ok := c.config[name]

	if !ok {
		return 0, errors.New("No such computed player property")
	}

	if definition.PropType != TypeInt {
		return 0, errors.New("That name is not an intprop")
	}

	//Compute it

	val, err := definition.compute(c.state, c.i)

	if err != nil {
		return 0, errors.New("Error computing calculated int prop: " + err.Error())
	}

	intVal, ok := val.(int)

	if !ok {
		return 0, errors.New("The compute function for that name did not return an int as expected")
	}

	c.bag.SetIntProp(name, intVal)

	return intVal, nil

}

func (c *computedPropertiesImpl) IntProp(name string) (int, error) {
	if val, err := c.bag.IntProp(name); err == nil {
		return val, nil
	}

	definition, ok := c.config.Properties[name]

	if !ok {
		return 0, errors.New("no such computed property")
	}

	if definition.PropType != TypeInt {
		return 0, errors.New("That name is not an IntProp.")
	}

	//Nope, gotta compute it.
	val, err := definition.compute(c.state)

	if err != nil {
		return 0, errors.New("Error computing calculated int prop:" + err.Error())
	}

	intVal, ok := val.(int)

	if !ok {
		return 0, errors.New("The compute function for that name did not return an int as expectd")
	}

	c.bag.SetIntProp(name, intVal)

	return intVal, nil

}

func (c *computedPlayerPropertiesImpl) BoolProp(name string) (bool, error) {
	if val, err := c.bag.BoolProp(name); err == nil {
		return val, nil
	}

	definition, ok := c.config[name]

	if !ok {
		return false, errors.New("No such computed player property")
	}

	if definition.PropType != TypeBool {
		return false, errors.New("That name is not an boolprop")
	}

	//Compute it

	val, err := definition.compute(c.state, c.i)

	if err != nil {
		return false, errors.New("Error computing calculated int prop: " + err.Error())
	}

	boolVal, ok := val.(bool)

	if !ok {
		return false, errors.New("The compute function for that name did not return an bool as expected")
	}

	c.bag.SetBoolProp(name, boolVal)

	return boolVal, nil

}

func (c *computedPropertiesImpl) BoolProp(name string) (bool, error) {
	if val, err := c.bag.BoolProp(name); err == nil {
		return val, nil
	}

	definition, ok := c.config.Properties[name]

	if !ok {
		return false, errors.New("no such computed property")
	}

	if definition.PropType != TypeBool {
		return false, errors.New("That name is not an BoolProp.")
	}

	//Nope, gotta compute it.
	val, err := definition.compute(c.state)

	if err != nil {
		return false, errors.New("Error computing calculated prop:" + err.Error())
	}

	boolVal, ok := val.(bool)

	if !ok {
		return false, errors.New("The compute function for that name did not return a bool as expectd")
	}

	c.bag.SetBoolProp(name, boolVal)

	return boolVal, nil

}

func (c *computedPlayerPropertiesImpl) StringProp(name string) (string, error) {
	if val, err := c.bag.StringProp(name); err == nil {
		return val, nil
	}

	definition, ok := c.config[name]

	if !ok {
		return "", errors.New("No such computed player property")
	}

	if definition.PropType != TypeString {
		return "", errors.New("That name is not an stringProp")
	}

	//Compute it

	val, err := definition.compute(c.state, c.i)

	if err != nil {
		return "", errors.New("Error computing calculated string prop: " + err.Error())
	}

	stringVal, ok := val.(string)

	if !ok {
		return "", errors.New("The compute function for that name did not return an string as expected")
	}

	c.bag.SetStringProp(name, stringVal)

	return stringVal, nil

}

func (c *computedPropertiesImpl) StringProp(name string) (string, error) {
	if val, err := c.bag.StringProp(name); err == nil {
		return val, nil
	}

	definition, ok := c.config.Properties[name]

	if !ok {
		return "", errors.New("no such computed property")
	}

	if definition.PropType != TypeString {
		return "", errors.New("That name is not a stringProp.")
	}

	//Nope, gotta compute it.
	val, err := definition.compute(c.state)

	if err != nil {
		return "", errors.New("Error computing calculated prop:" + err.Error())
	}

	stringVal, ok := val.(string)

	if !ok {
		return "", errors.New("The compute function for that name did not return a string as expectd")
	}

	c.bag.SetStringProp(name, stringVal)

	return stringVal, nil

}

func (c *computedPlayerPropertiesImpl) GrowableStackProp(name string) (*GrowableStack, error) {
	if val, err := c.bag.GrowableStackProp(name); err == nil {
		return val, nil
	}

	definition, ok := c.config[name]

	if !ok {
		return nil, errors.New("No such computed player property")
	}

	if definition.PropType != TypeGrowableStack {
		return nil, errors.New("That name is not an Growable Stack prop")
	}

	//Compute it

	val, err := definition.compute(c.state, c.i)

	if err != nil {
		return nil, errors.New("Error computing calculated growable stack prop: " + err.Error())
	}

	growableStackVal, ok := val.(*GrowableStack)

	if !ok {
		return nil, errors.New("The compute function for that name did not return an growable stack as expected")
	}

	c.bag.SetGrowableStackProp(name, growableStackVal)

	return growableStackVal, nil

}

func (c *computedPropertiesImpl) GrowableStackProp(name string) (*GrowableStack, error) {
	if val, err := c.bag.GrowableStackProp(name); err == nil {
		return val, nil
	}

	definition, ok := c.config.Properties[name]

	if !ok {
		return nil, errors.New("no such computed property")
	}

	if definition.PropType != TypeGrowableStack {
		return nil, errors.New("That name is not an growable stack prop.")
	}

	//Nope, gotta compute it.
	val, err := definition.compute(c.state)

	if err != nil {
		return nil, errors.New("Error computing calculated prop:" + err.Error())
	}

	growableStackVal, ok := val.(*GrowableStack)

	if !ok {
		return nil, errors.New("The compute function for that name did not return a growableStackVal as expectd")
	}

	c.bag.SetGrowableStackProp(name, growableStackVal)

	return growableStackVal, nil

}

func (c *computedPlayerPropertiesImpl) SizedStackProp(name string) (*SizedStack, error) {
	if val, err := c.bag.SizedStackProp(name); err == nil {
		return val, nil
	}

	definition, ok := c.config[name]

	if !ok {
		return nil, errors.New("No such computed player property")
	}

	if definition.PropType != TypeSizedStack {
		return nil, errors.New("That name is not an sized stack prop")
	}

	//Compute it

	val, err := definition.compute(c.state, c.i)

	if err != nil {
		return nil, errors.New("Error computing calculated sized stack prop: " + err.Error())
	}

	sizedStackVal, ok := val.(*SizedStack)

	if !ok {
		return nil, errors.New("The compute function for that name did not return an sized stack as expected")
	}

	c.bag.SetSizedStackProp(name, sizedStackVal)

	return sizedStackVal, nil

}

func (c *computedPropertiesImpl) SizedStackProp(name string) (*SizedStack, error) {
	if val, err := c.bag.SizedStackProp(name); err == nil {
		return val, nil
	}

	definition, ok := c.config.Properties[name]

	if !ok {
		return nil, errors.New("no such computed property")
	}

	if definition.PropType != TypeSizedStack {
		return nil, errors.New("That name is not an sized stack prop.")
	}

	//Nope, gotta compute it.
	val, err := definition.compute(c.state)

	if err != nil {
		return nil, errors.New("Error computing calculated prop:" + err.Error())
	}

	sizedStackVal, ok := val.(*SizedStack)

	if !ok {
		return nil, errors.New("The compute function for that name did not return a sizedStackVal as expectd")
	}

	c.bag.SetSizedStackProp(name, sizedStackVal)

	return sizedStackVal, nil

}

func (c *computedPlayerPropertiesImpl) Prop(name string) (interface{}, error) {
	if val, err := c.bag.Prop(name); err == nil {
		return val, nil
	}

	definition, ok := c.config[name]

	if !ok {
		return nil, errors.New("No such computed property")
	}

	switch definition.PropType {
	case TypeBool:
		return c.BoolProp(name)
	case TypeInt:
		return c.IntProp(name)
	case TypeString:
		return c.StringProp(name)
	case TypeGrowableStack:
		return c.GrowableStackProp(name)
	case TypeSizedStack:
		return c.SizedStackProp(name)
	}

	//If we get to here, it's a TypeUnknown

	val, err := definition.compute(c.state, c.i)

	if err != nil {
		return nil, errors.New("Error computing calculated prop" + err.Error())
	}

	c.bag.SetProp(name, val)

	return val, nil
}

func (c *computedPropertiesImpl) Prop(name string) (interface{}, error) {
	if val, err := c.bag.Prop(name); err == nil {
		return val, nil
	}

	definition, ok := c.config.Properties[name]

	if !ok {
		return nil, errors.New("No such computed property")
	}

	switch definition.PropType {
	case TypeBool:
		return c.BoolProp(name)
	case TypeInt:
		return c.IntProp(name)
	case TypeString:
		return c.StringProp(name)
	case TypeGrowableStack:
		return c.GrowableStackProp(name)
	case TypeSizedStack:
		return c.SizedStackProp(name)
	}

	//If we get to here, it's a TypeUnknown

	val, err := definition.compute(c.state)

	if err != nil {
		return nil, errors.New("Error computing calculated prop" + err.Error())
	}

	c.bag.SetProp(name, val)

	return val, nil
}

func newComputedPropertiesBag() *computedPropertiesBag {
	return &computedPropertiesBag{
		unknownProps:       make(map[string]interface{}),
		intProps:           make(map[string]int),
		boolProps:          make(map[string]bool),
		stringProps:        make(map[string]string),
		growableStackProps: make(map[string]*GrowableStack),
		sizedStackProps:    make(map[string]*SizedStack),
	}
}

func (c *computedPropertiesBag) Props() map[string]PropertyType {
	result := make(map[string]PropertyType)

	//TODO: memoize this

	for key, _ := range c.unknownProps {
		//TODO: shouldn't this be TypeUnknown?
		result[key] = TypeIllegal
	}

	for key, _ := range c.intProps {
		result[key] = TypeInt
	}

	for key, _ := range c.boolProps {
		result[key] = TypeBool
	}

	for key, _ := range c.stringProps {
		result[key] = TypeString
	}

	return result
}

func (c *computedPropertiesBag) GrowableStackProp(name string) (*GrowableStack, error) {
	result, ok := c.growableStackProps[name]

	if !ok {
		return nil, errors.New("No such growable stack prop")
	}

	return result, nil
}

func (c *computedPropertiesBag) SizedStackProp(name string) (*SizedStack, error) {
	result, ok := c.sizedStackProps[name]

	if !ok {
		return nil, errors.New("No such sized stack prop")
	}

	return result, nil
}

func (c *computedPropertiesBag) IntProp(name string) (int, error) {
	result, ok := c.intProps[name]

	if !ok {
		return 0, errors.New("No such int prop")
	}

	return result, nil
}

func (c *computedPropertiesBag) BoolProp(name string) (bool, error) {
	result, ok := c.boolProps[name]

	if !ok {
		return false, errors.New("No such bool prop")
	}

	return result, nil
}

func (c *computedPropertiesBag) StringProp(name string) (string, error) {
	result, ok := c.stringProps[name]

	if !ok {
		return "", errors.New("No such string prop")
	}

	return result, nil
}

func (c *computedPropertiesBag) Prop(name string) (interface{}, error) {
	props := c.Props()

	propType, ok := props[name]

	if !ok {
		return nil, errors.New("No prop with that name")
	}

	switch propType {
	case TypeString:
		return c.StringProp(name)
	case TypeBool:
		return c.BoolProp(name)
	case TypeInt:
		return c.IntProp(name)
	case TypeGrowableStack:
		return c.GrowableStackProp(name)
	case TypeSizedStack:
		return c.SizedStackProp(name)
	}

	val, ok := c.unknownProps[name]

	if !ok {
		return nil, errors.New("No such unknown prop")
	}

	return val, nil
}

func (c *computedPropertiesBag) SetIntProp(name string, value int) error {
	c.intProps[name] = value
	return nil
}

func (c *computedPropertiesBag) SetBoolProp(name string, value bool) error {
	c.boolProps[name] = value
	return nil
}

func (c *computedPropertiesBag) SetStringProp(name string, value string) error {
	c.stringProps[name] = value
	return nil
}

func (c *computedPropertiesBag) SetGrowableStackProp(name string, value *GrowableStack) error {
	c.growableStackProps[name] = value
	return nil
}

func (c *computedPropertiesBag) SetSizedStackProp(name string, value *SizedStack) error {
	c.sizedStackProps[name] = value
	return nil
}

func (c *computedPropertiesBag) SetProp(name string, value interface{}) error {
	c.unknownProps[name] = value
	return nil
}
