package legal

import (
	"fmt"

	"github.com/jkomoros/boardgame"
)

// Template keys for the presence/stack predicates in this file.
const (
	// TemplateComponentMissing is the default Fail template key for
	// ComponentPresentAt. Bindings: "index".
	TemplateComponentMissing = "legal.component_missing"
	// TemplateComponentMissingKey is the default Fail template key for
	// ComponentPresentAtKey. Bindings: "key".
	TemplateComponentMissingKey = "legal.component_missing_key"
	// TemplateNoComponentToMove is the default Fail template key MayMoveTo
	// and MayMoveToSlot use when there is no component at the source index
	// to move in the first place. Bindings: "index".
	TemplateNoComponentToMove = "legal.no_component_to_move"
	// TemplateMayNotMoveTo is the default Fail template key MayMoveTo and
	// MayMoveToSlot use when a component exists at the source index but
	// MayMoveTo/MayMoveToSlot itself rejects the move. Bindings: "detail",
	// the underlying error's message verbatim (component.go's MayMoveTo/
	// MayMoveToSlot error strings).
	TemplateMayNotMoveTo = "legal.may_not_move_to"
)

// DefaultTemplateKeys returns a copy of defaultTemplateKeys: every template
// key the catalog predicates in this package default to when Spec.Message is
// unset. Exported so this package's external test package (legal_test —
// necessary from Task 7 onward, since package legal cannot import package
// moves, but some of this package's own tests need real example games built
// on package moves as fixtures) can cross-check DefaultTemplates()' coverage
// against the same list catalog authors maintain by hand.
func DefaultTemplateKeys() []string {
	out := make([]string, len(defaultTemplateKeys))
	copy(out, defaultTemplateKeys)
	return out
}

// defaultTemplateKeys lists every template key the catalog predicates in
// this package (catalog_compare.go and catalog_stack.go) default to when
// Spec.Message is unset. This is the handoff for Task 6's
// legal.DefaultTemplates(): that function is expected to provide a default
// human-readable template string for every key in this list (games may
// still override any of them via their own ConfigureLegalTemplates()).
var defaultTemplateKeys = []string{
	TemplatePropAtLeast,
	TemplatePropCompare,
	TemplatePlayerBool,
	TemplateComponentMissing,
	TemplateComponentMissingKey,
	TemplateNoComponentToMove,
	TemplateMayNotMoveTo,
	// Task 5 (catalog_players.go / catalog_purpose.go) additions:
	TemplateAllActivePlayers,
	TemplateProposerTargetInvalid,
	TemplateProposerNotYourTurn,
	TemplateNoCardHere,
	TemplateAlreadyRevealed,
	TemplateComponentPropNotCurrentPlayer,
	// Task 7 (catalog_framework.go) additions: inPhase/stackConstraints
	// (predicate AND template both live in this package) and
	// inProgression (predicate lives in package moves; only its default
	// template body lives here — see catalog_framework.go's doc comment).
	TemplateInPhase,
	TemplateInProgression,
	TemplateStackConstraints,
	// Task 2 (legality-completeness round, catalog_count.go) additions:
	// count/emptiness predicates on FacetCount/FacetNonEmpty.
	TemplateStackCount,
	TemplateStackEmpty,
	TemplateStackNotEmpty,
}

// ComponentPresentAt returns a Spec for the "componentPresentAt" predicate:
// Passes if the stack at stackPath has a non-nil component at the int index
// named by idxField (typically a move.* path, e.g. "move.CardIndex"). Only
// occupancy is read, never the component's values.
func ComponentPresentAt(stackPath, idxField string) Spec {
	return Spec{Name: "componentPresentAt", Args: []string{stackPath, idxField}}
}

