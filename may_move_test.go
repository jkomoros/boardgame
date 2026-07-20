package boardgame

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame/errors"
	"github.com/workfit/tester/assert"
)

func TestMoveCountTo(t *testing.T) {
	t.Run("MovesExactCountInOrder", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())
		source := gs.DrawDeck
		destination := ps[0].Hand
		first := source.First()
		second := source.ComponentAt(1)

		if err := source.MayMoveCountTo(destination, 2); err != nil {
			t.Fatalf("MayMoveCountTo: %v", err)
		}
		if got := destination.NumComponents(); got != 0 {
			t.Fatalf("MayMoveCountTo mutated destination: got %d components", got)
		}
		if err := source.MoveCountTo(destination, 2); err != nil {
			t.Fatalf("MoveCountTo: %v", err)
		}
		if got := destination.ComponentAt(0); got != first {
			t.Fatalf("destination[0] = %v, want original first component", got)
		}
		if got := destination.ComponentAt(1); got != second {
			t.Fatalf("destination[1] = %v, want original second component", got)
		}
		verifyContainingComponent(t, game.CurrentState(), game.Manager().Chest().Deck("test"))
	})

	t.Run("RejectsInvalidCountsWithoutMutation", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())
		before, err := game.CurrentState().(*state).copy(false)
		if err != nil {
			t.Fatalf("copy state: %v", err)
		}

		for _, count := range []int{-1, gs.DrawDeck.NumComponents() + 1} {
			if err := gs.DrawDeck.MayMoveCountTo(ps[0].Hand, count); err == nil {
				t.Errorf("MayMoveCountTo count %d unexpectedly succeeded", count)
			}
			if err := gs.DrawDeck.MoveCountTo(ps[0].Hand, count); err == nil {
				t.Errorf("MoveCountTo count %d unexpectedly succeeded", count)
			}
			assertPersistedStatesEqual(t, game.CurrentState(), before)
		}
	})

	t.Run("ZeroStillValidatesEndpoints", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())
		if err := gs.DrawDeck.MoveCountTo(ps[0].Hand, 0); err != nil {
			t.Fatalf("zero-count move: %v", err)
		}
		if err := gs.DrawDeck.MoveCountTo(nil, 0); err == nil {
			t.Fatal("zero-count move to nil unexpectedly succeeded")
		}
		if err := gs.DrawDeck.MayMoveCountTo(gs.DrawDeck, 0); err == nil {
			t.Fatal("zero-count preflight to same stack unexpectedly succeeded")
		}
	})

	t.Run("RejectsInsufficientDestinationCapacity", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())
		if err := gs.DrawDeck.MayMoveCountTo(ps[0].Hand, 3); err == nil || !strings.Contains(err.Error(), "space") {
			t.Fatalf("capacity preflight error = %v", err)
		}
		if err := gs.DrawDeck.MoveCountTo(ps[0].Hand, 3); err == nil || !strings.Contains(err.Error(), "space") {
			t.Fatalf("capacity move error = %v", err)
		}
		if got := ps[0].Hand.NumComponents(); got != 0 {
			t.Fatalf("failed transfer moved %d components", got)
		}
	})

	t.Run("LateConstraintFailureIsAtomic", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())
		calls := 0
		if err := ps[0].Hand.AddConstraint(func(dest ImmutableStack, proposed []ImmutableComponentInstance, _ ImmutableState) error {
			calls++
			if dest.NumComponents()+len(proposed) > 1 {
				return errors.New("only one component allowed")
			}
			return nil
		}); err != nil {
			t.Fatalf("add constraint: %v", err)
		}
		before, err := game.CurrentState().(*state).copy(false)
		if err != nil {
			t.Fatalf("copy state: %v", err)
		}

		err = gs.DrawDeck.MayMoveCountTo(ps[0].Hand, 2)
		if err == nil || !strings.Contains(err.Error(), "only one") {
			t.Fatalf("MayMoveCountTo error = %v", err)
		}
		if calls != 2 {
			t.Fatalf("MayMoveCountTo constraint calls = %d, want 2", calls)
		}
		assertPersistedStatesEqual(t, game.CurrentState(), before)

		calls = 0
		err = gs.DrawDeck.MoveCountTo(ps[0].Hand, 2)
		if err == nil || !strings.Contains(err.Error(), "only one") {
			t.Fatalf("MoveCountTo error = %v", err)
		}
		if calls != 2 {
			t.Fatalf("constraint calls = %d, want 2", calls)
		}
		assertPersistedStatesEqual(t, game.CurrentState(), before)
		verifyContainingComponent(t, game.CurrentState(), game.Manager().Chest().Deck("test"))
	})

	t.Run("SuccessfulConstraintRunsOncePerComponent", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())
		calls := 0
		if err := ps[0].Hand.AddConstraint(func(dest ImmutableStack, proposed []ImmutableComponentInstance, _ ImmutableState) error {
			calls++
			if dest.NumComponents()+len(proposed) > 2 {
				return errors.New("too many components")
			}
			return nil
		}); err != nil {
			t.Fatalf("add constraint: %v", err)
		}

		if err := gs.DrawDeck.MoveCountTo(ps[0].Hand, 2); err != nil {
			t.Fatalf("MoveCountTo: %v", err)
		}
		if calls != 2 {
			t.Fatalf("constraint calls = %d, want 2", calls)
		}
		if got := ps[0].Hand.NumComponents(); got != 2 {
			t.Fatalf("destination count = %d, want 2", got)
		}
		verifyContainingComponent(t, game.CurrentState(), game.Manager().Chest().Deck("test"))
	})

	t.Run("SparseSizedSourceUsesComponentOrder", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())
		source := game.Manager().Chest().Deck("test").NewSizedStack(4)
		attachStackForPrimitiveTest(game.CurrentState().(*state), source)
		first := gs.DrawDeck.First()
		if err := first.MoveTo(source, 1); err != nil {
			t.Fatalf("seed source slot 1: %v", err)
		}
		second := gs.DrawDeck.First()
		if err := second.MoveTo(source, 3); err != nil {
			t.Fatalf("seed source slot 3: %v", err)
		}

		if err := source.MoveCountTo(ps[0].Hand, 2); err != nil {
			t.Fatalf("MoveCountTo: %v", err)
		}
		if ps[0].Hand.ComponentAt(0) != first || ps[0].Hand.ComponentAt(1) != second {
			t.Fatal("sparse sized source did not preserve first-to-last component order")
		}
	})

	t.Run("ExactAllEmptiesSource", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, _ := concreteStates(game.CurrentState())
		source, destination := gs.OtherStack, gs.MyBoard.SpaceAt(0)
		for slot := 0; slot < source.Len(); slot++ {
			if err := gs.DrawDeck.First().MoveTo(source, slot); err != nil {
				t.Fatalf("seed source slot %d: %v", slot, err)
			}
		}
		if err := source.MoveCountTo(destination, source.NumComponents()); err != nil {
			t.Fatalf("MoveCountTo exact-all: %v", err)
		}
		if source.NumComponents() != 0 || destination.NumComponents() != 2 {
			t.Fatalf("counts after exact-all = source %d, destination %d", source.NumComponents(), destination.NumComponents())
		}
	})

	t.Run("MayAndMoveErrorsMatch", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())
		otherDeckDestination := NewDeck().NewSizedStack(1)
		attachStackForPrimitiveTest(game.CurrentState().(*state), otherDeckDestination)

		tests := []struct {
			name  string
			dest  ImmutableStack
			count int
		}{
			{name: "nil", dest: nil, count: 0},
			{name: "same", dest: gs.DrawDeck, count: 0},
			{name: "cross deck", dest: otherDeckDestination, count: 0},
			{name: "insufficient source", dest: ps[0].Hand, count: gs.DrawDeck.NumComponents() + 1},
			{name: "insufficient capacity", dest: ps[0].Hand, count: ps[0].Hand.Len() + 1},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				mayErr := gs.DrawDeck.MayMoveCountTo(test.dest, test.count)
				moveDest, _ := test.dest.(Stack)
				moveErr := gs.DrawDeck.MoveCountTo(moveDest, test.count)
				if mayErr == nil || moveErr == nil {
					t.Fatalf("errors = MayMoveCountTo %v, MoveCountTo %v; want both non-nil", mayErr, moveErr)
				}
				if mayErr.Error() != moveErr.Error() {
					t.Fatalf("errors differ: MayMoveCountTo %q, MoveCountTo %q", mayErr, moveErr)
				}
			})
		}
	})

	t.Run("MergedEndpointsAreRejected", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, _ := concreteStates(game.CurrentState())
		if err := gs.DrawDeck.MayMoveCountTo(gs.MyMergedStack, 0); err == nil || !strings.Contains(err.Error(), "physical stack") {
			t.Fatalf("merged destination error = %v", err)
		}
		if err := gs.MyMergedStack.MayMoveCountTo(gs.DrawDeck, 0); err == nil || !strings.Contains(err.Error(), "MergedStacks") {
			t.Fatalf("merged source error = %v", err)
		}
	})
}

