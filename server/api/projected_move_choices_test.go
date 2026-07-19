package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
	"github.com/jkomoros/boardgame/moves/choice"
)

type projectedChoicesDelegate struct {
	*legalLedgerDelegate
	rejectAll         bool
	projectionCount   int
	namePolicy        string
	moveName          string
	legalCalls        atomic.Int32
	fixupUntilVersion atomic.Int32
}

func newProjectedChoicesDelegate() *projectedChoicesDelegate {
	return &projectedChoicesDelegate{
		legalLedgerDelegate: &legalLedgerDelegate{},
		projectionCount:     1,
	}
}

func (d *projectedChoicesDelegate) ConfigureMoves() []boardgame.MoveConfig {
	auto := moves.NewAutoConfigurer(d)
	result := make([]boardgame.MoveConfig, 0, d.projectionCount)
	for i := 0; i < d.projectionCount; i++ {
		name := "Choose Player"
		if d.moveName != "" {
			name = d.moveName
		}
		if d.projectionCount > 1 {
			name = fmt.Sprintf("Choose Player %d", i)
		}
		options := []moves.CustomConfigurationOption{
			moves.WithMoveName(name),
			moves.WithMoveInputFieldOverride("TargetPlayerIndex", boardgame.MoveInputRequired),
			moves.WithChoiceProjection(choice.PlayerIndexes("TargetPlayerIndex").
				DiscloseExactAvailabilityToActor("player identities and target legality are intentionally visible to the acting player")),
		}
		if d.namePolicy != "" {
			options = append(options, moves.WithMoveNameSanitization(d.namePolicy))
		}
		result = append(result, auto.MustConfig(new(projectedChoicesMove), options...))
	}
	result = append(result, auto.MustConfig(new(moves.NoOp), moves.WithMoveName("Recovery FixUp")))
	return result
}

func (d *projectedChoicesDelegate) ProposeFixUpMove(state boardgame.ImmutableState) boardgame.Move {
	if int32(state.Version()) >= d.fixupUntilVersion.Load() {
		return nil
	}
	return state.Game().MoveByNameForState("Recovery FixUp", state)
}

type projectedChoicesMove struct {
	moves.AnyPlayer
}

func (m *projectedChoicesMove) Apply(boardgame.State) error { return nil }

func (m *projectedChoicesMove) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	delegate, ok := state.Manager().Delegate().(*projectedChoicesDelegate)
	if !ok {
		return errors.New("unexpected projected-choice delegate")
	}
	delegate.legalCalls.Add(1)
	if err := m.Default.Legal(state, proposer); err != nil {
		return err
	}
	if delegate.rejectAll {
		return errors.New("secret rejection detail must not be published")
	}
	if m.TargetPlayerIndex == proposer {
		return errors.New("self is unavailable")
	}
	if m.TargetPlayerIndex < 0 || int(m.TargetPlayerIndex) >= len(state.ImmutablePlayerStates()) {
		return errors.New("target is invalid")
	}
	return nil
}

