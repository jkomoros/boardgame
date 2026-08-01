package moves

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/storage/memory"
)

/*
The end-to-end regression net for [ActivateEmptySeat]: can a game that closed
its empty seats for a round actually re-open them for the next one?

This is the third instance of the PlayerIndex.EnsureValid retargeting class
(after SeatPlayer, fixed in "Let a player who joins mid-game actually take a
turn"). ActivateEmptySeat resolved its target through EnsureValid, which
advances past any player GameDelegate.PlayerMayBeActive rejects -- and an
inactive seat is exactly such a player. So the move's own Legal, which then
REQUIRES the target to be inactive, could never see one: EnsureValid had
already walked away from it onto an active player, and Legal failed with
"Player is already active." The move was legal only in a game where no player
at all may be active, which is to say never.

The fixture reuses drop_in_joining_test.go's game, which is the shape the
scaffolding stub generates, and changes exactly one thing: the normal-play
phase makes ActivateEmptySeat legal instead of ActivateFilledSeat. Setup
therefore closes the two seats nobody sat in (DefaultRoundSetup's
InactivateEmptySeat), and normal play must re-open them.

The load-bearing assertion is not the Inactive flag. It is that
PlayerIndex.Next reaches the re-opened seats: "inactive" means "turn order
skips you", so a seat whose flag flipped but which turn order still walks past
has not actually been re-opened.
*/

// reopenDelegate is dropInDelegate with one substitution in ConfigureMoves.
type reopenDelegate struct {
	dropInDelegate
}

func (g *reopenDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := NewAutoConfigurer(g)
	return Combine(
		AddForPhase(dropInPhaseSetUp,
			auto.MustConfig(new(SeatPlayer)),
		),
		AddOrderedForPhase(dropInPhaseSetUp,
			DefaultRoundSetup(auto),
			auto.MustConfig(new(StartPhase),
				WithPhaseToStart(dropInPhaseNormal, dropInPhaseEnum),
			),
		),
		AddForPhase(dropInPhaseNormal,
			//The move under test. It is a FixUpMulti, so the engine proposes
			//it itself once the phase begins; nothing below calls it directly.
			auto.MustConfig(new(ActivateEmptySeat)),
			auto.MustConfig(new(moveDropInAct),
				WithMoveName("Drop In Act"),
			),
			auto.MustConfig(new(FinishTurn)),
		),
	)
}

func TestActivateEmptySeatReopensClosedSeats(t *testing.T) {

	storage := &dropInStorage{StorageManager: memory.NewStorageManager()}

	manager, err := boardgame.NewGameManager(&reopenDelegate{}, storage)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	//Two of the four seats are filled before the game starts; the other two
	//are the empty seats this test is about.
	storage.setSeat(&dropInSeat{storage: storage, remaining: 2})

	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("new game: %v", err)
	}

	players := dropInPlayers(t, game)

	if !players[0].SeatFilled || !players[1].SeatFilled {
		t.Fatalf("the two pre-game seats were not filled: %v %v", players[0].SeatFilled, players[1].SeatFilled)
	}
	if players[2].SeatFilled || players[3].SeatFilled {
		t.Fatal("more seats were filled than the rendezvous offered")
	}
	if got := game.CurrentState().ImmutableGameState().(*dropInGameState).Phase.Value(); got != dropInPhaseNormal {
		t.Fatalf("the game did not reach normal play; phase = %v", got)
	}

	//This is the state ActivateEmptySeat exists to undo: DefaultRoundSetup's
	//InactivateEmptySeat closed both seats nobody is sitting in.
	//ActivateEmptySeat is legal in the phase the game just entered, and it is
	//a fix-up, so by the time the game settled it must already have run.
	if players[2].PlayerInactive || players[3].PlayerInactive {
		t.Fatalf("ActivateEmptySeat never re-opened the empty seats: seat 2 inactive=%v, seat 3 inactive=%v",
			players[2].PlayerInactive, players[3].PlayerInactive)
	}

	//And it must not have disturbed the seats real people are in.
	if players[0].PlayerInactive || players[1].PlayerInactive {
		t.Fatal("re-opening the empty seats deactivated a filled one")
	}
	if players[2].SeatFilled || players[3].SeatFilled {
		t.Fatal("re-opening a seat must not fill it")
	}

	//The load-bearing check: turn order actually reaches a re-opened seat.
	//"Inactive" means PlayerIndex.Next skips you, so a flag that flipped
	//without turn order following is not a re-opened seat.
	state := game.CurrentState()
	reached := map[boardgame.PlayerIndex]bool{}
	index := boardgame.PlayerIndex(0)
	for i := 0; i < len(state.ImmutablePlayerStates()); i++ {
		index = index.Next(state)
		reached[index] = true
	}
	for _, seat := range []boardgame.PlayerIndex{2, 3} {
		if !reached[seat] {
			t.Fatalf("turn order still skips re-opened seat %v, so it was never really activated", seat)
		}
	}
}

