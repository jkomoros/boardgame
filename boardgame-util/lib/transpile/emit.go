package transpile

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
)

// Modules the generated bodies import, relative to the engine directory.
const (
	modAPI    = "go/api.ts"
	modNum    = "go/num.ts"
	modSlice  = "go/slice.ts"
	modStr    = "go/str.ts"
	modPanic  = "go/panic.ts"
	modState  = "rules/state.ts"
	modTypes  = "rules/types.ts"
	modLegal  = "legal/types.ts"
	modRand   = "rand.ts"
	typesOnly = "type:"
)

// facadeNames are every name the emitter may import: a Go local of the
// same name is renamed, so an import is never shadowed.
var facadeNames = map[string]string{
	"player": modAPI, "playerIndexes": modAPI, "gameStack": modAPI, "playerStack": modAPI, "stackView": modAPI,
	"componentAt": modAPI, "first": modAPI, "last": modAPI, "components": modAPI, "componentAtKey": modAPI,
	"values": modAPI, "dynamic": modAPI, "valuesOf": modAPI, "dynamicOf": modAPI, "containingStack": modAPI,
	"moveTo": modAPI, "secretMoveTo": modAPI, "moveToFirstSlot": modAPI, "moveToLastSlot": modAPI, "moveToNextSlot": modAPI,
	"slideToFirstSlot": modAPI, "slideToLastSlot": modAPI, "mayMoveTo": modAPI, "mayMoveToSlot": modAPI,
	"moveAllTo": modAPI, "moveCountTo": modAPI, "mayMoveAllTo": modAPI, "mayMoveCountTo": modAPI,
	"shuffle": modAPI, "publicShuffle": modAPI, "swapComponents": modAPI, "swapComponentsByKey": modAPI, "sortComponents": modAPI,
	"setSize": modAPI, "expandSize": modAPI, "contractSize": modAPI, "sizeToFit": modAPI,
	"variantValue": modAPI, "setEliminated": modAPI, "setPlayerInactive": modAPI, "setSeatFilled": modAPI,
	"enumKey": modAPI, "setEnumKey": modAPI, "enumEquals": modAPI, "rangeValue": modAPI, "rangeToValue": modAPI,
	"valueToRange": modAPI, "neighborsByKey": modAPI, "stockApply": modAPI, "legalError": modAPI, "diceRoll": modAPI,
	"idiv": modNum, "imod": modNum, "imul": modNum,
	"appendTo": modSlice, "lenOf": modSlice, "at": modSlice, "setAt": modSlice, "makeSlice": modSlice, "copyInto": modSlice, "subslice": modSlice,
	"join": modStr, "repeat": modStr, "split": modStr, "toLower": modStr, "trimSpace": modStr, "hasPrefix": modStr, "hasSuffix": modStr,
	"contains": modStr, "strCompare": modStr, "utf8Len": modStr, "itoa": modStr, "atoi": modStr, "sprintf": modStr,
	"nonNil":     modPanic,
	"RulesState": typesOnly + modState, "TimerHandle": typesOnly + modState, "MoveInstance": typesOnly + modTypes,
	"LegalVerdict": typesOnly + modLegal, "GoRand": typesOnly + modRand,
	"StackRef": typesOnly + modAPI, "Comp": typesOnly + modAPI, "DynamicRef": typesOnly + modAPI, "GraphRef": typesOnly + modAPI, "valuesAreNil": modAPI,
}

// reservedNames are names a Go local may not take in TypeScript.
var reservedNames = func() map[string]bool {
	words := strings.Fields(`break case catch class const continue debugger default delete do else enum export extends
		false finally for function if import in instanceof new null return super switch this throw true try typeof var void
		while with yield let static implements interface package private protected public await async of
		undefined NaN Infinity Math Map Error Object Array Number String Boolean JSON arguments eval
		state base deck index`)
	set := make(map[string]bool)
	for _, w := range words {
		set[w] = true
	}
	for name := range facadeNames {
		set[name] = true
	}
	return set
}()

// refusalPanic aborts one function's translation.
type refusalPanic struct{ r Refusal }

type local struct {
	name string
	k    kind
	// erased locals have no TypeScript value (the game handle, the state).
	erased bool
	// code is what a read emits for an erased-but-addressable local
	// (`state`), or "" to use name.
	code string
	// nullable marks a stack local that may hold nil.
	nullable bool
	t        types.Type
}

type emitter struct {
	p    *Package
	ref  string
	role string
	node ast.Node
	sig  *types.Signature
	body *ast.BlockStmt

	locals  map[types.Object]*local
	used    map[string]bool
	imports map[string]map[string]bool
	temps   int

	out    strings.Builder
	indent int

	// results are the Go results; named results are locals.
	results     *types.Tuple
	namedResult []*local
	// proposer is the legal role's proposer parameter, for base().
	proposer types.Object
	// moveParam is the TypeScript name of the move parameter.
	moveParam string
}

// emitFunction translates one function, or returns the emitter's refusal.
func (p *Package) emitFunction(ref, role string) (code string, imports map[string][]string, refused *Refusal) {
	node, ok := p.units[ref]
	if !ok {
		return "", nil, &Refusal{Func: ref, Reason: ReasonUnknownCall, Detail: "emitter: no syntax for " + ref}
	}
	e := &emitter{p: p, ref: ref, role: role, node: node, locals: make(map[types.Object]*local), used: make(map[string]bool), imports: make(map[string]map[string]bool)}
	defer func() {
		if r := recover(); r != nil {
			rp, ok := r.(refusalPanic)
			if !ok {
				panic(r)
			}
			code, imports, refused = "", nil, &rp.r
		}
	}()
	e.function()
	return e.out.String(), sortedImports(e.imports), nil
}

