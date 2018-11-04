package boardgame

import (
	"encoding/json"
	"github.com/jkomoros/boardgame/errors"
	"hash/fnv"
	"log"
	"math/rand"
	"strconv"
)

//ImmutableState is a version of State, but minus any mutator methods. Because
//states may not be modified except by moves, in almost every case where a
//state is passed to game logic you define (whether on your GameDelegate
//methods, or Legal() on your move structs), an ImmutableState will be passed
//instead. If an ImmutableState is passed to your method, it's a strong signal
//that you shouldn't modify the state. Note that idiomatic use (e.g.
//concreteStates) will cast an ImmutableState to a State immediately in order
//to retrieve the concrete structs underneath, but if you do that you have to
//be careful not to inadvertently modify the state because the changes won't
//be persisted. See the documentation for State for more about states in
//general.
type ImmutableState interface {

	//ImmutableGameState returns the ImmutableGameState for this State. See
	//State.GameState for more.
	ImmutableGameState() ImmutableSubState
	//ImmutablePlayerStates returns a slice of all ImmutablePlayerStates for
	//this State. See State.PlayerStates for more.
	ImmutablePlayerStates() []ImmutablePlayerState
	//DynamicComponentValues returns a map of deck name to array of component
	//values, one per component in that deck. See State.DynamicComponentValues
	//for more.
	ImmutableDynamicComponentValues() map[string][]ImmutableSubState

	//ImmutableCurrentPlayer returns the ImmutablePlayerState corresponding to the
	//result of delegate.CurrentPlayerIndex(), or nil if the index isn't
	//valid. See State.CurrentPlayer for more.
	ImmutableCurrentPlayer() ImmutablePlayerState
	//CurrentPlayerIndex is a simple convenience wrapper around
	//delegate.CurrentPlayerIndex(state) for this state.
	CurrentPlayerIndex() PlayerIndex

	//Version returns the version number the state is (or will be once
	//committed).
	Version() int

	//Copy returns a deep copy of the State, including copied version of the
	//Game and Player States. Note that copying uses the
	//ProperyReadSetConfigurer interface, so any properties not enumerated
	//there or otherwise defined in the constructors on your GameDelegate will
	//not be copied.
	Copy(sanitized bool) (ImmutableState, error)

	//Diagram returns a basic, ascii rendering of the state for debug rendering.
	//It thunks out to Delegate.Diagram.
	Diagram() string

	//Santizied will return false if this is a full-fidelity State object, or
	//true if it has been sanitized, which means that some properties might be
	//hidden or otherwise altered. This should return true if the object was
	//created with Copy(true)
	Sanitized() bool

	//Computed returns the computed properties for this state.
	computed() *computedProperties

	//SanitizedForPlayer produces a copy state object that has been sanitized
	//for the player at the given index. The state object returned will have
	//Sanitized() return true. Will call GameDelegate.SanitizationPolicy to
	//construct the effective policy to apply. See the documentation for
	//Policy for more on sanitization.
	SanitizedForPlayer(player PlayerIndex) ImmutableState

	//Game is the Game that this state is part of. Calling
	//Game.State(state.Version()) should return a state equivalent to this State
	//(modulo sanitization, if applied).
	Game() *Game

	//StorageRecord returns a StateStorageRecord representing the state.
	StorageRecord() StateStorageRecord

	//containingImmutableStack will return the stack and slot index for the associated
	//component, if that location is not sanitized. If no error is returned,
	//stack.ComponentAt(slotIndex) == c will evaluate to true.
	containingImmutableStack(c Component) (stack ImmutableStack, slotIndex int, err error)
}

type computedProperties struct {
	Global  PropertyCollection
	Players []PropertyCollection
}

//StateGroupType is the top-level grouping object used in a StatePropertyRef.
type StateGroupType int

const (
	StateGroupGame StateGroupType = iota
	StateGroupPlayer
	StateGroupDynamicComponentValues
)

