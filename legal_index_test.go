package boardgame

import (
	"strconv"
	"testing"

	"github.com/jkomoros/boardgame/enum"
	"github.com/workfit/tester/assert"
)

/*
legal_index_test.go exercises design spec §5's phase-bucketing engine win
(legal_index.go): buildLegalIndex, legalPlanInPhases, and CandidateMoves. Its
deliverable test is TestCandidateMovesSupersetProperty — see that test's
doc comment.

This package (core) cannot import package moves or package legal (moves
imports both; legal imports core — see legal_registry.go's layering doc
comment), so the fixture below hand-rolls the two predicate shapes it needs
("inPhase", built directly from LegalInPhaseCheck, and a trivial
always-passing predicate) rather than reusing the real catalog. It also
hand-rolls a minimal legalDeclarer Move that reaches
LegalProbeActive/LegalEvaluatePlan exactly like moves.Default.Legal() does,
since moves.Default itself lives in the unimportable package moves.
*/

// registerLegalIndexTestPredicates registers this file's two fixture
// predicate constructors ("inPhase", using the SAME LegalInPhaseCheck the
// real catalog and the frozen chain both call — see legal_framework.go —
// and a trivial always-passing predicate) plus the one template key the
// "inPhase" fixture can emit. RegisterDefaultLegalPredicateConstructors and
// RegisterDefaultLegalTemplates are idempotent last-write-wins process
// globals, so calling this from multiple tests in this file is harmless.
func registerLegalIndexTestPredicates() {
	RegisterDefaultLegalPredicateConstructors(
		&LegalPredicateConstructor{
			Name: legalInPhaseSpecName,
			Constructor: func(spec LegalSpec, chest *ComponentChest, resolve func(LegalSpec) (*LegalPredicate, error)) (*LegalPredicate, error) {
				phases := make([]enum.EnumKey, len(spec.Args))
				for i, a := range spec.Args {
					n, err := strconv.Atoi(a)
					if err != nil {
						return nil, err
					}
					phases[i] = enum.EnumKey(n)
				}
				return &LegalPredicate{
					Name: legalInPhaseSpecName,
					Args: spec.Args,
					Reads: []LegalRead{
						{Path: LegalPropPath("game.Phase"), Facet: LegalFacetValues},
					},
					Cost:             LegalCostCheap,
					EmittedTemplates: []string{"legalindextest.in_phase"},
					Evaluate: func(ctx LegalContext) LegalVerdict {
						if ctx.State == nil {
							return LegalVerdict{Outcome: LegalUnknown, Reason: "state was nil"}
						}
						if err := LegalInPhaseCheck(ctx.State, phases); err != nil {
							return LegalVerdict{
								Outcome: LegalFail,
								Message: &LegalMessage{Template: "legalindextest.in_phase"},
							}
						}
						return LegalVerdict{Outcome: LegalPass}
					},
				}, nil
			},
		},
		&LegalPredicateConstructor{
			Name: "legalIndexTestAlwaysPass",
			Constructor: func(spec LegalSpec, chest *ComponentChest, resolve func(LegalSpec) (*LegalPredicate, error)) (*LegalPredicate, error) {
				return &LegalPredicate{
					Name: "legalIndexTestAlwaysPass",
					Evaluate: func(ctx LegalContext) LegalVerdict {
						return LegalVerdict{Outcome: LegalPass}
					},
				}, nil
			},
		},
	)
	RegisterDefaultLegalTemplates(map[string]string{
		"legalindextest.in_phase": "not legal in this phase",
	})
}

// legalIndexOpaqueMove is a plain, non-opted-in move (no DeclaredPreconditions
// method at all): it must ALWAYS be a CandidateMoves() result, in every
// phase — the superset property's base case — even though its OWN Legal()
// imperatively restricts itself to phaseSetUp. The index can't see (and
// mustn't need to see) into an opaque move's imperative logic; that's
// exactly why this move's Legal() is sometimes non-nil while it's always a
// candidate.
type legalIndexOpaqueMove struct {
	baseMove
}

func (m *legalIndexOpaqueMove) Reader() PropertyReader { return getDefaultReader(m) }
func (m *legalIndexOpaqueMove) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(m)
}
func (m *legalIndexOpaqueMove) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(m)
}
func (m *legalIndexOpaqueMove) Apply(state State) error { return nil }

