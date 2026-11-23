package boardgame

import (
	"fmt"
	"strconv"
	"strings"
)

/*
This file implements the component hop of the legal path grammar (design spec
2026-11-22-legality-dynamic-component-values-design.md):

	KIND.Stack[selector].Prop
	KIND.Board[move.Space][selector].Prop

selector is "move.Field" (the slot at an int field's index, or an enum field's
key), "first" (the first occupied slot) or "occupied:N" (the N-th occupied
slot, zero based). Prop resolves against the selected component's static
Values() or its DynamicComponentValues — exactly one of the two must define
it on every component of every reachable deck, which boot validation proves.
Authors never name the storage split; a value is just a property of the
component.

Every predicate that resolves through LegalContext.ResolvePath therefore
works over static and dynamic component values alike, without a parallel
predicate family.
*/

// legalComponentSelectorKind identifies which slot a component selector picks.
type legalComponentSelectorKind int

const (
	legalSelectNone legalComponentSelectorKind = iota
	legalSelectFirst
	legalSelectOrdinal
	legalSelectMoveField
)

// legalComponentSelector is the parsed form of a component bracket.
type legalComponentSelector struct {
	kind      legalComponentSelectorKind
	ordinal   int
	moveField string
}

// legalComponentValueSource says where a component property lives.
type legalComponentValueSource int

const (
	legalComponentStatic legalComponentValueSource = iota
	legalComponentDynamic
)

func parseLegalComponentSelector(path, content string) (legalComponentSelector, error) {
	switch {
	case content == "first":
		return legalComponentSelector{kind: legalSelectFirst}, nil
	case strings.HasPrefix(content, "occupied:"):
		raw := strings.TrimPrefix(content, "occupied:")
		ordinal, err := strconv.Atoi(raw)
		if err != nil || ordinal < 0 || strconv.Itoa(ordinal) != raw {
			return legalComponentSelector{}, fmt.Errorf("boardgame: invalid legal path %q: occupied:N selector needs a non-negative integer, got %q", path, raw)
		}
		return legalComponentSelector{kind: legalSelectOrdinal, ordinal: ordinal}, nil
	case strings.HasPrefix(content, "move."):
		field := strings.TrimPrefix(content, "move.")
		if field == "" || strings.ContainsAny(field, ".[]") {
			return legalComponentSelector{}, fmt.Errorf("boardgame: invalid legal path %q: component selector move.Field needs a field name", path)
		}
		return legalComponentSelector{kind: legalSelectMoveField, moveField: field}, nil
	default:
		return legalComponentSelector{}, fmt.Errorf("boardgame: invalid legal path %q: unknown component selector %q (expected move.Field, first, or occupied:N)", path, content)
	}
}

// legalPathOwnerReaders returns every reader that can own parsed's stack:
// the game state for game.* paths and every player for the player-scoped
// kinds, since which player a path names is only known at evaluation time.
func legalPathOwnerReaders(p LegalPropPath, parsed parsedLegalPath, st ImmutableState) ([]PropertyReader, error) {
	if st == nil {
		return nil, fmt.Errorf("boardgame: legal path %q: no state provided to validate against", p)
	}
	switch parsed.kind {
	case pathGame:
		return []PropertyReader{st.ImmutableGameState().Reader()}, nil
	case pathPlayer, pathProposer, pathPlayersMoveField:
		players := st.ImmutablePlayerStates()
		if len(players) == 0 {
			return nil, fmt.Errorf("boardgame: legal path %q: state has no player states to validate against", p)
		}
		readers := make([]PropertyReader, 0, len(players))
		for _, player := range players {
			readers = append(readers, player.Reader())
		}
		return readers, nil
	default:
		return nil, fmt.Errorf("boardgame: legal path %q: component selectors are only supported on game and player state properties", p)
	}
}

