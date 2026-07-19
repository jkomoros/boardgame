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
