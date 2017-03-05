package boardgame

//ComputedProperties represents a collection of compute properties for a given
//state.
type ComputedProperties interface {
	PropertyReader
	Player(index int) PropertyReader
}

type ComputedPropertiesConfig struct {
	Properties map[string]ComputedPropertyDefinition
}

type ShadowPlayerState struct {
	PropertyReader
	Computed PropertyReader
}

type ShadowState struct {
	Game     PropertyReader
	Players  []*ShadowPlayerState
	Computed PropertyReader
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