// legalComponentPathDecks returns the distinct decks that the stack hop of
// parsed can hold, across every owner reader.
func legalComponentPathDecks(p LegalPropPath, parsed parsedLegalPath, st ImmutableState) ([]*Deck, error) {
	readers, err := legalPathOwnerReaders(p, parsed, st)
	if err != nil {
		return nil, err
	}
	var decks []*Deck
	seen := make(map[*Deck]bool)
	for _, reader := range readers {
		deck, err := legalComponentDeck(parsed, reader)
		if err != nil {
			return nil, fmt.Errorf("boardgame: legal path %q: cannot resolve component deck: %w", p, err)
		}
		if !seen[deck] {
			seen[deck] = true
			decks = append(decks, deck)
		}
	}
	return decks, nil
}

// legalComponentPathType validates the component hop of parsed at boot and
// returns the PropertyType of the property it reads. The stack hop must be a
// stack (or an indexed board space), a move.Field selector must be an int or
// enum move field, and the component property must resolve to exactly one
// source with one type across every reachable deck.
func legalComponentPathType(p LegalPropPath, parsed parsedLegalPath, exampleState ImmutableState, moveReader PropertyReader) (PropertyType, error) {
	if parsed.boardIndexField == "" {
		readers, err := legalPathOwnerReaders(p, parsed, exampleState)
		if err != nil {
			return TypeIllegal, err
		}
		if actual := readers[0].Props()[parsed.prop]; actual != TypeStack {
			return TypeIllegal, fmt.Errorf("boardgame: legal path %q: property %q has PropertyType %v, expected TypeStack before a component selector", p, parsed.prop, actual)
		}
	}
	if parsed.component.kind == legalSelectMoveField {
		if moveReader == nil {
			return TypeIllegal, fmt.Errorf("boardgame: legal path %q: component selector [move.%s] requires a move reader", p, parsed.component.moveField)
		}
		fieldType, ok := moveReader.Props()[parsed.component.moveField]
		if !ok {
			return TypeIllegal, fmt.Errorf("boardgame: legal path %q: move field %q does not exist", p, parsed.component.moveField)
		}
		if fieldType != TypeInt && fieldType != TypeEnum {
			return TypeIllegal, fmt.Errorf("boardgame: legal path %q: move field %q has PropertyType %v, expected TypeInt or TypeEnum for a component selector", p, parsed.component.moveField, fieldType)
		}
	}
	decks, err := legalComponentPathDecks(p, parsed, exampleState)
	if err != nil {
		return TypeIllegal, err
	}
	dynamic := exampleState.ImmutableDynamicComponentValues()
	result := TypeIllegal
	for _, deck := range decks {
		typeOf, _, err := legalDeckComponentProp(deck, dynamic, parsed.componentProp)
		if err != nil {
			return TypeIllegal, fmt.Errorf("boardgame: legal path %q: %w", p, err)
		}
		if result != TypeIllegal && result != typeOf {
			return TypeIllegal, fmt.Errorf("boardgame: legal path %q: component property %q has PropertyType %v in deck %q but %v in another reachable deck", p, parsed.componentProp, typeOf, deck.Name(), result)
		}
		result = typeOf
	}
	if result == TypeIllegal {
		return TypeIllegal, fmt.Errorf("boardgame: legal path %q: no reachable deck to validate component property %q against", p, parsed.componentProp)
	}
	return result, nil
}