func (m *legalIndexOpaqueMove) Legal(state ImmutableState, proposer PlayerIndex) error {
	return LegalInPhaseCheck(state, []enum.EnumKey{phaseSetUp})
}

var legalIndexOpaqueMoveConfig = NewMoveConfig(
	"Legal Index Opaque",
	func() Move { return new(legalIndexOpaqueMove) },
	nil,
)

// legalIndexDeclarerMove is a minimal Move implementing legalDeclarer
// directly. It hand-rolls the frozen-chain probe contract (design spec
// "prime guarantee" rule 4) moves.Default.Legal() implements: reach
// LegalProbeActive/LegalEvaluatePlan as literally the first thing Legal()
// does, so assembleLegalPlans's boot probe (legal_plan.go) can confirm its
// declarations are reachable.
type legalIndexDeclarerMove struct {
	baseMove
	authored []LegalSpec
}

func (m *legalIndexDeclarerMove) Reader() PropertyReader { return getDefaultReader(m) }
func (m *legalIndexDeclarerMove) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(m)
}
func (m *legalIndexDeclarerMove) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(m)
}
func (m *legalIndexDeclarerMove) Apply(state State) error { return nil }

func (m *legalIndexDeclarerMove) DeclaredPreconditions() ([]LegalSpec, []string) {
	return m.authored, nil
}

func (m *legalIndexDeclarerMove) Legal(state ImmutableState, proposer PlayerIndex) error {
	if manager := state.Manager(); manager != nil {
		if manager.LegalProbeActive() {
			return nil
		}
		if handled, err := manager.LegalEvaluatePlan(m.Info().Name(), state, m, proposer); handled {
			return err
		}
	}
	return nil
}

// legalIndexPhaseGatedMoveConfig is opted in with a top-level "inPhase"
// atom naming phaseNormal — legal (per the plan) only when the current
// phase is phaseNormal or a DESCENDANT of it (phaseNormalPlayerStart /
// phaseNormalActivateCard), via TreeEnum ancestor semantics.
var legalIndexPhaseGatedMoveConfig = NewMoveConfig(
	"Legal Index Phase Gated",
	func() Move {
		return &legalIndexDeclarerMove{
			authored: []LegalSpec{{Name: legalInPhaseSpecName, Args: []string{strconv.Itoa(int(phaseNormal))}}},
		}
	},
	nil,
)

// legalIndexAgnosticOptedInMoveConfig is opted in, but declares no
// "inPhase" atom at all — legal (per the plan) in every phase, and so
// phase-agnostic in the index too, despite being opted in.
var legalIndexAgnosticOptedInMoveConfig = NewMoveConfig(
	"Legal Index Agnostic Opted In",
	func() Move {
		return &legalIndexDeclarerMove{
			authored: []LegalSpec{{Name: "legalIndexTestAlwaysPass"}},
		}
	},
	nil,
)

// newLegalIndexTestManager boots a GameManager with the three fixture move
// types above (one opaque, one phase-gated opted-in, one phase-agnostic
// opted-in) installed on the standard core test delegate/phase-tree
// (testGameDelegate / testPhaseEnum, game_manager_test.go).
func newLegalIndexTestManager(t *testing.T) *GameManager {
	t.Helper()
	registerLegalIndexTestPredicates()

	moveInstaller := func(manager *GameManager) []MoveConfig {
		return []MoveConfig{
			legalIndexOpaqueMoveConfig,
			legalIndexPhaseGatedMoveConfig,
			legalIndexAgnosticOptedInMoveConfig,
		}
	}

	manager, err := NewGameManager(&testGameDelegate{moveInstaller: moveInstaller}, newTestStorageManager())
	assert.For(t).ThatActual(err).IsNil()
	return manager
}

// legalIndexStateAtPhase returns a fresh (never applied/committed) state
// for manager with GameState.Phase set to phaseKey. Bypassing the normal
// ProposeMove/applyMove pipeline means validateBeforeSave's leaf-phase
// check never runs, so this can set a non-leaf (branch) phase directly,
// which is exactly what the ancestor-lookup test below needs.
func legalIndexStateAtPhase(t *testing.T, manager *GameManager, phaseKey enum.EnumKey) ImmutableState {
	t.Helper()
	s, err := manager.emptyState(1)
	assert.For(t).ThatActual(err).IsNil()

	gs, ok := s.GameState().(*testGameState)
	assert.For(t).ThatActual(ok).Equals(true)
	assert.For(t).ThatActual(gs.Phase.SetValue(phaseKey)).IsNil()

	return s
}

