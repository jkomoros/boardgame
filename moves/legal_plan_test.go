package moves

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
	"github.com/jkomoros/boardgame/storage/memory"
	"github.com/workfit/tester/assert"
)

/*
This file holds Task 8's purely-sugar property tests (design spec §9): the
boot probe cases (b), the unsupported-base seam check (c), the registry-merge
boot test, and end-to-end opt-in plan evaluation. Test (a)'s floor — a game
using only frozen-chain moves is byte-identical — is the whole existing
go test ./... passing, plus TestLegalChainStringFreeze
(preconditions_test.go), which this task keeps green.

All test moves embed moves.Default or moves.CurrentPlayer with no persistable
fields, so they inherit base.Move's reflection-based reader (the same way
moveFreezeInPhase does) and need no generated reader. The unsupported-base
tests reuse the existing moveDealCards (embeds DealCountComponents) and
moveStartPhaseDrawAgain (embeds StartPhase) fixtures from moves_test.go.
*/

// alwaysFailPrecondition is a declarative precondition that can never pass for
// the shared moves-test fixture: game.Counter starts at 0, so requiring it to
// be at least 1000 always fails, and its template (TemplatePropAtLeast) is a
// registered default so it renders.
func alwaysFailPrecondition() legal.Spec {
	return legal.PropAtLeast("game.Counter", 1000)
}

// --- Fully declarative opt-in (no Legal override) ---

//boardgame:codegen
type moveDeclarativeOptIn struct {
	Default
}

func (m *moveDeclarativeOptIn) Apply(state boardgame.State) error { return nil }

// --- Wholesale Legal override that never super-calls (orphans declarations) ---

//boardgame:codegen
type moveWholesaleOverrideOptIn struct {
	Default
}

func (m *moveWholesaleOverrideOptIn) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	// Deliberately does NOT super-call Default.Legal — this orphans any
	// declared preconditions, which the boot probe must catch.
	return nil
}

func (m *moveWholesaleOverrideOptIn) Apply(state boardgame.State) error { return nil }

// --- Super-calling Legal override (declarations remain live) ---

//boardgame:codegen
type moveSuperCallOverrideOptIn struct {
	Default
}

func (m *moveSuperCallOverrideOptIn) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}
	// Own imperative residue would go here; the plan already evaluated inside
	// the super-call above.
	return nil
}

func (m *moveSuperCallOverrideOptIn) Apply(state boardgame.State) error { return nil }

// --- CustomLegaler implemented WITHOUT opting in (footgun batch F5) ---
//
// A move that implements boardgame.CustomLegaler (LegalCustom) but declares no
// WithPreconditions specs is not opted in, so no plan is assembled and its
// LegalCustom is never wrapped or consulted — the residue silently never runs
// (fails open). legal/doc.go documents this as a hard author requirement; the
// F5 boot check turns the honor-system requirement into a fail-closed boot
// error.

//boardgame:codegen
type moveCustomLegalerNoOptIn struct {
	Default
}

func (m *moveCustomLegalerNoOptIn) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	return nil
}

func (m *moveCustomLegalerNoOptIn) Apply(state boardgame.State) error { return nil }

// TestPurelySugarOptInEvaluatesPlan: a fully declarative move (no Legal
// override) has its plan evaluated in place of the frozen chain — an
// always-fail precondition makes Legal() return that failure, rendered
// through the game's template table.
func TestPurelySugarOptInEvaluatesPlan(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveDeclarativeOptIn),
				WithMoveName("Declarative Opt In"),
				WithPreconditions(alwaysFailPrecondition()),
			),
		)
	}

	manager, err := newGameManager(installer)
	assert.For(t, "manager").ThatActual(err).IsNil()
	if manager == nil {
		return
	}
	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	move := game.MoveByName("Declarative Opt In")
	assert.For(t, "move").ThatActual(move).IsNotNil()

	legalErr := move.Legal(game.CurrentState(), 0)
	assert.For(t, "plan failure surfaces").ThatActual(legalErr).IsNotNil()
	if legalErr != nil {
		// Rendered via TemplatePropAtLeast: "requires at least 1000, ...".
		assert.For(t, "rendered template").ThatActual(strings.Contains(legalErr.Error(), "1000")).IsTrue()
	}
}

// TestProbeWholesaleOverrideIsBootError (design spec §9(c) probe case): a move
// that declares preconditions but wholesale-overrides Legal() without
// super-calling orphans its declarations, so NewGameManager fails naming the
// move.
func TestProbeWholesaleOverrideIsBootError(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveWholesaleOverrideOptIn),
				WithMoveName("Wholesale Override Opt In"),
				WithPreconditions(alwaysFailPrecondition()),
			),
		)
	}

	_, err := newGameManager(installer)
	assert.For(t, "boot error").ThatActual(err).IsNotNil()
	if err != nil {
		assert.For(t, "names the move").ThatActual(strings.Contains(err.Error(), "Wholesale Override Opt In")).IsTrue()
		assert.For(t, "explains orphaning").ThatActual(strings.Contains(err.Error(), "Default.Legal")).IsTrue()
	}
}

// TestProbeSuperCallOverrideBootsAndEvaluates (design spec §9(c)): a move that
// declares preconditions and super-calls the chain in its Legal() override
// boots fine, AND its plan actually evaluates — proven by an always-fail
// precondition whose template surfaces through the override's super-call.
func TestProbeSuperCallOverrideBootsAndEvaluates(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveSuperCallOverrideOptIn),
				WithMoveName("Super Call Override Opt In"),
				WithPreconditions(alwaysFailPrecondition()),
			),
		)
	}

	manager, err := newGameManager(installer)
	assert.For(t, "boots").ThatActual(err).IsNil()
	if manager == nil {
		return
	}
	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	move := game.MoveByName("Super Call Override Opt In")
	legalErr := move.Legal(game.CurrentState(), 0)
	assert.For(t, "plan evaluates through super-call").ThatActual(legalErr).IsNotNil()
	if legalErr != nil {
		assert.For(t, "rendered template").ThatActual(strings.Contains(legalErr.Error(), "1000")).IsTrue()
	}
}

