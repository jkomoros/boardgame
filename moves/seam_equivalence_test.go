package moves

import (
	"sort"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
	"github.com/jkomoros/boardgame/legal/legaltest"
	"github.com/jkomoros/boardgame/storage/memory"
)

/*
Differential golden tests for the widened seam (design spec
2026-11-22-legality-seam-widening): every framework base type, opted in with
zero authored preconditions, must reach the same verdict, first-failure text
and post-call move values through its plan as through its frozen Legal() --
compared on the SAME state by legaltest (frozen under
LegalWithPlansSuspended, plan normally), for every proposer and
TargetPlayerIndex variant, at every state a short walk reaches.

The fixtures are the package's existing scenario installers and delegates,
so each base type is exercised in the configurations its own tests already
established. TestSeamEquivalenceCoversEveryFrameworkCheck then requires that,
across all of them, every framework check atom was actually part of some
compared plan.
*/

func optInInstaller(installer func(*boardgame.GameManager) []boardgame.MoveConfig) func(*boardgame.GameManager) []boardgame.MoveConfig {
	return func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		return legaltest.OptInAll(installer(manager), WithLegalPreconditions())
	}
}

type optedInGathering struct{ gatheringDelegate }

func (d *optedInGathering) ConfigureMoves() []boardgame.MoveConfig {
	return legaltest.OptInAll(d.gatheringDelegate.ConfigureMoves(), WithLegalPreconditions())
}

type optedInCloseAllSeats struct{ closeAllSeatsDelegate }

func (d *optedInCloseAllSeats) ConfigureMoves() []boardgame.MoveConfig {
	return legaltest.OptInAll(d.closeAllSeatsDelegate.ConfigureMoves(), WithLegalPreconditions())
}

type optedInCloseAllSeatsTarget struct{ closeAllSeatsTargetDelegate }

func (d *optedInCloseAllSeatsTarget) ConfigureMoves() []boardgame.MoveConfig {
	return legaltest.OptInAll(d.closeAllSeatsTargetDelegate.ConfigureMoves(), WithLegalPreconditions())
}

type optedInDropIn struct{ dropInDelegate }

func (d *optedInDropIn) ConfigureMoves() []boardgame.MoveConfig {
	return legaltest.OptInAll(d.dropInDelegate.ConfigureMoves(), WithLegalPreconditions())
}

type optedInReopen struct{ reopenDelegate }

func (d *optedInReopen) ConfigureMoves() []boardgame.MoveConfig {
	return legaltest.OptInAll(d.reopenDelegate.ConfigureMoves(), WithLegalPreconditions())
}

type optedInCloseSeats struct{ closeSeatsDelegate }

func (d *optedInCloseSeats) ConfigureMoves() []boardgame.MoveConfig {
	return legaltest.OptInAll(d.closeSeatsDelegate.ConfigureMoves(), WithLegalPreconditions())
}

type optedInBaseDelegation struct{ baseDelegationDelegate }

func (d *optedInBaseDelegation) ConfigureMoves() []boardgame.MoveConfig {
	return legaltest.OptInAll(d.baseDelegationDelegate.ConfigureMoves(), WithLegalPreconditions())
}

type seamScenario struct {
	name     string
	delegate boardgame.GameDelegate
}

func seamScenarios() []seamScenario {
	installers := map[string]func(*boardgame.GameManager) []boardgame.MoveConfig{
		"default":                     defaultMoveInstaller,
		"counted transfer":            countedTransferMoveInstaller(2),
		"counted fix-up transfer":     countedFixUpTransferMoveInstaller(2),
		"until reached":               thresholdTransferMoveInstaller(new(MoveComponentsUntilCountReached), 3, "Until Reached"),
		"until left":                  thresholdTransferMoveInstaller(new(MoveComponentsUntilCountLeft), 50, "Until Left"),
		"round robin preflight":       roundRobinTransferPreflightInstaller,
		"zero iteration round robin":  zeroIterationRoundRobinMoveInstaller,
		"hop along path":              hopMoveInstaller,
		"advance token":               advanceTokenMoveInstaller,
		"player scoped transfer":      playerScopedTransferInstaller("game.DrawStack", "player.Hand", 2),
		"force finish turn":           forceFinishTurnSeamInstaller(false),
		"force finish turn off-phase": forceFinishTurnSeamInstaller(true),
		"two failures":                twoFailuresInstaller,
	}
	var scenarios []seamScenario
	for name, installer := range installers {
		scenarios = append(scenarios, seamScenario{name, &gameDelegate{moveInstaller: optInInstaller(installer)}})
	}
	scenarios = append(scenarios,
		seamScenario{"gathering", &optedInGathering{}},
		seamScenario{"close all seats", &optedInCloseAllSeats{}},
		seamScenario{"close all seats target", &optedInCloseAllSeatsTarget{}},
		seamScenario{"drop in", &optedInDropIn{}},
		seamScenario{"reopen", &optedInReopen{}},
		seamScenario{"close seats", &optedInCloseSeats{}},
		seamScenario{"base delegation", &optedInBaseDelegation{}},
	)
	sort.Slice(scenarios, func(i, j int) bool { return scenarios[i].name < scenarios[j].name })
	return scenarios
}

