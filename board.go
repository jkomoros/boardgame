package boardgame

import (
	"encoding/json"
	"errors"
	"strconv"
)

//ImmutableBoard represents an array of growable Stacks. They're useful for
//representing spaces on a board, which may allow unlimited components to
//reside in them, or have a maxium number of occupants. If each board's space
//only allows a single item, it's often equivalent--and simpler--to just use a
//single Stack of a FixedSize. Get one from deck.NewBoard(). See also Board,
//which is the same, but adds mutator methods.
type ImmutableBoard interface {
	ImmutableSpaces() []ImmutableStack
	ImmutableSpaceAt(index int) ImmutableStack
	Len() int
	state() *state
	setState(st *state)
}

//Board represents a mutable array of growable stacks. See the documentation
//for ImmutableBoard for more.
type Board interface {
	ImmutableBoard
	Spaces() []Stack
	SpaceAt(index int) Stack

	applySanitizationPolicy(policy Policy)
	//Used to copy from other boards. See mutableStack.importFrom for more about how these work.
	importFrom(other ImmutableBoard) error
}

type board struct {
	spaces []*growableStack
}

//NewBoard returns a new board associated with the given deck. length is the
//number of spaces to create. maxSize is the maximum size for each growable
//Stack in the board. 0 means "no limitation". If you pass maxSize of 1,
//consider simply using a sized Stack for that property instead, as those are
//semantically equivalent, and a sized Stack is simpler. Typically you'd use
//this in your GameDelegate's GameStateConstructor() and other similar
//methods; although in practice it is much more common to use struct-tag based
//inflation, making direct use of this constructor unnecessary. See
//StructInflater for more.
func (d *Deck) NewBoard(length int, maxSize int) Board {
	if length <= 0 {
		return nil
	}

	spaces := make([]*growableStack, length)

	board := &board{}

	for i := 0; i < length; i++ {
		gStack := d.NewStack(maxSize).(*growableStack)
		gStack.board = board
		gStack.boardIndex = i
		spaces[i] = gStack
	}

	board.spaces = spaces

	return board
}

func (b *board) setState(st *state) {
	for _, stack := range b.spaces {
		stack.setState(st)
	}
}

func (b *board) state() *state {
	if len(b.spaces) < 1 {
		return nil
	}
	return b.spaces[0].state()
}

func (b *board) importFrom(other ImmutableBoard) error {

	otherB, ok := other.(*board)

	if !ok {
		return errors.New("other isn't a board")
	}

	for i, space := range b.spaces {
		if err := space.importFrom(otherB.spaces[i]); err != nil {
			return errors.New("Couldn't import, stack " + strconv.Itoa(i) + " errored: " + err.Error())
		}
	}

	return nil
}

func (b *board) ImmutableSpaces() []ImmutableStack {
	result := make([]ImmutableStack, len(b.spaces))

	for i, item := range b.spaces {
		result[i] = item
	}

	return result
}

func (b *board) ImmutableSpaceAt(index int) ImmutableStack {
	if index < 0 || index > b.Len() {
		return nil
	}
	return b.spaces[index]
}

func (b *board) Spaces() []Stack {
	result := make([]Stack, len(b.spaces))

	for i, item := range b.spaces {
		result[i] = item
	}

	return result
}

func (b *board) SpaceAt(index int) Stack {
	if index < 0 || index > b.Len() {
		return nil
	}
	return b.spaces[index]
}

func (b *board) Len() int {
	return len(b.spaces)
}

type boardJSONObj struct {
	Spaces []json.RawMessage
}

func (b *board) MarshalJSON() ([]byte, error) {
	spaces := make([]json.RawMessage, len(b.spaces))
	for i, space := range b.spaces {
		blob, err := json.Marshal(space)
		if err != nil {
			return nil, err
		}
		spaces[i] = blob
	}

	obj := &boardJSONObj{
		Spaces: spaces,
	}

	return json.Marshal(obj)

}

func (b *board) UnmarshalJSON(blob []byte) error {
	obj := &boardJSONObj{}

	if err := json.Unmarshal(blob, obj); err != nil {
		return err
	}

	for i, blob := range obj.Spaces {
		if err := b.spaces[i].UnmarshalJSON(blob); err != nil {
			return err
		}
	}

	return nil
}
