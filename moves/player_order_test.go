package moves

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/storage/memory"
)

/*
The connected half of PlayerOrderBehavior's tests.

`behaviors/player_order.go` scored 14% under a mutation pass -- the second
worst file measured -- and the survivors were not scattered: EVERY guard in
buildAndValidateOrder lived, including all five mutants on
`if v < 0 || v >= numPlayers` (disable the guard, drop either clause, `>=` ->
`>`, and the literal +-1), plus both loop bodies and the cache guard. A
behavior whose entire job is to reject a malformed player order was
indistinguishable from one that accepts anything.

The reason none of it was covered is stated in behaviors/main_test.go: the
validator needs a real game state to count players from, and `base` imports
`behaviors`, so the behaviors package's own test cannot build one. It builds
here instead, in the package that already has the machinery.

What the behavior promises, and what each test below pins:

  - A valid permutation comes back as given.
  - An order of the wrong LENGTH, one with an out-of-range index, or one with
    a duplicate is REJECTED SILENTLY -- PlayerOrder returns nil, meaning
    "default sequential", and logs a warning. Silent is the deliberate choice
    (a corrupt saved state must still be playable) and it is exactly what
    makes the rejection invisible to a test that only asks for the happy path.
  - SetPlayerOrder, unlike PlayerOrder, rejects LOUDLY, and its validation is
    a separate loop with its own survivors.
*/

//boardgame:codegen
type playerOrderGameState struct {
	base.SubState
	behaviors.PlayerOrderBehavior
}

//boardgame:codegen
type playerOrderPlayerState struct {
	base.SubState
}

// playerOrderNoOpMove exists only because a manager needs at least one move
// installed. It is a player move rather than a fix-up so the engine does not
// spin on it.
//
//boardgame:codegen
type playerOrderNoOpMove struct {
	Default
}

func (p *playerOrderNoOpMove) Apply(state boardgame.State) error { return nil }

type playerOrderDelegate struct {
	base.GameDelegate
}

func (p *playerOrderDelegate) Name() string { return "moves" }

func (p *playerOrderDelegate) DefaultNumPlayers() int { return 3 }

func (p *playerOrderDelegate) MinNumPlayers() int { return 3 }

func (p *playerOrderDelegate) MaxNumPlayers() int { return 3 }

func (p *playerOrderDelegate) ConfigureEnums() *enum.Set { return enum.NewSet() }

func (p *playerOrderDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{}
}

func (p *playerOrderDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(playerOrderGameState)
}

func (p *playerOrderDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return new(playerOrderPlayerState)
}

func (p *playerOrderDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := NewAutoConfigurer(p)
	return Add(
		auto.MustConfig(new(playerOrderNoOpMove)),
	)
}

func (p *playerOrderDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	return nil, nil
}

// newPlayerOrderGameState returns the game state of a fresh three-player game,
// with its PlayerOrderBehavior connected the way the framework connects it.
func newPlayerOrderGameState(t *testing.T) *playerOrderGameState {
	t.Helper()
	manager, err := boardgame.NewGameManager(&playerOrderDelegate{}, memory.NewStorageManager())
	if err != nil {
		t.Fatal("couldn't create manager:", err)
	}
	game, err := manager.NewGame(3, nil, nil)
	if err != nil {
		t.Fatal("couldn't create game:", err)
	}
	return game.CurrentState().ImmutableGameState().(*playerOrderGameState)
}

