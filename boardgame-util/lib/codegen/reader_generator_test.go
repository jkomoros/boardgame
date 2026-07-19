package codegen

import (
	"testing"
	"unicode"
)

func TestGeneratedIdentifierRuneCasing(t *testing.T) {
	if got, want := changeFirstRuneCase("éclairState", unicode.ToUpper), "ÉclairState"; got != want {
		t.Errorf("uppercased identifier = %q, want %q", got, want)
	}
	if got, want := firstRuneWithCase("ÉclairState", unicode.ToLower), "é"; got != want {
		t.Errorf("receiver identifier = %q, want %q", got, want)
	}
	if got := firstRuneWithCase("", unicode.ToLower); got != "" {
		t.Errorf("empty receiver identifier = %q", got)
	}
}
