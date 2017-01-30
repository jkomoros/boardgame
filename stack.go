package boardgame

import (
	"math"
)

//TODO: create a Stack interface that both GrowableStack and SizedStack
//implement.

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

//Len returns the number of items in the stack.
func (s *GrowableStack) Len() int {
	return len(s.Indexes)
}

//ComponentAt fetches the component object representing the n-th object in
//this stack.
func (s *GrowableStack) ComponentAt(index int) *Component {

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

//SlotsRemaining returns the count of slots left in this stack. If Cap is 0
//(inifinite) this will be MaxInt64.
func (s *GrowableStack) SlotsRemaining() int {
	if s.Cap <= 0 {
		return math.MaxInt64
	}
	return s.Cap - s.Len()
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