func TestMayMoveTo(t *testing.T) {
	game := testGameWithMutableConstraints(t)

	gameState, playerStates := concreteStates(game.CurrentState())

	drawStack := gameState.DrawDeck
	hand := playerStates[0].Hand

	t.Run("HappyPath", func(t *testing.T) {
		first := drawStack.First()
		assert.For(t).ThatActual(first).IsNotNil()

		err := first.MayMoveTo(hand)
		assert.For(t).ThatActual(err).IsNil()
	})

	t.Run("NilDestination", func(t *testing.T) {
		first := drawStack.First()
		err := first.MayMoveTo(nil)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(err.Error()).Equals("destination stack is nil")
	})

	t.Run("SameStack", func(t *testing.T) {
		first := drawStack.First()
		err := first.MayMoveTo(drawStack)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(err.Error()).Equals("source and destination are the same stack")
	})

	t.Run("NoSlotsRemaining", func(t *testing.T) {
		// Fill up the hand (2 slots)
		err := drawStack.First().MoveTo(hand, hand.SizedStack().FirstSlot())
		assert.For(t).ThatActual(err).IsNil()
		err = drawStack.First().MoveTo(hand, hand.SizedStack().FirstSlot())
		assert.For(t).ThatActual(err).IsNil()

		assert.For(t).ThatActual(hand.SlotsRemaining()).Equals(0)

		first := drawStack.First()
		err = first.MayMoveTo(hand)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(err.Error()).Equals("destination stack has no slots remaining")
	})

	t.Run("ConstraintViolation", func(t *testing.T) {
		game2 := testGameWithMutableConstraints(t)
		gs2, ps2 := concreteStates(game2.CurrentState())

		ps2[0].Hand.AddConstraint(func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
			return errors.New("always rejected")
		})

		first := gs2.DrawDeck.First()
		err := first.MayMoveTo(ps2[0].Hand)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(err.Error()).Equals("always rejected")
	})

	t.Run("ConstraintPasses", func(t *testing.T) {
		game2 := testGameWithMutableConstraints(t)
		gs2, ps2 := concreteStates(game2.CurrentState())

		ps2[0].Hand.AddConstraint(func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
			return nil
		})

		first := gs2.DrawDeck.First()
		err := first.MayMoveTo(ps2[0].Hand)
		assert.For(t).ThatActual(err).IsNil()
	})
}

