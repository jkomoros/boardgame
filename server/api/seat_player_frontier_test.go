package api

import (
	"errors"
	"sync"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/users"
	"github.com/sirupsen/logrus"
)

// autoSeatTestStorage backs the auto-seat race regression test. It reuses
// legalLedgerStorage for the core boardgame.StorageManager surface (including
// its version-checked SaveProposalFrontier) and adds just enough of the
// server-level StorageManager interface for the auto-seat flow:
// UserIDsForGame / SetPlayerForGame (with bolt's semantics: idempotent for the
// same user+seat, rejecting an occupied seat) and a minimal ExtendedGame.
// Every other method comes from the embedded nil interface and panics if
// reached, which is what we want in a test.
type autoSeatTestStorage struct {
	StorageManager
	core *legalLedgerStorage

	mu      sync.Mutex
	players map[string][]string
	users   map[string]*users.StorageRecord
}

func newAutoSeatTestStorage() *autoSeatTestStorage {
	return &autoSeatTestStorage{
		core:    newLegalLedgerStorage(),
		players: make(map[string][]string),
		users:   make(map[string]*users.StorageRecord),
	}
}

//The core boardgame.StorageManager surface forwards to legalLedgerStorage.

func (a *autoSeatTestStorage) State(gameID string, version int) (boardgame.StateStorageRecord, error) {
	return a.core.State(gameID, version)
}

func (a *autoSeatTestStorage) Move(gameID string, version int) (*boardgame.MoveStorageRecord, error) {
	return a.core.Move(gameID, version)
}

func (a *autoSeatTestStorage) Moves(gameID string, fromVersion, toVersion int) ([]*boardgame.MoveStorageRecord, error) {
	return a.core.Moves(gameID, fromVersion, toVersion)
}

// Game returns a copy: legalLedgerStorage both hands out and mutates (in
// SaveProposalFrontier) the same record pointer, which is fine for the
// serialized legal-ledger tests but a data race for this test's concurrent
// readers.
func (a *autoSeatTestStorage) Game(id string) (*boardgame.GameStorageRecord, error) {
	a.core.mu.Lock()
	defer a.core.mu.Unlock()
	record, ok := a.core.games[id]
	if !ok {
		return nil, errors.New("no such game")
	}
	copied := *record
	return &copied, nil
}

func (a *autoSeatTestStorage) AgentState(gameID string, player boardgame.PlayerIndex) ([]byte, error) {
	return a.core.AgentState(gameID, player)
}

func (a *autoSeatTestStorage) SaveGameAndCurrentState(game *boardgame.GameStorageRecord, state boardgame.StateStorageRecord, move *boardgame.MoveStorageRecord) error {
	return a.core.SaveGameAndCurrentState(game, state, move)
}

func (a *autoSeatTestStorage) SaveProposalFrontier(gameID string, stateVersion, frontierVersion int) error {
	return a.core.SaveProposalFrontier(gameID, stateVersion, frontierVersion)
}

func (a *autoSeatTestStorage) SaveAgentState(gameID string, player boardgame.PlayerIndex, state []byte) error {
	return a.core.SaveAgentState(gameID, player, state)
}

func (a *autoSeatTestStorage) PlayerMoveApplied(game *boardgame.GameStorageRecord) error {
	return a.core.PlayerMoveApplied(game)
}

func (a *autoSeatTestStorage) FetchInjectedDataForGame(gameID string, dataType string) interface{} {
	return a.core.FetchInjectedDataForGame(gameID, dataType)
}

//Server-level methods the auto-seat flow touches.

func (a *autoSeatTestStorage) UserIDsForGame(gameID string) []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	rec, err := a.core.Game(gameID)
	if err != nil {
		return nil
	}
	ids, ok := a.players[gameID]
	if !ok {
		return make([]string, rec.NumPlayers)
	}
	result := make([]string, len(ids))
	copy(result, ids)
	return result
}

// SetPlayerForGame mirrors storage/bolt's semantics: seating the same user in
// the seat they already hold succeeds silently (this is exactly what lets two
// racing auto-seat requests both get past the storage write), seating them in
// a different seat or taking an occupied seat errors.
func (a *autoSeatTestStorage) SetPlayerForGame(gameID string, playerIndex boardgame.PlayerIndex, userID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	rec, err := a.core.Game(gameID)
	if err != nil {
		return err
	}
	ids, ok := a.players[gameID]
	if !ok {
		ids = make([]string, rec.NumPlayers)
	}
	if int(playerIndex) < 0 || int(playerIndex) >= len(ids) {
		return errors.New("invalid player index")
	}
	for i, existing := range ids {
		if existing == userID {
			if boardgame.PlayerIndex(i) == playerIndex {
				return nil
			}
			return errors.New("user already assigned to another seat")
		}
	}
	if ids[playerIndex] != "" {
		return errors.New("seat already taken")
	}
	ids[playerIndex] = userID
	a.players[gameID] = ids
	return nil
}

func (a *autoSeatTestStorage) ExtendedGame(id string) (*extendedgame.StorageRecord, error) {
	return &extendedgame.StorageRecord{Open: true, Visible: true}, nil
}

