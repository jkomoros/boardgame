package api

import (
	"errors"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const (
	displayNameMinLen = 2
	displayNameMaxLen = 24
)

// NormalizeDisplayName applies the spec §6.6 validation pipeline to a
// user-supplied display name for a companion-mode seat presentation:
//
//   - NFKC-normalize (collapses compatibility forms before width check)
//   - Trim leading/trailing whitespace
//   - Length 2..24 inclusive after trim
//   - Allow only [A-Za-z0-9 ] (space allowed in interior, doubled spaces OK)
//   - Reject zero-width characters, RTL-override, combining marks, control
//     characters, surrogate runes
//
// Returns the normalized name on success, or an error describing what was
// wrong. The error messages are intentionally user-friendly because the
// phone surfaces them.
func NormalizeDisplayName(input string) (string, error) {
	// NFKC first so that fullwidth letters and similar are folded into their
	// ASCII equivalents before we apply the ASCII-only rule.
	normalized := norm.NFKC.String(input)
	normalized = strings.TrimSpace(normalized)

	if normalized == "" {
		return "", errors.New("display name is empty")
	}

	// Walk runes once to do the length + character checks in one pass.
	runeCount := 0
	for _, r := range normalized {
		runeCount++
		if isDisallowedNameRune(r) {
			return "", errors.New("display name contains a disallowed character")
		}
	}
	if runeCount < displayNameMinLen {
		return "", errors.New("display name is too short (minimum 2 characters)")
	}
	if runeCount > displayNameMaxLen {
		return "", errors.New("display name is too long (maximum 24 characters)")
	}
	return normalized, nil
}

// isDisallowedNameRune returns true if the rune should be rejected in a
// display name. We allow only A-Z, a-z, 0-9, and ASCII space; everything else
// is rejected. This includes — explicitly — zero-width characters, RTL
// overrides, combining marks, control chars, and any non-ASCII letter
// (because the NFKC normalize earlier should have folded the legitimate
// fullwidth-letter case into ASCII; anything still non-ASCII is intentional
// non-ASCII input and we reject for V1 per the spec).
func isDisallowedNameRune(r rune) bool {
	switch {
	case r >= 'A' && r <= 'Z':
		return false
	case r >= 'a' && r <= 'z':
		return false
	case r >= '0' && r <= '9':
		return false
	case r == ' ':
		return false
	}
	// Anything else (including all control / combining / RTL / zero-width /
	// surrogate / non-ASCII letters) is disallowed.
	_ = unicode.IsPrint // keep the import live for future tightening
	return true
}
