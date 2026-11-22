package transpile

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamedef"
	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile/subset"
)

// value is one translated expression.
type value struct {
	code string
	k    kind
	t    types.Type
	// compound is set for an operator expression, which needs parentheses
	// inside another.
	compound bool
	// nullable marks a stack read from a local that may hold nil.
	nullable bool
	// owner and prop name an enum property (kEnumVal).
	owner, prop string
	// inner is, for a values or dynamic handle made by valuesOf/dynamicOf,
	// the component expression it checked, which values()/dynamic() take
	// directly.
	inner string
	// lvalue and cast are, for a property read, the assignable expression
	// and the type a read casts to.
	lvalue, cast string
}

func (e *emitter) paren(v value) string {
	if v.compound {
		return "(" + v.code + ")"
	}
	return v.code
}

// exprAs is expr where a value of type want is expected (the context's
// type decides the representation of an untyped nil).
func (e *emitter) exprAs(expr ast.Expr, want types.Type) value {
	v := e.expr(expr)
	if want != nil && e.p.kindOf(want) == kStack && v.nullable {
		// Assigning a possibly-nil stack keeps it nullable; the local's
		// declaration decides.
		return v
	}
	return v
}

func (e *emitter) expr(expr ast.Expr) value {
	v := e.exprRaw(expr)
	if v.nullable && v.k == kStack {
		return value{code: e.use("nonNil") + "(" + v.code + ")", k: kStack, t: v.t}
	}
	return v
}

// exprRaw is expr without the nil check of a nullable stack local.
func (e *emitter) exprRaw(expr ast.Expr) value {
	info := e.p.info
	t := info.TypeOf(expr)
	if tv, ok := info.Types[expr]; ok && tv.Value != nil {
		return e.constant(expr, tv)
	}
	if tv, ok := info.Types[expr]; ok && tv.IsNil() {
		return value{code: "null", t: t}
	}
	switch x := expr.(type) {
	case *ast.ParenExpr:
		return e.exprRaw(x.X)
	case *ast.Ident:
		return e.ident(x)
	case *ast.BinaryExpr:
		return e.binary(x)
	case *ast.UnaryExpr:
		v := e.expr(x.X)
		switch x.Op {
		case token.NOT:
			return value{code: "!" + e.paren(v), t: t, compound: true}
		case token.SUB:
			return value{code: "-" + e.paren(v), t: t, compound: true}
		case token.ADD:
			return v
		}
		e.fail(x, ReasonUnknownType, "operator "+x.Op.String())
	case *ast.CallExpr:
		return e.call(x, false)
	case *ast.SelectorExpr:
		return e.selector(x, false)
	case *ast.IndexExpr:
		return e.index(x)
	case *ast.TypeAssertExpr:
		return e.assertion(x)
	case *ast.CompositeLit:
		return e.composite(x)
	case *ast.StarExpr:
		e.fail(x, ReasonPointerLocal, types.ExprString(x))
	case *ast.SliceExpr:
		e.fail(x, ReasonSliceAlias, "slice expression")
	case *ast.FuncLit:
		e.fail(x, ReasonClosureEscape, "function literal")
	}
	e.fail(expr, ReasonUnknownType, "expression "+types.ExprString(expr))
	return value{}
}

// constant folds a constant expression to its literal (go/types already
// evaluated iota, DIM*DIM and the rest).
func (e *emitter) constant(expr ast.Expr, tv types.TypeAndValue) value {
	t := tv.Type
	val := tv.Value
	switch val.Kind() {
	case constant.Bool:
		return value{code: strconv.FormatBool(constant.BoolVal(val)), t: t}
	case constant.String:
		return value{code: jsString(constant.StringVal(val)), t: t}
	case constant.Int:
		if isDuration(t) {
			ns, _ := new(big.Int).SetString(val.ExactString(), 10)
			ms, rem := new(big.Int).QuoRem(ns, big.NewInt(1000000), new(big.Int))
			if rem.Sign() != 0 {
				e.fail(expr, ReasonSubMsDuration, val.ExactString()+"ns")
			}
			return intLiteral(ms.String(), t)
		}
		if isFloat(t) {
			f, _ := constant.Float64Val(val)
			return floatLiteral(f, t)
		}
		return intLiteral(val.ExactString(), t)
	case constant.Float:
		if isInteger(t) {
			return intLiteral(constant.ToInt(val).ExactString(), t)
		}
		f, _ := constant.Float64Val(val)
		return floatLiteral(f, t)
	}
	e.fail(expr, ReasonUnknownType, "constant "+val.String())
	return value{}
}

func intLiteral(text string, t types.Type) value {
	if strings.HasPrefix(text, "-") {
		return value{code: "(" + text + ")", t: t}
	}
	return value{code: text, t: t}
}

func floatLiteral(f float64, t types.Type) value {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return value{code: "NaN", t: t}
	}
	text := strconv.FormatFloat(f, 'g', -1, 64)
	if f < 0 {
		return value{code: "(" + text + ")", t: t}
	}
	return value{code: text, t: t}
}

