package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/tablelease"
	"github.com/sirupsen/logrus"
	"github.com/workfit/tester/assert"
)

type readTrackingBody struct{ read bool }

func (b *readTrackingBody) Read([]byte) (int, error) { b.read = true; return 0, io.EOF }
func (b *readTrackingBody) Close() error             { return nil }

type fencingHostStorage struct {
	StorageManager
	extended *extendedgame.StorageRecord
	lease    *tablelease.StorageRecord
	reads    int
}

func (s *fencingHostStorage) ExtendedGame(string) (*extendedgame.StorageRecord, error) {
	return s.extended, nil
}

func (s *fencingHostStorage) CompanionTableLease(string) (*tablelease.StorageRecord, error) {
	s.reads++
	record := s.lease.Clone()
	if s.reads >= 3 {
		record.DeviceID = "fedcba9876543210fedcba9876543210"
		record.SecretDigest = strings.Repeat("0", 64)
	}
	return record, nil
}

type hostLeaseStorage struct {
	StorageManager
	extended *extendedgame.StorageRecord
	lease    *tablelease.StorageRecord
}

type actionFenceStorage struct {
	StorageManager
	mu               sync.Mutex
	extended         *extendedgame.StorageRecord
	lease            *tablelease.StorageRecord
	mutationSawFence bool
}

func (s *actionFenceStorage) ExtendedGame(string) (*extendedgame.StorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := *s.extended
	return &copy, nil
}

func (s *actionFenceStorage) UpdateExtendedGame(_ string, record *extendedgame.StorageRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mutationSawFence = s.lease.TransitionKind == tablelease.TransitionHostAction
	copy := *record
	s.extended = &copy
	return nil
}

func (s *actionFenceStorage) CompanionTableLease(string) (*tablelease.StorageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lease.Clone(), nil
}

func (s *actionFenceStorage) CompareAndSwapCompanionTableLease(_ string, expected uint64, replacement *tablelease.StorageRecord) (*tablelease.StorageRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lease.Generation != expected {
		return s.lease.Clone(), false, nil
	}
	next := replacement.Clone()
	next.Generation = expected + 1
	s.lease = next
	return next.Clone(), true, nil
}

func (s *hostLeaseStorage) ExtendedGame(string) (*extendedgame.StorageRecord, error) {
	return s.extended, nil
}

func (s *hostLeaseStorage) CompanionTableLease(string) (*tablelease.StorageRecord, error) {
	return s.lease.Clone(), nil
}

func hostLeaseTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	credentialServer := &Server{tableLeaseKey: []byte("0123456789abcdef0123456789abcdef")}
	deviceID, secret, digest, err := credentialServer.newTableLeaseCredential("gameXYZ")
	if err != nil {
		t.Fatal(err)
	}
	storage := &hostLeaseStorage{
		extended: &extendedgame.StorageRecord{CompanionRoomCode: "ROOM"},
		lease: &tablelease.StorageRecord{
			DeviceID: deviceID, SecretDigest: digest, Expires: time.Now().Add(time.Minute).UnixMilli(),
		},
	}
	return &Server{storage: NewServerStorageManager(storage), logger: logrus.New()}, deviceID + "." + secret
}

// TestIsHostRequiresSurfaceTableCookie pins the rule that a valid capability
// on a browser currently acting as a Hand cannot keep the Table alive.
func TestIsHostRequiresSurfaceTableCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s, credential := hostLeaseTestServer(t)

	// Build a gin.Context with NO surface cookie set. Even with a
	// matching userID, isHost should reject.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/game/x/y/hostSkipTurn", nil)
	c.Request.AddCookie(&http.Cookie{Name: tableLeaseCookieName("gameXYZ"), Value: credential})

	// (We can't easily wire getUser() without a full Server; this test
	// covers the cookie-absent branch where the function returns false
	// before getUser is consulted.)
	got := s.isHost(c, "gameXYZ", "ownerUID")
	if got {
		t.Error("isHost returned true with no surface cookie; expected false (spec §9.4)")
	}
}

// TestIsHostRequiresSurfaceTableValue covers the complete authority tuple:
// companion game + unexpired lease + exact secret + explicit Table surface.
func TestIsHostRequiresSurfaceTableValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s, credential := hostLeaseTestServer(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/game/x/y/hostSkipTurn", nil)
	req.AddCookie(&http.Cookie{Name: surfaceCookieName("gameXYZ"), Value: "hand"})
	req.AddCookie(&http.Cookie{Name: tableLeaseCookieName("gameXYZ"), Value: credential})
	c.Request = req

	got := s.isHost(c, "gameXYZ", "ownerUID")
	if got {
		t.Error("isHost returned true with surface=hand cookie; expected false (spec §9.4)")
	}

	req = httptest.NewRequest(http.MethodPost, "/api/game/x/y/hostSkipTurn", nil)
	req.AddCookie(&http.Cookie{Name: surfaceCookieName("gameXYZ"), Value: "table"})
	req.AddCookie(&http.Cookie{Name: tableLeaseCookieName("gameXYZ"), Value: credential})
	c.Request = req
	if !s.isHost(c, "gameXYZ", "ignored") {
		t.Error("isHost rejected a valid fenced Table credential")
	}
}

