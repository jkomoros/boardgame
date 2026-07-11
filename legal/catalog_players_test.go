package legal

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
)

// TestAny covers the "any" builder's shape only: per Any's doc comment,
// builders stay dumb — the >= 2 subs requirement, depth-1 enforcement, and
// Kleene evaluation semantics are all owned and exhaustively tested by core
// (boardgame's legal_predicate_test.go: TestResolveLegalSpecsAnyRequiresTwoSubs,
// TestResolveLegalSpecsAnyDepthTwoRejected, TestLegalAnyKleeneTruthTable*).
func TestAny(t *testing.T) {
	sub1 := PlayerBool("Eliminated")
	sub2 := PlayerBool("Stood")
	spec := Any(sub1, sub2)
	if spec.Name != "any" {
		t.Fatalf("Name = %q, want any", spec.Name)
	}
	if len(spec.Sub) != 2 || spec.Sub[0].Name != "playerBool" || spec.Sub[1].Name != "playerBool" {
		t.Fatalf("Sub = %+v", spec.Sub)
	}

	// "any" is intentionally NOT a registered PredicateConstructor: core
	// intercepts the name directly (legalAnyCompositorName in
	// legal_predicate.go) so it can never be shadowed by a registry entry.
	for _, c := range DefaultConstructors() {
		if c.Name == "any" {
			t.Fatal(`DefaultConstructors() must not register a constructor named "any" — core intercepts it directly`)
		}
	}
}

