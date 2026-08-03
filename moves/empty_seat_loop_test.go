package moves

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/storage/memory"
)

/*
The boot check for the empty-seat activation loop
(validateEmptySeatActivationLoop, behavior_pairings.go).

[ActivateInactivePlayer]'s doc comment has said not to make it always legal in a
game that also inactivates its empty seats since the move was written, and
validateBehaviorPairings' seating-deadlock remediation says it again in its last
sentence. Prose about a required invariant is not enforcement.

Every fixture below was run with the check disabled before the check was
written, and the comment on each records what the game ACTUALLY did. That
mattered: the first draft of this check asserted the failure was always
ErrTooManyFixUps, and the sabotage run showed it is not -- one of the three
configurations boots cleanly and breaks later, in the quieter way the prose
describes. A boot error whose stated reason has never been watched happen is
prose with extra steps.
*/

/*
loopingActivatorDelegate is dropInDelegate plus an unrestricted
ActivateEmptySeat -- the shape a creator reaches for when they want empty seats
reopened "as soon as possible."

Observed with the check disabled: the game BOOTS, because the unrestricted
activator sits after StartPhase in the move list, so the fix-up pass leaves the
setup phase before it can undo InactivateEmptySeat. It then reopens both empty
seats in normal play, and turn order walks straight into them:

	turn 0: current player 0 (seat filled = true)
	turn 1: current player 1 (seat filled = true)
	turn 2: current player 2 (seat filled = false)
	turn 3: current player 3 (seat filled = false)

Seats 2 and 3 have no user behind them, so nobody can propose their move and
play stops there forever. That is verbatim the harm ActivateInactivePlayer's doc
comment names: the round waits on players who do not exist.
*/
type loopingActivatorDelegate struct {
	dropInDelegate
}

func (g *loopingActivatorDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := NewAutoConfigurer(g)
	return Combine(
		g.dropInDelegate.ConfigureMoves(),
		Add(
			auto.MustConfig(new(ActivateEmptySeat)),
		),
	)
}

/*
loopingActivatorFirstDelegate is the SAME game with one difference: the
unrestricted ActivateEmptySeat is listed before StartPhase instead of after.

Observed with the check disabled: NewDefaultGame fails outright with "we
recursed deeply in fixup, which implies that ProposeFixUp has a move that is
always legal" (boardgame.ErrTooManyFixUps). The activator is now reachable while
the game is still in the setup phase, so it and InactivateEmptySeat ping-pong
until maxRecurseCount.

This fixture exists to pin the reason the check refuses BOTH orders: which of
the two failures a creator gets is decided by nothing but the order of their
ConfigureMoves entries. A silent coin flip between "unplayable now" and
"unplayable four turns from now" is exactly what a boot error is for.
*/
type loopingActivatorFirstDelegate struct {
	dropInDelegate
}

func (g *loopingActivatorFirstDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := NewAutoConfigurer(g)
	return Combine(
		AddForPhase(dropInPhaseSetUp,
			auto.MustConfig(new(SeatPlayer)),
		),
		Add(
			auto.MustConfig(new(ActivateEmptySeat)),
		),
		AddForPhase(dropInPhaseNormal,
			auto.MustConfig(new(ActivateFilledSeat)),
			auto.MustConfig(new(moveDropInAct),
				WithMoveName("Drop In Act"),
			),
			auto.MustConfig(new(FinishTurn)),
		),
		AddOrderedForPhase(dropInPhaseSetUp,
			DefaultRoundSetup(auto),
			auto.MustConfig(new(StartPhase),
				WithPhaseToStart(dropInPhaseNormal, dropInPhaseEnum),
			),
		),
	)
}

/*
loopingInactivatorDelegate is the mirror image: reopenDelegate's phase-scoped
ActivateEmptySeat (a configuration that works, and is played end-to-end in
empty_seat_moves_test.go) plus an unrestricted SECOND InactivateEmptySeat.

Observed with the check disabled: NewDefaultGame fails with ErrTooManyFixUps.
The loop is the same loop, so the check has to catch this direction too --
"unrestricted" is a property of the pair, not of one particular move.
*/
type loopingInactivatorDelegate struct {
	reopenDelegate
}

func (g *loopingInactivatorDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := NewAutoConfigurer(g)
	return Combine(
		g.reopenDelegate.ConfigureMoves(),
		Add(
			auto.MustConfig(new(InactivateEmptySeat),
				WithMoveNameSuffix("Anytime"),
			),
		),
	)
}

func TestEmptySeatActivationLoopIsABootError(t *testing.T) {

	tests := map[string]struct {
		delegate boardgame.GameDelegate
		//The move name the error must identify as the unrestricted one, since
		//an error that names the wrong move sends the reader to the wrong file.
		culprit string
	}{
		"unrestricted activator, listed late": {
			delegate: &loopingActivatorDelegate{},
			culprit:  "Activate Empty Seat",
		},
		"unrestricted activator, listed early": {
			delegate: &loopingActivatorFirstDelegate{},
			culprit:  "Activate Empty Seat",
		},
		"unrestricted inactivator": {
			delegate: &loopingInactivatorDelegate{},
			culprit:  "Inactivate Empty Seat - Anytime",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			storage := &dropInStorage{StorageManager: memory.NewStorageManager()}

			_, err := boardgame.NewGameManager(test.delegate, storage)
			if err == nil {
				t.Fatal("NewGameManager accepted a move list whose fix-up moves undo each other")
			}

			message := err.Error()
			for _, want := range []string{test.culprit, "installed unrestricted", "ActivateFilledSeat"} {
				if !strings.Contains(message, want) {
					t.Errorf("the boot error does not mention %q, so it does not tell the reader %v: %v", want, whyItMatters(want), message)
				}
			}
		})
	}
}

