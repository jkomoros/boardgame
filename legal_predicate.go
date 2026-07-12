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
	// EmittedTemplates is the conservative set of template keys this
	// predicate's Evaluate may emit in a Fail (or Message-carrying Unknown)
	// verdict — the effective keys AFTER any Spec.Message override was
	// applied at construction time. It is populated by the constructor that
	// built the predicate (each catalog constructor lists the key(s) its
	// Evaluate can FailT with; a game-registered constructor should do the
	// same). Boot-time validation (validateLegalEmittedTemplates) checks
	// every key here against the game's merged template table, so a
	// predicate that can emit an unregistered template key fails at
	// NewGameManager naming the owning move rather than rendering a bare key
	// mid-game — closing the spec §3 "unregistered keys are a boot error"
	// invariant for game-registered predicates and implicit catalog
	// defaults alike. Empty/nil is allowed (an opaque escape-hatch wrapper,
	// or a predicate that never fails with a template, declares none).
	EmittedTemplates []string
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
// it via the resolve callback); an "any" is enforced to depth 1 — no "any"
// may be resolved anywhere beneath another "any" on the same resolution
// path; an "any" with fewer than 2 Sub is rejected.
//
// The depth-1 rule is enforced at RESOLUTION time, not by literally
// inspecting each sub-spec's Name: the resolve closure handed to every
// constructor (via LegalPredicateConstructor.Constructor's resolve
// parameter) closes over whether the current resolution is already inside
// an "any". A constructor that internally calls resolve on an
// {Name: "any", ...} spec — even though its own spec's Name is something
// else entirely, e.g. a game-registered "wrapsAnAny" — hits the same check
// as a spec whose Sub literally names "any" directly, because the
// enforcement lives in what gets resolved, not in what the literal spec
// tree looks like. This closes a bypass where a constructor could plant a
// depth-2 "any" by resolving one internally and returning it as its own
// predicate. See checkNoNestedAny below for a second, independent
// belt-and-suspenders check over the resulting *LegalPredicate tree, which
// also catches a constructor that hand-builds a nested-"any"
// *LegalPredicate without ever calling resolve.
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

	// resolve is the internal resolver. insideAny is true whenever this
	// call is resolving something that is, directly or via any number of
	// constructor hops, beneath an "any" that is already being resolved.
	// The resolve closure bound and handed to a constructor (below) always
	// captures the insideAny value in effect at the point the constructor
	// was invoked, so a constructor cannot escape the depth-1 rule by
	// resolving an "any" itself: that call re-enters resolve with the same
	// insideAny=true, and resolving "any" while insideAny is an error.
	var resolve func(spec LegalSpec, insideAny bool) (*LegalPredicate, error)
	resolve = func(spec LegalSpec, insideAny bool) (*LegalPredicate, error) {
		if spec.Name == legalAnyCompositorName {
			if insideAny {
				return nil, fmt.Errorf("boardgame: legal spec %q: nested %q compositor is not allowed (any is depth-1 only, no nested any anywhere beneath another any)", spec.Name, legalAnyCompositorName)
			}
			// The subs of this "any" resolve with insideAny=true, whether
			// they are literally {Name: "any", ...} in spec.Sub or a
			// constructor that internally resolves an "any" of its own.
			boundResolve := func(sub LegalSpec) (*LegalPredicate, error) {
				return resolve(sub, true)
			}
			return resolveLegalAnySpec(spec, boundResolve)
		}

		ctor, ok := registry[spec.Name]
		if !ok {
			return nil, fmt.Errorf("boardgame: legal spec %q: unknown predicate name %q", spec.Name, spec.Name)
		}

		// The resolve closure handed to this constructor carries forward
		// the CURRENT insideAny — so if this constructor is itself being
		// resolved as a sub of an "any" (insideAny is already true), any
		// "any" it resolves internally is caught too.
		boundResolve := func(sub LegalSpec) (*LegalPredicate, error) {
			return resolve(sub, insideAny)
		}

		pred, err := ctor.Constructor(spec, chest, boundResolve)
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
		pred, err := resolve(spec, false)
		if err != nil {
			return nil, err
		}
		// Belt-and-suspenders: walk the resolved predicate tree itself,
		// independent of how it was built, and reject a nested "any" a
		// hand-built *LegalPredicate might have smuggled past resolve
		// entirely (a constructor that builds a LegalPredicate struct
		// literal rather than calling resolve).
		if err := checkNoNestedAny(pred); err != nil {
			return nil, fmt.Errorf("boardgame: legal spec %q: %w", spec.Name, err)
		}
		if err := validateLegalReadsForBoot(pred.Reads, exampleState, moveReader); err != nil {
			return nil, fmt.Errorf("boardgame: legal spec %q: %w", spec.Name, err)
		}
		result = append(result, pred)
	}
	return result, nil
}

