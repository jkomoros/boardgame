package boardgame

//Stack is one of the fundamental types in BoardGame. It represents an ordered
//stack of 0 or more Components, all from the same Deck. Each deck has 0 or
//more Stacks based off of it, and together they include all components in
//that deck, with no component residing in more than one stack. Stacks model
//things like a stack of cards, a collection of resource tokens, etc.
type Stack struct {
	//DeckName is the name of the deck in the game's ComponentChest that this
	//Stack is tied to.
	DeckName string
	//The indexes from the given deck that this stack contains, in order.
	Indexes []int
}
