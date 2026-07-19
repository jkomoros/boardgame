package boardgame

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"io/ioutil"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jkomoros/boardgame/internal/patchtree"
	"github.com/workfit/tester/assert"
)

func TestProposeMoveAtVersionRejectsInsideSerializedLoop(t *testing.T) {
	game := testDefaultGame(t, true)
	move := game.MoveByName("Draw Card")
	if move == nil {
		t.Fatal("Couldn't find move Draw Card")
	}
	expected := game.Version()
	if err := <-game.ProposeMoveAtVersion(move, AdminPlayerIndex, expected); err != nil {
		t.Fatalf("first version-bound proposal failed: %v", err)
	}
	err := <-game.ProposeMoveAtVersion(game.MoveByName("Draw Card"), AdminPlayerIndex, expected)
	var stale *StaleVersionError
	if !stderrors.As(err, &stale) {
		t.Fatalf("second proposal error = %v; want StaleVersionError", err)
	}
	if stale.Expected != expected || stale.Actual != game.Version() {
		t.Fatalf("stale error = %+v; current version = %d", stale, game.Version())
	}
	if !game.AtProposalFrontier() {
		t.Fatal("stale proposal cleared the settled proposal frontier")
	}
}

func TestIllegalProposalPreservesProposalFrontier(t *testing.T) {
	game := testDefaultGame(t, true)
	version := game.Version()
	move := game.MoveByName("Test").(*testMove)
	move.TargetPlayerIndex = (game.CurrentState().CurrentPlayerIndex() + 1) % PlayerIndex(game.NumPlayers())
	if err := <-game.ProposeMove(move, move.TargetPlayerIndex); err == nil {
		t.Fatal("illegal proposal unexpectedly succeeded")
	}
	if game.Version() != version {
		t.Fatalf("illegal proposal advanced version to %d, want %d", game.Version(), version)
	}
	if !game.AtProposalFrontier() {
		t.Fatal("illegal proposal cleared the settled proposal frontier")
	}
}

func TestMoveByNameForStateUsesPinnedDefaultsInsteadOfCurrentState(t *testing.T) {
	game := testDefaultGame(t, true)
	pinned := game.CurrentState()
	pinnedCurrent := pinned.CurrentPlayerIndex()

	if err := <-game.ProposeMove(game.MoveByName("Test"), pinnedCurrent); err != nil {
		t.Fatalf("make turn-consuming test move: %v", err)
	}
	current := game.CurrentState().CurrentPlayerIndex()
	if current == pinnedCurrent {
		t.Fatalf("test requires current player to advance; remained %d", current)
	}

	pinnedMove, ok := game.MoveByNameForState("Test", pinned).(*testMove)
	if !ok {
		t.Fatal("MoveByNameForState did not return Test move")
	}
	liveMove, ok := game.MoveByName("Test").(*testMove)
	if !ok {
		t.Fatal("MoveByName did not return Test move")
	}
	if pinnedMove.TargetPlayerIndex != pinnedCurrent {
		t.Fatalf("pinned default = %d, want %d", pinnedMove.TargetPlayerIndex, pinnedCurrent)
	}
	if liveMove.TargetPlayerIndex != current {
		t.Fatalf("live default = %d, want %d", liveMove.TargetPlayerIndex, current)
	}
}

func TestMoveByNameForStateRejectsStateFromAnotherGame(t *testing.T) {
	game := testDefaultGame(t, true)
	other, err := game.Manager().NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	if move := game.MoveByNameForState("Test", other.CurrentState()); move != nil {
		t.Fatalf("MoveByNameForState accepted another game's state: %T", move)
	}
}

type proposalFrontierBlockingStorage struct {
	*testStorageManager
	blockVersion int
	saved        chan struct{}
	release      chan struct{}
}

type proposalFrontierAgentFailureStorage struct {
	*testStorageManager
	failAgentState atomic.Bool
}

type proposalFrontierInjectedFailureStorage struct {
	*testStorageManager
	failGameSave atomic.Bool
	failFrontier atomic.Bool
}

func (s *proposalFrontierInjectedFailureStorage) SaveGameAndCurrentState(game *GameStorageRecord, state StateStorageRecord, move *MoveStorageRecord) error {
	if s.failGameSave.Load() {
		return stderrors.New("deliberate game save failure")
	}
	return s.testStorageManager.SaveGameAndCurrentState(game, state, move)
}

func (s *proposalFrontierInjectedFailureStorage) SaveProposalFrontier(gameID string, stateVersion, frontierVersion int) error {
	if s.failFrontier.Load() {
		return stderrors.New("deliberate frontier marker failure")
	}
	return s.testStorageManager.SaveProposalFrontier(gameID, stateVersion, frontierVersion)
}