func newProjectedChoicesGame(t *testing.T, delegate *projectedChoicesDelegate) *boardgame.Game {
	t.Helper()
	manager, err := boardgame.NewGameManager(delegate, newLegalLedgerStorage())
	if err != nil {
		t.Fatalf("build projected-choice manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("build projected-choice game: %v", err)
	}
	return game
}

func TestProjectMoveChoicesBindsEachCandidateAndUsesFullLegalityOnce(t *testing.T) {
	delegate := newProjectedChoicesDelegate()
	game := newProjectedChoicesGame(t, delegate)
	state := game.CurrentState()
	delegate.legalCalls.Store(0)

	snapshot, err := projectMoveChoicesSnapshot(game, state, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot == nil || snapshot.Status != projectedMoveChoicesStatusReady || len(snapshot.Sets) != 1 {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	set := snapshot.Sets[0]
	if set.MoveName != "Choose Player" || set.FieldName != "TargetPlayerIndex" || set.Source != boardgame.MoveChoiceSourcePlayers {
		t.Fatalf("set identity = %#v", set)
	}
	if len(set.Candidates) != 2 {
		t.Fatalf("candidates = %#v, want two players", set.Candidates)
	}
	if set.Candidates[0].Value != boardgame.PlayerIndex(0) || set.Candidates[0].Available {
		t.Fatalf("self candidate = %#v, want unavailable", set.Candidates[0])
	}
	if set.Candidates[1].Value != boardgame.PlayerIndex(1) || !set.Candidates[1].Available {
		t.Fatalf("other candidate = %#v, want available", set.Candidates[1])
	}
	if got := delegate.legalCalls.Load(); got != 2 {
		t.Fatalf("Legal calls = %d, want one per candidate", got)
	}
	if game.Version() != state.Version() {
		t.Fatalf("projection mutated game version from %d to %d", state.Version(), game.Version())
	}

	encoded, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	wire := string(encoded)
	for _, forbidden := range []string{"OfferKey", "Prompt", "Title", "DisabledReason", "self is unavailable", "AuditRationale"} {
		if strings.Contains(wire, forbidden) {
			t.Fatalf("wire leaked %q: %s", forbidden, wire)
		}
	}
}

func TestProjectMoveChoicesIsActorOnlyAndRespectsMoveNameVisibility(t *testing.T) {
	delegate := newProjectedChoicesDelegate()
	game := newProjectedChoicesGame(t, delegate)
	state := game.CurrentState()
	delegate.legalCalls.Store(0)

	for _, audience := range []struct {
		actor  boardgame.PlayerIndex
		viewer boardgame.PlayerIndex
	}{{0, 1}, {0, boardgame.ObserverPlayerIndex}, {boardgame.AdminPlayerIndex, boardgame.AdminPlayerIndex}} {
		snapshot, err := projectMoveChoicesSnapshot(game, state, audience.actor, audience.viewer)
		if err != nil || snapshot != nil {
			t.Fatalf("audience %+v got %#v, %v; want omission", audience, snapshot, err)
		}
	}
	if delegate.legalCalls.Load() != 0 {
		t.Fatal("ineligible audience reached Legal")
	}

	hiddenDelegate := newProjectedChoicesDelegate()
	hiddenDelegate.namePolicy = "self:hidden"
	hiddenGame := newProjectedChoicesGame(t, hiddenDelegate)
	hiddenDelegate.legalCalls.Store(0)
	snapshot, err := projectMoveChoicesSnapshot(hiddenGame, hiddenGame.CurrentState(), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot == nil || snapshot.Status != projectedMoveChoicesStatusReady || len(snapshot.Sets) != 0 {
		t.Fatalf("hidden move snapshot = %#v; want authoritative empty ready snapshot", snapshot)
	}
	if hiddenDelegate.legalCalls.Load() != 0 {
		t.Fatal("hidden move name reached Legal")
	}
}

func TestProjectMoveChoicesOmitsSetWithNoLegalCandidate(t *testing.T) {
	delegate := newProjectedChoicesDelegate()
	delegate.rejectAll = true
	game := newProjectedChoicesGame(t, delegate)
	delegate.legalCalls.Store(0)

	snapshot, err := projectMoveChoicesSnapshot(game, game.CurrentState(), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot == nil || snapshot.Status != projectedMoveChoicesStatusReady || len(snapshot.Sets) != 0 {
		t.Fatalf("all-illegal snapshot = %#v", snapshot)
	}
	if delegate.legalCalls.Load() != 2 {
		t.Fatalf("Legal calls = %d, want two candidates", delegate.legalCalls.Load())
	}
}

func TestProjectMoveChoicesOmitsEnvelopeForLegacyGameWithoutProjections(t *testing.T) {
	game, _ := newLegalLedgerGame(t)
	snapshot, err := projectMoveChoicesSnapshot(game, game.CurrentState(), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot != nil {
		t.Fatalf("legacy game got projected-choice envelope: %#v", snapshot)
	}
}

func TestProjectMoveChoicesBudgetsFailBeforeExcessLegalEvaluation(t *testing.T) {
	delegate := newProjectedChoicesDelegate()
	game := newProjectedChoicesGame(t, delegate)
	schema, err := boardgame.BuildMoveChoiceProjectionSchema(game.Manager())
	if err != nil || len(schema) != 1 {
		t.Fatalf("schema = %#v, %v", schema, err)
	}
	delegate.legalCalls.Store(0)
	budget := &projectedMoveChoiceBudget{legalEvaluations: maxProjectedMoveLegalEvaluations}
	if _, err := projectMoveChoiceSet(game, game.CurrentState(), 0, schema[0], budget); err == nil || !strings.Contains(err.Error(), "Legal evaluations") {
		t.Fatalf("exhausted-budget error = %v", err)
	}
	if delegate.legalCalls.Load() != 0 {
		t.Fatal("exhausted budget performed Legal evaluation")
	}

	tooMany := schema[0]
	tooMany.Source = boardgame.MoveChoiceSourceEnumValues
	for i := 0; i <= maxProjectedMoveCandidatesPerSet; i++ {
		tooMany.CandidateValues = append(tooMany.CandidateValues, fmt.Sprintf("candidate-%d", i))
	}
	if _, err := projectMoveChoiceSet(game, game.CurrentState(), 0, tooMany, new(projectedMoveChoiceBudget)); err == nil || !strings.Contains(err.Error(), "projected candidates") {
		t.Fatalf("candidate-limit error = %v", err)
	}
	if delegate.legalCalls.Load() != 0 {
		t.Fatal("oversized candidate universe performed Legal evaluation")
	}
}

func TestProjectedMoveChoicesByteBudget(t *testing.T) {
	snapshot := &projectedMoveChoicesSnapshot{
		StateVersion:                          1,
		MoveChoiceProjectionSchemaFingerprint: "sha256:test",
		ProjectionSchemaVersion:               1,
		Status:                                projectedMoveChoicesStatusReady,
		Sets: []projectedMoveChoiceSet{{
			MoveName:  "Huge",
			FieldName: "Value",
			Source:    boardgame.MoveChoiceSourceEnumValues,
			Candidates: []projectedMoveChoiceCandidate{{
				Value:     strings.Repeat("x", maxProjectedMoveChoicesBytes),
				Available: true,
			}},
		}},
	}
	if err := validateProjectedMoveChoicesSize(snapshot); err == nil || !strings.Contains(err.Error(), "bytes") {
		t.Fatalf("byte-limit error = %v", err)
	}
}

func TestProjectedMoveChoicesFailureIsExplicitAndGeneric(t *testing.T) {
	tooMany := newProjectedChoicesDelegate()
	tooMany.projectionCount = maxProjectedMoveChoiceSets + 1
	if _, err := boardgame.NewGameManager(tooMany, newLegalLedgerStorage()); err == nil || !strings.Contains(err.Error(), "limit") {
		t.Fatalf("manager boot error = %v; want projection-set limit", err)
	}

	delegate := newProjectedChoicesDelegate()
	delegate.rejectAll = true
	delegate.moveName = strings.Repeat("oversized", maxProjectedMoveChoicesBytes/len("oversized")+1)
	game := newProjectedChoicesGame(t, delegate)
	delegate.legalCalls.Store(0)

	snapshot := (&Server{}).projectedMoveChoicesForBundle(game, game.CurrentState(), 0)
	if snapshot == nil || snapshot.Status != projectedMoveChoicesStatusFailed || len(snapshot.Sets) != 0 {
		t.Fatalf("failed snapshot = %#v", snapshot)
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "limit") || strings.Contains(string(encoded), "game configures") {
		t.Fatalf("failed snapshot leaked internal diagnosis: %s", encoded)
	}
	if snapshot.MoveChoiceProjectionSchemaFingerprint == "" || snapshot.ProjectionSchemaVersion == 0 {
		t.Fatalf("failed snapshot omitted schema identity: %#v", snapshot)
	}
	if delegate.legalCalls.Load() != 0 {
		t.Fatalf("wire preflight performed %d Legal calls before failing", delegate.legalCalls.Load())
	}
}

type projectedChoicesBlockingStorage struct {
	*legalLedgerStorage
	blockVersion int
	saved        chan struct{}
	release      chan struct{}
}

func (s *projectedChoicesBlockingStorage) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {
	if err := s.legalLedgerStorage.SaveGameAndCurrentState(game, state, move); err != nil {
		return err
	}
	if move != nil && move.Version == s.blockVersion {
		s.saved <- struct{}{}
		<-s.release
	}
	return nil
}

func TestProjectedMoveChoicesDeliveryOnlyAtRecoverableFrontier(t *testing.T) {
	delegate := newProjectedChoicesDelegate()
	storage := &projectedChoicesBlockingStorage{
		legalLedgerStorage: newLegalLedgerStorage(),
		saved:              make(chan struct{}, 1),
		release:            make(chan struct{}),
	}
	manager, err := boardgame.NewGameManager(delegate, storage)
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{}

	initial, err := s.moveBundles(game, nil, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot, ok := initial[0]["ProjectedMoveChoices"].(*projectedMoveChoicesSnapshot); !ok || snapshot.Status != projectedMoveChoicesStatusReady {
		t.Fatalf("initial bundle choices = %#v", initial[0]["ProjectedMoveChoices"])
	}

	storage.blockVersion = game.Version() + 1
	move := game.MoveByName("Choose Player")
	if err := move.ReadSetter().SetPlayerIndexProp("TargetPlayerIndex", 1); err != nil {
		t.Fatal(err)
	}
	delayed := game.ProposeMove(move, 0)
	<-storage.saved
	intermediate := game.State(game.Version())
	if snapshot := s.projectedMoveChoicesForBundle(game, intermediate, 0); snapshot != nil {
		t.Fatalf("intermediate durable state advertised choices: %#v", snapshot)
	}

	close(storage.release)
	if err := <-delayed; err != nil {
		t.Fatal(err)
	}
	current := game.CurrentState()
	if snapshot := s.projectedMoveChoicesForBundle(game, current, 0); snapshot == nil || snapshot.Status != projectedMoveChoicesStatusReady {
		t.Fatalf("settled frontier choices = %#v", snapshot)
	}

	movesSinceStart, err := game.Manager().Storage().Moves(game.ID(), 0, game.Version())
	if err != nil {
		t.Fatal(err)
	}
	// Add another move so the first bundle is historical and only the final
	// bundle may carry the version-pinned projection.
	move = game.MoveByName("Choose Player")
	if err := move.ReadSetter().SetPlayerIndexProp("TargetPlayerIndex", 1); err != nil {
		t.Fatal(err)
	}
	if err := <-game.ProposeMove(move, 0); err != nil {
		t.Fatal(err)
	}
	movesSinceStart, err = game.Manager().Storage().Moves(game.ID(), 0, game.Version())
	if err != nil {
		t.Fatal(err)
	}
	bundles, err := s.moveBundles(game, movesSinceStart, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := bundles[0]["ProjectedMoveChoices"]; exists {
		t.Fatalf("historical bundle carried choices: %#v", bundles[0]["ProjectedMoveChoices"])
	}
	if _, ok := bundles[len(bundles)-1]["ProjectedMoveChoices"].(*projectedMoveChoicesSnapshot); !ok {
		t.Fatalf("final bundle omitted choices: %#v", bundles[len(bundles)-1])
	}
}

func TestProjectedMoveChoicesAutoCurrentDisplayDoesNotGrantActorAudience(t *testing.T) {
	delegate := newProjectedChoicesDelegate()
	game := newProjectedChoicesGame(t, delegate)
	s := &Server{}

	// The fixture's current player is 0. Player 1 asking for the auto-current
	// display may see player 0's sanitized board, but must not receive player
	// 0's actor-exact choices.
	bundles, err := s.moveBundles(game, nil, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	if bundles[0]["ViewingAsPlayer"] != boardgame.PlayerIndex(0) {
		t.Fatalf("fixture did not switch display perspective: %#v", bundles[0]["ViewingAsPlayer"])
	}
	if _, exists := bundles[0]["ProjectedMoveChoices"]; exists {
		t.Fatalf("auto-current display granted another actor's choices: %#v", bundles[0]["ProjectedMoveChoices"])
	}
}

type projectedChoicesReconcileStorage struct {
	*legalLedgerStorage
	frontierFailures atomic.Int32
}

func (s *projectedChoicesReconcileStorage) SaveProposalFrontier(gameID string, stateVersion, frontierVersion int) error {
	for {
		remaining := s.frontierFailures.Load()
		if remaining <= 0 {
			break
		}
		if s.frontierFailures.CompareAndSwap(remaining, remaining-1) {
			return errors.New("injected stale proposal-frontier CAS")
		}
	}
	return s.legalLedgerStorage.SaveProposalFrontier(gameID, stateVersion, frontierVersion)
}

func setProjectedChoicesFrontierUnknown(t *testing.T, storage *legalLedgerStorage, gameID string) {
	t.Helper()
	storage.mu.Lock()
	defer storage.mu.Unlock()
	record := storage.games[gameID]
	if record == nil {
		t.Fatalf("missing stored game %q", gameID)
	}
	record.ProposalFrontierKnown = false
	record.ProposalFrontierVersion = 0
}

func newProjectedChoicesReconcileGame(t *testing.T, delegate *projectedChoicesDelegate, storage *projectedChoicesReconcileStorage) (*boardgame.GameManager, *boardgame.Game) {
	t.Helper()
	manager, err := boardgame.NewGameManager(delegate, storage)
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	return manager, game
}

func restartProjectedChoicesManager(t *testing.T, delegate *projectedChoicesDelegate, storage boardgame.StorageManager) *boardgame.GameManager {
	t.Helper()
	manager, err := boardgame.NewGameManager(delegate, storage)
	if err != nil {
		t.Fatal(err)
	}
	return manager
}

func TestReconcileProjectedMoveChoiceFrontierRepairsLegacyMarkerAndStaleCAS(t *testing.T) {
	delegate := newProjectedChoicesDelegate()
	storage := &projectedChoicesReconcileStorage{legalLedgerStorage: newLegalLedgerStorage()}
	_, game := newProjectedChoicesReconcileGame(t, delegate, storage)
	setProjectedChoicesFrontierUnknown(t, storage.legalLedgerStorage, game.ID())
	storage.frontierFailures.Store(1)
	manager := restartProjectedChoicesManager(t, newProjectedChoicesDelegate(), storage)

	reconciled, err := reconcileProjectedMoveChoiceFrontier(manager.Game(game.ID()))
	if err != nil {
		t.Fatal(err)
	}
	if reconciled == nil || !projectedMoveChoiceFrontierKnown(reconciled) || reconciled.Version() != game.Version() {
		t.Fatalf("reconciled game = %#v", reconciled)
	}
	if storage.frontierFailures.Load() != 0 {
		t.Fatal("injected stale CAS was not exercised")
	}
}

func TestReconcileProjectedMoveChoiceFrontierCompletesPartialFixUpChain(t *testing.T) {
	delegate := newProjectedChoicesDelegate()
	storage := &projectedChoicesReconcileStorage{legalLedgerStorage: newLegalLedgerStorage()}
	_, game := newProjectedChoicesReconcileGame(t, delegate, storage)
	start := game.Version()
	setProjectedChoicesFrontierUnknown(t, storage.legalLedgerStorage, game.ID())
	restartedDelegate := newProjectedChoicesDelegate()
	restartedDelegate.fixupUntilVersion.Store(int32(start + 2))
	manager := restartProjectedChoicesManager(t, restartedDelegate, storage)

	reconciled, err := reconcileProjectedMoveChoiceFrontier(manager.Game(game.ID()))
	if err != nil {
		t.Fatal(err)
	}
	if reconciled.Version() != start+2 || !projectedMoveChoiceFrontierKnown(reconciled) {
		t.Fatalf("partial chain recovered to version %d, frontier=%v; want %d", reconciled.Version(), projectedMoveChoiceFrontierKnown(reconciled), start+2)
	}
}

func TestReconcileProjectedMoveChoiceFrontierSerializesConcurrentRequests(t *testing.T) {
	delegate := newProjectedChoicesDelegate()
	storage := &projectedChoicesReconcileStorage{legalLedgerStorage: newLegalLedgerStorage()}
	_, game := newProjectedChoicesReconcileGame(t, delegate, storage)
	setProjectedChoicesFrontierUnknown(t, storage.legalLedgerStorage, game.ID())
	manager := restartProjectedChoicesManager(t, newProjectedChoicesDelegate(), storage)

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reconciled, err := reconcileProjectedMoveChoiceFrontier(manager.Game(game.ID()))
			if err == nil && (reconciled == nil || !projectedMoveChoiceFrontierKnown(reconciled)) {
				err = errors.New("reconciliation returned an unknown frontier")
			}
			errs <- err
		}()
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("concurrent frontier reconciliations did not terminate")
	}
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestReconcileProjectedMoveChoiceFrontierSkipsLegacyGamesWithoutProjection(t *testing.T) {
	storage := newLegalLedgerStorage()
	manager, err := boardgame.NewGameManager(&legalLedgerDelegate{}, storage)
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	setProjectedChoicesFrontierUnknown(t, storage, game.ID())
	frozen := manager.Game(game.ID())
	got, err := reconcileProjectedMoveChoiceFrontier(frozen)
	if err != nil {
		t.Fatal(err)
	}
	if got != frozen {
		t.Fatal("legacy game without projections was unnecessarily refreshed")
	}
	if manager.Game(game.ID()).AtProposalFrontier() {
		t.Fatal("legacy game without projections was mutated")
	}
}

func TestReconcileProjectedMoveChoiceFrontierFailsAfterBoundedAttempts(t *testing.T) {
	delegate := newProjectedChoicesDelegate()
	storage := &projectedChoicesReconcileStorage{legalLedgerStorage: newLegalLedgerStorage()}
	_, game := newProjectedChoicesReconcileGame(t, delegate, storage)
	setProjectedChoicesFrontierUnknown(t, storage.legalLedgerStorage, game.ID())
	storage.frontierFailures.Store(maxProjectedMoveReconcileAttempts + 1)
	manager := restartProjectedChoicesManager(t, newProjectedChoicesDelegate(), storage)

	_, err := reconcileProjectedMoveChoiceFrontier(manager.Game(game.ID()))
	if err == nil || !strings.Contains(err.Error(), "after 3 attempts") {
		t.Fatalf("reconciliation error = %v", err)
	}
}

type projectedChoicesNoopFrontierStorage struct {
	boardgame.StorageManager
}

func (*projectedChoicesNoopFrontierStorage) SaveProposalFrontier(string, int, int) error {
	return nil
}

func TestReconcileProjectedMoveChoiceFrontierAcceptsActiveMarkerWhenStorageCannotPersistIt(t *testing.T) {
	delegate := newProjectedChoicesDelegate()
	underlying := newLegalLedgerStorage()
	// Model ServerStorageManager around a legacy/custom backend: the wrapper
	// advertises ProposalFrontierStorage but its successful write cannot make
	// the marker durable in the underlying game record.
	storage := &projectedChoicesNoopFrontierStorage{StorageManager: underlying}
	manager, err := boardgame.NewGameManager(delegate, storage)
	if err != nil {
		t.Fatal(err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
	if manager.Game(game.ID()).AtProposalFrontier() {
		t.Fatal("no-op frontier storage unexpectedly persisted the marker")
	}

	reconciled, err := reconcileProjectedMoveChoiceFrontier(manager.Game(game.ID()))
	if err != nil {
		t.Fatal(err)
	}
	if !projectedMoveChoiceFrontierKnown(reconciled) {
		t.Fatal("active-process frontier was ignored because the frozen record remained unknown")
	}
}
