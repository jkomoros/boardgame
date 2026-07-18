package boardgame

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func assertPersistedStatesEqual(t *testing.T, actual, expected ImmutableState) {
	t.Helper()
	actualState, actualOK := actual.(*state)
	expectedState, expectedOK := expected.(*state)
	if !actualOK || !expectedOK {
		t.Fatalf("states are not framework states: %T, %T", actual, expected)
	}
	if !bytes.Equal(actualState.StorageRecord(), expectedState.StorageRecord()) {
		t.Fatalf("persisted states differ:\nactual: %s\nexpected: %s", actualState.StorageRecord(), expectedState.StorageRecord())
	}
}

func attachStackForPrimitiveTest(st *state, stack Stack) {
	stack.setState(st)
	if st.stackOwners == nil {
		st.stackOwners = make(map[Stack]stackOwner)
	}
	path := fmt.Sprintf("TestOnly[%p]", stack)
	st.stackOwners[stack] = stackOwner{path: path, current: func() (Stack, error) { return stack, nil }}
}

type detachedDistributionDelegate struct {
	*testGameDelegate
}

type wrongDeckDistributionDelegate struct {
	*testGameDelegate
}

func (d *wrongDeckDistributionDelegate) ConfigureDecks() map[string]*Deck {
	decks := d.testGameDelegate.ConfigureDecks()
	other := NewDeck()
	other.AddComponent(&testingComponent{String: "wrong deck"})
	decks["other"] = other
	return decks
}

func (d *detachedDistributionDelegate) DistributeComponentToStarterStack(state ImmutableState, c Component) (ImmutableStack, error) {
	return c.Deck().NewStack(0), nil
}

func (d *detachedDistributionDelegate) FinishSetUp(state State) error {
	return nil
}

func TestSetupRejectsDetachedDistributionStack(t *testing.T) {
	delegate := &detachedDistributionDelegate{testGameDelegate: defaultTestGameDelegate(0)}
	manager, err := NewGameManager(delegate, newTestStorageManager())
	if err != nil {
		t.Fatalf("NewGameManager: %v", err)
	}
	game, err := manager.newGameImpl("", "")
	if err != nil {
		t.Fatalf("newGameImpl: %v", err)
	}

	err = game.setUp(delegate.DefaultNumPlayers(), nil, nil)
	if err == nil {
		t.Fatal("setUp accepted a detached distribution stack")
	}
	if !strings.Contains(err.Error(), "attached") {
		t.Fatalf("setUp error %q does not explain attachment", err)
	}
}

func TestSetupRejectsWrongDeckBeforeStorageWrite(t *testing.T) {
	storage := newTestStorageManager()
	delegate := &wrongDeckDistributionDelegate{testGameDelegate: defaultTestGameDelegate(0)}
	manager, err := NewGameManager(delegate, storage)
	if err != nil {
		t.Fatalf("NewGameManager: %v", err)
	}
	if _, err := manager.NewDefaultGame(); err == nil || !strings.Contains(err.Error(), "deck") {
		t.Fatalf("NewDefaultGame error = %v, want deck mismatch", err)
	}
	if len(storage.games) != 0 || len(storage.states) != 0 || len(storage.moves) != 0 {
		t.Fatalf("failed setup wrote storage: games=%d states=%d moves=%d", len(storage.games), len(storage.states), len(storage.moves))
	}
}

func TestMoveRejectsStackFromDifferentState(t *testing.T) {
	manager := newTestGameManger(t)
	first, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("first game: %v", err)
	}
	second, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("second game: %v", err)
	}

	firstState, _ := concreteStates(first.CurrentState())
	secondState, _ := concreteStates(second.CurrentState())
	beforeSource := firstState.DrawDeck.NumComponents()
	beforeDestination := secondState.DrawDeck.NumComponents()

	err = firstState.DrawDeck.First().MoveToNextSlot(secondState.DrawDeck)
	if err == nil {
		t.Fatal("cross-state move unexpectedly succeeded")
	}
	if !strings.Contains(err.Error(), "state") {
		t.Fatalf("cross-state error %q does not identify the state mismatch", err)
	}
	if got := firstState.DrawDeck.NumComponents(); got != beforeSource {
		t.Fatalf("source changed after rejected move: got %d, want %d", got, beforeSource)
	}
	if got := secondState.DrawDeck.NumComponents(); got != beforeDestination {
		t.Fatalf("destination changed after rejected move: got %d, want %d", got, beforeDestination)
	}
}

