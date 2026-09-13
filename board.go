package boardgame

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/jkomoros/boardgame/enum"
)

// ImmutableBoard is a version of a Board without any of the mutator methods.
// See Board for more.
type ImmutableBoard interface {
	ImmutableSpaces() []ImmutableStack
	ImmutableSpaceAt(index int) ImmutableStack
	// Enum returns the enum that identifies the board's spaces, or nil when
	// this board is indexed only by position.
	Enum() enum.Enum
	// ImmutableSpaceAtKey returns the space identified by key. It returns nil
	// when this board has no associated enum or key is not part of that enum.
	ImmutableSpaceAtKey(key enum.EnumKey) ImmutableStack
	Len() int
	state() *state
	setState(st *state)
}

// Board represents an array of growable Stacks. They're useful for
// representing spaces on a board, which may allow unlimited components to
// reside in them, or have a maxium number of occupants. If each board's space
// only allows a single item, it's often equivalent--and simpler--to just use a
// single Stack of a FixedSize. Get one from deck.NewBoard(). See also
// ImmutableBiard, which is the same, but without mutator methods.
type Board interface {
	ImmutableBoard
	Spaces() []Stack
	SpaceAt(index int) Stack
	// SpaceAtKey is the mutable counterpart of ImmutableSpaceAtKey.
	SpaceAtKey(key enum.EnumKey) Stack

	applySanitizationPolicy(policy Policy)
	//Used to copy from other boards. See mutableStack.importFrom for more about how these work.
	importFrom(other ImmutableBoard) error
}

type board struct {
	spaces []*growableStack
	enum   enum.Enum
}

// NewBoard returns a new board associated with the given deck. length is the
// number of spaces to create. maxSize is the maximum size for each growable
// Stack in the board. 0 means "no limitation". If you pass maxSize of 1,
// consider simply using a sized Stack for that property instead, as those are
// semantically equivalent, and a sized Stack is simpler. Typically you'd use
// this in your GameDelegate's GameStateConstructor() and other similar
// methods; although in practice it is much more common to use struct-tag based
// inflation, making direct use of this constructor unnecessary. See
// StructInflater for more.
func (d *Deck) NewBoard(length int, maxSize int) Board {
	return d.newBoard(length, maxSize, nil)
}

// NewBoardForEnum returns a board with one space for each value in the enum.
// Enum keys need not be contiguous: the enum's ascending Values order defines
// the stable positional order used by Spaces and persisted state.
func (d *Deck) NewBoardForEnum(theEnum enum.Enum, maxSize int) Board {
	if theEnum == nil {
		return nil
	}
	return d.newBoard(len(theEnum.Values()), maxSize, theEnum)
}

func (d *Deck) newBoard(length int, maxSize int, theEnum enum.Enum) Board {
	if length <= 0 {
		return nil
	}

	spaces := make([]*growableStack, length)

	board := &board{enum: theEnum}

	for i := 0; i < length; i++ {
		gStack := d.NewStack(maxSize).(*growableStack)
		gStack.board = board
		gStack.boardIndex = i
		spaces[i] = gStack
	}

	board.spaces = spaces

	return board
}

func (b *board) Enum() enum.Enum {
	return b.enum
}

func (b *board) indexForKey(key enum.EnumKey) int {
	if b.enum == nil || !b.enum.Valid(key) {
		return -1
	}
	for i, candidate := range b.enum.Values() {
		if candidate == key {
			return i
		}
	}
	return -1
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
	if index < 0 || index >= b.Len() {
		return nil
	}
	return b.spaces[index]
}

func (b *board) ImmutableSpaceAtKey(key enum.EnumKey) ImmutableStack {
	return b.ImmutableSpaceAt(b.indexForKey(key))
}

func (b *board) Spaces() []Stack {
	result := make([]Stack, len(b.spaces))

	for i, item := range b.spaces {
		result[i] = item
	}

	return result
}

func (b *board) SpaceAt(index int) Stack {
	if index < 0 || index >= b.Len() {
		return nil
	}
	return b.spaces[index]
}

func (b *board) SpaceAtKey(key enum.EnumKey) Stack {
	return b.SpaceAt(b.indexForKey(key))
}

func (b *board) Len() int {
	return len(b.spaces)
}

type boardJSONObj struct {
	Spaces []json.RawMessage
	Enum   string   `json:",omitempty"`
	Keys   []string `json:",omitempty"`
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
	if b.enum != nil {
		obj.Enum = b.enum.Name()
		for _, key := range b.enum.Values() {
			obj.Keys = append(obj.Keys, b.enum.String(key))
		}
	}

	return json.Marshal(obj)

}

func (b *board) UnmarshalJSON(blob []byte) error {
	obj := &boardJSONObj{}

	if err := json.Unmarshal(blob, obj); err != nil {
		return err
	}
	if len(obj.Spaces) != len(b.spaces) {
		return errors.New("board has " + strconv.Itoa(len(obj.Spaces)) + " persisted spaces, want " + strconv.Itoa(len(b.spaces)))
	}
	if obj.Enum != "" && (b.enum == nil || obj.Enum != b.enum.Name()) {
		return errors.New("board persisted enum " + strconv.Quote(obj.Enum) + " does not match configured board enum")
	}

	for i, blob := range obj.Spaces {
		if err := b.spaces[i].UnmarshalJSON(blob); err != nil {
			return err
		}
	}

	return nil
}
