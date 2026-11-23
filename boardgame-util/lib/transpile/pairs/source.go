package pairs

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/fuzz"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
)

// Span is one manifest function's place in its Go file.
type Span struct {
	Func string
	// File is relative to the package directory.
	File string
	// Start and End are the function's first and last positions (1-based
	// line and column), as go/token reports them.
	StartLine, StartCol, EndLine, EndCol int
	// Source is the function's text: a declaration from its func keyword
	// (the doc comment excluded) to its closing brace, or a literal, exactly
	// what gamedef hashes.
	Source []byte
}

// Spans locates every manifest function of a game package in its source
// and checks each against the manifest's hash.
func Spans(report *gamedef.StaticReport) (map[string]Span, error) {
	fset := token.NewFileSet()
	files := map[string]*ast.File{}
	sources := map[string][]byte{}
	out := map[string]Span{}
	for _, entry := range report.Functions {
		file, ok := files[entry.File]
		if !ok {
			src, err := os.ReadFile(filepath.Join(report.Dir, entry.File))
			if err != nil {
				return nil, err
			}
			parsed, err := parser.ParseFile(fset, entry.File, src, parser.ParseComments)
			if err != nil {
				return nil, err
			}
			files[entry.File], sources[entry.File] = parsed, src
			file = parsed
		}
		node, err := findUnit(fset, file, entry)
		if err != nil {
			return nil, err
		}
		start, end := fset.Position(node.Pos()), fset.Position(node.End())
		span := Span{
			Func: entry.Func, File: entry.File,
			StartLine: start.Line, StartCol: start.Column, EndLine: end.Line, EndCol: end.Column,
			Source: sources[entry.File][start.Offset:end.Offset],
		}
		if got := gamedef.SourceHash(span.Source); got != entry.SourceHash {
			return nil, fmt.Errorf("%s: the source found at %s:%d hashes to %s, the manifest says %s", entry.Func, entry.File, entry.Line, got, entry.SourceHash)
		}
		out[entry.Func] = span
	}
	return out, nil
}

// findUnit is the declaration (or, for a FnRef ending in .funcN, the
// function literal) the manifest entry names, by its line.
func findUnit(fset *token.FileSet, file *ast.File, entry gamedef.FunctionEntry) (ast.Node, error) {
	literal := 0
	if i := strings.LastIndex(entry.Func, ".func"); i >= 0 {
		if n, err := strconv.Atoi(entry.Func[i+len(".func"):]); err == nil && n > 0 {
			literal = n
		}
	}
	var found ast.Node
	ast.Inspect(file, func(n ast.Node) bool {
		if found != nil || n == nil {
			return false
		}
		switch node := n.(type) {
		case *ast.FuncDecl:
			if literal == 0 && fset.Position(node.Pos()).Line == entry.Line {
				found = node
				return false
			}
		case *ast.FuncLit:
			if literal > 0 && fset.Position(node.Pos()).Line == entry.Line {
				found = node
				return false
			}
		}
		return true
	})
	if found == nil {
		return nil, fmt.Errorf("%s: no function at %s:%d", entry.Func, entry.File, entry.Line)
	}
	return found, nil
}

// Blocks is Go block coverage of one function.
type Blocks struct {
	Covered, Total int
	Unreached      []string
}

// BlockCoverage assigns a text coverage profile's blocks of the game's
// package to each manifest function: a block counts for the innermost
// manifest function whose span holds it, so a function literal that is its
// own manifest entry (ConfigureComputedProperties.func1) is not counted
// again in its enclosing declaration.
func BlockCoverage(blocks []fuzz.ProfileBlock, game string, spans map[string]Span) map[string]Blocks {
	importPath := fuzz.GamePackage(game)
	before := func(l1, c1, l2, c2 int) bool { return l1 < l2 || l1 == l2 && c1 <= c2 }
	inside := func(b fuzz.ProfileBlock, s Span) bool {
		return before(s.StartLine, s.StartCol, b.StartLine, b.StartCol) && before(b.EndLine, b.EndCol, s.EndLine, s.EndCol)
	}
	funcs := make([]string, 0, len(spans))
	for f := range spans {
		funcs = append(funcs, f)
	}
	sort.Strings(funcs)
	out := map[string]Blocks{}
	for _, b := range blocks {
		if filepath.ToSlash(filepath.Dir(b.File)) != importPath {
			continue
		}
		name := filepath.Base(b.File)
		best := ""
		for _, f := range funcs {
			s := spans[f]
			if s.File != name || !inside(b, s) {
				continue
			}
			if best == "" || inside(fuzz.ProfileBlock{StartLine: s.StartLine, StartCol: s.StartCol, EndLine: s.EndLine, EndCol: s.EndCol}, spans[best]) {
				best = f
			}
		}
		if best == "" {
			continue
		}
		c := out[best]
		c.Total++
		if b.Count > 0 {
			c.Covered++
		} else {
			c.Unreached = append(c.Unreached, fmt.Sprintf("%s:%d.%d-%d.%d", name, b.StartLine, b.StartCol, b.EndLine, b.EndCol))
		}
		out[best] = c
	}
	for f, c := range out {
		sort.Strings(c.Unreached)
		out[f] = c
	}
	return out
}

// Seeds are the seeds of a game's checked-in fuzz documents.
func Seeds(root, game string) ([]int, error) {
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(fuzz.CorpusDir), game))
	if err != nil {
		return nil, err
	}
	seeds := []int{}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "seed-") || !strings.HasSuffix(name, ".json.gz") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(name, "seed-"), ".json.gz"))
		if err != nil {
			return nil, fmt.Errorf("fuzz document %s: %w", name, err)
		}
		seeds = append(seeds, n)
	}
	sort.Ints(seeds)
	return seeds, nil
}
