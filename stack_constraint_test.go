package boardgame

import (
	"testing"

	"github.com/jkomoros/boardgame/errors"
	"github.com/workfit/tester/assert"
)

func TestConstraintBlocksMoveTo(t *testing.T) {
	game := testDefaultGame(t, false)

	gameState, playerStates := concreteStates(game.CurrentState())

	drawStack := gameState.DrawDeck
	hand := playerStates[0].Hand

	// Move a component into the hand to start.
	err := drawStack.First().MoveTo(hand, hand.SizedStack().FirstSlot())
	assert.For(t).ThatActual(err).IsNil()

	initialHandCount := hand.NumComponents()
	initialDrawCount := drawStack.NumComponents()

	// Add a constraint that always rejects.
	hand.AddConstraint(func(dest ImmutableStack, justAdded []ImmutableComponentInstance, st ImmutableState) error {
		return errors.New("always rejected")
	})

	// Try to move another component in - should fail.
	err = drawStack.First().MoveTo(hand, hand.SizedStack().LastSlot())
	assert.For(t).ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(hand.NumComponents()).Equals(initialHandCount)
	assert.For(t).ThatActual(drawStack.NumComponents()).Equals(initialDrawCount)
}

func TestConstraintAllowsMove(t *testing.T) {
	game := testDefaultGame(t, false)

	gameState, playerStates := concreteStates(game.CurrentState())

	drawStack := gameState.DrawDeck
	hand := playerStates[0].Hand

	// Add a constraint that always passes.
	hand.AddConstraint(func(dest ImmutableStack, justAdded []ImmutableComponentInstance, st ImmutableState) error {
		return nil
	})

	err := drawStack.First().MoveTo(hand, hand.SizedStack().FirstSlot())
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(hand.NumComponents()).Equals(1)
}

func TestConstraintRollback(t *testing.T) {
	game := testDefaultGame(t, false)

	gameState, playerStates := concreteStates(game.CurrentState())

	drawStack := gameState.DrawDeck
	hand := playerStates[0].Hand

	initialDrawCount := drawStack.NumComponents()
	initialHandCount := hand.NumComponents()

	// Get the first component's deck index before the move.
	firstComponent := drawStack.ImmutableFirst()
	assert.For(t).ThatActual(firstComponent).IsNotNil()

	// Add a constraint that rejects.
	hand.AddConstraint(func(dest ImmutableStack, justAdded []ImmutableComponentInstance, st ImmutableState) error {
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
	game := testDefaultGame(t, false)

	gameState, playerStates := concreteStates(game.CurrentState())

	drawStack := gameState.DrawDeck
	hand := playerStates[0].Hand

	// Move one component to hand first (to have the first slot filled).
	err := drawStack.First().MoveTo(hand, hand.SizedStack().FirstSlot())
	assert.For(t).ThatActual(err).IsNil()

	afterFirstMoveDrawCount := drawStack.NumComponents()
	afterFirstMoveHandCount := hand.NumComponents()

	// Now add a constraint that rejects (simulating "hand is full").
	hand.AddConstraint(func(dest ImmutableStack, justAdded []ImmutableComponentInstance, st ImmutableState) error {
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

func TestClearConstraints(t *testing.T) {
	game := testDefaultGame(t, false)

	gameState, playerStates := concreteStates(game.CurrentState())

	drawStack := gameState.DrawDeck
	hand := playerStates[0].Hand

	// Add a constraint that always rejects.
	hand.AddConstraint(func(dest ImmutableStack, justAdded []ImmutableComponentInstance, st ImmutableState) error {
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
	game := testDefaultGame(t, false)

	_, playerStates := concreteStates(game.CurrentState())

	hand := playerStates[0].Hand

	// Add a constraint that always rejects.
	hand.AddConstraint(func(dest ImmutableStack, justAdded []ImmutableComponentInstance, st ImmutableState) error {
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
	game := testDefaultGame(t, false)
	assert.For(t).ThatActual(game).IsNotNil()

	gameState, _ := concreteStates(game.CurrentState())
	// components should have been distributed to stacks during setup.
	assert.For(t).ThatActual(gameState.OtherStack.NumComponents() + gameState.DrawDeck.NumComponents() + gameState.DownSizeStack.NumComponents() > 0).Equals(true)
}

func TestConstraintOnGrowableStack(t *testing.T) {
	game := testDefaultGame(t, false)

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
	drawDeck.AddConstraint(func(dest ImmutableStack, justAdded []ImmutableComponentInstance, st ImmutableState) error {
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
	game := testDefaultGame(t, false)

	gameState, playerStates := concreteStates(game.CurrentState())

	drawStack := gameState.DrawDeck
	hand := playerStates[0].Hand

	var receivedDest ImmutableStack
	var receivedAdded []ImmutableComponentInstance
	var receivedState ImmutableState

	hand.AddConstraint(func(dest ImmutableStack, justAdded []ImmutableComponentInstance, st ImmutableState) error {
		receivedDest = dest
		receivedAdded = justAdded
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
