package api

import (
	"encoding/json"
	"testing"

	"github.com/jkomoros/boardgame"
)

/*
This file tests Task 10's server ledger: (*Server).generateFormsWithLegality
and the preconditionEntry/legalMessageJSON wire types (main.go), against the
legal_ledger_fixture_test.go fixture -- a manager with one opted-in move
("Opted In": moves.CurrentPlayer + an authored precondition reading a Hidden
game property) and one opaque move ("Opaque": moves.Default, no
preconditions). legal_ledger_frozen_wire_test.go covers the OPAQUE byte-
identity guarantee separately, against examples/memory.
*/

func findMoveForm(forms []*moveForm, name string) *moveForm {
	for _, f := range forms {
		if f.Name == name {
			return f
		}
	}
	return nil
}

func findPrecondition(entries []preconditionEntry, name string) *preconditionEntry {
	for i := range entries {
		if entries[i].Name == name {
			return &entries[i]
		}
	}
	return nil
}

// assertLegalForAnyoneParity asserts the universal invariant an opted-in
// move's LegalForAnyone must maintain: it is exactly
// (move.Legal(state, boardgame.AdminPlayerIndex) == nil) -- the plan-path
// LegalForAnyone (legalFormFromLedger, main.go) must agree with the
// move's own imperative Legal() call under an admin proposer, for ANY
// predicate mix. This is the invariant the exemption-based derivation this
// replaced got wrong: it assumed the ONLY place ctx.Proposer can matter is
// the contributed proposerIsCurrentPlayer atom's proposer-BYPASSED
// sub-check ("target != proposer"), and unconditionally exempted that
// entire entry -- silently also exempting its proposer-INDEPENDENT
// sub-checks (target invalid, target != current player), which Admin does
// NOT bypass. Asserting this everywhere a form is already built is cheap
// and catches that class of bug for any move mix, not just the specific
// no-current-player scenario TestGenerateFormsWithLegalityAdminDoesNotBypassProposerIndependentChecks
// exercises directly.
func assertLegalForAnyoneParity(t *testing.T, move boardgame.Move, state boardgame.ImmutableState, form *moveForm) {
	t.Helper()
	wantLegalForAnyone := move.Legal(state, boardgame.AdminPlayerIndex) == nil
	if form.LegalForAnyone != wantLegalForAnyone {
		t.Errorf("LegalForAnyone = %v, want %v (move.Legal(state, AdminPlayerIndex) == nil)", form.LegalForAnyone, wantLegalForAnyone)
	}
}

func TestBuildPreconditionEntryRequiresClientCatalogSupport(t *testing.T) {
	entry := boardgame.LegalVerdictEntry{
		Name:            "gameOnlyPredicate",
		Verdict:         boardgame.LegalVerdict{Outcome: boardgame.LegalPass},
		Serializable:    true,
		ClientEvaluable: false,
	}
	got := buildPreconditionEntry(nil, boardgame.AdminPlayerIndex, entry)
	if got.Evaluable {
		t.Fatal("serializable game predicate without client support was marked evaluable")
	}

	entry.ClientEvaluable = true
	got = buildPreconditionEntry(nil, boardgame.AdminPlayerIndex, entry)
	if !got.Evaluable {
		t.Fatal("client-known serializable predicate was not marked evaluable for admin")
	}
}

func TestGenerateFormsWithLegalityOptedInLedgerShape(t *testing.T) {
	game, _ := newLegalLedgerGame(t)
	s := &Server{}

	forms := s.generateFormsWithLegality(game, game.CurrentState(), 0)

	form := findMoveForm(forms, "Opted In")
	if form == nil {
		t.Fatal("could not find Opted In move form")
	}
	assertLegalForAnyoneParity(t, game.MoveByName("Opted In"), game.CurrentState(), form)

	if len(form.Preconditions) != 2 {
		t.Fatalf("expected 2 preconditions (authored propAtLeast + contributed proposerIsCurrentPlayer), got %d: %+v", len(form.Preconditions), form.Preconditions)
	}

	// The authored propAtLeast(game.HiddenCounter, 1000) always fails
	// (HiddenCounter starts at 0).
	propAtLeast := findPrecondition(form.Preconditions, "propAtLeast")
	if propAtLeast == nil {
		t.Fatal("could not find propAtLeast entry")
	}
	if propAtLeast.Verdict != "fail" {
		t.Errorf("propAtLeast verdict = %q, want fail", propAtLeast.Verdict)
	}
	if propAtLeast.Provisional {
		t.Error("propAtLeast is field-independent (reads game.*, not move.*) -- should not be provisional")
	}

	// The contributed proposerIsCurrentPlayer atom is field-dependent
	// (reads move.TargetPlayerIndex) -- provisional per the design spec §6
	// wire example.
	proposer := findPrecondition(form.Preconditions, "proposerIsCurrentPlayer")
	if proposer == nil {
		t.Fatal("could not find proposerIsCurrentPlayer entry")
	}
	if !proposer.Provisional {
		t.Error("proposerIsCurrentPlayer reads move.TargetPlayerIndex -- should be provisional")
	}
	// DefaultsForState binds TargetPlayerIndex to the actual current player,
	// and we asked for viewer/proposer 0, so this should pass as long as
	// player 0 is the current player at game start.
	if proposer.Verdict != "pass" {
		t.Errorf("proposerIsCurrentPlayer verdict = %q, want pass", proposer.Verdict)
	}
}