//A StatePropertyRef is a reference to a particular property or item in a
//Property in a State, in a structured way. Currently used primarily as an
//input to your GameDelegate's SanitizationPolicy method. Get a new generic
//one, with all properties set to reasonable defaults, from
//NewStatePropertyRef.
type StatePropertyRef struct {
	//Group is which of Game, Player, or DynamicComponentValues this is a
	//reference to.
	Group StateGroupType
	//PropName is the specific property on the given SubStateObject specified
	//by the rest of the StatePropertyRef.
	PropName string

	//PlayerIndex is the index of the player, if Group is StateGroupPlayer.
	PlayerIndex int
	//DeckName is only used when Group is StateGroupDynamicComponentValues
	DeckName string

	//StackIndex specifies the index of the component within the stack (if it
	//is a stack) that is intended. Negative values signify "all components in
	//stack"
	StackIndex int
	//BoardIndex specifies the index of the Stack within the Board (if it is a
	//board) that is intended. Negative values signify "all stacks within the
	//board".
	BoardIndex int
	//DeckIndex is used only when the Group is
	//StateGroupDynamicComponentValues. Negative values mean "all values in
	//deck".
	DynamicComponentIndex int
}

//NewStatePropertyRef returns an initalized StatePropertyRef with all fields
//set to reasonable defaults. In particular, all of the Index properties are
//set to -1. It is rare for users of the library to need to create their own
//StatePropertyRefs.
func NewStatePropertyRef() StatePropertyRef {
	return StatePropertyRef{
		StateGroupGame,
		"",
		-1,
		-1,
		-1,
		"",
		-1,
	}
}

//getReader returns the reader associated with the StatePropertyRef in the
//given state, or errors if the StatePropertyRef does not refer to a valid
//reader.
func (r StatePropertyRef) associatedReadSetter(st State) (PropertyReadSetter, error) {
	switch r.Group {
	case StateGroupGame:
		gameState := st.GameState()
		if gameState == nil {
			return nil, errors.New("GameState selected, but was nil")
		}
		return gameState.ReadSetter(), nil
	case StateGroupPlayer:

		players := st.PlayerStates()
		if len(players) == 0 {
			return nil, errors.New("PlayerState selected, but no players in state")
		}

		if r.PlayerIndex < 0 {
			return nil, errors.New("PlayerState selected, but negative value for PlayerIndex")
		}

		if r.PlayerIndex >= len(players) {
			return nil, errors.New("PlayerState selected, but with a non-existent PlayerIndex")
		}

		player := players[r.PlayerIndex]

		return player.ReadSetter(), nil
	case StateGroupDynamicComponentValues:

		allDecks := st.DynamicComponentValues()

		if allDecks == nil {
			return nil, errors.New("DynamicComponentValues selected, but was nil")
		}

		values, ok := allDecks[r.DeckName]

		if !ok {
			return nil, errors.New("DeckName did not refer to any component values: " + r.DeckName)
		}

		if r.DynamicComponentIndex < 0 || r.DynamicComponentIndex >= len(values) {
			return nil, errors.New("DynamicComponentIndex referred to a component that didn't exist")
		}

		return values[r.DynamicComponentIndex].ReadSetter(), nil

	}
	return nil, errors.New("Invalid Group type")
}

//PlayerIndex is an int that represents the index of a given player in a game.
//Normal values are [0, game.NumPlayers). Special values are AdminPlayerIndex
//and ObserverPlayerIndex.
type PlayerIndex int

//ObserverPlayerIndex is a special PlayerIndex that denotes that the player in
//question is not one of the normal players, but someone generically watching.
//All hidden state should be hidden to them, and GroupSelf will never trigger
//for them.
const ObserverPlayerIndex PlayerIndex = -1

//AdminPlayerIndex is a special PlayerIndex that denotes the omniscient admin
//who can see all state and make moves whenever they want. This PlayerIndex is
//used for example to apply moves that your GameDelegate.ProposeFixUpMove
//returns, as well as when Timer's fire. It is also used when the server is in
//debug mode, allowing the given player to operate as the admin.
const AdminPlayerIndex PlayerIndex = -2

