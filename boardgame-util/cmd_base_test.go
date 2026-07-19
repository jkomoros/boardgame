package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewSystemTempDirUsesOperatingSystemTempDirectory(t *testing.T) {
	tempRoot := t.TempDir()
	t.Setenv("TMPDIR", tempRoot)

	dir, err := newSystemTempDir("temp_serve_")
	if err != nil {
		t.Fatalf("newSystemTempDir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("remove temporary directory: %v", err)
		}
	})

	if got := filepath.Dir(dir); got != tempRoot {
		t.Fatalf("temporary directory parent = %q, want OS temp directory %q", got, tempRoot)
	}
}

func TestTrackedTempDirIsRemovedByCleanup(t *testing.T) {
	tempRoot := t.TempDir()
	t.Setenv("TMPDIR", tempRoot)
	b := new(boardgameUtil)
	dir, err := b.newTrackedTempDir("temp_serve_")
	if err != nil {
		t.Fatalf("newTrackedTempDir: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("temporary directory was not created: %v", err)
	}
	b.Cleanup()
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("temporary directory remains after cleanup: %v", err)
	}
}

func TestTrackedTempDirRefusesAllocationAfterCleanupStarts(t *testing.T) {
	tempRoot := t.TempDir()
	t.Setenv("TMPDIR", tempRoot)
	b := new(boardgameUtil)
	b.Cleanup()
	if _, err := b.newTrackedTempDir("temp_serve_"); err == nil {
		t.Fatal("newTrackedTempDir succeeded after cleanup")
	}
	matches, err := filepath.Glob(filepath.Join(tempRoot, "temp_serve_*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("untracked temporary directories remain: %v", matches)
	}
}
