package gamedef

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile/refusal"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile/subset"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

// StaticInfo is the static half of the exporter: what the game package's
// source says that the live manager cannot. Static produces it from
// go/packages and go/types; Export joins it with the run-time half.
//
// It is the interface between workstream 2 (static.go) and workstream 1
// (export.go). It is not itself serialized into GameData; Export copies its
// parts into the places GameData names.
type StaticInfo struct {
	// Package is the import path that was loaded.
	Package string
	// PackageName is the package's name, the prefix of every FnRef.
	PackageName string
	// Moves is every move type the package declares or configures, keyed
	// by MoveStatic.GoType ("blackjack.moveCurrentPlayerHit"). Export finds
	// each configured move's entry from the reflected type of its move.
	Moves map[string]MoveStatic
	// Traits is what the state structs embed and which interface methods
	// the game declares.
	Traits Traits
	// Computed is computed property name to the function that computes it.
	Computed map[string]FnRef
	// Predicates is custom predicate name to its constructor.
	Predicates map[string]FnRef
	// Counters is a custom progression counter's closure, by the position
	// Export reports for it ("<move name>/<path>"), for a counter that fits
	// no stock family. Empty in v1, where such a counter is refused
	// (custom-counter).
	Counters map[string]FnRef
	// SetUp and End carry the FnRefs of the hooks the game declares; the
	// non-FnRef fields of End (LowScoreWins) come from the run-time half.
	SetUp SetUpData
	End   EndData
	// CurrentPlayerIndex is the delegate's override, if declared.
	CurrentPlayerIndex FnRef
	// Functions is every function reaching engine output, sorted by Func.
	Functions []FunctionEntry
}

// MoveStatic is one move type, as its source declares it.
type MoveStatic struct {
	// GoType is "pkg.typeName".
	GoType string
	// StockType is the first embedded type declared in package moves, by
	// its name there ("CurrentPlayer", "StartPhase", ...), found by walking
	// the embedding chain with types.LookupFieldOrMethod. "" when the only
	// ancestor is moves.Default (then MoveData.Stock is null).
	StockType string
	// Bodies are the methods whose declaring package is the game's: Apply,
	// Legal, LegalCustom, DefaultsForState. A promoted method is absent.
	Bodies MoveBodies
}

// Roles this file adds to the frozen list in schema.go: functions the
// framework calls through an interface that no frozen role names.
const (
	// RoleHook is a method the stock moves or the core call through a
	// framework interface (moves/interfaces: ConditionMet,
	// RoundRobinAction, SourceStack, BeforeLeavePhase, ...).
	RoleHook = "hook"
	// RoleReadyToStart is the delegate's ReadyToStart override.
	RoleReadyToStart = "readyToStart"
	// RoleProposeFixUpMove is the delegate's ProposeFixUpMove override.
	RoleProposeFixUpMove = "proposeFixUpMove"
)

// Buckets of a counted function (StaticReport.Buckets), the baseline
// table's last columns.
const (
	// BucketSubset is in the subset with no refusal.
	BucketSubset = "subset"
	// BucketEngineGap is translatable but blocked only on engine features:
	// every refusal is engine-gap, or callee-refused naming a function that
	// is itself in this bucket. Counted in the subset.
	BucketEngineGap = "engineGap"
	// BucketOutside has a refusal of any other reason.
	BucketOutside = "outside"
)

// StaticReport is Static's result with everything the declarativeness lint
// also reports. StaticInfo is what Export consumes; the rest is the lint's.
type StaticReport struct {
	*StaticInfo
	// Dir is the package directory (absolute); never serialized.
	Dir string `json:"-"`
	// MoveConfigs are the entries ConfigureMoves installs, in source order,
	// with moves.DefaultRoundSetup expanded to the configs it installs.
	MoveConfigs []MoveConfigStatic
	// Buckets is FnRef to bucket for every entry of Functions.
	Buckets map[FnRef]string
	// Order is Functions' FnRefs callee first: every in-package callee
	// before its callers, ties and cycles broken by FnRef.
	Order []FnRef
	// NotCounted are game functions the exporter evaluates on the live
	// manager instead (Name, ConfigureMoves, the constructors, HelpText, and
	// the helpers only they call).
	NotCounted []FunctionRef
	// NotPorted are Diagram and agent code, and the helpers only they call.
	NotPorted []FunctionRef
	// Erased are concreteStates and functions like it: their results are
	// the game handle and the player indexes.
	Erased []FunctionRef
	// Unreached are game functions nothing above reaches.
	Unreached []FunctionRef
	// Findings are the constructs with no deterministic answer anywhere in
	// the package, by function.
	Findings []Finding
}

// FunctionRef is a game function outside Functions, with why.
type FunctionRef struct {
	Func FnRef
	// Why is the slot or the root that reaches it ("Diagram", "agent",
	// "ConfigureDecks", ...).
	Why  string
	File string
	Line int
}

// Finding is a nondeterministic construct (subset.Finding) in one function.
type Finding struct {
	Func   FnRef
	Kind   string
	Detail string
	File   string
	Line   int
	Column int
}

// MoveConfigStatic is one entry ConfigureMoves installs, as its source
// says.
type MoveConfigStatic struct {
	// GoType is the configured move's type, "pkg.typeName" ("moves.SeatPlayer"
	// for a stock type configured directly).
	GoType string
	// StockType is MoveStatic.StockType.
	StockType string
	// StockPorted is false when StockType is a stock move the engine lacks.
	StockPorted bool
	// MoveName and MoveNameSuffix are the literal WithMoveName and
	// WithMoveNameSuffix options, when given.
	MoveName       string `json:",omitempty"`
	MoveNameSuffix string `json:",omitempty"`
	// ExpandedFrom names the stock installer this config came from
	// ("moves.DefaultRoundSetup"), or "".
	ExpandedFrom string `json:",omitempty"`
	// PreconditionSpecs is the number of top-level legal.Spec values passed
	// to WithLegalPreconditions; -1 when a call spreads a slice.
	PreconditionSpecs int
	// DataOnly is true when the type is a stock type, or a game type that
	// declares no method (none of its methods' declaring package is the game).
	DataOnly bool
	// Methods are the methods the game declares on the type (directly or on
	// an embedded game type), sorted.
	Methods []string
	// Bodies are MoveStatic.Bodies.
	Bodies MoveBodies
	// File and Line are the config call's position.
	File string
	Line int
}

// Static loads the game package at importPath and returns its static half.
func Static(importPath string) (*StaticInfo, error) {
	report, err := LoadStatic(importPath, "")
	if err != nil {
		return nil, err
	}
	return report.StaticInfo, nil
}

const (
	modulePath    = "github.com/jkomoros/boardgame"
	movesPath     = modulePath + "/moves"
	behaviorsPath = modulePath + "/behaviors"
	basePath      = modulePath + "/base"
)

