package gamepkg

import (
	"strings"
	"testing"
)

func TestAnalyzeTypedDelegateContracts(t *testing.T) {
	tests := []struct {
		name        string
		packagePath string
		wantValid   bool
		wantMessage string
	}{
		{name: "valid build-tag package", packagePath: "buildtag", wantValid: true},
		{name: "wrong return", packagePath: "wrongreturn", wantMessage: "does not implement boardgame.GameDelegate"},
		{name: "method", packagePath: "methoddelegate", wantMessage: "must be a package-level function"},
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
