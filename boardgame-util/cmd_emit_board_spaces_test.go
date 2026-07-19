package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAndGenerateAuthoredBoardSpaces(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.svg")
	source := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <path data-board-space="study:east" data-board-label="East Study" data-board-order="1" d="M0 0h1v1z"/>
  <rect data-board-space="2" data-board-label="Library" data-board-order="0" x="0" y="0" width="1" height="1"/>
</svg>`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	spaces, err := parseAuthoredBoardSpaces(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(spaces) != 2 || spaces[0].Key != "2" || spaces[1].Key != "study:east" {
		t.Fatalf("unexpected spaces: %#v", spaces)
	}
	generated := string(generateBoardSpacesTypeScript(path, spaces))
	for _, want := range []string{
		`export const BoardSpaceKeys = ["2", "study:east"] as const`,
		`export type BoardSpaceKey = (typeof BoardSpaceKeys)[number]`,
		`"study:east": "East Study"`,
	} {
		if !strings.Contains(generated, want) {
			t.Fatalf("generated contract did not contain %q:\n%s", want, generated)
		}
	}
}

func TestAuthoredBoardSpaceExtractionFailsLoudly(t *testing.T) {
	for name, body := range map[string]string{
		"missing label":   `<rect data-board-space="1"/>`,
		"duplicate key":   `<rect data-board-space="1" data-board-label="A"/><rect data-board-space="1" data-board-label="B"/>`,
		"duplicate order": `<rect data-board-space="1" data-board-label="A" data-board-order="2"/><rect data-board-space="2" data-board-label="B" data-board-order="2"/>`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "board.svg")
			source := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">` + body + `</svg>`
			if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := parseAuthoredBoardSpaces(path); err == nil {
				t.Fatal("expected extraction to fail")
			}
		})
	}
}

func TestAuthoredBoardSpaceNumericKeysAreExplicitAndSafe(t *testing.T) {
	for name, key := range map[string]string{"unsafe": "9007199254740992", "noncanonical": "02", "not number": "study"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "board.svg")
			source := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10" data-board-key-type="number"><rect data-board-space="` + key + `" data-board-label="Room"/></svg>`
			if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := parseAuthoredBoardSpaces(path); err == nil {
				t.Fatal("expected unsafe numeric key to fail")
			}
		})
	}
	path := filepath.Join(t.TempDir(), "board.svg")
	source := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10" data-board-key-type="number"><rect data-board-space="2"><title>Library</title></rect></svg>`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	spaces, err := parseAuthoredBoardSpaces(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(generateBoardSpacesTypeScript(path, spaces)); !strings.Contains(got, `BoardSpaceKeys = [2]`) {
		t.Fatalf("expected explicit numeric literal:\n%s", got)
	}
}

func TestOrphanBoardSpaceContractsOnlyClaimsGeneratedFiles(t *testing.T) {
	dir := t.TempDir()
	orphan := filepath.Join(dir, "_old_spaces.ts")
	creatorFile := filepath.Join(dir, "_custom_spaces.ts")
	if err := os.WriteFile(orphan, []byte(generatedBoardSpaceHeader+"old.svg. DO NOT EDIT. */\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(creatorFile, []byte("export const custom = true;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	found, err := orphanBoardSpaceContracts(dir, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0] != orphan {
		t.Fatalf("unexpected orphan set: %v", found)
	}
}

func TestCollectGeneratedBoardSpacesRejectsOutputCollisions(t *testing.T) {
	dir := t.TempDir()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><rect data-board-space="a" data-board-label="A"/></svg>`
	for _, name := range []string{"foo-bar.svg", "foo_bar.svg"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(svg), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := collectGeneratedBoardSpaces(dir, map[string]string{}); err == nil || !strings.Contains(err.Error(), "output collision") {
		t.Fatalf("expected an explicit output collision, got %v", err)
	}
}

func TestAuthoredBoardSpaceSizeLimitIsEnforced(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.svg")
	oversized := strings.Repeat(" ", maxAuthoredBoardBytes+1)
	if err := os.WriteFile(path, []byte(oversized), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := parseAuthoredBoardSpaces(path); err == nil || !strings.Contains(err.Error(), "byte limit") {
		t.Fatalf("expected size-limit failure, got %v", err)
	}
}

func TestInstallGeneratedBoardSpacesIsOneTransaction(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "a_spaces.ts")
	orphan := filepath.Join(dir, "b_spaces.ts")
	invalidParent := filepath.Join(dir, "z-not-a-directory")
	for path, contents := range map[string]string{
		existing: "old", orphan: "orphan", invalidParent: "sentinel",
	} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	err := installGeneratedBoardSpaces([]generatedBoardSpacesFile{
		{path: existing, contents: []byte("new")},
		{path: filepath.Join(invalidParent, "later_spaces.ts"), contents: []byte("later")},
	}, []string{orphan}, false)
	if err == nil {
		t.Fatal("install succeeded despite invalid destination")
	}
	for path, want := range map[string]string{existing: "old", orphan: "orphan"} {
		got, readErr := os.ReadFile(path)
		if readErr != nil || string(got) != want {
			t.Fatalf("%s after failed transaction = %q, %v; want %q", path, got, readErr, want)
		}
	}
}

func TestCheckGeneratedBoardSpacesReportsAllStalePaths(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, "a_spaces.ts")
	orphan := filepath.Join(dir, "b_spaces.ts")
	for _, path := range []string{stale, orphan} {
		if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	err := installGeneratedBoardSpaces(
		[]generatedBoardSpacesFile{{path: stale, contents: []byte("new")}},
		[]string{orphan}, true,
	)
	if err == nil || !strings.Contains(err.Error(), stale+", "+orphan) {
		t.Fatalf("error = %v, want complete sorted stale set", err)
	}
}
