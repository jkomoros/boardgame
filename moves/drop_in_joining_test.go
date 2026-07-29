package moves

import (
	"errors"
	"sync"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/storage/memory"
)

/*
This is the end-to-end regression net for drop-in joining: can a player seated
AFTER the game has started actually take a turn?

The game configured below is the shape the scaffolding stub generates.
SeatPlayer is legal in both phases; phaseSetUp gets DefaultRoundSetup (whose
ActivateInactivePlayer activates everyone, correctly, because no seat has been
closed yet, and whose InactivateEmptySeat then closes the unfilled ones); and
phaseNormal gets ActivateFilledSeat.

The load-bearing assertion is not that the late joiner's Inactive flag flips.
It is that PlayerIndex.Next reaches them: turn order skips inactive players, so
a player who is seated but never activated is a permanent spectator no matter
what the seat flags say. The test therefore plays turns until the late joiner is
the current player and has them apply a real CurrentPlayer move.
*/

const (
	dropInPhaseSetUp = iota
	dropInPhaseNormal
)

var dropInEnums = enum.NewSet()

var dropInPhaseEnum = dropInEnums.MustAdd("phase", map[enum.EnumKey]string{
	dropInPhaseSetUp:  "Set Up",
	dropInPhaseNormal: "Normal",
})

//boardgame:codegen
type dropInGameState struct {
	base.SubState
	behaviors.PhaseBehavior
	behaviors.CurrentPlayerBehavior
}

//boardgame:codegen
type dropInPlayerState struct {
	base.SubState
	behaviors.Seat
	behaviors.InactivePlayer
	Acted      bool
	TimesActed int
}

func (p *dropInPlayerState) TurnDone() error {
	if !p.Acted {
		return errors.New("they have not acted yet")
	}
	return nil
}

func (p *dropInPlayerState) ResetForTurnStart() error {
	p.Acted = false
	return nil
}

func (p *dropInPlayerState) ResetForTurnEnd() error {
	return nil
}

//boardgame:codegen
type moveDropInAct struct {
	CurrentPlayer
}

func (m *moveDropInAct) Apply(state boardgame.State) error {
	player := state.PlayerStates()[m.TargetPlayerIndex.EnsureValid(state)].(*dropInPlayerState)
	player.Acted = true
	player.TimesActed++
	return nil
}

type dropInDelegate struct {
	base.GameDelegate
}

func (g *dropInDelegate) Name() string { return "moves" }

func (g *dropInDelegate) DefaultNumPlayers() int { return 4 }

func (g *dropInDelegate) MinNumPlayers() int { return 2 }

func (g *dropInDelegate) ConfigureEnums() *enum.Set { return dropInEnums }

func (g *dropInDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{}
}

func (g *dropInDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(dropInGameState)
}

func (g *dropInDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return new(dropInPlayerState)
}

func (g *dropInDelegate) CurrentPlayerIndex(state boardgame.ImmutableState) boardgame.PlayerIndex {
	return state.ImmutableGameState().(*dropInGameState).CurrentPlayer
}

func (g *dropInDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	return nil, nil
}

// ConfigureMoves is the generated stub's shape, verbatim in structure.
func (g *dropInDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := NewAutoConfigurer(g)
	return Combine(
		AddForPhase(dropInPhaseSetUp,
			auto.MustConfig(new(SeatPlayer)),
		),
		AddForPhase(dropInPhaseNormal,
			auto.MustConfig(new(SeatPlayer),
				WithMoveNameSuffix("Mid Game"),
			),
			auto.MustConfig(new(ActivateFilledSeat)),
		),
		AddOrderedForPhase(dropInPhaseSetUp,
			DefaultRoundSetup(auto),
			auto.MustConfig(new(StartPhase),
				WithPhaseToStart(dropInPhaseNormal, dropInPhaseEnum),
			),
		),
		AddForPhase(dropInPhaseNormal,
			auto.MustConfig(new(moveDropInAct),
				WithMoveName("Drop In Act"),
			),
			auto.MustConfig(new(FinishTurn)),
		),
	)
}

// dropInSeat is the server's side of the seating rendezvous. Like the real
// server it hands back AdminPlayerIndex, meaning "any open seat", and it
// withdraws itself once the seating commits -- which is what stops SeatPlayer's
// fix-up pass from filling every remaining seat at once.
type dropInSeat struct {
	storage   *dropInStorage
	remaining int
}