func (e *emitter) fail(node ast.Node, reason, detail string) {
	var pos token.Pos
	if node != nil {
		pos = node.Pos()
	}
	file, line, column := e.p.position(pos)
	panic(refusalPanic{Refusal{Func: e.ref, Reason: reason, Detail: "emitter: " + detail, File: file, Line: line, Column: column}})
}

func (e *emitter) use(name string) string {
	module, ok := facadeNames[name]
	if !ok {
		panic("transpile: unknown facade name " + name)
	}
	if e.imports[module] == nil {
		e.imports[module] = make(map[string]bool)
	}
	e.imports[module][name] = true
	return name
}

func (e *emitter) line(format string, args ...any) {
	e.out.WriteString(strings.Repeat("\t", e.indent))
	fmt.Fprintf(&e.out, format, args...)
	e.out.WriteByte('\n')
}

func (e *emitter) temp(prefix string) string {
	e.temps++
	return "$" + prefix + strconv.Itoa(e.temps)
}

// declare gives a Go object its TypeScript name.
func (e *emitter) declare(obj types.Object, ident *ast.Ident) *local {
	name := obj.Name()
	if reservedNames[name] || strings.HasPrefix(name, e.p.Name()+"_") {
		name += "_"
	}
	base := name
	for n := 2; e.used[name]; n++ {
		name = base + strconv.Itoa(n)
	}
	e.used[name] = true
	l := &local{name: name, k: e.p.kindOf(obj.Type()), t: obj.Type()}
	e.locals[obj] = l
	return l
}

func (e *emitter) erase(obj types.Object, code string) {
	if obj == nil {
		return
	}
	e.locals[obj] = &local{name: obj.Name(), k: e.p.kindOf(obj.Type()), erased: true, code: code, t: obj.Type()}
}

// ----- the function ---------------------------------------------------------

type param struct {
	obj  types.Object
	name string
	ts   string
}

func (e *emitter) function() {
	var recv *ast.FieldList
	var ftype *ast.FuncType
	switch n := e.node.(type) {
	case *ast.FuncDecl:
		recv, ftype, e.body = n.Recv, n.Type, n.Body
		fn := e.p.info.Defs[n.Name].(*types.Func)
		e.sig = fn.Type().(*types.Signature)
	case *ast.FuncLit:
		ftype, e.body = n.Type, n.Body
		e.sig = e.p.info.TypeOf(n).(*types.Signature)
	}
	if e.body == nil {
		e.fail(e.node, ReasonUnknownCall, "function without a body")
	}
	if e.sig.Variadic() {
		e.fail(e.node, ReasonUnknownType, "variadic function")
	}
	for name := range map[string]bool{"state": true, "base": true, "deck": true, "index": true} {
		e.used[name] = true
	}
	var goParams []types.Object
	if recv != nil && len(recv.List) > 0 {
		goParams = append(goParams, e.fieldObjects(recv)...)
	}
	goParams = append(goParams, e.fieldObjects(ftype.Params)...)
	e.results = e.sig.Results()

	name, err := Mangle(e.ref)
	if err != nil {
		e.fail(e.node, ReasonUnknownType, err.Error())
	}
	params, ret := e.signature(goParams)

	// Named results are locals initialized to their zero values.
	if ftype.Results != nil {
		for _, field := range ftype.Results.List {
			for _, id := range field.Names {
				obj := e.p.info.Defs[id]
				if obj == nil {
					continue
				}
				if id.Name == "_" {
					e.fail(id, ReasonUnknownType, "blank named result")
				}
				l := e.declare(obj, id)
				e.namedResult = append(e.namedResult, l)
			}
		}
	}

	var list []string
	for _, prm := range params {
		list = append(list, prm.name+": "+prm.ts)
	}
	e.line("export function %s(%s): %s {", name, strings.Join(list, ", "), ret)
	e.indent++
	if e.role == gamedef.RoleDistribute {
		for _, obj := range goParams {
			if l := e.locals[obj]; l != nil && l.k == kComponent && !l.erased && e.referenced(obj) {
				e.line("const %s: %s = { deck, index };", l.name, e.use("DynamicRef"))
			}
		}
	}
	for _, l := range e.namedResult {
		zero := e.p.zeroValue(l.t)
		ts := e.p.tsType(l.t)
		if zero == "" || ts == "" {
			e.fail(e.node, ReasonUnknownType, "named result of type "+l.t.String())
		}
		e.line("let %s: %s = %s;", l.name, e.typeRef(ts), zero)
	}
	e.stmts(e.body.List, true)
	e.indent--
	e.line("}")
}

// referenced reports whether the body uses obj.
func (e *emitter) referenced(obj types.Object) bool {
	found := false
	ast.Inspect(e.body, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && e.p.info.Uses[id] == obj {
			found = true
		}
		return !found
	})
	return found
}

func (e *emitter) fieldObjects(list *ast.FieldList) []types.Object {
	var out []types.Object
	if list == nil {
		return out
	}
	for _, field := range list.List {
		if len(field.Names) == 0 {
			// An unnamed parameter: a placeholder object of its type.
			out = append(out, types.NewParam(field.Pos(), e.p.tpkg, "_", e.p.info.TypeOf(field.Type)))
			continue
		}
		for _, id := range field.Names {
			obj := e.p.info.Defs[id]
			if obj == nil {
				obj = types.NewParam(id.Pos(), e.p.tpkg, "_", e.p.info.TypeOf(field.Type))
			}
			out = append(out, obj)
		}
	}
	return out
}

