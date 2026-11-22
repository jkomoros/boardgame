package transpile

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/transpile/subset"
)

// packageVarInit is the initializer of a package variable of the game (in
// any file, generated ones included: auto_enum.go declares the enums), and
// whether the package writes the variable anywhere.
func (p *Package) packageVarInit(v *types.Var) (ast.Expr, bool, bool) {
	var init ast.Expr
	found, written := false, false
	for _, file := range p.pkg.Syntax {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				value := spec.(*ast.ValueSpec)
				for i, name := range value.Names {
					if p.info.Defs[name] != v {
						continue
					}
					found = true
					if len(value.Values) == len(value.Names) {
						init = value.Values[i]
					} else if len(value.Values) > 0 {
						written = true
					}
				}
			}
		}
		ast.Inspect(file, func(n ast.Node) bool {
			var targets []ast.Expr
			switch n := n.(type) {
			case *ast.AssignStmt:
				targets = n.Lhs
			case *ast.IncDecStmt:
				targets = []ast.Expr{n.X}
			case *ast.UnaryExpr:
				if n.Op == token.AND {
					targets = []ast.Expr{n.X}
				}
			}
			for _, target := range targets {
				if root := rootIdent(target); root != nil && p.info.ObjectOf(root) == v {
					written = true
				}
			}
			return true
		})
	}
	return init, found, written
}

// frameworkEnum is the enum a framework package variable holds: its name
// and, for a range enum, its dimensions.
type frameworkEnum struct {
	name       string
	dimensions []string
}

func constantInt(info *types.Info, expr ast.Expr) (string, bool) {
	tv, ok := info.Types[expr]
	if !ok || tv.Value == nil {
		return "", false
	}
	value := constant.ToInt(tv.Value)
	if value.Kind() != constant.Int {
		return "", false
	}
	return value.ExactString(), true
}

func constantString(info *types.Info, expr ast.Expr) (string, bool) {
	tv, ok := info.Types[expr]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return "", false
	}
	return constant.StringVal(tv.Value), true
}

// enumVar resolves a package variable built by enums.MustAdd(name, ...),
// MustAddRange(name, dims...) or MustAddTree(name, ...).
func (p *Package) enumVar(v *types.Var) (frameworkEnum, bool) {
	init, found, written := p.packageVarInit(v)
	if !found || written || init == nil {
		return frameworkEnum{}, false
	}
	call, ok := ast.Unparen(init).(*ast.CallExpr)
	if !ok || len(call.Args) == 0 {
		return frameworkEnum{}, false
	}
	fn := p.callee(call)
	if fn == nil {
		return frameworkEnum{}, false
	}
	qualified := subset.QualifiedName(fn)
	switch qualified {
	case "(*enum.Set).MustAdd", "(*enum.Set).MustAddRange", "(*enum.Set).MustAddTree":
	default:
		return frameworkEnum{}, false
	}
	name, ok := constantString(p.info, call.Args[0])
	if !ok {
		return frameworkEnum{}, false
	}
	result := frameworkEnum{name: name}
	if qualified == "(*enum.Set).MustAddRange" {
		for _, arg := range call.Args[1:] {
			dim, ok := constantInt(p.info, arg)
			if !ok {
				return frameworkEnum{}, false
			}
			result.dimensions = append(result.dimensions, dim)
		}
	}
	return result, true
}

// frameworkValue is the TypeScript value of a package variable holding a
// framework object (an enum, a grid graph over a range enum), or ok false.
func (p *Package) frameworkValue(v *types.Var) (value, bool) {
	switch p.kindOf(v.Type()) {
	case kEnum:
		e, ok := p.enumVar(v)
		if !ok {
			return value{}, false
		}
		return value{code: jsString(e.name), k: kEnum, t: v.Type()}, true
	case kGraph:
		init, found, written := p.packageVarInit(v)
		if !found || written || init == nil {
			return value{}, false
		}
		call, ok := ast.Unparen(init).(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return value{}, false
		}
		fn := p.callee(call)
		if fn == nil || subset.QualifiedName(fn) != "graph.MustNewEnumGridConnectedness" {
			return value{}, false
		}
		id, ok := ast.Unparen(call.Args[0]).(*ast.Ident)
		if !ok {
			return value{}, false
		}
		ev, ok := p.info.Uses[id].(*types.Var)
		if !ok {
			return value{}, false
		}
		e, ok := p.enumVar(ev)
		if !ok || len(e.dimensions) == 0 {
			return value{}, false
		}
		var directions []string
		for _, arg := range call.Args[1:] {
			var name *ast.Ident
			switch a := ast.Unparen(arg).(type) {
			case *ast.SelectorExpr:
				name = a.Sel
			case *ast.Ident:
				name = a
			}
			if name == nil {
				return value{}, false
			}
			dir, ok := p.info.Uses[name].(*types.Func)
			if !ok || !subset.FrameworkFuncValues[subset.QualifiedName(dir)] {
				return value{}, false
			}
			directions = append(directions, jsString(dir.Name()))
		}
		code := "{ enum: " + jsString(e.name) + ", dimensions: [" + strings.Join(e.dimensions, ", ") + "], directions: [" + strings.Join(directions, ", ") + "] }"
		return value{code: code, k: kGraph, t: v.Type()}, true
	}
	return value{}, false
}

// constantVarValue is the value of a never-written package variable
// initialized from constants (or not at all).
func (e *emitter) constantVarValue(v *types.Var, at ast.Node) (value, bool) {
	init, found, written := e.p.packageVarInit(v)
	if !found || written {
		return value{}, false
	}
	if init == nil {
		zero := e.p.zeroValue(v.Type())
		if zero == "" {
			return value{}, false
		}
		return value{code: zero, t: v.Type()}, true
	}
	if tv, ok := e.p.info.Types[init]; ok && tv.Value != nil {
		return e.constant(init, tv), true
	}
	if lit, ok := ast.Unparen(init).(*ast.CompositeLit); ok {
		if _, isSlice := e.p.info.TypeOf(lit).Underlying().(*types.Slice); isSlice {
			for _, elt := range lit.Elts {
				if tv, ok := e.p.info.Types[elt]; !ok || tv.Value == nil {
					return value{}, false
				}
			}
			return e.composite(lit), true
		}
	}
	return value{}, false
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
