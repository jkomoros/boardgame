package werewolf

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/moves/interfaces"
	"github.com/jkomoros/boardgame/storage/memory"
)

// Rendezvous data type strings the SeatPlayer move uses to coordinate
// with a server-like context. Duplicated from moves/seat_player.go (they
// are also duplicated in server/api and lib/golden — the framework keys
// them by string, not by exported const).
const playerToSeatRendevousDataType = "github.com/jkomoros/boardgame/server/api.PlayerToSeat"
const willSeatPlayerRendevousDataType = "github.com/jkomoros/boardgame/server/api.WillSeatPlayer"

type testPlayerToSeat struct {
	index boardgame.PlayerIndex
	s     *testStorageManager
}

func (p *testPlayerToSeat) SeatIndex() boardgame.PlayerIndex { return p.index }
func (p *testPlayerToSeat) Committed()                       { p.s.playerToSeat = nil }

var _ interfaces.SeatPlayerSignaler = &testPlayerToSeat{}

// testStorageManager wraps the memory storage manager and implements the
// SeatPlayer rendezvous protocol so tests can drive the gathering flow
// the same way the real server does.
type testStorageManager struct {
	*memory.StorageManager
	playerToSeat *testPlayerToSeat
}

func (s *testStorageManager) FetchInjectedDataForGame(gameID string, dataType string) interface{} {
	if dataType == willSeatPlayerRendevousDataType {
		return true
	}
	if dataType == playerToSeatRendevousDataType {
		if s.playerToSeat == nil {
			return nil
		}
		return s.playerToSeat
	}
	return s.StorageManager.FetchInjectedDataForGame(gameID, dataType)
}

// newSeatedGame creates a werewolf game and seats numToSeat players via
// the real SeatPlayer flow. The game has the default 5 slots; once
// MinNumPlayers (4) are seated, WaitForEnoughPlayers fires,
// InactivateEmptySeat marks unfilled seats inactive, and moveBeginGame
// assigns roles among the active players.
func newSeatedGame(t *testing.T, numToSeat int) (*boardgame.GameManager, *boardgame.Game) {
	t.Helper()

	storage := &testStorageManager{memory.NewStorageManager(), nil}

	manager, err := boardgame.NewGameManager(NewDelegate(), storage)
	if err != nil {
		t.Fatalf("NewGameManager: %v", err)
	}

	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("NewDefaultGame: %v", err)
	}

	for i := 0; i < numToSeat; i++ {
		storage.playerToSeat = &testPlayerToSeat{boardgame.PlayerIndex(i), storage}
		if err := <-manager.Internals().ForceFixUp(game); err != nil {
			t.Fatalf("ForceFixUp seating player %d: %v", i, err)
		}
	}

	return manager, game
}

func TestNewGameManager(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), memory.NewStorageManager())
	if err != nil {
		t.Fatalf("NewGameManager: %v", err)
	}
	if manager == nil {
		t.Fatal("NewGameManager returned nil")
	}

	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("NewDefaultGame: %v", err)
	}
	if game == nil {
		t.Fatal("NewDefaultGame returned nil")
	}
}

// TestRolesOnlyAssignedToActivePlayers pins the fix for the
// empty-seat-werewolf bug: a 5-slot game seating only 4 players must
// deal exactly 1 werewolf among the 4 ACTIVE players. If roles were
// dealt across all 5 slots (the old FinishSetUp behavior), the wolf
// could land on the never-filled seat — instantly unwinnable.
func TestRolesOnlyAssignedToActivePlayers(t *testing.T) {
	_, game := newSeatedGame(t, 4)

	state := game.CurrentState()
	gameState, players := concreteStates(state)

	if gameState.Phase.Value() != phaseDay {
		t.Fatalf("Expected game to reach phaseDay after seating 4 players, got phase %d", gameState.Phase.Value())
	}

	activeWerewolves := 0
	for i, p := range players {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		if p.Role.Value() == roleWerewolf {
			activeWerewolves++
		}
		_ = i
	}
	if activeWerewolves != 1 {
		t.Errorf("Expected exactly 1 werewolf among 4 active players, got %d", activeWerewolves)
	}
}