// TestActivateEmptySeatIsLegalOnAnInactiveEmptySeat is the unit-level
// statement of the same bug, without a whole game: DefaultsForState picks the
// empty inactive seat, and Legal must AGREE with the target its own defaulter
// just chose. Before the fix it disagreed with itself, because EnsureValid
// moved the target between the two.
func TestActivateEmptySeatIsLegalOnAnInactiveEmptySeat(t *testing.T) {

	storage := &dropInStorage{StorageManager: memory.NewStorageManager()}

	manager, err := boardgame.NewGameManager(&reopenDelegate{}, storage)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	storage.setSeat(&dropInSeat{storage: storage, remaining: 2})

	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("new game: %v", err)
	}

	//Put the two empty seats back to closed by hand, so this test states the
	//precondition itself rather than depending on when the fix-up ran.
	state := game.CurrentState().(boardgame.State)
	for _, index := range []int{2, 3} {
		state.PlayerStates()[index].(*dropInPlayerState).SetPlayerInactive()
	}

	move := new(ActivateEmptySeat)
	move.DefaultsForState(state)

	if move.TargetPlayerIndex != 2 {
		t.Fatalf("DefaultsForState chose %v, not the first empty inactive seat", move.TargetPlayerIndex)
	}

	//Legal's FixUpMulti/Default half needs real move info to check the phase,
	//so only the seat reasoning is exercised here, directly.
	if err := activateEmptySeatTarget.legalTarget(state, move.TargetPlayerIndex); err != nil {
		t.Fatalf("ActivateEmptySeat rejected the very seat its own DefaultsForState chose: %v", err)
	}

	//A filled seat is not this move's business, even when inactive.
	state.PlayerStates()[0].(*dropInPlayerState).SetPlayerInactive()
	move.TargetPlayerIndex = 0
	if err := activateEmptySeatTarget.legalTarget(state, move.TargetPlayerIndex); err == nil {
		t.Fatal("ActivateEmptySeat accepted a FILLED seat; that is ActivateFilledSeat's job")
	}
}

// closeSeatsDelegate is the same game again, with CloseEmptySeat in normal
// play: the audit of the EnsureValid class found the identical bug there.
type closeSeatsDelegate struct {
	dropInDelegate
}

func (g *closeSeatsDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := NewAutoConfigurer(g)
	return Combine(
		AddForPhase(dropInPhaseSetUp,
			auto.MustConfig(new(SeatPlayer)),
		),
		AddOrderedForPhase(dropInPhaseSetUp,
			DefaultRoundSetup(auto),
			auto.MustConfig(new(StartPhase),
				WithPhaseToStart(dropInPhaseNormal, dropInPhaseEnum),
			),
		),
		AddForPhase(dropInPhaseNormal,
			auto.MustConfig(new(CloseEmptySeat)),
			auto.MustConfig(new(moveDropInAct),
				WithMoveName("Drop In Act"),
			),
			auto.MustConfig(new(FinishTurn)),
		),
	)
}

