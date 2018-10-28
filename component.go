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

	//Instance returns a mutable ComponentInstance representing this component
	//in the given state. Will never return nil, even if the component isn't
	//valid in this state----although later things like ContainingStack may
	//error later in that case.
	Instance(st State) ComponentInstance

	//ImmutableInstance returns an ImmutableComponentInstance representing
	//this component in the given state. Will never return nil, even if the
	//component isn't valid in this state--although later things like
	//ContainingStack may error later in that case.
	ImmutableInstance(st ImmutableState) ImmutableComponentInstance

	//Generic returns true if this Component is the generic component for this
	//deck. You might get this component if you ask for a component from a
	//sanitized stack. This method is a convenience method equivalent to
	//checking for equivalency to Deck().GenericComponent().
	Generic() bool

	//ptr() is used for when the component itself is used as a key into
	//an index and equality is important.
	ptr() *component
}

//ImmutableComponentInstance is a specific instantiation of a component as it
//exists in the particular State it is associated with.
//ImmutableComponentInstances also implement all of the Component information,
//as a convenience you often need both bits of inforamation.  The downside of
//this is that two Component values can't be compared directly for equality
//because they may be different underlying objects and wrappers. If you want
//to see if two Components that might be from different states refer to the
//same underlying conceptual Component, use Equivalent(). However,
//ImmutableComponentInstances compared with another ImmutableComponentInstance
//for the same component in the same state will be equal. See also
//ComponentInstance, which extends this interface with mutators as well.
type ImmutableComponentInstance interface {
	//ImmutableComponentInstances have all of the information of a base Component, as
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

	//ImmutableDynamicValues returns the Dynamic Values for this component in the state
	//this instance is associated with. A convenience so you don't have to go
	//find them within the DynamicComponentValues yourself.
	ImmutableDynamicValues() ImmutableSubState

	//ContainingImmutableStack will return the stack and slot index for the
	//associated component, if that location is not sanitized, and the
	//componentinstance is legal for the state it's in. If no error is
	//returned, stack.ComponentAt(slotIndex) == c will evaluate to true.
	ContainingImmutableStack() (stack ImmutableStack, slotIndex int, err error)

	//ImmutableState returns the State object that this ComponentInstance is affiliated
	//with.
	ImmutableState() ImmutableState

	secretMoveCount() int
	movedSecretly()
}

//Note that a ComponentInstance doesn't actually guarantee that the component
//is in a mutable context at this moment, just that it was at some point on
//thie state (otherwise you couldn't have gotten a reference to it). In
//practice though, if a Component was ever in a mutable context in a given
//state, it must remain that way, because it can't be moved from a Stack to a
//ImmutableStack.

