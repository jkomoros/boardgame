package boardgame

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"strconv"
)

//A Component represents a movable resource in the game. Cards, dice, meeples,
//resource tokens, etc are all components. Values is a struct that stores the
//specific values for the component. Components are the same across all games
//of this type. Component references should not be compared directly for
//equality, as sometimes different underlying objects will represent the same
//notional component (in order to satisfy both the Component and
//ComponentInstance interfaces simultaneously). Instead, use Equivalent() to
//test that two Components refer to the same conceptual Component.
type Component interface {
	Values() ComponentValues
	Deck() *Deck
	DeckIndex() int
	//Equivalent checks whether two components are equivalent--that is,
	//whether they represent the same index in the same deck.
	Equivalent(other Component) bool

	//Instance returns a ComponentInstance representing this component in the
	//given state.
	Instance(st State) ComponentInstance

	//ptr() is used for when the component itself is used as a key into
	//an index and equality is important.
	ptr() *component
}

//ComponentInstance is a specific instantiation of a component as it exists in
//the particular State it is associated with. ComponentInstances also
//implement all of the Component information, as a convenience you often need
//both bits of inforamation.  The downside of this is that two Component
//values can't be compared directly for equality because they may be different
//underlying objects and wrappers. If you want to see if two Components that
//might be from different states refer to the same underlying conceptual
//Component, use Equivalent(). However, ComponentInstances compared with
//another ComponentInstance for the same component in the same state will be
//equal.
type ComponentInstance interface {
	//ComponentInstances have all of the information of a base Component, as
	//often that's the information you most need.
	Component

	//ID returns a semi-stable ID for this component within this game and the
	//current state this component instance is associated with. Within this
	//game, it will only change when the shuffleCount for this component
	//changes. Across games the Id for the "same" component will be different,
	//in a way that cannot be guessed without access to game.SecretSalt. See
	//the package doc for more on semi-stable Ids for components, what they
	//can be used for, and when they do (and don't) change.
	ID() string
	//DynamicValues returns the Dynamic Values for this component in the state
	//this instance is associated with. A convenience so you don't have to go
	//find them within the DynamicComponentValues yourself.
	DynamicValues() SubState

	//State returns the State object that this ComponentInstance is affiliated
	//with.
	State() State

	//Mutable returns a mutable version of this component instance. It's sugar
	//for state.ContainingMutableComponent. You must pass in a Mutable version
	//of the state associated with this ComponentInstance to prove that you
	//have a mutable state.
	Mutable(state MutableState) (MutableComponentInstance, error)

	//mutable is only to be used to up-cast to mutablecomponentindex when you
	//know that's what it is as a quick convenience for use only within this package.
	mutable() MutableComponentInstance

	secretMoveCount() int
	movedSecretly()
}

//Note that a MutableComponentInstance doesn't actually guarantee that the
//component is in a mutable context at this moment, just that it was at some
//point on thie state (otherwise you couldn't have gotten a reference to it).
//In practice though, if a Component was ever in a mutable context in a given
//state, it must remain that way, because it can't be moved from a
//MutableStack to a non-Mutable stack.

