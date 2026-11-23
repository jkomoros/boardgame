package moves

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame/legal"
)

/*
TestEveryFrameworkLegalIsItsContribution is the structural half of the
widened composition seam (design spec 2026-11-22-legality-seam-widening):
there is no base-type allowlist, because every framework move type's
legality IS its contribution. That is only true if each type that declares
its own Legal() keeps two things in lockstep, which this test checks by
parsing every non-test moves/*.go file:

 1. Its Legal() has exactly one of the sanctioned shapes:

    super-call, then guard, then its own check --

    if err := <super-call>; err != nil { return err }
    if legalPlanOwnsChain(state, r.Name(), Precondition<T>) { return nil }
    return r.legal<T>(state, proposer)

    plan entry, then its own check (a type that inherits nothing) --

    if handled, err := r.legalPlanEntry(state, proposer, Precondition<T>); handled { return err }
    return r.legal<T>(state, proposer)

    or a single `return r.<helper>(...)` delegating to a method of the same
    type that has the first shape (ApplyUntil.legalWithBase).

 2. It declares ContributedPreconditions() and that body names
    FrameworkPrecondition(Precondition<T>) -- the atom whose check the guard
    skips. A type that guards but forgets to contribute would otherwise drop
    its check silently for every opted-in move.

Default (the plan entry itself) and CurrentPlayer (whose own atom is the
catalog's proposerIsCurrentPlayer, verified by TestLegalChainStringFreeze)
are the two documented exceptions.

TestFrameworkGuardAtomsAreRegistered then checks every atom named by a guard
is registered by FrameworkConstructors, so a guard can never name a check no
plan can evaluate.
*/

var seamExemptTypes = map[string]bool{"Default": true, "CurrentPlayer": true}

type seamMethod struct {
	file string
	decl *ast.FuncDecl
}

func parseSeamMethods(t *testing.T) (map[string]map[string]seamMethod, *token.FileSet) {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("glob moves/*.go found nothing (err %v): the working directory should be the moves package", err)
	}
	fset := token.NewFileSet()
	methods := make(map[string]map[string]seamMethod)
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatalf("parse %v: %v", file, err)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}
			typeName := receiverBaseTypeName(fn.Recv.List[0].Type)
			if typeName == "" {
				continue
			}
			if methods[typeName] == nil {
				methods[typeName] = make(map[string]seamMethod)
			}
			methods[typeName][fn.Name.Name] = seamMethod{file, fn}
		}
	}
	return methods, fset
}

func TestEveryFrameworkLegalIsItsContribution(t *testing.T) {
	methods, fset := parseSeamMethods(t)

	render := func(node ast.Node) string {
		var b strings.Builder
		printer.Fprint(&b, fset, node)
		return strings.Join(strings.Fields(b.String()), " ")
	}

	checked := 0
	for typeName, typeMethods := range methods {
		legalMethod, ok := typeMethods["Legal"]
		if !ok || seamExemptTypes[typeName] {
			continue
		}
		checked++
		recv := legalMethod.decl.Recv.List[0].Names[0].Name
		atom := "Precondition" + typeName
		guard := "if legalPlanOwnsChain(state, " + recv + ".Name(), " + atom + ") { return nil }"
		ownCheck := "return " + recv + ".legal" + typeName + "(state, proposer)"
		entry := "if handled, err := " + recv + ".legalPlanEntry(state, proposer, " + atom + "); handled { return err }"

		hasGuardShape := func(body *ast.BlockStmt) bool {
			if len(body.List) != 3 {
				return false
			}
			first, ok := body.List[0].(*ast.IfStmt)
			if !ok || first.Init == nil || !strings.HasSuffix(render(first.Cond), "err != nil") || render(first.Body) != "{ return err }" {
				return false
			}
			return render(body.List[1]) == guard && render(body.List[2]) == ownCheck
		}

		body := legalMethod.decl.Body
		shapeOK := hasGuardShape(body)
		if !shapeOK && len(body.List) == 2 {
			shapeOK = render(body.List[0]) == entry && render(body.List[1]) == ownCheck
		}
		if !shapeOK && len(body.List) == 1 {
			if ret, ok := body.List[0].(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
				if call, ok := ret.Results[0].(*ast.CallExpr); ok {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == recv {
							if helper, ok := typeMethods[sel.Sel.Name]; ok {
								shapeOK = hasGuardShape(helper.decl.Body)
							}
						}
					}
				}
			}
		}
		if !shapeOK {
			t.Errorf("moves.%v.Legal (%v) does not have a sanctioned seam shape. A framework type's Legal must super-call, then `%v`, then `%v` (or, for a type that inherits nothing, `%v` then the same return) -- otherwise its own check either runs a second time after an opted-in plan or silently disappears from it. See framework_checks.go", typeName, legalMethod.file, guard, ownCheck, entry)
		}

		if _, ok := typeMethods["legal"+typeName]; !ok {
			t.Errorf("moves.%v declares Legal but no legal%v holding its own check (%v)", typeName, typeName, legalMethod.file)
		}
		contributed, ok := typeMethods["ContributedPreconditions"]
		if !ok {
			t.Errorf("moves.%v declares Legal (%v) but not ContributedPreconditions: it would inherit its parent's contribution, so an opted-in move's plan would silently drop %v's own check", typeName, legalMethod.file, typeName)
			continue
		}
		if !strings.Contains(render(contributed.decl.Body), "FrameworkPrecondition("+atom+")") {
			t.Errorf("moves.%v.ContributedPreconditions (%v) does not contribute FrameworkPrecondition(%v), the check its Legal guard skips for opted-in moves", typeName, contributed.file, atom)
		}
	}

	if checked < 20 {
		t.Fatalf("only found %d framework types declaring Legal; the parse is broken, not the invariant", checked)
	}
}

func TestFrameworkGuardAtomsAreRegistered(t *testing.T) {
	methods, _ := parseSeamMethods(t)
	registered := make(map[string]bool)
	for _, ctor := range FrameworkConstructors() {
		registered[ctor.Name] = true
	}
	defaults := make(map[string]bool)
	for _, ctor := range legal.DefaultConstructors() {
		defaults[ctor.Name] = true
	}
	for typeName, typeMethods := range methods {
		if _, ok := typeMethods["Legal"]; !ok || seamExemptTypes[typeName] {
			continue
		}
		atom := strings.ToLower(typeName[:1]) + typeName[1:]
		if !registered[atom] {
			t.Errorf("moves.%v guards framework check %q, but FrameworkConstructors registers no constructor with that name; add it to frameworkCheckConstructors", typeName, atom)
		}
		if defaults[atom] {
			t.Errorf("framework check %q collides with a legal catalog predicate of the same name", atom)
		}
	}
}

// receiverBaseTypeName extracts the base type name from a method receiver's
// type expression, unwrapping a single leading pointer (e.g. "*Default" ->
// "Default"). Returns "" for any shape this package's receivers don't use.
func receiverBaseTypeName(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name
}
