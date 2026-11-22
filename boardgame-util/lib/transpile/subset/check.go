/*
Package subset is the translator's subset checker: given one type-checked
game function, it returns the reasons the translator would refuse it, or
none. It implements fabric/TRANSPILE-PROPOSAL.md, "The translator": the
Types, Statements and Expressions sections, the package-variable rule, the
API table (Table, unknown-call against engine-gap via EngineGaps) and the
refusals of "Semantics that differ".

The static half of the exporter (lib/gamedef) runs it over every function
that reaches engine output; the translator (lib/transpile) reuses it and
its tables. It imports nothing from either.
*/
package subset

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile/refusal"
	"golang.org/x/tools/go/types/typeutil"
)

const (
	modulePath    = "github.com/jkomoros/boardgame"
	movesPath     = modulePath + "/moves"
	behaviorsPath = modulePath + "/behaviors"
)

// Config is what the checker needs to know about the package.
type Config struct {
	Fset *token.FileSet
	Pkg  *types.Package
	Info *types.Info
	// Files are all of the package's non-test files, for the
	// package-variable rule (a variable written anywhere is not constant).
	Files []*ast.File
	// Dir is the package directory; refusal File fields are relative to it.
	Dir string
	// Erased reports whether a game function is erased like concreteStates:
	// its results have no value of their own. May be nil.
	Erased func(*types.Func) bool
}

// Checker checks functions of one package.
type Checker struct {
	cfg     Config
	varKind map[*types.Var]varKind
}

type varKind int

const (
	varOther varKind = iota
	varConstant
	varFramework
)

// New prepares a checker for one package, classifying its package
// variables.
func New(cfg Config) *Checker {
	c := &Checker{cfg: cfg, varKind: make(map[*types.Var]varKind)}
	c.classifyVars()
	return c
}

// Kinds of body (CheckKind) that have no v1 rule whatever their content.
const (
	// BodyFunction is an ordinary function body.
	BodyFunction = ""
	// BodyPredicateConstructor is a custom legal predicate constructor.
	BodyPredicateConstructor = refusal.PredicateConstructor
	// BodyCustomCounter is a progression counter fitting no stock family.
	BodyCustomCounter = refusal.CustomCounter
)

// CheckKind is Check for a body of the given kind. A predicate constructor
// or custom counter is refused for its kind alone, at its first position,
// with detail; its contents are not examined further.
func (c *Checker) CheckKind(fn string, node ast.Node, kind, detail string) []refusal.Refusal {
	if kind == BodyFunction {
		return c.Check(fn, node)
	}
	return []refusal.Refusal{c.refusalAt(fn, node.Pos(), kind, detail)}
}

// Check returns the refusals for one function: node is its *ast.FuncDecl or
// *ast.FuncLit. fn is its FnRef, copied into each refusal. The result is
// sorted by position, then reason and detail, without duplicates; it is
// empty (not nil) when the function is in the subset.
func (c *Checker) Check(fn string, node ast.Node) []refusal.Refusal {
	w := &walker{c: c, fn: fn, parents: make(map[ast.Node]ast.Node)}
	var funcType *ast.FuncType
	var body *ast.BlockStmt
	switch decl := node.(type) {
	case *ast.FuncDecl:
		funcType, body = decl.Type, decl.Body
		if decl.Recv != nil {
			for _, field := range decl.Recv.List {
				w.checkFieldTypes(field)
			}
		}
	case *ast.FuncLit:
		funcType, body = decl.Type, decl.Body
	default:
		return []refusal.Refusal{}
	}
	if funcType.TypeParams != nil && len(funcType.TypeParams.List) > 0 {
		w.refuse(funcType.TypeParams.Pos(), refusal.Generic, "type parameters")
	}
	for _, list := range []*ast.FieldList{funcType.Params, funcType.Results} {
		if list == nil {
			continue
		}
		for _, field := range list.List {
			w.checkFieldTypes(field)
		}
	}
	if body != nil {
		w.body = body
		ast.Inspect(body, func(n ast.Node) bool {
			if n == nil {
				return true
			}
			for _, child := range children(n) {
				w.parents[child] = n
			}
			return true
		})
		w.parents[body] = node
		ast.Inspect(body, w.visit)
	}
	return w.result()
}

// Finding is one construct with no deterministic answer, reported as a
// warning whatever the function's verdict.
type Finding struct {
	// Kind is "map-range", "global-rand" or "value-from-string".
	Kind   string
	Detail string
	File   string
	Line   int
	Column int
}

// Nondeterminism finds the constructs in node (any syntax: a function, a
// file) with no single answer in Go: range over a map, the package-level
// math/rand functions, and enum ValueFromString.
func (c *Checker) Nondeterminism(node ast.Node) []Finding {
	var findings []Finding
	add := func(pos token.Pos, kind, detail string) {
		file, line, column := c.position(pos)
		findings = append(findings, Finding{Kind: kind, Detail: detail, File: file, Line: line, Column: column})
	}
	ast.Inspect(node, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.RangeStmt:
			if t := c.cfg.Info.TypeOf(n.X); t != nil {
				if _, ok := t.Underlying().(*types.Map); ok {
					add(n.Pos(), "map-range", types.ExprString(n.X))
				}
			}
		case *ast.CallExpr:
			fn, _ := typeutil.Callee(c.cfg.Info, n).(*types.Func)
			if fn == nil || fn.Pkg() == nil {
				return true
			}
			if isRandPackage(fn.Pkg().Path()) && fn.Type().(*types.Signature).Recv() == nil {
				add(n.Pos(), "global-rand", QualifiedName(fn))
			}
			if isValueFromString(fn) {
				add(n.Pos(), "value-from-string", QualifiedName(fn))
			}
		}
		return true
	})
	return findings
}

// QualifiedName is the API table's spelling of a function: "pkg.Func" for a
// function, "(pkg.Type).Method" for a method on a named type or interface,
// "(*pkg.Type).Method" for a pointer receiver. pkg is the package name.
func QualifiedName(fn *types.Func) string {
	sig, _ := fn.Type().(*types.Signature)
	if sig != nil && sig.Recv() != nil {
		t := sig.Recv().Type()
		star := ""
		if pointer, ok := t.(*types.Pointer); ok {
			t = pointer.Elem()
			star = "*"
		}
		return "(" + star + typeName(t) + ")." + fn.Name()
	}
	if fn.Pkg() == nil {
		return fn.Name()
	}
	return fn.Pkg().Name() + "." + fn.Name()
}