//State represents the entire semantic state of a game at a given version. For
//your specific game, GameState and PlayerStates will actually be concrete
//structs to your particular game. State is a container of gameStates,
//playerStates, and dynamicComponentValues for your game. Games often define a
//top-level concreteStates() *myGameState, []*myPlayerState so at the top of
//methods that accept a State they can quickly get concrete, type-checked
//types with only a single conversion leap of faith at the top. States contain
//mutable refrences to their contained SubStates, whereas ImmutableState does
//not. Most of the methods you define that accept states from the core game
//engine will be an ImmutableState, because the only time States should be
//modified is when the game is initally being set up before the first move,
//and during a move's Apply()  method.
type State interface {
	//State contains all of the methods of a read-only state.
	ImmutableState
	//GameState is a reference to the MutableGameState for this MutableState.
	GameState() SubState
	//Playerstates returns a slice of PlayerStates for this State.
	PlayerStates() []PlayerState
	DynamicComponentValues() map[string][]SubState

	//CurrentPlayer returns the PlayerState corresponding to the
	//result of delegate.CurrentPlayerIndex(), or nil if the index isn't
	//valid.
	CurrentPlayer() PlayerState

	//Rand returns a source of randomness. All game logic should use this rand
	//source. It is deterministically seeded when it is created for this state
	//based on the game's ID, the game's secret salt, and the version number
	//of the state. Repeated calls to Rand() on the same state will return the
	//same random generator. If games use this source for all of their
	//randomness it allows the game to be played back detrministically, which
	//is useful in some testing scenarios. Rand is only available on State,
	//not ImmutableState, because all methods that aren't mutators in your
	//game logic should be deterministic.
	Rand() *rand.Rand

	//containingStack will return the stack and slot index for the
	//associated component, if that location is not sanitized. If no error is
	//returned, stack.ComponentAt(slotIndex) == c will evaluate to true.
	containingStack(c Component) (stack Stack, slotIndex int, err error)
}

//Valid returns true if the PlayerIndex's value is legal in the context of the
//current State--that is, it is either AdminPlayerIndex, ObserverPlayerIndex,
//or between 0 (inclusive) and game.NumPlayers().
func (p PlayerIndex) Valid(state ImmutableState) bool {
	if p == AdminPlayerIndex || p == ObserverPlayerIndex {
		return true
	}
	if state == nil {
		return false
	}
	if p < 0 || int(p) >= len(state.ImmutablePlayerStates()) {
		return false
	}
	return true
}

//Next returns the next PlayerIndex, wrapping around back to 0 if it
//overflows. PlayerIndexes of AdminPlayerIndex and Observer PlayerIndex will
//not be affected.
func (p PlayerIndex) Next(state ImmutableState) PlayerIndex {
	if p == AdminPlayerIndex || p == ObserverPlayerIndex {
		return p
	}
	p++
	if int(p) >= len(state.ImmutablePlayerStates()) {
		p = 0
	}
	return p
}

//Previous returns the previous PlayerIndex, wrapping around back to len(players -1) if it
//goes below 0. PlayerIndexes of AdminPlayerIndex and Observer PlayerIndex will
//not be affected.
func (p PlayerIndex) Previous(state ImmutableState) PlayerIndex {
	if p == AdminPlayerIndex || p == ObserverPlayerIndex {
		return p
	}
	p--
	if int(p) < 0 {
		p = PlayerIndex(len(state.ImmutablePlayerStates()) - 1)
	}
	return p
}

//Equivalent checks whether the two playerIndexes are equivalent. For most
//indexes it checks if both are the same. ObserverPlayerIndex returns false
//when compared to any other PlayerIndex. AdminPlayerIndex returns true when
//compared to any other index (other than ObserverPlayerIndex). This method is
//useful for verifying that a given TargerPlayerIndex is equivalent to the
//proposer PlayerIndex in a move's Legal method. moves.CurrentPlayer handles
//that logic for you.
func (p PlayerIndex) Equivalent(other PlayerIndex) bool {

	//Sanity check obviously-illegal values
	if p < AdminPlayerIndex || other < AdminPlayerIndex {
		return false
	}

	if p == ObserverPlayerIndex || other == ObserverPlayerIndex {
		return false
	}
	if p == AdminPlayerIndex || other == AdminPlayerIndex {
		return true
	}
	return p == other
}

//String returns the int value of the PlayerIndex.
func (p PlayerIndex) String() string {
	return strconv.Itoa(int(p))
}

//componentIndexItem represents one item in the componentIndex.s
type componentIndexItem struct {
	stack     Stack
	slotIndex int
}

