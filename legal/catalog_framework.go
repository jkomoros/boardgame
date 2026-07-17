package legal

import (
	"fmt"
	"strconv"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

/*
This file holds two of the three "framework" wrapper predicates promoting
moves.Default's buried imperative chain to declarative form (design spec §2):
"inPhase" and "stackConstraints". The third, "inProgression", is NOT here —
see moves/catalog_framework.go's doc comment for why (short version:
inProgression needs moves.MoveProgressionGroup, which package legal cannot
import). Both predicates in this file wrap boardgame.LegalInPhaseCheck /
boardgame.LegalStackConstraintsCheck (legal_framework.go, core) rather than
reimplementing anything: those core functions are the SAME code
moves/default.go's frozen legalInPhase/legalStackConstraints methods call,
extracted once so both call sites observably agree by construction. See the
Task 7 report for the full layering decision.

"proposerIsCurrentPlayer" (the fourth stable name from design spec §2) lives
in catalog_players.go, from an earlier task.
*/

// Template keys for the framework predicates in this file (and, via
// TemplateInProgression, for moves/catalog_framework.go's "inProgression"
// predicate too — its default template BODY lives here even though its
// predicate CONSTRUCTOR lives in package moves, so that legal.DefaultTemplates()
// remains the single source of truth for every catalog predicate's default
// rendering, regardless of which package registers the predicate itself).
const (
	// TemplateInPhase is the default Fail template key for "inPhase".
	// Bindings: "detail", the verbatim legacy string from
	// boardgame.LegalInPhaseCheck ("Move is not legal in phase X").
	TemplateInPhase = "legal.in_phase"
	// TemplateInProgression is the default Fail template key for
	// "inProgression" (moves/catalog_framework.go). Bindings: "detail", the
	// verbatim legacy string from moves/default.go's matchTape.
	TemplateInProgression = "legal.in_progression"
	// TemplateStackConstraints is the default Fail template key for
	// "stackConstraints". Bindings: "detail", the verbatim legacy string
	// from component.go's MayMoveTo (via boardgame.LegalStackConstraintsCheck).
	TemplateStackConstraints = "legal.stack_constraints"
)

// InPhase returns a Spec for the "inPhase" predicate: Passes if the game's
// current phase (per the game's delegate, walking TreeEnum ancestors) is one
// of phases. A zero-length phases is legal in every phase. This wraps
// moves/default.go's legalInPhase check verbatim (via
// boardgame.LegalInPhaseCheck); Default.ContributedPreconditions constructs
// this spec automatically from a move type's WithLegalPhases configuration,
// but it may also be authored directly via WithLegalPreconditions.
func InPhase(phases ...enum.EnumKey) Spec {
	args := make([]string, len(phases))
	for i, p := range phases {
		args[i] = strconv.Itoa(p.Int())
	}
	return Spec{Name: "inPhase", Args: args}
}

// StackConstraints returns a Spec for the "stackConstraints" predicate:
// Passes if the first component of the srcProperty stack (read from
// GameState) would be accepted by the dstProperty stack, per
// ImmutableComponentInstance.MayMoveTo. srcProperty/dstProperty are bare
// GameState property names (e.g. "DrawStack"), matching
// moves.WithSourceProperty/WithDestinationProperty's argument shape — NOT
// full "game.X" paths. This wraps moves/default.go's legalStackConstraints
// check verbatim (via boardgame.LegalStackConstraintsCheck);
// Default.ContributedPreconditions constructs this spec automatically when
// both WithSourceProperty and WithDestinationProperty are configured.
func StackConstraints(srcProperty, dstProperty string) Spec {
	return Spec{Name: "stackConstraints", Args: []string{srcProperty, dstProperty}}
}

// inPhaseConstructor returns the registry entry for "inPhase".
func inPhaseConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "inPhase",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {

			phases := make([]enum.EnumKey, len(spec.Args))
			for i, a := range spec.Args {
				n, err := strconv.Atoi(a)
				if err != nil {
					return nil, fmt.Errorf("legal: inPhase: arg %d (%q) is not an integer enum key: %w", i, a, err)
				}
				phases[i] = enum.EnumKey(n)
			}

			template := spec.Message
			if template == "" {
				template = TemplateInPhase
			}

			return &Predicate{
				Name:            "inPhase",
				ClientEvaluable: true,
				Args:            spec.Args,
				Reads: []Read{
					// By-convention Read, in the same spirit as
					// proposerIsCurrentPlayer's "game.CurrentPlayer"
					// (catalog_players.go): base.GameDelegate.CurrentPhase
					// reads gameState.Phase by convention
					// (behaviors.PhaseBehavior). Evaluate below never reads
					// this path directly — it calls
					// boardgame.LegalInPhaseCheck, which is
					// delegate-correct even for a delegate that overrides
					// CurrentPhase non-conventionally. A delegate that does
					// so without backing CurrentPhase by a "Phase"
					// game-state property would fail boot-time path
					// validation for this specific declared Read — a known
					// v1 limitation, not exercised by any in-repo game
					// today (see the design spec §1's Reads-conservativeness
					// risk note, §10).
					{Path: PropPath("game.Phase"), Facet: boardgame.LegalFacetValues},
				},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{PropPath("game.Phase"): boardgame.TypeEnum},
				Cost:              boardgame.LegalCostCheap,
				EmittedTemplates:  []string{template},
				EmittedBindings:   map[string][]string{template: {"detail"}},
				Evaluate: func(ctx Context) Verdict {
					if ctx.State == nil {
						return UnknownVerdict("legal: inPhase: state was nil")
					}
					if err := boardgame.LegalInPhaseCheck(ctx.State, phases); err != nil {
						return FailT(template, map[string]BindingValue{
							"detail": String(err.Error()),
						})
					}
					return PassVerdict()
				},
			}, nil
		},
	}
}

// stackConstraintsConstructor returns the registry entry for
// "stackConstraints".
func stackConstraintsConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "stackConstraints",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 2 {
				return nil, fmt.Errorf("legal: stackConstraints requires 2 args (srcProperty, dstProperty), got %d", len(spec.Args))
			}
			srcName := spec.Args[0]
			dstName := spec.Args[1]

			template := spec.Message
			if template == "" {
				template = TemplateStackConstraints
			}

			return &Predicate{
				Name:            "stackConstraints",
				ClientEvaluable: true,
				Args:            spec.Args,
				Reads: []Read{
					{Path: PropPath("game." + srcName), Facet: boardgame.LegalFacetValues},
					{Path: PropPath("game." + dstName), Facet: boardgame.LegalFacetValues},
				},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{
					PropPath("game." + srcName): boardgame.TypeStack,
					PropPath("game." + dstName): boardgame.TypeStack,
				},
				Cost:             boardgame.LegalCostModerate,
				EmittedTemplates: []string{template},
				EmittedBindings:  map[string][]string{template: {"detail"}},
				Evaluate: func(ctx Context) Verdict {
					if ctx.State == nil {
						return UnknownVerdict("legal: stackConstraints: state was nil")
					}
					if err := boardgame.LegalStackConstraintsCheck(ctx.State, srcName, dstName); err != nil {
						return FailT(template, map[string]BindingValue{
							"detail": String(err.Error()),
						})
					}
					return PassVerdict()
				},
			}, nil
		},
	}
}
