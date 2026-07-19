package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRawConfigCreateOnlyToleratesMissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.json")
	config, err := NewRawConfig(missing, true)
	if err != nil {
		t.Fatalf("NewRawConfig(missing, true): %v", err)
	}
	if config.Path() != missing {
		t.Fatalf("config path = %q, want %q", config.Path(), missing)
	}

	_, err = NewRawConfig(t.TempDir(), true)
	if err == nil || !strings.Contains(err.Error(), "couldn't read config file") {
		t.Fatalf("NewRawConfig(directory, true) error = %v, want read error", err)
	}
}

func TestRawConfigSaveAtomicallyReplacesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("old"), 0o640); err != nil {
		t.Fatal(err)
	}
	config := &RawConfig{
		Base: &RawConfigMode{},
		path: path,
	}
	if err := config.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), `"base"`) {
		t.Fatalf("saved contents = %q, want serialized base config", contents)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o640 {
		t.Fatalf("saved permissions = %o, want 640", got)
	}
}
