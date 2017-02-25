package boardgame

//TODO: consider making Deck be an interface again (in some cases it
//might be nice to be able to cast the Deck directly to its underlying type to
//minimize later casts)

//A Deck represents an immutable collection of a certain type of components.
//Every component lives in one deck. 1 or more Stacks index into every Deck,
//and cover every item in the deck, with no items in more than one deck. The
//zero-value of Deck is useful. The Deck will not return items until it has
//been added to a ComponentChest, which helps enforce that Decks' values never
//change. Create a new Deck with NewDeck()
type Deck struct {
	chest *ComponentChest
	//Name is only set when it's added to the component chest.
	name string
	//Components should only ever be added at initalization time. After
	//initalization, Components should be read-only.
	components             []*Component
	vendedShadowComponents map[int]*Component
	//TODO: protect shadowComponents cache with mutex to make threadsafe.
}

func NewDeck() *Deck {
	return &Deck{
		vendedShadowComponents: make(map[int]*Component),
	}
}

//AddComponent adds a new component with the given values to the next spot in
//the deck. If the deck has already been added to a componentchest, this will
//do nothing.
func (d *Deck) AddComponent(v ComponentValues) {
	if d.chest != nil {
		return
	}

	c := &Component{
		Deck:      d,
		DeckIndex: len(d.components),
		Values:    v,
	}

	d.components = append(d.components, c)
}

//AddComponentMulti is like AddComponent, but creates multiple versions of the
//same component. The exact same ComponentValues will be re-used, which is
//reasonable becasue components are read-only anyway.
func (d *Deck) AddComponentMulti(v ComponentValues, count int) {
	for i := 0; i < count; i++ {
		d.AddComponent(v)
	}
}

//Components returns a list of Components in order in this deck, but only if
//this Deck has already been added to its ComponentChest.
func (d *Deck) Components() []*Component {
	if d.chest == nil {
		return nil
	}
	return d.components
}

//Chest points back to the chest we're part of.
func (d *Deck) Chest() *ComponentChest {
	return d.chest
}

func (d *Deck) Name() string {
	return d.name
}

//ComponentAt returns the component at a given index. It handles empty indexes
//and shadow indexes correctly.
func (d *Deck) ComponentAt(index int) *Component {
	if d.chest == nil {
		return nil
	}
	if index >= len(d.components) {
		return nil
	}
	if index >= 0 {
		return d.components[index]
	}

	//d.ShadowComponent handles all negative indexes correctly, which is what
	//we have.
	return d.ShadowComponent(index)

}

//ShadowComponent takes an index that is negative and returns a component that
//is empty but when compared to the result of previous calls to
//ShadowComponent with that index will have equality. This is important for
//sanitized states, where depending on the policy for that property, the stack
//might have its order revealed but not its contents, which requires throwaway
//but stable indexes.
func (d *Deck) ShadowComponent(index int) *Component {
	if index >= 0 {
		return nil
	}
	if index == emptyIndexSentinel {
		return nil
	}

	shadow, ok := d.vendedShadowComponents[index]

	if !ok {
		shadow = &Component{
			Deck:      d,
			DeckIndex: index,
			Values:    nil,
		}
		d.vendedShadowComponents[index] = shadow
	}

	return shadow

}

//finish is called when the deck is added to a component chest. It signifies that no more items may be added.
func (d *Deck) finish(chest *ComponentChest, name string) {
	d.chest = chest
	//If a deck has a name, it cannot receive any more items.
	d.name = name
}
