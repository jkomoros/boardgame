package boardgame

import (
	"encoding/json"
	"testing"

	"github.com/jkomoros/boardgame/errors"
	"github.com/workfit/tester/assert"
)

func testGameWithMutableConstraints(t *testing.T) *Game {
	t.Helper()
	game := testDefaultGame(t, false)
	game.Manager().Internals().AllowMutableConstraints(game)
	return game
}

func mustAddStackConstraint(t testing.TB, stack Stack, constraint StackConstraint) {
	t.Helper()
	if err := stack.AddConstraint(constraint); err != nil {
		t.Fatal("add stack constraint:", err)
	}
}

func TestConstraintBlocksMoveTo(t *testing.T) {
	game := testGameWithMutableConstraints(t)

	gameState, playerStates := concreteStates(game.CurrentState())

	drawStack := gameState.DrawDeck
	hand := playerStates[0].Hand

	// Move a component into the hand to start.
	err := drawStack.First().MoveTo(hand, hand.SizedStack().FirstSlot())
	assert.For(t).ThatActual(err).IsNil()

	initialHandCount := hand.NumComponents()
	initialDrawCount := drawStack.NumComponents()

	// Add a constraint that always rejects.
	hand.AddConstraint(func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
		return errors.New("always rejected")
	})

	// Try to move another component in - should fail.
	err = drawStack.First().MoveTo(hand, hand.SizedStack().LastSlot())
	assert.For(t).ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(hand.NumComponents()).Equals(initialHandCount)
	assert.For(t).ThatActual(drawStack.NumComponents()).Equals(initialDrawCount)
}

func TestConstraintAllowsMove(t *testing.T) {
	game := testGameWithMutableConstraints(t)

	gameState, playerStates := concreteStates(game.CurrentState())

	drawStack := gameState.DrawDeck
	hand := playerStates[0].Hand

	// Add a constraint that always passes.
	hand.AddConstraint(func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
		return nil
	})

	err := drawStack.First().MoveTo(hand, hand.SizedStack().FirstSlot())
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(hand.NumComponents()).Equals(1)
}

func TestConstraintRollback(t *testing.T) {
	game := testGameWithMutableConstraints(t)

	gameState, playerStates := concreteStates(game.CurrentState())

	drawStack := gameState.DrawDeck
	hand := playerStates[0].Hand

	initialDrawCount := drawStack.NumComponents()
	initialHandCount := hand.NumComponents()

	// Get the first component's deck index before the move.
	firstComponent := drawStack.ImmutableFirst()
	assert.For(t).ThatActual(firstComponent).IsNotNil()

	// Add a constraint that rejects.
	hand.AddConstraint(func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
		return errors.New("rejected")
	})

	err := drawStack.First().MoveTo(hand, hand.SizedStack().FirstSlot())
	assert.For(t).ThatActual(err).IsNotNil()

	// Verify source and destination are unchanged.
	assert.For(t, "draw count after rollback").ThatActual(drawStack.NumComponents()).Equals(initialDrawCount)
	assert.For(t, "hand count after rollback").ThatActual(hand.NumComponents()).Equals(initialHandCount)

	// The first component in draw should still be the same one.
	assert.For(t, "first component preserved").ThatActual(drawStack.ImmutableFirst()).Equals(firstComponent)
}

func TestMoveAllToRespectsConstraints(t *testing.T) {
	game := testGameWithMutableConstraints(t)

	gameState, playerStates := concreteStates(game.CurrentState())

	drawStack := gameState.DrawDeck
	hand := playerStates[0].Hand

	// Move one component to hand first (to have the first slot filled).
	err := drawStack.First().MoveTo(hand, hand.SizedStack().FirstSlot())
	assert.For(t).ThatActual(err).IsNil()

	afterFirstMoveDrawCount := drawStack.NumComponents()
	afterFirstMoveHandCount := hand.NumComponents()

	// Now add a constraint that rejects (simulating "hand is full").
	hand.AddConstraint(func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
		return errors.New("hand full")
	})

	// MoveAllTo should fail because constraint is violated on the next move.
	err = drawStack.MoveAllTo(hand)
	assert.For(t).ThatActual(err).IsNotNil()

	// Draw count should still be the same as before MoveAllTo was attempted.
	// Note: MoveAllTo moves components one at a time, so the first attempted
	// move in the MoveAllTo loop will fail, leaving counts unchanged.
	assert.For(t, "draw count unchanged").ThatActual(drawStack.NumComponents()).Equals(afterFirstMoveDrawCount)
	assert.For(t, "hand count unchanged").ThatActual(hand.NumComponents()).Equals(afterFirstMoveHandCount)
}

