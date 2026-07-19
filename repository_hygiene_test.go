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
