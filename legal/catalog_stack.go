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
	// to move in the first place. Bindings: "index" — but only on THIS
	// branch's own key: a single .WithMessage override collapses this branch
	// and TemplateMayNotMoveTo's onto one key, whose guaranteed bindings are
	// the two branches' intersection (empty), so an overriding template body
	// may not reference {index} or any other placeholder (see
	// TemplateMayNotMoveTo).
	TemplateNoComponentToMove = "legal.no_component_to_move"
	// TemplateMayNotMoveTo is the default Fail template key MayMoveTo and
	// MayMoveToSlot use when a component exists at the source index but
	// MayMoveTo/MayMoveToSlot itself rejects the move. Bindings: "detail",
	// the underlying error's message verbatim (component.go's MayMoveTo/
	// MayMoveToSlot error strings). Note the per-branch bindings above and
	// here apply only to the DEFAULT keys: MayMoveTo/MayMoveToSlot fail on
	// two branches ({index} on one, {detail} on the other), and a single
	// .WithMessage override retargets BOTH branches at your one key, so the
	// bindings guaranteed to render with that key shrink to the branches'
	// intersection — for this predicate, none — and an overriding template
	// body referencing ANY placeholder is a boot error (see
	// boardgame.LegalPredicate.EmittedBindings and legal/doc.go's template
	// tables section).
	TemplateMayNotMoveTo = "legal.may_not_move_to"
	// TemplateMayNotMoveAllTo is the default Fail template key for
	// MayMoveAllTo. Bindings: "detail", the underlying error message.
	TemplateMayNotMoveAllTo = "legal.may_not_move_all_to"
	// TemplateMayNotMoveCountTo is the default Fail template key for
	// MayMoveCountTo. Bindings: "detail", the underlying error message.
	TemplateMayNotMoveCountTo = "legal.may_not_move_count_to"
	// TemplateMayNotSwapComponents is the default Fail template key for
	// MaySwapComponents and MaySwapComponentsByKey. Bindings: "detail", the
	// underlying error message.
	TemplateMayNotSwapComponents = "legal.may_not_swap_components"
	// TemplateComponentPresentUnexpected is the default Fail template key
	// for ComponentAbsentAt (spec §4's negation leaf, the exact inverse of
	// ComponentPresentAt): fired when a component IS present at the index
	// where the predicate requires absence. Bindings: "index".
	TemplateComponentPresentUnexpected = "legal.component_present_unexpected"
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
	TemplatePlayerAlreadySubmitted,
	TemplatePlayerNotSubmitted,
	TemplatePlayerInactive,
	TemplatePlayerActive,
	TemplateSeatNotFilled,
	TemplateSeatNotClosed,
	TemplatePlayerNotAdmin,
	TemplateComponentMissing,
	TemplateComponentMissingKey,
	TemplateNoComponentToMove,
	TemplateMayNotMoveTo,
	TemplateMayNotMoveAllTo,
	TemplateMayNotMoveCountTo,
	TemplateMayNotSwapComponents,
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
	// Task 4 (legality-completeness round, catalog_compare.go) additions:
	// typed equality predicates.
	TemplatePropEquals,
	TemplatePropNotEquals,
	// Task 5 (legality-completeness round, catalog_players.go /
	// catalog_stack.go) additions: negation leaves. playerBool's want-false
	// arm reuses TemplatePlayerBool (already listed above, catalog_compare.go)
	// rather than adding a new key — see playerBoolConstructor's doc comment.
	// componentAbsentAt gets its own key, below.
	TemplateComponentPresentUnexpected,
}

// ComponentPresentAt returns a Spec for the "componentPresentAt" predicate:
// Passes if the stack at stackPath has a non-nil component at the int index
// named by idxField (typically a move.* path, e.g. "move.CardIndex"). Only
// occupancy is read, never the component's values.
func ComponentPresentAt(stackPath, idxField string) Spec {
	return Spec{Name: "componentPresentAt", Args: []string{stackPath, idxField}}
}

