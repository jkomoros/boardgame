/*
Package legaltest is the golden-equivalence harness for declarative legality:
it checks, on a real opted-in game, that every move's plan reaches the same
verdict as the frozen imperative chain it replaced.

The comparison runs both paths on the SAME state with the same move values:

  - frozen: move.Legal() under GameManager.LegalWithPlansSuspended, which
    hides the assembled plans so every framework Legal() level runs its
    original imperative code;
  - plan: move.Legal() normally, which evaluates the plan once inside
    moves.Default.Legal and lets every framework level return after its
    super-call.

They must agree on the first failure's text (or both pass), on the values of
the move's properties after the call (framework checks such as
moves.AnyPlayer's normalize the move), and on failing when the frozen chain
panics (the plan fails closed instead of panicking).

Typical use, in a game's own tests:

	manager, _ := boardgame.NewGameManager(delegateWhoseMovesAreAllOptedIn, storage)
	game, _ := manager.NewDefaultGame()
	legaltest.Walk(t, game, 20)

OptInAll opts every move in a ConfigureMoves result in with zero authored
preconditions, so a game can run the harness without editing its move list.

This package depends only on core, so package moves' own tests can use it.
*/
package legaltest

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
)

// OptInAll returns configs with optIn applied to a copy of each config's
// custom configuration. Pass moves.WithLegalPreconditions() (no specs) to
// give every move a plan made purely of its base types' contributions.
func OptInAll(configs []boardgame.MoveConfig, optIn func(boardgame.PropertyCollection)) []boardgame.MoveConfig {
	result := make([]boardgame.MoveConfig, len(configs))
	for i, config := range configs {
		custom := make(boardgame.PropertyCollection)
		for key, val := range config.CustomConfiguration() {
			custom[key] = val
		}
		optIn(custom)
		result[i] = boardgame.NewMoveConfig(config.Name(), config.Constructor(), custom)
	}
	return result
}

// Mismatch is one (move, proposer, variant) at which frozen and plan
// legality disagreed.
type Mismatch struct {
	Version  int
	Move     string
	Proposer boardgame.PlayerIndex
	Variant  string
	Frozen   string
	Plan     string
}

func (m Mismatch) String() string {
	return fmt.Sprintf("version %d, move %q, proposer %v, %v: frozen %v, plan %v", m.Version, m.Move, m.Proposer, m.Variant, m.Frozen, m.Plan)
}

type variant struct {
	name  string
	apply func(boardgame.Move)
}

// variantsFor returns the move-field variants worth comparing: the defaults,
// and, for a move with a TargetPlayerIndex, the sentinel, every seat, and one
// past the last seat.
func variantsFor(move boardgame.Move, numPlayers int) []variant {
	variants := []variant{{"defaults", func(boardgame.Move) {}}}
	if move.Reader().Props()["TargetPlayerIndex"] != boardgame.TypePlayerIndex {
		return variants
	}
	set := func(index boardgame.PlayerIndex) func(boardgame.Move) {
		return func(m boardgame.Move) {
			_ = m.ReadSetter().SetPlayerIndexProp("TargetPlayerIndex", index)
		}
	}
	variants = append(variants, variant{"TargetPlayerIndex=observer", set(boardgame.ObserverPlayerIndex)})
	for i := 0; i <= numPlayers; i++ {
		variants = append(variants, variant{fmt.Sprintf("TargetPlayerIndex=%d", i), set(boardgame.PlayerIndex(i))})
	}
	return variants
}

func proposersFor(numPlayers int) []boardgame.PlayerIndex {
	proposers := []boardgame.PlayerIndex{boardgame.ObserverPlayerIndex, boardgame.AdminPlayerIndex}
	for i := 0; i < numPlayers; i++ {
		proposers = append(proposers, boardgame.PlayerIndex(i))
	}
	return proposers
}

type outcome struct {
	text     string
	panicked bool
	props    map[string]interface{}
}

func runLegal(move boardgame.Move, state boardgame.ImmutableState, proposer boardgame.PlayerIndex, wrap func(func())) (result outcome) {
	wrap(func() {
		defer func() {
			if r := recover(); r != nil {
				result.panicked = true
				result.text = fmt.Sprintf("panic: %v", r)
			}
		}()
		if err := move.Legal(state, proposer); err != nil {
			result.text = "error: " + err.Error()
		} else {
			result.text = "legal"
		}
	})
	result.props = make(map[string]interface{})
	reader := move.Reader()
	for name := range reader.Props() {
		val, err := reader.Prop(name)
		if err == nil {
			result.props[name] = fmt.Sprint(val)
		}
	}
	return result
}

type declarer interface {
	DeclaredPreconditions() ([]boardgame.LegalSpec, []string)
}

