package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/tablelease"
	"github.com/sirupsen/logrus"
	"github.com/workfit/tester/assert"
)

type hostLeaseStorage struct {
	StorageManager
	extended *extendedgame.StorageRecord
	lease    *tablelease.StorageRecord
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
