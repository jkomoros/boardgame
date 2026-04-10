package boardgame

import (
	"testing"

	"github.com/jkomoros/boardgame/errors"
	"github.com/workfit/tester/assert"
)

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
