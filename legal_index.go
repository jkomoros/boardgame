package boardgame

import (
	"strconv"

	"github.com/jkomoros/boardgame/enum"
)

/*
This file implements design spec §5's "Phase bucketing" engine win: a
boot-assembled index letting the fixup loop (base/game_delegate.go's
ProposeFixUpMove, via GameManager.CandidateMoves) skip declaratively-
impossible moves with zero Legal() evaluations, without EVER excluding a
move that could still turn out to be legal — the SUPERSET PROPERTY, this
file's deliverable test (legal_index_test.go).

candidateMoves = phaseIndex[currentPhase ∪ TreeEnum ancestors] ∪ phaseAgnostic

phaseAgnostic contains every opaque (non-opted-in) move — an opaque move has
no plan to consult, so it must always be considered a candidate, which is
also why opaque moves see ZERO behavior change from this file: they always
land in phaseAgnostic, exactly as if this index didn't exist — plus every
opted-in move with no top-level "inPhase" atom (legal in every phase, same
semantics as LegalInPhaseCheck's own zero-length-legalPhases case).

phaseIndex is built from each opted-in move's assembled plan (legal_plan.go):
plans preserve their serializable spec list, and "inPhase" is one of the
stable framework atom names (design spec §2) — Name "inPhase", Args the
phase EnumKeys as decimal strings, contributed automatically by
moves.Default.ContributedPreconditions from WithLegalPhases, or authored
directly. See legalPlanInPhases below for the exact extraction and its
conservativeness rules.
*/

// legalInPhaseSpecName is the stable registry name of the framework "inPhase"
// atom (design spec §2), matching legal.InPhase's Spec.Name
// (legal/catalog_framework.go). Duplicated as a literal here (core cannot
// import package legal — see legal_registry.go's doc comment) rather than
// shared, the same layering trade every other framework-atom-name reference
// in this package makes.
const legalInPhaseSpecName = "inPhase"

// legalIndex is the boot-assembled phase index (design spec §5), built once
// by GameManager.buildLegalIndex (called from NewGameManager right after
// assembleLegalPlans, since it consumes g.legalPlans) and read-only
// thereafter.
type legalIndex struct {
	// phaseIndex maps a phase key to the names of opted-in move types whose
	// plan declares a top-level "inPhase" atom admitting that phase.
	phaseIndex map[enum.EnumKey][]string
	// phaseAgnostic holds the names of every move type that is a candidate
	// in EVERY phase: every opaque move, and every opted-in move with no
	// top-level "inPhase" atom.
	phaseAgnostic []string
}

// buildLegalIndex assembles g.legalIndex from g.moves/g.legalPlans (design
// spec §5's "Phase bucketing"). Must run after assembleLegalPlans, since it
// reads the plans assembleLegalPlans populates.
func (g *GameManager) buildLegalIndex() {
	idx := &legalIndex{phaseIndex: make(map[enum.EnumKey][]string)}

	for _, mType := range g.moves {
		name := mType.Name()

		plan := g.legalPlans[name]
		if plan == nil {
			// Opaque: never opted in to declarative legality. Always a
			// candidate — this is the superset property's base case.
			idx.phaseAgnostic = append(idx.phaseAgnostic, name)
			continue
		}

		phases := legalPlanInPhases(plan)
		if len(phases) == 0 {
			// Opted in, but declares no top-level inPhase atom: legal in
			// every phase (matches LegalInPhaseCheck's own
			// zero-length-legalPhases semantics), so also phase-agnostic.
			idx.phaseAgnostic = append(idx.phaseAgnostic, name)
			continue
		}

		for _, phase := range phases {
			idx.phaseIndex[phase] = append(idx.phaseIndex[phase], name)
		}
	}

	g.legalIndex = idx
}

