package boardgame

//TODO: consider making Deck be an interface again (in some cases it
//might be nice to be able to cast the Deck directly to its underlying type to
//minimize later casts)

//A Deck represents an immutable collection of a certain type of components.
//Every component lives in one deck. 1 or more Stacks index into every Deck,
//and cover every item in the deck, with no items in more than one deck.
type Deck struct {
	Name string
	//Components should only ever be added at initalization time. After
	//initalization, Components should be read-only.
	Components []Component
}