// TestUnsupportedBaseTypeStartPhaseBootsAndEvaluates (design spec §5, Task 6
// Step 2): moves.StartPhase joined the seam allowlist (it declares no Legal()
// override of its own — its legality IS Default.Legal, exactly like a bare
// Default-embedding move). A move embedding StartPhase with WithPreconditions
// now boots cleanly AND its plan actually evaluates, proven the same way
// TestProbeSuperCallOverrideBootsAndEvaluates proves it for a super-calling
// Default override: an always-fail declared precondition's template surfaces
// through Legal(). This test used to be
// TestUnsupportedBaseTypeStartPhaseIsBootError (StartPhase was outside the
// v1 seam); it inverts here because Task 6 widened the seam, not because the
// underlying mechanism changed.
func TestUnsupportedBaseTypeStartPhaseBootsAndEvaluates(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		// moveStartPhaseDrawAgain (moves_test.go) embeds StartPhase and
		// implements PhaseToStart, so it passes ValidConfiguration and
		// reaches the seam check.
		return Add(
			auto.MustConfig(
				new(moveStartPhaseDrawAgain),
				WithMoveName("Start Phase Opt In"),
				WithPreconditions(alwaysFailPrecondition()),
			),
		)
	}

	manager, err := newGameManager(installer)
	assert.For(t, "boots").ThatActual(err).IsNil()
	if manager == nil {
		return
	}

	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	move := game.MoveByName("Start Phase Opt In")
	assert.For(t, "move").ThatActual(move).IsNotNil()

	legalErr := move.Legal(game.CurrentState(), 0)
	assert.For(t, "plan evaluates").ThatActual(legalErr).IsNotNil()
	if legalErr != nil {
		assert.For(t, "rendered template").ThatActual(strings.Contains(legalErr.Error(), "1000")).IsTrue()
	}
}

// --- FixUp / FixUpMulti opt-in (design spec §5, Task 6 Step 2): both joined
// the seam allowlist alongside StartPhase, for the same reason — neither
// declares its own Legal() override. ---

//boardgame:codegen
type moveFixUpOptIn struct {
	FixUp
}

func (m *moveFixUpOptIn) Apply(state boardgame.State) error { return nil }

// TestFixUpBaseTypeBootsAndEvaluates: a move embedding moves.FixUp with
// WithPreconditions boots cleanly and its plan evaluates.
func TestFixUpBaseTypeBootsAndEvaluates(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveFixUpOptIn),
				WithMoveName("FixUp Opt In"),
				WithPreconditions(alwaysFailPrecondition()),
			),
		)
	}

	manager, err := newGameManager(installer)
	assert.For(t, "boots").ThatActual(err).IsNil()
	if manager == nil {
		return
	}

	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	move := game.MoveByName("FixUp Opt In")
	assert.For(t, "move").ThatActual(move).IsNotNil()

	legalErr := move.Legal(game.CurrentState(), 0)
	assert.For(t, "plan evaluates").ThatActual(legalErr).IsNotNil()
	if legalErr != nil {
		assert.For(t, "rendered template").ThatActual(strings.Contains(legalErr.Error(), "1000")).IsTrue()
	}
}

//boardgame:codegen
type moveFixUpMultiOptIn struct {
	FixUpMulti
}

func (m *moveFixUpMultiOptIn) Apply(state boardgame.State) error { return nil }

// TestFixUpMultiBaseTypeBootsAndEvaluates: a move embedding moves.FixUpMulti
// with WithPreconditions boots cleanly and its plan evaluates — the full
// opt-in path the design spec §5 FixUpMulti precondition (proven separately
// by moves/preconditions_test.go's TestFixUpMultiProgressionAtomEquivalence,
// which this test does not duplicate: that test proves the "inProgression"
// atom agrees with the frozen chain across a repeated-move tape; this test
// only proves a FixUpMulti-embedding move can opt in and its plan runs at
// all) now unlocks.
func TestFixUpMultiBaseTypeBootsAndEvaluates(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveFixUpMultiOptIn),
				WithMoveName("FixUpMulti Opt In"),
				WithPreconditions(alwaysFailPrecondition()),
			),
		)
	}

	manager, err := newGameManager(installer)
	assert.For(t, "boots").ThatActual(err).IsNil()
	if manager == nil {
		return
	}

	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	move := game.MoveByName("FixUpMulti Opt In")
	assert.For(t, "move").ThatActual(move).IsNotNil()

	legalErr := move.Legal(game.CurrentState(), 0)
	assert.For(t, "plan evaluates").ThatActual(legalErr).IsNotNil()
	if legalErr != nil {
		assert.For(t, "rendered template").ThatActual(strings.Contains(legalErr.Error(), "1000")).IsTrue()
	}
}

// TestUnsupportedBaseTypeDealCountComponentsIsBootError (design spec §9(c)):
// a move embedding moves.DealCountComponents (which has its own Legal logic)
// cannot opt in — boot fails naming the base type. This complements the
// StartPhase case: the seam check is embed-graph-based, so it catches both a
// base type without its own Legal override and one with.
func TestUnsupportedBaseTypeDealCountComponentsIsBootError(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		// moveDealCards (moves_test.go) embeds DealCountComponents and
		// implements the PlayerStacker/GameStacker/TargetCounter interfaces,
		// so it passes ValidConfiguration and reaches the seam check.
		return Add(
			auto.MustConfig(
				new(moveDealCards),
				WithMoveName("Deal Opt In"),
				WithPreconditions(alwaysFailPrecondition()),
			),
		)
	}

	_, err := newGameManager(installer)
	assert.For(t, "boot error").ThatActual(err).IsNotNil()
	if err != nil {
		assert.For(t, "names DealCountComponents").ThatActual(strings.Contains(err.Error(), "DealCountComponents")).IsTrue()
	}
}

