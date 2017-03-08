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
//will return a state.
type State struct {
	Game      GameState
	Players   []PlayerState
	computed  *computedPropertiesImpl
	sanitized bool
	delegate  GameDelegate
}

//Copy returns a deep copy of the State, including copied version of the Game
//and Player States.
func (s *State) copy(sanitized bool) *State {

	players := make([]PlayerState, len(s.Players))

	for i, player := range s.Players {
		players[i] = player.Copy()
	}

	result := &State{
		Game:      s.Game.Copy(),
		Players:   players,
		sanitized: sanitized,
		delegate:  s.delegate,
	}

	//FixUp stacks to make sure they point to this new state.
	if err := verifyReaderStacks(result.Game.Reader(), result); err != nil {
		return nil
	}
	for _, player := range result.Players {
		if err := verifyReaderStacks(player.Reader(), result); err != nil {
			return nil
		}
	}

	return result

}

func (s *State) MarshalJSON() ([]byte, error) {
	obj := map[string]interface{}{
		"Game":     s.Game,
		"Players":  s.Players,
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
		config := s.delegate.ComputedPropertiesConfig()
		var playerConfig map[string]ComputedPlayerPropertyDefinition
		if config != nil {
			playerConfig = config.PlayerProperties
		}

		playerBags := make([]*computedPlayerPropertiesImpl, len(s.Players))

		for i, player := range s.Players {
			playerBags[i] = &computedPlayerPropertiesImpl{
				newComputedPropertiesBag(),
				playerConfig,
				player,
			}
		}

		s.computed = &computedPropertiesImpl{
			newComputedPropertiesBag(),
			playerBags,
			s,
			config,
		}
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
	if playerIndex < 0 || playerIndex >= len(s.Players) {
		return s
	}

	policy := s.delegate.StateSanitizationPolicy()

	if policy == nil {
		policy = &StatePolicy{}
	}

	sanitized := s.copy(true)

	sanitizeStateObj(sanitized.Game.Reader(), policy.Game, -1, playerIndex)

	playerStates := sanitized.Players

	for i := 0; i < len(playerStates); i++ {
		sanitizeStateObj(playerStates[i].Reader(), policy.Player, i, playerIndex)
	}

	return sanitized

}

//BaseState is the interface that all state objects--UserStates and GameStates
//--implement.
type BaseState interface {
	Reader() PropertyReadSetter
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

//GameState represents the state of a game that is not associated with a
//particular user. For example, the draw stack of cards, who the current
//player is, and other properites.
type GameState interface {
	//Copy returns a copy of our current state. Be sure it's a deep copy that
	//makes a copy of any pointer arguments.
	Copy() GameState
	BaseState
}

//DefaultMarshalJSON is a simple wrapper around json.MarshalIndent, with the
//right defaults set. If your structs need to implement MarshaLJSON to output
//JSON, use this to encode it.
func DefaultMarshalJSON(obj interface{}) ([]byte, error) {
	return json.MarshalIndent(obj, "", "  ")
}