// LoadStatic loads the one game package pattern names (an import path or a
// directory pattern such as "."), resolved from dir ("" for the current
// directory), and returns its full static report.
func LoadStatic(pattern, dir string) (*StaticReport, error) {
	config := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:   dir,
		Tests: false,
	}
	loaded, err := packages.Load(config, pattern)
	if err != nil {
		return nil, fmt.Errorf("gamedef: load %s: %w", pattern, err)
	}
	if len(loaded) != 1 {
		return nil, fmt.Errorf("gamedef: %s matched %d packages, want 1", pattern, len(loaded))
	}
	pkg := loaded[0]
	if len(pkg.Errors) > 0 {
		var messages []string
		for _, problem := range pkg.Errors {
			messages = append(messages, problem.Error())
		}
		return nil, fmt.Errorf("gamedef: %s does not type-check: %s", pattern, strings.Join(messages, "; "))
	}
	if pkg.Types == nil || pkg.TypesInfo == nil {
		return nil, fmt.Errorf("gamedef: %s has no type information", pattern)
	}
	a := newAnalysis(pkg)
	return a.run()
}

type analysis struct {
	pkg   *packages.Package
	tpkg  *types.Package
	info  *types.Info
	fset  *token.FileSet
	dir   string
	name  string
	files []*ast.File

	sources map[string][]byte

	delegate    *types.Named
	gameState   *types.Named
	playerState *types.Named

	units map[FnRef]*unit
	byObj map[*types.Func]*unit
	// order is every unit's FnRef in source order.
	order []FnRef

	checker   *subset.Checker
	erased    map[*types.Func]bool
	generated map[*ast.File]bool
}

type unit struct {
	ref  FnRef
	fn   *types.Func // nil for a function literal
	node ast.Node    // *ast.FuncDecl or *ast.FuncLit
	// parent is the enclosing declaration of a literal unit.
	parent *unit
	calls  map[FnRef]token.Pos

	role   string
	kind   string // subset.Body*
	detail string
	why    string
}

func newAnalysis(pkg *packages.Package) *analysis {
	dir := pkg.Dir
	if dir == "" && len(pkg.GoFiles) > 0 {
		dir = filepath.Dir(pkg.GoFiles[0])
	}
	a := &analysis{
		pkg:       pkg,
		tpkg:      pkg.Types,
		info:      pkg.TypesInfo,
		fset:      pkg.Fset,
		dir:       dir,
		name:      pkg.Types.Name(),
		files:     pkg.Syntax,
		sources:   make(map[string][]byte),
		units:     make(map[FnRef]*unit),
		byObj:     make(map[*types.Func]*unit),
		erased:    make(map[*types.Func]bool),
		generated: make(map[*ast.File]bool),
	}
	for _, file := range pkg.Syntax {
		if isGeneratedFile(file) {
			a.generated[file] = true
		}
	}
	return a
}

func (a *analysis) run() (*StaticReport, error) {
	a.findStateTypes()
	if a.delegate == nil {
		return nil, fmt.Errorf("%w: %s", ErrNotGame, a.pkg.PkgPath)
	}
	a.collectUnits()
	a.findErased()
	a.checker = subset.New(subset.Config{
		Fset: a.fset, Pkg: a.tpkg, Info: a.info, Files: a.files, Dir: a.dir,
		Erased: func(fn *types.Func) bool { return a.erased[fn] },
	})
	for _, ref := range a.order {
		a.collectCalls(a.units[ref])
	}

	info := &StaticInfo{
		Package:     a.pkg.PkgPath,
		PackageName: a.name,
		Moves:       make(map[string]MoveStatic),
		Computed:    make(map[string]FnRef),
		Predicates:  make(map[string]FnRef),
		Counters:    make(map[string]FnRef),
		Functions:   []FunctionEntry{},
	}
	report := &StaticReport{StaticInfo: info, Dir: a.dir, Buckets: make(map[FnRef]string)}

	roots := a.roots(info)
	configs := a.moveConfigs()
	report.MoveConfigs = configs
	a.moveStatics(info, configs)
	a.traits(info)

	counted, notCounted, notPorted := a.classify(roots)
	a.functions(report, counted)

	for _, ref := range a.order {
		u := a.units[ref]
		if u.fn == nil {
			continue
		}
		fr := a.functionRef(u)
		switch {
		case a.erased[u.fn]:
			fr.Why = "erased"
			report.Erased = append(report.Erased, fr)
		case counted[ref] != "":
		case notPorted[ref] != "":
			fr.Why = notPorted[ref]
			report.NotPorted = append(report.NotPorted, fr)
		case notCounted[ref] != "":
			fr.Why = notCounted[ref]
			report.NotCounted = append(report.NotCounted, fr)
		default:
			fr.Why = "unreached"
			report.Unreached = append(report.Unreached, fr)
		}
	}
	for _, list := range []*[]FunctionRef{&report.NotCounted, &report.NotPorted, &report.Erased, &report.Unreached} {
		if *list == nil {
			*list = []FunctionRef{}
		}
		sort.Slice(*list, func(i, j int) bool { return (*list)[i].Func < (*list)[j].Func })
	}

	report.Findings = []Finding{}
	for _, ref := range a.order {
		u := a.units[ref]
		decl, ok := u.node.(*ast.FuncDecl)
		if !ok {
			continue
		}
		for _, finding := range a.checker.Nondeterminism(decl) {
			report.Findings = append(report.Findings, Finding{
				Func: ref, Kind: finding.Kind, Detail: finding.Detail,
				File: finding.File, Line: finding.Line, Column: finding.Column,
			})
		}
	}
	return report, nil
}

func (a *analysis) source(filename string) []byte {
	if src, ok := a.sources[filename]; ok {
		return src
	}
	src, _ := os.ReadFile(filename)
	a.sources[filename] = src
	return src
}

func (a *analysis) relFile(pos token.Pos) (string, int) {
	p := a.fset.Position(pos)
	file := p.Filename
	if rel, err := filepath.Rel(a.dir, file); err == nil && !strings.HasPrefix(rel, "..") {
		file = filepath.ToSlash(rel)
	} else {
		file = filepath.Base(file)
	}
	return file, p.Line
}

func (a *analysis) functionRef(u *unit) FunctionRef {
	file, line := a.relFile(u.node.Pos())
	return FunctionRef{Func: u.ref, File: file, Line: line}
}

// fnRef is the FnRef of a declared function.
func (a *analysis) fnRef(fn *types.Func) FnRef {
	sig := fn.Type().(*types.Signature)
	pkgName := a.name
	if fn.Pkg() != nil {
		pkgName = fn.Pkg().Name()
	}
	if recv := sig.Recv(); recv != nil {
		t := recv.Type()
		star := ""
		if pointer, ok := t.(*types.Pointer); ok {
			t = pointer.Elem()
			star = "*"
		}
		name := types.TypeString(t, func(*types.Package) string { return "" })
		if named, ok := types.Unalias(t).(*types.Named); ok {
			name = named.Obj().Name()
		}
		return pkgName + ".(" + star + name + ")." + fn.Name()
	}
	return pkgName + "." + fn.Name()
}

