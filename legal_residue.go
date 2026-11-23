package boardgame

import (
	stderrors "errors"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
)

/*
This file is the core half of the widened composition seam (design spec
docs/superpowers/specs/2026-11-22-legality-seam-widening-design.md).

Every framework base type in package moves expresses its legality as its
contribution. Where a base type's own check cannot be a catalog relation --
it calls game-supplied Go such as FinishTurn's TurnDone or ApplyUntil's
ConditionMet -- it contributes a RESIDUE atom: LegalCustom with a name and a
position. A residue atom is

  - opaque: unserializable, never client-evaluable, never memoized (its reads
    cannot be declared; some read the move's own fields);
  - rendered exactly like a LegalCustom error, so the frozen chain's error
    text survives byte-for-byte;
  - movable but not removable: suppressing it requires re-authoring it
    (legalValidateResidueAtoms), the same "reorder, don't delete" recipe the
    component verbs' defining atoms follow;
  - reached: the boot probe requires every residue atom's owning Legal()
    level to actually be on the move's call path (legalProbeReachResidue), so
    a game override that super-calls a non-immediate ancestor cannot silently
    gain the skipped level's check by opting in.
*/

// NewLegalResidueConstructor registers a framework base type's imperative
// check as a named, opaque plan atom: the LegalCustom escape hatch,
// contributable by a base type at a fixed position in its contribution. The
// spec takes no args and no WithMessage; check receives the evaluation
// context (ctx.Move is the concrete move) and its error is translated exactly
// as a LegalCustom error is (a *LegalError keeps its structured verdict; any
// other error renders as its own text), so a migrated chain keeps its
// first-failure messages.
//
// It exists for package moves' framework atoms, which wrap code the frozen
// chain also runs. A game's own imperative residue belongs in LegalCustom.
func NewLegalResidueConstructor(name string, check func(ctx LegalContext) error) *LegalPredicateConstructor {
	return &LegalPredicateConstructor{
		Name:    name,
		residue: true,
		Constructor: func(spec LegalSpec, chest *ComponentChest, resolve func(LegalSpec) (*LegalPredicate, error)) (*LegalPredicate, error) {
			if len(spec.Args) != 0 || len(spec.Sub) != 0 {
				return nil, fmt.Errorf("framework check %q takes no args", name)
			}
			if spec.Message != "" {
				return nil, fmt.Errorf("framework check %q renders the base type's own error text and cannot take WithMessage; wrap the check in LegalCustom if you need a different message", name)
			}
			return &LegalPredicate{
				Name:    name,
				Cost:    LegalCostExpensive,
				opaque:  true,
				residue: true,
				Evaluate: func(ctx LegalContext) LegalVerdict {
					return legalVerdictFromError(check(ctx))
				},
			}, nil
		},
	}
}

// IsResidue reports whether c was built by NewLegalResidueConstructor: a
// framework base type's opaque, server-only check with no client semantics
// (so, for example, no conformance-corpus fixture applies to it).
func (c *LegalPredicateConstructor) IsResidue() bool {
	return c != nil && c.residue
}

// legalVerdictFromError is the one translation from an imperative error to a
// verdict, shared by the LegalCustom wrapper and residue atoms.
func legalVerdictFromError(err error) LegalVerdict {
	if err == nil {
		return LegalVerdict{Outcome: LegalPass}
	}
	var le *LegalError
	if stderrors.As(err, &le) {
		return le.Verdict
	}
	return LegalVerdict{
		Outcome: LegalFail,
		Message: &LegalMessage{Template: err.Error()},
	}
}

// legalPredicateOpaque reports whether p, or any predicate in its Sub tree,
// is opaque. An opaque predicate's read-set is unknown, so plan assembly must
// never memoize it -- including when it is nested in a compositor, whose
// declared Reads are only the union of its children's.
func legalPredicateOpaque(p *LegalPredicate) bool {
	if p == nil {
		return false
	}
	if p.opaque {
		return true
	}
	for _, sub := range p.Sub {
		if legalPredicateOpaque(sub) {
			return true
		}
	}
	return false
}

// legalNotEvaluatedVerdict is what full-ledger evaluation reports for a
// residue atom after an earlier failure: residues assume their predecessors
// passed (an admin check indexing a target an earlier atom validated), and
// the hot path would never have run them.
func legalNotEvaluatedVerdict() LegalVerdict {
	return LegalVerdict{Outcome: LegalUnknown, Reason: "not evaluated: an earlier precondition failed"}
}