//MutableComponentInstance is a component instance that's in a context that is
//mutable. You generally get these from a MutableStack that contains them. The
//instance contains many methods to move the component to other stacks or
//locations.
type MutableComponentInstance interface {
	ComponentInstance

	//MoveTo moves the specified component in its current stack to the
	//specified slot in the destination stack. The destination stack must be
	//different than the one the component's currently in--if you're moving
	//components within a stack, use SwapComponent. In destination,
	//slotIndex must point to a valid "slot" to put a component, such that
	//after insertion, using that index on the destination will return that
	//component. In defaults Stacks, slots are any index from 0 up to and
	//including stack.Len(), because the stack will grow to insert the
	//component between existing components if necessary. For SizedStacks,
	//slotIndex must point to a currently empty slot. Use
	//{First,Last}SlotIndex constants to automatically set these
	//indexes to common values. If you want the precise location of the
	//inserted component to not be visible, see SecretMoveTo.
	MoveTo(other MutableStack, slotIndex int) error

	//SecretMoveTo is equivalent to MoveTo, but after the move the Ids of all
	//components in destination will be scrambled. SecretMoveTo is useful when
	//the destination stack will be sanitized with something like PolicyOrder,
	//but the precise location of this insertion should not be observable.
	//Read the package doc for more about when this is useful.
	SecretMoveTo(other MutableStack, slotIndex int) error

	//MoveToStart takes the given component and moves it to the start
	//of the same stack, moving everything else up. It is equivalent to
	//removing the component, moving it to a temporary stack, and then moving
	//it back to the original stack with a slotIndex of FirstSlotIndex--but of
	//course without needing the extra scratch stack.
	MoveToStart() error

	//MoveToEnd takes the given component and moves it to the end of
	//the same stack, moving everything else down. It is equivalent to
	//removing the component, moving it to a temporary stack, and then moving
	//it back to the original stack with a slotIndex of LastSlotIndex--but of
	//course without needing the extra scratch stack.
	MoveToEnd() error
}

type component struct {
	values ComponentValues
	//The deck we're a part of.
	deck *Deck
	//The index we are in the deck we're in.
	deckIndex int
}

//componentInstance has value method receivers so two that are configured the
//same will test as equal even if they were created separately, and because we
//never need to mutate the values within--all of the mutable state is handled
//on the state object.
type componentInstance struct {
	*component
	statePtr *state
}

func (c *component) Values() ComponentValues {
	if c == nil {
		return nil
	}
	return c.values
}

func (c *component) Deck() *Deck {
	if c == nil {
		return nil
	}
	return c.deck
}

func (c *component) DeckIndex() int {
	if c == nil {
		return -1
	}
	return c.deckIndex
}

func (c *component) Equivalent(other Component) bool {
	if c == nil && other == nil {
		return true
	}
	if c.Deck().Name() != other.Deck().Name() {
		return false
	}
	return c.DeckIndex() == other.DeckIndex()
}

func (c *component) ptr() *component {
	return c
}

func (c *component) Instance(st State) ComponentInstance {

	var ptr *state

	if st != nil {
		ptr = st.(*state)
	}

	return componentInstance{
		c,
		ptr,
	}
}

//ComponentValues is the interface that the Values property of a Component
//must implement. BaseComponentValues is designed to be anonymously embedded
//in your component to implement the latter part of the interface. autoreader
//can be used to implement Reader.
type ComponentValues interface {
	Reader
	//ContainingComponent is the component that this ComponentValues is
	//embedded in. It should return the component that was passed to
	//SetContainingComponent.
	ContainingComponent() Component
	//SetContainingComponent is called to let the component values know what
	//its containing component is.
	SetContainingComponent(c Component)
}

//BaseComponentValues is an optional convenience struct designed to be
//embedded anoymously in your component values to implement
//ContainingComponent() and SetContainingComponent() automatically.
type BaseComponentValues struct {
	c Component
}

func (b *BaseComponentValues) ContainingComponent() Component {
	return b.c
}

func (b *BaseComponentValues) SetContainingComponent(c Component) {
	b.c = c
}

func (c componentInstance) ID() string {

	//Shadow components shouldn't get an Id
	if c.Equivalent(c.Deck().GenericComponent()) {
		return ""
	}

	//S should  never be nil in normal circumstances, but if it is, return an
	//obviously-special Id so it doesn't appear to be the actual Id for this
	//component.
	if c.statePtr == nil {
		return ""
	}

	input := "insecuredefaultinput"

	game := c.statePtr.game
	input = game.Id() + game.secretSalt

	input += c.Deck().Name() + strconv.Itoa(c.DeckIndex())

	//The id only ever changes when the item has moved secretly.
	input += strconv.Itoa(c.secretMoveCount())

	hash := sha1.Sum([]byte(input))

	return fmt.Sprintf("%x", hash)
}

