package static

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClientCheckResultIsDeterministic(t *testing.T) {
	result := NewClientCheckResult([]ClientDiagnostic{
		{Source: "typescript", Code: "TS2", Severity: "error", Package: "z", File: "b.ts", Line: 2, Message: "second"},
		{Source: "boardgame", Code: "BGCLIENT001", Severity: "error", Package: "a", File: "a.ts", Line: 1, Message: "first", Remediation: "repair it"},
	})
	if result.OK || result.Version != 1 || result.Diagnostics[0].Package != "a" {
		t.Fatalf("unexpected normalized result: %#v", result)
	}
	var encoded bytes.Buffer
	if err := result.WriteJSON(&encoded); err != nil {
		t.Fatal(err)
	}
	var decoded ClientCheckResult
	if err := json.Unmarshal(encoded.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Diagnostics[0].Code != "BGCLIENT001" {
		t.Fatalf("unexpected JSON order: %s", encoded.String())
	}
	var human bytes.Buffer
	if err := result.WriteHuman(&human); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"a.ts:1: error BGCLIENT001: first", "Fix: repair it", "b.ts:2: error TS2: second"} {
		if !strings.Contains(human.String(), want) {
			t.Fatalf("human output missing %q:\n%s", want, human.String())
		}
	}
}

func TestEmptyClientCheckResultIsGreen(t *testing.T) {
	result := NewClientCheckResult(nil)
	if !result.OK || result.Diagnostics == nil {
		t.Fatalf("unexpected empty result: %#v", result)
	}
}

func TestPackageDirectoryNameDisambiguatesDuplicateNames(t *testing.T) {
	first := packageDirectoryName(ClientCheckPackage{Name: "game", ImportPath: "example.com/one/game"})
	second := packageDirectoryName(ClientCheckPackage{Name: "game", ImportPath: "example.com/two/game"})
	if first == second {
		t.Fatalf("duplicate package names shared an assembly directory: %q", first)
	}
	if !strings.HasPrefix(first, "game-") || !strings.HasPrefix(second, "game-") {
		t.Fatalf("directory names should remain recognizable: %q, %q", first, second)
	}
}

func TestCheckClientReportsPackageScopedTypeScriptDiagnostic(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	client := t.TempDir()
	if err := os.WriteFile(filepath.Join(client, "renderer.ts"), []byte("const wrong: number = 'wrong';\n"), 0600); err != nil {
		t.Fatal(err)
	}
	game := ClientCheckPackage{ImportPath: "example.com/fixture/game", Name: "game", ClientFolder: client}
	first, err := CheckClient(filepath.Join(root, "server", "static"), []ClientCheckPackage{game})
	if err != nil {
		t.Fatal(err)
	}
	second, err := CheckClient(filepath.Join(root, "server", "static"), []ClientCheckPackage{game})
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Diagnostics) != 1 {
		t.Fatalf("got %d diagnostics, want 1: %#v", len(first.Diagnostics), first.Diagnostics)
	}
	diagnostic := first.Diagnostics[0]
	if diagnostic.Package != game.ImportPath || diagnostic.Code != "TS2322" || diagnostic.File != "client/renderer.ts" {
		t.Fatalf("unexpected diagnostic: %#v", diagnostic)
	}
	if len(second.Diagnostics) != 1 || second.Diagnostics[0] != diagnostic {
		t.Fatalf("diagnostics were not deterministic:\nfirst: %#v\nsecond: %#v", first.Diagnostics, second.Diagnostics)
	}
}

func TestCheckClientInfrastructureErrorsHaveStableCode(t *testing.T) {
	report, err := CheckClient(t.TempDir(), nil)
	if err == nil || !strings.HasPrefix(err.Error(), "BGCLIENT0001:") {
		t.Fatalf("got report %#v and error %v, want BGCLIENT0001", report, err)
	}
}