func typeName(t types.Type) string {
	t = types.Unalias(t)
	if named, ok := t.(*types.Named); ok {
		obj := named.Obj()
		if obj.Pkg() == nil {
			return obj.Name()
		}
		return obj.Pkg().Name() + "." + obj.Name()
	}
	return types.TypeString(t, func(p *types.Package) string { return p.Name() })
}

func isRandPackage(path string) bool {
	return path == "math/rand" || path == "math/rand/v2"
}

func isValueFromString(fn *types.Func) bool {
	return fn.Name() == "ValueFromString" && fn.Pkg() != nil && fn.Pkg().Path() == modulePath+"/enum"
}

// ValueFromStringDetail is the unknown-call detail for enum ValueFromString,
// which has no reason of its own: on names that normalize alike it has no
// single answer.
const ValueFromStringDetail = "enum.ValueFromString (nondeterministic)"

type walker struct {
	c       *Checker
	fn      string
	body    *ast.BlockStmt
	parents map[ast.Node]ast.Node
	out     []refusal.Refusal
	// writes are identifiers on the left of an assignment, not reads.
	writes map[*ast.Ident]bool
}

func (w *walker) refuse(pos token.Pos, reason, detail string) {
	w.out = append(w.out, w.c.refusalAt(w.fn, pos, reason, detail))
}

func (c *Checker) refusalAt(fn string, pos token.Pos, reason, detail string) refusal.Refusal {
	file, line, column := c.position(pos)
	return refusal.Refusal{Func: fn, Reason: reason, Detail: detail, File: file, Line: line, Column: column}
}

func (c *Checker) position(pos token.Pos) (string, int, int) {
	if c.cfg.Fset == nil || !pos.IsValid() {
		return "", 0, 0
	}
	p := c.cfg.Fset.Position(pos)
	file := p.Filename
	if c.cfg.Dir != "" {
		if rel, err := filepath.Rel(c.cfg.Dir, file); err == nil && !strings.HasPrefix(rel, "..") {
			file = filepath.ToSlash(rel)
		} else {
			file = filepath.Base(file)
		}
	}
	return file, p.Line, p.Column
}

func (w *walker) result() []refusal.Refusal {
	sort.Slice(w.out, func(i, j int) bool {
		a, b := w.out[i], w.out[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		if a.Column != b.Column {
			return a.Column < b.Column
		}
		if a.Reason != b.Reason {
			return a.Reason < b.Reason
		}
		return a.Detail < b.Detail
	})
	result := make([]refusal.Refusal, 0, len(w.out))
	for i, r := range w.out {
		if i > 0 && r == w.out[i-1] {
			continue
		}
		result = append(result, r)
	}
	return result
}

func (w *walker) checkFieldTypes(field *ast.Field) {
	t := w.c.cfg.Info.TypeOf(field.Type)
	if t == nil {
		return
	}
	if reason, detail := w.c.typeRefusal(t); reason != "" {
		w.refuse(field.Type.Pos(), reason, detail)
	}
}

// allowedNamed are named types from other packages a value may have.
var allowedNamed = map[string]bool{
	"boardgame.PlayerIndex":                true,
	"boardgame.ImmutableState":             true,
	"boardgame.State":                      true,
	"boardgame.ImmutableSubState":          true,
	"boardgame.SubState":                   true,
	"boardgame.ConfigurableSubState":       true,
	"boardgame.ImmutableStack":             true,
	"boardgame.Stack":                      true,
	"boardgame.ImmutableSizedStack":        true,
	"boardgame.SizedStack":                 true,
	"boardgame.ImmutableMergedStack":       true,
	"boardgame.MergedStack":                true,
	"boardgame.Component":                  true,
	"boardgame.ComponentInstance":          true,
	"boardgame.ImmutableComponentInstance": true,
	"boardgame.ComponentValues":            true,
	"boardgame.Timer":                      true,
	"boardgame.ImmutableTimer":             true,
	"boardgame.Move":                       true,
	"boardgame.Variant":                    true,
	"enum.EnumKey":                         true,
	"enum.Val":                             true,
	"enum.ImmutableVal":                    true,
	"enum.RangeVal":                        true,
	"enum.ImmutableRangeVal":               true,
	"enum.Enum":                            true,
	"enum.RangeEnum":                       true,
	"enum.TreeEnum":                        true,
	"graph.Graph":                          true,
	"time.Duration":                        true,
	"error":                                true,
}

// allowedPointers are pointers to other packages' structs a value may be.
var allowedPointers = map[string]bool{
	"rand.Rand":             true,
	"graph.EnumGraph":       true,
	"dice.DynamicValue":     true,
	"playingcards.Card":     true,
	"boardgame.Game":        true,
	"boardgame.Deck":        true,
	"boardgame.GameManager": true,
}

// typeRefusal is the Types section of the proposal: the reason a value of
// type t is outside the subset, or "".
func (c *Checker) typeRefusal(t types.Type) (string, string) {
	t = types.Unalias(t)
	switch t := t.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.Int, types.Bool, types.String, types.Float64,
			types.UntypedInt, types.UntypedBool, types.UntypedString,
			types.UntypedFloat, types.UntypedRune, types.UntypedNil, types.Invalid:
			return "", ""
		case types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr:
			return refusal.SizedInt, t.Name()
		case types.UnsafePointer:
			return refusal.Unsafe, t.Name()
		}
		return refusal.UnknownType, t.Name()
	case *types.TypeParam:
		return refusal.Generic, t.String()
	case *types.Named:
		if t.TypeArgs() != nil && t.TypeArgs().Len() > 0 {
			return refusal.Generic, typeName(t)
		}
		obj := t.Obj()
		if obj.Pkg() == c.cfg.Pkg {
			switch u := t.Underlying().(type) {
			case *types.Struct:
				return refusal.StructCopy, typeName(t)
			case *types.Interface:
				return refusal.UnknownType, typeName(t)
			default:
				return c.typeRefusal(u)
			}
		}
		if allowedNamed[typeName(t)] {
			return "", ""
		}
		if basic, ok := t.Underlying().(*types.Basic); ok {
			if reason, _ := c.typeRefusal(basic); reason == "" {
				return "", ""
			}
		}
		if _, ok := t.Underlying().(*types.Struct); ok {
			return refusal.StructCopy, typeName(t)
		}
		return refusal.UnknownType, typeName(t)
	case *types.Pointer:
		elem := types.Unalias(t.Elem())
		if named, ok := elem.(*types.Named); ok {
			if _, isStruct := named.Underlying().(*types.Struct); isStruct {
				if named.Obj().Pkg() == c.cfg.Pkg || allowedPointers[typeName(named)] {
					return "", ""
				}
				return refusal.UnknownType, "*" + typeName(named)
			}
		}
		return refusal.PointerLocal, types.TypeString(t, qualifier)
	case *types.Slice:
		elem := types.Unalias(t.Elem())
		if _, ok := elem.Underlying().(*types.Slice); ok {
			return refusal.UnknownType, types.TypeString(t, qualifier)
		}
		return c.typeRefusal(elem)
	case *types.Array:
		return refusal.FixedArray, types.TypeString(t, qualifier)
	case *types.Map:
		key, ok := t.Key().Underlying().(*types.Basic)
		if !ok || (key.Info()&(types.IsInteger|types.IsString)) == 0 {
			return refusal.UnknownType, types.TypeString(t, qualifier)
		}
		if reason, detail := c.typeRefusal(t.Key()); reason != "" {
			return reason, detail
		}
		return c.typeRefusal(t.Elem())
	case *types.Chan:
		return refusal.Goroutine, types.TypeString(t, qualifier)
	case *types.Signature:
		return refusal.ClosureEscape, types.TypeString(t, qualifier)
	case *types.Struct:
		return refusal.StructCopy, types.TypeString(t, qualifier)
	case *types.Interface:
		return refusal.UnknownType, types.TypeString(t, qualifier)
	case *types.Tuple:
		for i := 0; i < t.Len(); i++ {
			if reason, detail := c.typeRefusal(t.At(i).Type()); reason != "" {
				return reason, detail
			}
		}
	}
	return "", ""
}