// --- CurrentPlayer opt-in (contributed proposer atom in the plan) ---

//boardgame:codegen
type moveCurrentPlayerOptIn struct {
	CurrentPlayer
}

func (m *moveCurrentPlayerOptIn) Apply(state boardgame.State) error { return nil }

// TestCurrentPlayerOptInProposerAtom verifies the CurrentPlayer opt-in path
// Tasks 11+ depend on: a move embedding CurrentPlayer contributes the
// proposer atom into its plan, and CurrentPlayer.Legal's super-call evaluates
// that plan. The proposer atom (field-dependent) rejects a wrong proposer with
// the verbatim legacy "it's not your turn" string, and the correct proposer
// passes — identical to the frozen chain's behavior, now via the plan.
func TestCurrentPlayerOptInProposerAtom(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveCurrentPlayerOptIn),
				WithMoveName("Current Player Opt In"),
				// A trivially-satisfiable authored precondition, so the only
				// thing that can fail is the contributed proposer atom.
				WithPreconditions(legal.PropAtLeast("game.Counter", 0)),
			),
		)
	}

	manager, err := newGameManager(installer)
	assert.For(t, "boots").ThatActual(err).IsNil()
	if manager == nil {
		return
	}
	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	state := game.CurrentState()

	// Correct proposer (0), target defaults to current player 0: legal.
	correct := game.MoveByName("Current Player Opt In")
	assert.For(t, "correct proposer legal").ThatActual(correct.Legal(state, 0)).IsNil()

	// Wrong proposer/target: the contributed proposer atom fails with the
	// verbatim legacy string.
	wrong := game.MoveByName("Current Player Opt In").(*moveCurrentPlayerOptIn)
	wrong.TargetPlayerIndex = 1
	wrongErr := wrong.Legal(state, 1)
	assert.For(t, "wrong proposer illegal").ThatActual(wrongErr).IsNotNil()
	if wrongErr != nil {
		assert.For(t, "verbatim legacy string").ThatActual(wrongErr.Error()).Equals("it's not your turn")
	}
}

// --- CurrentPlayer opt-in suppressing the proposer atom (final-review boot
// guard: this suppression is ineffective on Legal() itself, since
// CurrentPlayer.Legal's imperative proposer check runs regardless of the
// plan) ---

//boardgame:codegen
type moveCurrentPlayerSuppressedProposer struct {
	CurrentPlayer
}

func (m *moveCurrentPlayerSuppressedProposer) Apply(state boardgame.State) error { return nil }

// TestCurrentPlayerSuppressedProposerIsBootError (final review finding):
// WithoutPrecondition(PreconditionProposerIsCurrentPlayer) on a move embedding
// CurrentPlayer would remove the contributed proposer atom from the plan
// (so the ledger/client would report the move legal for any proposer) while
// CurrentPlayer.Legal's imperative proposer-equivalence check keeps running
// unconditionally after its super-call — a ledger/actual divergence. Boot
// must reject this combination naming the move, rather than let a client
// silently disagree with the server about legality.
func TestCurrentPlayerSuppressedProposerIsBootError(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveCurrentPlayerSuppressedProposer),
				WithMoveName("Current Player Suppressed Proposer"),
				WithPreconditions(legal.PropAtLeast("game.Counter", 0)),
				WithoutPrecondition(PreconditionProposerIsCurrentPlayer),
			),
		)
	}

	_, err := newGameManager(installer)
	assert.For(t, "boot error").ThatActual(err).IsNotNil()
	if err != nil {
		assert.For(t, "names the move").ThatActual(strings.Contains(err.Error(), "Current Player Suppressed Proposer")).IsTrue()
		assert.For(t, "names the mechanism").ThatActual(strings.Contains(err.Error(), "proposerIsCurrentPlayer")).IsTrue()
		assert.For(t, "names CurrentPlayer").ThatActual(strings.Contains(err.Error(), "CurrentPlayer")).IsTrue()
	}
}

// TestDefaultSuppressedProposerIsBootError (contrast case, updated by the
// footgun batch's F2 suppression validation): the identical
// WithoutPrecondition(PreconditionProposerIsCurrentPlayer) call on a
// moves.Default-embedding move used to boot cleanly as a silent no-op —
// Default never contributes that atom, so the suppression matched nothing
// and was dropped at plan assembly. Under F2 an unmatched suppression name
// is a boot error (it can only ever be a typo or an opt-out of a check the
// move never had), so the same call now fails boot — via the GENERIC
// unmatched-name guard, not the CurrentPlayer-specific divergence guard its
// counterpart above exercises. The two errors are deliberately distinct:
// this one says the atom was never contributed; that one says the atom IS
// contributed but suppressing it would desynchronize the client ledger from
// CurrentPlayer.Legal's imperative check.
func TestDefaultSuppressedProposerIsBootError(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveDeclarativeOptIn),
				WithMoveName("Default Suppressed Proposer"),
				WithPreconditions(legal.PropAtLeast("game.Counter", 0)),
				WithoutPrecondition(PreconditionProposerIsCurrentPlayer),
			),
		)
	}

	_, err := newGameManager(installer)
	assert.For(t, "boot error").ThatActual(err).IsNotNil()
	if err != nil {
		assert.For(t, "names the move").ThatActual(strings.Contains(err.Error(), "Default Suppressed Proposer")).IsTrue()
		assert.For(t, "generic unmatched-name guard, not the CurrentPlayer guard").ThatActual(strings.Contains(err.Error(), "contributes no spec with that name")).IsTrue()
	}
}