// forceFinishTurnSeamInstaller installs FinishTurn and ForceFinishTurn. With
// offPhase, ForceFinishTurn is also restricted to a phase the game is not in:
// its frozen chain inherits nothing and so ignores that, and its plan must
// too.
func forceFinishTurnSeamInstaller(offPhase bool) func(*boardgame.GameManager) []boardgame.MoveConfig {
	return func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		options := []CustomConfigurationOption{WithMoveName("Force Finish Turn"), WithIsFixUp(false)}
		if offPhase {
			options = append(options, WithLegalPhases(phaseNormalPlayDrawCard))
		}
		return Add(
			auto.MustConfig(new(FinishTurn)),
			auto.MustConfig(new(ForceFinishTurn), options...),
		)
	}
}

// twoFailuresInstaller restricts several framework moves to a phase the game
// does not start in AND makes their own check fail too (an empty source
// stack, a turn that is never done, a draw stack that needs no reshuffle).
// Both the contributed phase atom and the type's own check fail, so only the
// ORDER of the contribution decides which message wins -- the states a walk
// of a well-formed game rarely reaches, and the ones that catch a
// contribution assembled in the wrong order.
func twoFailuresInstaller(manager *boardgame.GameManager) []boardgame.MoveConfig {
	auto := NewAutoConfigurer(manager.Delegate())
	return AddForPhase(phaseNormalPlay,
		auto.MustConfig(
			new(MoveCountComponents),
			WithMoveName("Move From Empty Discard"),
			WithSourceProperty("DiscardStack"),
			WithDestinationProperty("DrawStack"),
			WithTargetCount(2),
			WithIsFixUp(false),
		),
		auto.MustConfig(
			new(DealComponentsUntilPlayerCountReached),
			WithMoveName("Deal From Empty Discard"),
			WithGameProperty("DiscardStack"),
			WithPlayerProperty("Hand"),
			WithTargetCount(1),
		),
		auto.MustConfig(new(FinishTurn)),
		auto.MustConfig(new(ShuffleDiscardIntoDraw)),
	)
}

// seamCoverage records every framework check atom a compared plan contained.
func seamPlanAtoms(manager *boardgame.GameManager, game *boardgame.Game) []string {
	var atoms []string
	for _, move := range game.Moves() {
		if !manager.LegalPlanAssembled(move.Info().Name()) {
			continue
		}
		provider, ok := move.(PreconditionsProvider)
		if !ok {
			continue
		}
		for _, spec := range provider.ContributedPreconditions() {
			atoms = append(atoms, spec.Name)
		}
	}
	return atoms
}

func TestSeamEquivalence(t *testing.T) {
	for _, scenario := range seamScenarios() {
		t.Run(scenario.name, func(t *testing.T) {
			manager, err := boardgame.NewGameManager(scenario.delegate, memory.NewStorageManager())
			if err != nil {
				t.Fatalf("every move opted in must still boot: %v", err)
			}
			game, err := manager.NewDefaultGame()
			if err != nil {
				t.Fatalf("new game: %v", err)
			}
			if compared := legaltest.Walk(t, game, 12); compared == 0 {
				t.Fatal("no opted-in move was compared")
			}
		})
	}
}

func TestSeamEquivalenceCoversEveryFrameworkCheck(t *testing.T) {
	covered := make(map[string]bool)
	for _, scenario := range seamScenarios() {
		manager, err := boardgame.NewGameManager(scenario.delegate, memory.NewStorageManager())
		if err != nil {
			t.Fatalf("%v: %v", scenario.name, err)
		}
		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("%v: %v", scenario.name, err)
		}
		for _, atom := range seamPlanAtoms(manager, game) {
			covered[atom] = true
		}
	}

	// Checks whose base types have no fixture in this package's tests are
	// covered by the example games' own differential tests (the
	// player-facing AnyPlayer family, the behavior-discovered stack moves,
	// graph movement, component-dispatched legality). Every name here must
	// stay honest: if it becomes covered, the test asks for its removal.
	coveredElsewhere := map[string]string{
		string(PreconditionAdminPlayer):               "no in-repo fixture; same shape as anyPlayer, which is covered",
		string(PreconditionSelectColor):               "no in-repo fixture; same shape as anyPlayer, which is covered",
		string(PreconditionAllPlayersSubmitted):       "no in-repo fixture",
		string(PreconditionResetAllPlayerSubmissions): "no in-repo fixture",
		string(PreconditionMoveOnGraph):               "games/murdermrmonroe",
		string(PreconditionDefaultComponent):          "no in-repo fixture",
		string(PreconditionReplenishMarket):           "no in-repo fixture",
	}

	for _, ctor := range frameworkCheckConstructors() {
		_, elsewhere := coveredElsewhere[ctor.Name]
		switch {
		case covered[ctor.Name] && elsewhere:
			t.Errorf("framework check %q is now covered here; remove it from coveredElsewhere", ctor.Name)
		case !covered[ctor.Name] && !elsewhere:
			t.Errorf("framework check %q was not part of any differential scenario's plan", ctor.Name)
		}
	}
}

