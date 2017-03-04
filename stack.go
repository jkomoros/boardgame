package boardgame

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"strconv"
)

const emptyIndexSentinel = -1

//Stack is one of the fundamental types in BoardGame. It represents an ordered
//stack of 0 or more Components, all from the same Deck. Each deck has 0 or
//more Stacks based off of it, and together they include all components in
//that deck, with no component residing in more than one stack. Stacks model
//things like a stack of cards, a collection of resource tokens, etc. There
//are two concrete types of Stacks: GrowableStack's and SizedStack's.
type Stack interface {
	//Len returns the number of slots in the Stack. For a GrowableStack this
	//is the number of items in the stack. For SizedStacks, this is the number
	//of slots--even if some are unfilled.
	Len() int

	//NumComponents returns the number of components that are in this stack.
	//For GrowableStacks this is the same as Len(); for SizedStacks, this is
	//the number of non-nil slots.
	NumComponents() int

	//Inflated returns true if we are inflated--that is, we have a connection
	//to the underlying deck we reference. ComponentAt and ComponentValues()
	//will fail if we are not inflated.
	Inflated() bool

	//Stacks that are not inflated will become inflated by grabbing a
	//reference to the associated deck in the provided chest.
	Inflate(chest *ComponentChest) error

	//ComponentAt retrieves the component at the given index in the stack.
	ComponentAt(index int) *Component
	//Components returns all components. Equivalent to calling ComponentAt
	//from 0 to Len(), and extracting the Values of each.
	ComponentValues() []ComponentValues
	//SlotsRemaining returns how many slots there are left in this stack to
	//add items.
	SlotsRemaining() int
	//InsertFront inserts the component at the first position in the stack
	//that it can. In GrowableStacks, this will be the front of the stack (if
	//there aren't already MaxLen), and in SizedStacks this will be in the
	//first unfilled slot. The component you insert should not currently be a
	//member of any other Stack, in order to maintain the Deck/Stack
	//invariant.
	InsertFront(c *Component) error

	//InsertBack starts at the back of the stack and then walks forward and
	//inserts the component at the first slot it fits. In GrowableStack, this
	//will merely append to the end. In SizedStacks it will be the first slot,
	//walking from the back, that is empty. The component you insert should
	//not currently be a member of any other Stack, in order to maintain the
	//Deck/Stack invariant.
	InsertBack(c *Component) error

	//RemoveFirst removes the first component in the stack. For GrowableStacks
	//this will always be the first component in the stack. For SizedStacks,
	//this will be the component in the first filled slot. Remember to insert
	//the component in another stack to maintain the Deck/Stack invariant.
	RemoveFirst() *Component

	//RemoveLast removes the last component in the stack. For GrowableStacks
	//this will always be the last component in the stack. For SizedStacks,
	//this will be the component in the last filled slot. Remember to insert
	//the component in another stack to maintain the Deck/Stack invariant.
	RemoveLast() *Component

	//MoveAllTo moves all of the components in this stack to the other stack,
	//by repeatedly calling RemoveFirst() and InsertBack(). Errors and does
	//not complete the move if there's not enough space in the target stack.
	MoveAllTo(other Stack) error

	//Shuffle shuffles the order of the stack, so that it has the same items,
	//but in a different order. In a SizedStack, the empty slots will move
	//around as part of a shuffle.
	Shuffle() error

	//SwapComponents swaps the position of two components within this stack
	//without changing the size of the stack (in SizedStacks, it is legal to
	//swap empty slots). i,j must be between [0, stack.Len()).
	SwapComponents(i, j int) error

	//MoveComponent moves the specified component in the source stack to the
	//specified slot in the destination stack. The source and destination
	//stacks must be different--if you're moving components within a stack,
	//use SwapComponent. Components and Slots are overlapping concepts but are
	//distinct. For the source you must provide a componentIndex--that is, an
	//index that computes to an index that, when passed to
	//source.ComponentAt() will return a component. In destination, slotIndex
	//must point to a valid "slot" to put a component, such that after
	//insertion, using that index on the destination will return that
	//component. In GrowableStacks, slots are any index from 0 up to and
	//including stack.Len(), because the growable stack will insert the
	//component between existing componnets if necessary. For SizedStack,
	//slotIndex must point to a currently empty slot. Use
	//{First,Last}{Component,Slot}Index constants to automatically set these
	//indexes to common values.
	MoveComponent(componentIndex int, destination Stack, slotIndex int) error

	//applySanitizationPolicy applies the given policy to ourselves. This
	//should only be called by methods in sanitization.go.
	applySanitizationPolicy(policy Policy)

	//Whether or not the stack is set up to be modified right now.
	modificationsAllowed() error

	//Takes the given index, and expands it--either returns the given index,
	//or, if it's one of {First,Last}{Component,Slot}Index, what that computes
	//to in this case.
	effectiveIndex(index int) int

	//legalSlot will return true if the provided index points to a valid slot
	//to insert a component at. For growableStacks, this is simply a check to
	//ensure it's in the range [0, stack.Len()]. For SizedStacks, it is a
	//check that the slot is valid and is currently empty. Does not expand the
	//special index constants.
	legalSlot(index int) bool

	//removeComponentAt returns the component at componentIndex, and removes
	//it from the stack. For GrowableStacks, this will splice out the
	//component. For SizedStacks it will simply vacate that slot. This should
	//only be called by MoveComponent. Performs minimal error checking because
	//it is only used inside of MoveComponent.
	removeComponentAt(componentIndex int) *Component

	//insertComponentAt inserts the given component at the given slot index,
	//such that calling ComponentAt with slotIndex would return that
	//component. For GrowableStacks, this splices in the component. For
	//SizedStacks, this just inserts the component in the slot. This should
	//only be called by MoveComponent. Performs minimal error checking because
	//it is only used inside of Movecomponent and game.SetUp.
	insertComponentAt(slotIndex int, component *Component)

	//Returns the state that this Stack is currently part of. Mainly a
	//convenience method when you have a Stack but don't know its underlying
	//type.
	state() *State

	//deck returns the Deck in this stack. Just a conveniene wrapper if you
	//don't know what kind of stack you have.
	deck() *Deck
}