// collectUnits makes a unit of every function declaration, and of every
// function literal in a known slot: a computed property's closure, a
// progression counter.
func (a *analysis) collectUnits() {
	for _, file := range a.files {
		if a.generated[file] {
			continue
		}
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			fn, ok := a.info.Defs[fd.Name].(*types.Func)
			if !ok {
				continue
			}
			u := &unit{ref: a.fnRef(fn), fn: fn, node: fd}
			a.add(u)
			a.collectLiteralUnits(u)
		}
	}
}

func (a *analysis) add(u *unit) {
	if _, exists := a.units[u.ref]; exists {
		return
	}
	a.units[u.ref] = u
	if u.fn != nil {
		a.byObj[u.fn] = u
	}
	a.order = append(a.order, u.ref)
}

// collectLiteralUnits finds, inside one declaration, the function literals
// passed as a computed property's function or as a progression counter.
// They are numbered funcN in source order among the declaration's
// top-level literals, as the Go toolchain names closures.
func (a *analysis) collectLiteralUnits(parent *unit) {
	decl := parent.node.(*ast.FuncDecl)
	if decl.Body == nil {
		return
	}
	numbers := make(map[*ast.FuncLit]int)
	count := 0
	var number func(n ast.Node) bool
	number = func(n ast.Node) bool {
		if lit, ok := n.(*ast.FuncLit); ok {
			count++
			numbers[lit] = count
			return false // nested literals are funcN.M; not slots we know
		}
		return true
	}
	ast.Inspect(decl.Body, number)
	ast.Inspect(decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		for i, arg := range call.Args {
			lit, ok := ast.Unparen(arg).(*ast.FuncLit)
			if !ok || numbers[lit] == 0 {
				continue
			}
			role, kind := a.literalSlot(call, i)
			if role == "" {
				continue
			}
			u := &unit{
				ref:    parent.ref + ".func" + strconv.Itoa(numbers[lit]),
				node:   lit,
				parent: parent,
				role:   role,
				kind:   kind,
			}
			a.add(u)
		}
		return true
	})
}

// literalSlot is the role of a function literal passed as argument i of
// call: computed for a boardgame computed-property constructor's function,
// counter for a moves.ValidCounter parameter.
func (a *analysis) literalSlot(call *ast.CallExpr, i int) (string, string) {
	fn, _ := typeutil.Callee(a.info, call).(*types.Func)
	if fn == nil || fn.Pkg() == nil {
		return "", ""
	}
	if fn.Pkg().Path() == modulePath && isComputedConstructor(fn.Name()) {
		return RoleComputed, subset.BodyFunction
	}
	sig := fn.Type().(*types.Signature)
	if i < sig.Params().Len() {
		if named, ok := types.Unalias(sig.Params().At(i).Type()).(*types.Named); ok {
			if named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == movesPath && named.Obj().Name() == "ValidCounter" {
				return RoleCounter, subset.BodyCustomCounter
			}
		}
	}
	return "", ""
}

func isComputedConstructor(name string) bool {
	return (strings.HasPrefix(name, "GlobalComputed") || strings.HasPrefix(name, "PlayerComputed"))
}

// findStateTypes finds the delegate type (NewDelegate's result) and the
// game and player state types (what its constructors return).
func (a *analysis) findStateTypes() {
	if fd := a.topFunc("NewDelegate"); fd != nil {
		a.delegate = a.returnedNamed(fd)
	}
	if a.delegate == nil {
		return
	}
	if fd := a.delegateMethodDecl("GameStateConstructor"); fd != nil {
		a.gameState = a.returnedNamed(fd)
	}
	if fd := a.delegateMethodDecl("PlayerStateConstructor"); fd != nil {
		a.playerState = a.returnedNamed(fd)
	}
}

func (a *analysis) topFunc(name string) *ast.FuncDecl {
	for _, file := range a.files {
		for _, decl := range file.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok && fd.Recv == nil && fd.Name.Name == name {
				return fd
			}
		}
	}
	return nil
}

// delegateMethod is the delegate's method name when the game declares it
// (on the delegate type or an embedded game type).
func (a *analysis) delegateMethod(name string) *types.Func {
	if a.delegate == nil {
		return nil
	}
	return a.declaredMethod(a.delegate, name)
}

func (a *analysis) delegateMethodDecl(name string) *ast.FuncDecl {
	fn := a.delegateMethod(name)
	if fn == nil {
		return nil
	}
	for _, file := range a.files {
		for _, decl := range file.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok && a.info.Defs[fd.Name] == fn {
				return fd
			}
		}
	}
	return nil
}

// declaredMethod is named's method name (pointer method set) when its
// declaring package is the game's, else nil.
func (a *analysis) declaredMethod(named *types.Named, name string) *types.Func {
	obj, _, _ := types.LookupFieldOrMethod(types.NewPointer(named), true, a.tpkg, name)
	fn, ok := obj.(*types.Func)
	if !ok || fn.Pkg() != a.tpkg || a.inGenerated(fn.Pos()) {
		return nil
	}
	return fn
}

// isGeneratedFile reports whether file is generated code: the standard
// "Code generated ... DO NOT EDIT." header, or the boardgame codegen header
// (auto_reader.go, auto_enum.go). The baseline counts only hand-written Go.
func isGeneratedFile(file *ast.File) bool {
	if ast.IsGenerated(file) {
		return true
	}
	for _, group := range file.Comments {
		if group.Pos() > file.Package {
			break
		}
		text := group.Text()
		if strings.Contains(text, "'boardgame-util codegen'") && strings.Contains(text, "DO NOT EDIT") {
			return true
		}
	}
	return false
}

func (a *analysis) inGenerated(pos token.Pos) bool {
	for file := range a.generated {
		if file.FileStart <= pos && pos <= file.FileEnd {
			return true
		}
	}
	return false
}

// declaredMethods are all methods of named's pointer method set declared in
// the game's package, by name.
func (a *analysis) declaredMethods(named *types.Named) []*types.Func {
	set := types.NewMethodSet(types.NewPointer(named))
	var result []*types.Func
	for i := 0; i < set.Len(); i++ {
		fn, ok := set.At(i).Obj().(*types.Func)
		if ok && fn.Pkg() == a.tpkg && !a.inGenerated(fn.Pos()) {
			result = append(result, fn)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name() < result[j].Name() })
	return result
}

// returnedNamed is the game-declared named type (through a pointer) of the
// first return statement in fd that has one.
func (a *analysis) returnedNamed(fd *ast.FuncDecl) *types.Named {
	var result *types.Named
	if fd.Body == nil {
		return nil
	}
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		if result != nil {
			return false
		}
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		ret, ok := n.(*ast.ReturnStmt)
		if !ok || len(ret.Results) == 0 {
			return true
		}
		t := a.info.TypeOf(ret.Results[0])
		if pointer, ok := t.(*types.Pointer); ok {
			t = pointer.Elem()
		}
		if named, ok := types.Unalias(t).(*types.Named); ok && named.Obj().Pkg() == a.tpkg {
			result = named
		}
		return true
	})
	return result
}

