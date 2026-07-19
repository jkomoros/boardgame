package fileutil

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWriteFilesAtomicCreatesNestedSetAndOutputRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "new-root")
	files := map[string][]byte{
		"game/main.go":        []byte("main"),
		"game/client/view.ts": []byte("view"),
	}
	if err := WriteFilesAtomic(root, files, false, 0o644); err != nil {
		t.Fatalf("WriteFilesAtomic: %v", err)
	}
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestWriteFilesAtomicRejectsUnsafeAndAliasedPaths(t *testing.T) {
	root := t.TempDir()
	tests := []map[string][]byte{
		{"../escape": []byte("bad")},
		{filepath.Join(t.TempDir(), "absolute"): []byte("bad")},
		{"same": []byte("one"), "dir/../same": []byte("two")},
	}
	for _, files := range tests {
		if err := WriteFilesAtomic(root, files, false, 0o644); err == nil {
			t.Fatalf("WriteFilesAtomic(%v) succeeded", files)
		}
	}
}

func TestWriteFileSetAtomicAbsoluteWritesAcrossPackageDirectories(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "games", "alpha", "client", "_types.ts")
	second := filepath.Join(root, "examples", "beta", "client", "_types.ts")
	err := WriteFileSetAtomicAbsolute(map[string]FileSpec{
		first:  {Contents: []byte("alpha"), Mode: 0o644},
		second: {Contents: []byte("beta"), Mode: 0o644},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{first: "alpha", second: "beta"} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != want {
			t.Fatalf("%s = %q, %v; want %q", path, got, err, want)
		}
	}
}

func TestWriteFileSetAtomicAbsoluteRejectsRelativePath(t *testing.T) {
	err := WriteFileSetAtomicAbsolute(map[string]FileSpec{
		"relative/output": {Contents: []byte("bad"), Mode: 0o644},
	}, true)
	if err == nil || !strings.Contains(err.Error(), "must be absolute") {
		t.Fatalf("error = %v, want absolute-path error", err)
	}
}

func TestWriteFilesAtomicRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation generally requires elevated privileges on Windows")
	}
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}
	err := WriteFilesAtomic(root, map[string][]byte{"linked/escape": []byte("bad")}, false, 0o644)
	if err == nil || !strings.Contains(err.Error(), "escapes the output root") {
		t.Fatalf("error = %v, want symlink escape error", err)
	}
}

func TestWriteFilesAtomicRejectsOutputsAliasedBySymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation generally requires elevated privileges on Windows")
	}
	root := t.TempDir()
	realDir := filepath.Join(root, "real")
	if err := os.Mkdir(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realDir, filepath.Join(root, "alias")); err != nil {
		t.Fatal(err)
	}
	err := WriteFilesAtomic(root, map[string][]byte{
		"real/output":  []byte("first"),
		"alias/output": []byte("second"),
	}, true, 0o644)
	if err == nil || !strings.Contains(err.Error(), "resolve to the same file") {
		t.Fatalf("error = %v, want symlink-alias collision", err)
	}
	if _, statErr := os.Stat(filepath.Join(realDir, "output")); !os.IsNotExist(statErr) {
		t.Fatalf("collision preflight wrote output: %v", statErr)
	}
}

func TestWriteFilesAtomicPreflightLeavesSetUntouched(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "b"), []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := WriteFilesAtomic(root, map[string][]byte{"a": []byte("new"), "b": []byte("replace")}, false, 0o644)
	if err == nil {
		t.Fatal("WriteFilesAtomic succeeded despite existing output")
	}
	if _, err := os.Stat(filepath.Join(root, "a")); !os.IsNotExist(err) {
		t.Fatal("preflight failure installed an earlier file")
	}
}

func TestWriteFilesAtomicRollsBackLateRenameFailure(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a", "b"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("old-"+name), 0o640); err != nil {
			t.Fatal(err)
		}
	}
	originalRename := rename
	calls := 0
	rename = func(oldPath, newPath string) error {
		calls++
		if calls == 4 {
			return errors.New("injected late rename failure")
		}
		return originalRename(oldPath, newPath)
	}
	t.Cleanup(func() { rename = originalRename })

	err := WriteFilesAtomic(root, map[string][]byte{"a": []byte("new-a"), "b": []byte("new-b")}, true, 0o644)
	if err == nil {
		t.Fatal("WriteFilesAtomic succeeded, want injected failure")
	}
	for _, name := range []string{"a", "b"} {
		contents, readErr := os.ReadFile(filepath.Join(root, name))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if got, want := string(contents), "old-"+name; got != want {
			t.Errorf("%s after rollback = %q, want %q", name, got, want)
		}
	}
	assertNoSetArtifacts(t, root)
}

