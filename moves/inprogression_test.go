package moves

import (
	"testing"

	"github.com/jkomoros/boardgame/legal"
	"github.com/workfit/tester/assert"
)

/*
This file functionally tests the "inProgression" framework wrapper predicate
(catalog_framework.go): the one framework precondition predicate (of the
four stable names in design spec §2) that lives in package moves rather than
package legal — see catalog_framework.go's doc comment for why. It reuses
preconditions_test.go's freezeMoveInstaller fixture (the same
Freeze-Progression-A/B moves the TestLegalChainStringFreeze string-freeze
test exercises via the frozen imperative chain), so this predicate and that
frozen chain are provably being evaluated against the exact same game setup
— just through two different entry points (Predicate.Evaluate here,
Default.Legal() there).
*/

// resolveInProgressionForTest resolves spec (built via inProgressionSpec)
// against inProgressionConstructor directly, failing the test on error.
func resolveInProgressionForTest(t *testing.T, moveName string) *legal.Predicate {
	t.Helper()
	pred, err := inProgressionConstructor().Constructor(inProgressionSpec(moveName), nil, nil)
	if err != nil {
		t.Fatalf("moves: resolving inProgression spec for %q: %v", moveName, err)
	}
	return pred
}

// TestInProgressionMatch covers the "inProgression" predicate's Pass path:
// the next move in an empty tape matches the progression's first entry.
func TestInProgressionMatch(t *testing.T) {
	manager, err := newGameManager(freezeMoveInstaller)
	assert.For(t, "manager").ThatActual(err).IsNil()

	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	// Transition into phaseNormalPlayDrawCard, same as
	// TestLegalChainStringFreeze, so the progression moves are reachable.
	startMove := game.MoveByName("Freeze Start Normal Play")
	assert.For(t, "start move").ThatActual(startMove).IsNotNil()
	startErr := <-game.ProposeMove(startMove, 0)
	assert.For(t, "start move propose").ThatActual(startErr).IsNil()

	state := game.CurrentState()

	pred := resolveInProgressionForTest(t, "Freeze Progression A")
	if len(pred.Reads) != 1 || pred.Reads[0].Path != "game.Phase" {
		t.Fatalf("Reads = %+v, want a single game.Phase entry", pred.Reads)
	}

	// Pass: with an empty tape (no progression moves proposed yet since the
	// phase transition), "Freeze Progression A" is next.
	v := pred.Evaluate(legal.Context{State: state, ProposerPlayerIndex: 0})
	if v.Outcome != legal.Pass {
		t.Fatalf("Freeze Progression A, empty tape: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}
}

// TestInProgressionReject covers the "inProgression" predicate's Fail path,
// pinning that its detail binding is byte-identical to the frozen chain's
// own string (TestLegalChainStringFreeze in preconditions_test.go pins the
// exact same string via Default.Legal(), not Predicate.Evaluate — this test
// proves the two entry points agree).
func TestInProgressionReject(t *testing.T) {
	manager, err := newGameManager(freezeMoveInstaller)
	assert.For(t, "manager").ThatActual(err).IsNil()

	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	startMove := game.MoveByName("Freeze Start Normal Play")
	assert.For(t, "start move").ThatActual(startMove).IsNotNil()
	startErr := <-game.ProposeMove(startMove, 0)
	assert.For(t, "start move propose").ThatActual(startErr).IsNil()

	state := game.CurrentState()

	// Fail: "Freeze Progression B" proposed before "Freeze Progression A"
	// (same case TestLegalChainStringFreeze pins via the frozen chain).
	pred := resolveInProgressionForTest(t, "Freeze Progression B")
	v := pred.Evaluate(legal.Context{State: state, ProposerPlayerIndex: 0})
	if v.Outcome != legal.Fail {
		t.Fatalf("Freeze Progression B, empty tape: Outcome = %v, want Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplateInProgression {
		t.Fatalf("Fail Message = %+v, want template %q", v.Message, legal.TemplateInProgression)
	}
	wantDetail := "Move name does not match: Freeze Progression B is not Freeze Progression A"
	if got := v.Message.Bindings["detail"]; got.S == nil || *got.S != wantDetail {
		t.Fatalf("detail binding = %+v, want %q", got, wantDetail)
	}
}

// TestInProgressionUnknown covers the "inProgression" predicate's Unknown
// paths: nil state, and a move name with no registered move type.
func TestInProgressionUnknown(t *testing.T) {
	manager, err := newGameManager(freezeMoveInstaller)
	assert.For(t, "manager").ThatActual(err).IsNil()

	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	state := game.CurrentState()

	pred := resolveInProgressionForTest(t, "Freeze Progression A")

	if v := pred.Evaluate(legal.Context{State: nil, ProposerPlayerIndex: 0}); v.Outcome != legal.Unknown {
		t.Fatalf("nil state: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}

	unregistered := resolveInProgressionForTest(t, "No Such Move")
	if v := unregistered.Evaluate(legal.Context{State: state, ProposerPlayerIndex: 0}); v.Outcome != legal.Unknown {
		t.Fatalf("unregistered move name: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}
}