// typeRef records the type names a TypeScript type mentions as imports.
func (e *emitter) typeRef(ts string) string {
	for _, name := range []string{"RulesState", "TimerHandle", "MoveInstance", "LegalVerdict", "GoRand", "StackRef", "Comp", "DynamicRef", "GraphRef"} {
		if strings.Contains(ts, name) {
			for _, word := range strings.FieldsFunc(ts, func(r rune) bool {
				return !(r == '_' || r == '$' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9')
			}) {
				if word == name {
					e.use(name)
				}
			}
		}
	}
	return ts
}

// signature binds the Go parameters to TypeScript ones by role (the body
// calling convention of rules/data.ts) and returns the TypeScript return
// type.
func (e *emitter) signature(goParams []types.Object) ([]param, string) {
	params := []param{{name: "state", ts: e.typeRef("RulesState")}}
	bindOrdinary := func(obj types.Object) {
		k := e.p.kindOf(obj.Type())
		switch k {
		case kState:
			e.erase(obj, "state")
			return
		case kGame, kDelegate:
			e.erase(obj, "")
			return
		}
		if obj.Name() == "_" {
			ts := e.p.tsType(obj.Type())
			if ts == "" {
				e.fail(e.node, ReasonUnknownType, "parameter of type "+obj.Type().String())
			}
			params = append(params, param{name: e.temp("unused"), ts: e.typeRef(ts)})
			return
		}
		ts := e.p.tsType(obj.Type())
		if ts == "" || k == kEnumVal {
			e.fail(e.node, ReasonUnknownType, "parameter of type "+obj.Type().String())
		}
		l := e.declare(obj, nil)
		params = append(params, param{obj: obj, name: l.name, ts: e.typeRef(ts)})
		if k == kMove {
			e.moveParam = l.name
		}
	}
	errorResult := func() {
		if e.results.Len() != 1 || !isError(e.results.At(0).Type()) {
			e.fail(e.node, ReasonUnknownType, "role "+e.role+" without a single error result")
		}
	}
	switch e.role {
	case gamedef.RoleApply, gamedef.RoleSetUpBegin, gamedef.RoleSetUpFinish, gamedef.RoleResetForTurnStart, gamedef.RoleResetForTurnEnd, gamedef.RoleTurnDone:
		for _, obj := range goParams {
			bindOrdinary(obj)
		}
		errorResult()
		return params, "string | null"
	case gamedef.RoleLegal, gamedef.RoleLegalCustom:
		for _, obj := range goParams {
			if isPlayerIndex(obj.Type()) && e.proposer == nil {
				l := e.declare(obj, nil)
				e.proposer = obj
				params = append(params, param{obj: obj, name: l.name, ts: "number"})
				continue
			}
			bindOrdinary(obj)
		}
		errorResult()
		if e.role == gamedef.RoleLegal {
			params = append(params, param{name: "base", ts: "() => string | null"})
			return params, "string | null"
		}
		return params, "string | " + e.typeRef("LegalVerdict") + " | null"
	case gamedef.RoleDefaults:
		for _, obj := range goParams {
			bindOrdinary(obj)
		}
		if e.results.Len() != 0 {
			e.fail(e.node, ReasonUnknownType, "DefaultsForState with results")
		}
		return params, "void"
	case gamedef.RoleDistribute:
		for _, obj := range goParams {
			if e.p.kindOf(obj.Type()) == kComponent {
				if obj.Name() == "_" {
					continue
				}
				e.declare(obj, nil)
				continue
			}
			bindOrdinary(obj)
		}
		params = append(params, param{name: "deck", ts: "string"}, param{name: "index", ts: "number"})
		if e.results.Len() != 2 || e.p.kindOf(e.results.At(0).Type()) != kStack || !isError(e.results.At(1).Type()) {
			e.fail(e.node, ReasonUnknownType, "distribute without (stack, error) results")
		}
		return params, e.typeRef("StackRef")
	case gamedef.RoleComponent:
		// A component values type's Legal(state, legalType): the
		// DeckSpec.componentLegal shape, with the values handle first.
		// legalType is the key Go's enum value holds, or null for Go's nil;
		// the body may not use it (it has no owner to read).
		for _, obj := range goParams {
			if e.p.kindOf(obj.Type()) == kEnumVal {
				if e.referenced(obj) {
					e.fail(e.node, ReasonUnknownType, "a component Legal that reads its legalType")
				}
				params = append(params, param{name: e.temp("legalType"), ts: "number | null"})
				continue
			}
			bindOrdinary(obj)
		}
		errorResult()
		return params, "string | null"
	case gamedef.RoleCheckGameFinished:
		for _, obj := range goParams {
			bindOrdinary(obj)
		}
		return params, e.resultType()
	}
	for _, obj := range goParams {
		bindOrdinary(obj)
	}
	return params, e.resultType()
}

func isPlayerIndex(t types.Type) bool {
	named, ok := types.Unalias(t).(*types.Named)
	return ok && pkgPathOf(named) == boardgamePath && named.Obj().Name() == "PlayerIndex"
}

func (e *emitter) resultType() string {
	switch e.results.Len() {
	case 0:
		return "void"
	case 1:
		ts := e.p.tsType(e.results.At(0).Type())
		if ts == "" {
			e.fail(e.node, ReasonUnknownType, "result of type "+e.results.At(0).Type().String())
		}
		return e.typeRef(ts)
	}
	ts := e.p.tsType(e.results)
	if ts == "" {
		e.fail(e.node, ReasonUnknownType, "results "+e.results.String())
	}
	return e.typeRef(ts)
}

// calleeParams is how a call passes arguments to an in-package function
// translated by the generic rule: state first, then the receiver and the
// parameters whose kind has a value. It returns which Go parameter indexes
// (receiver as -1) are passed.
func (p *Package) passed(fn *types.Func) (recv bool, params []bool) {
	sig := fn.Type().(*types.Signature)
	if r := sig.Recv(); r != nil {
		k := p.kindOf(r.Type())
		recv = !(k == kState || k == kGame || k == kDelegate)
	}
	for i := 0; i < sig.Params().Len(); i++ {
		k := p.kindOf(sig.Params().At(i).Type())
		params = append(params, !(k == kState || k == kGame || k == kDelegate))
	}
	return recv, params
}

// ----- statements -----------------------------------------------------------

func (e *emitter) stmts(list []ast.Stmt, top bool) {
	for i, s := range list {
		if top && i == e.firstNonErased(list) {
			if e.droppedDefaults(s) {
				continue
			}
		}
		e.stmt(s)
	}
}

// firstNonErased is the index of the first statement that does more than
// bind erased handles (concreteStates).
func (e *emitter) firstNonErased(list []ast.Stmt) int {
	for i, s := range list {
		if !e.erasedBinding(s) {
			return i
		}
	}
	return len(list)
}

func (e *emitter) erasedBinding(s ast.Stmt) bool {
	assign, ok := s.(*ast.AssignStmt)
	if !ok || len(assign.Rhs) != 1 {
		return false
	}
	return e.erasedCall(assign.Rhs[0]) || e.erasedValue(assign.Rhs[0])
}

// droppedDefaults reports whether s is m.<Stock>.DefaultsForState(state),
// which the engine runs before the body.
func (e *emitter) droppedDefaults(s ast.Stmt) bool {
	expr, ok := s.(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := expr.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "DefaultsForState" || !e.stockReceiver(sel.X) {
		return false
	}
	return e.role == gamedef.RoleDefaults
}

func (e *emitter) block(list []ast.Stmt) {
	e.indent++
	e.stmts(list, false)
	e.indent--
}

func (e *emitter) stmt(s ast.Stmt) {
	switch s := s.(type) {
	case *ast.BlockStmt:
		e.line("{")
		e.block(s.List)
		e.line("}")
	case *ast.ExprStmt:
		call, ok := ast.Unparen(s.X).(*ast.CallExpr)
		if !ok {
			e.fail(s, ReasonUnknownCall, "expression statement that is not a call")
		}
		if sel, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr); ok && sel.Sel.Name == "DefaultsForState" && e.stockReceiver(sel.X) {
			e.fail(s, ReasonUnknownCall, "stock DefaultsForState not first")
		}
		v := e.call(call, true)
		if v.code == "" {
			return
		}
		e.line("%s;", v.code)
	case *ast.AssignStmt:
		e.assign(s)
	case *ast.IncDecStmt:
		op := "+"
		if s.Tok == token.DEC {
			op = "-"
		}
		e.compound(s.X, op, value{code: "1", t: e.p.info.TypeOf(s.X)}, s)
	case *ast.DeclStmt:
		gen, ok := s.Decl.(*ast.GenDecl)
		if !ok {
			e.fail(s, ReasonUnknownType, "declaration")
		}
		switch gen.Tok {
		case token.CONST, token.TYPE:
			// Constants are folded where they are used; types have no code.
			return
		case token.VAR:
		default:
			e.fail(s, ReasonUnknownType, "declaration "+gen.Tok.String())
		}
		for _, spec := range gen.Specs {
			e.varSpec(spec.(*ast.ValueSpec))
		}
	case *ast.IfStmt:
		e.ifStmt(s, false)
	case *ast.ForStmt:
		e.forStmt(s, "")
	case *ast.RangeStmt:
		e.rangeStmt(s, "")
	case *ast.SwitchStmt:
		e.switchStmt(s, "")
	case *ast.ReturnStmt:
		e.returnStmt(s)
	case *ast.BranchStmt:
		label := ""
		if s.Label != nil {
			label = " " + s.Label.Name
		}
		switch s.Tok {
		case token.BREAK:
			e.line("break%s;", label)
		case token.CONTINUE:
			e.line("continue%s;", label)
		default:
			e.fail(s, ReasonGoto, s.Tok.String())
		}
	case *ast.LabeledStmt:
		switch inner := s.Stmt.(type) {
		case *ast.ForStmt:
			e.forStmt(inner, s.Label.Name)
		case *ast.RangeStmt:
			e.rangeStmt(inner, s.Label.Name)
		case *ast.SwitchStmt:
			e.switchStmt(inner, s.Label.Name)
		default:
			e.fail(s, ReasonGoto, "label on a statement that is not a loop or switch")
		}
	case *ast.EmptyStmt:
	default:
		e.fail(s, ReasonUnknownType, fmt.Sprintf("statement %T", s))
	}
}

func (e *emitter) varSpec(spec *ast.ValueSpec) {
	if len(spec.Values) == 1 && len(spec.Names) > 1 {
		e.multiAssign(identsToExprs(spec.Names), spec.Values[0], true, spec)
		return
	}
	for i, id := range spec.Names {
		obj := e.p.info.Defs[id]
		if obj == nil {
			continue
		}
		var init ast.Expr
		if i < len(spec.Values) {
			init = spec.Values[i]
		}
		if id.Name == "_" {
			if init != nil {
				v := e.expr(init)
				if v.code != "" {
					e.line("void %s;", v.code)
				}
			}
			continue
		}
		k := e.p.kindOf(obj.Type())
		if erasedKind(k) {
			code := ""
			if k == kState {
				code = "state"
			}
			e.erase(obj, code)
			continue
		}
		if init == nil {
			zero := e.p.zeroValue(obj.Type())
			ts := e.p.tsType(obj.Type())
			if k == kStack {
				l := e.declare(obj, id)
				l.nullable = true
				e.line("let %s: %s | null = null;", l.name, e.typeRef("StackRef"))
				continue
			}
			if zero == "" || ts == "" {
				e.fail(id, ReasonUnknownType, "var without a value of type "+obj.Type().String())
			}
			l := e.declare(obj, id)
			e.line("let %s: %s = %s;", l.name, e.typeRef(ts), zero)
			continue
		}
		v := e.exprAs(init, obj.Type())
		e.declareWith(obj, id, v)
	}
}

func identsToExprs(ids []*ast.Ident) []ast.Expr {
	out := make([]ast.Expr, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}

// declareWith declares a new local initialized to v.
func (e *emitter) declareWith(obj types.Object, id *ast.Ident, v value) {
	k := e.p.kindOf(obj.Type())
	if erasedKind(k) {
		e.erase(obj, v.code)
		return
	}
	ts := e.p.tsType(obj.Type())
	if ts == "" {
		e.fail(id, ReasonUnknownType, "local of type "+obj.Type().String())
	}
	l := e.declare(obj, id)
	if k == kStack && e.assignedNil(obj) {
		l.nullable = true
		ts = "StackRef | null"
	}
	e.line("let %s: %s = %s;", l.name, e.typeRef(ts), v.code)
}

// assignedNil reports whether the function ever assigns nil to obj.
func (e *emitter) assignedNil(obj types.Object) bool {
	found := false
	ast.Inspect(e.body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return !found
		}
		for i, lhs := range assign.Lhs {
			id, ok := lhs.(*ast.Ident)
			if !ok || e.p.info.ObjectOf(id) != obj || i >= len(assign.Rhs) {
				continue
			}
			if tv, ok := e.p.info.Types[assign.Rhs[i]]; ok && tv.IsNil() {
				found = true
			}
		}
		return !found
	})
	return found
}

