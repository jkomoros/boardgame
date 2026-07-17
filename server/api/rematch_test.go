package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame/server/api/extendedgame"
	"github.com/jkomoros/boardgame/server/api/tablelease"
	"github.com/jkomoros/boardgame/server/api/users"
)

func TestRematchAuthorityAcceptsOwnerOrLastTableAfterExpiry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const gameID = "finished-game"
	key := []byte("0123456789abcdef0123456789abcdef")
	credentials := &Server{tableLeaseKey: key}
	deviceID, secret, digest, err := credentials.newTableLeaseCredential(gameID)
	if err != nil {
		t.Fatal(err)
	}
	storage := &hostLeaseStorage{lease: &tablelease.StorageRecord{
		DeviceID: deviceID, SecretDigest: digest,
		Expires: time.Now().Add(-24 * time.Hour).UnixMilli(),
	}}
	s := &Server{storage: NewServerStorageManager(storage)}
	info := &extendedgame.StorageRecord{Owner: "owner"}

	context := func(surface, credential string) *gin.Context {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodPost, "/api/game/x/y/rematch", nil)
		if surface != "" {
			req.AddCookie(&http.Cookie{Name: surfaceCookieName(gameID), Value: surface})
		}
		if credential != "" {
			req.AddCookie(&http.Cookie{Name: tableLeaseCookieName(gameID), Value: credential})
		}
		c.Request = req
		return c
	}

	owner, table := s.rematchAuthority(context("", ""), gameID, info, &users.StorageRecord{ID: "owner"})
	if !owner || table {
		t.Fatalf("owner/table = %t/%t; want true/false", owner, table)
	}
	owner, table = s.rematchAuthority(context("table", deviceID+"."+secret), gameID, info, nil)
	if owner || !table {
		t.Fatalf("expired exact Table owner/table = %t/%t; want false/true", owner, table)
	}
	for name, c := range map[string]*gin.Context{
		"Hand surface": context("hand", deviceID+"."+secret),
		"wrong secret": context("table", deviceID+".wrong"),
		"no surface":   context("", deviceID+"."+secret),
	} {
		owner, table = s.rematchAuthority(c, gameID, info, nil)
		if owner || table {
			t.Errorf("%s owner/table = %t/%t; want false/false", name, owner, table)
		}
	}
}

func TestBeginRematchFenceRecoversExpiredCreator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const gameID = "finished-game"
	key := []byte("0123456789abcdef0123456789abcdef")
	credentials := &Server{tableLeaseKey: key}
	deviceID, secret, digest, err := credentials.newTableLeaseCredential(gameID)
	if err != nil {
		t.Fatal(err)
	}
	storage := &actionFenceStorage{lease: &tablelease.StorageRecord{
		Generation: 4, DeviceID: deviceID, SecretDigest: digest,
		Expires:          time.Now().Add(-time.Minute).UnixMilli(),
		TransitionKind:   tablelease.TransitionHostAction,
		PreviousDeviceID: "stale-transfer-source",
	}}
	s := &Server{storage: NewServerStorageManager(storage)}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/game/x/y/rematch", nil)
	req.AddCookie(&http.Cookie{Name: surfaceCookieName(gameID), Value: "table"})
	req.AddCookie(&http.Cookie{Name: tableLeaseCookieName(gameID), Value: deviceID + "." + secret})
	c.Request = req

	fence, renewal := s.beginRematchFence(c, gameID, false)
	if renewal != tableLeaseRenewed || fence == nil {
		t.Fatalf("expired fence recovery = %v, %#v; want renewed fence", renewal, fence)
	}
	current, _ := storage.CompanionTableLease(gameID)
	if current.TransitionKind != tablelease.TransitionHostAction || current.Expires <= time.Now().UnixMilli() {
		t.Fatalf("recovered fence was not reacquired: %+v", current)
	}
	if !s.endRematchFence(gameID, fence) {
		t.Fatal("could not release recovered rematch fence")
	}
}