// pureContribution reports whether move's plan is exactly its framework
// bases' contributions: no authored specs, no suppressions, no LegalCustom,
// and no plan-only base (moves.DrawToPlayer, moves.MoveComponentToSlot),
// whose defining check has no frozen counterpart.
// Only then is the frozen chain the same legality the plan claims to be. A
// move with game-authored legality is fenced by its own golden test, which
// compares against a snapshot of the Legal() it replaced.
func pureContribution(move boardgame.Move) bool {
	if planOnly, ok := move.(interface{ LegalPlanOnly() bool }); ok && planOnly.LegalPlanOnly() {
		return false
	}
	if _, custom := move.(boardgame.CustomLegaler); custom {
		return false
	}
	if d, ok := move.(declarer); ok {
		authored, suppressed := d.DeclaredPreconditions()
		if len(authored) > 0 || len(suppressed) > 0 {
			return false
		}
	}
	return true
}

// Compare checks every move type whose assembled plan is pure contribution
// (see pureContribution) at game's current state, for every proposer (observer, admin, each seat) and move-field
// variant. It returns every disagreement and the number of comparisons made.
func Compare(game *boardgame.Game) ([]Mismatch, int) {
	return CompareAt(game, game.CurrentState())
}

// CompareAt is Compare at any of game's states, including a historical one
// (game.State(version)). Fix-ups are judged at intermediate versions a
// settled game never shows, so Walk compares at every version.
func CompareAt(game *boardgame.Game, state boardgame.ImmutableState) ([]Mismatch, int) {
	manager := game.Manager()
	numPlayers := game.NumPlayers()
	frozen := manager.LegalWithPlansSuspended
	plan := func(fn func()) { fn() }

	var mismatches []Mismatch
	compared := 0
	for _, proto := range game.Moves() {
		name := proto.Info().Name()
		if !manager.LegalPlanAssembled(name) || !pureContribution(proto) {
			continue
		}
		for _, v := range variantsFor(proto, numPlayers) {
			for _, proposer := range proposersFor(numPlayers) {
				frozenMove := game.MoveByNameForState(name, state)
				planMove := game.MoveByNameForState(name, state)
				v.apply(frozenMove)
				v.apply(planMove)

				f := runLegal(frozenMove, state, proposer, frozen)
				p := runLegal(planMove, state, proposer, plan)
				compared++

				mismatch := Mismatch{Version: state.Version(), Move: name, Proposer: proposer, Variant: v.name, Frozen: f.text, Plan: p.text}
				switch {
				case f.panicked:
					// The plan converts a panicking check into a fail-closed
					// Unknown. It must still refuse the move.
					if p.text == "legal" {
						mismatches = append(mismatches, mismatch)
					}
				case f.text != p.text:
					mismatches = append(mismatches, mismatch)
				case !reflect.DeepEqual(f.props, p.props):
					mismatch.Frozen = fmt.Sprintf("%v (props %v)", f.text, f.props)
					mismatch.Plan = fmt.Sprintf("%v (props %v)", p.text, p.props)
					mismatches = append(mismatches, mismatch)
				}
			}
		}
	}
	return mismatches, compared
}

// AssertEquivalent fails t for every disagreement Compare finds at game's
// current state, and returns the number of comparisons made.
func AssertEquivalent(t testing.TB, game *boardgame.Game) int {
	t.Helper()
	return assertAt(t, game, game.CurrentState())
}

func assertAt(t testing.TB, game *boardgame.Game, state boardgame.ImmutableState) int {
	t.Helper()
	mismatches, compared := CompareAt(game, state)
	for _, m := range mismatches {
		t.Errorf("declarative plan disagrees with the frozen Legal(): %v", m)
	}
	return compared
}

// assertVersions compares at every version after from up to the current one,
// returning the comparisons made and the next version to start from.
func assertVersions(t testing.TB, game *boardgame.Game, from int) (int, int) {
	t.Helper()
	compared := 0
	current := game.Version()
	for version := from; version <= current; version++ {
		state := game.State(version)
		if state == nil {
			continue
		}
		compared += assertAt(t, game, state)
	}
	return compared, current + 1
}

// Walk asserts equivalence at every version game has reached, then up to
// maxSteps times proposes the first legal non-fix-up move (seats first, then
// admin), lets fix-ups settle, and asserts again at every new version --
// including the intermediate ones fix-ups were judged at. It returns the
// total number of comparisons, so callers can require the walk actually
// covered something.
func Walk(t testing.TB, game *boardgame.Game, maxSteps int) int {
	t.Helper()
	compared, next := assertVersions(t, game, 0)
	proposers := proposersFor(game.NumPlayers())[2:]
	proposers = append(proposers, boardgame.AdminPlayerIndex)

	for step := 0; step < maxSteps && !game.Finished(); step++ {
		proposed := false
		for _, proto := range game.Moves() {
			if base.IsFixUp(proto) {
				continue
			}
			for _, proposer := range proposers {
				move := game.MoveByName(proto.Info().Name())
				if move.Legal(game.CurrentState(), proposer) != nil {
					continue
				}
				if err := <-game.ProposeMove(move, proposer); err != nil {
					continue
				}
				proposed = true
				break
			}
			if proposed {
				break
			}
		}
		if !proposed {
			break
		}
		more, after := assertVersions(t, game, next)
		compared += more
		next = after
	}
	return compared
}