func TestMoveAllToLateConstraintRejectionIsAtomic(t *testing.T) {
	game := testGameWithMutableConstraints(t)
	gameState, playerStates := concreteStates(game.CurrentState())

	source := gameState.OtherStack
	destination := playerStates[0].Hand
	for slot := 0; slot < 2; slot++ {
		if err := gameState.DrawDeck.First().MoveTo(source, slot); err != nil {
			t.Fatalf("populate source slot %d: %v", slot, err)
		}
	}

	moved := source.ImmutableComponents()
	for slot, component := range moved {
		stack, gotSlot, err := component.ContainingImmutableStack()
		if err != nil {
			t.Fatalf("build component index for source slot %d: %v", slot, err)
		}
		if stack != source || gotSlot != slot {
			t.Fatalf("component in source slot %d reported at %v slot %d", slot, stack, gotSlot)
		}
	}

	mustAddStackConstraint(t, destination, func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
		if dest.NumComponents()+len(proposed) > 1 {
			return errors.New("only one component accepted")
		}
		return nil
	})

	before, err := json.Marshal(game.CurrentState())
	if err != nil {
		t.Fatalf("marshal state before MoveAllTo: %v", err)
	}

	err = source.MoveAllTo(destination)
	if err == nil || err.Error() != "only one component accepted" {
		t.Fatalf("MoveAllTo error = %v, want late constraint rejection", err)
	}

	after, err := json.Marshal(game.CurrentState())
	if err != nil {
		t.Fatalf("marshal state after MoveAllTo: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("state changed after rejected MoveAllTo\nbefore: %s\n after: %s", before, after)
	}

	for slot, component := range moved {
		stack, gotSlot, err := component.ContainingImmutableStack()
		if err != nil {
			t.Fatalf("locate component from source slot %d after rejection: %v", slot, err)
		}
		if stack != source || gotSlot != slot {
			t.Fatalf("component from source slot %d moved after rejection: stack %v slot %d", slot, stack, gotSlot)
		}
	}
}

func TestMoveAllToValidatesOnceOnCopiedState(t *testing.T) {
	game := testGameWithMutableConstraints(t)
	gameState, playerStates := concreteStates(game.CurrentState())

	source := gameState.OtherStack
	destination := playerStates[0].Hand
	for slot := 0; slot < 2; slot++ {
		if err := gameState.DrawDeck.First().MoveTo(source, slot); err != nil {
			t.Fatalf("populate source slot %d: %v", slot, err)
		}
	}

	liveState := game.CurrentState()
	var destinationCounts []int
	mustAddStackConstraint(t, destination, func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
		if st == liveState || dest.state() == liveState {
			return errors.New("constraint received live state")
		}
		if dest.state() != st {
			return errors.New("constraint destination and state do not match")
		}
		for _, component := range proposed {
			if component.ImmutableState() != st {
				return errors.New("proposed component and state do not match")
			}
		}
		destinationCounts = append(destinationCounts, dest.NumComponents())
		return nil
	})

	if err := source.MoveAllTo(destination); err != nil {
		t.Fatalf("MoveAllTo: %v", err)
	}
	if len(destinationCounts) != 2 || destinationCounts[0] != 0 || destinationCounts[1] != 1 {
		t.Fatalf("constraint destination counts = %v, want [0 1]", destinationCounts)
	}
	if source.NumComponents() != 0 || destination.NumComponents() != 2 {
		t.Fatalf("counts after MoveAllTo = source %d, destination %d; want 0, 2", source.NumComponents(), destination.NumComponents())
	}
}

