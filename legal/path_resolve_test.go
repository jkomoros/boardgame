package legal_test

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

// TestTypedResolversUnderEnforcement exercises the exported typed resolvers
// through legal.Evaluate, i.e. with the engine's read scope applied.
func TestTypedResolversUnderEnforcement(t *testing.T) {
	fixture := buildLegalFixture(t, "memoryDefault")

	var (
		numCards    int
		hidden      boardgame.ImmutableStack
		current     boardgame.PlayerIndex
		wrongType   error
		undeclared  error
		declaredErr error
	)
	pred := &legal.Predicate{
		Name: "test.typedResolvers",
		Reads: []legal.Read{
			{Path: "game.NumCards", Facet: boardgame.LegalFacetValues},
			{Path: "game.HiddenCards", Facet: boardgame.LegalFacetValues},
			{Path: "game.CurrentPlayer", Facet: boardgame.LegalFacetValues},
		},
		Evaluate: func(ctx legal.Context) legal.Verdict {
			var err error
			if numCards, err = legal.ResolveInt(ctx, "game.NumCards"); err != nil {
				declaredErr = err
			}
			if hidden, err = legal.ResolveStack(ctx, "game.HiddenCards"); err != nil {
				declaredErr = err
			}
			if current, err = legal.ResolvePlayerIndex(ctx, "game.CurrentPlayer"); err != nil {
				declaredErr = err
			}
			_, wrongType = legal.ResolveInt(ctx, "game.HiddenCards")
			_, undeclared = legal.ResolveInt(ctx, "player.CardsLeftToReveal")
			return legal.PassVerdict()
		},
	}

	verdict := legal.Evaluate(pred, fixture.context(0))

	if declaredErr != nil {
		t.Fatalf("resolving a declared path failed: %v", declaredErr)
	}
	if numCards <= 0 || hidden == nil || current != 0 {
		t.Fatalf("typed values wrong: numCards=%d hidden=%v current=%v", numCards, hidden, current)
	}
	if wrongType == nil || !strings.Contains(wrongType.Error(), "not an int") {
		t.Fatalf("ResolveInt on a stack: want a type error, got %v", wrongType)
	}
	if undeclared == nil {
		t.Fatal("ResolveInt on an undeclared path succeeded")
	}
	// The undeclared read wins over the Pass Evaluate returned.
	if verdict.Outcome != legal.Unknown || !strings.Contains(verdict.Reason, "player.CardsLeftToReveal") {
		t.Fatalf("verdict = %+v, want Unknown naming the undeclared path", verdict)
	}
}
