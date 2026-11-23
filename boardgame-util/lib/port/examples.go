package port

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile"
)

// Example is one translation the brief shows: a Go function and its
// accepted TypeScript (fabric/TRANSPILE-PROPOSAL.md, "Stored pairs").
type Example struct {
	Game string
	Func string
	Role string
	// Producer is "transpile@N", "agent:<model>" or "hand".
	Producer string
	// Reasons are the refusal reasons the translator gave the Go source
	// when the pair was stored; empty for a mechanical translation.
	Reasons    []string
	GoSource   string
	TypeScript string
	// Calls are the Go calls outside the game's package the source makes
	// (subset.QualifiedName spelling), for nearness.
	Calls []string
}

// ExampleSource supplies the stored translations a brief chooses its three
// examples from. PairStore reads lib/transpile/testdata/pairs/ (the
// proposal's store); GeneratedExamples pairs the checked-in generated
// games' blocks with their Go source until that store exists.
type ExampleSource interface {
	// Examples are every stored translation, in any order.
	Examples() ([]Example, error)
	// Name says where the examples come from, for the brief.
	Name() string
}

// PairsDir is the proposal's pair store, relative to the module root.
const PairsDir = "boardgame-util/lib/transpile/testdata/pairs"

// DefaultExamples is the pair store when it holds any pair, else the
// checked-in generated games.
func DefaultExamples(root string) ExampleSource {
	store := PairStore{Root: root}
	if examples, err := store.Examples(); err == nil && len(examples) > 0 {
		return store
	}
	return GeneratedExamples{Root: root}
}

// PairStore reads <PairsDir>/<game>/<ts name>/{source.go,accepted.ts,meta.json}.
type PairStore struct{ Root string }

// Name implements ExampleSource.
func (s PairStore) Name() string { return PairsDir }

// pairMeta is meta.json's part the brief reads.
type pairMeta struct {
	Func     string `json:"func"`
	Role     string `json:"role"`
	Producer string `json:"producer"`
	Refusals []struct {
		Reason string `json:"reason"`
		Detail string `json:"detail"`
	} `json:"refusals"`
}

// Examples implements ExampleSource.
func (s PairStore) Examples() ([]Example, error) {
	dir := filepath.Join(s.Root, filepath.FromSlash(PairsDir))
	games, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Example
	for _, game := range games {
		if !game.IsDir() {
			continue
		}
		fns, err := os.ReadDir(filepath.Join(dir, game.Name()))
		if err != nil {
			return nil, err
		}
		for _, fn := range fns {
			base := filepath.Join(dir, game.Name(), fn.Name())
			src, err1 := os.ReadFile(filepath.Join(base, "source.go"))
			ts, err2 := os.ReadFile(filepath.Join(base, "accepted.ts"))
			metaBlob, err3 := os.ReadFile(filepath.Join(base, "meta.json"))
			if err1 != nil || err2 != nil || err3 != nil {
				continue
			}
			var meta pairMeta
			if json.Unmarshal(metaBlob, &meta) != nil {
				continue
			}
			ex := Example{Game: game.Name(), Func: meta.Func, Role: meta.Role, Producer: meta.Producer, GoSource: string(src), TypeScript: string(ts)}
			for _, r := range meta.Refusals {
				ex.Reasons = append(ex.Reasons, r.Reason)
			}
			ex.Calls = goCallNames(string(src))
			out = append(out, ex)
		}
	}
	return out, nil
}

// GeneratedExamples are the blocks of every checked-in generated game
// (games/<game>/bodies.gen.ts) whose header's source hash is the Go
// function's current one, each with that Go function's source (located by
// the checked-in definition's FunctionEntry).
type GeneratedExamples struct{ Root string }

// Name implements ExampleSource.
func (g GeneratedExamples) Name() string {
	return "the checked-in generated games (" + transpile.GamesDir + "/<game>/bodies.gen.ts), until " + PairsDir + " holds pairs"
}