//state implements both State and MutableState, so it can always be passed for
//either, and what it's interpreted as is primarily a function of what the
//method signature is that it's passed to
type state struct {
	gameState              ConfigurableSubState
	playerStates           []ConfigurablePlayerState
	computedValues         *computedProperties
	dynamicComponentValues map[string][]ConfigurableSubState
	//We hang onto these because otherwise we'd have to create them on the fly
	//whenever MutablePlayerStates() and MutableDynamicComponentValues are
	//called. They're populated in setStateForSubStates.
	mutablePlayerStates           []PlayerState
	mutableDynamicComponentValues map[string][]SubState
	secretMoveCount               map[string][]int
	sanitized                     bool
	version                       int
	game                          *Game

	memoizedRand *rand.Rand

	//componentIndex keeps track of the current location of all components in
	//stacks in this state. It is not persisted, but is rebuilt the first time
	//it's asked for, and then all modifications are kept track of as things
	//move around.
	componentIndex map[Component]componentIndexItem

	//Set to true while computed is being calculating computed. Primarily so
	//if you marshal JSON in that time we know to just elide computed.
	calculatingComputed bool
	//If TimerProp.Start() is called, it prepares a timer, but doesn't
	//actually start ticking it until this state is committed. This is where
	//we accumulate the timers that still need to be fully started at that
	//point.
	timersToStart []string
}

func (s *state) Rand() *rand.Rand {
	if s.memoizedRand == nil {

		input := "insecurestarterdefault"

		if game := s.game; game != nil {
			//Sometimes, like exampleState, we don't have the game reference.
			//But those are rare and it's OK to have deterministic behavior.
			input = game.Id() + game.secretSalt
		}

		input += strconv.Itoa(s.version)

		hasher := fnv.New64()

		hasher.Write([]byte(input))

		val := hasher.Sum64()

		s.memoizedRand = rand.New(rand.NewSource(int64(val)))
	}
	return s.memoizedRand
}

func (s *state) containingImmutableStack(c Component) (stack ImmutableStack, slotIndex int, err error) {
	return s.containingStack(c)
}

func (s *state) containingStack(c Component) (stack Stack, slotIndex int, err error) {

	if s.componentIndex == nil {
		s.buildComponentIndex()
	}

	if c == nil {
		return nil, 0, errors.New("Nil component doesn't exist in any stack")
	}

	if c.Deck().GenericComponent().Equivalent(c) {
		return nil, 0, errors.New("The generic component for that deck isn't in any stack")
	}

	item, ok := s.componentIndex[c.ptr()]
	if !ok {
		//This can happen if the state is sanitized, after
		//buildComponentIndex, which won't be able to see the component.
		if s.Sanitized() {
			return nil, 0, errors.New("That component's location is not public information.")
		}
		//If this happened and the state isn't expected, then something bad happened.
		//TODO: remove this once debugging that it doesn't happen
		log.Println("WARNING: Component didn't exist in index")
		return nil, 0, errors.New("Unexpectedly that component was not found in the index")
	}

	//Sanity check that we're allowed to see that component in that location.
	otherC := item.stack.ComponentAt(item.slotIndex)

	if otherC == nil || otherC.Generic() {
		return nil, 0, errors.New("That component's location is not public information.")
	}

	//This check should always work if the stack has been sanitized, because
	//every Policy other than PolicyVisible replaces ComponentAt with generic
	//component.
	if !otherC.Equivalent(c) {
		//If this happened and the state isn't expected, then something bad happened.
		//TODO: remove this once debugging that it doesn't happen
		log.Println("WARNING: Component didn't exist; wrong component in index")
		return nil, 0, errors.New("Unexpectedly that component was not found in the index")
	}

	return item.stack, item.slotIndex, nil
}

//buildComponentIndex creates the component index by force. Should be called
//if an operation is called on the componentIndex but it's nil.
func (s *state) buildComponentIndex() {
	s.componentIndex = make(map[Component]componentIndexItem)

	if s.gameState != nil {
		s.reportComponentLocationsForReader(s.gameState.ReadSetter())
	}
	for _, player := range s.playerStates {
		if player != nil {
			s.reportComponentLocationsForReader(player.ReadSetter())
		}
	}
	for _, dynamicValues := range s.dynamicComponentValues {
		for _, value := range dynamicValues {
			if value != nil {
				s.reportComponentLocationsForReader(value.ReadSetter())
			}
		}
	}
}

