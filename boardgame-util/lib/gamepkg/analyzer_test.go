package gamepkg

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestAnalyzeTypedDelegateContracts(t *testing.T) {
	tests := []struct {
		name        string
		packagePath string
		wantValid   bool
		wantMessage string
	}{
		{name: "valid build-tag package", packagePath: "buildtag", wantValid: true},
		{name: "alias return", packagePath: "aliasreturn", wantValid: true},
		{name: "transitive concrete return", packagePath: "transitivereturn", wantValid: true},
		{name: "wrong return", packagePath: "wrongreturn", wantMessage: "does not implement boardgame.GameDelegate"},
		{name: "parameters", packagePath: "delegateparams", wantMessage: "must not accept parameters"},
		{name: "compile error", packagePath: "compileerror", wantMessage: "undefined: deliberatelyMissing"},
		{name: "syntax error", packagePath: "syntaxerror", wantMessage: "expected"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			analyses, err := Analyze([]string{"./testdata/" + test.packagePath}, "", Options{ReadOnly: true})
			if err != nil {
				t.Fatal(err)
			}
			if len(analyses) != 1 {
				t.Fatalf("got %d analyses, want 1", len(analyses))
			}
			analysis := analyses[0]
			if !analysis.Candidate {
				t.Fatal("package containing NewDelegate was not classified as a candidate")
			}
			if analysis.ValidGame() != test.wantValid {
				t.Fatalf("ValidGame = %v, diagnostics: %+v", analysis.ValidGame(), analysis.Diagnostics)
			}
			if test.wantMessage == "" {
				return
			}
			var found bool
			for _, diagnostic := range analysis.Diagnostics {
				if strings.Contains(diagnostic.Message, test.wantMessage) {
					found = true
					if diagnostic.Kind == "" {
						t.Errorf("diagnostic did not include a stable kind: %+v", diagnostic)
					}
					if diagnostic.Position.File == "" || diagnostic.Position.Line == 0 {
						t.Errorf("diagnostic did not include a source position: %+v", diagnostic)
					}
				}
			}
			if !found {
				t.Fatalf("diagnostics %+v did not contain %q", analysis.Diagnostics, test.wantMessage)
			}
		})
	}
}

func TestAnalyzeIgnoresMethodOnlyNearMisses(t *testing.T) {
	for _, packagePath := range []string{"methoddelegate", "syntaxordinary"} {
		analyses, err := Analyze([]string{"./testdata/" + packagePath}, "", Options{ReadOnly: true})
		if err != nil {
			t.Fatal(err)
		}
		if len(analyses) != 1 || analyses[0].Candidate {
			t.Fatalf("%s was classified as a game candidate: %+v", packagePath, analyses)
		}
	}
}

func TestAnalysisGamePackageRejectsSyntheticValues(t *testing.T) {
	_, err := (Analysis{ImportPath: "example/game", Name: "game", Dir: t.TempDir(), Candidate: true, Diagnostics: []Diagnostic{}}).GamePackage()
	if err == nil {
		t.Fatal("caller-constructed Analysis bypassed package validation")
	}
}

func TestPackageDirPrefersAuthoritativeDirectory(t *testing.T) {
	want := filepath.Join(t.TempDir(), "source")
	got := packageDir(&packages.Package{Dir: want, CompiledGoFiles: []string{"/tmp/cgo-generated.go"}})
	if got != want {
		t.Fatalf("packageDir = %q, want authoritative Dir %q", got, want)
	}
}

func TestAnalyzeDistinguishesOrdinaryPackage(t *testing.T) {
	analyses, err := Analyze([]string{"."}, "", Options{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(analyses) != 1 {
		t.Fatalf("got %d analyses, want 1", len(analyses))
	}
	if analyses[0].Candidate || analyses[0].ValidGame() {
		t.Fatalf("gamepkg package classified as a game: %+v", analyses[0])
	}
}

func TestAnalyzeDeduplicatesOverlappingPatterns(t *testing.T) {
	analyses, err := Analyze([]string{"../../../examples/pig", "../../../examples/pig"}, "", Options{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(analyses) != 1 || !analyses[0].ValidGame() {
		t.Fatalf("overlapping analysis = %+v, want one valid Pig package", analyses)
	}
}

func TestAnalyzeReadOnlyDoesNotModifyModuleFiles(t *testing.T) {
	moduleRoot := filepath.Clean("../../..")
	paths := []string{filepath.Join(moduleRoot, "go.mod"), filepath.Join(moduleRoot, "go.sum")}
	before := make(map[string][]byte, len(paths))
	for _, path := range paths {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before[path] = contents
	}
	if _, err := Analyze([]string{"../../../examples/pig"}, "", Options{ReadOnly: true}); err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before[path], after) {
			t.Fatalf("read-only analysis modified %s", path)
		}
	}
}

func TestParsePosition(t *testing.T) {
	for _, test := range []struct {
		input string
		want  Position
	}{
		{"game.go:12:7", Position{File: "game.go", Line: 12, Column: 7}},
		{"game.go:12", Position{File: "game.go", Line: 12}},
		{"C:\\game.go:12:7", Position{File: "C:\\game.go", Line: 12, Column: 7}},
	} {
		if got := parsePosition(test.input); got != test.want {
			t.Errorf("parsePosition(%q) = %+v, want %+v", test.input, got, test.want)
		}
	}
}
