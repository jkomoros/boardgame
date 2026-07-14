package legal_test

import (
	"testing"

	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/legal"
)

/*
This file functionally tests the two framework wrapper predicates that live
in package legal (legal/catalog_framework.go): "inPhase" and
"stackConstraints". Both wrap a helper extracted to core
(boardgame.LegalInPhaseCheck / boardgame.LegalStackConstraintsCheck,
legal_framework.go) — see that file's doc comment and catalog_framework.go's
doc comment for the layering rationale. The third framework predicate,
"inProgression", lives in package moves instead (moves/catalog_framework.go)
and is functionally tested there (moves/inprogression_test.go), for the same
reason it isn't registered in this package's DefaultConstructors(): it needs
moves.MoveProgressionGroup, which package legal cannot import.
*/

// TestInPhase covers the "inPhase" predicate's pass/fail/unknown paths, its
// declared Read, and that its Fail detail binding is byte-identical to
// boardgame.LegalInPhaseCheck's error string. It uses examples/blackjack as
// its fixture: blackjack's phase enum (examples/blackjack/auto_enum.go) is a
// flat (non-tree) enum.Enum, so this only pins the flat pass/fail branches —
// TreeEnum ancestor-walking is already covered against boardgame's own
// LegalInPhaseCheck by that function's pre-existing callers/tests in package
// boardgame and moves/default.go's long-standing Legal() behavior; no
// in-repo example game's phase enum is itself a tree, so it isn't cheaply
// reachable from this external test package's fixtures.
func TestInPhase(t *testing.T) {
	// examples/blackjack/state.go: phaseGathering, phaseInitialDeal,
	// phaseNormalPlay, phaseRoundCleanup = iota (0, 1, 2, 3).
	// manager.NewDefaultGame() doesn't stop in phaseGathering: blackjack's
	// phaseGathering/phaseInitialDeal moves are FixUp (auto-applied), so a
	// freshly built game has already fast-forwarded to the first phase that
	// needs a real player decision — phaseNormalPlay (2) — by the time
	// NewDefaultGame returns (confirmed via CurrentPhase(state), not
	// assumed).
	const phaseGathering = enum.EnumKey(0)
	const phaseNormalPlay = enum.EnumKey(2)

	spec := legal.InPhase(phaseGathering, phaseNormalPlay)
	if spec.Name != "inPhase" {
		t.Fatalf("Name = %q, want inPhase", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 1 {
		t.Fatalf("Reads = %+v, want 1 entry", pred.Reads)
	}
	if pred.Reads[0].Path != "game.Phase" {
		t.Fatalf("Reads[0] = %+v, want game.Phase", pred.Reads[0])
	}

	_, state := newBlackjackGame(t)
	ctx := legal.Context{State: state, ProposerPlayerIndex: 0}

	// Pass: current phase (phaseNormalPlay) is in the list.
	if v := pred.Evaluate(ctx); v.Outcome != legal.Pass {
		t.Fatalf("phaseNormalPlay in [Gathering, NormalPlay]: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	// Fail: current phase (phaseNormalPlay) is NOT in a list that excludes
	// it.
	failPred := resolvePredicateForTest(t, legal.InPhase(phaseGathering))
	v := failPred.Evaluate(ctx)
	if v.Outcome != legal.Fail {
		t.Fatalf("phaseNormalPlay not in [Gathering]: Outcome = %v, want Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplateInPhase {
		t.Fatalf("Fail Message = %+v, want template %q", v.Message, legal.TemplateInPhase)
	}
	// Byte-identical to the legacy moves/default.go string (see
	// legal_framework.go's LegalInPhaseCheck doc comment).
	wantDetail := "Move is not legal in phase Normal Play"
	if got := v.Message.Bindings["detail"]; got.S == nil || *got.S != wantDetail {
		t.Fatalf("detail binding = %+v, want %q", got, wantDetail)
	}

	// Unknown: nil state.
	if v := pred.Evaluate(legal.Context{State: nil, ProposerPlayerIndex: 0}); v.Outcome != legal.Unknown {
		t.Fatalf("nil state: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}
}

// TestStackConstraints covers the "stackConstraints" predicate's
// pass/fail/unknown paths, its declared Reads, and that its Fail detail
// binding is byte-identical to boardgame.LegalStackConstraintsCheck's
// underlying component.MayMoveTo error string.
func TestStackConstraints(t *testing.T) {
	spec := legal.StackConstraints("HiddenCards", "VisibleCards")
	if spec.Name != "stackConstraints" {
		t.Fatalf("Name = %q, want stackConstraints", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 2 {
		t.Fatalf("Reads = %+v, want 2 entries", pred.Reads)
	}
	if pred.Reads[0].Path != "game.HiddenCards" || pred.Reads[1].Path != "game.VisibleCards" {
		t.Fatalf("Reads = %+v, want game.HiddenCards then game.VisibleCards", pred.Reads)
	}

	fixture := buildLegalFixture(t, "memoryDefault")

	// Pass: HiddenCards[0]'s component (memoryDefault has one) may move to
	// VisibleCards (a distinct, unconstrained stack).
	if v := pred.Evaluate(fixture.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("HiddenCards->VisibleCards: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	// Fail: source and destination are the same stack — MayMoveTo itself
	// rejects this (mirrors legal.MayMoveTo's own same-stack case in
	// catalog_test.go's TestMayMoveTo).
	sameStackPred := resolvePredicateForTest(t, legal.StackConstraints("HiddenCards", "HiddenCards"))
	v := sameStackPred.Evaluate(fixture.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("same-stack: Outcome = %v, want Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplateStackConstraints {
		t.Fatalf("same-stack Message = %+v, want template %q", v.Message, legal.TemplateStackConstraints)
	}
	if got := v.Message.Bindings["detail"]; got.S == nil || *got.S == "" {
		t.Fatalf("same-stack detail binding = %+v, want a non-empty error string", got)
	}

	// Unknown: nil state.
	if v := pred.Evaluate(legal.Context{State: nil, ProposerPlayerIndex: 0}); v.Outcome != legal.Unknown {
		t.Fatalf("nil state: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}
}
