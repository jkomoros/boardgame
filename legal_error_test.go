package boardgame

import (
	"errors"
	"testing"
)

// TestLegalVerdictErrorPassIsLiteralNil pins the nil-interface guarantee:
// LegalVerdict.Error() for LegalPass must return a literal nil, not a
// typed-nil *LegalError boxed into the error interface — the classic Go
// footgun where `var p *LegalError; var err error = p; err != nil` is true.
// Assigning the return value to a plain `error` variable and comparing
// against nil is exactly the trap this guards against.
func TestLegalVerdictErrorPassIsLiteralNil(t *testing.T) {
	v := LegalVerdict{Outcome: LegalPass}
	var err error = v.Error()
	if err != nil {
		t.Fatalf("LegalPass Verdict.Error() = %v, want nil", err)
	}
}

// TestLegalVerdictErrorFailReturnsLegalError verifies LegalFail produces a
// non-nil error whose concrete type is *LegalError, retrievable via
// errors.As, carrying the original Verdict.
func TestLegalVerdictErrorFailReturnsLegalError(t *testing.T) {
	v := LegalVerdict{
		Outcome: LegalFail,
		Message: &LegalMessage{Template: "some.key", Bindings: map[string]LegalBindingValue{
			"n": {I: intPtr(3)},
		}},
	}
	err := v.Error()
	if err == nil {
		t.Fatal("LegalFail Verdict.Error() = nil, want non-nil")
	}
	var target *LegalError
	if !errors.As(err, &target) {
		t.Fatalf("errors.As(err, &target) = false, want true (err: %v, %T)", err, err)
	}
	if target.Verdict.Outcome != LegalFail {
		t.Fatalf("target.Verdict.Outcome = %v, want LegalFail", target.Verdict.Outcome)
	}
	if target.Verdict.Message == nil || target.Verdict.Message.Template != "some.key" {
		t.Fatalf("target.Verdict.Message = %+v, want Template \"some.key\"", target.Verdict.Message)
	}
}

// TestLegalVerdictErrorUnknownReturnsLegalError verifies LegalUnknown also
// produces a non-nil *LegalError (fail-closed: only LegalPass is nil).
func TestLegalVerdictErrorUnknownReturnsLegalError(t *testing.T) {
	v := LegalVerdict{Outcome: LegalUnknown, Reason: "reads hidden state"}
	err := v.Error()
	if err == nil {
		t.Fatal("LegalUnknown Verdict.Error() = nil, want non-nil")
	}
	if got := err.Error(); got != "reads hidden state" {
		t.Fatalf("err.Error() = %q, want %q", got, "reads hidden state")
	}
}

// TestLegalVerdictErrorInvalidZeroValueReturnsLegalError verifies the
// zero-value (legalOutcomeInvalid) Outcome fails closed through Error() too
// — it must not be treated as LegalPass.
func TestLegalVerdictErrorInvalidZeroValueReturnsLegalError(t *testing.T) {
	var v LegalVerdict
	if err := v.Error(); err == nil {
		t.Fatal("zero-value Verdict.Error() = nil, want non-nil (fail closed)")
	}
}

// TestLegalErrorErrorNilSafe verifies (*LegalError)(nil).Error() does not
// panic and returns "".
func TestLegalErrorErrorNilSafe(t *testing.T) {
	var e *LegalError
	if got := e.Error(); got != "" {
		t.Fatalf("nil *LegalError.Error() = %q, want \"\"", got)
	}
}

// TestLegalErrorAttachTableNilSafe verifies AttachTable on a nil *LegalError
// returns nil rather than panicking.
func TestLegalErrorAttachTableNilSafe(t *testing.T) {
	var e *LegalError
	if got := e.AttachTable(map[string]string{"x": "y"}); got != nil {
		t.Fatalf("nil *LegalError.AttachTable(...) = %v, want nil", got)
	}
}

// TestLegalErrorFallsBackToTemplateKeyWithoutTable verifies Error() renders
// the bare template key when no table has been attached.
func TestLegalErrorFallsBackToTemplateKeyWithoutTable(t *testing.T) {
	v := LegalVerdict{Outcome: LegalFail, Message: &LegalMessage{Template: "reveal.no_cards_left"}}
	e := &LegalError{Verdict: v}
	if got := e.Error(); got != "reveal.no_cards_left" {
		t.Fatalf("Error() = %q, want the bare template key %q", got, "reveal.no_cards_left")
	}
}