// TestAllActivePlayers covers AllActivePlayers' shape, Reads (the
// "players[*].X" quantifier-only paths, one per inner leaf), and
// pass/fail/unknown against the blackjack fixtures — the design spec §8
// acid test (AllActivePlayers(Any(PlayerBool("Eliminated"),
// PlayerBool("Stood")))).
func TestAllActivePlayers(t *testing.T) {
	spec := AllActivePlayers(Any(PlayerBool("Eliminated"), PlayerBool("Stood")))
	if spec.Name != "allActivePlayers" {
		t.Fatalf("Name = %q, want allActivePlayers", spec.Name)
	}
	if len(spec.Sub) != 1 || spec.Sub[0].Name != "any" {
		t.Fatalf("Sub = %+v", spec.Sub)
	}

	pred := resolvePredicateForTest(t, spec)
	wantReads := map[string]bool{
		"players[*].Eliminated": true,
		"players[*].Stood":      true,
	}
	if len(pred.Reads) != len(wantReads) {
		t.Fatalf("Reads = %+v, want %d entries", pred.Reads, len(wantReads))
	}
	for _, r := range pred.Reads {
		if !wantReads[string(r.Path)] {
			t.Errorf("unexpected Read %+v", r)
		}
		if r.Facet != boardgame.LegalFacetValues {
			t.Errorf("Read %+v: Facet = %v, want LegalFacetValues", r, r.Facet)
		}
	}

	// Pass: every active player has Stood (blackjackAllFinished).
	allFinished := buildLegalFixture(t, "blackjackAllFinished")
	if v := pred.Evaluate(allFinished.context(0)); v.Outcome != Pass {
		t.Fatalf("blackjackAllFinished: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	// Fail: player 0 has neither Eliminated nor Stood.
	oneUnfinished := buildLegalFixture(t, "blackjackOneUnfinished")
	v := pred.Evaluate(oneUnfinished.context(0))
	if v.Outcome != Fail {
		t.Fatalf("blackjackOneUnfinished: Outcome = %v, want Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != TemplateAllActivePlayers {
		t.Fatalf("blackjackOneUnfinished: Message = %+v, want template %q", v.Message, TemplateAllActivePlayers)
	}

	// Pass: player 0 is the same as blackjackOneUnfinished (neither
	// condition true), but it's also PlayerInactive, so it's skipped
	// entirely — every ACTIVE player has Stood.
	inactiveSkipped := buildLegalFixture(t, "blackjackInactiveSkipped")
	if v := pred.Evaluate(inactiveSkipped.context(0)); v.Outcome != Pass {
		t.Fatalf("blackjackInactiveSkipped: Outcome = %v, want Pass (%+v) — an inactive player's unfinished state must not fail the quantifier", v.Outcome, v)
	}

	// Unknown: inner references a bool prop that doesn't exist on
	// blackjack's playerState.
	unknownPred := resolvePredicateForTest(t, AllActivePlayers(PlayerBool("NoSuchBoolProp")))
	if v := unknownPred.Evaluate(allFinished.context(0)); v.Outcome != Unknown {
		t.Fatalf("nonexistent inner prop: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}

	// Custom message overrides the default template.
	overridden := resolvePredicateForTest(t, AllActivePlayers(Any(PlayerBool("Eliminated"), PlayerBool("Stood"))).WithMessage("custom.key"))
	if v := overridden.Evaluate(oneUnfinished.context(0)); v.Message == nil || v.Message.Template != "custom.key" {
		t.Fatalf("WithMessage override: Message = %+v, want template custom.key", v.Message)
	}
}

// TestAllActivePlayersSingleLeaf covers the non-"any" inner shape (a bare
// playerBool/propAtLeast/propCompare, no compositor).
func TestAllActivePlayersSingleLeaf(t *testing.T) {
	spec := AllActivePlayers(PlayerBool("Stood"))
	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 1 || pred.Reads[0].Path != "players[*].Stood" {
		t.Fatalf("Reads = %+v", pred.Reads)
	}

	allFinished := buildLegalFixture(t, "blackjackAllFinished")
	if v := pred.Evaluate(allFinished.context(0)); v.Outcome != Pass {
		t.Fatalf("blackjackAllFinished: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	oneUnfinished := buildLegalFixture(t, "blackjackOneUnfinished")
	if v := pred.Evaluate(oneUnfinished.context(0)); v.Outcome != Fail {
		t.Fatalf("blackjackOneUnfinished: Outcome = %v, want Fail", v.Outcome)
	}
}

// TestAllActivePlayersV1InnerRestriction pins the v1 boot error for any
// inner spec name beyond playerBool/propAtLeast/propCompare/any, including
// a nested "any" beneath the top-level "any" (depth-1, same rule as core's
// own any-compositor).
func TestAllActivePlayersV1InnerRestriction(t *testing.T) {
	t.Run("unsupported leaf name", func(t *testing.T) {
		_, err := resolveSpecViaRegistry(AllActivePlayers(ComponentPresentAt("game.Spaces", "move.Idx")), DefaultConstructors(), nil)
		if err == nil {
			t.Fatal("expected a boot error for an unsupported inner predicate name")
		}
	})

	t.Run("nested any beneath any is rejected (depth-1)", func(t *testing.T) {
		nested := AllActivePlayers(Any(
			Any(PlayerBool("Eliminated"), PlayerBool("Stood")),
			PlayerBool("Stood"),
		))
		_, err := resolveSpecViaRegistry(nested, DefaultConstructors(), nil)
		if err == nil {
			t.Fatal("expected a boot error for a nested any beneath AllActivePlayers' any")
		}
	})

	t.Run("any with fewer than 2 subs is rejected", func(t *testing.T) {
		// Hand-build a Spec with a single-element Sub — Any() itself won't
		// stop a caller from doing this (builders stay dumb).
		spec := AllActivePlayers(Spec{Name: "any", Sub: []Spec{PlayerBool("Stood")}})
		_, err := resolveSpecViaRegistry(spec, DefaultConstructors(), nil)
		if err == nil {
			t.Fatal("expected a boot error for an any with fewer than 2 subs")
		}
	})

	t.Run("propAtLeast/propCompare must be player-path", func(t *testing.T) {
		_, err := resolveSpecViaRegistry(AllActivePlayers(PropAtLeast("game.NumCards", 1)), DefaultConstructors(), nil)
		if err == nil {
			t.Fatal("expected a boot error for a non-player-path propAtLeast inner spec")
		}
		_, err2 := resolveSpecViaRegistry(AllActivePlayers(PropCompare("game.NumCards", "==", 1)), DefaultConstructors(), nil)
		if err2 == nil {
			t.Fatal("expected a boot error for a non-player-path propCompare inner spec")
		}
	})

	t.Run("wrong Sub count on allActivePlayers itself", func(t *testing.T) {
		_, err := resolveSpecViaRegistry(Spec{Name: "allActivePlayers", Sub: []Spec{PlayerBool("Stood"), PlayerBool("Eliminated")}}, DefaultConstructors(), nil)
		if err == nil {
			t.Fatal("expected a boot error for allActivePlayers with more than 1 Sub")
		}
	})
}

// TestProposerIsCurrentPlayer covers ProposerIsCurrentPlayer's shape, Reads
// (incl. the FIELD-DEPENDENT move.TargetPlayerIndex read — spec §4),
// pass/fail(x2 distinct branches)/unknown, and the string-parity guarantee:
// each Fail Verdict's "detail" binding carries the EXACT legacy string from
// moves/current_player.go, verbatim.
func TestProposerIsCurrentPlayer(t *testing.T) {
	spec := ProposerIsCurrentPlayer()
	if spec.Name != "proposerIsCurrentPlayer" {
		t.Fatalf("Name = %q, want proposerIsCurrentPlayer", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 2 {
		t.Fatalf("Reads = %+v, want 2 entries", pred.Reads)
	}
	moveRead, ok := findRead(pred.Reads, "move.TargetPlayerIndex")
	if !ok {
		t.Fatal("no Read found for move.TargetPlayerIndex")
	}
	if moveRead.Facet != boardgame.LegalFacetValues {
		t.Fatalf("move.TargetPlayerIndex Facet = %v, want LegalFacetValues", moveRead.Facet)
	}
	// FIELD-DEPENDENT (spec §4): the presence of a move.* Read is exactly
	// what puts this predicate in a plan's fieldDependent bucket rather
	// than fieldIndependent — see legalReadsIncludeMovePath in core
	// (unexported; this is the external-package pin of the same property).
	if !strings.HasPrefix(string(moveRead.Path), "move.") {
		t.Fatalf("expected move.TargetPlayerIndex's Read.Path to have a move.* prefix, got %q", moveRead.Path)
	}
	if _, ok := findRead(pred.Reads, "game.CurrentPlayer"); !ok {
		t.Fatal("no Read found for game.CurrentPlayer")
	}

	// Pass: move.TargetPlayerIndex defaults to player 0 (its PlayerIndex
	// zero value), matching both the current player (0) and the proposer
	// (0) in memoryDefault.
	pass := buildLegalFixture(t, "memoryDefault")
	if v := pred.Evaluate(pass.context(0)); v.Outcome != Pass {
		t.Fatalf("memoryDefault proposer=0: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	// Fail branch 1: TargetPlayerIndex (1) != current player (0). Legacy
	// string: moves/current_player.go:56.
	targetMismatch := buildLegalFixture(t, "memoryTargetPlayerOne")
	v := pred.Evaluate(targetMismatch.context(0))
	if v.Outcome != Fail {
		t.Fatalf("memoryTargetPlayerOne: Outcome = %v, want Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != TemplateProposerNotYourTurn {
		t.Fatalf("memoryTargetPlayerOne: Message = %+v, want template %q", v.Message, TemplateProposerNotYourTurn)
	}
	if got := v.Message.Bindings["detail"]; got.S == nil || *got.S != "it's not your turn" {
		t.Fatalf("memoryTargetPlayerOne: detail binding = %+v, want the verbatim legacy string %q", got, "it's not your turn")
	}

	// Fail branch 2: TargetPlayerIndex (0) == current player (0), but
	// proposer (1) != TargetPlayerIndex. Same legacy string, different
	// triggering condition (moves/current_player.go:60).
	v2 := pred.Evaluate(pass.context(1))
	if v2.Outcome != Fail {
		t.Fatalf("memoryDefault proposer=1: Outcome = %v, want Fail (%+v)", v2.Outcome, v2)
	}
	if got := v2.Message.Bindings["detail"]; got.S == nil || *got.S != "it's not your turn" {
		t.Fatalf("memoryDefault proposer=1: detail binding = %+v, want the verbatim legacy string %q", got, "it's not your turn")
	}

	// Fail branch 3: TargetPlayerIndex is a special negative index
	// (ObserverPlayerIndex) — PlayerIndex.Valid() alone treats it as
	// "valid", but the explicit `< 0` check in moves/current_player.go:51-53
	// rejects it as a move target. Legacy string: moves/current_player.go:48/52.
	targetObserver := buildLegalFixture(t, "memoryTargetObserver")
	v3 := pred.Evaluate(targetObserver.context(0))
	if v3.Outcome != Fail {
		t.Fatalf("memoryTargetObserver: Outcome = %v, want Fail (%+v)", v3.Outcome, v3)
	}
	if v3.Message == nil || v3.Message.Template != TemplateProposerTargetInvalid {
		t.Fatalf("memoryTargetObserver: Message = %+v, want template %q", v3.Message, TemplateProposerTargetInvalid)
	}
	if got := v3.Message.Bindings["detail"]; got.S == nil || *got.S != "The specified target player is not valid" {
		t.Fatalf("memoryTargetObserver: detail binding = %+v, want the verbatim legacy string %q", got, "The specified target player is not valid")
	}

	// Unknown: nil move (Reads declares move.TargetPlayerIndex, so this
	// exercises resolvePlayerIndexPath's own nil-move error, not the
	// runtime undeclared-move-read guard).
	noMove := buildLegalFixture(t, "memoryNoMove")
	if v := pred.Evaluate(noMove.context(0)); v.Outcome != Unknown {
		t.Fatalf("memoryNoMove: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}
}

// TestProposerIsCurrentPlayerTakesNoArgs pins that the predicate rejects a
// spec with args (it has none to accept).
func TestProposerIsCurrentPlayerTakesNoArgs(t *testing.T) {
	spec := Spec{Name: "proposerIsCurrentPlayer", Args: []string{"unexpected"}}
	if _, err := resolveSpecViaRegistry(spec, DefaultConstructors(), nil); err == nil {
		t.Fatal("expected an error constructing proposerIsCurrentPlayer with args")
	}
}
