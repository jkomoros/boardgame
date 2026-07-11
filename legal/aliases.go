package legal

import "github.com/jkomoros/boardgame"

// Outcome is an alias for boardgame.LegalOutcome. Core owns the underlying
// type; this package exists so game authors write legal.Outcome instead of
// boardgame.LegalOutcome.
type Outcome = boardgame.LegalOutcome

// BindingValue is an alias for boardgame.LegalBindingValue.
type BindingValue = boardgame.LegalBindingValue

// Message is an alias for boardgame.LegalMessage.
type Message = boardgame.LegalMessage

// Verdict is an alias for boardgame.LegalVerdict.
type Verdict = boardgame.LegalVerdict

// Cost is an alias for boardgame.LegalCost.
type Cost = boardgame.LegalCost

// Facet is an alias for boardgame.LegalFacet.
type Facet = boardgame.LegalFacet

// PropPath is an alias for boardgame.LegalPropPath.
type PropPath = boardgame.LegalPropPath

// Read is an alias for boardgame.LegalRead.
type Read = boardgame.LegalRead

// Spec is an alias for boardgame.LegalSpec.
type Spec = boardgame.LegalSpec

// Context is an alias for boardgame.LegalContext: the entire vocabulary a
// Predicate's Evaluate func may reference.
type Context = boardgame.LegalContext

// Predicate is an alias for boardgame.LegalPredicate: one resolved
// legality question, either a leaf built by a PredicateConstructor or a
// compositor ("any" in v1) whose Sub holds its children.
type Predicate = boardgame.LegalPredicate

// PredicateConstructor is an alias for boardgame.LegalPredicateConstructor:
// a named factory that turns a Spec into a *Predicate. Games register their
// own predicates through the same registry mechanism the built-in catalog
// uses.
type PredicateConstructor = boardgame.LegalPredicateConstructor

// Pass, Fail, and Unknown alias boardgame's LegalOutcome values, so callers
// that need to inspect a Verdict's Outcome (e.g. in tests) can compare
// against legal.Pass rather than boardgame.LegalPass. Prefer the
// PassVerdict/FailT/UnknownVerdict constructors for building Verdicts.
const (
	// Pass aliases boardgame.LegalPass.
	Pass = boardgame.LegalPass
	// Fail aliases boardgame.LegalFail.
	Fail = boardgame.LegalFail
	// Unknown aliases boardgame.LegalUnknown.
	Unknown = boardgame.LegalUnknown
)

// PassVerdict returns a Verdict with Outcome set to boardgame.LegalPass and
// no Message.
func PassVerdict() Verdict {
	return Verdict{
		Outcome: boardgame.LegalPass,
	}
}

// FailT returns a Verdict with Outcome set to boardgame.LegalFail and a
// Message carrying the given template key and, if provided, bindings. At
// most one bindings map is used; a bindings map is optional so callers can
// write legal.FailT("some.key") for templates with no bindings.
func FailT(template string, bindings ...map[string]boardgame.LegalBindingValue) Verdict {
	var b map[string]boardgame.LegalBindingValue
	if len(bindings) > 0 {
		b = bindings[0]
	}
	return Verdict{
		Outcome: boardgame.LegalFail,
		Message: &Message{
			Template: template,
			Bindings: b,
		},
	}
}

// UnknownVerdict returns a Verdict with Outcome set to
// boardgame.LegalUnknown and Reason set to the given explanation.
func UnknownVerdict(reason string) Verdict {
	return Verdict{
		Outcome: boardgame.LegalUnknown,
		Reason:  reason,
	}
}

// String returns a BindingValue with its string field set to s, for use in
// a LegalMessage's Bindings map.
func String(s string) boardgame.LegalBindingValue {
	return boardgame.LegalBindingValue{S: &s}
}

// Int returns a BindingValue with its int field set to i, for use in a
// LegalMessage's Bindings map.
func Int(i int) boardgame.LegalBindingValue {
	return boardgame.LegalBindingValue{I: &i}
}

// Bool returns a BindingValue with its bool field set to b, for use in a
// LegalMessage's Bindings map.
func Bool(b bool) boardgame.LegalBindingValue {
	return boardgame.LegalBindingValue{B: &b}
}