// ComponentAbsentAt returns a Spec for the "componentAbsentAt" predicate:
// the exact negation of ComponentPresentAt (spec §4's negation leaf) —
// Passes if the stack at stackPath has NO component (a nil slot) at the int
// index named by idxField. Same Reads shape as ComponentPresentAt (occupancy
// facet on the stack, values facet on idxField — only occupancy is read,
// never the component's values) with the presence check inverted.
func ComponentAbsentAt(stackPath, idxField string) Spec {
	return Spec{Name: "componentAbsentAt", Args: []string{stackPath, idxField}}
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
// the component at sourceIndexField in srcPath could legally be moved to
// destinationSlotField in dstPath, per
// ImmutableComponentInstance.MayMoveToSlot.
//
// Facet honesty: like MayMoveTo, the resolved predicate declares
// LegalFacetValues (not LegalFacetOccupancy) on dstPath. See the doc comment
// on mayMoveConstructor for why.
func MayMoveToSlot(srcPath, dstPath, sourceIndexField, destinationSlotField string) Spec {
	return Spec{Name: "mayMoveToSlot", Args: []string{srcPath, dstPath, sourceIndexField, destinationSlotField}}
}

// MayMoveToSameSlot is a convenience for mirrored stack layouts where the
// same move field selects both the source component and destination slot.
func MayMoveToSameSlot(srcPath, dstPath, indexField string) Spec {
	return MayMoveToSlot(srcPath, dstPath, indexField, indexField)
}

// MayMoveAllTo returns a server-evaluated Spec that passes when every
// component in srcPath could be moved to dstPath transactionally. It is not
// client-evaluable because custom stack constraints may inspect other runtime
// state while simulating the transfer.
func MayMoveAllTo(srcPath, dstPath string) Spec {
	return Spec{Name: "mayMoveAllTo", Args: []string{srcPath, dstPath}}
}

// MayMoveCountTo returns a server-evaluated Spec that passes when exactly the
// number of components in countField could be moved from srcPath to dstPath
// transactionally. It is not client-evaluable because custom stack
// constraints may inspect other runtime state while simulating the transfer.
func MayMoveCountTo(srcPath, dstPath, countField string) Spec {
	return Spec{Name: "mayMoveCountTo", Args: []string{srcPath, dstPath, countField}}
}

// MaySwapComponents returns a Spec that passes when the two int-valued move
// fields identify distinct, in-range slots in stackPath.
func MaySwapComponents(stackPath, firstIndexField, secondIndexField string) Spec {
	return Spec{Name: "maySwapComponents", Args: []string{stackPath, firstIndexField, secondIndexField}}
}

// MaySwapComponentsByKey is the enum-keyed counterpart of MaySwapComponents.
func MaySwapComponentsByKey(stackPath, firstKeyField, secondKeyField string) Spec {
	return Spec{Name: "maySwapComponentsByKey", Args: []string{stackPath, firstKeyField, secondKeyField}}
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
				Name:            "componentPresentAt",
				ClientEvaluable: true,
				Args:            spec.Args,
				Reads: []Read{
					{Path: PropPath(stackPath), Facet: boardgame.LegalFacetOccupancy},
					{Path: PropPath(idxField), Facet: boardgame.LegalFacetValues},
				},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{
					PropPath(stackPath): boardgame.TypeStack,
					PropPath(idxField):  boardgame.TypeInt,
				},
				Cost:             boardgame.LegalCostCheap,
				EmittedTemplates: []string{template},
				EmittedBindings:  map[string][]string{template: {"index"}},
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

// componentAbsentAtConstructor returns the registry entry for
// "componentAbsentAt". Mirrors componentPresentAtConstructor exactly (same
// Reads, same Cost, same nil-move/unresolvable-path -> Unknown handling)
// with the presence check inverted.
func componentAbsentAtConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "componentAbsentAt",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 2 {
				return nil, fmt.Errorf("legal: componentAbsentAt requires 2 args (stackPath, idxField), got %d", len(spec.Args))
			}
			stackPath := spec.Args[0]
			idxField := spec.Args[1]

			template := spec.Message
			if template == "" {
				template = TemplateComponentPresentUnexpected
			}

			return &Predicate{
				Name:            "componentAbsentAt",
				ClientEvaluable: true,
				Args:            spec.Args,
				Reads: []Read{
					{Path: PropPath(stackPath), Facet: boardgame.LegalFacetOccupancy},
					{Path: PropPath(idxField), Facet: boardgame.LegalFacetValues},
				},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{
					PropPath(stackPath): boardgame.TypeStack,
					PropPath(idxField):  boardgame.TypeInt,
				},
				Cost:             boardgame.LegalCostCheap,
				EmittedTemplates: []string{template},
				EmittedBindings:  map[string][]string{template: {"index"}},
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
					if stack.ImmutableComponentAt(idx) == nil {
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
				Name:            "componentPresentAtKey",
				ClientEvaluable: true,
				Args:            spec.Args,
				Reads: []Read{
					{Path: PropPath(stackPath), Facet: boardgame.LegalFacetOccupancy},
					{Path: PropPath(keyField), Facet: boardgame.LegalFacetValues},
				},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{
					PropPath(stackPath): boardgame.TypeStack,
					PropPath(keyField):  boardgame.TypeEnum,
				},
				Cost:             boardgame.LegalCostCheap,
				EmittedTemplates: []string{template},
				EmittedBindings:  map[string][]string{template: {"key"}},
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

// mayMoveConstructor builds either "mayMoveTo" or "mayMoveToSlot".
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
			wantArgs := 3
			argDescription := "srcPath, dstPath, sourceIndexField"
			if useSlot {
				wantArgs = 4
				argDescription = "srcPath, dstPath, sourceIndexField, destinationSlotField"
			}
			if len(spec.Args) != wantArgs {
				return nil, fmt.Errorf("legal: %s requires %d args (%s), got %d", name, wantArgs, argDescription, len(spec.Args))
			}
			srcPath := spec.Args[0]
			dstPath := spec.Args[1]
			sourceIndexField := spec.Args[2]
			destinationSlotField := ""
			if useSlot {
				destinationSlotField = spec.Args[3]
			}

			noComponentTemplate := spec.Message
			if noComponentTemplate == "" {
				noComponentTemplate = TemplateNoComponentToMove
			}
			mayNotMoveTemplate := spec.Message
			if mayNotMoveTemplate == "" {
				mayNotMoveTemplate = TemplateMayNotMoveTo
			}

			// The two branches emit DIFFERENT bindings ({index} vs {detail}).
			// When a Spec.Message override collapses both onto one key, the
			// guaranteed set for that key is the branches' INTERSECTION —
			// empty — so an overriding template may not reference any
			// placeholder (see boardgame.LegalPredicate.EmittedBindings).
			emittedBindings := map[string][]string{
				noComponentTemplate: {"index"},
				mayNotMoveTemplate:  {"detail"},
			}
			if noComponentTemplate == mayNotMoveTemplate {
				emittedBindings = map[string][]string{noComponentTemplate: nil}
			}

			reads := []Read{
				{Path: PropPath(srcPath), Facet: boardgame.LegalFacetOccupancy},
				{Path: PropPath(dstPath), Facet: boardgame.LegalFacetValues},
				{Path: PropPath(sourceIndexField), Facet: boardgame.LegalFacetValues},
			}
			requiredReadTypes := map[PropPath]boardgame.PropertyType{
				PropPath(srcPath):          boardgame.TypeStack,
				PropPath(dstPath):          boardgame.TypeStack,
				PropPath(sourceIndexField): boardgame.TypeInt,
			}
			if useSlot {
				if destinationSlotField != sourceIndexField {
					reads = append(reads, Read{Path: PropPath(destinationSlotField), Facet: boardgame.LegalFacetValues})
				}
				requiredReadTypes[PropPath(destinationSlotField)] = boardgame.TypeInt
			}

			return &Predicate{
				Name:              name,
				ClientEvaluable:   true,
				Args:              spec.Args,
				Reads:             reads,
				RequiredReadTypes: requiredReadTypes,
				Cost:              boardgame.LegalCostModerate,
				EmittedTemplates:  []string{noComponentTemplate, mayNotMoveTemplate},
				EmittedBindings:   emittedBindings,
				Evaluate: func(ctx Context) Verdict {
					sourceIndex, err := resolveIntPath(sourceIndexField, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					destinationSlot := sourceIndex
					if useSlot {
						destinationSlot, err = resolveIntPath(destinationSlotField, ctx)
						if err != nil {
							return UnknownVerdict(err.Error())
						}
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

					comp := src.ImmutableComponentAt(sourceIndex)
					if comp == nil {
						return FailT(noComponentTemplate, map[string]BindingValue{
							"index": Int(sourceIndex),
						})
					}

					var moveErr error
					if useSlot {
						moveErr = comp.MayMoveToSlot(dst, destinationSlot)
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

func mayMoveAllToConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "mayMoveAllTo",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 2 {
				return nil, fmt.Errorf("legal: mayMoveAllTo requires 2 args (srcPath, dstPath), got %d", len(spec.Args))
			}
			srcPath, dstPath := spec.Args[0], spec.Args[1]
			template := spec.Message
			if template == "" {
				template = TemplateMayNotMoveAllTo
			}
			return &Predicate{
				Name:            "mayMoveAllTo",
				ClientEvaluable: false,
				Args:            spec.Args,
				Reads: []Read{
					{Path: PropPath(srcPath), Facet: boardgame.LegalFacetValues},
					{Path: PropPath(dstPath), Facet: boardgame.LegalFacetValues},
				},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{
					PropPath(srcPath): boardgame.TypeStack,
					PropPath(dstPath): boardgame.TypeStack,
				},
				Cost:             boardgame.LegalCostExpensive,
				EmittedTemplates: []string{template},
				EmittedBindings:  map[string][]string{template: {"detail"}},
				Evaluate: func(ctx Context) Verdict {
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
					if err := src.MayMoveAllTo(dst); err != nil {
						return FailT(template, map[string]BindingValue{"detail": String(err.Error())})
					}
					return PassVerdict()
				},
			}, nil
		},
	}
}

func mayMoveCountToConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "mayMoveCountTo",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 3 {
				return nil, fmt.Errorf("legal: mayMoveCountTo requires 3 args (srcPath, dstPath, countField), got %d", len(spec.Args))
			}
			srcPath, dstPath, countField := spec.Args[0], spec.Args[1], spec.Args[2]
			template := spec.Message
			if template == "" {
				template = TemplateMayNotMoveCountTo
			}
			return &Predicate{
				Name:            "mayMoveCountTo",
				ClientEvaluable: false,
				Args:            spec.Args,
				Reads: []Read{
					{Path: PropPath(srcPath), Facet: boardgame.LegalFacetValues},
					{Path: PropPath(dstPath), Facet: boardgame.LegalFacetValues},
					{Path: PropPath(countField), Facet: boardgame.LegalFacetValues},
				},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{
					PropPath(srcPath):    boardgame.TypeStack,
					PropPath(dstPath):    boardgame.TypeStack,
					PropPath(countField): boardgame.TypeInt,
				},
				Cost:             boardgame.LegalCostExpensive,
				EmittedTemplates: []string{template},
				EmittedBindings:  map[string][]string{template: {"detail"}},
				Evaluate: func(ctx Context) Verdict {
					src, err := resolveStackPath(srcPath, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					dst, err := resolveStackPath(dstPath, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					count, err := resolveIntPath(countField, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if src == nil || dst == nil {
						return UnknownVerdict("legal: source or destination stack path resolved to nil")
					}
					if err := src.MayMoveCountTo(dst, count); err != nil {
						return FailT(template, map[string]BindingValue{"detail": String(err.Error())})
					}
					return PassVerdict()
				},
			}, nil
		},
	}
}

func maySwapComponentsConstructor(name string, keyed bool) *PredicateConstructor {
	return &PredicateConstructor{
		Name: name,
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 3 {
				return nil, fmt.Errorf("legal: %s requires 3 args (stackPath, firstIndexField, secondIndexField), got %d", name, len(spec.Args))
			}
			stackPath, firstField, secondField := spec.Args[0], spec.Args[1], spec.Args[2]
			template := spec.Message
			if template == "" {
				template = TemplateMayNotSwapComponents
			}
			fieldType := boardgame.TypeInt
			if keyed {
				fieldType = boardgame.TypeEnum
			}
			reads := []Read{
				{Path: PropPath(stackPath), Facet: boardgame.LegalFacetCount},
				{Path: PropPath(firstField), Facet: boardgame.LegalFacetValues},
			}
			if secondField != firstField {
				reads = append(reads, Read{Path: PropPath(secondField), Facet: boardgame.LegalFacetValues})
			}
			return &Predicate{
				Name:            name,
				ClientEvaluable: true,
				Args:            spec.Args,
				Reads:           reads,
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{
					PropPath(stackPath):   boardgame.TypeStack,
					PropPath(firstField):  fieldType,
					PropPath(secondField): fieldType,
				},
				Cost:             boardgame.LegalCostCheap,
				EmittedTemplates: []string{template},
				EmittedBindings:  map[string][]string{template: {"detail"}},
				Evaluate: func(ctx Context) Verdict {
					stack, err := resolveStackPath(stackPath, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if stack == nil {
						return UnknownVerdict(fmt.Sprintf("legal: stack path %q resolved to nil", stackPath))
					}
					var first, second int
					if keyed {
						firstVal, firstErr := resolveEnumPath(firstField, ctx)
						secondVal, secondErr := resolveEnumPath(secondField, ctx)
						if firstErr != nil {
							return UnknownVerdict(firstErr.Error())
						}
						if secondErr != nil {
							return UnknownVerdict(secondErr.Error())
						}
						if firstVal == nil || secondVal == nil {
							return UnknownVerdict("legal: swap enum path resolved to nil")
						}
						first, second = firstVal.Value().Int(), secondVal.Value().Int()
					} else {
						first, err = resolveIntPath(firstField, ctx)
						if err != nil {
							return UnknownVerdict(err.Error())
						}
						second, err = resolveIntPath(secondField, ctx)
						if err != nil {
							return UnknownVerdict(err.Error())
						}
					}
					if err := stack.MaySwapComponents(first, second); err != nil {
						return FailT(template, map[string]BindingValue{"detail": String(err.Error())})
					}
					return PassVerdict()
				},
			}, nil
		},
	}
}

// ConstructorConfigurer is implemented optionally by a game's GameDelegate to
// register its own predicate constructors on top of the universal catalog
// (design spec §1's checkers.spaceIsBlack example). This package never calls
// ConfigurePredicateConstructors itself: like TemplateConfigurer, it is
// consumed via a type-assertion on the delegate by NewGameManager's plan
// assembly (boardgame/legal_plan.go), which overlays the returned
// constructors (by Name) on the process-wide defaults. Absence means the game
// uses only the default catalog. Return only additions or intentional
// overrides; the built-in catalog is retained automatically.
type ConstructorConfigurer interface {
	// ConfigurePredicateConstructors returns this game's predicate
	// constructors. Names matching a built-in override it; new names extend
	// the catalog.
	ConfigurePredicateConstructors() []*PredicateConstructor
}

// DefaultConstructors returns the full set of pre-built
// LegalPredicateConstructors provided by this package.
func DefaultConstructors() []*PredicateConstructor {
	return []*PredicateConstructor{
		propAtLeastConstructor(),
		propCompareConstructor(),
		playerBoolConstructor(),
		playerBoolAtConstructor(),
		componentPresentAtConstructor(),
		componentAbsentAtConstructor(),
		componentPresentAtKeyConstructor(),
		mayMoveToConstructor(),
		mayMoveToSlotConstructor(),
		mayMoveAllToConstructor(),
		mayMoveCountToConstructor(),
		maySwapComponentsConstructor("maySwapComponents", false),
		maySwapComponentsConstructor("maySwapComponentsByKey", true),
		allActivePlayersConstructor(),
		proposerIsCurrentPlayerConstructor(),
		proposerIsPlayerFromMoveConstructor(),
		revealableCardAtConstructor(),
		componentPropEqualsCurrentPlayerConstructor(),
		inPhaseConstructor(),
		stackConstraintsConstructor(),
		stackCountConstructor(),
		stackEmptyConstructor(),
		stackNotEmptyConstructor(),
		propEqualsConstructor(),
		propNotEqualsConstructor(),
	}
}
