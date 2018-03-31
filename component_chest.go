package boardgame

import (
	"errors"
	"github.com/jkomoros/boardgame/enum"
)

//TODO: consider making ComponentChest be an interface again (in some cases it
//might be nice to be able to cast the Deck directly to its underlying type to
//minimize later casts)

//ComponentChest is a list of all decks for this game type. Each game has one
//ComponentChest, which is an immutable set of all components in this game,
//configured into 0 or more Decks. A chest has two phases: construction and
//serving. During consruction, decks may be added but non may be retrieved.
//After consruction decks may be retrieved but not added. This helps ensure
//that Decks always give a consistent view of the world. You do not create
//ComponentChests yourself; they are created when a new GameManager is created
//and populated based on what the GameDelegate returns for ConfigureEnums and
//ConfigureDecks().
type ComponentChest struct {
	initialized bool
	deckNames   []string
	decks       map[string]*Deck
	enums       *enum.Set

	manager *GameManager
}

//NewComponentChest returns a new ComponentChest with the given enumset. If no
//enumset is provided, an empty one will be created. Calls Finish() on the
//enumset to verify that it cannot be modified.
func newComponentChest(enums *enum.Set) *ComponentChest {
	//TODO: now that component chest constructor can't be called outside this
	//package, a lot of the init nonsense can be gotten rid of.
	if enums == nil {
		enums = enum.NewSet()
	}
	enums.Finish()
	return &ComponentChest{
		enums: enums,
	}
}

//Enums returns the enum.Set in use in this chest.
func (c *ComponentChest) Enums() *enum.Set {
	return c.enums
}

//Manager returns the GameManager that is associated with this ComponentChest.
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
func (c *ComponentChest) AddDeck(name string, deck *Deck) error {
	//Only add the deck if we haven't finished initalizing
	if c.initialized {
		return errors.New("The chest was already finished, so no new decks may be added.")
	}
	if c.decks == nil {
		c.decks = make(map[string]*Deck)
	}

	if name == "" {
		name = "NONAMEPROVIDED"
	}

	if _, ok := c.decks[name]; ok {
		return errors.New("A deck with name " + name + " was already in the deck.")
	}

	//Tell the deck that no more items will be added to it.
	if err := deck.finish(c, name); err != nil {
		return errors.New("Couldn't finish deck: " + err.Error())
	}

	c.decks[name] = deck

	return nil

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

	c.initialized = true

	//Now that no more decks are coming, we can create deckNames once and be
	//done with it.
	c.deckNames = make([]string, len(c.decks))

	i := 0

	for name := range c.decks {
		c.deckNames[i] = name
		i++
	}
}

func (c *ComponentChest) MarshalJSON() ([]byte, error) {
	obj := struct {
		Decks map[string]*Deck
	}{
		c.decks,
	}
	return DefaultMarshalJSON(obj)
}
