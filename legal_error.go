package boardgame

import (
	"fmt"
	"regexp"
	"strconv"
)

// legalAnyFailedTemplate is the default Fail template key the "any"
// compositor emits (resolveLegalAnySpec) when its own Spec has no Message
// override. It is declared here, in core, rather than alongside the
// catalog's other default-template constants (legal/catalog_stack.go's
// defaultTemplateKeys), because "any" is a core-level compositor — resolved
// directly by resolveLegalSpecs, never through the LegalPredicateConstructor
// registry package legal populates — not a legal-package catalog predicate.
// validateLegalTemplates (below) and legal.DefaultTemplates() (legal/templates.go)
// both need this exact key: validateLegalTemplates checks it directly since
// it lives in this package; legal.DefaultTemplates() carries its own copy of
// the string literal (package legal cannot reference this unexported
// constant), and a test on both sides pins them equal.
const legalAnyFailedTemplate = "legal.any_failed"

// legalPlaceholderPattern matches a "{name}" placeholder in a template body:
// name is one or more letters, digits, or underscores.
var legalPlaceholderPattern = regexp.MustCompile(`\{([A-Za-z0-9_]+)\}`)

// RenderLegalMessage renders m into human-readable text: it looks up
// m.Template in table to get the template body (falling back to the bare
// template key itself if table is nil or has no entry for it — this is the
// "template key as fallback text if unregistered" behavior the design spec
// §3 describes), then substitutes every "{name}" placeholder in that body
// with the string form of m.Bindings["name"] (whichever of S/I/B is set). A
// placeholder whose name has no entry in m.Bindings (or whose entry is a
// zero-value LegalBindingValue with none of S/I/B set) renders as the bare
// placeholder name itself, e.g. "{value}" -> "value" — this is deliberate,
// not an error: a missing binding is a caller bug, but RenderLegalMessage
// must never panic or produce garbled output over it (see the design spec
// §6 and the legal package's RenderMessage, which is the public,
// game-author-facing wrapper around this function). RenderLegalMessage is
// nil-safe: a nil m renders as "".
//
// package legal's RenderMessage delegates to this function directly
// (package legal already imports package boardgame, so there is no cycle);
// this keeps the rendering algorithm defined exactly once, since
// LegalError.Error() (below) also needs it and cannot import package legal
// (that would be the cycle: legal depends on boardgame, never the reverse).
func RenderLegalMessage(m *LegalMessage, table map[string]string) string {
	if m == nil {
		return ""
	}
	body, ok := table[m.Template]
	if !ok {
		body = m.Template
	}
	return legalPlaceholderPattern.ReplaceAllStringFunc(body, func(match string) string {
		name := match[1 : len(match)-1]
		val, ok := m.Bindings[name]
		if !ok {
			return name
		}
		switch {
		case val.S != nil:
			return *val.S
		case val.I != nil:
			return strconv.Itoa(*val.I)
		case val.B != nil:
			return strconv.FormatBool(*val.B)
		default:
			return name
		}
	})
}

// LegalError is the concrete error type LegalVerdict.Error() returns for a
// LegalFail or LegalUnknown Verdict (never for LegalPass — see
// LegalVerdict.Error()'s doc comment for the nil-interface guarantee).
// Error() renders Verdict.Message through an attached template table (see
// AttachTable) when one has been set, falling back to the raw template key
// (via RenderLegalMessage's own fallback) when no table is attached.
type LegalError struct {
	// Verdict is the LegalFail or LegalUnknown verdict this error wraps.
	Verdict LegalVerdict
	// table is the template table Error() renders Verdict.Message against.
	// Unset (nil) until AttachTable is called. Unexported: callers attach a
	// table only through AttachTable, never by field access, so this
	// mechanism can evolve without an API break.
	table map[string]string
}