// findErased marks concreteStates and functions shaped like it: every
// result is the game state, a player state or a slice of player states,
// and the body calls nothing in the package.
func (a *analysis) findErased() {
	if a.gameState == nil && a.playerState == nil {
		return
	}
	isState := func(t types.Type) bool {
		if slice, ok := t.(*types.Slice); ok {
			t = slice.Elem()
		}
		pointer, ok := t.(*types.Pointer)
		if !ok {
			return false
		}
		named, ok := types.Unalias(pointer.Elem()).(*types.Named)
		return ok && (named == a.gameState || named == a.playerState)
	}
	for _, ref := range a.order {
		u := a.units[ref]
		if u.fn == nil {
			continue
		}
		sig := u.fn.Type().(*types.Signature)
		if sig.Recv() != nil || sig.Results().Len() == 0 {
			continue
		}
		all := true
		for i := 0; i < sig.Results().Len(); i++ {
			if !isState(sig.Results().At(i).Type()) {
				all = false
			}
		}
		if !all {
			continue
		}
		callsGame := false
		ast.Inspect(u.node.(*ast.FuncDecl).Body, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if fn, ok := typeutil.Callee(a.info, call).(*types.Func); ok && fn.Pkg() == a.tpkg {
					callsGame = true
				}
			}
			return !callsGame
		})
		if !callsGame {
			a.erased[u.fn] = true
		}
	}
}

// collectCalls records the in-package functions u refers to (calls or
// function values), excluding erased ones and literal units nested in it,
// plus the game's own delegate method for a call through
// state.Manager().Delegate().
func (a *analysis) collectCalls(u *unit) {
	u.calls = make(map[FnRef]token.Pos)
	var body ast.Node
	switch n := u.node.(type) {
	case *ast.FuncDecl:
		if n.Body == nil {
			return
		}
		body = n.Body
	case *ast.FuncLit:
		body = n.Body
	}
	ast.Inspect(body, func(n ast.Node) bool {
		if lit, ok := n.(*ast.FuncLit); ok && lit != u.node && a.isLiteralUnit(lit) {
			return false
		}
		switch n := n.(type) {
		case *ast.Ident:
			fn, ok := a.info.Uses[n].(*types.Func)
			if !ok || fn.Pkg() != a.tpkg || a.erased[fn] {
				return true
			}
			if callee, ok := a.byObj[fn]; ok {
				a.addCall(u, callee.ref, n.Pos())
			}
		case *ast.CallExpr:
			fn, ok := typeutil.Callee(a.info, n).(*types.Func)
			if !ok || fn.Pkg() == nil || fn.Pkg().Path() != modulePath {
				return true
			}
			if entry, known := subset.Lookup(subset.QualifiedName(fn)); known && entry.Kind == subset.KindDelegate {
				if own := a.delegateMethod(fn.Name()); own != nil {
					if callee, ok := a.byObj[own]; ok {
						a.addCall(u, callee.ref, n.Pos())
					}
				}
			}
		}
		return true
	})
}

func (a *analysis) addCall(u *unit, ref FnRef, pos token.Pos) {
	if ref == u.ref {
		return
	}
	if existing, ok := u.calls[ref]; !ok || pos < existing {
		u.calls[ref] = pos
	}
}

func (a *analysis) isLiteralUnit(lit *ast.FuncLit) bool {
	for _, ref := range a.order {
		if a.units[ref].node == lit {
			return true
		}
	}
	return false
}

// Delegate methods by what the exporter does with them.
var (
	delegateRoles = map[string]string{
		"BeginSetUp":                        RoleSetUpBegin,
		"DistributeComponentToStarterStack": RoleDistribute,
		"FinishSetUp":                       RoleSetUpFinish,
		"GameEndConditionMet":               RoleConditionMet,
		"PlayerScore":                       RolePlayerScore,
		"CheckGameFinished":                 RoleCheckGameFinished,
		"CurrentPlayerIndex":                RoleCurrentPlayerIndex,
		"ConfigurePredicateConstructors":    RolePredicate,
		"ReadyToStart":                      RoleReadyToStart,
		"ProposeFixUpMove":                  RoleProposeFixUpMove,
	}
	// delegateNotCounted are evaluated by the exporter on the live manager.
	delegateNotCounted = map[string]bool{
		"Name": true, "DisplayName": true, "Description": true,
		"DefaultNumPlayers": true, "MinNumPlayers": true, "MaxNumPlayers": true, "LegalNumPlayers": true,
		"Variants": true, "ConfigureMoves": true, "ConfigureDecks": true, "ConfigureConstants": true,
		"ConfigureLegalTemplates": true, "ConfigureEnums": true, "ConfigureAgents": true,
		"ConfigureComputedProperties": true, "GameStateConstructor": true, "PlayerStateConstructor": true,
		"DynamicComponentValuesConstructor": true, "LowScoreWins": true,
	}
	moveRoles = map[string]string{
		"Apply":            RoleApply,
		"Legal":            RoleLegal,
		"LegalCustom":      RoleLegalCustom,
		"DefaultsForState": RoleDefaults,
	}
	// hookNames are moves/interfaces methods the stock moves or the core
	// call on a game's move, state or component.
	hookNames = map[string]bool{
		"ConditionMet": true, "RoundRobinAction": true, "TargetCount": true,
		"PlayerStack": true, "GameStack": true, "SourceStack": true, "DestinationStack": true,
		"ShouldAdvance": true, "AfterAdvance": true, "SpaceIsLegal": true,
		"MovesRemaining": true, "ConsumeMovement": true, "IsFreeMove": true, "ApplyFreeMove": true,
		"BeforeLeavePhase": true, "BeforeEnterPhase": true,
	}
	playerRoles = map[string]string{
		"TurnDone":          RoleTurnDone,
		"ResetForTurnStart": RoleResetForTurnStart,
		"ResetForTurnEnd":   RoleResetForTurnEnd,
		"GameScore":         RoleGameScore,
	}
)

type rootSet struct {
	counted    map[FnRef]string // ref -> role
	notCounted map[FnRef]string // ref -> why
	notPorted  map[FnRef]string // ref -> why
}

