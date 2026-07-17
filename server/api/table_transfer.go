package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame/server/api/tablelease"
	"github.com/skip2/go-qrcode"
)

const (
	tableTransferTTL          = 5 * time.Minute
	tableTransferMaxBodyBytes = 4 << 10
	tableTransferCodeLength   = 10
)

const crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

type tableTransferRequest struct {
	Token      string `json:"token"`
	RoomCode   string `json:"roomCode"`
	ManualCode string `json:"manualCode"`
	PairingID  string `json:"pairingID"`
	DeviceID   string `json:"deviceID"`
}

type tableTransferInput struct {
	gameID, pairingID, tokenSecret, manualCode string
}

func decodeStrictJSON(c *gin.Context, destination interface{}) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, tableTransferMaxBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request must contain exactly one JSON value")
		}
		return err
	}
	return nil
}

func (s *Server) transferMAC(domain string, values ...string) []byte {
	mac := hmac.New(sha256.New, s.tableLeaseKey)
	mac.Write([]byte(domain))
	for _, value := range values {
		mac.Write([]byte{0})
		mac.Write([]byte(value))
	}
	return mac.Sum(nil)
}

func (s *Server) transferTokenSecret(gameID, pairingID string) string {
	return hex.EncodeToString(s.transferMAC("boardgame-table-transfer-token", gameID, pairingID))
}

func (s *Server) transferDigest(domain, value string) string {
	return hex.EncodeToString(s.transferMAC(domain, value))
}

func (s *Server) transferManualCode(gameID, pairingID string) string {
	bits := s.transferMAC("boardgame-table-transfer-manual-code", gameID, pairingID)
	result := make([]byte, tableTransferCodeLength)
	// Ten independent five-bit symbols provide 50 bits without modulo bias.
	for i := range result {
		bit := i * 5
		word := uint16(bits[bit/8]) << 8
		if bit/8+1 < len(bits) {
			word |= uint16(bits[bit/8+1])
		}
		result[i] = crockfordAlphabet[(word>>uint(11-bit%8))&31]
	}
	return string(result)
}

func tableTransferToken(gameID, pairingID, secret string) string {
	encodedGame := base64.RawURLEncoding.EncodeToString([]byte(gameID))
	return "v1." + encodedGame + "." + pairingID + "." + secret
}

func parseTableTransferToken(token string) (gameID, pairingID, secret string, ok bool) {
	if len(token) == 0 || len(token) > 512 {
		return "", "", "", false
	}
	parts := strings.Split(token, ".")
	if len(parts) != 4 || parts[0] != "v1" || !validLowerHex(parts[2], 32) || !validLowerHex(parts[3], 64) {
		return "", "", "", false
	}
	decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(decoded) == 0 || len(decoded) > 128 || strings.ContainsRune(string(decoded), 0) {
		return "", "", "", false
	}
	return string(decoded), parts[2], parts[3], true
}

func validLowerHex(value string, length int) bool {
	if len(value) != length || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func normalizeTransferCode(value string) (string, bool) {
	value = strings.ToUpper(strings.NewReplacer("-", "", " ", "").Replace(value))
	if len(value) != tableTransferCodeLength {
		return "", false
	}
	for _, char := range value {
		if !strings.ContainsRune(crockfordAlphabet, char) {
			return "", false
		}
	}
	return value, true
}

func constantStringEqual(left, right string) bool {
	return len(left) == len(right) && subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func tableTransferProblem(c *gin.Context, status int, code, message string) {
	c.Header("Cache-Control", "no-store")
	c.JSON(status, gin.H{"code": code, "error": message})
}

func (s *Server) tableTransferClaimOrigin(c *gin.Context) (string, bool) {
	origin := strings.TrimSpace(c.GetHeader("Origin"))
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", false
	}
	if s.config == nil || !s.config.OriginAllowed(origin) {
		return "", false
	}
	return strings.TrimSuffix(origin, "/"), true
}

func (s *Server) tableTransferOffer(c *gin.Context, origin, gameID, pairingID string, expires int64) {
	tokenSecret := s.transferTokenSecret(gameID, pairingID)
	token := tableTransferToken(gameID, pairingID, tokenSecret)
	manualCode := s.transferManualCode(gameID, pairingID)
	claimURL := origin + "/table#transfer=" + url.QueryEscape(token)
	png, err := qrcode.Encode(claimURL, qrcode.Medium, 320)
	if err != nil {
		tableTransferProblem(c, http.StatusInternalServerError, "TABLE_TRANSFER_QR", "could not create the transfer QR code")
		return
	}
	now := time.Now().UnixMilli()
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"ok": true, "pairingID": pairingID, "token": token,
		"manualCode": manualCode, "claimURL": claimURL,
		"qrDataURL":   "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
		"expiresAtMs": expires, "serverNowMs": now,
	})
}

