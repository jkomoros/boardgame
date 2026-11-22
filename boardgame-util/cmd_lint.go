package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bobziuchkovski/writ"
	lintpkg "github.com/jkomoros/boardgame/boardgame-util/lib/lint"
)

type lintCmd struct {
	baseSubCommand
	JSON        bool
	Fix         bool
	Declarative bool
	MaxRefused  int
}

func (l *lintCmd) Name() string { return "lint" }

func (l *lintCmd) Description() string {
	return "Preflights game packages and reports actionable authoring errors"
}

func (l *lintCmd) HelpText() string {
	return l.Name() + ` validates game packages before they are served or deployed.
It checks package shape and deterministic randomness, verifies generated Go
readers and enums without rewriting them, then constructs a real GameManager
to exercise stack schemas, constraints, moves, phases, legal preconditions,
and progression configuration. All independent diagnostics are reported in a
stable order and any diagnostic produces a nonzero exit status.

PKG may be an import path, directory, or go-list pattern such as ./.... If no
PKG is supplied, the current package is checked. Non-game packages inside a
pattern are ignored. Pass --fix to safely refresh boardgame-owned generated
Go files before runtime validation.

With --declarative, lint instead reports how declarative each game is and how
much of its code the Go-to-TypeScript translator would take: the move configs
that are data and those with game code, the declared precondition specs, the
moves whose legality is an imperative Legal override, every function reaching
engine output with its verdict and refusals, and constructs with no single
answer (range over a map, the global math/rand, ValueFromString). It writes
nothing and exits 0, unless --max-refused N is given and more than N
functions are refused.`
}

func (l *lintCmd) Usage() string { return "[PKG]..." }

func (l *lintCmd) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"json"},
			Description: "Emit one deterministic machine-readable report.",
			Decoder:     writ.NewFlagDecoder(&l.JSON),
			Flag:        true,
		},
		{
			Names:       []string{"fix"},
			Description: "Safely refresh stale boardgame-generated Go files before validation.",
			Decoder:     writ.NewFlagDecoder(&l.Fix),
			Flag:        true,
		},
		{
			Names:       []string{"declarative"},
			Description: "Report declarativeness and translator verdicts instead of preflighting.",
			Decoder:     writ.NewFlagDecoder(&l.Declarative),
			Flag:        true,
		},
		{
			Names:       []string{"max-refused"},
			Description: "With --declarative, exit nonzero when more than this many functions are refused.",
			Decoder:     writ.NewDefaulter(writ.NewOptionDecoder(&l.MaxRefused), "-1"),
			Placeholder: "N",
		},
	}
}

func (l *lintCmd) Run(_ writ.Path, positional []string) {
	if l.Declarative {
		l.runDeclarative(positional)
		return
	}
	report := lintpkg.Check(positional, lintpkg.Options{Fix: l.Fix})
	if l.JSON {
		_ = json.NewEncoder(os.Stdout).Encode(report)
	} else if report.OK {
		fmt.Printf("Game author preflight passed for %d package(s)\n", len(report.Packages))
	} else {
		for _, diagnostic := range report.Diagnostics {
			location := diagnostic.Package
			if diagnostic.File != "" {
				fileLocation := diagnostic.File
				if diagnostic.Line > 0 {
					fileLocation += fmt.Sprintf(":%d", diagnostic.Line)
					if diagnostic.Column > 0 {
						fileLocation += fmt.Sprintf(":%d", diagnostic.Column)
					}
				}
				if location != "" {
					location += ": "
				}
				location += fileLocation
			}
			if location != "" {
				location += ": "
			}
			fmt.Printf("%s%s %s\n", location, diagnostic.Code, diagnostic.Message)
			if diagnostic.Remediation != "" {
				fmt.Printf("  Fix: %s\n", diagnostic.Remediation)
			}
		}
		fmt.Printf("Game author preflight failed with %d diagnostic(s) across %d package(s)\n", len(report.Diagnostics), len(report.Packages))
	}
	if !report.OK {
		quit(1)
	}
}

func (l *lintCmd) runDeclarative(positional []string) {
	report := lintpkg.Declarative(positional, lintpkg.DeclarativeOptions{MaxRefused: l.MaxRefused})
	if l.JSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(report)
	} else {
		lintpkg.WriteDeclarativeText(os.Stdout, report)
	}
	if !report.OK {
		quit(1)
	}
}