// TestCloseEmptySeatClosesAnInactivatedSeat is the second confirmed instance of
// the same class. DefaultRoundSetup inactivates the seats nobody sat in, and
// "an empty seat that has been closed for the round" is exactly the seat a game
// then wants to stop offering to new players. CloseEmptySeat resolved its
// target through EnsureValid, which stepped off that inactive seat onto the
// next player who may be active -- a FILLED one, in any partly-seated game --
// so Legal failed with "already filled, not empty" and the seats stayed open
// forever.
func TestCloseEmptySeatClosesAnInactivatedSeat(t *testing.T) {

	storage := &dropInStorage{StorageManager: memory.NewStorageManager()}

	manager, err := boardgame.NewGameManager(&closeSeatsDelegate{}, storage)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	storage.setSeat(&dropInSeat{storage: storage, remaining: 2})

	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("new game: %v", err)
	}

	players := dropInPlayers(t, game)

	if players[2].SeatFilled || players[3].SeatFilled {
		t.Fatal("more seats were filled than the rendezvous offered")
	}
	if !players[2].PlayerInactive || !players[3].PlayerInactive {
		t.Fatal("setup did not inactivate the empty seats, so this test is not exercising the bug")
	}
	if !players[2].SeatClosed || !players[3].SeatClosed {
		t.Fatalf("CloseEmptySeat never closed the inactivated empty seats: seat 2 closed=%v, seat 3 closed=%v",
			players[2].SeatClosed, players[3].SeatClosed)
	}
}

/*
The three activation moves are one verb with three seat filters, and the two
facts below are what "one verb" has to mean if the taxonomy is to be honest.

The first is that ActivateInactivePlayer really is the UNION: for every
(seat filled?, player inactive?) combination it accepts a target exactly when
one of the two specific moves would. If that ever stops being true, the three
types have started drifting into three different rules wearing similar names,
which is the thing the shared activationTarget exists to prevent.

The second is that the marker interface names a CAPABILITY, not a type.
ActivateFilledSeat implements it because it genuinely undoes SeatPlayer's
inactivation of a real player; ActivateEmptySeat must NOT, because reopening an
empty seat undoes nothing SeatPlayer did to anyone.
*/
func TestActivateInactivePlayerIsTheUnionOfTheOtherTwo(t *testing.T) {

	manager, err := boardgame.NewGameManager(&dropInDelegate{}, memory.NewStorageManager())
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("new game: %v", err)
	}
	state := game.CurrentState().(boardgame.State)

	//Player 0 is the one whose seat/activity we vary; every combination is
	//reachable because both flags are plain behavior fields.
	player := state.PlayerStates()[0].(*dropInPlayerState)

	for _, filled := range []bool{false, true} {
		for _, inactive := range []bool{false, true} {
			player.SeatFilled = filled
			if inactive {
				player.SetPlayerInactive()
			} else {
				player.SetPlayerActive()
			}

			unionAccepts := activateAnySeat.legalTarget(state, 0) == nil
			emptyAccepts := activateEmptySeatTarget.legalTarget(state, 0) == nil
			filledAccepts := activateFilledSeatTarget.legalTarget(state, 0) == nil

			if unionAccepts != (emptyAccepts || filledAccepts) {
				t.Errorf("filled=%v inactive=%v: ActivateInactivePlayer accepts=%v but ActivateEmptySeat=%v / ActivateFilledSeat=%v; the union has drifted from its two halves",
					filled, inactive, unionAccepts, emptyAccepts, filledAccepts)
			}
			if emptyAccepts && filledAccepts {
				t.Errorf("filled=%v inactive=%v: both specific moves accepted the same seat, so they are no longer complements", filled, inactive)
			}
		}
	}
}

func TestSeatedPlayerActivatorNamesACapabilityNotAType(t *testing.T) {

	if activator, ok := interface{}(new(ActivateInactivePlayer)).(interfaces.SeatedPlayerActivator); !ok || !activator.ActivatesSeatedPlayers() {
		t.Error("ActivateInactivePlayer should satisfy the boot-time pairing check: it activates filled seats among others")
	}

	if activator, ok := interface{}(new(ActivateFilledSeat)).(interfaces.SeatedPlayerActivator); !ok || !activator.ActivatesSeatedPlayers() {
		t.Error("ActivateFilledSeat should satisfy the boot-time pairing check: undoing SeatPlayer's inactivation is its whole job")
	}

	if _, ok := interface{}(new(ActivateEmptySeat)).(interfaces.SeatedPlayerActivator); ok {
		t.Error("ActivateEmptySeat must NOT satisfy the pairing check: it never activates a seat anyone is sitting in, so accepting it would let a game boot with no way to activate the players it seats")
	}
}