// --- Registry-merge boot test ---

//boardgame:codegen
type moveProgressionOptInA struct {
	Default
}

func (m *moveProgressionOptInA) Apply(state boardgame.State) error { return nil }

// TestRegistryMergeInProgressionBoots is the merge test the Task 8 brief
// mandates: a move that opts in AND sits in an ordered phase contributes an
// "inProgression" spec, whose constructor lives in package moves
// (FrameworkConstructors), not package legal. If NewGameManager's registry
// assembly failed to merge FrameworkConstructors alongside
// legal.DefaultConstructors, this move would boot-fail with "unknown
// predicate name inProgression". A clean boot proves the merge.
func TestRegistryMergeInProgressionBoots(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Combine(
			AddOrderedForPhase(phaseSetUp,
				auto.MustConfig(
					new(StartPhase),
					WithMoveName("Registry Merge Start Normal Play"),
					WithPhaseToStart(phaseNormalPlayDrawCard, phaseEnum),
					WithIsFixUp(false),
				),
			),
			AddOrderedForPhase(phaseNormalPlayDrawCard,
				// This move opts in (WithPreconditions) AND, via
				// AddOrderedForPhase, gets a move progression — so its
				// contributed specs include "inProgression".
				auto.MustConfig(
					new(moveProgressionOptInA),
					WithMoveName("Registry Merge Progression A"),
					WithPreconditions(legal.PropAtLeast("game.Counter", 0)),
				),
				auto.MustConfig(
					new(NoOp),
					WithMoveName("Registry Merge Guard"),
				),
			),
		)
	}

	manager, err := newGameManager(installer)
	assert.For(t, "boots with merged FrameworkConstructors").ThatActual(err).IsNil()
	assert.For(t, "manager").ThatActual(manager != nil).IsTrue()
}

// TestContributedOnlyIsNotOptedIn (design spec §2 "declaring is
// implementing"): a move configured with WithLegalPhases but NO
// WithPreconditions does NOT opt in, so it is not probed and runs its frozen
// chain. Proven by pairing WithLegalPhases with a wholesale Legal() override:
// were it treated as opted-in, the probe would boot-fail (as the wholesale
// override case does); because contributions alone don't opt in, it boots.
func TestContributedOnlyIsNotOptedIn(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveWholesaleOverrideOptIn),
				WithMoveName("Contributed Only Not Opted In"),
				// Contributes an inPhase spec, but declares NO
				// WithPreconditions — so it is not opted in.
				WithLegalPhases(phaseNormalPlayDrawCard),
			),
		)
	}

	_, err := newGameManager(installer)
	assert.For(t, "boots (not opted in, no probe)").ThatActual(err).IsNil()
}

// --- WithoutPrecondition boot validation (footgun batch F2) ---
//
// WithoutPrecondition is fire-and-forget config: before these guards, a
// suppression whose name matched nothing (a typo, or a check the move never
// contributed) was silently dropped at plan assembly, and a suppression on a
// move that never opted in via WithPreconditions was dead config while the
// frozen chain kept enforcing the "suppressed" check. All three flavors are
// now boot errors naming the move.

// TestSuppressionTypoIsBootError (F2 flavor 1a): a WithoutPrecondition name
// that matches no contributed spec name because it is misspelled ("inphase"
// for "inPhase") must fail boot naming the move, the unmatched name, and the
// move's actual contributed names — not silently no-op while the phase check
// keeps running.
func TestSuppressionTypoIsBootError(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveDeclarativeOptIn),
				WithMoveName("Suppression Typo"),
				WithLegalPhases(phaseSetUp),
				WithPreconditions(legal.PropAtLeast("game.Counter", 0)),
				WithoutPrecondition("inphase"), // typo: real name is "inPhase"
			),
		)
	}

	_, err := newGameManager(installer)
	assert.For(t, "boot error").ThatActual(err).IsNotNil()
	if err != nil {
		assert.For(t, "names the move").ThatActual(strings.Contains(err.Error(), "Suppression Typo")).IsTrue()
		assert.For(t, "names the unmatched name").ThatActual(strings.Contains(err.Error(), `"inphase"`)).IsTrue()
		assert.For(t, "lists the contributed names").ThatActual(strings.Contains(err.Error(), `"inPhase"`)).IsTrue()
	}
}

// TestSuppressionNotContributedIsBootError (F2 flavor 1b): a perfectly valid
// framework name ("inProgression") suppressed on a move that never
// contributes it (no move progression configured) must fail boot — the
// author believes they opted out of a check that was never going to run,
// which almost certainly means they meant a different name or a different
// move.
func TestSuppressionNotContributedIsBootError(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveDeclarativeOptIn),
				WithMoveName("Suppression Not Contributed"),
				WithLegalPhases(phaseSetUp),
				WithPreconditions(legal.PropAtLeast("game.Counter", 0)),
				WithoutPrecondition(PreconditionInProgression),
			),
		)
	}

	_, err := newGameManager(installer)
	assert.For(t, "boot error").ThatActual(err).IsNotNil()
	if err != nil {
		assert.For(t, "names the move").ThatActual(strings.Contains(err.Error(), "Suppression Not Contributed")).IsTrue()
		assert.For(t, "names the unmatched name").ThatActual(strings.Contains(err.Error(), `"inProgression"`)).IsTrue()
	}
}