// jsString is a TypeScript string literal with Go's string's code points.
func jsString(s string) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	out := strings.TrimSuffix(buf.String(), "\n")
	// Single quotes, the engine's house style.
	inner := out[1 : len(out)-1]
	inner = strings.ReplaceAll(inner, `\"`, `"`)
	inner = strings.ReplaceAll(inner, `'`, `\'`)
	return "'" + inner + "'"
}

func (e *emitter) ident(id *ast.Ident) value {
	obj := e.p.info.ObjectOf(id)
	t := e.p.info.TypeOf(id)
	if l := e.locals[obj]; l != nil {
		if l.erased {
			return value{code: l.code, k: l.k, t: l.t}
		}
		return value{code: l.name, k: l.k, t: l.t, nullable: l.nullable}
	}
	switch obj := obj.(type) {
	case *types.Var:
		if obj.Parent() == e.p.tpkg.Scope() {
			// The package-variable rule: a framework object built from
			// constants is looked up by name or rebuilt; a never-written
			// constant-initialized variable is its value.
			if v, ok := e.p.frameworkValue(obj); ok {
				return v
			}
			if v, ok := e.constantVarValue(obj, id); ok {
				return v
			}
			e.fail(id, ReasonPackageVarRead, id.Name)
		}
	case *types.Nil:
		return value{code: "null", t: t}
	}
	e.fail(id, ReasonUnknownType, "identifier "+id.Name)
	return value{}
}

func (e *emitter) binary(x *ast.BinaryExpr) value {
	t := e.p.info.TypeOf(x)
	switch x.Op {
	case token.LAND, token.LOR:
		l, r := e.expr(x.X), e.expr(x.Y)
		return value{code: e.paren(l) + " " + x.Op.String() + " " + e.paren(r), t: t, compound: true}
	case token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ:
		return e.comparison(x)
	case token.ADD, token.SUB, token.MUL, token.QUO, token.REM:
		return e.arith(x.Op.String(), e.expr(x.X), e.expr(x.Y), t, x)
	}
	e.fail(x, ReasonUnknownType, "operator "+x.Op.String())
	return value{}
}

// arith is Go integer and float arithmetic and string concatenation.
func (e *emitter) arith(op string, l, r value, t types.Type, at ast.Node) value {
	if t == nil {
		t = l.t
	}
	switch {
	case isString(t):
		if op != "+" {
			e.fail(at, ReasonUnknownType, "string operator "+op)
		}
		return value{code: e.paren(l) + " + " + e.paren(r), t: t, compound: true}
	case isInteger(t):
		switch op {
		case "+", "-":
			return value{code: e.paren(l) + " " + op + " " + e.paren(r), t: t, compound: true}
		case "*":
			return value{code: e.use("imul") + "(" + l.code + ", " + r.code + ")", t: t}
		case "/":
			return value{code: e.use("idiv") + "(" + l.code + ", " + r.code + ")", t: t}
		case "%":
			return value{code: e.use("imod") + "(" + l.code + ", " + r.code + ")", t: t}
		}
	case isFloat(t):
		switch op {
		case "+", "-", "*", "/":
			return value{code: e.paren(l) + " " + op + " " + e.paren(r), t: t, compound: true}
		}
	}
	e.fail(at, ReasonUnknownType, "operator "+op+" on "+typeString(t))
	return value{}
}

func typeString(t types.Type) string {
	if t == nil {
		return "?"
	}
	return t.String()
}

func (e *emitter) comparison(x *ast.BinaryExpr) value {
	info := e.p.info
	lt, rt := info.TypeOf(x.X), info.TypeOf(x.Y)
	lnil := info.Types[x.X].IsNil()
	rnil := info.Types[x.Y].IsNil()
	t := info.TypeOf(x)
	op := x.Op.String()
	if lnil || rnil {
		if x.Op != token.EQL && x.Op != token.NEQ {
			e.fail(x, ReasonUnknownType, "ordered comparison with nil")
		}
		other := x.X
		if lnil {
			other = x.Y
		}
		v := e.exprRaw(other)
		if v.code == "" {
			e.fail(x, ReasonUnknownType, "nil comparison of an erased value")
		}
		if v.k == kValues && v.inner != "" {
			// c.Values() == nil: a generic component with no values.
			not := ""
			if x.Op == token.NEQ {
				not = "!"
			}
			return value{code: not + e.use("valuesAreNil") + "(state, " + v.inner + ")", t: t, compound: true}
		}
		if v.k == kValues || v.k == kDynamic || v.k == kPlayer || v.k == kComponent {
			e.fail(x, ReasonUnknownType, "nil comparison of a handle that is never nil")
		}
		jsop := "==="
		if x.Op == token.NEQ {
			jsop = "!=="
		}
		return value{code: e.paren(v) + " " + jsop + " null", t: t, compound: true}
	}
	l, r := e.expr(x.X), e.expr(x.Y)
	operand := lt
	if operand == nil || isUntyped(operand) {
		operand = rt
	}
	scalar := func(t types.Type) bool {
		return t != nil && (isInteger(t) || isString(t) || isBool(t) || isFloat(t))
	}
	if !scalar(lt) || !scalar(rt) {
		e.fail(x, ReasonUnknownType, "comparison of "+typeString(lt))
	}
	if isString(operand) && x.Op != token.EQL && x.Op != token.NEQ {
		return value{code: e.use("strCompare") + "(" + l.code + ", " + r.code + ") " + op + " 0", t: t, compound: true}
	}
	switch x.Op {
	case token.EQL:
		op = "==="
	case token.NEQ:
		op = "!=="
	}
	return value{code: e.paren(l) + " " + op + " " + e.paren(r), t: t, compound: true}
}

