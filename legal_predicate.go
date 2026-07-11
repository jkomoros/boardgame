package boardgame

import (
	"fmt"
)

// LegalContext is the entire vocabulary a LegalPredicate's Evaluate func may
// reference — the wall before the Turing tarpit: no I/O, no mutation,
// nothing beyond these four fields. See the design spec §1 "Predicates are
// structs-with-a-func, not interfaces".
type LegalContext struct {
	// State is the state being evaluated against.
	State ImmutableState
	// Move is the move being evaluated, or nil during field-independent
	// evaluation (a predicate whose declared Reads include no move.* path
	// may be evaluated with a nil Move). See evalLegalPredicate for the
	// runtime guard that protects a predicate that touches ctx.Move despite
	// not declaring a move.* read.
	Move Move
	// Proposer is the player proposing the move.
	Proposer PlayerIndex
	// Chest is the game's component chest.
	Chest *ComponentChest
}

// LegalPredicate is one resolved legality question: either a leaf, built by
// a LegalPredicateConstructor from a LegalSpec, or a compositor (only "any"
// in v1), whose Sub holds its children. See the design spec §1.
type LegalPredicate struct {
	// Name is the registry identity of this predicate, e.g.
	// "playerPropAtLeast", or "any" for the any-compositor.
	Name string
	// Args are the string arguments this predicate was constructed with.
	// With Name, this round-trips the registry. Empty for compositors.
	Args []string
	// Reads is the declared read-set: a conservative over-approximation of
	// the state this predicate's Evaluate depends on, used for client
	// evaluability and (in the future) dirty-tracking. For the "any"
	// compositor, Reads is the union of its children's Reads.
	Reads []LegalRead
	// Cost is a coarse cost tier, metadata only in v1. For "any", Cost is
	// the max of its children's Cost.
	Cost LegalCost
	// Evaluate is the pure function that computes this predicate's verdict.
	// Callers should invoke it via evalLegalPredicate rather than directly,
	// so that a zero-value (invalid) verdict, a panic, or an undeclared
	// touch of a nil ctx.Move are all converted to a fail-closed
	// LegalUnknown instead of propagating.
	Evaluate func(ctx LegalContext) LegalVerdict
	// Sub holds child predicates; only set for compositors ("any" in v1).
	Sub []*LegalPredicate
	// opaque marks an escape-hatch wrapper predicate (e.g. LegalCustom) that
	// has no serialized form: it did not come from resolving a LegalSpec
	// and cannot be turned back into one. Unexported: only this package's
	// escape-hatch wrapper (added in a later task) sets it.
	opaque bool
}

// Serializable reports whether p (and, recursively, every predicate in its
// Sub tree) has a serialized form: it is false if p itself is an opaque
// escape-hatch wrapper, or if any of its Sub predicates is not itself
// Serializable. A nil p is not Serializable.
func (p *LegalPredicate) Serializable() bool {
	if p == nil {
		return false
	}
	if p.opaque {
		return false
	}
	for _, sub := range p.Sub {
		if !sub.Serializable() {
			return false
		}
	}
	return true
}

// LegalPredicateConstructor is a named factory that turns a LegalSpec into a
// *LegalPredicate. Constructors mirror constraints.StackConstraintConstructor:
// games register their own predicates through the same registry mechanism
// resolveLegalSpecs consumes, which is how a game-specific check (e.g.
// checkers' board-geometry predicate) stays serializable. See the design
// spec §1.
type LegalPredicateConstructor struct {
	// Name is the registry key this constructor is registered under. It
	// should match the Name every LegalSpec it constructs from carries.
	Name string
	// Constructor builds a *LegalPredicate from spec. resolve is provided so
	// a constructor whose spec has its own notion of sub-specs (distinct
	// from the "any" compositor) can recursively resolve them through the
	// same registry and validation path as top-level specs.
	Constructor func(spec LegalSpec, chest *ComponentChest,
		resolve func(LegalSpec) (*LegalPredicate, error)) (*LegalPredicate, error)
}

// legalAnyCompositorName is the registry-reserved name of the sole v1
// compositor. It is intercepted directly in resolveLegalSpecs's resolve
// closure rather than going through the registry map, so that "any" cannot
// be shadowed or omitted by a caller-supplied registry. See spec §1
// "Anti-tarpit rules": any is the only compositor in v1, enforced to depth
// 1 (no sub of an any may itself be an any), and requires at least 2 subs.
const legalAnyCompositorName = "any"