// TestRoleHiddenFromObserver pins the core companion-mode privacy
// contract: the Table surface connects as ObserverPlayerIndex, so the
// framework's sanitization with behaviors.PlayerRole's default
// sanitize:"other:hidden" tag hides every player's role from the
// projector. If this test fails, the projector would display everyone's
// secret role on the shared screen.
func TestRoleHiddenFromObserver(t *testing.T) {
	_, game := newSeatedGame(t, 4)

	state := game.CurrentState()

	// Sanity: roles assigned (at least one active werewolf exists).
	_, players := concreteStates(state)
	hasWerewolf := false
	for _, p := range players {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		if p.Role.Value() == roleWerewolf {
			hasWerewolf = true
			break
		}
	}
	if !hasWerewolf {
		t.Fatal("Expected at least one werewolf after game setup")
	}

	sanitized, err := state.SanitizedForPlayer(boardgame.ObserverPlayerIndex)
	if err != nil {
		t.Fatalf("SanitizedForPlayer(Observer): %v", err)
	}

	for i, ps := range sanitized.ImmutablePlayerStates() {
		reader := ps.Reader()
		roleProp, err := reader.ImmutableEnumProp("Role")
		if err != nil {
			t.Fatalf("Player %d: reading Role prop: %v", i, err)
		}
		if roleProp != nil && roleProp.Value() != 0 {
			t.Errorf("Player %d: observer can see Role=%d; expected hidden (0)", i, roleProp.Value())
		}
	}
}

// TestRoleVisibleToSelf pins the other half of the companion-mode
// contract: each player's phone (Hand surface) connects as
// PlayerIndex(n), and sanitize:"other:hidden" only hides OTHER players'
// roles — you can always see your own. If this test fails, phones would
// show "Unknown" for the player's own role.
func TestRoleVisibleToSelf(t *testing.T) {
	_, game := newSeatedGame(t, 4)

	state := game.CurrentState()

	_, players := concreteStates(state)
	werewolfIdx := -1
	for i, p := range players {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		if p.Role.Value() == roleWerewolf {
			werewolfIdx = i
			break
		}
	}
	if werewolfIdx < 0 {
		t.Fatal("No werewolf found")
	}

	sanitized, err := state.SanitizedForPlayer(boardgame.PlayerIndex(werewolfIdx))
	if err != nil {
		t.Fatalf("SanitizedForPlayer(%d): %v", werewolfIdx, err)
	}

	selfState := sanitized.ImmutablePlayerStates()[werewolfIdx]
	reader := selfState.Reader()
	roleProp, err := reader.ImmutableEnumProp("Role")
	if err != nil {
		t.Fatalf("Reading own Role: %v", err)
	}
	if roleProp == nil || roleProp.Value() != roleWerewolf {
		t.Errorf("Player %d cannot see their own werewolf role; got %v", werewolfIdx, roleProp)
	}
}