// TestHostActionAllowedRateLimit pins the 1-action-per-second rate limit
// added in P3.5. Two rapid calls for the same (gameID, userID): first
// succeeds, second is throttled.
func TestHostActionAllowedRateLimit(t *testing.T) {
	// Clear the package-level locks to ensure deterministic state.
	hostActionLocksMu.Lock()
	hostActionLocks = map[string]time.Time{}
	hostActionLocksMu.Unlock()

	assert.For(t).ThatActual(hostActionAllowed("g1", "u1")).IsTrue()
	// Immediately after — should be throttled.
	assert.For(t).ThatActual(hostActionAllowed("g1", "u1")).IsFalse()

	// Different game OR different user is not throttled.
	assert.For(t).ThatActual(hostActionAllowed("g2", "u1")).IsTrue()
	assert.For(t).ThatActual(hostActionAllowed("g1", "u2")).IsTrue()
}

// TestHostActionAllowedRefillsAfterDelay confirms the rate limit window
// actually advances. Wait past hostActionRateLimit and the same key
// should succeed again.
func TestHostActionAllowedRefillsAfterDelay(t *testing.T) {
	hostActionLocksMu.Lock()
	hostActionLocks = map[string]time.Time{}
	hostActionLocksMu.Unlock()

	assert.For(t).ThatActual(hostActionAllowed("g1", "u1")).IsTrue()
	// Manually backdate the recorded timestamp so we don't actually
	// sleep for a second in the test.
	hostActionLocksMu.Lock()
	hostActionLocks["g1:u1"] = time.Now().Add(-2 * time.Second)
	hostActionLocksMu.Unlock()

	assert.For(t).ThatActual(hostActionAllowed("g1", "u1")).IsTrue()
}

func TestHostActionFenceBlocksTransferAndCoversMetadataMutation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	game, _ := newLegalLedgerGame(t)
	key := []byte("0123456789abcdef0123456789abcdef")
	credentialServer := &Server{tableLeaseKey: key}
	deviceID, secret, digest, err := credentialServer.newTableLeaseCredential(game.ID())
	if err != nil {
		t.Fatal(err)
	}
	storage := &actionFenceStorage{
		extended: &extendedgame.StorageRecord{CompanionRoomCode: "ABCD"},
		lease: &tablelease.StorageRecord{
			Generation: 1, DeviceID: deviceID, SecretDigest: digest, HolderUserID: "host",
			Expires:             time.Now().Add(time.Minute).UnixMilli(),
			TransferID:          "0123456789abcdef0123456789abcdef",
			TransferTokenDigest: strings.Repeat("a", 64), TransferCodeDigest: strings.Repeat("b", 64),
			TransferExpires: time.Now().Add(time.Minute).UnixMilli(),
		},
	}
	s := &Server{
		tableLeaseKey: key, storage: NewServerStorageManager(storage), logger: logrus.New(),
		seatJoinLocks: map[string]*sync.Mutex{},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/game/test/id/setRoomLock", strings.NewReader(`{"locked":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: surfaceCookieName(game.ID()), Value: "table"})
	req.AddCookie(&http.Cookie{Name: tableLeaseCookieName(game.ID()), Value: deviceID + "." + secret})
	c.Request = req
	s.setGame(c, game)
	hostActionLocksMu.Lock()
	hostActionLocks = map[string]time.Time{}
	hostActionLocksMu.Unlock()

	s.setRoomLockHandler(c)
	if w.Code != http.StatusOK || !storage.mutationSawFence {
		t.Fatalf("room lock status/fence = %d/%t", w.Code, storage.mutationSawFence)
	}
	lease, _ := storage.CompanionTableLease(game.ID())
	if tableTransferBlockedByHostAction(lease) || !lease.TransferPending(time.Now().UnixMilli()) {
		t.Fatalf("action fence did not release while preserving transfer: %+v", lease)
	}

	// Both lock and skip use the same primitive. While it is held, a second
	// host action and transfer redemption must fail closed; after release the
	// original pending offer is usable again.
	fence, result := s.beginTableLeaseAction(c, game.ID())
	if result != tableLeaseRenewed {
		t.Fatalf("begin fence = %v", result)
	}
	lease, _ = storage.CompanionTableLease(game.ID())
	if !tableTransferBlockedByHostAction(lease) {
		t.Fatal("live host action did not block transfer redemption")
	}
	fenceDeadline := lease.Expires
	if renewal := s.renewTableLeaseCredential(game.ID(), deviceID+"."+secret); renewal != tableLeaseRenewRetryable {
		t.Fatalf("heartbeat during host action = %v; want retryable without renewal", renewal)
	}
	lease, _ = storage.CompanionTableLease(game.ID())
	if lease.Expires != fenceDeadline {
		t.Fatal("heartbeat extended a host-action crash-recovery deadline")
	}
	if _, second := s.beginTableLeaseAction(c, game.ID()); second != tableLeaseRenewRetryable {
		t.Fatalf("concurrent host action = %v; want retryable", second)
	}
	if !s.endTableLeaseAction(game.ID(), fence) {
		t.Fatal("could not release action fence")
	}
}

