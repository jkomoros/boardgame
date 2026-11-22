package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFuzzRunRefusesTheCheckedInCorpus(t *testing.T) {
	root, err := boardgameRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	for _, out := range []string{
		filepath.Join(root, "conformance"),
		filepath.Join(root, "conformance", "fuzz"),
		filepath.Join(root, "conformance", "new", "dir"),
	} {
		if err := refuseCorpusDir("", out); err == nil {
			t.Errorf("refuseCorpusDir accepted %s", out)
		}
	}
	for _, out := range []string{t.TempDir(), filepath.Join(root, "conformance-elsewhere")} {
		if err := refuseCorpusDir("", out); err != nil {
			t.Errorf("refuseCorpusDir refused %s: %v", out, err)
		}
	}
}

func TestFuzzRunWritesDocuments(t *testing.T) {
	_, _ = captureCommandOutput(t)
	out := t.TempDir()
	if code := mainImpl([]string{"boardgame-util", "fuzz", "run", "--game", "tictactoe", "--seeds", "1..2", "--out", out}); code != 0 {
		t.Fatalf("fuzz run exited %d", code)
	}
	for _, name := range []string{"tictactoe/seed-0001.json.gz", "tictactoe/seed-0002.json.gz", "index.json"} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(name))); err != nil {
			t.Errorf("fuzz run did not write %s: %v", name, err)
		}
	}
}

func TestFuzzRunFailsInsideCorpus(t *testing.T) {
	_, stderr := captureCommandOutput(t)
	root, err := boardgameRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "conformance", "fuzz-scratch")
	if code := mainImpl([]string{"boardgame-util", "fuzz", "run", "--game", "tictactoe", "--seed", "1", "--out", out}); code == 0 {
		t.Fatal("fuzz run inside conformance/ succeeded")
	}
	if !strings.Contains(stderr.String(), "outside the checked-in corpus") {
		t.Fatalf("error did not explain the refusal: %q", stderr.String())
	}
	if _, err := os.Stat(out); err == nil {
		t.Fatalf("fuzz run created %s", out)
	}
}