// TestSuppressionWithoutOptInIsBootError (F2 flavor 2): WithoutPrecondition
// on a move with NO authored WithPreconditions specs is dead config — the
// move is not opted in, no plan exists, and the frozen imperative chain
// runs unchanged (whether or not any of its checks corresponds to the
// suppressed name). Boot must reject it rather than let the author believe
// the check is off.
func TestSuppressionWithoutOptInIsBootError(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveDeclarativeOptIn),
				WithMoveName("Suppression Without Opt In"),
				WithLegalPhases(phaseSetUp),
				// NO WithPreconditions: not opted in, so this suppression
				// could never take effect.
				WithoutPrecondition(PreconditionInPhase),
			),
		)
	}

	_, err := newGameManager(installer)
	assert.For(t, "boot error").ThatActual(err).IsNotNil()
	if err != nil {
		assert.For(t, "names the move").ThatActual(strings.Contains(err.Error(), "Suppression Without Opt In")).IsTrue()
		assert.For(t, "points at the missing opt-in").ThatActual(strings.Contains(err.Error(), "WithPreconditions")).IsTrue()
	}
}

// TestSuppressionWithoutOptInListsAllNames (wave-1 review M1): when a
// not-opted-in move carries MULTIPLE dead suppressions, the flavor-2 boot
// error must report every dead name, not just the first — an author fixing
// the error by deleting only the named call would just trade one boot error
// for another, once per remaining suppression.
func TestSuppressionWithoutOptInListsAllNames(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveDeclarativeOptIn),
				WithMoveName("Suppression Multi Without Opt In"),
				WithLegalPhases(phaseSetUp),
				// NO WithPreconditions: not opted in, so BOTH suppressions
				// are dead config.
				WithoutPrecondition(PreconditionInPhase),
				WithoutPrecondition(PreconditionInProgression),
			),
		)
	}

	_, err := newGameManager(installer)
	assert.For(t, "boot error").ThatActual(err).IsNotNil()
	if err != nil {
		assert.For(t, "names the move").ThatActual(strings.Contains(err.Error(), "Suppression Multi Without Opt In")).IsTrue()
		assert.For(t, "lists the first dead name").ThatActual(strings.Contains(err.Error(), `"inPhase"`)).IsTrue()
		assert.For(t, "lists the second dead name").ThatActual(strings.Contains(err.Error(), `"inProgression"`)).IsTrue()
	}
}

// --- CustomLegaler-without-opt-in boot validation (footgun batch F5) ---

// TestCustomLegalerWithoutOptInIsBootError (F5): a move that implements
// boardgame.CustomLegaler (LegalCustom) but declares no WithPreconditions is
// not opted in — no plan is assembled, so LegalCustom is never wrapped and
// never runs. Every check the author put in LegalCustom silently stops being
// enforced (fails open), with zero boot signal. legal/doc.go documents that a
// LegalCustom move must opt in via WithPreconditions; boot must enforce it,
// naming the move and pointing at the missing opt-in.
func TestCustomLegalerWithoutOptInIsBootError(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveCustomLegalerNoOptIn),
				WithMoveName("Custom Legaler Without Opt In"),
				WithLegalPhases(phaseSetUp),
				// NO WithPreconditions: not opted in, so LegalCustom is dead.
			),
		)
	}

	_, err := newGameManager(installer)
	assert.For(t, "boot error").ThatActual(err).IsNotNil()
	if err != nil {
		assert.For(t, "names the move").ThatActual(strings.Contains(err.Error(), "Custom Legaler Without Opt In")).IsTrue()
		assert.For(t, "names LegalCustom").ThatActual(strings.Contains(err.Error(), "LegalCustom")).IsTrue()
		assert.For(t, "points at the missing opt-in").ThatActual(strings.Contains(err.Error(), "WithPreconditions")).IsTrue()
	}
}

// TestValidSuppressionBootsAndSuppresses (F2 contrast case): a suppression
// that names a check the move actually contributes is the supported pattern
// and must keep working — the game boots, and the suppressed check really
// is gone from the plan (the move is legal even though the game is NOT in
// the move's configured phase).
func TestValidSuppressionBootsAndSuppresses(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveDeclarativeOptIn),
				WithMoveName("Valid Suppression"),
				// The game never enters phaseDrawAgain in these tests, so if
				// the contributed inPhase atom survived, Legal() would fail.
				WithLegalPhases(phaseDrawAgain),
				WithPreconditions(legal.PropAtLeast("game.Counter", 0)),
				WithoutPrecondition(PreconditionInPhase),
			),
		)
	}

	manager, err := newGameManager(installer)
	assert.For(t, "boots").ThatActual(err).IsNil()
	if manager == nil {
		return
	}
	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	move := game.MoveByName("Valid Suppression")
	assert.For(t, "move").ThatActual(move).IsNotNil()
	// Legal despite the wrong phase: the inPhase atom was suppressed.
	assert.For(t, "phase check suppressed").ThatActual(move.Legal(game.CurrentState(), 0)).IsNil()
}

// --- Footgun-batch F4: template placeholder / emitted-binding boot checks ---

// legalBindingTestDelegate embeds the standard moves-test delegate and adds
// the optional legal.TemplateConfigurer surface, so the F4 boot tests below
// can register game templates whose bodies reference placeholders and prove
// the boot gauntlet validates those placeholders against the owning
// predicate's declared emitted bindings.
type legalBindingTestDelegate struct {
	gameDelegate
	templates map[string]string
}

func (d *legalBindingTestDelegate) ConfigureLegalTemplates() map[string]string {
	return d.templates
}