// roots finds every function that fills a slot, by role, and the roots of
// what is not counted and not ported.
func (a *analysis) roots(info *StaticInfo) rootSet {
	roots := rootSet{counted: map[FnRef]string{}, notCounted: map[FnRef]string{}, notPorted: map[FnRef]string{}}
	setRole := func(fn *types.Func, role string) FnRef {
		if fn == nil || a.erased[fn] {
			return ""
		}
		u, ok := a.byObj[fn]
		if !ok {
			return ""
		}
		if _, done := roots.counted[u.ref]; !done {
			roots.counted[u.ref] = role
		}
		return u.ref
	}

	if a.delegate != nil {
		for _, fn := range a.declaredMethods(a.delegate) {
			name := fn.Name()
			switch {
			case delegateRoles[name] != "":
				ref := setRole(fn, delegateRoles[name])
				switch name {
				case "BeginSetUp":
					info.SetUp.Begin = ref
				case "DistributeComponentToStarterStack":
					info.SetUp.Distribute = ref
				case "FinishSetUp":
					info.SetUp.Finish = ref
				case "GameEndConditionMet":
					info.End.ConditionMet = ref
				case "PlayerScore":
					info.End.PlayerScore = ref
				case "CheckGameFinished":
					info.End.Check = ref
				case "CurrentPlayerIndex":
					info.CurrentPlayerIndex = ref
				case "ConfigurePredicateConstructors":
					names := a.predicateNames(fn)
					for _, predicate := range names {
						info.Predicates[predicate] = ref
					}
					if u := a.byObj[fn]; u != nil {
						u.kind = subset.BodyPredicateConstructor
						u.detail = strings.Join(names, ",")
					}
				}
			case name == "Diagram":
				if u := a.byObj[fn]; u != nil {
					roots.notPorted[u.ref] = "Diagram"
				}
			case delegateNotCounted[name]:
				if u := a.byObj[fn]; u != nil {
					roots.notCounted[u.ref] = name
				}
			}
		}
	}
	if fd := a.topFunc("NewDelegate"); fd != nil {
		if fn, ok := a.info.Defs[fd.Name].(*types.Func); ok {
			roots.notCounted[a.byObj[fn].ref] = "NewDelegate"
		}
	}

	// Literal units: computed closures and counters.
	for _, ref := range a.order {
		u := a.units[ref]
		if u.fn != nil {
			continue
		}
		roots.counted[ref] = u.role
	}
	for name, ref := range a.computedFunctions(roots) {
		info.Computed[name] = ref
	}

	// Moves: every game type whose pointer is a boardgame.Move.
	for _, named := range a.gameNamedTypes() {
		switch {
		case a.isMove(named):
			for _, fn := range a.declaredMethods(named) {
				if role := moveRoles[fn.Name()]; role != "" {
					setRole(fn, role)
				} else if hookNames[fn.Name()] {
					setRole(fn, RoleHook)
				} else if fn.Name() == "HelpText" {
					if u := a.byObj[fn]; u != nil {
						roots.notCounted[u.ref] = "HelpText"
					}
				}
			}
		case named == a.playerState:
			for _, fn := range a.declaredMethods(named) {
				if role := playerRoles[fn.Name()]; role != "" {
					setRole(fn, role)
				} else if hookNames[fn.Name()] {
					setRole(fn, RoleHook)
				}
			}
		case named == a.gameState:
			for _, fn := range a.declaredMethods(named) {
				if fn.Name() == "PlayerOrder" {
					setRole(fn, RolePlayerOrder)
				} else if hookNames[fn.Name()] {
					setRole(fn, RoleHook)
				}
			}
		case a.isComponentValues(named):
			for _, fn := range a.declaredMethods(named) {
				if fn.Name() == "Legal" {
					setRole(fn, RoleComponent)
				} else if hookNames[fn.Name()] {
					setRole(fn, RoleHook)
				}
			}
		case a.isAgent(named):
			for _, fn := range a.declaredMethods(named) {
				if u := a.byObj[fn]; u != nil {
					roots.notPorted[u.ref] = "agent"
				}
			}
		}
	}
	return roots
}

// computedFunctions are the computed properties' names and functions: a
// literal unit, or a named function passed in the slot.
func (a *analysis) computedFunctions(roots rootSet) map[string]FnRef {
	result := make(map[string]FnRef)
	for _, ref := range a.order {
		u := a.units[ref]
		decl, ok := u.node.(*ast.FuncDecl)
		if !ok || decl.Body == nil {
			continue
		}
		ast.Inspect(decl.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			fn, _ := typeutil.Callee(a.info, call).(*types.Func)
			if fn == nil || fn.Pkg() == nil || fn.Pkg().Path() != modulePath || !isComputedConstructor(fn.Name()) || len(call.Args) < 2 {
				return true
			}
			name := stringLiteral(a.info, call.Args[0])
			last := ast.Unparen(call.Args[len(call.Args)-1])
			switch arg := last.(type) {
			case *ast.FuncLit:
				for _, r := range a.order {
					if a.units[r].node == arg {
						result[name] = r
					}
				}
			case *ast.Ident, *ast.SelectorExpr:
				var id *ast.Ident
				if sel, ok := arg.(*ast.SelectorExpr); ok {
					id = sel.Sel
				} else {
					id = arg.(*ast.Ident)
				}
				if target, ok := a.info.Uses[id].(*types.Func); ok {
					if callee, ok := a.byObj[target]; ok {
						result[name] = callee.ref
						if _, done := roots.counted[callee.ref]; !done {
							roots.counted[callee.ref] = RoleComputed
						}
					}
				}
			}
			return true
		})
	}
	return result
}

func stringLiteral(info *types.Info, expr ast.Expr) string {
	if tv, ok := info.Types[expr]; ok && tv.Value != nil {
		if s, err := strconv.Unquote(tv.Value.ExactString()); err == nil {
			return s
		}
	}
	return ""
}

// predicateNames are the Name fields of the legal.PredicateConstructor
// literals ConfigurePredicateConstructors returns.
func (a *analysis) predicateNames(fn *types.Func) []string {
	u := a.byObj[fn]
	if u == nil {
		return nil
	}
	var names []string
	decl := u.node.(*ast.FuncDecl)
	if decl.Body == nil {
		return nil
	}
	ast.Inspect(decl.Body, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		t := a.info.TypeOf(lit)
		if pointer, ok := t.(*types.Pointer); ok {
			t = pointer.Elem()
		}
		named, ok := types.Unalias(t).(*types.Named)
		if !ok || (named.Obj().Name() != "LegalPredicateConstructor" && named.Obj().Name() != "PredicateConstructor") {
			return true
		}
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			if key, ok := kv.Key.(*ast.Ident); ok && key.Name == "Name" {
				if name := stringLiteral(a.info, kv.Value); name != "" {
					names = append(names, name)
				}
			}
		}
		return true
	})
	sort.Strings(names)
	return names
}

// gameNamedTypes are the package's named types, in source order.
func (a *analysis) gameNamedTypes() []*types.Named {
	var result []*types.Named
	for _, file := range a.files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts := spec.(*ast.TypeSpec)
				if obj, ok := a.info.Defs[ts.Name].(*types.TypeName); ok {
					if named, ok := obj.Type().(*types.Named); ok {
						result = append(result, named)
					}
				}
			}
		}
	}
	return result
}