// createTableTransferHandler creates or resumes one transfer offer without
// interrupting the current Table. The transfer becomes authoritative only at
// redemption's lease CAS.
func (s *Server) createTableTransferHandler(c *gin.Context) {
	var body struct{}
	if err := decodeStrictJSON(c, &body); err != nil {
		tableTransferProblem(c, http.StatusBadRequest, "TABLE_TRANSFER_INVALID_REQUEST", "the request body must be an empty JSON object")
		return
	}
	origin, originAllowed := s.tableTransferClaimOrigin(c)
	if !originAllowed {
		tableTransferProblem(c, http.StatusForbidden, "TABLE_TRANSFER_ORIGIN", "the shared Table origin is not allowed")
		return
	}
	game := s.getGame(c)
	if game == nil {
		tableTransferProblem(c, http.StatusNotFound, "GAME_NOT_FOUND", "no such game")
		return
	}
	gameID := game.ID()
	if game.Finished() {
		tableTransferProblem(c, http.StatusConflict, "GAME_FINISHED", "the game is finished")
		return
	}
	eGame, err := s.storage.ExtendedGame(gameID)
	if err != nil || eGame == nil || eGame.CompanionRoomCode == "" {
		tableTransferProblem(c, http.StatusConflict, "GAME_NOT_COMPANION", "this game has no shared Table")
		return
	}
	credential, cookieErr := c.Cookie(tableLeaseCookieName(gameID))
	for attempts := 0; attempts < 5; attempts++ {
		now := time.Now()
		current, err := s.storage.CompanionTableLease(gameID)
		if err != nil {
			tableTransferProblem(c, http.StatusInternalServerError, "TABLE_TRANSFER_STORAGE", "could not inspect the shared Table")
			return
		}
		if cookieErr != nil || !tableLeaseActive(current, now) || !tableLeaseCredentialMatches(current, credential) {
			tableTransferProblem(c, http.StatusForbidden, "TABLE_LEASE_LOST", "this screen no longer controls the shared Table")
			return
		}
		if current.TransitionKind == tablelease.TransitionSolo {
			tableTransferProblem(c, http.StatusConflict, "GAME_NOT_COMPANION", "this game is switching to solo mode")
			return
		}
		if current.TransferPending(now.UnixMilli()) {
			s.tableTransferOffer(c, origin, gameID, current.TransferID, current.TransferExpires)
			return
		}
		pairingID, err := randomHex(16)
		if err != nil {
			tableTransferProblem(c, http.StatusInternalServerError, "TABLE_TRANSFER_RANDOM", "could not create the transfer")
			return
		}
		replacement := current.Clone()
		replacement.ClearTransfer()
		replacement.TransferID = pairingID
		replacement.TransferTokenDigest = s.transferDigest("boardgame-table-transfer-token-digest", s.transferTokenSecret(gameID, pairingID))
		replacement.TransferCodeDigest = s.transferDigest("boardgame-table-transfer-code-digest", s.transferManualCode(gameID, pairingID))
		replacement.TransferExpires = now.Add(tableTransferTTL).UnixMilli()
		replacement.Expires = now.Add(tableLeaseTTL).UnixMilli()
		replacement.PreviousDeviceID = ""
		replacement.TransitionKind = ""
		_, swapped, err := s.storage.CompareAndSwapCompanionTableLease(gameID, current.Generation, replacement)
		if err != nil {
			tableTransferProblem(c, http.StatusInternalServerError, "TABLE_TRANSFER_STORAGE", "could not create the transfer")
			return
		}
		if swapped {
			s.tableTransferOffer(c, origin, gameID, pairingID, replacement.TransferExpires)
			return
		}
	}
	tableTransferProblem(c, http.StatusConflict, "TABLE_TRANSFER_CONTENTION", "the shared Table changed; try again")
}