func (a *autoSeatTestStorage) UpdateExtendedGame(id string, eGame *extendedgame.StorageRecord) error {
	return nil
}

func (a *autoSeatTestStorage) GetUserByID(uid string) *users.StorageRecord {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.users[uid]
}

func (a *autoSeatTestStorage) UpdateUser(user *users.StorageRecord) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.users[user.ID] = user
	return nil
}

// newAutoSeatTestServer wires the minimal Server needed by the first-viewer
// auto-seat path (autoSeatFirstViewer -> doSeatPlayer -> forceSeatPlayer ->
// ForceFixUp), with a real memory-game manager whose storage is the
// ServerStorageManager so SeatPlayer's rendezvous injection works exactly as
// in production.
func newAutoSeatTestServer(t *testing.T) (*Server, *autoSeatTestStorage) {
	t.Helper()

	underlying := newAutoSeatTestStorage()
	storage := NewServerStorageManager(underlying)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	server := &Server{
		managers:        make(managerMap),
		playersToSeat:   make(map[string][]*playerToSeat),
		gameOverEmitted: make(map[string]bool),
		storage:         storage,
		logger:          logger,
		seatJoinLocks:   make(map[string]*sync.Mutex),
	}
	server.config = &config.Mode{}
	storage.server = server
	server.notifier = newVersionNotifier(server)

	manager, err := boardgame.NewGameManager(memory.NewDelegate(), storage)
	if err != nil {
		t.Fatal(err)
	}
	manager.SetLogger(logger)

	server.managers[manager.Delegate().Name()] = &managerInfo{
		manager:         manager,
		seatPlayerMoves: managerSeatPlayerMoves(manager),
		playerHasSeat:   true,
	}

	t.Cleanup(func() { server.notifier.done() })

	return server, underlying
}

// TestAutoSeatFirstViewerRaceSeatsExactlyOnce reproduces the offline-dev bug
// where the game page's initial burst of API requests (e.g. /info and /socket)
// all hit gameAPISetup's first-viewer auto-seat concurrently. Without
// serialization, several requests pass the "every seat is empty" check, each
// queues its own pending SeatPlayer fix-up, and: (1) duplicate SeatPlayer
// moves apply, corrupting the game's seating state, and (2) a loser's forced
// fix-up races the winner's in-flight commit and fails with "proposal frontier
// used a stale game version", aborting the request. The auto-seat must take
// the same per-game seat-claim lock as /api/join/seat and re-check under it.
func TestAutoSeatFirstViewerRaceSeatsExactlyOnce(t *testing.T) {

	const attempts = 25
	const concurrency = 8

	for i := 0; i < attempts; i++ {
		server, underlying := newAutoSeatTestServer(t)
		manager := server.managers["memory"].manager

		game, err := manager.NewDefaultGame()
		if err != nil {
			t.Fatal(err)
		}

		user := &users.StorageRecord{ID: "creator"}
		if err := underlying.UpdateUser(user); err != nil {
			t.Fatal(err)
		}

		closedSeats := server.closedSeatsForGame(game)

		var wg sync.WaitGroup
		start := make(chan struct{})
		errs := make([]error, concurrency)
		seatedCount := 0
		var seatedMu sync.Mutex

		for j := 0; j < concurrency; j++ {
			wg.Add(1)
			go func(j int) {
				defer wg.Done()
				// Each concurrent request works from its own snapshot, like
				// each HTTP request's gameAPISetup does.
				snapshot := manager.Game(game.ID())
				<-start
				_, _, seated, err := server.autoSeatFirstViewer(snapshot, user, closedSeats)
				errs[j] = err
				if seated {
					seatedMu.Lock()
					seatedCount++
					seatedMu.Unlock()
				}
			}(j)
		}
		close(start)
		wg.Wait()

		for j, err := range errs {
			if err != nil {
				t.Fatalf("attempt %d: concurrent auto-seat %d failed: %v", i, j, err)
			}
		}

		if seatedCount != 1 {
			t.Fatalf("attempt %d: %d requests performed the auto-seat; want exactly 1", i, seatedCount)
		}

		// Exactly one SeatPlayer move must have been applied for the creator.
		rec, err := underlying.Game(game.ID())
		if err != nil {
			t.Fatal(err)
		}
		moves, err := underlying.Moves(game.ID(), 0, rec.Version)
		if err != nil {
			t.Fatal(err)
		}
		seatMoves := 0
		for _, move := range moves {
			if move != nil && move.Name == "Seat Player" {
				seatMoves++
			}
		}
		if seatMoves != 1 {
			t.Fatalf("attempt %d: %d Seat Player moves were applied; want exactly 1", i, seatMoves)
		}

		// No ghost pending SeatPlayer work may remain.
		server.mu.Lock()
		pending := len(server.playersToSeat[game.ID()])
		server.mu.Unlock()
		if pending != 0 {
			t.Fatalf("attempt %d: %d pending playersToSeat left behind; want 0", i, pending)
		}
	}
}
