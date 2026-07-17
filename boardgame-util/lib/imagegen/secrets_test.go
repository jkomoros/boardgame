package imagegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveAPIKeyFromDevSecretsJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.SECRET.json")
	if err := os.WriteFile(path, []byte(`{"dev":{"gemini_api_key":"dev-key"},"prod":{"gemini_api_key":"prod-key"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	key, err := ResolveAPIKey("", path, "dev.gemini_api_key")
	if err != nil {
		t.Fatal(err)
	}
	if key != "dev-key" {
		t.Fatalf("key = %q", key)
	}
}

func TestResolveAPIKeyPrefersEnvironmentAndNeverLeaks(t *testing.T) {
	key, err := ResolveAPIKey(" env-key ", "missing", "dev.gemini_api_key")
	if err != nil || key != "env-key" {
		t.Fatalf("key=%q err=%v", key, err)
	}
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte(`{"dev":{"gemini_api_key":"super-secret"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = ResolveAPIKey("", path, "missing.field")
	if err == nil || strings.Contains(err.Error(), "super-secret") {
		t.Fatalf("unsafe error: %v", err)
	}
}