func (s *Server) cancelTableTransferHandler(c *gin.Context) {
	var body tableTransferRequest
	if err := decodeStrictJSON(c, &body); err != nil || !validLowerHex(body.PairingID, 32) || body.Token != "" || body.RoomCode != "" || body.ManualCode != "" || body.DeviceID != "" {
		tableTransferProblem(c, http.StatusBadRequest, "TABLE_TRANSFER_INVALID_REQUEST", "a valid pairingID is required")
		return
	}
	game := s.getGame(c)
	if game == nil {
		tableTransferProblem(c, http.StatusNotFound, "GAME_NOT_FOUND", "no such game")
		return
	}
	gameID := game.ID()
	credential, err := c.Cookie(tableLeaseCookieName(gameID))
	if err != nil {
		tableTransferProblem(c, http.StatusForbidden, "TABLE_LEASE_LOST", "this screen no longer controls the shared Table")
		return
	}
	for attempts := 0; attempts < 5; attempts++ {
		current, err := s.storage.CompanionTableLease(gameID)
		if err != nil {
			tableTransferProblem(c, http.StatusInternalServerError, "TABLE_TRANSFER_STORAGE", "could not cancel the transfer")
			return
		}
		if !tableLeaseActive(current, time.Now()) || !tableLeaseCredentialMatches(current, credential) {
			tableTransferProblem(c, http.StatusForbidden, "TABLE_LEASE_LOST", "this screen no longer controls the shared Table")
			return
		}
		if current.TransferID == "" {
			c.Header("Cache-Control", "no-store")
			c.JSON(http.StatusOK, gin.H{"ok": true})
			return
		}
		if current.TransferID != body.PairingID {
			tableTransferProblem(c, http.StatusConflict, "TABLE_TRANSFER_REPLACED", "a newer transfer is active")
			return
		}
		replacement := current.Clone()
		replacement.ClearTransfer()
		_, swapped, err := s.storage.CompareAndSwapCompanionTableLease(gameID, current.Generation, replacement)
		if err != nil {
			tableTransferProblem(c, http.StatusInternalServerError, "TABLE_TRANSFER_STORAGE", "could not cancel the transfer")
			return
		}
		if swapped {
			c.Header("Cache-Control", "no-store")
			c.JSON(http.StatusOK, gin.H{"ok": true})
			return
		}
	}
	tableTransferProblem(c, http.StatusConflict, "TABLE_TRANSFER_CONTENTION", "the shared Table changed; try again")
}

