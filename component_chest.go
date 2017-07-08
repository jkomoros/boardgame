package boardgame

//TODO: consider making ComponentChest be an interface again (in some cases it
//might be nice to be able to cast the Deck directly to its underlying type to
//minimize later casts)

//Each game has one ComponentChest, which is an immutable set of all
//components in this game, configured into 0 or more Decks. A chest has two
//phases: construction and serving. During consruction, decks may be added but
//non may be retrieved. After consruction decks may be retrieved but not
//added. This helps ensure that Decks always give a consistent view of the
//world.
type ComponentChest struct {
	initialized bool
	deckNames   []string
	decks       map[string]*Deck
	enum        *EnumManager

	manager *GameManager
}

func NewComponentChest() *ComponentChest {
	return &ComponentChest{
		enum: NewEnumManager(),
	}
}

func (c *ComponentChest) Enum() *EnumManager {
	return c.enum
}

func (c *ComponentChest) Manager() *GameManager {
	return c.manager
}

//DeckNames returns all of the valid deck names, if the chest has finished initalization.
func (c *ComponentChest) DeckNames() []string {
	//If it's not finished being initalized then no decks are valid.
	if !c.initialized {
		return nil
	}
	return c.deckNames
}

//Deck returns the deck with a given name, if the chest has finished initalization.
func (c *ComponentChest) Deck(name string) *Deck {
	if !c.initialized {
		return nil
	}
	return c.decks[name]
}

//AddDeck adds a deck with a given name, but only if Freeze() has not yet been called.
func (c *ComponentChest) AddDeck(name string, deck *Deck) {
	//Only add the deck if we haven't finished initalizing
	if c.initialized {
		return
	}
	if c.decks == nil {
		c.decks = make(map[string]*Deck)
	}

	if name == "" {
		name = "NONAMEPROVIDED"
	}

	//Tell the deck that no more items will be added to it.
	deck.finish(c, name)

	c.decks[name] = deck

}

//Finish switches the chest from constructing to serving. Before freeze is
//called, decks may be added but not retrieved. After it is called, decks may
//be retrieved but not added. Finish() is called automatically when a Chest is
//added to a game via SetChest(), but you can call it before then if you'd
//like.
func (c *ComponentChest) Finish() {

	//Check if Finish() has already been called
	if c.initialized {
		return
	}

	c.enum.Finish()

	c.initialized = true

	//Now that no more decks are coming, we can create deckNames once and be
	//done with it.
	c.deckNames = make([]string, len(c.decks))

	i := 0

	for name, _ := range c.decks {
		c.deckNames[i] = name
		i++
	}
}