func (e *emitter) ifStmt(s *ast.IfStmt, elseIf bool) {
	if s.Init != nil {
		if elseIf {
			e.line("} else {")
		} else {
			e.line("{")
		}
		e.indent++
		e.stmt(s.Init)
		e.ifStmt(&ast.IfStmt{If: s.If, Cond: s.Cond, Body: s.Body, Else: s.Else}, false)
		e.indent--
		e.line("}")
		return
	}
	cond := e.expr(s.Cond)
	if elseIf {
		e.line("} else if (%s) {", cond.code)
	} else {
		e.line("if (%s) {", cond.code)
	}
	e.block(s.Body.List)
	switch els := s.Else.(type) {
	case nil:
		e.line("}")
	case *ast.IfStmt:
		if els.Init != nil {
			e.line("} else {")
			e.indent++
			e.ifStmt(els, false)
			e.indent--
			e.line("}")
			return
		}
		e.ifStmt(els, true)
	case *ast.BlockStmt:
		e.line("} else {")
		e.block(els.List)
		e.line("}")
	}
}

func labelPrefix(label string) string {
	if label == "" {
		return ""
	}
	return label + ": "
}

func (e *emitter) forStmt(s *ast.ForStmt, label string) {
	opened := false
	init := ""
	if s.Init != nil {
		// A simple `i := x` (or several) goes in the head, with JavaScript's
		// per-iteration `let`, which is Go 1.22's loop variable.
		if assign, ok := s.Init.(*ast.AssignStmt); ok && assign.Tok == token.DEFINE && len(assign.Lhs) == len(assign.Rhs) {
			var parts []string
			for i, lhs := range assign.Lhs {
				id := lhs.(*ast.Ident)
				obj := e.p.info.Defs[id]
				if obj == nil || id.Name == "_" || erasedKind(e.p.kindOf(obj.Type())) {
					e.fail(id, ReasonUnknownType, "loop variable")
				}
				v := e.exprAs(assign.Rhs[i], obj.Type())
				ts := e.p.tsType(obj.Type())
				l := e.declare(obj, id)
				parts = append(parts, l.name+": "+e.typeRef(ts)+" = "+v.code)
			}
			init = "let " + strings.Join(parts, ", ")
		} else {
			e.line("{")
			e.indent++
			opened = true
			e.stmt(s.Init)
		}
	}
	cond := ""
	if s.Cond != nil {
		cond = " " + e.expr(s.Cond).code
	}
	post := ""
	if s.Post != nil {
		post = " " + e.simpleStmt(s.Post)
	}
	e.line("%sfor (%s;%s;%s) {", labelPrefix(label), init, cond, post)
	e.block(s.Body.List)
	e.line("}")
	if opened {
		e.indent--
		e.line("}")
	}
}

