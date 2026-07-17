package legal

import (
	"fmt"

	"github.com/jkomoros/boardgame"
)

// Template keys for the purpose-built predicates in this file.
const (
	// TemplateNoCardHere is the default Fail template key RevealableCardAt
	// uses when there is no component at idx in EITHER stack (hidden or
	// visible). No bindings. Preserves examples/memory/moves.go:56's legacy
	// string verbatim: "there is no card at that index".
	TemplateNoCardHere = "legal.no_card_here"
	// TemplateAlreadyRevealed is the default Fail template key
	// RevealableCardAt uses when idx's component has already moved to the
	// visible stack. No bindings. Preserves
	// examples/memory/moves.go:58's legacy string verbatim: "that card has
	// already been revealed".
	TemplateAlreadyRevealed = "legal.already_revealed"
	// TemplateComponentPropNotCurrentPlayer is the default Fail template key
	// for ComponentPropEqualsCurrentPlayer. Bindings: "prop" (the property
	// name compared).
	TemplateComponentPropNotCurrentPlayer = "legal.component_prop_not_current_player"
)

// RevealableCardAt returns a Spec for the "revealableCardAt" predicate: the
// design spec §8 two-branch disambiguation, verbatim. idxField (typically a
// move.* path, e.g. "move.CardIndex") names the SAME index used to look up
// both hiddenPath and visiblePath — the mirrored-stacks pattern memory's
// HiddenCards/VisibleCards uses.
//
//   - If hiddenPath has a component at idx: Pass (the card is still hidden
//     and can be revealed).
//   - Else if visiblePath has NO component at idx either: Fail
//     (TemplateNoCardHere) — idx doesn't name a card at all.
//   - Else (visiblePath has a component at idx): Fail
//     (TemplateAlreadyRevealed) — the card at idx was already revealed.
//
// Reads declare occupancy facets ONLY on both stacks (never values): this
// predicate never inspects what's IN either slot, only whether a slot is
// occupied, which is why it's client-evaluable even under memory's
// sanitize:"order" policy (design spec §8).
func RevealableCardAt(hiddenPath, visiblePath, idxField string) Spec {
	return Spec{Name: "revealableCardAt", Args: []string{hiddenPath, visiblePath, idxField}}
}

// ComponentPropEqualsCurrentPlayer returns a Spec for the
// "componentPropEqualsCurrentPlayer" predicate: checkers' token-ownership
// check (examples/checkers/moves.go:93-138's "!p.Color.Equals(t.Color)"),
// generalized. Passes if the component at the enum-keyed slot keyField in
// the stack at stackPath has an enum property named prop whose value
// equals the CURRENT player's own property of the same name (resolved via
// the "player.<prop>" path — i.e. both sides read a property literally
// named prop, just on different objects).
func ComponentPropEqualsCurrentPlayer(stackPath, keyField, prop string) Spec {
	return Spec{Name: "componentPropEqualsCurrentPlayer", Args: []string{stackPath, keyField, prop}}
}

