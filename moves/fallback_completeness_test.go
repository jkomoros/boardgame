package moves

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

/*
ensureAllMovesSatisfyFallBack (default.go) is this package's completeness net:
a func that fails to COMPILE if a move type in this package does not satisfy
autoConfigFallbackMoveType, which is what AutoConfigurer.Config falls back on
for a move's name and help text.

A compile-time check only covers the types somebody remembered to list, so the
net silently develops holes: a new move type gets a FallbackName, nobody adds a
line, and the net keeps compiling while no longer asserting anything about it.
ActivateEmptySeat was one such hole, which is part of why nothing noticed it
had never worked.

This test is the net for the net. It reads the package's own source: every type
that declares FallbackName must be named in ensureAllMovesSatisfyFallBack.
*/
func TestEveryFallbackMoveIsInTheCompletenessNet(t *testing.T) {

	fset := token.NewFileSet()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("couldn't list package dir: %v", err)
	}

	declared := map[string]string{}
	var listed map[string]bool

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
		if err != nil {
			t.Fatalf("couldn't parse %v: %v", name, err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) != 1 {
				continue
			}
			if fn.Name.Name == "ensureAllMovesSatisfyFallBack" {
				continue
			}
			receiver := receiverTypeName(fn.Recv.List[0].Type)
			if receiver == "" {
				continue
			}
			if fn.Name.Name == "FallbackName" {
				declared[receiver] = name
			}
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.Name != "ensureAllMovesSatisfyFallBack" {
				continue
			}
			listed = assignedMoveTypes(fn)
		}
	}

	if listed == nil {
		t.Fatal("couldn't find ensureAllMovesSatisfyFallBack in the package source")
	}
	if len(declared) == 0 {
		t.Fatal("found no FallbackName declarations, so this test is asserting nothing")
	}

	for name, file := range declared {
		if !listed[name] {
			t.Errorf("%v declares FallbackName (%v) but is missing from ensureAllMovesSatisfyFallBack in default.go, so nothing checks that it satisfies the fallback contract", name, file)
		}
	}
}

// receiverTypeName returns the bare type name of a (possibly pointer) receiver.
func receiverTypeName(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name
}

// assignedMoveTypes collects the T of every `m = new(T)` in the function body.
func assignedMoveTypes(fn *ast.FuncDecl) map[string]bool {
	result := map[string]bool{}
	ast.Inspect(fn, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := call.Fun.(*ast.Ident)
		if !ok || ident.Name != "new" || len(call.Args) != 1 {
			return true
		}
		if arg, ok := call.Args[0].(*ast.Ident); ok {
			result[arg.Name] = true
		}
		return true
	})
	return result
}