func TestMoveAllRejectsInvalidOwnersEvenWhenEmpty(t *testing.T) {
	manager := newTestGameManger(t)
	first, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("first game: %v", err)
	}
	second, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("second game: %v", err)
	}
	firstState, _ := concreteStates(first.CurrentState())
	secondState, _ := concreteStates(second.CurrentState())
	emptyFirst := firstState.MyBoard.SpaceAt(2)
	emptySecond := secondState.MyBoard.SpaceAt(2)

	if err := emptyFirst.MayMoveAllTo(emptySecond); err == nil || !strings.Contains(err.Error(), "different states") {
		t.Fatalf("cross-state empty MayMoveAllTo error = %v", err)
	}
	if err := emptyFirst.MoveAllTo(emptySecond); err == nil || !strings.Contains(err.Error(), "different states") {
		t.Fatalf("cross-state empty MoveAllTo error = %v", err)
	}
	if err := firstState.DrawDeck.MoveAllTo(nil); err == nil {
		t.Fatal("MoveAllTo(nil) unexpectedly succeeded")
	}
	detached := manager.Chest().Deck("test").NewStack(0)
	if err := detached.MoveAllTo(firstState.DrawDeck); err == nil || !strings.Contains(err.Error(), "state") {
		t.Fatalf("detached empty MoveAllTo error = %v", err)
	}
}

func TestMayMoveAllToSupportsBoardSpaces(t *testing.T) {
	game := testDefaultGame(t, false)
	gameState, players := concreteStates(game.CurrentState())
	if err := gameState.MyBoard.SpaceAt(1).MayMoveAllTo(players[0].Hand); err != nil {
		t.Fatalf("board-space source preflight: %v", err)
	}
	if err := gameState.DrawDeck.MayMoveAllTo(gameState.MyBoard.SpaceAt(2)); err != nil {
		t.Fatalf("board-space destination preflight: %v", err)
	}
}

func TestMoveRejectsReplacedStaleStack(t *testing.T) {
	game := testDefaultGame(t, false)
	gameState, players := concreteStates(game.CurrentState())
	stale := gameState.DrawDeck
	gameState.DrawDeck = game.Manager().Chest().Deck("test").NewStack(0)

	beforeSource := stale.NumComponents()
	beforeDestination := players[0].Hand.NumComponents()
	err := stale.First().MoveToNextSlot(players[0].Hand)
	if err == nil {
		t.Fatal("move from a replaced stale stack unexpectedly succeeded")
	}
	if !strings.Contains(err.Error(), "attached") && !strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale-stack error %q does not explain ownership", err)
	}
	if got := stale.NumComponents(); got != beforeSource {
		t.Fatalf("stale source changed after rejected move: got %d, want %d", got, beforeSource)
	}
	if got := players[0].Hand.NumComponents(); got != beforeDestination {
		t.Fatalf("destination changed after rejected move: got %d, want %d", got, beforeDestination)
	}
}

