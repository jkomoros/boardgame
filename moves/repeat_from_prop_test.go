package moves

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/workfit/tester/assert"
)

// This file tests RepeatFromProp (#644, design spec §7): a repeat count
// resolved DYNAMICALLY from live state at match time, rather than baked in
// at auto.Config time via Repeat(CountExactly(n), ...). It reuses the shared
// gameState/playerState/phaseEnum/newGameManager fixture from game_test.go —
// gameState.Counter (a plain int field) stands in for the design spec §7
// example's "game.RoundsThisTurn".

//boardgame:codegen
type moveRepeatFromPropUnit struct {
	Default
}

func (m *moveRepeatFromPropUnit) Apply(state boardgame.State) error {
	return nil
}

// moveRepeatGuard is a NON-fixup stand-in for moves.NoOp: it satisfies the
// unexported isNoOper interface (no_op.go) so AddOrderedForPhase's "ends
// with StartPhase or NoOp" sanity check accepts it as the group's terminal
// item, but — unlike moves.NoOp, which embeds FixUp — it embeds Default
// directly, so it does NOT auto-fire via the fixup loop. That distinction
// matters here: gameState.Counter (this fixture's dynamic repeat-count
// property) defaults to its zero value, 0, and RepeatFromProp("game.Counter",
// ...) is legal with ZERO repeats when Counter is 0 — meaning the guard
// would otherwise become immediately legal-and-fixup-eligible the moment the
// game is created, before the test gets a chance to set Counter, silently
// consuming itself and corrupting every assertion below.
//
//boardgame:codegen
type moveRepeatGuard struct {
	Default
}

func (m *moveRepeatGuard) Apply(state boardgame.State) error {
	return nil
}

func (m *moveRepeatGuard) isNoOp() bool {
	return true
}

// repeatFromPropMoveInstaller wires a single move type, "Repeat Unit", whose
// legal move-tape progression is RepeatFromProp("game.Counter", ...): legal
// to propose as long as fewer than gameState.Counter copies of it have
// already appeared since the last phase transition. The Serial's second
// group ("Repeat Guard", moveRepeatGuard) exists only to satisfy
// AddOrderedForPhase's "ends with StartPhase or NoOp" sanity check — see
// moveRepeatGuard's doc comment for why it isn't moves.NoOp itself.
func repeatFromPropMoveInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {
	auto := NewAutoConfigurer(manager.Delegate())
	return AddOrderedForPhase(phaseSetUp,
		RepeatFromProp("game.Counter", auto.MustConfig(
			new(moveRepeatFromPropUnit),
			WithMoveName("Repeat Unit"),
		)),
		auto.MustConfig(new(moveRepeatGuard), WithMoveName("Repeat Guard")),
	)
}

// setCounter sets gameState.Counter on game's live current state (via
// ReadSetter, not direct field access, to match this package's other tests'
// idiom for mutating a fixture state in place).
func setCounter(t *testing.T, game *boardgame.Game, n int) {
	t.Helper()
	state, ok := game.CurrentState().(boardgame.State)
	if !ok {
		t.Fatal("moves: repeatFromProp fixture: CurrentState() was not mutable")
	}
	if err := state.GameState().ReadSetter().SetIntProp("Counter", n); err != nil {
		t.Fatalf("moves: setting Counter: %v", err)
	}
}

// TestRepeatFromPropDynamicCount is the state-driven repeat count test the
// Task 7 brief calls for: the SAME move type ("Repeat Unit") is legal or
// illegal to propose a second time in a row depending ENTIRELY on the
// CURRENT value of gameState.Counter at evaluation time — proof that the
// count is resolved dynamically (per legal.Context.State), not baked in once
// at auto.Config time.
func TestRepeatFromPropDynamicCount(t *testing.T) {

	manager, err := newGameManager(repeatFromPropMoveInstaller)
	assert.For(t, "manager").ThatActual(err).IsNil()

	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	// Counter = 2: the FIRST "Repeat Unit" proposal (empty tape so far) is
	// legal — 0 -> 1 is within [0, 2].
	setCounter(t, game, 2)
	move := game.MoveByName("Repeat Unit")
	assert.For(t, "move lookup").ThatActual(move).IsNotNil()
	firstErr := move.Legal(game.CurrentState(), 0)
	assert.For(t, "first proposal, Counter=2").ThatActual(firstErr).IsNil()

	// Counter = 1: same empty-tape check, ALSO legal — 0 -> 1 is within
	// [0, 1] too (the boundary case).
	setCounter(t, game, 1)
	move = game.MoveByName("Repeat Unit")
	boundaryErr := move.Legal(game.CurrentState(), 0)
	assert.For(t, "first proposal, Counter=1").ThatActual(boundaryErr).IsNil()

	// Actually apply the first "Repeat Unit" (Counter is still 1 on the live
	// state at proposal time), so the move tape now has one real entry.
	applyErr := <-game.ProposeMove(move, 0)
	assert.For(t, "apply first Repeat Unit").ThatActual(applyErr).IsNil()

	// With Counter RE-SET to 1 on the post-apply state, a SECOND "Repeat
	// Unit" proposal must now be illegal: the tape already has one match,
	// which is the full count.
	setCounter(t, game, 1)
	secondMove := game.MoveByName("Repeat Unit")
	secondErr := secondMove.Legal(game.CurrentState(), 0)
	assert.For(t, "second proposal, Counter=1").ThatActual(secondErr).IsNotNil()

	// But with Counter set to 2 on that SAME post-apply state (one real
	// prior "Repeat Unit" already on the tape), a second proposal IS legal —
	// changing only the live state's Counter value flips the verdict,
	// proving the count is re-resolved from state at match time, not cached
	// from the first evaluation.
	setCounter(t, game, 2)
	thirdMove := game.MoveByName("Repeat Unit")
	thirdErr := thirdMove.Legal(game.CurrentState(), 0)
	assert.For(t, "second proposal, Counter=2").ThatActual(thirdErr).IsNil()
}

// TestRepeatFromPropSatisfiedWithoutContextFailsClosed covers RepeatFromProp's
// backward-compatibility stub: calling Satisfied directly (bypassing
// matchTape, so no legal.Context is available to resolve the count) must
// fail closed with a descriptive error, never silently treat the count as 0.
func TestRepeatFromPropSatisfiedWithoutContextFailsClosed(t *testing.T) {
	inner := NewGroupableMoveConfig(newMoveConfig("Repeat Unit", new(moveRepeatFromPropUnit), nil))
	group := RepeatFromProp("game.Counter", inner)

	tape := &MoveGroupHistoryItem{MoveName: "Repeat Unit"}

	_, err := group.Satisfied(tape)
	assert.For(t, "Satisfied without context").ThatActual(err).IsNotNil()
}