const (
	//FirstComponentIndex is computed to be the first  index, from the left,
	//where ComponentAt() will return a component. For GrowableStacks this is
	//always 0 (for non-empty stacks); for SizedStacks, it's the first non-
	//empty slot from the left.
	FirstComponentIndex = -1
	//LastComponentIndex is computed to be the largest index where
	//ComponentAt() will return a component. For GrowableStacks, this is
	//always Len() - 1 (for non-empty stacks); for SizedStacks, it's the first
	//non-empty slot from the right.
	LastComponentIndex = -2
	//FirstSlotIndex is computed to be the first index that it is valid to
	//insert a component at (a "slot"). For GrowableStacks, this is always 0.
	//For SizedStacks, this is the first empty slot from the left.
	FirstSlotIndex = -3
	//LastSlotIndex is computed to be the last index that it is valid to
	//insert a component at (a "slot"). For GrowableStacks, this is always
	//Len(). For SizedStacks, this is the first empty slot from the right.
	LastSlotIndex = -4
	//NextSlotIndex returns the next slot index, from the left, where a
	//component could be inserted without splicing--that is, without shifting
	//other components to the right. For SizedStacks, this is equivalent to
	//FirstSlotIndex. For GrowableStacks, this is equivalent to LastSlotIndex.
	NextSlotIndex = -5
)

type GrowableStack struct {
	//Deck is the deck that we're a part of. This will be nil if we aren't
	//inflated.
	deckPtr *Deck
	//We need to maintain the name of deck because sometimes we aren't
	//inflated yet (like after being deserialized from disk)
	deckName string
	//The indexes from the given deck that this stack contains, in order.
	indexes []int
	//size, if set, says the maxmimum number of items allowed in the Stack. 0
	//means that the Stack may grow without bound.
	maxLen int
	//Each stack is associated with precisely one state. This is consulted to
	//verify that components being transfered between stacks are part of a
	//single state. Set in empty{Game,Player}State.
	statePtr *State
}

