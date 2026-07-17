package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/tablelease"
)

func TestTableLeaseCredentialRoundTripAndTampering(t *testing.T) {
	s := &Server{tableLeaseKey: []byte("0123456789abcdef0123456789abcdef")}
	deviceID, secret, digest, err := s.newTableLeaseCredential("GAME")
	if err != nil {
		t.Fatal(err)
	}
	credential := deviceID + "." + secret
	record := &tablelease.StorageRecord{DeviceID: deviceID, SecretDigest: digest}
	if !tableLeaseCredentialMatches(record, credential) {
		t.Fatal("fresh credential did not match its digest")
	}
	tamperAt := func(input string, index int) string {
		result := []byte(input)
		result[index] = '0'
		if input[index] == '0' {
			result[index] = '1'
		}
		return string(result)
	}
	tests := []string{
		"", credential + "x", "not.hex", strings.Repeat("x", 257),
		tamperAt(credential, 0), tamperAt(credential, len(credential)-1),
	}
	for _, invalid := range tests {
		if tableLeaseCredentialMatches(record, invalid) {
			t.Fatalf("tampered credential unexpectedly matched: %q", invalid)
		}
	}
}

func TestTableLeaseCredentialIsStableForIdempotentDeviceRetries(t *testing.T) {
	s := &Server{tableLeaseKey: []byte("0123456789abcdef0123456789abcdef")}
	const deviceID = "0123456789abcdef0123456789abcdef"
	first, firstDigest, err := s.tableLeaseCredentialForDevice("GAME", deviceID)
	if err != nil {
		t.Fatal(err)
	}
	second, secondDigest, err := s.tableLeaseCredentialForDevice("GAME", deviceID)
	if err != nil {
		t.Fatal(err)
	}
	if first != second || firstDigest != secondDigest {
		t.Fatal("same game/device retry rotated its credential")
	}
	other, _, err := s.tableLeaseCredentialForDevice("OTHER", deviceID)
	if err != nil {
		t.Fatal(err)
	}
	if first == other {
		t.Fatal("credential was not bound to its game")
	}
}

func TestTableLeaseCredentialIgnoresJoinTicketRotationAcrossServers(t *testing.T) {
	sharedTableKey := []byte("table-key-shared-by-every-api-instance")
	oldServer := &Server{tableLeaseKey: sharedTableKey, joinTicketKey: []byte("old-join-key-old-join-key-old-key")}
	newServer := &Server{tableLeaseKey: sharedTableKey, joinTicketKey: []byte("new-join-key-new-join-key-new-key")}
	const deviceID = "fedcba9876543210fedcba9876543210"
	oldCredential, _, err := oldServer.tableLeaseCredentialForDevice("GAME", deviceID)
	if err != nil {
		t.Fatal(err)
	}
	newCredential, _, err := newServer.tableLeaseCredentialForDevice("GAME", deviceID)
	if err != nil {
		t.Fatal(err)
	}
	if oldCredential != newCredential {
		t.Fatal("join-ticket rotation changed a Table recovery credential")
	}
}

func TestTableLeaseActiveRequiresCredentialAndFutureExpiry(t *testing.T) {
	now := time.UnixMilli(10_000)
	record := &tablelease.StorageRecord{DeviceID: "device", SecretDigest: "digest", Expires: 10_001}
	if !tableLeaseActive(record, now) {
		t.Fatal("future complete lease should be active")
	}
	record.Expires = 10_000
	if tableLeaseActive(record, now) {
		t.Fatal("lease must expire at the exact boundary")
	}
	record.Expires = 10_001
	record.SecretDigest = ""
	if tableLeaseActive(record, now) {
		t.Fatal("lease without a secret digest must be inactive")
	}
}

func TestSoloTransitionBlocksOnlyUntilItsCrashRecoveryDeadline(t *testing.T) {
	now := time.UnixMilli(10_000)
	record := &tablelease.StorageRecord{
		DeviceID: "0123456789abcdef0123456789abcdef", SecretDigest: strings.Repeat("0", 64),
		Expires: 10_001, PreviousDeviceID: "0123456789abcdef0123456789abcdef",
		TransitionKind: tablelease.TransitionSolo,
	}
	if !tableSoloTransitionActive(record, now) {
		t.Fatal("live solo transition did not fence recovery")
	}
	record.Expires = now.UnixMilli()
	if tableSoloTransitionActive(record, now) {
		t.Fatal("expired solo transition permanently wedged companion recovery")
	}
}

func TestActiveTableAlwaysReceivesObserverState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	game, _ := newLegalLedgerGame(t)
	key := []byte("0123456789abcdef0123456789abcdef")
	credentialServer := &Server{tableLeaseKey: key}
	deviceID, secret, digest, err := credentialServer.newTableLeaseCredential(game.ID())
	if err != nil {
		t.Fatal(err)
	}
	storage := &hostLeaseStorage{
		extended: &extendedgame.StorageRecord{CompanionRoomCode: "ABCD"},
		lease: &tablelease.StorageRecord{
			DeviceID: deviceID, SecretDigest: digest, Expires: time.Now().Add(time.Minute).UnixMilli(),
		},
	}
	s := &Server{tableLeaseKey: key, storage: NewServerStorageManager(storage)}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/game/test/id/info?admin=1&player=0&autoCurrentPlayer=1", nil)
	req.AddCookie(&http.Cookie{Name: surfaceCookieName(game.ID()), Value: "table"})
	req.AddCookie(&http.Cookie{Name: tableLeaseCookieName(game.ID()), Value: deviceID + "." + secret})
	c.Request = req
	s.setGame(c, game)
	s.setAdminAllowed(c, true)
	s.setViewingAsPlayer(c, 0)

	if got := s.effectivePlayerIndex(c); got != boardgame.ObserverPlayerIndex {
		t.Fatalf("effective player = %d; want observer despite seated/admin overrides", got)
	}
	if s.effectiveAutoCurrentPlayer(c) {
		t.Fatal("active Table honored auto-current-player and could expose private state")
	}
}

func TestDisplacedOrExpiredTableStillReceivesObserverState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	game, _ := newLegalLedgerGame(t)
	storage := &hostLeaseStorage{
		extended: &extendedgame.StorageRecord{CompanionRoomCode: "ABCD"},
		lease: &tablelease.StorageRecord{
			DeviceID:     "fedcba9876543210fedcba9876543210",
			SecretDigest: strings.Repeat("0", 64), Expires: time.Now().Add(-time.Minute).UnixMilli(),
		},
	}
	s := &Server{tableLeaseKey: []byte("0123456789abcdef0123456789abcdef"), storage: NewServerStorageManager(storage)}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/game/test/id/info?admin=1&player=0&autoCurrentPlayer=1", nil)
	req.AddCookie(&http.Cookie{Name: surfaceCookieName(game.ID()), Value: "table"})
	req.AddCookie(&http.Cookie{Name: tableLeaseCookieName(game.ID()), Value: "tampered"})
	c.Request = req
	s.setGame(c, game)
	s.setAdminAllowed(c, true)
	s.setViewingAsPlayer(c, 0)

	if got := s.effectivePlayerIndex(c); got != boardgame.ObserverPlayerIndex {
		t.Fatalf("fenced Table effective player = %d; want observer", got)
	}
	if s.effectiveAutoCurrentPlayer(c) {
		t.Fatal("fenced Table honored auto-current-player and could expose private state")
	}
}