// resolveLegalSpecs resolves specs against registry, producing the
// corresponding *LegalPredicate slice in the same order. It enforces the
// anti-tarpit rules from spec §1: an unknown Name is an error naming the
// offending spec; "any" is the only compositor (any other Name with Sub set
// is not itself special-cased — Sub is only meaningful to the "any"
// resolution path and to whatever a specific constructor chooses to do with
// it via the resolve callback); an "any" whose Sub contains another "any"
// is rejected (depth 1 only), naming both the outer spec and the offending
// nested "any"; an "any" with fewer than 2 Sub is rejected.
//
// resolveLegalSpecs is also where every resolved predicate's declared Reads
// paths are validated via validateLegalPath, so a typo in a Read's path
// fails at boot (naming the offending predicate and path) rather than
// surfacing mid-game. Because validateLegalPath needs both an example state
// (for game.*/player.*/players[*].* paths) and a move reader (for move.*
// paths), resolveLegalSpecs takes both as explicit parameters, beyond the
// three the brief sketched (specs, registry, chest) — a deviation
// documented here and forwarded from Task 2's reviewer note. Either may be
// nil to skip the corresponding class of path validation: exampleState nil
// skips game.*/player.*/players[*].* validation, moveReader nil skips
// move.* validation. nil is also today's production default for moveReader
// until a later task wires NewGameManager to pass the real example move's
// reader; passing nil for both is convenient for isolated
// construction/registry tests that don't want to stand up a full example
// state.
func resolveLegalSpecs(specs []LegalSpec, registry map[string]*LegalPredicateConstructor,
	chest *ComponentChest, exampleState ImmutableState, moveReader PropertyReader) ([]*LegalPredicate, error) {

	var resolve func(spec LegalSpec) (*LegalPredicate, error)
	resolve = func(spec LegalSpec) (*LegalPredicate, error) {
		if spec.Name == legalAnyCompositorName {
			return resolveLegalAnySpec(spec, resolve)
		}

		ctor, ok := registry[spec.Name]
		if !ok {
			return nil, fmt.Errorf("boardgame: legal spec %q: unknown predicate name %q", spec.Name, spec.Name)
		}

		pred, err := ctor.Constructor(spec, chest, resolve)
		if err != nil {
			return nil, fmt.Errorf("boardgame: legal spec %q: constructor failed: %w", spec.Name, err)
		}
		if pred == nil {
			return nil, fmt.Errorf("boardgame: legal spec %q: constructor returned a nil predicate", spec.Name)
		}
		return pred, nil
	}

	result := make([]*LegalPredicate, 0, len(specs))
	for _, spec := range specs {
		pred, err := resolve(spec)
		if err != nil {
			return nil, err
		}
		if err := validateLegalReadsForBoot(pred.Reads, exampleState, moveReader); err != nil {
			return nil, fmt.Errorf("boardgame: legal spec %q: %w", spec.Name, err)
		}
		result = append(result, pred)
	}
	return result, nil
}

// resolveLegalAnySpec resolves an "any"-named spec into the built-in
// any-compositor predicate: depth-1 enforced (no sub may itself be "any"),
// at least 2 subs required. Its Reads is the union of its children's Reads,
// its Cost the max of its children's Cost, and its Evaluate implements the
// Kleene truth table (spec §6): any child Pass -> Pass; else if any child
// Unknown -> Unknown; else Fail, with the spec's Message key if set, else
// the "legal.any_failed" template.
func resolveLegalAnySpec(spec LegalSpec, resolve func(LegalSpec) (*LegalPredicate, error)) (*LegalPredicate, error) {
	if len(spec.Sub) < 2 {
		return nil, fmt.Errorf("boardgame: legal spec %q: %q compositor requires at least 2 sub-specs, got %d", spec.Name, legalAnyCompositorName, len(spec.Sub))
	}

	subs := make([]*LegalPredicate, 0, len(spec.Sub))
	for _, subSpec := range spec.Sub {
		if subSpec.Name == legalAnyCompositorName {
			return nil, fmt.Errorf("boardgame: legal spec %q: sub-spec %q may not itself be %q (any is depth-1 only, no nested any)", spec.Name, subSpec.Name, legalAnyCompositorName)
		}
		sub, err := resolve(subSpec)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}

	template := spec.Message
	if template == "" {
		template = "legal.any_failed"
	}

	return &LegalPredicate{
		Name:  legalAnyCompositorName,
		Reads: unionLegalReads(subs),
		Cost:  maxLegalCost(subs),
		Sub:   subs,
		Evaluate: func(ctx LegalContext) LegalVerdict {
			return evalLegalAnyKleene(subs, ctx, template)
		},
	}, nil
}

