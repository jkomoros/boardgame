package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile"
	"golang.org/x/tools/go/packages"
)

type transpileCmd struct {
	baseSubCommand

	Check  bool
	DryRun bool
	Stats  bool
	Out    string
	Commit string

	AssumeGapsClosed bool
}

func (t *transpileCmd) Name() string {
	return "transpile"
}

func (t *transpileCmd) Description() string {
	return "Translates a game's small Go functions to TypeScript and writes the generated game"
}

func (t *transpileCmd) HelpText() string {
	return t.Name() + ` [--check] [--dry-run] [--stats] [--out DIR] PKG...

For each game package (an import path or a directory pattern; "./examples/..." takes every game under it), translates every function of its manifest (boardgame-util/lib/gamedef, the static half) that the subset checker accepts, and writes the generated game to ` + transpile.GamesDir + `/<game>/ at the module root: definition.gen.ts (the exported definition from ` + transpile.DefinitionsDir + `/<game>.json, as a literal), bodies.gen.ts (the translated functions and the bodies resolver) and index.ts. See fabric/TRANSPILE-PROPOSAL.md, "The translator".

Each generated function's header names the Go function, the SHA-256 of its source and the commit at which its text was first written; an unchanged function keeps its commit, so two runs write identical files.

--check writes nothing and fails if a checked-in generated game differs from what the translator produces now. --dry-run writes nothing and prints each function's verdict. --stats prints the refusals across the packages by reason and detail. --out writes the games under DIR/<game>/ instead.`
}

func (t *transpileCmd) WritOptions() []*writ.Option {
	return []*writ.Option{
		{Names: []string{"check"}, Description: "Verify the checked-in generated games are current without writing files.", Decoder: writ.NewFlagDecoder(&t.Check), Flag: true},
		{Names: []string{"dry-run"}, Description: "Write nothing; print each function's verdict.", Decoder: writ.NewFlagDecoder(&t.DryRun), Flag: true},
		{Names: []string{"stats"}, Description: "Print refusals by reason and detail across the packages.", Decoder: writ.NewFlagDecoder(&t.Stats), Flag: true},
		{Names: []string{"out"}, Description: "Write (or check) generated games under DIR/<game>/ instead of the engine's games directory.", Decoder: writ.NewOptionDecoder(&t.Out), Placeholder: "DIR"},
		{Names: []string{"assume-gaps-closed"}, Description: "Also translate functions refused only for engine gaps, as if the gap table were empty. For measurement with --out or --dry-run; never for checked-in output.", Decoder: writ.NewFlagDecoder(&t.AssumeGapsClosed), Flag: true},
		{Names: []string{"commit"}, Description: "The commit recorded for functions written or changed now. Defaults to the repository's HEAD.", Decoder: writ.NewOptionDecoder(&t.Commit), Placeholder: "SHA"},
	}
}

// transpilePackages expands the patterns to game packages: an explicit
// pattern must be a game; one ending in "..." skips packages that are not.
func transpilePackages(patterns []string) ([]string, error) {
	var result []string
	for _, pattern := range patterns {
		loaded, err := packages.Load(&packages.Config{Mode: packages.NeedName | packages.NeedFiles}, pattern)
		if err != nil {
			return nil, err
		}
		for _, pkg := range loaded {
			if strings.HasSuffix(pattern, "...") && !declaresNewDelegate(pkg) {
				continue
			}
			result = append(result, pkg.PkgPath)
		}
	}
	return result, nil
}

func declaresNewDelegate(pkg *packages.Package) bool {
	for _, file := range pkg.GoFiles {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		src, err := os.ReadFile(file)
		if err == nil && bytes.Contains(src, []byte("func NewDelegate()")) {
			return true
		}
	}
	return false
}

