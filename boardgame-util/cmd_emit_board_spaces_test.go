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
		`export const BoardSpaceKeys = [2, "study:east"] as const`,
		`export type BoardSpaceKey = (typeof BoardSpaceKeys)[number]`,
		`"study:east": "East Study"`,
	} {
		if !strings.Contains(generated, want) {
			t.Fatalf("generated contract did not contain %q:\n%s", want, generated)
		}
	}
}

func TestAuthoredBoardSpaceExtractionFailsLoudly(t *testing.T) {
	for name, source := range map[string]string{
		"missing label":   `<svg><rect data-board-space="1"/></svg>`,
		"duplicate key":   `<svg><rect data-board-space="1" data-board-label="A"/><rect data-board-space="1" data-board-label="B"/></svg>`,
		"duplicate order": `<svg><rect data-board-space="1" data-board-label="A" data-board-order="2"/><rect data-board-space="2" data-board-label="B" data-board-order="2"/></svg>`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "board.svg")
			if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := parseAuthoredBoardSpaces(path); err == nil {
				t.Fatal("expected extraction to fail")
			}
		})
	}
}
