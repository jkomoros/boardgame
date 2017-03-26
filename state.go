package boardgame

import (
	"encoding/json"
)

//State represents the entire semantic state of a game at a given version. For
//your specific game, Game and Players will actually be concrete structs to
//your particular game. Games often define a top-level concreteStates()
//*myGameState, []*myPlayerState so at the top of methods that accept a State
//they can quickly get concrete, type-checked types with only a single
//conversion leap of faith at the top. States are intended to be read-only;
//methods where you are allowed to mutate the state (e.g. Move.Apply()) will
//take a MutableState instead as a signal that it is permissable to modify the
//state. That is why the states only return non-mutable states
//(PropertyReaders, not PropertyReadSetters, although realistically it is
//possible to cast them and modify directly. The MarshalJSON output of a State
//is appropriate for sending to a client or serializing a state to be put in
//storage. Given a blob serialized in that fashion, GameManager.StateFromBlob
//will return a state.
type State interface {
	//Game returns the GameState for this State
	Game() GameState
	//Players returns a slice of all PlayerStates for this State
	Players() []PlayerState
	//DynamicComponentValues returns a map of deck name to array of component
	//values, one per component in that deck.
	DynamicComponentValues() map[string][]DynamicComponentValues
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
	Computed() ComputedProperties
	//SanitizedForPlayer produces a copy state object that has been sanitized for
	//the player at the given index. The state object returned will have
	//Sanitized() return true. Will call GameDelegate.StateSanitizationPolicy to
	//retrieve the policy in place. See the package level comment for an overview
	//of how state sanitization works.
	SanitizedForPlayer(playerIndex int) State
}

//A MutableState is a state that is designed to be modified in place. These
//are passed to methods (instead of normal States) as a signal that
//modifications are intended to be done on the state.
type MutableState interface {
	//MutableState contains all of the methods of a read-only state.
	State
	//MutableGame is a reference to the MutableGameState for this MutableState.
	MutableGame() MutableGameState
	//MutablePlayers returns a slice of MutablePlayerStates for this MutableState.
	MutablePlayers() []MutablePlayerState

	MutableDynamicComponentValues() map[string][]MutableDynamicComponentValues
}

//state implements both State and MutableState, so it can always be passed for
//either, and what it's interpreted as is primarily a function of what the
//method signature is that it's passed to
type state struct {
	game                   MutableGameState
	players                []MutablePlayerState
	computed               *computedPropertiesImpl
	dynamicComponentValues map[string][]MutableDynamicComponentValues
	sanitized              bool
	delegate               GameDelegate
}

func (s *state) MutableGame() MutableGameState {
	return s.game
}

func (s *state) MutablePlayers() []MutablePlayerState {
	return s.players
}

func (s *state) MutableDynamicComponentValues() map[string][]MutableDynamicComponentValues {
	return s.dynamicComponentValues
}

func (s *state) Game() GameState {
	return s.game
}

func (s *state) Players() []PlayerState {
	result := make([]PlayerState, len(s.players))
	for i := 0; i < len(s.players); i++ {
		result[i] = s.players[i]
	}
	return result
}

func (s *state) Copy(sanitized bool) State {
	return s.copy(sanitized)
}

func (s *state) copy(sanitized bool) *state {
	players := make([]MutablePlayerState, len(s.players))

	for i, player := range s.players {
		players[i] = player.MutableCopy()
	}

	result := &state{
		game:                   s.game.MutableCopy(),
		players:                players,
		dynamicComponentValues: make(map[string][]MutableDynamicComponentValues),
		sanitized:              sanitized,
		delegate:               s.delegate,
	}

	for deckName, values := range s.dynamicComponentValues {
		arr := make([]MutableDynamicComponentValues, len(values))
		for i := 0; i < len(values); i++ {
			arr[i] = values[i].Copy()
			if err := verifyReaderStacks(arr[i].Reader(), result); err != nil {
				return nil
			}
		}
		result.dynamicComponentValues[deckName] = arr
	}

	//FixUp stacks to make sure they point to this new state.
	if err := verifyReaderStacks(result.game.Reader(), result); err != nil {
		return nil
	}
	for _, player := range result.players {
		if err := verifyReaderStacks(player.Reader(), result); err != nil {
			return nil
		}
	}

	return result
}

func (s *state) MarshalJSON() ([]byte, error) {
	obj := map[string]interface{}{
		"Game":     s.game,
		"Players":  s.players,
		"Computed": s.Computed(),
	}

	dynamic := s.DynamicComponentValues()

	if dynamic != nil && len(dynamic) != 0 {
		obj["Components"] = dynamic
	}

	return json.Marshal(obj)
}

func (s *state) Diagram() string {
	return s.delegate.Diagram(s)
}

func (s *state) Sanitized() bool {
	return s.sanitized
}

func (s *state) DynamicComponentValues() map[string][]DynamicComponentValues {

	result := make(map[string][]DynamicComponentValues)

	for key, val := range s.dynamicComponentValues {
		slice := make([]DynamicComponentValues, len(val))
		for i := 0; i < len(slice); i++ {
			slice[i] = val[i]
		}
		result[key] = slice
	}

	return result
}

func (s *state) Computed() ComputedProperties {
	if s.computed == nil {
		s.computed = newComputedPropertiesImpl(s.delegate.ComputedPropertiesConfig(), s)
	}
	return s.computed
}

//SanitizedForPlayer is in sanitized.go

//BaseState is the interface that all state objects--PlayerStates and GameStates
//--implement.
type BaseState interface {
	Reader() PropertyReader
}

//MutableBaseState is the interface that Mutable{Game,Player}State's
//implement.
type MutableBaseState interface {
	ReadSetter() PropertyReadSetter
}

//PlayerState represents the state of a game associated with a specific user.
type PlayerState interface {
	//PlayerIndex encodes the index this user's state is in the containing
	//state object.
	PlayerIndex() int
	//Copy produces a copy of our current state. Be sure it's a deep copy that
	//makes a copy of any pointer arguments.
	Copy() PlayerState
	BaseState
}

//A MutablePlayerState is a PlayerState that is allowed to be mutated.
type MutablePlayerState interface {
	PlayerState
	MutableCopy() MutablePlayerState
	MutableBaseState
}

//GameState represents the state of a game that is not associated with a
//particular user. For example, the draw stack of cards, who the current
//player is, and other properites.
type GameState interface {
	//Copy returns a copy of our current state. Be sure it's a deep copy that
	//makes a copy of any pointer arguments.
	Copy() GameState
	BaseState
}

//A MutableGameState is a GameState that is allowed to be mutated.
type MutableGameState interface {
	GameState
	MutableCopy() MutableGameState
	MutableBaseState
}

//DefaultMarshalJSON is a simple wrapper around json.MarshalIndent, with the
//right defaults set. If your structs need to implement MarshaLJSON to output
//JSON, use this to encode it.
func DefaultMarshalJSON(obj interface{}) ([]byte, error) {
	return json.MarshalIndent(obj, "", "  ")
}
