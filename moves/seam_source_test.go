package moves

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/workfit/tester/assert"
)

/*
TestSeamAllowlistTypesDeclareNoLegalMethod is the design spec §5 structural
enforcement for Task 6's widened seam allowlist (legal_plan.go's
legalSupportedMovesBaseTypes): FixUp, FixUpMulti, and StartPhase are safe to
allowlist ONLY because none of them declares its own Legal() method — their
legality IS moves.Default.Legal, verbatim, so plan evaluation composes
exactly as it does for a bare Default-embedding move (see legal_plan.go's
doc comment on legalSupportedMovesBaseTypes). Default and CurrentPlayer
legitimately declare their own Legal() — that is the ORIGINAL v1 seam
(design spec §2); their imperative chains are frozen and verified by
TestLegalChainStringFreeze (preconditions_test.go) — so they are
deliberately excluded from this check.

This is a structural, not textual, guarantee: it parses every moves/*.go
file with go/parser (not a hand-maintained list of "files I remember to
check") and walks every top-level FuncDecl with a receiver, collecting the
receiver's base type name wherever the method name is "Legal". If a future
change gives FixUp, FixUpMulti, or StartPhase its own Legal() override, this
test goes red — forcing whoever makes that change to consciously decide
whether the type should stay on the seam allowlist (its frozen chain would
no longer be provably equivalent to plan evaluation) rather than silently
letting imperative and declarative evaluation start to interleave.

Proof this test actually works (not just passes vacuously): during Task 6
development, seamAllowlistTypesRequiringNoLegalMethod below was temporarily
widened to also include "CurrentPlayer" (which DOES declare its own Legal(),
moves/current_player.go) — the test went RED, naming CurrentPlayer and the
declaring file — then was restored to the real allowlist and confirmed
GREEN again. See the Task 6 report (.superpowers/sdd/task-6-report.md) for
the captured RED failure text.
*/

// seamAllowlistTypesRequiringNoLegalMethod is the design spec §5 seam
// allowlist's members BEYOND Default/CurrentPlayer (which legitimately
// declare Legal — see this file's doc comment) — kept in sync BY HAND with
// legal_plan.go's legalSupportedMovesBaseTypes map minus {"Default",
// "CurrentPlayer"}. A future widening of that allowlist must add the new
// type here too, or this test silently fails to cover it.
var seamAllowlistTypesRequiringNoLegalMethod = map[string]bool{
	"FixUp":      true,
	"FixUpMulti": true,
	"StartPhase": true,
}

func TestSeamAllowlistTypesDeclareNoLegalMethod(t *testing.T) {
	sourceFiles, err := filepath.Glob("*.go")
	assert.For(t, "glob moves/*.go").ThatActual(err).IsNil()
	if len(sourceFiles) == 0 {
		t.Fatal("seam_source_test.go: filepath.Glob(\"*.go\") found no files — the glob is broken, not the invariant; go test's working directory should be the moves package directory")
	}

	fset := token.NewFileSet()

	// typeToFiles maps a receiver base type name that declares a Legal
	// method to the file(s) where that declaration was found, for a
	// legible failure message.
	typeToFiles := make(map[string][]string)

	for _, filename := range sourceFiles {
		f, parseErr := parser.ParseFile(fset, filename, nil, 0)
		if !assert.For(t, "parse", filename).ThatActual(parseErr).IsNil().Passed() {
			continue
		}

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}
			if fn.Name.Name != "Legal" {
				continue
			}
			typeName := receiverBaseTypeName(fn.Recv.List[0].Type)
			if typeName == "" {
				continue
			}
			typeToFiles[typeName] = append(typeToFiles[typeName], filename)
		}
	}

	for name := range seamAllowlistTypesRequiringNoLegalMethod {
		declFiles := typeToFiles[name]
		if len(declFiles) == 0 {
			continue
		}
		t.Errorf("moves.%s declares its own Legal() method (in %v), but it is on the design spec §5 seam allowlist (legal_plan.go's legalSupportedMovesBaseTypes) as a type that must NOT override Legal — the allowlist's whole justification for %s is that its legality IS moves.Default.Legal, verbatim, so plan evaluation composes safely. This is a conscious-seam-decision gate, not necessarily a bug: if moves.%s genuinely needs its own Legal() now, either (a) rewrite it as a super-call into the embedded chain (so plan evaluation stays live — see moves.CurrentPlayer.Legal for the pattern) and keep it on the allowlist, or (b) remove moves.%s from legalSupportedMovesBaseTypes in legal_plan.go (and update seamAllowlistTypesRequiringNoLegalMethod here to match) so a move embedding it can no longer opt in to declarative legality at all.", name, declFiles, name, name, name)
	}
}

// receiverBaseTypeName extracts the base type name from a method receiver's
// type expression, unwrapping a single leading pointer (e.g. "*Default" ->
// "Default"). Returns "" for any shape this package's receivers don't use
// (there are none today, but a shape this function doesn't recognize should
// be ignored rather than crash the test).
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
