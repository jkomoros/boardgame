package boardgame

import (
	"encoding/json"
	"github.com/jkomoros/boardgame/errors"
	"math"
	"math/rand"
	"sort"
	"strconv"
)

const emptyIndexSentinel = -1

//Stack is one of the fundamental types in BoardGame. It represents an ordered
//stack of 0 or more Components, all from the same Deck. Each deck has 0 or
//more Stacks based off of it, and together they include all components in
//that deck, with no component residing in more than one stack. Stacks model
//things like a stack of cards, a collection of resource tokens, etc. Stacks
//can either be growable (the default), or of a fixed size (called
//SizedStacks). The default stacks have a FixedSize() of false and can grow to
//accomodate as many components as desired (up to maxSize), with no gaps in
//between components. An insertion at an index in the middle of a stack will
//simply move later components down. SizedStacks, however, have a specific
//size, with empty slots being allowed. Each insertion puts the component at
//precisely that slot, and will fail if it is already taken. Stack contains
//only read-only methods, and MutableStack extends with mutator methods. In
//general you retrieve new Stack objects from the associated deck's NewStack
//or NewSizedStack and install them in your Constructor methods (if you don't
//use tag-based auto-inflation). NewOverlappedStack and NewConcatenatedStack
//are advanced techniques.
type Stack interface {
	//Len returns the number of slots in the Stack. For a normal Stack this is
	//the number of items in the stack. For SizedStacks, this is the number of
	//slots--even if some are unfilled.
	Len() int

	//FixedSize returns if the stack has a fixed number of slots (any number
	//of which may be empty), or a non-fixed size that can grow up to MaxSize
	//and not have any nil slots. Stacks that return FixedSize() false are
	//considered default stacks, and stacks that return FixedSize() true are
	//referred to as SizedStacks.
	FixedSize() bool

	//NumComponents returns the number of components that are in this stack.
	//For default Stacks this is the same as Len(); for SizedStacks, this is
	//the number of non-nil slots.
	NumComponents() int

	//ComponentAt retrieves the component at the given index in the stack.
	ComponentAt(index int) ComponentInstance

	//Components returns all of the components. Equivalent to calling
	//ComponentAt from 0 to Len().
	Components() []ComponentInstance

	//Ids returns a slice of strings representing the Ids of each component at
	//each index. Under normal circumstances this will be the results of
	//calling c.Id() on each component in order. This information will be
	//elided if the Sanitization policy in effect is more restrictive than
	//PolicyOrder, and tweaked if PolicyOrder is in effect.
	Ids() []string

	//LastSeen represents an unordered list of the last version number at
	//which the given ID was seen in this stack. A component is "seen" at
	//three moments: 1) when it is moved to this stack, 2) immediately before
	//its Id is scrambled, and 3) immediately after its Id is scrambled.
	//LastSeen thus represents the last time that we knew for sure it was in
	//this stack --although it may have been in this stack after that, and may
	//no longer be in this stack.
	IdsLastSeen() map[string]int

	//SlotsRemaining returns how many slots there are left in this stack to
	//add items. For default stacks this will be the number of slots until
	//maxSize is reached (or MaxInt64 if there is no maxSize). For SizedStacks
	//this will be the number of empty slots.
	SlotsRemaining() int

	//MaxSize returns the Maxium Size, if set, for default stacks. For sized
	//stacks it will return the number of total current slots (filled and
	//unfilled), which is equivalent to Len().
	MaxSize() int

	//Deck returns the Deck associated with this stack.
	Deck() *Deck

	//Returns the state that this Stack is currently part of. Mainly a
	//convenience method when you have a Stack but don't know its underlying
	//type.
	state() *state

	//setState sets the state ptr that will be returned by state().
	setState(state *state)

	//Valid will return a non-nil error if the stack isn't valid currently.
	//Normal stacks always reutrn nil, but MergedStacks might return non-nil,
	//for example if the two stacks being merged are different sizes for an
	//overlapped stack. Valid is checked just before state is saved. If any
	//stack returns any non-nil for this then the state will not be saved.
	Valid() error

	//Board will return the Board that this Stack is part of, or nil if it is
	//not part of a board.
	Board() Board

	//If Board returns a non-nil Board, this will return the index within the
	//Board that this stack is.
	BoardIndex() int
}