// checkNoNestedAny walks pred's Sub tree (the already-resolved
// *LegalPredicate structure, not LegalSpec) and returns an error if any
// "any" node has another "any" node anywhere among its descendants. This is
// independent of, and a backstop for, resolve's insideAny enforcement
// above: it catches a constructor that hand-assembles a *LegalPredicate
// (Name: "any", Sub: [...]) containing a nested "any" without ever routing
// through the resolve closure it was given.
func checkNoNestedAny(pred *LegalPredicate) error {
	var walk func(p *LegalPredicate, insideAny bool) error
	walk = func(p *LegalPredicate, insideAny bool) error {
		if p == nil {
			return nil
		}
		isAny := p.Name == legalAnyCompositorName
		if isAny && insideAny {
			return fmt.Errorf("resolved predicate %q: nested %q compositor is not allowed (any is depth-1 only, no nested any anywhere beneath another any)", p.Name, legalAnyCompositorName)
		}
		for _, sub := range p.Sub {
			if err := walk(sub, insideAny || isAny); err != nil {
				return err
			}
		}
		return nil
	}
	return walk(pred, false)
}

// resolveLegalAnySpec resolves an "any"-named spec into the built-in
// any-compositor predicate: at least 2 subs required. Its Reads is the
// union of its children's Reads, its Cost the max of its children's Cost,
// and its Evaluate implements the Kleene truth table (spec §6): any child
// Pass -> Pass; else if any child Unknown -> Unknown; else Fail. Both
// non-Pass outcomes carry the spec's Message key if set, else the
// "legal.any_failed" template (see evalLegalAnyKleene).
//
// resolve is expected to already be bound (by the caller, resolveLegalSpecs)
// so that resolving any of spec's subs — including a nested "any", however
// it's reached — is treated as inside this "any"; resolveLegalAnySpec does
// not itself re-check subSpec.Name against "any" literally, since that
// literal check is bypassable by a constructor that resolves an "any"
// internally rather than naming it directly in Sub. See resolveLegalSpecs's
// doc comment for the full depth-1 enforcement rationale.
func resolveLegalAnySpec(spec LegalSpec, resolve func(LegalSpec) (*LegalPredicate, error)) (*LegalPredicate, error) {
	if len(spec.Sub) < 2 {
		return nil, fmt.Errorf("boardgame: legal spec %q: %q compositor requires at least 2 sub-specs, got %d", spec.Name, legalAnyCompositorName, len(spec.Sub))
	}

	subs := make([]*LegalPredicate, 0, len(spec.Sub))
	for _, subSpec := range spec.Sub {
		sub, err := resolve(subSpec)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}

	template := spec.Message
	if template == "" {
		template = legalAnyFailedTemplate
	}

	return &LegalPredicate{
		Name:             legalAnyCompositorName,
		Reads:            unionLegalReads(subs),
		Cost:             maxLegalCost(subs),
		Sub:              subs,
		EmittedTemplates: []string{template},
		Evaluate: func(ctx LegalContext) LegalVerdict {
			return evalLegalAnyKleene(subs, ctx, template)
		},
	}, nil
}

// evalLegalAnyKleene implements the Kleene truth table for the "any"
// compositor (spec §6): any child Pass -> Pass; else if any child Unknown
// -> Unknown; else Fail (all children Fail). Both non-Pass outcomes carry a
// LegalMessage built from template (the spec's Message override, or the
// "legal.any_failed" default) — LegalVerdict explicitly permits a Message on
// LegalUnknown, and an Unknown here means "could not confirm that any
// alternative holds", which renders to the player exactly like the all-Fail
// case; the Unknown's Reason additionally names the FIRST unknown
// sub-predicate for logs/debugging (footgun-batch F6: previously the Unknown
// verdict carried only a bare anonymous Reason and dropped the template
// entirely). Each child is evaluated via evalLegalPredicate so a panicking
// or invalid-verdict child degrades to Unknown for that child rather than
// propagating.
func evalLegalAnyKleene(subs []*LegalPredicate, ctx LegalContext, template string) LegalVerdict {
	sawUnknown := false
	var firstUnknown *LegalPredicate
	for _, sub := range subs {
		v := evalLegalPredicate(sub, ctx)
		switch v.Outcome {
		case LegalPass:
			return LegalVerdict{Outcome: LegalPass}
		case LegalUnknown:
			if !sawUnknown {
				firstUnknown = sub
			}
			sawUnknown = true
		}
	}
	if sawUnknown {
		name := "<nil>"
		if firstUnknown != nil {
			name = firstUnknown.Name
		}
		return LegalVerdict{
			Outcome: LegalUnknown,
			Message: &LegalMessage{Template: template},
			Reason:  fmt.Sprintf("sub-predicate %q of %q was unknown", name, legalAnyCompositorName),
		}
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
// path that parses as a move.* path OR a players[move.<Field>].* path (spec
// §3): the latter resolves to a value that depends on the move's own
// <Field>, so a predicate reading one is field-dependent by construction
// exactly like a bare move.* read — it must land in the plan's
// field-dependent bucket (buildLegalPlanFromPredicates, legal_plan.go) and
// count as a declared move read for evalLegalPredicate's nil-Move panic
// guard, above. A malformed path is treated as not a move path
// (parseLegalPath's error is ignored here — reads that failed to parse are
// a construction-time bug caught by resolveLegalSpecs' boot-time
// validation, not something evalLegalPredicate should mask or amplify).
func legalReadsIncludeMovePath(reads []LegalRead) bool {
	for _, r := range reads {
		if parsed, err := parseLegalPath(r.Path); err == nil && (parsed.kind == pathMove || parsed.kind == pathPlayersMoveField) {
			return true
		}
	}
	return false
}
