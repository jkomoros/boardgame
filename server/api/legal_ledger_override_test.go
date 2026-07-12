package api

import (
	"errors"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
	"github.com/jkomoros/boardgame/moves"
)

/*
This file is the regression harness for the audit's F1 finding (docs/
superpowers/plans/2026-07-12-legality-footgun-batch.md, Task 2): an
opted-in move's LegalForPlayer/LegalForPlayerError/LegalForAnyone must be
derived from the move's REAL Legal() (the exact call game.ProposeMove
makes -- game.go's applyMove calls move.Legal(currentState, proposer)),
not from the declarative plan's verdict alone. A super-calling Legal()
override with imperative residue is explicitly blessed by the design
spec's prime-guarantee rule 4 -- the plan verdict does not see the
residue, so deriving the booleans from the plan enables buttons
ProposeMove would reject (residue direction) and disables buttons
ProposeMove would accept (early-return-nil direction).

The fixture reuses legal_ledger_fixture_test.go's state/reader/storage
scaffolding, adding a delegate whose FinishSetUp seeds game.HiddenCounter
to a sentinel value NO example state ever has (GameManager.ExampleState is
an emptyState -- FinishSetUp never runs on it, so HiddenCounter is 0 during
the boot reachability probe). That lets both override moves super-call on
the probe path (keeping boot green) while behaving differently on a real
game's state.
*/

// legalLedgerOverrideCounterSeed is the game.HiddenCounter value
// legalLedgerOverrideDelegate.FinishSetUp seeds into every REAL game. The
// boot probe runs against ExampleState (HiddenCounter zero-value 0), so a
// fixture move can key "am I on a real game state?" off this sentinel.
const legalLedgerOverrideCounterSeed = 7

// legalLedgerResidueError is the exact error text legalLedgerMoveResidue's
// imperative residue rejects with, so tests can assert the served
// LegalForPlayerError is the residue's own text, byte-for-byte.
const legalLedgerResidueError = "the hidden counter is still warming up"

// legalLedgerOverrideDelegate is legalLedgerDelegate with two moves that
// override Legal() around their declarative plan:
//
//   - "Residue": plan PASSES (propAtLeast(game.HiddenCounter, 5); seeded
//     counter is 7) but a super-calling override's imperative residue
//     REJECTS (counter < 100). Ground truth: illegal. Plan alone: legal.
//   - "Early Nil": plan FAILS (propAtLeast(game.HiddenCounter, 1000)) but
//     the override early-returns nil on any seeded (real-game) state
//     without consulting the plan. Ground truth: legal. Plan alone:
//     illegal.
type legalLedgerOverrideDelegate struct {
	legalLedgerDelegate
}

func (d *legalLedgerOverrideDelegate) FinishSetUp(state boardgame.State) error {
	if err := d.legalLedgerDelegate.FinishSetUp(state); err != nil {
		return err
	}
	gameState, ok := state.GameState().(*legalLedgerGameState)
	if !ok {
		return errors.New("legal ledger override fixture: game state was not *legalLedgerGameState")
	}
	gameState.HiddenCounter = legalLedgerOverrideCounterSeed
	return nil
}

func (d *legalLedgerOverrideDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := moves.NewAutoConfigurer(d)
	return moves.Add(
		auto.MustConfig(
			new(legalLedgerMoveResidue),
			moves.WithMoveName("Residue"),
			moves.WithPreconditions(
				// Passes on every real game (FinishSetUp seeds 7 >= 5) --
				// the plan's verdict is Pass, so ONLY the imperative
				// residue makes this move illegal.
				legal.PropAtLeast("game.HiddenCounter", 5),
			),
		),
		auto.MustConfig(
			new(legalLedgerMoveEarlyNil),
			moves.WithMoveName("Early Nil"),
			moves.WithPreconditions(
				// Fails on every real game (7 < 1000) -- the plan's verdict
				// is Fail, so ONLY the override's early-return-nil makes
				// this move legal.
				legal.PropAtLeast("game.HiddenCounter", 1000),
			),
		),
	)
}

// legalLedgerMoveResidue opts in via WithPreconditions AND keeps a
// super-calling Legal() override with imperative residue -- the exact
// composition the design spec's prime-guarantee rule 4 blesses. The
// super-call runs first (so the boot reachability probe reaches
// moves.Default.Legal), then the residue rejects whenever
// game.HiddenCounter < 100 -- true on every real game this fixture builds
// (seeded to 7), so the move's ground-truth Legal() ALWAYS errors even
// though its plan verdict is Pass.
type legalLedgerMoveResidue struct {
	moves.CurrentPlayer
}

func (m *legalLedgerMoveResidue) Apply(state boardgame.State) error { return nil }

func (m *legalLedgerMoveResidue) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}
	gameState, ok := state.ImmutableGameState().(*legalLedgerGameState)
	if !ok {
		return errors.New("legal ledger override fixture: game state was not *legalLedgerGameState")
	}
	if gameState.HiddenCounter < 100 {
		return errors.New(legalLedgerResidueError)
	}
	return nil
}

// legalLedgerMoveEarlyNil opts in via WithPreconditions but its Legal()
// override conditionally early-returns nil BEFORE super-calling -- legal by
// ground truth while its plan verdict is Fail. The early return only
// triggers on a FinishSetUp-seeded (real-game) state; on the probe's
// ExampleState (HiddenCounter 0) it falls through to the super-call, so
// the boot reachability probe still reaches moves.Default.Legal and boot
// stays green.
type legalLedgerMoveEarlyNil struct {
	moves.CurrentPlayer
}