//MutableStack is a Stack that also has mutator methods.
type MutableStack interface {
	Stack

	//MoveAllTo moves all of the components in this stack to the other stack,
	//by repeatedly calling RemoveFirst() and InsertBack(). Errors and does
	//not complete the move if there's not enough space in the target stack.
	MoveAllTo(other MutableStack) error

	//Shuffle shuffles the order of the stack, so that it has the same items,
	//but in a different order. In a SizedStack, the empty slots will move
	//around as part of a shuffle. Shuffling will scramble all of the ids in
	//the stack, such that the Ids of all items in the stack change. See the
	//package doc section on sanitization for more on Id scrambling.
	Shuffle() error

	//PublicShuffle is the same as Shuffle, but the Ids are not scrambled
	//after the shuffle. PublicShuffle makes sense in cases where only a small
	//number of cards are shuffled and a preternaturally savvy observer should
	//be able to keep track of them. The normal Shuffle() is almost always
	//what you want.
	PublicShuffle() error

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
	//component. In defaults Stacks, slots are any index from 0 up to and
	//including stack.Len(), because the stack will grow toinsert the
	//component between existing components if necessary. For SizedStacks,
	//slotIndex must point to a currently empty slot. Use
	//{First,Last}{Component,Slot}Index constants to automatically set these
	//indexes to common values. If you want the precise location of the
	//inserted component to not be visible, see SecretMoveComponent.
	MoveComponent(componentIndex int, destination MutableStack, slotIndex int) error

	//SecretMoveComponent is equivalent to MoveComponent, but after the move
	//the Ids of all components in destination will be scrambled.
	//SecretMoveComponent is useful when the destination stack will be
	//sanitized with something like PolicyOrder, but the precise location of
	//this insertion should not be observable. Read the package doc for more
	//about when this is useful.
	SecretMoveComponent(componentIndex int, destination MutableStack, slotIndex int) error

	//MoveComponentToEnd takes the given component and moves it to the end of
	//the same stack, moving everything else down. It is equivalent to
	//removing the component, moving it to a temporary stack, and then moving
	//it back to the original stack with a slotIndex of LastSlotIndex--but of
	//course without needing the extra scratch stack.
	MoveComponentToEnd(componentIndex int) error

	//MoveComponentToStart takes the given component and moves it to the start
	//of the same stack, moving everything else up. It is equivalent to
	//removing the component, moving it to a temporary stack, and then moving
	//it back to the original stack with a slotIndex of FirstSlotIndex--but of
	//course without needing the extra scratch stack.
	MoveComponentToStart(componentIndex int) error

	//SortComponents sorts the stack's components in the order implied by less
	//by repeatedly calling SwapComponents. Errors if any SwapComponents
	//errors. If error is non-nil, the stack may be left in an arbitrary order.
	SortComponents(less func(i, j Component) bool) error

	//Resizable returns true if calls to any of the methods that change the
	//Size of the stack are legal to call in general. Currently only stacks
	//within a Board return Resizable false. If this returns false, any of
	//those size mutating methods will fail with an error.
	Resizable() bool

	//ContractSize changes the size of the stack. For default stacks it
	//contracts the MaxSize, if non-zero. For sized stacks it will reduce the
	//size by removing the given number of slots, starting from the right.
	//This method will fail if there are more components in the stack
	//currently than would fit in newSize.
	ContractSize(newSize int) error

	//ExpandSize changes the size of the stack. For default stacks it
	//increases MaxSize (unless it is zero). For sized stacks it does it by
	//adding the given number of newSlots to the end.
	ExpandSize(newSlots int) error

	//SetSize is a convenience method that will call ContractSize or
	//ExpandSize depending on the current Len() and the target len. Calling
	//SetSize on a stack that is already that size is a no-op. For default
	//stacks, this is the only sway to switch from a zero MaxSize (no limit)
	//to a non-zero MaxSize().
	SetSize(newSize int) error

	//SizeToFit is a simple convenience wrapper around ContractSize. It
	//automatically sizes the stack down so that there are no empty slots.
	SizeToFit() error

	//MutableBoard will return a mutable reference to the Board we're part of,
	//if we're part of a board.
	MutableBoard() MutableBoard

	//removeComponentAt returns the component at componentIndex, and removes
	//it from the stack. For GrowableStacks, this will splice `out the
	//component. For SizedStacks it will simply vacate that slot. This should
	//only be called by MoveComponent. Performs minimal error checking because
	//it is only used inside of MoveComponent.
	removeComponentAt(componentIndex int) ComponentInstance

	//insertComponentAt inserts the given component at the given slot index,
	//such that calling ComponentAt with slotIndex would return that
	//component. For GrowableStacks, this splices in the component. For
	//SizedStacks, this just inserts the component in the slot. This should
	//only be called by MoveComponent. Performs minimal error checking because
	//it is only used inside of Movecomponent and game.SetUp.
	insertComponentAt(slotIndex int, component ComponentInstance)

	//insertNext is a convenience wrapper around insertComponentAt.
	insertNext(c ComponentInstance)

	//Whether or not the stack is set up to be modified right now.
	modificationsAllowed() error

	//applySanitizationPolicy applies the given policy to ourselves. This
	//should only be called by methods in sanitization.go.
	applySanitizationPolicy(policy Policy)

	//idSeen is called when an Id is seen (that is, either when added to the
	//item or right before being scrambled)
	idSeen(id string)

	//scrambleIds copies all component ids to persistentPossibleIds, then
	//increments all components secretMoveCount.
	scrambleIds()

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

	//used to import the state from another stack into this one. This allows
	//stacks to be phsyically the same within a state as what was returned
	//from the constructor.
	importFrom(other Stack) error
}

//These special Indexes are designed to be provided to stack.MoveComponent.
const (
	//FirstComponentIndex is computed to be the first  index, from the left,
	//where ComponentAt() will return a component. For default Stacks this is
	//always 0 (for non-empty stacks); for SizedStacks, it's the first non-
	//empty slot from the left.
	FirstComponentIndex = -1
	//LastComponentIndex is computed to be the largest index where
	//ComponentAt() will return a component. For default Stacks, this is
	//always Len() - 1 (for non-empty stacks); for SizedStacks, it's the first
	//non-empty slot from the right.
	LastComponentIndex = -2
	//FirstSlotIndex is computed to be the first index that it is valid to
	//insert a component at (a "slot"). For default Stacks, this is always 0.
	//For SizedStacks, this is the first empty slot from the left.
	FirstSlotIndex = -3
	//LastSlotIndex is computed to be the last index that it is valid to
	//insert a component at (a "slot"). For default Stacks, this is always
	//Len(). For SizedStacks, this is the first empty slot from the right.
	LastSlotIndex = -4
	//NextSlotIndex returns the next slot index, from the left, where a
	//component could be inserted without splicing--that is, without shifting
	//other components to the right. For SizedStacks, this is equivalent to
	//FirstSlotIndex. For default Stacks, this is equivalent to LastSlotIndex.
	NextSlotIndex = -5
)