// revealableCardAtConstructor returns the registry entry for
// "revealableCardAt".
func revealableCardAtConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "revealableCardAt",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 3 {
				return nil, fmt.Errorf("legal: revealableCardAt requires 3 args (hiddenPath, visiblePath, idxField), got %d", len(spec.Args))
			}
			hiddenPath := spec.Args[0]
			visiblePath := spec.Args[1]
			idxField := spec.Args[2]

			// Both branches' templates share the single Spec.Message
			// override, if set — matching mayMoveConstructor's precedent
			// in catalog_stack.go (a caller who overrides the message gets
			// one message for the whole predicate, not per-branch
			// control; per-branch overrides aren't a v1 need).
			noCardTemplate := spec.Message
			if noCardTemplate == "" {
				noCardTemplate = TemplateNoCardHere
			}
			alreadyRevealedTemplate := spec.Message
			if alreadyRevealedTemplate == "" {
				alreadyRevealedTemplate = TemplateAlreadyRevealed
			}

			return &Predicate{
				Name:            "revealableCardAt",
				ClientEvaluable: true,
				Args:            spec.Args,
				Reads: []Read{
					{Path: PropPath(hiddenPath), Facet: boardgame.LegalFacetOccupancy},
					{Path: PropPath(visiblePath), Facet: boardgame.LegalFacetOccupancy},
					{Path: PropPath(idxField), Facet: boardgame.LegalFacetValues},
				},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{
					PropPath(hiddenPath):  boardgame.TypeStack,
					PropPath(visiblePath): boardgame.TypeStack,
					PropPath(idxField):    boardgame.TypeInt,
				},
				Cost:             boardgame.LegalCostCheap,
				EmittedTemplates: []string{noCardTemplate, alreadyRevealedTemplate},
				// Neither branch emits any bindings, so this stays correct
				// (still the empty intersection) when a Spec.Message override
				// collapses both keys onto one.
				EmittedBindings: map[string][]string{
					noCardTemplate:          nil,
					alreadyRevealedTemplate: nil,
				},
				Evaluate: func(ctx Context) Verdict {
					idx, err := resolveIntPath(idxField, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					hidden, err := resolveStackPath(hiddenPath, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					visible, err := resolveStackPath(visiblePath, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if hidden == nil {
						return UnknownVerdict(fmt.Sprintf("legal: stack path %q resolved to nil", hiddenPath))
					}
					if visible == nil {
						return UnknownVerdict(fmt.Sprintf("legal: stack path %q resolved to nil", visiblePath))
					}

					if hidden.ImmutableComponentAt(idx) != nil {
						return PassVerdict()
					}
					if visible.ImmutableComponentAt(idx) == nil {
						return FailT(noCardTemplate)
					}
					return FailT(alreadyRevealedTemplate)
				},
			}, nil
		},
	}
}

// componentPropEqualsCurrentPlayerConstructor returns the registry entry for
// "componentPropEqualsCurrentPlayer".
func componentPropEqualsCurrentPlayerConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "componentPropEqualsCurrentPlayer",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 3 {
				return nil, fmt.Errorf("legal: componentPropEqualsCurrentPlayer requires 3 args (stackPath, keyField, prop), got %d", len(spec.Args))
			}
			stackPath := spec.Args[0]
			keyField := spec.Args[1]
			prop := spec.Args[2]
			playerPath := "player." + prop

			template := spec.Message
			if template == "" {
				template = TemplateComponentPropNotCurrentPlayer
			}

			return &Predicate{
				Name:            "componentPropEqualsCurrentPlayer",
				ClientEvaluable: true,
				Args:            spec.Args,
				Reads: []Read{
					// FacetValues on stackPath (not FacetOccupancy): this
					// predicate reads the VALUE of a property on the
					// component found there, not merely whether a slot is
					// occupied (see design spec §1 Read.Facet: "a
					// stack-size check needs only the count facet" — this
					// is the opposite case, a values check).
					{Path: PropPath(stackPath), Facet: boardgame.LegalFacetValues},
					{Path: PropPath(keyField), Facet: boardgame.LegalFacetValues},
					{Path: PropPath(playerPath), Facet: boardgame.LegalFacetValues},
				},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{
					PropPath(stackPath):  boardgame.TypeStack,
					PropPath(keyField):   boardgame.TypeEnum,
					PropPath(playerPath): boardgame.TypeEnum,
				},
				Cost:             boardgame.LegalCostCheap,
				EmittedTemplates: []string{template},
				EmittedBindings:  map[string][]string{template: {"prop"}},
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
					comp := stack.ImmutableComponentAtKey(key.Value())
					if comp == nil {
						return UnknownVerdict(fmt.Sprintf("legal: no component at key %v in stack %q", key, stackPath))
					}
					compVal, err := comp.Values().Reader().ImmutableEnumProp(prop)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if compVal == nil {
						return UnknownVerdict(fmt.Sprintf("legal: component's enum property %q resolved to nil", prop))
					}
					playerVal, err := resolveEnumPath(playerPath, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if playerVal == nil {
						return UnknownVerdict(fmt.Sprintf("legal: enum path %q resolved to nil", playerPath))
					}
					if compVal.Equals(playerVal) {
						return PassVerdict()
					}
					return FailT(template, map[string]BindingValue{
						"prop": String(prop),
					})
				},
			}, nil
		},
	}
}