func isUntyped(t types.Type) bool {
	b, ok := t.(*types.Basic)
	return ok && b.Info()&types.IsUntyped != 0
}

// ----- selectors ------------------------------------------------------------

// selector is x.F: a property of the game, a player, a move or a
// component's values. A method value (x.M without a call) is refused.
func (e *emitter) selector(sel *ast.SelectorExpr, write bool) value {
	info := e.p.info
	if _, isPkg := info.Uses[identOf(sel.X)].(*types.PkgName); isPkg {
		obj := info.Uses[sel.Sel]
		if _, ok := obj.(*types.Var); ok {
			e.fail(sel, ReasonPackageVarRead, types.ExprString(sel))
		}
		e.fail(sel, ReasonUnknownType, "package member "+types.ExprString(sel))
	}
	selection := info.Selections[sel]
	if selection == nil || selection.Kind() != types.FieldVal {
		e.fail(sel, ReasonClosureEscape, "method value "+types.ExprString(sel))
	}
	field := selection.Obj().(*types.Var)
	recv := e.expr(sel.X)
	return e.property(recv, field, sel)
}

func identOf(expr ast.Expr) *ast.Ident {
	id, _ := ast.Unparen(expr).(*ast.Ident)
	return id
}

func (e *emitter) property(recv value, field *types.Var, at ast.Node) value {
	prop := field.Name()
	ft := field.Type()
	fk := e.p.kindOf(ft)
	quoted := jsString(prop)
	cast := e.p.tsType(ft)
	switch recv.k {
	case kGame:
		switch fk {
		case kStack:
			return value{code: e.use("gameStack") + "(state, " + quoted + ")", k: kStack, t: ft}
		case kTimer:
			return value{code: "state.timer(" + quoted + ")", k: kTimer, t: ft}
		case kEnumVal:
			return value{k: kEnumVal, t: ft, owner: "'game'", prop: prop}
		case kPlain:
			return e.scalarProp("state.g."+prop, cast, ft, at)
		}
	case kPlayer:
		switch fk {
		case kStack:
			return value{code: e.use("playerStack") + "(state, " + recv.code + ", " + quoted + ")", k: kStack, t: ft}
		case kEnumVal:
			return value{k: kEnumVal, t: ft, owner: recv.code, prop: prop}
		case kPlain:
			return e.scalarProp(e.use("player")+"(state, "+recv.code+")."+prop, cast, ft, at)
		}
	case kMove:
		switch fk {
		case kEnumVal:
			return value{k: kEnumVal, t: ft, owner: "{ fieldsOf: " + recv.code + " }", prop: prop}
		case kPlain:
			return e.scalarProp(recv.code+".fields."+prop, cast, ft, at)
		}
	case kValues:
		inner := recv.inner
		if inner == "" {
			inner = recv.code
		}
		switch fk {
		case kEnumVal:
			return value{k: kEnumVal, t: ft, owner: "{ valuesOf: " + inner + " }", prop: prop}
		case kPlain:
			v := e.scalarProp(e.use("values")+"(state, "+inner+")."+prop, cast, ft, at)
			v.lvalue = ""
			return v
		}
	case kDynamic:
		inner := recv.inner
		if inner == "" {
			inner = recv.code
		}
		switch fk {
		case kEnumVal:
			return value{k: kEnumVal, t: ft, owner: "{ dynamicOf: " + inner + " }", prop: prop}
		case kPlain:
			return e.scalarProp(e.use("dynamic")+"(state, "+inner+")."+prop, cast, ft, at)
		}
	}
	e.fail(at, ReasonUnknownType, "property "+prop+" of "+typeString(recv.t))
	return value{}
}

func (e *emitter) scalarProp(path, cast string, t types.Type, at ast.Node) value {
	if cast == "" {
		e.fail(at, ReasonUnknownType, "property of type "+t.String())
	}
	if _, isStruct := t.Underlying().(*types.Struct); isStruct {
		e.fail(at, ReasonStructCopy, t.String())
	}
	return value{code: "(" + path + " as " + e.typeRef(cast) + ")", t: t, lvalue: path, cast: cast}
}

// ----- indexing -------------------------------------------------------------

func (e *emitter) index(x *ast.IndexExpr) value {
	t := e.p.info.TypeOf(x)
	xt := e.p.info.TypeOf(x.X)
	// state.Game().Variant()[k], or the variant parameter's [k].
	if e.p.kindOf(xt) == kVariant {
		key := e.expr(x.Index)
		return value{code: e.use("variantValue") + "(state, " + key.code + ")", t: t}
	}
	switch u := xt.Underlying().(type) {
	case *types.Slice:
		obj := e.expr(x.X)
		i := e.expr(x.Index)
		return value{code: e.use("at") + "(" + obj.code + ", " + i.code + ")", t: t, k: e.p.kindOf(u.Elem())}
	case *types.Map:
		obj := e.expr(x.X)
		key := e.expr(x.Index)
		zero := e.p.zeroValue(u.Elem())
		if zero == "" {
			e.fail(x, ReasonUnknownType, "map of "+u.Elem().String())
		}
		return value{code: "(" + obj.code + ".get(" + key.code + ") ?? " + zero + ")", t: t}
	}
	e.fail(x, ReasonUnknownType, "index of "+typeString(xt))
	return value{}
}