func (m *legalLedgerMoveEarlyNil) Apply(state boardgame.State) error { return nil }

func (m *legalLedgerMoveEarlyNil) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	gameState, ok := state.ImmutableGameState().(*legalLedgerGameState)
	if ok && gameState.HiddenCounter == legalLedgerOverrideCounterSeed {
		return nil
	}
	return m.CurrentPlayer.Legal(state, proposer)
}

// newLegalLedgerOverrideGame builds a fresh two-player game on
// legalLedgerOverrideDelegate.
func newLegalLedgerOverrideGame(t interface {
	Fatalf(format string, args ...interface{})
}) (*boardgame.Game, *boardgame.GameManager) {
	manager, err := boardgame.NewGameManager(&legalLedgerOverrideDelegate{}, newLegalLedgerStorage())
	if err != nil {
		t.Fatalf("legal ledger override fixture: building manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("legal ledger override fixture: building game: %v", err)
	}
	return game, manager
}

// TestGenerateFormsWithLegalityImperativeResidueRejects is the F1
// residue-direction regression: the plan verdict is Pass but the
// super-calling override's residue rejects, so ProposeMove would reject
// this move for EVERY proposer -- the served booleans must say so, with
// the residue's own error text, or the client renders an enabled button
// that can only ever bounce.
func TestGenerateFormsWithLegalityImperativeResidueRejects(t *testing.T) {
	game, _ := newLegalLedgerOverrideGame(t)
	s := &Server{}

	move := game.MoveByName("Residue")
	legalErr := move.Legal(game.CurrentState(), 0)
	if legalErr == nil {
		t.Fatal("test setup bug: the residue should reject on a freshly seeded game (HiddenCounter 7 < 100)")
	}
	if legalErr.Error() != legalLedgerResidueError {
		t.Fatalf("test setup bug: expected the residue's own error, got %q -- the plan or the CurrentPlayer imperative checks rejected first", legalErr.Error())
	}

	forms := s.generateFormsWithLegality(game, game.CurrentState(), 0)
	form := findMoveForm(forms, "Residue")
	if form == nil {
		t.Fatal("could not find Residue move form")
	}

	if form.LegalForPlayer {
		t.Error("LegalForPlayer = true, but move.Legal(state, 0) rejects (imperative residue) -- ProposeMove would bounce this move")
	}
	if form.LegalForPlayerError != legalErr.Error() {
		t.Errorf("LegalForPlayerError = %q, want the residue's ground-truth error %q", form.LegalForPlayerError, legalErr.Error())
	}
	if form.LegalForAnyone {
		t.Error("LegalForAnyone = true, but the residue rejects regardless of proposer -- move.Legal(state, AdminPlayerIndex) errors too")
	}
	assertLegalForAnyoneParity(t, move, game.CurrentState(), form)

	// The ledger stays advisory and plan-derived: the residue is invisible
	// to it by design, so its entries still reflect the (passing) plan.
	if len(form.Preconditions) == 0 {
		t.Error("opted-in move must still ship its advisory Preconditions ledger")
	}
	propAtLeast := findPrecondition(form.Preconditions, "propAtLeast")
	if propAtLeast == nil {
		t.Fatal("could not find propAtLeast ledger entry")
	}
	if propAtLeast.Verdict != "pass" {
		t.Errorf("propAtLeast ledger verdict = %q, want pass (the plan itself passes; only the residue rejects)", propAtLeast.Verdict)
	}
}

// TestGenerateFormsWithLegalityEarlyReturnNilOverride is the F1
// inverse-direction regression: the plan verdict is Fail but the override
// early-returns nil, so ProposeMove would ACCEPT this move -- the served
// booleans must say legal, or the client renders a disabled button for a
// perfectly legal move.
func TestGenerateFormsWithLegalityEarlyReturnNilOverride(t *testing.T) {
	game, _ := newLegalLedgerOverrideGame(t)
	s := &Server{}

	move := game.MoveByName("Early Nil")
	if err := move.Legal(game.CurrentState(), 0); err != nil {
		t.Fatalf("test setup bug: the override should early-return nil on a freshly seeded game, got %v", err)
	}

	forms := s.generateFormsWithLegality(game, game.CurrentState(), 0)
	form := findMoveForm(forms, "Early Nil")
	if form == nil {
		t.Fatal("could not find Early Nil move form")
	}

	if !form.LegalForPlayer {
		t.Errorf("LegalForPlayer = false (error %q), but move.Legal(state, 0) returns nil (early-return override) -- ProposeMove would accept this move", form.LegalForPlayerError)
	}
	if form.LegalForPlayerError != "" {
		t.Errorf("LegalForPlayerError = %q, want empty: the move is legal by ground truth", form.LegalForPlayerError)
	}
	if !form.LegalForAnyone {
		t.Error("LegalForAnyone = false, but move.Legal(state, AdminPlayerIndex) returns nil (early-return override)")
	}
	assertLegalForAnyoneParity(t, move, game.CurrentState(), form)

	// The ledger stays advisory and plan-derived: the plan's own view is
	// still a Fail, and that explanation detail must keep shipping.
	propAtLeast := findPrecondition(form.Preconditions, "propAtLeast")
	if propAtLeast == nil {
		t.Fatal("could not find propAtLeast ledger entry")
	}
	if propAtLeast.Verdict != "fail" {
		t.Errorf("propAtLeast ledger verdict = %q, want fail (the plan's advisory view is unchanged by the override)", propAtLeast.Verdict)
	}
}