// TestNightVoteSanitization pins the more subtle hidden-role privacy
// contract: whether a player has voted at night identifies them as a wolf,
// and their target is secret too. Day votes remain public by design.
func TestNightVoteSanitization(t *testing.T) {
	manager, game := newSeatedGame(t, 4)
	inflater := manager.Internals().StructInflater(boardgame.StatePropertyRef{Group: boardgame.StateGroupPlayer})
	if inflater == nil {
		t.Fatal("player-state inflater not found")
	}
	if got := inflater.PropertySanitizationPolicy("NightVote")[boardgame.SanitizationDefaultPlayerGroup]; got != boardgame.PolicyHidden {
		t.Fatalf("NightVote schema policy for other players = %v; want hidden", got)
	}
	policy := manager.Delegate().SanitizationPolicy(
		boardgame.StatePropertyRef{Group: boardgame.StateGroupPlayer, PropName: "NightVote"},
		map[string]bool{boardgame.SanitizationDefaultGroup: true, boardgame.SanitizationDefaultPlayerGroup: true},
	)
	if policy != boardgame.PolicyHidden {
		t.Fatalf("NightVote effective policy for another player = %v; want hidden", policy)
	}

	state := game.CurrentState()
	_, players := concreteStates(state)
	wolfIndex := -1
	for i, p := range players {
		if !behaviors.PlayerIsInactive(p) && p.Role.Value() == roleWerewolf {
			wolfIndex = i
			break
		}
	}
	if wolfIndex < 0 {
		t.Fatal("no werewolf assigned in seated game")
	}

	// Cast one public day vote and prove observers receive it.
	dayMove := game.MoveByName("Cast Vote").(*moveCastVote)
	dayMove.VoteTarget = 1
	if err := <-game.ProposeMove(dayMove, 0); err != nil {
		t.Fatalf("cast first day vote: %v", err)
	}
	state = game.CurrentState()
	dayState, err := state.SanitizedForPlayer(boardgame.ObserverPlayerIndex)
	if err != nil {
		t.Fatalf("sanitize day state for observer: %v", err)
	}
	dayVote, err := dayState.ImmutablePlayerStates()[0].Reader().PlayerIndexProp("DayVote")
	if err != nil || dayVote != 1 {
		t.Fatalf("observer DayVote = %d, %v; want public target 1", dayVote, err)
	}

	// Finish day voting through the real proposal/fix-up flow so the committed
	// state reaches Night before the wolf votes.
	for i := 1; i < 4; i++ {
		move := game.MoveByName("Cast Vote").(*moveCastVote)
		move.VoteTarget = boardgame.PlayerIndex((i + 1) % 4)
		if err := <-game.ProposeMove(move, boardgame.PlayerIndex(i)); err != nil {
			t.Fatalf("cast day vote for player %d: %v", i, err)
		}
	}
	state = game.CurrentState()
	gameState, _ := concreteStates(state)
	if gameState.Phase.Value() != phaseNight {
		t.Fatalf("phase after all day votes = %d; want Night", gameState.Phase.Value())
	}
	// The four-player fixture normally has one wolf, so its first night vote
	// would immediately resolve and disappear. Promote one active villager in
	// this test fixture to keep a real, committed mid-night state available.
	_, players = concreteStates(state)
	secondWolf := -1
	for i, p := range players {
		if i != wolfIndex && !behaviors.PlayerIsInactive(p) {
			p.Role.SetValue(roleWerewolf)
			secondWolf = i
			break
		}
	}
	if secondWolf < 0 {
		t.Fatal("could not create second wolf for mid-night privacy fixture")
	}
	populateFellowWolves(players)

	target := boardgame.PlayerIndex((wolfIndex + 1) % 4)
	if target == 0 || int(target) == secondWolf {
		target = 1 // Hidden PlayerIndex values sanitize to 0; use a distinct target.
	}
	if int(target) == wolfIndex || int(target) == secondWolf {
		target = 2
	}
	nightMove := game.MoveByName("Cast Night Vote").(*moveCastVote)
	nightMove.VoteTarget = target
	if err := <-game.ProposeMove(nightMove, boardgame.PlayerIndex(wolfIndex)); err != nil {
		t.Fatalf("cast night vote: %v", err)
	}
	state = game.CurrentState()

	assertVotes := func(viewer boardgame.PlayerIndex, wantNight boardgame.PlayerIndex) {
		t.Helper()
		sanitized, err := state.SanitizedForPlayer(viewer)
		if err != nil {
			t.Fatalf("SanitizedForPlayer(%d): %v", viewer, err)
		}
		reader := sanitized.ImmutablePlayerStates()[wolfIndex].Reader()
		nightVote, err := reader.PlayerIndexProp("NightVote")
		if err != nil {
			t.Fatalf("reading NightVote: %v", err)
		}
		if nightVote != wantNight {
			t.Errorf("viewer %d sees NightVote=%d; want %d", viewer, nightVote, wantNight)
		}
	}

	assertVotes(boardgame.ObserverPlayerIndex, 0)
	assertVotes(boardgame.PlayerIndex((wolfIndex+2)%4), 0)
	assertVotes(boardgame.PlayerIndex(wolfIndex), target)
}