// candidateMoveNames returns the Info().Name() of every move CandidateMoves
// returned, as a set, for convenient membership assertions.
func candidateMoveNames(moves []Move) map[string]bool {
	out := make(map[string]bool, len(moves))
	for _, m := range moves {
		out[m.Info().Name()] = true
	}
	return out
}

// TestCandidateMovesSupersetProperty is legal_index.go's deliverable test
// (design spec §5): across a mixed fixture of an opaque move, an opted-in
// move with an inPhase atom, and an opted-in move with none, CandidateMoves
// must be a SUPERSET of every move whose Legal() actually returns nil in
// that state — in every phase tried, including a phase reached only via
// TreeEnum ancestor walking.
func TestCandidateMovesSupersetProperty(t *testing.T) {
	manager := newLegalIndexTestManager(t)

	phasesToTry := []enum.EnumKey{
		phase, // root
		phaseSetUp,
		phaseNormal,
		phaseNormalPlayerStart,   // descendant of phaseNormal
		phaseNormalActivateCard, // descendant of phaseNormal
		phaseScoring,
	}

	for _, ph := range phasesToTry {
		state := legalIndexStateAtPhase(t, manager, ph)
		candidates := candidateMoveNames(manager.CandidateMoves(state))

		for _, mType := range manager.moves {
			move := mType.NewMove(state)
			assert.For(t, "phase", ph, "move", mType.Name()).ThatActual(move).IsNotNil()
			if move.Legal(state, AdminPlayerIndex) == nil {
				// Legal in this state: MUST be a candidate (the superset
				// property itself).
				assert.For(t, "phase", ph, "move", mType.Name()).ThatActual(candidates[mType.Name()]).Equals(true)
			}
		}
	}
}

// TestCandidateMovesOpaqueAlwaysIncluded proves the superset property's
// most load-bearing special case directly: an opaque move is a candidate in
// EVERY phase, even phases where its own (imperative, index-invisible)
// Legal() would reject it.
func TestCandidateMovesOpaqueAlwaysIncluded(t *testing.T) {
	manager := newLegalIndexTestManager(t)

	for _, ph := range []enum.EnumKey{phase, phaseSetUp, phaseNormal, phaseScoring} {
		state := legalIndexStateAtPhase(t, manager, ph)
		candidates := candidateMoveNames(manager.CandidateMoves(state))
		assert.For(t, "phase", ph).ThatActual(candidates["Legal Index Opaque"]).Equals(true)
	}
}

// TestCandidateMovesPhaseAgnosticOptedInAlwaysIncluded proves an opted-in
// move with no inPhase atom is a candidate in every phase too.
func TestCandidateMovesPhaseAgnosticOptedInAlwaysIncluded(t *testing.T) {
	manager := newLegalIndexTestManager(t)

	for _, ph := range []enum.EnumKey{phase, phaseSetUp, phaseNormal, phaseScoring} {
		state := legalIndexStateAtPhase(t, manager, ph)
		candidates := candidateMoveNames(manager.CandidateMoves(state))
		assert.For(t, "phase", ph).ThatActual(candidates["Legal Index Agnostic Opted In"]).Equals(true)
	}
}

// TestCandidateMovesPhaseGatedExcludedElsewhere proves the index actually
// FILTERS, not just includes: the phase-gated move is ABSENT from
// candidates in an unrelated phase (phaseScoring), even though it's present
// in its own declared phase and that phase's descendants.
func TestCandidateMovesPhaseGatedExcludedElsewhere(t *testing.T) {
	manager := newLegalIndexTestManager(t)

	state := legalIndexStateAtPhase(t, manager, phaseScoring)
	candidates := candidateMoveNames(manager.CandidateMoves(state))
	assert.For(t).ThatActual(candidates["Legal Index Phase Gated"]).Equals(false)

	state = legalIndexStateAtPhase(t, manager, phaseSetUp)
	candidates = candidateMoveNames(manager.CandidateMoves(state))
	assert.For(t).ThatActual(candidates["Legal Index Phase Gated"]).Equals(false)
}