func qualifier(p *types.Package) string { return p.Name() }

func (w *walker) visit(n ast.Node) bool {
	info := w.c.cfg.Info
	switch n := n.(type) {
	case *ast.GoStmt:
		w.refuse(n.Pos(), refusal.Goroutine, "go")
	case *ast.SelectStmt:
		w.refuse(n.Pos(), refusal.Goroutine, "select")
	case *ast.SendStmt:
		w.refuse(n.Pos(), refusal.Goroutine, "send")
	case *ast.DeferStmt:
		w.refuse(n.Pos(), refusal.Defer, "defer")
	case *ast.BranchStmt:
		switch n.Tok {
		case token.GOTO:
			w.refuse(n.Pos(), refusal.Goto, "goto")
		case token.FALLTHROUGH:
			w.refuse(n.Pos(), refusal.Fallthrough, "fallthrough")
		}
	case *ast.TypeSwitchStmt:
		w.refuse(n.Pos(), refusal.TypeSwitch, "type switch")
	case *ast.RangeStmt:
		w.visitRange(n)
	case *ast.AssignStmt:
		w.visitAssign(n.Lhs, n.Tok)
		if n.Tok == token.DEFINE {
			for _, lhs := range n.Lhs {
				w.checkDefinedIdent(lhs)
			}
		}
	case *ast.IncDecStmt:
		w.visitAssign([]ast.Expr{n.X}, n.Tok)
	case *ast.ValueSpec:
		for _, name := range n.Names {
			w.checkDefinedIdent(name)
		}
	case *ast.CallExpr:
		w.visitCall(n)
	case *ast.FuncLit:
		if !w.funcArgAllowed(n) {
			w.refuse(n.Pos(), refusal.ClosureEscape, "function literal")
		}
		for _, list := range []*ast.FieldList{n.Type.Params, n.Type.Results} {
			if list == nil {
				continue
			}
			for _, field := range list.List {
				w.checkFieldTypes(field)
			}
		}
	case *ast.SelectorExpr:
		w.visitSelector(n)
	case *ast.Ident:
		w.visitIdent(n)
	case *ast.UnaryExpr:
		switch n.Op {
		case token.AND:
			if lit, ok := ast.Unparen(n.X).(*ast.CompositeLit); ok {
				w.refuse(n.Pos(), refusal.StructCopy, typeName(info.TypeOf(lit)))
			} else {
				w.refuse(n.Pos(), refusal.PointerLocal, types.ExprString(n.X))
			}
		case token.ARROW:
			w.refuse(n.Pos(), refusal.Goroutine, "receive")
		}
	case *ast.StarExpr:
		if tv, ok := info.Types[n]; ok && !tv.IsType() {
			if structOf(tv.Type) != nil {
				w.refuse(n.Pos(), refusal.StructCopy, "*"+types.ExprString(n.X))
			} else {
				w.refuse(n.Pos(), refusal.PointerLocal, "*"+types.ExprString(n.X))
			}
		}
	case *ast.IndexExpr:
		w.visitIndex(n.X, n)
	case *ast.IndexListExpr:
		w.refuse(n.Pos(), refusal.Generic, types.ExprString(n))
	case *ast.SliceExpr:
		w.visitSlice(n)
	case *ast.CompositeLit:
		w.visitComposite(n)
	case *ast.TypeAssertExpr:
		if n.Type != nil {
			w.visitAssert(n)
		}
	}
	if expr, ok := n.(ast.Expr); ok {
		w.checkDuration(expr)
	}
	return true
}

func (w *walker) checkDefinedIdent(expr ast.Expr) {
	id, ok := expr.(*ast.Ident)
	if !ok || id.Name == "_" {
		return
	}
	obj, ok := w.c.cfg.Info.Defs[id].(*types.Var)
	if !ok || obj == nil {
		return
	}
	if reason, detail := w.c.typeRefusal(obj.Type()); reason != "" {
		w.refuse(id.Pos(), reason, detail)
	}
}

