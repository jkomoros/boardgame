package base

import (
	"github.com/jkomoros/boardgame"
)

//go:generate boardgame-util codegen

type isFixUpper interface {
	IsFixUp() bool
}

//IsFixUp is a convenience method that takes the given move and returns
//whehter its IsFixUp method returns true. If no IsFixUp exists, will return
//false. Used by base.GameDelegate, since IsFixUp() isn't defined in the core
//library, which means that moves fetched via the GameManager will have to be
//casted to an interface.
func IsFixUp(move boardgame.Move) bool {

	fixUpper, ok := move.(isFixUpper)
	if !ok {
		return false
	}

	return fixUpper.IsFixUp()

}

/*
Move is an optional, convenience struct designed to be embedded anonymously
in your own Moves. Although technically optional to embed this struct, in
almost all cases you will embed it or a move that transitively embeds it.

It provides minimal stubs for all of the expected methods Moves should have,
other than Legal() and Apply(), which generally have logic specific to the
move.

See also moves.Default, which embeds this but adds logic about overriding
configuration via auto.Config(), as well as robust base logic for phases and
phase progressions. Typically your moves will use that (or something that
embeds that).

boardgame:codegen
*/
type Move struct {
	info           *boardgame.MoveInfo
	topLevelStruct boardgame.Move
}

func (m *Move) SetInfo(info *boardgame.MoveInfo) {
	m.info = info
}

//Type simply returns BaseMove.MoveInfo
func (m *Move) Info() *boardgame.MoveInfo {
	return m.info
}

func (m *Move) SetTopLevelStruct(t boardgame.Move) {
	m.topLevelStruct = t
}

//TopLevelStruct returns the object that was set via SetTopLevelStruct.
func (m *Move) TopLevelStruct() boardgame.Move {
	return m.topLevelStruct
}

//DefaultsForState doesn't do anything
func (m *Move) DefaultsForState(state boardgame.ImmutableState) {
	return
}

//Name returns the name of this move according to MoveInfo.Name(). A simple
//convenience wrapper that allows you to avoid a nil check.
func (m *Move) Name() string {
	if m.info == nil {
		return ""
	}
	return m.info.Name()
}

//CustomConfiguration returns the custom configuration associated with this
//move, according to MoveInfo.CustomConfiguration(). A simple convenience
//wrapper that allows you to avoid a nil check.
func (m *Move) CustomConfiguration() boardgame.PropertyCollection {

	if m.info == nil {
		return nil
	}

	return m.info.CustomConfiguration()

}

//HelpText returns ""
func (m *Move) HelpText() string {
	return ""
}

//IsFixUp always returns false; it's designed ot be overriden. It is designed
//to work well with base.IsFixUp, for use in base.GameDelegate.ProposeFixUp.
func (m *Move) IsFixUp() bool {
	return false
}
