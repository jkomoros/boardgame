package boardgame

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"testing"
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

	//Components returns all of the components. Equivalent to calling
	//ComponentAt from 0 to Len().
	Components() []*Component

	//Components returns all components. Equivalent to calling ComponentAt
	//from 0 to Len(), and extracting the Values of each.
	ComponentValues() []SubState

	//Ids returns a slice of strings representing the Ids of each component at
	//each index. Under normal circumstances this will be the results of
	//calling c.Id() on each component in order. This information will be
	//elided if the Sanitization policy in effect is more restrictive than
	//PolicyOrder, and tweaked if PolicyOrder is in effect.
	Ids() []string

	//PossibleIds represents an undordered list of Ids that MAY be in this
	//Stack currently. There are two classes of items that may be represented
	//in this collection: 1) items whose last known location, before their Id
	//changed, was this stack, and 2) items who are in this stack, but the
	//current Sanitization policy does not allow their precise order or even
	//the length of the stack to be recovered.
	PossibleIds() map[string]bool

	//SlotsRemaining returns how many slots there are left in this stack to
	//add items.
	SlotsRemaining() int

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

	//SortComponents sorts the stack's components in the order implied by less
	//by repeatedly calling SwapComponents. Errors if any SwapComponents
	//errors. If error is non-nil, the stack may be left in an arbitrary order.
	SortComponents(less func(i, j *Component) bool) error

	//UnsafeInsertNextComponent is designed only to be used in tests, because
	//it makes it trivial to violate the component-in-one-stack invariant. It
	//inserts the given component to the NextSlotIndex in the given stack. You
	//must pass a non-nil testing.T in order to reinforce that this is only
	//intended to be used in tests.
	UnsafeInsertNextComponent(t *testing.T, c *Component) error

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

	//insertNext is a convenience wrapper around insertComponentAt.
	insertNext(c *Component)

	//addPersistentPossibleId adds the id to possibleIds in a way that will persist.
	addPersistentPossibleId(id string)

	//Returns the state that this Stack is currently part of. Mainly a
	//convenience method when you have a Stack but don't know its underlying
	//type.
	state() *state

	//deck returns the Deck in this stack. Just a conveniene wrapper if you
	//don't know what kind of stack you have.
	deck() *Deck
}

//These special Indexes are designed to be provided to stack.MoveComponent.
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

	//We don't need to store Ids because it can be recovered fully
	//from the components in the stack.

	//We do need to store the possibleIds--at least, the ones that were last
	//known to be in this location.
	possibleIds map[string]bool

	//size, if set, says the maxmimum number of items allowed in the Stack. 0
	//means that the Stack may grow without bound.
	maxLen int
	//Each stack is associated with precisely one state. This is consulted to
	//verify that components being transfered between stacks are part of a
	//single state. Set in empty{Game,Player}State.
	statePtr *state
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

	//We don't need to store Ids because it can be recovered fully
	//from the components in the stack.

	//We do need to store the possibleIds--at least, the ones that were last
	//known to be in this location.
	possibleIds map[string]bool

	//Size is the number of slots.
	size int
	//Each stack is associated with precisely one state. This is consulted to
	//verify that components being transfered between stacks are part of a
	//single state. Set in empty{Game,Player}State.
	statePtr *state
}

//stackJSONObj is an internal struct that we populate and use to implement
//MarshalJSON so stacks can be saved in output JSON with minimum fuss.
type stackJSONObj struct {
	Deck        string
	Indexes     []int
	Ids         []string
	PossibleIds map[string]bool
	Size        int `json:",omitempty"`
	MaxLen      int `json:",omitempty"`
}

//NewGrowableStack creates a new growable stack with the given Deck and Cap.
func NewGrowableStack(deck *Deck, maxLen int) *GrowableStack {

	if maxLen < 0 {
		maxLen = 0
	}

	return &GrowableStack{
		deckPtr:     deck,
		deckName:    deck.Name(),
		indexes:     make([]int, 0),
		possibleIds: make(map[string]bool),
		maxLen:      maxLen,
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
		deckPtr:     deck,
		deckName:    deck.Name(),
		indexes:     indexes,
		possibleIds: make(map[string]bool),
		size:        size,
	}
}

func (g *GrowableStack) Copy() *GrowableStack {

	var result GrowableStack
	result = *g
	result.indexes = make([]int, len(g.indexes))
	copy(result.indexes, g.indexes)
	result.possibleIds = make(map[string]bool, len(g.possibleIds))
	for key, val := range g.possibleIds {
		result.possibleIds[key] = val
	}
	return &result
}

