package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFuzzCoverageReportsTheCorpus(t *testing.T) {
	stdout, _ := captureCommandOutput(t)
	if code := mainImpl([]string{"boardgame-util", "fuzz", "coverage", "--game", "tictactoe"}); code != 0 {
		t.Fatalf("fuzz coverage exited %d", code)
	}
	for _, want := range []string{"## tictactoe", "8 documents", "| Place Token |", "### Predicates", "### Gaps"} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, stdout.String())
		}
	}
}

func TestFuzzCoverageJSON(t *testing.T) {
	stdout, _ := captureCommandOutput(t)
	if code := mainImpl([]string{"boardgame-util", "fuzz", "coverage", "--json"}); code != 0 {
		t.Fatalf("fuzz coverage --json exited %d", code)
	}
	var reports []struct {
		Game      string `json:"game"`
		Documents int    `json:"documents"`
	}
	if err := json.Unmarshal([]byte(stdout.String()), &reports); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if len(reports) != 7 {
		t.Fatalf("%d reports, want one per corpus game", len(reports))
	}
	for _, report := range reports {
		if report.Documents != 8 {
			t.Errorf("%s: %d documents", report.Game, report.Documents)
		}
	}
}
