package moves

import (
	"errors"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/storage/memory"
)

/*
The authorization tests for seating, which a mutation pass found were missing
entirely.

`moves/seat_player.go` scored 32% -- the worst file measured -- and its three
most alarming survivors were all authorization:

  - `CloseAllSeats.Legal`'s `checkRequireAdmin` call site could be turned into
    `return nil` and nothing failed. `checkRequireAdmin` itself IS unit tested
    (TestCheckRequireAdmin, in gathering_test.go), which is what made the gap
    invisible: three assertions on the helper read as coverage for a guard no
    test ever routed a proposer through. In fact `CloseAllSeats` appeared in no
    test in the repo at all, so its whole Legal was never evaluated -- the
    `ReadyToStart` check could be inverted and the "enough players are seated"
    comparison loosened, both unnoticed.

  - `SeatPlayer.Legal`'s "only an admin may propose this" guard could be
    deleted. `SeatPlayer` is the move that binds a real user to a seat, so that
    is a security property. The existing tests execute the line on every run;
    they simply never vary the proposer.

  - The "is there already a game admin?" loop in `SeatPlayer.Apply` could be
    skipped, which makes EVERY newly seated player game admin and silently
    displaces the first. A plausible wrong answer rather than a crash.

Every test below states its proposer or its ordering explicitly, because that
is the variable none of the existing tests moved.
*/

// closeAllSeatsDelegate configures CloseAllSeats the way a game that wants an
// explicit start does: only the game administrator may propose it.
//
// WithTargetCount(0) is not incidental. CloseAllSeats.Legal checks the seated
// count BEFORE the admin check, and the default target is MinNumPlayers, so
// with nobody seated every proposer would be rejected for the wrong reason and
// the admin guard would never be reached -- which is the shape of the bug
// these tests exist to catch, one layer up.
type closeAllSeatsDelegate struct {
	gatheringDelegate
}

var errTestNotReady = errors.New("the test delegate says the game is not ready")

func (c *closeAllSeatsDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := NewAutoConfigurer(c)
	return Combine(
		Add(
			auto.MustConfig(new(SeatPlayer)),
		),
		AddForPhase(gatheringPhaseGathering,
			auto.MustConfig(new(CloseAllSeats),
				WithRequireAdmin(),
				WithTargetCount(0),
			),
		),
	)
}

func newCloseAllSeatsGame(t *testing.T) (*boardgame.Game, *closeAllSeatsDelegate, boardgame.Move) {
	t.Helper()
	delegate := &closeAllSeatsDelegate{}
	manager, err := boardgame.NewGameManager(delegate, memory.NewStorageManager())
	if err != nil {
		t.Fatal("Couldn't create manager:", err)
	}
	game, err := manager.NewGame(2, nil, nil)
	if err != nil {
		t.Fatal("Couldn't create game:", err)
	}
	move := manager.ExampleMoveByName("Close All Seats")
	if move == nil {
		t.Fatal("Close All Seats was not configured")
	}
	return game, delegate, move
}

func TestCloseAllSeatsRequiresAdminOfItsProposer(t *testing.T) {

	game, _, move := newCloseAllSeatsGame(t)
	state := game.CurrentState()

	//Player 0 is a real player and is not the game administrator -- nobody has
	//been seated, so nobody is. WithRequireAdmin means they may not close the
	//seats.
	err := move.Legal(state, boardgame.PlayerIndex(0))
	if err == nil {
		t.Fatal("a non-admin player was allowed to close all seats")
	}
	if !strings.Contains(err.Error(), "administrator") {
		t.Errorf("a non-admin was rejected, but for the wrong reason: %v", err)
	}

	//AdminPlayerIndex is the engine itself, which always passes.
	if err := move.Legal(state, boardgame.AdminPlayerIndex); err != nil {
		t.Errorf("the engine was not allowed to close all seats: %v", err)
	}

	//ObserverPlayerIndex is not a player at all. checkRequireAdmin's own test
	//only ever passes player 0 and AdminPlayerIndex, so the `proposer < 0`
	//branch it takes is reached from here and nowhere else.
	if err := move.Legal(state, boardgame.ObserverPlayerIndex); err == nil {
		t.Error("an observer was allowed to close all seats")
	}
}