func (s *SizedStack) Copy() *SizedStack {
	var result SizedStack
	result = *s
	result.indexes = make([]int, len(s.indexes))
	copy(result.indexes, s.indexes)
	result.possibleIds = make(map[string]bool, len(s.possibleIds))
	for key, val := range s.possibleIds {
		result.possibleIds[key] = val
	}
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

func (g *GrowableStack) Components() []*Component {
	result := make([]*Component, len(g.indexes))

	for i := 0; i < len(result); i++ {
		result[i] = g.ComponentAt(i)
	}

	return result
}

func (s *SizedStack) Components() []*Component {
	result := make([]*Component, len(s.indexes))

	for i := 0; i < len(result); i++ {
		result[i] = s.ComponentAt(i)
	}

	return result
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
func (g *GrowableStack) ComponentValues() []SubState {
	//TODO: memoize this, as long as indexes hasn't changed

	if !g.Inflated() {
		return nil
	}

	//Substantially recreated in SizedStack.ComponentValues
	result := make([]SubState, g.Len())
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
func (s *SizedStack) ComponentValues() []SubState {
	//TODO: memoize this, as long as indexes hasn't changed

	if !s.Inflated() {
		return nil
	}

	//Substantially recreated in GrowableStack.ComponentValues
	result := make([]SubState, s.Len())
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

func (g *GrowableStack) Ids() []string {
	return stackIdsImpl(g)
}

func (s *SizedStack) Ids() []string {
	return stackIdsImpl(s)
}

func stackIdsImpl(s Stack) []string {
	result := make([]string, s.Len())
	for i, c := range s.Components() {
		if c == nil {
			continue
		}
		result[i] = c.Id(s.state())
	}
	return result
}

func (g *GrowableStack) PossibleIds() map[string]bool {
	//return a copy because this is important state to preserve, just in case
	//someone messes with it.
	result := make(map[string]bool, len(g.possibleIds))
	for key, val := range g.possibleIds {
		result[key] = val
	}
	return result
}

func (s *SizedStack) PossibleIds() map[string]bool {
	//return a copy because this is important state to preserve, just in case
	//someone messes with it.
	result := make(map[string]bool, len(s.possibleIds))
	for key, val := range s.possibleIds {
		result[key] = val
	}
	return result
}

func (g *GrowableStack) addPersistentPossibleId(id string) {
	g.possibleIds[id] = true
}

func (s *SizedStack) addPersistentPossibleId(id string) {
	s.possibleIds[id] = true
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

func (g *GrowableStack) state() *state {
	return g.statePtr
}

func (s *SizedStack) state() *state {
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

func (g *GrowableStack) UnsafeInsertNextComponent(t *testing.T, c *Component) error {
	if t == nil {
		return errors.New("You must provide a non-nil testing.T")
	}
	if g.SlotsRemaining() < 1 {
		return errors.New("There are not enough slots remaining")
	}
	g.insertNext(c)
	return nil
}

func (s *SizedStack) UnsafeInsertNextComponent(t *testing.T, c *Component) error {
	if t == nil {
		return errors.New("You must provide a non-nil testing.T")
	}
	if s.SlotsRemaining() < 1 {
		return errors.New("There are not enough slots remaining")
	}
	s.insertNext(c)
	return nil
}

func (g *GrowableStack) insertNext(c *Component) {
	g.insertComponentAt(g.effectiveIndex(NextSlotIndex), c)
}

func (s *SizedStack) insertNext(c *Component) {
	s.insertComponentAt(s.effectiveIndex(NextSlotIndex), c)
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

type stackSorter struct {
	stack Stack
	less  func(i, j *Component) bool
	err   error
}

func (s *stackSorter) Len() int {
	return s.stack.Len()
}

func (s *stackSorter) Swap(i, j int) {
	err := s.stack.SwapComponents(i, j)

	if err != nil {
		s.err = err
	}
}

func (s *stackSorter) Less(i, j int) bool {
	return s.less(s.stack.ComponentAt(i), s.stack.ComponentAt(j))
}

func (g *GrowableStack) SortComponents(less func(i, j *Component) bool) error {
	return sortComponentsImpl(g, less)
}

func (s *SizedStack) SortComponents(less func(i, j *Component) bool) error {
	return sortComponentsImpl(s, less)
}

func sortComponentsImpl(s Stack, less func(i, j *Component) bool) error {
	sorter := &stackSorter{
		stack: s,
		less:  less,
		err:   nil,
	}

	sort.Sort(sorter)

	return sorter.err
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
		Deck:        g.deckName,
		Indexes:     g.indexes,
		Ids:         g.Ids(),
		PossibleIds: g.possibleIds,
		MaxLen:      g.maxLen,
	}
	return json.Marshal(obj)
}

func (s *SizedStack) MarshalJSON() ([]byte, error) {
	//TODO: test this, including Size
	obj := &stackJSONObj{
		Deck:        s.deckName,
		Indexes:     s.indexes,
		Ids:         s.Ids(),
		PossibleIds: s.possibleIds,
		Size:        s.size,
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
	g.possibleIds = obj.PossibleIds
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
	s.possibleIds = obj.PossibleIds
	s.size = obj.Size
	return nil
}
