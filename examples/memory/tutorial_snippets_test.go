package memory

import (
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
	"github.com/jkomoros/boardgame/moves"
	storagememory "github.com/jkomoros/boardgame/storage/memory"
)

/*
Compile-checked source for TUTORIAL.md's declarative-legality section (Task
14, user-mandated tutorial integration). The tutorial's before/after for
moveRevealCard and its LegalCustom example (checkers' moveMoveToken) are
copied VERBATIM from real branch source, cited by file:line in HTML
comments right in TUTORIAL.md -- those don't need a second compiled copy
here. This file exists for the tutorial's remaining snippets, which are
illustrative rather than lifted from one exact call site: the catalog
predicate builder cheat-sheet (every constructor signature the tutorial's
table claims exists, called for real) and the WithoutLegalPrecondition
escape-from-inheritance pattern (design spec §2's "ForceFinishTurn, now
expressible declaratively" -- there is no in-repo game that actually calls
WithoutLegalPrecondition yet, so the tutorial's example is necessarily
synthetic and must be proven to compile and Config() cleanly here).
*/

// TestTutorialSnippetCatalogPredicates proves every catalog builder call
// used in TUTORIAL.md's "common predicates" table compiles and returns a
// legal.Spec with the expected registry Name -- the table's whole point is
// that these names/signatures are real, not documentation prose.
func TestTutorialSnippetCatalogPredicates(t *testing.T) {
	specs := map[string]legal.Spec{
		"propAtLeast":                      legal.PropAtLeast("player.CardsLeftToReveal", 1),
		"propCompare":                      legal.PropCompare("game.AttackStrength", ">", 0),
		"playerBool":                       legal.PlayerBool("Eliminated"),
		"playerBoolIs":                     legal.PlayerBoolIs("Eliminated", false),
		"playerBoolAt":                     legal.PlayerBoolAt(legal.Proposer(), "Eliminated", false),
		"stackCount":                       legal.StackCount("game.VisibleCards", "==", 2),
		"stackEmpty":                       legal.StackEmpty("game.VisibleCards"),
		"stackNotEmpty":                    legal.StackNotEmpty("game.HiddenCards"),
		"propEquals":                       legal.PropEquals("game.NumCards", "20"),
		"propNotEquals":                    legal.PropNotEquals("game.NumCards", "0"),
		"componentPresentAt":               legal.ComponentPresentAt("game.HiddenCards", "move.CardIndex"),
		"componentAbsentAt":                legal.ComponentAbsentAt("game.VisibleCards", "move.CardIndex"),
		"componentPresentAtKey":            legal.ComponentPresentAtKey("game.Spaces", "move.TokenIndexToMove"),
		"mayMoveTo":                        legal.MayMoveTo("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
		"mayMoveToSlot":                    legal.MayMoveToSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex", "move.TargetSlot"),
		"mayMoveToSameSlot":                legal.MayMoveToSameSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
		"mayMoveAllTo":                     legal.MayMoveAllTo("game.HiddenCards", "game.VisibleCards"),
		"mayMoveCountTo":                   legal.MayMoveCountTo("game.HiddenCards", "game.VisibleCards", "player.CardsLeftToReveal"),
		"mayMoveFixedCountTo":              legal.MayMoveFixedCountTo("game.HiddenCards", "game.VisibleCards", 2),
		"maySwapComponents":                legal.MaySwapComponents("game.Spaces", "move.FromIndex", "move.ToIndex"),
		"maySwapComponentsByKey":           legal.MaySwapComponentsByKey("game.Spaces", "move.FromSpace", "move.ToSpace"),
		"any":                              legal.Any(legal.PlayerBool("Eliminated"), legal.PlayerBool("Stood")),
		"allActivePlayers":                 legal.AllActivePlayers(legal.PlayerBool("Eliminated")),
		"proposerIsCurrentPlayer":          legal.ProposerIsCurrentPlayer(),
		"proposerIsPlayerFromMove":         legal.ProposerIsPlayerFromMove("TargetPlayerIndex"),
		"revealableCardAt":                 legal.RevealableCardAt("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
		"componentPropEqualsCurrentPlayer": legal.ComponentPropEqualsCurrentPlayer("game.Spaces", "move.TokenIndexToMove", "Color"),
	}

	for name, spec := range specs {
		if spec.Name == "" {
			t.Errorf("%s: Spec.Name was empty", name)
		}
	}
}

// TestTutorialSnippetWithoutLegalPrecondition proves the tutorial's
// WithoutLegalPrecondition example -- a suppression that itself opts the move
// into declarative legality, the moves.ForceFinishTurn "inherit nothing" pattern
// design spec §2 calls out -- actually compiles and Config()s cleanly.
// This needs a real, manager-attached delegate (AutoConfigurer.Config
// calls into a.delegate.Manager()), so it builds a throwaway
// GameManager rather than calling auto.Config in isolation.
func TestTutorialSnippetWithoutLegalPrecondition(t *testing.T) {
	manager, err := boardgame.NewGameManager(NewDelegate(), storagememory.NewStorageManager())
	if err != nil {
		t.Fatalf("tutorial_snippets: building manager: %v", err)
	}

	auto := moves.NewAutoConfigurer(manager.Delegate())

	// moveCaptureCards is a real move type in this package (moves.go),
	// embedding moves.FixUp; it is NOT actually reconfigured in the running
	// game by this call -- Config() just builds a MoveConfig value, it does
	// not register or mutate anything on the live manager. This exercises
	// exactly the syntax the tutorial teaches: WithoutLegalPrecondition both
	// opts in and suppresses an inherited contribution by its stable name.
	//
	// The WithLegalPhases call matters (footgun-batch F2): boot validates
	// every suppression against the move's ACTUAL contributed spec names, so
	// suppressing "inPhase" on a move that never contributes an inPhase
	// check (no WithLegalPhases, no AddForPhase) is a boot error, as is any
	// misspelled name. In a real game the phase config usually arrives via
	// moves.AddForPhase/AddOrderedForPhase rather than an explicit
	// WithLegalPhases; memory has no phase enum, so a literal stands in
	// here — this config is never installed, only Config()'d.
	//
	// Deliberately a moves.FixUp / moves.Default-family embed: suppressing
	// moves.PreconditionProposerIsCurrentPlayer on a
	// moves.CurrentPlayer-embedding move is a BOOT ERROR (the imperative
	// proposer check in CurrentPlayer.Legal would still run, desyncing the
	// client ledger from actual legality — see legal/doc.go's
	// WithoutLegalPrecondition section). Suppress only checks the plan actually
	// controls; on a CurrentPlayer move, embed moves.Default instead if you
	// truly need proposer-free semantics.
	_, err = auto.Config(
		new(moveCaptureCards),
		moves.WithLegalPhases(1),
		moves.WithoutLegalPrecondition(moves.PreconditionInPhase),
	)
	if err != nil {
		t.Fatalf("tutorial_snippets: WithoutLegalPrecondition example failed to Config: %v", err)
	}
}

// TestTutorialSnippetCustomLegaler is a compile-time proof that the shape
// TUTORIAL.md describes for the LegalCustom escape hatch --
// `func (m *T) LegalCustom(state boardgame.ImmutableState, proposer
// boardgame.PlayerIndex) error` -- actually satisfies boardgame.CustomLegaler.
// checkers' moveMoveToken (examples/checkers/moves.go) is the real,
// verbatim-cited example in the tutorial; this just pins the interface
// shape independent of that one call site.
func TestTutorialSnippetCustomLegaler(t *testing.T) {
	var _ boardgame.CustomLegaler = (*tutorialSnippetCustomLegalerExample)(nil)
}

type tutorialSnippetCustomLegalerExample struct{}

func (t *tutorialSnippetCustomLegalerExample) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	return legal.Errorf("tutorial.example_residue", nil)
}
