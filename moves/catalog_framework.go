package moves

import (
	"fmt"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

/*
This file holds the "inProgression" wrapper predicate: the ONE framework
precondition predicate (of the four stable names in design spec §2 —
"inPhase", "inProgression", "stackConstraints", "proposerIsCurrentPlayer")
that lives in package moves rather than package legal.

LAYERING DECISION (Task 7): inPhase and stackConstraints
(legal/catalog_framework.go) wrap a helper extracted to core
(boardgame.LegalInPhaseCheck / boardgame.LegalStackConstraintsCheck,
legal_framework.go) because their logic only ever touches core types
(ImmutableState, PropertyReader, enum.EnumKey) — core is a package both
moves and legal can import, so extracting there lets both the frozen chain
and the new predicate call one implementation.

inProgression cannot follow that pattern. Move-tape matching fundamentally
depends on moves.MoveProgressionGroup (the Satisfied/tape-matching
interface driving Serial/Parallel/Repeat) and moves/default.go's private
historicalMovesSincePhaseTransition/matchTape machinery (which itself uses
a moves-package-private cache keyed by move name). Moving
MoveProgressionGroup to core would be a much larger, out-of-scope
structural change (the design spec's §2 composition seam is explicitly
Default/CurrentPlayer only for v1, and #644's RepeatFromProp — this same
task — extends MoveProgressionGroup further, which would only compound a
core move). Since package legal cannot import package moves (moves already
imports legal — see the design spec §3's layering diagram: core ← legal ←
moves), the "inProgression" predicate constructor is registered here
instead, in package moves.

This means the registry resolveLegalSpecs consumes for any game with an
opted-in move declaring an "inProgression" spec (i.e., ANY opted-in move
configured with WithLegalMoveProgression — see
Default.ContributedPreconditions) MUST include this file's
FrameworkConstructors() merged alongside legal.DefaultConstructors() (or a
delegate's legal.ConstructorConfigurer override). That merge is NewGameManager
plan-assembly's job (a later task in this campaign) — it is called out
prominently here, and in the Task 7 report, because omitting it produces a
boot-time "unknown predicate name" error for every such move, not a subtle
runtime bug.

Unlike inPhase/stackConstraints, inProgressionConstructor needs no core
extraction and no reimplementation at all: its Evaluate looks up a live
instance of the owning move type (via game.MoveByName, keyed by the move
name captured in the spec at ContributedPreconditions time) and calls that
instance's EXISTING legalMoveInProgression method directly — the exact same
method moves.Default.Legal()'s frozen chain calls. The frozen chain and this
predicate are provably calling the same code, not just similar code.
*/

// inProgressionSpec returns a Spec for the "inProgression" predicate.
// moveName names the move type this precondition was contributed for (see
// Default.ContributedPreconditions). Unlike this package's other
// declarative surface, this is deliberately NOT exported: a hand-authored
// "inProgression" spec naming a DIFFERENT move type than the one it's
// attached to would silently check the wrong move's place in the tape, so
// this is only reachable via ContributedPreconditions, which always passes
// the owning move's own configured name.
func inProgressionSpec(moveName string) legal.Spec {
	return legal.Spec{Name: "inProgression", Args: []string{moveName}}
}

// legalMoveInProgressionChecker is satisfied by any move embedding
// moves.Default (legalMoveInProgression is defined on *Default and promotes
// through embedding). inProgressionConstructor's Evaluate uses it to call
// the frozen chain's own progression check directly, rather than
// reimplementing it.
type legalMoveInProgressionChecker interface {
	legalMoveInProgression(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error
}

// FrameworkConstructors returns the legal.PredicateConstructor for
// "inProgression" — the one framework precondition predicate that must be
// registered from package moves rather than package legal. See this file's
// doc comment for why. Whatever builds the registry resolveLegalSpecs
// consumes (NewGameManager plan-assembly, a later task) must merge this into
// that registry alongside legal.DefaultConstructors() (or a delegate's
// ConfigurePredicateConstructors override).
func FrameworkConstructors() []*legal.PredicateConstructor {
	return []*legal.PredicateConstructor{
		inProgressionConstructor(),
	}
}

// inProgressionConstructor returns the registry entry for "inProgression".
func inProgressionConstructor() *legal.PredicateConstructor {
	return &legal.PredicateConstructor{
		Name: "inProgression",
		Constructor: func(spec legal.Spec, chest *boardgame.ComponentChest, resolve func(legal.Spec) (*legal.Predicate, error)) (*legal.Predicate, error) {
			if len(spec.Args) != 1 {
				return nil, fmt.Errorf("moves: inProgression requires 1 arg (moveName), got %d", len(spec.Args))
			}
			moveName := spec.Args[0]

			template := spec.Message
			if template == "" {
				template = legal.TemplateInProgression
			}

			return &legal.Predicate{
				Name: "inProgression",
				Args: spec.Args,
				// game.moveHistory (design spec §7's sketch) has no
				// PropPath grammar equivalent in v1 — there is no path
				// kind representing derived move-tape data (only
				// game/player/players[*]/move — see boardgame/legal_path.go).
				// Declaring it literally would fail boot-time path
				// validation for every opted-in game. game.Phase is
				// declared instead, since the tape-matching window is
				// bounded by the last phase transition; this is a
				// documented, deliberate deviation from the spec's literal
				// sketch — see the Task 7 report's Reads-honesty note. Like
				// inPhase's "game.Phase" Read, this is by-convention and
				// shares that same known v1 limitation.
				Reads: []legal.Read{
					{Path: legal.PropPath("game.Phase"), Facet: boardgame.LegalFacetValues},
				},
				RequiredReadTypes: map[legal.PropPath]boardgame.PropertyType{legal.PropPath("game.Phase"): boardgame.TypeEnum},
				Cost:              boardgame.LegalCostModerate,
				EmittedTemplates:  []string{template},
				EmittedBindings:   map[string][]string{template: {"detail"}},
				Evaluate: func(ctx legal.Context) legal.Verdict {
					if ctx.State == nil {
						return legal.UnknownVerdict("moves: inProgression: state was nil")
					}
					game := ctx.State.Game()
					if game == nil {
						return legal.UnknownVerdict("moves: inProgression: state's Game() was nil")
					}
					mv := game.MoveByName(moveName)
					if mv == nil {
						return legal.UnknownVerdict("moves: inProgression: no move type registered with name " + moveName)
					}
					checker, ok := mv.(legalMoveInProgressionChecker)
					if !ok {
						return legal.UnknownVerdict("moves: inProgression: move " + moveName + " does not embed moves.Default")
					}
					if err := checker.legalMoveInProgression(ctx.State, ctx.ProposerPlayerIndex); err != nil {
						return legal.FailT(template, map[string]legal.BindingValue{
							"detail": legal.String(err.Error()),
						})
					}
					return legal.PassVerdict()
				},
			}, nil
		},
	}
}
