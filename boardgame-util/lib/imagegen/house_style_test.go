package imagegen

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAddAndResolveHouseStyle(t *testing.T) {
	dir := t.TempDir()
	selected := filepath.Join(dir, "selected.png")
	locked := filepath.Join(dir, "locked.png")
	brief := filepath.Join(dir, "brief.md")
	if err := writeTestPNG(selected); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(brief, []byte("Original graphite field archive."), 0o644); err != nil {
		t.Fatal(err)
	}
	now := func() time.Time { return time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC) }
	if _, err := CreateStyleLock(selected, locked, false, now); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(dir, "catalog")
	style, err := AddHouseStyle(root, "Field Archive", "field-archive", brief, locked+".style-lock.json", false, now)
	if err != nil {
		t.Fatal(err)
	}
	if style.Name != "Field Archive" {
		t.Fatalf("style = %#v", style)
	}
	resolved, err := ResolveHouseStyle("field-archive", root)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.ImageSHA256 != style.ImageSHA256 {
		t.Fatal("resolved house style hash mismatch")
	}
	if _, err := os.Stat(filepath.Join(root, "catalog.json")); err != nil {
		t.Fatal(err)
	}
	staleManifest := filepath.Join(root, "field-archive", "style.png.imagegen.json")
	if err := os.WriteFile(staleManifest, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := AddHouseStyle(root, "Field Archive", "field-archive", brief, locked+".style-lock.json", true, now); err != nil {
		t.Fatalf("force replacement: %v", err)
	}
	if _, err := os.Stat(staleManifest); !os.IsNotExist(err) {
		t.Fatalf("stale copied source manifest remains after replacement: %v", err)
	}
}

func TestHouseStyleRejectsUnsafeSlug(t *testing.T) {
	if _, err := AddHouseStyle(t.TempDir(), "Bad", "../bad", "missing", "missing", false, nil); err == nil {
		t.Fatal("expected unsafe slug rejection")
	}
}

func TestAddHouseStylePreflightFailureCreatesNoStyleFiles(t *testing.T) {
	dir := t.TempDir()
	selected := filepath.Join(dir, "selected.png")
	locked := filepath.Join(dir, "locked.png")
	brief := filepath.Join(dir, "brief.md")
	if err := writeTestPNG(selected); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(brief, []byte("brief"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateStyleLock(selected, locked, false, nil); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(dir, "catalog")
	if err := os.MkdirAll(filepath.Join(root, "catalog.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := AddHouseStyle(root, "Field Archive", "field-archive", brief, locked+".style-lock.json", false, nil); err == nil {
		t.Fatal("AddHouseStyle succeeded with a directory at catalog.json")
	}
	if _, err := os.Stat(filepath.Join(root, "field-archive")); !os.IsNotExist(err) {
		t.Fatalf("style directory was created despite transaction preflight failure: %v", err)
	}
}
