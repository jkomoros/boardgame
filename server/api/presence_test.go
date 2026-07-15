package api

import (
	"testing"
	"time"

	"github.com/jkomoros/boardgame"
	"github.com/sirupsen/logrus"
	"github.com/workfit/tester/assert"
)

// Tests for the presence state machine on versionNotifier (spec §9.1).
// We exercise markAbsent / clearAbsentIfPresent / AbsentPlayers /
// scanStaleHeartbeats directly without spinning up the full workLoop
// goroutine. The notifier under test is constructed with a stub server
// (just the logger field) to satisfy the calls into v.server.logger.

func newTestNotifier(t *testing.T) *versionNotifier {
	t.Helper()
	// Construct a minimal Server that has a logger; everything else can
	// stay zero. newVersionNotifier reads s.server.logger.
	srv := &Server{logger: logrus.New()}
	v := &versionNotifier{
		sockets:           map[string]map[*socket]bool{},
		register:          make(chan *socket),
		unregister:        make(chan *socket),
		notifyVersion:     make(chan gameVersionChanged),
		notifyChat:        make(chan chatBroadcast),
		heartbeat:         make(chan heartbeatRecord, 4),
		doneChan:          make(chan bool),
		server:            srv,
		lastHeartbeat:     map[string]map[boardgame.PlayerIndex]time.Time{},
		absent:            map[string]map[boardgame.PlayerIndex]bool{},
		animationLaneTail: map[string]animationLaneEntry{},
	}
	return v
}

func TestMarkAbsentReturnsTrueOnTransition(t *testing.T) {
	v := newTestNotifier(t)
	first := v.markAbsent("game1", 2)
	assert.For(t).ThatActual(first).IsTrue()
	second := v.markAbsent("game1", 2)
	// Already absent — no transition.
	assert.For(t).ThatActual(second).IsFalse()
}

func TestClearAbsentIfPresent(t *testing.T) {
	v := newTestNotifier(t)
	v.markAbsent("game1", 2)
	cleared := v.clearAbsentIfPresent("game1", 2)
	assert.For(t).ThatActual(cleared).IsTrue()
	// Idempotent — clearing again is a no-op.
	cleared2 := v.clearAbsentIfPresent("game1", 2)
	assert.For(t).ThatActual(cleared2).IsFalse()
}

func TestAbsentPlayersReturnsSnapshot(t *testing.T) {
	v := newTestNotifier(t)
	v.markAbsent("game1", 0)
	v.markAbsent("game1", 2)
	v.markAbsent("game2", 1)

	g1 := v.AbsentPlayers("game1")
	assert.For(t).ThatActual(len(g1)).Equals(2)
	hasZero, hasTwo := false, false
	for _, pi := range g1 {
		if pi == 0 {
			hasZero = true
		}
		if pi == 2 {
			hasTwo = true
		}
	}
	assert.For(t).ThatActual(hasZero).IsTrue()
	assert.For(t).ThatActual(hasTwo).IsTrue()

	g2 := v.AbsentPlayers("game2")
	assert.For(t).ThatActual(len(g2)).Equals(1)
	assert.For(t).ThatActual(g2[0]).Equals(boardgame.PlayerIndex(1))

	// Unknown game returns nil.
	gNone := v.AbsentPlayers("never-seen")
	assert.For(t).ThatActual(len(gNone)).Equals(0)
}

func TestScanStaleHeartbeatsPromotesAndEvicts(t *testing.T) {
	v := newTestNotifier(t)

	// Game A: fresh heartbeat (within threshold) — should stay live.
	// Game B: stale heartbeat (older than absentThreshold) — should be
	//         marked absent.
	// Game C: very stale + no connected sockets — should be evicted.
	now := time.Now()
	v.lastHeartbeat["A"] = map[boardgame.PlayerIndex]time.Time{
		0: now.Add(-5 * time.Second), // fresh
	}
	v.lastHeartbeat["B"] = map[boardgame.PlayerIndex]time.Time{
		1: now.Add(-absentThreshold - 5*time.Second), // stale → absent
	}
	v.lastHeartbeat["C"] = map[boardgame.PlayerIndex]time.Time{
		0: now.Add(-10 * time.Minute), // very stale + no sockets → evict
	}
	// Sockets bucket: A has one (mocked); C has none.
	v.sockets["A"] = map[*socket]bool{{gameID: "A"}: true}
	// B has a socket too so it's just stale, not dormant.
	v.sockets["B"] = map[*socket]bool{{gameID: "B"}: true}

	v.scanStaleHeartbeats()

	// A: still live, not absent.
	assert.For(t).ThatActual(len(v.AbsentPlayers("A"))).Equals(0)
	// B: marked absent.
	abs := v.AbsentPlayers("B")
	assert.For(t).ThatActual(len(abs)).Equals(1)
	assert.For(t).ThatActual(abs[0]).Equals(boardgame.PlayerIndex(1))
	// C: evicted from lastHeartbeat.
	_, hasC := v.lastHeartbeat["C"]
	assert.For(t).ThatActual(hasC).IsFalse()
}

func TestPresenceClearOnHeartbeatTransitionsBack(t *testing.T) {
	v := newTestNotifier(t)
	v.markAbsent("game1", 0)
	// A heartbeat arriving for an absent player should be a transition.
	cleared := v.clearAbsentIfPresent("game1", 0)
	assert.For(t).ThatActual(cleared).IsTrue()
	assert.For(t).ThatActual(len(v.AbsentPlayers("game1"))).Equals(0)
}
