package api

import (
	"testing"

	"github.com/jkomoros/boardgame"
)

// TestLegalMoveFormPreviewMatchesMoveLegalAndDoesNotApply pins the two core
// promises of the movePreview endpoint's legality builder (legalMoveForm):
// (1) it reports exactly the authoritative move.Legal() verdict + error the
// real ProposeMove gate would, and (2) it is side-effect-free — previewing must
// never advance the game, so a client can call it on every keystroke.
func TestLegalMoveFormPreviewMatchesMoveLegalAndDoesNotApply(t *testing.T) {
	game, _ := newLegalLedgerGame(t)
	s := &Server{}

	move := game.MoveByName("Opted In")
	if move == nil {
		t.Fatal("could not find the Opted In move")
	}
	state := game.CurrentState()
	versionBefore := game.Version()

	form := s.legalMoveForm(game, state, move, 0)

	// (1) legality parity with the authoritative move.Legal (the ground truth
	// game.ProposeMove itself gates on).
	legalErr := move.Legal(state, 0)
	wantLegal := legalErr == nil
	if form.LegalForPlayer != wantLegal {
		t.Errorf("preview LegalForPlayer = %v, want %v (move.Legal(state, 0) == nil)", form.LegalForPlayer, wantLegal)
	}
	if !wantLegal {
		if form.LegalForPlayerError != legalErr.Error() {
			t.Errorf("preview LegalForPlayerError = %q, want %q (verbatim move.Legal error)", form.LegalForPlayerError, legalErr.Error())
		}
	}

	// LegalForAnyone parity with move.Legal(state, AdminPlayerIndex).
	wantAnyone := move.Legal(state, boardgame.AdminPlayerIndex) == nil
	if form.LegalForAnyone != wantAnyone {
		t.Errorf("preview LegalForAnyone = %v, want %v", form.LegalForAnyone, wantAnyone)
	}

	// (2) side-effect-free: previewing must not advance the game version.
	if got := game.Version(); got != versionBefore {
		t.Errorf("preview advanced the game version %d -> %d; the preview path must never apply a move", versionBefore, got)
	}
}