// TestBootValidatesTemplatePlaceholdersAgainstEmittedBindings (footgun-batch
// F4): a WithMessage retarget pointing a catalog predicate at a game
// template whose body references a placeholder the predicate never emits is
// a boot error naming the move, the template key, and the missing binding —
// today that placeholder would render as its own bare name mid-game
// ("you need frobs more"), silently garbled. Templates whose placeholders
// are all genuinely emitted must keep booting.
func TestBootValidatesTemplatePlaceholdersAgainstEmittedBindings(t *testing.T) {
	newManagerWithTemplates := func(moveName string, templates map[string]string, specs ...legal.Spec) (*boardgame.GameManager, error) {
		installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
			auto := NewAutoConfigurer(manager.Delegate())
			return Add(
				auto.MustConfig(
					new(moveDeclarativeOptIn),
					WithMoveName(moveName),
					WithPreconditions(specs...),
				),
			)
		}
		return boardgame.NewGameManager(&legalBindingTestDelegate{
			gameDelegate{moveInstaller: installer},
			templates,
		}, memory.NewStorageManager())
	}

	t.Run("placeholder the predicate never emits is a boot error naming move, key, and binding", func(t *testing.T) {
		_, err := newManagerWithTemplates(
			"Binding Validated Move",
			map[string]string{"test.bad_binding": "you need {frobs} more"},
			legal.PropAtLeast("game.Counter", 1000).WithMessage("test.bad_binding"),
		)
		assert.For(t, "boot error").ThatActual(err).IsNotNil()
		if err == nil {
			return
		}
		for _, want := range []string{"Binding Validated Move", "test.bad_binding", "frobs"} {
			assert.For(t, "error names", want).ThatActual(strings.Contains(err.Error(), want)).IsTrue()
		}
	})

	t.Run("placeholders the predicate does emit boot cleanly", func(t *testing.T) {
		_, err := newManagerWithTemplates(
			"Binding Valid Move",
			map[string]string{"test.good_binding": "have {value}, need at least {min}"},
			legal.PropAtLeast("game.Counter", 1000).WithMessage("test.good_binding"),
		)
		assert.For(t, "boots").ThatActual(err).IsNil()
	})

	t.Run("any-compositor WithMessage retarget with a placeholder is a boot error", func(t *testing.T) {
		// The "any" compositor never attaches bindings to its Fail/Unknown
		// Message, so ANY placeholder in a retargeted template body is
		// unrenderable and must be rejected at boot.
		_, err := newManagerWithTemplates(
			"Any Binding Move",
			map[string]string{"test.any_bad": "none of it worked: {detail}"},
			legal.Any(
				legal.PropAtLeast("game.Counter", 1000),
				legal.StackNotEmpty("game.DrawStack"),
			).WithMessage("test.any_bad"),
		)
		assert.For(t, "boot error").ThatActual(err).IsNotNil()
		if err == nil {
			return
		}
		for _, want := range []string{"Any Binding Move", "test.any_bad", "detail"} {
			assert.For(t, "error names", want).ThatActual(strings.Contains(err.Error(), want)).IsTrue()
		}
	})

	t.Run("overriding a catalog DEFAULT template's body with an unemitted placeholder is a boot error", func(t *testing.T) {
		// ConfigureLegalTemplates overrides of a catalog default key are
		// validated too: the resolved body is what renders, wherever it came
		// from.
		_, err := newManagerWithTemplates(
			"Default Override Move",
			map[string]string{legal.TemplatePropAtLeast: "you need {frobs} more"},
			legal.PropAtLeast("game.Counter", 1000),
		)
		assert.For(t, "boot error").ThatActual(err).IsNotNil()
		if err == nil {
			return
		}
		for _, want := range []string{"Default Override Move", legal.TemplatePropAtLeast, "frobs"} {
			assert.For(t, "error names", want).ThatActual(strings.Contains(err.Error(), want)).IsTrue()
		}
	})
}

// legalRegisteredBindingTestDelegate embeds the standard moves-test delegate
// and adds BOTH optional legal surfaces — ConstructorConfigurer and
// TemplateConfigurer — so the tests below can register a game-registered
// predicate carrying EmittedBindings metadata alongside the game templates
// its keys resolve to.
type legalRegisteredBindingTestDelegate struct {
	gameDelegate
	ctors     []*legal.PredicateConstructor
	templates map[string]string
}

func (d *legalRegisteredBindingTestDelegate) ConfigurePredicateConstructors() []*legal.PredicateConstructor {
	return d.ctors
}

func (d *legalRegisteredBindingTestDelegate) ConfigureLegalTemplates() map[string]string {
	return d.templates
}

// bindingMetadataConstructor returns a game-registered predicate constructor
// whose predicates declare exactly the given EmittedTemplates/EmittedBindings
// metadata and whose Evaluate fails with the first emitted key carrying a
// single {who} binding — the minimal shape for exercising the F4
// placeholder/binding boot checks through the game-registered (rather than
// catalog) path.
func bindingMetadataConstructor(name string, emittedTemplates []string,
	emittedBindings map[string][]string) *legal.PredicateConstructor {
	return &legal.PredicateConstructor{
		Name: name,
		Constructor: func(spec legal.Spec, _ *boardgame.ComponentChest,
			_ func(legal.Spec) (*legal.Predicate, error)) (*legal.Predicate, error) {
			return &legal.Predicate{
				Name: name,
				Args: spec.Args,
				Reads: []legal.Read{
					{Path: "game.Counter", Facet: boardgame.LegalFacetValues},
				},
				Cost:             boardgame.LegalCostTrivial,
				EmittedTemplates: emittedTemplates,
				EmittedBindings:  emittedBindings,
				Evaluate: func(ctx legal.Context) legal.Verdict {
					return legal.FailT(emittedTemplates[0], map[string]legal.BindingValue{
						"who": legal.String("someone"),
					})
				},
			}, nil
		},
	}
}