func (w *walker) visitRange(n *ast.RangeStmt) {
	t := w.c.cfg.Info.TypeOf(n.X)
	if t == nil {
		return
	}
	switch u := t.Underlying().(type) {
	case *types.Map:
		w.refuse(n.Pos(), refusal.MapRange, types.ExprString(n.X))
	case *types.Chan:
		w.refuse(n.Pos(), refusal.Goroutine, "range over channel")
	case *types.Signature:
		w.refuse(n.Pos(), refusal.ClosureEscape, "range over function")
	case *types.Basic:
		if u.Info()&types.IsString != 0 {
			w.refuse(n.Pos(), refusal.StringBytes, "range over string")
		}
	}
	if n.Tok == token.DEFINE {
		w.checkDefinedIdent(n.Key)
		if n.Value != nil {
			w.checkDefinedIdent(n.Value)
		}
	} else if n.Tok == token.ASSIGN {
		var lhs []ast.Expr
		for _, e := range []ast.Expr{n.Key, n.Value} {
			if e != nil {
				lhs = append(lhs, e)
			}
		}
		w.visitAssign(lhs, token.ASSIGN)
	}
}

// rootIdent is the variable an assignment target writes through: x for x,
// x.F, x[i], *x.
func rootIdent(expr ast.Expr) *ast.Ident {
	for {
		switch e := ast.Unparen(expr).(type) {
		case *ast.Ident:
			return e
		case *ast.SelectorExpr:
			expr = e.X
		case *ast.IndexExpr:
			expr = e.X
		case *ast.StarExpr:
			expr = e.X
		default:
			return nil
		}
	}
}

func (w *walker) visitAssign(lhs []ast.Expr, tok token.Token) {
	for _, target := range lhs {
		root := rootIdent(target)
		if root == nil {
			continue
		}
		obj := w.c.cfg.Info.ObjectOf(root)
		if v, ok := obj.(*types.Var); ok && isPackageVar(v) {
			if w.writes == nil {
				w.writes = make(map[*ast.Ident]bool)
			}
			w.writes[root] = true
			w.refuse(target.Pos(), refusal.PackageVarWrite, qualifiedVar(v))
		}
	}
}

func isPackageVar(v *types.Var) bool {
	return v != nil && !v.IsField() && v.Pkg() != nil && v.Parent() == v.Pkg().Scope()
}

func qualifiedVar(v *types.Var) string {
	return v.Pkg().Name() + "." + v.Name()
}

func (w *walker) visitIdent(id *ast.Ident) {
	info := w.c.cfg.Info
	obj := info.Uses[id]
	switch obj := obj.(type) {
	case *types.Var:
		if !isPackageVar(obj) || w.writes[id] {
			return
		}
		if obj.Pkg() != w.c.cfg.Pkg {
			w.refuse(id.Pos(), refusal.PackageVarRead, qualifiedVar(obj))
			return
		}
		if w.c.varKind[obj] == varOther {
			w.refuse(id.Pos(), refusal.PackageVarRead, qualifiedVar(obj))
		}
	case *types.PkgName:
		switch obj.Imported().Path() {
		case "reflect":
			w.refuse(id.Pos(), refusal.Reflection, "reflect")
		case "unsafe":
			w.refuse(id.Pos(), refusal.Unsafe, "unsafe")
		}
	case *types.Func:
		if w.isCallee(id) {
			return
		}
		if w.funcArgAllowed(id) {
			return
		}
		w.refuse(id.Pos(), refusal.ClosureEscape, "function value "+QualifiedName(obj))
	}
}

// isCallee reports whether id names the function a call calls (f(), x.f(),
// pkg.f()).
func (w *walker) isCallee(id *ast.Ident) bool {
	var node ast.Node = id
	if parent, ok := w.parents[id].(*ast.SelectorExpr); ok && parent.Sel == id {
		node = parent
	}
	for {
		parent := w.parents[node]
		if paren, ok := parent.(*ast.ParenExpr); ok {
			node = paren
			continue
		}
		if call, ok := parent.(*ast.CallExpr); ok && ast.Unparen(call.Fun) == ast.Unparen(node.(ast.Expr)) {
			return true
		}
		return false
	}
}

// funcArgAllowed reports whether expr (a function literal or a function
// name) is passed straight to a call whose table entry takes a function.
func (w *walker) funcArgAllowed(expr ast.Expr) bool {
	var node ast.Node = expr
	if parent, ok := w.parents[expr].(*ast.SelectorExpr); ok && parent.Sel == expr {
		node = parent
	}
	call, ok := w.parents[node].(*ast.CallExpr)
	if !ok || ast.Unparen(call.Fun) == node {
		return false
	}
	fn, _ := typeutil.Callee(w.c.cfg.Info, call).(*types.Func)
	if fn == nil {
		return false
	}
	entry, ok := Lookup(QualifiedName(fn))
	return ok && entry.Kind == KindFunc
}

func (w *walker) visitSelector(sel *ast.SelectorExpr) {
	info := w.c.cfg.Info
	selection := info.Selections[sel]
	if selection == nil {
		return
	}
	if key := gapOnPath(selection); key != "" {
		w.refuse(sel.Sel.Pos(), refusal.EngineGap, key)
	}
	switch selection.Kind() {
	case types.MethodVal:
		if !w.isCallee(sel.Sel) && !w.funcArgAllowed(sel.Sel) {
			w.refuse(sel.Pos(), refusal.ClosureEscape, "method value "+types.ExprString(sel))
		}
	case types.MethodExpr:
		w.refuse(sel.Pos(), refusal.ClosureEscape, "method expression "+types.ExprString(sel))
	}
}

// gapOnPath returns the EngineGaps key of an unported stock move or
// behavior that a field or method selection goes through (an embedded
// field on the way, or the type that declares the field), or "".
func gapOnPath(selection *types.Selection) string {
	t := selection.Recv()
	path := selection.Index()
	for i, index := range path {
		strct := structOf(t)
		if strct == nil || index >= strct.NumFields() {
			return ""
		}
		if named := namedOf(t); named != nil {
			if key := frameworkKey(named); key != "" && IsGap(key) {
				return key
			}
		}
		field := strct.Field(index)
		if i == len(path)-1 {
			return ""
		}
		if field.Embedded() {
			if named := namedOf(field.Type()); named != nil {
				if key := frameworkKey(named); key != "" && IsGap(key) {
					return key
				}
			}
		}
		t = field.Type()
	}
	return ""
}

