package stub

import "testing"

func TestUppercaseFirstRunePreservesRestOfName(t *testing.T) {
	for input, want := range map[string]string{
		"checkers":  "Checkers",
		"ticTacToe": "TicTacToe",
		"échecs":    "Échecs",
		"":          "",
	} {
		if got := uppercaseFirstRune(input); got != want {
			t.Errorf("uppercaseFirstRune(%q) = %q, want %q", input, got, want)
		}
	}
}