func TestStaleAndSanitizedStacksRejectNonMoveMutators(t *testing.T) {
	game := testDefaultGame(t, false)
	gameState, _ := concreteStates(game.CurrentState())
	stale := gameState.OtherStack
	gameState.OtherStack = game.Manager().Chest().Deck("test").NewSizedStack(stale.Len())

	checks := []struct {
		name string
		run  func() error
	}{
		{"resize", func() error { return stale.ExpandSize(1) }},
		{"constraint", func() error {
			return stale.AddConstraint(func(ImmutableStack, []ImmutableComponentInstance, ImmutableState) error { return nil })
		}},
		{"swap preflight", func() error { return stale.MaySwapComponents(0, 1) }},
	}
	for _, check := range checks {
		if err := check.run(); err == nil || (!strings.Contains(err.Error(), "stale") && !strings.Contains(err.Error(), "attached")) {
			t.Errorf("%s error = %v, want ownership failure", check.name, err)
		}
	}

	sanitized, err := game.CurrentState().SanitizedForPlayer(0)
	if err != nil {
		t.Fatalf("sanitize: %v", err)
	}
	sanitizedGame, _ := concreteStates(sanitized)
	if err := sanitizedGame.OtherStack.ExpandSize(1); err == nil || !strings.Contains(err.Error(), "sanitized") {
		t.Fatalf("sanitized resize error = %v", err)
	}
}

func TestMergedSetStateNeverReassignsBackingOwners(t *testing.T) {
	manager := newTestGameManger(t)
	first, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("first game: %v", err)
	}
	second, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("second game: %v", err)
	}
	firstGameState, _ := concreteStates(first.CurrentState())
	fresh := manager.Chest().Deck("test").NewStack(0)
	view := NewConcatenatedStack(fresh, firstGameState.DrawDeck)
	view.setState(second.CurrentState().(*state))
	if got := firstGameState.DrawDeck.state(); got != first.CurrentState() {
		t.Fatalf("merged view reassigned owned leaf to %p, want %p", got, first.CurrentState())
	}
	if got := fresh.state(); got != nil {
		t.Fatalf("merged view assigned state to detached leaf: %p", got)
	}
}

func TestMergedValidationRejectsNilLeafWithoutPanic(t *testing.T) {
	game := testDefaultGame(t, false)
	st := game.CurrentState().(*state)
	view := NewConcatenatedStack(nil)
	if err := view.Valid(); err == nil {
		t.Fatal("merged view with nil leaf unexpectedly valid")
	}
	if err := st.validateMergedOwnerLeaves("Game.BadView", view, make(map[MergedStack]bool), make(map[Stack]string)); err == nil || !strings.Contains(err.Error(), "nil") {
		t.Fatalf("merged owner validation error = %v", err)
	}
}

func TestOwnerRegistryRejectsAliasesAndSwaps(t *testing.T) {
	t.Run("property aliases board space", func(t *testing.T) {
		game := testDefaultGame(t, false)
		st := game.CurrentState().(*state)
		gameState, _ := concreteStates(st)
		gameState.DrawDeck = gameState.MyBoard.SpaceAt(0)
		if err := st.initializeStackOwners(); err == nil || !strings.Contains(err.Error(), "both") {
			t.Fatalf("alias error = %v", err)
		}
	})

	t.Run("swapped properties become stale", func(t *testing.T) {
		game := testDefaultGame(t, false)
		st := game.CurrentState().(*state)
		gameState, _ := concreteStates(st)
		gameState.DownSizeStack, gameState.OtherStack = gameState.OtherStack, gameState.DownSizeStack
		if err := st.validateComponentConservation(); err == nil || !strings.Contains(err.Error(), "stale") {
			t.Fatalf("swapped-owner error = %v", err)
		}
	})
}

