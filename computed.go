package boardgame

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
)

type StatePropertyRef struct {
	Group    StateGroupType
	PropName string
}

//Computed returns the computed properties for this state.
func (s *State) Computed() ComputedProperties {
	//TODO: implement
	return nil
}
