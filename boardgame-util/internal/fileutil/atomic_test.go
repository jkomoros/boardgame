package fileutil

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWriteFileAtomicReplacesContentsAndPreservesPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("old"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(path, []byte("new"), 0o644); err != nil {
		t.Fatalf("WriteFileAtomic: %v", err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "new" {
		t.Fatalf("contents = %q, want new", contents)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o640 {
		t.Fatalf("permissions = %o, want 640", got)
	}
}

func TestWriteFileAtomicCreatesWithRequestedPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new.json")
	if err := WriteFileAtomic(path, []byte("new"), 0o644); err != nil {
		t.Fatalf("WriteFileAtomic: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o644 {
		t.Fatalf("permissions = %o, want 644", got)
	}
}

func TestWriteFileAtomicRenameFailureLeavesOriginalUntouched(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	originalRename := rename
	rename = func(string, string) error { return errors.New("injected rename failure") }
	t.Cleanup(func() { rename = originalRename })

	if err := WriteFileAtomic(path, []byte("new"), 0o644); err == nil {
		t.Fatal("WriteFileAtomic succeeded, want rename failure")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "old" {
		t.Fatalf("contents after failed replacement = %q, want old", contents)
	}
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".boardgame-write-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary files left after failure: %v", matches)
	}
}

func TestWriteFileAtomicRefusesSymlinkDestination(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation generally requires elevated privileges on Windows")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "target.json")
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}

	err := WriteFileAtomic(path, []byte("new"), 0o644)
	if err == nil || !strings.Contains(err.Error(), "non-regular") {
		t.Fatalf("WriteFileAtomic error = %v, want non-regular destination error", err)
	}
	contents, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "old" {
		t.Fatalf("symlink target contents = %q, want old", contents)
	}
}

func TestWriteFileExclusiveDoesNotReplaceExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "existing.go")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileExclusive(path, []byte("new"), 0o644); err == nil {
		t.Fatal("WriteFileExclusive succeeded for an existing path")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "old" {
		t.Fatalf("existing contents = %q, want old", contents)
	}
}

func TestWriteFileExclusivePublishesCompleteFileWithoutStagingArtifacts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.go")
	if err := WriteFileExclusive(path, []byte("complete"), 0o640); err != nil {
		t.Fatalf("WriteFileExclusive: %v", err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "complete" {
		t.Fatalf("contents = %q, want complete", contents)
	}
	matches, err := filepath.Glob(filepath.Join(dir, ".boardgame-exclusive-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("exclusive staging artifacts remain: %v", matches)
	}
}