// Error implements the error interface. It is nil-safe (a nil *LegalError
// renders as ""). For LegalFail (and any LegalUnknown that carries a
// Message), it renders Verdict.Message via RenderLegalMessage against the
// attached table (nil if none was attached, which RenderLegalMessage
// handles by falling back to the bare template key). For a Message-less
// LegalUnknown, it returns Verdict.Reason, or a generic fallback if Reason
// is empty too.
func (e *LegalError) Error() string {
	if e == nil {
		return ""
	}
	if e.Verdict.Message != nil {
		return RenderLegalMessage(e.Verdict.Message, e.table)
	}
	if e.Verdict.Reason != "" {
		return e.Verdict.Reason
	}
	return "boardgame: move is not legal"
}

// AttachTable returns a copy of e with table set as the template table
// Error() renders against; e itself is not modified. This is the handoff
// mechanism a later task's engine uses: it resolves the game's merged
// template table once (legal.DefaultTemplates() overlaid with the
// delegate's legal.TemplateConfigurer table, if the delegate implements
// that optional interface) and calls AttachTable on every *LegalError it
// hands back from a move's Legal(), so LegalForPlayerError's rendered text
// reflects the owning game's own templates rather than falling back to bare
// template keys. AttachTable is nil-safe (AttachTable on a nil *LegalError
// returns nil).
func (e *LegalError) AttachTable(table map[string]string) *LegalError {
	if e == nil {
		return nil
	}
	clone := *e
	clone.table = table
	return &clone
}

// Error converts v into an error: a literal nil for LegalPass — not a
// typed-nil *LegalError boxed into the error interface, which is why this
// method's return type is the error INTERFACE rather than *LegalError: a
// `return nil` from a function declared to return error produces a true nil
// interface value, so `var err error = someVerdict.Error(); err == nil`
// holds whenever someVerdict.Outcome is LegalPass. Every other Outcome
// (LegalFail, LegalUnknown, and the invalid zero value — fail-closed, per
// the design spec §1) returns a non-nil *LegalError wrapping v, with no
// template table attached yet (see LegalError.AttachTable).
func (v LegalVerdict) Error() error {
	if v.Outcome == LegalPass {
		return nil
	}
	return &LegalError{Verdict: v}
}

// validateLegalEmittedTemplates walks each predicate (and, recursively, its
// Sub tree) and returns an error if any key in a predicate's
// EmittedTemplates is absent from table. This is the completion of the
// spec §3 "unregistered keys are a boot error" invariant that
// validateLegalTemplates could not cover: validateLegalTemplates only sees
// EXPLICIT Spec.Message overrides (plus the one core-level "any" default),
// whereas EmittedTemplates is populated by every predicate's constructor
// with the effective template key(s) its Evaluate can actually emit —
// including a leaf catalog predicate's implicit default and a
// game-registered predicate's own hardcoded key. table must already be the
// caller-merged union of legal.DefaultTemplates() and the delegate's
// legal.TemplateConfigurer table (same layering constraint as
// validateLegalTemplates). The boot call site wraps the returned error with
// the owning MOVE's name (this function, like the rest of core's legal
// plumbing, does not know it).
func validateLegalEmittedTemplates(predicates []*LegalPredicate, table map[string]string) error {
	for _, pred := range predicates {
		if err := validateLegalEmittedTemplatesTree(pred, table); err != nil {
			return err
		}
	}
	return nil
}

func validateLegalEmittedTemplatesTree(pred *LegalPredicate, table map[string]string) error {
	if pred == nil {
		return nil
	}
	for _, key := range pred.EmittedTemplates {
		if _, ok := table[key]; !ok {
			return fmt.Errorf("boardgame: predicate %q may emit template key %q, which is not found in the game's template table", pred.Name, key)
		}
	}
	for _, sub := range pred.Sub {
		if err := validateLegalEmittedTemplatesTree(sub, table); err != nil {
			return err
		}
	}
	return nil
}

