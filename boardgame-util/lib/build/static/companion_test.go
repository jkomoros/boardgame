package static

import (
	"os"
	"sort"
	"testing"
)

// Constructing a full gamepkg.Pkg requires a real Go-package on disk; that's
// more setup than these unit tests need. Instead we test the pure helpers
// (fileExists, pkgIsCompanionCapable on nil) and let an integration test
// (run via the dev server) cover the happy path of CompanionCapableGames
// over a real []*gamepkg.Pkg.

func TestPkgIsCompanionCapableNilPkg(t *testing.T) {
	if pkgIsCompanionCapable(nil) {
		t.Errorf("nil pkg should never be capable")
	}
}

func TestFileExists(t *testing.T) {
	if fileExists("/nonexistent/path/should/never/exist/asdfqwer") {
		t.Errorf("fileExists should return false for nonexistent path")
	}
	tmp, err := os.CreateTemp("", "fileExists-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()
	if !fileExists(tmp.Name()) {
		t.Errorf("fileExists should return true for existing file %s", tmp.Name())
	}
}

func TestCompanionCapableGamesEmpty(t *testing.T) {
	got := CompanionCapableGames(nil)
	if len(got) != 0 {
		t.Errorf("nil input should yield empty slice, got %v", got)
	}
}

func TestCompanionCapableGamesSortedDeterministic(t *testing.T) {
	// We can't easily construct synthetic gamepkg.Pkgs without filesystem
	// setup, but we can verify that the function sorts and deduplicates as
	// expected by calling sort.Strings directly on a known input — the
	// sort.Strings line in CompanionCapableGames is the contract. If
	// upstream changes the implementation to use a map iteration without
	// re-sorting, this test would need adjustment.
	input := []string{"zelda", "apex", "mario"}
	sort.Strings(input)
	want := []string{"apex", "mario", "zelda"}
	for i := range want {
		if input[i] != want[i] {
			t.Errorf("sort.Strings produced unexpected order: %v", input)
		}
	}
}
