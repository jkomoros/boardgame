package boardgame

import (
	"encoding/json"
)

//State represents the entire semantic state of a game at a given version. For
//your specific game, Game and Players will actually be concrete structs to
//your particular game. Games often define a top-level concreteStates()
//*myGameState, []*myPlayerState so at the top of methods that accept a *State
//they can quickly get concrete, type-checked types with only a single
//conversion leap of faith at the top. States are generally read-only; the
//exception is in Move.Apply() and Delegate.BeginSetup() and FinishSetup(),
//when you may modify the provided state. The MarshalJSON output of a State is
//appropriate for sending to a client or serializing a state to be put in
//storage. Given a blob serialized in that fashion, GameManager.StateFromBlob
//will return a state. States are intended to be read-only: that is, you
//should not modify the underlying properties of the game or player states.
//That is why the states only return non-mutable states (PropertyReaders, not
//PropertyReadSetters, although realistically it is possible to cast them and
//modify directly.
type State struct {
	game      MutableGameState
	players   []MutablePlayerState
	computed  *computedPropertiesImpl
	sanitized bool
	delegate  GameDelegate
}

//A MutableState is a state that is designed to be modified in place. These
//are passed to methods (instead of normal States) as a signal that
//modifications are intended to be done on the state.
type MutableState struct {
	//MutableState wraps a Read-Only state object.
	*State
}

//MutableGame is a reference to the MutableGameState for this MutableState.
func (m *MutableState) MutableGame() MutableGameState {
	return m.game
}

//MutablePlayers returns a slice of MutablePlayerStates for this MutableState.
func (m *MutableState) MutablePlayers() []MutablePlayerState {
	return m.players
}

//Game returns the GameState for this State
func (s *State) Game() GameState {
	return s.game
}

//Players returns a slice of all PlayerStates for this State
func (s *State) Players() []PlayerState {
	result := make([]PlayerState, len(s.players))
	for i := 0; i < len(s.players); i++ {
		result[i] = s.players[i]
	}
	return result
}

//Copy returns a deep copy of the State, including copied version of the Game
//and Player States.
func (s *State) Copy(sanitized bool) *State {

	players := make([]MutablePlayerState, len(s.players))

	for i, player := range s.players {
		players[i] = player.MutableCopy()
	}

	result := &State{
		game:      s.game.MutableCopy(),
		players:   players,
		sanitized: sanitized,
		delegate:  s.delegate,
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

func (s *State) MarshalJSON() ([]byte, error) {
	obj := map[string]interface{}{
		"Game":     s.game,
		"Players":  s.players,
		"Computed": s.Computed(),
	}
	return json.Marshal(obj)
}

//Diagram returns a basic, ascii rendering of the state for debug rendering.
//It thunks out to Delegate.Diagram.
func (s *State) Diagram() string {
	return s.delegate.Diagram(s)
}

//Santizied will return false if this is a full-fidelity State object, or
//true if it has been sanitized, which means that some properties might be
//hidden or otherwise altered. This should return true if the object was
//created with Copy(true)
func (s *State) Sanitized() bool {
	return s.sanitized
}

//Computed returns the computed properties for this state.
func (s *State) Computed() ComputedProperties {
	if s.computed == nil {
		s.computed = newComputedPropertiesImpl(s.delegate.ComputedPropertiesConfig(), s)
	}
	return s.computed
}

//SanitizedForPlayer produces a copy state object that has been sanitized for
//the player at the given index. The state object returned will have
//Sanitized() return true. Will call GameDelegate.StateSanitizationPolicy to
//retrieve the policy in place. See the package level comment for an overview
//of how state sanitization works.
func (s *State) SanitizedForPlayer(playerIndex int) *State {

	//If the playerIndex isn't an actuall player's index, just return self.
	if playerIndex < 0 || playerIndex >= len(s.players) {
		return s
	}

	policy := s.delegate.StateSanitizationPolicy()

	if policy == nil {
		policy = &StatePolicy{}
	}

	sanitized := s.Copy(true)

	sanitizeStateObj(sanitized.game.ReadSetter(), policy.Game, -1, playerIndex, PolicyVisible)

	playerStates := sanitized.players

	for i := 0; i < len(playerStates); i++ {
		sanitizeStateObj(playerStates[i].ReadSetter(), policy.Player, i, playerIndex, PolicyVisible)
	}

	return sanitized

}

//sanitizedWithExceptions will return a Sanitized() State where properties
//that are not in the passed policy are treated as PolicyRandom. Useful in
//computing properties.
func (s *State) sanitizedWithExceptions(policy *StatePolicy) *State {

	sanitized := s.Copy(true)

	sanitizeStateObj(sanitized.game.ReadSetter(), policy.Game, -1, -1, PolicyRandom)

	playerStates := sanitized.players

	for i := 0; i < len(playerStates); i++ {
		sanitizeStateObj(playerStates[i].ReadSetter(), policy.Player, -1, -1, PolicyRandom)
	}

	return sanitized

}

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
