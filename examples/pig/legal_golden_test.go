package pig

import (
	"errors"
	"testing"

	"github.com/jkomoros/boardgame"
	storagememory "github.com/jkomoros/boardgame/storage/memory"
)

/*
Golden-equivalence harness for moveRollDice and moveDoneTurn (design spec
§8/§9, Task 12 brief). Follows examples/memory/legal_golden_test.go's
pattern (Task 11 precedent): a hand-copied legacy oracle per move type,
crossed against every proposer worth distinguishing, checked against the
migrated move's ACTUAL Legal().

moveCountDie is NOT covered here: it stays fully opaque (its only gate is a
negated boolean with no catalog primitive — see moves.go's doc comment), so
its Legal() is byte-for-byte unchanged and there is nothing to diff.
*/

func newPigGame(t *testing.T) (*boardgame.Game, boardgame.State) {
	t.Helper()
	manager, err := boardgame.NewGameManager(NewDelegate(), storagememory.NewStorageManager())
	if err != nil {
		t.Fatalf("legal_golden: building manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("legal_golden: building default game: %v", err)
	}
	state, ok := game.CurrentState().(boardgame.State)
	if !ok {
		t.Fatalf("legal_golden: CurrentState() was not mutable")
	}
	return game, state
}

func pigProposers(t *testing.T, state boardgame.State) map[string]boardgame.PlayerIndex {
	t.Helper()
	currentPlayer := state.CurrentPlayerIndex()
	var otherPlayer boardgame.PlayerIndex = -1
	for i := range state.ImmutablePlayerStates() {
		pIdx := boardgame.PlayerIndex(i)
		if pIdx != currentPlayer {
			otherPlayer = pIdx
			break
		}
	}
	if otherPlayer < 0 {
		t.Fatal("legal_golden: could not find a non-current player")
	}
	return map[string]boardgame.PlayerIndex{
		"currentPlayer": currentPlayer,
		"otherPlayer":   otherPlayer,
		"admin":         boardgame.AdminPlayerIndex,
		"observer":      boardgame.ObserverPlayerIndex,
	}
}

/**************************************************
 *
 * moveRollDice golden coverage
 *
 **************************************************/

// legacyLegalMoveRollDice is a hand-copied snapshot of moveRollDice's
// Legal() method exactly as it read before this migration (see moves.go's
// comment block for the original source), BUG INCLUDED: the pre-migration
// body discarded the proposer-check's error ("return nil" instead of
// "return err"), so that check's outcome could never affect the result —
// only the DieCounted gate below mattered. This oracle replicates that
// exactly (the proposer check is intentionally not computed at all: its
// result was unreachable dead code in the original, and computing it here
// without using it would misleadingly imply it mattered).
func legacyLegalMoveRollDice(m *moveRollDice, state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	game, players := concreteStates(state)
	p := players[game.CurrentPlayer.EnsureValid(state)]
	if !p.DieCounted {
		return errors.New("Your most recent roll has not yet been counted")
	}
	return nil
}

// knownBugFixNilnessDivergence names (fixture, proposer) combinations where
// the migrated plan is EXPECTED to disagree with the legacy oracle on
// NIL-NESS itself (not just message text): moveRollDice's legacy Legal()
// had a bug (see legacyLegalMoveRollDice's doc comment) that made it always
// report "legal" for a non-current-player proposer, regardless of the
// actual proposer-check outcome. The migrated plan's contributed proposer
// atom (legal.ProposerIsCurrentPlayer) has no such bug, so it correctly
// rejects a wrong proposer even when DieCounted is true (the "default"
// fixture). This is a deliberate, documented behavior IMPROVEMENT — see
// moves.go's doc comment and the Task 12 report — not a regression.
var knownBugFixNilnessDivergence = map[string]bool{
	"default/otherPlayer": true,
	"default/observer":    true,
}

func TestGoldenLegalMoveRollDice(t *testing.T) {
	type fixture struct {
		name string
		game *boardgame.Game
	}

	var fixtures []fixture

	// default: a fresh game (DieCounted is true for every player at turn
	// start — ResetForTurnStart). Legal for the current player.
	{
		game, _ := newPigGame(t)
		fixtures = append(fixtures, fixture{"default", game})
	}

	// dieNotCounted: the current player's DieCounted is forced to false —
	// legal.PlayerBool("DieCounted")'s Fail branch ("pig.roll_not_counted").
	{
		game, state := newPigGame(t)
		rs := state.CurrentPlayer().ReadSetter()
		if err := rs.SetBoolProp("DieCounted", false); err != nil {
			t.Fatalf("legal_golden: setting DieCounted: %v", err)
		}
		fixtures = append(fixtures, fixture{"dieNotCounted", game})
	}

	for _, fx := range fixtures {
		fx := fx
		state := fx.game.CurrentState()
		move := fx.game.MoveByName("Roll Dice")
		if move == nil {
			t.Fatal("legal_golden: no \"Roll Dice\" move found")
		}
		mv, ok := move.(*moveRollDice)
		if !ok {
			t.Fatal("legal_golden: \"Roll Dice\" move was not a *moveRollDice")
		}

		for proposerName, proposer := range pigProposers(t, state.(boardgame.State)) {
			t.Run(fx.name+"/"+proposerName, func(t *testing.T) {
				legacyErr := legacyLegalMoveRollDice(mv, state, proposer)
				actualErr := mv.Legal(state, proposer)

				if knownBugFixNilnessDivergence[fx.name+"/"+proposerName] {
					if legacyErr != nil {
						t.Fatalf("expected legacy (buggy) oracle to report legal, got %v", legacyErr)
					}
					if actualErr == nil {
						t.Fatalf("expected migrated Legal() to correctly reject a non-current-player proposer, got nil")
					}
					return
				}

				if (legacyErr == nil) != (actualErr == nil) {
					t.Fatalf("nil-ness mismatch: legacy=%v actual=%v", legacyErr, actualErr)
				}
				if legacyErr != nil && legacyErr.Error() != actualErr.Error() {
					t.Fatalf("message mismatch:\n legacy: %q\n actual: %q", legacyErr.Error(), actualErr.Error())
				}
			})
		}
	}
}

/**************************************************
 *
 * moveDoneTurn golden coverage
 *
 **************************************************/

// legacyLegalMoveDoneTurn is a hand-copied snapshot of moveDoneTurn's
// Legal() method exactly as it read before this migration (see moves.go's
// comment block for the original source): unlike moveRollDice, this one has
// no bug (it correctly returns the super-call's error).
func legacyLegalMoveDoneTurn(m *moveDoneTurn, state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	currentPlayer := state.CurrentPlayerIndex()
	targetPlayerIndex := m.TargetPlayerIndex.EnsureValid(state)

	if !targetPlayerIndex.Valid(state) {
		return errors.New("The specified target player is not valid")
	}
	if targetPlayerIndex < 0 {
		return errors.New("The specified target player is not valid")
	}
	if !targetPlayerIndex.Equivalent(currentPlayer) {
		return errors.New("it's not your turn")
	}
	if !targetPlayerIndex.Equivalent(proposer) {
		return errors.New("it's not your turn")
	}

	game, players := concreteStates(state)
	p := players[game.CurrentPlayer.EnsureValid(state)]

	if !p.DieCounted {
		return errors.New("your most recent roll has not yet been counted")
	}

	if p.Done {
		return errors.New("you already signaled that you are done")
	}

	return nil
}

// knownMessageOrderingDivergence names (fixture, proposer) combinations
// where the migrated plan is EXPECTED to disagree with the legacy oracle on
// WHICH message wins (nil-ness always matches), for the SAME
// bucket-reordering reason documented in
// examples/memory/legal_golden_test.go (Task 11): legal.PlayerBool
// ("DieCounted") reads no move.* path, so it lands in moveDoneTurn's
// field-INDEPENDENT bucket and evaluates before the contributed proposer
// check (field-dependent), reversing their legacy relative order. Only
// "dieNotCounted" combined with a failing proposer check exercises this
// (the "default" and "alreadyDone" fixtures have DieCounted true, so the
// proposer check — when it fails — always wins in both legacy and migrated
// order, matching trivially).
var knownMessageOrderingDivergence = map[string]bool{
	"dieNotCounted/otherPlayer": true,
	"dieNotCounted/observer":    true,
}

func TestGoldenLegalMoveDoneTurn(t *testing.T) {
	type fixture struct {
		name string
		game *boardgame.Game
	}

	var fixtures []fixture

	// default: a fresh game (DieCounted true, Done false). Legal for the
	// current player.
	{
		game, _ := newPigGame(t)
		fixtures = append(fixtures, fixture{"default", game})
	}

	// dieNotCounted: DieCounted forced to false —
	// legal.PlayerBool("DieCounted")'s Fail branch ("pig.done_roll_not_counted").
	{
		game, state := newPigGame(t)
		rs := state.CurrentPlayer().ReadSetter()
		if err := rs.SetBoolProp("DieCounted", false); err != nil {
			t.Fatalf("legal_golden: setting DieCounted: %v", err)
		}
		fixtures = append(fixtures, fixture{"dieNotCounted", game})
	}

	// alreadyDone: Done forced to true (DieCounted stays true) —
	// LegalCustom's residue Fail branch ("pig.already_done").
	{
		game, state := newPigGame(t)
		rs := state.CurrentPlayer().ReadSetter()
		if err := rs.SetBoolProp("Done", true); err != nil {
			t.Fatalf("legal_golden: setting Done: %v", err)
		}
		fixtures = append(fixtures, fixture{"alreadyDone", game})
	}

	for _, fx := range fixtures {
		fx := fx
		state := fx.game.CurrentState()
		move := fx.game.MoveByName("Done Turn")
		if move == nil {
			t.Fatal("legal_golden: no \"Done Turn\" move found")
		}
		mv, ok := move.(*moveDoneTurn)
		if !ok {
			t.Fatal("legal_golden: \"Done Turn\" move was not a *moveDoneTurn")
		}

		for proposerName, proposer := range pigProposers(t, state.(boardgame.State)) {
			t.Run(fx.name+"/"+proposerName, func(t *testing.T) {
				legacyErr := legacyLegalMoveDoneTurn(mv, state, proposer)
				actualErr := mv.Legal(state, proposer)

				if (legacyErr == nil) != (actualErr == nil) {
					t.Fatalf("nil-ness mismatch: legacy=%v actual=%v", legacyErr, actualErr)
				}
				if knownMessageOrderingDivergence[fx.name+"/"+proposerName] {
					return
				}
				if legacyErr != nil && legacyErr.Error() != actualErr.Error() {
					t.Fatalf("message mismatch:\n legacy: %q\n actual: %q", legacyErr.Error(), actualErr.Error())
				}
			})
		}
	}
}