func (d *dropInSeat) SeatIndex() boardgame.PlayerIndex { return boardgame.AdminPlayerIndex }

func (d *dropInSeat) Committed() {
	d.remaining--
	if d.remaining <= 0 {
		d.storage.setSeat(nil)
	}
}

// dropInStorage stands in for the server storage manager, which is the only
// thing that ever tells a game it will seat players.
type dropInStorage struct {
	*memory.StorageManager
	mu   sync.Mutex
	seat *dropInSeat
}

func (d *dropInStorage) setSeat(seat *dropInSeat) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seat = seat
}

func (d *dropInStorage) FetchInjectedDataForGame(gameID string, dataType string) interface{} {
	d.mu.Lock()
	defer d.mu.Unlock()
	switch dataType {
	case willSeatPlayerRendevousDataType:
		return true
	case playerToSeatRendevousDataType:
		if d.seat == nil {
			return nil
		}
		return d.seat
	}
	return nil
}

func dropInPlayers(t *testing.T, game *boardgame.Game) []*dropInPlayerState {
	t.Helper()
	result := make([]*dropInPlayerState, len(game.CurrentState().ImmutablePlayerStates()))
	for i, p := range game.CurrentState().ImmutablePlayerStates() {
		result[i] = p.(*dropInPlayerState)
	}
	return result
}

func TestDropInJoining(t *testing.T) {

	storage := &dropInStorage{StorageManager: memory.NewStorageManager()}

	manager, err := boardgame.NewGameManager(&dropInDelegate{}, storage)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	//Two of the four seats get filled before the game starts. SeatPlayer is a
	//fix-up, so the rendezvous being present during the game's first fix-up
	//pass is all it takes.
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

	//DefaultRoundSetup's InactivateEmptySeat closed the two unfilled seats.
	if !players[2].PlayerInactive || !players[3].PlayerInactive {
		t.Fatal("the unfilled seats were not inactivated during setup")
	}
	if players[0].PlayerInactive || players[1].PlayerInactive {
		t.Fatal("the seated players were not activated during setup")
	}

	//Now the actual question: seat someone into an open seat AFTER the game
	//started, and let the normal-play fix-up pass handle them.
	storage.setSeat(&dropInSeat{storage: storage, remaining: 1})
	if err := <-game.ProposeMove(actMoveFor(t, game), 0); err != nil {
		t.Fatalf("player 0 acting: %v", err)
	}

	players = dropInPlayers(t, game)

	if !players[2].SeatFilled {
		t.Fatal("the late joiner was never seated")
	}
	if players[2].PlayerInactive {
		t.Fatal("the late joiner was seated but never activated: a silent permanent spectator")
	}
	//And the seat nobody is sitting in stayed closed. This is the whole reason
	//ActivateInactivePlayer could not be used here.
	if !players[3].PlayerInactive {
		t.Fatal("activating the late joiner also reopened an empty seat")
	}

	//The load-bearing check: turn order actually reaches them. Turn order
	//skips inactive players, so a player who is seated but not activated can
	//never become the current player no matter what their seat flags say.
	var reached bool
	for i := 0; i < 8; i++ {
		current := game.CurrentState().(boardgame.State).CurrentPlayerIndex()
		if current == 2 {
			reached = true
			break
		}
		if err := <-game.ProposeMove(actMoveFor(t, game), current); err != nil {
			t.Fatalf("player %v acting: %v", current, err)
		}
	}

	if !reached {
		t.Fatal("turn order never reached the late joiner, so they can never take a turn")
	}

	if err := <-game.ProposeMove(actMoveFor(t, game), 2); err != nil {
		t.Fatalf("the late joiner could not take their turn: %v", err)
	}

	players = dropInPlayers(t, game)
	if players[2].TimesActed < 1 {
		t.Fatal("the late joiner's move did not apply")
	}
	if players[3].TimesActed != 0 {
		t.Fatal("an empty seat took a turn")
	}
}

func actMoveFor(t *testing.T, game *boardgame.Game) boardgame.Move {
	t.Helper()
	move := game.MoveByName("Drop In Act")
	if move == nil {
		t.Fatal("no \"Drop In Act\" move installed")
	}
	return move
}
