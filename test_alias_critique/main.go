package main

// This file tests the type alias approach for the boardgame framework generics migration

// Simulate the boardgame package types
type SubState interface{}
type PlayerIndex int

// Simulated TypedImmutableState and TypedState from the migration plan
type TypedImmutableState[G, P any] interface {
	GameState() *G
	PlayerStates() []*P
	CurrentPlayer() *P
}

type TypedState[G, P any] interface {
	TypedImmutableState[G, P]
	MutableGameState() *G
	MutablePlayerStates() []*P
	MutableCurrentPlayer() *P
}

// Simulated moves package Default type
type MovesDefault[G, P any] struct {
	// Base move logic
}

func (m *MovesDefault[G, P]) Legal(state TypedImmutableState[G, P], proposer PlayerIndex) error {
	return nil
}

// Example game-specific state types (like examples/pig)
type gameState struct {
	CurrentPlayerValue int
	TargetScore        int
}

type playerState struct {
	Score  int
	Busted bool
}

// TEST 1: Type aliases in game package
type State = TypedImmutableState[gameState, playerState]
type MutableState = TypedState[gameState, playerState]

// TEST 2: Can we use aliases in move definitions?
type moveRollDice struct {
	MovesDefault[gameState, playerState]
}

func (m *moveRollDice) Legal(state State, proposer PlayerIndex) error {
	// Can we call embedded methods with the alias?
	if err := m.MovesDefault.Legal(state, proposer); err != nil {
		return err
	}

	game := state.GameState()
	_ = game.TargetScore
	return nil
}

func (m *moveRollDice) Apply(state MutableState) error {
	game := state.MutableGameState()
	game.CurrentPlayerValue++
	return nil
}

// TEST 3: What if two games are in the same binary?
// This is the CRITICAL issue - name conflicts!

type game2GameState struct {
	Value string
}

type game2PlayerState struct {
	Name string
}

// This WILL NOT WORK - name conflict with the aliases above
// type State = TypedImmutableState[game2GameState, game2PlayerState]
// type MutableState = TypedState[game2GameState, game2PlayerState]

// Instead, each game needs unique names:
type Game2State = TypedImmutableState[game2GameState, game2PlayerState]
type Game2MutableState = TypedState[game2GameState, game2PlayerState]

// TEST 4: Can IDE autocomplete work with aliases?
func exampleUsage() {
	var state State
	// When you type "state.", does the IDE show GameState(), PlayerStates(), etc.?
	// This depends on IDE implementation and may not work well.
	_ = state.GameState()
}

// TEST 5: Error messages - what do they look like?
func causeError() TypedImmutableState[gameState, playerState] {
	// If there's a type error, will it show:
	// - "expected TypedImmutableState[gameState, playerState]"
	// - "expected State"
	// The former is more helpful for debugging
	return nil
}

func main() {
	println("Type alias approach compiles successfully")
	println("But has significant issues - see analysis below")
}
