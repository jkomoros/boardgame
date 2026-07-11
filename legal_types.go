package boardgame

import (
	"encoding/json"
	"fmt"
	"math"
)

// LegalCatalogVersion stamps the shape of the declarative-legality wire
// format the server ledger ships (design spec §6: "A catalogVersion stamp
// ships with the ledger; a client with an older catalog treats unknown
// predicate names as evaluable: false and defers to server verdicts
// (graceful skew)"). It has no relationship to the module/package version —
// it is bumped only when the CATALOG (predicate names/args shapes, the
// PreconditionEntry wire shape itself) changes in a way an older client's
// evaluator couldn't safely interpret. Consumers: server/api's info
// response ships it as LegalCatalogVersion; a future TS evaluator reads it
// to decide whether it trusts its own bundled catalog for this game.
const LegalCatalogVersion = 1

// LegalOutcome is the three-valued verdict returned by legality evaluation.
// The zero value is deliberately invalid (neither LegalPass, LegalFail, nor
// LegalUnknown) so that a forgotten or zero-initialized LegalVerdict fails
// closed instead of silently reading as legal. See
// docs/superpowers/specs/2026-07-10-declarative-legality-design.md §1.
type LegalOutcome int

const (
	// legalOutcomeInvalid is the zero value of LegalOutcome. It is
	// unexported because it is not a valid outcome to construct
	// deliberately; it exists only so that a zero-value LegalVerdict fails
	// closed.
	legalOutcomeInvalid LegalOutcome = iota
	// LegalPass indicates the predicate was satisfied.
	LegalPass
	// LegalFail indicates the predicate was not satisfied.
	LegalFail
	// LegalUnknown indicates the predicate could not be evaluated (for
	// example because it depends on state hidden from the viewer, or is an
	// opaque escape-hatch check).
	LegalUnknown
)

// LegalBindingValue is a single named value substituted into a
// LegalMessage's template. Exactly one of S, I, or B must be set; this is
// enforced by MarshalJSON, which errors if none is set. Keeping this a
// closed set of typed pointers (rather than interface{}) keeps
// LegalMessage.Bindings JSON-round-trippable and conformant with a future
// TypeScript evaluator: no arbitrary `any` in the wire format.
type LegalBindingValue struct {
	// S is set if this binding is a string value.
	S *string
	// I is set if this binding is an int value.
	I *int
	// B is set if this binding is a bool value.
	B *bool
}

// MarshalJSON emits the bare JSON value (string, number, or bool) of
// whichever field is set. It returns an error if no field is set, since a
// zero-value LegalBindingValue violates the exactly-one-field-set
// invariant.
func (b LegalBindingValue) MarshalJSON() ([]byte, error) {
	switch {
	case b.S != nil:
		return json.Marshal(*b.S)
	case b.I != nil:
		return json.Marshal(*b.I)
	case b.B != nil:
		return json.Marshal(*b.B)
	default:
		return nil, fmt.Errorf("boardgame: LegalBindingValue has no field set")
	}
}

// UnmarshalJSON infers which field to populate from the JSON value's type:
// a JSON string sets S, a JSON number sets I, and a JSON bool sets B.
func (b *LegalBindingValue) UnmarshalJSON(data []byte) error {
	b.S, b.I, b.B = nil, nil, nil
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch val := raw.(type) {
	case string:
		b.S = &val
	case bool:
		b.B = &val
	case float64:
		if val != math.Trunc(val) {
			return fmt.Errorf("boardgame: LegalBindingValue: non-integer number %v not representable", val)
		}
		i := int(val)
		b.I = &i
	default:
		return fmt.Errorf("boardgame: LegalBindingValue JSON has unsupported type %T", raw)
	}
	return nil
}

// LegalMessage is a template key plus named bindings, never a pre-baked
// string, so failures are localizable, greppable, and re-renderable on
// server or client. Template keys resolve through the game's template
// table (see the design spec §6).
type LegalMessage struct {
	// Template is the template key, e.g. "reveal.no_cards_left".
	Template string
	// Bindings are the named values substituted into the template.
	Bindings map[string]LegalBindingValue
}

// LegalVerdict is the result of evaluating one predicate. Its zero value is
// intentionally invalid; use the constructors in the legal package
// (legal.PassVerdict, legal.FailT, legal.UnknownVerdict) to build one.
type LegalVerdict struct {
	// Outcome is the three-valued result of evaluation.
	Outcome LegalOutcome
	// Message is set on LegalFail (and optionally LegalUnknown); nil on
	// LegalPass.
	Message *LegalMessage
	// Reason explains why the outcome is LegalUnknown, e.g. "reads hidden
	// property HiddenCards".
	Reason string
}

// LegalCost is a coarse cost tier attached to a predicate as metadata. It
// does not affect evaluation order in v1 (declared order is evaluation
// order); it exists for docs and future lints/opt-in reordering.
type LegalCost int

const (
	// LegalCostTrivial marks a predicate as effectively free to evaluate.
	LegalCostTrivial LegalCost = iota
	// LegalCostCheap marks a predicate as inexpensive to evaluate.
	LegalCostCheap
	// LegalCostModerate marks a predicate as moderately expensive to
	// evaluate.
	LegalCostModerate
	// LegalCostExpensive marks a predicate as expensive to evaluate.
	LegalCostExpensive
)

// LegalFacet identifies which aspect of a property a predicate's read
// depends on. This is what makes client evaluability precise under
// sanitization: a stack-size check needs only LegalFacetCount, which
// survives sanitization policies that a full LegalFacetValues read would
// not.
type LegalFacet int

const (
	// LegalFacetValues means the predicate reads the actual values held at
	// the path.
	LegalFacetValues LegalFacet = iota
	// LegalFacetCount means the predicate reads only a count (e.g. stack
	// size) at the path.
	LegalFacetCount
	// LegalFacetOccupancy means the predicate reads only whether slots at
	// the path are occupied, not their values.
	LegalFacetOccupancy
	// LegalFacetOrder means the predicate reads only the order of items at
	// the path, not their values.
	LegalFacetOrder
)

// LegalPropPath is a property path in the state resolver's path grammar,
// e.g. "game.HiddenCards" or "move.CardIndex".
type LegalPropPath string

// LegalRead declares one property a predicate depends on, along with the
// facet of that property it actually needs. Read-sets are a conservative
// over-approximation used for client evaluability and, in the future,
// dirty-tracking.
type LegalRead struct {
	// Path is the property path being read.
	Path LegalPropPath
	// Facet is the aspect of that property being read.
	Facet LegalFacet
}

// LegalSpec is the serializable, registry-resolvable form of a predicate or
// compositor. A leaf spec has Name/Args; a compositor spec (only "any" in
// v1) has Name/Sub. Message optionally overrides the predicate's default
// template key.
type LegalSpec struct {
	// Name is the registry identity of the predicate or compositor, e.g.
	// "playerPropAtLeast" or "any".
	Name string `json:"name"`
	// Args are the string arguments to the named predicate.
	Args []string `json:"args,omitempty"`
	// Sub holds child specs for a compositor.
	Sub []LegalSpec `json:"sub,omitempty"`
	// Message, if set, overrides the predicate's default template key.
	Message string `json:"message,omitempty"`
}

// WithMessage returns a copy of s with Message set to templateKey. s itself
// is not modified.
func (s LegalSpec) WithMessage(templateKey string) LegalSpec {
	s.Message = templateKey
	return s
}