// legalDeckComponentProp classifies prop on deck: static when every
// component's Values() defines it with one type, dynamic when the deck's
// DynamicComponentValues define it. Defined in both places is ambiguous and
// defined in neither is missing; both are errors, as is a static property
// that only some components define.
func legalDeckComponentProp(deck *Deck, dynamic map[string][]ImmutableSubState, prop string) (PropertyType, legalComponentValueSource, error) {
	staticType := TypeIllegal
	staticCount := 0
	components := deck.Components()
	firstMissing := -1
	for i, component := range components {
		var reader PropertyReader
		if component != nil && component.Values() != nil {
			reader = component.Values().Reader()
		}
		if reader == nil {
			if firstMissing == -1 {
				firstMissing = i
			}
			continue
		}
		typeOf, ok := reader.Props()[prop]
		if !ok {
			if firstMissing == -1 {
				firstMissing = i
			}
			continue
		}
		if staticCount > 0 && typeOf != staticType {
			return TypeIllegal, 0, fmt.Errorf("component property %q has PropertyType %v on deck %q component %d but %v on an earlier component", prop, typeOf, deck.Name(), i, staticType)
		}
		staticType = typeOf
		staticCount++
	}
	if staticCount > 0 && firstMissing != -1 {
		return TypeIllegal, 0, fmt.Errorf("component property %q is defined on only some of deck %q's component Values (component %d lacks it)", prop, deck.Name(), firstMissing)
	}

	dynamicType := TypeIllegal
	dynamicOK := false
	if values := dynamic[deck.Name()]; len(values) > 0 && values[0] != nil {
		if reader := values[0].Reader(); reader != nil {
			dynamicType, dynamicOK = reader.Props()[prop]
		}
	}

	switch {
	case staticCount > 0 && dynamicOK:
		return TypeIllegal, 0, fmt.Errorf("component property %q is defined in both deck %q's component Values and its DynamicComponentValues; rename one so the path is unambiguous", prop, deck.Name())
	case staticCount > 0:
		return staticType, legalComponentStatic, nil
	case dynamicOK:
		return dynamicType, legalComponentDynamic, nil
	default:
		return TypeIllegal, 0, fmt.Errorf("component property %q is not defined in deck %q's component Values or DynamicComponentValues", prop, deck.Name())
	}
}

// selectLegalComponent picks the component that parsed's selector names in
// stack. An empty slot or an out-of-range selector is an error, which a
// predicate reports as a fail-closed Unknown.
func selectLegalComponent(path LegalPropPath, parsed parsedLegalPath, stack ImmutableStack, move Move) (ImmutableComponentInstance, error) {
	if stack == nil {
		return nil, fmt.Errorf("boardgame: legal path %q: stack %q was nil", path, parsed.prop)
	}
	var component ImmutableComponentInstance
	switch parsed.component.kind {
	case legalSelectFirst:
		component = stack.ImmutableFirst()
	case legalSelectOrdinal:
		seen := 0
		for i := 0; i < stack.Len(); i++ {
			candidate := stack.ImmutableComponentAt(i)
			if candidate == nil {
				continue
			}
			if seen == parsed.component.ordinal {
				component = candidate
				break
			}
			seen++
		}
	case legalSelectMoveField:
		if move == nil {
			return nil, fmt.Errorf("boardgame: legal path %q: component selector [move.%s] requires a move", path, parsed.component.moveField)
		}
		field := parsed.component.moveField
		moveReader := move.Reader()
		var index int
		switch moveReader.Props()[field] {
		case TypeInt:
			value, err := moveReader.IntProp(field)
			if err != nil {
				return nil, fmt.Errorf("boardgame: legal path %q: could not read move field %q: %w", path, field, err)
			}
			index = value
		case TypeEnum:
			value, err := moveReader.ImmutableEnumProp(field)
			if err != nil || value == nil {
				return nil, fmt.Errorf("boardgame: legal path %q: could not read enum move field %q: %v", path, field, err)
			}
			index = int(value.Value())
		default:
			return nil, fmt.Errorf("boardgame: legal path %q: move field %q is not an int or enum field", path, field)
		}
		if index < 0 || index >= stack.Len() {
			return nil, fmt.Errorf("boardgame: legal path %q: component selector [move.%s] = %d is out of range for a stack of length %d", path, field, index, stack.Len())
		}
		component = stack.ImmutableComponentAt(index)
	default:
		return nil, fmt.Errorf("boardgame: legal path %q: unknown component selector", path)
	}
	if component == nil {
		return nil, fmt.Errorf("boardgame: legal path %q: component selector selected no component", path)
	}
	return component, nil
}