// TestBootValidatesGameRegisteredEmittedBindings: the F4 placeholder/binding
// boot checks through a GAME-REGISTERED predicate that declares
// EmittedBindings metadata (the catalog path is
// TestBootValidatesTemplatePlaceholdersAgainstEmittedBindings above; the
// metadata-free game-registered path deliberately skips validation and is
// pinned by the root package's TestValidateLegalEmittedBindings). Also pins
// the malformed-metadata boot error: an EmittedBindings key absent from
// EmittedTemplates used to be silently skipped by validation entirely.
func TestBootValidatesGameRegisteredEmittedBindings(t *testing.T) {
	newRegisteredBindingManager := func(moveName string, templates map[string]string,
		ctor *legal.PredicateConstructor, spec legal.Spec) (*boardgame.GameManager, error) {
		installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
			auto := NewAutoConfigurer(manager.Delegate())
			return Add(
				auto.MustConfig(
					new(moveDeclarativeOptIn),
					WithMoveName(moveName),
					WithPreconditions(spec),
				),
			)
		}
		return boardgame.NewGameManager(&legalRegisteredBindingTestDelegate{
			gameDelegate{moveInstaller: installer},
			[]*legal.PredicateConstructor{ctor},
			templates,
		}, memory.NewStorageManager())
	}

	t.Run("template body referencing a binding not in the metadata is a boot error", func(t *testing.T) {
		_, err := newRegisteredBindingManager(
			"Registered Binding Move",
			map[string]string{"test.registered_binding": "blocked by {villain}"},
			bindingMetadataConstructor("test.registeredBinding",
				[]string{"test.registered_binding"},
				map[string][]string{"test.registered_binding": {"who"}}),
			legal.Spec{Name: "test.registeredBinding"},
		)
		assert.For(t, "boot error").ThatActual(err).IsNotNil()
		if err == nil {
			return
		}
		for _, want := range []string{"Registered Binding Move", "test.registered_binding", "villain"} {
			assert.For(t, "error names", want).ThatActual(strings.Contains(err.Error(), want)).IsTrue()
		}
	})

	t.Run("EmittedBindings key absent from EmittedTemplates is a boot error", func(t *testing.T) {
		// The metadata key is typo'd ("bidning"): before the
		// malformed-metadata check, the typo'd entry was never looked at AND
		// the real key was skipped as metadata-free, so the {villain}
		// mismatch above would sail through boot unvalidated.
		_, err := newRegisteredBindingManager(
			"Typo Metadata Move",
			map[string]string{"test.registered_binding": "blocked by {villain}"},
			bindingMetadataConstructor("test.typoMetadata",
				[]string{"test.registered_binding"},
				map[string][]string{"test.registered_bidning": {"who"}}),
			legal.Spec{Name: "test.typoMetadata"},
		)
		assert.For(t, "boot error").ThatActual(err).IsNotNil()
		if err == nil {
			return
		}
		for _, want := range []string{"Typo Metadata Move", "test.typoMetadata", "test.registered_bidning", "EmittedTemplates"} {
			assert.For(t, "error names", want).ThatActual(strings.Contains(err.Error(), want)).IsTrue()
		}
	})

	t.Run("matching metadata and template body boot cleanly", func(t *testing.T) {
		manager, err := newRegisteredBindingManager(
			"Valid Registered Binding Move",
			map[string]string{"test.registered_binding": "blocked by {who}"},
			bindingMetadataConstructor("test.validBinding",
				[]string{"test.registered_binding"},
				map[string][]string{"test.registered_binding": {"who"}}),
			legal.Spec{Name: "test.validBinding"},
		)
		assert.For(t, "boots").ThatActual(err).IsNil()
		assert.For(t, "manager").ThatActual(manager != nil).IsTrue()
	})
}

// --- Footgun-batch F3: boot smoke probe catching a game-registered
// predicate that reads move properties without declaring any move.* Read ---

// legalProbeTestDelegate embeds the standard moves-test delegate and adds the
// optional legal.ConstructorConfigurer surface, so the F3 probe tests below
// can register game-registered predicates and prove the boot gauntlet probes
// their field-independent ones against a sentinel move.
type legalProbeTestDelegate struct {
	gameDelegate
	ctors []*legal.PredicateConstructor
}

func (d *legalProbeTestDelegate) ConfigurePredicateConstructors() []*legal.PredicateConstructor {
	return d.ctors
}

// newProbeTestManager builds a manager whose delegate registers ctors and
// whose single configured move is whatever moveConfig returns.
func newProbeTestManager(ctors []*legal.PredicateConstructor,
	moveConfig func(auto *AutoConfigurer) boardgame.MoveConfig) (*boardgame.GameManager, error) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(moveConfig(auto))
	}
	return boardgame.NewGameManager(&legalProbeTestDelegate{
		gameDelegate{moveInstaller: installer},
		ctors,
	}, memory.NewStorageManager())
}

// probeTestConstructor returns a game-registered predicate constructor whose
// predicates declare the given reads and evaluate with eval.
func probeTestConstructor(name string, reads []legal.Read,
	eval func(ctx legal.Context) legal.Verdict) *legal.PredicateConstructor {
	return &legal.PredicateConstructor{
		Name: name,
		Constructor: func(spec legal.Spec, _ *boardgame.ComponentChest,
			_ func(legal.Spec) (*legal.Predicate, error)) (*legal.Predicate, error) {
			return &legal.Predicate{
				Name:     name,
				Args:     spec.Args,
				Reads:    reads,
				Cost:     boardgame.LegalCostTrivial,
				Evaluate: eval,
			}, nil
		},
	}
}

// sneakyMoveReadConstructor is the F3 footgun in miniature: a game-registered
// predicate whose Evaluate reads a move property while declaring NO Reads at
// all. Plan assembly sorts it into the field-independent bucket, whose
// verdict is memoized without the move's fields in the key — so before the
// boot probe, the server itself would serve stale verdicts as the move's
// fields changed.
func sneakyMoveReadConstructor() *legal.PredicateConstructor {
	return probeTestConstructor("test.sneakyMoveRead", nil,
		func(ctx legal.Context) legal.Verdict {
			if _, _, err := ctx.ResolvePath("move.TargetPlayerIndex"); err != nil {
				return legal.UnknownVerdict(err.Error())
			}
			return legal.PassVerdict()
		})
}