// TestLegalErrorAttachTableRenders verifies AttachTable causes Error() to
// render Verdict.Message through the attached table, and that AttachTable
// returns a COPY (the original *LegalError, and any error value already
// referencing it, is unaffected).
func TestLegalErrorAttachTableRenders(t *testing.T) {
	v := LegalVerdict{
		Outcome: LegalFail,
		Message: &LegalMessage{Template: "reveal.no_cards_left", Bindings: map[string]LegalBindingValue{
			"left": {I: intPtr(0)},
		}},
	}
	original := &LegalError{Verdict: v}
	table := map[string]string{"reveal.no_cards_left": "You have {left} cards left to reveal"}
	attached := original.AttachTable(table)

	if got := attached.Error(); got != "You have 0 cards left to reveal" {
		t.Fatalf("attached.Error() = %q, want %q", got, "You have 0 cards left to reveal")
	}
	if got := original.Error(); got != "reveal.no_cards_left" {
		t.Fatalf("original.Error() = %q, want unchanged fallback %q (AttachTable must not mutate the receiver)", got, "reveal.no_cards_left")
	}
}

// TestRenderLegalMessageNilSafe verifies RenderLegalMessage(nil, ...) is "".
func TestRenderLegalMessageNilSafe(t *testing.T) {
	if got := RenderLegalMessage(nil, map[string]string{"x": "y"}); got != "" {
		t.Fatalf("RenderLegalMessage(nil, ...) = %q, want \"\"", got)
	}
	if got := RenderLegalMessage(nil, nil); got != "" {
		t.Fatalf("RenderLegalMessage(nil, nil) = %q, want \"\"", got)
	}
}

// TestRenderLegalMessageMissingBindingRendersPlaceholderName verifies a
// placeholder with no corresponding binding renders as the bare placeholder
// name, never panics, and never leaves the braces in the output.
func TestRenderLegalMessageMissingBindingRendersPlaceholderName(t *testing.T) {
	m := &LegalMessage{Template: "some.key", Bindings: map[string]LegalBindingValue{
		"present": {S: stringPtr("here")},
	}}
	table := map[string]string{"some.key": "{present} but {missing} is absent"}
	got := RenderLegalMessage(m, table)
	want := "here but missing is absent"
	if got != want {
		t.Fatalf("RenderLegalMessage = %q, want %q", got, want)
	}
}

// TestRenderLegalMessageNilBindingsMap verifies a template with placeholders
// but a nil Bindings map renders every placeholder as its own name, without
// panicking.
func TestRenderLegalMessageNilBindingsMap(t *testing.T) {
	m := &LegalMessage{Template: "some.key"}
	table := map[string]string{"some.key": "value is {value}"}
	if got := RenderLegalMessage(m, table); got != "value is value" {
		t.Fatalf("RenderLegalMessage = %q, want %q", got, "value is value")
	}
}

// TestRenderLegalMessageAllBindingKinds verifies string, int, and bool
// bindings all render correctly.
func TestRenderLegalMessageAllBindingKinds(t *testing.T) {
	m := &LegalMessage{Template: "k", Bindings: map[string]LegalBindingValue{
		"s": {S: stringPtr("hello")},
		"i": {I: intPtr(42)},
		"b": {B: boolPtr(true)},
	}}
	table := map[string]string{"k": "{s} {i} {b}"}
	if got := RenderLegalMessage(m, table); got != "hello 42 true" {
		t.Fatalf("RenderLegalMessage = %q, want %q", got, "hello 42 true")
	}
}

// TestRenderLegalMessageUnregisteredTemplateFallsBackToKey verifies that
// when table has no entry for m.Template at all, the bare key is used as
// the template body (and, if the key itself happens to contain no
// placeholders, is returned verbatim).
func TestRenderLegalMessageUnregisteredTemplateFallsBackToKey(t *testing.T) {
	m := &LegalMessage{Template: "unregistered.key"}
	if got := RenderLegalMessage(m, map[string]string{"other.key": "irrelevant"}); got != "unregistered.key" {
		t.Fatalf("RenderLegalMessage = %q, want the bare key %q", got, "unregistered.key")
	}
}

// --- validateLegalTemplates ---

// TestValidateLegalTemplatesExplicitMessageOverride verifies a Spec.Message
// override is checked against table.
func TestValidateLegalTemplatesExplicitMessageOverride(t *testing.T) {
	specs := []LegalSpec{{Name: "propAtLeast", Message: "custom.key"}}
	preds := []*LegalPredicate{{Name: "propAtLeast"}}

	if err := validateLegalTemplates(specs, preds, map[string]string{"custom.key": "text"}); err != nil {
		t.Fatalf("validateLegalTemplates with custom.key present = %v, want nil", err)
	}
	if err := validateLegalTemplates(specs, preds, map[string]string{}); err == nil {
		t.Fatal("validateLegalTemplates with custom.key missing = nil, want an error naming custom.key")
	}
}

