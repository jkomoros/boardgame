package boardgame

import (
	"encoding/json"
	"errors"
	"math"
)

const emptyIndexSentinel = -1

//Stack is one of the fundamental types in BoardGame. It represents an ordered
//stack of 0 or more Components, all from the same Deck. Each deck has 0 or
//more Stacks based off of it, and together they include all components in
//that deck, with no component residing in more than one stack. Stacks model
//things like a stack of cards, a collection of resource tokens, etc. There
//are two concrete types of Stacks: GrowableStack's and SizedStack's.
type Stack interface {
	//Len returns the number of slots in the Stack. For a GrowableStack this
	//is the number of items in the stack. For SizedStacks, this is the number
	//of slots--even if some are unfilled.
	Len() int
	//ComponentAt retrieves the component at the given index in the stack.
	ComponentAt(index int) *Component
	//SlotsRemaining returns how many slots there are left in this stack to
	//add items.
	SlotsRemaining() int
	//InsertFront inserts the component at the first position in the stack
	//that it can. In GrowableStacks, this will be the front of the stack (if
	//there aren't already MaxLen), and in SizedStacks this will be in the
	//first unfilled slot. The component you insert should not currently be a
	//member of any other Stack, in order to maintain the Deck/Stack
	//invariant.
	InsertFront(c *Component) error

	//RemoveFirst removes the first component in the stack. For GrowableStacks
	//this will always be the first component in the stack. For SizedStacks,
	//this will be the component in the first filled slot. Remember to insert
	//the component in another stack to maintain the Deck/Stack invariant.
	RemoveFirst() *Component
}

type GrowableStack struct {
	//Deck is the deck that we're a part of.
	deck *Deck
	//The indexes from the given deck that this stack contains, in order.
	indexes []int
	//size, if set, says the maxmimum number of items allowed in the Stack. 0
	//means that the Stack may grow without bound.
	maxLen int
}

//SizedStack is a Stack that has a fixed number of slots, any of which may be
//empty. Create a new one with NewSizedStack.
type SizedStack struct {
	//Deck is the deck we're a part of.
	deck *Deck
	//Indexes will always have a len of size. Slots that are "empty" will have
	//index of -1.
	indexes []int
	//Size is the number of slots.
	size int
}

//stackJSONObj is an internal struct that we populate and use to implement
//MarshalJSON so stacks can be saved in output JSON with minimum fuss.
type stackJSONObj struct {
	Deck    string
	Indexes []int
	Size    int `json:",omitempty"`
	MaxLen  int `json:",omitempty"`
}

//NewGrowableStack creates a new growable stack with the given Deck and Cap.
func NewGrowableStack(deck *Deck, maxLen int) *GrowableStack {

	if maxLen < 0 {
		maxLen = 0
	}

	return &GrowableStack{
		deck:    deck,
		indexes: make([]int, 0),
		maxLen:  maxLen,
	}
}

//NewSizedStack creates a new SizedStack for the given deck, with the
//specified size.
func NewSizedStack(deck *Deck, size int) *SizedStack {
	if size < 0 {
		size = 0
	}

	indexes := make([]int, size)

	for i := 0; i < size; i++ {
		indexes[i] = emptyIndexSentinel
	}

	return &SizedStack{
		deck:    deck,
		indexes: indexes,
		size:    size,
	}
}

//Len returns the number of items in the stack.
func (s *GrowableStack) Len() int {
	return len(s.indexes)
}

//Len returns the number of slots in the stack. It will always equal Size.
func (s *SizedStack) Len() int {
	return len(s.indexes)
}

//ComponentAt fetches the component object representing the n-th object in
//this stack.
func (s *GrowableStack) ComponentAt(index int) *Component {

	//Substantially recreated in SizedStack.ComponentAt()
	if index >= s.Len() || index < 0 {
		return nil
	}

	if s.deck == nil {
		return nil
	}

	//We don't need to check that s.Indexes[index] is valid because it was
	//checked when it was set, and Decks are immutable.
	return s.deck.Components()[s.indexes[index]]
}

//ComponentAt fetches the component object representing the n-th object in
//this stack.
func (s *SizedStack) ComponentAt(index int) *Component {

	//Substantially recreated in GrowableStack.ComponentAt()

	if index >= s.Len() || index < 0 {
		return nil
	}

	if s.deck == nil {
		return nil
	}

	deckIndex := s.indexes[index]

	//Check if this is an empty slot
	if deckIndex == emptyIndexSentinel {
		return nil
	}

	//We don't need to check that s.Indexes[index] is valid because it was
	//checked when it was set, and Decks are immutable.
	return s.deck.Components()[deckIndex]
}