//SizedStack is a Stack that has a fixed number of slots, any of which may be
//empty. Create a new one with NewSizedStack.
type SizedStack struct {
	//Deck is the deck we're a part of. This will be nil if we aren't inflated.
	deckPtr *Deck
	//We need to maintain the name of deck because sometimes we aren't
	//inflated yet (like after being deserialized from disk)
	deckName string
	//Indexes will always have a len of size. Slots that are "empty" will have
	//index of -1.
	indexes []int
	//Size is the number of slots.
	size int
	//Each stack is associated with precisely one state. This is consulted to
	//verify that components being transfered between stacks are part of a
	//single state. Set in empty{Game,Player}State.
	statePtr *State
}

//stackJSONObj is an internal struct that we populate and use to implement
//MarshalJSON so stacks can be saved in output JSON with minimum fuss.
type stackJSONObj struct {
	Deck    string
	Indexes []int
	Size    int `json:",omitempty"`
	MaxLen  int `json:",omitempty"`
}

//NewGrowableStack creates a new growable stack with the given Deck and Cap.
func NewGrowableStack(deck *Deck, maxLen int) *GrowableStack {

	if maxLen < 0 {
		maxLen = 0
	}

	return &GrowableStack{
		deckPtr:  deck,
		deckName: deck.Name(),
		indexes:  make([]int, 0),
		maxLen:   maxLen,
	}
}

//NewSizedStack creates a new SizedStack for the given deck, with the
//specified size.
func NewSizedStack(deck *Deck, size int) *SizedStack {
	if size < 0 {
		size = 0
	}

	indexes := make([]int, size)

	for i := 0; i < size; i++ {
		indexes[i] = emptyIndexSentinel
	}

	return &SizedStack{
		deckPtr:  deck,
		deckName: deck.Name(),
		indexes:  indexes,
		size:     size,
	}
}

func (g *GrowableStack) Copy() *GrowableStack {

	var result GrowableStack
	result = *g
	result.indexes = make([]int, len(g.indexes))
	copy(result.indexes, g.indexes)
	return &result
}

func (s *SizedStack) Copy() *SizedStack {
	var result SizedStack
	result = *s
	result.indexes = make([]int, len(s.indexes))
	copy(result.indexes, s.indexes)
	return &result
}

//Len returns the number of items in the stack.
func (s *GrowableStack) Len() int {
	return len(s.indexes)
}

//Len returns the number of slots in the stack. It will always equal Size.
func (s *SizedStack) Len() int {
	return len(s.indexes)
}

func (g *GrowableStack) NumComponents() int {
	return len(g.indexes)
}

func (s *SizedStack) NumComponents() int {
	count := 0
	for _, index := range s.indexes {
		if index != emptyIndexSentinel {
			count++
		}
	}
	return count
}

func (s *GrowableStack) Inflated() bool {
	return s.deck() != nil
}

func (s *SizedStack) Inflated() bool {
	return s.deck() != nil
}

func (g *GrowableStack) Inflate(chest *ComponentChest) error {

	if g.Inflated() {
		return errors.New("Stack already inflated")
	}

	deck := chest.Deck(g.deckName)

	if deck == nil {
		return errors.New("Chest did not contain deck with name " + g.deckName)
	}

	g.deckPtr = deck

	return nil

}

func (s *SizedStack) Inflate(chest *ComponentChest) error {

	if s.Inflated() {
		return errors.New("Stack already inflated")
	}

	deck := chest.Deck(s.deckName)

	if deck == nil {
		return errors.New("Chest did not contain deck with name " + s.deckName)
	}

	s.deckPtr = deck

	return nil

}

//ComponentAt fetches the component object representing the n-th object in
//this stack.
func (s *GrowableStack) ComponentAt(index int) *Component {

	if !s.Inflated() {
		return nil
	}

	//Substantially recreated in SizedStack.ComponentAt()
	if index >= s.Len() || index < 0 {
		return nil
	}

	if s.deck == nil {
		return nil
	}

	deckIndex := s.indexes[index]

	//ComponentAt will handle negative values and empty sentinel correctly.
	return s.deck().ComponentAt(deckIndex)
}

//ComponentAt fetches the component object representing the n-th object in
//this stack.
func (s *SizedStack) ComponentAt(index int) *Component {

	if !s.Inflated() {
		return nil
	}

	//Substantially recreated in GrowableStack.ComponentAt()

	if index >= s.Len() || index < 0 {
		return nil
	}

	if s.deck == nil {
		return nil
	}

	deckIndex := s.indexes[index]

	//ComponentAt will handle negative values and empty sentinel correctly.
	return s.deck().ComponentAt(deckIndex)
}

