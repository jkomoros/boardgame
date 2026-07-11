package legal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jkomoros/boardgame"
)

// TestDefaultTemplatesCoversDefaultTemplateKeys is a sanity check that
// DefaultTemplates() has an entry for every key declared in
// defaultTemplateKeys (legal/catalog_stack.go — the hand-maintained list
// every catalog file appends its own template keys to as Tasks land). This
// is NOT the primary proof of coverage (a hand-list check could pass with
// both lists wrong in the same way); see
// TestDefaultTemplatesCoversCorpusFailCases below for that.
func TestDefaultTemplatesCoversDefaultTemplateKeys(t *testing.T) {
	table := DefaultTemplates()
	for _, key := range defaultTemplateKeys {
		if _, ok := table[key]; !ok {
			t.Errorf("DefaultTemplates() is missing an entry for %q (listed in defaultTemplateKeys)", key)
		}
	}
}

// TestDefaultTemplatesCoversCorpusFailCases is the primary coverage proof
// the task brief requires: rather than trusting a hand-maintained list, it
// evaluates every case in every testdata/conformance/*.json file (the same
// corpus TestConformanceCorpus in conformance_test.go exercises), and for
// every case whose Verdict.Outcome comes back Fail, asserts the ACTUAL
// emitted Verdict.Message.Template is a key DefaultTemplates() covers. This
// catches a real gap even if defaultTemplateKeys and DefaultTemplates()
// were both wrong in the same way, because it never consults
// defaultTemplateKeys at all — it only looks at what predicates actually
// emit when evaluated.
func TestDefaultTemplatesCoversCorpusFailCases(t *testing.T) {
	table := DefaultTemplates()

	paths, err := filepath.Glob("testdata/conformance/*.json")
	if err != nil {
		t.Fatalf("legal: globbing conformance corpus: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("legal: no conformance corpus files found under testdata/conformance/")
	}

	checked := 0
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("legal: reading %s: %v", path, err)
		}
		var cf conformanceFile
		if err := json.Unmarshal(data, &cf); err != nil {
			t.Fatalf("legal: parsing %s: %v", path, err)
		}
		for i, c := range cf.Cases {
			if c.Expect != "fail" {
				continue
			}
			fixture := buildLegalFixture(t, c.Fixture)
			pred := resolvePredicateForTest(t, c.Spec)
			verdict := pred.Evaluate(fixture.context(boardgame.PlayerIndex(c.Proposer)))
			if verdict.Outcome != Fail {
				t.Fatalf("%s case %d (%s): expected fail per corpus, evaluator returned %v", path, i, c.Spec.Name, verdict.Outcome)
			}
			if verdict.Message == nil {
				t.Fatalf("%s case %d (%s): Fail verdict has no Message", path, i, c.Spec.Name)
			}
			if _, ok := table[verdict.Message.Template]; !ok {
				t.Errorf("%s case %d (%s): emitted template key %q is not covered by DefaultTemplates()", path, i, c.Spec.Name, verdict.Message.Template)
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("legal: no corpus fail-cases were evaluated — the corpus or this test's fixture-building is broken")
	}
	t.Logf("legal: checked %d corpus fail-case emitted template keys against DefaultTemplates()", checked)
}

// TestDefaultTemplatesReturnsCopy verifies a caller mutating the returned
// map cannot corrupt the package's own default table or a subsequent call's
// result.
func TestDefaultTemplatesReturnsCopy(t *testing.T) {
	first := DefaultTemplates()
	first["injected.key"] = "should not leak"
	second := DefaultTemplates()
	if _, ok := second["injected.key"]; ok {
		t.Fatal("DefaultTemplates() shares mutable state across calls")
	}
}

// TestErrorfRoundTripsThroughErrorsAs verifies Errorf's result implements
// error, is retrievable via errors.As as a *boardgame.LegalError, and
// carries the given template key and bindings.
func TestErrorfRoundTripsThroughErrorsAs(t *testing.T) {
	err := Errorf("checkers.illegal_dest", map[string]boardgame.LegalBindingValue{
		"from": String("A1"),
	})
	if err == nil {
		t.Fatal("Errorf(...) = nil, want non-nil error")
	}

	var target *boardgame.LegalError
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, want true (err: %v, %T)", err, err)
	}
	if target.Verdict.Outcome != Fail {
		t.Fatalf("target.Verdict.Outcome = %v, want Fail", target.Verdict.Outcome)
	}
	if target.Verdict.Message == nil || target.Verdict.Message.Template != "checkers.illegal_dest" {
		t.Fatalf("target.Verdict.Message = %+v, want Template %q", target.Verdict.Message, "checkers.illegal_dest")
	}
	if got := target.Verdict.Message.Bindings["from"]; got.S == nil || *got.S != "A1" {
		t.Fatalf("target.Verdict.Message.Bindings[\"from\"] = %+v, want S=%q", got, "A1")
	}
}

// TestErrorfNilBindings verifies Errorf accepts a nil bindings map for a
// template with no placeholders.
func TestErrorfNilBindings(t *testing.T) {
	err := Errorf("some.key", nil)
	if err == nil {
		t.Fatal("Errorf(\"some.key\", nil) = nil, want non-nil error")
	}
}

