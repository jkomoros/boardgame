package legal_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/constraints"
	"github.com/jkomoros/boardgame/legal"
)

// TestPropAtLeast covers PropAtLeast's pass/fail/unknown paths, its
// declared Reads, and the template key used on Fail.
func TestPropAtLeast(t *testing.T) {
	spec := legal.PropAtLeast("player.CardsLeftToReveal", 1)
	if spec.Name != "propAtLeast" {
		t.Fatalf("Name = %q, want propAtLeast", spec.Name)
	}
	if len(spec.Args) != 2 || spec.Args[0] != "player.CardsLeftToReveal" || spec.Args[1] != "1" {
		t.Fatalf("Args = %v", spec.Args)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 1 || pred.Reads[0].Path != "player.CardsLeftToReveal" || pred.Reads[0].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads = %+v", pred.Reads)
	}

	// Pass: memoryDefault has CardsLeftToReveal == 2 >= 1.
	pass := buildLegalFixture(t, "memoryDefault")
	if v := pred.Evaluate(pass.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("pass fixture: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	// Fail: memoryZeroCardsLeft has CardsLeftToReveal == 0 < 1.
	fail := buildLegalFixture(t, "memoryZeroCardsLeft")
	v := pred.Evaluate(fail.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("fail fixture: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplatePropAtLeast {
		t.Fatalf("fail fixture: legal.Message = %+v, want template %q", v.Message, legal.TemplatePropAtLeast)
	}
	if got := v.Message.Bindings["value"]; got.I == nil || *got.I != 0 {
		t.Fatalf("fail fixture: value binding = %+v, want 0", got)
	}
	if got := v.Message.Bindings["min"]; got.I == nil || *got.I != 1 {
		t.Fatalf("fail fixture: min binding = %+v, want 1", got)
	}

	// Unknown: property doesn't exist.
	unknownSpec := legal.PropAtLeast("game.NoSuchIntProp", 1)
	unknownPred := resolvePredicateForTest(t, unknownSpec)
	if v := unknownPred.Evaluate(pass.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("unknown-path: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}

	// Custom message overrides the default template.
	overridden := resolvePredicateForTest(t, legal.PropAtLeast("player.CardsLeftToReveal", 1).WithMessage("custom.key"))
	if v := overridden.Evaluate(fail.context(0)); v.Message == nil || v.Message.Template != "custom.key" {
		t.Fatalf("WithMessage override: legal.Message = %+v, want template custom.key", v.Message)
	}
}

// TestPropCompare covers PropCompare's ops, pass/fail/unknown, Reads, and
// bad-op construction-time validation.
func TestPropCompare(t *testing.T) {
	spec := legal.PropCompare("player.CardsLeftToReveal", "==", 2)
	if spec.Name != "propCompare" {
		t.Fatalf("Name = %q, want propCompare", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 1 || pred.Reads[0].Path != "player.CardsLeftToReveal" || pred.Reads[0].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads = %+v", pred.Reads)
	}

	pass := buildLegalFixture(t, "memoryDefault")
	if v := pred.Evaluate(pass.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("pass fixture: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	fail := buildLegalFixture(t, "memoryZeroCardsLeft")
	v := pred.Evaluate(fail.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("fail fixture: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplatePropCompare {
		t.Fatalf("fail fixture: legal.Message = %+v, want template %q", v.Message, legal.TemplatePropCompare)
	}

	unknownPred := resolvePredicateForTest(t, legal.PropCompare("game.NoSuchIntProp", "==", 0))
	if v := unknownPred.Evaluate(pass.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("unknown-path: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}

	// Every operator exercised against a known value (CardsLeftToReveal == 2).
	opCases := []struct {
		op   string
		n    int
		want legal.Outcome
	}{
		{"==", 2, legal.Pass},
		{"==", 3, legal.Fail},
		{"!=", 3, legal.Pass},
		{"!=", 2, legal.Fail},
		{"<", 3, legal.Pass},
		{"<", 2, legal.Fail},
		{"<=", 2, legal.Pass},
		{"<=", 1, legal.Fail},
		{">", 1, legal.Pass},
		{">", 2, legal.Fail},
		{">=", 2, legal.Pass},
		{">=", 3, legal.Fail},
	}
	for _, oc := range opCases {
		p := resolvePredicateForTest(t, legal.PropCompare("player.CardsLeftToReveal", oc.op, oc.n))
		if v := p.Evaluate(pass.context(0)); v.Outcome != oc.want {
			t.Errorf("op %s %d: legal.Outcome = %v, want %v", oc.op, oc.n, v.Outcome, oc.want)
		}
	}

	// Bad op is a construction-time error, not a runtime Unknown.
	if _, err := resolveSpecViaRegistry(legal.PropCompare("game.NumCards", "~=", 1), legal.DefaultConstructors(), nil); err == nil {
		t.Fatal("expected an error constructing propCompare with an invalid op")
	}
}

// TestPlayerBool covers PlayerBool's pass/fail/unknown, Reads, and template
// key.
func TestPlayerBool(t *testing.T) {
	spec := legal.PlayerBool("SeatFilled")
	if spec.Name != "playerBool" {
		t.Fatalf("Name = %q, want playerBool", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 1 || pred.Reads[0].Path != "player.SeatFilled" || pred.Reads[0].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads = %+v", pred.Reads)
	}

	pass := buildLegalFixture(t, "memorySeatFilled")
	if v := pred.Evaluate(pass.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("pass fixture: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	fail := buildLegalFixture(t, "memoryDefault")
	v := pred.Evaluate(fail.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("fail fixture: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplatePlayerBool {
		t.Fatalf("fail fixture: legal.Message = %+v, want template %q", v.Message, legal.TemplatePlayerBool)
	}
	if got := v.Message.Bindings["prop"]; got.S == nil || *got.S != "SeatFilled" {
		t.Fatalf("fail fixture: prop binding = %+v, want SeatFilled", got)
	}

	unknownPred := resolvePredicateForTest(t, legal.PlayerBool("NoSuchBoolProp"))
	if v := unknownPred.Evaluate(fail.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("unknown-prop: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}
}

// TestPlayerBoolIs covers PlayerBoolIs (spec §4's negation leaf): the 2-arg
// "playerBool" spelling, both want=true (byte-for-byte the same registry
// entry and args-count PlayerBool's own 1-arg spelling resolves to a
// backward-compatible want=true) and want=false, plus the bad-want
// construction-time error.
func TestPlayerBoolIs(t *testing.T) {
	trueSpec := legal.PlayerBoolIs("SeatFilled", true)
	if trueSpec.Name != "playerBool" {
		t.Fatalf("Name = %q, want playerBool", trueSpec.Name)
	}
	if len(trueSpec.Args) != 2 || trueSpec.Args[0] != "SeatFilled" || trueSpec.Args[1] != "true" {
		t.Fatalf("Args = %v, want [SeatFilled true]", trueSpec.Args)
	}

	falseSpec := legal.PlayerBoolIs("SeatFilled", false)
	if len(falseSpec.Args) != 2 || falseSpec.Args[0] != "SeatFilled" || falseSpec.Args[1] != "false" {
		t.Fatalf("Args = %v, want [SeatFilled false]", falseSpec.Args)
	}

	seatFilled := buildLegalFixture(t, "memorySeatFilled") // SeatFilled == true
	seatEmpty := buildLegalFixture(t, "memoryDefault")     // SeatFilled == false

	// want=true behaves exactly like PlayerBool("SeatFilled").
	truePred := resolvePredicateForTest(t, trueSpec)
	if v := truePred.Evaluate(seatFilled.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("want=true, SeatFilled=true: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}
	v := truePred.Evaluate(seatEmpty.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("want=true, SeatFilled=false: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplatePlayerBool {
		t.Fatalf("legal.Message = %+v, want template %q", v.Message, legal.TemplatePlayerBool)
	}
	if got := v.Message.Bindings["want"]; got.S == nil || *got.S != "true" {
		t.Fatalf("want binding = %+v, want \"true\"", got)
	}

	// want=false is the negation: Pass when SeatFilled is false, Fail (with
	// the want=false binding) when SeatFilled is true.
	falsePred := resolvePredicateForTest(t, falseSpec)
	if v := falsePred.Evaluate(seatEmpty.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("want=false, SeatFilled=false: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}
	v2 := falsePred.Evaluate(seatFilled.context(0))
	if v2.Outcome != legal.Fail {
		t.Fatalf("want=false, SeatFilled=true: legal.Outcome = %v, want legal.Fail (%+v)", v2.Outcome, v2)
	}
	if v2.Message == nil || v2.Message.Template != legal.TemplatePlayerBool {
		t.Fatalf("legal.Message = %+v, want template %q", v2.Message, legal.TemplatePlayerBool)
	}
	if got := v2.Message.Bindings["want"]; got.S == nil || *got.S != "false" {
		t.Fatalf("want binding = %+v, want \"false\"", got)
	}

	// A hand-built spec with a bad 2nd arg is a construction-time error.
	badSpec := legal.Spec{Name: "playerBool", Args: []string{"SeatFilled", "nope"}}
	if _, err := resolveSpecViaRegistry(badSpec, legal.DefaultConstructors(), nil); err == nil {
		t.Fatal("expected an error constructing playerBool with an invalid want arg")
	}
}

// TestComponentPresentAt covers ComponentPresentAt's pass/fail/unknown,
// Reads (occupancy facet on the stack, values facet on the index), and
// template key.
func TestComponentPresentAt(t *testing.T) {
	spec := legal.ComponentPresentAt("game.HiddenCards", "move.CardIndex")
	if spec.Name != "componentPresentAt" {
		t.Fatalf("Name = %q, want componentPresentAt", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 2 {
		t.Fatalf("Reads = %+v, want 2 entries", pred.Reads)
	}
	if pred.Reads[0].Path != "game.HiddenCards" || pred.Reads[0].Facet != boardgame.LegalFacetOccupancy {
		t.Fatalf("Reads[0] = %+v, want game.HiddenCards/FacetOccupancy", pred.Reads[0])
	}
	if pred.Reads[1].Path != "move.CardIndex" || pred.Reads[1].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads[1] = %+v, want move.CardIndex/FacetValues", pred.Reads[1])
	}

	fixture := buildLegalFixture(t, "memoryDefault")
	if v := pred.Evaluate(fixture.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("HiddenCards[0] present: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	failSpec := legal.ComponentPresentAt("game.VisibleCards", "move.CardIndex")
	failPred := resolvePredicateForTest(t, failSpec)
	v := failPred.Evaluate(fixture.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("VisibleCards[0] absent: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplateComponentMissing {
		t.Fatalf("legal.Message = %+v, want template %q", v.Message, legal.TemplateComponentMissing)
	}
	if got := v.Message.Bindings["index"]; got.I == nil || *got.I != 0 {
		t.Fatalf("index binding = %+v, want 0", got)
	}

	noMove := buildLegalFixture(t, "memoryNoMove")
	if v := pred.Evaluate(noMove.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("nil move: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}
}

// TestComponentAbsentAt covers ComponentAbsentAt (spec §4's negation leaf,
// the exact inverse of ComponentPresentAt): pass/fail/unknown, Reads
// (occupancy facet on the stack, values facet on the index, mirroring
// ComponentPresentAt exactly), and template key.
func TestComponentAbsentAt(t *testing.T) {
	spec := legal.ComponentAbsentAt("game.VisibleCards", "move.CardIndex")
	if spec.Name != "componentAbsentAt" {
		t.Fatalf("Name = %q, want componentAbsentAt", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 2 {
		t.Fatalf("Reads = %+v, want 2 entries", pred.Reads)
	}
	if pred.Reads[0].Path != "game.VisibleCards" || pred.Reads[0].Facet != boardgame.LegalFacetOccupancy {
		t.Fatalf("Reads[0] = %+v, want game.VisibleCards/FacetOccupancy", pred.Reads[0])
	}
	if pred.Reads[1].Path != "move.CardIndex" || pred.Reads[1].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads[1] = %+v, want move.CardIndex/FacetValues", pred.Reads[1])
	}

	// Pass: VisibleCards[0] is empty in memoryDefault.
	fixture := buildLegalFixture(t, "memoryDefault")
	if v := pred.Evaluate(fixture.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("VisibleCards[0] absent: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	// Fail: HiddenCards[0] is occupied in memoryDefault.
	failSpec := legal.ComponentAbsentAt("game.HiddenCards", "move.CardIndex")
	failPred := resolvePredicateForTest(t, failSpec)
	v := failPred.Evaluate(fixture.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("HiddenCards[0] present: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplateComponentPresentUnexpected {
		t.Fatalf("legal.Message = %+v, want template %q", v.Message, legal.TemplateComponentPresentUnexpected)
	}
	if got := v.Message.Bindings["index"]; got.I == nil || *got.I != 0 {
		t.Fatalf("index binding = %+v, want 0", got)
	}

	// Unknown: nil move (Reads declares move.CardIndex).
	noMove := buildLegalFixture(t, "memoryNoMove")
	if v := pred.Evaluate(noMove.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("nil move: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}
}

// TestComponentPresentAtKey covers ComponentPresentAtKey's pass/fail/unknown,
// Reads, and template key, using checkers' enum-keyed Spaces stack.
func TestComponentPresentAtKey(t *testing.T) {
	spec := legal.ComponentPresentAtKey("game.Spaces", "move.TokenIndexToMove")
	if spec.Name != "componentPresentAtKey" {
		t.Fatalf("Name = %q, want componentPresentAtKey", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 2 {
		t.Fatalf("Reads = %+v, want 2 entries", pred.Reads)
	}
	if pred.Reads[0].Path != "game.Spaces" || pred.Reads[0].Facet != boardgame.LegalFacetOccupancy {
		t.Fatalf("Reads[0] = %+v, want game.Spaces/FacetOccupancy", pred.Reads[0])
	}
	if pred.Reads[1].Path != "move.TokenIndexToMove" || pred.Reads[1].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads[1] = %+v, want move.TokenIndexToMove/FacetValues", pred.Reads[1])
	}

	occupied := buildLegalFixture(t, "checkersDefault")
	if v := pred.Evaluate(occupied.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("occupied space: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	empty := buildLegalFixture(t, "checkersEmptySpace")
	v := pred.Evaluate(empty.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("empty space: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplateComponentMissingKey {
		t.Fatalf("legal.Message = %+v, want template %q", v.Message, legal.TemplateComponentMissingKey)
	}

	noMove := buildLegalFixture(t, "checkersNoMove")
	if v := pred.Evaluate(noMove.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("nil move: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}
}

// TestMayMoveTo covers MayMoveTo's pass/fail(x2 templates)/unknown, Reads,
// and template keys.
func TestMayMoveTo(t *testing.T) {
	spec := legal.MayMoveTo("game.HiddenCards", "game.VisibleCards", "move.CardIndex")
	if spec.Name != "mayMoveTo" {
		t.Fatalf("Name = %q, want mayMoveTo", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 3 {
		t.Fatalf("Reads = %+v, want 3 entries", pred.Reads)
	}
	if pred.Reads[0].Path != "game.HiddenCards" || pred.Reads[0].Facet != boardgame.LegalFacetOccupancy {
		t.Fatalf("Reads[0] = %+v", pred.Reads[0])
	}
	// dstPath declares FacetValues, not FacetOccupancy: MayMoveTo's
	// Evaluate calls comp.MayMoveTo(dst), which ends by calling
	// dst.CheckConstraints(...) — a stack constraint (e.g.
	// constraints.Same/constraints.Unique) may read the destination
	// stack's component VALUES, not just its occupancy. See
	// mayMoveConstructor's doc comment in catalog_stack.go for the full
	// rationale and TestMayMoveFacetHonesty below for a dedicated pin.
	if pred.Reads[1].Path != "game.VisibleCards" || pred.Reads[1].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads[1] = %+v", pred.Reads[1])
	}
	if pred.Reads[2].Path != "move.CardIndex" || pred.Reads[2].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads[2] = %+v", pred.Reads[2])
	}

	fixture := buildLegalFixture(t, "memoryDefault")
	if v := pred.Evaluate(fixture.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("HiddenCards->VisibleCards: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	// Fail: component present at source, but MayMoveTo itself rejects
	// (same stack for source and destination).
	sameStackPred := resolvePredicateForTest(t, legal.MayMoveTo("game.HiddenCards", "game.HiddenCards", "move.CardIndex"))
	v := sameStackPred.Evaluate(fixture.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("same-stack: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplateMayNotMoveTo {
		t.Fatalf("same-stack legal.Message = %+v, want template %q", v.Message, legal.TemplateMayNotMoveTo)
	}
	if got := v.Message.Bindings["detail"]; got.S == nil || *got.S == "" {
		t.Fatalf("same-stack detail binding = %+v, want a non-empty error string", got)
	}

	// Fail: no component at the source index (VisibleCards[0] is empty in
	// memoryDefault).
	noComponentPred := resolvePredicateForTest(t, legal.MayMoveTo("game.VisibleCards", "game.HiddenCards", "move.CardIndex"))
	v2 := noComponentPred.Evaluate(fixture.context(0))
	if v2.Outcome != legal.Fail {
		t.Fatalf("no-component: legal.Outcome = %v, want legal.Fail (%+v)", v2.Outcome, v2)
	}
	if v2.Message == nil || v2.Message.Template != legal.TemplateNoComponentToMove {
		t.Fatalf("no-component legal.Message = %+v, want template %q", v2.Message, legal.TemplateNoComponentToMove)
	}

	noMove := buildLegalFixture(t, "memoryNoMove")
	if v := pred.Evaluate(noMove.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("nil move: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}
}

// TestMayMoveToSlot covers MayMoveToSlot's pass/fail/unknown and its
// slot-specific rejection (occupied destination slot), which MayMoveTo
// itself does not check.
func TestMayMoveToSlot(t *testing.T) {
	spec := legal.MayMoveToSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex", "move.CardIndex")
	if spec.Name != "mayMoveToSlot" {
		t.Fatalf("Name = %q, want mayMoveToSlot", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 3 {
		t.Fatalf("Reads = %+v, want 3 entries", pred.Reads)
	}

	fixture := buildLegalFixture(t, "memoryDefault")
	if v := pred.Evaluate(fixture.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("HiddenCards->VisibleCards[0] (empty slot): legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	// Fail: destination slot is occupied (MayMoveTo alone wouldn't catch
	// this — it's MayMoveToSlot's own slot-specific check).
	occupied := buildLegalFixture(t, "memoryVisibleOccupied")
	v := pred.Evaluate(occupied.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("occupied dest slot: legal.Outcome = %v, want legal.Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != legal.TemplateMayNotMoveTo {
		t.Fatalf("occupied dest slot legal.Message = %+v, want template %q", v.Message, legal.TemplateMayNotMoveTo)
	}

	// Distinct source/destination fields are the general case: source index 0
	// remains occupied while player.CardsLeftToReveal selects empty slot 2.
	distinct := legal.MayMoveToSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex", "player.CardsLeftToReveal")
	distinctPred := resolvePredicateForTest(t, distinct)
	if v := distinctPred.Evaluate(occupied.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("distinct destination slot: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}
	if got := legal.MayMoveToSameSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex"); !reflect.DeepEqual(got, spec) {
		t.Fatalf("MayMoveToSameSlot = %+v, want %+v", got, spec)
	}

	noMove := buildLegalFixture(t, "memoryNoMove")
	if v := pred.Evaluate(noMove.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("nil move: legal.Outcome = %v, want legal.Unknown (%+v)", v.Outcome, v)
	}
}

func TestMayMoveCountTo(t *testing.T) {
	spec := legal.MayMoveCountTo("game.HiddenCards", "game.VisibleCards", "player.CardsLeftToReveal")
	if spec.Name != "mayMoveCountTo" {
		t.Fatalf("Name = %q, want mayMoveCountTo", spec.Name)
	}
	pred := resolvePredicateForTest(t, spec)
	if pred.ClientEvaluable {
		t.Fatal("MayMoveCountTo unexpectedly marked client-evaluable")
	}
	if len(pred.Reads) != 3 {
		t.Fatalf("Reads = %+v, want source, destination, and count", pred.Reads)
	}
	if got := pred.RequiredReadTypes["player.CardsLeftToReveal"]; got != boardgame.TypeInt {
		t.Fatalf("count required type = %v, want TypeInt", got)
	}

	fixture := buildLegalFixture(t, "memoryDefault")
	gameStateReader := fixture.state.ImmutableGameState().Reader()
	hidden, err := gameStateReader.ImmutableStackProp("HiddenCards")
	if err != nil {
		t.Fatal("read HiddenCards:", err)
	}
	visible, err := gameStateReader.ImmutableStackProp("VisibleCards")
	if err != nil {
		t.Fatal("read VisibleCards:", err)
	}
	beforeHidden, beforeVisible := hidden.NumComponents(), visible.NumComponents()
	if v := pred.Evaluate(fixture.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("two-component transfer: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}
	if hidden.NumComponents() != beforeHidden || visible.NumComponents() != beforeVisible {
		t.Fatalf("predicate mutated stacks: hidden %d→%d, visible %d→%d", beforeHidden, hidden.NumComponents(), beforeVisible, visible.NumComponents())
	}

	game, state := newMemoryGame(t)
	zero := legalFixture{state: state, move: memoryMoveWithCardIndex(t, game, 0), chest: game.Manager().Chest()}
	zeroPred := resolvePredicateForTest(t, legal.MayMoveCountTo("game.HiddenCards", "game.VisibleCards", "move.CardIndex"))
	if v := zeroPred.Evaluate(zero.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("zero count verdict = %+v, want Pass", v)
	}

	sameStack := resolvePredicateForTest(t, legal.MayMoveCountTo("game.HiddenCards", "game.HiddenCards", "player.CardsLeftToReveal"))
	if v := sameStack.Evaluate(fixture.context(0)); v.Outcome != legal.Fail || v.Message == nil || v.Message.Template != legal.TemplateMayNotMoveCountTo {
		t.Fatalf("same-stack verdict = %+v, want %q failure", v, legal.TemplateMayNotMoveCountTo)
	}

	negative := legalFixture{state: state, move: memoryMoveWithCardIndex(t, game, -1), chest: game.Manager().Chest()}
	negativePred := resolvePredicateForTest(t, legal.MayMoveCountTo("game.HiddenCards", "game.VisibleCards", "move.CardIndex"))
	if v := negativePred.Evaluate(negative.context(0)); v.Outcome != legal.Fail {
		t.Fatalf("negative count verdict = %+v, want Fail", v)
	}

	noMove := buildLegalFixture(t, "memoryNoMove")
	if v := negativePred.Evaluate(noMove.context(0)); v.Outcome != legal.Unknown {
		t.Fatalf("missing count move verdict = %+v, want Unknown", v)
	}

	game.Manager().Internals().AllowMutableConstraints(game)
	visibleMutable, err := state.GameState().ReadSetter().StackProp("VisibleCards")
	if err != nil {
		t.Fatal("read mutable VisibleCards:", err)
	}
	if err := visibleMutable.AddConstraint(func(boardgame.ImmutableStack, []boardgame.ImmutableComponentInstance, boardgame.ImmutableState) error {
		return errors.New("declarative constraint rejection")
	}); err != nil {
		t.Fatal("add constraint:", err)
	}
	constraintFixture := legalFixture{state: state, move: memoryMoveWithCardIndex(t, game, 1), chest: game.Manager().Chest()}
	hiddenMutable, err := state.GameState().ReadSetter().StackProp("HiddenCards")
	if err != nil {
		t.Fatal("read mutable HiddenCards:", err)
	}
	beforeHidden = hiddenMutable.NumComponents()
	beforeVisible = visibleMutable.NumComponents()
	if v := negativePred.Evaluate(constraintFixture.context(0)); v.Outcome != legal.Fail {
		t.Fatalf("constraint verdict = %+v, want Fail", v)
	}
	if hiddenMutable.NumComponents() != beforeHidden || visibleMutable.NumComponents() != beforeVisible {
		t.Fatal("constraint evaluation mutated live stacks")
	}
}

func TestMayMoveFixedCountTo(t *testing.T) {
	spec := legal.MayMoveFixedCountTo("game.HiddenCards", "game.VisibleCards", 2)
	if spec.Name != "mayMoveFixedCountTo" {
		t.Fatalf("Name = %q, want mayMoveFixedCountTo", spec.Name)
	}
	pred := resolvePredicateForTest(t, spec)
	if pred.ClientEvaluable {
		t.Fatal("MayMoveFixedCountTo unexpectedly marked client-evaluable")
	}
	if len(pred.Reads) != 2 {
		t.Fatalf("Reads = %+v, want only source and destination", pred.Reads)
	}
	if v := pred.Evaluate(buildLegalFixture(t, "memoryDefault").context(0)); v.Outcome != legal.Pass {
		t.Fatalf("fixed two-component transfer = %+v, want Pass", v)
	}
	zero := resolvePredicateForTest(t, legal.MayMoveFixedCountTo("game.HiddenCards", "game.VisibleCards", 0))
	if v := zero.Evaluate(buildLegalFixture(t, "memoryDefault").context(0)); v.Outcome != legal.Pass {
		t.Fatalf("fixed zero-component transfer = %+v, want Pass", v)
	}
	overCapacity := resolvePredicateForTest(t, legal.MayMoveFixedCountTo("game.HiddenCards", "game.VisibleCards", 41))
	if v := overCapacity.Evaluate(buildLegalFixture(t, "memoryDefault").context(0)); v.Outcome != legal.Fail {
		t.Fatalf("over-capacity fixed transfer = %+v, want Fail", v)
	}

	if _, err := resolveSpecViaRegistry(legal.MayMoveFixedCountTo("game.HiddenCards", "game.VisibleCards", -1), legal.DefaultConstructors(), nil); err == nil {
		t.Fatal("negative fixed count unexpectedly constructed")
	}

	malformed := legal.Spec{Name: "mayMoveFixedCountTo", Args: []string{"game.HiddenCards", "game.VisibleCards", "not-an-int"}}
	if _, err := resolveSpecViaRegistry(malformed, legal.DefaultConstructors(), nil); err == nil {
		t.Fatal("malformed fixed count unexpectedly constructed")
	}
}

// findRead returns the Read in reads whose Path matches path, and whether
// one was found.
func findRead(reads []legal.Read, path string) (legal.Read, bool) {
	for _, r := range reads {
		if string(r.Path) == path {
			return r, true
		}
	}
	return legal.Read{}, false
}

// TestMayMoveFacetHonesty pins the facet-honesty fix for MayMoveTo/
// MayMoveToSlot (see mayMoveConstructor's doc comment in catalog_stack.go):
// both predicates declare LegalFacetValues, never LegalFacetOccupancy, on
// dstPath. This is unconditional — the declaration doesn't change based on
// whether the destination stack actually carries a constraint at
// construction time — because Evaluate ends by calling
// comp.MayMoveTo/MayMoveToSlot (component.go), which itself ends by calling
// dst.CheckConstraints(...), and a constraint such as constraints.Same or
// constraints.Unique is free to read dst's component VALUES, not just its
// occupancy.
func TestMayMoveFacetHonesty(t *testing.T) {
	// Pin the declaration for a destination that, at construction time,
	// carries no constraint at all — the common case, and the one
	// LegalFacetOccupancy would have been tempting to declare.
	for _, name := range []string{"mayMoveTo", "mayMoveToSlot"} {
		args := []string{"game.HiddenCards", "game.VisibleCards", "move.CardIndex"}
		if name == "mayMoveToSlot" {
			args = append(args, "move.CardIndex")
		}
		spec := legal.Spec{Name: name, Args: args}
		pred := resolvePredicateForTest(t, spec)
		dst, ok := findRead(pred.Reads, "game.VisibleCards")
		if !ok {
			t.Fatalf("%s: no legal.Read found for dstPath game.VisibleCards", name)
		}
		if dst.Facet != boardgame.LegalFacetValues {
			t.Fatalf("%s: dstPath legal.Facet = %v, want LegalFacetValues", name, dst.Facet)
		}
	}

	// Prove the declaration isn't just conservative paperwork: attach a
	// values-reading constraint (constraints.Same) to VisibleCards after
	// construction, then show mayMoveTo's verdict genuinely depends on
	// VisibleCards' component VALUES, not just its occupancy — exactly the
	// information LegalFacetOccupancy would have told a client evaluability
	// check was safe to hide.
	game, state := newMemoryGame(t)
	game.Manager().Internals().AllowMutableConstraints(game)

	gameRS := state.GameState().ReadSetter()
	hidden, err := gameRS.StackProp("HiddenCards")
	if err != nil {
		t.Fatalf("reading HiddenCards: %v", err)
	}
	visible, err := gameRS.StackProp("VisibleCards")
	if err != nil {
		t.Fatalf("reading VisibleCards: %v", err)
	}

	// Establish VisibleCards[1]'s Type as the "same" baseline by moving a
	// specific hidden card there directly (bypassing MayMoveTo, which is
	// under test below).
	established := hidden.ComponentAt(hidden.Len() - 1)
	if established == nil {
		t.Fatal("expected HiddenCards' last slot to be occupied")
	}
	establishedType, err := established.Values().Reader().StringProp("Type")
	if err != nil {
		t.Fatalf("reading established card's Type: %v", err)
	}
	if err := established.MoveTo(visible, 1); err != nil {
		t.Fatalf("establishing VisibleCards[1]: %v", err)
	}

	visible.AddConstraint(constraints.Same("Type"))

	// Find a still-hidden card whose Type MATCHES establishedType (memory
	// pairs every Type, so the established card's pair partner is still in
	// HiddenCards) and one whose Type DIFFERS.
	matchIdx, mismatchIdx := -1, -1
	for i := 0; i < hidden.Len(); i++ {
		c := hidden.ComponentAt(i)
		if c == nil {
			continue
		}
		typ, err := c.Values().Reader().StringProp("Type")
		if err != nil {
			t.Fatalf("reading Type at HiddenCards[%d]: %v", i, err)
		}
		if typ == establishedType && matchIdx < 0 {
			matchIdx = i
		}
		if typ != establishedType && mismatchIdx < 0 {
			mismatchIdx = i
		}
	}
	if matchIdx < 0 || mismatchIdx < 0 {
		t.Fatal("expected both a Type-matching and a Type-mismatching card left in HiddenCards")
	}

	pred := resolvePredicateForTest(t, legal.MayMoveTo("game.HiddenCards", "game.VisibleCards", "move.CardIndex"))

	matchMove := memoryMoveWithCardIndex(t, game, matchIdx)
	matchFixture := legalFixture{state: state, move: matchMove, chest: game.Manager().Chest()}
	if v := pred.Evaluate(matchFixture.context(0)); v.Outcome != legal.Pass {
		t.Fatalf("matching Type: legal.Outcome = %v, want legal.Pass (%+v)", v.Outcome, v)
	}

	mismatchMove := memoryMoveWithCardIndex(t, game, mismatchIdx)
	mismatchFixture := legalFixture{state: state, move: mismatchMove, chest: game.Manager().Chest()}
	v := pred.Evaluate(mismatchFixture.context(0))
	if v.Outcome != legal.Fail {
		t.Fatalf("mismatching Type: legal.Outcome = %v, want legal.Fail (%+v) — this values-dependence is exactly what the FacetValues declaration on dstPath exists to make honest", v.Outcome, v)
	}
}

// TestDefaultConstructors verifies DefaultConstructors returns exactly the
// catalog built in this task.
func TestDefaultConstructors(t *testing.T) {
	constructors := legal.DefaultConstructors()
	wantNames := map[string]bool{
		"propAtLeast":                      true,
		"propCompare":                      true,
		"playerBool":                       true,
		"playerBoolAt":                     true,
		"componentPresentAt":               true,
		"componentAbsentAt":                true,
		"componentPresentAtKey":            true,
		"mayMoveTo":                        true,
		"mayMoveToSlot":                    true,
		"mayMoveAllTo":                     true,
		"mayMoveCountTo":                   true,
		"mayMoveFixedCountTo":              true,
		"maySwapComponents":                true,
		"maySwapComponentsByKey":           true,
		"allActivePlayers":                 true,
		"proposerIsCurrentPlayer":          true,
		"proposerIsPlayerFromMove":         true,
		"revealableCardAt":                 true,
		"componentPropEqualsCurrentPlayer": true,
		"inPhase":                          true,
		"stackConstraints":                 true,
		"stackCount":                       true,
		"stackEmpty":                       true,
		"stackNotEmpty":                    true,
		"propEquals":                       true,
		"propNotEquals":                    true,
	}
	if len(constructors) != len(wantNames) {
		t.Fatalf("len(legal.DefaultConstructors()) = %d, want %d", len(constructors), len(wantNames))
	}
	for _, c := range constructors {
		if !wantNames[c.Name] {
			t.Errorf("unexpected constructor name %q", c.Name)
		}
		delete(wantNames, c.Name)
	}
	if len(wantNames) != 0 {
		t.Errorf("missing constructor names: %v", wantNames)
	}
}

// TestDefaultTemplateKeysCoversAllTemplates pins defaultTemplateKeys
// (catalog_stack.go's handoff list for Task 6's legal.DefaultTemplates())
// against the `want` list below. Despite the name, this does NOT
// independently verify that every TemplateXxx constant defined in this
// package is present: `want` is a second, hand-maintained copy of the same
// set catalog_stack.go's authors had to remember to update, not something
// derived by scanning the package for TemplateXxx declarations. So it
// guards against defaultTemplateKeys and this test's own `want` drifting
// out of sync with each other (e.g. a typo, or one being edited without the
// other) — it does not guard against BOTH lists omitting a template key a
// new predicate introduces. A future predicate constructor's author must
// still remember to add its template constant to defaultTemplateKeys AND
// here by hand; nothing enforces that automatically today.
func TestDefaultTemplateKeysCoversAllTemplates(t *testing.T) {
	want := []string{
		legal.TemplatePropAtLeast,
		legal.TemplatePropCompare,
		legal.TemplatePlayerBool,
		legal.TemplatePlayerAlreadySubmitted,
		legal.TemplatePlayerNotSubmitted,
		legal.TemplatePlayerInactive,
		legal.TemplatePlayerActive,
		legal.TemplateSeatNotFilled,
		legal.TemplateSeatNotClosed,
		legal.TemplatePlayerNotAdmin,
		legal.TemplateComponentMissing,
		legal.TemplateComponentMissingKey,
		legal.TemplateNoComponentToMove,
		legal.TemplateMayNotMoveTo,
		legal.TemplateMayNotMoveAllTo,
		legal.TemplateMayNotMoveCountTo,
		legal.TemplateMayNotSwapComponents,
		legal.TemplateAllActivePlayers,
		legal.TemplateProposerTargetInvalid,
		legal.TemplateProposerNotYourTurn,
		legal.TemplateNoCardHere,
		legal.TemplateAlreadyRevealed,
		legal.TemplateComponentPropNotCurrentPlayer,
		legal.TemplateInPhase,
		legal.TemplateInProgression,
		legal.TemplateStackConstraints,
		legal.TemplateStackCount,
		legal.TemplateStackEmpty,
		legal.TemplateStackNotEmpty,
		legal.TemplatePropEquals,
		legal.TemplatePropNotEquals,
		legal.TemplateComponentPresentUnexpected,
	}
	keys := legal.DefaultTemplateKeys()
	if len(keys) != len(want) {
		t.Fatalf("len(legal.DefaultTemplateKeys()) = %d, want %d", len(keys), len(want))
	}
	got := make(map[string]bool)
	for _, k := range keys {
		got[k] = true
	}
	for _, k := range want {
		if !got[k] {
			t.Errorf("legal.DefaultTemplateKeys() is missing %q", k)
		}
	}
}
