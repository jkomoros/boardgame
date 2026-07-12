package moves

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
	"github.com/workfit/tester/assert"
)

/*
This file tests moves.Default.ContributedPreconditions,
moves.CurrentPlayer.ContributedPreconditions, and
moves.Default.DeclaredPreconditions (design spec §2/§3) — the plan-assembly
data a later task's NewGameManager wiring will consume. It reuses the shared
gameState/playerState/phaseEnum/newGameManager fixture from game_test.go, but
with its own small, purpose-built move types and installer so each test
exercises EXACTLY one configuration combination (no config at all / phase
only / stack constraint only / phase + progression / CurrentPlayer's extra
proposer atom), which the shared freezeMoveInstaller fixture
(preconditions_test.go) doesn't isolate as cleanly (its moves compose
multiple config keys, or route every move through AddOrderedForPhase, which
always sets BOTH legalPhases and legalMoveProgression).
*/

//boardgame:codegen
type moveContribNone struct {
	Default
}

func (m *moveContribNone) Apply(state boardgame.State) error {
	return nil
}

//boardgame:codegen
type moveContribPhaseOnly struct {
	Default
}

func (m *moveContribPhaseOnly) Apply(state boardgame.State) error {
	return nil
}

//boardgame:codegen
type moveContribStackOnly struct {
	Default
}

func (m *moveContribStackOnly) Apply(state boardgame.State) error {
	return nil
}

//boardgame:codegen
type moveContribProgression struct {
	Default
}

func (m *moveContribProgression) Apply(state boardgame.State) error {
	return nil
}

//boardgame:codegen
type moveContribCurrentPlayer struct {
	CurrentPlayer
}

func (m *moveContribCurrentPlayer) Apply(state boardgame.State) error {
	return nil
}

func contributedPreconditionsMoveInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {
	auto := NewAutoConfigurer(manager.Delegate())
	return Combine(
		Add(
			auto.MustConfig(
				new(moveContribNone),
				WithMoveName("Contrib None"),
			),
			auto.MustConfig(
				new(moveContribPhaseOnly),
				WithMoveName("Contrib Phase Only"),
				WithLegalPhases(phaseSetUp),
			),
			auto.MustConfig(
				new(moveContribStackOnly),
				WithMoveName("Contrib Stack Only"),
				WithSourceProperty("DrawStack"),
				WithDestinationProperty("DiscardStack"),
			),
			auto.MustConfig(
				new(moveContribCurrentPlayer),
				WithMoveName("Contrib Current Player"),
				WithLegalPhases(phaseSetUp),
			),
		),
		AddOrderedForPhase(phaseDrawAgain,
			auto.MustConfig(
				new(moveContribProgression),
				WithMoveName("Contrib Progression"),
			),
			auto.MustConfig(new(NoOp), WithMoveName("Contrib Progression Guard")),
		),
	)
}

func TestDefaultContributedPreconditionsEmpty(t *testing.T) {
	manager, err := newGameManager(contributedPreconditionsMoveInstaller)
	assert.For(t, "manager").ThatActual(err).IsNil()
	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	move := game.MoveByName("Contrib None").(*moveContribNone)
	specs := move.ContributedPreconditions()
	assert.For(t, "no config -> no contributed specs").ThatActual(len(specs)).Equals(0)
}

func TestDefaultContributedPreconditionsPhaseOnly(t *testing.T) {
	manager, err := newGameManager(contributedPreconditionsMoveInstaller)
	assert.For(t, "manager").ThatActual(err).IsNil()
	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	move := game.MoveByName("Contrib Phase Only").(*moveContribPhaseOnly)
	specs := move.ContributedPreconditions()
	assert.For(t, "spec count").ThatActual(len(specs)).Equals(1)
	assert.For(t, "spec name").ThatActual(specs[0].Name).Equals("inPhase")
	assert.For(t, "spec args").ThatActual(len(specs[0].Args)).Equals(1)
}

func TestDefaultContributedPreconditionsStackOnly(t *testing.T) {
	manager, err := newGameManager(contributedPreconditionsMoveInstaller)
	assert.For(t, "manager").ThatActual(err).IsNil()
	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	move := game.MoveByName("Contrib Stack Only").(*moveContribStackOnly)
	specs := move.ContributedPreconditions()
	assert.For(t, "spec count").ThatActual(len(specs)).Equals(1)
	assert.For(t, "spec name").ThatActual(specs[0].Name).Equals("stackConstraints")
	assert.For(t, "spec args").ThatActual(specs[0].Args).Equals([]string{"DrawStack", "DiscardStack"})
}

func TestDefaultContributedPreconditionsPhaseAndProgression(t *testing.T) {
	manager, err := newGameManager(contributedPreconditionsMoveInstaller)
	assert.For(t, "manager").ThatActual(err).IsNil()
	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	move := game.MoveByName("Contrib Progression").(*moveContribProgression)
	specs := move.ContributedPreconditions()
	assert.For(t, "spec count").ThatActual(len(specs)).Equals(2)
	// Base-first, deterministic order: inPhase before inProgression (design
	// spec §2's "Plan assembly" note / §4's base-first ordering rule) —
	// matching the frozen chain's own evaluation order (legalInPhase before
	// legalMoveInProgression, see moves/default.go's Legal()).
	assert.For(t, "spec 0 name").ThatActual(specs[0].Name).Equals("inPhase")
	assert.For(t, "spec 1 name").ThatActual(specs[1].Name).Equals("inProgression")
	assert.For(t, "spec 1 args").ThatActual(specs[1].Args).Equals([]string{"Contrib Progression"})
}

