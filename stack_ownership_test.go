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