func TestGenerateFormsWithLegalityBindingsStrippedWhenInevaluable(t *testing.T) {
	game, _ := newLegalLedgerGame(t)
	s := &Server{}

	// Viewer 0 cannot see game.HiddenCounter (PolicyHidden for every
	// non-admin viewer, legal_ledger_fixture_test.go's SanitizationPolicy).
	forms := s.generateFormsWithLegality(game, game.CurrentState(), 0)
	form := findMoveForm(forms, "Opted In")
	assertLegalForAnyoneParity(t, game.MoveByName("Opted In"), game.CurrentState(), form)
	propAtLeast := findPrecondition(form.Preconditions, "propAtLeast")
	if propAtLeast == nil {
		t.Fatal("could not find propAtLeast entry")
	}

	if propAtLeast.Evaluable {
		t.Error("propAtLeast reads game.HiddenCounter (Hidden) -- should be evaluable:false for a non-admin viewer")
	}
	if propAtLeast.Message == nil {
		t.Fatal("propAtLeast failed -- should carry a Message")
	}
	if propAtLeast.Message.Template == "" {
		t.Error("#693 guard over-corrected: template key must still ship even when inevaluable")
	}
	if propAtLeast.Message.Bindings != nil {
		t.Errorf("#693 guard: bindings must be stripped when evaluable is false, got %+v", propAtLeast.Message.Bindings)
	}

	// Prove the #693 guard is enforced at the JSON boundary too, not just
	// in the in-memory struct (the review risk called out explicitly).
	data, err := json.Marshal(form)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	preconditions, ok := raw["Preconditions"].([]interface{})
	if !ok {
		t.Fatal("Preconditions was not a JSON array")
	}
	for _, p := range preconditions {
		entry, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		if entry["name"] != "propAtLeast" {
			continue
		}
		msg, ok := entry["message"].(map[string]interface{})
		if !ok {
			t.Fatal("propAtLeast JSON entry had no message object")
		}
		if _, hasBindings := msg["bindings"]; hasBindings {
			t.Errorf("propAtLeast JSON message carried bindings despite evaluable:false: %+v", msg)
		}
		if _, hasTemplate := msg["template"]; !hasTemplate {
			t.Error("propAtLeast JSON message is missing its template key")
		}
	}
}

func TestGenerateFormsWithLegalityAdminSeesEvaluableAndBindings(t *testing.T) {
	game, _ := newLegalLedgerGame(t)
	s := &Server{}

	// Admin is omniscient (boardgame.LegalReadEvaluable's own AdminPlayerIndex
	// bypass), so the SAME failing propAtLeast entry should be evaluable
	// with its bindings intact when the viewer is Admin.
	forms := s.generateFormsWithLegality(game, game.CurrentState(), boardgame.AdminPlayerIndex)
	form := findMoveForm(forms, "Opted In")
	assertLegalForAnyoneParity(t, game.MoveByName("Opted In"), game.CurrentState(), form)
	propAtLeast := findPrecondition(form.Preconditions, "propAtLeast")
	if propAtLeast == nil {
		t.Fatal("could not find propAtLeast entry")
	}

	if !propAtLeast.Evaluable {
		t.Error("Admin viewer should see propAtLeast as evaluable (omniscient)")
	}
	if propAtLeast.Message == nil || propAtLeast.Message.Bindings == nil {
		t.Error("Admin viewer should see propAtLeast's bindings")
	}
}

