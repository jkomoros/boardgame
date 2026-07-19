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
