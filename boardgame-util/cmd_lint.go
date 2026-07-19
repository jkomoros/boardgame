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
	JSON bool
	Fix  bool
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
Go files before runtime validation.`
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
	}
}

func (l *lintCmd) Run(_ writ.Path, positional []string) {
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
