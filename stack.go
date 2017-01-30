package boardgame

import (
	"math"
)

//TODO: create a Stack interface that both GrowableStack and SizedStack
//implement.

//TODO: should deck, index, and cap all be private?

//Stack is one of the fundamental types in BoardGame. It represents an ordered
//stack of 0 or more Components, all from the same Deck. Each deck has 0 or
//more Stacks based off of it, and together they include all components in
//that deck, with no component residing in more than one stack. Stacks model
//things like a stack of cards, a collection of resource tokens, etc.
type GrowableStack struct {
	//Deck is the deck that we're a part of.
	Deck *Deck
	//The indexes from the given deck that this stack contains, in order.
	Indexes []int
	//Cap, if set, says the maxmimum number of items allowed in the Stack. 0
	//means that the Stack may grow without bound.
	Cap int
}

//SizedStack is a Stack that has a fixed number of slots, any of which may be
//empty. Create a new one with NewSizedStack.
type SizedStack struct {
	//Deck is the deck we're a part of.
	Deck *Deck
	//Indexes will always have a len of size. Slots that are "empty" will have
	//index of -1.
	Indexes []int
	//Size is the number of slots.
	Size int
}

//NewGrowableStack creates a new growable stack with the given Deck and Cap.
func NewGrowableStack(deck *Deck, max int) *GrowableStack {

	if max < 0 {
		max = 0
	}

	return &GrowableStack{
		Deck:    deck,
		Indexes: nil,
		Cap:     max,
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
		indexes[i] = -1
	}

	return &SizedStack{
		Deck:    deck,
		Indexes: indexes,
		Size:    size,
	}
}

//Len returns the number of items in the stack.
func (s *GrowableStack) Len() int {
	return len(s.Indexes)
}

//Len returns the number of slots in the stack. It will always equal Size.
func (s *SizedStack) Len() int {
	return len(s.Indexes)
}

//ComponentAt fetches the component object representing the n-th object in
//this stack.
func (s *GrowableStack) ComponentAt(index int) *Component {

	//Substantially recreated in SizedStack.ComponentAt()
	if index >= s.Len() || index < 0 {
		return nil
	}

	if s.Deck == nil {
		return nil
	}

	//We don't need to check that s.Indexes[index] is valid because it was
	//checked when it was set, and Decks are immutable.
	return s.Deck.Components()[s.Indexes[index]]
}

//ComponentAt fetches the component object representing the n-th object in
//this stack.
func (s *SizedStack) ComponentAt(index int) *Component {

	//Substantially recreated in GrowableStack.ComponentAt()

	if index >= s.Len() || index < 0 {
		return nil
	}

	if s.Deck == nil {
		return nil
	}

	deckIndex := s.Indexes[index]

	//Check if this is an empty slot
	if deckIndex == -1 {
		return nil
	}

	//We don't need to check that s.Indexes[index] is valid because it was
	//checked when it was set, and Decks are immutable.
	return s.Deck.Components()[deckIndex]
}

//SlotsRemaining returns the count of slots left in this stack. If Cap is 0
//(inifinite) this will be MaxInt64.
func (s *GrowableStack) SlotsRemaining() int {
	if s.Cap <= 0 {
		return math.MaxInt64
	}
	return s.Cap - s.Len()
}

//SlotsRemaining returns the count of unfilled slots in this stack.
func (s *SizedStack) SlotsRemaining() int {
	count := 0
	for _, index := range s.Indexes {
		if index == -1 {
			count++
		}
	}
	return count
}

//InsertFront puts the component at index 0 in this stack, moving all other
//items down by one. The Component you insert should not currently be a member
//of any other stacks, to maintain the deck invariant.
func (s *GrowableStack) InsertFront(c *Component) {

	//Based on how Decks and Chests are constructed, we know the components in
	//the chest hae the right gamename, so no need to check.

	if c.Deck.Name() != s.Deck.Name() {
		//We can only add items that are in our deck.

		//TODO: communicate an error
		return
	}

	if s.SlotsRemaining() < 1 {
		return
	}

	s.Indexes = append([]int{c.DeckIndex}, s.Indexes...)
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