func (a *analysis) frameworkInterface(pkgPath, name string) *types.Interface {
	for _, imported := range a.tpkg.Imports() {
		if imported.Path() != pkgPath {
			continue
		}
		if obj := imported.Scope().Lookup(name); obj != nil {
			if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
				return iface
			}
		}
	}
	// Not imported directly: search the import graph.
	seen := map[*types.Package]bool{}
	var search func(p *types.Package) *types.Interface
	search = func(p *types.Package) *types.Interface {
		if seen[p] {
			return nil
		}
		seen[p] = true
		if p.Path() == pkgPath {
			if obj := p.Scope().Lookup(name); obj != nil {
				if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
					return iface
				}
			}
			return nil
		}
		for _, imported := range p.Imports() {
			if found := search(imported); found != nil {
				return found
			}
		}
		return nil
	}
	return search(a.tpkg)
}

func (a *analysis) implements(named *types.Named, pkgPath, iface string) bool {
	if _, isInterface := named.Underlying().(*types.Interface); isInterface {
		return false
	}
	target := a.frameworkInterface(pkgPath, iface)
	return target != nil && types.Implements(types.NewPointer(named), target)
}

func (a *analysis) isMove(named *types.Named) bool {
	return a.implements(named, modulePath, "Move")
}

func (a *analysis) isAgent(named *types.Named) bool {
	return a.implements(named, modulePath, "Agent")
}

// isComponentValues reports whether named embeds base.ComponentValues.
func (a *analysis) isComponentValues(named *types.Named) bool {
	if named == a.gameState || named == a.playerState {
		return false
	}
	obj, _, _ := types.LookupFieldOrMethod(named, true, a.tpkg, "ComponentValues")
	field, ok := obj.(*types.Var)
	if !ok || !field.Embedded() {
		return false
	}
	fieldNamed, ok := types.Unalias(field.Type()).(*types.Named)
	return ok && fieldNamed.Obj().Pkg() != nil && fieldNamed.Obj().Pkg().Path() == basePath
}

// classify spreads the roots over the call graph. A function reached from
// a counted root is counted, with the root's role if it is a root itself
// and helper otherwise; then not-ported, then not-counted.
func (a *analysis) classify(roots rootSet) (counted, notCounted, notPorted map[FnRef]string) {
	spread := func(start map[FnRef]string, skip ...map[FnRef]string) map[FnRef]string {
		result := make(map[FnRef]string)
		var queue []FnRef
		refs := make([]FnRef, 0, len(start))
		for ref := range start {
			refs = append(refs, ref)
		}
		sort.Strings(refs)
		for _, ref := range refs {
			result[ref] = start[ref]
			queue = append(queue, ref)
		}
		for len(queue) > 0 {
			ref := queue[0]
			queue = queue[1:]
			u := a.units[ref]
			callees := make([]FnRef, 0, len(u.calls))
			for callee := range u.calls {
				callees = append(callees, callee)
			}
			sort.Strings(callees)
			for _, callee := range callees {
				if _, done := result[callee]; done {
					continue
				}
				skipped := false
				for _, other := range skip {
					if _, in := other[callee]; in {
						skipped = true
					}
				}
				if skipped {
					continue
				}
				result[callee] = ""
				queue = append(queue, callee)
			}
		}
		return result
	}
	counted = spread(roots.counted)
	for ref, role := range counted {
		if role == "" {
			counted[ref] = RoleHelper
		}
	}
	notPorted = spread(roots.notPorted, counted)
	for ref, why := range notPorted {
		if why == "" {
			notPorted[ref] = "helper of Diagram or an agent"
		}
	}
	notCounted = spread(roots.notCounted, counted, notPorted)
	for ref, why := range notCounted {
		if why == "" {
			notCounted[ref] = "helper of a function the exporter evaluates"
		}
	}
	return counted, notCounted, notPorted
}