// TestCandidateMovesAncestorLookup proves the TreeEnum ancestor walk (design
// spec §5's "∪ TreeEnum ancestors"): a move whose inPhase atom names
// phaseNormal (a branch node) is a candidate when the CURRENT phase is a
// DESCENDANT of phaseNormal (phaseNormalPlayerStart / phaseNormalActivateCard),
// not just when it's phaseNormal exactly.
func TestCandidateMovesAncestorLookup(t *testing.T) {
	manager := newLegalIndexTestManager(t)

	for _, ph := range []enum.EnumKey{phaseNormal, phaseNormalPlayerStart, phaseNormalActivateCard} {
		state := legalIndexStateAtPhase(t, manager, ph)
		candidates := candidateMoveNames(manager.CandidateMoves(state))
		assert.For(t, "phase", ph).ThatActual(candidates["Legal Index Phase Gated"]).Equals(true)
	}
}

// TestBuildLegalIndexBuckets is a direct, unit-level check of
// buildLegalIndex's bucket assignment, independent of CandidateMoves'
// lookup logic.
func TestBuildLegalIndexBuckets(t *testing.T) {
	manager := newLegalIndexTestManager(t)

	idx := manager.legalIndex
	assert.For(t).ThatActual(idx).IsNotNil()

	agnostic := make(map[string]bool, len(idx.phaseAgnostic))
	for _, n := range idx.phaseAgnostic {
		agnostic[n] = true
	}
	assert.For(t).ThatActual(agnostic["Legal Index Opaque"]).Equals(true)
	assert.For(t).ThatActual(agnostic["Legal Index Agnostic Opted In"]).Equals(true)
	assert.For(t).ThatActual(agnostic["Legal Index Phase Gated"]).Equals(false)

	assert.For(t).ThatActual(len(idx.phaseIndex[phaseNormal])).Equals(1)
	assert.For(t).ThatActual(idx.phaseIndex[phaseNormal][0]).Equals("Legal Index Phase Gated")
	assert.For(t).ThatActual(len(idx.phaseIndex[phaseScoring])).Equals(0)
}

// TestLegalPlanInPhasesUnionsMultipleTopLevelAtoms proves
// legalPlanInPhases's conservativeness rule directly: a plan with more than
// one top-level "inPhase" spec is indexed under the UNION of every declared
// phase (a safe over-approximation of the specs' actual intersection — see
// legalPlanInPhases's doc comment).
func TestLegalPlanInPhasesUnionsMultipleTopLevelAtoms(t *testing.T) {
	plan := &legalPlan{
		specs: []LegalSpec{
			{Name: legalInPhaseSpecName, Args: []string{strconv.Itoa(int(phaseSetUp))}},
			{Name: "somethingElse"},
			{Name: legalInPhaseSpecName, Args: []string{strconv.Itoa(int(phaseScoring)), strconv.Itoa(int(phaseSetUp))}},
		},
	}
	phases := legalPlanInPhases(plan)
	seen := make(map[enum.EnumKey]bool)
	for _, p := range phases {
		seen[p] = true
	}
	assert.For(t).ThatActual(len(phases)).Equals(2)
	assert.For(t).ThatActual(seen[phaseSetUp]).Equals(true)
	assert.For(t).ThatActual(seen[phaseScoring]).Equals(true)
}

// TestLegalPlanInPhasesIgnoresNestedAny proves the conservative "only
// top-level specs" rule: an "inPhase" spec nested inside an "any"
// compositor's Sub is NOT extracted (legalPlanInPhases only scans
// plan.specs, which is the top-level list — a Sub-nested spec never
// appears there), so such a plan reports no inPhase atom at all — landing
// the move in phaseAgnostic, the safe default.
func TestLegalPlanInPhasesIgnoresNestedAny(t *testing.T) {
	plan := &legalPlan{
		specs: []LegalSpec{
			{
				Name: legalAnyCompositorName,
				Sub: []LegalSpec{
					{Name: legalInPhaseSpecName, Args: []string{strconv.Itoa(int(phaseSetUp))}},
					{Name: "somethingElse"},
				},
			},
		},
	}
	phases := legalPlanInPhases(plan)
	assert.For(t).ThatActual(len(phases)).Equals(0)
}
