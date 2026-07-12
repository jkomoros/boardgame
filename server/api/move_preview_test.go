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
