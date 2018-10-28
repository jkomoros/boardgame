/*

base contains a number of base classes for common objects in boardgame.
Technically all of these base objects are fully optional, but in practice
almost every game will use them (or a class that embeds them).

*/
package base

import (
	"github.com/jkomoros/boardgame"
)

//SubState is a simple struct designed to be anonymously embedded in the
//boardgame.SubStates you create, so you don't have to implement SetState yourself.
type SubState struct {
	//Ugh it's really annoying to have to hold onto the same state in two
	//references...
	immutableState boardgame.ImmutableState
	state          boardgame.State
}

func (s *SubState) SetState(state boardgame.State) {
	s.state = state
}

func (s *SubState) State() boardgame.State {
	return s.state
}

func (s *SubState) SetImmutableState(state boardgame.ImmutableState) {
	s.immutableState = state
}

func (s *SubState) ImmutableState() boardgame.ImmutableState {
	return s.immutableState
}

//ComponentValues is an optional convenience struct designed to be embedded
//anoymously in your component values to implement
//boardgame.ContainingComponent() and boardgame.SetContainingComponent()
//automatically.
type ComponentValues struct {
	c boardgame.Component
}

func (v *ComponentValues) ContainingComponent() boardgame.Component {
	return v.c
}

func (v *ComponentValues) SetContainingComponent(c boardgame.Component) {
	v.c = c
}