// validateLegalTemplates is the boot-time template-key validation contract,
// consumed by a later task's NewGameManager wiring (design spec §6:
// "validated at NewGameManager: every template key referenced by any
// Spec/FailT/Errorf must exist"). specs and predicates must be parallel
// slices in the same order and shape resolveLegalSpecs produces
// (predicates[i] is what specs[i] resolved to; a compositor's Spec.Sub and
// LegalPredicate.Sub line up the same way) — a length mismatch is itself a
// validation error. table must already be the CALLER-merged union of
// legal.DefaultTemplates() and the delegate's legal.TemplateConfigurer
// table, if any: this function lives in core, one layer below package
// legal, so it cannot compute that union itself (see RenderLegalMessage's
// doc comment for the same layering constraint).
//
// Coverage decision (v1, deliberately scoped down — see the task-6 report
// for the full justification): this function verifies every EXPLICIT
// Spec.Message override against table, plus the one core-level IMPLICIT
// default it knows about (the "any" compositor's legalAnyFailedTemplate,
// since "any" is resolved by this package, not by a registered
// LegalPredicateConstructor). It does NOT attempt to verify every implicit
// default template key a leaf catalog or game-registered predicate's
// Evaluate might emit when Spec.Message is unset (e.g. propAtLeast's
// TemplatePropAtLeast, or a game-registered predicate's own hardcoded
// key): LegalPredicate carries no metadata field declaring "the template
// keys my Evaluate might return", and introspecting an Evaluate closure is
// not possible in Go. For the built-in catalog, this gap is covered by
// convention instead: every catalog predicate's default key is listed in
// legal/catalog_stack.go's defaultTemplateKeys, which legal.DefaultTemplates()
// is built from and tested (legal/templates_test.go) to cover completely —
// so as long as the caller merges legal.DefaultTemplates() into table
// before calling this function (which the design spec §6 requires), every
// catalog predicate's implicit default is already present. A
// game-registered predicate's own implicit default key (e.g. checkers'
// "checkers.black_spaces_only") is NOT verified at boot by this function;
// an unregistered key there degrades gracefully at render time
// (RenderLegalMessage falls back to the bare key) rather than failing
// boot. A future enhancement could close this gap by adding an
// EmittedTemplates []string-style declaration to
// LegalPredicateConstructor, matching Reads' by-convention declaration
// model — noted as follow-up, not implemented here.
func validateLegalTemplates(specs []LegalSpec, predicates []*LegalPredicate, table map[string]string) error {
	if len(specs) != len(predicates) {
		return fmt.Errorf("boardgame: validateLegalTemplates: specs and predicates must be parallel slices of equal length, got %d specs and %d predicates", len(specs), len(predicates))
	}
	for i, spec := range specs {
		if err := validateLegalTemplateTree(spec, predicates[i], table); err != nil {
			return err
		}
	}
	return nil
}

// validateLegalTemplateTree validates one (spec, pred) pair, then recurses
// into spec.Sub/pred.Sub when both are present and the same length (which
// in v1 is only ever true for the "any" compositor — see
// validateLegalTemplates's doc comment on why other Sub-bearing specs, like
// AllActivePlayers, don't line up this way and are simply not recursed
// into: they hand-compile their inner spec rather than preserving it as a
// resolved Sub predicate).
func validateLegalTemplateTree(spec LegalSpec, pred *LegalPredicate, table map[string]string) error {
	if spec.Message != "" {
		if _, ok := table[spec.Message]; !ok {
			return fmt.Errorf("boardgame: legal spec %q: template key %q (Spec.Message override) not found in template table", spec.Name, spec.Message)
		}
	} else if pred != nil && pred.Name == legalAnyCompositorName {
		if _, ok := table[legalAnyFailedTemplate]; !ok {
			return fmt.Errorf("boardgame: legal spec %q: default template key %q not found in template table", spec.Name, legalAnyFailedTemplate)
		}
	}

	if pred != nil && len(spec.Sub) == len(pred.Sub) {
		for i, subSpec := range spec.Sub {
			if err := validateLegalTemplateTree(subSpec, pred.Sub[i], table); err != nil {
				return err
			}
		}
	}

	return nil
}