func TestMoveAllToLateRejectionAtomicAcrossStackKinds(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, gameState *testGameState, playerStates []*testPlayerState) (Stack, Stack)
	}{
		{
			name: "growable to growable",
			setup: func(t *testing.T, gameState *testGameState, _ []*testPlayerState) (Stack, Stack) {
				source := gameState.MyBoard.SpaceAt(0)
				for i := 0; i < 2; i++ {
					if err := gameState.DrawDeck.First().MoveToNextSlot(source); err != nil {
						t.Fatalf("populate growable source: %v", err)
					}
				}
				return source, gameState.MyBoard.SpaceAt(2)
			},
		},
		{
			name: "growable to sized",
			setup: func(t *testing.T, gameState *testGameState, playerStates []*testPlayerState) (Stack, Stack) {
				source := gameState.MyBoard.SpaceAt(0)
				for i := 0; i < 2; i++ {
					if err := gameState.DrawDeck.First().MoveToNextSlot(source); err != nil {
						t.Fatalf("populate growable source: %v", err)
					}
				}
				return source, playerStates[0].Hand
			},
		},
		{
			name: "sparse sized to growable",
			setup: func(t *testing.T, gameState *testGameState, _ []*testPlayerState) (Stack, Stack) {
				for _, slot := range []int{0, 2} {
					if err := gameState.DrawDeck.First().MoveTo(gameState.DownSizeStack, slot); err != nil {
						t.Fatalf("populate sized source slot %d: %v", slot, err)
					}
				}
				return gameState.DownSizeStack, gameState.MyBoard.SpaceAt(0)
			},
		},
		{
			name: "sparse sized to sized",
			setup: func(t *testing.T, gameState *testGameState, playerStates []*testPlayerState) (Stack, Stack) {
				for _, slot := range []int{0, 2} {
					if err := gameState.DrawDeck.First().MoveTo(gameState.DownSizeStack, slot); err != nil {
						t.Fatalf("populate sized source slot %d: %v", slot, err)
					}
				}
				return gameState.DownSizeStack, playerStates[0].Hand
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			game := testGameWithMutableConstraints(t)
			gameState, playerStates := concreteStates(game.CurrentState())
			source, destination := test.setup(t, gameState, playerStates)

			mustAddStackConstraint(t, destination, func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
				if dest.NumComponents()+len(proposed) > 1 {
					return errors.New("only one component accepted")
				}
				return nil
			})

			before, err := json.Marshal(game.CurrentState())
			if err != nil {
				t.Fatalf("marshal state before MoveAllTo: %v", err)
			}
			if err := source.MoveAllTo(destination); err == nil || err.Error() != "only one component accepted" {
				t.Fatalf("MoveAllTo error = %v, want late constraint rejection", err)
			}
			after, err := json.Marshal(game.CurrentState())
			if err != nil {
				t.Fatalf("marshal state after MoveAllTo: %v", err)
			}
			if string(after) != string(before) {
				t.Fatal("state changed after rejected MoveAllTo")
			}
			if err := destination.ClearConstraints(); err != nil {
				t.Fatal("clear rejecting constraint:", err)
			}
			component := source.First()
			if err := component.MoveToNextSlot(destination); err != nil {
				t.Fatal("valid move after rejected transaction:", err)
			}
			stack, _, err := component.ContainingImmutableStack()
			if err != nil {
				t.Fatal("locate component after valid follow-up move:", err)
			}
			if stack != destination {
				t.Fatal("component index did not point at destination after valid follow-up move")
			}
		})
	}
}

func TestMayMoveAllToAndMoveAllToReturnSameConstraintError(t *testing.T) {
	game := testGameWithMutableConstraints(t)
	gameState, playerStates := concreteStates(game.CurrentState())
	source := gameState.OtherStack
	destination := playerStates[0].Hand
	for slot := 0; slot < 2; slot++ {
		if err := gameState.DrawDeck.First().MoveTo(source, slot); err != nil {
			t.Fatal(err)
		}
	}
	mustAddStackConstraint(t, destination, func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
		if dest.NumComponents()+len(proposed) > 1 {
			return errors.New("bulk limit reached")
		}
		return nil
	})

	mayErr := source.MayMoveAllTo(destination)
	moveErr := source.MoveAllTo(destination)
	if mayErr == nil || moveErr == nil {
		t.Fatalf("errors = MayMoveAllTo %v, MoveAllTo %v; want both non-nil", mayErr, moveErr)
	}
	if mayErr.Error() != moveErr.Error() {
		t.Fatalf("errors differ: MayMoveAllTo %q, MoveAllTo %q", mayErr, moveErr)
	}
	if source.NumComponents() != 2 || destination.NumComponents() != 0 {
		t.Fatalf("failed checks changed stacks: source %d, destination %d", source.NumComponents(), destination.NumComponents())
	}
}

