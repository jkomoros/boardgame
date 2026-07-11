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

func TestGenerateFormsWithLegalityOptedInLedgerShape(t *testing.T) {
	game, _ := newLegalLedgerGame(t)
	s := &Server{}

	forms := s.generateFormsWithLegality(game, game.CurrentState(), 0)

	form := findMoveForm(forms, "Opted In")
	if form == nil {
		t.Fatal("could not find Opted In move form")
	}

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
