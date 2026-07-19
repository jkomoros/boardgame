package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallGeneratedGameTypesPreservesOldFilesWhenStagingFails(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "a-game", "client", "_types.ts")
	if err := os.MkdirAll(filepath.Dir(oldPath), 0700); err != nil {
		t.Fatal(err)
	}
	const oldContents = "old complete generation"
	if err := os.WriteFile(oldPath, []byte(oldContents), 0600); err != nil {
		t.Fatal(err)
	}

	invalidParent := filepath.Join(dir, "z-not-a-directory")
	if err := os.WriteFile(invalidParent, []byte("sentinel"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := installGeneratedGameTypes([]generatedGameTypeFile{
		{path: oldPath, contents: []byte("new generation"), gameName: "a-game"},
		{path: filepath.Join(invalidParent, "_game_renderer.ts"), contents: []byte("new renderer"), gameName: "z-missing"},
	})
	if err == nil {
		t.Fatal("installGeneratedGameTypes() succeeded, want staging error")
	}
	got, readErr := os.ReadFile(oldPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != oldContents {
		t.Fatalf("old output changed on staging failure: got %q", got)
	}
}

func TestCheckGeneratedGameTypesIsNonMutating(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "_types.ts")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := checkGeneratedGameTypes([]generatedGameTypeFile{{path: path, contents: []byte("new")}}); err == nil {
		t.Fatal("checkGeneratedGameTypes() succeeded for stale output")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "old" {
		t.Fatalf("freshness check mutated output: %q", contents)
	}
	if err := checkGeneratedGameTypes([]generatedGameTypeFile{{path: path, contents: []byte("old")}}); err != nil {
		t.Fatalf("current output failed check: %v", err)
	}
}

func TestInstallGeneratedGameTypesReplacesCompletePair(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "client")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	typesPath := filepath.Join(dir, "_types.ts")
	rendererPath := filepath.Join(dir, "_game_renderer.ts")
	if err := installGeneratedGameTypes([]generatedGameTypeFile{
		{path: typesPath, contents: []byte("new types"), gameName: "game"},
		{path: rendererPath, contents: []byte("new renderer"), gameName: "game"},
	}); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{typesPath: "new types", rendererPath: "new renderer"} {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("%s = %q, want %q", path, got, want)
		}
	}
}
