package legal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/examples/checkers"
	memorygame "github.com/jkomoros/boardgame/examples/memory"
	storagememory "github.com/jkomoros/boardgame/storage/memory"
)

/*
Fixture decision (see task-4-report.md for the full writeup): the design
spec's conformance corpus (§6, §9) is meant to be checked against real named
state fixtures, and the brief's suggestion was "the core test game" —
package boardgame's own unexported _test.go fixture (main_test.go's
testGameState/newTestGame machinery). That's not reachable from this
package: it's unexported, and package legal is a separate package from
package boardgame, so no amount of internal-vs-external test file placement
changes that (Go's export boundary is per-package, not per-file).

Rather than hand-build a brand-new minimal fixture game (which would need
its own boardgame:codegen Reader() boilerplate, since this package can't
reach package boardgame's unexported getDefaultReader either), this file
reuses two existing, fully-codegen'd example games as the standing
conformance fixture: examples/memory (HiddenCards/VisibleCards mirrored
SizedStacks, player.CardsLeftToReveal int, player.PlayerInactive bool,
game.NumCards int — covers PropAtLeast/PropCompare/PlayerBool/
ComponentPresentAt/MayMoveTo/MayMoveToSlot) and examples/checkers
(Spaces SizedStack keyed by an enum — covers ComponentPresentAtKey).
Both are already NewDelegate()-constructible with real Reader()
implementations; constraints/constraints_test.go established the precedent
of importing an example game from an external test package for exactly
this reason. Later tasks' conformance corpora (quantifiers, custom
predicates, etc.) can extend this fixture set the same way.
*/

// legalFixture bundles what a conformance/catalog test needs to build a
// Context: a real state, an optional move (nil exercises the "no move
// provided" Unknown path), and the chest backing that state.
type legalFixture struct {
	state boardgame.ImmutableState
	move  boardgame.Move
	chest *boardgame.ComponentChest
}

// context builds a legal.Context from the fixture for the given proposer.
func (f legalFixture) context(proposer boardgame.PlayerIndex) Context {
	return Context{
		State:    f.state,
		Move:     f.move,
		Proposer: proposer,
		Chest:    f.chest,
	}
}