// simpleStmt is a loop's post statement as one expression.
func (e *emitter) simpleStmt(s ast.Stmt) string {
	saved := e.out
	savedIndent := e.indent
	e.out = strings.Builder{}
	e.indent = 0
	e.stmt(s)
	text := strings.TrimSpace(e.out.String())
	e.out = saved
	e.indent = savedIndent
	if strings.Contains(text, "\n") || !strings.HasSuffix(text, ";") {
		e.fail(s, ReasonUnknownType, "loop post statement")
	}
	return strings.TrimSuffix(text, ";")
}

func (e *emitter) rangeStmt(s *ast.RangeStmt, label string) {
	xt := e.p.info.TypeOf(s.X)
	var keyObj, valObj types.Object
	var keyExpr, valExpr ast.Expr
	if s.Key != nil {
		keyExpr = s.Key
	}
	if s.Value != nil {
		valExpr = s.Value
	}
	if s.Tok == token.DEFINE {
		if id, ok := keyExpr.(*ast.Ident); ok && id.Name != "_" {
			keyObj = e.p.info.Defs[id]
		}
		if id, ok := valExpr.(*ast.Ident); ok && id.Name != "_" {
			valObj = e.p.info.Defs[id]
		}
	}
	index := e.temp("i")
	switch u := xt.Underlying().(type) {
	case *types.Basic:
		if u.Info()&types.IsInteger == 0 {
			e.fail(s, ReasonStringBytes, "range over "+xt.String())
		}
		n := e.expr(s.X)
		limit := e.temp("n")
		e.line("%sfor (let %s = 0, %s = %s; %s < %s; %s++) {", labelPrefix(label), index, limit, n.code, index, limit, index)
		e.indent++
		e.rangeVar(s, keyObj, keyExpr, index)
		e.indent--
	case *types.Slice:
		x := e.expr(s.X)
		slice := e.temp("s")
		e.line("{")
		e.indent++
		e.line("const %s = %s;", slice, x.code)
		e.line("%sfor (let %s = 0; %s < %s(%s); %s++) {", labelPrefix(label), index, index, e.use("lenOf"), slice, index)
		e.indent++
		e.rangeVar(s, keyObj, keyExpr, index)
		if valExpr != nil && !isBlank(valExpr) {
			elem := value{code: slice + "![" + index + "]", t: u.Elem(), k: e.p.kindOf(u.Elem())}
			if valObj != nil {
				e.declareWith(valObj, valExpr.(*ast.Ident), elem)
			} else if s.Tok == token.ASSIGN {
				e.assignTo(valExpr, elem, s)
			}
		}
		e.indent--
		defer func() {
			e.indent--
			e.line("}")
		}()
	default:
		e.fail(s, ReasonMapRange, "range over "+xt.String())
	}
	e.block(s.Body.List)
	e.line("}")
}

