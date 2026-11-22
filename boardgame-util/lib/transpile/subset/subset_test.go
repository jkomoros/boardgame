package subset

import (
	"encoding/json"
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
	// The proposal's "none (gap)" rows. The engine now exports every one
	// (engine/go/api.ts), so none is a gap; a facade gap is checked against
	// the engine in TestEngineGapsMatchStockInventory.
	facades := map[string]bool{}
	for _, entry := range Table {
		facades[entry.Facade] = true
	}
	for _, name := range []string{"componentAtKey", "neighborsByKey", "rangeToValue", "rangeValue", "swapComponentsByKey", "valueToRange"} {
		if !facades[name] {
			t.Errorf("no API table row has facade %s", name)
		}
	}
	for key := range EngineGaps {
		if !strings.Contains(key, ".") && !facades[key] {
			t.Errorf("EngineGaps lists facade %s, which no API table row emits", key)
		}
	}
}

// inventoryEntry is one row of server/static/src/engine/rules/stock-inventory.json.
type inventoryEntry struct {
	GoType       string   `json:"goType"`
	Kind         string   `json:"kind"`
	Ported       bool     `json:"ported"`
	RegistryName *string  `json:"registryName"`
	VerifiedBy   []string `json:"verifiedBy"`
	Notes        string   `json:"notes"`
}

var tsExport = regexp.MustCompile(`(?m)^export (?:function|const) ([A-Za-z_][A-Za-z0-9_]*)`)

// TestEngineGapsMatchStockInventory holds EngineGaps to the engine's own
// machine-readable inventory, rules/stock-inventory.json, the one source of
// truth for what is ported. The inventory must list exactly the stock move
// types of package moves and the behavior structs Go has. A stock move type
// is a gap unless the inventory has it ported and names its registryName (a
// type ported only as parts, which stockMove cannot build from a definition,
// is a gap: see EngineGaps); a behavior is a gap unless it is ported. A
// facade name in EngineGaps must not be exported by engine/go or engine/lib.
func TestEngineGapsMatchStockInventory(t *testing.T) {
	engineDir := filepath.Join("..", "..", "..", "..", "server", "static", "src", "engine")
	raw, err := os.ReadFile(filepath.Join(engineDir, "rules", "stock-inventory.json"))
	if err != nil {
		t.Fatal(err)
	}
	var inventory []inventoryEntry
	if err := json.Unmarshal(raw, &inventory); err != nil {
		t.Fatal(err)
	}
	inventoried := map[string]inventoryEntry{}
	for _, entry := range inventory {
		if entry.Kind != "move" && entry.Kind != "behavior" {
			continue
		}
		if _, dup := inventoried[entry.GoType]; dup {
			t.Errorf("%s is in the inventory twice", entry.GoType)
		}
		if entry.Ported && len(entry.VerifiedBy) == 0 {
			t.Errorf("%s: the inventory says ported but names nothing that verifies it", entry.GoType)
		}
		inventoried[entry.GoType] = entry
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
			entry, ok := inventoried[key]
			if !ok {
				t.Errorf("%s is a stock type Go has but stock-inventory.json does not list", key)
				continue
			}
			wantKind := map[string]string{"moves": "move", "behaviors": "behavior"}[pkg.Types.Name()]
			if entry.Kind != wantKind {
				t.Errorf("%s: inventory kind %q, want %q", key, entry.Kind, wantKind)
			}
			ported := entry.Ported
			if entry.Kind == "move" {
				ported = ported && entry.RegistryName != nil
			}
			if gap := IsGap(key); gap == ported {
				t.Errorf("%s: inventory ported=%v registryName=%v, in EngineGaps=%v; want a gap exactly when it is not buildable", key, entry.Ported, entry.RegistryName != nil, gap)
			}
		}
	}
	for key := range inventoried {
		if !counted[key] {
			t.Errorf("stock-inventory.json lists %s, which is not a Go stock move type or behavior", key)
		}
	}
	for key := range EngineGaps {
		if strings.HasPrefix(key, "moves.") || strings.HasPrefix(key, "behaviors.") {
			if !counted[key] {
				t.Errorf("%s is a gap but is not a stock move type or behavior", key)
			}
		}
	}
	if count["behaviors"] != 18 {
		t.Errorf("found %d behaviors, want 18 (the proposal's count)", count["behaviors"])
	}

	exported := map[string]string{}
	for _, dir := range []string{"go", "lib"} {
		entries, err := os.ReadDir(filepath.Join(engineDir, dir))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range entries {
			if !strings.HasSuffix(file.Name(), ".ts") || strings.HasSuffix(file.Name(), ".test.ts") {
				continue
			}
			src, err := os.ReadFile(filepath.Join(engineDir, dir, file.Name()))
			if err != nil {
				t.Fatal(err)
			}
			for _, match := range tsExport.FindAllStringSubmatch(string(src), -1) {
				exported[match[1]] = dir + "/" + file.Name()
			}
		}
	}
	for key := range EngineGaps {
		if strings.Contains(key, ".") {
			continue
		}
		if file, ok := exported[key]; ok {
			t.Errorf("EngineGaps lists facade %s, but %s exports it", key, file)
		}
	}
	gaps := 0
	for key := range EngineGaps {
		if strings.HasPrefix(key, "moves.") {
			gaps++
		}
	}
	t.Logf("stock move types %d (%d gaps), behaviors %d (%d gaps)", count["moves"], gaps, count["behaviors"], len(EngineGaps)-gaps)
}