func namedOf(t types.Type) *types.Named {
	t = types.Unalias(t)
	if pointer, ok := t.(*types.Pointer); ok {
		t = types.Unalias(pointer.Elem())
	}
	named, _ := t.(*types.Named)
	return named
}

func structOf(t types.Type) *types.Struct {
	t = types.Unalias(t)
	if pointer, ok := t.Underlying().(*types.Pointer); ok {
		t = pointer.Elem()
	}
	s, _ := t.Underlying().(*types.Struct)
	return s
}

// frameworkKey is "moves.<Type>" or "behaviors.<Type>" for a type declared
// in one of those packages, else "".
func frameworkKey(named *types.Named) string {
	pkg := named.Obj().Pkg()
	if pkg == nil {
		return ""
	}
	switch pkg.Path() {
	case movesPath:
		return "moves." + named.Obj().Name()
	case behaviorsPath:
		return "behaviors." + named.Obj().Name()
	}
	return ""
}

func (w *walker) visitCall(call *ast.CallExpr) {
	info := w.c.cfg.Info
	if tv, ok := info.Types[call.Fun]; ok && tv.IsType() {
		w.visitConversion(call, tv.Type)
		return
	}
	if id, ok := ast.Unparen(call.Fun).(*ast.Ident); ok {
		if builtin, ok := info.Uses[id].(*types.Builtin); ok {
			w.visitBuiltin(call, builtin.Name())
			return
		}
	}
	if _, ok := ast.Unparen(call.Fun).(*ast.FuncLit); ok {
		w.refuse(call.Pos(), refusal.ClosureEscape, "immediately invoked function literal")
		return
	}
	if inst, ok := info.Instances[calleeIdent(call.Fun)]; ok && inst.TypeArgs != nil && inst.TypeArgs.Len() > 0 {
		w.refuse(call.Pos(), refusal.Generic, types.ExprString(call.Fun))
	}
	callee := typeutil.Callee(info, call)
	fn, ok := callee.(*types.Func)
	if !ok {
		if callee != nil {
			w.refuse(call.Pos(), refusal.ClosureEscape, "call of function value "+types.ExprString(call.Fun))
		} else {
			w.refuse(call.Pos(), refusal.ClosureEscape, "call of "+types.ExprString(call.Fun))
		}
		return
	}
	if fn.Pkg() == w.c.cfg.Pkg {
		if fn.Type().(*types.Signature).Recv() != nil && w.interfaceDeclaredByGame(fn) {
			w.refuse(call.Pos(), refusal.InterfaceDispatch, QualifiedName(fn))
		}
		return
	}
	if fn.Pkg() != nil && fn.Pkg().Path() == movesPath {
		if sel, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr); ok && w.isStockReceiver(sel.X) {
			w.visitStockCall(call, sel, fn)
			return
		}
	}
	qualified := QualifiedName(fn)
	pkgPath := ""
	if fn.Pkg() != nil {
		pkgPath = fn.Pkg().Path()
	}
	isMethod := fn.Type().(*types.Signature).Recv() != nil
	switch {
	case pkgPath == "reflect":
		w.refuse(call.Pos(), refusal.Reflection, qualified)
		return
	case pkgPath == "unsafe":
		w.refuse(call.Pos(), refusal.Unsafe, qualified)
		return
	case isRandPackage(pkgPath) && !isMethod:
		w.refuse(call.Pos(), refusal.ForeignRand, qualified)
		return
	case isIO(pkgPath, fn.Name(), isMethod):
		w.refuse(call.Pos(), refusal.IO, qualified)
		return
	case isValueFromString(fn):
		w.refuse(call.Pos(), refusal.UnknownCall, ValueFromStringDetail)
		return
	}
	entry, ok := Lookup(qualified)
	if !ok {
		if isMethod && w.interfaceDeclaredByGame(fn) {
			w.refuse(call.Pos(), refusal.InterfaceDispatch, qualified)
			return
		}
		w.refuse(call.Pos(), refusal.UnknownCall, qualified)
		return
	}
	switch entry.Kind {
	case KindChain:
		if !w.chainContinues(call) {
			w.refuse(call.Pos(), refusal.UnknownCall, qualified)
			return
		}
	case KindFormat:
		w.checkFormat(call, qualified)
	}
	if entry.Gap() {
		w.refuse(call.Pos(), refusal.EngineGap, entry.Facade)
	}
}

func calleeIdent(fun ast.Expr) *ast.Ident {
	switch f := ast.Unparen(fun).(type) {
	case *ast.Ident:
		return f
	case *ast.SelectorExpr:
		return f.Sel
	case *ast.IndexExpr:
		return calleeIdent(f.X)
	case *ast.IndexListExpr:
		return calleeIdent(f.X)
	}
	return nil
}

func isIO(pkgPath, name string, isMethod bool) bool {
	switch pkgPath {
	case "os", "log", "log/slog", "encoding/json", "io", "io/ioutil", "net/http", "bufio":
		return true
	case "time":
		switch name {
		case "Now", "Since", "Until", "Sleep", "After", "Tick", "NewTimer", "NewTicker", "AfterFunc":
			return !isMethod
		}
	case "fmt":
		return strings.HasPrefix(name, "Print") || strings.HasPrefix(name, "Fprint") ||
			strings.HasPrefix(name, "Scan") || strings.HasPrefix(name, "Fscan") || strings.HasPrefix(name, "Sscan")
	}
	return false
}

func (w *walker) interfaceDeclaredByGame(fn *types.Func) bool {
	recv := fn.Type().(*types.Signature).Recv()
	named := namedOf(recv.Type())
	if named == nil {
		_, isInterface := recv.Type().Underlying().(*types.Interface)
		return isInterface
	}
	_, isInterface := named.Underlying().(*types.Interface)
	return isInterface && named.Obj().Pkg() == w.c.cfg.Pkg
}

