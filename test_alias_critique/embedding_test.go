package main

import "fmt"

// TEST: Can we embed moves.Default with type parameters and use aliases?

// Simulating what the moves package would look like
type Default[G, P any] struct {
	name string
}

func (d *Default[G, P]) Legal(state TypedImmutableState[G, P], proposer PlayerIndex) error {
	fmt.Println("Default.Legal called")
	return nil
}

func (d *Default[G, P]) GetName() string {
	return d.name
}

// Game-specific types
type pigGameState struct {
	Die int
}

type pigPlayerState struct {
	Score int
}

// Type aliases in the pig package
type PigState = TypedImmutableState[pigGameState, pigPlayerState]
type PigMutableState = TypedState[pigGameState, pigPlayerState]

// Move that embeds Default with explicit types
type moveRollDiceExplicit struct {
	Default[pigGameState, pigPlayerState]
}

func (m *moveRollDiceExplicit) Legal(state PigState, proposer PlayerIndex) error {
	// This SHOULD work - calling embedded method with alias type
	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}

	game := state.GameState()
	fmt.Println("Die value:", game.Die)
	return nil
}

// CRITICAL TEST: Can we define a move using the alias for the embedded type?
// This would be the ideal syntax for game developers

// This WILL NOT WORK because type aliases can't be used in struct field types with generics
// type moveRollDiceWithAlias struct {
//     Default[PigState, PigMutableState]  // ERROR: PigState is not a type, it's a type alias
// }

// The embedded struct MUST use the concrete types, not aliases
// This means game developers STILL need to write the full type parameters

func testEmbedding() {
	move := &moveRollDiceExplicit{}
	move.name = "Roll Dice"

	fmt.Println("Move name:", move.GetName())

	// This works - the alias can be used in method parameters
	var state PigState
	move.Legal(state, 0)
}
