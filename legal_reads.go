package boardgame

import "fmt"

/*
This file makes a predicate's declared Reads ENFORCED rather than trusted
(docs/superpowers/specs/2026-11-22-legality-enforced-reads-design.md).

The soundness problem it closes: plan assembly sorts a predicate that declares
no move-dependent read into the field-independent bucket, whose verdict is
memoized keyed on (move type, state version, proposer) — WITHOUT the move's
field values. Reads used to be an honor system, so a predicate that read the
move anyway (unconditionally, conditionally, or through a concrete type
assertion on ctx.Move) made the server serve a verdict computed for some other
move's fields.

The fix is to make the memo key and the predicate's inputs the same thing by
construction, instead of hoping they agree:

 1. Move withholding. A predicate that declares no move-dependent read is
    evaluated with ctx.Move == nil, ALWAYS — in the memoized hot path, in
    full-ledger mode, and with no Game to memoize against. Its verdict can
    therefore only be a function of (state, proposer, chest), which is exactly
    what the memo key covers. An undeclared move read cannot be stale; it can
    only fail, deterministically and identically in tests and in production.

 2. Read scoping. ctx.ResolvePath resolves only the paths the predicate
    declared. Resolving any other path returns an error AND marks the
    evaluation violated; evalLegalPredicate then reports a fail-closed
    LegalUnknown naming the predicate and the path, regardless of what
    Evaluate went on to return. A predicate that swallows the error and
    returns Pass still cannot pass.

 3. Derived reads. A predicate that leaves Reads nil gets one LegalFacetValues
    read per Arg that parses as a legal path (see legalReadsFromArgs), so the
    common game-registered shape — args are paths, Evaluate resolves them —
    needs no hand-written Reads at all. Enforcement (2) makes a wrong
    derivation loud rather than silent.

Catalog predicates share the same LegalPredicate type and get the same
enforcement; nothing here distinguishes game-registered from built-in. The
opaque CustomLegaler wrapper is exempt: it is never memoized, always
field-dependent, and has no read-set to scope against.
*/

// legalReadScope is the boot-computed enforcement view of one predicate's
// declared Reads: which paths ctx.ResolvePath may resolve, and whether the
// predicate may see the move at all.
type legalReadScope struct {
	declared  map[LegalPropPath]bool
	readsMove bool
}

func newLegalReadScope(reads []LegalRead) *legalReadScope {
	scope := &legalReadScope{declared: make(map[LegalPropPath]bool, len(reads))}
	for _, r := range reads {
		scope.declared[r.Path] = true
		if legalPathReadsMove(r.Path) {
			scope.readsMove = true
		}
	}
	return scope
}

// legalReadEnforcement is the per-evaluation state carried on a LegalContext
// while evalLegalPredicate runs one predicate. It is a pointer so a
// ResolvePath call on a by-value copy of the context still records into the
// one evaluation it belongs to.
type legalReadEnforcement struct {
	predicate string
	scope     *legalReadScope
	// moveWithheld is true when the predicate declares no move-dependent
	// read, so ctx.Move was replaced with nil for this evaluation.
	moveWithheld bool
	// violation is the first out-of-scope path ResolvePath was asked for.
	violation LegalPropPath
}

func (e *legalReadEnforcement) check(p LegalPropPath) error {
	if e == nil || e.scope.declared[p] {
		return nil
	}
	if e.violation == "" {
		e.violation = p
	}
	if e.moveWithheld && legalPathReadsMove(p) {
		return fmt.Errorf("boardgame: legal path %q: predicate %q declares no move-dependent read, so it is evaluated without the move (declare %q in Reads, or leave Reads unset to derive it from Args)", p, e.predicate, p)
	}
	return fmt.Errorf("boardgame: legal path %q: predicate %q may only resolve the paths in its declared Reads (declare %q in Reads, or leave Reads unset to derive it from Args)", p, e.predicate, p)
}

func (e *legalReadEnforcement) violationVerdict() LegalVerdict {
	return LegalVerdict{
		Outcome: LegalUnknown,
		Reason:  fmt.Sprintf("undeclared read: predicate %q resolved %q, which is not in its declared Reads", e.predicate, e.violation),
	}
}

// legalPathReadsMove reports whether resolving p depends on the move's field
// values: a move.X path, a players[move.Field].X path, or any path carrying a
// [move.Field] board selector (game.Board[move.Space]) or component selector
// (game.Stack[move.Slot].Prop) — see parsedLegalPath.readsMove. A malformed
// path is not a move path; boot validation reports it separately.
func legalPathReadsMove(p LegalPropPath) bool {
	parsed, err := parseLegalPath(p)
	if err != nil {
		return false
	}
	return parsed.readsMove()
}

// legalReadsFromArgs derives a conservative read-set for a predicate that
// left Reads nil: one LegalFacetValues read for every arg that parses as a
// legal path. Args that are not path-shaped (literals, operators, phase
// names) are skipped. FacetValues is the most conservative facet, so the
// derivation can only over-approximate what a client needs.
func legalReadsFromArgs(args []string) []LegalRead {
	var reads []LegalRead
	seen := make(map[LegalPropPath]bool, len(args))
	for _, arg := range args {
		path := LegalPropPath(arg)
		if seen[path] {
			continue
		}
		if _, err := parseLegalPath(path); err != nil {
			continue
		}
		seen[path] = true
		reads = append(reads, LegalRead{Path: path, Facet: LegalFacetValues})
	}
	if reads == nil {
		// Non-nil and empty: the predicate's read-set has been decided.
		reads = []LegalRead{}
	}
	return reads
}

// sealLegalPredicate fills in derived Reads (when Reads is nil) and caches
// the enforcement scope for p and its Sub tree. It runs during single-threaded
// boot (plan resolution and assembly) so evaluation never mutates a predicate.
// It is idempotent.
func sealLegalPredicate(p *LegalPredicate, args []string) {
	if p == nil || p.opaque {
		return
	}
	if p.Reads == nil {
		if len(p.Args) > 0 {
			args = p.Args
		}
		p.Reads = legalReadsFromArgs(args)
	}
	if p.readScope == nil {
		p.readScope = newLegalReadScope(p.Reads)
	}
	for _, sub := range p.Sub {
		sealLegalPredicate(sub, sub.Args)
	}
}

// EvaluateLegalPredicate runs p against ctx exactly as a legality plan does:
// fail-closed on a nil, panicking, or zero-valued verdict, with the move
// withheld unless p declares a move-dependent read, and with ctx.ResolvePath
// scoped to p's declared Reads. Use it (legal.Evaluate in package legal) to
// unit-test a predicate under the same contract the engine enforces; calling
// p.Evaluate directly skips that enforcement.
func EvaluateLegalPredicate(p *LegalPredicate, ctx LegalContext) LegalVerdict {
	return evalLegalPredicate(p, ctx)
}