// chainContinues reports whether call is the receiver of a further call the
// table knows: state.Game() in state.Game().MoveByName(n).
func (w *walker) chainContinues(call *ast.CallExpr) bool {
	sel, ok := w.parents[call].(*ast.SelectorExpr)
	if !ok || sel.X != call {
		return false
	}
	outer, ok := w.parents[sel].(*ast.CallExpr)
	if !ok || outer.Fun != sel {
		return false
	}
	fn, _ := typeutil.Callee(w.c.cfg.Info, outer).(*types.Func)
	if fn == nil {
		return false
	}
	_, known := Lookup(QualifiedName(fn))
	return known
}

// isStockReceiver reports whether expr selects an embedded field whose type
// is declared in package moves: m.CurrentPlayer, m.FixUp.
func (w *walker) isStockReceiver(expr ast.Expr) bool {
	sel, ok := ast.Unparen(expr).(*ast.SelectorExpr)
	if !ok {
		return false
	}
	selection := w.c.cfg.Info.Selections[sel]
	if selection == nil || selection.Kind() != types.FieldVal {
		return false
	}
	field, ok := selection.Obj().(*types.Var)
	if !ok || !field.Embedded() {
		return false
	}
	named := namedOf(field.Type())
	return named != nil && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == movesPath
}

func (w *walker) visitStockCall(call *ast.CallExpr, sel *ast.SelectorExpr, fn *types.Func) {
	stock := namedOf(w.c.cfg.Info.TypeOf(sel.X))
	key := frameworkKey(stock)
	qualified := "(" + key + ")." + fn.Name()
	if _, ok := StockParts[fn.Name()]; !ok {
		w.refuse(call.Pos(), refusal.UnknownCall, qualified)
		return
	}
	if fn.Name() == "DefaultsForState" && !w.isFirstStatement(call) {
		w.refuse(call.Pos(), refusal.UnknownCall, qualified+" (not the first statement)")
		return
	}
	if IsGap(key) {
		w.refuse(call.Pos(), refusal.EngineGap, key)
	}
}

// isFirstStatement reports whether call is an expression statement that is
// the first of the function body once erased statements (game, players :=
// concreteStates(state)) are skipped.
func (w *walker) isFirstStatement(call *ast.CallExpr) bool {
	stmt, ok := w.parents[call].(*ast.ExprStmt)
	if !ok || w.body == nil {
		return false
	}
	for _, s := range w.body.List {
		if s == stmt {
			return true
		}
		if !w.erasedStmt(s) {
			return false
		}
	}
	return false
}

func (w *walker) erasedStmt(stmt ast.Stmt) bool {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		for _, rhs := range s.Rhs {
			if !w.erasedExpr(rhs) {
				return false
			}
		}
		return len(s.Rhs) > 0
	case *ast.DeclStmt:
		gen, ok := s.Decl.(*ast.GenDecl)
		if !ok {
			return false
		}
		for _, spec := range gen.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok || len(value.Values) == 0 {
				return false
			}
			for _, rhs := range value.Values {
				if !w.erasedExpr(rhs) {
					return false
				}
			}
		}
		return true
	}
	return false
}

func (w *walker) erasedExpr(expr ast.Expr) bool {
	switch e := ast.Unparen(expr).(type) {
	case *ast.TypeAssertExpr:
		return w.erasedExpr(e.X)
	case *ast.CallExpr:
		fn, _ := typeutil.Callee(w.c.cfg.Info, e).(*types.Func)
		if fn == nil {
			return false
		}
		if fn.Pkg() == w.c.cfg.Pkg {
			return w.c.cfg.Erased != nil && w.c.cfg.Erased(fn)
		}
		entry, ok := Lookup(QualifiedName(fn))
		return ok && entry.Kind == KindErased
	}
	return false
}

func (w *walker) visitConversion(call *ast.CallExpr, target types.Type) {
	if len(call.Args) != 1 {
		return
	}
	source := w.c.cfg.Info.TypeOf(call.Args[0])
	if source == nil {
		return
	}
	targetBasic, targetIsBasic := target.Underlying().(*types.Basic)
	if _, ok := target.Underlying().(*types.Slice); ok {
		if b, ok := source.Underlying().(*types.Basic); ok && b.Info()&types.IsString != 0 {
			w.refuse(call.Pos(), refusal.StringBytes, types.TypeString(target, qualifier)+"(string)")
			return
		}
	}
	if targetIsBasic && targetBasic.Info()&types.IsString != 0 {
		if b, ok := source.Underlying().(*types.Basic); ok && b.Info()&types.IsInteger != 0 {
			w.refuse(call.Pos(), refusal.StringBytes, "string("+types.TypeString(source, qualifier)+")")
			return
		}
		if _, ok := source.Underlying().(*types.Slice); ok {
			w.refuse(call.Pos(), refusal.StringBytes, "string("+types.TypeString(source, qualifier)+")")
			return
		}
	}
	if reason, detail := w.c.typeRefusal(target); reason != "" {
		w.refuse(call.Pos(), reason, detail)
	}
}

func (w *walker) visitBuiltin(call *ast.CallExpr, name string) {
	info := w.c.cfg.Info
	switch name {
	case "len":
		if len(call.Args) == 1 {
			if t := info.TypeOf(call.Args[0]); t != nil {
				switch t.Underlying().(type) {
				case *types.Map:
					w.refuse(call.Pos(), refusal.UnknownCall, "len(map)")
				case *types.Chan:
					w.refuse(call.Pos(), refusal.Goroutine, "len(chan)")
				}
			}
		}
	case "append":
		if !w.selfAppend(call) {
			w.refuse(call.Pos(), refusal.SliceAlias, "append not in x = append(x, ...) form")
		}
	case "make", "copy", "delete", "min", "max":
	case "print", "println":
		w.refuse(call.Pos(), refusal.IO, name)
	case "close":
		w.refuse(call.Pos(), refusal.Goroutine, name)
	case "recover":
		w.refuse(call.Pos(), refusal.Defer, name)
	case "complex", "real", "imag":
		w.refuse(call.Pos(), refusal.UnknownType, name)
	default:
		w.refuse(call.Pos(), refusal.UnknownCall, name)
	}
}

// selfAppend reports whether call is the right side of x = append(x, ...).
func (w *walker) selfAppend(call *ast.CallExpr) bool {
	assign, ok := w.parents[call].(*ast.AssignStmt)
	if !ok || assign.Tok != token.ASSIGN || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 || assign.Rhs[0] != call {
		return false
	}
	if len(call.Args) == 0 {
		return false
	}
	return types.ExprString(assign.Lhs[0]) == types.ExprString(call.Args[0])
}

