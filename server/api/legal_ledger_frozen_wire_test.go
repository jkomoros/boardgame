package api

import (
	"encoding/json"
	"testing"

	"github.com/jkomoros/boardgame"
	werewolfgame "github.com/jkomoros/boardgame/examples/werewolf"
)

/*
This file is Task 10's frozen-wire test (spec §6's "prime guarantee",
extended to the wire format): the fixture game used here has not opted in to
declarative legality, so EVERY one of its moves takes the opaque
(legalFormOpaque) path through generateFormsWithLegality, and that path's
output must be byte-identical to what the pre-Task-10 server produced.

Fixture choice (Task 11 update): this originally used examples/memory, which
was accurate when written (Task 10, before any game had migrated). Task 11
migrated memory's "Reveal Card" move to declarative legality (design spec
§8), so memory is no longer a fully-opaque game and this test's own "None of
memory's moves have opted in yet" assumption below would be false against
it. Task 11's brief calls this out explicitly ("handle deliberately and
document"): rather than weaken this test to special-case one opted-in move,
the fixture was swapped to examples/checkers, which stayed fully un-migrated
through Task 11 and so kept this test's original all-opaque guarantee intact
and meaningful. (memory's own opted-in move now has its OWN golden-
equivalence coverage instead — examples/memory/legal_golden_test.go.)

Fixture choice (Task 12 update): Task 12 migrated checkers' "Move Token"
move (moveMoveToken) to declarative legality (design spec §8's checkers acid
test — see examples/checkers/moves.go/main.go), so checkers is no longer a
fully-opaque game either, for the identical reason memory stopped being one
in Task 11. Per the Task 12 brief's explicit instruction ("re-point it to a
game that REMAINS fully opaque after this task... document the choice"),
the fixture is swapped again, to examples/werewolf, which Task 12's survey
found gains ZERO opted-in moves, not because no natural catalog gate exists,
but because every one of werewolf's three own move types embeds a
moves-package framework base type OTHER than Default/CurrentPlayer — the
v1 declarative-legality composition seam (design spec §2) only supports
those two:
  - moveBeginGame embeds moves.StartPhase
  - moveCastVote embeds moves.AnyPlayer
  - moveResolveVotes embeds moves.FixUp
(plus moves.SeatPlayer/ActivateInactivePlayer/WaitForEnoughPlayers/
InactivateEmptySeat from main.go's ConfigureMoves, which are unmodified
framework move types werewolf never overrides Legal() on, and which
legalUnsupportedMovesBaseType would flag too — moot here since none of them
declare WithLegalPreconditions in the first place). Every one of legalUnsupported-
MovesBaseType's checks is a hard boot-time gate (legal_plan.go), not a
judgment call, so this is a durable "stays opaque" guarantee, not a
survey-cycle-scoped one — unlike checkers/memory, no genuinely-natural
catalog gate was left un-migrated here; werewolf's moves.go documents the
same finding (Task 12 report has the full per-game migration table).

There is no literal "before" JSON blob checked in here to diff against
(this package had no test harness capable of building a real game before
this task), so byte-identity is proven differentially instead, which is
actually the stronger guarantee: legalFormOpaqueExpected below
re-implements the OLD two-move.Legal()-call logic independently, at test
time, against the SAME live game/state generateFormsWithLegality itself
uses. If legalFormOpaque (main.go) ever drifted from that logic, this test
would catch it even without a pre-recorded fixture.
*/

// legalFormOpaqueExpected independently recomputes what the pre-Task-10
// generateFormsWithLegality would have produced for one move, by calling
// move.Legal() exactly the way the old code did (two calls: playerIndex,
// then AdminPlayerIndex). It deliberately does NOT call any Task-10
// production code (legalFormOpaque/legalFormFromLedger/buildPreconditionEntry) --
// it is the independent oracle those functions are checked against.
func legalFormOpaqueExpected(move boardgame.Move, state boardgame.ImmutableState, playerIndex boardgame.PlayerIndex) *moveForm {
	item := &moveForm{}
	if playerIndex != boardgame.ObserverPlayerIndex {
		if err := move.Legal(state, playerIndex); err != nil {
			item.LegalForPlayerError = err.Error()
		} else {
			item.LegalForPlayer = true
		}
	}
	if err := move.Legal(state, boardgame.AdminPlayerIndex); err == nil {
		item.LegalForAnyone = true
	}
	return item
}

func TestGenerateFormsWithLegalityOpaqueGameByteIdentical(t *testing.T) {
	manager, err := boardgame.NewGameManager(werewolfgame.NewDelegate(), newLegalLedgerStorage())
	if err != nil {
		t.Fatalf("building werewolf manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("building werewolf game: %v", err)
	}

	s := &Server{}
	state := game.CurrentState()

	for _, playerIndex := range []boardgame.PlayerIndex{0, 1, boardgame.ObserverPlayerIndex} {
		forms := s.generateFormsWithLegality(game, state, playerIndex)
		if len(forms) == 0 {
			t.Fatalf("expected at least one move form for player %d", playerIndex)
		}

		for _, form := range forms {
			// None of werewolf's moves have opted in, and structurally never
			// will under the v1 seam (see this file's doc comment) -- every
			// single one must take the frozen opaque path: no Preconditions
			// at all.
			if form.Preconditions != nil {
				t.Errorf("player %d, move %q: opaque game produced a non-nil Preconditions ledger: %+v", playerIndex, form.Name, form.Preconditions)
			}

			move := game.MoveByName(form.Name)
			if move == nil {
				t.Fatalf("player %d: could not re-look-up move %q", playerIndex, form.Name)
			}
			expected := legalFormOpaqueExpected(move, state, playerIndex)

			if form.LegalForPlayer != expected.LegalForPlayer {
				t.Errorf("player %d, move %q: LegalForPlayer = %v, want %v", playerIndex, form.Name, form.LegalForPlayer, expected.LegalForPlayer)
			}
			if form.LegalForPlayerError != expected.LegalForPlayerError {
				t.Errorf("player %d, move %q: LegalForPlayerError = %q, want %q", playerIndex, form.Name, form.LegalForPlayerError, expected.LegalForPlayerError)
			}
			if form.LegalForAnyone != expected.LegalForAnyone {
				t.Errorf("player %d, move %q: LegalForAnyone = %v, want %v", playerIndex, form.Name, form.LegalForAnyone, expected.LegalForAnyone)
			}
		}

		// The JSON wire form itself must never carry a "Preconditions" key
		// for any of these moves (omitempty on a nil slice) -- the actual
		// byte-identity guarantee, checked at the JSON boundary rather than
		// just the Go struct.
		data, err := json.Marshal(forms)
		if err != nil {
			t.Fatalf("player %d: marshal error: %v", playerIndex, err)
		}
		var raw []map[string]interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("player %d: unmarshal error: %v", playerIndex, err)
		}
		for _, entry := range raw {
			if _, ok := entry["Preconditions"]; ok {
				t.Errorf("player %d: JSON move form %v carries a Preconditions key on an opaque game", playerIndex, entry["Name"])
			}
		}
	}
}
