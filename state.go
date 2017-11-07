package boardgame

import (
	"encoding/json"
	"github.com/jkomoros/boardgame/errors"
	"strconv"
)

//State represents the entire semantic state of a game at a given version. For
//your specific game, GameState and PlayerStates will actually be concrete
//structs to your particular game. Games often define a top-level
//concreteStates() *myGameState, []*myPlayerState so at the top of methods
//that accept a State they can quickly get concrete, type-checked types with
//only a single conversion leap of faith at the top. States are intended to be
//read-only; methods where you are allowed to mutate the state (e.g.
//Move.Apply()) will take a MutableState instead as a signal that it is
//permissable to modify the state. That is why the states only return non-
//mutable states (PropertyReaders, not PropertyReadSetters, although
//realistically it is possible to cast them and modify directly. The
//MarshalJSON output of a State is appropriate for sending to a client or
//serializing a state to be put in storage. Given a blob serialized in that
//fashion, GameManager.StateFromBlob will return a state.
type State interface {
	//GameState returns the GameState for this State
	GameState() SubState
	//PlayerStates returns a slice of all PlayerStates for this State
	PlayerStates() []PlayerState
	//DynamicComponentValues returns a map of deck name to array of component
	//values, one per component in that deck.
	DynamicComponentValues() map[string][]SubState

	//CurrentPlayer returns the PlayerState corresponding to the result of
	//delegate.CurrentPlayerIndex(), or nil if the index isn't valid.
	CurrentPlayer() PlayerState
	//CurrentPlayerIndex is a simple convenience wrapper around
	//delegate.CurrentPlayerIndex for this state.
	CurrentPlayerIndex() PlayerIndex

	//Version returns the version number the state is (or will be once
	//committed).
	Version() int

	//Copy returns a deep copy of the State, including copied version of the Game
	//and Player States.
	Copy(sanitized bool) State
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
	//construct the effective policy to apply. See the package level comment
	//for an overview of how state sanitization works.
	SanitizedForPlayer(player PlayerIndex) State

	//Game is the Game that this state is part of. Calling
	//Game.State(s.Version()) should return a state equivalent to this State
	//(module sanitization, if applied).
	Game() *Game

	//StorageRecord returns a StateStorageRecord representing the state.
	StorageRecord() StateStorageRecord
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

//A StatePropertyRef is a reference to a particular property in a State, in a
//structured way. Currently used when defining your dependencies for computed
//properties.
type StatePropertyRef struct {
	Group StateGroupType
	//DeckName is only used when Group is StateGroupDynamicComponentValues
	DeckName string
	//PropName is the specific property on the given SubStateObject specified
	//by the rest of the StatePropertyRef.
	PropName string
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

//AdminPlayerINdex is a special PlayerIndex that denotes the omniscient admin
//who can see all state and make moves whenever they want. This PlayerIndex
//should only be used in rare or debug circumstances.
const AdminPlayerIndex PlayerIndex = -2

//A MutableState is a state that is designed to be modified in place. These
//are passed to methods (instead of normal States) as a signal that
//modifications are intended to be done on the state.
type MutableState interface {
	//MutableState contains all of the methods of a read-only state.
	State
	//MutableGameState is a reference to the MutableGameState for this MutableState.
	MutableGameState() MutableSubState
	//MutablePlayerstates returns a slice of MutablePlayerStates for this MutableState.
	MutablePlayerStates() []MutablePlayerState

	MutableDynamicComponentValues() map[string][]MutableSubState
}

//Valid returns true if the PlayerIndex's value is legal in the context of the
//current State.
func (p PlayerIndex) Valid(state State) bool {
	if p == AdminPlayerIndex || p == ObserverPlayerIndex {
		return true
	}
	if state == nil {
		return false
	}
	if p < 0 || int(p) >= len(state.PlayerStates()) {
		return false
	}
	return true
}

//Next returns the next PlayerIndex, wrapping around back to 0 if it
//overflows. PlayerIndexes of AdminPlayerIndex and Observer PlayerIndex will
//not be affected.
func (p PlayerIndex) Next(state State) PlayerIndex {
	if p == AdminPlayerIndex || p == ObserverPlayerIndex {
		return p
	}
	p++
	if int(p) >= len(state.PlayerStates()) {
		p = 0
	}
	return p
}

//Previous returns the previous PlayerIndex, wrapping around back to len(players -1) if it
//goes below 0. PlayerIndexes of AdminPlayerIndex and Observer PlayerIndex will
//not be affected.
func (p PlayerIndex) Previous(state State) PlayerIndex {
	if p == AdminPlayerIndex || p == ObserverPlayerIndex {
		return p
	}
	p--
	if int(p) < 0 {
		p = PlayerIndex(len(state.PlayerStates()) - 1)
	}
	return p
}

//Equivalent checks whether the two playerIndexes are equivalent. For most
//indexes it checks if both are the same. ObserverPlayerIndex returns false
//when compared to any other PlayerIndex. AdminPlayerIndex returns true when
//compared to any other index (other than ObserverPlayerIndex). This method is
//useful for verifying that a given TargerPlayerIndex is equivalent to the
//proposer PlayerIndex in a move's Legal method.
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

func (p PlayerIndex) String() string {
	return strconv.Itoa(int(p))
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
	//called. They're populated in state.finish().
	mutablePlayerStates           []MutablePlayerState
	mutableDynamicComponentValues map[string][]MutableSubState
	secretMoveCount               map[string][]int
	sanitized                     bool
	version                       int
	game                          *Game
	//Set to true while computed is being calculating computed. Primarily so
	//if you marshal JSON in that time we know to just elide computed.
	calculatingComputed bool
	//If TimerProp.Start() is called, it prepares a timer, but doesn't
	//actually start ticking it until this state is committed. This is where
	//we accumulate the timers that still need to be fully started at that
	//point.
	timersToStart []int
}

func (s *state) Version() int {
	return s.version
}

func (s *state) MutableGameState() MutableSubState {
	return s.gameState
}

func (s *state) MutablePlayerStates() []MutablePlayerState {
	return s.mutablePlayerStates
}

func (s *state) MutableDynamicComponentValues() map[string][]MutableSubState {
	return s.mutableDynamicComponentValues
}

func (s *state) Game() *Game {
	return s.game
}

func (s *state) GameState() SubState {
	return s.gameState
}

func (s *state) PlayerStates() []PlayerState {
	result := make([]PlayerState, len(s.playerStates))
	for i := 0; i < len(s.playerStates); i++ {
		result[i] = s.playerStates[i]
	}
	return result
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

func (s *state) Copy(sanitized bool) State {
	return s.copy(sanitized)
}

func (s *state) copy(sanitized bool) *state {
	players := make([]ConfigurablePlayerState, len(s.playerStates))

	for i, player := range s.playerStates {
		players[i] = s.copyPlayerState(player)
	}

	result := &state{
		gameState:              s.copyGameState(s.gameState),
		playerStates:           players,
		dynamicComponentValues: make(map[string][]ConfigurableSubState),
		secretMoveCount:        make(map[string][]int),
		sanitized:              sanitized,
		version:                s.version,
		game:                   s.game,
		//We copy this over, because this should only be set when computed is
		//being calculated, and during that time we'll be creating sanitized
		//copies of ourselves. However, if there are other copies created when
		//this flag is set that outlive the original flag being unset, that
		//state would be in a bad state long term...
		calculatingComputed: s.calculatingComputed,
	}

	for deckName, values := range s.dynamicComponentValues {
		arr := make([]ConfigurableSubState, len(values))
		for i := 0; i < len(values); i++ {
			arr[i] = s.copyDynamicComponentValues(values[i], deckName)
			if err := setReaderStatePtr(arr[i].Reader(), result); err != nil {
				return nil
			}
		}
		result.dynamicComponentValues[deckName] = arr
	}

	result.setStateForSubStates()

	//FixUp stacks to make sure they point to this new state.
	if err := setReaderStatePtr(result.gameState.Reader(), result); err != nil {
		return nil
	}
	for _, player := range result.playerStates {
		if err := setReaderStatePtr(player.Reader(), result); err != nil {
			return nil
		}
	}

	return result
}

//finish should be called when the state has all of its sub-states set. It
//goes through each subState on s and calls SetState on it, and also sets the
//mutable*States once.
func (s *state) setStateForSubStates() {
	s.gameState.SetState(s)
	s.gameState.SetMutableState(s)
	for i := 0; i < len(s.playerStates); i++ {
		s.playerStates[i].SetState(s)
		s.playerStates[i].SetMutableState(s)
	}
	for _, dynamicComponents := range s.dynamicComponentValues {
		for _, component := range dynamicComponents {
			component.SetState(s)
			component.SetMutableState(s)
		}
	}

	mutablePlayerStates := make([]MutablePlayerState, len(s.playerStates))
	for i := 0; i < len(s.playerStates); i++ {
		mutablePlayerStates[i] = s.playerStates[i]
	}

	s.mutablePlayerStates = mutablePlayerStates

	dynamicComponentValues := make(map[string][]MutableSubState)

	for key, arr := range s.dynamicComponentValues {
		resultArr := make([]MutableSubState, len(arr))
		for i := 0; i < len(arr); i++ {
			resultArr[i] = arr[i]
		}
		dynamicComponentValues[key] = resultArr
	}

	s.mutableDynamicComponentValues = dynamicComponentValues
}

func (s *state) copyDynamicComponentValues(input MutableSubState, deckName string) ConfigurableSubState {
	deck := s.game.manager.chest.Deck(deckName)
	if deck == nil {
		s.game.manager.Logger().WithField("deckname", deckName).Error("Invalid deck")
		return nil
	}
	output := s.game.Manager().delegate.DynamicComponentValuesConstructor(deck)

	inputReader := input.ReadSetter()

	if inputReader == nil {
		s.game.manager.Logger().Error("InputReader was unexpectedly nil")
		return nil
	}

	outputReadSetConfigurer := output.ReadSetConfigurer()

	if outputReadSetConfigurer == nil {
		s.game.manager.Logger().Error("Output read set configurer was unexpectedly nil")
		return nil
	}

	if err := copyReader(inputReader, outputReadSetConfigurer); err != nil {
		s.game.manager.Logger().Error("WARNING: couldn't copy dynamic value state: " + err.Error())
	}
	return output
}

func (s *state) copyPlayerState(input MutablePlayerState) ConfigurablePlayerState {
	output := s.game.manager.delegate.PlayerStateConstructor(input.PlayerIndex())

	inputReader := input.ReadSetter()

	if inputReader == nil {
		s.game.manager.Logger().Error("InputReader was unexpectedly nil")
		return nil
	}

	outputReadSetConfigurer := output.ReadSetConfigurer()

	if outputReadSetConfigurer == nil {
		s.game.manager.Logger().Error("Output read set configurer was unexpectedly nil")
		return nil
	}

	if err := copyReader(inputReader, outputReadSetConfigurer); err != nil {
		s.game.manager.Logger().Error("WARNING: couldn't copy player state: " + err.Error())
	}

	return output
}

func (s *state) copyGameState(input MutableSubState) ConfigurableSubState {
	output := s.game.manager.delegate.GameStateConstructor()

	inputReader := input.ReadSetter()

	if inputReader == nil {
		s.game.manager.Logger().Error("InputReader was unexpectedly nil")
		return nil
	}

	outputReadSetConfigurer := output.ReadSetConfigurer()

	if outputReadSetConfigurer == nil {
		s.game.manager.Logger().Error("Output read set configurer was unexpectedly nil")
		return nil
	}

	if err := copyReader(inputReader, outputReadSetConfigurer); err != nil {
		s.game.manager.Logger().Error("WARNING: couldn't copy game state: " + err.Error())
	}
	return output
}

//validatePlayerIndexes checks all of the PropertyReaders in State and
//verifies that PlayerIndexes are within legal bounds.
func (s *state) validatePlayerIndexes() error {

	var errs []error

	errs = append(errs, validatePlayerIndexesForReader(s.GameState().Reader(), "Game", s))

	for i, player := range s.PlayerStates() {
		errs = append(errs, validatePlayerIndexesForReader(player.Reader(), "Player "+strconv.Itoa(i), s))
	}

	for name, deck := range s.DynamicComponentValues() {
		for i, values := range deck {
			errs = append(errs, validatePlayerIndexesForReader(values.Reader(), "DynamicComponentValues "+name+" "+strconv.Itoa(i), s))
		}
	}

	//TODO: check computed, too?

	var firstErr error
	var numErrors int

	for _, err := range errs {
		if err == nil {
			continue
		}
		if firstErr == nil {
			firstErr = err
		}
		numErrors++
	}

	if firstErr != nil {
		return errors.New("Found " + strconv.Itoa(numErrors) + " PlayerIndexes that were not valid. For example, " + firstErr.Error())
	}

	return nil
}

func validatePlayerIndexesForReader(reader PropertyReader, name string, state State) error {

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

func (s *state) DynamicComponentValues() map[string][]SubState {

	result := make(map[string][]SubState)

	for key, val := range s.dynamicComponentValues {
		slice := make([]SubState, len(val))
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

//Reader is the interface that any object that can return a PropertyReader
//implements.
type Reader interface {
	Reader() PropertyReader
}

//ReadSetter is the interface that any object that can return a
//PropertyReadSetter implements. Objects that implement ReadSetter also
//implement Reader.
type ReadSetter interface {
	Reader
	ReadSetter() PropertyReadSetter
}

//ReadSetCongigurer is the interface that any object that can return a
//PropertyReadSetConfigurer implements. Objects that implement
//ReadSetConfigurer also implement ReadSetter.
type ReadSetConfigurer interface {
	ReadSetter
	ReadSetConfigurer() PropertyReadSetConfigurer
}

//StateSetter is included in SubState, MutableSubState, and
//ConfigureableSubState as the way to keep track of which State a given
//SubState is part of.
type StateSetter interface {
	//SetState is called to give the SubState object a pointer back to the
	//State that contains it. You can implement it yourself, or anonymously
	//embed BaseSubState to get it for free.
	SetState(state State)
	//State() returns the state that was set via SetState().
	State() State
}

//MutableStateSetter is like StateSetter but it also includes Mutable methods.
type MutableStateSetter interface {
	StateSetter
	SetMutableState(state MutableState)
	MutableState() MutableState
}

//SubState is the interface that all sub-state objects (PlayerStates.
//GameStates, and DynamicComponentValues) implement.
type SubState interface {
	StateSetter
	Reader
}

//MutableSubState is the interface that Mutable{Game,Player}State's
//implement.
type MutableSubState interface {
	MutableStateSetter
	ReadSetter
}

//ConfigurableSubState is the interface that Configurable{Game,Player}State's
//implement.
type ConfigurableSubState interface {
	MutableStateSetter
	ReadSetConfigurer
}

//PlayerIndexer is implemented by all PlayerStates, which differentiates them
//from a generic SubState.
type PlayerIndexer interface {
	//PlayerIndex encodes the index this user's state is in the containing
	//state object.
	PlayerIndex() PlayerIndex
}

//PlayerState represents the state of a game associated with a specific user.
//It is just a SubState with the addition of a PlayerIndex().
type PlayerState interface {
	PlayerIndexer
	SubState
}

//A MutablePlayerState is a PlayerState that is allowed to be mutated.
type MutablePlayerState interface {
	PlayerIndexer
	MutableSubState
}

//A ConfigurablePlayerState is a PlayerState that is allowed to be mutated and
//configured.
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

//BaseSubState is a simple struct designed to be anonymously embedded in the
//SubStates you create, so you don't have to implement SetState yourself.
type BaseSubState struct {
	//Ugh it's really annoying to have to hold onto the same state in two
	//references...
	state        State
	mutableState MutableState
}

func (b *BaseSubState) SetState(state State) {
	b.state = state
}

func (b *BaseSubState) State() State {
	return b.state
}

func (b *BaseSubState) SetMutableState(state MutableState) {
	b.mutableState = state
}

func (b *BaseSubState) MutableState() MutableState {
	return b.mutableState
}
