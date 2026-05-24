package api

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
)

// roomCodeAlphabet is 22 uppercase letters chosen for readability when spoken
// aloud or typed quickly. O, I, L, Z are omitted because of typo / homoglyph
// risk with 0, 1, and 2.
const roomCodeAlphabet = "ABCDEFGHJKMNPQRSTUVWXY"

const (
	defaultRoomCodeLength  = 4
	fallbackRoomCodeLength = 5
	roomCodeMaxRetries     = 10
)

// ErrRoomCodeNamespaceExhausted is returned when GenerateRoomCode cannot find
// a free code after retrying at both 4-letter and 5-letter lengths. In
// practice this only fires when the active-room namespace is near-saturated
// (>~200k concurrent Table+Hand games for the 4-letter space) or when the
// codeInUse predicate is misbehaving.
var ErrRoomCodeNamespaceExhausted = errors.New("room code namespace exhausted: 10 4-letter and 10 5-letter retries all collided")

// GenerateRoomCode returns a fresh random room code that codeInUse reports as
// not-in-use. It tries up to roomCodeMaxRetries 4-letter codes first; if those
// all collide it falls back to roomCodeMaxRetries 5-letter codes. Returns
// ErrRoomCodeNamespaceExhausted if both lengths exhaust.
//
// codeInUse must return true iff the code is currently bound to a game that
// is either active OR still within the 24h post-Finished grace period (see
// spec §6.1). Implementations typically wrap a StorageManager.GameByRoomCode
// call; see server/api/main.go for the binding.
func GenerateRoomCode(codeInUse func(code string) (bool, error)) (string, error) {
	for length := defaultRoomCodeLength; length <= fallbackRoomCodeLength; length++ {
		for i := 0; i < roomCodeMaxRetries; i++ {
			code, err := randomRoomCode(length)
			if err != nil {
				return "", err
			}
			inUse, err := codeInUse(code)
			if err != nil {
				return "", err
			}
			if !inUse {
				return code, nil
			}
		}
	}
	return "", ErrRoomCodeNamespaceExhausted
}

func randomRoomCode(length int) (string, error) {
	result := make([]byte, length)
	max := big.NewInt(int64(len(roomCodeAlphabet)))
	for i := range result {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = roomCodeAlphabet[n.Int64()]
	}
	return string(result), nil
}

// NormalizeRoomCode trims whitespace, uppercases, and validates the input
// against the confusion-resistant alphabet and the allowed length range.
// Returns the normalized code or an error describing what was wrong.
//
// This is the canonical input-side normalizer for /api/join — phones that
// enter codes with stray whitespace or lowercase letters get a forgiving
// experience; non-alphabet characters are rejected cleanly with a useful
// error rather than silently letting through a 404 from a malformed lookup.
func NormalizeRoomCode(input string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(input))
	if normalized == "" {
		return "", errors.New("room code is empty")
	}
	if len(normalized) < defaultRoomCodeLength || len(normalized) > fallbackRoomCodeLength {
		return "", errors.New("room code must be 4 or 5 characters")
	}
	for _, r := range normalized {
		if !strings.ContainsRune(roomCodeAlphabet, r) {
			return "", errors.New("room code contains an invalid character (allowed: " + roomCodeAlphabet + ")")
		}
	}
	return normalized, nil
}