//reportComponnentLocationsForReader goes through the given reader, and for
//each component it finds, reports its location into the index. Used to help
//build up the index when it's first created.
func (s *state) reportComponentLocationsForReader(readSetter PropertyReadSetter) {
	for propName, propType := range readSetter.Props() {

		if !readSetter.PropMutable(propName) {
			continue
		}

		if propType == TypeStack {
			stack, err := readSetter.StackProp(propName)
			if err != nil {
				continue
			}
			for i, c := range stack.Components() {
				//can't use updateIndexForAllComponents because we don't want
				//to clal buildComponents.
				s.componentAddedImpl(c, stack, i)
			}
		} else if propType == TypeBoard {
			board, err := readSetter.BoardProp(propName)
			if err != nil {
				continue
			}
			for _, stack := range board.Spaces() {
				//can't use updateIndexForAllComponents because we don't want
				//to clal buildComponents.
				for i, c := range stack.Components() {
					s.componentAddedImpl(c, stack, i)
				}
			}
		}
	}
}

func (s *state) componentAddedImpl(c Component, stack Stack, slotIndex int) {
	if c == nil {
		return
	}
	if c.Deck() != nil && c.Deck().GenericComponent().Equivalent(c) {
		return
	}
	s.componentIndex[c.ptr()] = componentIndexItem{
		stack,
		slotIndex,
	}
}

//componetAdded should be called by stacks when a component is added to them,
//by non-merged stacks.
func (s *state) componentAdded(c Component, stack Stack, slotIndex int) {
	if s.componentIndex == nil {
		s.buildComponentIndex()
	}

	s.componentAddedImpl(c, stack, slotIndex)
}

func (s *state) updateIndexForAllComponents(stack Stack) {
	for i, c := range stack.Components() {
		s.componentAdded(c, stack, i)
	}
}

func (s *state) Version() int {
	return s.version
}

func (s *state) GameState() SubState {
	return s.gameState
}

func (s *state) PlayerStates() []PlayerState {
	return s.mutablePlayerStates
}

func (s *state) DynamicComponentValues() map[string][]SubState {
	return s.mutableDynamicComponentValues
}

func (s *state) Game() *Game {
	return s.game
}

func (s *state) ImmutableGameState() ImmutableSubState {
	return s.gameState
}

func (s *state) ImmutablePlayerStates() []ImmutablePlayerState {
	result := make([]ImmutablePlayerState, len(s.playerStates))
	for i := 0; i < len(s.playerStates); i++ {
		result[i] = s.playerStates[i]
	}
	return result
}

func (s *state) ImmutableCurrentPlayer() ImmutablePlayerState {
	return s.CurrentPlayer()
}

func (s *state) CurrentPlayer() PlayerState {
	index := s.CurrentPlayerIndex()
	if index < 0 || int(index) >= len(s.playerStates) {
		return nil
	}
	return s.playerStates[index]
}

func (s *state) CurrentPlayerIndex() PlayerIndex {
	return s.game.manager.delegate.CurrentPlayerIndex(s)
}

func (s *state) Copy(sanitized bool) (ImmutableState, error) {
	//TODO: just make copy() be public
	return s.copy(sanitized)
}