// LegalProbeReachResidue is engine-internal plumbing for the boot probe. A
// framework Legal() level that owns residue atom name calls it (through
// moves' legalPlanOwnsChain) after its super-call. During a probe it records
// that the level is on this move's call path and returns true; otherwise it
// returns false. Like LegalProbeActive, it is only written during
// NewGameManager's single-threaded boot.
func (g *GameManager) LegalProbeReachResidue(name string) bool {
	if g == nil || !g.legalProbing {
		return false
	}
	g.legalProbeReached = true
	if g.legalProbeReachedResidues == nil {
		g.legalProbeReachedResidues = make(map[string]bool)
	}
	g.legalProbeReachedResidues[name] = true
	return true
}

// legalProbeCall runs the probe's Legal call, containing any panic. During a
// probe every framework Legal() level returns right after its super-call, so
// a game override's own code after its super-call runs on a move no
// framework check normalized (moves.AnyPlayer, for one, would have defaulted
// TargetPlayerIndex) and may panic. The probe only observes which levels were
// reached, and those were recorded before any such panic.
func legalProbeCall(move Move, exampleState ImmutableState) {
	defer func() {
		_ = recover()
	}()
	_ = move.Legal(exampleState, ObserverPlayerIndex)
}

// legalProbeUnreachedResidues returns, sorted, every residue atom in plan
// whose owning Legal() level the probe did not reach.
func legalProbeUnreachedResidues(plan *legalPlan, reached map[string]bool) []string {
	var missing []string
	for _, step := range plan.ordered {
		if step.predicate.residue && !reached[step.predicate.Name] {
			missing = append(missing, step.predicate.Name)
		}
	}
	sort.Strings(missing)
	return missing
}

func legalProbeUnreachedResidueError(moveName string, missing []string) error {
	return fmt.Errorf("move %q opts in to declarative legality, but its Legal() never runs the framework Legal() level that owns %s: your override super-calls a further ancestor and skips that level, so opting in would add a check your frozen chain never ran. Super-call the immediately embedded framework move's Legal() instead", moveName, strings.Join(quoteAll(missing), ", "))
}

func quoteAll(names []string) []string {
	out := make([]string, len(names))
	for i, name := range names {
		out[i] = fmt.Sprintf("%q", name)
	}
	return out
}

// legalValidateResidueAtoms enforces "move, don't remove" for residue atoms,
// generically: a suppressed residue atom must be re-authored (suppression
// only reorders it), and an authored residue atom must be one this move
// contributes and suppresses (anywhere else its owning level is absent, so it
// could never pass).
func legalValidateResidueAtoms(moveName string, contributed, authored []LegalSpec, suppressions []string, registry map[string]*LegalPredicateConstructor) error {
	isResidue := func(name string) bool {
		ctor := registry[name]
		return ctor != nil && ctor.residue
	}

	suppressed := make(map[string]bool, len(suppressions))
	for _, name := range suppressions {
		suppressed[name] = true
	}
	contributedNames := make(map[string]bool, len(contributed))
	for _, spec := range contributed {
		contributedNames[spec.Name] = true
	}
	authoredNames := make(map[string]bool, len(authored))
	for _, spec := range authored {
		authoredNames[spec.Name] = true
	}

	for _, name := range suppressions {
		if isResidue(name) && !authoredNames[name] {
			return fmt.Errorf("move %q suppresses %q, a framework base type's own legality check, but never re-authors it: suppression can only MOVE a framework check (re-author it with moves.FrameworkPrecondition in WithLegalPreconditions at the position you want), never remove it -- to skip it, embed a different base type", moveName, name)
		}
	}
	for _, spec := range authored {
		for _, sub := range spec.Sub {
			if isResidue(sub.Name) {
				return fmt.Errorf("move %q nests framework check %q inside %q: a framework check is a base type's own legality and only runs at its own position in the plan, never as an alternative inside a compositor", moveName, sub.Name, spec.Name)
			}
		}
		if !isResidue(spec.Name) {
			continue
		}
		if !contributedNames[spec.Name] || !suppressed[spec.Name] {
			return fmt.Errorf("move %q authors framework check %q, which only makes sense to re-position a check this move already contributes: it must embed the base type that contributes %q and suppress it with WithoutLegalPrecondition", moveName, spec.Name, spec.Name)
		}
	}
	return nil
}

// LegalWithPlansSuspended runs fn with this manager's assembled plans made
// invisible to the frozen chain: LegalPlanAssembled and LegalEvaluatePlan
// report "no plan", so every opted-in move's Legal() runs its frozen
// imperative chain instead. It is engine-internal plumbing for differential
// golden tests (frozen vs plan on the same state); nothing else should call
// it. It is safe to call concurrently with itself but affects every
// goroutine using this manager while fn runs.
func (g *GameManager) LegalWithPlansSuspended(fn func()) {
	atomic.AddInt32(&g.legalPlansSuspended, 1)
	defer atomic.AddInt32(&g.legalPlansSuspended, -1)
	fn()
}

func (g *GameManager) legalPlansAreSuspended() bool {
	return atomic.LoadInt32(&g.legalPlansSuspended) > 0
}