func TestCloseAllSeatsRefusesWhenTheGameIsNotReady(t *testing.T) {

	game, delegate, move := newCloseAllSeatsGame(t)

	//The admin passes today...
	if err := move.Legal(game.CurrentState(), boardgame.AdminPlayerIndex); err != nil {
		t.Fatalf("the engine could not close the seats to begin with: %v", err)
	}

	//...and stops passing the moment the delegate says the game is not ready.
	//This is the deadlock guard: seats must not close before the configuration
	//they depend on validates.
	delegate.readyToStartErr = errTestNotReady

	err := move.Legal(game.CurrentState(), boardgame.AdminPlayerIndex)
	if err == nil {
		t.Fatal("the seats closed on a game the delegate said was not ready")
	}
	if !strings.Contains(err.Error(), "not ready to start") {
		t.Errorf("wrong rejection: %v", err)
	}
}

func TestCloseAllSeatsCountsSeatedPlayersAgainstItsTarget(t *testing.T) {

	//Zero seated players against a target of zero is legal: "at least" means
	//at least. A move that reads this as "more than" blocks the only case a
	//manual-start game begins from.
	_, _, move := newCloseAllSeatsGame(t)

	delegate := &closeAllSeatsTargetDelegate{}
	manager, err := boardgame.NewGameManager(delegate, memory.NewStorageManager())
	if err != nil {
		t.Fatal("Couldn't create manager:", err)
	}
	game, err := manager.NewGame(2, nil, nil)
	if err != nil {
		t.Fatal("Couldn't create game:", err)
	}
	strict := manager.ExampleMoveByName("Close All Seats")
	if strict == nil {
		t.Fatal("Close All Seats was not configured")
	}

	//A target of one, with nobody seated, must be refused.
	err = strict.Legal(game.CurrentState(), boardgame.AdminPlayerIndex)
	if err == nil {
		t.Fatal("the seats closed with fewer players seated than the target")
	}
	if !strings.Contains(err.Error(), "are seated") {
		t.Errorf("wrong rejection: %v", err)
	}

	//And the move really does say it starts a gathering, which is what puts a
	//"Start Game" button in front of a player rather than a generic move form.
	concrete, ok := move.(*CloseAllSeats)
	if !ok {
		t.Fatalf("Close All Seats was a %T", move)
	}
	if !concrete.IsGatheringStartMove() {
		t.Error("CloseAllSeats does not report itself as a gathering start move")
	}
}

// closeAllSeatsTargetDelegate is closeAllSeatsDelegate with a target of one
// seated player rather than zero, so the count check can be seen to bind in
// both directions.
type closeAllSeatsTargetDelegate struct {
	gatheringDelegate
}

func (c *closeAllSeatsTargetDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := NewAutoConfigurer(c)
	return Combine(
		Add(
			auto.MustConfig(new(SeatPlayer)),
		),
		AddForPhase(gatheringPhaseGathering,
			auto.MustConfig(new(CloseAllSeats),
				WithTargetCount(1),
			),
		),
	)
}

func TestSeatPlayerMayOnlyBeProposedByAnAdmin(t *testing.T) {

	//SeatPlayer binds a real user to a seat, and the server is the only thing
	//that may ask for that. The rendezvous has to be present or Legal stops at
	//"No player to seat" long before it looks at the proposer -- which is why
	//no existing test reaches this guard.
	storage := &dropInStorage{StorageManager: memory.NewStorageManager()}
	manager, err := boardgame.NewGameManager(&dropInDelegate{}, storage)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	//One seat only: the game then sits in setup with three seats still open
	//and unclosed, which is the state SeatPlayer.Legal is asked about.
	storage.setSeat(&dropInSeat{storage: storage, remaining: 1})

	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("new game: %v", err)
	}

	//The rendezvous withdrew itself once that one seating committed, so put a
	//fresh one back: Legal reads it live, and without it the move stops at
	//"No player to seat" before it ever looks at who proposed.
	storage.setSeat(&dropInSeat{storage: storage, remaining: 1})

	move := manager.ExampleMoveByName("Seat Player")
	if move == nil {
		t.Fatal("Seat Player was not configured")
	}
	seat, ok := move.(*SeatPlayer)
	if !ok {
		t.Fatalf("Seat Player was a %T", move)
	}

	state := game.CurrentState()
	//Whichever seat is still open; there are four and the rendezvous above
	//filled at most four, so find one rather than assuming an index.
	seat.TargetPlayerIndex = boardgame.ObserverPlayerIndex
	for i, p := range state.ImmutablePlayerStates() {
		player := p.(*dropInPlayerState)
		if !player.SeatFilled && !player.SeatClosed {
			seat.TargetPlayerIndex = boardgame.PlayerIndex(i)
			break
		}
	}
	if seat.TargetPlayerIndex < 0 {
		t.Skip("every seat was filled during setup, so there is no open seat to aim at")
	}

	if err := move.Legal(state, boardgame.PlayerIndex(0)); err == nil {
		t.Fatal("a player was allowed to propose seating someone into a seat")
	} else if !strings.Contains(err.Error(), "admin") {
		t.Errorf("player 0 was rejected, but for the wrong reason: %v", err)
	}

	if err := move.Legal(state, boardgame.AdminPlayerIndex); err != nil {
		t.Errorf("the server was not allowed to seat a player: %v", err)
	}
}

