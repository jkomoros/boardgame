package legal

import (
	"testing"

	"github.com/jkomoros/boardgame"
)

// TestRevealableCardAt covers the design spec §8 acid test's two-branch
// disambiguation, Reads (occupancy facets ONLY, never values — this is what
// keeps the predicate client-evaluable under memory's sanitize:"order"
// policy), and the two Fail templates, which pin the exact legacy strings
// from examples/memory/moves.go:56/58 by template-key mapping (see
// TemplateNoCardHere/TemplateAlreadyRevealed's doc comments in
// catalog_purpose.go for the verbatim strings themselves).
func TestRevealableCardAt(t *testing.T) {
	spec := RevealableCardAt("game.HiddenCards", "game.VisibleCards", "move.CardIndex")
	if spec.Name != "revealableCardAt" {
		t.Fatalf("Name = %q, want revealableCardAt", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 3 {
		t.Fatalf("Reads = %+v, want 3 entries", pred.Reads)
	}
	if pred.Reads[0].Path != "game.HiddenCards" || pred.Reads[0].Facet != boardgame.LegalFacetOccupancy {
		t.Fatalf("Reads[0] = %+v, want game.HiddenCards/FacetOccupancy", pred.Reads[0])
	}
	if pred.Reads[1].Path != "game.VisibleCards" || pred.Reads[1].Facet != boardgame.LegalFacetOccupancy {
		t.Fatalf("Reads[1] = %+v, want game.VisibleCards/FacetOccupancy", pred.Reads[1])
	}
	if pred.Reads[2].Path != "move.CardIndex" || pred.Reads[2].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads[2] = %+v, want move.CardIndex/FacetValues", pred.Reads[2])
	}

	// Pass: the card is still hidden (memoryDefault's HiddenCards[0] is
	// occupied).
	pass := buildLegalFixture(t, "memoryDefault")
	if v := pred.Evaluate(pass.context(0)); v.Outcome != Pass {
		t.Fatalf("memoryDefault: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	// Fail (already revealed): hidden empty, visible occupied, same idx.
	alreadyRevealed := buildLegalFixture(t, "memoryCardAlreadyRevealed")
	v := pred.Evaluate(alreadyRevealed.context(0))
	if v.Outcome != Fail {
		t.Fatalf("memoryCardAlreadyRevealed: Outcome = %v, want Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != TemplateAlreadyRevealed {
		t.Fatalf("memoryCardAlreadyRevealed: Message = %+v, want template %q", v.Message, TemplateAlreadyRevealed)
	}

	// Fail (no card here): hidden empty, visible ALSO empty, same idx.
	neverThere := buildLegalFixture(t, "memoryCardNeverThere")
	v2 := pred.Evaluate(neverThere.context(0))
	if v2.Outcome != Fail {
		t.Fatalf("memoryCardNeverThere: Outcome = %v, want Fail (%+v)", v2.Outcome, v2)
	}
	if v2.Message == nil || v2.Message.Template != TemplateNoCardHere {
		t.Fatalf("memoryCardNeverThere: Message = %+v, want template %q", v2.Message, TemplateNoCardHere)
	}

	// Unknown: nil move (idxField can't resolve).
	noMove := buildLegalFixture(t, "memoryNoMove")
	if v := pred.Evaluate(noMove.context(0)); v.Outcome != Unknown {
		t.Fatalf("memoryNoMove: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}

	// Bad arg count is a construction-time error.
	if _, err := resolveSpecViaRegistry(Spec{Name: "revealableCardAt", Args: []string{"a", "b"}}, DefaultConstructors(), nil); err == nil {
		t.Fatal("expected an error constructing revealableCardAt with 2 args")
	}
}

// TestComponentPropEqualsCurrentPlayer covers checkers' token-ownership
// check (examples/checkers/moves.go:93-138's "!p.Color.Equals(t.Color)"),
// generalized. Reads declare FacetValues on the stack path (not
// FacetOccupancy — this predicate reads a component's property VALUE, not
// merely its presence).
func TestComponentPropEqualsCurrentPlayer(t *testing.T) {
	spec := ComponentPropEqualsCurrentPlayer("game.Spaces", "move.TokenIndexToMove", "Color")
	if spec.Name != "componentPropEqualsCurrentPlayer" {
		t.Fatalf("Name = %q, want componentPropEqualsCurrentPlayer", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 3 {
		t.Fatalf("Reads = %+v, want 3 entries", pred.Reads)
	}
	if pred.Reads[0].Path != "game.Spaces" || pred.Reads[0].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads[0] = %+v, want game.Spaces/FacetValues", pred.Reads[0])
	}
	if pred.Reads[1].Path != "move.TokenIndexToMove" || pred.Reads[1].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads[1] = %+v, want move.TokenIndexToMove/FacetValues", pred.Reads[1])
	}
	if pred.Reads[2].Path != "player.Color" || pred.Reads[2].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads[2] = %+v, want player.Color/FacetValues", pred.Reads[2])
	}

	// Pass: the token at the space belongs to the current player.
	own := buildLegalFixture(t, "checkersOwnToken")
	if v := pred.Evaluate(own.context(0)); v.Outcome != Pass {
		t.Fatalf("checkersOwnToken: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	// Fail: the token at the space belongs to the opponent.
	opponent := buildLegalFixture(t, "checkersOpponentToken")
	v := pred.Evaluate(opponent.context(0))
	if v.Outcome != Fail {
		t.Fatalf("checkersOpponentToken: Outcome = %v, want Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != TemplateComponentPropNotCurrentPlayer {
		t.Fatalf("checkersOpponentToken: Message = %+v, want template %q", v.Message, TemplateComponentPropNotCurrentPlayer)
	}
	if got := v.Message.Bindings["prop"]; got.S == nil || *got.S != "Color" {
		t.Fatalf("checkersOpponentToken: prop binding = %+v, want Color", got)
	}

	// Unknown: no component at the keyed space (presence is a SEPARATE
	// predicate's job — ComponentPresentAtKey — so this predicate can't
	// evaluate a color comparison against nothing).
	empty := buildLegalFixture(t, "checkersEmptySpace")
	if v := pred.Evaluate(empty.context(0)); v.Outcome != Unknown {
		t.Fatalf("checkersEmptySpace: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}

	// Unknown: nil move.
	noMove := buildLegalFixture(t, "checkersNoMove")
	if v := pred.Evaluate(noMove.context(0)); v.Outcome != Unknown {
		t.Fatalf("checkersNoMove: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}

	// Bad arg count is a construction-time error.
	if _, err := resolveSpecViaRegistry(Spec{Name: "componentPropEqualsCurrentPlayer", Args: []string{"a", "b"}}, DefaultConstructors(), nil); err == nil {
		t.Fatal("expected an error constructing componentPropEqualsCurrentPlayer with 2 args")
	}
}