func headCommit(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--short=8", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (t *transpileCmd) Run(p writ.Path, positional []string) {
	if len(positional) == 0 {
		exitHelp(p.Last(), errors.New("PKG is required"))
	}
	paths, err := transpilePackages(positional)
	if err != nil {
		t.Base().errAndQuit(err.Error())
	}
	if len(paths) == 0 {
		t.Base().errAndQuit("no game package matched")
	}
	type stat struct{ reason, detail string }
	stats := map[stat]int{}
	var stale []string
	written := 0
	for _, path := range paths {
		pkg, err := transpile.Load(path, "")
		if err != nil {
			if errors.Is(err, gamedef.ErrNotGame) {
				t.Base().errAndQuit(path + " is not a game package")
			}
			t.Base().errAndQuit(err.Error())
		}
		if t.AssumeGapsClosed {
			if t.Out == "" && !t.DryRun {
				t.Base().errAndQuit("--assume-gaps-closed writes only with --out or --dry-run")
			}
			pkg.AssumeGapsClosed = true
		}
		root, err := transpile.ModuleRoot(pkg.Dir())
		if err != nil {
			t.Base().errAndQuit(err.Error())
		}
		game := pkg.Name()
		definition, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(transpile.DefinitionsDir), game+".json"))
		if err != nil && t.DryRun && os.IsNotExist(err) {
			definition = []byte("null")
		} else if err != nil {
			t.Base().errAndQuit("reading " + game + "'s exported definition (run `boardgame-util gamedef export`): " + err.Error())
		}
		outDir := filepath.Join(root, filepath.FromSlash(transpile.GamesDir), game)
		if t.Out != "" {
			outDir = filepath.Join(t.Out, game)
		}
		existing, err := os.ReadFile(filepath.Join(outDir, "bodies.gen.ts"))
		if err != nil && !os.IsNotExist(err) {
			t.Base().errAndQuit(err.Error())
		}
		commit := t.Commit
		if commit == "" {
			if commit, err = headCommit(root); err != nil {
				t.Base().errAndQuit(err.Error())
			}
		}
		output, err := pkg.Generate(definition, existing, commit)
		if err != nil {
			t.Base().errAndQuit(err.Error())
		}
		translated, refused := 0, 0
		for _, fn := range output.Functions {
			if len(fn.Refusals) == 0 {
				translated++
			} else {
				refused++
			}
			for _, r := range fn.Refusals {
				stats[stat{r.Reason, r.Detail}]++
			}
			if t.DryRun {
				if len(fn.Refusals) == 0 {
					fmt.Printf("%s\ttranslated\t%s\n", fn.Func, fn.Name)
					continue
				}
				who := "checker"
				if fn.EmitterRefused {
					who = "emitter"
				}
				for _, r := range fn.Refusals {
					fmt.Printf("%s\trefused (%s)\t%s\t%s\t%s:%d\n", fn.Func, who, r.Reason, r.Detail, r.File, r.Line)
				}
			}
		}
		if t.DryRun || t.Stats {
			fmt.Printf("%s: %d translated, %d refused\n", game, translated, refused)
		}
		if t.Stats {
			// What the emitter alone makes of each function the checker
			// refuses (its own body only: a callee-refused caller's callees
			// are not looked at). Nothing is written from this.
			for _, fn := range output.Functions {
				if len(fn.Refusals) == 0 || fn.EmitterRefused {
					continue
				}
				ok, refusal := pkg.TryRefused(fn.Func)
				if ok {
					fmt.Printf("emitter alone\t%s\twould translate (checker: %s %s)\n", fn.Func, fn.Refusals[0].Reason, fn.Refusals[0].Detail)
				} else {
					fmt.Printf("emitter alone\t%s\twould refuse: %s %s\n", fn.Func, refusal.Reason, refusal.Detail)
				}
			}
		}
		if t.DryRun {
			continue
		}
		for _, name := range transpile.FileNames {
			target := filepath.Join(outDir, name)
			if t.Check {
				got, err := os.ReadFile(target)
				if err != nil || !bytes.Equal(got, output.Files[name]) {
					stale = append(stale, target)
				}
				continue
			}
			if err := os.MkdirAll(outDir, 0755); err != nil {
				t.Base().errAndQuit(err.Error())
			}
			if err := os.WriteFile(target, output.Files[name], 0644); err != nil {
				t.Base().errAndQuit(err.Error())
			}
			written++
		}
	}
	if t.Stats {
		var keys []stat
		for key := range stats {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			if keys[i].reason != keys[j].reason {
				return keys[i].reason < keys[j].reason
			}
			return keys[i].detail < keys[j].detail
		})
		byReason := map[string]int{}
		for _, key := range keys {
			byReason[key.reason] += stats[key]
		}
		for _, key := range keys {
			fmt.Printf("%s\t%d\t%s\n", key.reason, stats[key], key.detail)
		}
		var reasons []string
		for reason := range byReason {
			reasons = append(reasons, reason)
		}
		sort.Strings(reasons)
		for _, reason := range reasons {
			fmt.Printf("total %s\t%d\n", reason, byReason[reason])
		}
	}
	if t.Check {
		if len(stale) > 0 {
			for _, path := range stale {
				fmt.Fprintln(commandStderr, path+": missing or differs from what the translator produces now")
			}
			t.Base().errAndQuit("The generated games are stale; run `boardgame-util transpile`")
		}
		fmt.Println("The generated games are current")
		return
	}
	if !t.DryRun {
		fmt.Printf("Wrote %d files\n", written)
	}
}