func TestPlayerOrderAcceptsAValidPermutation(t *testing.T) {

	gameState := newPlayerOrderGameState(t)

	if err := gameState.SetPlayerOrder([]boardgame.PlayerIndex{2, 0, 1}); err != nil {
		t.Fatalf("a valid permutation was rejected: %v", err)
	}

	order := gameState.PlayerOrder()

	if len(order) != 3 {
		t.Fatalf("expected three entries, got %v", order)
	}
	for i, want := range []boardgame.PlayerIndex{2, 0, 1} {
		if order[i] != want {
			t.Errorf("position %v is %v, expected %v", i, order[i], want)
		}
	}

	//Read twice: the second call comes from the cache, and a cache that never
	//gets built hands back nil (i.e. default sequential) instead.
	again := gameState.PlayerOrder()
	for i, want := range []boardgame.PlayerIndex{2, 0, 1} {
		if again[i] != want {
			t.Errorf("cached position %v is %v, expected %v", i, again[i], want)
		}
	}

	//And SetPlayerOrder invalidates that cache rather than leaving the old
	//answer in place.
	if err := gameState.SetPlayerOrder([]boardgame.PlayerIndex{1, 2, 0}); err != nil {
		t.Fatalf("a second valid permutation was rejected: %v", err)
	}
	if got := gameState.PlayerOrder()[0]; got != 1 {
		t.Errorf("the order was changed but PlayerOrder still says %v", got)
	}
}

func TestPlayerOrderRejectsAMalformedOrderSlice(t *testing.T) {

	//These go straight into OrderSlice rather than through SetPlayerOrder,
	//because SetPlayerOrder would refuse them -- and the point is what happens
	//when a state carrying a corrupt order is LOADED. The behavior falls back
	//to default sequential order and warns; it does not error, and it does not
	//hand back a corrupt order.
	for _, test := range []struct {
		description string
		slice       []int
	}{
		{"too short", []int{0, 1}},
		{"too long", []int{0, 1, 2, 0}},
		{"an index at the number of players", []int{0, 1, 3}},
		{"an index above the number of players", []int{0, 1, 9}},
		{"a negative index", []int{0, -1, 2}},
		{"a duplicate", []int{0, 1, 1}},
	} {
		gameState := newPlayerOrderGameState(t)
		gameState.OrderSlice = test.slice

		if order := gameState.PlayerOrder(); order != nil {
			t.Errorf("%v (%v) was accepted as a player order: %v",
				test.description, test.slice, order)
		}
	}

	//The boundary the range check is written on: the LAST valid index is
	//numPlayers - 1, and it has to be accepted, or `>=` -> `>` would be the
	//only reading that passes.
	gameState := newPlayerOrderGameState(t)
	gameState.OrderSlice = []int{2, 1, 0}
	if order := gameState.PlayerOrder(); order == nil {
		t.Error("a permutation using the highest valid index was rejected")
	}
}

func TestSetPlayerOrderRejectsLoudly(t *testing.T) {

	for _, test := range []struct {
		description string
		order       []boardgame.PlayerIndex
		wantMessage string
	}{
		{"too short", []boardgame.PlayerIndex{0, 1}, "does not match"},
		{"out of range", []boardgame.PlayerIndex{0, 1, 3}, "out of range"},
		{"negative", []boardgame.PlayerIndex{0, -1, 2}, "out of range"},
		{"a duplicate", []boardgame.PlayerIndex{0, 1, 1}, "duplicate"},
	} {
		gameState := newPlayerOrderGameState(t)
		err := gameState.SetPlayerOrder(test.order)
		if err == nil {
			t.Errorf("SetPlayerOrder accepted %v (%v)", test.description, test.order)
			continue
		}
		if !strings.Contains(err.Error(), test.wantMessage) {
			t.Errorf("%v was rejected, but for the wrong reason: %v", test.description, err)
		}
		//And nothing was written: a refused order must not half-apply.
		if len(gameState.OrderSlice) != 0 {
			t.Errorf("%v was refused but left OrderSlice = %v", test.description, gameState.OrderSlice)
		}
	}
}

func TestReversePlayerOrder(t *testing.T) {

	gameState := newPlayerOrderGameState(t)

	if err := gameState.ReversePlayerOrder(gameState.State()); err != nil {
		t.Fatalf("ReversePlayerOrder failed: %v", err)
	}

	order := gameState.PlayerOrder()
	for i, want := range []boardgame.PlayerIndex{2, 1, 0} {
		if order[i] != want {
			t.Errorf("reversed position %v is %v, expected %v", i, order[i], want)
		}
	}
}