// assertion is x.(*T): the game's own state, move and component types are
// erased; the value keeps its representation.
func (e *emitter) assertion(x *ast.TypeAssertExpr) value {
	if x.Type == nil {
		e.fail(x, ReasonTypeSwitch, "type switch guard")
	}
	t := e.p.info.TypeOf(x.Type)
	v := e.expr(x.X)
	want := e.p.kindOf(t)
	switch {
	case v.k == want:
		v.t = t
		return v
	case want == kGame && (v.k == kGame || v.code == ""):
		return value{k: kGame, t: t}
	case want == kPlayer && v.k == kPlayer:
		return value{code: v.code, k: kPlayer, t: t}
	}
	e.fail(x, ReasonUnknownType, "type assertion to "+t.String()+" from "+typeString(v.t))
	return value{}
}

func (e *emitter) composite(x *ast.CompositeLit) value {
	t := e.p.info.TypeOf(x)
	s, ok := t.Underlying().(*types.Slice)
	if !ok {
		e.fail(x, ReasonStructCopy, "composite literal of "+t.String())
	}
	var parts []string
	for _, elt := range x.Elts {
		if _, ok := elt.(*ast.KeyValueExpr); ok {
			e.fail(elt, ReasonUnknownType, "indexed slice literal")
		}
		parts = append(parts, e.exprAs(elt, s.Elem()).code)
	}
	return value{code: "[" + strings.Join(parts, ", ") + "]", t: t}
}

// ----- calls ----------------------------------------------------------------

func (e *emitter) isBuiltin(call *ast.CallExpr, name string) bool {
	id, ok := ast.Unparen(call.Fun).(*ast.Ident)
	if !ok {
		return false
	}
	b, ok := e.p.info.Uses[id].(*types.Builtin)
	return ok && b.Name() == name
}

// stockReceiver reports whether expr selects an embedded stock move field
// (m.FixUp, m.CurrentPlayer).
func (e *emitter) stockReceiver(expr ast.Expr) bool {
	sel, ok := ast.Unparen(expr).(*ast.SelectorExpr)
	if !ok {
		return false
	}
	selection := e.p.info.Selections[sel]
	if selection == nil || selection.Kind() != types.FieldVal {
		return false
	}
	field := selection.Obj().(*types.Var)
	if !field.Embedded() {
		return false
	}
	named, _ := namedOf(field.Type())
	return named != nil && pkgPathOf(named) == movesPath
}

// call translates a call. statement is true for an expression statement,
// where a call with no TypeScript effect (a dropped erased call) may emit
// nothing.
func (e *emitter) call(call *ast.CallExpr, statement bool) value {
	info := e.p.info
	t := info.TypeOf(call)
	if tv, ok := info.Types[call.Fun]; ok && tv.IsType() {
		return e.conversion(call, tv.Type)
	}
	if id, ok := ast.Unparen(call.Fun).(*ast.Ident); ok {
		if b, ok := info.Uses[id].(*types.Builtin); ok {
			return e.builtin(call, b.Name())
		}
	}
	fn := e.p.callee(call)
	if fn == nil {
		e.fail(call, ReasonClosureEscape, "call of a function value")
	}
	if call.Ellipsis.IsValid() && subset.QualifiedName(fn) != "(enum.RangeEnum).RangeToValue" {
		e.fail(call, ReasonUnknownType, "call with ...")
	}
	if fn.Pkg() == e.p.tpkg {
		if e.p.erased[fn] {
			e.fail(call, ReasonUnknownType, "erased function used as a value")
		}
		return e.gameCall(call, fn, t)
	}
	sel, _ := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if fn.Pkg() != nil && fn.Pkg().Path() == movesPath && sel != nil && e.stockReceiver(sel.X) {
		return e.stockCall(call, sel, fn)
	}
	entry, ok := subset.Lookup(subset.QualifiedName(fn))
	if !ok {
		e.fail(call, ReasonUnknownCall, subset.QualifiedName(fn))
	}
	if entry.Kind == subset.KindDelegate {
		own := e.p.delegateMethod(fn.Name())
		if own == nil {
			e.fail(call, ReasonUnknownCall, "delegate method "+fn.Name()+" the game does not declare")
		}
		return e.inPackage(call, own, "", t)
	}
	return e.tableCall(call, sel, fn, t)
}