// whyItMatters keeps the assertion above honest about what each required
// substring is FOR, so a future edit that drops one has to argue with a reason
// rather than with a string literal.
func whyItMatters(want string) string {
	switch want {
	case "installed unrestricted":
		return "which property of the configuration is the problem"
	case "ActivateFilledSeat":
		return "the move to use instead for always-legal activation"
	default:
		return "which move to go change"
	}
}

// TestEmptySeatActivationCheckAcceptsRealConfigurations is the other half, and
// the more important one: this check would be worse than the prose it replaced
// if it fired on a game that works. Both fixtures below are played end-to-end
// elsewhere in this package (drop_in_joining_test.go and
// empty_seat_moves_test.go), so they are known-good, not merely
// plausible-looking.
func TestEmptySeatActivationCheckAcceptsRealConfigurations(t *testing.T) {

	tests := map[string]boardgame.GameDelegate{
		//DefaultRoundSetup puts ActivateInactivePlayer and InactivateEmptySeat
		//in ONE ordered progression, plus an always-legal ActivateFilledSeat --
		//which touches no empty seat and so must stay invisible to the check.
		"DefaultRoundSetup plus always-legal ActivateFilledSeat": &dropInDelegate{},
		//The same game with a phase-scoped ActivateEmptySeat: a genuine
		//empty-seat activator, legal only in the phase after setup closed them.
		"phase-scoped ActivateEmptySeat": &reopenDelegate{},
	}

	for name, delegate := range tests {
		t.Run(name, func(t *testing.T) {
			storage := &dropInStorage{StorageManager: memory.NewStorageManager()}

			if _, err := boardgame.NewGameManager(delegate, storage); err != nil {
				t.Fatalf("the empty-seat loop check rejected a configuration that works: %v", err)
			}
		})
	}
}

/*
samePhaseLoopDelegate is dropInDelegate's own move list with ONE addition:
ActivateEmptySeat scoped to dropInPhaseSetUp, the same phase DefaultRoundSetup's
InactivateEmptySeat runs its progression in.

Neither move is installed unrestricted -- the activator carries a legal phase,
the inactivator carries a progression -- so the check's original "is either one
unrestricted" test was silent on this, and the check's own remedy text said to
"restrict them to phases where they cannot both apply". A creator who followed
that advice INTO the same phase got no diagnostic at all.

Observed with the check disabled, and with a seating rendezvous injected so
InactivateEmptySeat actually applies (storage.setSeat, two of four seats
filled): NewDefaultGame fails with "we recursed deeply in fixup, which implies
that ProposeFixUp has a move that is always legal" -- boardgame.ErrTooManyFixUps.
The control matters here: the identical fixture with the activator deleted boots
clean, so it is the activator doing it and not the trimmed move list.

The activator is legal at every point of the setup phase because it is in no
progression, which includes the point the progression fires the inactivator at.
Sharing a phase is exactly as fatal as sharing the whole game.
*/
type samePhaseLoopDelegate struct {
	dropInDelegate
}

func (g *samePhaseLoopDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := NewAutoConfigurer(g)
	return Combine(
		AddForPhase(dropInPhaseSetUp,
			auto.MustConfig(new(SeatPlayer)),
			auto.MustConfig(new(ActivateEmptySeat)),
		),
		AddForPhase(dropInPhaseNormal,
			auto.MustConfig(new(SeatPlayer),
				WithMoveNameSuffix("Mid Game"),
			),
			auto.MustConfig(new(ActivateFilledSeat)),
			auto.MustConfig(new(moveDropInAct),
				WithMoveName("Drop In Act"),
			),
			auto.MustConfig(new(FinishTurn)),
		),
		AddOrderedForPhase(dropInPhaseSetUp,
			DefaultRoundSetup(auto),
			auto.MustConfig(new(StartPhase),
				WithPhaseToStart(dropInPhaseNormal, dropInPhaseEnum),
			),
		),
	)
}

// TestSamePhaseEmptySeatActivationIsABootError covers the shape the check
// originally missed: both moves restricted, to the SAME phase. The question the
// check has to answer is not "is either move unrestricted" but "can these two be
// candidates at the same moment", and phase-scoping two fix-up moves together
// answers yes.
func TestSamePhaseEmptySeatActivationIsABootError(t *testing.T) {

	storage := &dropInStorage{StorageManager: memory.NewStorageManager()}

	_, err := boardgame.NewGameManager(&samePhaseLoopDelegate{}, storage)
	if err == nil {
		t.Fatal("NewGameManager accepted an activator and an inactivator scoped to the same phase, which recurses until NewDefaultGame fails with ErrTooManyFixUps")
	}

	message := err.Error()

	//The error must name the phase, or the reader has to guess which of their
	//AddForPhase blocks is the problem.
	if !strings.Contains(message, "Set Up") {
		t.Errorf("the boot error does not name the shared phase, so it does not say which AddForPhase block to go change: %v", message)
	}

	//It must name the move that is NOT in a progression, since that is the one
	//that is a candidate at every point of the phase and the one to move.
	if !strings.Contains(message, "\"Activate Empty Seat\" is not in a move progression") {
		t.Errorf("the boot error does not identify the move that is loose in the phase, so it does not say which of the two to change: %v", message)
	}

	//It must NOT claim either move is unrestricted -- both carry restrictions,
	//and sending the reader to look for a missing phase would waste their time.
	if strings.Contains(message, "installed unrestricted") {
		t.Errorf("the boot error blames being installed unrestricted, but both of these moves ARE restricted: %v", message)
	}

	//The remedy that caused this must not be the remedy offered for it.
	if !strings.Contains(message, "scoping both to the SAME phase does not help") {
		t.Errorf("the boot error does not warn against the remedy that produced this configuration: %v", message)
	}
}