//growableStack is a Stack that has a variable number of slots, none of which
//may be empty. It can optionally have a max size. Create a new one with
//deck.NewGrowableStack.
type growableStack struct {
	//Deck is the deck that we're a part of. This will be nil if we aren't
	//inflated.
	deckPtr *Deck
	//We need to maintain the name of deck because sometimes we aren't
	//inflated yet (like after being deserialized from disk)
	deckName string
	//The indexes from the given deck that this stack contains, in order.
	indexes []int

	//If overrideIds is nil, we'll just fetch them from all of component's Id.
	overrideIds []string

	idsLastSeen map[string]int

	//size, if set, says the maxmimum number of items allowed in the Stack. 0
	//means that the Stack may grow without bound.
	maxSize int
	//Each stack is associated with precisely one state. This is consulted to
	//verify that components being transfered between stacks are part of a
	//single state. Set in empty{Game,Player}State.
	statePtr *state

	board      MutableBoard
	boardIndex int
}

//sizedStack is a Stack that has a fixed number of slots, any of which may be
//empty. Create a new one with deck.NewSizedStack.
type sizedStack struct {
	//Deck is the deck we're a part of. This will be nil if we aren't inflated.
	deckPtr *Deck
	//We need to maintain the name of deck because sometimes we aren't
	//inflated yet (like after being deserialized from disk)
	deckName string
	//Indexes will always have a len of size. Slots that are "empty" will have
	//index of -1.
	indexes []int

	//If overrideIds is nil, we'll just fetch them from all of component's Id.
	overrideIds []string

	idsLastSeen map[string]int

	//Size is the number of slots.
	size int
	//Each stack is associated with precisely one state. This is consulted to
	//verify that components being transfered between stacks are part of a
	//single state. Set in empty{Game,Player}State.
	statePtr *state
}

//mergedStack is a derived stack that is made of two stacks, either in
//concatenate mode (default) or overlap mode.
type mergedStack struct {
	stacks  []Stack
	overlap bool
}

//stackJSONObj is an internal struct that we populate and use to implement
//MarshalJSON so stacks can be saved in output JSON with minimum fuss.
type stackJSONObj struct {
	Deck        string
	Indexes     []int
	Ids         []string
	IdsLastSeen map[string]int
	Size        int `json:",omitempty"`
	MaxSize     int `json:",omitempty"`
}

//NewConcatenatedStack returns a new merged stack where all of the components
//in the first stack will show up, then all of the components in the second
//stack, and on down the list of stacks. In practice this is useful as a
//computed property when you have a logical stack made up of components that
//are santiized followed by components that are not sanitized, like in a
//blackjack hand. All stacks must be from the same deck.
func NewConcatenatedStack(stack ...Stack) Stack {
	return &mergedStack{
		stacks:  stack,
		overlap: false,
	}
}

//NewOverlappedStack returns a new merged stack where any gaps in the first
//stack will be filled with whatever is in the same position in the second
//stack, and so on down the line. In practice this is useful as a computed
//property when you have a logical stack made up of components where some are
//sanitized and some are not, like the grid of cards in Memory. All stacks
//must be from the same deck, and all stacks must be FixedSize.
func NewOverlappedStack(stack ...Stack) Stack {

	return &mergedStack{
		stacks:  stack,
		overlap: true,
	}

}

//NewGrowableStack creates a new growable stack with the given Deck and Cap.
func newGrowableStack(deck *Deck, maxSize int) *growableStack {

	if maxSize < 0 {
		maxSize = 0
	}

	return &growableStack{
		deckPtr:     deck,
		deckName:    deck.Name(),
		indexes:     make([]int, 0),
		idsLastSeen: make(map[string]int),
		maxSize:     maxSize,
	}
}

//NewSizedStack creates a new SizedStack for the given deck, with the
//specified size.
func newSizedStack(deck *Deck, size int) *sizedStack {
	if size < 0 {
		size = 0
	}

	indexes := make([]int, size)

	for i := 0; i < size; i++ {
		indexes[i] = emptyIndexSentinel
	}

	return &sizedStack{
		deckPtr:     deck,
		deckName:    deck.Name(),
		indexes:     indexes,
		idsLastSeen: make(map[string]int),
		size:        size,
	}
}

func (g *growableStack) importFrom(other Stack) error {
	otherGrowable, ok := other.(*growableStack)

	if !ok {
		return errors.New("The other stack provided was not a growable.")
	}

	myState := g.statePtr
	g.copyFrom(otherGrowable)
	g.statePtr = myState
	return nil

}

func (g *growableStack) copyFrom(other *growableStack) {
	(*g) = *other
	g.indexes = make([]int, len(other.indexes))
	copy(g.indexes, other.indexes)
	g.idsLastSeen = make(map[string]int, len(other.idsLastSeen))
	for key, val := range other.idsLastSeen {
		g.idsLastSeen[key] = val
	}
}