//ComponentValues returns the Values of each Component in order. Useful for
//then running through a converter to the underlying struct type you know it
//is.
func (g *GrowableStack) ComponentValues() []ComponentValues {
	//TODO: memoize this, as long as indexes hasn't changed

	if !g.Inflated() {
		return nil
	}

	//Substantially recreated in SizedStack.ComponentValues
	result := make([]ComponentValues, g.Len())
	for i := 0; i < g.Len(); i++ {
		c := g.ComponentAt(i)
		if c == nil {
			result[i] = nil
			continue
		}
		result[i] = c.Values
	}
	return result
}

//ComponentValues returns the Values of each Component in order. Useful for
//then running through a converter to the underlying struct type you know it
//is.
func (s *SizedStack) ComponentValues() []ComponentValues {
	//TODO: memoize this, as long as indexes hasn't changed

	if !s.Inflated() {
		return nil
	}

	//Substantially recreated in GrowableStack.ComponentValues
	result := make([]ComponentValues, s.Len())
	for i := 0; i < s.Len(); i++ {
		c := s.ComponentAt(i)
		if c == nil {
			result[i] = nil
			continue
		}
		result[i] = c.Values
	}
	return result
}

//SlotsRemaining returns the count of slots left in this stack. If Cap is 0
//(inifinite) this will be MaxInt64.
func (s *GrowableStack) SlotsRemaining() int {
	if s.maxLen <= 0 {
		return math.MaxInt64
	}
	return s.maxLen - s.Len()
}

//SlotsRemaining returns the count of unfilled slots in this stack.
func (s *SizedStack) SlotsRemaining() int {
	count := 0
	for _, index := range s.indexes {
		if index == emptyIndexSentinel {
			count++
		}
	}
	return count
}

//InsertAtSlot inserts the given component at the specified slot in the stack,
//as long as that slot is not currently occupied.
func (s *SizedStack) InsertAtSlot(c *Component, index int) error {
	//Based on how Decks and Chests are constructed, we know the components in
	//the chest hae the right gamename, so no need to check.

	if c.Deck.Name() != s.deck().Name() {
		//We can only add items that are in our deck.

		return errors.New("The component is not part of this stack's deck.")
	}

	if index > s.Len() || index < 0 {
		return errors.New("The index is not valid")
	}

	if s.indexes[index] != emptyIndexSentinel {
		//That slot is taken!
		return errors.New("That slot is taken!")
	}

	s.indexes[index] = c.DeckIndex

	return nil
}

//InsertFront inserts the component in the first slot that is empty.
func (s *SizedStack) InsertFront(c *Component) error {

	//TODO: shouldn't this just be InsertFront, and then we pop it into the
	//Stack interface?

	//Based on how Decks and Chests are constructed, we know the components in
	//the chest hae the right gamename, so no need to check.

	if c.Deck.Name() != s.deck().Name() {
		//We can only add items that are in our deck.

		return errors.New("The component is not part of this stack's deck.")
	}

	if s.SlotsRemaining() < 1 {
		return errors.New("There are no available slots.")
	}

	for i, index := range s.indexes {
		if index == emptyIndexSentinel {
			//Found it!
			s.indexes[i] = c.DeckIndex
			return nil
		}
	}

	return nil
}

//InsertBack adds the item to the end of the stack.The Component you insert
//should not currently be a member of any other stacks, to maintain the deck
//invariant.
func (g *GrowableStack) InsertBack(c *Component) error {

	//Based on how Decks and Chests are constructed, we know the components in
	//the chest hae the right gamename, so no need to check.

	if c.Deck.Name() != g.deck().Name() {
		//We can only add items that are in our deck.

		return errors.New("The component is not part of this stack's deck.")
	}

	if g.SlotsRemaining() < 1 {
		return errors.New("There's no more room in the stack.")
	}

	g.indexes = append(g.indexes, c.DeckIndex)

	return nil
}

