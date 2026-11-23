/*
Package port is the orchestrator and agent harness of the Go-to-TypeScript
pipeline (fabric/TRANSPILE-PROPOSAL.md, "The agent loop" and "The ratchet
and its metrics"): for one game package it translates (lib/transpile),
lists the refusals, writes a brief for each refused function, checks an
agent's translation of one function against the game's recordings
(`port check`), copies an accepted one into the generated game (`port
accept`) or replaces it with the translator's output (`port --adopt`), and
writes the per-game report.

The TypeScript half runs in Node (server/static/scripts/port-check.mjs),
as a child process that is waited for. Nothing here is bounded by time.
*/
package port

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile"
)

// ModulePath is the repository's Go module.
const ModulePath = "github.com/jkomoros/boardgame"

// ExampleGames are the seven example games, in the order reports list them.
var ExampleGames = []string{"memory", "blackjack", "pig", "tictactoe", "werewolf", "debuganimations", "checkers"}

// GamePackage is a game's import path.
func GamePackage(game string) string { return ModulePath + "/examples/" + game }

// Game is one game package loaded for the pipeline.
type Game struct {
	Root       string
	Name       string
	ImportPath string
	Pkg        *transpile.Package
	// Definition is the exported definition (conformance/definitions/<game>.json).
	Definition []byte
	// Existing is the checked-in bodies.gen.ts, nil when the game has no
	// generated game checked in.
	Existing []byte
	// Output is what the translator writes now, keeping Existing's
	// agent-written blocks.
	Output *transpile.Output
	// Commit is the commit recorded for a block written now.
	Commit string

	src  *source
	def  definitionDoc
	byFn map[string]gamedef.FunctionEntry
}

// definitionDoc is the part of the definition the brief and the report read.
type definitionDoc struct {
	Definition struct {
		Game                   subStateDef
		Player                 subStateDef
		DynamicComponentValues map[string]subStateDef
		Decks                  map[string]struct {
			ComponentProps map[string]json.RawMessage
		}
	}
	Moves []struct {
		Name    string
		GoType  string
		IsFixUp bool
		Stock   *struct{ Type string }
		Base    string
		Fields  map[string]json.RawMessage
	}
	Predicates []struct{ Name, Construct string }
	Functions  []gamedef.FunctionEntry
}

type subStateDef struct {
	Props map[string]json.RawMessage
}

// Load loads a game package (an import path) for the pipeline. definition
// is its exported definition; commit is recorded for blocks written now.
func Load(root, importPath string, definition []byte, commit string) (*Game, error) {
	pkg, err := transpile.Load(importPath, root)
	if err != nil {
		return nil, err
	}
	g := &Game{Root: root, Name: pkg.Name(), ImportPath: importPath, Pkg: pkg, Definition: definition, Commit: commit, byFn: map[string]gamedef.FunctionEntry{}}
	if err := json.Unmarshal(definition, &g.def); err != nil {
		return nil, fmt.Errorf("port: %s's definition: %w", g.Name, err)
	}
	for _, entry := range pkg.Static.Functions {
		g.byFn[entry.Func] = entry
	}
	existing, err := os.ReadFile(g.GeneratedPath("bodies.gen.ts"))
	if err == nil {
		g.Existing = existing
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if g.Output, err = pkg.Generate(definition, g.Existing, commit); err != nil {
		return nil, err
	}
	if g.src, err = loadSource(importPath, root); err != nil {
		return nil, err
	}
	return g, nil
}

// GeneratedDir is the game's generated game directory (absolute).
func (g *Game) GeneratedDir() string {
	return filepath.Join(g.Root, filepath.FromSlash(transpile.GamesDir), g.Name)
}

// GeneratedPath is one file of the generated game.
func (g *Game) GeneratedPath(name string) string { return filepath.Join(g.GeneratedDir(), name) }

// HasGenerated reports whether a generated game is checked in.
func (g *Game) HasGenerated() bool { return g.Existing != nil }

// Entry is the manifest entry of ref.
func (g *Game) Entry(ref string) (gamedef.FunctionEntry, bool) {
	e, ok := g.byFn[ref]
	return e, ok
}

// Function is the translator's answer for ref.
func (g *Game) Function(ref string) (transpile.Function, bool) {
	for _, fn := range g.Output.Functions {
		if fn.Func == ref {
			return fn, true
		}
	}
	return transpile.Function{}, false
}

// Untranslated are the manifest functions with no block in Output: refused
// by the translator and not agent-written. Callee first.
func (g *Game) Untranslated() []transpile.Function {
	var out []transpile.Function
	for _, fn := range g.Output.Functions {
		if _, agent := g.Output.Agent[fn.Func]; agent {
			continue
		}
		if len(fn.Refusals) > 0 {
			out = append(out, fn)
		}
	}
	return out
}

// Documents are the recordings a function is checked against: the game's
// checked-in fuzz corpus (conformance/fuzz/<game>/seed-*.json.gz), then its
// conformance documents (conformance/index.json), as absolute paths.
func (g *Game) Documents() ([]string, error) {
	var out []string
	fuzzDir := filepath.Join(g.Root, "conformance", "fuzz", g.Name)
	entries, err := os.ReadDir(fuzzDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	var fuzz []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "seed-") && strings.HasSuffix(e.Name(), ".json.gz") {
			fuzz = append(fuzz, filepath.Join(fuzzDir, e.Name()))
		}
	}
	sort.Strings(fuzz)
	out = append(out, fuzz...)
	blob, err := os.ReadFile(filepath.Join(g.Root, "conformance", "index.json"))
	if err != nil {
		return nil, err
	}
	var index []struct{ File, Game string }
	if err := json.Unmarshal(blob, &index); err != nil {
		return nil, err
	}
	for _, entry := range index {
		if entry.Game == g.Name {
			out = append(out, filepath.Join(g.Root, "conformance", filepath.FromSlash(entry.File)))
		}
	}
	return out, nil
}

// Virtual is the generated game's three files, as the Node helper serves
// them (paths relative to the engine directory), from files.
func (g *Game) Virtual(files map[string][]byte) map[string]string {
	out := map[string]string{}
	for name, text := range files {
		out["games/"+g.Name+"/"+name] = string(text)
	}
	return out
}

// tsName is the TypeScript name of ref.
func tsName(ref string) string {
	name, err := transpile.Mangle(ref)
	if err != nil {
		return ""
	}
	return name
}