func (e *emitter) conversion(call *ast.CallExpr, target types.Type) value {
	if len(call.Args) != 1 {
		e.fail(call, ReasonUnknownType, "conversion")
	}
	v := e.expr(call.Args[0])
	source := e.p.info.TypeOf(call.Args[0])
	switch {
	case isInteger(target) && source != nil && isFloat(source):
		return value{code: "Math.trunc(" + v.code + ")", t: target}
	case (isInteger(target) || isFloat(target)) && source != nil && (isInteger(source) || isFloat(source)):
		v.t = target
		return v
	case isString(target) && source != nil && isString(source), isBool(target) && source != nil && isBool(source):
		v.t = target
		return v
	}
	if e.p.kindOf(target) == e.p.kindOf(source) && e.p.kindOf(target) != kPlain {
		v.t = target
		return v
	}
	e.fail(call, ReasonUnknownType, "conversion to "+target.String())
	return value{}
}

func (e *emitter) builtin(call *ast.CallExpr, name string) value {
	t := e.p.info.TypeOf(call)
	switch name {
	case "len":
		arg := call.Args[0]
		// len(state.PlayerStates())
		if inner, ok := ast.Unparen(arg).(*ast.CallExpr); ok {
			if fn := e.p.callee(inner); fn != nil {
				if entry, ok := subset.Lookup(subset.QualifiedName(fn)); ok && entry.Facade == "playerStates" {
					return value{code: "state.numPlayers()", t: t}
				}
			}
		}
		at := e.p.info.TypeOf(arg)
		v := e.expr(arg)
		switch at.Underlying().(type) {
		case *types.Slice:
			return value{code: e.use("lenOf") + "(" + v.code + ")", t: t}
		case *types.Basic:
			if isString(at) {
				return value{code: e.use("utf8Len") + "(" + v.code + ")", t: t}
			}
		}
		e.fail(call, ReasonUnknownCall, "len of "+at.String())
	case "make":
		mt := e.p.info.TypeOf(call.Args[0])
		switch u := mt.Underlying().(type) {
		case *types.Slice:
			if len(call.Args) != 2 {
				e.fail(call, ReasonUnknownCall, "make with a capacity")
			}
			zero := e.p.zeroValue(u.Elem())
			if zero == "" {
				e.fail(call, ReasonUnknownType, "make of "+u.Elem().String())
			}
			n := e.expr(call.Args[1])
			return value{code: e.use("makeSlice") + "(" + n.code + ", " + zero + ")", t: t}
		case *types.Map:
			ts := e.p.tsType(mt)
			if ts == "" {
				e.fail(call, ReasonUnknownType, "make of "+mt.String())
			}
			return value{code: "new " + e.typeRef(ts) + "()", t: t}
		}
	case "copy":
		dst, src := e.expr(call.Args[0]), e.expr(call.Args[1])
		return value{code: e.use("copyInto") + "(" + dst.code + ", " + src.code + ")", t: t}
	case "delete":
		m, k := e.expr(call.Args[0]), e.expr(call.Args[1])
		return value{code: m.code + ".delete(" + k.code + ")", t: t}
	case "min", "max":
		if !isInteger(t) {
			e.fail(call, ReasonUnknownType, name+" of "+t.String())
		}
		var args []string
		for _, a := range call.Args {
			args = append(args, e.expr(a).code)
		}
		return value{code: "Math." + name + "(" + strings.Join(args, ", ") + ")", t: t}
	case "append":
		e.fail(call, ReasonSliceAlias, "append not in x = append(x, ...) form")
	}
	e.fail(call, ReasonUnknownCall, name)
	return value{}
}

// gameCall is a call of one of the game's own functions or methods.
func (e *emitter) gameCall(call *ast.CallExpr, fn *types.Func, t types.Type) value {
	recv := ""
	if sig := fn.Type().(*types.Signature); sig.Recv() != nil {
		sel, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
		if !ok {
			e.fail(call, ReasonUnknownType, "method call without a receiver")
		}
		selection := e.p.info.Selections[sel]
		if selection == nil || selection.Kind() != types.MethodVal {
			e.fail(call, ReasonClosureEscape, "method expression")
		}
		r := e.expr(sel.X)
		if !erasedKind(e.p.kindOf(sig.Recv().Type())) {
			if r.code == "" {
				e.fail(call, ReasonUnknownType, "receiver without a value")
			}
			recv = r.code
		}
	}
	return e.inPackage(call, fn, recv, t)
}

func (e *emitter) inPackage(call *ast.CallExpr, fn *types.Func, recv string, t types.Type) value {
	ref := e.p.refs[fn]
	if ref == "" {
		e.fail(call, ReasonUnknownCall, "no declaration for "+fn.Name())
	}
	switch e.p.roles[ref] {
	case gamedef.RoleLegal, gamedef.RoleDistribute:
		e.fail(call, ReasonUnknownCall, "direct call of a "+e.p.roles[ref]+" body")
	}
	name, err := Mangle(ref)
	if err != nil {
		e.fail(call, ReasonUnknownType, err.Error())
	}
	args := []string{"state"}
	hasRecv, passed := e.p.passed(fn)
	if hasRecv {
		args = append(args, recv)
	}
	sig := fn.Type().(*types.Signature)
	for i, a := range call.Args {
		v := e.exprAs(a, sig.Params().At(i).Type())
		if !passed[i] {
			continue
		}
		if v.code == "" {
			e.fail(a, ReasonUnknownType, "argument without a value")
		}
		args = append(args, v.code)
	}
	return value{code: name + "(" + strings.Join(args, ", ") + ")", t: t, k: e.p.kindOf(t)}
}

