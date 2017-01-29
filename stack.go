package boardgame

//Stack is one of the fundamental types in BoardGame. It represents an ordered
//stack of 0 or more Components, all from the same Deck. Each deck has 0 or
//more Stacks based off of it, and together they include all components in
//that deck, with no component residing in more than one stack. Stacks model
//things like a stack of cards, a collection of resource tokens, etc.
type Stack struct {
	//The Game that we're associated with
	Game *Game
	//DeckName is the name of the deck in the game's ComponentChest that this
	//Stack is tied to.
	DeckName string
	//The indexes from the given deck that this stack contains, in order.
	Indexes []int
}

//Len returns the number of items in the stack.
func (s *Stack) Len() int {
	return len(s.Indexes)
}

//ComponentAt fetches the component object representing the n-th object in
//this stack.
func (s *Stack) ComponentAt(index int) *Component {

	if index >= s.Len() || index < 0 {
		return nil
	}

	deck := s.Game.Chest.Deck(s.DeckName)

	if deck == nil {
		return nil
	}

	//We don't need to check that s.Indexes[index] is valid because it was
	//checked when it was set, and Decks are immutable.
	return deck.Components()[s.Indexes[index]]
}

//InsertFront puts the component at index 0 in this stack, moving all other
//items down by one. The Component you insert should not currently be a member
//of any other stacks, to maintain the deck invariant.
func (s *Stack) InsertFront(c *Component) {

	//Based on how Decks and Chests are constructed, we know the components in
	//the chest hae the right gamename, so no need to check.

	if c.Address.Deck != s.DeckName {
		//We can only add items that are in our deck.

		//TODO: communicate an error
		return
	}

	s.Indexes = append([]int{c.Address.Index}, s.Indexes...)
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