func (w *walker) visitIndex(x ast.Expr, n ast.Node) {
	t := w.c.cfg.Info.TypeOf(x)
	if t == nil {
		return
	}
	if b, ok := t.Underlying().(*types.Basic); ok && b.Info()&types.IsString != 0 {
		w.refuse(n.Pos(), refusal.StringBytes, "index of string")
	}
	if id := calleeIdent(x); id != nil {
		if _, ok := w.c.cfg.Info.Instances[id]; ok {
			w.refuse(n.Pos(), refusal.Generic, types.ExprString(x))
		}
	}
}

func (w *walker) visitSlice(n *ast.SliceExpr) {
	t := w.c.cfg.Info.TypeOf(n.X)
	if t == nil {
		return
	}
	if b, ok := t.Underlying().(*types.Basic); ok && b.Info()&types.IsString != 0 {
		w.refuse(n.Pos(), refusal.StringBytes, "slice of string")
		return
	}
	// A slice expression may only be read: never the first argument of
	// append or the destination of copy, and never bound to a variable that
	// is later appended to, index-assigned or copied into.
	parent := w.parents[n]
	if call, ok := parent.(*ast.CallExpr); ok && len(call.Args) > 0 && call.Args[0] == n {
		if id, ok := ast.Unparen(call.Fun).(*ast.Ident); ok {
			if builtin, ok := w.c.cfg.Info.Uses[id].(*types.Builtin); ok && (builtin.Name() == "append" || builtin.Name() == "copy") {
				w.refuse(n.Pos(), refusal.SliceAlias, types.ExprString(n))
				return
			}
		}
	}
	var bound *ast.Ident
	switch p := parent.(type) {
	case *ast.AssignStmt:
		for i, rhs := range p.Rhs {
			if rhs == n && i < len(p.Lhs) {
				bound, _ = p.Lhs[i].(*ast.Ident)
			}
		}
	case *ast.ValueSpec:
		for i, value := range p.Values {
			if value == n && i < len(p.Names) {
				bound = p.Names[i]
			}
		}
	}
	if bound == nil || w.body == nil {
		return
	}
	obj := w.c.cfg.Info.ObjectOf(bound)
	if obj == nil {
		return
	}
	if w.writtenThrough(obj) {
		w.refuse(n.Pos(), refusal.SliceAlias, types.ExprString(n))
	}
}

// writtenThrough reports whether the function appends to, index-assigns or
// copies into the variable obj.
func (w *walker) writtenThrough(obj types.Object) bool {
	info := w.c.cfg.Info
	is := func(expr ast.Expr) bool {
		id, ok := ast.Unparen(expr).(*ast.Ident)
		return ok && info.ObjectOf(id) == obj
	}
	found := false
	ast.Inspect(w.body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.CallExpr:
			if id, ok := ast.Unparen(n.Fun).(*ast.Ident); ok && len(n.Args) > 0 {
				if builtin, ok := info.Uses[id].(*types.Builtin); ok && (builtin.Name() == "append" || builtin.Name() == "copy") && is(n.Args[0]) {
					found = true
				}
			}
		case *ast.AssignStmt:
			for _, lhs := range n.Lhs {
				if index, ok := ast.Unparen(lhs).(*ast.IndexExpr); ok && is(index.X) {
					found = true
				}
			}
		case *ast.IncDecStmt:
			if index, ok := ast.Unparen(n.X).(*ast.IndexExpr); ok && is(index.X) {
				found = true
			}
		}
		return !found
	})
	return found
}

func (w *walker) visitComposite(n *ast.CompositeLit) {
	t := w.c.cfg.Info.TypeOf(n)
	if t == nil {
		return
	}
	switch t.Underlying().(type) {
	case *types.Slice:
		if reason, detail := w.c.typeRefusal(t); reason != "" {
			w.refuse(n.Pos(), reason, detail)
		}
	case *types.Array:
		w.refuse(n.Pos(), refusal.FixedArray, types.TypeString(t, qualifier))
	case *types.Map:
		w.refuse(n.Pos(), refusal.UnknownType, "map literal "+types.TypeString(t, qualifier))
	case *types.Struct:
		if _, ok := w.parents[n].(*ast.UnaryExpr); ok {
			return // reported once, at the &
		}
		w.refuse(n.Pos(), refusal.StructCopy, typeName(t))
	default:
		w.refuse(n.Pos(), refusal.UnknownType, types.TypeString(t, qualifier))
	}
}

func (w *walker) visitAssert(n *ast.TypeAssertExpr) {
	t := w.c.cfg.Info.TypeOf(n.Type)
	if t == nil {
		return
	}
	if named := namedOf(t); named != nil {
		if _, isPointer := types.Unalias(t).(*types.Pointer); isPointer {
			if _, isStruct := named.Underlying().(*types.Struct); isStruct {
				if named.Obj().Pkg() == w.c.cfg.Pkg || allowedPointers[typeName(named)] {
					return
				}
			}
		}
	}
	w.refuse(n.Pos(), refusal.UnknownType, "type assertion to "+types.TypeString(t, qualifier))
}

var durationsPerMillisecond = constant.MakeInt64(1000000)

func (w *walker) checkDuration(expr ast.Expr) {
	tv, ok := w.c.cfg.Info.Types[expr]
	if !ok || tv.Value == nil || tv.Type == nil {
		return
	}
	named := namedOf(tv.Type)
	if named == nil || named.Obj().Pkg() == nil || named.Obj().Pkg().Path() != "time" || named.Obj().Name() != "Duration" {
		return
	}
	if parent, ok := w.parents[expr].(ast.Expr); ok {
		if ptv, ok := w.c.cfg.Info.Types[parent]; ok && ptv.Value != nil {
			return // checked at the outermost constant
		}
	}
	value := constant.ToInt(tv.Value)
	if value.Kind() != constant.Int {
		return
	}
	remainder := constant.BinaryOp(value, token.REM, durationsPerMillisecond)
	if constant.Sign(remainder) != 0 {
		w.refuse(expr.Pos(), refusal.SubMsDuration, value.ExactString()+"ns")
	}
}

