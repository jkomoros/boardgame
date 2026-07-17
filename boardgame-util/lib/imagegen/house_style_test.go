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
	if err := os.WriteFile(selected, []byte("style image"), 0o644); err != nil {
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
}

func TestHouseStyleRejectsUnsafeSlug(t *testing.T) {
	if _, err := AddHouseStyle(t.TempDir(), "Bad", "../bad", "missing", "missing", false, nil); err == nil {
		t.Fatal("expected unsafe slug rejection")
	}
}
