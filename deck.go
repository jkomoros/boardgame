package boardgame

import (
	"github.com/jkomoros/boardgame/errors"
	"strconv"
)

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
	shadowValues           Reader
	vendedShadowComponents map[int]*Component
	//TODO: protect shadowComponents cache with mutex to make threadsafe.
}

const genericComponentSentinel = -2

func NewDeck() *Deck {
	return &Deck{
		vendedShadowComponents: make(map[int]*Component),
	}
}

//NewSizedStack returns a new sized stack with the given size based on this
//deck. You normally do this in Empty*State delegate methods, if you aren't
//using the auto-inflating struct tags to configure your stacks.
func (d *Deck) NewSizedStack(size int) *SizedStack {
	return newSizedStack(d, size)
}

//NewGrowableStack returns a new growable stack with the given size based on
//this deck. If the growable stack has no size limit, pass 0 for maxLen. You
//normally do this in Empty*State delegate methods, if you aren't using the
//auto-inflating struct tags to configure your stacks.
func (d *Deck) NewGrowableStack(maxLen int) *GrowableStack {
	return newGrowableStack(d, maxLen)
}

//AddComponent adds a new component with the given values to the next spot in
//the deck. If the deck has already been added to a componentchest, this will
//do nothing.
func (d *Deck) AddComponent(v Reader) {
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
func (d *Deck) AddComponentMulti(v Reader, count int) {
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

//SetShadowValues sets the SubState to return for every shadow
//component that is returned. May only be set before added to a chest. Should
//generally be the same shape of componentValues as used for other components
//in the deck.
func (d *Deck) SetShadowValues(v Reader) {
	if d.chest != nil {
		return
	}
	d.shadowValues = v
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
			Values:    d.shadowValues,
		}
		d.vendedShadowComponents[index] = shadow
	}

	return shadow

}

//GenericComponent returns the component that is considereed fully generic for
//this deck. This is the component that every component will be if a Stack is
//sanitized with PolicyLen, for example. If you want to figure out if a Stack
//was sanitized according to that policy, you can compare the component to
//this.
func (d *Deck) GenericComponent() *Component {
	return d.ShadowComponent(genericComponentSentinel)
}

var illegalComponentValuesProps = map[PropertyType]bool{
	TypeEnumVar: true,
	TypeStack:   true,
	TypeTimer:   true,
}

//finish is called when the deck is added to a component chest. It signifies that no more items may be added.
func (d *Deck) finish(chest *ComponentChest, name string) error {

	for i, c := range d.components {
		if c.Values == nil {
			continue
		}
		validator, err := newReaderValidator(c.Values.Reader(), c.Values, illegalComponentValuesProps, chest, false)
		if err != nil {
			return errors.New("Component " + strconv.Itoa(i) + "failed to validate: " + err.Error())
		}
		if err := validator.Valid(c.Values.Reader()); err != nil {
			return errors.New("Component " + strconv.Itoa(i) + " failed to validate: " + err.Error())
		}
	}

	d.chest = chest
	//If a deck has a name, it cannot receive any more items.
	d.name = name
	return nil
}