func TestGameAdminGoesToTheFirstSeatedPlayerOnly(t *testing.T) {

	//`gatheringPlayerState` embeds behaviors.GameAdministrator, so seating
	//through it exercises the auto-assignment in SeatPlayer.Apply. The loop
	//that asks "is there already an admin?" can be skipped without any
	//existing test noticing, and skipping it makes every newly seated player
	//admin in turn -- so the assertion has to be that the SECOND player is
	//not, which no test asserted because nothing seated two players into a
	//state that has the behavior.
	storage := &dropInStorage{StorageManager: memory.NewStorageManager()}
	manager, err := boardgame.NewGameManager(&gatheringDelegate{}, storage)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	storage.setSeat(&dropInSeat{storage: storage, remaining: 2})

	game, err := manager.NewGame(4, nil, nil)
	if err != nil {
		t.Fatalf("new game: %v", err)
	}

	states := game.CurrentState().ImmutablePlayerStates()
	seated := 0
	admins := 0
	firstSeated := -1
	for i, p := range states {
		player := p.(*gatheringPlayerState)
		if player.SeatFilled {
			if firstSeated < 0 {
				firstSeated = i
			}
			seated++
		}
		if behaviors.PlayerIsAdmin(p) {
			admins++
		}
	}

	if seated < 2 {
		t.Fatalf("only %v players were seated; this test needs two to say anything", seated)
	}
	if admins != 1 {
		t.Fatalf("%v of %v seated players are game admin; exactly one may be", admins, seated)
	}
	if !behaviors.PlayerIsAdmin(states[firstSeated]) {
		t.Errorf("player %v was seated first but is not the game admin", firstSeated)
	}
}

// baseDelegationDelegate configures ActivateEmptySeat so that its EMBEDDED
// base's check is the one that fails: the move is legal only in the playing
// phase, and every game below sits in the gathering phase.
type baseDelegationDelegate struct {
	gatheringDelegate
}

func (b *baseDelegationDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := NewAutoConfigurer(b)
	return Add(
		auto.MustConfig(new(SeatPlayer)),
		auto.MustConfig(new(ActivateEmptySeat),
			WithLegalPhases(gatheringPhasePlaying),
		),
	)
}

// TestASubclassSurfacesItsBaseClassError covers the framework's central
// composition mechanism, which a mutation pass found systematically
// unasserted: `moves/seat_player.go:540` (`a.FixUpMulti.Legal`), `:779`
// (`c.Default.ValidConfiguration`), and `moves/round_robin.go:254,316` are all
// a subclass forwarding to its embedded base and propagating the error, and
// every one of them could be disabled without a test noticing.
//
// The shape of the test is the generalizable part: configure the move so the
// BASE's check is what fails, then assert the subclass hands that failure back
// rather than reaching its own logic. Here the base is `Default.Legal`'s phase
// gate, reached through FixUpMulti, and the subclass's own check (a target
// seat that is unfilled and inactive) would pass.
func TestASubclassSurfacesItsBaseClassError(t *testing.T) {

	manager, err := boardgame.NewGameManager(&baseDelegationDelegate{}, memory.NewStorageManager())
	if err != nil {
		t.Fatal("Couldn't create manager:", err)
	}
	game, err := manager.NewGame(2, nil, nil)
	if err != nil {
		t.Fatal("Couldn't create game:", err)
	}

	move := manager.ExampleMoveByName("Activate Empty Seat")
	if move == nil {
		t.Fatal("Activate Empty Seat was not configured")
	}
	activate, ok := move.(*ActivateEmptySeat)
	if !ok {
		t.Fatalf("Activate Empty Seat was a %T", move)
	}

	state := game.CurrentState()

	//Its own check would pass on this target: seat 0 is unfilled, and every
	//player of a fresh game is inactive. So anything the move returns comes
	//from the base.
	activate.DefaultsForState(state)

	err = move.Legal(state, boardgame.AdminPlayerIndex)
	if err == nil {
		t.Fatal("a move legal only in the playing phase was legal in the gathering phase")
	}
	if !strings.Contains(err.Error(), "phase") {
		t.Errorf("the move was refused, but not by its base's phase check: %v", err)
	}
}
