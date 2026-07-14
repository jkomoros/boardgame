package api

import "testing"

// TestLegalLedgerFixtureBoots is a smoke test for legal_ledger_fixture_test.go:
// proves the hand-rolled delegate/readers/storage actually boot a game and
// that the "Opted In" move type really did opt in to declarative legality
// (LegalEvaluateLedger reports opted==true), before the real ledger-shape
// tests build on top of it.
func TestLegalLedgerFixtureBoots(t *testing.T) {
	game, manager := newLegalLedgerGame(t)

	move := game.MoveByName("Opted In")
	if move == nil {
		t.Fatal("could not find move Opted In")
	}

	_, _, opted := manager.LegalEvaluateLedger("Opted In", game.CurrentState(), move, 0)
	if !opted {
		t.Fatal("Opted In move did not opt in to declarative legality")
	}

	opaque := game.MoveByName("Opaque")
	if opaque == nil {
		t.Fatal("could not find move Opaque")
	}
	_, _, opaqueOpted := manager.LegalEvaluateLedger("Opaque", game.CurrentState(), opaque, 0)
	if opaqueOpted {
		t.Fatal("Opaque move should not have opted in")
	}
}