// checkFormat refuses a format string the facade's sprintf does not port:
// a non-constant format, or a verb other than %d %s %v %q.
func (w *walker) checkFormat(call *ast.CallExpr, qualified string) {
	if len(call.Args) == 0 {
		return
	}
	tv, ok := w.c.cfg.Info.Types[call.Args[0]]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		w.refuse(call.Pos(), refusal.UnknownCall, qualified+" (non-constant format)")
		return
	}
	format := constant.StringVal(tv.Value)
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}
		i++
		if i >= len(format) {
			break
		}
		switch format[i] {
		case '%', 'd', 's', 'v', 'q':
		default:
			w.refuse(call.Pos(), refusal.UnknownCall, qualified+" (verb %"+string(format[i])+")")
		}
	}
}

// classifyVars applies the package-variable rule to every package variable:
// a variable written anywhere outside its declaration is not constant; one
// initialized from constants (or not at all) is; one built by a
// FrameworkConstructor from constant arguments and other framework
// variables is a framework object.
func (c *Checker) classifyVars() {
	info := c.cfg.Info
	written := make(map[*types.Var]bool)
	for _, file := range c.cfg.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			var targets []ast.Expr
			switch n := n.(type) {
			case *ast.AssignStmt:
				targets = n.Lhs
			case *ast.IncDecStmt:
				targets = []ast.Expr{n.X}
			case *ast.RangeStmt:
				if n.Tok == token.ASSIGN {
					targets = []ast.Expr{n.Key, n.Value}
				}
			case *ast.UnaryExpr:
				if n.Op == token.AND {
					targets = []ast.Expr{n.X}
				}
			}
			for _, target := range targets {
				if target == nil {
					continue
				}
				if root := rootIdent(target); root != nil {
					if v, ok := info.ObjectOf(root).(*types.Var); ok && isPackageVar(v) {
						written[v] = true
					}
				}
			}
			return true
		})
	}
	type pending struct {
		v    *types.Var
		init ast.Expr
	}
	var vars []pending
	for _, file := range c.cfg.Files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				value := spec.(*ast.ValueSpec)
				for i, name := range value.Names {
					v, ok := info.Defs[name].(*types.Var)
					if !ok {
						continue
					}
					var init ast.Expr
					if len(value.Values) == len(value.Names) {
						init = value.Values[i]
					} else if len(value.Values) > 0 {
						init = value.Values[0] // a tuple assignment: never constant
						written[v] = true
					}
					vars = append(vars, pending{v, init})
				}
			}
		}
	}
	// Framework variables may depend on each other (enums, then spacesEnum,
	// then a graph over it): iterate to a fixed point.
	for changed := true; changed; {
		changed = false
		for _, p := range vars {
			if _, done := c.varKind[p.v]; done || written[p.v] {
				continue
			}
			kind := varOther
			switch {
			case p.init == nil || c.constantExpr(p.init):
				kind = varConstant
			case c.frameworkExpr(p.init):
				kind = varFramework
			default:
				continue
			}
			c.varKind[p.v] = kind
			changed = true
		}
	}
	for _, p := range vars {
		if _, done := c.varKind[p.v]; !done {
			c.varKind[p.v] = varOther
		}
	}
}

// constantExpr reports whether expr is built only from constants: a
// constant expression, or a slice or map literal of them.
func (c *Checker) constantExpr(expr ast.Expr) bool {
	if tv, ok := c.cfg.Info.Types[expr]; ok && tv.Value != nil {
		return true
	}
	switch e := ast.Unparen(expr).(type) {
	case *ast.CompositeLit:
		t := c.cfg.Info.TypeOf(e)
		if t == nil {
			return false
		}
		switch t.Underlying().(type) {
		case *types.Slice, *types.Map:
		default:
			return false
		}
		for _, elt := range e.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				if !c.constantExpr(kv.Key) || !c.constantExpr(kv.Value) {
					return false
				}
				continue
			}
			if !c.constantExpr(elt) {
				return false
			}
		}
		return true
	case *ast.Ident:
		return e.Name == "nil"
	}
	return false
}

// frameworkExpr reports whether expr is a FrameworkConstructor call whose
// receiver and arguments are constants, recognized framework variables or
// FrameworkFuncValues.
func (c *Checker) frameworkExpr(expr ast.Expr) bool {
	call, ok := ast.Unparen(expr).(*ast.CallExpr)
	if !ok {
		return false
	}
	fn, _ := typeutil.Callee(c.cfg.Info, call).(*types.Func)
	if fn == nil || !FrameworkConstructors[QualifiedName(fn)] {
		return false
	}
	if sel, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr); ok {
		if _, isPkg := c.cfg.Info.Uses[identOf(sel.X)].(*types.PkgName); !isPkg {
			if !c.frameworkOperand(sel.X) {
				return false
			}
		}
	}
	for _, arg := range call.Args {
		if !c.frameworkOperand(arg) {
			return false
		}
	}
	return true
}

func identOf(expr ast.Expr) *ast.Ident {
	id, _ := ast.Unparen(expr).(*ast.Ident)
	return id
}

func (c *Checker) frameworkOperand(expr ast.Expr) bool {
	if c.constantExpr(expr) {
		return true
	}
	info := c.cfg.Info
	var id *ast.Ident
	switch e := ast.Unparen(expr).(type) {
	case *ast.Ident:
		id = e
	case *ast.SelectorExpr:
		id = e.Sel
	default:
		return c.frameworkExpr(expr)
	}
	switch obj := info.Uses[id].(type) {
	case *types.Var:
		if obj.Pkg() == c.cfg.Pkg && isPackageVar(obj) {
			kind, ok := c.varKind[obj]
			return ok && kind != varOther
		}
	case *types.Func:
		return FrameworkFuncValues[QualifiedName(obj)]
	}
	return false
}

// children lists the direct child nodes of n, for the parent map.
func children(n ast.Node) []ast.Node {
	var result []ast.Node
	first := true
	ast.Inspect(n, func(child ast.Node) bool {
		if first {
			first = false
			return true
		}
		if child != nil {
			result = append(result, child)
		}
		return false
	})
	return result
}