func TestOwnerRegistryValidatesNestedMergedTopology(t *testing.T) {
	t.Run("nested empty view", func(t *testing.T) {
		view := NewConcatenatedStack(NewConcatenatedStack())
		if err := view.Valid(); err == nil || !strings.Contains(err.Error(), "no sub-stacks") {
			t.Fatalf("nested empty error = %v", err)
		}
	})

	t.Run("nested nil leaf does not panic", func(t *testing.T) {
		view := NewConcatenatedStack(NewConcatenatedStack(nil))
		if err := view.Valid(); err == nil || !strings.Contains(err.Error(), "nil") {
			t.Fatalf("nested nil error = %v", err)
		}
	})

	t.Run("nested mixed decks", func(t *testing.T) {
		game := testDefaultGame(t, false)
		gameState, _ := concreteStates(game.CurrentState())
		otherDeck := NewDeck()
		otherDeck.AddComponent(&testingComponent{String: "other"})
		view := NewConcatenatedStack(NewConcatenatedStack(gameState.DownSizeStack, otherDeck.NewSizedStack(1)), gameState.OtherStack)
		if err := view.Valid(); err == nil || !strings.Contains(err.Error(), "different deck") {
			t.Fatalf("nested mixed-deck error = %v", err)
		}
	})

	t.Run("nested invalid overlap", func(t *testing.T) {
		game := testDefaultGame(t, false)
		gameState, _ := concreteStates(game.CurrentState())
		view := NewConcatenatedStack(NewOverlappedStack(gameState.DownSizeStack, gameState.DrawDeck), gameState.OtherStack)
		if err := view.Valid(); err == nil || !strings.Contains(err.Error(), "fixed size") {
			t.Fatalf("nested overlap error = %v", err)
		}
	})

	t.Run("valid nested view", func(t *testing.T) {
		game := testDefaultGame(t, false)
		st := game.CurrentState().(*state)
		gameState, _ := concreteStates(st)
		gameState.MyMergedStack = NewConcatenatedStack(NewConcatenatedStack(gameState.DownSizeStack), gameState.OtherStack)
		if err := st.initializeStackOwners(); err != nil {
			t.Fatalf("valid nested view: %v", err)
		}
	})

	t.Run("repeated leaf", func(t *testing.T) {
		game := testDefaultGame(t, false)
		st := game.CurrentState().(*state)
		gameState, _ := concreteStates(st)
		gameState.MyMergedStack = NewConcatenatedStack(NewConcatenatedStack(gameState.DownSizeStack), gameState.DownSizeStack)
		if err := st.initializeStackOwners(); err == nil || !strings.Contains(err.Error(), "repeats") {
			t.Fatalf("repeated leaf error = %v", err)
		}
	})

	t.Run("hidden detached leaf", func(t *testing.T) {
		game := testDefaultGame(t, false)
		st := game.CurrentState().(*state)
		gameState, _ := concreteStates(st)
		detached := game.Manager().Chest().Deck("test").NewStack(0)
		gameState.MyMergedStack = NewConcatenatedStack(detached, gameState.DownSizeStack)
		if err := st.initializeStackOwners(); err == nil || !strings.Contains(err.Error(), "not a declared") {
			t.Fatalf("hidden leaf error = %v", err)
		}
	})

	t.Run("foreign leaf remains foreign", func(t *testing.T) {
		manager := newTestGameManger(t)
		first, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("first game: %v", err)
		}
		second, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatalf("second game: %v", err)
		}
		firstGameState, _ := concreteStates(first.CurrentState())
		secondState := second.CurrentState().(*state)
		secondGameState, _ := concreteStates(secondState)
		secondGameState.MyMergedStack = NewConcatenatedStack(secondGameState.DownSizeStack, firstGameState.DrawDeck)
		if err := secondState.initializeStackOwners(); err == nil {
			t.Fatal("foreign merged leaf unexpectedly accepted")
		}
		if firstGameState.DrawDeck.state() != first.CurrentState() {
			t.Fatal("foreign merged leaf was reassigned")
		}
	})
}

func TestOwnerRegistryIncludesPlayerAndDynamicStacks(t *testing.T) {
	game := testDefaultGame(t, false)
	st := game.CurrentState().(*state)
	var playerOwner, dynamicOwner bool
	for _, owner := range st.stackOwners {
		playerOwner = playerOwner || strings.HasPrefix(owner.path, "Players[")
		dynamicOwner = dynamicOwner || strings.HasPrefix(owner.path, "DynamicComponentValues[")
	}
	if !playerOwner || !dynamicOwner {
		t.Fatalf("owner registry coverage: player=%v dynamic=%v", playerOwner, dynamicOwner)
	}
}