func isBlank(expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "_"
}

func (e *emitter) rangeVar(s *ast.RangeStmt, keyObj types.Object, keyExpr ast.Expr, index string) {
	if keyExpr == nil || isBlank(keyExpr) {
		return
	}
	v := value{code: index, t: types.Typ[types.Int]}
	if keyObj != nil {
		e.declareWith(keyObj, keyExpr.(*ast.Ident), v)
	} else if s.Tok == token.ASSIGN {
		e.assignTo(keyExpr, v, s)
	}
}

func (e *emitter) switchStmt(s *ast.SwitchStmt, label string) {
	opened := false
	if s.Init != nil {
		e.line("{")
		e.indent++
		opened = true
		e.stmt(s.Init)
	}
	tag := "true"
	var tagType types.Type
	if s.Tag != nil {
		v := e.expr(s.Tag)
		tag = v.code
		tagType = e.p.info.TypeOf(s.Tag)
		if tagType != nil && !(isInteger(tagType) || isString(tagType) || isBool(tagType)) {
			e.fail(s.Tag, ReasonUnknownType, "switch on "+tagType.String())
		}
	}
	e.line("%sswitch (%s) {", labelPrefix(label), tag)
	e.indent++
	for _, clause := range s.Body.List {
		cc := clause.(*ast.CaseClause)
		if cc.List == nil {
			e.line("default: {")
		} else {
			for i, expr := range cc.List {
				v := e.expr(expr)
				suffix := ":"
				if i == len(cc.List)-1 {
					suffix = ": {"
				}
				e.line("case %s%s", v.code, suffix)
			}
		}
		e.block(cc.Body)
		if !terminates(cc.Body) {
			e.indent++
			e.line("break;")
			e.indent--
		}
		e.line("}")
	}
	e.indent--
	e.line("}")
	if opened {
		e.indent--
		e.line("}")
	}
}

func terminates(list []ast.Stmt) bool {
	if len(list) == 0 {
		return false
	}
	switch last := list[len(list)-1].(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.BranchStmt:
		return last.Tok == token.BREAK || last.Tok == token.CONTINUE
	}
	return false
}

func (e *emitter) returnStmt(s *ast.ReturnStmt) {
	results := s.Results
	if len(results) == 0 && len(e.namedResult) > 0 {
		for _, l := range e.namedResult {
			results = append(results, &ast.Ident{Name: l.name})
		}
		var names []string
		for _, l := range e.namedResult {
			names = append(names, l.name)
		}
		if len(names) == 1 {
			e.line("return %s;", names[0])
		} else {
			e.line("return [%s];", strings.Join(names, ", "))
		}
		return
	}
	switch {
	case len(results) == 0:
		e.line("return;")
		return
	case e.role == gamedef.RoleDistribute:
		e.distributeReturn(s)
		return
	case e.role == gamedef.RoleLegalCustom && len(results) == 1:
		if call, ok := ast.Unparen(results[0]).(*ast.CallExpr); ok {
			if fn := e.p.callee(call); fn != nil && fn.Pkg() != nil && fn.Pkg().Path() == boardgamePath+"/legal" && fn.Name() == "Errorf" {
				e.line("return %s;", e.legalErrorf(call).code)
				return
			}
		}
	}
	if len(results) == 1 {
		want := e.results
		if want.Len() > 1 {
			// return f() where f has the same results.
			v := e.expr(results[0])
			e.line("return %s;", v.code)
			return
		}
		v := e.exprAs(results[0], want.At(0).Type())
		e.line("return %s;", v.code)
		return
	}
	var parts []string
	for i, r := range results {
		parts = append(parts, e.exprAs(r, e.results.At(i).Type()).code)
	}
	e.line("return [%s];", strings.Join(parts, ", "))
}

// distributeReturn: Go's (stack, error); the convention returns the stack
// and throws the error.
func (e *emitter) distributeReturn(s *ast.ReturnStmt) {
	if len(s.Results) != 2 {
		e.fail(s, ReasonUnknownType, "distribute return of one value")
	}
	errExpr := s.Results[1]
	if tv, ok := e.p.info.Types[errExpr]; ok && tv.IsNil() {
		stack := e.expr(s.Results[0])
		if tv0, ok := e.p.info.Types[s.Results[0]]; ok && tv0.IsNil() {
			e.line("return %s(null);", e.use("nonNil"))
			return
		}
		e.line("return %s;", e.nonNullStack(stack))
		return
	}
	if tv0, ok := e.p.info.Types[s.Results[0]]; ok && tv0.IsNil() {
		err := e.expr(errExpr)
		e.line("throw new Error(%s(%s));", e.use("nonNil"), err.code)
		return
	}
	stack := e.expr(s.Results[0])
	err := e.expr(errExpr)
	r, x := e.temp("r"), e.temp("err")
	e.line("{")
	e.indent++
	e.line("const %s = %s;", r, stack.code)
	e.line("const %s = %s;", x, err.code)
	e.line("if (%s !== null) throw new Error(%s);", x, x)
	e.line("return %s(%s);", e.use("nonNil"), r)
	e.indent--
	e.line("}")
}

