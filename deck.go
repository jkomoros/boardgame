package boardgame

//TODO: is there a reason Deck isn't just a Struct?

//A Deck represents an immutable collection of a certain type of components.
//Every component lives in one deck. 1 or more Stacks index into every Deck,
//and cover every item in the deck, with no items in more than one deck.
type Deck interface {
	//The name of the deck we are in the ComponentChest for this game.
	Name() string
	//The number of components in this deck
	Len() int
	//The component at the given index in this deck
	ComponentAt(index int) Component
}
