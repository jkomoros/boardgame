package boardgame

//Each game has one ComponentChest, which is an immutable set of all
//components in this game, configured into 0 or more Decks.
type ComponentChest interface {
	//Deck returns the Deck with the given name
	Deck(name string) Deck
	//Names returns the names of all of the decks in this chest.
	Names() []string
}