func (e *emitter) nonNullStack(v value) string {
	if v.nullable {
		return e.use("nonNil") + "(" + v.code + ")"
	}
	return v.code
}

// ----- assignment -----------------------------------------------------------

func (e *emitter) assign(s *ast.AssignStmt) {
	switch s.Tok {
	case token.ASSIGN, token.DEFINE:
	default:
		if len(s.Lhs) != 1 || len(s.Rhs) != 1 {
			e.fail(s, ReasonUnknownType, "compound tuple assignment")
		}
		op := strings.TrimSuffix(s.Tok.String(), "=")
		e.compound(s.Lhs[0], op, e.expr(s.Rhs[0]), s)
		return
	}
	define := s.Tok == token.DEFINE
	if len(s.Rhs) == 1 && len(s.Lhs) > 1 {
		e.multiAssign(s.Lhs, s.Rhs[0], define, s)
		return
	}
	if len(s.Lhs) != len(s.Rhs) {
		e.fail(s, ReasonUnknownType, "assignment count")
	}
	if len(s.Lhs) == 1 {
		e.single(s.Lhs[0], s.Rhs[0], define, s)
		return
	}
	// a, b = x, y: every right side first, as Go does.
	var temps []string
	for i, rhs := range s.Rhs {
		v := e.exprAs(rhs, e.p.info.TypeOf(s.Lhs[i]))
		t := e.temp("t")
		e.line("const %s = %s;", t, v.code)
		temps = append(temps, t)
	}
	for i, lhs := range s.Lhs {
		v := value{code: temps[i], t: e.p.info.TypeOf(s.Rhs[i]), k: e.p.kindOf(e.p.info.TypeOf(s.Rhs[i]))}
		e.bind(lhs, v, define, s)
	}
}

func (e *emitter) single(lhs, rhs ast.Expr, define bool, s ast.Node) {
	if isBlank(lhs) {
		v := e.expr(rhs)
		if v.code != "" {
			e.line("void %s;", v.code)
		}
		return
	}
	// x = append(x, ...)
	if call, ok := ast.Unparen(rhs).(*ast.CallExpr); ok && e.isBuiltin(call, "append") {
		target := e.lvalue(lhs)
		args := []string{e.readLvalue(target, lhs)}
		for _, a := range call.Args[1:] {
			args = append(args, e.exprAs(a, e.sliceElem(e.p.info.TypeOf(lhs))).code)
		}
		if call.Ellipsis.IsValid() {
			e.fail(call, ReasonSliceAlias, "append with ...")
		}
		e.writeLvalue(target, e.use("appendTo")+"("+strings.Join(args, ", ")+")")
		return
	}
	if e.erasedCall(rhs) {
		e.fail(rhs, ReasonUnknownType, "single value from an erased function")
	}
	t := e.p.info.TypeOf(lhs)
	if define {
		if id, ok := lhs.(*ast.Ident); ok {
			if obj := e.p.info.Defs[id]; obj != nil {
				e.declareWith(obj, id, e.exprAs(rhs, obj.Type()))
				return
			}
		}
	}
	v := e.exprAs(rhs, t)
	e.assignTo(lhs, v, s)
}

func (e *emitter) bind(lhs ast.Expr, v value, define bool, s ast.Node) {
	if isBlank(lhs) {
		return
	}
	if define {
		if id, ok := lhs.(*ast.Ident); ok {
			if obj := e.p.info.Defs[id]; obj != nil {
				e.declareWith(obj, id, v)
				return
			}
		}
	}
	e.assignTo(lhs, v, s)
}

func (e *emitter) sliceElem(t types.Type) types.Type {
	if s, ok := t.Underlying().(*types.Slice); ok {
		return s.Elem()
	}
	return nil
}

// multiAssign is a, b := f() (or = f()), and v, ok := m[k].
func (e *emitter) multiAssign(lhs []ast.Expr, rhs ast.Expr, define bool, s ast.Node) {
	if e.erasedCall(rhs) {
		// game, players := concreteStates(state)
		call := ast.Unparen(rhs).(*ast.CallExpr)
		fn := e.p.callee(call)
		results := fn.Type().(*types.Signature).Results()
		for i, target := range lhs {
			if isBlank(target) {
				continue
			}
			rt := results.At(i).Type()
			switch {
			case e.p.kindOf(rt) == kGame:
				e.bindErased(target, define)
			case e.isPlayersSlice(rt):
				e.bind(target, value{code: e.use("playerIndexes") + "(state)", t: rt}, define, s)
			default:
				e.fail(target, ReasonUnknownType, "result of an erased function")
			}
		}
		return
	}
	if index, ok := ast.Unparen(rhs).(*ast.IndexExpr); ok {
		if m, ok := e.p.info.TypeOf(index.X).Underlying().(*types.Map); ok && len(lhs) == 2 {
			mv := e.expr(index.X)
			key := e.exprAs(index.Index, m.Key())
			k := e.temp("k")
			mapName := e.temp("m")
			e.line("const %s = %s;", mapName, mv.code)
			e.line("const %s = %s;", k, key.code)
			zero := e.p.zeroValue(m.Elem())
			e.bind(lhs[0], value{code: "(" + mapName + ".get(" + k + ") ?? " + zero + ")", t: m.Elem()}, define, s)
			e.bind(lhs[1], value{code: mapName + ".has(" + k + ")", t: types.Typ[types.Bool]}, define, s)
			return
		}
	}
	v := e.expr(rhs)
	tuple, ok := e.p.info.TypeOf(rhs).(*types.Tuple)
	if !ok || tuple.Len() != len(lhs) {
		e.fail(rhs, ReasonUnknownType, "multiple values")
	}
	t := e.temp("r")
	e.line("const %s = %s;", t, v.code)
	for i, target := range lhs {
		et := tuple.At(i).Type()
		e.bind(target, value{code: t + "[" + strconv.Itoa(i) + "]", t: et, k: e.p.kindOf(et)}, define, s)
	}
}

