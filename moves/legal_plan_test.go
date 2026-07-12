package moves

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
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

// TestUnsupportedBaseTypeStartPhaseIsBootError (design spec §9(c) / §2 seam):
// a move embedding moves.StartPhase (no Legal override of its own) cannot opt
// in — boot fails naming the base type.
func TestUnsupportedBaseTypeStartPhaseIsBootError(t *testing.T) {
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

	_, err := newGameManager(installer)
	assert.For(t, "boot error").ThatActual(err).IsNotNil()
	if err != nil {
		assert.For(t, "names StartPhase").ThatActual(strings.Contains(err.Error(), "StartPhase")).IsTrue()
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
// WithoutPrecondition("proposerIsCurrentPlayer") on a move embedding
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
				WithoutPrecondition("proposerIsCurrentPlayer"),
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

// TestDefaultSuppressedProposerStillBoots (final review finding, contrast
// case): the identical WithoutPrecondition("proposerIsCurrentPlayer") call
// on a moves.Default-embedding move (which never contributes that atom in
// the first place, and has no imperative proposer check to diverge from)
// still boots cleanly — the boot guard above is CurrentPlayer-specific, not
// a blanket ban on suppressing this name. moveDeclaredPreconditions
// (contributed_preconditions_test.go) already exercises this exact
// suppression on a Default-embedding move via
// TestWithPreconditionsRoundTrip; this test pins the "boots without error"
// half directly and by name, next to its CurrentPlayer counterpart above.
func TestDefaultSuppressedProposerStillBoots(t *testing.T) {
	installer := func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(
			auto.MustConfig(
				new(moveDeclarativeOptIn),
				WithMoveName("Default Suppressed Proposer"),
				WithPreconditions(legal.PropAtLeast("game.Counter", 0)),
				WithoutPrecondition("proposerIsCurrentPlayer"),
			),
		)
	}

	_, err := newGameManager(installer)
	assert.For(t, "boots (Default has no proposer check to diverge from)").ThatActual(err).IsNil()
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