func (s *Server) resolveTableTransferRequest(req tableTransferRequest) (*tableTransferInput, *tablelease.StorageRecord, int, string, string) {
	var input tableTransferInput
	if req.Token != "" && req.RoomCode == "" && req.ManualCode == "" {
		var ok bool
		input.gameID, input.pairingID, input.tokenSecret, ok = parseTableTransferToken(req.Token)
		if !ok {
			return nil, nil, http.StatusNotFound, "TABLE_TRANSFER_INVALID", "that transfer is not valid"
		}
	} else if req.Token == "" && req.RoomCode != "" && req.ManualCode != "" {
		roomCode, err := NormalizeRoomCode(req.RoomCode)
		if err != nil {
			return nil, nil, http.StatusNotFound, "TABLE_TRANSFER_INVALID", "that transfer is not valid"
		}
		manualCode, ok := normalizeTransferCode(req.ManualCode)
		if !ok {
			return nil, nil, http.StatusNotFound, "TABLE_TRANSFER_INVALID", "that transfer is not valid"
		}
		input.gameID, err = s.storage.GameByRoomCode(roomCode)
		if err != nil {
			return nil, nil, http.StatusInternalServerError, "TABLE_TRANSFER_STORAGE", "could not inspect the transfer"
		}
		if input.gameID == "" {
			return nil, nil, http.StatusNotFound, "TABLE_TRANSFER_INVALID", "that transfer is not valid"
		}
		input.manualCode = manualCode
	} else {
		return nil, nil, http.StatusBadRequest, "TABLE_TRANSFER_INVALID_REQUEST", "provide either a transfer token or both manual codes"
	}
	record, err := s.storage.CompanionTableLease(input.gameID)
	if err != nil {
		return nil, nil, http.StatusInternalServerError, "TABLE_TRANSFER_STORAGE", "could not inspect the transfer"
	}
	if record == nil || record.TransferID == "" || record.ValidateTransfer() != nil {
		return nil, nil, http.StatusNotFound, "TABLE_TRANSFER_INVALID", "that transfer is not valid"
	}
	if input.pairingID == "" {
		input.pairingID = record.TransferID
		expected := s.transferDigest("boardgame-table-transfer-code-digest", input.manualCode)
		if !constantStringEqual(expected, record.TransferCodeDigest) {
			return nil, nil, http.StatusNotFound, "TABLE_TRANSFER_INVALID", "that transfer is not valid"
		}
	} else {
		expected := s.transferDigest("boardgame-table-transfer-token-digest", input.tokenSecret)
		if !constantStringEqual(input.pairingID, record.TransferID) || !constantStringEqual(expected, record.TransferTokenDigest) {
			return nil, nil, http.StatusNotFound, "TABLE_TRANSFER_INVALID", "that transfer is not valid"
		}
	}
	if record.TransferExpires <= time.Now().UnixMilli() {
		return nil, nil, http.StatusGone, "TABLE_TRANSFER_EXPIRED", "that transfer has expired"
	}
	return &input, record, 0, "", ""
}

func (s *Server) inspectTableTransferHandler(c *gin.Context) {
	var req tableTransferRequest
	if err := decodeStrictJSON(c, &req); err != nil || req.PairingID != "" || req.DeviceID != "" {
		tableTransferProblem(c, http.StatusBadRequest, "TABLE_TRANSFER_INVALID_REQUEST", "the transfer request is malformed")
		return
	}
	input, record, status, code, message := s.resolveTableTransferRequest(req)
	if status != 0 {
		tableTransferProblem(c, status, code, message)
		return
	}
	combined, err := s.storage.CombinedGame(input.gameID)
	if err != nil || combined == nil || combined.Finished || combined.CompanionRoomCode == "" {
		tableTransferProblem(c, http.StatusConflict, "GAME_FINISHED", "the game is no longer available")
		return
	}
	manager := s.managers[combined.Name]
	if manager == nil || manager.manager == nil {
		tableTransferProblem(c, http.StatusNotFound, "TABLE_TRANSFER_INVALID", "that transfer is not valid")
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"ok": true, "pairingID": input.pairingID, "gameID": input.gameID,
		"gameName": combined.Name, "gameDisplayName": manager.manager.Delegate().DisplayName(),
		"expiresAtMs": record.TransferExpires, "serverNowMs": time.Now().UnixMilli(),
		"alreadyRedeemed": record.TransferTargetDeviceID != "",
	})
}