// ComponentPresentAtKey returns a Spec for the "componentPresentAtKey"
// predicate: like ComponentPresentAt, but the stack slot is identified by
// an enum-valued keyField (e.g. checkers' board-position enum) rather than
// a plain int, for stacks whose slots are keyed by enum values.
func ComponentPresentAtKey(stackPath, keyField string) Spec {
	return Spec{Name: "componentPresentAtKey", Args: []string{stackPath, keyField}}
}

// MayMoveTo returns a Spec for the "mayMoveTo" predicate: Passes if the
// component at index idxField in the stack at srcPath could legally be
// moved to the stack at dstPath, per
// ImmutableComponentInstance.MayMoveTo (component.go). This does not check
// a specific destination slot; see MayMoveToSlot for that.
//
// Facet honesty: unlike ComponentPresentAt, the resolved predicate declares
// LegalFacetValues (not LegalFacetOccupancy) on dstPath. See the doc comment
// on mayMoveConstructor for why.
func MayMoveTo(srcPath, dstPath, idxField string) Spec {
	return Spec{Name: "mayMoveTo", Args: []string{srcPath, dstPath, idxField}}
}

// MayMoveToSlot returns a Spec for the "mayMoveToSlot" predicate: Passes if
// the component at index idxField in the stack at srcPath could legally be
// moved to slot idxField in the stack at dstPath, per
// ImmutableComponentInstance.MayMoveToSlot (component.go). idxField names
// the SAME index used for both the source lookup and the destination slot
// — the mirrored-stacks pattern memory's HiddenCards/VisibleCards uses
// (design spec §8).
//
// Facet honesty: like MayMoveTo, the resolved predicate declares
// LegalFacetValues (not LegalFacetOccupancy) on dstPath. See the doc comment
// on mayMoveConstructor for why.
func MayMoveToSlot(srcPath, dstPath, idxField string) Spec {
	return Spec{Name: "mayMoveToSlot", Args: []string{srcPath, dstPath, idxField}}
}

// componentPresentAtConstructor returns the registry entry for
// "componentPresentAt".
func componentPresentAtConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "componentPresentAt",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 2 {
				return nil, fmt.Errorf("legal: componentPresentAt requires 2 args (stackPath, idxField), got %d", len(spec.Args))
			}
			stackPath := spec.Args[0]
			idxField := spec.Args[1]

			template := spec.Message
			if template == "" {
				template = TemplateComponentMissing
			}

			return &Predicate{
				Name: "componentPresentAt",
				Args: spec.Args,
				Reads: []Read{
					{Path: PropPath(stackPath), Facet: boardgame.LegalFacetOccupancy},
					{Path: PropPath(idxField), Facet: boardgame.LegalFacetValues},
				},
				Cost:             boardgame.LegalCostCheap,
				EmittedTemplates: []string{template},
				Evaluate: func(ctx Context) Verdict {
					idx, err := resolveIntPath(idxField, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					stack, err := resolveStackPath(stackPath, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if stack == nil {
						return UnknownVerdict(fmt.Sprintf("legal: stack path %q resolved to nil", stackPath))
					}
					if stack.ImmutableComponentAt(idx) != nil {
						return PassVerdict()
					}
					return FailT(template, map[string]BindingValue{
						"index": Int(idx),
					})
				},
			}, nil
		},
	}
}