func (e *emitter) bindErased(target ast.Expr, define bool) {
	id, ok := target.(*ast.Ident)
	if !ok {
		e.fail(target, ReasonUnknownType, "erased value assigned to a non-variable")
	}
	obj := e.p.info.ObjectOf(id)
	if obj == nil {
		return
	}
	if !define {
		if l := e.locals[obj]; l != nil && l.erased {
			return
		}
	}
	e.erase(obj, "")
}

func (e *emitter) isPlayersSlice(t types.Type) bool {
	s, ok := t.Underlying().(*types.Slice)
	return ok && e.p.kindOf(s.Elem()) == kPlayer
}

// erasedCall reports whether expr calls an erased function (concreteStates).
func (e *emitter) erasedCall(expr ast.Expr) bool {
	call, ok := ast.Unparen(expr).(*ast.CallExpr)
	if !ok {
		return false
	}
	fn := e.p.callee(call)
	return fn != nil && e.p.erased[fn]
}

// erasedValue reports whether expr is a value with no TypeScript value
// (state.ImmutableGameState().(*gameState)).
func (e *emitter) erasedValue(expr ast.Expr) bool {
	t := e.p.info.TypeOf(expr)
	return t != nil && (e.p.kindOf(t) == kGame || e.p.kindOf(t) == kDelegate)
}

// lv is an assignment target.
type lv struct {
	kind   string // "local", "prop", "index", "map"
	code   string // the local's name, or the property expression
	cast   string // the TypeScript type a read casts to
	object string // index/map: the slice or map
	key    string // index/map: the index or key
	zero   string // map: the zero value
}

func (e *emitter) lvalue(expr ast.Expr) lv {
	switch x := ast.Unparen(expr).(type) {
	case *ast.Ident:
		obj := e.p.info.ObjectOf(x)
		l := e.locals[obj]
		if l == nil {
			if v, ok := obj.(*types.Var); ok && v.Parent() == e.p.tpkg.Scope() {
				e.fail(x, ReasonPackageVarWrite, x.Name)
			}
			e.fail(x, ReasonUnknownType, "assignment to "+x.Name)
		}
		if l.erased {
			e.fail(x, ReasonUnknownType, "assignment to an erased value")
		}
		return lv{kind: "local", code: l.name}
	case *ast.SelectorExpr:
		target := e.selector(x, true)
		if target.lvalue == "" {
			e.fail(x, ReasonUnknownType, "assignment to "+types.ExprString(x))
		}
		return lv{kind: "prop", code: target.lvalue, cast: target.cast}
	case *ast.IndexExpr:
		xt := e.p.info.TypeOf(x.X)
		switch u := xt.Underlying().(type) {
		case *types.Slice:
			obj := e.expr(x.X)
			index := e.exprAs(x.Index, types.Typ[types.Int])
			return lv{kind: "index", object: obj.code, key: index.code}
		case *types.Map:
			obj := e.expr(x.X)
			key := e.exprAs(x.Index, u.Key())
			return lv{kind: "map", object: obj.code, key: key.code, zero: e.p.zeroValue(u.Elem())}
		}
	}
	e.fail(expr, ReasonUnknownType, "assignment target "+types.ExprString(expr))
	return lv{}
}

func (e *emitter) readLvalue(target lv, expr ast.Expr) string {
	switch target.kind {
	case "local":
		return target.code
	case "prop":
		return "(" + target.code + " as " + target.cast + ")"
	case "index":
		return e.use("at") + "(" + target.object + ", " + target.key + ")"
	case "map":
		return "(" + target.object + ".get(" + target.key + ") ?? " + target.zero + ")"
	}
	e.fail(expr, ReasonUnknownType, "read of an assignment target")
	return ""
}

func (e *emitter) writeLvalue(target lv, code string) {
	switch target.kind {
	case "local", "prop":
		e.line("%s = %s;", target.code, code)
	case "index":
		e.line("%s(%s, %s, %s);", e.use("setAt"), target.object, target.key, code)
	case "map":
		e.line("%s.set(%s, %s);", target.object, target.key, code)
	}
}

// simpleKey reports whether code is a name or a literal, safe to evaluate
// twice.
func simpleKey(code string) bool {
	for _, r := range code {
		if !(r == '_' || r == '$' || r == '-' || r == '.' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return code != ""
}

func (e *emitter) assignTo(target ast.Expr, v value, s ast.Node) {
	if isBlank(target) {
		return
	}
	if id, ok := ast.Unparen(target).(*ast.Ident); ok {
		if l := e.locals[e.p.info.ObjectOf(id)]; l != nil && l.erased {
			if erasedKind(e.p.kindOf(v.t)) || v.code == "" {
				return
			}
		}
	}
	t := e.lvalue(target)
	e.writeLvalue(t, v.code)
}

// compound is x op= y and x++ / x--.
func (e *emitter) compound(target ast.Expr, op string, rhs value, s ast.Node) {
	t := e.p.info.TypeOf(target)
	lt := e.lvalue(target)
	if (lt.kind == "index" || lt.kind == "map") && (!simpleKey(lt.key) || !simpleKey(lt.object)) {
		obj, key := e.temp("o"), e.temp("k")
		e.line("const %s = %s;", obj, lt.object)
		e.line("const %s = %s;", key, lt.key)
		lt.object, lt.key = obj, key
	}
	current := value{code: e.readLvalue(lt, target), t: t}
	result := e.arith(op, current, rhs, t, s)
	if lt.kind == "local" && (op == "+" || op == "-") && rhs.code == "1" {
		e.line("%s%s%s;", lt.code, op, op)
		return
	}
	e.writeLvalue(lt, result.code)
}
