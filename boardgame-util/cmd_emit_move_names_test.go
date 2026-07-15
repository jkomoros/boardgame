package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallGeneratedMoveNamesCheckIsNonMutating(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "_move_names.ts")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	generated := []generatedMoveNamesFile{{path: path, contents: []byte("new"), gameName: "game"}}
	if err := installGeneratedMoveNames(generated, true); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("check error = %v, want stale diagnostic", err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "old" {
		t.Fatalf("check mutated file: %q", contents)
	}
	if err := installGeneratedMoveNames([]generatedMoveNamesFile{{path: path, contents: []byte("old"), gameName: "game"}}, true); err != nil {
		t.Fatalf("current file failed check: %v", err)
	}
}
