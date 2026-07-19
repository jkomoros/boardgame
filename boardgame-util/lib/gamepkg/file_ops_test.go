package gamepkg

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPackageFileOperationsRejectEscapingPaths(t *testing.T) {
	root := t.TempDir()
	pkg := &Pkg{absolutePath: root}
	escapeName := filepath.Base(root) + "-outside.txt"
	escapePath := filepath.Join("..", escapeName)
	outside := filepath.Join(filepath.Dir(root), escapeName)

	operations := map[string]func() error{
		"ensure directory": func() error { return pkg.EnsureDir("../outside") },
		"write file":       func() error { return pkg.WriteFile(escapePath, []byte("bad"), true) },
		"remove file":      func() error { return pkg.RemoveFile(escapePath) },
		"remove directory": func() error { return pkg.RemoveDirIfEmpty("../outside") },
	}
	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			err := operation()
			if err == nil || !strings.Contains(err.Error(), "escapes package root") {
				t.Fatalf("error = %v, want package-root escape error", err)
			}
		})
	}
	if pkg.Has(escapePath) {
		t.Fatal("Has reported an escaping path as present")
	}
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatalf("escaping operation created %s", outside)
	}
}

func TestPackageFileOperationsRejectAbsolutePaths(t *testing.T) {
	pkg := &Pkg{absolutePath: t.TempDir()}
	err := pkg.WriteFile(filepath.Join(t.TempDir(), "outside.txt"), []byte("bad"), true)
	if err == nil || !strings.Contains(err.Error(), "must be relative") {
		t.Fatalf("error = %v, want relative-path error", err)
	}
}

func TestPackageFileOperationsRejectSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation generally requires elevated privileges on Windows")
	}
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}
	pkg := &Pkg{absolutePath: root}
	err := pkg.WriteFile(filepath.Join("linked", "escaped.txt"), []byte("bad"), true)
	if err == nil || !strings.Contains(err.Error(), "through a symlink") {
		t.Fatalf("error = %v, want symlink-escape error", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "escaped.txt")); !os.IsNotExist(err) {
		t.Fatal("package write escaped through a symlink")
	}
}

func TestPackageWriteFileOverwriteModesAreSafe(t *testing.T) {
	root := t.TempDir()
	pkg := &Pkg{absolutePath: root}
	path := filepath.Join(root, "generated.go")
	if err := os.WriteFile(path, []byte("old"), 0o640); err != nil {
		t.Fatal(err)
	}

	if err := pkg.WriteFile("generated.go", []byte("unexpected"), false); err == nil {
		t.Fatal("WriteFile overwrite=false replaced an existing file")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "old" {
		t.Fatalf("contents after exclusive write = %q, want old", contents)
	}

	if err := pkg.WriteFile("generated.go", []byte("new"), true); err != nil {
		t.Fatalf("WriteFile overwrite=true: %v", err)
	}
	contents, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "new" {
		t.Fatalf("contents after overwrite = %q, want new", contents)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o640 {
		t.Fatalf("permissions after overwrite = %o, want 640", got)
	}
}

func TestPathWithinDirectoryUsesPathBoundaries(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "cache", "pkg", "mod")
	tests := map[string]bool{
		root: true,
		filepath.Join(root, "example.com", "game"): true,
		root + "-backup": false,
		filepath.Join(filepath.Dir(root), "model"): false,
	}
	for path, want := range tests {
		if got := pathWithinDirectory(root, path); got != want {
			t.Errorf("pathWithinDirectory(%q, %q) = %v, want %v", root, path, got, want)
		}
	}
}

func TestPackageReadOnlyDetectsFilesystemPermissions(t *testing.T) {
	root := t.TempDir()
	if err := os.Chmod(root, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o755) })
	if !(&Pkg{absolutePath: root}).ReadOnly() {
		t.Fatal("ReadOnly returned false for a directory with no write bits")
	}
}