func (s *state) copy(sanitized bool) (*state, error) {

	result, err := s.game.manager.emptyState(len(s.playerStates))

	if err != nil {
		return nil, err
	}

	moveCounts := make(map[string][]int)

	for deck, counts := range s.secretMoveCount {
		newCounts := make([]int, len(counts))
		for i, count := range counts {
			newCounts[i] = count
		}
		moveCounts[deck] = newCounts
	}

	result.secretMoveCount = moveCounts
	result.sanitized = sanitized
	result.version = s.version
	result.game = s.game
	//We copy this over, because this should only be set when computed is
	//being calculated, and during that time we'll be creating sanitized
	//copies of ourselves. However, if there are other copies created when
	//this flag is set that outlive the original flag being unset, that
	//state would be in a bad state long term...
	result.calculatingComputed = s.calculatingComputed

	//Note: we can't copy componentIndex, because all of those items point to
	//MutableStacks in the original state, and we don't have an easy way to
	//figure out which ones they correspond to in the new one.

	if err := copyReader(s.gameState.ReadSetter(), result.gameState.ReadSetter()); err != nil {
		return nil, err
	}

	for i := 0; i < len(s.playerStates); i++ {
		if err := copyReader(s.playerStates[i].ReadSetter(), result.playerStates[i].ReadSetter()); err != nil {
			return nil, err
		}
	}

	for deckName, values := range s.dynamicComponentValues {
		for i := 0; i < len(values); i++ {
			if err := copyReader(s.dynamicComponentValues[deckName][i].ReadSetter(), result.dynamicComponentValues[deckName][i].ReadSetter()); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

//finish should be called when the state has all of its sub-states set. It
//goes through each subState on s and calls SetState on it, and also sets the
//mutable*States once.
func (s *state) setStateForSubStates() {

	s.gameState.SetState(s)
	s.gameState.SetImmutableState(s)

	for i := 0; i < len(s.playerStates); i++ {
		s.playerStates[i].SetState(s)
		s.playerStates[i].SetImmutableState(s)

	}

	for _, dynamicComponents := range s.dynamicComponentValues {
		for _, component := range dynamicComponents {
			component.SetState(s)
			component.SetImmutableState(s)
		}
	}

	mutablePlayerStates := make([]PlayerState, len(s.playerStates))
	for i := 0; i < len(s.playerStates); i++ {
		mutablePlayerStates[i] = s.playerStates[i]
	}

	s.mutablePlayerStates = mutablePlayerStates

	dynamicComponentValues := make(map[string][]SubState)

	for key, arr := range s.dynamicComponentValues {
		resultArr := make([]SubState, len(arr))
		for i := 0; i < len(arr); i++ {
			resultArr[i] = arr[i]
		}
		dynamicComponentValues[key] = resultArr
	}

	s.mutableDynamicComponentValues = dynamicComponentValues
}

//validateBeforeSave insures that for all readers, the playerIndexes are
//valid, and the stacks are too.
func (s *state) validateBeforeSave() error {

	if err := validateReaderBeforeSave(s.GameState().Reader(), "Game", s); err != nil {
		return err
	}

	for i, player := range s.PlayerStates() {
		if err := validateReaderBeforeSave(player.Reader(), "Player "+strconv.Itoa(i), s); err != nil {
			return err
		}
	}

	for name, deck := range s.DynamicComponentValues() {
		for i, values := range deck {
			if err := validateReaderBeforeSave(values.Reader(), "DynamicComponentValues "+name+" "+strconv.Itoa(i), s); err != nil {
				return err
			}
		}
	}

	//If delegate.PhaseEnum returns a tree, ensure it's in a leaf state.

	delegate := s.Game().Manager().Delegate()

	e := delegate.PhaseEnum()

	if e == nil {
		return nil
	}

	t := e.TreeEnum()

	if t == nil {
		return nil
	}

	if t.IsLeaf(delegate.CurrentPhase(s)) {
		return nil
	}

	return errors.New("PhaseEnum is a TreeEnum, but CurrentPhase is not a leaf value.")
}

func validateReaderBeforeSave(reader PropertyReader, name string, state State) error {

	for propName, propType := range reader.Props() {
		if propType == TypePlayerIndex {
			val, err := reader.PlayerIndexProp(propName)
			if err != nil {
				return errors.New("Error reading property " + propName + ": " + err.Error())
			}
			if !val.Valid(state) {
				return errors.New(propName + " was an invalid PlayerIndex, with value " + strconv.Itoa(int(val)))
			}
		}
		if propType == TypeStack {
			stack, err := reader.ImmutableStackProp(propName)
			if err != nil {
				return errors.New("Error reading property " + propName + ": " + err.Error())
			}
			if merged := stack.MergedStack(); merged != nil {
				if err := merged.Valid(); err != nil {
					return errors.New(propName + " was a merged stack that did not validate: " + err.Error())
				}
			}
		}
		//We don't need to check TypeBoard here, because TypeBoard never has
		//merged stacks within it, and those are the only ones who could be invalid here.
	}

	return nil
}

//committed is called right after the state has been committed to the database
//and we're sure it will stick. This is the time to do any actions that were
//triggered during the state manipulation. currently that is only timers.
func (s *state) committed() {
	for _, id := range s.timersToStart {
		s.game.manager.timers.StartTimer(id)
	}
}

func (s *state) StorageRecord() StateStorageRecord {
	record, _ := s.customMarshalJSON(false, true)
	return record
}

func (s *state) customMarshalJSON(includeComputed bool, indent bool) ([]byte, error) {
	obj := map[string]interface{}{
		"Game":    s.gameState,
		"Players": s.playerStates,
		"Version": s.version,
	}

	if includeComputed {
		obj["Computed"] = s.computed()
	}

	//We emit the secretMoveCount only when the state isn't sanitized. Any
	//time the state is sent via StateForPlayer sanitized will be true, so
	//this has the effect of persisting SecretMoveCount when serialized for
	//storage layer, but not when sanitized state.
	if !s.sanitized {
		if len(s.secretMoveCount) > 0 {
			obj["SecretMoveCount"] = s.secretMoveCount
		}
	}

	dynamic := s.DynamicComponentValues()

	if dynamic != nil && len(dynamic) != 0 {
		obj["Components"] = dynamic
	} else {
		obj["Components"] = map[string]interface{}{}
	}

	if indent {
		return DefaultMarshalJSON(obj)
	}

	return json.Marshal(obj)

}

func (s *state) MarshalJSON() ([]byte, error) {
	return s.customMarshalJSON(true, false)
}

func (s *state) Diagram() string {
	return s.game.manager.delegate.Diagram(s)
}

func (s *state) Sanitized() bool {
	return s.sanitized
}

func (s *state) ImmutableDynamicComponentValues() map[string][]ImmutableSubState {

	result := make(map[string][]ImmutableSubState)

	for key, val := range s.dynamicComponentValues {
		slice := make([]ImmutableSubState, len(val))
		for i := 0; i < len(slice); i++ {
			slice[i] = val[i]
		}
		result[key] = slice
	}

	return result
}

func (s *state) computed() *computedProperties {

	if s.calculatingComputed {
		//This might be called in a Compute() callback either directly, or
		//implicitly via MarshalJSON.
		return nil
	}

	if s.computedValues == nil {

		s.calculatingComputed = true

		playerProperties := make([]PropertyCollection, len(s.playerStates))

		for i, player := range s.playerStates {
			playerProperties[i] = s.game.manager.delegate.ComputedPlayerProperties(player)
		}

		s.computedValues = &computedProperties{
			Global:  s.game.manager.delegate.ComputedGlobalProperties(s),
			Players: playerProperties,
		}

		s.calculatingComputed = false
	}
	return s.computedValues
}

//SanitizedForPlayer is in sanitized.go

//Reader is the interface to fetch a PropertyReader from an object. See
//ConfigurableSubState and PropertyReadSetConfigurer for more.
type Reader interface {
	Reader() PropertyReader
}

//ReadSetter is the interface to fetch a PropertyReadSetter from an object.
//See ConfigurableSubState and PropertyReadSetConfigurer for more.
type ReadSetter interface {
	Reader
	ReadSetter() PropertyReadSetter
}

//ReadSetConfigurer is the interface to fetch a PropertyReadSetConfigurer from
//an object. See ConfigurableSubState and PropertyReadSetConfigurer for more.
type ReadSetConfigurer interface {
	ReadSetter
	ReadSetConfigurer() PropertyReadSetConfigurer
}

//ImmutableStateSetter is included in ImmutableSubState, SubState, and
//ConfigureableSubState as the way to keep track of which ImmutableState a
//given SubState is part of. See also StateSetter, which adds getters/setters
//for mutable States. Typically you use base.SubState to implement this
//automatically.
type ImmutableStateSetter interface {
	//SetImmutableState is called to give the SubState object a pointer back
	//to the State that contains it. You can implement it yourself, or
	//anonymously embed base.SubState to get it for free.
	SetImmutableState(state ImmutableState)
	//ImmutableState() returns the state that was set via SetState().
	ImmutableState() ImmutableState
}

//StateSetter is included in SubState and ConfigureableSubState as the way to
//keep track of which State a given SubState is part of. See also
//ImmutableStateSetter, which adds getters/setters for ImmutableStates.
//Typically you use base.SubState to implement this automatically.
type StateSetter interface {
	ImmutableStateSetter
	SetState(state State)
	State() State
}

//ImmutableSubState is the interface that all non-modifiable sub-state objects
//(PlayerStates. GameStates, and DynamicComponentValues) implement. It is like
//SubState, but minus any mutator methods. See ConfigurableSubState for more
//on the SubState type hierarchy.
type ImmutableSubState interface {
	ImmutableStateSetter
	Reader
}

//ImmutableSubState is the interface that mutable {Game,Player}State's
//implement.
type SubState interface {
	StateSetter
	ReadSetter
}

/*

ConfigurableSubState is the interface for many types of structs that store
properties and configuration specific to your game type. The values returned
from your GameDelegate's GameStateConstructor, PlayerStateConstructor, and
DynamicComponentValues constructor must all implement this interface.
(PlayerStateConstructor also adds PlayerIndex())

A ConfigurableSubState is a struct that has a collection of properties all of
a given small set of legal types, enumerated in PropertyType. These are the
core objects to maintain state in your game type. The types of properties on
these objects are strictly defined to ensure the shapes of the objects are
simple and knowable.

The engine in general doesn't know the shape of your underlying structs, so it
uses the ProeprtyReadSetConfigurer interface to interact with your objects.
See the documetnation for PropertyReadSetConfigurer for more.

Many legal property types, like string and int, are simple and can be Read and
Set as you'd expect. But some, called interface types, are more complex
because they denote objects that carry configuration information in their
instantiation. Stacks, Timers, and Enums are examples of these. These
interface types can be Read and have their sub-properties Set. But they also
must be able to be Configured, which is to say instantied and set onto the
underlying struct.

ConfigurableSubState is the most powerful interface for interacting with these
types of objects, because it has methods to Read, Set, and Configure all
properties. In certain cases, however, for example with an ImmutableState, it
might not be appropriate to allow Setting or Configuring propeties. For this
reason, the interfaces are split into a series of layers, building up from
only Reader methods up to adding Set proeprties, and then terminating by
layering on Configure methods.

ConfigurablePlayerSubState is an interface that extends ConfigurableSubState
with one extra method, PlayerIndex(). There are also player-state versions for
SubState and ImmutableSubState.

Typically your game's sub-states satisfy this interface by embedding
base.SubState, and then using `boardgame-util codegen` to generate the
underlying code for the PropertyReadSetConfigurer for your object type.

*/
type ConfigurableSubState interface {
	//Every SubState should be able to have its containing State set and read
	//back, so each sub-state knows how to reach up and over into other parts
	//of the over-arching state. You can implement this interface by emedding
	//base.SubState in your struct.
	StateSetter

	//ReadSetConfigurer defines the method to retrieve the
	//PropertyReadSetConfigurer for this object type. Typically this getter--
	//and the underlying PropertyReadSetConfigurer it returns--are generated
	//via `boardgame-util codegen`.
	ReadSetConfigurer
}

//PlayerIndexer is implemented by all PlayerStates, which differentiates them
//from a generic SubState.
type PlayerIndexer interface {
	//PlayerIndex encodes the index this user's state is in the containing
	//state object, allowing the SubState to know how to fetch itself from its
	//containing State.
	PlayerIndex() PlayerIndex
}

//PlayerState represents the state of a game associated with a specific user.
//It is just a SubState with the addition of a PlayerIndex(). See
//ConfigurableSubState for more on the SubState type hierarchy.
type PlayerState interface {
	PlayerIndexer
	SubState
}

//ImmutablePlayerState represents a PlayerState SubState that is not in a
//context where mutating is legal. It is simply an ImmutableSubState that also
//has a PlayerIndex method. See more on substates at the documentation for
//ConfigurableSubState.
type ImmutablePlayerState interface {
	PlayerIndexer
	ImmutableSubState
}

//A ConfigurablePlayerState is a PlayerState that is allowed to be mutated and
//configured. It is simply a ConfigurableSubState that also has a
//PlayerIndex() method. See ConfigurableSubState for more on this hierarchy of
//objects.
type ConfigurablePlayerState interface {
	PlayerIndexer
	ConfigurableSubState
}

//DefaultMarshalJSON is a simple wrapper around json.MarshalIndent, with the
//right defaults set. If your structs need to implement MarshaLJSON to output
//JSON, use this to encode it.
func DefaultMarshalJSON(obj interface{}) ([]byte, error) {
	return json.MarshalIndent(obj, "", "  ")
}