// resolveLegalComponentProp reads parsed's component property from the
// component its selector picks in stack: static Values() first, then
// DynamicComponentValues. Boot validation has proved exactly one defines the
// property, so the order never changes which value is read.
func resolveLegalComponentProp(path LegalPropPath, parsed parsedLegalPath, stack ImmutableStack, move Move) (interface{}, PropertyType, error) {
	component, err := selectLegalComponent(path, parsed, stack, move)
	if err != nil {
		return nil, TypeIllegal, err
	}
	if values := component.Values(); values != nil {
		if reader := values.Reader(); reader != nil {
			if _, ok := reader.Props()[parsed.componentProp]; ok {
				return resolveProp(path, parsed.componentProp, reader)
			}
		}
	}
	if dynamic := component.ImmutableDynamicValues(); dynamic != nil {
		return resolveProp(path, parsed.componentProp, dynamic.Reader())
	}
	return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: component property %q does not exist on the selected component", path, parsed.componentProp)
}

// legalComponentReadEvaluable is LegalReadEvaluable's component-path arm. It
// mirrors sanitizeDynamicComponentValues: a component is only visible inside
// a PolicyVisible stack, and its dynamic values are PolicyHidden unless the
// component is visible, in which case the deck's configured dynamic policy
// applies.
//
//  1. The stack (or board) hop must survive FacetValues — i.e. be
//     PolicyVisible — for every owner the path can name, since selecting a
//     component and seeing anything about it needs its identity and values.
//  2. A static terminal property is then visible.
//  3. A dynamic terminal property must additionally survive read.Facet under
//     the deck's dynamic-value policy for this viewer.
//
// Anything that cannot be classified fails closed.
func legalComponentReadEvaluable(st ImmutableState, transformation *sanitizationTransformation, parsed parsedLegalPath, read LegalRead, movePlayerIsProposer bool) bool {
	var owners []subStateSanitizationTransformation
	var readers []PropertyReader
	kind := parsed.kind
	if kind == pathPlayer && (movePlayerIsProposer || legalPlayerIndexIsWildcard(st.CurrentPlayerIndex())) {
		// The move's player is the proposer here, which this viewer-keyed
		// check cannot trust: judge it like proposer.X, across every player
		// (see legalReadEvaluable's pathPlayer case).
		kind = pathProposer
	}
	switch kind {
	case pathGame:
		owners = []subStateSanitizationTransformation{transformation.Game}
		readers = []PropertyReader{st.ImmutableGameState().Reader()}
	case pathPlayer:
		current := st.CurrentPlayerIndex()
		players := st.ImmutablePlayerStates()
		if current < 0 || int(current) >= len(transformation.Players) || int(current) >= len(players) {
			return false
		}
		owners = []subStateSanitizationTransformation{transformation.Players[current]}
		readers = []PropertyReader{players[current].Reader()}
	case pathProposer, pathPlayersMoveField:
		players := st.ImmutablePlayerStates()
		if len(transformation.Players) == 0 || len(players) != len(transformation.Players) {
			return false
		}
		owners = transformation.Players
		for _, player := range players {
			readers = append(readers, player.Reader())
		}
	default:
		return false
	}
	for _, policies := range owners {
		if !facetSurvives(legalPolicyForProp(policies, parsed.prop), LegalFacetValues) {
			return false
		}
	}
	dynamic := st.ImmutableDynamicComponentValues()
	for _, reader := range readers {
		deck, err := legalComponentDeck(parsed, reader)
		if err != nil {
			return false
		}
		_, source, err := legalDeckComponentProp(deck, dynamic, parsed.componentProp)
		if err != nil {
			return false
		}
		if source != legalComponentDynamic {
			continue
		}
		deckPolicies, ok := transformation.DynamicComponentValues[deck.Name()]
		if !ok {
			return false
		}
		if !facetSurvives(legalPolicyForProp(deckPolicies, parsed.componentProp), read.Facet) {
			return false
		}
	}
	return true
}
