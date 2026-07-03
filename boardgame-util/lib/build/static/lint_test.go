package static

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTypeCheckGameSrcNoGameSrc verifies the graceful no-op when the
// assembled dir has no game-src folder (nothing to check, no tsc run).
func TestTypeCheckGameSrcNoGameSrc(t *testing.T) {
	dir := t.TempDir()
	diagnostics, err := TypeCheckGameSrc(dir)
	if err != nil {
		t.Fatalf("expected nil err for dir without game-src, got %v", err)
	}
	if diagnostics != nil {
		t.Fatalf("expected nil diagnostics for dir without game-src, got %v", diagnostics)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "tsconfig.gamesrc.json")); statErr == nil {
		t.Fatal("TypeCheckGameSrc wrote a tsconfig despite no game-src folder")
	}
}