func TestMoveAllToSuccessMatchesCheckedSequenceAcrossStackKinds(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, gameState *testGameState, playerStates []*testPlayerState) (Stack, Stack)
	}{
		{
			name: "growable to growable with existing destination",
			setup: func(t *testing.T, gameState *testGameState, _ []*testPlayerState) (Stack, Stack) {
				source := gameState.MyBoard.SpaceAt(0)
				destination := gameState.MyBoard.SpaceAt(2)
				if err := gameState.DrawDeck.First().MoveToNextSlot(destination); err != nil {
					t.Fatal(err)
				}
				for i := 0; i < 2; i++ {
					if err := gameState.DrawDeck.First().MoveToNextSlot(source); err != nil {
						t.Fatal(err)
					}
				}
				return source, destination
			},
		},
		{
			name: "growable to sparse sized",
			setup: func(t *testing.T, gameState *testGameState, _ []*testPlayerState) (Stack, Stack) {
				source := gameState.MyBoard.SpaceAt(0)
				destination := gameState.DownSizeStack
				if err := gameState.DrawDeck.First().MoveTo(destination, 1); err != nil {
					t.Fatal(err)
				}
				for i := 0; i < 2; i++ {
					if err := gameState.DrawDeck.First().MoveToNextSlot(source); err != nil {
						t.Fatal(err)
					}
				}
				return source, destination
			},
		},
		{
			name: "sparse sized to growable with existing destination",
			setup: func(t *testing.T, gameState *testGameState, _ []*testPlayerState) (Stack, Stack) {
				source := gameState.DownSizeStack
				destination := gameState.MyBoard.SpaceAt(0)
				if err := gameState.DrawDeck.First().MoveToNextSlot(destination); err != nil {
					t.Fatal(err)
				}
				for _, slot := range []int{0, 2} {
					if err := gameState.DrawDeck.First().MoveTo(source, slot); err != nil {
						t.Fatal(err)
					}
				}
				return source, destination
			},
		},
		{
			name: "sparse sized to sparse sized",
			setup: func(t *testing.T, gameState *testGameState, _ []*testPlayerState) (Stack, Stack) {
				source := gameState.DownSizeStack
				destination := gameState.OtherStack
				if err := destination.ExpandSize(2); err != nil {
					t.Fatal(err)
				}
				if err := gameState.DrawDeck.First().MoveTo(destination, 1); err != nil {
					t.Fatal(err)
				}
				for _, slot := range []int{0, 2} {
					if err := gameState.DrawDeck.First().MoveTo(source, slot); err != nil {
						t.Fatal(err)
					}
				}
				return source, destination
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			game := testGameWithMutableConstraints(t)
			gameState, playerStates := concreteStates(game.CurrentState())
			source, destination := test.setup(t, gameState, playerStates)
			moved := make([]ComponentInstance, 0, source.NumComponents())
			for _, component := range source.Components() {
				if component != nil {
					moved = append(moved, component)
				}
			}

			mustAddStackConstraint(t, destination, func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
				if dest.state() != st {
					return errors.New("destination is not owned by supplied state")
				}
				for _, component := range proposed {
					if component.ImmutableState() != st {
						return errors.New("proposed component is not owned by supplied state")
					}
				}
				return nil
			})

			originalState := game.CurrentState().(*state)
			expectedState, err := originalState.copy(false)
			if err != nil {
				t.Fatal(err)
			}
			expectedSource, err := findCorrespondingStack(source, originalState, expectedState)
			if err != nil {
				t.Fatal(err)
			}
			expectedDestination, err := findCorrespondingStack(destination, originalState, expectedState)
			if err != nil {
				t.Fatal(err)
			}
			if err := moveComponentsToChecked(expectedSource, expectedDestination, source.NumComponents()); err != nil {
				t.Fatal("checked sequence:", err)
			}

			if err := source.MoveAllTo(destination); err != nil {
				t.Fatal("transactional MoveAllTo:", err)
			}
			gotJSON, err := json.Marshal(originalState)
			if err != nil {
				t.Fatal(err)
			}
			wantJSON, err := json.Marshal(expectedState)
			if err != nil {
				t.Fatal(err)
			}
			if string(gotJSON) != string(wantJSON) {
				t.Fatalf("transactional state differs from checked sequence\n got: %s\nwant: %s", gotJSON, wantJSON)
			}

			expectedSlots := make(map[int]int)
			for slot, component := range expectedDestination.Components() {
				if component != nil {
					expectedSlots[component.DeckIndex()] = slot
				}
			}
			for _, component := range moved {
				stack, slot, err := component.ContainingImmutableStack()
				if err != nil {
					t.Fatal(err)
				}
				if stack != destination || slot != expectedSlots[component.DeckIndex()] {
					t.Fatalf("component %d location = %v[%d], want destination[%d]", component.DeckIndex(), stack, slot, expectedSlots[component.DeckIndex()])
				}
			}
		})
	}
}