// evalLegalAnyKleene implements the Kleene truth table for the "any"
// compositor (spec §6): any child Pass -> Pass; else if any child Unknown
// -> Unknown; else Fail (all children Fail), with a LegalMessage built from
// template. Each child is evaluated via evalLegalPredicate so a panicking
// or invalid-verdict child degrades to Unknown for that child rather than
// propagating.
func evalLegalAnyKleene(subs []*LegalPredicate, ctx LegalContext, template string) LegalVerdict {
	sawUnknown := false
	for _, sub := range subs {
		v := evalLegalPredicate(sub, ctx)
		switch v.Outcome {
		case LegalPass:
			return LegalVerdict{Outcome: LegalPass}
		case LegalUnknown:
			sawUnknown = true
		}
	}
	if sawUnknown {
		return LegalVerdict{Outcome: LegalUnknown, Reason: "a sub-predicate of \"any\" was unknown"}
	}
	return LegalVerdict{
		Outcome: LegalFail,
		Message: &LegalMessage{Template: template},
	}
}

// unionLegalReads returns the de-duplicated union of every predicate in
// subs's Reads, in first-seen order.
func unionLegalReads(subs []*LegalPredicate) []LegalRead {
	seen := make(map[LegalRead]bool)
	var out []LegalRead
	for _, sub := range subs {
		for _, r := range sub.Reads {
			if !seen[r] {
				seen[r] = true
				out = append(out, r)
			}
		}
	}
	return out
}

// maxLegalCost returns the highest LegalCost among subs. subs must be
// non-empty.
func maxLegalCost(subs []*LegalPredicate) LegalCost {
	max := subs[0].Cost
	for _, sub := range subs[1:] {
		if sub.Cost > max {
			max = sub.Cost
		}
	}
	return max
}

// validateLegalReadsForBoot validates every path in reads via
// validateLegalPath, skipping a given Read if the validation target it
// needs is nil: a move.* Read is skipped if moveReader is nil, and a
// game.*/player.*/players[*].* Read is skipped if exampleState is nil. A
// malformed path (one that fails to parse at all) is always an error,
// regardless of exampleState/moveReader.
func validateLegalReadsForBoot(reads []LegalRead, exampleState ImmutableState, moveReader PropertyReader) error {
	for _, r := range reads {
		parsed, err := parseLegalPath(r.Path)
		if err != nil {
			return err
		}
		if parsed.kind == pathMove {
			if moveReader == nil {
				continue
			}
		} else if exampleState == nil {
			continue
		}
		if err := validateLegalPath(r.Path, exampleState, moveReader); err != nil {
			return err
		}
	}
	return nil
}

// evalLegalPredicate is the sole sanctioned way to run a *LegalPredicate's
// Evaluate: it fails closed. Three failure modes are all converted to
// LegalUnknown rather than propagating:
//
//   - p is nil, or p.Evaluate is nil: LegalUnknown naming p.
//   - p.Evaluate panics: recovered. If ctx.Move is nil and p's declared
//     Reads contain no move.* path, the panic is assumed to be an
//     undeclared touch of ctx.Move and reported as
//     LegalUnknown{Reason: "undeclared move read"} (spec §1's "Runtime
//     guard"). Any other panic is reported as LegalUnknown with the
//     predicate's Name and the recovered value in Reason.
//   - p.Evaluate returns a zero-value (LegalOutcome-invalid) LegalVerdict:
//     LegalUnknown{Reason: "predicate returned invalid verdict"} — a
//     forgotten Outcome must never silently read as legal.
func evalLegalPredicate(p *LegalPredicate, ctx LegalContext) (verdict LegalVerdict) {
	if p == nil {
		return LegalVerdict{Outcome: LegalUnknown, Reason: "predicate was nil"}
	}
	if p.Evaluate == nil {
		return LegalVerdict{Outcome: LegalUnknown, Reason: fmt.Sprintf("predicate %q has no Evaluate function", p.Name)}
	}

	defer func() {
		if r := recover(); r != nil {
			if ctx.Move == nil && !legalReadsIncludeMovePath(p.Reads) {
				verdict = LegalVerdict{Outcome: LegalUnknown, Reason: "undeclared move read"}
				return
			}
			verdict = LegalVerdict{Outcome: LegalUnknown, Reason: fmt.Sprintf("predicate %q panicked: %v", p.Name, r)}
		}
	}()

	v := p.Evaluate(ctx)
	if v.Outcome == legalOutcomeInvalid {
		return LegalVerdict{Outcome: LegalUnknown, Reason: "predicate returned invalid verdict"}
	}
	return v
}

// legalReadsIncludeMovePath reports whether reads contains at least one
// path that parses as a move.* path. A malformed path is treated as not a
// move path (parseLegalPath's error is ignored here — reads that failed to
// parse are a construction-time bug caught by resolveLegalSpecs' boot-time
// validation, not something evalLegalPredicate should mask or amplify).
func legalReadsIncludeMovePath(reads []LegalRead) bool {
	for _, r := range reads {
		if parsed, err := parseLegalPath(r.Path); err == nil && parsed.kind == pathMove {
			return true
		}
	}
	return false
}