func (s *sizedStack) importFrom(other Stack) error {
	otherSized, ok := other.(*sizedStack)

	if !ok {
		return errors.New("The other stack provided was not a sized.")
	}

	myState := s.statePtr
	s.copyFrom(otherSized)
	s.statePtr = myState
	return nil

}

func (s *sizedStack) copyFrom(other *sizedStack) {
	*s = *other
	s.indexes = make([]int, len(other.indexes))
	copy(s.indexes, other.indexes)
	s.idsLastSeen = make(map[string]int, len(other.idsLastSeen))
	for key, val := range other.idsLastSeen {
		s.idsLastSeen[key] = val
	}
}

//Len returns the number of items in the stack.
func (s *growableStack) Len() int {
	return len(s.indexes)
}

//Len returns the number of slots in the stack. It will always equal Size.
func (s *sizedStack) Len() int {
	return len(s.indexes)
}

func (m *mergedStack) Len() int {
	if len(m.stacks) == 0 {
		return 0
	}
	if m.overlap {
		return m.stacks[0].Len()
	}
	result := 0
	for _, stack := range m.stacks {
		result += stack.Len()
	}
	return result
}

func (g *growableStack) NumComponents() int {
	return len(g.indexes)
}

func (s *sizedStack) NumComponents() int {
	count := 0
	for _, index := range s.indexes {
		if index != emptyIndexSentinel {
			count++
		}
	}
	return count
}

func (m *mergedStack) NumComponents() int {
	if len(m.stacks) == 0 {
		return 0
	}
	if m.overlap {
		count := 0
		for i, c := range m.stacks[0].Components() {
			if c != nil {
				count++
				continue
			}
			for depth := 1; depth < len(m.stacks); depth++ {
				if c := m.stacks[depth].ComponentAt(i); c != nil {
					count++
					break
				}
			}
		}
		return count
	}
	result := 0
	for _, stack := range m.stacks {
		result += stack.NumComponents()
	}
	return result
}

func (g *growableStack) FixedSize() bool {
	return false
}

func (s *sizedStack) FixedSize() bool {
	return true
}

func (m *mergedStack) FixedSize() bool {
	for _, stack := range m.stacks {
		if !stack.FixedSize() {
			return false
		}
	}
	return true
}

func (g *growableStack) Valid() error {
	return nil
}

func (s *sizedStack) Valid() error {
	return nil
}

func (m *mergedStack) Valid() error {
	if len(m.stacks) == 0 {
		return errors.New("No sub-stacks provided.")
	}
	for i, stack := range m.stacks {
		if stack == nil {
			return errors.New("stack " + strconv.Itoa(i) + " is nil")
		}
	}
	deck := m.stacks[0].Deck()
	for i, stack := range m.stacks {
		if stack.Deck() != deck {
			return errors.New("stack " + strconv.Itoa(i) + " had a different deck than other sub-stacks")
		}
	}

	if !m.overlap {
		return nil
	}

	for i, stack := range m.stacks {
		if !stack.FixedSize() {
			return errors.New("stack " + strconv.Itoa(i) + " was not fixed size, but overlap stacks require them all to be fixed size")
		}
	}

	stackLen := m.stacks[0].Len()

	for i, stack := range m.stacks {
		if stack.Len() != stackLen {
			return errors.New("stack " + strconv.Itoa(i) + " was not the same length as the others")
		}
	}

	return nil

}

func (g *growableStack) Components() []ComponentInstance {
	result := make([]ComponentInstance, len(g.indexes))

	for i := 0; i < len(result); i++ {
		result[i] = g.ComponentAt(i)
	}

	return result
}

func (s *sizedStack) Components() []ComponentInstance {
	result := make([]ComponentInstance, len(s.indexes))

	for i := 0; i < len(result); i++ {
		result[i] = s.ComponentAt(i)
	}

	return result
}

func (m *mergedStack) Components() []ComponentInstance {
	if len(m.stacks) == 0 {
		return []ComponentInstance{}
	}
	if m.overlap {

		result := make([]ComponentInstance, len(m.stacks[0].Components()))

		for i, _ := range m.stacks[0].Components() {
			result[i] = m.ComponentAt(i)
		}
		return result

	}

	var result []ComponentInstance

	for _, stack := range m.stacks {
		result = append(result, stack.Components()...)
	}

	return result
}

//ComponentAt fetches the component object representing the n-th object in
//this stack.
func (s *growableStack) ComponentAt(index int) ComponentInstance {

	//Substantially recreated in SizedStack.ComponentAt()
	if index >= s.Len() || index < 0 {
		return nil
	}

	if s.deckPtr == nil {
		return nil
	}

	deckIndex := s.indexes[index]

	//ComponentAt will handle negative values and empty sentinel correctly.
	component := s.Deck().ComponentAt(deckIndex)
	if component == nil {
		return nil
	}
	return component.Instance(s.state())

}

//ComponentAt fetches the component object representing the n-th object in
//this stack.
func (s *sizedStack) ComponentAt(index int) ComponentInstance {

	//Substantially recreated in GrowableStack.ComponentAt()

	if index >= s.Len() || index < 0 {
		return nil
	}

	if s.deckPtr == nil {
		return nil
	}

	deckIndex := s.indexes[index]

	//ComponentAt will handle negative values and empty sentinel correctly.
	component := s.Deck().ComponentAt(deckIndex)
	if component == nil {
		return nil
	}
	return component.Instance(s.state())
}

