package static

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAdjacentNodeModulesPath(t *testing.T) {
	root := t.TempDir()
	packageJSON := filepath.Join(root, packageJSONFileName)

	path, exists, err := adjacentNodeModulesPath(packageJSON)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("missing adjacent node_modules was reported as present")
	}
	if want := filepath.Join(root, nodeModulesFolder); path != want {
		t.Fatalf("path was %q, want %q", path, want)
	}

	if err := os.WriteFile(path, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := adjacentNodeModulesPath(packageJSON); err == nil {
		t.Fatal("non-directory adjacent node_modules was accepted")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	got, exists, err := adjacentNodeModulesPath(packageJSON)
	if err != nil {
		t.Fatal(err)
	}
	if !exists || got != path {
		t.Fatalf("adjacent node_modules = (%q, %v), want (%q, true)", got, exists, path)
	}
}
