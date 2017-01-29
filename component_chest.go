package boardgame

//TODO: consider making ComponentChest be an interface again (in some cases it
//might be nice to be able to cast the Deck directly to its underlying type to
//minimize later casts)

//Each game has one ComponentChest, which is an immutable set of all
//components in this game, configured into 0 or more Decks.
type ComponentChest map[string]*Deck
