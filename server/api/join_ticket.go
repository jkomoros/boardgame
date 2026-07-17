package api

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	joinTicketHeader = "X-Boardgame-Join-Ticket"
	joinTicketTTL    = 10 * time.Minute
)

type joinTicketClaims struct {
	GameID  string `json:"gameID"`
	Expires int64  `json:"expires"`
}

func newJoinTicketKey() []byte {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic("could not initialize join-ticket signing key: " + err.Error())
	}
	return key
}

func joinTicketKeyFromSecret(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return append([]byte(nil), sum[:]...)
}

func (s *Server) issueJoinTicket(gameID string, now time.Time) (string, error) {
	claims := joinTicketClaims{GameID: gameID, Expires: now.Add(joinTicketTTL).Unix()}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, s.joinTicketKey)
	_, _ = mac.Write([]byte(encoded))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encoded + "." + signature, nil
}

func (s *Server) verifyJoinTicket(ticket, gameID string, now time.Time) error {
	if len(ticket) > 4096 {
		return errors.New("join ticket is too large")
	}
	parts := strings.Split(ticket, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return errors.New("malformed join ticket")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return errors.New("malformed join ticket signature")
	}
	validSignature := false
	for _, key := range [][]byte{s.joinTicketKey, s.joinTicketPreviousKey} {
		if len(key) == 0 {
			continue
		}
		mac := hmac.New(sha256.New, key)
		_, _ = mac.Write([]byte(parts[0]))
		validSignature = validSignature || hmac.Equal(signature, mac.Sum(nil))
	}
	if !validSignature {
		return errors.New("invalid join ticket signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return errors.New("malformed join ticket payload")
	}
	var claims joinTicketClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return errors.New("malformed join ticket claims")
	}
	if claims.GameID != gameID {
		return errors.New("join ticket is for another game")
	}
	if claims.Expires <= now.Unix() {
		return errors.New("join ticket expired")
	}
	return nil
}
