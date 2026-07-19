package moves

import (
	"github.com/jkomoros/boardgame/legal"
)

/*
This file wires Default, CurrentPlayer, RecordCurrentPlayerChoice, FixUp,
FixUpMulti, and StartPhase into
the declarative legality composition seam: PreconditionsProvider is the
optional interface core consults (a later task) to derive a move type's
plan, base-first contributed specs (ContributedPreconditions) plus authored
specs from WithLegalPreconditions (DeclaredPreconditions), minus
WithoutLegalPrecondition suppressions.

Other framework move types (DealCountComponents, FinishTurn, RoundRobin, etc.)
remain opaque; configuring them declaratively is a named boot error.
*/

// The stable names of the framework-contributed precondition checks — the
// exact names moves.Default/CurrentPlayer's ContributedPreconditions specs
// carry, and therefore the only names WithoutLegalPrecondition can meaningfully
// suppress. Pass these constants instead of raw string literals: a
// misspelled or non-contributed name is a boot error (NewGameManager's
// gauntlet validates every suppression against the move's actual
// contributed spec names), and the constants make that class of typo a
// compile-time non-issue.
type PreconditionName string

const (
	// PreconditionInPhase is the contributed phase check, present when the
	// move was configured with WithLegalPhases (including via
	// AddForPhase/AddOrderedForPhase).
	PreconditionInPhase PreconditionName = "inPhase"
	// PreconditionInProgression is the contributed move-progression check,
	// present when the move was configured with WithLegalMoveProgression
	// (including via AddOrderedForPhase).
	PreconditionInProgression PreconditionName = "inProgression"
	// PreconditionStackConstraints is the contributed source/destination
	// stack-size check, present when the move was configured with BOTH
	// WithSourceProperty and WithDestinationProperty.
	PreconditionStackConstraints PreconditionName = "stackConstraints"
	// PreconditionProposerIsCurrentPlayer is the proposer-equivalence check
	// moves.CurrentPlayer contributes on top of Default's specs. Note that
	// suppressing it on a CurrentPlayer-embedding move is a boot error in
	// its own right (CurrentPlayer.Legal's imperative check would still
	// run, desynchronizing the client ledger — embed moves.Default
	// instead), and on a Default-embedding move it is never contributed at
	// all, so suppressing it there is an unmatched-name boot error too.
	PreconditionProposerIsCurrentPlayer PreconditionName = "proposerIsCurrentPlayer"
)

// PreconditionsProvider is the optional interface core consults (design spec
// §2/§3) to derive a move type's declarative precondition plan. Default
// implements it directly; CurrentPlayer overrides it to append the proposer
// atom on top of Default's own contributions.
type PreconditionsProvider interface {
	// ContributedPreconditions returns this move type's own base-first
	// specs, derived from whatever legality configuration it was given via
	// auto.Config — the exact same configuration bag the frozen imperative
	// chain already reads. A move type that received none of that
	// configuration returns nil.
	ContributedPreconditions() []legal.Spec
}

// ContributedPreconditions derives inPhase/inProgression/stackConstraints
// specs from whatever legality configuration was passed to auto.Config, ONLY
// for the configuration keys actually present: a move type configured with
// just WithLegalPhases contributes only an inPhase spec, not a zero-value
// inProgression/stackConstraints one alongside it. This mirrors, and is
// derived directly from, the exact same config bag moves.Default's frozen
// Legal() chain reads — legalPhases (WithLegalPhases), legalMoveProgression
// (WithLegalMoveProgression), and sourceProperty+destinationProperty
// (WithSourceProperty/WithDestinationProperty; see moves/with.go:8-31 and
// default.go's Legal()). Specs are returned base-first, in the same
// deterministic order the frozen chain evaluates them in (phase,
// progression, stack constraints) — see the design spec §2's "Plan assembly"
// note and §4's "base-first" ordering rule.
func (d *Default) ContributedPreconditions() []legal.Spec {

	var specs []legal.Spec

	config := d.CustomConfiguration()

	if _, ok := config[configPropLegalPhases]; ok {
		if phases := d.legalPhases(); len(phases) > 0 {
			specs = append(specs, legal.InPhase(phases...))
		}
	}

	if _, ok := config[configPropLegalMoveProgression]; ok {
		if group := d.legalMoveProgression(); group != nil {
			specs = append(specs, inProgressionSpec(d.Name()))
		}
	}

	srcName, hasSrc := config[configPropSourceProperty].(string)
	dstName, hasDst := config[configPropDestinationProperty].(string)
	if hasSrc && hasDst {
		specs = append(specs, legal.StackConstraints(srcName, dstName))
	}

	return specs
}

// DeclaredPreconditions returns this move type's authored specs (from
// WithLegalPreconditions, in declaration order) and suppression names (from
// WithoutLegalPrecondition), as configured via auto.Config. A nil specs return
// means there are no authored specs; LegalPlanEnabled separately reports
// whether configuration explicitly requested a plan.
func (d *Default) DeclaredPreconditions() ([]legal.Spec, []string) {
	config := d.CustomConfiguration()

	specs, _ := config[configPropPreconditions].([]legal.Spec)
	suppressions, _ := config[configPropSuppressedPreconditions].([]string)

	return specs, suppressions
}

// LegalPlanEnabled reports whether WithLegalPreconditions or
// WithoutLegalPrecondition requested a plan. It distinguishes an intentional
// zero-authored-spec plan from a move that never opted in.
func (d *Default) LegalPlanEnabled() bool {
	enabled, _ := d.CustomConfiguration()[configPropLegalPlanEnabled].(bool)
	return enabled
}

// ContributedPreconditions returns Default's own contributed specs
// (inPhase/inProgression/stackConstraints, derived from configuration) plus
// legal.ProposerIsCurrentPlayer() appended last — CurrentPlayer.Legal()'s
// proposer checks, beyond its Default.Legal() super-call (design spec §2).
func (c *CurrentPlayer) ContributedPreconditions() []legal.Spec {
	return append(c.Default.ContributedPreconditions(), legal.ProposerIsCurrentPlayer())
}

// Compile-time interface satisfaction checks.
var _ PreconditionsProvider = (*Default)(nil)
var _ PreconditionsProvider = (*CurrentPlayer)(nil)