//InsertBack inserts the component in the first slot that is empty, starting
//from the end of the stack.
func (s *SizedStack) InsertBack(c *Component) error {

	//Based on how Decks and Chests are constructed, we know the components in
	//the chest hae the right gamename, so no need to check.

	if c.Deck.Name() != s.deck().Name() {
		//We can only add items that are in our deck.

		return errors.New("The component is not part of this stack's deck.")
	}

	if s.SlotsRemaining() < 1 {
		return errors.New("There are no available slots.")
	}

	for i := len(s.indexes) - 1; i >= 0; i-- {
		index := s.indexes[i]
		if index == emptyIndexSentinel {
			//Found it!
			s.indexes[i] = c.DeckIndex
			return nil
		}
	}

	return errors.New("Couldn't find the empty slot, even though it should have existed")
}

//InsertFront puts the component at index 0 in this stack, moving all other
//items down by one. The Component you insert should not currently be a member
//of any other stacks, to maintain the deck invariant.
func (s *GrowableStack) InsertFront(c *Component) error {

	//Based on how Decks and Chests are constructed, we know the components in
	//the chest hae the right gamename, so no need to check.

	if c.Deck.Name() != s.deck().Name() {
		//We can only add items that are in our deck.

		return errors.New("The component is not part of this stack's deck.")
	}

	if s.SlotsRemaining() < 1 {
		return errors.New("There's no more room in the stack.")
	}

	s.indexes = append([]int{c.DeckIndex}, s.indexes...)

	return nil
}

func (s *SizedStack) RemoveFirst() *Component {
	for i, index := range s.indexes {
		if index != emptyIndexSentinel {
			//Found it!
			component := s.ComponentAt(i)
			s.indexes[i] = emptyIndexSentinel
			return component
		}
	}
	//didn't find any.
	return nil
}

func (g *GrowableStack) RemoveFirst() *Component {
	if len(g.indexes) == 0 {
		return nil
	}
	component := g.ComponentAt(0)

	g.indexes = g.indexes[1:]
	return component

}

func (s *SizedStack) RemoveLast() *Component {
	for i := len(s.indexes) - 1; i >= 0; i-- {
		index := s.indexes[i]
		if index != emptyIndexSentinel {
			//Found it!
			component := s.ComponentAt(i)
			s.indexes[i] = emptyIndexSentinel
			return component
		}
	}

	//didn't find any.
	return nil
}

func (g *GrowableStack) RemoveLast() *Component {
	if len(g.indexes) == 0 {
		return nil
	}
	component := g.ComponentAt(g.Len() - 1)

	g.indexes = g.indexes[:g.Len()-1]
	return component

}

func (s *SizedStack) MoveAllTo(other Stack) error {
	return moveAllToImpl(s, other)
}

func (g *GrowableStack) MoveAllTo(other Stack) error {
	return moveAllToImpl(g, other)
}

func moveAllToImpl(from Stack, to Stack) error {

	if to.SlotsRemaining() < from.NumComponents() {
		return errors.New("Not enough space in the target stack")
	}

	for from.NumComponents() > 0 {
		if err := from.MoveComponent(FirstComponentIndex, to, NextSlotIndex); err != nil {
			return err
		}
	}

	return nil
}

func (g *GrowableStack) state() *State {
	return g.statePtr
}

func (s *SizedStack) state() *State {
	return s.statePtr
}

func (g *GrowableStack) deck() *Deck {
	return g.deckPtr
}

func (s *SizedStack) deck() *Deck {
	return s.deckPtr
}

func (g *GrowableStack) modificationsAllowed() error {
	if !g.Inflated() {
		return errors.New("Modifications not allowed: stack is not inflated")
	}
	if g.state() == nil {
		return errors.New("Modifications not allowed: stack's state not set")
	}
	if g.state().Sanitized() {
		return errors.New("Modifications not allowed: stack's state is sanitized")
	}
	return nil
}

func (s *SizedStack) modificationsAllowed() error {
	if !s.Inflated() {
		return errors.New("Modifications not allowed: stack is not inflated")
	}
	if s.state() == nil {
		return errors.New("Modifications not allowed: stack's state not set")
	}
	if s.state().Sanitized() {
		return errors.New("Modifications not allowed: stack's state is sanitized")
	}
	return nil
}

func (g *GrowableStack) legalSlot(index int) bool {
	if index < 0 {
		return false
	}
	if index > g.Len() {
		return false
	}

	return true
}