// TestBootProbeCatchesUndeclaredMoveRead (footgun-batch F3): at boot, every
// field-independent game-registered predicate is evaluated once against the
// example state with a sentinel move whose PropertyReader panics on every
// property access; a predicate that touches the move despite declaring no
// move.* Read is a boot error naming the move and the predicate.
func TestBootProbeCatchesUndeclaredMoveRead(t *testing.T) {
	t.Run("top-level field-independent predicate", func(t *testing.T) {
		_, err := newProbeTestManager(
			[]*legal.PredicateConstructor{sneakyMoveReadConstructor()},
			func(auto *AutoConfigurer) boardgame.MoveConfig {
				return auto.MustConfig(
					new(moveDeclarativeOptIn),
					WithMoveName("Sneaky Move Read"),
					WithPreconditions(legal.Spec{Name: "test.sneakyMoveRead"}),
				)
			},
		)
		assert.For(t, "boot error").ThatActual(err).IsNotNil()
		if err == nil {
			return
		}
		for _, want := range []string{"Sneaky Move Read", "test.sneakyMoveRead", "move.*"} {
			assert.For(t, "error names", want).ThatActual(strings.Contains(err.Error(), want)).IsTrue()
		}
	})

	t.Run("game-registered predicate nested inside an any compositor", func(t *testing.T) {
		_, err := newProbeTestManager(
			[]*legal.PredicateConstructor{sneakyMoveReadConstructor()},
			func(auto *AutoConfigurer) boardgame.MoveConfig {
				return auto.MustConfig(
					new(moveDeclarativeOptIn),
					WithMoveName("Sneaky Any Move Read"),
					WithPreconditions(legal.Any(
						legal.Spec{Name: "test.sneakyMoveRead"},
						legal.PropAtLeast("game.Counter", 1000),
					)),
				)
			},
		)
		assert.For(t, "boot error").ThatActual(err).IsNotNil()
		if err == nil {
			return
		}
		for _, want := range []string{"Sneaky Any Move Read", "test.sneakyMoveRead"} {
			assert.For(t, "error names", want).ThatActual(strings.Contains(err.Error(), want)).IsTrue()
		}
	})
}

// TestBootProbeAllowsStateOnlyGameRegistered (F3 contrast case): a
// game-registered field-independent predicate that genuinely never touches
// the move IS probed but never trips the sentinel, so the game boots.
func TestBootProbeAllowsStateOnlyGameRegistered(t *testing.T) {
	stateOnly := probeTestConstructor("test.stateOnly",
		[]legal.Read{{Path: "game.Counter", Facet: boardgame.LegalFacetValues}},
		func(ctx legal.Context) legal.Verdict {
			if _, _, err := ctx.ResolvePath("game.Counter"); err != nil {
				return legal.UnknownVerdict(err.Error())
			}
			return legal.PassVerdict()
		})

	manager, err := newProbeTestManager(
		[]*legal.PredicateConstructor{stateOnly},
		func(auto *AutoConfigurer) boardgame.MoveConfig {
			return auto.MustConfig(
				new(moveDeclarativeOptIn),
				WithMoveName("State Only Game Registered"),
				WithPreconditions(legal.Spec{Name: "test.stateOnly"}),
			)
		},
	)
	assert.For(t, "boots").ThatActual(err).IsNil()
	assert.For(t, "manager").ThatActual(manager != nil).IsTrue()
}

// TestBootProbeSkipsFieldDependentGameRegistered (F3 contrast case): a
// game-registered predicate that reads the move AND declares that read lands
// in the field-dependent bucket, which the probe does not touch — declaring
// honestly is exactly what the probe exists to encourage.
func TestBootProbeSkipsFieldDependentGameRegistered(t *testing.T) {
	declared := probeTestConstructor("test.declaredMoveRead",
		[]legal.Read{{Path: "move.TargetPlayerIndex", Facet: boardgame.LegalFacetValues}},
		func(ctx legal.Context) legal.Verdict {
			if _, _, err := ctx.ResolvePath("move.TargetPlayerIndex"); err != nil {
				return legal.UnknownVerdict(err.Error())
			}
			return legal.PassVerdict()
		})

	manager, err := newProbeTestManager(
		[]*legal.PredicateConstructor{declared},
		func(auto *AutoConfigurer) boardgame.MoveConfig {
			// moveCurrentPlayerOptIn embeds CurrentPlayer, so
			// move.TargetPlayerIndex exists for boot path validation.
			return auto.MustConfig(
				new(moveCurrentPlayerOptIn),
				WithMoveName("Declared Move Read"),
				WithPreconditions(legal.Spec{Name: "test.declaredMoveRead"}),
			)
		},
	)
	assert.For(t, "boots").ThatActual(err).IsNil()
	assert.For(t, "manager").ThatActual(manager != nil).IsTrue()
}

// TestBootProbeIgnoresUnrelatedPanics (F3 precision case): the probe must
// only fire on the sentinel's own panic — a game-registered predicate that
// panics at boot for some unrelated reason keeps booting (at runtime
// evalLegalPredicate degrades that panic to a fail-closed Unknown, exactly
// as before the probe existed).
func TestBootProbeIgnoresUnrelatedPanics(t *testing.T) {
	panicky := probeTestConstructor("test.unrelatedPanic", nil,
		func(ctx legal.Context) legal.Verdict {
			panic("boom: nothing to do with the move")
		})

	manager, err := newProbeTestManager(
		[]*legal.PredicateConstructor{panicky},
		func(auto *AutoConfigurer) boardgame.MoveConfig {
			return auto.MustConfig(
				new(moveDeclarativeOptIn),
				WithMoveName("Unrelated Panic"),
				WithPreconditions(legal.Spec{Name: "test.unrelatedPanic"}),
			)
		},
	)
	assert.For(t, "boots despite unrelated panic").ThatActual(err).IsNil()
	assert.For(t, "manager").ThatActual(manager != nil).IsTrue()
}
