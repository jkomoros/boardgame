package boardgame

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoGeneratedWorkspacesAreTracked is a backstop for .gitignore. In
// particular, it rejects generated workspaces added with git add --force.
func TestNoGeneratedWorkspacesAreTracked(t *testing.T) {
	if _, err := os.Stat(".git"); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skip("not running from a Git checkout")
		}
		t.Fatalf("inspect .git: %v", err)
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable is unavailable")
	}

	out, err := exec.Command("git", "ls-files", "-z").Output()
	if err != nil {
		t.Fatalf("list tracked files: %v", err)
	}

	var forbidden []string
	for _, path := range strings.Split(string(out), "\x00") {
		if generatedWorkspacePath(path) {
			forbidden = append(forbidden, path)
		}
	}
	if len(forbidden) > 0 {
		t.Fatalf("generated boardgame-util workspaces must not be tracked:\n%s", strings.Join(forbidden, "\n"))
	}
}

func generatedWorkspacePath(path string) bool {
	prefixes := [...]string{
		"temp_serve_",
		"temp_gametypes_",
		"temp_moveargs_",
		"temp_movenames_",
	}
	for _, segment := range strings.Split(filepath.ToSlash(path), "/") {
		for _, prefix := range prefixes {
			if strings.HasPrefix(segment, prefix) {
				return true
			}
		}
	}
	return false
}

func TestGeneratedWorkspacePath(t *testing.T) {
	tests := map[string]bool{
		"temp_serve_123/main.go":                         true,
		"boardgame-util/temp_gametypes_123/main.go":      true,
		"nested/deeper/temp_moveargs_123/main.go":        true,
		"nested/deeper/temp_movenames_123/main.go":       true,
		"docs/how-temp_serve_workspaces-are-created.md":  false,
		"boardgame-util/temporary-serve-fixture/main.go": false,
	}
	for path, want := range tests {
		t.Run(path, func(t *testing.T) {
			if got := generatedWorkspacePath(path); got != want {
				t.Fatalf("generatedWorkspacePath(%q) = %v, want %v", path, got, want)
			}
		})
	}
}

func TestBuildSourcePackageIsNotIgnored(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable is unavailable")
	}
	path := "boardgame-util/lib/build/static/future_source_test.go"
	err := exec.Command("git", "check-ignore", "--quiet", "--no-index", path).Run()
	if err == nil {
		t.Fatalf("%s is ignored; new build-library source files could be silently lost", path)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("check ignore policy for %s: %v", path, err)
	}
}

func TestLocalBuildCacheIsIgnored(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable is unavailable")
	}
	path := ".cache/go-build/README"
	if err := exec.Command("git", "check-ignore", "--quiet", "--no-index", path).Run(); err != nil {
		t.Fatalf("%s must stay ignored: %v", path, err)
	}
}

func TestGoLocalUsesRequestedCache(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go executable is unavailable")
	}
	cacheDir := filepath.Join(t.TempDir(), "go-build")
	cmd := exec.Command("./scripts/go-local", "env", "GOCACHE")
	cmd.Env = append(os.Environ(), "BOARDGAME_GO_CACHE_DIR="+cacheDir)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("run scripts/go-local: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != cacheDir {
		t.Fatalf("GOCACHE = %q, want %q", got, cacheDir)
	}
	if info, err := os.Stat(cacheDir); err != nil || !info.IsDir() {
		t.Fatalf("cache directory was not created: %v", err)
	}
}

/*
TestSeatingRendezvousKeysHaveOneDefinition keeps the seating rendezvous keys
from growing a fifth copy.

moves.SeatPlayer and the server meet through
StorageManager.FetchInjectedDataForGame, keyed by two strings, and a string is
the whole contract: an unrecognized key returns nil with no error anywhere, so a
drifted copy means "No player to seat" forever and no diagnostic pointing at
why.

They had FOUR independent literal definitions -- moves/seat_player.go,
server/api/storage.go, boardgame-util/lib/golden/storage.go, and
examples/werewolf/main_test.go -- held together by a comment on three of them.
The comments had already drifted: two named exactly one other copy, and none
named the werewolf one. That is what a comment-enforced invariant looks like
after a while.

There is now one definition, in moves/interfaces, which all four sites import.
That makes drift a compile error rather than a silent runtime nothing -- but
only for the copies that exist today. This test is what stops a fifth from being
typed out fresh: the literal may appear only where it is defined.
*/
func TestSeatingRendezvousKeysHaveOneDefinition(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable is unavailable")
	}

	// The definition site, relative to the repo root this test runs from.
	const definitionFile = "moves/interfaces/main.go"

	// Assembled from pieces rather than written whole, so that this file does
	// not itself become a fifth copy of the literal it is policing.
	const keyPrefix = "github.com/jkomoros/boardgame/server/api"
	keys := []string{
		keyPrefix + ".PlayerToSeat",
		keyPrefix + ".WillSeatPlayer",
	}

	out, err := exec.Command("git", "ls-files", "-z", "*.go").Output()
	if err != nil {
		t.Fatalf("list tracked Go files: %v", err)
	}

	sawDefinition := false
	for _, path := range strings.Split(string(out), "\x00") {
		if path == "" {
			continue
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			//A tracked file we cannot read is not this test's business.
			continue
		}
		text := string(contents)

		for _, key := range keys {
			if !strings.Contains(text, `"`+key+`"`) {
				continue
			}
			if filepath.ToSlash(path) == definitionFile {
				sawDefinition = true
				continue
			}
			t.Errorf("%v spells the seating rendezvous key %q as a literal. Use interfaces.PlayerToSeatRendezvousDataType / interfaces.WillSeatPlayerRendezvousDataType instead: this key is a protocol between moves.SeatPlayer and the server, FetchInjectedDataForGame answers nil for an unrecognized one with no error anywhere, and four independent copies of it is how it got here.", path, key)
		}
	}

	if !sawDefinition {
		t.Fatalf("no tracked file spells the rendezvous keys, not even %v. Either the definition moved (point this test at it) or the keys changed; leaving this unmatched would make the test pass without checking anything.", definitionFile)
	}
}