// TestValidateLegalTemplatesLeafDefaultNotChecked verifies v1's documented
// scope decision: a leaf predicate's IMPLICIT default template key (no
// Spec.Message override) is not independently checked by this function —
// callers are expected to have merged legal.DefaultTemplates() into table
// already.
func TestValidateLegalTemplatesLeafDefaultNotChecked(t *testing.T) {
	specs := []LegalSpec{{Name: "propAtLeast"}}
	preds := []*LegalPredicate{{Name: "propAtLeast"}}
	if err := validateLegalTemplates(specs, preds, map[string]string{}); err != nil {
		t.Fatalf("validateLegalTemplates for an unoverridden leaf spec against an empty table = %v, want nil (leaf defaults are not independently verified in v1)", err)
	}
}

// TestValidateLegalTemplatesAnyCompositorDefault verifies the "any"
// compositor's implicit default key (legalAnyFailedTemplate) IS checked,
// since "any" is resolved by this package directly rather than through a
// registered constructor.
func TestValidateLegalTemplatesAnyCompositorDefault(t *testing.T) {
	specs := []LegalSpec{{Name: legalAnyCompositorName, Sub: []LegalSpec{
		{Name: "playerBool"}, {Name: "playerBool"},
	}}}
	preds := []*LegalPredicate{{Name: legalAnyCompositorName, Sub: []*LegalPredicate{
		{Name: "playerBool"}, {Name: "playerBool"},
	}}}

	if err := validateLegalTemplates(specs, preds, map[string]string{}); err == nil {
		t.Fatal("validateLegalTemplates for an any without legalAnyFailedTemplate in table = nil, want an error")
	}
	if err := validateLegalTemplates(specs, preds, map[string]string{legalAnyFailedTemplate: "text"}); err != nil {
		t.Fatalf("validateLegalTemplates for an any with legalAnyFailedTemplate present = %v, want nil", err)
	}
}

// TestValidateLegalTemplatesRecursesIntoAnySubs verifies a Message override
// on a sub-spec beneath "any" is itself validated.
func TestValidateLegalTemplatesRecursesIntoAnySubs(t *testing.T) {
	specs := []LegalSpec{{Name: legalAnyCompositorName, Sub: []LegalSpec{
		{Name: "playerBool", Message: "sub.override"},
		{Name: "playerBool"},
	}}}
	preds := []*LegalPredicate{{Name: legalAnyCompositorName, Sub: []*LegalPredicate{
		{Name: "playerBool"}, {Name: "playerBool"},
	}}}

	if err := validateLegalTemplates(specs, preds, map[string]string{legalAnyFailedTemplate: "text"}); err == nil {
		t.Fatal("validateLegalTemplates missing sub.override = nil, want an error naming sub.override")
	}
	full := map[string]string{legalAnyFailedTemplate: "text", "sub.override": "text2"}
	if err := validateLegalTemplates(specs, preds, full); err != nil {
		t.Fatalf("validateLegalTemplates with sub.override present = %v, want nil", err)
	}
}

// TestValidateLegalTemplatesLengthMismatch verifies specs/predicates length
// mismatch is itself a validation error rather than an index panic.
func TestValidateLegalTemplatesLengthMismatch(t *testing.T) {
	specs := []LegalSpec{{Name: "propAtLeast"}, {Name: "playerBool"}}
	preds := []*LegalPredicate{{Name: "propAtLeast"}}
	if err := validateLegalTemplates(specs, preds, map[string]string{}); err == nil {
		t.Fatal("validateLegalTemplates with mismatched slice lengths = nil, want an error")
	}
}

func intPtr(i int) *int          { return &i }
func stringPtr(s string) *string { return &s }
func boolPtr(b bool) *bool       { return &b }

// TestLegalTemplatePlaceholders pins the exported placeholder extractor
// (footgun-batch F4): distinct names, first-appearance order, matching
// exactly the {name} pattern RenderLegalMessage substitutes.
func TestLegalTemplatePlaceholders(t *testing.T) {
	tests := []struct {
		body string
		want []string
	}{
		{"no placeholders here", nil},
		{"", nil},
		{"one {value}", []string{"value"}},
		{"{value} then {min} then {value} again", []string{"value", "min"}},
		{"underscores and digits: {some_name_2}", []string{"some_name_2"}},
		{"not a placeholder: {with space} {with-dash} {}", nil},
	}
	for _, tc := range tests {
		got := LegalTemplatePlaceholders(tc.body)
		if len(got) != len(tc.want) {
			t.Errorf("LegalTemplatePlaceholders(%q) = %v, want %v", tc.body, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("LegalTemplatePlaceholders(%q) = %v, want %v", tc.body, got, tc.want)
				break
			}
		}
	}
}