func newMemoryGame(t *testing.T) (*boardgame.Game, boardgame.State) {
	t.Helper()
	manager, err := boardgame.NewGameManager(memorygame.NewDelegate(), storagememory.NewStorageManager())
	if err != nil {
		t.Fatalf("legal: building memory fixture manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("legal: building memory fixture game: %v", err)
	}
	state, ok := game.CurrentState().(boardgame.State)
	if !ok {
		t.Fatalf("legal: memory fixture CurrentState() was not mutable")
	}
	return game, state
}

func newCheckersGame(t *testing.T) (*boardgame.Game, boardgame.State) {
	t.Helper()
	manager, err := boardgame.NewGameManager(checkers.NewDelegate(), storagememory.NewStorageManager())
	if err != nil {
		t.Fatalf("legal: building checkers fixture manager: %v", err)
	}
	game, err := manager.NewDefaultGame()
	if err != nil {
		t.Fatalf("legal: building checkers fixture game: %v", err)
	}
	state, ok := game.CurrentState().(boardgame.State)
	if !ok {
		t.Fatalf("legal: checkers fixture CurrentState() was not mutable")
	}
	return game, state
}

// memoryMoveWithCardIndex returns a fresh example move from game's manager
// that has a CardIndex int property (memory's RevealCard move), with
// CardIndex set to cardIndex.
func memoryMoveWithCardIndex(t *testing.T, game *boardgame.Game, cardIndex int) boardgame.Move {
	t.Helper()
	for _, mv := range game.Manager().ExampleMoves() {
		if _, ok := mv.Reader().Props()["CardIndex"]; ok {
			if err := mv.ReadSetter().SetIntProp("CardIndex", cardIndex); err != nil {
				t.Fatalf("legal: setting CardIndex: %v", err)
			}
			return mv
		}
	}
	t.Fatal("legal: no memory move with a CardIndex property found")
	return nil
}

// checkersMoveWithTokenIndex returns a fresh example move from game's
// manager that has a TokenIndexToMove enum property (checkers' MoveToken
// move), with TokenIndexToMove set to the given key.
func checkersMoveWithTokenIndex(t *testing.T, game *boardgame.Game, key int) boardgame.Move {
	t.Helper()
	for _, mv := range game.Manager().ExampleMoves() {
		if _, ok := mv.Reader().Props()["TokenIndexToMove"]; ok {
			val, err := mv.ReadSetter().EnumProp("TokenIndexToMove")
			if err != nil {
				t.Fatalf("legal: reading TokenIndexToMove enum prop: %v", err)
			}
			if err := val.SetValue(enum.EnumKey(key)); err != nil {
				t.Fatalf("legal: setting TokenIndexToMove: %v", err)
			}
			return mv
		}
	}
	t.Fatal("legal: no checkers move with a TokenIndexToMove property found")
	return nil
}

// firstOccupiedIndex returns the lowest index in stack with a non-nil
// component, or -1 if stack is entirely empty.
func firstOccupiedIndex(stack boardgame.ImmutableStack) int {
	for i := 0; i < stack.Len(); i++ {
		if stack.ImmutableComponentAt(i) != nil {
			return i
		}
	}
	return -1
}

// firstEmptyIndex returns the lowest index in stack with a nil component,
// or -1 if stack is entirely full.
func firstEmptyIndex(stack boardgame.ImmutableStack) int {
	for i := 0; i < stack.Len(); i++ {
		if stack.ImmutableComponentAt(i) == nil {
			return i
		}
	}
	return -1
}

// legalFixtureBuilders is the named-fixture registry the conformance corpus
// JSON files (testdata/conformance/*.json) reference by name, and that
// catalog_test.go's unit tests also draw on directly.
var legalFixtureBuilders = map[string]func(t *testing.T) legalFixture{
	// memoryDefault: an untouched fresh memory game. HiddenCards[0..19] are
	// all occupied (FinishSetUp distributes every card), VisibleCards is
	// entirely empty, player 0's CardsLeftToReveal is 2 (ResetForTurnStart),
	// game.NumCards is 20. move.CardIndex is 0.
	"memoryDefault": func(t *testing.T) legalFixture {
		game, state := newMemoryGame(t)
		move := memoryMoveWithCardIndex(t, game, 0)
		return legalFixture{state: state, move: move, chest: game.Manager().Chest()}
	},
	// memoryZeroCardsLeft: memoryDefault, but the current player's
	// CardsLeftToReveal is forced to 0.
	"memoryZeroCardsLeft": func(t *testing.T) legalFixture {
		game, state := newMemoryGame(t)
		rs := state.CurrentPlayer().ReadSetter()
		if err := rs.SetIntProp("CardsLeftToReveal", 0); err != nil {
			t.Fatalf("legal: setting CardsLeftToReveal: %v", err)
		}
		move := memoryMoveWithCardIndex(t, game, 0)
		return legalFixture{state: state, move: move, chest: game.Manager().Chest()}
	},
	// memorySeatFilled: memoryDefault, but the current player's SeatFilled
	// is forced to true. (Deliberately NOT PlayerInactive: setting the
	// CURRENT player's PlayerInactive to true changes who counts as current
	// player — PlayerMayBeActive/EnsureValid skip inactive players — which
	// would make the fixture self-defeating for a player.* path predicate.
	// SeatFilled/SeatClosed (behaviors.Seat) don't feed PlayerMayBeActive,
	// so they're safe for this purpose.)
	"memorySeatFilled": func(t *testing.T) legalFixture {
		game, state := newMemoryGame(t)
		rs := state.CurrentPlayer().ReadSetter()
		if err := rs.SetBoolProp("SeatFilled", true); err != nil {
			t.Fatalf("legal: setting SeatFilled: %v", err)
		}
		move := memoryMoveWithCardIndex(t, game, 0)
		return legalFixture{state: state, move: move, chest: game.Manager().Chest()}
	},
	// memoryVisibleOccupied: memoryDefault, but VisibleCards[0] is occupied
	// by moving a DIFFERENT hidden card (the last one) there directly, so
	// HiddenCards[0] stays occupied too. move.CardIndex is 0.
	"memoryVisibleOccupied": func(t *testing.T) legalFixture {
		game, state := newMemoryGame(t)
		gameRS := state.GameState().ReadSetter()
		hidden, err := gameRS.StackProp("HiddenCards")
		if err != nil {
			t.Fatalf("legal: reading HiddenCards: %v", err)
		}
		visible, err := gameRS.StackProp("VisibleCards")
		if err != nil {
			t.Fatalf("legal: reading VisibleCards: %v", err)
		}
		other := hidden.ComponentAt(hidden.Len() - 1)
		if err := other.MoveTo(visible, 0); err != nil {
			t.Fatalf("legal: moving a hidden card into VisibleCards[0]: %v", err)
		}
		move := memoryMoveWithCardIndex(t, game, 0)
		return legalFixture{state: state, move: move, chest: game.Manager().Chest()}
	},
	// memoryNoMove: memoryDefault's state, but with a nil Move — exercises
	// the Unknown-on-missing-move-context path for any predicate whose
	// Reads includes a move.* path.
	"memoryNoMove": func(t *testing.T) legalFixture {
		game, state := newMemoryGame(t)
		return legalFixture{state: state, move: nil, chest: game.Manager().Chest()}
	},
	// checkersDefault: an untouched fresh checkers game (SeatPlayer +
	// PlaceToken fixups have already run, so 24 of the 64 Spaces are
	// occupied). move.TokenIndexToMove is set to the first occupied space.
	"checkersDefault": func(t *testing.T) legalFixture {
		game, state := newCheckersGame(t)
		gameRS := state.GameState().ReadSetter()
		spaces, err := gameRS.StackProp("Spaces")
		if err != nil {
			t.Fatalf("legal: reading Spaces: %v", err)
		}
		idx := firstOccupiedIndex(spaces)
		if idx < 0 {
			t.Fatal("legal: checkers fixture has no occupied space")
		}
		move := checkersMoveWithTokenIndex(t, game, idx)
		return legalFixture{state: state, move: move, chest: game.Manager().Chest()}
	},
	// checkersEmptySpace: checkersDefault, but move.TokenIndexToMove is set
	// to the first EMPTY space instead.
	"checkersEmptySpace": func(t *testing.T) legalFixture {
		game, state := newCheckersGame(t)
		gameRS := state.GameState().ReadSetter()
		spaces, err := gameRS.StackProp("Spaces")
		if err != nil {
			t.Fatalf("legal: reading Spaces: %v", err)
		}
		idx := firstEmptyIndex(spaces)
		if idx < 0 {
			t.Fatal("legal: checkers fixture has no empty space")
		}
		move := checkersMoveWithTokenIndex(t, game, idx)
		return legalFixture{state: state, move: move, chest: game.Manager().Chest()}
	},
	// checkersNoMove: checkersDefault's state, but with a nil Move.
	"checkersNoMove": func(t *testing.T) legalFixture {
		game, state := newCheckersGame(t)
		return legalFixture{state: state, move: nil, chest: game.Manager().Chest()}
	},
}

// buildLegalFixture looks up name in legalFixtureBuilders and builds it,
// failing the test if name is unregistered.
func buildLegalFixture(t *testing.T, name string) legalFixture {
	t.Helper()
	builder, ok := legalFixtureBuilders[name]
	if !ok {
		t.Fatalf("legal: unknown fixture %q", name)
	}
	return builder(t)
}

// resolveSpecViaRegistry resolves spec against registry directly — a
// minimal stand-in for package boardgame's unexported resolveLegalSpecs,
// which this (external) package cannot call. It's sufficient for this
// package's own predicates, none of which are compositors: it looks spec.Name
// up in registry and invokes that constructor, handing it a resolve closure
// that recurses through the same registry (for forward-compatibility with a
// future constructor that composes sub-specs of its own).
func resolveSpecViaRegistry(spec Spec, registry []*PredicateConstructor, chest *boardgame.ComponentChest) (*Predicate, error) {
	for _, c := range registry {
		if c.Name != spec.Name {
			continue
		}
		resolve := func(sub Spec) (*Predicate, error) {
			return resolveSpecViaRegistry(sub, registry, chest)
		}
		return c.Constructor(spec, chest, resolve)
	}
	return nil, fmt.Errorf("legal: unknown predicate name %q", spec.Name)
}

// resolvePredicateForTest resolves spec against DefaultConstructors(),
// failing the test on error.
func resolvePredicateForTest(t *testing.T, spec Spec) *Predicate {
	t.Helper()
	pred, err := resolveSpecViaRegistry(spec, DefaultConstructors(), nil)
	if err != nil {
		t.Fatalf("legal: resolving spec %+v: %v", spec, err)
	}
	return pred
}

// outcomeString renders an Outcome the way the conformance corpus JSON
// spells verdicts ("pass", "fail", "unknown").
func outcomeString(o Outcome) string {
	switch o {
	case Pass:
		return "pass"
	case Fail:
		return "fail"
	case Unknown:
		return "unknown"
	default:
		return "invalid"
	}
}

// conformanceCase is one row of a conformance corpus file's "cases" array.
type conformanceCase struct {
	Spec     Spec   `json:"spec"`
	Fixture  string `json:"fixture"`
	Proposer int    `json:"proposer"`
	Expect   string `json:"expect"`
	// Template, if set, pins the Fail template key this case's Verdict must
	// carry (Verdict.Message.Template). Optional — most useful (and, by
	// convention, always populated) on "fail" cases, where it is what
	// actually distinguishes e.g. mayMoveTo's TemplateNoComponentToMove
	// case from its TemplateMayNotMoveTo case; both expect "fail" but for
	// different reasons, and without pinning the template a corpus edit
	// that silently swapped one for the other would go undetected. Empty
	// (the zero value) means "don't check the template" — used for "pass"
	// (Message is nil) and "unknown" (Reason, not Message, carries the
	// explanation) cases, where there is nothing meaningful to pin.
	Template string `json:"template,omitempty"`
}

// conformanceFile is the top-level shape of a
// testdata/conformance/<name>.json file. This format IS the future Go<->TS
// conformance contract (design spec §6/§9) — keep it dumb JSON.
type conformanceFile struct {
	Predicate string            `json:"predicate"`
	Cases     []conformanceCase `json:"cases"`
}

// TestConformanceCorpus loads every file in testdata/conformance/, resolves
// each case's spec through DefaultConstructors(), builds the named fixture,
// and asserts the resulting Verdict's Outcome matches the case's expected
// verdict.
func TestConformanceCorpus(t *testing.T) {
	paths, err := filepath.Glob("testdata/conformance/*.json")
	if err != nil {
		t.Fatalf("legal: globbing conformance corpus: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("legal: no conformance corpus files found under testdata/conformance/")
	}

	for _, path := range paths {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("legal: reading %s: %v", path, err)
			}
			var cf conformanceFile
			if err := json.Unmarshal(data, &cf); err != nil {
				t.Fatalf("legal: parsing %s: %v", path, err)
			}
			if len(cf.Cases) < 3 {
				t.Fatalf("legal: %s has %d cases, want >= 3", path, len(cf.Cases))
			}
			for i, c := range cf.Cases {
				c := c
				t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
					if c.Spec.Name != cf.Predicate {
						t.Fatalf("legal: case spec name %q does not match file predicate %q", c.Spec.Name, cf.Predicate)
					}
					fixture := buildLegalFixture(t, c.Fixture)
					pred := resolvePredicateForTest(t, c.Spec)
					verdict := pred.Evaluate(fixture.context(boardgame.PlayerIndex(c.Proposer)))
					if got := outcomeString(verdict.Outcome); got != c.Expect {
						t.Errorf("legal: %s case %d (%s, fixture %s): got %s, want %s (verdict: %+v)", path, i, c.Spec.Name, c.Fixture, got, c.Expect, verdict)
					}
					if c.Template != "" {
						if verdict.Message == nil || verdict.Message.Template != c.Template {
							t.Errorf("legal: %s case %d (%s, fixture %s): template = %+v, want %q", path, i, c.Spec.Name, c.Fixture, verdict.Message, c.Template)
						}
					}
				})
			}
		})
	}
}
