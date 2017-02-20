package blackjack

import (
	"github.com/jkomoros/boardgame"
)

type MoveShuffleDiscardToDraw struct{}

func (m *MoveShuffleDiscardToDraw) Legal(boardgame.State) error {
	return nil
}

func (m *MoveShuffleDiscardToDraw) Apply(state boardgame.State) error {
	return nil
}

func (m *MoveShuffleDiscardToDraw) Copy() boardgame.Move {
	var result MoveShuffleDiscardToDraw
	result = *m
	return &result
}

func (m *MoveShuffleDiscardToDraw) DefaultsForState(state boardgame.State) {
	//TODO: implement
}

func (m *MoveShuffleDiscardToDraw) Name() string {
	return "Shuffle Discard To Draw"
}

func (m *MoveShuffleDiscardToDraw) Description() string {
	return "When the draw deck is empty, shuffles the discard deck into draw deck."
}

func (m *MoveShuffleDiscardToDraw) Props() []string {
	return boardgame.PropertyReaderPropsImpl(m)
}

func (m *MoveShuffleDiscardToDraw) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(m, name)
}

func (m *MoveShuffleDiscardToDraw) SetProp(name string, val interface{}) error {
	return boardgame.PropertySetImpl(m, name, val)
}
