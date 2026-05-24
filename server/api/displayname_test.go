package api

import (
	"strings"
	"testing"

	"github.com/workfit/tester/assert"
)

func TestNormalizeDisplayNameForgiving(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Alice", "Alice"},
		{"  Alice  ", "Alice"},
		{"Brave Lion", "Brave Lion"},
		{"User42", "User42"},
		// Fullwidth letters get folded by NFKC and accepted.
		{"Ａlice", "Alice"},
	}
	for _, c := range cases {
		got, err := NormalizeDisplayName(c.input)
		if err != nil {
			t.Errorf("input %q: unexpected error %v", c.input, err)
			continue
		}
		assert.For(t, c.input).ThatActual(got).Equals(c.want)
	}
}

func TestNormalizeDisplayNameRejects(t *testing.T) {
	cases := []string{
		"",            // empty
		"  ",          // whitespace-only
		"A",           // too short
		strings.Repeat("a", 25), // too long
		"Alice!",      // punctuation
		"héllo",       // diacritic
		"Alice​", // zero-width space
		"Alice‮", // RTL override
		"Alice\nNext", // newline
		"Alice\x00",   // NUL
		"日本語",         // non-ASCII letters
	}
	for _, c := range cases {
		_, err := NormalizeDisplayName(c)
		if err == nil {
			t.Errorf("input %q: expected error, got nil", c)
		}
	}
	_ = assert.For
}
