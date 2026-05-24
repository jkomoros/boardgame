package api

import (
	"errors"
	"strings"
	"testing"

	"github.com/workfit/tester/assert"
)

func TestRoomCodeAlphabetExcludesConfusables(t *testing.T) {
	for _, banned := range []rune{'O', 'I', 'L', 'Z'} {
		if strings.ContainsRune(roomCodeAlphabet, banned) {
			t.Errorf("alphabet should not contain %q (confusable with digits or letters)", banned)
		}
	}
	assert.For(t).ThatActual(len(roomCodeAlphabet)).Equals(22)
}

func TestRandomRoomCodeUsesOnlyAlphabet(t *testing.T) {
	for i := 0; i < 200; i++ {
		code, err := randomRoomCode(defaultRoomCodeLength)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(len(code)).Equals(defaultRoomCodeLength)
		for _, r := range code {
			if !strings.ContainsRune(roomCodeAlphabet, r) {
				t.Fatalf("random code %q contained non-alphabet rune %q", code, r)
			}
		}
	}
}

func TestGenerateRoomCodeReturnsCodeWhenAllFree(t *testing.T) {
	codeInUse := func(code string) (bool, error) { return false, nil }
	code, err := GenerateRoomCode(codeInUse)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(len(code)).Equals(defaultRoomCodeLength)
}

func TestGenerateRoomCodeFallsBackToFiveLetterAfterCollisions(t *testing.T) {
	// codeInUse returns true for every 4-letter code, false for 5-letter
	codeInUse := func(code string) (bool, error) {
		return len(code) == defaultRoomCodeLength, nil
	}
	code, err := GenerateRoomCode(codeInUse)
	assert.For(t).ThatActual(err).IsNil()
	assert.For(t).ThatActual(len(code)).Equals(fallbackRoomCodeLength)
}

func TestGenerateRoomCodeExhaustionError(t *testing.T) {
	codeInUse := func(code string) (bool, error) { return true, nil }
	_, err := GenerateRoomCode(codeInUse)
	assert.For(t).ThatActual(errors.Is(err, ErrRoomCodeNamespaceExhausted)).IsTrue()
}

func TestGenerateRoomCodePropagatesPredicateError(t *testing.T) {
	sentinel := errors.New("db down")
	codeInUse := func(code string) (bool, error) { return false, sentinel }
	_, err := GenerateRoomCode(codeInUse)
	assert.For(t).ThatActual(errors.Is(err, sentinel)).IsTrue()
}

func TestNormalizeRoomCodeForgivingInputs(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"abcd", "ABCD"},
		{"  ABCD  ", "ABCD"},
		{"AbCd", "ABCD"},
		{"abcde", "ABCDE"},
	}
	for _, c := range cases {
		got, err := NormalizeRoomCode(c.input)
		assert.For(t, c.input).ThatActual(err).IsNil()
		assert.For(t, c.input).ThatActual(got).Equals(c.want)
	}
}

func TestNormalizeRoomCodeRejectsBadInputs(t *testing.T) {
	cases := []string{
		"",     // empty
		"   ",  // whitespace-only
		"ABC",  // too short
		"ABCDEF", // too long
		"ABCO", // contains O (banned)
		"ABCI", // contains I (banned)
		"ABCL", // contains L (banned)
		"ABCZ", // contains Z (banned)
		"AB1D", // contains digit
		"AB!D", // contains punctuation
	}
	for _, c := range cases {
		_, err := NormalizeRoomCode(c)
		assert.For(t, c).ThatActual(err).IsNotNil()
	}
}
