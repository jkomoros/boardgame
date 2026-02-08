package main

import "fmt"

// CRITICAL LIMITATION TEST
// This demonstrates WHY type aliases don't actually help much

// ============================================================================
// What game developers WANT to write (with aliases):
// ============================================================================

// In pig/state.go:
// type State = TypedImmutableState[gameState, playerState]
// type MutableState = TypedState[gameState, playerState]

// In pig/moves.go:
// type moveRollDice struct {
//     moves.CurrentPlayer[State, MutableState]  // ERROR: State is an interface type alias, not a concrete type
// }

// ============================================================================
// What they ACTUALLY have to write:
// ============================================================================

// Setup
type DemoGameState struct{ Value int }
type DemoPlayerState struct{ Score int }

type DemoState = TypedImmutableState[DemoGameState, DemoPlayerState]
type DemoMutableState = TypedState[DemoGameState, DemoPlayerState]

// The ACTUAL move definition - can't use the aliases for embedding!
// type moveDemoActual struct {
// 	   You MUST use the concrete types here, NOT the aliases
// 	   moves.Default[DemoGameState, DemoPlayerState]  // Must use concrete types!
// }

// Method signatures CAN use the aliases
// func (m *moveDemoActual) Legal(state DemoState, proposer PlayerIndex) error {
// 	   This works - aliases work in method signatures
// 	   return m.Default.Legal(state, proposer)
// }

// ============================================================================
// BOILERPLATE ANALYSIS
// ============================================================================

func boilerplateAnalysis() {
	fmt.Println(`
WITHOUT ALIASES (current approach):
-----------------------------------
type moveRollDice struct {
    moves.CurrentPlayer
}

func (m *moveRollDice) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    game, players := concreteStates(state)  // 1 line of boilerplate
    // ... use game and players
}

WITH ALIASES (proposed approach):
-----------------------------------
type moveRollDice struct {
    moves.CurrentPlayer[gameState, playerState]  // STILL need to write type params!
}

func (m *moveRollDice) Legal(state State, proposer boardgame.PlayerIndex) error {
    game := state.GameState()     // No concreteStates() needed
    players := state.PlayerStates()
    // ... use game and players
}

SAVINGS:
- Struct definition: NO SAVINGS (still need type parameters)
- Method body: Save ~1 line per method (no concreteStates() call)
- Method signature: Save ~20 characters (State vs boardgame.ImmutableState)

TRADE-OFFS:
- Error messages become longer and less clear
- IDE autocomplete may not work as well
- Each game needs unique alias names (PigState, MemoryState, etc.)
- Still need to write [gameState, playerState] in every move struct

VERDICT: Aliases provide MINIMAL benefit (~1 line per method) but you
         still write [gameState, playerState] in every move definition.
         The main boilerplate is NOT eliminated.
`)
}

// ============================================================================
// ALTERNATIVE: What if we don't use type aliases at all?
// ============================================================================

// Option 1: Just use full types everywhere
// type moveNoAlias1 struct {
// 	   moves.Default[DemoGameState, DemoPlayerState]
// }
//
// func (m *moveNoAlias1) Legal(state TypedImmutableState[DemoGameState, DemoPlayerState], proposer PlayerIndex) error {
// 	   game := state.GameState()
// 	   return nil
// }

// Option 2: Use a helper at package level to create moves
// This doesn't work in Go - you can't parameterize at package level

// Option 3: Code generation
// Generate move types with the correct type parameters
// This is probably the best approach!

func alternativeApproaches() {
	fmt.Println(`
BETTER ALTERNATIVES TO TYPE ALIASES:
====================================

1. CODE GENERATION (Best option)
   - Update boardgame-util codegen to generate typed move structs
   - Generate with correct type parameters automatically
   - No manual type parameter writing needed
   - Clean error messages

2. FULLY QUALIFIED TYPES (Simple, clear)
   - Just write TypedImmutableState[gameState, playerState] everywhere
   - Long but explicit and clear
   - No aliases to confuse things
   - Better error messages

3. BASE TYPES WITHOUT GENERICS (Current approach, keep it)
   - Keep current interface{} approach
   - Add Optional generic wrappers for those who want them
   - Don't force generics on everyone
   - Most games don't have type safety issues

RECOMMENDATION: Code generation is the way forward, not type aliases.
`)
}
