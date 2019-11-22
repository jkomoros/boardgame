/*

Package base contains a number of base classes for common objects in boardgame.
Technically all of these base objects are fully optional, but in practice almost
every game will use them (or a class that embeds them).

*/
package base

import (
	"github.com/jkomoros/boardgame"
)

//SubState is a simple struct designed to be anonymously embedded in the
//boardgame.SubStates you create, so you don't have to implement SetState yourself.
type SubState struct {
	state boardgame.State
	ref   boardgame.StatePropertyRef
}

//ConnectContainingState sets the State to the given state and the ref to the
//given ref.
func (s *SubState) ConnectContainingState(state boardgame.State, ref boardgame.StatePropertyRef) {
	s.state = state
	s.ref = ref
}

//State returns the state set with SetState.
func (s *SubState) State() boardgame.State {
	return s.state
}

//ImmutableState returns the immutablestate set via SetImmutableState.
func (s *SubState) ImmutableState() boardgame.ImmutableState {
	return s.state
}

//StatePropertyRef returns the ref passed via SetStatePropertyRef()
func (s *SubState) StatePropertyRef() boardgame.StatePropertyRef {
	return s.ref
}

//PlayerIndex is a conveniece warpper that returns
//StatePropertyRef().PlayerIndex. Only really useful for when the SubState is of
//type PlayerState.
func (s *SubState) PlayerIndex() boardgame.PlayerIndex {
	return s.ref.PlayerIndex
}

//FinishStateSetUp doesn't do anything. Typically if you embed a Connectable behavior, you
//override this and call ConnectBehavior() within it.
func (s *SubState) FinishStateSetUp() {
	//pass
}

//ComponentValues is an optional convenience struct designed to be embedded
//anoymously in your component values to implement
//boardgame.ContainingComponent() and boardgame.SetContainingComponent()
//automatically.
type ComponentValues struct {
	c boardgame.Component
}

//ContainingComponent returns the component set via SetContainingComponent.
func (v *ComponentValues) ContainingComponent() boardgame.Component {
	return v.c
}

//SetContainingComponent sets the return value of ContainingComponent.
func (v *ComponentValues) SetContainingComponent(c boardgame.Component) {
	v.c = c
}