func TestUnauthorisedRoomLockDoesNotReadRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	game, _ := newLegalLedgerGame(t)
	s, _ := hostLeaseTestServer(t)
	s.logger = logrus.New()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := &readTrackingBody{}
	req := httptest.NewRequest(http.MethodPost, "/api/game/test/id/setRoomLock", nil)
	req.Body = body
	c.Request = req
	s.setGame(c, game)
	s.setRoomLockHandler(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d; want 403", w.Code)
	}
	if body.read {
		t.Fatal("unauthorised public request body was read before capability validation")
	}
}

func TestFencedHostCannotMutateRoomMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	game, _ := newLegalLedgerGame(t)
	key := []byte("0123456789abcdef0123456789abcdef")
	credentialServer := &Server{tableLeaseKey: key}
	deviceID, secret, digest, err := credentialServer.newTableLeaseCredential(game.ID())
	if err != nil {
		t.Fatal(err)
	}
	newStorage := func() *fencingHostStorage {
		return &fencingHostStorage{
			extended: &extendedgame.StorageRecord{CompanionRoomCode: "ABCD"},
			lease:    &tablelease.StorageRecord{DeviceID: deviceID, SecretDigest: digest, HolderUserID: "host", Expires: time.Now().Add(time.Minute).UnixMilli()},
		}
	}
	request := func(path, body string, handler func(*Server, *gin.Context)) (*fencingHostStorage, int) {
		t.Helper()
		storage := newStorage()
		s := &Server{
			tableLeaseKey: key, storage: NewServerStorageManager(storage), logger: logrus.New(),
			seatJoinLocks: map[string]*sync.Mutex{},
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: surfaceCookieName(game.ID()), Value: "table"})
		req.AddCookie(&http.Cookie{Name: tableLeaseCookieName(game.ID()), Value: deviceID + "." + secret})
		c.Request = req
		s.setGame(c, game)
		hostActionLocksMu.Lock()
		hostActionLocks = map[string]time.Time{}
		hostActionLocksMu.Unlock()
		handler(s, c)
		return storage, w.Code
	}

	roomStorage, status := request("/api/game/test/id/setRoomLock", `{"locked":true}`, func(s *Server, c *gin.Context) { s.setRoomLockHandler(c) })
	if status != http.StatusConflict || roomStorage.extended.CompanionLocked {
		t.Fatalf("fenced room-lock status/record = %d/%+v", status, roomStorage.extended)
	}
	soloStorage, status := request("/api/game/test/id/switchToSolo", `{}`, func(s *Server, c *gin.Context) { s.switchToSoloHandler(c) })
	if status != http.StatusConflict || soloStorage.extended.CompanionRoomCode != "ABCD" {
		t.Fatalf("fenced solo-switch status/record = %d/%+v", status, soloStorage.extended)
	}
}

// TestHostActionLocksEvictionTriggersAtThreshold confirms Polish 7's
// opportunistic eviction inside hostActionAllowed. Populate the map
// past the threshold with old entries; one fresh call should sweep
// them out.
func TestHostActionLocksEvictionTriggersAtThreshold(t *testing.T) {
	hostActionLocksMu.Lock()
	hostActionLocks = map[string]time.Time{}
	// 33 stale entries (> threshold of 32). All are an hour+1min old —
	// past hostActionLocksEvictAge.
	stale := time.Now().Add(-(hostActionLocksEvictAge + time.Minute))
	for i := 0; i < 33; i++ {
		hostActionLocks["g"+string(rune('a'+i))+":u"] = stale
	}
	hostActionLocksMu.Unlock()

	// Fresh call triggers the eviction sweep.
	hostActionAllowed("freshGame", "freshUser")

	hostActionLocksMu.Lock()
	remaining := len(hostActionLocks)
	hostActionLocksMu.Unlock()

	// The stale 33 should be evicted; only the fresh entry remains.
	if remaining > 1 {
		t.Errorf("Expected at most 1 fresh entry after eviction, got %d", remaining)
	}
}
