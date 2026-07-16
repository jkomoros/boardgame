package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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

	err := installGeneratedGameTypes([]generatedGameTypeFile{
		{path: oldPath, contents: []byte("new generation"), gameName: "a-game"},
		{path: filepath.Join(dir, "z-missing", "client", "_game_renderer.ts"), contents: []byte("new renderer"), gameName: "z-missing"},
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

func TestInstallGeneratedGameTypesReportsAndPreservesFailedRestore(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "client")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "_types.ts")
	if err := os.WriteFile(path, []byte("old types"), 0600); err != nil {
		t.Fatal(err)
	}

	originalRename := renameGeneratedGameTypeFile
	originalRestore := restoreGeneratedGameTypeFile
	defer func() {
		renameGeneratedGameTypeFile = originalRename
		restoreGeneratedGameTypeFile = originalRestore
	}()
	calls := 0
	renameGeneratedGameTypeFile = func(oldPath, newPath string) error {
		calls++
		if calls == 2 {
			return errors.New("injected install failure")
		}
		return os.Rename(oldPath, newPath)
	}
	restoreGeneratedGameTypeFile = func(oldPath, newPath string) error {
		return errors.New("injected restore failure")
	}

	err := installGeneratedGameTypes([]generatedGameTypeFile{{
		path: path, contents: []byte("new types"), gameName: "game",
	}})
	if err == nil || !strings.Contains(err.Error(), "current restore: injected restore failure") {
		t.Fatalf("error = %v, want reported restore failure", err)
	}
	backups, globErr := filepath.Glob(filepath.Join(dir, ".game-types-*.backup"))
	if globErr != nil || len(backups) != 1 {
		t.Fatalf("preserved backups = %v, err = %v; want one", backups, globErr)
	}
	contents, readErr := os.ReadFile(backups[0])
	if readErr != nil || string(contents) != "old types" {
		t.Fatalf("backup contents = %q, err = %v; want old types", contents, readErr)
	}
}

func TestInstallGeneratedGameTypesRollsBackLateRenameFailure(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "client")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	typesPath := filepath.Join(dir, "_types.ts")
	rendererPath := filepath.Join(dir, "_game_renderer.ts")
	for path, contents := range map[string]string{typesPath: "old types", rendererPath: "old renderer"} {
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}

	originalRename := renameGeneratedGameTypeFile
	defer func() { renameGeneratedGameTypeFile = originalRename }()
	calls := 0
	renameGeneratedGameTypeFile = func(oldPath, newPath string) error {
		calls++
		// Each existing destination is backed up before its replacement. Fail
		// while installing the second file, after the first was replaced.
		if calls == 4 {
			return errors.New("injected late rename failure")
		}
		return os.Rename(oldPath, newPath)
	}

	err := installGeneratedGameTypes([]generatedGameTypeFile{
		{path: typesPath, contents: []byte("new types"), gameName: "game"},
		{path: rendererPath, contents: []byte("new renderer"), gameName: "game"},
	})
	if err == nil {
		t.Fatal("installGeneratedGameTypes() succeeded, want injected rename failure")
	}
	for path, want := range map[string]string{typesPath: "old types", rendererPath: "old renderer"} {
		got, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if string(got) != want {
			t.Fatalf("%s after rollback = %q, want %q", path, got, want)
		}
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
