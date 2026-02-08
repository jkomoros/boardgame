package main

import "fmt"

// Realistic test of how the type alias pattern would work in practice
// Simulating the pig game from examples/pig

// ============================================================================
// Simulated boardgame package
// ============================================================================

type ImmutableSubState interface {
	State() TypedImmutableState[any, any] // problematic - can't properly type this
}

type MutableSubState interface {
	ImmutableSubState
}

// ============================================================================
// Simulated moves package with generics
// ============================================================================

type CurrentPlayer[G, P any] struct {
	Default[G, P]
}

func (c *CurrentPlayer[G, P]) Legal(state TypedImmutableState[G, P], proposer PlayerIndex) error {
	fmt.Println("CurrentPlayer.Legal called")
	return c.Default.Legal(state, proposer)
}

// ============================================================================
// pig package - examples/pig
// ============================================================================

// Game state types
type pigGameStateStruct struct {
	CurrentPlayerValue int
	TargetScore        int
}

type pigPlayerStateStruct struct {
	Score      int
	Busted     bool
	DieCounted bool
	RoundScore int
	TotalScore int
}

// TYPE ALIASES - the pattern being evaluated
type PigImmutableState = TypedImmutableState[pigGameStateStruct, pigPlayerStateStruct]
type PigState = TypedState[pigGameStateStruct, pigPlayerStateStruct]

// Move definitions
type moveRollDice struct {
	CurrentPlayer[pigGameStateStruct, pigPlayerStateStruct] // MUST use full types, not aliases
}

func (m *moveRollDice) Legal(state PigImmutableState, proposer PlayerIndex) error {
	// Call embedded Legal - does this work with the alias?
	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game := state.GameState()
	players := state.PlayerStates()

	p := players[game.CurrentPlayerValue]

	if !p.DieCounted {
		return fmt.Errorf("most recent roll not counted")
	}

	return nil
}

func (m *moveRollDice) Apply(state PigState) error {
	game := state.MutableGameState()
	players := state.MutablePlayerStates()

	p := players[game.CurrentPlayerValue]
	p.DieCounted = false

	return nil
}

// ============================================================================
// memory package - examples/memory
// ============================================================================

type memoryGameStateStruct struct {
	NumCards int
}

type memoryPlayerStateStruct struct {
	CardsLeftToReveal int
}

// Each game needs its own uniquely-named aliases
type MemoryImmutableState = TypedImmutableState[memoryGameStateStruct, memoryPlayerStateStruct]
type MemoryState = TypedState[memoryGameStateStruct, memoryPlayerStateStruct]

type moveRevealCard struct {
	CurrentPlayer[memoryGameStateStruct, memoryPlayerStateStruct]
	CardIndex int
}

func (m *moveRevealCard) Legal(state MemoryImmutableState, proposer PlayerIndex) error {
	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game := state.GameState()
	players := state.PlayerStates()

	p := players[game.NumCards] // accessing typed state

	if p.CardsLeftToReveal < 1 {
		return fmt.Errorf("no cards left to reveal")
	}

	return nil
}

// ============================================================================
// Real-world usage scenario: Multiple games in one binary
// ============================================================================

func testMultipleGames() {
	// Can we use both games in the same function?
	var pigState PigImmutableState
	var memoryState MemoryImmutableState

	// These are completely different types - good!
	_ = pigState
	_ = memoryState

	// But we can't write a generic function that works with both:
	// processSomeState(pigState)    // What type would processSomeState accept?
	// processSomeState(memoryState)

	// We'd need something like:
	// func processSomeState[G, P any](state TypedImmutableState[G, P]) { }
	// But then we're back to writing the type parameters everywhere
}

// ============================================================================
// CRITICAL ISSUE: moves.Default embedding
// ============================================================================

// In the real boardgame framework, moves embed moves.Default or moves.CurrentPlayer
// Those would need to be generic: moves.Default[G, P]
// But game developers can't use their aliases when embedding:

// This does NOT work:
// type moveExample struct {
//     moves.Default[PigImmutableState, PigState]  // ERROR: these are not types, they're aliases to interfaces
// }

// Game developers must write:
type moveExample struct {
	Default[pigGameStateStruct, pigPlayerStateStruct] // Full type parameters required
}

// This defeats the purpose of aliases - you still need to write the full
// type parameters in every move struct definition!

// ============================================================================
// ANALYSIS SUMMARY
// ============================================================================

func analysisSummary() {
	fmt.Println(`
PROS:
1. ✓ Type aliases DO work syntactically in Go
2. ✓ Can reduce typing in METHOD SIGNATURES
3. ✓ Type-safe at compile time
4. ✓ Each game can have its own aliases (PigState, MemoryState)

CONS:
1. ✗ CANNOT use aliases when embedding generic structs
   - moves.Default[gameState, playerState] - must use full types
   - This is where most of the boilerplate lives!
2. ✗ Error messages show full expanded types, not aliases
   - Less readable error messages
3. ✗ IDE autocomplete may not work as well
   - Depends on IDE implementation
4. ✗ Reflection shows underlying type, not alias name
   - Debugging/logging shows long type names
5. ✗ Still need to write full type parameters in struct embedding
   - The MAIN source of boilerplate is not eliminated
6. ✗ Multiple games in same binary need different alias names
   - Can't all use "State" - need PigState, MemoryState, etc.

VERDICT: Type aliases provide MINIMAL benefit because you can't use them
where you need them most - in struct embedding. The boilerplate reduction
is only in method signatures, not in move struct definitions.
`)
}