func TestProposalFrontierRestoredWhenStateSaveRejects(t *testing.T) {
	storage := &proposalFrontierInjectedFailureStorage{testStorageManager: newTestStorageManager()}
	manager, err := NewGameManager(defaultTestGameDelegate(0), storage)
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	version := game.Version()
	storage.failGameSave.Store(true)
	move := game.MoveByName("Test")
	if err := <-game.ProposeMove(move, game.CurrentState().CurrentPlayerIndex()); err == nil || !strings.Contains(err.Error(), "game save failure") {
		t.Fatalf("proposal error = %v; want deliberate game save failure", err)
	}
	if game.Version() != version {
		t.Fatalf("failed save left in-memory version %d, want %d", game.Version(), version)
	}
	if !game.AtProposalFrontier() {
		t.Fatal("failed save did not restore the settled in-memory frontier")
	}
	stored, err := storage.Game(game.ID())
	if err != nil {
		t.Fatal(err)
	}
	if stored.Version != version || !stored.ProposalFrontierKnown || stored.ProposalFrontierVersion != version {
		t.Fatalf("failed save changed durable frontier: %+v", stored)
	}

	storage.failGameSave.Store(false)
	if err := <-game.ProposeMove(game.MoveByName("Test"), game.CurrentState().CurrentPlayerIndex()); err != nil {
		t.Fatalf("proposal after restored save failure: %v", err)
	}
}

func TestProposalFrontierMarkerFailureDoesNotFailCommittedMove(t *testing.T) {
	for _, tc := range []struct {
		name     string
		finished bool
	}{
		{name: "normal"},
		{name: "finished", finished: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			storage := &proposalFrontierInjectedFailureStorage{testStorageManager: newTestStorageManager()}
			manager, err := NewGameManager(defaultTestGameDelegate(0), storage)
			if err != nil {
				t.Fatal(err)
			}
			var logs bytes.Buffer
			manager.Logger().SetOutput(&logs)
			game, err := manager.NewDefaultGame()
			if err != nil {
				t.Fatal(err)
			}
			version := game.Version()
			move := game.MoveByName("Test").(*testMove)
			if tc.finished {
				move.ScoreIncrement = 6
			}
			storage.failFrontier.Store(true)
			if err := <-game.ProposeMove(move, game.CurrentState().CurrentPlayerIndex()); err != nil {
				t.Fatalf("committed proposal reported marker failure: %v", err)
			}
			committedVersion := game.Version()
			if committedVersion <= version {
				t.Fatalf("committed version = %d, want greater than %d", committedVersion, version)
			}
			if game.Finished() != tc.finished {
				t.Fatalf("finished = %v, want %v", game.Finished(), tc.finished)
			}
			if game.AtProposalFrontier() {
				t.Fatal("failed marker write was incorrectly certified in memory")
			}
			stored, err := storage.Game(game.ID())
			if err != nil {
				t.Fatal(err)
			}
			if stored.Version != committedVersion || stored.ProposalFrontierKnown {
				t.Fatalf("durable committed head/frontier = %+v", stored)
			}
			if !strings.Contains(logs.String(), "Could not persist proposal frontier after committed move") || !strings.Contains(logs.String(), "deliberate frontier marker failure") {
				t.Fatalf("marker failure was not logged: %q", logs.String())
			}

			storage.failFrontier.Store(false)
			if err := <-manager.Internals().ForceFixUp(game); err != nil {
				t.Fatalf("recover marker after transient failure: %v", err)
			}
			if !game.AtProposalFrontier() {
				t.Fatal("recovery did not certify the committed head")
			}
		})
	}
}

func (s *proposalFrontierAgentFailureStorage) AgentState(gameID string, player PlayerIndex) ([]byte, error) {
	if s.failAgentState.Load() {
		return nil, stderrors.New("deliberate agent-state failure")
	}
	return s.testStorageManager.AgentState(gameID, player)
}

func TestProposalFrontierSurvivesAgentFailureAfterTerminalFixUp(t *testing.T) {
	storage := &proposalFrontierAgentFailureStorage{testStorageManager: newTestStorageManager()}
	fixups := make(map[int]string)
	delegate := newProposalFrontierScriptDelegate(fixups)
	manager, err := NewGameManager(delegate, storage)
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewGame(3, nil, []string{"Test", "", ""})
	if err != nil {
		t.Fatal(err)
	}
	// The initiating player move commits version 1; require one recursive fix-up
	// before the terminal check so this pins the deepest-frame regression.
	fixups[1] = "Legal Memo Test Move"
	storage.failAgentState.Store(true)
	err = <-game.ProposeMove(game.MoveByName("Legal Memo Test Move"), 0)
	if err == nil || !strings.Contains(err.Error(), "agent") {
		t.Fatalf("proposal error = %v; want agent failure", err)
	}
	if !game.AtProposalFrontier() {
		t.Fatal("agent failure erased completed recursive fix-up frontier")
	}
	restarted, err := NewGameManager(newProposalFrontierScriptDelegate(fixups), storage)
	if err != nil {
		t.Fatal(err)
	}
	if frontier, ok := restarted.ProposalFrontierVersion(game.ID()); !ok || frontier != game.Version() {
		t.Fatalf("reloaded frontier after agent failure = %d, %v; want %d, true", frontier, ok, game.Version())
	}
}

// frontierUnsupportedStorage deliberately hides the optional durable marker
// capability exposed by its underlying test storage.
type frontierUnsupportedStorage struct{ StorageManager }

func TestProposalFrontierLegacyStorageReloadFailsClosed(t *testing.T) {
	underlying := newTestStorageManager()
	storage := &frontierUnsupportedStorage{StorageManager: underlying}
	manager, err := NewGameManager(defaultTestGameDelegate(0), storage)
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	if !game.AtProposalFrontier() {
		t.Fatal("active game should retain an in-memory frontier")
	}
	restarted, err := NewGameManager(defaultTestGameDelegate(0), storage)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := restarted.ProposalFrontierVersion(game.ID()); ok {
		t.Fatal("legacy storage record without durable evidence was certified")
	}
}