// TestRenderMessageMissingBindingRendersPlaceholderName verifies
// RenderMessage (the package legal wrapper) has the same missing-binding
// behavior as core's RenderLegalMessage: the bare placeholder name, never a
// panic.
func TestRenderMessageMissingBindingRendersPlaceholderName(t *testing.T) {
	m := &Message{Template: "some.key"}
	table := map[string]string{"some.key": "needs {value}"}
	if got := RenderMessage(m, table); got != "needs value" {
		t.Fatalf("RenderMessage = %q, want %q", got, "needs value")
	}
}

// TestRenderMessageNilSafe verifies RenderMessage(nil, ...) never panics.
func TestRenderMessageNilSafe(t *testing.T) {
	if got := RenderMessage(nil, DefaultTemplates()); got != "" {
		t.Fatalf("RenderMessage(nil, ...) = %q, want \"\"", got)
	}
}

// TestProposerTemplateRenderingParity is the byte-for-byte string-parity
// test the task brief requires: rendering ProposerIsCurrentPlayer's Fail
// Verdicts through DefaultTemplates() must reproduce the EXACT legacy
// strings from moves/current_player.go, since un-migrated games' clients
// depend on seeing those exact strings (design spec "prime guarantee").
func TestProposerTemplateRenderingParity(t *testing.T) {
	table := DefaultTemplates()
	pred := resolvePredicateForTest(t, ProposerIsCurrentPlayer())

	targetMismatch := buildLegalFixture(t, "memoryTargetPlayerOne")
	v := pred.Evaluate(targetMismatch.context(0))
	if v.Outcome != Fail {
		t.Fatalf("memoryTargetPlayerOne: Outcome = %v, want Fail", v.Outcome)
	}
	if got, want := RenderMessage(v.Message, table), "it's not your turn"; got != want {
		t.Fatalf("rendered = %q, want the verbatim legacy string %q", got, want)
	}

	targetObserver := buildLegalFixture(t, "memoryTargetObserver")
	v2 := pred.Evaluate(targetObserver.context(0))
	if v2.Outcome != Fail {
		t.Fatalf("memoryTargetObserver: Outcome = %v, want Fail", v2.Outcome)
	}
	if got, want := RenderMessage(v2.Message, table), "The specified target player is not valid"; got != want {
		t.Fatalf("rendered = %q, want the verbatim legacy string %q", got, want)
	}
}

// TestRevealableCardAtTemplateRenderingParity pins the other verbatim-string
// migration acid test from the design spec §8: RevealableCardAt's two Fail
// branches must render exactly examples/memory/moves.go's legacy strings
// through DefaultTemplates().
func TestRevealableCardAtTemplateRenderingParity(t *testing.T) {
	table := DefaultTemplates()
	pred := resolvePredicateForTest(t, RevealableCardAt("game.HiddenCards", "game.VisibleCards", "move.CardIndex"))

	neverThere := buildLegalFixture(t, "memoryCardNeverThere")
	v := pred.Evaluate(neverThere.context(0))
	if v.Outcome != Fail {
		t.Fatalf("memoryCardNeverThere: Outcome = %v, want Fail", v.Outcome)
	}
	if got, want := RenderMessage(v.Message, table), "there is no card at that index"; got != want {
		t.Fatalf("rendered = %q, want %q", got, want)
	}

	alreadyRevealed := buildLegalFixture(t, "memoryCardAlreadyRevealed")
	v2 := pred.Evaluate(alreadyRevealed.context(0))
	if v2.Outcome != Fail {
		t.Fatalf("memoryCardAlreadyRevealed: Outcome = %v, want Fail", v2.Outcome)
	}
	if got, want := RenderMessage(v2.Message, table), "that card has already been revealed"; got != want {
		t.Fatalf("rendered = %q, want %q", got, want)
	}
}

// fakeTemplateConfigurer is a minimal TemplateConfigurer implementation used
// only to prove the interface's shape is usable via type-assertion, the way
// a later task's NewGameManager wiring will consume it from a delegate.
type fakeTemplateConfigurer struct{}

func (fakeTemplateConfigurer) ConfigureLegalTemplates() map[string]string {
	return map[string]string{"my.key": "my text"}
}

// TestTemplateConfigurerTypeAssertion verifies TemplateConfigurer is
// consumable via the optional-interface type-assertion pattern this
// package's other optional interfaces (ConstructorConfigurer) use.
func TestTemplateConfigurerTypeAssertion(t *testing.T) {
	var delegate interface{} = fakeTemplateConfigurer{}
	tc, ok := delegate.(TemplateConfigurer)
	if !ok {
		t.Fatal("fakeTemplateConfigurer does not satisfy TemplateConfigurer via type-assertion")
	}
	got := tc.ConfigureLegalTemplates()
	if got["my.key"] != "my text" {
		t.Fatalf("ConfigureLegalTemplates() = %+v, want my.key = \"my text\"", got)
	}

	// A delegate that doesn't implement it must fail the assertion cleanly,
	// not panic.
	var other interface{} = struct{}{}
	if _, ok := other.(TemplateConfigurer); ok {
		t.Fatal("struct{}{} unexpectedly satisfies TemplateConfigurer")
	}
}
