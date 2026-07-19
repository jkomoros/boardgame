package static

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
)

func TestStaticBuildDirRejectsFilePaths(t *testing.T) {
	root := t.TempDir()
	buildFile := filepath.Join(root, "build-file")
	if err := os.WriteFile(buildFile, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := staticBuildDir(buildFile); err == nil {
		t.Fatal("staticBuildDir accepted a file as its build directory")
	}
	staticFile := filepath.Join(root, staticSubFolder)
	if err := os.WriteFile(staticFile, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := staticBuildDir(root); err == nil {
		t.Fatal("staticBuildDir accepted a file as its static directory")
	}
}

func TestStaticBuildDirRejectsSymlinkedStaticDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation generally requires elevated privileges on Windows")
	}
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, staticSubFolder)); err != nil {
		t.Fatal(err)
	}
	if _, err := staticBuildDir(root); err == nil {
		t.Fatal("staticBuildDir accepted a symlinked static directory")
	}
}

func TestClientConfigForBuildDoesNotMutateCaller(t *testing.T) {
	source := &config.ClientConfig{TableHandSupportedGames: []string{"caller-owned"}}
	result := clientConfigForBuild(source, nil)
	if result == source {
		t.Fatal("clientConfigForBuild returned the caller's pointer")
	}
	if len(source.TableHandSupportedGames) != 1 || source.TableHandSupportedGames[0] != "caller-owned" {
		t.Fatalf("caller config was mutated: %#v", source)
	}
	if len(result.TableHandSupportedGames) != 0 {
		t.Fatalf("build config capabilities = %v, want empty", result.TableHandSupportedGames)
	}
}

func TestEnsureSymlinkReconcilesStaleDestinations(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation generally requires elevated privileges on Windows")
	}
	dir := t.TempDir()
	first := filepath.Join(dir, "first")
	second := filepath.Join(dir, "second")
	local := filepath.Join(dir, "local")
	for _, path := range []string{first, second} {
		if err := os.WriteFile(path, []byte(filepath.Base(path)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(first, local); err != nil {
		t.Fatal(err)
	}
	changed, err := ensureSymlink(local, second, second, true)
	if err != nil || !changed {
		t.Fatalf("reconcile stale link: changed=%v err=%v", changed, err)
	}
	contents, readErr := os.ReadFile(local)
	if readErr != nil || string(contents) != "second" {
		t.Fatalf("reconciled link contents = %q, %v; want second", contents, readErr)
	}
	changed, err = ensureSymlink(local, second, second, true)
	if err != nil || changed {
		t.Fatalf("retain current link: changed=%v err=%v", changed, err)
	}
	if err := os.Remove(local); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(local, []byte("owned output"), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err = ensureSymlink(local, first, first, true)
	if err != nil || !changed {
		t.Fatalf("replace owned regular file: changed=%v err=%v", changed, err)
	}
}