func TestMayMoveToSlot(t *testing.T) {
	t.Run("SizedStackEmptySlot", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())

		first := gs.DrawDeck.First()
		err := first.MayMoveToSlot(ps[0].Hand, 0)
		assert.For(t).ThatActual(err).IsNil()
	})

	t.Run("SizedStackOccupiedSlot", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())

		// Fill slot 0
		err := gs.DrawDeck.First().MoveTo(ps[0].Hand, 0)
		assert.For(t).ThatActual(err).IsNil()

		// Try to move to the same slot
		second := gs.DrawDeck.First()
		err = second.MayMoveToSlot(ps[0].Hand, 0)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(err.Error()).Equals("slot 0 is already occupied")
	})

	t.Run("NegativeSlotIndex", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())

		first := gs.DrawDeck.First()
		err := first.MayMoveToSlot(ps[0].Hand, -1)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(err.Error()).Equals("slot index must be non-negative")
	})

	t.Run("OutOfRangeSizedStack", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())

		first := gs.DrawDeck.First()
		err := first.MayMoveToSlot(ps[0].Hand, 99)
		assert.For(t).ThatActual(err).IsNotNil()
	})

	t.Run("GrowableStackValidInsertion", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())

		// Move a component from hand to draw deck (growable) at end position
		err := gs.DrawDeck.First().MoveTo(ps[0].Hand, 0)
		assert.For(t).ThatActual(err).IsNil()

		// MayMoveToSlot at index 0 (beginning) of draw deck
		first := ps[0].Hand.First()
		err = first.MayMoveToSlot(gs.DrawDeck, 0)
		assert.For(t).ThatActual(err).IsNil()

		// MayMoveToSlot at Len() (end) should also work
		err = first.MayMoveToSlot(gs.DrawDeck, gs.DrawDeck.Len())
		assert.For(t).ThatActual(err).IsNil()

		// MayMoveToSlot past Len() should fail
		err = first.MayMoveToSlot(gs.DrawDeck, gs.DrawDeck.Len()+1)
		assert.For(t).ThatActual(err).IsNotNil()
	})
}

