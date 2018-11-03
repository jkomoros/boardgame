package boardgame

import (
	"errors"
	"github.com/jkomoros/boardgame/enum"
	"sort"
)

//TODO: consider making ComponentChest be an interface again (in some cases it
//might be nice to be able to cast the Deck directly to its underlying type to
//minimize later casts)

//ComponentChest is a list of all decks, enums, and constants, for this game
//type. Each game type has one ComponentChest, which is an immutable set of
//all components in this game type. Every game of a given game type uses the
//same ComponentChest. NewGameManager creates a new ComponentChest, filling it
//with values returned from your GameDelegate's ConfigureDecks(),
//ConfigureConstants(), and ConfigureEnums().
type ComponentChest struct {
	initialized bool
	deckNames   []string
	decks       map[string]*Deck
	constants   map[string]interface{}
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

//Enums returns the enum.Set in use in this chest--the set that was returned
//from your GameDelegate's ConfigureEnums().
func (c *ComponentChest) Enums() *enum.Set {
	return c.enums
}

//Manager returns the GameManager that this ComponentChest is associated with.
func (c *ComponentChest) Manager() *GameManager {
	return c.manager
}

//DeckNames returns all of the valid deck names--the names that would return a
//valid deck if passed to Deck().
func (c *ComponentChest) DeckNames() []string {
	//If it's not finished being initalized then no decks are valid.
	if !c.initialized {
		return nil
	}
	return c.deckNames
}

//Deck returns the deck in the chest with the given name. Those Decks() are
//associated with this ComponentChest based on what your GameDelegate returns
//from ConfigureDecks().
func (c *ComponentChest) Deck(name string) *Deck {
	if !c.initialized {
		return nil
	}
	return c.decks[name]
}

//ConstantNames returns all of the names of constants in this chest that would
//return something from Constant().
func (c *ComponentChest) ConstantNames() []string {
	if !c.initialized {
		return nil
	}
	var result []string

	for name := range c.constants {
		result = append(result, name)
	}

	sort.Strings(result)

	return result
}

//Constant returns the constant of that name, that was configured via your
//GameDelegate's ConfigureConstants().
func (c *ComponentChest) Constant(name string) interface{} {
	if !c.initialized {
		return nil
	}
	if c.constants == nil {
		return nil
	}
	return c.constants[name]
}

//AddConstant adds a constant to the chest. Will error if the chest is already
//finished, or if the val is not a bool, int, or string.
func (c *ComponentChest) addConstant(name string, val interface{}) error {
	if c.initialized {
		return errors.New("Couldn't add constant because the chest was already finished")
	}

	if c.constants == nil {
		c.constants = make(map[string]interface{})
	}

	if _, exists := c.constants[name]; exists {
		return errors.New(name + " is already set as a constant")
	}

	switch val.(type) {
	case int:
		//OK
	case bool:
		//OK
	case string:
		//OK
	default:
		return errors.New("Unsupported type. Val must be int, bool, or string")
	}

	c.constants[name] = val

	return nil
}

//AddDeck adds a deck with a given name, but only if Freeze() has not yet been called.
func (c *ComponentChest) addDeck(name string, deck *Deck) error {
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
func (c *ComponentChest) finish() {

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

//MarshalJSON returns all of the decks, connstants, and enums in this chest.
func (c *ComponentChest) MarshalJSON() ([]byte, error) {
	obj := struct {
		Decks     map[string]*Deck
		Enums     *enum.Set
		Constants map[string]interface{}
	}{
		c.decks,
		c.enums,
		c.constants,
	}
	return DefaultMarshalJSON(obj)
}