// TestSeamBypassedLevelIsBootError: a creator override that super-calls a
// NON-immediate ancestor (moveBypassesMoveCountLegal calls
// ApplyCountTimes.Legal, skipping MoveCountComponents.Legal) has a frozen
// chain that never runs MoveCountComponents' own check. Opting it in would
// silently add that check, so boot fails naming the skipped atom.
func TestSeamBypassedLevelIsBootError(t *testing.T) {
	_, err := newGameManager(optInInstaller(bypassedCountLegalMoveInstaller))
	if err == nil {
		t.Fatal("expected a boot error for an override that skips a framework Legal level")
	}
	if !strings.Contains(err.Error(), string(PreconditionMoveCountComponents)) {
		t.Errorf("the boot error should name the skipped check: %v", err)
	}
}

func finishTurnInstaller(options ...CustomConfigurationOption) func(*boardgame.GameManager) []boardgame.MoveConfig {
	return func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		return Add(auto.MustConfig(new(FinishTurn), options...))
	}
}

// TestFrameworkCheckCanBeMovedNotRemoved pins "move, don't remove".
func TestFrameworkCheckCanBeMovedNotRemoved(t *testing.T) {
	t.Run("suppressing without re-authoring is a boot error", func(t *testing.T) {
		_, err := newGameManager(finishTurnInstaller(WithoutLegalPrecondition(PreconditionFinishTurn)))
		if err == nil || !strings.Contains(err.Error(), "never re-authors it") {
			t.Fatalf("expected a move-don't-remove boot error, got %v", err)
		}
	})
	t.Run("authoring a check the move does not contribute is a boot error", func(t *testing.T) {
		_, err := newGameManager(finishTurnInstaller(WithLegalPreconditions(FrameworkPrecondition(PreconditionApplyUntil))))
		if err == nil || !strings.Contains(err.Error(), string(PreconditionApplyUntil)) {
			t.Fatalf("expected a boot error naming applyUntil, got %v", err)
		}
	})
	t.Run("suppress then re-author moves it after an authored gate", func(t *testing.T) {
		manager, err := newGameManager(finishTurnInstaller(
			WithoutLegalPrecondition(PreconditionFinishTurn),
			WithLegalPreconditions(alwaysFailPrecondition(), FrameworkPrecondition(PreconditionFinishTurn)),
		))
		if err != nil {
			t.Fatalf("boot: %v", err)
		}
		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("new game: %v", err)
		}
		move := game.MoveByName("Finish Turn")
		_, entries, ok := manager.LegalEvaluateLedger(move.Info().Name(), game.CurrentState(), move, boardgame.AdminPlayerIndex)
		if !ok {
			t.Fatal("no plan")
		}
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name)
		}
		if strings.Join(names, ",") != "propAtLeast,finishTurn" {
			t.Fatalf("plan order = %v, want the authored gate then finishTurn", names)
		}
		// The gate failed first, so the ledger does not run the residue.
		last := entries[len(entries)-1]
		if last.Verdict.Outcome != legal.Unknown || !strings.Contains(last.Verdict.Reason, "not evaluated") {
			t.Errorf("a framework check after a failure should be reported not evaluated, got %+v", last.Verdict)
		}
		if last.Serializable || last.ClientEvaluable || !last.FieldDependent {
			t.Errorf("a framework check is opaque and never memoized, got %+v", last)
		}
	})
}

// TestMoveCountComponentsContributionSkipsStackConstraints pins that the
// contribution mirrors the call path, not the embedding: MoveCountComponents'
// Legal calls legalWithoutStackConstraints, so its plan has no
// stackConstraints atom even though Default would derive one.
func TestMoveCountComponentsContributionSkipsStackConstraints(t *testing.T) {
	manager, err := newGameManager(countedTransferMoveInstaller(2))
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	move := game.MoveByName("Move Counted Components").(*MoveCountComponents)
	var names []string
	for _, spec := range move.ContributedPreconditions() {
		names = append(names, spec.Name)
	}
	if strings.Contains(strings.Join(names, ","), string(PreconditionStackConstraints)) {
		t.Fatalf("contribution %v includes stackConstraints", names)
	}
	if got := strings.Join(names[len(names)-2:], ","); got != "applyUntil,moveCountComponents" {
		t.Fatalf("contribution tail = %v, want applyUntil,moveCountComponents", got)
	}
}