// stockCall is m.<Stock>.Apply(state), .Legal(state, proposer).
func (e *emitter) stockCall(call *ast.CallExpr, sel *ast.SelectorExpr, fn *types.Func) value {
	root := sel.X
	for {
		inner, ok := ast.Unparen(root).(*ast.SelectorExpr)
		if !ok {
			break
		}
		root = inner.X
	}
	m := e.expr(root)
	if m.k != kMove {
		e.fail(call, ReasonUnknownCall, "stock call on something other than the move")
	}
	switch fn.Name() {
	case "Apply":
		return value{code: e.use("stockApply") + "(state, " + m.code + ")", t: e.p.info.TypeOf(call)}
	case "Legal":
		if e.role != gamedef.RoleLegal || len(call.Args) != 2 || e.p.info.ObjectOf(identOf(call.Args[1])) != e.proposer || e.proposer == nil {
			e.fail(call, ReasonUnknownCall, "stock Legal outside a Legal body, or with another proposer")
		}
		return value{code: "base()", t: e.p.info.TypeOf(call)}
	}
	e.fail(call, ReasonUnknownCall, "stock "+fn.Name())
	return value{}
}

// legalErrorf is legal.Errorf(key, bindings) in a LegalCustom return.
func (e *emitter) legalErrorf(call *ast.CallExpr) value {
	if len(call.Args) != 2 {
		e.fail(call, ReasonUnknownCall, "legal.Errorf arguments")
	}
	key := e.expr(call.Args[0])
	if tv, ok := e.p.info.Types[call.Args[1]]; ok && tv.IsNil() {
		return value{code: e.use("legalError") + "(" + key.code + ")"}
	}
	e.fail(call, ReasonUnknownCall, "legal.Errorf with bindings")
	return value{}
}

func (e *emitter) args(call *ast.CallExpr) []string {
	var out []string
	for _, a := range call.Args {
		out = append(out, e.expr(a).code)
	}
	return out
}