//ComponentInstance is a component instance that's in a context that is
//mutable. You generally get these from a Stack that contains them. The
//instance contains many methods to move the component to other stacks or
//locations. See also ImmutableComponentInstance, which is the same, but
//without mutator methods.
type ComponentInstance interface {

	//ComponentInstance can be used anywhere that ImmutableComponentInstance
	//can be.
	ImmutableComponentInstance

	//DynamicValues returns the Dynamic Values for this component in the state
	//this instance is associated with. A convenience so you don't have to go
	//find them within the DynamicComponentValues yourself.
	DynamicValues() SubState

	//ContainingStack will return the stack and slot index for the associated
	//component, if that location is not sanitized, and the componentinstance
	//is legal for the state it's in. If no error is returned,
	//stack.ComponentAt(slotIndex) == c will evaluate to true.
	ContainingStack() (stack Stack, slotIndex int, err error)

	//MoveTo moves the specified component in its current stack to the
	//specified slot in the destination stack. The destination stack must be
	//different than the one the component's currently in--if you're moving
	//components within a stack, use SwapComponent. In destination, slotIndex
	//must point to a valid "slot" to put a component, such that after
	//insertion, using that index on the destination will return that
	//component. In defaults Stacks, slots are any index from 0 up to and
	//including stack.Len(), because the stack will grow to insert the
	//component between existing components if necessary. For SizedStacks,
	//slotIndex must point to a currently empty slot.
	//MoveTo{First,Last,Next}Slot methods are useful if you want to move to
	//those locations. If you want the precise location of the inserted
	//component to not be visible, see SecretMoveTo.
	MoveTo(other Stack, slotIndex int) error

	//SecretMoveTo is equivalent to MoveTo, but after the move the Ids of all
	//components in destination will be scrambled. SecretMoveTo is useful when
	//the destination stack will be sanitized with something like PolicyOrder,
	//but the precise location of this insertion should not be observable.
	//Read the package doc for more about when this is useful.
	SecretMoveTo(other Stack, slotIndex int) error

	//MoveToFirstSlot moves the component to the first valid slot in the other
	//stack. For default Stacks, this is always 0. For SizedStacks, this is
	//the first empty slot from the left. A convenience wrapper around
	//stack.FirstSlot.
	MoveToFirstSlot(other Stack) error

	//MoveToLastSlot moves the component to the last valid slot in the other
	//stack. For default Stacks, this is always Len(). For SizedStacks, this
	//is the first empty slot from the right. A convenience wrappar around
	//stack.LastSlot().
	MoveToLastSlot(other Stack) error

	//MoveToNextSlot moves the component to the next valid slot in the other
	//stack where the component could be added without splicing. For default
	//stacks this is equivalent to MoveToLastSlot. For fixed size stacks this
	//is equivalent to MoveToFirstSlot. A convenience wrapper arond
	//stack.NextSlot().
	MoveToNextSlot(other Stack) error

	//SlideToFirstSlot takes the given component and moves it to the start of
	//the same stack, moving everything else up. It is equivalent to removing
	//the component, moving it to a temporary stack, and then moving it back
	//to the original stack with MoveToFirstSlot--but of course without
	//needing the extra scratch stack.
	SlideToFirstSlot() error

	//SlideToLastSlot takes the given component and moves it to the end of the
	//same stack, moving everything else down. It is equivalent to removing
	//the component, moving it to a temporary stack, and then moving it back
	//to the original stack with MoveToLastSlot--but of course without needing
	//the extra scratch stack.
	SlideToLastSlot() error
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

func (c *component) Generic() bool {
	if c == nil {
		return false
	}
	return c.Equivalent(c.Deck().GenericComponent())
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

func (c *component) ImmutableInstance(st ImmutableState) ImmutableComponentInstance {
	var ptr *state

	if st != nil {
		ptr = st.(*state)
	}

	return componentInstance{
		c,
		ptr,
	}
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
//must implement. base.ComponentValues is designed to be anonymously embedded
//in your component to implement the latter part of the interface. 'boardgame-
//util codegen' can be used to implement Reader.
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

func (c componentInstance) ContainingStack() (Stack, int, error) {
	if c.statePtr == nil {
		return nil, 0, errors.New("State is non-existent")
	}
	return c.statePtr.containingStack(c)
}

func (c componentInstance) ContainingImmutableStack() (ImmutableStack, int, error) {
	return c.ContainingStack()
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

func (c componentInstance) MoveTo(other Stack, slotIndex int) error {
	if slotIndex < 0 {
		return errors.New("Invalid slotIndex")
	}
	source, sourceIndex, err := c.ContainingStack()
	if err != nil {
		return errors.New("The source component was not in a mutable stack: " + err.Error())
	}
	return source.moveComponent(sourceIndex, other, slotIndex)
}

func (c componentInstance) SecretMoveTo(other Stack, slotIndex int) error {
	if slotIndex < 0 {
		return errors.New("Invalid slotIndex")
	}
	source, sourceIndex, err := c.ContainingStack()
	if err != nil {
		return errors.New("The source component was not in a mutable stack: " + err.Error())
	}
	return source.secretMoveComponent(sourceIndex, other, slotIndex)
}

func (c componentInstance) MoveToFirstSlot(other Stack) error {
	source, sourceIndex, err := c.ContainingStack()
	if err != nil {
		return errors.New("The source component was not in a mutable stack: " + err.Error())
	}
	return source.moveComponent(sourceIndex, other, other.firstSlot())
}

func (c componentInstance) MoveToLastSlot(other Stack) error {
	source, sourceIndex, err := c.ContainingStack()
	if err != nil {
		return errors.New("The source component was not in a mutable stack: " + err.Error())
	}
	return source.moveComponent(sourceIndex, other, other.lastSlot())
}

func (c componentInstance) MoveToNextSlot(other Stack) error {
	source, sourceIndex, err := c.ContainingStack()
	if err != nil {
		return errors.New("The source component was not in a mutable stack: " + err.Error())
	}
	return source.moveComponent(sourceIndex, other, other.nextSlot())
}

func (c componentInstance) SlideToFirstSlot() error {
	source, sourceIndex, err := c.ContainingStack()
	if err != nil {
		return errors.New("The source component was not in a mutable stack: " + err.Error())
	}
	return source.moveComponentToStart(sourceIndex)
}

func (c componentInstance) SlideToLastSlot() error {
	source, sourceIndex, err := c.ContainingStack()
	if err != nil {
		return errors.New("The source component was not in a mutable stack: " + err.Error())
	}
	return source.moveComponentToEnd(sourceIndex)
}

func (c componentInstance) ImmutableState() ImmutableState {
	return c.statePtr
}

func (c componentInstance) State() State {
	return c.statePtr
}

func (c componentInstance) ImmutableDynamicValues() ImmutableSubState {
	return c.DynamicValues()
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