type proposalFrontierExternalDelegate struct {
	testGameDelegate
	block   atomic.Bool
	entered chan struct{}
	release chan struct{}
}

func newProposalFrontierExternalDelegate() *proposalFrontierExternalDelegate {
	d := &proposalFrontierExternalDelegate{
		entered: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	d.moveInstaller = defaultTestGameDelegate(0).moveInstaller
	return d
}

func (d *proposalFrontierExternalDelegate) ProposeFixUpMove(ImmutableState) Move {
	if d.block.Load() {
		d.entered <- struct{}{}
		<-d.release
	}
	return nil
}

func TestForceFixUpInvalidatesDurableFrontierUntilTerminalCheck(t *testing.T) {
	storage := newTestStorageManager()
	delegate := newProposalFrontierExternalDelegate()
	manager, err := NewGameManager(delegate, storage)
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	delegate.block.Store(true)
	delayed := manager.Internals().ForceFixUp(game)
	<-delegate.entered
	if game.AtProposalFrontier() {
		t.Fatal("active game retained frontier while external fix-up check was queued")
	}
	record, err := storage.Game(game.ID())
	if err != nil {
		t.Fatal(err)
	}
	if record.ProposalFrontierKnown {
		t.Fatal("durable frontier remained known during external fix-up check")
	}
	close(delegate.release)
	if err := <-delayed; err != nil {
		t.Fatal(err)
	}
	if !game.AtProposalFrontier() {
		t.Fatal("terminal external fix-up check did not restore frontier")
	}
	restarted, err := NewGameManager(newProposalFrontierExternalDelegate(), storage)
	if err != nil {
		t.Fatal(err)
	}
	if frontier, ok := restarted.ProposalFrontierVersion(game.ID()); !ok || frontier != game.Version() {
		t.Fatalf("reloaded external frontier = %d, %v; want %d, true", frontier, ok, game.Version())
	}
}

func (s *proposalFrontierBlockingStorage) SaveGameAndCurrentState(game *GameStorageRecord, state StateStorageRecord, move *MoveStorageRecord) error {
	if err := s.testStorageManager.SaveGameAndCurrentState(game, state, move); err != nil {
		return err
	}
	if move != nil && move.Version == s.blockVersion {
		s.saved <- struct{}{}
		<-s.release
	}
	return nil
}

func TestProposalFrontierDoesNotAdvanceWithIntermediateDurableState(t *testing.T) {
	storage := &proposalFrontierBlockingStorage{
		testStorageManager: newTestStorageManager(),
		saved:              make(chan struct{}, 1),
		release:            make(chan struct{}),
	}
	manager, err := NewGameManager(defaultTestGameDelegate(0), storage)
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.newGameImpl("FRONTIERTEST", "FRONTIERSALT")
	if err != nil {
		t.Fatal(err)
	}
	if err := game.setUp(0, nil, nil); err != nil {
		t.Fatal(err)
	}
	settled := game.Version()
	if !game.AtProposalFrontier() {
		t.Fatalf("setup version %d was not marked as a proposal frontier", settled)
	}

	storage.blockVersion = settled + 1
	delayed := game.ProposeMove(game.MoveByName("Test"), game.CurrentState().CurrentPlayerIndex())
	<-storage.saved

	// Save completed before the hook blocks, so a fresh read can observe the
	// newer durable version. The active serialized loop has not completed the
	// proposal/fix-up chain and remains the manager's frontier authority.
	stored := manager.Game(game.ID())
	if stored == nil || stored.Version() != settled+1 {
		t.Fatalf("stored intermediate version = %v, want %d", stored, settled+1)
	}
	frontier, ok := manager.ProposalFrontierVersion(game.ID())
	if ok {
		t.Fatalf("manager advertised stale frontier %d during active version %d", frontier, stored.Version())
	}

	close(storage.release)
	if err := <-delayed; err != nil {
		t.Fatal(err)
	}
	frontier, ok = manager.ProposalFrontierVersion(game.ID())
	if !ok || frontier != game.Version() || !game.AtProposalFrontier() {
		t.Fatalf("completed frontier = %d, %v; game version %d", frontier, ok, game.Version())
	}
}

type proposalFrontierFailingMove struct {
	baseMove
}

var proposalFrontierFailingMoveConfig = NewMoveConfig(
	"Proposal Frontier Failing Move",
	func() Move { return new(proposalFrontierFailingMove) },
	nil,
)

func (m *proposalFrontierFailingMove) Reader() PropertyReader { return getDefaultReader(m) }
func (m *proposalFrontierFailingMove) ReadSetter() PropertyReadSetter {
	return getDefaultReadSetter(m)
}
func (m *proposalFrontierFailingMove) ReadSetConfigurer() PropertyReadSetConfigurer {
	return getDefaultReadSetConfigurer(m)
}
func (m *proposalFrontierFailingMove) Legal(ImmutableState, PlayerIndex) error { return nil }
func (m *proposalFrontierFailingMove) Apply(State) error {
	return stderrors.New("deliberate later fix-up failure")
}

// proposalFrontierScriptDelegate makes recovery deterministic: each durable
// version either names the next pending fix-up, or is a settled head.
type proposalFrontierScriptDelegate struct {
	testGameDelegate
	fixupAtVersion map[int]string
}

func newProposalFrontierScriptDelegate(fixups map[int]string) *proposalFrontierScriptDelegate {
	d := &proposalFrontierScriptDelegate{fixupAtVersion: fixups}
	d.moveInstaller = func(*GameManager) []MoveConfig {
		return []MoveConfig{legalMemoTestMoveConfig, proposalFrontierFailingMoveConfig}
	}
	return d
}

func (d *proposalFrontierScriptDelegate) ProposeFixUpMove(state ImmutableState) Move {
	name := d.fixupAtVersion[state.Version()]
	if name == "" {
		return nil
	}
	return state.Game().MoveByNameForState(name, state)
}

func TestProposalFrontierSetupIncludesAllFixUps(t *testing.T) {
	storage := newTestStorageManager()
	fixups := map[int]string{
		0: "Legal Memo Test Move",
		1: "Legal Memo Test Move",
	}
	delegate := newProposalFrontierScriptDelegate(fixups)
	manager, err := NewGameManager(delegate, storage)
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.newGameImpl("SETUPFRONTIER", "SETUPFRONTIERSALT")
	if err != nil {
		t.Fatal(err)
	}
	if err := game.setUp(0, nil, nil); err != nil {
		t.Fatal(err)
	}
	if game.Version() != 2 || !game.AtProposalFrontier() {
		t.Fatalf("setup frontier = version %d, frontier %d; want 2, 2", game.Version(), game.ProposalFrontierVersion())
	}
	restarted, err := NewGameManager(newProposalFrontierScriptDelegate(fixups), storage)
	if err != nil {
		t.Fatal(err)
	}
	if frontier, ok := restarted.ProposalFrontierVersion(game.ID()); !ok || frontier != 2 {
		t.Fatalf("recovered completed setup frontier = %d, %v; want 2, true", frontier, ok)
	}
}

func TestProposalFrontierFailedLaterFixUpRemainsUnsettledAcrossManager(t *testing.T) {
	storage := newTestStorageManager()
	fixups := map[int]string{
		1: "Legal Memo Test Move",
		2: "Proposal Frontier Failing Move",
	}
	manager, err := NewGameManager(newProposalFrontierScriptDelegate(fixups), storage)
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.newGameImpl("FAILEDCHAIN", "FAILEDCHAINSALT")
	if err != nil {
		t.Fatal(err)
	}
	if err := game.setUp(0, nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := <-game.ProposeMove(game.MoveByName("Legal Memo Test Move"), AdminPlayerIndex); err == nil || !strings.Contains(err.Error(), "deliberate later fix-up failure") {
		t.Fatalf("proposal error = %v; want deliberate later fix-up failure", err)
	}
	if game.Version() != 2 {
		t.Fatalf("durable partial chain ended at version %d; want 2", game.Version())
	}
	if _, ok := manager.ProposalFrontierVersion(game.ID()); ok {
		t.Fatal("active failed partial chain was certified settled")
	}

	// A second manager has no process-local knowledge of the first manager's
	// chain. It must rediscover the pending failed fix-up from durable state.
	restarted, err := NewGameManager(newProposalFrontierScriptDelegate(fixups), storage)
	if err != nil {
		t.Fatal(err)
	}
	reloaded := restarted.Game(game.ID())
	if reloaded == nil || reloaded.Version() != 2 {
		t.Fatalf("reloaded partial game = %#v", reloaded)
	}
	if reloaded.AtProposalFrontier() {
		t.Fatal("reloaded partial chain was certified settled")
	}
	if _, ok := restarted.ProposalFrontierVersion(game.ID()); ok {
		t.Fatal("second manager advertised actions for a pending durable fix-up")
	}
}

type testInfiniteLoopGameDelegate struct {
	testGameDelegate
}

func (t *testInfiniteLoopGameDelegate) ProposeFixUpMove(state ImmutableState) Move {
	return state.Game().MoveByName("Test Always Legal Move")
}

func TestGameDelegateConstants(t *testing.T) {
	game := testDefaultGame(t, false)

	assert.For(t).ThatActual(game.Manager().Chest().ConstantNames()).Equals([]string{
		"ConstantStackSize",
		"MyBool",
	})
}

func TestComputedGroupNames(t *testing.T) {
	manager := newTestGameManger(t)
	names := manager.playerValidator.sanitizationPolicyGroupNames(manager.delegate.GroupEnum())
	assert.For(t).ThatActual(names).Equals(map[string]bool{"other": true})
}

func TestIllegalPhase(t *testing.T) {
	game := testDefaultGame(t, false)

	m := game.MoveByName("Make Illegal Phase")

	assert.For(t).ThatActual(m).IsNotNil()

	err := <-game.ProposeMove(m, AdminPlayerIndex)

	//Ensure that the move was rejected because of tree enum, not for example
	//not being a Legal move.
	rejectedBecausePhase := strings.Contains(err.Error(), "TreeEnum")

	assert.For(t).ThatActual(rejectedBecausePhase).IsTrue()
}

func TestGameScorer(t *testing.T) {
	d := &defaultGameDelegate{}
	p := &testPlayerState{
		Score: 10,
	}
	result := d.PlayerScore(p)
	assert.For(t).ThatActual(result).Equals(10)
}

func TestMoveModifyDynamicValues(t *testing.T) {
	game := testDefaultGame(t, true)

	drawCardMove := game.MoveByName("Draw Card")

	if drawCardMove == nil {
		t.Fatal("Couldn't find move draw card")
	}

	if err := <-game.ProposeMove(drawCardMove, ObserverPlayerIndex); err == nil {
		t.Error("Expected error proposing move from ObserverPlayerIndex")
	}

	if err := <-game.ProposeMove(drawCardMove, PlayerIndex(5)); err == nil {
		t.Error("Expected error proposing move with invalid Proposer")
	}

	if err := <-game.ProposeMove(drawCardMove, AdminPlayerIndex); err != nil {
		t.Error("Unexpected error trying to draw card: " + err.Error())
	}

	move := game.MoveByName("Increment IntValue of Card in Hand")

	if move == nil {
		t.Fatal("Couldn't find move Increment IntValue of Card in Hand")
	}

	if err := <-game.ProposeMove(move, AdminPlayerIndex); err != nil {
		t.Error("Unexpected error trying to increment dynamic component state: " + err.Error())
	}

	//Apply the move again. This implicitly tests that deserializing a non-zero dynamic component value works.

	if err := <-game.ProposeMove(move, AdminPlayerIndex); err != nil {
		t.Error("unexpected error trying to increment dynamic component state a second time: ", err.Error())
	}

	gameState, playerStates := concreteStates(game.CurrentState())

	player := playerStates[gameState.CurrentPlayer]

	component := player.Hand.ComponentAt(0)

	dynamic := component.DynamicValues()

	if dynamic == nil {
		t.Error("Component unexpectedly had nil dynamic values")
	}

	easyDynamic := dynamic.(*testingComponentDynamic)

	if easyDynamic.IntVar != 7 {
		t.Error("Dynamic state of component unexpected value: ", easyDynamic.IntVar)
	}

	//Test that SetContainingComponent was set.
	assert.For(t).ThatActual(easyDynamic.ContainingComponent()).Equals(component.ptr())

	var stateNil *state

	assert.For(t).ThatActual(easyDynamic.Stack.state()).DoesNotEqual(stateNil)

	currentJSON, _ := json.MarshalIndent(game.CurrentState(), "", "\t")

	compareJSONObjects(currentJSON, patchtree.MustJSON("./testdata/after_dynamic_component_move"), "Comparing json after two dynamic moves", t)

}

func TestProposeMoveNonModifiableGame(t *testing.T) {
	game := testDefaultGame(t, false)

	manager := game.Manager()

	id := game.ID()

	//At this point, the game has stored state in storage.

	refriedGame := manager.Game(id)

	if refriedGame == nil {
		t.Fatal("Couldn't get a game out refried")
	}

	rawMove := game.MoveByName("test")

	move := rawMove.(*testMove)

	move.AString = "foo"
	move.ScoreIncrement = 3
	move.TargetPlayerIndex = 0
	move.ABool = true

	if err := <-refriedGame.ProposeMove(move, AdminPlayerIndex); err != nil {
		t.Error("Propose move on refried game failed:", err)
	}

	//No refresh necessary becuase we should have refreshed it automatically.

	if refriedGame.Version() != 2 {
		t.Error("The proposed move didn't actually modify the underlying game in storage: ", refriedGame.Version())
	}

}

func TestGameSetUp(t *testing.T) {

	//TODO: really now this test should be repeated calls to NewGame() on manager

	manager := newTestGameManger(t)

	game, err := manager.newGameImpl("", "")

	assert.For(t).ThatActual(err).IsNil()

	id := game.ID()

	if len(id) != gameIDLength {
		t.Error("Game didn't have an ID of correct length. Wanted", gameIDLength, "got", id)
	}

	if game.Moves() != nil {
		t.Error("Got moves back before SetUp was called")
	}

	if game.MoveByName("Test") != nil {
		t.Error("Move by name returned a move before SetUp was called")
	}

	move := &testMove{
		AString:           "foo",
		ScoreIncrement:    3,
		TargetPlayerIndex: 0,
		ABool:             true,
	}

	originalTestMove := move

	delayedError := game.ProposeMove(move, AdminPlayerIndex)

	select {
	case <-delayedError:
		//Good!
	case <-time.After(time.Millisecond * 5):
		t.Error("We never got an error from proposing a move on a game that hadn't even started")
	}

	if err := game.setUp(15, nil, nil); err == nil {
		t.Error("Calling set up with an illegal number of players didn't fail")
	}

	if err := game.setUp(2, Variant{"color": "illegal"}, nil); err == nil {
		t.Error("Calling game set up with an illegal config did not fail")
	}

	if err := game.setUp(-5, nil, nil); err == nil {
		t.Error("Calling set up with negative number of players didn't fail")
	}

	if err := game.setUp(3, nil, []string{"", "bam"}); err == nil {
		t.Error("Calling set up with wrong-sized agent config didn't fail")
	}

	//TODO: we no longer test that SetUp calls the Component distribution logic.

	//Blue is a legal color according to our delegate's Configs()
	if err := game.setUp(0, map[string]string{"color": "blue"}, nil); err != nil {
		t.Error("Calling SetUp on a previously errored game did not succeed", err)
	}

	if wrapper, err := game.Manager().Storage().State(game.ID(), 0); wrapper == nil {
		t.Error("State 0 was not saved in storage when game set up")
	} else if err != nil {
		t.Error("Storing state 0 failed: " + err.Error())
	}

	if game.CurrentState() == nil {
		t.Error("Game had no current state after saving")
	}

	if game.MoveByName("Test") == nil {
		t.Error("MoveByName didn't return a valid move when provided the proper name after calling setup")
	}

	if game.MoveByName("test") == nil {
		t.Error("MoveByName didn't return a valid move when provided with a lowercase name after calling SetUp.")
	}

	if originalTestMove == game.MoveByName("Test") {
		t.Error("MoveByName returned a non-copy")
	}

	//Test to verify that game has stack's state property set
	currentState := game.CurrentState()
	gameState, playerStates := concreteStates(currentState)

	if gameState.DrawDeck.state() != currentState {
		t.Error("GameState's drawdeck didn't have state set correctly. Got", gameState.DrawDeck.state(), "wanted", currentState)
	}

	if playerStates[0].Hand.state() != currentState {
		t.Error("PlayerStates Hand didn't have state set correctly. Got", playerStates[0].Hand.state(), "wanted", currentState)
	}

	deck := game.Manager().Chest().Deck("test").Components()

	//We put one of the components in MyBoard[1]
	if gameState.DrawDeck.Len() != len(deck)-1 {
		t.Error("All of the components were not distributed in SetUp")
	}

	stateCopy, err := currentState.Copy(false)

	assert.For(t).ThatActual(err).IsNil()

	gameState, playerStates = concreteStates(stateCopy)

	if gameState.DrawDeck.state() != stateCopy {
		t.Error("The copy of state's stacks had the old state in gamestate")
	}

	if playerStates[0].Hand.state() != stateCopy {
		t.Error("The copy of state's stacks had the old state in playerstate")
	}

}

func TestApplyMove(t *testing.T) {
	game := testDefaultGame(t, true)

	rawMove := game.MoveByName("test")

	move := rawMove.(*testMove)

	move.AString = "foo"
	move.ScoreIncrement = 3
	move.TargetPlayerIndex = 0
	move.ABool = true

	manager := game.Manager()

	oldMoves := manager.moves
	oldMovesByName := manager.movesByName

	manager.moves = nil
	manager.movesByName = make(map[string]*moveType)

	if err := <-game.ProposeMove(move, AdminPlayerIndex); err == nil {
		t.Error("Game allowed a move that wasn't configured as part of game to be applied")
	}

	manager.moves = oldMoves
	manager.movesByName = oldMovesByName

	//testMove checks to make sure game.state.currentPlayerIndex is targetplayerindex

	move.TargetPlayerIndex = 1

	if err := <-game.ProposeMove(move, AdminPlayerIndex); err == nil {
		t.Error("Game allowed a move to be applied where the wrong playe was current")
	}

	move.TargetPlayerIndex = 0

	if err := <-game.ProposeMove(move, AdminPlayerIndex); err != nil {
		t.Error("Game didn't allow a legal move to be made")
	}

	//Verify that the move was made. Note that because our Delegate has a
	//FixUp move, this is also testing that not just the main move, but also
	//the fixup move was made.

	record, err := game.Manager().Storage().State(game.ID(), game.Version())

	if err != nil {
		t.Error("Unexpected error", err)
	}

	wrapper, err := game.Manager().stateFromRecord(record, game.Version())

	if err != nil {
		t.Error("Error state from from blob", err)
	}

	wrapper.game = game

	currentJSON, _ := json.Marshal(wrapper)

	compareJSONObjects(currentJSON, patchtree.MustJSON("./testdata/after_move"), "Basic state after test move", t)

	//Apply a move that should finish the game (any player has score > 5)
	newRawMove := game.MoveByName("test")

	newMove := newRawMove.(*testMove)

	newMove.AString = "foo"
	newMove.ScoreIncrement = 6
	newMove.TargetPlayerIndex = 1
	newMove.ABool = true

	if err := <-game.ProposeMove(newMove, AdminPlayerIndex); err != nil {
		t.Error("Game didn't allow a move to be made even though it was legal: ", err)
	}

	if wrapper, _ := game.Manager().Storage().State(game.ID(), 1); wrapper == nil {
		t.Error("We didn't get back state for state 1; game must not be persisting states to DB.")
	}

	//By the time err has resolved above, any fixup moves have been applied.

	if !game.Finished() {
		t.Error("Game didn't notice that a user had won")
	}

	if !reflect.DeepEqual(game.Winners(), []PlayerIndex{PlayerIndex(1)}) {
		t.Error("Game thought the wrong players had won")
	}

	moveAfterFinished := &testMove{
		AString:           "foo",
		ScoreIncrement:    3,
		TargetPlayerIndex: 2,
		ABool:             true,
	}

	if err := <-game.ProposeMove(moveAfterFinished, AdminPlayerIndex); err == nil {
		t.Error("Game allowed a move to be applied after the game was finished")
	}
}

func TestMoveRoundTrip(t *testing.T) {
	game := testDefaultGame(t, false)

	move := game.MoveByName("test")

	testMove := move.(*testMove)

	testMove.AString = "foo"
	testMove.ScoreIncrement = 3
	testMove.TargetPlayerIndex = 0
	testMove.ABool = true

	err := <-game.ProposeMove(move, AdminPlayerIndex)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(testMove.Info().Timestamp().IsZero()).IsFalse()

	refriedMove, err := game.Move(1)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(move).Equals(refriedMove)

	assert.For(t).ThatActual(refriedMove.Info().Timestamp()).Equals(move.Info().Timestamp())

	fixUpMove, err := game.Move(2)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(fixUpMove.Info().Initiator()).Equals(1)

}

func TestIllegalMove(t *testing.T) {

	manager := newTestGameManger(t)

	_, err := newMoveType(testIllegalMoveConfig, manager)

	assert.For(t).ThatActual(err).IsNotNil()

}

func TestInfiniteProposeFixUp(t *testing.T) {
	//This test makes sure that if our GameDelegate is going to always return
	//moves that are legal, we'll bail at a certain point.

	moveInstaller := func(manager *GameManager) []MoveConfig {
		return []MoveConfig{
			testMoveConfig,
			testAlwaysLegalMoveConfig,
		}
	}

	delegate := &testInfiniteLoopGameDelegate{
		testGameDelegate{
			moveInstaller: moveInstaller,
		},
	}

	manager, err := NewGameManager(delegate, newTestStorageManager())

	assert.For(t).ThatActual(err).IsNil()

	_, err = manager.NewDefaultGame()

	assert.For(t).ThatActual(err).Equals(ErrTooManyFixUps)

}

func TestIllegalPlayerIndex(t *testing.T) {
	game := testDefaultGame(t, false)

	previousVersion := game.Version()

	move := game.MoveByName("Invalid PlayerIndex")

	assert.For(t).ThatActual(move).IsNotNil()

	move.(*testMoveInvalidPlayerIndex).CurrentlyLegal = true

	err := <-game.ProposeMove(move, AdminPlayerIndex)

	assert.For(t).ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(game.Version()).Equals(previousVersion)

}

func TestAgent(t *testing.T) {

	manager := newTestGameManger(t)

	game, err := manager.newGameImpl("", "")

	assert.For(t).ThatActual(err).IsNil()

	game.instantAgentMoves = true

	assert.For(t).ThatActual(game.NumAgentPlayers()).Equals(0)

	err = game.setUp(3, nil, []string{"", "Test", "Test"})

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(game.Version()).Equals(0)

	assert.For(t).ThatActual(game.NumAgentPlayers()).Equals(2)

	move := game.MoveByName("Test")

	assert.For(t).ThatActual(move).IsNotNil()

	err = <-game.ProposeMove(move, 0)

	assert.For(t).ThatActual(err).IsNil()

	//After we make that move, the next two players will make moves and it
	//will advance back to main player.

	<-time.After(time.Millisecond * 50)

	assert.For(t).ThatActual(game.Version()).Equals(6)

	gameState, _ := concreteStates(game.CurrentState())

	assert.For(t).ThatActual(gameState.CurrentPlayer).Equals(PlayerIndex(0))

}

func TestGameSalt(t *testing.T) {
	game := testDefaultGame(t, false)

	assert.For(t).ThatActual(game.secretSalt).DoesNotEqual("")

	refriedGame := game.Manager().Game(game.ID())

	if !assert.For(t).ThatActual(refriedGame).IsNotNil().Passed() {
		t.FailNow()
	}

	assert.For(t).ThatActual(game.secretSalt).Equals(refriedGame.secretSalt)

	mainC := game.Manager().Chest().Deck("test").ComponentAt(0).ImmutableInstance(game.CurrentState())
	refriedC := refriedGame.Manager().Chest().Deck("test").ComponentAt(0).ImmutableInstance(refriedGame.CurrentState())

	mainCId := mainC.ID()

	assert.For(t).ThatActual(mainCId).DoesNotEqual("")
	assert.For(t).ThatActual(mainCId).Equals(refriedC.ID())

	otherGame := testDefaultGame(t, false)

	otherC := otherGame.Manager().Chest().Deck("test").ComponentAt(0).ImmutableInstance(otherGame.CurrentState())

	assert.For(t).ThatActual(mainCId).DoesNotEqual(otherC.ID())
}

func goldenGameBlob() []byte {
	gameBlob, err := ioutil.ReadFile("testdata/game_blob.json")
	if err != nil {
		return nil
	}
	baseState, err := ioutil.ReadFile("testdata/base.json")
	if err != nil {
		return nil
	}

	var gameBlobJSON map[string]interface{}
	var baseStateJSON map[string]interface{}

	if err := json.Unmarshal(gameBlob, &gameBlobJSON); err != nil {
		return nil
	}

	if err := json.Unmarshal(baseState, &baseStateJSON); err != nil {
		return nil
	}

	gameBlobJSON["CurrentState"] = baseStateJSON

	blob, err := json.Marshal(gameBlobJSON)

	if err != nil {
		return nil
	}

	return blob

}

func TestGameState(t *testing.T) {
	game := testDefaultGame(t, true)

	if game.Name() != testGameName {
		t.Error("Game name was not correct")
	}

	blob, err := json.MarshalIndent(game, "", "  ")

	if err != nil {
		t.Error("Json marshal of game failed:", err)
	}

	goldenBlob := goldenGameBlob()

	if err != nil {
		t.Error("Couldn't load golden file", err)
	}

	compareJSONObjects(blob, goldenBlob, "Sanity checking game json", t)

	//Getting this now helps verify that we invalidate currentState cache when
	//we apply a move.
	state := game.CurrentState()

	state0 := game.State(0)

	//This is lame, but when you create json for a State, it touches Computed,
	//which will make it non-nil, so if you're doing direct comparison they
	//won't compare equal even though they basically are. At this point
	//CurrentState has already been touched above by creating a json blob. So
	//just touch state0, too. ¯\_(ツ)_/¯
	_, _ = json.Marshal(state0)

	assertPersistedStatesEqual(t, state, state0)

	move := game.MoveByName("Test")

	if move == nil {
		t.Fatal("Couldn't find a move to make")
	}

	if err := <-game.ProposeMove(move, AdminPlayerIndex); err != nil {
		t.Error("Couldn't make move")
	}

	state = game.State(-1)

	if state != nil {
		t.Error("Returned a state for a non-sensiscal version -1", state)
	}

	state = game.State(game.Version() + 1)

	if state != nil {
		t.Error("Returned a state for a too-high version", state)
	}

	currentState := game.CurrentState()
	state = game.State(game.Version())

	assertPersistedStatesEqual(t, currentState, state)

}

func TestGameFreeze(t *testing.T) {
	game := testDefaultGame(t, false)

	assert.For(t).ThatActual(game.Frozen()).IsFalse()

	// Freeze the game
	game.manager.freezeGame(game)

	// Give mainLoop time to exit after done is closed
	<-time.After(10 * time.Millisecond)

	assert.For(t).ThatActual(game.Frozen()).IsTrue()

	// Proposing a move on a frozen modifiable game should return an error
	move := game.MoveByName("Test")
	assert.For(t).ThatActual(move).IsNotNil()

	err := <-game.ProposeMove(move, AdminPlayerIndex)
	assert.For(t).ThatActual(err).IsNotNil()
	assert.For(t).ThatActual(strings.Contains(err.Error(), "frozen")).IsTrue()

	// Calling markFrozen again should not panic (sync.Once protects double-close)
	game.markFrozen()
	assert.For(t).ThatActual(game.Frozen()).IsTrue()
}

func TestFrozenGameReload(t *testing.T) {
	manager := newTestGameManger(t)

	game, err := manager.NewDefaultGame()
	assert.For(t).ThatActual(err).IsNil()

	gameID := game.ID()
	originalVersion := game.Version()

	// Freeze the game
	manager.freezeGame(game)
	assert.For(t).ThatActual(game.Frozen()).IsTrue()

	// The frozen game should no longer be in modifiableGames
	manager.modifiableGamesLock.RLock()
	_, ok := manager.modifiableGames[strings.ToUpper(gameID)]
	manager.modifiableGamesLock.RUnlock()
	assert.For(t).ThatActual(ok).IsFalse()

	// ModifiableGame should transparently reload the game from storage
	reloadedGame := manager.ModifiableGame(gameID)
	assert.For(t).ThatActual(reloadedGame).IsNotNil()
	assert.For(t).ThatActual(reloadedGame.Frozen()).IsFalse()
	assert.For(t).ThatActual(reloadedGame.Version()).Equals(originalVersion)

	// The reloaded game should be a different object from the frozen one
	assert.For(t).ThatActual(reloadedGame != game).IsTrue()

	// The reloaded game should be in the cache
	manager.modifiableGamesLock.RLock()
	cachedGame := manager.modifiableGames[strings.ToUpper(gameID)]
	manager.modifiableGamesLock.RUnlock()
	assert.For(t).ThatActual(cachedGame).Equals(reloadedGame)

	// Should be able to propose moves on the reloaded game
	move := reloadedGame.MoveByName("Test")
	assert.For(t).ThatActual(move).IsNotNil()

	moveErr := <-reloadedGame.ProposeMove(move, AdminPlayerIndex)
	assert.For(t).ThatActual(moveErr).IsNil()
	// Version increases by more than 1 because fix-up moves are also applied.
	assert.For(t).ThatActual(reloadedGame.Version() > originalVersion).IsTrue()
}

func TestGameCacheEviction(t *testing.T) {
	manager := newTestGameManger(t)

	// Create enough games to fill the cache
	games := make([]*Game, maxResidentGames+1)
	for i := 0; i < maxResidentGames+1; i++ {
		game, err := manager.NewDefaultGame()
		assert.For(t).ThatActual(err).IsNil()
		games[i] = game

		// Space out lastActivity so the first game is the oldest
		games[i].lastActivity = time.Now().Add(time.Duration(i) * time.Millisecond)
	}

	// The cache should have at most maxResidentGames entries
	manager.modifiableGamesLock.RLock()
	cacheLen := len(manager.modifiableGames)
	manager.modifiableGamesLock.RUnlock()
	assert.For(t).ThatActual(cacheLen <= maxResidentGames).IsTrue()

	// The first game (oldest lastActivity) should have been evicted
	assert.For(t).ThatActual(games[0].Frozen()).IsTrue()
}