func (s *Server) redeemTableTransferHandler(c *gin.Context) {
	var req tableTransferRequest
	if err := decodeStrictJSON(c, &req); err != nil || !validLowerHex(req.PairingID, 32) || !validLowerHex(req.DeviceID, 32) {
		tableTransferProblem(c, http.StatusBadRequest, "TABLE_TRANSFER_INVALID_REQUEST", "valid pairing and device IDs are required")
		return
	}
	input, _, status, code, message := s.resolveTableTransferRequest(req)
	if status != 0 {
		tableTransferProblem(c, status, code, message)
		return
	}
	if !constantStringEqual(req.PairingID, input.pairingID) {
		tableTransferProblem(c, http.StatusNotFound, "TABLE_TRANSFER_INVALID", "that transfer is not valid")
		return
	}
	combined, err := s.storage.CombinedGame(input.gameID)
	if err != nil || combined == nil || combined.Finished || combined.CompanionRoomCode == "" {
		tableTransferProblem(c, http.StatusConflict, "GAME_FINISHED", "the game is no longer available")
		return
	}
	secret, digest, err := s.tableLeaseCredentialForDevice(input.gameID, req.DeviceID)
	if err != nil {
		tableTransferProblem(c, http.StatusInternalServerError, "TABLE_TRANSFER_CREDENTIAL", "could not connect this screen")
		return
	}
	for attempts := 0; attempts < 5; attempts++ {
		current, status, code, message := func() (*tablelease.StorageRecord, int, string, string) {
			_, rec, st, cd, msg := s.resolveTableTransferRequest(req)
			return rec, st, cd, msg
		}()
		if status != 0 {
			tableTransferProblem(c, status, code, message)
			return
		}
		if current.TransferTargetDeviceID != "" {
			if current.TransferTargetDeviceID != req.DeviceID {
				tableTransferProblem(c, http.StatusConflict, "TABLE_TRANSFER_ALREADY_REDEEMED", "that transfer was already used by another screen")
				return
			}
			// The first response may have been lost before the receiver could
			// navigate and open a renewing socket. Re-fence the exact same device
			// through CAS and restore a full lease window before reissuing cookies.
			replacement := current.Clone()
			replacement.Expires = time.Now().Add(tableLeaseTTL).UnixMilli()
			_, swapped, err := s.storage.CompareAndSwapCompanionTableLease(input.gameID, current.Generation, replacement)
			if err != nil {
				tableTransferProblem(c, http.StatusInternalServerError, "TABLE_TRANSFER_STORAGE", "could not reconnect this screen")
				return
			}
			if !swapped {
				continue
			}
			s.setTableLeaseCookie(c, input.gameID, req.DeviceID, secret)
			s.setSurfaceCookie(c, input.gameID, "table")
			s.notifier.broadcastTableSessionChanged(input.gameID)
			c.Header("Cache-Control", "no-store")
			c.JSON(http.StatusOK, gin.H{"ok": true, "gameID": input.gameID, "gameName": combined.Name, "gameURL": fmt.Sprintf("/game/%s/%s", url.PathEscape(combined.Name), url.PathEscape(input.gameID))})
			return
		}
		now := time.Now()
		replacement := current.Clone()
		replacement.PreviousDeviceID = current.DeviceID
		replacement.DeviceID = req.DeviceID
		replacement.SecretDigest = digest
		replacement.Expires = now.Add(tableLeaseTTL).UnixMilli()
		replacement.TransferTargetDeviceID = req.DeviceID
		replacement.TransitionKind = tablelease.TransitionTransfer
		stored, swapped, err := s.storage.CompareAndSwapCompanionTableLease(input.gameID, current.Generation, replacement)
		if err != nil {
			tableTransferProblem(c, http.StatusInternalServerError, "TABLE_TRANSFER_STORAGE", "could not connect this screen")
			return
		}
		if !swapped {
			_ = stored
			continue
		}
		s.setTableLeaseCookie(c, input.gameID, req.DeviceID, secret)
		s.setSurfaceCookie(c, input.gameID, "table")
		s.notifier.broadcastTableSessionChanged(input.gameID)
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, gin.H{"ok": true, "gameID": input.gameID, "gameName": combined.Name, "gameURL": fmt.Sprintf("/game/%s/%s", url.PathEscape(combined.Name), url.PathEscape(input.gameID))})
		return
	}
	tableTransferProblem(c, http.StatusConflict, "TABLE_TRANSFER_CONTENTION", "the shared Table changed; try again")
}
