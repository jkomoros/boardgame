package boardgame

//Board represents an array of growable Stacks. They're useful for
//representing spaces on a board, which may allow unlimited components to
//reside in them, or have a maxium number of occupants. If each board's space
//only allows a single item, it's often equivalent--and simpler--to just use a
//single Stack of a FixedSize. Get one from deck.NewBoard(). See also
//MutableBoard, which is the same, but adds Mutators.
type Board interface {
	Spaces() []Stack
	SpaceAt(index int) Stack
	Len() int
}

type MutableBoard interface {
	Board
	MutableSpaces() []MutableStack
	MutableSpaceAt(index int) MutableStack
}

type board struct {
	spaces []*growableStack
}

//NewBoard returns a new board associated with the given deck. length is the
//number of spaces to create. maxSize is the maximum size for each growable
//Stack in the board. 0 means "no limitation". If you pass maxSize of 1,
//consider simply using a sized Stack for that property instead. Boards can be
//created with struct tags as well.
func (d *Deck) NewBoard(length int, maxSize int) MutableBoard {
	if length <= 0 {
		return nil
	}

	spaces := make([]*growableStack, length)

	for i := 0; i < length; i++ {
		spaces[i] = d.NewStack(maxSize).(*growableStack)
	}

	return &board{
		spaces: spaces,
	}
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

func (b *board) MutableSpaces() []MutableStack {
	result := make([]MutableStack, len(b.spaces))

	for i, item := range b.spaces {
		result[i] = item
	}

	return result
}

func (b *board) MutableSpaceAt(index int) MutableStack {
	if index < 0 || index > b.Len() {
		return nil
	}
	return b.spaces[index]
}

func (b *board) Len() int {
	return len(b.spaces)
}
