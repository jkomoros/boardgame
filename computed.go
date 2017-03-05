package boardgame

import (
	"errors"
)

//ComputedProperties represents a collection of compute properties for a given
//state.
type ComputedProperties interface {
	PropertyReader
}

type ComputedPropertiesConfig struct {
	Properties map[string]ComputedPropertyDefinition
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
	Compute      func(shadow *ShadowState) (interface{}, error)
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
	*computedPropertiesBag
	state  *State
	config *ComputedPropertiesConfig
}

type computedPropertiesBag struct {
	unknownProps map[string]interface{}
	intProps     map[string]int
	boolProps    map[string]bool
	stringProps  map[string]string
}

//Computed returns the computed properties for this state.
func (s *State) Computed() ComputedProperties {
	if s.computed == nil {
		config := s.delegate.ComputedPropertiesConfig()
		s.computed = &computedPropertiesImpl{
			newComputedPropertiesBag(),
			s,
			config,
		}
	}
	return s.computed
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