//secretMoveCount returns the secret move count for this component in the
//given state.
func (c componentInstance) secretMoveCount() int {

	if c == c.Deck().GenericComponent() {
		return 0
	}

	if c.statePtr == nil {
		return 0
	}

	s := c.statePtr

	deckMoveCount, ok := s.secretMoveCount[c.Deck().Name()]

	//No components in that deck have been moved secretly, I guess.
	if !ok {
		return 0
	}

	if c.DeckIndex() >= len(deckMoveCount) {
		//TODO: warn?
		return 0
	}

	if c.DeckIndex() < 0 {
		//This should never happen
		return 0
	}

	return deckMoveCount[c.DeckIndex()]
}

//movedSecretly increments the secretMoveCount for this component.
func (c componentInstance) movedSecretly() {
	if c.Equivalent(c.Deck().GenericComponent()) {
		return
	}

	s := c.statePtr

	if s == nil {
		return
	}

	deckMoveCount, ok := s.secretMoveCount[c.Deck().Name()]

	//We must be the first component in this deck that has been secretly
	//moved. Create the whole int stack for this group.
	if !ok {
		//The zero-value will be fine
		deckMoveCount = make([]int, len(c.Deck().Components()))
		s.secretMoveCount[c.Deck().Name()] = deckMoveCount
	}

	if c.DeckIndex() >= len(deckMoveCount) {
		//TODO: warn?
		return
	}

	if c.DeckIndex() < 0 {
		//This should never happen
		return
	}

	deckMoveCount[c.DeckIndex()]++

}

func (c componentInstance) MoveTo(other MutableStack, slotIndex int) error {
	source, sourceIndex, err := c.statePtr.ContainingMutableStack(c)
	if err != nil {
		return errors.New("The source component was not in a mutable stack: " + err.Error())
	}
	return source.MoveComponent(sourceIndex, other, slotIndex)
}

func (c componentInstance) SecretMoveTo(other MutableStack, slotIndex int) error {
	source, sourceIndex, err := c.statePtr.ContainingMutableStack(c)
	if err != nil {
		return errors.New("The source component was not in a mutable stack: " + err.Error())
	}
	return source.SecretMoveComponent(sourceIndex, other, slotIndex)
}

func (c componentInstance) MoveToStart() error {
	source, sourceIndex, err := c.statePtr.ContainingMutableStack(c)
	if err != nil {
		return errors.New("The source component was not in a mutable stack: " + err.Error())
	}
	return source.MoveComponentToStart(sourceIndex)
}

func (c componentInstance) MoveToEnd() error {
	source, sourceIndex, err := c.statePtr.ContainingMutableStack(c)
	if err != nil {
		return errors.New("The source component was not in a mutable stack: " + err.Error())
	}
	return source.MoveComponentToEnd(sourceIndex)
}

func (c componentInstance) Mutable(mState MutableState) (MutableComponentInstance, error) {
	if mState == nil {
		return nil, errors.New("Passed nil MutableState")
	}

	st := mState.(*state)

	if st != c.statePtr {
		return nil, errors.New("The MutableState passed did not match the state this componentinstance was originally from.")
	}

	stack, index, err := mState.ContainingMutableStack(c)

	if err != nil {
		return nil, err
	}

	return stack.MutableComponentAt(index), nil
}

func (c componentInstance) mutable() MutableComponentInstance {
	return c
}

func (c componentInstance) State() State {
	return c.statePtr
}

func (c componentInstance) DynamicValues() SubState {

	if c.statePtr == nil {
		return nil
	}

	dynamic := c.statePtr.DynamicComponentValues()

	values := dynamic[c.Deck().Name()]

	if values == nil {
		return nil
	}

	if len(values) <= c.DeckIndex() {
		return nil
	}

	if c.DeckIndex() < 0 {
		//TODO: is this the right beahvior now that we have auto-inflation?
		return c.Deck().Chest().Manager().Delegate().DynamicComponentValuesConstructor(c.Deck())
	}

	return values[c.DeckIndex()]
}