// TestFellowWolvesPrivateToOwner verifies both halves of the teammate
// contract: a wolf receives the exact other-wolf roster, while villagers,
// observers, and other player-state records receive an empty slice.
func TestFellowWolvesPrivateToOwner(t *testing.T) {
	_, game := newSeatedGame(t, 4)

	state := game.CurrentState()
	_, players := concreteStates(state)
	for i, p := range players {
		if behaviors.PlayerIsInactive(p) {
			continue
		}
		if i == 0 || i == 1 {
			p.Role.SetValue(roleWerewolf)
		} else {
			p.Role.SetValue(roleVillager)
		}
	}
	populateFellowWolves(players)

	assertFellows := func(viewer, owner boardgame.PlayerIndex, want []boardgame.PlayerIndex) {
		t.Helper()
		sanitized, err := state.SanitizedForPlayer(viewer)
		if err != nil {
			t.Fatalf("SanitizedForPlayer(%d): %v", viewer, err)
		}
		got, err := sanitized.ImmutablePlayerStates()[owner].Reader().PlayerIndexSliceProp("FellowWolves")
		if err != nil {
			t.Fatalf("reading FellowWolves: %v", err)
		}
		if len(got) != len(want) {
			t.Fatalf("viewer %d sees player %d fellows %v; want %v", viewer, owner, got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("viewer %d sees player %d fellows %v; want %v", viewer, owner, got, want)
			}
		}
	}

	assertFellows(0, 0, []boardgame.PlayerIndex{1})
	assertFellows(1, 1, []boardgame.PlayerIndex{0})
	assertFellows(2, 0, nil)
	assertFellows(boardgame.ObserverPlayerIndex, 0, nil)
	assertFellows(0, 1, nil)
}

// TestCheckGameFinishedTeamWinners pins the team-based winner computation:
// the winning ROLE's members (eliminated teammates included) are the
// winners — never "everyone", which is what base.GameDelegate's default
// scoring produced before werewolf implemented CheckGameFinished itself.
func TestCheckGameFinishedTeamWinners(t *testing.T) {
	delegate := NewDelegate().(*gameDelegate)
	_, game := newSeatedGame(t, 4)

	state := game.CurrentState()
	_, players := concreteStates(state)

	var wolfIndex boardgame.PlayerIndex = -1
	for i, p := range players {
		if p.Role.Value() == roleWerewolf {
			wolfIndex = boardgame.PlayerIndex(i)
			break
		}
	}
	if wolfIndex < 0 {
		t.Fatal("no werewolf assigned in seated game")
	}

	// Mid-game: nobody eliminated → not finished.
	if finished, _ := delegate.CheckGameFinished(state); finished {
		t.Fatal("game reported finished with all players alive")
	}

	// Villagers win: eliminate the wolf. Winners must be exactly the
	// non-wolves, and must NOT include the wolf.
	players[wolfIndex].Eliminated = true

	finished, winners := delegate.CheckGameFinished(state)
	if !finished {
		t.Fatal("game not finished after last wolf eliminated")
	}
	activeCount := 0
	for _, p := range players {
		if !behaviors.PlayerIsInactive(p) {
			activeCount++
		}
	}
	if len(winners) != activeCount-1 {
		t.Fatalf("expected %d villager winners, got %d", activeCount-1, len(winners))
	}
	for _, w := range winners {
		if w == wolfIndex {
			t.Fatal("losing werewolf listed among winners")
		}
	}

	// Werewolves win: revive the wolf, eliminate villagers until parity.
	players[wolfIndex].Eliminated = false
	eliminated := 0
	for i, p := range players {
		if boardgame.PlayerIndex(i) == wolfIndex {
			continue
		}
		if eliminated < 2 {
			p.Eliminated = true
			eliminated++
		}
	}
	// 1 wolf vs 1 villager alive → wolves win.
	finished, winners = delegate.CheckGameFinished(state)
	if !finished {
		t.Fatal("game not finished at wolf parity")
	}
	if len(winners) != 1 || winners[0] != wolfIndex {
		t.Fatalf("expected sole winner %v (the wolf), got %v", wolfIndex, winners)
	}
}