func TestMoveCountToSuccessMatchesCheckedSequenceAcrossStackKinds(t *testing.T) {
	const count = 2
	tests := []struct {
		name  string
		setup func(t *testing.T, gameState *testGameState) (Stack, Stack)
	}{
		{
			name: "growable to growable with existing destination",
			setup: func(t *testing.T, gameState *testGameState) (Stack, Stack) {
				source, destination := gameState.MyBoard.SpaceAt(0), gameState.MyBoard.SpaceAt(2)
				if err := gameState.DrawDeck.First().MoveToNextSlot(destination); err != nil {
					t.Fatal(err)
				}
				for i := 0; i < 3; i++ {
					if err := gameState.DrawDeck.First().MoveToNextSlot(source); err != nil {
						t.Fatal(err)
					}
				}
				return source, destination
			},
		},
		{
			name: "growable to sparse sized",
			setup: func(t *testing.T, gameState *testGameState) (Stack, Stack) {
				source, destination := gameState.MyBoard.SpaceAt(0), gameState.DownSizeStack
				if err := gameState.DrawDeck.First().MoveTo(destination, 1); err != nil {
					t.Fatal(err)
				}
				for i := 0; i < 3; i++ {
					if err := gameState.DrawDeck.First().MoveToNextSlot(source); err != nil {
						t.Fatal(err)
					}
				}
				return source, destination
			},
		},
		{
			name: "sparse sized to growable with existing destination",
			setup: func(t *testing.T, gameState *testGameState) (Stack, Stack) {
				source, destination := gameState.DownSizeStack, gameState.MyBoard.SpaceAt(0)
				if err := gameState.DrawDeck.First().MoveToNextSlot(destination); err != nil {
					t.Fatal(err)
				}
				for _, slot := range []int{0, 2, 3} {
					if err := gameState.DrawDeck.First().MoveTo(source, slot); err != nil {
						t.Fatal(err)
					}
				}
				return source, destination
			},
		},
		{
			name: "sparse sized to sparse sized",
			setup: func(t *testing.T, gameState *testGameState) (Stack, Stack) {
				source, destination := gameState.DownSizeStack, gameState.OtherStack
				if err := destination.ExpandSize(1); err != nil {
					t.Fatal(err)
				}
				if err := gameState.DrawDeck.First().MoveTo(destination, 1); err != nil {
					t.Fatal(err)
				}
				for _, slot := range []int{0, 2, 3} {
					if err := gameState.DrawDeck.First().MoveTo(source, slot); err != nil {
						t.Fatal(err)
					}
				}
				return source, destination
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			game := testGameWithMutableConstraints(t)
			gameState, _ := concreteStates(game.CurrentState())
			source, destination := test.setup(t, gameState)
			unmoved := source.Last()

			mustAddStackConstraint(t, destination, func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
				if dest.state() != st {
					return errors.New("destination is not owned by supplied state")
				}
				for _, component := range proposed {
					if component.ImmutableState() != st {
						return errors.New("proposed component is not owned by supplied state")
					}
				}
				return nil
			})

			originalState := game.CurrentState().(*state)
			expectedState, err := originalState.copy(false)
			if err != nil {
				t.Fatal(err)
			}
			expectedSource, err := findCorrespondingStack(source, originalState, expectedState)
			if err != nil {
				t.Fatal(err)
			}
			expectedDestination, err := findCorrespondingStack(destination, originalState, expectedState)
			if err != nil {
				t.Fatal(err)
			}
			if err := moveComponentsToChecked(expectedSource, expectedDestination, count); err != nil {
				t.Fatal("checked sequence:", err)
			}

			if err := source.MoveCountTo(destination, count); err != nil {
				t.Fatal("transactional MoveCountTo:", err)
			}
			gotJSON, err := json.Marshal(originalState)
			if err != nil {
				t.Fatal(err)
			}
			wantJSON, err := json.Marshal(expectedState)
			if err != nil {
				t.Fatal(err)
			}
			if string(gotJSON) != string(wantJSON) {
				t.Fatalf("transactional state differs from checked sequence\n got: %s\nwant: %s", gotJSON, wantJSON)
			}
			if source.NumComponents() != 1 || source.First() != unmoved {
				t.Fatalf("partial transfer did not leave the final source component in place")
			}
		})
	}
}