// componentPresentAtKeyConstructor returns the registry entry for
// "componentPresentAtKey".
func componentPresentAtKeyConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "componentPresentAtKey",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 2 {
				return nil, fmt.Errorf("legal: componentPresentAtKey requires 2 args (stackPath, keyField), got %d", len(spec.Args))
			}
			stackPath := spec.Args[0]
			keyField := spec.Args[1]

			template := spec.Message
			if template == "" {
				template = TemplateComponentMissingKey
			}

			return &Predicate{
				Name: "componentPresentAtKey",
				Args: spec.Args,
				Reads: []Read{
					{Path: PropPath(stackPath), Facet: boardgame.LegalFacetOccupancy},
					{Path: PropPath(keyField), Facet: boardgame.LegalFacetValues},
				},
				Cost:             boardgame.LegalCostCheap,
				EmittedTemplates: []string{template},
				Evaluate: func(ctx Context) Verdict {
					key, err := resolveEnumPath(keyField, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if key == nil {
						return UnknownVerdict(fmt.Sprintf("legal: enum path %q resolved to nil", keyField))
					}
					stack, err := resolveStackPath(stackPath, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if stack == nil {
						return UnknownVerdict(fmt.Sprintf("legal: stack path %q resolved to nil", stackPath))
					}
					if stack.ImmutableComponentAtKey(key.Value()) != nil {
						return PassVerdict()
					}
					return FailT(template, map[string]BindingValue{
						"key": String(key.String()),
					})
				},
			}, nil
		},
	}
}

