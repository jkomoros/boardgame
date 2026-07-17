package imagegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStyleFunnelAlwaysOffersFourCandidates(t *testing.T) {
	for name, candidates := range map[string][]StyleCandidate{
		"explore": ExploreStyles("brief"), "iterate": IterateStyles("refine"),
	} {
		if len(candidates) != 4 {
			t.Fatalf("%s candidates = %d, want 4", name, len(candidates))
		}
		seen := map[string]bool{}
		for _, candidate := range candidates {
			if seen[candidate.Slug] || !strings.Contains(candidate.Prompt, nameAxis(name)) {
				t.Fatalf("bad %s candidate: %#v", name, candidate)
			}
			seen[candidate.Slug] = true
		}
	}
}

func nameAxis(name string) string {
	if name == "explore" {
		return "STYLE AXIS"
	}
	return "ITERATION AXIS"
}

func TestCreateAndReadStyleLock(t *testing.T) {
	dir := t.TempDir()
	selected := filepath.Join(dir, "candidate.png")
	output := filepath.Join(dir, "locked.png")
	if err := os.WriteFile(selected, []byte("image"), 0o644); err != nil {
		t.Fatal(err)
	}
	now := func() time.Time { return time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC) }
	lock, err := CreateStyleLock(selected, output, false, now)
	if err != nil {
		t.Fatal(err)
	}
	read, err := ReadStyleLock(output + ".style-lock.json")
	if err != nil {
		t.Fatal(err)
	}
	if read.ImageSHA256 != lock.ImageSHA256 || read.Image != output {
		t.Fatalf("lock mismatch: %#v %#v", lock, read)
	}
}

func TestGalleryListsEveryCandidate(t *testing.T) {
	dir := t.TempDir()
	candidates := ExploreStyles("brief")
	if err := WriteGallery(dir, "Round one", candidates); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "gallery.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range candidates {
		if !strings.Contains(string(data), candidate.Slug+".png") {
			t.Errorf("gallery missing %s", candidate.Slug)
		}
	}
}
