package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/workfit/tester/assert"
)

// TestIsHostRequiresSurfaceTableCookie pins the spec §9.4 rule: a user
// (even the original game Owner) on a phone with no surface=table
// cookie is NOT host. This is the load-bearing gate that keeps host
// privileges with the projector.
func TestIsHostRequiresSurfaceTableCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &Server{
		logger: logrus.New(),
	}

	// Build a gin.Context with NO surface cookie set. Even with a
	// matching userID, isHost should reject.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/game/x/y/hostSkipTurn", nil)

	// (We can't easily wire getUser() without a full Server; this test
	// covers the cookie-absent branch where the function returns false
	// before getUser is consulted.)
	got := s.isHost(c, "gameXYZ", "ownerUID")
	if got {
		t.Error("isHost returned true with no surface cookie; expected false (spec §9.4)")
	}
}

// TestIsHostRequiresSurfaceTableValue: cookie present but with the wrong
// value (surface=hand) should also reject — the phone surface is not the
// table surface.
func TestIsHostRequiresSurfaceTableValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &Server{
		logger: logrus.New(),
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/game/x/y/hostSkipTurn", nil)
	req.AddCookie(&http.Cookie{Name: surfaceCookieName("gameXYZ"), Value: "hand"})
	c.Request = req

	got := s.isHost(c, "gameXYZ", "ownerUID")
	if got {
		t.Error("isHost returned true with surface=hand cookie; expected false (spec §9.4)")
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