func (m *mergedStack) ComponentAt(index int) ComponentInstance {
	if len(m.stacks) == 0 {
		return nil
	}
	if m.overlap {
		for _, stack := range m.stacks {
			if c := stack.ComponentAt(index); c != nil {
				return c
			}
		}
		return nil
	}

	for _, stack := range m.stacks {
		if index < stack.Len() {
			return stack.ComponentAt(index)
		}
		index -= stack.Len()
	}

	return nil
}

func (g *growableStack) Ids() []string {
	if g.overrideIds != nil {
		return g.overrideIds
	}
	return stackIdsImpl(g)
}

func (s *sizedStack) Ids() []string {
	if s.overrideIds != nil {
		return s.overrideIds
	}
	return stackIdsImpl(s)
}

func (m *mergedStack) Ids() []string {

	if len(m.stacks) == 0 {
		return []string{}
	}

	cachedIds := make([][]string, len(m.stacks))
	for i, stack := range m.stacks {
		cachedIds[i] = stack.Ids()
	}

	if m.overlap {

		result := make([]string, len(cachedIds[0]))
		for i, ID := range cachedIds[0] {
			if ID != "" {
				result[i] = ID
				continue
			}
			for depth := 1; depth < len(m.stacks); depth++ {
				if cachedIds[depth][i] != "" {
					result[i] = cachedIds[depth][i]
					break
				}
			}
		}
		return result
	}

	var result []string

	for _, ids := range cachedIds {
		result = append(result, ids...)
	}

	return result
}

func stackIdsImpl(s Stack) []string {

	result := make([]string, s.Len())
	for i, c := range s.Components() {
		if c == nil {
			continue
		}
		result[i] = c.ID()
	}
	return result
}

func (g *growableStack) IdsLastSeen() map[string]int {
	//return a copy because this is important state to preserve, just in case
	//someone messes with it.
	result := make(map[string]int, len(g.idsLastSeen))
	for key, val := range g.idsLastSeen {
		result[key] = val
	}
	return result
}

func (s *sizedStack) IdsLastSeen() map[string]int {
	//return a copy because this is important state to preserve, just in case
	//someone messes with it.
	result := make(map[string]int, len(s.idsLastSeen))
	for key, val := range s.idsLastSeen {
		result[key] = val
	}
	return result
}

func (m *mergedStack) IdsLastSeen() map[string]int {
	result := make(map[string]int)

	for _, stack := range m.stacks {
		for key, val := range stack.IdsLastSeen() {
			//If there is a conflict, always prefer the highest last version
			//number seen, because that's the semantic expectation of
			//IdsLastSeen.
			if val > result[key] {
				result[key] = val
			}
		}
	}
	return result
}

func (g *growableStack) idSeen(id string) {
	if id == "" {
		return
	}
	if g.statePtr == nil {
		//Should only happen in weird tests
		return
	}
	g.idsLastSeen[id] = g.statePtr.Version()
}

func (s *sizedStack) idSeen(id string) {
	if id == "" {
		return
	}
	if s.statePtr == nil {
		//Should only happen in weird tests
		return
	}
	s.idsLastSeen[id] = s.statePtr.Version()
}

func (g *growableStack) scrambleIds() {
	for _, c := range g.Components() {
		if c == nil {
			continue
		}
		g.idSeen(c.ID())
		c.movedSecretly(g.state())
		g.idSeen(c.ID())
	}
}

func (s *sizedStack) scrambleIds() {
	for _, c := range s.Components() {
		if c == nil {
			continue
		}
		s.idSeen(c.ID())
		c.movedSecretly(s.state())
		s.idSeen(c.ID())
	}
}

//SlotsRemaining returns the count of slots left in this stack. If Cap is 0
//(inifinite) this will be MaxInt64.
func (s *growableStack) SlotsRemaining() int {
	if s.maxSize <= 0 {
		return math.MaxInt64
	}
	return s.maxSize - s.Len()
}

//SlotsRemaining returns the count of unfilled slots in this stack.
func (s *sizedStack) SlotsRemaining() int {
	count := 0
	for _, index := range s.indexes {
		if index == emptyIndexSentinel {
			count++
		}
	}
	return count
}

func (m *mergedStack) SlotsRemaining() int {

	if len(m.stacks) == 0 {
		return 0
	}

	if m.overlap {
		count := 0
		for _, c := range m.Components() {
			if c == nil {
				count++
			}
		}
		return count
	}

	count := 0

	for _, stack := range m.stacks {
		if stack.SlotsRemaining() == math.MaxInt64 {
			return math.MaxInt64
		}
		count += stack.SlotsRemaining()
	}

	return count

}

func (s *sizedStack) MoveAllTo(other MutableStack) error {
	return moveAllToImpl(s, other)
}

func (g *growableStack) MoveAllTo(other MutableStack) error {
	return moveAllToImpl(g, other)
}

