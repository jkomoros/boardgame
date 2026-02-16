package main

// TEST: What do error messages look like with type aliases?

func testErrorMessages() {
	// Test 1: Type mismatch error
	var pigState PigState
	var game2State Game2State

	// This should error - what does the error message say?
	// Does it say "cannot use game2State (type Game2State = TypedImmutableState[game2GameState, game2PlayerState])"
	// or does it expand the alias?
	_ = pigState
	_ = game2State

	// Uncommenting this should produce an error:
	// pigState = game2State  // Type mismatch

	// Test 2: Method call error
	// What happens if we pass the wrong alias to a method?
	move := &moveRollDiceExplicit{}
	// move.Legal(game2State, 0)  // Should error - what message?

	// Test 3: Return type error
	// testReturnError()
}

func testReturnError() TypedImmutableState[pigGameState, pigPlayerState] {
	// If we try to return the wrong type, what's the error message?
	// var state Game2State
	// return state  // Should error
	return nil
}

// TEST: Type identity - are aliases truly transparent?
func testTypeIdentity() {
	// In Go, type aliases should be completely transparent at compile time
	// This means PigState and TypedImmutableState[pigGameState, pigPlayerState] are the SAME type

	var state1 PigState
	var state2 TypedImmutableState[pigGameState, pigPlayerState]

	// These should be interchangeable
	state1 = state2
	state2 = state1

	// Both should work with the same function
	processState(state1)
	processState(state2)
}

func processState(state PigState) {
	// works with both aliased and non-aliased types
}

// TEST: Reflection and type names
func testReflection() {
	// When using reflection, what name appears?
	// reflect.TypeOf(state).String() will show the UNDERLYING type, not the alias
	// This could be confusing in error messages or debugging
}

// TEST: Import conflicts
// If multiple games each define "State" as a type alias,
// and you try to import both packages, you get conflicts:

// import (
//     "github.com/jkomoros/boardgame/examples/pig"
//     "github.com/jkomoros/boardgame/examples/memory"
// )
//
// pig.State vs memory.State - works fine as they're in different packages
// But if you want to use both in the same function, you need fully qualified names
