package legal

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/constraints"
)

// TestPropAtLeast covers PropAtLeast's pass/fail/unknown paths, its
// declared Reads, and the template key used on Fail.
func TestPropAtLeast(t *testing.T) {
	spec := PropAtLeast("player.CardsLeftToReveal", 1)
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
	if v := pred.Evaluate(pass.context(0)); v.Outcome != Pass {
		t.Fatalf("pass fixture: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	// Fail: memoryZeroCardsLeft has CardsLeftToReveal == 0 < 1.
	fail := buildLegalFixture(t, "memoryZeroCardsLeft")
	v := pred.Evaluate(fail.context(0))
	if v.Outcome != Fail {
		t.Fatalf("fail fixture: Outcome = %v, want Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != TemplatePropAtLeast {
		t.Fatalf("fail fixture: Message = %+v, want template %q", v.Message, TemplatePropAtLeast)
	}
	if got := v.Message.Bindings["value"]; got.I == nil || *got.I != 0 {
		t.Fatalf("fail fixture: value binding = %+v, want 0", got)
	}
	if got := v.Message.Bindings["min"]; got.I == nil || *got.I != 1 {
		t.Fatalf("fail fixture: min binding = %+v, want 1", got)
	}

	// Unknown: property doesn't exist.
	unknownSpec := PropAtLeast("game.NoSuchIntProp", 1)
	unknownPred := resolvePredicateForTest(t, unknownSpec)
	if v := unknownPred.Evaluate(pass.context(0)); v.Outcome != Unknown {
		t.Fatalf("unknown-path: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}

	// Custom message overrides the default template.
	overridden := resolvePredicateForTest(t, PropAtLeast("player.CardsLeftToReveal", 1).WithMessage("custom.key"))
	if v := overridden.Evaluate(fail.context(0)); v.Message == nil || v.Message.Template != "custom.key" {
		t.Fatalf("WithMessage override: Message = %+v, want template custom.key", v.Message)
	}
}

// TestPropCompare covers PropCompare's ops, pass/fail/unknown, Reads, and
// bad-op construction-time validation.
func TestPropCompare(t *testing.T) {
	spec := PropCompare("player.CardsLeftToReveal", "==", 2)
	if spec.Name != "propCompare" {
		t.Fatalf("Name = %q, want propCompare", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 1 || pred.Reads[0].Path != "player.CardsLeftToReveal" || pred.Reads[0].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads = %+v", pred.Reads)
	}

	pass := buildLegalFixture(t, "memoryDefault")
	if v := pred.Evaluate(pass.context(0)); v.Outcome != Pass {
		t.Fatalf("pass fixture: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	fail := buildLegalFixture(t, "memoryZeroCardsLeft")
	v := pred.Evaluate(fail.context(0))
	if v.Outcome != Fail {
		t.Fatalf("fail fixture: Outcome = %v, want Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != TemplatePropCompare {
		t.Fatalf("fail fixture: Message = %+v, want template %q", v.Message, TemplatePropCompare)
	}

	unknownPred := resolvePredicateForTest(t, PropCompare("game.NoSuchIntProp", "==", 0))
	if v := unknownPred.Evaluate(pass.context(0)); v.Outcome != Unknown {
		t.Fatalf("unknown-path: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}

	// Every operator exercised against a known value (CardsLeftToReveal == 2).
	opCases := []struct {
		op   string
		n    int
		want Outcome
	}{
		{"==", 2, Pass},
		{"==", 3, Fail},
		{"!=", 3, Pass},
		{"!=", 2, Fail},
		{"<", 3, Pass},
		{"<", 2, Fail},
		{"<=", 2, Pass},
		{"<=", 1, Fail},
		{">", 1, Pass},
		{">", 2, Fail},
		{">=", 2, Pass},
		{">=", 3, Fail},
	}
	for _, oc := range opCases {
		p := resolvePredicateForTest(t, PropCompare("player.CardsLeftToReveal", oc.op, oc.n))
		if v := p.Evaluate(pass.context(0)); v.Outcome != oc.want {
			t.Errorf("op %s %d: Outcome = %v, want %v", oc.op, oc.n, v.Outcome, oc.want)
		}
	}

	// Bad op is a construction-time error, not a runtime Unknown.
	if _, err := resolveSpecViaRegistry(PropCompare("game.NumCards", "~=", 1), DefaultConstructors(), nil); err == nil {
		t.Fatal("expected an error constructing propCompare with an invalid op")
	}
}

// TestPlayerBool covers PlayerBool's pass/fail/unknown, Reads, and template
// key.
func TestPlayerBool(t *testing.T) {
	spec := PlayerBool("SeatFilled")
	if spec.Name != "playerBool" {
		t.Fatalf("Name = %q, want playerBool", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 1 || pred.Reads[0].Path != "player.SeatFilled" || pred.Reads[0].Facet != boardgame.LegalFacetValues {
		t.Fatalf("Reads = %+v", pred.Reads)
	}

	pass := buildLegalFixture(t, "memorySeatFilled")
	if v := pred.Evaluate(pass.context(0)); v.Outcome != Pass {
		t.Fatalf("pass fixture: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	fail := buildLegalFixture(t, "memoryDefault")
	v := pred.Evaluate(fail.context(0))
	if v.Outcome != Fail {
		t.Fatalf("fail fixture: Outcome = %v, want Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != TemplatePlayerBool {
		t.Fatalf("fail fixture: Message = %+v, want template %q", v.Message, TemplatePlayerBool)
	}
	if got := v.Message.Bindings["prop"]; got.S == nil || *got.S != "SeatFilled" {
		t.Fatalf("fail fixture: prop binding = %+v, want SeatFilled", got)
	}

	unknownPred := resolvePredicateForTest(t, PlayerBool("NoSuchBoolProp"))
	if v := unknownPred.Evaluate(fail.context(0)); v.Outcome != Unknown {
		t.Fatalf("unknown-prop: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}
}

// TestComponentPresentAt covers ComponentPresentAt's pass/fail/unknown,
// Reads (occupancy facet on the stack, values facet on the index), and
// template key.
func TestComponentPresentAt(t *testing.T) {
	spec := ComponentPresentAt("game.HiddenCards", "move.CardIndex")
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
	if v := pred.Evaluate(fixture.context(0)); v.Outcome != Pass {
		t.Fatalf("HiddenCards[0] present: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	failSpec := ComponentPresentAt("game.VisibleCards", "move.CardIndex")
	failPred := resolvePredicateForTest(t, failSpec)
	v := failPred.Evaluate(fixture.context(0))
	if v.Outcome != Fail {
		t.Fatalf("VisibleCards[0] absent: Outcome = %v, want Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != TemplateComponentMissing {
		t.Fatalf("Message = %+v, want template %q", v.Message, TemplateComponentMissing)
	}
	if got := v.Message.Bindings["index"]; got.I == nil || *got.I != 0 {
		t.Fatalf("index binding = %+v, want 0", got)
	}

	noMove := buildLegalFixture(t, "memoryNoMove")
	if v := pred.Evaluate(noMove.context(0)); v.Outcome != Unknown {
		t.Fatalf("nil move: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}
}

// TestComponentPresentAtKey covers ComponentPresentAtKey's pass/fail/unknown,
// Reads, and template key, using checkers' enum-keyed Spaces stack.
func TestComponentPresentAtKey(t *testing.T) {
	spec := ComponentPresentAtKey("game.Spaces", "move.TokenIndexToMove")
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
	if v := pred.Evaluate(occupied.context(0)); v.Outcome != Pass {
		t.Fatalf("occupied space: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	empty := buildLegalFixture(t, "checkersEmptySpace")
	v := pred.Evaluate(empty.context(0))
	if v.Outcome != Fail {
		t.Fatalf("empty space: Outcome = %v, want Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != TemplateComponentMissingKey {
		t.Fatalf("Message = %+v, want template %q", v.Message, TemplateComponentMissingKey)
	}

	noMove := buildLegalFixture(t, "checkersNoMove")
	if v := pred.Evaluate(noMove.context(0)); v.Outcome != Unknown {
		t.Fatalf("nil move: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}
}

// TestMayMoveTo covers MayMoveTo's pass/fail(x2 templates)/unknown, Reads,
// and template keys.
func TestMayMoveTo(t *testing.T) {
	spec := MayMoveTo("game.HiddenCards", "game.VisibleCards", "move.CardIndex")
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
	if v := pred.Evaluate(fixture.context(0)); v.Outcome != Pass {
		t.Fatalf("HiddenCards->VisibleCards: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	// Fail: component present at source, but MayMoveTo itself rejects
	// (same stack for source and destination).
	sameStackPred := resolvePredicateForTest(t, MayMoveTo("game.HiddenCards", "game.HiddenCards", "move.CardIndex"))
	v := sameStackPred.Evaluate(fixture.context(0))
	if v.Outcome != Fail {
		t.Fatalf("same-stack: Outcome = %v, want Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != TemplateMayNotMoveTo {
		t.Fatalf("same-stack Message = %+v, want template %q", v.Message, TemplateMayNotMoveTo)
	}
	if got := v.Message.Bindings["detail"]; got.S == nil || *got.S == "" {
		t.Fatalf("same-stack detail binding = %+v, want a non-empty error string", got)
	}

	// Fail: no component at the source index (VisibleCards[0] is empty in
	// memoryDefault).
	noComponentPred := resolvePredicateForTest(t, MayMoveTo("game.VisibleCards", "game.HiddenCards", "move.CardIndex"))
	v2 := noComponentPred.Evaluate(fixture.context(0))
	if v2.Outcome != Fail {
		t.Fatalf("no-component: Outcome = %v, want Fail (%+v)", v2.Outcome, v2)
	}
	if v2.Message == nil || v2.Message.Template != TemplateNoComponentToMove {
		t.Fatalf("no-component Message = %+v, want template %q", v2.Message, TemplateNoComponentToMove)
	}

	noMove := buildLegalFixture(t, "memoryNoMove")
	if v := pred.Evaluate(noMove.context(0)); v.Outcome != Unknown {
		t.Fatalf("nil move: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}
}

// TestMayMoveToSlot covers MayMoveToSlot's pass/fail/unknown and its
// slot-specific rejection (occupied destination slot), which MayMoveTo
// itself does not check.
func TestMayMoveToSlot(t *testing.T) {
	spec := MayMoveToSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex")
	if spec.Name != "mayMoveToSlot" {
		t.Fatalf("Name = %q, want mayMoveToSlot", spec.Name)
	}

	pred := resolvePredicateForTest(t, spec)
	if len(pred.Reads) != 3 {
		t.Fatalf("Reads = %+v, want 3 entries", pred.Reads)
	}

	fixture := buildLegalFixture(t, "memoryDefault")
	if v := pred.Evaluate(fixture.context(0)); v.Outcome != Pass {
		t.Fatalf("HiddenCards->VisibleCards[0] (empty slot): Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	// Fail: destination slot is occupied (MayMoveTo alone wouldn't catch
	// this — it's MayMoveToSlot's own slot-specific check).
	occupied := buildLegalFixture(t, "memoryVisibleOccupied")
	v := pred.Evaluate(occupied.context(0))
	if v.Outcome != Fail {
		t.Fatalf("occupied dest slot: Outcome = %v, want Fail (%+v)", v.Outcome, v)
	}
	if v.Message == nil || v.Message.Template != TemplateMayNotMoveTo {
		t.Fatalf("occupied dest slot Message = %+v, want template %q", v.Message, TemplateMayNotMoveTo)
	}

	noMove := buildLegalFixture(t, "memoryNoMove")
	if v := pred.Evaluate(noMove.context(0)); v.Outcome != Unknown {
		t.Fatalf("nil move: Outcome = %v, want Unknown (%+v)", v.Outcome, v)
	}
}

// findRead returns the Read in reads whose Path matches path, and whether
// one was found.
func findRead(reads []Read, path string) (Read, bool) {
	for _, r := range reads {
		if string(r.Path) == path {
			return r, true
		}
	}
	return Read{}, false
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
		spec := Spec{Name: name, Args: []string{"game.HiddenCards", "game.VisibleCards", "move.CardIndex"}}
		pred := resolvePredicateForTest(t, spec)
		dst, ok := findRead(pred.Reads, "game.VisibleCards")
		if !ok {
			t.Fatalf("%s: no Read found for dstPath game.VisibleCards", name)
		}
		if dst.Facet != boardgame.LegalFacetValues {
			t.Fatalf("%s: dstPath Facet = %v, want LegalFacetValues", name, dst.Facet)
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

	pred := resolvePredicateForTest(t, MayMoveTo("game.HiddenCards", "game.VisibleCards", "move.CardIndex"))

	matchMove := memoryMoveWithCardIndex(t, game, matchIdx)
	matchFixture := legalFixture{state: state, move: matchMove, chest: game.Manager().Chest()}
	if v := pred.Evaluate(matchFixture.context(0)); v.Outcome != Pass {
		t.Fatalf("matching Type: Outcome = %v, want Pass (%+v)", v.Outcome, v)
	}

	mismatchMove := memoryMoveWithCardIndex(t, game, mismatchIdx)
	mismatchFixture := legalFixture{state: state, move: mismatchMove, chest: game.Manager().Chest()}
	v := pred.Evaluate(mismatchFixture.context(0))
	if v.Outcome != Fail {
		t.Fatalf("mismatching Type: Outcome = %v, want Fail (%+v) — this values-dependence is exactly what the FacetValues declaration on dstPath exists to make honest", v.Outcome, v)
	}
}

// TestDefaultConstructors verifies DefaultConstructors returns exactly the
// catalog built in this task, and that ExtendDefaults appends without
// mutating DefaultConstructors' own result.
func TestDefaultConstructors(t *testing.T) {
	constructors := DefaultConstructors()
	wantNames := map[string]bool{
		"propAtLeast":           true,
		"propCompare":           true,
		"playerBool":            true,
		"componentPresentAt":    true,
		"componentPresentAtKey": true,
		"mayMoveTo":             true,
		"mayMoveToSlot":         true,
	}
	if len(constructors) != len(wantNames) {
		t.Fatalf("len(DefaultConstructors()) = %d, want %d", len(constructors), len(wantNames))
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

func TestExtendDefaults(t *testing.T) {
	custom := &PredicateConstructor{
		Name: "custom",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			return nil, nil
		},
	}

	extended := ExtendDefaults(custom)
	defaultLen := len(DefaultConstructors())
	if len(extended) != defaultLen+1 {
		t.Fatalf("len(extended) = %d, want %d", len(extended), defaultLen+1)
	}
	if extended[len(extended)-1].Name != "custom" {
		t.Fatalf("last constructor = %q, want custom", extended[len(extended)-1].Name)
	}
	if len(DefaultConstructors()) != defaultLen {
		t.Fatalf("DefaultConstructors() was mutated by ExtendDefaults")
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
		TemplatePropAtLeast,
		TemplatePropCompare,
		TemplatePlayerBool,
		TemplateComponentMissing,
		TemplateComponentMissingKey,
		TemplateNoComponentToMove,
		TemplateMayNotMoveTo,
	}
	if len(defaultTemplateKeys) != len(want) {
		t.Fatalf("len(defaultTemplateKeys) = %d, want %d", len(defaultTemplateKeys), len(want))
	}
	got := make(map[string]bool)
	for _, k := range defaultTemplateKeys {
		got[k] = true
	}
	for _, k := range want {
		if !got[k] {
			t.Errorf("defaultTemplateKeys is missing %q", k)
		}
	}
}
