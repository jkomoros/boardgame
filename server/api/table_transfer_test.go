package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame/server/api/tablelease"
)

type transferTestStorage struct {
	StorageManager
	gameID   string
	roomCode string
	lease    *tablelease.StorageRecord
}

func (s *transferTestStorage) GameByRoomCode(code string) (string, error) {
	if code == s.roomCode {
		return s.gameID, nil
	}
	return "", nil
}

func (s *transferTestStorage) CompanionTableLease(gameID string) (*tablelease.StorageRecord, error) {
	if gameID != s.gameID {
		return nil, nil
	}
	return s.lease.Clone(), nil
}

func transferTestServer(t *testing.T) (*Server, string, string, string) {
	t.Helper()
	s := &Server{tableLeaseKey: []byte("0123456789abcdef0123456789abcdef")}
	const gameID = "game.with punctuation/and spaces"
	const pairingID = "0123456789abcdef0123456789abcdef"
	tokenSecret := s.transferTokenSecret(gameID, pairingID)
	manualCode := s.transferManualCode(gameID, pairingID)
	record := &tablelease.StorageRecord{
		TransferID:          pairingID,
		TransferTokenDigest: s.transferDigest("boardgame-table-transfer-token-digest", tokenSecret),
		TransferCodeDigest:  s.transferDigest("boardgame-table-transfer-code-digest", manualCode),
		TransferExpires:     time.Now().Add(time.Minute).UnixMilli(),
	}
	storage := &transferTestStorage{gameID: gameID, roomCode: "ABCD", lease: record}
	s.storage = NewServerStorageManager(storage)
	return s, gameID, pairingID, manualCode
}

func TestTableTransferTokenRoundTripIsGameBoundAndDeterministic(t *testing.T) {
	s, gameID, pairingID, _ := transferTestServer(t)
	secret := s.transferTokenSecret(gameID, pairingID)
	token := tableTransferToken(gameID, pairingID, secret)
	parsedGame, parsedPairing, parsedSecret, ok := parseTableTransferToken(token)
	if !ok || parsedGame != gameID || parsedPairing != pairingID || parsedSecret != secret {
		t.Fatalf("token did not round-trip: %q %q %t", parsedGame, parsedPairing, ok)
	}
	if secret != s.transferTokenSecret(gameID, pairingID) {
		t.Fatal("retry changed the deterministic token secret")
	}
	if secret == s.transferTokenSecret(gameID+"other", pairingID) {
		t.Fatal("token secret was not bound to its game")
	}
	for _, invalid := range []string{"", token + "x", strings.Replace(token, "v1.", "v2.", 1), strings.Repeat("x", 513)} {
		if _, _, _, ok := parseTableTransferToken(invalid); ok {
			t.Fatalf("accepted malformed token %q", invalid)
		}
	}
}

func TestTableTransferManualCodeIsUnambiguousAndNormalizesPresentation(t *testing.T) {
	s, _, _, code := transferTestServer(t)
	if len(code) != tableTransferCodeLength {
		t.Fatalf("manual code length = %d", len(code))
	}
	for _, ambiguous := range "ILOU" {
		if strings.ContainsRune(code, ambiguous) {
			t.Fatalf("manual code contains ambiguous character %q", ambiguous)
		}
	}
	formatted := strings.ToLower(code[:5] + "-" + code[5:])
	if normalized, ok := normalizeTransferCode(formatted); !ok || normalized != code {
		t.Fatalf("normalizeTransferCode(%q) = %q, %t", formatted, normalized, ok)
	}
	if _, ok := normalizeTransferCode(code[:9] + "I"); ok {
		t.Fatal("accepted ambiguous manual-code character")
	}
	if code != s.transferManualCode("game.with punctuation/and spaces", "0123456789abcdef0123456789abcdef") {
		t.Fatal("manual code changed across an idempotent retry")
	}
}

func TestResolveTableTransferAcceptsEitherCapabilityAndRejectsTampering(t *testing.T) {
	s, gameID, pairingID, manualCode := transferTestServer(t)
	token := tableTransferToken(gameID, pairingID, s.transferTokenSecret(gameID, pairingID))
	for _, request := range []tableTransferRequest{
		{Token: token},
		{RoomCode: " abcd ", ManualCode: strings.ToLower(manualCode[:5] + "-" + manualCode[5:])},
	} {
		input, _, status, _, _ := s.resolveTableTransferRequest(request)
		if status != 0 || input == nil || input.gameID != gameID || input.pairingID != pairingID {
			t.Fatalf("valid request failed: input=%+v status=%d", input, status)
		}
	}

	tamperedToken := token[:len(token)-1] + "0"
	if tamperedToken == token {
		tamperedToken = token[:len(token)-1] + "1"
	}
	tamperedCode := manualCode[:9] + "0"
	if tamperedCode == manualCode {
		tamperedCode = manualCode[:9] + "1"
	}
	for _, request := range []tableTransferRequest{
		{Token: tamperedToken},
		{RoomCode: "ABCD", ManualCode: tamperedCode},
		{Token: token, RoomCode: "ABCD", ManualCode: manualCode},
	} {
		if _, _, status, code, _ := s.resolveTableTransferRequest(request); status == 0 || code == "" {
			t.Fatalf("tampered/ambiguous request unexpectedly passed: %+v", request)
		}
	}
}

func TestResolveTableTransferRetainsBoundedRedeemReceipt(t *testing.T) {
	s, gameID, pairingID, _ := transferTestServer(t)
	storage := s.storage.StorageManager.(*transferTestStorage)
	storage.lease.PreviousDeviceID = "0123456789abcdef0123456789abcdef"
	storage.lease.DeviceID = "fedcba9876543210fedcba9876543210"
	storage.lease.TransferTargetDeviceID = storage.lease.DeviceID
	storage.lease.TransitionKind = tablelease.TransitionTransfer
	token := tableTransferToken(gameID, pairingID, s.transferTokenSecret(gameID, pairingID))
	if _, record, status, _, _ := s.resolveTableTransferRequest(tableTransferRequest{Token: token}); status != 0 || record.TransferTargetDeviceID == "" {
		t.Fatal("valid redeemed receipt was not available for a lost-response retry")
	}
	storage.lease.TransferExpires = time.Now().UnixMilli()
	if _, _, status, code, _ := s.resolveTableTransferRequest(tableTransferRequest{Token: token}); status != http.StatusGone || code != "TABLE_TRANSFER_EXPIRED" {
		t.Fatalf("expired receipt status/code = %d/%q", status, code)
	}
}

func TestDecodeStrictJSONRejectsUnknownTrailingAndOversizedBodies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, body := range []string{
		`{"pairingID":"x","unknown":true}`,
		`{"pairingID":"x"} {}`,
		`{"pairingID":"` + strings.Repeat("x", tableTransferMaxBodyBytes) + `"}`,
	} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/table-transfer/redeem", strings.NewReader(body))
		var request tableTransferRequest
		if err := decodeStrictJSON(c, &request); err == nil {
			t.Fatalf("accepted malformed body of length %d", len(body))
		}
	}
}