func TestMoveAllToDiscardsConstraintMutationOfSuppliedState(t *testing.T) {
	game := testGameWithMutableConstraints(t)
	gameState, playerStates := concreteStates(game.CurrentState())
	source := gameState.OtherStack
	destination := playerStates[0].Hand
	for slot := 0; slot < 2; slot++ {
		if err := gameState.DrawDeck.First().MoveTo(source, slot); err != nil {
			t.Fatalf("populate source slot %d: %v", slot, err)
		}
	}

	originalPlayer := gameState.CurrentPlayer
	mustAddStackConstraint(t, destination, func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
		copiedGameState, _ := concreteStates(st)
		copiedGameState.CurrentPlayer = 2 // Deliberate contract violation on the supplied copy.
		if dest.NumComponents()+len(proposed) > 1 {
			return errors.New("late rejection")
		}
		return nil
	})

	if err := source.MoveAllTo(destination); err == nil || err.Error() != "late rejection" {
		t.Fatalf("MoveAllTo error = %v, want late rejection", err)
	}
	if gameState.CurrentPlayer != originalPlayer {
		t.Fatalf("live CurrentPlayer = %d, want unchanged %d", gameState.CurrentPlayer, originalPlayer)
	}
}

func TestMoveAllToLateConstraintPanicLeavesLiveStateUnchanged(t *testing.T) {
	game := testGameWithMutableConstraints(t)
	gameState, playerStates := concreteStates(game.CurrentState())
	source := gameState.OtherStack
	destination := playerStates[0].Hand
	for slot := 0; slot < 2; slot++ {
		if err := gameState.DrawDeck.First().MoveTo(source, slot); err != nil {
			t.Fatalf("populate source slot %d: %v", slot, err)
		}
	}
	mustAddStackConstraint(t, destination, func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
		if dest.NumComponents() > 0 {
			panic("constraint panic")
		}
		return nil
	})

	before, err := json.Marshal(game.CurrentState())
	if err != nil {
		t.Fatal(err)
	}
	func() {
		defer func() {
			if recovered := recover(); recovered != "constraint panic" {
				t.Fatalf("recovered %v, want constraint panic", recovered)
			}
		}()
		_ = source.MoveAllTo(destination)
		t.Fatal("MoveAllTo did not propagate the constraint panic")
	}()
	after, err := json.Marshal(game.CurrentState())
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("live state changed before constraint panic propagated")
	}
}