func TestMayMoveAllTo(t *testing.T) {
	t.Run("HappyPath", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())

		// OtherStack has 2 slots. Move 2 components from draw to it.
		err := gs.DrawDeck.First().MoveTo(gs.OtherStack, 0)
		assert.For(t).ThatActual(err).IsNil()
		err = gs.DrawDeck.First().MoveTo(gs.OtherStack, 1)
		assert.For(t).ThatActual(err).IsNil()

		// Now MayMoveAllTo from OtherStack to hand (also 2 slots)
		err = gs.OtherStack.MayMoveAllTo(ps[0].Hand)
		assert.For(t).ThatActual(err).IsNil()
	})

	t.Run("NotEnoughSlots", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())

		// DrawDeck has many components, hand only 2 slots
		err := gs.DrawDeck.MayMoveAllTo(ps[0].Hand)
		assert.For(t).ThatActual(err).IsNotNil()
	})

	t.Run("SameStack", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, _ := concreteStates(game.CurrentState())

		err := gs.DrawDeck.MayMoveAllTo(gs.DrawDeck)
		assert.For(t).ThatActual(err).IsNotNil()
	})

	t.Run("EmptySource", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())

		// OtherStack starts empty
		assert.For(t).ThatActual(gs.OtherStack.NumComponents()).Equals(0)

		err := gs.OtherStack.MayMoveAllTo(ps[0].Hand)
		assert.For(t).ThatActual(err).IsNil()
	})

	t.Run("ConstraintRejectsMidSequence", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())

		// Move 2 components from draw to OtherStack
		err := gs.DrawDeck.First().MoveTo(gs.OtherStack, 0)
		assert.For(t).ThatActual(err).IsNil()
		err = gs.DrawDeck.First().MoveTo(gs.OtherStack, 1)
		assert.For(t).ThatActual(err).IsNil()

		// Add a constraint to hand that accepts only the first component
		callCount := 0
		ps[0].Hand.AddConstraint(func(dest ImmutableStack, proposed []ImmutableComponentInstance, st ImmutableState) error {
			callCount++
			if dest.NumComponents()+len(proposed) > 1 {
				return errors.New("hand accepts at most 1 component")
			}
			return nil
		})

		err = gs.OtherStack.MayMoveAllTo(ps[0].Hand)
		assert.For(t).ThatActual(err).IsNotNil()
	})

	t.Run("OriginalStateUnchanged", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())

		// Move 2 components from draw to OtherStack
		err := gs.DrawDeck.First().MoveTo(gs.OtherStack, 0)
		assert.For(t).ThatActual(err).IsNil()
		err = gs.DrawDeck.First().MoveTo(gs.OtherStack, 1)
		assert.For(t).ThatActual(err).IsNil()

		otherCount := gs.OtherStack.NumComponents()
		handCount := ps[0].Hand.NumComponents()

		// MayMoveAllTo should succeed
		err = gs.OtherStack.MayMoveAllTo(ps[0].Hand)
		assert.For(t).ThatActual(err).IsNil()

		// But original stacks should be unchanged
		assert.For(t, "other unchanged").ThatActual(gs.OtherStack.NumComponents()).Equals(otherCount)
		assert.For(t, "hand unchanged").ThatActual(ps[0].Hand.NumComponents()).Equals(handCount)
	})

	t.Run("MergedStackSource", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, ps := concreteStates(game.CurrentState())

		err := gs.MyMergedStack.MayMoveAllTo(ps[0].Hand)
		assert.For(t).ThatActual(err).IsNotNil()
	})
}

func TestMaySwapComponents(t *testing.T) {
	t.Run("HappyPath", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, _ := concreteStates(game.CurrentState())

		// DrawDeck has multiple components
		assert.For(t).ThatActual(gs.DrawDeck.Len() >= 2).IsTrue()

		err := gs.DrawDeck.MaySwapComponents(0, 1)
		assert.For(t).ThatActual(err).IsNil()
	})

	t.Run("NegativeIndex", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, _ := concreteStates(game.CurrentState())

		err := gs.DrawDeck.MaySwapComponents(-1, 0)
		assert.For(t).ThatActual(err).IsNotNil()
	})

	t.Run("OutOfRange", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, _ := concreteStates(game.CurrentState())

		err := gs.DrawDeck.MaySwapComponents(0, 999)
		assert.For(t).ThatActual(err).IsNotNil()
	})

	t.Run("SameIndex", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, _ := concreteStates(game.CurrentState())

		err := gs.DrawDeck.MaySwapComponents(0, 0)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(err.Error()).Equals("i and j were the same")
	})

	t.Run("SizedStack", func(t *testing.T) {
		game := testGameWithMutableConstraints(t)
		gs, _ := concreteStates(game.CurrentState())

		// OtherStack is a SizedStack with 2 slots - swap works even on empty slots
		err := gs.OtherStack.MaySwapComponents(0, 1)
		assert.For(t).ThatActual(err).IsNil()
	})
}