// mayMoveConstructor builds either "mayMoveTo" or "mayMoveToSlot",
// depending on useSlot: both share arg shape (srcPath, dstPath, idxField)
// and Reads, differing only in whether MayMoveTo or MayMoveToSlot is
// called on the component found at the source index.
//
// Facet honesty (dstPath declares LegalFacetValues, not LegalFacetOccupancy,
// unlike srcPath): Evaluate calls comp.MayMoveTo(dst) / comp.MayMoveToSlot
// (component.go), which end by calling dst.CheckConstraints(...)
// (stack.go). CheckConstraints runs every StackConstraint attached to dst
// (struct-tag-attached via constraints.Same/constraints.Unique/etc, or
// added programmatically via Stack.AddConstraint) — see stack_constraint.go
// — and those constraint functions are handed the destination stack itself,
// which they are free to inspect by VALUE (constraints.Same and
// constraints.Unique, for two examples that ship in this repo, both compare
// component property VALUES already in the destination stack against the
// proposed component). So a client that only sanitizes dst down to
// PolicyOrder (which LegalFacetOccupancy would call safe) could still leak a
// values-comparison through this predicate's verdict, if dst happens to
// carry a values-reading constraint.
//
// Ideally this predicate would declare LegalFacetOccupancy on dst when dst
// provably carries no constraints, and only fall back to LegalFacetValues
// when it does — narrowing client evaluability instead of pessimizing it
// unconditionally. That requires inspecting dst's attached constraints at
// LegalPredicate construction time, using an example state (resolveLegalSpecs
// already threads one through for boot-time path validation). Two things
// stand in the way today: (1) ImmutableStack/Stack expose no way to ask "do
// you have any constraints attached" short of actually invoking
// CheckConstraints with a real proposed component (which reports whether a
// SPECIFIC move would be accepted, not whether the stack merely carries a
// values-reading constraint) — there is no exported NumConstraints/
// HasConstraints/Constraints accessor in core to grep for; and (2)
// LegalPredicateConstructor.Constructor's signature (spec, chest, resolve)
// does not receive the example state resolveLegalSpecs holds, so even a
// constraint-count accessor wouldn't be reachable from here without a wider
// signature change touching every registered constructor, not just this
// one. Absent either of those, this predicate declares LegalFacetValues on
// dstPath unconditionally, which is always honest (never under-declares) at
// the cost of sometimes being more conservative than necessary. Narrowing
// this via one or both of the above is legitimate future work.
func mayMoveConstructor(name string, useSlot bool) *PredicateConstructor {
	return &PredicateConstructor{
		Name: name,
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 3 {
				return nil, fmt.Errorf("legal: %s requires 3 args (srcPath, dstPath, idxField), got %d", name, len(spec.Args))
			}
			srcPath := spec.Args[0]
			dstPath := spec.Args[1]
			idxField := spec.Args[2]

			noComponentTemplate := spec.Message
			if noComponentTemplate == "" {
				noComponentTemplate = TemplateNoComponentToMove
			}
			mayNotMoveTemplate := spec.Message
			if mayNotMoveTemplate == "" {
				mayNotMoveTemplate = TemplateMayNotMoveTo
			}

			return &Predicate{
				Name: name,
				Args: spec.Args,
				Reads: []Read{
					{Path: PropPath(srcPath), Facet: boardgame.LegalFacetOccupancy},
					{Path: PropPath(dstPath), Facet: boardgame.LegalFacetValues},
					{Path: PropPath(idxField), Facet: boardgame.LegalFacetValues},
				},
				Cost:             boardgame.LegalCostModerate,
				EmittedTemplates: []string{noComponentTemplate, mayNotMoveTemplate},
				Evaluate: func(ctx Context) Verdict {
					idx, err := resolveIntPath(idxField, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					src, err := resolveStackPath(srcPath, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					dst, err := resolveStackPath(dstPath, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if src == nil || dst == nil {
						return UnknownVerdict("legal: source or destination stack path resolved to nil")
					}

					comp := src.ImmutableComponentAt(idx)
					if comp == nil {
						return FailT(noComponentTemplate, map[string]BindingValue{
							"index": Int(idx),
						})
					}

					var moveErr error
					if useSlot {
						moveErr = comp.MayMoveToSlot(dst, idx)
					} else {
						moveErr = comp.MayMoveTo(dst)
					}
					if moveErr != nil {
						return FailT(mayNotMoveTemplate, map[string]BindingValue{
							"detail": String(moveErr.Error()),
						})
					}
					return PassVerdict()
				},
			}, nil
		},
	}
}

func mayMoveToConstructor() *PredicateConstructor {
	return mayMoveConstructor("mayMoveTo", false)
}

func mayMoveToSlotConstructor() *PredicateConstructor {
	return mayMoveConstructor("mayMoveToSlot", true)
}

// ConstructorConfigurer is implemented optionally by a game's GameDelegate to
// register its own predicate constructors on top of the universal catalog
// (design spec §1's checkers.spaceIsBlack example). This package never calls
// ConfigurePredicateConstructors itself: like TemplateConfigurer, it is
// consumed via a type-assertion on the delegate by NewGameManager's plan
// assembly (boardgame/legal_plan.go), which overlays the returned
// constructors (by Name) on the process-wide defaults. Absence means the game
// uses only the default catalog. A delegate that wants the built-in catalog
// plus its own returns ExtendDefaults(...) here.
type ConstructorConfigurer interface {
	// ConfigurePredicateConstructors returns this game's predicate
	// constructors. Names matching a built-in override it; new names extend
	// the catalog.
	ConfigurePredicateConstructors() []*PredicateConstructor
}

// DefaultConstructors returns the full set of pre-built
// LegalPredicateConstructors provided by this package. Games consuming the
// default catalog get these for free; games registering additional
// predicates of their own use ExtendDefaults instead.
func DefaultConstructors() []*PredicateConstructor {
	return []*PredicateConstructor{
		propAtLeastConstructor(),
		propCompareConstructor(),
		playerBoolConstructor(),
		componentPresentAtConstructor(),
		componentPresentAtKeyConstructor(),
		mayMoveToConstructor(),
		mayMoveToSlotConstructor(),
		allActivePlayersConstructor(),
		proposerIsCurrentPlayerConstructor(),
		revealableCardAtConstructor(),
		componentPropEqualsCurrentPlayerConstructor(),
		inPhaseConstructor(),
		stackConstraintsConstructor(),
		stackCountConstructor(),
		stackEmptyConstructor(),
		stackNotEmptyConstructor(),
	}
}

// ExtendDefaults returns DefaultConstructors() with extra appended. This is
// a convenience for a delegate's ConfigurePredicateConstructors when a game
// wants the built-in catalog plus its own predicates (e.g. checkers'
// spaceIsBlack, design spec §8).
func ExtendDefaults(extra ...*PredicateConstructor) []*PredicateConstructor {
	return append(DefaultConstructors(), extra...)
}