func TestComponentConservationRejectsForgedDeckName(t *testing.T) {
	game := testDefaultGame(t, false)
	gameState, _ := concreteStates(game.CurrentState())
	gameState.DrawDeck.(*growableStack).deckName = "forged"
	if err := game.CurrentState().(*state).validateComponentConservation(); err == nil || !strings.Contains(err.Error(), "unknown deck") {
		t.Fatalf("forged-deck error = %v", err)
	}
}

func TestStateCopyPreservesBoardSpaceIdentity(t *testing.T) {
	game := testDefaultGame(t, false)
	copyState, err := game.CurrentState().(*state).copy(false)
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	gameCopy, _ := concreteStates(copyState)

	for i, space := range gameCopy.MyBoard.Spaces() {
		if got := space.Board(); got != gameCopy.MyBoard {
			t.Fatalf("space %d points at board %p, want copied board %p", i, got, gameCopy.MyBoard)
		}
	}
}

func TestComponentConservationRejectsMissingComponent(t *testing.T) {
	game := testDefaultGame(t, false)
	gameState, _ := concreteStates(game.CurrentState())
	draw := gameState.DrawDeck.(*growableStack)
	draw.indexes = draw.indexes[1:]

	err := game.CurrentState().(*state).validateComponentConservation()
	if err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("missing component error = %v", err)
	}
}

func TestComponentConservationRejectsDuplicateComponent(t *testing.T) {
	game := testDefaultGame(t, false)
	gameState, players := concreteStates(game.CurrentState())
	draw := gameState.DrawDeck.(*growableStack)
	hand := players[0].Hand.(*sizedStack)
	hand.indexes[0] = draw.indexes[0]

	err := game.CurrentState().(*state).validateComponentConservation()
	if err == nil || !strings.Contains(err.Error(), "appears at both") {
		t.Fatalf("duplicate component error = %v", err)
	}
}

func TestComponentConservationRejectsInvalidRawIndex(t *testing.T) {
	game := testDefaultGame(t, false)
	gameState, _ := concreteStates(game.CurrentState())
	draw := gameState.DrawDeck.(*growableStack)
	draw.indexes[0] = -99

	err := game.CurrentState().(*state).validateComponentConservation()
	if err == nil || !strings.Contains(err.Error(), "invalid component index -99") {
		t.Fatalf("invalid component error = %v", err)
	}
}

func TestSanitizedStateSkipsExactConservation(t *testing.T) {
	game := testDefaultGame(t, false)
	sanitized, err := game.CurrentState().SanitizedForPlayer(0)
	if err != nil {
		t.Fatalf("sanitize: %v", err)
	}
	if err := sanitized.(*state).validateComponentConservation(); err != nil {
		t.Fatalf("sanitized conservation should be skipped: %v", err)
	}
}

func TestStateFromRecordRejectsCorruptComponents(t *testing.T) {
	game := testDefaultGame(t, false)
	gameState, players := concreteStates(game.CurrentState())
	draw := gameState.DrawDeck.(*growableStack)
	players[0].Hand.(*sizedStack).indexes[0] = draw.indexes[0]

	_, err := game.Manager().stateFromRecord(game.CurrentState().StorageRecord(), game.Version())
	if err == nil || !strings.Contains(err.Error(), "appears at both") {
		t.Fatalf("corrupt load error = %v", err)
	}
}

func BenchmarkComponentConservation(b *testing.B) {
	manager, err := NewGameManager(defaultTestGameDelegate(0), newTestStorageManager())
	if err != nil {
		b.Fatalf("NewGameManager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		b.Fatalf("NewDefaultGame: %v", err)
	}
	st := game.CurrentState().(*state)
	gameState, _ := concreteStates(st)
	if err := gameState.OtherStack.ExpandSize(10_000); err != nil {
		b.Fatalf("expand sparse sized stack: %v", err)
	}

	b.Run("10000SparseSlots", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := st.validateComponentConservation(); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("StorageRecordComparison", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if record := st.StorageRecord(); len(record) == 0 {
				b.Fatal("empty storage record")
			}
		}
	})
}
