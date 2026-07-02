package blackjack

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/workfit/tester/assert"
)

// TestForceFinishTurnAdminBypassesTurnDone exercises the spec §9.3 host-
// SkipTurn flow: in a live blackjack game, the current player hasn't
// busted or chosen to stand, so playerState.TurnDone() returns an error.
// Vanilla FinishTurn would be rejected for that reason; ForceFinishTurn
// proposed by AdminPlayerIndex should accept and advance the current
// player anyway. This is the bug-fix-purpose of the move; without this
// test a refactor that broke the FinishTurn.Apply inheritance would
// pass the existing legality-check-only tests.
func TestForceFinishTurnAdminBypassesTurnDone(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())
	if err != nil {
		t.Fatalf("NewGameManager: %v", err)
	}
	if manager == nil {
		t.Fatal("NewGameManager returned nil")
	}

	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("NewDefaultGame: %v", err)
	}
	if game == nil {
		t.Fatal("NewDefaultGame returned nil")
	}

	// Confirm the initial setup deals cards and the game is in play
	// state — i.e. we're at a point where the current player has cards
	// in hand and TurnDone() returns "they have neither busted nor
	// decided to stand".
	initialCurrent := game.CurrentState().CurrentPlayerIndex()
	if initialCurrent < 0 {
		t.Fatalf("Expected a valid current player after game setup, got %v", initialCurrent)
	}

	// Vanilla FinishTurn should be rejected — TurnDone() refuses.
	finishTurn := game.MoveByName("Finish Turn")
	assert.For(t).ThatActual(finishTurn).IsNotNil()
	finishErr := <-game.ProposeMove(finishTurn, initialCurrent)
	// Either Legal() refuses at proposal time, or the move applies
	// (degenerate — the current player IS done somehow). Tolerate
	// either since the absolute state depends on dealer's hand luck.
	if finishErr == nil {
		t.Log("Note: vanilla FinishTurn happened to apply — current player was apparently done. Test still exercises ForceFinishTurn below.")
	}

	currentBeforeForce := game.CurrentState().CurrentPlayerIndex()

	// ForceFinishTurn proposed by AdminPlayerIndex must succeed regardless
	// of TurnDone — that's the whole point of the move (spec §9.3).
	force := game.MoveByName("Force Finish Turn")
	assert.For(t).ThatActual(force).IsNotNil()
	forceErr := <-game.ProposeMove(force, boardgame.AdminPlayerIndex)
	assert.For(t).ThatActual(forceErr).IsNil()

	// And the current player should have advanced (or the game finished,
	// in which case CurrentPlayerIndex may be -1).
	currentAfterForce := game.CurrentState().CurrentPlayerIndex()
	if currentAfterForce == currentBeforeForce {
		t.Errorf("ForceFinishTurn applied but current player did not advance: %v -> %v",
			currentBeforeForce, currentAfterForce)
	}
}

// TestForceFinishTurnRejectsNonAdmin pins the security contract: a regular
// player cannot use ForceFinishTurn to skip their own (or anyone else's)
// turn. Only AdminPlayerIndex (server-initiated) is accepted.
func TestForceFinishTurnRejectsNonAdmin(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())
	if err != nil {
		t.Fatalf("NewGameManager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("NewDefaultGame: %v", err)
	}

	force := game.MoveByName("Force Finish Turn")
	assert.For(t).ThatActual(force).IsNotNil()

	// Propose as player 0 (or whoever the current player is). Should be
	// rejected with the "AdminPlayerIndex required" error.
	proposeErr := <-game.ProposeMove(force, boardgame.PlayerIndex(0))
	assert.For(t).ThatActual(proposeErr).IsNotNil()
}
