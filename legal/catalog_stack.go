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
				Cost: boardgame.LegalCostCheap,
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
				Cost: boardgame.LegalCostCheap,
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
					{Path: PropPath(dstPath), Facet: boardgame.LegalFacetOccupancy},
					{Path: PropPath(idxField), Facet: boardgame.LegalFacetValues},
				},
				Cost: boardgame.LegalCostModerate,
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
	}
}

// ExtendDefaults returns DefaultConstructors() with extra appended. This is
// a convenience for a delegate's ConfigurePredicateConstructors when a game
// wants the built-in catalog plus its own predicates (e.g. checkers'
// spaceIsBlack, design spec §8).
func ExtendDefaults(extra ...*PredicateConstructor) []*PredicateConstructor {
	return append(DefaultConstructors(), extra...)
}