//SlotsRemaining returns the count of slots left in this stack. If Cap is 0
//(inifinite) this will be MaxInt64.
func (s *GrowableStack) SlotsRemaining() int {
	if s.maxLen <= 0 {
		return math.MaxInt64
	}
	return s.maxLen - s.Len()
}

//SlotsRemaining returns the count of unfilled slots in this stack.
func (s *SizedStack) SlotsRemaining() int {
	count := 0
	for _, index := range s.indexes {
		if index == emptyIndexSentinel {
			count++
		}
	}
	return count
}

//InsertAtSlot inserts the given component at the specified slot in the stack,
//as long as that slot is not currently occupied.
func (s *SizedStack) InsertAtSlot(c *Component, index int) error {
	//Based on how Decks and Chests are constructed, we know the components in
	//the chest hae the right gamename, so no need to check.

	if c.Deck.Name() != s.deck.Name() {
		//We can only add items that are in our deck.

		return errors.New("The component is not part of this stack's deck.")
	}

	if index > s.Len() || index < 0 {
		return errors.New("The index is not valid")
	}

	if s.indexes[index] != emptyIndexSentinel {
		//That slot is taken!
		return errors.New("That slot is taken!")
	}

	s.indexes[index] = c.DeckIndex

	return nil
}

//InsertFront inserts the component in the first slot that is empty.
func (s *SizedStack) InsertFront(c *Component) error {

	//TODO: shouldn't this just be InsertFront, and then we pop it into the
	//Stack interface?

	//Based on how Decks and Chests are constructed, we know the components in
	//the chest hae the right gamename, so no need to check.

	if c.Deck.Name() != s.deck.Name() {
		//We can only add items that are in our deck.

		return errors.New("The component is not part of this stack's deck.")
	}

	if s.SlotsRemaining() < 1 {
		return errors.New("There are no available slots.")
	}

	for i, index := range s.indexes {
		if index == emptyIndexSentinel {
			//Found it!
			s.indexes[i] = c.DeckIndex
			return nil
		}
	}

	return nil
}

//InsertFront puts the component at index 0 in this stack, moving all other
//items down by one. The Component you insert should not currently be a member
//of any other stacks, to maintain the deck invariant.
func (s *GrowableStack) InsertFront(c *Component) error {

	//Based on how Decks and Chests are constructed, we know the components in
	//the chest hae the right gamename, so no need to check.

	if c.Deck.Name() != s.deck.Name() {
		//We can only add items that are in our deck.

		return errors.New("The component is not part of this stack's deck.")
	}

	if s.SlotsRemaining() < 1 {
		return errors.New("There's no more room in the stack.")
	}

	s.indexes = append([]int{c.DeckIndex}, s.indexes...)

	return nil
}

func (s *SizedStack) RemoveFirst() *Component {
	for i, index := range s.indexes {
		if index != emptyIndexSentinel {
			//Found it!
			component := s.ComponentAt(i)
			s.indexes[i] = emptyIndexSentinel
			return component
		}
	}
	//didn't find any.
	return nil
}

func (g *GrowableStack) RemoveFirst() *Component {
	if len(g.indexes) == 0 {
		return nil
	}
	component := g.ComponentAt(0)

	g.indexes = g.indexes[1:]
	return component

}

func (g *GrowableStack) MarshalJSON() ([]byte, error) {
	obj := &stackJSONObj{
		Deck:    g.deck.Name(),
		Indexes: g.indexes,
		MaxLen:  g.maxLen,
	}
	return json.Marshal(obj)
}

func (s *SizedStack) MarshalJSON() ([]byte, error) {
	//TODO: test this, including Size
	obj := &stackJSONObj{
		Deck:    s.deck.Name(),
		Indexes: s.indexes,
		Size:    s.size,
	}
	return json.Marshal(obj)
}

/*

//InsertBack puts the component at the last index in this stack. The
//Component you insert should not currently be a member of any other stacks,
//to maintain the deck invariant.
func (s *Stack) InsertBack(c Component) {

}

//RemoveFront removes the component from the first slot in this stack,
//shifting all later components down by 1. You should then insert the
//component in another stack to maintain the deck invariant.
func (s *Stack) RemoveFront() Component {

}

//RemoveBack removes the component from the last slot in this stack. You
//should then insert the component in another stack to maintain the deck
//invariant.
func (s *Stack) RemoveBack() Component {

}

*/
