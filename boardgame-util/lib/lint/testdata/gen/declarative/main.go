// Command declarative regenerates the declarativeness lint's goldens,
// boardgame-util/lib/lint/testdata/declarative/<game>.json, by running the
// lint over the seven example games. Run from the repository root:
//
//	./scripts/go-local run ./boardgame-util/lib/lint/testdata/gen/declarative
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jkomoros/boardgame/boardgame-util/lib/lint"
)

func main() {
	out := filepath.Join("boardgame-util", "lib", "lint", "testdata", "declarative")
	var inputs []string
	for _, game := range lint.ExampleGames {
		inputs = append(inputs, "./"+filepath.Join("examples", game))
	}
	report := lint.Declarative(inputs, lint.DeclarativeOptions{MaxRefused: -1})
	if !report.OK {
		fmt.Fprintf(os.Stderr, "lint failed: %+v\n", report.Diagnostics)
		os.Exit(1)
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, pkg := range report.Packages {
		contents, err := lint.MarshalDeclarativePackage(pkg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.WriteFile(filepath.Join(out, pkg.Name+".json"), contents, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	fmt.Printf("wrote %d goldens to %s\n", len(report.Packages), out)
}
