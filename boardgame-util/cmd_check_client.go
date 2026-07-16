package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bobziuchkovski/writ"
	staticbuild "github.com/jkomoros/boardgame/boardgame-util/lib/build/static"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

type checkClient struct {
	baseSubCommand
	JSON bool
	Fix  bool
}

func (c *checkClient) Name() string { return "check-client" }

func (c *checkClient) Description() string {
	return "Strictly type-checks each configured game's client in isolation"
}

func (c *checkClient) HelpText() string {
	return c.Name() + ` runs the framework's pinned strict TypeScript checker once,
then checks every configured game client in an isolated assembled package. It
is fatal when any diagnostic is reported and is suitable for local and CI use.
It also verifies that move names, move inputs, expanded state types, and the
bound renderer base are current without rewriting creator files. Creator code
is also checked for unsafe TypeScript escape hatches and for invalid Lit
property, attribute, and event bindings. Pass --fix to transactionally refresh
the complete generated surface before checking creator code.`
}

func (c *checkClient) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"json"},
			Description: "Emit one deterministic machine-readable diagnostic document.",
			Decoder:     writ.NewFlagDecoder(&c.JSON),
			Flag:        true,
		},
		{
			Names:       []string{"fix"},
			Description: "Transactionally refresh all generated client contracts before checking.",
			Decoder:     writ.NewFlagDecoder(&c.Fix),
			Flag:        true,
		},
	}
}

func (c *checkClient) Run(p writ.Path, positional []string) {
	frameworkDir, err := frameworkStaticDirectory()
	if err != nil {
		c.finish(staticbuild.ClientCheckReport{Version: staticbuild.ClientCheckVersion, Diagnostics: []staticbuild.ClientDiagnostic{}}, err)
		return
	}
	packages, err := c.Base().GetConfig(false).Dev.AllGamePackages()
	if err != nil {
		c.finish(staticbuild.ClientCheckReport{Version: staticbuild.ClientCheckVersion, Diagnostics: []staticbuild.ClientDiagnostic{}}, fmt.Errorf("BGCLIENT0001: couldn't load configured game packages: %w", err))
		return
	}
	contractsRefreshed := false
	if c.Fix {
		generated, generationErr := generateCompleteClientContracts(c.Base(), packages, false)
		if generationErr == nil {
			generationErr = generated.install()
		}
		if generationErr != nil {
			c.finish(staticbuild.ClientCheckReport{Version: staticbuild.ClientCheckVersion, Diagnostics: []staticbuild.ClientDiagnostic{}}, fmt.Errorf("BGCLIENT0001: couldn't refresh generated client contracts: %w", generationErr))
			return
		}
		contractsRefreshed = true
	}
	inputs := make([]staticbuild.ClientCheckPackage, 0, len(packages))
	for _, pkg := range packages {
		inputs = append(inputs, staticbuild.ClientCheckPackage{ImportPath: pkg.Import(), Name: pkg.Name(), ClientFolder: pkg.ClientFolder()})
	}
	report, err := staticbuild.CheckClient(frameworkDir, inputs)
	if err == nil && !contractsRefreshed {
		if freshnessErr := checkGeneratedClientContracts(c.Base(), packages); freshnessErr != nil {
			if isStaleGeneratedClientContracts(freshnessErr) {
				paths := generatedClientContractStalePaths(freshnessErr)
				if len(paths) == 0 {
					report = staticbuild.NewClientCheckResult(append(report.Diagnostics, staticbuild.ClientDiagnostic{
						Source: "boardgame", Code: "BGCLIENT0002", Severity: "error",
						Message: freshnessErr.Error(), Remediation: "Run boardgame-util check-client --fix and commit every generated client contract.",
					}))
				} else {
					for _, path := range paths {
						report.Diagnostics = append(report.Diagnostics, staticbuild.ClientDiagnostic{
							Source: "boardgame", Code: "BGCLIENT0002", Severity: "error", File: path,
							Message: "generated client contract is stale or orphaned", Remediation: "Run boardgame-util check-client --fix and commit every generated client contract.",
						})
					}
					report = staticbuild.NewClientCheckResult(report.Diagnostics)
				}
			} else {
				err = fmt.Errorf("BGCLIENT0001: generated-contract extraction or validation failed: %w", freshnessErr)
			}
		}
	}
	c.finish(report, err)
}

func checkGeneratedClientContracts(base *boardgameUtil, packages []*gamepkg.Pkg) error {
	generated, err := generateCompleteClientContracts(base, packages, true)
	if err != nil {
		return fmt.Errorf("couldn't extract or validate complete generated contracts: %w", err)
	}
	return generated.check()
}

func (c *checkClient) finish(report staticbuild.ClientCheckReport, infrastructureErr error) {
	failed := infrastructureErr != nil || len(report.Diagnostics) > 0
	if c.JSON {
		document := struct {
			Version             int                            `json:"version"`
			OK                  bool                           `json:"ok"`
			Diagnostics         []staticbuild.ClientDiagnostic `json:"diagnostics"`
			InfrastructureError string                         `json:"infrastructureError,omitempty"`
		}{Version: report.Version, OK: !failed, Diagnostics: report.Diagnostics}
		if infrastructureErr != nil {
			document.InfrastructureError = infrastructureErr.Error()
		}
		_ = json.NewEncoder(os.Stdout).Encode(document)
	} else if infrastructureErr != nil {
		fmt.Println(infrastructureErr)
	} else if len(report.Diagnostics) == 0 {
		fmt.Println("Client checks passed")
	} else {
		for _, diagnostic := range report.Diagnostics {
			location := diagnostic.File
			if diagnostic.Line > 0 {
				location += fmt.Sprintf(":%d:%d", diagnostic.Line, diagnostic.Column)
			}
			if location != "" {
				location += ": "
			}
			fmt.Printf("%s: %s%s %s\n", diagnostic.Package, location, diagnostic.Code, diagnostic.Message)
			if diagnostic.Remediation != "" {
				fmt.Printf("  Fix: %s\n", diagnostic.Remediation)
			}
		}
		fmt.Printf("Client checks failed with %d diagnostic(s)\n", len(report.Diagnostics))
	}
	if failed {
		c.Base().Cleanup()
		os.Exit(1)
	}
}

func frameworkStaticDirectory() (string, error) {
	command := exec.Command("go", "list", "-m", "-f={{.Dir}}", "github.com/jkomoros/boardgame")
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("BGCLIENT0001: couldn't locate the boardgame framework with go list: %w: %s", err, strings.TrimSpace(string(output)))
	}
	root := strings.TrimSpace(string(output))
	if root == "" {
		return "", fmt.Errorf("BGCLIENT0001: go list returned an empty boardgame framework directory")
	}
	return filepath.Join(root, "server", "static"), nil
}
