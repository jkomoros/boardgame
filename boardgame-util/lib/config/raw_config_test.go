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

func TestConfigSaveWritesPublicAndSecretWithSafeModes(t *testing.T) {
	dir := t.TempDir()
	publicPath := filepath.Join(dir, publicConfigFileName)
	secretPath := filepath.Join(dir, privateConfigFileName)
	config := &Config{
		rawPublicConfig: &RawConfig{Base: &RawConfigMode{}, path: publicPath},
		rawSecretConfig: &RawConfig{Base: &RawConfigMode{}, path: secretPath},
	}
	if err := config.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	for path, want := range map[string]os.FileMode{publicPath: 0o644, secretPath: 0o600} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != want {
			t.Errorf("%s mode = %o, want %o", path, got, want)
		}
	}
}

func TestConfigSaveTightensExistingSecretPermissions(t *testing.T) {
	dir := t.TempDir()
	secretPath := filepath.Join(dir, privateConfigFileName)
	if err := os.WriteFile(secretPath, []byte("old-secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	config := &Config{rawSecretConfig: &RawConfig{Base: &RawConfigMode{}, path: secretPath}}
	if err := config.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(secretPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("secret mode = %o, want 600", got)
	}
}

func TestConfigSavePreflightFailureLeavesPublicConfigUntouched(t *testing.T) {
	dir := t.TempDir()
	publicPath := filepath.Join(dir, publicConfigFileName)
	secretPath := filepath.Join(dir, privateConfigFileName)
	if err := os.WriteFile(publicPath, []byte("old-public"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(secretPath, 0o700); err != nil {
		t.Fatal(err)
	}
	config := &Config{
		rawPublicConfig: &RawConfig{Base: &RawConfigMode{}, path: publicPath},
		rawSecretConfig: &RawConfig{Base: &RawConfigMode{}, path: secretPath},
	}
	if err := config.Save(); err == nil {
		t.Fatal("Save succeeded with a directory at the private config path")
	}
	contents, err := os.ReadFile(publicPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "old-public" {
		t.Fatalf("public config changed despite pair preflight failure: %q", contents)
	}
}