// Examples implements ExampleSource.
func (g GeneratedExamples) Examples() ([]Example, error) {
	gamesDir := filepath.Join(g.Root, filepath.FromSlash(transpile.GamesDir))
	entries, err := os.ReadDir(gamesDir)
	if err != nil {
		return nil, err
	}
	var out []Example
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		game := entry.Name()
		bodies, err := os.ReadFile(filepath.Join(gamesDir, game, "bodies.gen.ts"))
		if err != nil {
			continue
		}
		defBlob, err := os.ReadFile(filepath.Join(g.Root, filepath.FromSlash(transpile.DefinitionsDir), game+".json"))
		if err != nil {
			continue
		}
		var def struct{ Functions []gamedef.FunctionEntry }
		if err := json.Unmarshal(defBlob, &def); err != nil {
			continue
		}
		blocks := transpile.Blocks(bodies)
		for _, entry := range def.Functions {
			block, ok := blocks[entry.Func]
			if !ok || block.SourceHash != entry.SourceHash {
				continue
			}
			src := functionSource(filepath.Join(g.Root, "examples", game, filepath.FromSlash(entry.File)), entry.Line, entry.Func)
			if src == "" || gamedef.SourceHash([]byte(src)) != entry.SourceHash {
				continue
			}
			out = append(out, Example{Game: game, Func: entry.Func, Role: entry.Role, Producer: block.Producer, GoSource: src, TypeScript: block.Code, Calls: goCallNames(src)})
		}
	}
	return out, nil
}

// functionSource is the function declaration (or, for a FnRef ending in
// .funcN, the function literal) starting at line of file, verbatim.
func functionSource(file string, line int, ref string) string {
	text, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, file, text, 0)
	if err != nil {
		return ""
	}
	literal := false
	if i := strings.LastIndex(ref, ".func"); i >= 0 {
		if _, err := strconv.Atoi(ref[i+len(".func"):]); err == nil {
			literal = true
		}
	}
	var found ast.Node
	ast.Inspect(parsed, func(n ast.Node) bool {
		if found != nil || n == nil {
			return false
		}
		switch x := n.(type) {
		case *ast.FuncDecl:
			if !literal && fset.Position(x.Pos()).Line == line {
				found = x
			}
		case *ast.FuncLit:
			if literal && fset.Position(x.Pos()).Line == line {
				found = x
			}
		}
		return true
	})
	if found == nil {
		return ""
	}
	return string(text[fset.Position(found.Pos()).Offset:fset.Position(found.End()).Offset])
}

// goCallNames approximates the calls of a Go source by the selector names
// it calls (`x.Start(`, `.MoveByName(`), for nearness between examples
// whose packages are not loaded. Sorted, without duplicates.
func goCallNames(src string) []string {
	seen := map[string]bool{}
	for i := 0; i < len(src); i++ {
		if src[i] != '(' || i == 0 {
			continue
		}
		j := i
		for j > 0 && isIdentChar(src[j-1]) {
			j--
		}
		if j == i || j == 0 || src[j-1] != '.' {
			continue
		}
		seen[src[j:i]] = true
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func isIdentChar(b byte) bool {
	return b == '_' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

// chooseExamples picks n examples for a target: never from the target's
// own game (a translation must not use knowledge of the rest of its game),
// nearest first by refusal reasons shared with the target, then the same
// role, then Go calls shared, then the shorter Go source, then game and
// FnRef. Deterministic for a given set.
func chooseExamples(all []Example, game, role string, reasons, calls []string, n int) []Example {
	type scored struct {
		ex                    Example
		reasons, role, shared int
	}
	reasonSet := map[string]bool{}
	for _, r := range reasons {
		reasonSet[r] = true
	}
	callSet := map[string]bool{}
	for _, c := range calls {
		callSet[c] = true
	}
	var list []scored
	for _, ex := range all {
		if ex.Game == game {
			continue
		}
		s := scored{ex: ex}
		seen := map[string]bool{}
		for _, r := range ex.Reasons {
			if reasonSet[r] && !seen[r] {
				seen[r] = true
				s.reasons++
			}
		}
		if ex.Role == role {
			s.role = 1
		}
		for _, c := range ex.Calls {
			if callSet[c] {
				s.shared++
			}
		}
		list = append(list, s)
	}
	sort.SliceStable(list, func(i, j int) bool {
		a, b := list[i], list[j]
		if a.reasons != b.reasons {
			return a.reasons > b.reasons
		}
		if a.role != b.role {
			return a.role > b.role
		}
		if a.shared != b.shared {
			return a.shared > b.shared
		}
		if len(a.ex.GoSource) != len(b.ex.GoSource) {
			return len(a.ex.GoSource) < len(b.ex.GoSource)
		}
		if a.ex.Game != b.ex.Game {
			return a.ex.Game < b.ex.Game
		}
		return a.ex.Func < b.ex.Func
	})
	var out []Example
	for i := 0; i < len(list) && i < n; i++ {
		out = append(out, list[i].ex)
	}
	return out
}