func (s *SizedStack) legalSlot(index int) bool {
	if index < 0 {
		return false
	}

	if index >= s.Len() {
		return false
	}

	if s.ComponentAt(index) != nil {
		return false
	}

	return true
}

func (g *GrowableStack) removeComponentAt(componentIndex int) *Component {

	component := g.ComponentAt(componentIndex)

	if componentIndex == 0 {
		g.indexes = g.indexes[1:]
	} else if componentIndex == g.Len()-1 {
		g.indexes = g.indexes[:g.Len()-1]
	} else {
		g.indexes = append(g.indexes[:componentIndex], g.indexes[componentIndex+1:]...)
	}

	return component

}

func (s *SizedStack) removeComponentAt(componentIndex int) *Component {
	component := s.ComponentAt(componentIndex)

	s.indexes[componentIndex] = emptyIndexSentinel

	return component
}

func (g *GrowableStack) insertComponentAt(slotIndex int, component *Component) {

	if slotIndex == 0 {
		g.indexes = append([]int{component.DeckIndex}, g.indexes...)
	} else if slotIndex == g.Len() {
		g.indexes = append(g.indexes, component.DeckIndex)
	} else {
		firstPart := g.indexes[:slotIndex]
		firstPartCopy := make([]int, len(firstPart))
		copy(firstPartCopy, firstPart)
		//If we just append, it will put the component.DeckIndex in the
		//underlying slice, which will then be copied again in the last append.
		firstPartCopy = append(firstPartCopy, component.DeckIndex)
		g.indexes = append(firstPartCopy, g.indexes[slotIndex:]...)
	}

}

func (s *SizedStack) insertComponentAt(slotIndex int, component *Component) {
	s.indexes[slotIndex] = component.DeckIndex
}

func (g *GrowableStack) effectiveIndex(index int) int {

	switch index {
	case FirstComponentIndex:
		return 0
	case LastComponentIndex:
		return g.Len() - 1
	case FirstSlotIndex:
		return 0
	case LastSlotIndex:
		return g.Len()
	case NextSlotIndex:
		return g.Len()
	}

	return index

}

func (s *SizedStack) effectiveIndex(index int) int {

	if index == FirstComponentIndex {
		for i, componentIndex := range s.indexes {
			if componentIndex != emptyIndexSentinel {
				return i
			}
		}
	}

	if index == LastComponentIndex {
		for i := len(s.indexes) - 1; i >= 0; i-- {
			if s.indexes[i] != emptyIndexSentinel {
				return i
			}
		}
	}

	if index == FirstSlotIndex || index == NextSlotIndex {
		for i, componentIndex := range s.indexes {
			if componentIndex == emptyIndexSentinel {
				return i
			}
		}
	}

	if index == LastSlotIndex {
		for i := len(s.indexes) - 1; i >= 0; i-- {
			if s.indexes[i] == emptyIndexSentinel {
				return i
			}
		}
	}

	//If we get to here either we were just provided index, or there were no
	//slots/components to return.
	return index

}

func (g *GrowableStack) MoveComponent(componentIndex int, destination Stack, slotIndex int) error {
	return moveComonentImpl(g, componentIndex, destination, slotIndex)
}

func (s *SizedStack) MoveComponent(componentIndex int, destination Stack, slotIndex int) error {
	return moveComonentImpl(s, componentIndex, destination, slotIndex)
}