func TestCurrentPlayerContributedPreconditions(t *testing.T) {
	manager, err := newGameManager(contributedPreconditionsMoveInstaller)
	assert.For(t, "manager").ThatActual(err).IsNil()
	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	move := game.MoveByName("Contrib Current Player").(*moveContribCurrentPlayer)
	specs := move.ContributedPreconditions()
	// Default's own contribution (inPhase, from WithLegalPhases) plus
	// legal.ProposerIsCurrentPlayer() appended LAST (design spec §2).
	assert.For(t, "spec count").ThatActual(len(specs)).Equals(2)
	assert.For(t, "spec 0 name").ThatActual(specs[0].Name).Equals("inPhase")
	assert.For(t, "spec 1 name").ThatActual(specs[1].Name).Equals("proposerIsCurrentPlayer")
	assert.For(t, "spec 1 matches builder").ThatActual(specs[1]).Equals(legal.ProposerIsCurrentPlayer())
}

//boardgame:codegen
type moveDeclaredPreconditions struct {
	Default
}

func (m *moveDeclaredPreconditions) Apply(state boardgame.State) error {
	return nil
}

//boardgame:codegen
type moveNoDeclaredPreconditions struct {
	Default
}

func (m *moveNoDeclaredPreconditions) Apply(state boardgame.State) error {
	return nil
}

func declaredPreconditionsMoveInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {
	auto := NewAutoConfigurer(manager.Delegate())
	return Add(
		auto.MustConfig(
			new(moveDeclaredPreconditions),
			WithMoveName("Declared Preconditions"),
			// Footgun-batch F2 note: WithoutPrecondition names must match
			// CONTRIBUTED spec names at boot, so this fixture contributes
			// the two checks it then suppresses (the original fixture
			// suppressed "proposerIsCurrentPlayer"/"stackConstraints"
			// without contributing either — a silent no-op then, a boot
			// error now).
			WithLegalPhases(phaseSetUp),
			WithSourceProperty("DrawStack"),
			WithDestinationProperty("DiscardStack"),
			WithPreconditions(
				legal.PropAtLeast("game.Counter", 1),
			),
			// Task 8 note: this second spec must reference a property that
			// actually resolves, because as of Task 8 NewGameManager
			// assembles and boot-validates the plan for every opted-in move
			// (resolving paths against the example state). The original Task 7
			// round-trip fixture used legal.PlayerBool("SeatFilled"), which
			// worked only while declarations were inert; the shared moves-test
			// playerState (game_test.go) carries no bool property, so that path
			// now fails boot validation. player.Counter (an int on playerState)
			// resolves, and a distinct predicate type still exercises the
			// accumulate-across-WithPreconditions-calls behavior this test
			// documents.
			WithPreconditions(
				legal.PropCompare("player.Counter", ">=", 0),
			),
			WithoutPrecondition(PreconditionInPhase),
			WithoutPrecondition(PreconditionStackConstraints),
		),
		auto.MustConfig(
			new(moveNoDeclaredPreconditions),
			WithMoveName("No Declared Preconditions"),
		),
	)
}

// TestWithPreconditionsRoundTrip covers WithPreconditions/WithoutPrecondition
// round-tripping through auto.Config and back out via
// Default.DeclaredPreconditions: specs from MULTIPLE WithPreconditions calls
// accumulate in declaration order (mirroring WithLegalPhases' own
// accumulate-across-calls behavior, moves/with.go:97-107), suppression names
// from multiple WithoutPrecondition calls likewise accumulate, and a move
// that never called WithPreconditions at all reports nil (not an empty
// non-nil slice) — the design spec §2's "declaring is implementing": nil
// specifically means "not opted in".
func TestWithPreconditionsRoundTrip(t *testing.T) {
	manager, err := newGameManager(declaredPreconditionsMoveInstaller)
	assert.For(t, "manager").ThatActual(err).IsNil()
	game, err := manager.NewDefaultGame()
	assert.For(t, "game").ThatActual(err).IsNil()

	declared := game.MoveByName("Declared Preconditions").(*moveDeclaredPreconditions)
	specs, suppressions := declared.DeclaredPreconditions()

	assert.For(t, "declared spec count").ThatActual(len(specs)).Equals(2)
	assert.For(t, "declared spec 0").ThatActual(specs[0].Name).Equals("propAtLeast")
	assert.For(t, "declared spec 1").ThatActual(specs[1].Name).Equals("propCompare")

	assert.For(t, "suppression count").ThatActual(len(suppressions)).Equals(2)
	assert.For(t, "suppression 0").ThatActual(suppressions[0]).Equals("inPhase")
	assert.For(t, "suppression 1").ThatActual(suppressions[1]).Equals("stackConstraints")

	notDeclared := game.MoveByName("No Declared Preconditions").(*moveNoDeclaredPreconditions)
	nilSpecs, nilSuppressions := notDeclared.DeclaredPreconditions()
	assert.For(t, "nil specs when not opted in").ThatActual(nilSpecs == nil).IsTrue()
	assert.For(t, "nil suppressions when not configured").ThatActual(nilSuppressions == nil).IsTrue()
}
