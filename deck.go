package boardgame

import (
	"strconv"

	"github.com/jkomoros/boardgame/errors"
)

/*
A Deck represents an immutable collection of a certain type of component for
a given game type. A Deck, like a Component, is immutable for a given game
type and shared by every game of that type. Every Component lives in one
deck in a position that is always the same; the Component's identity is
defined by the deck it is in and the index in that deck.

Stack is a related concept; they represent the collections in your GameState,
PlayerStates, and DynamicComponentValues where ComponentInstances live in a
given state.

Each State of every game in this game type must have each component from each
deck live in precisely one position in one stack at all times. This is
referred to as the "component invariant", and captures the notion that the
component instance is a physical entity that can only be in one place at one
time and must exist somewhere. This invariant is established via careeful
design of the library. When a given game is created,
GameDelegate.DistributeComponentToStarterStack ensures that every component
instance for the game has a home befofre the game starts, and after that the
move methods on ComponentInstance ensure that a component can't ever be copied.

You generally only create decks inside of your GameDelegate's ConfigureDecks()
method, which is called when the GameManager is being set up.

*/
type Deck struct {
	chest *ComponentChest
	//Name is only set when it's added to the component chest.
	name string
	//Components should only ever be added at initalization time. After
	//initalization, Components should be read-only.
	components             []Component
	genericValues          ComponentValues
	vendedGenericComponent Component
	//TODO: protect shadowComponents cache with mutex to make threadsafe.
}

const genericComponentSentinel = -2

//NewDeck returns a new deck, ready to have components added to it. Typically
//you call this within your GameDelegate's ConfigureDecks() method.
func NewDeck() *Deck {
	return &Deck{}
}

//NewStack returns a new default (growable Stack) with the given size based on
//this deck. The returned stack will allow up to maxSize items to be inserted.
//If you don't want to set a maxSize on the stack (you often don't) pass 0 for
//maxSize to allow it to grow without limit. Typically you'd use this in your
//GameDelegate's GameStateConstructor() and other similar methods; although in
//practice it is much more common to use struct-tag based inflation, making
//direct use of this constructor unnecessary. See StructInflater for more.
func (d *Deck) NewStack(maxSize int) Stack {
	return newGrowableStack(d, maxSize)
}

//NewSizedStack returns a new SizedStack (a stack whose FixedSize() will
//return true) associated with this deck. Typically you'd use this in your
//GameDelegate's GameStateConstructor() and other similar methods; although in
//practice it is much more common to use struct-tag based inflation, making
//direct use of this constructor unnecessary. See StructInflater for more.
func (d *Deck) NewSizedStack(size int) SizedStack {
	return newSizedStack(d, size)
}

//AddComponent adds a new component with the given values to the next spot in
//the deck. This is where you affiliate the specific immutable properties
//specific to this component and your game with the component. v may be nil.
//This method is only legal to be called before the Deck has been installed
//into a ComponentChest, which is to say generally only within your
//GameDelegate's ConfigureDecks() method. Although technically there is no
//problem with using a different underlying struct for different components in
//the same deck, in practice it is strongly discouraged because often you will
//blindly cast returned component values for a given deck into the underlying
//struct.
func (d *Deck) AddComponent(v ComponentValues) {
	if d.chest != nil {
		return
	}

	c := &component{
		deck:      d,
		deckIndex: len(d.components),
		values:    v,
	}

	if v != nil {
		v.SetContainingComponent(c)
	}

	d.components = append(d.components, c)
}

//Components returns a list of Components in order in this deck, equivalent to
//calling ComponentAt() from 0 to deck.Len()
func (d *Deck) Components() []Component {
	return d.components
}

//Len returns the number of components in this deck.
func (d *Deck) Len() int {
	return len(d.components)
}

//Chest points back to the chest we're part of.
func (d *Deck) Chest() *ComponentChest {
	return d.chest
}

//Name returns the name of this deck; the string by which it could be retrived
//from the ComponentChest it resides in. This name is implied by the string
//key the deck was associated with in the return value from
//GameDelegate.ConfigureDecks().
func (d *Deck) Name() string {
	return d.name
}

//ComponentAt returns the component at a given index. It handles empty indexes
//and shadow indexes correctly.
func (d *Deck) ComponentAt(index int) Component {
	if index >= len(d.components) {
		return nil
	}
	if index >= 0 {
		return d.components[index]
	}

	if index == emptyIndexSentinel {
		return nil
	}

	return d.GenericComponent()

}

//SetGenericValues sets the ComponentValues to return for every generic
//component that is returned via GenericComponent(). May only be set before
//added to a chest, that is within your GameDelegate's ConfigureDecks method.
//Should  be the same underlying struct type as the ComponentValues used for
//other components of this type.
func (d *Deck) SetGenericValues(v ComponentValues) {
	if d.chest != nil {
		return
	}
	d.genericValues = v
}

//GenericComponent returns the component that is considereed fully generic for
//this deck. This is the component that every component will be if a Stack
//affilated with this deck is sanitized with PolicyLen, for example. If you
//want to figure out if a Stack was sanitized according to that policy, you
//can compare the component to this. To override the ComponentValues in this
//GenericComponent, call SetGenericValues.
func (d *Deck) GenericComponent() Component {

	if d.vendedGenericComponent == nil {
		shadow := &component{
			deck:      d,
			deckIndex: genericComponentSentinel,
			values:    d.genericValues,
		}

		if shadow.Values() != nil {
			shadow.Values().SetContainingComponent(shadow)
		}

		d.vendedGenericComponent = shadow
	}

	return d.vendedGenericComponent
}

var illegalComponentValuesProps = map[PropertyType]bool{
	TypeStack: true,
	TypeBoard: true,
	TypeTimer: true,
}

//finish is called when the deck is added to a component chest. It signifies that no more items may be added.
func (d *Deck) finish(chest *ComponentChest, name string) error {

	for i, c := range d.components {
		if c.Values() == nil {
			continue
		}
		validator, err := NewStructInflater(c.Values(), illegalComponentValuesProps, chest)
		if err != nil {
			return errors.New("Component " + strconv.Itoa(i) + "failed to validate: " + err.Error())
		}
		if err := validator.Valid(c.Values()); err != nil {
			return errors.New("Component " + strconv.Itoa(i) + " failed to validate: " + err.Error())
		}
	}

	d.chest = chest
	//If a deck has a name, it cannot receive any more items.
	d.name = name
	return nil
}

//MarshalJSON marshasl in a format appropriate for use on the client.
func (d *Deck) MarshalJSON() ([]byte, error) {
	components := d.Components()

	values := make([]interface{}, len(components))

	for i, component := range components {
		values[i] = struct {
			Index  int
			Values interface{}
		}{
			i,
			component.Values(),
		}
	}

	return DefaultMarshalJSON(values)
}