func moveAllToImpl(from MutableStack, to MutableStack) error {

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

func (g *growableStack) state() *state {
	return g.statePtr
}

func (s *sizedStack) state() *state {
	return s.statePtr
}

func (m *mergedStack) state() *state {
	if len(m.stacks) == 0 {
		return nil
	}
	return m.stacks[0].state()
}

func (g *growableStack) setState(state *state) {
	g.statePtr = state
}

func (s *sizedStack) setState(state *state) {
	s.statePtr = state
}

func (m *mergedStack) setState(state *state) {
	for i, _ := range m.stacks {
		m.stacks[i].setState(state)
	}
}

func (g *growableStack) Deck() *Deck {
	if g.deckPtr == nil {
		if g.statePtr.game != nil {
			g.deckPtr = g.statePtr.game.Chest().Deck(g.deckName)
		}
	}
	return g.deckPtr
}

func (s *sizedStack) Deck() *Deck {
	if s.deckPtr == nil {
		if s.statePtr.game != nil {
			s.deckPtr = s.statePtr.game.Chest().Deck(s.deckName)
		}
	}
	return s.deckPtr
}

func (m *mergedStack) Deck() *Deck {
	if len(m.stacks) == 0 {
		return nil
	}
	return m.stacks[0].Deck()
}

func (g *growableStack) modificationsAllowed() error {
	if g.state() == nil {
		return errors.New("Modifications not allowed: stack's state not set")
	}
	if g.state().Sanitized() {
		return errors.New("Modifications not allowed: stack's state is sanitized")
	}
	return nil
}

func (s *sizedStack) modificationsAllowed() error {
	if s.state() == nil {
		return errors.New("Modifications not allowed: stack's state not set")
	}
	if s.state().Sanitized() {
		return errors.New("Modifications not allowed: stack's state is sanitized")
	}
	return nil
}

func (g *growableStack) legalSlot(index int) bool {
	if index < 0 {
		return false
	}
	if index > g.Len() {
		return false
	}

	return true
}

func (s *sizedStack) legalSlot(index int) bool {
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

func (g *growableStack) removeComponentAt(componentIndex int) ComponentInstance {

	component := g.ComponentAt(componentIndex)

	//This finicky code is replicated in s.ContractSize
	if componentIndex == 0 {
		g.indexes = g.indexes[1:]
	} else if componentIndex == g.Len()-1 {
		g.indexes = g.indexes[:g.Len()-1]
	} else {
		g.indexes = append(g.indexes[:componentIndex], g.indexes[componentIndex+1:]...)
	}

	return component

}

func (s *sizedStack) removeComponentAt(componentIndex int) ComponentInstance {
	component := s.ComponentAt(componentIndex)

	s.indexes[componentIndex] = emptyIndexSentinel

	return component
}

func (g *growableStack) insertComponentAt(slotIndex int, component ComponentInstance) {

	if slotIndex == 0 {
		g.indexes = append([]int{component.DeckIndex()}, g.indexes...)
	} else if slotIndex == g.Len() {
		g.indexes = append(g.indexes, component.DeckIndex())
	} else {
		firstPart := g.indexes[:slotIndex]
		firstPartCopy := make([]int, len(firstPart))
		copy(firstPartCopy, firstPart)
		//If we just append, it will put the component.DeckIndex in the
		//underlying slice, which will then be copied again in th`e last append.
		firstPartCopy = append(firstPartCopy, component.DeckIndex())
		g.indexes = append(firstPartCopy, g.indexes[slotIndex:]...)
	}

	g.idSeen(component.ID())

	//In some weird testing scenarios state can be nil
	if g.state() != nil {
		//TODO: only update the ids for ones after the insert point in the component stack.
		g.state().updateIndexForAllComponents(g)
	}

}

func (s *sizedStack) insertComponentAt(slotIndex int, component ComponentInstance) {
	s.indexes[slotIndex] = component.DeckIndex()
	s.idSeen(component.ID())
	//In some weird testing scenarios state can be nil
	if s.state() != nil {
		s.state().componentAdded(component, s, slotIndex)
	}
}

func (g *growableStack) insertNext(c ComponentInstance) {
	g.insertComponentAt(g.effectiveIndex(NextSlotIndex), c)
}

func (s *sizedStack) insertNext(c ComponentInstance) {
	s.insertComponentAt(s.effectiveIndex(NextSlotIndex), c)
}

func (g *growableStack) effectiveIndex(index int) int {

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

func (s *sizedStack) effectiveIndex(index int) int {

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
	stack MutableStack
	less  func(i, j Component) bool
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

func (g *growableStack) SortComponents(less func(i, j Component) bool) error {
	return sortComponentsImpl(g, less)
}

func (s *sizedStack) SortComponents(less func(i, j Component) bool) error {
	return sortComponentsImpl(s, less)
}

func sortComponentsImpl(s MutableStack, less func(i, j Component) bool) error {
	sorter := &stackSorter{
		stack: s,
		less:  less,
		err:   nil,
	}

	sort.Sort(sorter)

	return errors.NewWrapped(sorter.err)
}

func (g *growableStack) SecretMoveComponent(componentIndex int, destination MutableStack, slotIndex int) error {
	err := moveComonentImpl(g, componentIndex, destination, slotIndex)
	if err != nil {
		return err
	}
	destination.scrambleIds()
	return nil
}

func (s *sizedStack) SecretMoveComponent(componentIndex int, destination MutableStack, slotIndex int) error {
	err := moveComonentImpl(s, componentIndex, destination, slotIndex)
	if err != nil {
		return err
	}
	destination.scrambleIds()
	return nil
}

func (g *growableStack) MoveComponent(componentIndex int, destination MutableStack, slotIndex int) error {
	return moveComonentImpl(g, componentIndex, destination, slotIndex)
}

func (s *sizedStack) MoveComponent(componentIndex int, destination MutableStack, slotIndex int) error {
	return moveComonentImpl(s, componentIndex, destination, slotIndex)
}

func moveComonentImpl(source MutableStack, componentIndex int, destination MutableStack, slotIndex int) error {

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

	if source.Deck() != destination.Deck() {
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
		return errors.New("Unexpected nil component returned from removeComponentAt")
	}

	destination.insertComponentAt(slotIndex, c)

	return nil

}

func (g *growableStack) MoveComponentToEnd(componentIndex int) error {
	return moveComponentToExtremeImpl(g, componentIndex, false)
}

func (g *growableStack) MoveComponentToStart(componentIndex int) error {
	return moveComponentToExtremeImpl(g, componentIndex, true)
}

func (s *sizedStack) MoveComponentToEnd(componentIndex int) error {
	return moveComponentToExtremeImpl(s, componentIndex, false)
}

func (s *sizedStack) MoveComponentToStart(componentIndex int) error {
	return moveComponentToExtremeImpl(s, componentIndex, true)
}

func moveComponentToExtremeImpl(stack MutableStack, componentIndex int, isStart bool) error {

	scratchStack := stack.Deck().NewStack(0).(*growableStack)

	scratchStack.setState(stack.state())

	if err := stack.MoveComponent(componentIndex, scratchStack, FirstSlotIndex); err != nil {
		return errors.New("Couldn't move to scratch stack: " + err.Error())
	}

	targetSlot := LastSlotIndex

	if isStart {
		targetSlot = FirstSlotIndex
	}

	if err := scratchStack.MoveComponent(0, stack, targetSlot); err != nil {
		return errors.New("Couldn't move back from scratch stack: " + err.Error())
	}

	return nil

}

func (g *growableStack) PublicShuffle() error {
	if err := g.modificationsAllowed(); err != nil {
		return err
	}

	perm := rand.Perm(len(g.indexes))

	currentComponents := g.indexes
	g.indexes = make([]int, len(g.indexes))

	for i, j := range perm {
		g.indexes[i] = currentComponents[j]
	}

	g.state().updateIndexForAllComponents(g)

	return nil
}

func (g *growableStack) Shuffle() error {

	if err := g.PublicShuffle(); err != nil {
		return err
	}

	g.scrambleIds()

	return nil
}

func (s *sizedStack) PublicShuffle() error {
	if err := s.modificationsAllowed(); err != nil {
		return err
	}

	perm := rand.Perm(len(s.indexes))

	currentComponents := s.indexes
	s.indexes = make([]int, len(s.indexes))

	for i, j := range perm {
		s.indexes[i] = currentComponents[j]
	}

	s.state().updateIndexForAllComponents(s)

	return nil
}

func (s *sizedStack) Shuffle() error {

	if err := s.PublicShuffle(); err != nil {
		return err
	}

	s.scrambleIds()

	return nil
}

func (g *growableStack) SwapComponents(i, j int) error {
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

	g.state().componentAdded(g.ComponentAt(i), g, i)
	g.state().componentAdded(g.ComponentAt(j), g, j)

	return nil

}

func (s *sizedStack) SwapComponents(i, j int) error {
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

	s.state().componentAdded(s.ComponentAt(i), s, i)
	s.state().componentAdded(s.ComponentAt(j), s, j)

	return nil
}

func (m *mergedStack) MaxSize() int {
	if len(m.stacks) == 0 {
		return 0
	}
	if m.overlap {
		return m.stacks[0].MaxSize()
	}
	count := 0
	for _, stack := range m.stacks {
		count += stack.MaxSize()
	}
	return count
}

func (g *growableStack) MaxSize() int {
	return g.maxSize
}

func (g *growableStack) ExpandSize(newSlots int) error {

	if !g.Resizable() {
		return errors.New("Stack is not resizable.")
	}

	if g.MaxSize() == 0 {
		return errors.New("Can't expand maxSize; maxSize is currently not set. Use SetSize first")
	}

	if newSlots <= 0 {
		return errors.New("Expand size must have a positive non-zero number for newSlots")
	}

	g.maxSize = g.maxSize + newSlots

	return nil
}

func (g *growableStack) ContractSize(newSize int) error {
	if !g.Resizable() {
		return errors.New("Stack is not resizable.")
	}

	if g.MaxSize() == 0 {
		return errors.New("Can't expand maxSize; maxSize is currently not set. Use SetSize first")
	}

	if newSize <= 0 {
		return errors.New("Contract size must be given a positive, non-zero number for newSize")
	}

	if newSize > g.MaxSize() {
		return errors.New("Contract size must be passed a smaller size")
	}

	if newSize < g.NumComponents() {
		return errors.New("Can't set the max size to a size smaller than the current number of components")
	}

	g.maxSize = newSize
	return nil
}

func (g *growableStack) SetSize(newSize int) error {
	if g.MaxSize() == newSize {
		//No op!
		return nil
	}

	if newSize < 0 {
		newSize = 0
	}

	if g.MaxSize() == 0 {
		g.maxSize = newSize
		return nil
	}

	if g.MaxSize() > newSize {
		return g.ContractSize(newSize)
	}

	return g.ExpandSize(newSize - g.MaxSize())
}

func (g *growableStack) Board() Board {
	return g.board
}

func (g *growableStack) MutableBoard() MutableBoard {
	return g.board
}

func (g *growableStack) BoardIndex() int {
	return g.boardIndex
}

func (s *sizedStack) Board() Board {
	return nil
}

func (s *sizedStack) MutableBoard() MutableBoard {
	return nil
}

func (s *sizedStack) BoardIndex() int {
	return 0
}

func (m *mergedStack) Board() Board {
	return nil
}

func (m *mergedStack) BoardIndex() int {
	return 0
}

func (g *growableStack) SizeToFit() error {
	return g.SetSize(g.Len())
}

func (g *growableStack) Resizable() bool {
	return g.board == nil
}

func (s *sizedStack) Resizable() bool {
	return true
}

func (s *sizedStack) MaxSize() int {
	return s.Len()
}

func (s *sizedStack) ExpandSize(newSlots int) error {
	if !s.Resizable() {
		return errors.New("Stack is not resizable.")
	}

	if newSlots < 1 {
		return errors.New("Can't add 0 or negative slots to a sized stack")
	}

	slots := make([]int, newSlots)

	for i, _ := range slots {
		slots[i] = -1
	}

	s.indexes = append(s.indexes, slots...)

	s.size = s.size + newSlots

	return nil
}

func (s *sizedStack) ContractSize(newSize int) error {
	if !s.Resizable() {
		return errors.New("Stack is not resizable.")
	}
	if newSize > s.Len() {
		return errors.New("Contract size cannot be used to grow a sized stack")
	}
	if newSize < s.NumComponents() {
		return errors.New("The proposed newSize for stack would not be sufficient to contain the components currently in the stack")
	}

	if newSize < 0 {
		newSize = 0
	}

	for s.Len() > newSize {
		slotIndex := -1
		//Find the next slot from the right.
		for i := len(s.indexes) - 1; i >= 0; i-- {
			//TODO: this is not as time efficient as it could be, could start
			//the search from last known non-empty location.
			if s.indexes[i] == emptyIndexSentinel {
				slotIndex = i
				break
			}
		}

		if slotIndex == -1 {
			return errors.New("There was an unexpected error contracting size of stack: no more slots to remove!")
		}

		//This finicky code is replicated in g.removeComponentAt
		if slotIndex == 0 {
			s.indexes = s.indexes[1:]
		} else if slotIndex == s.Len()-1 {
			s.indexes = s.indexes[:s.Len()-1]
		} else {
			s.indexes = append(s.indexes[:slotIndex], s.indexes[slotIndex+1:]...)
		}
	}

	s.size = newSize

	return nil

}

func (s *sizedStack) SetSize(newSize int) error {

	if s.Len() == newSize {
		//No op!
		return nil
	}

	if s.Len() > newSize {
		return s.ContractSize(newSize)
	}

	return s.ExpandSize(newSize - s.Len())
}

func (s *sizedStack) SizeToFit() error {
	return s.ContractSize(s.NumComponents())
}

func (g *growableStack) MarshalJSON() ([]byte, error) {
	obj := &stackJSONObj{
		Deck:        g.deckName,
		Indexes:     g.indexes,
		Ids:         g.Ids(),
		IdsLastSeen: g.idsLastSeen,
		MaxSize:     g.maxSize,
	}
	return json.Marshal(obj)
}

func (s *sizedStack) MarshalJSON() ([]byte, error) {
	//TODO: test this, including Size
	obj := &stackJSONObj{
		Deck:        s.deckName,
		Indexes:     s.indexes,
		Ids:         s.Ids(),
		IdsLastSeen: s.idsLastSeen,
		Size:        s.size,
	}
	return json.Marshal(obj)
}

func (m *mergedStack) MarshalJSON() ([]byte, error) {

	components := m.Components()

	indexes := make([]int, len(components))

	for i, c := range components {
		if c == nil {
			indexes[i] = emptyIndexSentinel
			continue
		}
		indexes[i] = c.DeckIndex()
	}

	obj := &stackJSONObj{
		Deck:        m.Deck().Name(),
		Indexes:     indexes,
		Ids:         m.Ids(),
		IdsLastSeen: m.IdsLastSeen(),
	}
	if m.FixedSize() {
		obj.Size = m.Len()
	} else {
		obj.MaxSize = m.MaxSize()
	}
	return json.Marshal(obj)
}

func (m *mergedStack) UnmarshalJSON(blob []byte) error {
	//Just drop it on the floor; we have all of the cofig we need from the
	//constructor for our containing state.
	return nil
}

func (g *growableStack) UnmarshalJSON(blob []byte) error {
	obj := &stackJSONObj{}
	if err := json.Unmarshal(blob, obj); err != nil {
		return err
	}
	//TODO: what if any of these required fields are zero? Should we return
	//error?
	g.deckName = obj.Deck
	g.indexes = obj.Indexes
	g.idsLastSeen = obj.IdsLastSeen
	g.maxSize = obj.MaxSize
	return nil
}

func (s *sizedStack) UnmarshalJSON(blob []byte) error {
	obj := &stackJSONObj{}
	if err := json.Unmarshal(blob, obj); err != nil {
		return err
	}
	//TODO: what if any of these required fields are zero? Should we return
	//error?
	if len(obj.Indexes) != obj.Size {
		return errors.New("Couldn't unmarshal sized stack: lenght of indexes didn't agree with size")
	}
	s.deckName = obj.Deck
	s.indexes = obj.Indexes
	s.idsLastSeen = obj.IdsLastSeen
	s.size = obj.Size
	return nil
}