func TestGenerateFormsWithLegalityForPlayerAndForAnyone(t *testing.T) {
	game, _ := newLegalLedgerGame(t)
	s := &Server{}

	forms := s.generateFormsWithLegality(game, game.CurrentState(), 0)
	form := findMoveForm(forms, "Opted In")
	assertLegalForAnyoneParity(t, game.MoveByName("Opted In"), game.CurrentState(), form)

	// The authored precondition always fails, so the move can never be
	// legal for player 0, and LegalForPlayerError should be the RENDERED
	// text (byte-matching what move.Legal(state, 0) itself would produce).
	if form.LegalForPlayer {
		t.Error("LegalForPlayer should be false: the authored precondition always fails")
	}
	if form.LegalForPlayerError == "" {
		t.Error("LegalForPlayerError should be populated with the rendered failure")
	}

	legalErr := game.MoveByName("Opted In").Legal(game.CurrentState(), 0)
	if legalErr == nil {
		t.Fatal("expected move.Legal to also report a failure")
	}
	if form.LegalForPlayerError != legalErr.Error() {
		t.Errorf("LegalForPlayerError = %q, want byte-identical to move.Legal()'s error %q", form.LegalForPlayerError, legalErr.Error())
	}

	// LegalForAnyone is also false: the authored precondition (not the
	// proposer atom) is what's failing, and that's never exempted.
	if form.LegalForAnyone {
		t.Error("LegalForAnyone should be false: the authored (non-proposer) precondition fails regardless of who proposes")
	}
}

func TestGenerateFormsWithLegalityOpaqueMoveHasNoPreconditions(t *testing.T) {
	game, _ := newLegalLedgerGame(t)
	s := &Server{}

	forms := s.generateFormsWithLegality(game, game.CurrentState(), 0)
	form := findMoveForm(forms, "Opaque")
	if form == nil {
		t.Fatal("could not find Opaque move form")
	}
	if form.Preconditions != nil {
		t.Errorf("opaque move should have nil Preconditions, got %+v", form.Preconditions)
	}

	// And the JSON key itself must be absent (omitempty), not merely
	// null/empty -- this is what keeps an un-migrated move's moveForm JSON
	// byte-identical to the pre-Task-10 shape.
	data, err := json.Marshal(form)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if _, ok := raw["Preconditions"]; ok {
		t.Error("opaque move's JSON must omit the Preconditions key entirely")
	}
}

// TestGenerateFormsWithLegalityAdminDoesNotBypassProposerIndependentChecks is
// the reviewer's exact regression scenario for the Critical finding on
// legalForAnyoneFromLedger (now removed): that function derived
// LegalForAnyone by unconditionally exempting the whole
// proposerIsCurrentPlayer ledger entry, on the false premise that it would
// always pass under proposer=Admin. In fact that predicate bundles three
// checks (legal/catalog_players.go): (1) target invalid and (2) target !=
// current player are proposer-INDEPENDENT -- Admin does not bypass them;
// only (3) target != proposer is bypassed (via PlayerIndex.Equivalent's
// Admin-is-wildcard rule). legalLedgerObserverDelegate (fixture file) forces
// game.CurrentPlayer = boardgame.ObserverPlayerIndex, a documented framework
// pattern (base.GameDelegate.CurrentPlayerIndex's own doc comment) for "no
// one may move this round" -- and pre-seeds HiddenCounter past the authored
// propAtLeast threshold, so proposerIsCurrentPlayer is the ONLY precondition
// that can fail. moves.CurrentPlayer.DefaultsForState mirrors
// TargetPlayerIndex to the (Observer) current player, so the atom fails at
// its very first check ("target invalid" -- ObserverPlayerIndex < 0) --
// proposer-independent, so Admin does NOT bypass it, and LegalForAnyone must
// be false.
//
// Against the removed exemption-based code this test is RED: entries holds
// only the (exempted) proposerIsCurrentPlayer entry, so
// legalForAnyoneFromLedger's loop finds nothing else to check and returns
// true, even though move.Legal(state, AdminPlayerIndex) itself returns a
// non-nil error. Against the fix (GameManager.LegalEvaluatePlan re-run at
// proposer=AdminPlayerIndex) it is GREEN.
func TestGenerateFormsWithLegalityAdminDoesNotBypassProposerIndependentChecks(t *testing.T) {
	game, _ := newLegalLedgerObserverGame(t)
	s := &Server{}

	forms := s.generateFormsWithLegality(game, game.CurrentState(), 0)
	form := findMoveForm(forms, "Opted In")
	if form == nil {
		t.Fatal("could not find Opted In move form")
	}

	move := game.MoveByName("Opted In")
	adminErr := move.Legal(game.CurrentState(), boardgame.AdminPlayerIndex)
	if adminErr == nil {
		t.Fatal("test setup bug: move.Legal(state, AdminPlayerIndex) should fail when there is no current player -- Admin does not bypass the proposer-independent 'target invalid'/'not your turn' checks")
	}

	if form.LegalForAnyone {
		t.Error("LegalForAnyone should be false: no current player means proposerIsCurrentPlayer's proposer-INDEPENDENT sub-checks fail, and Admin does not bypass those")
	}

	assertLegalForAnyoneParity(t, move, game.CurrentState(), form)
}