// legalPlanInPhases extracts the union of phase keys declared by every
// TOP-LEVEL "inPhase" spec in plan's serializable spec list (legalPlan.specs,
// assembled by assembleLegalSpecList before bucket-splitting). It
// deliberately does not recurse into an "any" compositor's Sub: an inPhase
// nested inside an "any" does not by itself gate the move's legality (the
// "any" could still pass via a different sub-predicate regardless of phase),
// so treating such a move as having NO inPhase atom — landing it in
// phaseAgnostic — is the conservative, SAFE choice; it never excludes a move
// that could still be legal.
//
// A move with more than one top-level "inPhase" spec (unusual — one from
// WithLegalPhases plus a hand-authored one, say — but not forbidden) is
// indexed under the UNION of every declared phase. The move's ACTUAL legal
// phase set is those specs' INTERSECTION (every plan predicate must pass for
// the move to be legal), and union ⊇ intersection, so this over-
// approximates rather than under-approximates: still safe, per the superset
// property.
func legalPlanInPhases(plan *legalPlan) []enum.EnumKey {
	seen := make(map[enum.EnumKey]bool)
	var out []enum.EnumKey

	for _, spec := range plan.specs {
		if spec.Name != legalInPhaseSpecName {
			continue
		}
		for _, a := range spec.Args {
			n, err := strconv.Atoi(a)
			if err != nil {
				// A malformed inPhase arg would already have failed
				// boot-time resolution (legal/catalog_framework.go's
				// inPhaseConstructor parses every arg the same way and
				// errors there); unreachable in a booted manager. Skip
				// rather than propagate an error the index has no channel
				// to surface.
				continue
			}
			key := enum.EnumKey(n)
			if !seen[key] {
				seen[key] = true
				out = append(out, key)
			}
		}
	}

	return out
}

// CandidateMoves returns the moves that could possibly be legal against
// state's current phase (design spec §5): phaseIndex[currentPhase ∪
// TreeEnum ancestors] ∪ phaseAgnostic, in the SAME declaration order as
// g.moves (matching today's un-filtered iteration order exactly, so a fully
// un-migrated game — every move opaque, every move landing in
// phaseAgnostic — returns byte-identical results to iterating every move).
// Exported because base.GameDelegate.ProposeFixUpMove (a different package)
// is the v1 integration point (design spec §5 / the Task 9 runbook): the
// core fixup LOOP itself is untouched, only the candidate list it iterates
// is now pre-filtered.
//
// An uninitialized manager or a nil state has no legal index (and, in the
// nil-state case, no state to bucket by phase anyway), so this fails CLOSED
// in that case — returning nil rather than every move — since there is no
// well-defined "current phase" to look up. Once the manager IS initialized,
// though, a still-nil g.legalIndex (e.g. a hand-constructed *GameManager in
// a unit test that built moves without going through the normal
// buildLegalIndex path) fails OPEN instead — every move is a candidate —
// matching the pre-Task-9 behavior exactly rather than silently returning
// nothing.
func (g *GameManager) CandidateMoves(state ImmutableState) []Move {
	if !g.initialized || state == nil {
		return nil
	}

	if g.legalIndex == nil {
		result := make([]Move, 0, len(g.moves))
		for _, mType := range g.moves {
			if move := mType.NewMove(state); move != nil {
				result = append(result, move)
			}
		}
		return result
	}

	candidateNames := make(map[string]bool, len(g.legalIndex.phaseAgnostic))
	for _, name := range g.legalIndex.phaseAgnostic {
		candidateNames[name] = true
	}
	for _, phase := range legalCurrentPhaseAncestors(state) {
		for _, name := range g.legalIndex.phaseIndex[phase] {
			candidateNames[name] = true
		}
	}

	result := make([]Move, 0, len(candidateNames))
	for _, mType := range g.moves {
		if !candidateNames[mType.Name()] {
			continue
		}
		if move := mType.NewMove(state); move != nil {
			result = append(result, move)
		}
	}
	return result
}