func moveComonentImpl(source Stack, componentIndex int, destination Stack, slotIndex int) error {

	if source == nil {
		return errors.New("Source is a nil stack")
	}

	if destination == nil {
		return errors.New("Destination is a nil stack")
	}

	if err := source.modificationsAllowed(); err != nil {
		return errors.New("Source doesn't allow modifications: " + err.Error())
	}

	if err := source.modificationsAllowed(); err != nil {
		return errors.New("Destination doesn't allow modifications: " + err.Error())
	}

	if source == destination {
		return errors.New("Source and desintation stack are the same. Use SwapComponents instead.")
	}

	if source.state() != destination.state() {
		return errors.New("Source and destination are not members of the same state object.")
	}

	if source.deck() != destination.deck() {
		return errors.New("Source and destination are affiliated with two different decks.")
	}

	componentIndex = source.effectiveIndex(componentIndex)

	if c := source.ComponentAt(componentIndex); c == nil {
		return errors.New("The effective index, " + strconv.Itoa(componentIndex) + " does not point to an existing component in Source")
	}

	slotIndex = destination.effectiveIndex(slotIndex)

	if !destination.legalSlot(slotIndex) {
		return errors.New("The effective slot index, " + strconv.Itoa(slotIndex) + " does not point to a legal slot.")
	}

	if destination.SlotsRemaining() < 1 {
		return errors.New("The destination stack does not have any extra slots.")
	}

	c := source.removeComponentAt(componentIndex)

	if c == nil {
		panic("Unexpected nil component returned from removeComponentAt")
	}

	destination.insertComponentAt(slotIndex, c)

	return nil

}

func (g *GrowableStack) Shuffle() error {

	if err := g.modificationsAllowed(); err != nil {
		return err
	}

	perm := rand.Perm(len(g.indexes))

	currentComponents := g.indexes
	g.indexes = make([]int, len(g.indexes))

	for i, j := range perm {
		g.indexes[i] = currentComponents[j]
	}

	return nil

}

func (s *SizedStack) Shuffle() error {

	if err := s.modificationsAllowed(); err != nil {
		return err
	}

	perm := rand.Perm(len(s.indexes))

	currentComponents := s.indexes
	s.indexes = make([]int, len(s.indexes))

	for i, j := range perm {
		s.indexes[i] = currentComponents[j]
	}

	return nil
}

func (g *GrowableStack) SwapComponents(i, j int) error {
	if err := g.modificationsAllowed(); err != nil {
		return err
	}
	//check i j indexes are legal
	if i < 0 {
		return errors.New("i must be 0 or greater")
	}
	if j < 0 {
		return errors.New("j must be 0 or greater")
	}
	if i >= g.Len() {
		return errors.New("i must be less than or equal to the stack's length")
	}
	if j >= g.Len() {
		return errors.New("j must be less than or equal to the stack's length")
	}
	if i == j {
		return errors.New("i and j were the same")
	}

	g.indexes[i], g.indexes[j] = g.indexes[j], g.indexes[i]

	return nil

}

func (s *SizedStack) SwapComponents(i, j int) error {
	if err := s.modificationsAllowed(); err != nil {
		return err
	}
	//check i j indexes are legal
	if i < 0 {
		return errors.New("i must be 0 or greater")
	}
	if j < 0 {
		return errors.New("j must be 0 or greater")
	}
	if i >= s.Len() {
		return errors.New("i must be less than or equal to the stack's length")
	}
	if j >= s.Len() {
		return errors.New("j must be less than or equal to the stack's length")
	}
	if i == j {
		return errors.New("i and j were the same")
	}

	s.indexes[i], s.indexes[j] = s.indexes[j], s.indexes[i]

	return nil
}

func (g *GrowableStack) MarshalJSON() ([]byte, error) {
	obj := &stackJSONObj{
		Deck:    g.deckName,
		Indexes: g.indexes,
		MaxLen:  g.maxLen,
	}
	return json.Marshal(obj)
}

func (s *SizedStack) MarshalJSON() ([]byte, error) {
	//TODO: test this, including Size
	obj := &stackJSONObj{
		Deck:    s.deckName,
		Indexes: s.indexes,
		Size:    s.size,
	}
	return json.Marshal(obj)
}

func (g *GrowableStack) UnmarshalJSON(blob []byte) error {
	obj := &stackJSONObj{}
	if err := json.Unmarshal(blob, obj); err != nil {
		return err
	}
	//TODO: what if any of these required fields are zero? Should we return
	//error?
	g.deckName = obj.Deck
	g.indexes = obj.Indexes
	g.maxLen = obj.MaxLen
	return nil
}

func (s *SizedStack) UnmarshalJSON(blob []byte) error {
	obj := &stackJSONObj{}
	if err := json.Unmarshal(blob, obj); err != nil {
		return err
	}
	//TODO: what if any of these required fields are zero? Should we return
	//error?
	s.deckName = obj.Deck
	s.indexes = obj.Indexes
	s.size = obj.Size
	return nil
}