// TestLegalMoveFormsBatch pins the batch preview core (legalMoveFormsBatch),
// which computes legality for many candidate arg-sets of ONE move type against
// a single state in one call — the primitive that lets a client gray a whole
// board's candidate targets in a single round-trip. The promises: each result
// matches the authoritative move.Legal() for that candidate's arg-bound move,
// results stay in candidate order (so the client can correlate by index), a
// malformed candidate soft-fails to Legal:false without killing the batch or
// leaking into its neighbors, an invalid move type is a whole-batch error, and
// the whole thing never advances the game (safe on every keystroke).
func TestLegalMoveFormsBatch(t *testing.T) {
	game, _ := newLegalLedgerGame(t)
	s := &Server{}
	state := game.CurrentState()
	proposer := boardgame.PlayerIndex(0)

	// expectedFor is the independent oracle: bind a fresh move the same way the
	// batch does, then ask the authoritative move.Legal() (the exact call
	// ProposeMove gates on) for the verdict + verbatim error.
	expectedFor := func(t *testing.T, moveType string, args map[string]string) (bool, string) {
		t.Helper()
		m := game.MoveByName(moveType)
		if m == nil {
			t.Fatalf("could not find move %q", moveType)
		}
		if err := bindMoveFields(m, func(n string) string { return args[n] }); err != nil {
			t.Fatalf("independent bind of %q failed: %v", moveType, err)
		}
		legalErr := m.Legal(state, proposer)
		if legalErr != nil {
			return false, legalErr.Error()
		}
		return true, ""
	}

	t.Run("parity and order across varying args", func(t *testing.T) {
		// "Opted In" embeds moves.CurrentPlayer, so it has a TargetPlayerIndex
		// field to vary per candidate.
		candidates := []movePreviewBatchCandidate{
			{Args: map[string]string{"TargetPlayerIndex": "0"}},
			{Args: map[string]string{"TargetPlayerIndex": "1"}},
			{Args: map[string]string{"TargetPlayerIndex": "0"}},
		}
		versionBefore := game.Version()

		results, err := s.legalMoveFormsBatch(game, state, "Opted In", candidates, proposer)
		if err != nil {
			t.Fatalf("unexpected whole-batch error: %v", err)
		}
		if len(results) != len(candidates) {
			t.Fatalf("got %d results, want %d (exactly one per candidate, in order)", len(results), len(candidates))
		}
		for i, cand := range candidates {
			wantLegal, wantErr := expectedFor(t, "Opted In", cand.Args)
			if results[i].Legal != wantLegal {
				t.Errorf("candidate %d: Legal = %v, want %v (move.Legal parity)", i, results[i].Legal, wantLegal)
			}
			if results[i].Error != wantErr {
				t.Errorf("candidate %d: Error = %q, want %q (verbatim move.Legal error)", i, results[i].Error, wantErr)
			}
		}

		// side-effect-free.
		if got := game.Version(); got != versionBefore {
			t.Errorf("batch preview advanced game version %d -> %d; must never apply", versionBefore, got)
		}
	})

	t.Run("legal candidate (fieldless Opaque move)", func(t *testing.T) {
		// "Opaque" is moves.Default with no fields and no failing precondition
		// — it exercises the Legal:true branch.
		results, err := s.legalMoveFormsBatch(game, state, "Opaque", []movePreviewBatchCandidate{{Args: map[string]string{}}}, proposer)
		if err != nil {
			t.Fatalf("unexpected whole-batch error: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("got %d results, want 1", len(results))
		}
		wantLegal, wantErr := expectedFor(t, "Opaque", map[string]string{})
		if !wantLegal {
			t.Fatalf("test precondition: expected Opaque to be legal for proposer 0, but move.Legal said %q", wantErr)
		}
		if !results[0].Legal || results[0].Error != "" {
			t.Errorf("Opaque result = {Legal:%v Error:%q}, want {Legal:true Error:\"\"}", results[0].Legal, results[0].Error)
		}
	})

	t.Run("malformed candidate soft-fails without killing the batch", func(t *testing.T) {
		candidates := []movePreviewBatchCandidate{
			{Args: map[string]string{"TargetPlayerIndex": "0"}},           // fine
			{Args: map[string]string{"TargetPlayerIndex": "not-an-int"}},  // unbindable
			{Args: map[string]string{"TargetPlayerIndex": "1"}},           // fine, must still be evaluated
		}
		results, err := s.legalMoveFormsBatch(game, state, "Opted In", candidates, proposer)
		if err != nil {
			t.Fatalf("a malformed candidate must not fail the whole batch, got error: %v", err)
		}
		if len(results) != 3 {
			t.Fatalf("got %d results, want 3 (malformed candidate still occupies its slot)", len(results))
		}
		if results[1].Legal || results[1].Error == "" {
			t.Errorf("malformed candidate result = {Legal:%v Error:%q}, want Legal:false with a non-empty bind error", results[1].Legal, results[1].Error)
		}
		// Neighbors match their own oracle (the bad candidate didn't leak).
		for _, i := range []int{0, 2} {
			wantLegal, wantErr := expectedFor(t, "Opted In", candidates[i].Args)
			if results[i].Legal != wantLegal || results[i].Error != wantErr {
				t.Errorf("candidate %d contaminated by malformed neighbor: got {Legal:%v Error:%q}, want {Legal:%v Error:%q}", i, results[i].Legal, results[i].Error, wantLegal, wantErr)
			}
		}
	})

	t.Run("invalid move type is a whole-batch error", func(t *testing.T) {
		results, err := s.legalMoveFormsBatch(game, state, "No Such Move", []movePreviewBatchCandidate{{Args: map[string]string{}}}, proposer)
		if err == nil {
			t.Fatalf("expected a whole-batch error for an invalid move type, got results %#v", results)
		}
	})
}