func benchmarkMoveAllTo(b *testing.B, constrained bool, unrelatedSlots int) {
	manager, err := NewGameManager(defaultTestGameDelegate(0), newTestStorageManager())
	if err != nil {
		b.Fatal(err)
	}
	game, err := manager.newGameImpl("", "")
	if err != nil {
		b.Fatal(err)
	}
	if err := game.setUp(0, nil, nil); err != nil {
		b.Fatal(err)
	}
	manager.Internals().AllowMutableConstraints(game)
	gameState, playerStates := concreteStates(game.CurrentState())
	if unrelatedSlots > 0 {
		if err := gameState.DownSizeStack.ExpandSize(unrelatedSlots); err != nil {
			b.Fatal(err)
		}
	}
	left := gameState.OtherStack
	right := playerStates[0].Hand
	for slot := 0; slot < 2; slot++ {
		if err := gameState.DrawDeck.First().MoveTo(left, slot); err != nil {
			b.Fatalf("populate source slot %d: %v", slot, err)
		}
	}
	if constrained {
		accept := func(ImmutableStack, []ImmutableComponentInstance, ImmutableState) error { return nil }
		mustAddStackConstraint(b, left, accept)
		mustAddStackConstraint(b, right, accept)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := left.MoveAllTo(right); err != nil {
			b.Fatal(err)
		}
		if err := right.MoveAllTo(left); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMoveAllToUnconstrained(b *testing.B) {
	benchmarkMoveAllTo(b, false, 0)
}

func BenchmarkMoveAllToConstrained(b *testing.B) {
	benchmarkMoveAllTo(b, true, 0)
}

func BenchmarkMoveAllToUnconstrainedLargeState(b *testing.B) {
	benchmarkMoveAllTo(b, false, 10_000)
}

func BenchmarkMoveAllToConstrainedLargeState(b *testing.B) {
	benchmarkMoveAllTo(b, true, 10_000)
}

func benchmarkRemainingCountPreflight(b *testing.B, constrained bool) {
	manager, err := NewGameManager(defaultTestGameDelegate(50), newTestStorageManager())
	if err != nil {
		b.Fatal(err)
	}
	game, err := manager.newGameImpl("", "")
	if err != nil {
		b.Fatal(err)
	}
	if err := game.setUp(0, nil, nil); err != nil {
		b.Fatal(err)
	}
	manager.Internals().AllowMutableConstraints(game)
	gameState, _ := concreteStates(game.CurrentState())
	source, destination := gameState.DrawDeck, gameState.MyBoard.SpaceAt(0)
	if constrained {
		mustAddStackConstraint(b, destination, func(ImmutableStack, []ImmutableComponentInstance, ImmutableState) error { return nil })
	}
	const target = 40
	b.ReportAllocs()
	b.ReportMetric(float64(target*(target+1)/2), "components_checked/op")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for remaining := target; remaining > 0; remaining-- {
			if err := source.MayMoveCountTo(destination, remaining); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkRemainingCountPreflightUnconstrained(b *testing.B) {
	benchmarkRemainingCountPreflight(b, false)
}

func BenchmarkRemainingCountPreflightConstrained(b *testing.B) {
	benchmarkRemainingCountPreflight(b, true)
}

func benchmarkMoveCountOne(b *testing.B, constrained bool, unrelatedSlots int) {
	manager, err := NewGameManager(defaultTestGameDelegate(0), newTestStorageManager())
	if err != nil {
		b.Fatal(err)
	}
	game, err := manager.newGameImpl("", "")
	if err != nil {
		b.Fatal(err)
	}
	if err := game.setUp(0, nil, nil); err != nil {
		b.Fatal(err)
	}
	manager.Internals().AllowMutableConstraints(game)
	gameState, playerStates := concreteStates(game.CurrentState())
	if err := gameState.DownSizeStack.ExpandSize(unrelatedSlots); err != nil {
		b.Fatal(err)
	}
	left, right := gameState.OtherStack, playerStates[0].Hand
	if err := gameState.DrawDeck.First().MoveToNextSlot(left); err != nil {
		b.Fatal(err)
	}
	if constrained {
		accept := func(ImmutableStack, []ImmutableComponentInstance, ImmutableState) error { return nil }
		mustAddStackConstraint(b, left, accept)
		mustAddStackConstraint(b, right, accept)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := left.MoveCountTo(right, 1); err != nil {
			b.Fatal(err)
		}
		if err := right.MoveCountTo(left, 1); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMoveCountOneUnconstrainedLargeState(b *testing.B) {
	benchmarkMoveCountOne(b, false, 10_000)
}

func BenchmarkMoveCountOneConstrainedLargeState(b *testing.B) {
	benchmarkMoveCountOne(b, true, 10_000)
}

func TestClearConstraints(t *testing.T) {
	game := testGameWithMutableConstraints(t)

	gameState, playerStates := concreteStates(game.CurrentState())

	drawStack := gameState.DrawDeck
	hand := playerStates[0].Hand

	// Add a constraint that always rejects.
	hand.AddConstraint(func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
		return errors.New("rejected")
	})

	// Move should fail.
	err := drawStack.First().MoveTo(hand, hand.SizedStack().FirstSlot())
	assert.For(t).ThatActual(err).IsNotNil()

	// Clear constraints.
	hand.ClearConstraints()

	// Move should now succeed.
	err = drawStack.First().MoveTo(hand, hand.SizedStack().FirstSlot())
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(hand.NumComponents()).Equals(1)
}

func TestConstraintsSurviveStateCopy(t *testing.T) {
	game := testGameWithMutableConstraints(t)

	_, playerStates := concreteStates(game.CurrentState())

	hand := playerStates[0].Hand

	// Add a constraint that always rejects.
	hand.AddConstraint(func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
		return errors.New("rejected")
	})

	// Get a copy of the state.
	stateCopy, copyErr := game.CurrentState().(*state).copy(false)
	assert.For(t).ThatActual(copyErr).IsNil()

	copiedGameState := stateCopy.ImmutableGameState().(*testGameState)
	copiedPlayerStates := make([]*testPlayerState, len(stateCopy.ImmutablePlayerStates()))
	for i, p := range stateCopy.ImmutablePlayerStates() {
		copiedPlayerStates[i] = p.(*testPlayerState)
	}

	copiedDraw := copiedGameState.DrawDeck
	copiedHand := copiedPlayerStates[0].Hand

	// The constraint should have survived the copy.
	err := copiedDraw.First().MoveTo(copiedHand, copiedHand.SizedStack().FirstSlot())
	assert.For(t).ThatActual(err).IsNotNil()
}

func TestConstraintNotCheckedDuringSetup(t *testing.T) {
	// DistributeComponentToStarterStack uses insertComponentAt directly,
	// not moveComonentImpl, so constraints should NOT be checked during
	// setup. We verify this by observing that testDefaultGame succeeds
	// even though we can't add constraints before setup. The test is
	// really just verifying the design: constraints on stacks are only
	// enforced at move time.
	game := testGameWithMutableConstraints(t)
	assert.For(t).ThatActual(game).IsNotNil()

	gameState, _ := concreteStates(game.CurrentState())
	// components should have been distributed to stacks during setup.
	assert.For(t).ThatActual(gameState.OtherStack.NumComponents()+gameState.DrawDeck.NumComponents()+gameState.DownSizeStack.NumComponents() > 0).Equals(true)
}

func TestConstraintOnGrowableStack(t *testing.T) {
	game := testGameWithMutableConstraints(t)

	gameState, _ := concreteStates(game.CurrentState())

	// DrawDeck is a growable stack.
	drawDeck := gameState.DrawDeck

	// Use OtherStack (SizedStack) as a source for testing.
	// First move a component from draw to OtherStack.
	otherStack := gameState.OtherStack
	err := drawDeck.First().MoveTo(otherStack, otherStack.SizedStack().FirstSlot())
	assert.For(t).ThatActual(err).IsNil()

	initialDrawCount := drawDeck.NumComponents()

	// Add a constraint to the growable draw stack.
	drawDeck.AddConstraint(func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
		return errors.New("no more draws")
	})

	// Try to move back from OtherStack to draw — should fail.
	err = otherStack.First().MoveToNextSlot(drawDeck)
	assert.For(t).ThatActual(err).IsNotNil()
	assert.For(t, "draw count unchanged").ThatActual(drawDeck.NumComponents()).Equals(initialDrawCount)

	// Clear and retry.
	drawDeck.ClearConstraints()
	err = otherStack.First().MoveToNextSlot(drawDeck)
	assert.For(t).ThatActual(err).IsNil()
}

func TestConstraintReceivesCorrectArgs(t *testing.T) {
	game := testGameWithMutableConstraints(t)

	gameState, playerStates := concreteStates(game.CurrentState())

	drawStack := gameState.DrawDeck
	hand := playerStates[0].Hand

	var receivedDest ImmutableStack
	var receivedAdded []ImmutableComponentInstance
	var receivedState ImmutableState

	hand.AddConstraint(func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
		receivedDest = dest
		receivedAdded = proposed
		receivedState = st
		return nil // allow the move
	})

	componentToMove := drawStack.ImmutableFirst()

	err := drawStack.First().MoveTo(hand, hand.SizedStack().FirstSlot())
	assert.For(t).ThatActual(err).IsNil()

	// Verify the constraint received the correct arguments.
	assert.For(t, "destination is hand").ThatActual(receivedDest).Equals(ImmutableStack(hand))
	assert.For(t, "one component added").ThatActual(len(receivedAdded)).Equals(1)
	assert.For(t, "added component matches").ThatActual(receivedAdded[0].Deck()).Equals(componentToMove.Deck())
	assert.For(t, "state is non-nil").ThatActual(receivedState).IsNotNil()
}

func TestAddConstraintBlockedAfterSetup(t *testing.T) {
	// Use testDefaultGame directly — no mutable constraints override.
	game := testDefaultGame(t, false)

	_, playerStates := concreteStates(game.CurrentState())
	hand := playerStates[0].Hand

	err := hand.AddConstraint(func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
		return nil
	})
	assert.For(t, "AddConstraint after setup").ThatActual(err).IsNotNil()

	clearErr := hand.ClearConstraints()
	assert.For(t, "ClearConstraints after setup").ThatActual(clearErr).IsNotNil()
}
