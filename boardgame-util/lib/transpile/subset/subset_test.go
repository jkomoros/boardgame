package subset

import (
	"go/ast"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile/refusal"
	"golang.org/x/tools/go/packages"
)

func loadFixture(t *testing.T) (*packages.Package, *Checker) {
	t.Helper()
	config := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo,
	}
	loaded, err := packages.Load(config, "./testdata/fixture")
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || len(loaded[0].Errors) > 0 {
		t.Fatalf("fixture did not load: %v", loaded[0].Errors)
	}
	pkg := loaded[0]
	dir, _ := filepath.Abs("./testdata/fixture")
	return pkg, New(Config{Fset: pkg.Fset, Pkg: pkg.Types, Info: pkg.TypesInfo, Files: pkg.Syntax, Dir: dir})
}

var wantLine = regexp.MustCompile(`want: (.*)`)

// TestFixture checks every fixture function against its "want:" comment,
// and that the fixtures cover every refusal reason the checker itself
// decides (callee-refused is decided across functions, in lib/gamedef).
func TestFixture(t *testing.T) {
	pkg, checker := loadFixture(t)
	covered := map[string]bool{}
	functions := 0
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			functions++
			name := fd.Name.Name
			if fd.Recv != nil {
				name = types.ExprString(fd.Recv.List[0].Type) + "." + name
			}
			match := wantLine.FindStringSubmatch(fd.Doc.Text())
			if match == nil {
				t.Errorf("%s has no want comment", name)
				continue
			}
			want := []string{}
			if strings.TrimSpace(match[1]) != "none" {
				for _, reason := range strings.Split(match[1], ",") {
					reason = strings.TrimSpace(reason)
					if !refusal.ValidReason(reason) {
						t.Errorf("%s wants unknown reason %q", name, reason)
					}
					want = append(want, reason)
				}
			}
			sort.Strings(want)
			refusals := checker.Check("fixture."+name, fd)
			gotSet := map[string]bool{}
			for _, r := range refusals {
				gotSet[r.Reason] = true
				covered[r.Reason] = true
				if r.Func != "fixture."+name || r.File != "fixture.go" || r.Line == 0 || r.Column == 0 {
					t.Errorf("%s: refusal has a bad identity or position: %+v", name, r)
				}
			}
			got := []string{}
			for reason := range gotSet {
				got = append(got, reason)
			}
			sort.Strings(got)
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Errorf("%s: got reasons %v, want %v; refusals %+v", name, got, want, refusals)
			}
		}
	}
	if functions < 40 {
		t.Fatalf("only %d fixture functions", functions)
	}

	// The two kinds of body with no v1 rule.
	for _, kind := range []string{BodyPredicateConstructor, BodyCustomCounter} {
		var anyDecl ast.Node
		for _, decl := range pkg.Syntax[0].Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok {
				anyDecl = fd
				break
			}
		}
		refusals := checker.CheckKind("fixture.x", anyDecl, kind, "detail")
		if len(refusals) != 1 || refusals[0].Reason != kind || refusals[0].Detail != "detail" {
			t.Errorf("CheckKind(%s) = %+v", kind, refusals)
		}
		covered[kind] = true
	}

	for _, reason := range refusal.Reasons() {
		if reason == refusal.CalleeRefused {
			continue
		}
		if !covered[reason] {
			t.Errorf("no fixture covers refusal reason %s", reason)
		}
	}
}

func TestCheckIsDeterministic(t *testing.T) {
	pkg, checker := loadFixture(t)
	for _, decl := range pkg.Syntax[0].Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		first := checker.Check("f", fd)
		second := checker.Check("f", fd)
		if len(first) != len(second) {
			t.Fatalf("%s: two runs differ", fd.Name.Name)
		}
		for i := range first {
			if first[i] != second[i] {
				t.Fatalf("%s: two runs differ", fd.Name.Name)
			}
		}
		if first == nil {
			t.Fatalf("%s: Check returned nil, want an empty slice", fd.Name.Name)
		}
	}
}

func TestNondeterminism(t *testing.T) {
	pkg, checker := loadFixture(t)
	kinds := map[string]bool{}
	for _, file := range pkg.Syntax {
		for _, finding := range checker.Nondeterminism(file) {
			kinds[finding.Kind] = true
		}
	}
	for _, kind := range []string{"map-range", "global-rand", "value-from-string"} {
		if !kinds[kind] {
			t.Errorf("Nondeterminism found no %s", kind)
		}
	}
}

func TestTableIsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, entry := range Table {
		if entry.Facade == "" || entry.Kind == "" || len(entry.Go) == 0 {
			t.Errorf("incomplete row %+v", entry)
		}
		if entry.Source != proposal && !strings.HasPrefix(entry.Source, proposal) && !strings.HasPrefix(entry.Source, "added: ") {
			t.Errorf("row %v has source %q, want proposal or added: why", entry.Go, entry.Source)
		}
		for _, name := range entry.Go {
			if seen[name] {
				t.Errorf("%s is in two rows", name)
			}
			seen[name] = true
		}
	}
	// Exactly the proposal's "none (gap)" rows are gaps among the facade calls.
	var gaps []string
	for _, entry := range Table {
		if entry.Gap() {
			gaps = append(gaps, entry.Facade)
		}
	}
	sort.Strings(gaps)
	want := "componentAtKey,neighborsByKey,rangeToValue,rangeValue,swapComponentsByKey,valueToRange"
	if strings.Join(gaps, ",") != want {
		t.Errorf("gap rows = %v, want %s", gaps, want)
	}
}

var tsMention = regexp.MustCompile(`\b(moves|behaviors)\.([A-Z][A-Za-z]+)`)

// TestEngineGapsMatchStockTS keeps EngineGaps' stock and behavior entries
// equal to what the engine's rules/*.ts does not mention: every stock move
// type of package moves and every behavior struct is either mentioned there
// (ported) or listed as a gap, never both.
func TestEngineGapsMatchStockTS(t *testing.T) {
	rulesDir := filepath.Join("..", "..", "..", "..", "server", "static", "src", "engine", "rules")
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		t.Fatal(err)
	}
	mentioned := map[string]bool{
		// A fix-up moves.Default: MoveSpec.isFixUp.
		"moves.FixUp": true,
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(rulesDir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range tsMention.FindAllStringSubmatch(string(src), -1) {
			mentioned[match[1]+"."+match[2]] = true
		}
	}

	config := &packages.Config{Mode: packages.NeedName | packages.NeedTypes | packages.NeedImports}
	loaded, err := packages.Load(config, movesPath, behaviorsPath)
	if err != nil {
		t.Fatal(err)
	}
	var moveInterface *types.Interface
	for _, pkg := range loaded {
		for _, imported := range pkg.Types.Imports() {
			if imported.Path() == modulePath {
				moveInterface = imported.Scope().Lookup("Move").Type().Underlying().(*types.Interface)
			}
		}
	}
	if moveInterface == nil {
		t.Fatal("boardgame.Move not found")
	}
	count := map[string]int{}
	counted := map[string]bool{}
	for _, pkg := range loaded {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj, ok := scope.Lookup(name).(*types.TypeName)
			if !ok || !obj.Exported() {
				continue
			}
			named, ok := obj.Type().(*types.Named)
			if !ok {
				continue
			}
			if _, isStruct := named.Underlying().(*types.Struct); !isStruct {
				continue
			}
			key := pkg.Types.Name() + "." + name
			// A stock move type is a move, or an abstract base (DefaultComponent,
			// ApplyUntil) that lacks only Apply: it has Legal.
			if pkg.PkgPath == movesPath && !types.Implements(types.NewPointer(named), moveInterface) {
				if legal, _, _ := types.LookupFieldOrMethod(types.NewPointer(named), true, pkg.Types, "Legal"); legal == nil {
					continue
				}
			}
			count[pkg.Types.Name()]++
			counted[key] = true
			ported, gap := mentioned[key], IsGap(key)
			if ported == gap {
				t.Errorf("%s: mentioned in rules/*.ts = %v, in EngineGaps = %v; want exactly one", key, ported, gap)
			}
		}
	}
	for key := range EngineGaps {
		if !strings.HasPrefix(key, "moves.") && !strings.HasPrefix(key, "behaviors.") {
			continue
		}
		if mentioned[key] {
			t.Errorf("%s is a gap but rules/*.ts ports it", key)
		}
		if !counted[key] {
			t.Errorf("%s is a gap but is not a stock move type or behavior", key)
		}
	}
	if count["behaviors"] != 18 {
		t.Errorf("found %d behaviors, want 18 (the proposal's count)", count["behaviors"])
	}
	t.Logf("stock move types %d, behaviors %d", count["moves"], count["behaviors"])
}
