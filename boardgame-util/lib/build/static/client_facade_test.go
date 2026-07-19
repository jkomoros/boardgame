package static

import (
	"os"
	"path/filepath"
	"testing"
)

// Game clients are linked at static/game-src/<game>. From that assembled
// location ../../src/client.js must resolve to the framework facade exactly as
// it does for source renderers copied into the normal server tree.
func TestClientFacadeResolvesFromAssembledGameSrc(t *testing.T) {
	root := t.TempDir()
	if err := CopyStaticResources(root, true); err != nil {
		t.Fatalf("CopyStaticResources: %v", err)
	}

	gameClient := filepath.Join(root, staticSubFolder, gameSrcSubFolder, "fixture")
	if err := os.MkdirAll(gameClient, 0700); err != nil {
		t.Fatalf("create fixture game client: %v", err)
	}

	facadeFromGame := filepath.Join(gameClient, "..", "..", "src", "client.ts")
	info, err := os.Stat(facadeFromGame)
	if err != nil {
		t.Fatalf("../../src/client.js does not resolve from assembled game-src: %v", err)
	}
	if info.IsDir() {
		t.Fatalf("assembled facade path %q is a directory", facadeFromGame)
	}
}

func TestCopyStaticResourcesReplacesDevFileLinksForProduction(t *testing.T) {
	root := t.TempDir()
	if err := CopyStaticResources(root, false); err != nil {
		t.Fatalf("link development resources: %v", err)
	}
	indexPath := filepath.Join(root, staticSubFolder, "index.html")
	linked, err := os.Lstat(indexPath)
	if err != nil {
		t.Fatalf("stat linked index: %v", err)
	}
	if linked.Mode()&os.ModeSymlink == 0 {
		t.Fatal("development index should be linked")
	}

	if err := CopyStaticResources(root, true); err != nil {
		t.Fatalf("replace development resources for production: %v", err)
	}
	copied, err := os.Lstat(indexPath)
	if err != nil {
		t.Fatalf("stat copied index: %v", err)
	}
	if copied.Mode()&os.ModeSymlink != 0 || !copied.Mode().IsRegular() {
		t.Fatalf("production index mode = %v, want regular copied file", copied.Mode())
	}
}

func TestCopyStaticResourcesReplacesProductionCopiesForDevelopment(t *testing.T) {
	root := t.TempDir()
	if err := CopyStaticResources(root, true); err != nil {
		t.Fatal(err)
	}
	index := filepath.Join(root, staticSubFolder, "index.html")
	if info, err := os.Lstat(index); err != nil || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("production index shape = %v, %v; want regular file", info, err)
	}
	if err := CopyStaticResources(root, false); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Lstat(index); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("development index shape = %v, %v; want symlink", info, err)
	}
}
