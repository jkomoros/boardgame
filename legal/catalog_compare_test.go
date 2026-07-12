package legal_test

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

// TestPropEquals covers PropEquals' four type arms (int, bool, enum,
// PlayerIndex incl. the "observer"/"admin" specials), Reads, template key/
// bindings, wrong-typed-path Unknown, and unparseable-value Unknown per
// arm. See propEqualsFamilyConstructor's doc comment (catalog_compare.go)
// for why type dispatch happens at Evaluate rather than at construction —
// this is the task's recorded judgment call.
func TestPropEquals(t *testing.T) {
	spec := legal.PropEquals("player.CardsLeftToReveal", "2")
	if spec.Name != "propEquals" {
		t.Fatalf("Name = %q, want propEquals", spec.Name)
	}
	if len(spec.Args) != 2 || spec.Args[0] != "player.CardsLeftToReveal" || spec.Args[1] != "2" {
		t.Fatalf("Args = %v", spec.Args)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 1 || pred.Reads[0].Path != "player.CardsLeftToReveal" || pred.Reads[0].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads = %+v", pred.Reads)
	}

	memoryDefault := buildLegalFixture(t, "memoryDefault")

	// int arm: pass.
	if v := pred.Evaluate(memoryDefault.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("int pass: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	// int arm: fail, with the {value, want} bindings.
	failPred := resolvePredicateForTest(t, legal.PropEquals("player.CardsLeftToReveal", "0"))
	v := failPred.Evaluate(memoryDefault.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("int fail: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplatePropEquals {
		t.Fatalf("int fail: legal.Message = %+v, want template %q", v.Message, legal.TemplatePropEquals)
	}
	if got := v.Message.Bindings["value"]; got.S == nil || *got.S != "2" {
		t.Fatalf("int fail: value binding = %+v, want \"2\"", got)
	}
	if got := v.Message.Bindings["want"]; got.S == nil || *got.S != "0" {
		t.Fatalf("int fail: want binding = %+v, want \"0\"", got)
	}

	// int arm: value doesn't parse as an int -> Unknown, not a panic or a
	// construction error (the arm isn't known until Evaluate resolves the
	// path's PropertyType).
	badIntPred := resolvePredicateForTest(t, legal.PropEquals("player.CardsLeftToReveal", "not-a-number"))
	if v := badIntPred.Evaluate(memoryDefault.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("unparseable int value: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}

	// bool arm: pass and fail.
	boolPassPred := resolvePredicateForTest(t, legal.PropEquals("player.PlayerInactive", "false"))
	if v := boolPassPred.Evaluate(memoryDefault.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("bool pass: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}
	boolFailPred := resolvePredicateForTest(t, legal.PropEquals("player.PlayerInactive", "true"))
	if v := boolFailPred.Evaluate(memoryDefault.context(0)); v.Outcome != legal.Fail {
		t.Fatalf("bool fail: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}

	// bool arm: value isn't "true"/"false" -> Unknown.
	badBoolPred := resolvePredicateForTest(t, legal.PropEquals("player.PlayerInactive", "yes"))
	if v := badBoolPred.Evaluate(memoryDefault.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("unparseable bool value: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}

	// enum arm: checkersDefault's current player Color defaults to "Red".
	checkersDefault := buildLegalFixture(t, "checkersDefault")
	enumPassPred := resolvePredicateForTest(t, legal.PropEquals("player.Color", "Red"))
	if v := enumPassPred.Evaluate(checkersDefault.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("enum pass: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}
	enumFailPred := resolvePredicateForTest(t, legal.PropEquals("player.Color", "Black"))
	fv := enumFailPred.Evaluate(checkersDefault.context(0))
	if fv.Outcome != legal.Fail {
		t.Fatalf("enum fail: legal.Outcome = %v, want legal.Fail (%+v)", fv.Outcome, fv)
	}
	if got := fv.Message.Bindings["value"]; got.S == nil || *got.S != "Red" {
		t.Fatalf("enum fail: value binding = %+v, want \"Red\"", got)
	}

	// enum arm: unknown value NAME -> Unknown (the design spec's "boot
	// error naming move+path+value" aspiration is not deliverable without
	// an example state reaching the constructor -- see the doc comment on
	// propEqualsFamilyConstructor; this is the fail-closed runtime
	// fallback).
	enumUnknownPred := resolvePredicateForTest(t, legal.PropEquals("player.Color", "Chartreuse"))
	if v := enumUnknownPred.Evaluate(checkersDefault.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("unknown enum name: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}

	// PlayerIndex arm: plain int, and the "observer"/"admin" specials.
	piIntPred := resolvePredicateForTest(t, legal.PropEquals("move.TargetPlayerIndex", "0"))
	if v := piIntPred.Evaluate(memoryDefault.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("PlayerIndex int pass: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}
	piFailPred := resolvePredicateForTest(t, legal.PropEquals("move.TargetPlayerIndex", "1"))
	if v := piFailPred.Evaluate(memoryDefault.context(0)); v.Outcome != legal.Fail {
		t.Fatalf("PlayerIndex int fail: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}

	targetOne := buildLegalFixture(t, "memoryTargetPlayerOne")
	piOnePred := resolvePredicateForTest(t, legal.PropEquals("move.TargetPlayerIndex", "1"))
	if v := piOnePred.Evaluate(targetOne.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("PlayerIndex int 1 pass: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	targetObserver := buildLegalFixture(t, "memoryTargetObserver")
	piObserverPred := resolvePredicateForTest(t, legal.PropEquals("move.TargetPlayerIndex", "observer"))
	if v := piObserverPred.Evaluate(targetObserver.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("PlayerIndex observer pass: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}
	piAdminPred := resolvePredicateForTest(t, legal.PropEquals("move.TargetPlayerIndex", "admin"))
	if v := piAdminPred.Evaluate(targetObserver.context(0)); v.Outcome != legal.Fail {
		t.Fatalf("PlayerIndex admin vs observer fail: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}

	// PlayerIndex arm: unparseable value (not an int, not observer/admin)
	// -> Unknown.
	badPiPred := resolvePredicateForTest(t, legal.PropEquals("move.TargetPlayerIndex", "not-an-index"))
	if v := badPiPred.Evaluate(memoryDefault.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("unparseable PlayerIndex value: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}

	// Wrong-typed path (a stack property) -> Unknown, never a panic.
	stackTypedPred := resolvePredicateForTest(t, legal.PropEquals("game.HiddenCards", "foo"))
	if v := stackTypedPred.Evaluate(memoryDefault.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("stack-typed path: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}

	// Unresolvable path -> Unknown.
	unresolvablePred := resolvePredicateForTest(t, legal.PropEquals("game.NoSuchIntProp", "0"))
	if v := unresolvablePred.Evaluate(memoryDefault.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("unresolvable path: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}

	// Custom message overrides the default template.
	overridden := resolvePredicateForTest(t, legal.PropEquals("player.CardsLeftToReveal", "0").WithMessage("custom.key"))
	if v := overridden.Evaluate(memoryDefault.context(0)); v.Message == nil || v.Message.Template != "custom.key" {
		t.Fatalf("WithMessage override: legal.Message = %+v, want template custom.key", v.Message)
	}
}

// TestPropNotEquals covers PropNotEquals as PropEquals' negation: matching
// pass/fail flips relative to PropEquals, but an Unknown (unparseable
// value, unresolvable/wrong-typed path, unknown enum name) is never flipped
// to a Pass.
func TestPropNotEquals(t *testing.T) {
	spec := legal.PropNotEquals("player.CardsLeftToReveal", "0")
	if spec.Name != "propNotEquals" {
		t.Fatalf("Name = %q, want propNotEquals", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 1 || pred.Reads[0].Path != "player.CardsLeftToReveal" || pred.Reads[0].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads = %+v", pred.Reads)
	}

	memoryDefault := buildLegalFixture(t, "memoryDefault")

	// int arm: CardsLeftToReveal == 2, so != 0 passes and != 2 fails.
	if v := pred.Evaluate(memoryDefault.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("int pass: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}
	failPred := resolvePredicateForTest(t, legal.PropNotEquals("player.CardsLeftToReveal", "2"))
	v := failPred.Evaluate(memoryDefault.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("int fail: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplatePropNotEquals {
		t.Fatalf("int fail: legal.Message = %+v, want template %q", v.Message, legal.TemplatePropNotEquals)
	}

	// enum arm: unknown value NAME stays Unknown, not negated to Pass.
	checkersDefault := buildLegalFixture(t, "checkersDefault")
	enumUnknownPred := resolvePredicateForTest(t, legal.PropNotEquals("player.Color", "Chartreuse"))
	if v := enumUnknownPred.Evaluate(checkersDefault.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("unknown enum name: legal.Outcome = %v, want legal.Unknown (%+v) -- negation must not rescue an Unknown into a Pass", v.Outcome, v)
	}

	// Wrong-typed path stays Unknown under negation too.
	stackTypedPred := resolvePredicateForTest(t, legal.PropNotEquals("game.HiddenCards", "foo"))
	if v := stackTypedPred.Evaluate(memoryDefault.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("stack-typed path: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}

	// PlayerIndex arm negation.
	piPassPred := resolvePredicateForTest(t, legal.PropNotEquals("move.TargetPlayerIndex", "1"))
	if v := piPassPred.Evaluate(memoryDefault.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("PlayerIndex pass: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}
	piFailPred := resolvePredicateForTest(t, legal.PropNotEquals("move.TargetPlayerIndex", "0"))
	if v := piFailPred.Evaluate(memoryDefault.context(0)); v.Outcome != legal.Fail {
		t.Fatalf("PlayerIndex fail: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
}

// TestPropEqualsBadArgCount verifies both propEquals and propNotEquals
// reject a wrong arg count at construction time.
func TestPropEqualsBadArgCount(t *testing.T) {
	badSpec := legal.Spec{Name: "propEquals", Args: []string{"player.CardsLeftToReveal"}}
	if _, err := resolveSpecViaRegistry(badSpec, legal.DefaultConstructors(), nil); err == nil {
		t.Fatal("expected an error constructing propEquals with 1 arg")
	}
	badSpec2 := legal.Spec{Name: "propNotEquals", Args: []string{"player.CardsLeftToReveal"}}
	if _, err := resolveSpecViaRegistry(badSpec2, legal.DefaultConstructors(), nil); err == nil {
		t.Fatal("expected an error constructing propNotEquals with 1 arg")
	}
}
