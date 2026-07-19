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

func TestInstallGeneratedMoveNamesIsOneTransaction(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "a", "_move_names.ts")
	if err := os.Mkdir(filepath.Dir(first), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(first, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	invalidParent := filepath.Join(dir, "z-not-a-directory")
	if err := os.WriteFile(invalidParent, []byte("sentinel"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := installGeneratedMoveNames([]generatedMoveNamesFile{
		{path: first, contents: []byte("new"), gameName: "a"},
		{path: filepath.Join(invalidParent, "_move_names.ts"), contents: []byte("later"), gameName: "z"},
	}, false)
	if err == nil {
		t.Fatal("install succeeded despite invalid destination")
	}
	contents, readErr := os.ReadFile(first)
	if readErr != nil || string(contents) != "old" {
		t.Fatalf("first output after failed transaction = %q, %v; want old", contents, readErr)
	}
}