// tableCall is a call in the API table.
func (e *emitter) tableCall(call *ast.CallExpr, sel *ast.SelectorExpr, fn *types.Func, t types.Type) value {
	qualified := subset.QualifiedName(fn)
	k := e.p.kindOf(t)
	var recv value
	if fn.Type().(*types.Signature).Recv() != nil {
		if sel == nil {
			e.fail(call, ReasonUnknownCall, qualified)
		}
		recv = e.expr(sel.X)
	}
	ret := func(code string) value { return value{code: code, t: t, k: k} }
	stack := func() string {
		if recv.k != kStack {
			e.fail(call, ReasonUnknownType, qualified+" on "+typeString(recv.t))
		}
		return recv.code
	}
	comp := func() string {
		if recv.k != kComp {
			e.fail(call, ReasonUnknownType, qualified+" on "+typeString(recv.t))
		}
		return recv.code
	}
	facade := func(name string, first string) value {
		args := append([]string{first}, e.args(call)...)
		return ret(e.use(name) + "(" + strings.Join(args, ", ") + ")")
	}
	method := fn.Name()
	switch qualified {
	case "(boardgame.ImmutableState).ImmutablePlayerStates", "(boardgame.State).PlayerStates":
		return ret(e.use("playerIndexes") + "(state)")
	case "(boardgame.ImmutableState).CurrentPlayerIndex", "(boardgame.State).CurrentPlayer", "(boardgame.ImmutableState).ImmutableCurrentPlayer":
		return value{code: "state.currentPlayerIndex()", t: t, k: kPlayer}
	case "(boardgame.PlayerIndex).EnsureValid", "(boardgame.PlayerIndex).Next", "(boardgame.PlayerIndex).Previous",
		"(boardgame.PlayerIndex).Valid", "(boardgame.PlayerIndex).WithinBounds":
		name := strings.ToLower(method[:1]) + method[1:]
		return ret("state." + name + "(" + recv.code + ")")
	case "(boardgame.ImmutableState).Rand", "(boardgame.State).Rand":
		return value{code: "state.rand()", t: t, k: kRand}
	case "(*rand.Rand).Intn", "(*rand.Rand).Perm", "(*rand.Rand).Int63n", "(*rand.Rand).Float64":
		if recv.k != kRand {
			e.fail(call, ReasonForeignRand, qualified)
		}
		name := strings.ToLower(method[:1]) + method[1:]
		return ret(recv.code + "." + name + "(" + strings.Join(e.args(call), ", ") + ")")
	case "(boardgame.ImmutableState).Game":
		return value{k: kGameObject, t: t}
	case "(*boardgame.Game).Variant":
		return value{code: "state.variant()", k: kVariant, t: t}
	case "(*boardgame.Game).MoveByName":
		return ret("state.moveByName(" + e.args(call)[0] + ")")
	case "(*boardgame.Game).MoveByNameForState":
		return ret("state.newMove(" + e.expr(call.Args[0]).code + ")")
	case "(boardgame.ImmutableState).Manager", "(*boardgame.GameManager).Delegate":
		return value{k: kManager, t: t}
	case "(boardgame.ImmutableState).ImmutableGameState", "(boardgame.State).GameState":
		return value{k: kGame, t: t}
	case "(boardgame.ImmutableSubState).ImmutableState", "(boardgame.SubState).State", "(*base.SubState).State", "(*base.SubState).ImmutableState":
		return value{code: "state", k: kState, t: t}
	case "behaviors.PlayerIsInactive", "behaviors.PlayerIsEliminated":
		p := e.expr(call.Args[0])
		if p.k != kPlayer {
			e.fail(call, ReasonUnknownType, qualified+" of something other than a player")
		}
		return ret("state." + strings.ToLower(method[:1]) + method[1:] + "(" + p.code + ")")
	case "(*behaviors.PlayerElimination).SetEliminated", "(*behaviors.InactivePlayer).SetPlayerInactive", "(*behaviors.Seat).SetSeatFilled":
		if recv.k != kPlayer {
			e.fail(call, ReasonUnknownType, qualified+" on something other than a player")
		}
		return ret(e.use(strings.ToLower(method[:1])+method[1:]) + "(state, " + recv.code + ")")
	case "(*behaviors.PhaseBehavior).SetCurrentPhase":
		return ret("state.setPhase(" + e.expr(call.Args[0]).code + ")")
	case "(enum.ImmutableVal).Value":
		e.requireEnum(call, recv)
		return ret(e.use("enumKey") + "(state, " + recv.owner + ", " + jsString(recv.prop) + ")")
	case "(enum.Val).SetValue":
		e.requireEnum(call, recv)
		return ret(e.use("setEnumKey") + "(state, " + recv.owner + ", " + jsString(recv.prop) + ", " + e.expr(call.Args[0]).code + ")")
	case "(enum.ImmutableVal).Equals":
		e.requireEnum(call, recv)
		other := e.exprRaw(call.Args[0])
		e.requireEnum(call, other)
		return ret(e.use("enumEquals") + "(state, " + recv.owner + ", " + jsString(recv.prop) + ", " + other.owner + ", " + jsString(other.prop) + ")")
	case "(enum.EnumKey).Int":
		return ret(recv.code)
	case "(enum.ImmutableRangeVal).RangeValue", "(enum.RangeVal).RangeValue":
		e.requireEnum(call, recv)
		return ret(e.use("rangeValue") + "(state, " + recv.owner + ", " + jsString(recv.prop) + ")")
	case "(enum.RangeEnum).RangeToValue":
		if recv.k != kEnum {
			e.fail(call, ReasonUnknownType, qualified+" on something other than an enum variable")
		}
		indexes := "[" + strings.Join(e.args(call), ", ") + "]"
		if call.Ellipsis.IsValid() {
			indexes = "(" + e.expr(call.Args[0]).code + " ?? [])"
		}
		return ret(e.use("rangeToValue") + "(state, " + recv.code + ", " + indexes + ")")
	case "(enum.RangeEnum).ValueToRange":
		if recv.k != kEnum {
			e.fail(call, ReasonUnknownType, qualified+" on something other than an enum variable")
		}
		return ret(e.use("valueToRange") + "(state, " + recv.code + ", " + e.expr(call.Args[0]).code + ")")
	case "(graph.Graph).NeighborsByKey", "(*graph.EnumGraph).NeighborsByKey":
		if recv.k != kGraph {
			e.fail(call, ReasonUnknownType, qualified+" on something other than a graph variable")
		}
		return ret(e.use("neighborsByKey") + "(" + recv.code + ", " + e.expr(call.Args[0]).code + ")")
	case "(boardgame.ImmutableStack).Len", "(boardgame.ImmutableStack).NumComponents", "(boardgame.ImmutableStack).SlotsRemaining", "(boardgame.ImmutableStack).MaxSize":
		return ret(e.use("stackView") + "(" + stack() + ")." + strings.ToLower(method[:1]) + method[1:] + "()")
	case "(boardgame.Stack).ComponentAt", "(boardgame.ImmutableStack).ImmutableComponentAt":
		return facade("componentAt", stack())
	case "(boardgame.Stack).First", "(boardgame.ImmutableStack).ImmutableFirst":
		return facade("first", stack())
	case "(boardgame.Stack).Last", "(boardgame.ImmutableStack).ImmutableLast":
		return facade("last", stack())
	case "(boardgame.Stack).Components", "(boardgame.ImmutableStack).ImmutableComponents":
		return facade("components", stack())
	case "(boardgame.Stack).ComponentAtKey", "(boardgame.ImmutableStack).ImmutableComponentAtKey":
		return facade("componentAtKey", stack())
	case "(boardgame.Stack).SwapComponentsByKey", "(boardgame.Stack).MoveAllTo", "(boardgame.Stack).MoveCountTo",
		"(boardgame.Stack).MayMoveAllTo", "(boardgame.ImmutableStack).MayMoveAllTo", "(boardgame.Stack).MayMoveCountTo",
		"(boardgame.ImmutableStack).MayMoveCountTo", "(boardgame.Stack).Shuffle", "(boardgame.Stack).PublicShuffle",
		"(boardgame.Stack).SwapComponents", "(boardgame.SizedStack).SetSize", "(boardgame.Stack).SetSize",
		"(boardgame.SizedStack).ExpandSize", "(boardgame.Stack).ExpandSize", "(boardgame.SizedStack).ContractSize",
		"(boardgame.Stack).ContractSize", "(boardgame.SizedStack).SizeToFit", "(boardgame.Stack).SizeToFit":
		return facade(strings.ToLower(method[:1])+method[1:], stack())
	case "(boardgame.Component).Values":
		if recv.k != kComp && recv.k != kComponent {
			e.fail(call, ReasonUnknownType, qualified+" on "+typeString(recv.t))
		}
		return value{code: e.use("valuesOf") + "(state, " + recv.code + ")", k: kValues, t: t, inner: recv.code}
	case "(boardgame.ComponentInstance).DynamicValues", "(boardgame.ImmutableComponentInstance).ImmutableDynamicValues":
		if recv.k != kComp && recv.k != kComponent {
			e.fail(call, ReasonUnknownType, qualified+" on "+typeString(recv.t))
		}
		return value{code: e.use("dynamicOf") + "(state, " + recv.code + ")", k: kDynamic, t: t, inner: recv.code}
	case "(boardgame.ComponentInstance).MoveTo", "(boardgame.ComponentInstance).SecretMoveTo", "(boardgame.ComponentInstance).MoveToFirstSlot",
		"(boardgame.ComponentInstance).MoveToLastSlot", "(boardgame.ComponentInstance).MoveToNextSlot", "(boardgame.ComponentInstance).SlideToFirstSlot",
		"(boardgame.ComponentInstance).SlideToLastSlot", "(boardgame.ImmutableComponentInstance).MayMoveTo", "(boardgame.ImmutableComponentInstance).MayMoveToSlot":
		return facade(strings.ToLower(method[:1])+method[1:], comp())
	case "(boardgame.ImmutableComponentInstance).ContainingImmutableStack", "(boardgame.ComponentInstance).ContainingStack":
		if recv.k == kComponent {
			// ContainingComponent().ImmutableInstance(state): the handle of a
			// values struct, located in state.
			return ret(e.use("containingStack") + "(state, " + recv.code + ")")
		}
		return ret(e.use("containingStack") + "(state, " + comp() + ")")
	case "(boardgame.Component).Deck":
		if recv.k == kComp {
			return value{code: e.use("nonNil") + "(" + recv.code + ")", k: kComponent, t: t}
		}
		return value{code: recv.code, k: kComponent, t: t}
	case "(*boardgame.Deck).Name":
		if recv.k != kComponent {
			e.fail(call, ReasonUnknownType, qualified+" of a deck not from c.Deck()")
		}
		return ret(recv.code + ".deck")
	case "(boardgame.ComponentValues).ContainingComponent", "(*base.ComponentValues).ContainingComponent":
		if recv.k != kValues {
			e.fail(call, ReasonUnknownType, qualified+" on "+typeString(recv.t))
		}
		return value{code: recv.code, k: kComponent, t: t}
	case "(boardgame.Component).ImmutableInstance", "(boardgame.Component).Instance":
		return value{code: recv.code, k: kComponent, t: t}
	case "(boardgame.ImmutableTimer).Active", "(boardgame.Timer).Cancel":
		if recv.k != kTimer {
			e.fail(call, ReasonUnknownType, qualified)
		}
		return ret(recv.code + "." + strings.ToLower(method[:1]) + method[1:] + "()")
	case "(boardgame.Timer).Start":
		if recv.k != kTimer {
			e.fail(call, ReasonUnknownType, qualified)
		}
		d := e.expr(call.Args[0])
		m := e.expr(call.Args[1])
		mc := m.code
		if m.k == kMoveIface {
			mc = e.use("nonNil") + "(" + mc + ")"
		}
		return ret(recv.code + ".start(" + d.code + ", " + mc + ")")
	case "errors.New":
		return ret(e.expr(call.Args[0]).code)
	case "(error).Error":
		return ret(e.use("nonNil") + "(" + recv.code + ")")
	case "fmt.Errorf", "fmt.Sprintf":
		return ret(e.use("sprintf") + "(" + strings.Join(e.args(call), ", ") + ")")
	case "strconv.Itoa", "strconv.Atoi", "strings.Join", "strings.Repeat", "strings.Split", "strings.ToLower",
		"strings.TrimSpace", "strings.HasPrefix", "strings.HasSuffix", "strings.Contains":
		return ret(e.use(strings.ToLower(method[:1])+method[1:]) + "(" + strings.Join(e.args(call), ", ") + ")")
	case "(*dice.DynamicValue).Roll":
		if recv.k != kDynamic {
			e.fail(call, ReasonUnknownType, qualified)
		}
		r := e.expr(call.Args[0])
		if r.k != kRand || r.code != "state.rand()" {
			e.fail(call, ReasonForeignRand, "Roll with a generator other than state.Rand()")
		}
		inner := recv.inner
		if inner == "" {
			inner = recv.code
		}
		return ret(e.use("diceRoll") + "(state, " + inner + ")")
	}
	e.fail(call, ReasonUnknownCall, "emitter has no rule for "+qualified)
	return value{}
}

func (e *emitter) requireEnum(call *ast.CallExpr, v value) {
	if v.k != kEnumVal || v.owner == "" {
		e.fail(call, ReasonUnknownType, "enum call on something other than an enum property")
	}
}
