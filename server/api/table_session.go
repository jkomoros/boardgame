package api

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame/server/api/tablelease"
)

const tableLeaseTTL = 40 * time.Second

func tableLeaseCookieName(gameID string) string { return "table_lease_" + gameID }

func randomHex(bytes int) (string, error) {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func digestTableLeaseSecret(secret string) string {
	digest := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(digest[:])
}

func (s *Server) tableLeaseCredentialForDevice(gameID, deviceID string) (secret, digest string, err error) {
	if len(s.tableLeaseKey) < 16 || gameID == "" || len(deviceID) != 32 {
		return "", "", fmt.Errorf("Table lease credential key or identity is invalid")
	}
	mac := hmac.New(sha256.New, s.tableLeaseKey)
	mac.Write([]byte("boardgame-table-lease\x00"))
	mac.Write([]byte(gameID))
	mac.Write([]byte("\x00"))
	mac.Write([]byte(deviceID))
	secret = hex.EncodeToString(mac.Sum(nil))
	return secret, digestTableLeaseSecret(secret), nil
}

func (s *Server) newTableLeaseCredential(gameID string) (deviceID, secret, digest string, err error) {
	deviceID, err = randomHex(16)
	if err != nil {
		return "", "", "", err
	}
	secret, digest, err = s.tableLeaseCredentialForDevice(gameID, deviceID)
	return deviceID, secret, digest, err
}

func parseTableLeaseCredential(value string) (deviceID, secret string, ok bool) {
	if len(value) > 256 {
		return "", "", false
	}
	parts := strings.Split(value, ".")
	if len(parts) != 2 || len(parts[0]) != 32 || len(parts[1]) != 64 {
		return "", "", false
	}
	if _, err := hex.DecodeString(parts[0]); err != nil {
		return "", "", false
	}
	if _, err := hex.DecodeString(parts[1]); err != nil {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func tableLeaseCredentialMatches(record *tablelease.StorageRecord, value string) bool {
	if record == nil || record.DeviceID == "" || record.SecretDigest == "" {
		return false
	}
	deviceID, secret, ok := parseTableLeaseCredential(value)
	if !ok || subtle.ConstantTimeCompare([]byte(deviceID), []byte(record.DeviceID)) != 1 {
		return false
	}
	digest := digestTableLeaseSecret(secret)
	return subtle.ConstantTimeCompare([]byte(digest), []byte(record.SecretDigest)) == 1
}

func (s *Server) setTableLeaseCookie(c *gin.Context, gameID, deviceID, secret string) {
	secure := !s.config.OfflineDevMode
	c.SetCookie(tableLeaseCookieName(gameID), deviceID+"."+secret, int((30*24*time.Hour)/time.Second), "/", "", secure, true)
}

func clearTableLeaseCookie(c *gin.Context, gameID string) {
	c.SetCookie(tableLeaseCookieName(gameID), "", -1, "/", "", false, true)
}

func tableLeaseActive(record *tablelease.StorageRecord, now time.Time) bool {
	return record != nil && record.DeviceID != "" && record.SecretDigest != "" && record.Expires > now.UnixMilli()
}

func tableSoloTransitionActive(record *tablelease.StorageRecord, now time.Time) bool {
	return record != nil && record.TransitionKind == tablelease.TransitionSolo && tableLeaseActive(record, now)
}

// tableSurfaceForRequest identifies a browser that has explicitly been put on
// the shared/public surface. Unlike host authority, this deliberately does not
// depend on a live credential: expired, displaced, and tampered Tables must
// still fail closed to observer-only state while recovery UI is shown.
func tableSurfaceForRequest(c *gin.Context, gameID string) bool {
	surface, err := c.Cookie(surfaceCookieName(gameID))
	return err == nil && surface == "table"
}

// activeTableLeaseForRequest is the sole authority check for companion host
// actions. Renderer-selection state is deliberately irrelevant here.
func (s *Server) activeTableLeaseForRequest(c *gin.Context, gameID string) (*tablelease.StorageRecord, bool) {
	if s == nil || s.storage == nil {
		return nil, false
	}
	// A lease is a capability for the shared screen, not a sticky privilege
	// for this browser. Joining as a Hand overwrites the surface cookie and
	// immediately prevents that socket from keeping a zombie Table alive.
	if !tableSurfaceForRequest(c, gameID) {
		return nil, false
	}
	eGame, err := s.storage.ExtendedGame(gameID)
	if err != nil {
		return nil, false
	}
	if eGame == nil || eGame.CompanionRoomCode == "" {
		return nil, false
	}
	record, err := s.storage.CompanionTableLease(gameID)
	if err != nil || !tableLeaseActive(record, time.Now()) {
		return record, false
	}
	credential, err := c.Cookie(tableLeaseCookieName(gameID))
	if err != nil || !tableLeaseCredentialMatches(record, credential) {
		return record, false
	}
	return record, true
}

func (s *Server) tableLeaseEligible(gameID, userID string) bool {
	eGame, err := s.storage.ExtendedGame(gameID)
	if err == nil && eGame != nil && eGame.Owner == userID {
		return true
	}
	for _, seatedUserID := range s.storage.UserIDsForGame(gameID) {
		if seatedUserID == userID {
			return true
		}
	}
	return false
}

func tableLeaseProblem(c *gin.Context, status int, code, message string, extra gin.H) {
	body := gin.H{"code": code, "error": message}
	for key, value := range extra {
		body[key] = value
	}
	c.JSON(status, body)
}

// acquireTableLeaseHandler atomically resumes this browser's lease or, once
// the old Table lease expires, gives one eligible in-person participant a new
// fenced credential. The storage CAS—not a process mutex—selects the winner.
func (s *Server) acquireTableLeaseHandler(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4<<10)
	var body struct {
		DeviceID string `json:"deviceID"`
	}
	if err := c.BindJSON(&body); err != nil {
		tableLeaseProblem(c, http.StatusBadRequest, "TABLE_LEASE_INVALID_REQUEST", "a recovery device ID is required", nil)
		return
	}
	if len(body.DeviceID) != 32 {
		tableLeaseProblem(c, http.StatusBadRequest, "TABLE_LEASE_INVALID_REQUEST", "invalid recovery device ID", nil)
		return
	}
	if _, err := hex.DecodeString(body.DeviceID); err != nil {
		tableLeaseProblem(c, http.StatusBadRequest, "TABLE_LEASE_INVALID_REQUEST", "invalid recovery device ID", nil)
		return
	}
	game := s.getGame(c)
	if game == nil {
		tableLeaseProblem(c, http.StatusNotFound, "GAME_NOT_FOUND", "no such game", nil)
		return
	}
	gameID := game.ID()
	eGame, err := s.storage.ExtendedGame(gameID)
	if err != nil {
		tableLeaseProblem(c, http.StatusInternalServerError, "TABLE_LEASE_STORAGE", "could not inspect the game session", nil)
		return
	}
	if eGame == nil || eGame.CompanionRoomCode == "" {
		tableLeaseProblem(c, http.StatusConflict, "GAME_NOT_COMPANION", "this game has no shared screen", nil)
		return
	}
	if game.Finished() {
		tableLeaseProblem(c, http.StatusConflict, "GAME_FINISHED", "the game is finished", nil)
		return
	}
	user := s.getUser(c)
	if user == nil || !s.tableLeaseEligible(gameID, user.ID) {
		tableLeaseProblem(c, http.StatusForbidden, "TABLE_LEASE_NOT_ELIGIBLE", "only the owner or a seated player can restore the shared screen", nil)
		return
	}

	credential, _ := c.Cookie(tableLeaseCookieName(gameID))
	for attempts := 0; attempts < 5; attempts++ {
		now := time.Now()
		current, err := s.storage.CompanionTableLease(gameID)
		if err != nil {
			tableLeaseProblem(c, http.StatusInternalServerError, "TABLE_LEASE_STORAGE", "could not inspect the shared screen", nil)
			return
		}
		if tableSoloTransitionActive(current, now) {
			tableLeaseProblem(c, http.StatusConflict, "GAME_NOT_COMPANION", "this game is switching to solo mode", nil)
			return
		}
		credentialHeld := tableLeaseActive(current, now) && tableLeaseCredentialMatches(current, credential)
		idempotentRetry := tableLeaseActive(current, now) && current.DeviceID == body.DeviceID && current.HolderUserID == user.ID
		alreadyHeld := credentialHeld || idempotentRetry
		if tableLeaseActive(current, now) && !alreadyHeld {
			retryAfter := current.Expires - now.UnixMilli()
			if retryAfter < 0 {
				retryAfter = 0
			}
			tableLeaseProblem(c, http.StatusConflict, "TABLE_LEASE_ACTIVE", "the shared screen is still connected", gin.H{"retryAfterMs": retryAfter})
			return
		}

		deviceID, secret, digest := "", "", ""
		if credentialHeld {
			deviceID = current.DeviceID
			_, secret, _ = parseTableLeaseCredential(credential)
			digest = current.SecretDigest
		} else {
			secret, digest, err = s.tableLeaseCredentialForDevice(gameID, body.DeviceID)
			deviceID = body.DeviceID
			if err != nil {
				tableLeaseProblem(c, http.StatusInternalServerError, "TABLE_LEASE_RANDOM", "could not create shared-screen credentials", nil)
				return
			}
		}
		replacement := &tablelease.StorageRecord{
			DeviceID: deviceID, SecretDigest: digest, HolderUserID: user.ID,
			Expires: now.Add(tableLeaseTTL).UnixMilli(),
		}
		if current != nil && current.DeviceID != deviceID && validLowerHex(current.DeviceID, 32) {
			replacement.PreviousDeviceID = current.DeviceID
			replacement.TransitionKind = tablelease.TransitionRecovery
		}
		expected := uint64(0)
		if current != nil {
			expected = current.Generation
		}
		stored, swapped, err := s.storage.CompareAndSwapCompanionTableLease(gameID, expected, replacement)
		if err != nil {
			tableLeaseProblem(c, http.StatusInternalServerError, "TABLE_LEASE_STORAGE", "could not restore the shared screen", nil)
			return
		}
		if !swapped {
			_ = stored
			continue
		}
		s.setTableLeaseCookie(c, gameID, deviceID, secret)
		s.setSurfaceCookie(c, gameID, "table")
		s.notifier.broadcastTableSessionChanged(gameID)
		c.JSON(http.StatusOK, gin.H{"ok": true, "alreadyHeld": alreadyHeld, "expiresAtMs": stored.Expires})
		return
	}
	tableLeaseProblem(c, http.StatusConflict, "TABLE_LEASE_ALREADY_RESTORED", "another player restored the shared screen", nil)
}

// renewTableLeaseCredential extends the exact fenced credential. It retries a
// benign CAS loss from two sockets belonging to the same Table, but can never
// revive a credential after another device has acquired the lease.
type tableLeaseRenewal int

const (
	tableLeaseRenewed tableLeaseRenewal = iota
	tableLeaseRenewRetryable
	tableLeaseRenewLost
)

func (s *Server) renewTableLeaseCredential(gameID, credential string) tableLeaseRenewal {
	eGame, err := s.storage.ExtendedGame(gameID)
	if err != nil {
		return tableLeaseRenewRetryable
	}
	if eGame == nil || eGame.CompanionRoomCode == "" {
		return tableLeaseRenewLost
	}
	for attempts := 0; attempts < 3; attempts++ {
		now := time.Now()
		current, err := s.storage.CompanionTableLease(gameID)
		if err != nil {
			return tableLeaseRenewRetryable
		}
		if !tableLeaseActive(current, now) || !tableLeaseCredentialMatches(current, credential) {
			return tableLeaseRenewLost
		}
		if current.TransitionKind == tablelease.TransitionSolo {
			return tableLeaseRenewLost
		}
		replacement := current.Clone()
		replacement.Expires = now.Add(tableLeaseTTL).UnixMilli()
		_, swapped, err := s.storage.CompareAndSwapCompanionTableLease(gameID, current.Generation, replacement)
		if err != nil {
			return tableLeaseRenewRetryable
		}
		if swapped {
			return tableLeaseRenewed
		}
	}
	// Contention from another socket holding the same credential is benign.
	// Preserve this socket's capability and let its next heartbeat retry.
	current, err := s.storage.CompanionTableLease(gameID)
	if err != nil {
		return tableLeaseRenewRetryable
	}
	if tableLeaseActive(current, time.Now()) && tableLeaseCredentialMatches(current, credential) {
		return tableLeaseRenewRetryable
	}
	return tableLeaseRenewLost
}

func (s *Server) releaseTableLease(gameID string, credential string) bool {
	for attempts := 0; attempts < 5; attempts++ {
		current, err := s.storage.CompanionTableLease(gameID)
		if err != nil || current == nil || !tableLeaseCredentialMatches(current, credential) {
			return false
		}
		_, swapped, err := s.storage.CompareAndSwapCompanionTableLease(gameID, current.Generation, &tablelease.StorageRecord{})
		if err != nil {
			return false
		}
		if swapped {
			return true
		}
	}
	return false
}

// refreshTableLeaseForAction closes the practical check-then-mutate window by
// conditionally renewing the exact credential immediately before a host
// mutation. A takeover cannot begin until the renewed TTL expires.
func (s *Server) refreshTableLeaseForAction(c *gin.Context, gameID string) tableLeaseRenewal {
	if _, ok := s.activeTableLeaseForRequest(c, gameID); !ok {
		return tableLeaseRenewLost
	}
	credential, err := c.Cookie(tableLeaseCookieName(gameID))
	if err != nil {
		return tableLeaseRenewLost
	}
	return s.renewTableLeaseCredential(gameID, credential)
}
