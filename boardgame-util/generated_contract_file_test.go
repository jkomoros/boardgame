package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGeneratedFileCurrentHandlesCurrentMissingAndDifferentSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "contract.ts")
	if current, err := generatedFileCurrent(path, []byte("expected")); err != nil || current {
		t.Fatalf("missing file: current=%v err=%v", current, err)
	}
	if err := os.WriteFile(path, []byte("expected"), 0o644); err != nil {
		t.Fatal(err)
	}
	if current, err := generatedFileCurrent(path, []byte("expected")); err != nil || !current {
		t.Fatalf("matching file: current=%v err=%v", current, err)
	}
	if current, err := generatedFileCurrent(path, []byte("short")); err != nil || current {
		t.Fatalf("different-sized file: current=%v err=%v", current, err)
	}
}

func TestGeneratedFileCurrentRejectsNonRegularDestination(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation generally requires elevated privileges on Windows")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "target.ts")
	path := filepath.Join(dir, "contract.ts")
	if err := os.WriteFile(target, []byte("expected"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if _, err := generatedFileCurrent(path, []byte("expected")); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("error = %v, want non-regular destination", err)
	}
}
