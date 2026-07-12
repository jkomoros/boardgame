package legal_test

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

// TestStackCount covers StackCount's ops, pass/fail/unknown, Reads (a
// FacetCount read on the stack path), and template key/bindings.
func TestStackCount(t *testing.T) {
	spec := legal.StackCount("game.HiddenCards", ">=", 20)
	if spec.Name != "stackCount" {
		t.Fatalf("Name = %q, want stackCount", spec.Name)
	}
	if len(spec.Args) != 3 || spec.Args[0] != "game.HiddenCards" || spec.Args[1] != ">=" || spec.Args[2] != "20" {
		t.Fatalf("Args = %v", spec.Args)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 1 || pred.Reads[0].Path != "game.HiddenCards" || pred.Reads[0].Facet != boardgame.LegalFacetCount {
		t.Fatalf("Reads = %+v", pred.Reads)
	}

	// Pass: memoryDefault's HiddenCards has 20 components.
	pass := buildLegalFixture(t, "memoryDefault")
	if v := pred.Evaluate(pass.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("pass fixture: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	// Fail: HiddenCards has 20 components, not == 5.
	failPred := resolvePredicateForTest(t, legal.StackCount("game.HiddenCards", "==", 5))
	v := failPred.Evaluate(pass.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("fail fixture: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplateStackCount {
		t.Fatalf("fail fixture: legal.Message = %+v, want template %q", v.Message, legal.TemplateStackCount)
	}
	if got := v.Message.Bindings["count"]; got.I == nil || *got.I != 20 {
		t.Fatalf("fail fixture: count binding = %+v, want 20", got)
	}
	if got := v.Message.Bindings["op"]; got.S == nil || *got.S != "==" {
		t.Fatalf("fail fixture: op binding = %+v, want ==", got)
	}
	if got := v.Message.Bindings["n"]; got.I == nil || *got.I != 5 {
		t.Fatalf("fail fixture: n binding = %+v, want 5", got)
	}

	// Pass: VisibleCards has 0 components.
	emptyPred := resolvePredicateForTest(t, legal.StackCount("game.VisibleCards", "==", 0))
	if v := emptyPred.Evaluate(pass.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("VisibleCards==0: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	// Every operator exercised against a known count (HiddenCards == 20).
	opCases := []struct {
		op   string
		n    int
		want legal.Outcome
	}{
		{"==", 20, legal.Pass},
		{"==", 5, legal.Fail},
		{"!=", 5, legal.Pass},
		{"!=", 20, legal.Fail},
		{"<", 21, legal.Pass},
		{"<", 20, legal.Fail},
		{"<=", 20, legal.Pass},
		{"<=", 19, legal.Fail},
		{">", 19, legal.Pass},
		{">", 20, legal.Fail},
		{">=", 20, legal.Pass},
		{">=", 21, legal.Fail},
	}
	for _, oc := range opCases {
		p := resolvePredicateForTest(t, legal.StackCount("game.HiddenCards", oc.op, oc.n))
		if v := p.Evaluate(pass.context(0)); v.Outcome != oc.want {
			t.Errorf("op %s %d: legal.Outcome = %v, want %v", oc.op, oc.n, v.Outcome, oc.want)
		}
	}

	// Bad op is a construction-time error, not a runtime Unknown.
	if _, err := resolveSpecViaRegistry(legal.StackCount("game.HiddenCards", "~=", 1), legal.DefaultConstructors(), nil); err == nil {
		t.Fatal("expected an error constructing stackCount with an invalid op")
	}

	// Unknown: stack path doesn't exist.
	unknownPred := resolvePredicateForTest(t, legal.StackCount("game.NoSuchStackProp", ">=", 0))
	if v := unknownPred.Evaluate(pass.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("unknown-path: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}
}

// TestStackEmpty covers StackEmpty's pass/fail/unknown, Reads (a
// FacetNonEmpty read on the stack path), and template key.
func TestStackEmpty(t *testing.T) {
	spec := legal.StackEmpty("game.VisibleCards")
	if spec.Name != "stackEmpty" {
		t.Fatalf("Name = %q, want stackEmpty", spec.Name)
	}
	if len(spec.Args) != 1 || spec.Args[0] != "game.VisibleCards" {
		t.Fatalf("Args = %v", spec.Args)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 1 || pred.Reads[0].Path != "game.VisibleCards" || pred.Reads[0].Facet != legal.FacetNonEmpty {
		t.Fatalf("Reads = %+v", pred.Reads)
	}

	fixture := buildLegalFixture(t, "memoryDefault")
	if v := pred.Evaluate(fixture.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("VisibleCards empty: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	failPred := resolvePredicateForTest(t, legal.StackEmpty("game.HiddenCards"))
	v := failPred.Evaluate(fixture.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("HiddenCards not empty: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplateStackEmpty {
		t.Fatalf("legal.Message = %+v, want template %q", v.Message, legal.TemplateStackEmpty)
	}

	unknownPred := resolvePredicateForTest(t, legal.StackEmpty("game.NoSuchStackProp"))
	if v := unknownPred.Evaluate(fixture.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("unknown-path: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}
}

// TestStackNotEmpty covers StackNotEmpty's pass/fail/unknown, Reads (the
// same FacetNonEmpty facet StackEmpty uses), and template key.
func TestStackNotEmpty(t *testing.T) {
	spec := legal.StackNotEmpty("game.HiddenCards")
	if spec.Name != "stackNotEmpty" {
		t.Fatalf("Name = %q, want stackNotEmpty", spec.Name)
	}
	if len(spec.Args) != 1 || spec.Args[0] != "game.HiddenCards" {
		t.Fatalf("Args = %v", spec.Args)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 1 || pred.Reads[0].Path != "game.HiddenCards" || pred.Reads[0].Facet != legal.FacetNonEmpty {
		t.Fatalf("Reads = %+v", pred.Reads)
	}

	fixture := buildLegalFixture(t, "memoryDefault")
	if v := pred.Evaluate(fixture.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("HiddenCards not empty: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	failPred := resolvePredicateForTest(t, legal.StackNotEmpty("game.VisibleCards"))
	v := failPred.Evaluate(fixture.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("VisibleCards empty: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplateStackNotEmpty {
		t.Fatalf("legal.Message = %+v, want template %q", v.Message, legal.TemplateStackNotEmpty)
	}

	unknownPred := resolvePredicateForTest(t, legal.StackNotEmpty("game.NoSuchStackProp"))
	if v := unknownPred.Evaluate(fixture.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("unknown-path: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}
}
