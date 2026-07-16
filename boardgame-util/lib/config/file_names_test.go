package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileNamesWalksActualParentsFromNestedDirectory(t *testing.T) {
	root := t.TempDir()
	public := filepath.Join(root, "config.json")
	if err := os.WriteFile(public, []byte("{}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "examples", "game")
	if err := os.MkdirAll(nested, 0700); err != nil {
		t.Fatal(err)
	}

	gotPublic, gotPrivate, err := FileNames(nested, false)
	if err != nil {
		t.Fatal(err)
	}
	if gotPublic != public {
		t.Fatalf("public config = %q, want %q", gotPublic, public)
	}
	if gotPrivate != "" {
		t.Fatalf("private config = %q, want none", gotPrivate)
	}
}

func TestFileNamesTerminatesWhenNoAncestorHasConfig(t *testing.T) {
	_, _, err := FileNames(t.TempDir(), false)
	if err == nil {
		t.Fatal("config-less ancestor walk unexpectedly succeeded")
	}
}
