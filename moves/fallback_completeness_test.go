package moves

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
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

/*
Nothing in this repo checked a moves-package FallbackName against the type that
returns it, and the same class of bug has now bitten twice: ActivateInactivePlayer
returned the plural "Activate Inactive Players", which made TUTORIAL.md print
moves.ActivateInactivePlayers -- a symbol that does not compile -- and made a
test that filtered moves by display name into a test that could silently stop
filtering anything.

A FallbackName is prose, so it cannot be derived mechanically in every case
(half the moves in this package build theirs out of the stacks they were
configured with). But when it IS a plain string literal, it is a rendering of
the type's own name, and any drift between the two is a bug in one of them.
This test asserts exactly that, and only for the literal cases.
*/
func TestLiteralFallbackNamesMatchTheirTypeName(t *testing.T) {

	fset := token.NewFileSet()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("couldn't list package dir: %v", err)
	}

	checked := 0

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
			if !ok || fn.Recv == nil || len(fn.Recv.List) != 1 || fn.Name.Name != "FallbackName" {
				continue
			}
			receiver := receiverTypeName(fn.Recv.List[0].Type)
			if receiver == "" {
				continue
			}
			for _, literal := range returnedStringLiterals(fn) {
				checked++
				if collapseFallbackName(literal) != receiver {
					t.Errorf("%v.FallbackName returns the literal %q, which does not render the type name %v (%v). One of the two is wrong -- and a name nobody can guess from the type is a name the docs will get wrong",
						receiver, literal, receiver, name)
				}
			}
		}
	}

	if checked == 0 {
		t.Fatal("found no literal FallbackName returns, so this test is asserting nothing")
	}
}

// returnedStringLiterals collects every `return "..."` in fn whose returned
// expression is a lone string literal. Computed names (a literal concatenated
// with a configured stack name, say) are skipped: they are prose about the
// configuration, not about the type.
func returnedStringLiterals(fn *ast.FuncDecl) []string {
	var result []string
	ast.Inspect(fn, func(node ast.Node) bool {
		ret, ok := node.(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 1 {
			return true
		}
		lit, ok := ret.Results[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		unquoted, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}
		result = append(result, unquoted)
		return true
	})
	return result
}

// collapseFallbackName turns a display name back into the type name it should
// render: spaces removed, and a trailing "Move" (as in Default's "Default
// Move") dropped, since that suffix is a disambiguator for the base types
// rather than part of the type's name.
func collapseFallbackName(display string) string {
	collapsed := strings.ReplaceAll(display, " ", "")
	return strings.TrimSuffix(collapsed, "Move")
}
