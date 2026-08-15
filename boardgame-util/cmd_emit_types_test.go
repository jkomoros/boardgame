package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestShippedClientTypesCarryEnumValuesInDeclaredOrder checks the committed
// output of the whole pipeline -- enum.Values, the extractor binary, and the
// TypeScript generator -- rather than any single link in it. blackjack's Rank
// enum is the useful case because its declared order (Unknown, Ace, 2 .. King,
// Joker) matches no sort a downstream stage might apply by accident, so this
// fails if any stage sorts, drops, or reorders.
func TestShippedClientTypesCarryEnumValuesInDeclaredOrder(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join("..", "examples", "blackjack", "client", "_types.ts"))
	if err != nil {
		t.Fatal(err)
	}

	want := `export const RankValues = [
  { Key: 0, Value: "Unknown", Label: "Unknown" },
  { Key: 1, Value: "Ace", Label: "Ace" },
  { Key: 2, Value: "2", Label: "2" },
  { Key: 3, Value: "3", Label: "3" },
  { Key: 4, Value: "4", Label: "4" },
  { Key: 5, Value: "5", Label: "5" },
  { Key: 6, Value: "6", Label: "6" },
  { Key: 7, Value: "7", Label: "7" },
  { Key: 8, Value: "8", Label: "8" },
  { Key: 9, Value: "9", Label: "9" },
  { Key: 10, Value: "10", Label: "10" },
  { Key: 11, Value: "Jack", Label: "Jack" },
  { Key: 12, Value: "Queen", Label: "Queen" },
  { Key: 13, Value: "King", Label: "King" },
  { Key: 14, Value: "Joker", Label: "Joker" },
] as const satisfies readonly EnumValueInfo<RankValue>[];`

	if !strings.Contains(string(contents), want) {
		t.Fatalf("blackjack's shipped enum metadata is missing, empty, or misordered; regenerate with boardgame-util emit-types.\nwant block:\n%s", want)
	}
}

// TestShippedClientTypesOmitPresentationNobodyAttached is the other half of the
// contract: the framework's own examples attach no presentation, so their
// generated types must carry no presentation keys at all.
func TestShippedClientTypesOmitPresentationNobodyAttached(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join("..", "examples", "checkers", "client", "_types.ts"))
	if err != nil {
		t.Fatal(err)
	}
	// Keys are the enum's real keys, not the list index: checkers takes Red and
	// Black from the shared color palette, where Black is 4.
	if !strings.Contains(string(contents), `export const ColorValues = [
  { Key: 0, Value: "Red", Label: "Red" },
  { Key: 4, Value: "Black", Label: "Black" },
]`) {
		t.Fatalf("checkers' color metadata is not the expected clean round trip:\n%s", contents)
	}
	for _, unwanted := range []string{"Description:", "CSSColor:", "Art:"} {
		if strings.Contains(string(contents), unwanted) {
			t.Errorf("generated %q for a game that attached no presentation", unwanted)
		}
	}
}

func TestInstallGeneratedGameTypesPreservesOldFilesWhenStagingFails(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "a-game", "client", "_types.ts")
	if err := os.MkdirAll(filepath.Dir(oldPath), 0700); err != nil {
		t.Fatal(err)
	}
	const oldContents = "old complete generation"
	if err := os.WriteFile(oldPath, []byte(oldContents), 0600); err != nil {
		t.Fatal(err)
	}

	invalidParent := filepath.Join(dir, "z-not-a-directory")
	if err := os.WriteFile(invalidParent, []byte("sentinel"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := installGeneratedGameTypes([]generatedGameTypeFile{
		{path: oldPath, contents: []byte("new generation"), gameName: "a-game"},
		{path: filepath.Join(invalidParent, "_game_renderer.ts"), contents: []byte("new renderer"), gameName: "z-missing"},
	})
	if err == nil {
		t.Fatal("installGeneratedGameTypes() succeeded, want staging error")
	}
	got, readErr := os.ReadFile(oldPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != oldContents {
		t.Fatalf("old output changed on staging failure: got %q", got)
	}
}

func TestCheckGeneratedGameTypesIsNonMutating(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "_types.ts")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := checkGeneratedGameTypes([]generatedGameTypeFile{{path: path, contents: []byte("new")}}); err == nil {
		t.Fatal("checkGeneratedGameTypes() succeeded for stale output")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "old" {
		t.Fatalf("freshness check mutated output: %q", contents)
	}
	if err := checkGeneratedGameTypes([]generatedGameTypeFile{{path: path, contents: []byte("old")}}); err != nil {
		t.Fatalf("current output failed check: %v", err)
	}
}

func TestInstallGeneratedGameTypesReplacesCompletePair(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "client")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	typesPath := filepath.Join(dir, "_types.ts")
	rendererPath := filepath.Join(dir, "_game_renderer.ts")
	if err := installGeneratedGameTypes([]generatedGameTypeFile{
		{path: typesPath, contents: []byte("new types"), gameName: "game"},
		{path: rendererPath, contents: []byte("new renderer"), gameName: "game"},
	}); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{typesPath: "new types", rendererPath: "new renderer"} {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("%s = %q, want %q", path, got, want)
		}
	}
}