// functions checks every counted function, propagates callee-refused
// callee first, and fills Functions, Buckets and Order.
func (a *analysis) functions(report *StaticReport, counted map[FnRef]string) {
	refs := make([]FnRef, 0, len(counted))
	for ref := range counted {
		refs = append(refs, ref)
	}
	sort.Strings(refs)

	own := make(map[FnRef][]refusal.Refusal)
	for _, ref := range refs {
		u := a.units[ref]
		own[ref] = a.checker.CheckKind(ref, u.node, u.kind, u.detail)
	}

	order := calleeFirst(refs, func(ref FnRef) []FnRef {
		var callees []FnRef
		for callee := range a.units[ref].calls {
			if _, ok := counted[callee]; ok {
				callees = append(callees, callee)
			}
		}
		sort.Strings(callees)
		return callees
	})
	report.Order = order

	// Propagate to a fixed point (cycles): a function with a refused
	// callee gets callee-refused naming it, at the first reference.
	all := make(map[FnRef][]refusal.Refusal)
	for _, ref := range refs {
		all[ref] = append([]refusal.Refusal(nil), own[ref]...)
	}
	for changed := true; changed; {
		changed = false
		for _, ref := range order {
			u := a.units[ref]
			for _, callee := range sortedKeys(u.calls) {
				if _, ok := counted[callee]; !ok || len(all[callee]) == 0 {
					continue
				}
				if hasCalleeRefusal(all[ref], callee) {
					continue
				}
				file, line := a.relFile(u.calls[callee])
				column := a.fset.Position(u.calls[callee]).Column
				all[ref] = append(all[ref], refusal.Refusal{
					Func: ref, Reason: refusal.CalleeRefused, Detail: callee,
					File: file, Line: line, Column: column,
				})
				changed = true
			}
		}
	}

	// Buckets: outside wins; a callee-refused counts as outside when its
	// callee is outside.
	buckets := make(map[FnRef]string)
	for _, ref := range refs {
		buckets[ref] = BucketSubset
		if len(all[ref]) > 0 {
			buckets[ref] = BucketEngineGap
		}
		for _, r := range all[ref] {
			if r.Reason != refusal.EngineGap && r.Reason != refusal.CalleeRefused {
				buckets[ref] = BucketOutside
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for _, ref := range order {
			if buckets[ref] == BucketOutside {
				continue
			}
			for _, r := range all[ref] {
				if r.Reason == refusal.CalleeRefused && buckets[r.Detail] == BucketOutside {
					buckets[ref] = BucketOutside
					changed = true
				}
			}
		}
	}
	report.Buckets = buckets

	for _, ref := range refs {
		u := a.units[ref]
		refusals := all[ref]
		sortRefusals(refusals)
		if refusals == nil {
			refusals = []refusal.Refusal{}
		}
		verdict := VerdictSubset
		if len(refusals) > 0 {
			verdict = VerdictRefused
		}
		file, line := a.relFile(u.node.Pos())
		calls := []FnRef{}
		for _, callee := range sortedKeys(u.calls) {
			if _, ok := counted[callee]; ok {
				calls = append(calls, callee)
			}
		}
		report.Functions = append(report.Functions, FunctionEntry{
			Func: ref, Role: counted[ref], File: file, Line: line,
			SourceHash: SourceHash(a.nodeSource(u.node)),
			Calls:      calls, Verdict: verdict, Refusals: refusals,
		})
	}
}

func hasCalleeRefusal(refusals []refusal.Refusal, callee FnRef) bool {
	for _, r := range refusals {
		if r.Reason == refusal.CalleeRefused && r.Detail == callee {
			return true
		}
	}
	return false
}

func sortRefusals(refusals []refusal.Refusal) {
	sort.SliceStable(refusals, func(i, j int) bool {
		x, y := refusals[i], refusals[j]
		if x.File != y.File {
			return x.File < y.File
		}
		if x.Line != y.Line {
			return x.Line < y.Line
		}
		if x.Column != y.Column {
			return x.Column < y.Column
		}
		if x.Reason != y.Reason {
			return x.Reason < y.Reason
		}
		return x.Detail < y.Detail
	})
}

func sortedKeys(m map[FnRef]token.Pos) []FnRef {
	keys := make([]FnRef, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// calleeFirst orders refs so that every callee comes before its callers;
// among functions free to go next, and to break a cycle, the smallest FnRef
// goes first.
func calleeFirst(refs []FnRef, callees func(FnRef) []FnRef) []FnRef {
	done := make(map[FnRef]bool)
	visiting := make(map[FnRef]bool)
	var order []FnRef
	var visit func(ref FnRef)
	visit = func(ref FnRef) {
		if done[ref] || visiting[ref] {
			return
		}
		visiting[ref] = true
		for _, callee := range callees(ref) {
			visit(callee)
		}
		visiting[ref] = false
		done[ref] = true
		order = append(order, ref)
	}
	sorted := append([]FnRef(nil), refs...)
	sort.Strings(sorted)
	for _, ref := range sorted {
		visit(ref)
	}
	return order
}

// CalleeFirst orders functions so that every in-package callee comes
// before its callers (the order the translator and the agent loop handle
// them in); ties and cycles are broken by FnRef.
func CalleeFirst(functions []FunctionEntry) []FnRef {
	byRef := make(map[FnRef]FunctionEntry, len(functions))
	refs := make([]FnRef, 0, len(functions))
	for _, entry := range functions {
		byRef[entry.Func] = entry
		refs = append(refs, entry.Func)
	}
	return calleeFirst(refs, func(ref FnRef) []FnRef {
		var callees []FnRef
		for _, callee := range byRef[ref].Calls {
			if _, ok := byRef[callee]; ok {
				callees = append(callees, callee)
			}
		}
		return callees
	})
}

// nodeSource is one function's source: a declaration from its func keyword
// (the doc comment excluded) to its closing brace, or a literal.
func (a *analysis) nodeSource(node ast.Node) []byte {
	start := a.fset.Position(node.Pos())
	end := a.fset.Position(node.End())
	src := a.source(start.Filename)
	if start.Offset < 0 || end.Offset > len(src) || start.Offset > end.Offset {
		return nil
	}
	return src[start.Offset:end.Offset]
}

// moveConfigs finds the auto.MustConfig and auto.Config calls ConfigureMoves
// reaches (through in-package helpers), in source order, with stock
// installers expanded.
func (a *analysis) moveConfigs() []MoveConfigStatic {
	result := []MoveConfigStatic{}
	fn := a.delegateMethod("ConfigureMoves")
	if fn == nil {
		return result
	}
	visited := make(map[*types.Func]bool)
	var walk func(fn *types.Func)
	walk = func(fn *types.Func) {
		if visited[fn] {
			return
		}
		visited[fn] = true
		u := a.byObj[fn]
		if u == nil {
			return
		}
		decl := u.node.(*ast.FuncDecl)
		if decl.Body == nil {
			return
		}
		ast.Inspect(decl.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			callee, _ := typeutil.Callee(a.info, call).(*types.Func)
			if callee == nil {
				return true
			}
			if callee.Pkg() == a.tpkg {
				walk(callee)
				return true
			}
			qualified := subset.QualifiedName(callee)
			switch qualified {
			case "(*moves.AutoConfigurer).MustConfig", "(*moves.AutoConfigurer).Config":
				if config, ok := a.moveConfig(call); ok {
					result = append(result, config)
				}
				return false
			}
			if expansion, ok := stockInstallers[qualified]; ok {
				types := expansion.base
				if len(call.Args) > 1 {
					types = expansion.withOptions
				}
				file, line := a.relFile(call.Pos())
				for _, stock := range types {
					result = append(result, MoveConfigStatic{
						GoType: "moves." + stock, StockType: stock, StockPorted: !subset.IsGap("moves." + stock),
						ExpandedFrom: qualified, DataOnly: true, Methods: []string{},
						File: file, Line: line,
					})
				}
				return false
			}
			return true
		})
	}
	walk(fn)
	return result
}

type installer struct {
	base        []string
	withOptions []string
}

// stockInstallers are the moves functions that configure moves themselves
// (moves/seat_player.go), with the stock types they install. With any
// option (WithManualStart) DefaultRoundSetup also installs CloseAllSeats.
var stockInstallers = map[string]installer{
	"moves.DefaultRoundSetup": {
		base:        []string{"ActivateInactivePlayer", "WaitForEnoughPlayers", "InactivateEmptySeat"},
		withOptions: []string{"ActivateInactivePlayer", "CloseAllSeats", "WaitForEnoughPlayers", "InactivateEmptySeat"},
	},
}

func (a *analysis) moveConfig(call *ast.CallExpr) (MoveConfigStatic, bool) {
	if len(call.Args) == 0 {
		return MoveConfigStatic{}, false
	}
	t := a.info.TypeOf(call.Args[0])
	pointer, ok := t.(*types.Pointer)
	if !ok {
		return MoveConfigStatic{}, false
	}
	named, ok := types.Unalias(pointer.Elem()).(*types.Named)
	if !ok {
		return MoveConfigStatic{}, false
	}
	file, line := a.relFile(call.Pos())
	static := a.moveStatic(named)
	config := MoveConfigStatic{
		GoType: static.GoType, StockType: static.StockType,
		StockPorted: static.StockType == "" || !subset.IsGap("moves."+static.StockType),
		Bodies:      static.Bodies, Methods: []string{}, File: file, Line: line,
	}
	if named.Obj().Pkg() == a.tpkg {
		for _, fn := range a.declaredMethods(named) {
			config.Methods = append(config.Methods, fn.Name())
		}
	}
	config.DataOnly = len(config.Methods) == 0
	for _, arg := range call.Args[1:] {
		option, ok := ast.Unparen(arg).(*ast.CallExpr)
		if !ok {
			continue
		}
		fn, _ := typeutil.Callee(a.info, option).(*types.Func)
		if fn == nil {
			continue
		}
		switch subset.QualifiedName(fn) {
		case "moves.WithMoveName":
			if len(option.Args) == 1 {
				config.MoveName = stringLiteral(a.info, option.Args[0])
			}
		case "moves.WithMoveNameSuffix":
			if len(option.Args) == 1 {
				config.MoveNameSuffix = stringLiteral(a.info, option.Args[0])
			}
		case "moves.WithLegalPreconditions":
			if option.Ellipsis.IsValid() {
				config.PreconditionSpecs = -1
			} else if config.PreconditionSpecs >= 0 {
				config.PreconditionSpecs += len(option.Args)
			}
		}
	}
	return config, true
}

// moveStatic is one move type's GoType, nearest stock ancestor and bodies.
func (a *analysis) moveStatic(named *types.Named) MoveStatic {
	obj := named.Obj()
	static := MoveStatic{GoType: obj.Pkg().Name() + "." + obj.Name()}
	if obj.Pkg().Path() == movesPath {
		if obj.Name() != "Default" {
			static.StockType = obj.Name()
		}
		return static
	}
	static.StockType = a.stockAncestor(named)
	for name, role := range moveRoles {
		fn := a.declaredMethod(named, name)
		if fn == nil {
			continue
		}
		ref := a.fnRef(fn)
		switch role {
		case RoleApply:
			static.Bodies.Apply = ref
		case RoleLegal:
			static.Bodies.Legal = ref
		case RoleLegalCustom:
			static.Bodies.LegalCustom = ref
		case RoleDefaults:
			static.Bodies.Defaults = ref
		}
	}
	return static
}

// stockAncestor walks the embedding chain breadth first and returns the
// name of the nearest embedded type declared in package moves, "" when that
// is moves.Default or there is none.
func (a *analysis) stockAncestor(named *types.Named) string {
	queue := []*types.Named{named}
	seen := map[*types.Named]bool{}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if seen[current] {
			continue
		}
		seen[current] = true
		strct, ok := current.Underlying().(*types.Struct)
		if !ok {
			continue
		}
		var next []*types.Named
		for i := 0; i < strct.NumFields(); i++ {
			field := strct.Field(i)
			if !field.Embedded() {
				continue
			}
			t := types.Unalias(field.Type())
			if pointer, ok := t.(*types.Pointer); ok {
				t = types.Unalias(pointer.Elem())
			}
			embedded, ok := t.(*types.Named)
			if !ok {
				continue
			}
			// Confirm the promotion path with LookupFieldOrMethod, as the
			// proposal specifies.
			if obj, _, _ := types.LookupFieldOrMethod(current, true, a.tpkg, embedded.Obj().Name()); obj == nil {
				continue
			}
			if embedded.Obj().Pkg() != nil && embedded.Obj().Pkg().Path() == movesPath {
				if embedded.Obj().Name() == "Default" {
					return ""
				}
				return embedded.Obj().Name()
			}
			next = append(next, embedded)
		}
		queue = append(queue, next...)
	}
	return ""
}

// moveStatics fills StaticInfo.Moves: every configured type, and every
// move type the game declares.
func (a *analysis) moveStatics(info *StaticInfo, configs []MoveConfigStatic) {
	for _, config := range configs {
		info.Moves[config.GoType] = MoveStatic{GoType: config.GoType, StockType: config.StockType, Bodies: config.Bodies}
	}
	for _, named := range a.gameNamedTypes() {
		if a.isMove(named) {
			static := a.moveStatic(named)
			info.Moves[static.GoType] = static
		}
	}
}

// traits reads the state structs' embedded behaviors and the interface
// methods the game declares.
func (a *analysis) traits(info *StaticInfo) {
	if a.gameState != nil {
		if field := a.embeddedBehavior(a.gameState, "DrawDiscardPair"); field != nil {
			tag := reflect.StructTag(field.tag)
			info.Traits.Game.DrawDiscardPair = &DrawDiscardPair{Draw: tag.Get("draw"), Discard: tag.Get("discard")}
		}
		if a.embeddedBehavior(a.gameState, "RoundRobin") != nil {
			info.Traits.Game.RoundRobin = true
		}
		if fn := a.declaredMethod(a.gameState, "PlayerOrder"); fn != nil {
			info.Traits.Game.PlayerOrder = a.fnRef(fn)
		}
	}
	if a.playerState != nil {
		p := &info.Traits.Player
		p.Seat = a.embeddedBehavior(a.playerState, "Seat") != nil
		p.Inactive = a.embeddedBehavior(a.playerState, "InactivePlayer") != nil
		p.Elimination = a.embeddedBehavior(a.playerState, "PlayerElimination") != nil
		p.Role = a.embeddedBehavior(a.playerState, "PlayerRole") != nil
		p.Color = a.embeddedBehavior(a.playerState, "PlayerColor") != nil
		if fn := a.declaredMethod(a.playerState, "GameScore"); fn != nil {
			p.GameScore = a.fnRef(fn)
		} else if a.embeddedBehavior(a.playerState, "ScoreBehavior") != nil {
			p.GameScore = GameScoreFromScore
		}
		turn := TurnTraits{}
		declared := false
		for name, target := range map[string]*FnRef{"TurnDone": &turn.TurnDone, "ResetForTurnStart": &turn.ResetForTurnStart, "ResetForTurnEnd": &turn.ResetForTurnEnd} {
			if fn := a.declaredMethod(a.playerState, name); fn != nil {
				*target = a.fnRef(fn)
				declared = true
			}
		}
		if declared {
			p.Turn = &turn
		}
	}
}

type behaviorField struct {
	tag string
}

// embeddedBehavior finds an embedded field of type behaviors.<name> at any
// depth of named's struct, with its struct tag.
func (a *analysis) embeddedBehavior(named *types.Named, name string) *behaviorField {
	obj, index, _ := types.LookupFieldOrMethod(named, true, a.tpkg, name)
	field, ok := obj.(*types.Var)
	if !ok || !field.Embedded() {
		return nil
	}
	t := types.Unalias(field.Type())
	if pointer, ok := t.(*types.Pointer); ok {
		t = types.Unalias(pointer.Elem())
	}
	fieldNamed, ok := t.(*types.Named)
	if !ok || fieldNamed.Obj().Pkg() == nil || fieldNamed.Obj().Pkg().Path() != behaviorsPath {
		return nil
	}
	// Walk the index path for the tag of the last hop.
	var current types.Type = named
	tag := ""
	for _, i := range index {
		strct, ok := current.Underlying().(*types.Struct)
		if !ok {
			break
		}
		tag = strct.Tag(i)
		current = strct.Field(i).Type()
		if pointer, ok := current.(*types.Pointer); ok {
			current = pointer.Elem()
		}
	}
	return &behaviorField{tag: tag}
}

// ErrNotGame is returned when the package declares no NewDelegate.
var ErrNotGame = errors.New("gamedef: package declares no NewDelegate")
