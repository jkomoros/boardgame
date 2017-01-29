package boardgame

//TODO: consider making Deck be an interface again (in some cases it
//might be nice to be able to cast the Deck directly to its underlying type to
//minimize later casts)

//A Deck represents an immutable collection of a certain type of components.
//Every component lives in one deck. 1 or more Stacks index into every Deck,
//and cover every item in the deck, with no items in more than one deck. The
//zero-value of Deck is useful. The Deck will not return items until it has
//been added to a ComponentChest, which helps enforce that Decks' values never
//change.
type Deck struct {
	//Name is only set when it's added to the component chest.
	name string
	//Components should only ever be added at initalization time. After
	//initalization, Components should be read-only.
	components []*Component
}

//AddComponent adds the component to the next spot in the deck. If the deck
//has already been added to a componentchest, this will do nothing.
func (d *Deck) AddComponent(c *Component) {
	if d.name != "" {
		return
	}
	//We don't know our name yet. The component name will be set when the deck is added to the chest.
	c.Address.Index = len(d.components)
	d.components = append(d.components, c)
}

//Components returns a list of Components in order in this deck, but only if
//this Deck has already been added to its ComponentChest.
func (d *Deck) Components() []*Component {
	if d.name == "" {
		return nil
	}
	return d.components
}

//finish is called when the deck is added to a component chest. It signifies that no more items may be added.
func (d *Deck) finish(name string) {
	//If a deck has a name, it cannot receive any more items.
	d.name = name
}