func TestWriteFilesAtomicReportsFailedRestoreAndPreservesBackup(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "output")
	if err := os.WriteFile(path, []byte("old"), 0o640); err != nil {
		t.Fatal(err)
	}
	originalRename := rename
	calls := 0
	rename = func(oldPath, newPath string) error {
		calls++
		switch calls {
		case 2:
			return errors.New("injected install failure")
		case 3:
			return errors.New("injected restore failure")
		default:
			return originalRename(oldPath, newPath)
		}
	}
	t.Cleanup(func() { rename = originalRename })

	err := WriteFilesAtomic(root, map[string][]byte{"output": []byte("new")}, true, 0o644)
	if err == nil || !strings.Contains(err.Error(), "injected restore failure") {
		t.Fatalf("error = %v, want reported restore failure", err)
	}
	backups, globErr := filepath.Glob(filepath.Join(root, ".boardgame-set-*.backup"))
	if globErr != nil || len(backups) != 1 {
		t.Fatalf("preserved backups = %v, err = %v; want one", backups, globErr)
	}
	contents, readErr := os.ReadFile(backups[0])
	if readErr != nil || string(contents) != "old" {
		t.Fatalf("backup contents = %q, err = %v; want old", contents, readErr)
	}
}

func TestWriteFilesAtomicReportsPostCommitCleanupFailure(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "output")
	if err := os.WriteFile(path, []byte("old"), 0o640); err != nil {
		t.Fatal(err)
	}
	originalRemove := removeSetArtifact
	removeSetArtifact = func(path string) error {
		if strings.HasSuffix(path, ".backup") {
			return errors.New("injected cleanup failure")
		}
		return originalRemove(path)
	}
	t.Cleanup(func() { removeSetArtifact = originalRemove })

	err := WriteFilesAtomic(root, map[string][]byte{"output": []byte("new")}, true, 0o644)
	if err == nil || !strings.Contains(err.Error(), "outputs committed") || !strings.Contains(err.Error(), "injected cleanup failure") {
		t.Fatalf("error = %v, want explicit post-commit cleanup failure", err)
	}
	contents, readErr := os.ReadFile(path)
	if readErr != nil || string(contents) != "new" {
		t.Fatalf("committed output = %q, %v; want new", contents, readErr)
	}
	backups, globErr := filepath.Glob(filepath.Join(root, ".boardgame-set-*.backup"))
	if globErr != nil || len(backups) != 1 {
		t.Fatalf("preserved backups = %v, err = %v; want one", backups, globErr)
	}
}

func TestWriteFilesAtomicExclusiveInstallDoesNotClobberRaceWinner(t *testing.T) {
	root := t.TempDir()
	originalLink := link
	calls := 0
	link = func(oldPath, newPath string) error {
		calls++
		if calls == 2 {
			if err := os.WriteFile(newPath, []byte("race-winner"), 0o644); err != nil {
				return err
			}
		}
		return originalLink(oldPath, newPath)
	}
	t.Cleanup(func() { link = originalLink })

	err := WriteFilesAtomic(root, map[string][]byte{"a": []byte("new-a"), "b": []byte("new-b")}, false, 0o644)
	if err == nil {
		t.Fatal("WriteFilesAtomic succeeded, want exclusive-install race failure")
	}
	if _, err := os.Stat(filepath.Join(root, "a")); !os.IsNotExist(err) {
		t.Fatal("earlier output was not rolled back")
	}
	contents, readErr := os.ReadFile(filepath.Join(root, "b"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(contents) != "race-winner" {
		t.Fatalf("race winner was clobbered: %q", contents)
	}
	assertNoSetArtifacts(t, root)
}

func TestWriteFileSetAtomicDeletesAsPartOfTransaction(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "orphan.go")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileSetAtomic(root, map[string]FileSpec{"orphan.go": {Delete: true}}, true); err != nil {
		t.Fatalf("WriteFileSetAtomic: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("deleted file remains: %v", err)
	}
	assertNoSetArtifacts(t, root)
}

func TestWriteFileSetAtomicCanRequireDeletionTarget(t *testing.T) {
	root := t.TempDir()
	err := WriteFileSetAtomic(root, map[string]FileSpec{
		"disappeared.go": {Delete: true, RequireExisting: true},
	}, true)
	if err == nil || !strings.Contains(err.Error(), "no longer exists") {
		t.Fatalf("error = %v, want disappeared-output error", err)
	}
}

func TestWriteFileSetAtomicRestoresDeletionOnLaterFailure(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a-delete", "b-write"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("old-"+name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	originalRename := rename
	calls := 0
	rename = func(oldPath, newPath string) error {
		calls++
		if calls == 3 {
			return errors.New("injected install failure after deletion")
		}
		return originalRename(oldPath, newPath)
	}
	t.Cleanup(func() { rename = originalRename })

	err := WriteFileSetAtomic(root, map[string]FileSpec{
		"a-delete": {Delete: true},
		"b-write":  {Contents: []byte("new"), Mode: 0o644},
	}, true)
	if err == nil {
		t.Fatal("WriteFileSetAtomic succeeded, want injected failure")
	}
	for _, name := range []string{"a-delete", "b-write"} {
		contents, readErr := os.ReadFile(filepath.Join(root, name))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if got, want := string(contents), "old-"+name; got != want {
			t.Errorf("%s after rollback = %q, want %q", name, got, want)
		}
	}
	assertNoSetArtifacts(t, root)
}

func assertNoSetArtifacts(t *testing.T, root string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(root, ".boardgame-set-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("transaction artifacts remain: %v", matches)
	}
}
