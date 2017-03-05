package boardgame

import (
	"encoding/json"
	"errors"
	"strconv"
)

//ComputedProperties represents a collection of compute properties for a given
//state.
type ComputedProperties interface {
	PropertyReader
	Player(index int) PropertyReader
	MarshalJSON() ([]byte, error)
}

type ComputedPropertiesConfig struct {
	Properties       map[string]ComputedPropertyDefinition
	PlayerProperties map[string]ComputedPlayerPropertyDefinition
}

type ShadowPlayerState struct {
	PropertyReader
}

//ShadowState is an object roughly shaped like a State, but where instead of
//underlying types it has PropertyReaders. Passed in to the Compute method of
//a ComputedProperty, based on the dependencies they define.
type ShadowState struct {
	Game    PropertyReader
	Players []*ShadowPlayerState
}

type ComputedPropertyDefinition struct {
	Dependencies []StatePropertyRef
	//The thing we expect to be able to cast the result of Compute to.
	PropType PropertyType
	Compute  func(shadow *ShadowState) (interface{}, error)
}

type ComputedPlayerPropertyDefinition struct {
	//Only StateGroupPlayer group is valid for these
	Dependencies []StatePropertyRef
	PropType     PropertyType
	Compute      func(shadow *ShadowPlayerState) (interface{}, error)
}

type StateGroupType int

const (
	StateGroupGame StateGroupType = iota
	StateGroupPlayer
)

type StatePropertyRef struct {
	Group    StateGroupType
	PropName string
}

//The private impl for ComputedProperties
type computedPropertiesImpl struct {
	bag        *computedPropertiesBag
	playerBags []*computedPropertiesBag
	state      *State
	config     *ComputedPropertiesConfig
}

type computedPropertiesBag struct {
	unknownProps map[string]interface{}
	intProps     map[string]int
	boolProps    map[string]bool
	stringProps  map[string]string
}

func (c *ComputedPropertyDefinition) compute(state *State) (interface{}, error) {

	//First, prepare a shadow state with all of the dependencies.

	players := make([]*ShadowPlayerState, len(state.Players))

	for i := 0; i < len(state.Players); i++ {
		players[i] = &ShadowPlayerState{newComputedPropertiesBag()}
	}

	shadow := &ShadowState{
		Game:    newComputedPropertiesBag(),
		Players: players,
	}

	for _, dependency := range c.Dependencies {
		shadow.addDependency(state, dependency)
	}

	return c.Compute(shadow)

}

func (s *ShadowState) addDependency(state *State, ref StatePropertyRef) error {

	if ref.Group == StateGroupGame {
		return s.addGameDependency(state, ref.PropName)
	}

	if ref.Group == StateGroupPlayer {
		return s.addPlayerDependency(state, ref.PropName)
	}

	return errors.New("Unsupoorted Ref.Group")

}

func (s *ShadowState) addGameDependency(state *State, propName string) error {
	reader := state.Game.Reader()
	//TODO: this is hacky
	bag := s.Game.(*computedPropertiesBag)

	return s.addDependencyHelper(propName, reader, bag)

}

func (s *ShadowState) addDependencyHelper(propName string, reader PropertyReader, bag *computedPropertiesBag) error {
	props := reader.Props()

	propType, ok := props[propName]

	if !ok {
		return errors.New("No such property on state game")
	}

	switch propType {
	case TypeInt:
		if val, err := reader.IntProp(propName); err == nil {
			bag.SetIntProp(propName, val)
		} else {
			return errors.New("Error reading int prop" + err.Error())
		}
	case TypeBool:
		if val, err := reader.BoolProp(propName); err == nil {
			bag.SetBoolProp(propName, val)
		} else {
			return errors.New("Error reading bool prop" + err.Error())
		}
	case TypeString:
		if val, err := reader.StringProp(propName); err == nil {
			bag.SetStringProp(propName, val)
		} else {
			return errors.New("Error reading string prop" + err.Error())
		}
	case TypeGrowableStack:
		if val, err := reader.GrowableStackProp(propName); err == nil {
			bag.SetGrowableStackProp(propName, val)
		} else {
			return errors.New("Error reading growable stack prop" + err.Error())
		}
	case TypeSizedStack:
		if val, err := reader.SizedStackProp(propName); err == nil {
			bag.SetSizedStackProp(propName, val)
		} else {
			return errors.New("Error reading sized stack prop" + err.Error())
		}
	default:
		if val, err := reader.Prop(propName); err == nil {
			bag.SetProp(propName, val)
		} else {
			return errors.New("Error reading unknown prop" + err.Error())
		}
	}

	return nil
}

func (s *ShadowState) addPlayerDependency(state *State, propName string) error {

	for i, player := range state.Players {

		reader := player.Reader()
		//TODO: this is hacky
		bag := s.Players[i].PropertyReader.(*computedPropertiesBag)

		if err := s.addDependencyHelper(propName, reader, bag); err != nil {
			return errors.New("Error on " + strconv.Itoa(i) + ": " + err.Error())
		}
	}

	return nil

}

func (c *computedPropertiesImpl) MarshalJSON() ([]byte, error) {

	result := make(map[string]interface{})

	for propName, _ := range c.Props() {
		val, err := c.Prop(propName)

		if err != nil {
			continue
		}

		result[propName] = val
	}

	return json.Marshal(result)
}

func (c *computedPropertiesImpl) Player(index int) PropertyReader {
	return c.playerBags[index]
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
		unknownProps: make(map[string]interface{}),
		intProps:     make(map[string]int),
		boolProps:    make(map[string]bool),
		stringProps:  make(map[string]string),
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
	//We don't (yet?) support growable stack computed props
	return nil, errors.New("No such growable stack prop")
}

func (c *computedPropertiesBag) SizedStackProp(name string) (*SizedStack, error) {
	//We don't (yet?) support SizedStackProps.
	return nil, errors.New("No such sized stack prop")
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
	return errors.New("We don't currently support growable stacks")
}

func (c *computedPropertiesBag) SetSizedStackProp(name string, value *SizedStack) error {
	return errors.New("We don't currently support sized stacks")
}

func (c *computedPropertiesBag) SetProp(name string, value interface{}) error {
	c.unknownProps[name] = value
	return nil
}
