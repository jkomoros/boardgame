package imagegen

import (
	"image"
	"image/color"
	"image/png"
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
	if err := writeTestPNG(selected); err != nil {
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

func TestCreateStyleLockPreflightFailurePublishesNothing(t *testing.T) {
	dir := t.TempDir()
	selected := filepath.Join(dir, "candidate.png")
	output := filepath.Join(dir, "locked.png")
	sidecar := output + ".style-lock.json"
	if err := writeTestPNG(selected); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sidecar, []byte("creator-owned"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateStyleLock(selected, output, false, nil); err == nil {
		t.Fatal("CreateStyleLock succeeded despite an existing sidecar")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("style image was published despite sidecar preflight failure")
	}
	contents, err := os.ReadFile(sidecar)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "creator-owned" {
		t.Fatalf("existing sidecar changed: %q", contents)
	}
}

func writeTestPNG(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	image := image.NewNRGBA(image.Rect(0, 0, 3, 3))
	image.SetNRGBA(1, 1, color.NRGBA{R: 255, A: 255})
	encodeErr := png.Encode(file, image)
	closeErr := file.Close()
	if encodeErr != nil {
		return encodeErr
	}
	return closeErr
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
