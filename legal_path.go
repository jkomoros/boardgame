package boardgame

import (
	"fmt"
	"strings"
)

// legalPathKind identifies which of the four path grammars a parsed
// LegalPropPath uses. See the design spec §1 "Path grammar".
type legalPathKind int

const (
	// pathGame denotes a "game.X" path, resolved against the game state.
	pathGame legalPathKind = iota
	// pathPlayer denotes a "player.X" path, resolved against the current
	// player, i.e. the player state at state.CurrentPlayerIndex(). See
	// resolveLegalPath for how an invalid/Observer/Admin/AnyPlayerIndex
	// current player is handled.
	pathPlayer
	// pathPlayersAll denotes a "players[*].X" path. This kind is
	// quantifier-only (spec §1): it names a property to be read from every
	// player by a quantified predicate, but it does not itself resolve to a
	// single value. resolveLegalPath returns an error if asked to resolve
	// one directly.
	pathPlayersAll
	// pathMove denotes a "move.X" path, resolved against the move being
	// evaluated.
	pathMove
)

// parsedLegalPath is the result of successfully parsing a LegalPropPath: the
// path's kind, and the property name to look up within that kind's reader.
type parsedLegalPath struct {
	kind legalPathKind
	prop string
}

// parseLegalPath parses p according to the path grammar in spec §1:
// "game.X", "player.X" (current player), "players[*].X" (quantifier-only),
// or "move.X". The KIND segment is matched literally and case-sensitively
// against exactly those four spellings ("players[0]" or any other concrete
// index is not valid grammar — only the "*" quantifier is).
//
// parseLegalPath only checks the path's shape; it does not check that the
// named property actually exists on any reader. That is validateLegalPath's
// job, run once at boot (NewGameManager) against real example readers so a
// typo fails at boot naming the move and path, never mid-game.
func parseLegalPath(p LegalPropPath) (parsedLegalPath, error) {
	s := string(p)

	kindStr, prop, ok := strings.Cut(s, ".")
	if !ok {
		return parsedLegalPath{}, fmt.Errorf("boardgame: invalid legal path %q: expected KIND.Property (KIND is one of game, player, players[*], move)", s)
	}

	if prop == "" {
		return parsedLegalPath{}, fmt.Errorf("boardgame: invalid legal path %q: missing property name after %q", s, kindStr)
	}

	var kind legalPathKind
	switch kindStr {
	case "game":
		kind = pathGame
	case "player":
		kind = pathPlayer
	case "players[*]":
		kind = pathPlayersAll
	case "move":
		kind = pathMove
	default:
		return parsedLegalPath{}, fmt.Errorf("boardgame: invalid legal path %q: unknown path kind %q (expected game, player, players[*], or move)", s, kindStr)
	}

	return parsedLegalPath{kind: kind, prop: prop}, nil
}

// validateLegalPath parses p and checks that its named property actually
// exists on the appropriate reader: exampleState's game state reader for
// "game.X" paths, an example player state reader for "player.X" and
// "players[*].X" paths, or moveReader for "move.X" paths. moveReader may be
// nil for non-move paths, but validating a "move.X" path with a nil
// moveReader is itself an error.
//
// This is intended to be run once at boot (NewGameManager), against a
// manager's ExampleState() and an example move's Reader(), so that a typo in
// a declared path fails at boot naming the offending path and property,
// rather than surfacing mid-game.
func validateLegalPath(p LegalPropPath, exampleState ImmutableState, moveReader PropertyReader) error {
	parsed, err := parseLegalPath(p)
	if err != nil {
		return err
	}

	switch parsed.kind {
	case pathGame:
		if exampleState == nil {
			return fmt.Errorf("boardgame: legal path %q: no example state provided to validate against", p)
		}
		return validatePropOnReader(p, parsed.prop, exampleState.ImmutableGameState().Reader())
	case pathPlayer, pathPlayersAll:
		if exampleState == nil {
			return fmt.Errorf("boardgame: legal path %q: no example state provided to validate against", p)
		}
		players := exampleState.ImmutablePlayerStates()
		if len(players) == 0 {
			return fmt.Errorf("boardgame: legal path %q: example state has no player states to validate against", p)
		}
		return validatePropOnReader(p, parsed.prop, players[0].Reader())
	case pathMove:
		if moveReader == nil {
			return fmt.Errorf("boardgame: legal path %q: is a move.* path but no move reader was provided to validate against", p)
		}
		return validatePropOnReader(p, parsed.prop, moveReader)
	}

	return fmt.Errorf("boardgame: legal path %q: unknown path kind", p)
}

// validatePropOnReader checks that propName exists in reader.Props(),
// returning an error naming both the full path p and the missing property
// if not.
func validatePropOnReader(p LegalPropPath, propName string, reader PropertyReader) error {
	if reader == nil {
		return fmt.Errorf("boardgame: legal path %q: reader was nil", p)
	}
	if _, ok := reader.Props()[propName]; !ok {
		return fmt.Errorf("boardgame: legal path %q: property %q does not exist", p, propName)
	}
	return nil
}

// resolveLegalPath resolves p at evaluation time against state and, for
// "move.X" paths, move. It returns the raw property value, its
// PropertyType, and an error if resolution failed.
//
// NOTE on signature: the design spec's evaluation Context type
// (boardgame.LegalContext, wrapping State/Move/Proposer/Chest) does not
// exist yet — it is introduced in a later task. This task's brief was
// controller-approved to instead take (LegalPropPath, ImmutableState, Move)
// directly; a future task adapts callers once LegalContext lands.
//
// "player.X" paths resolve against state.ImmutableCurrentPlayer(), which
// already guards CurrentPlayerIndex() against being out of bounds or one of
// the special negative indices (ObserverPlayerIndex, AdminPlayerIndex,
// AnyPlayerIndex): in every one of those cases it returns nil rather than
// panicking, and this function turns that nil into an error. move may be
// nil for non-move paths; resolving a "move.X" path with a nil move is
// itself an error. "players[*].X" paths are quantifier-only (spec §1) and
// cannot be resolved to a single value here; resolving one is an error.
func resolveLegalPath(p LegalPropPath, state ImmutableState, move Move) (interface{}, PropertyType, error) {
	parsed, err := parseLegalPath(p)
	if err != nil {
		return nil, TypeIllegal, err
	}

	if state == nil {
		return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: state was nil", p)
	}

	switch parsed.kind {
	case pathGame:
		return resolveProp(p, parsed.prop, state.ImmutableGameState().Reader())
	case pathPlayer:
		current := state.ImmutableCurrentPlayer()
		if current == nil {
			return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: current player index is not a valid concrete player (may be out of bounds, Observer, Admin, or Any)", p)
		}
		return resolveProp(p, parsed.prop, current.Reader())
	case pathPlayersAll:
		return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: players[*] paths are quantifier-only and cannot be resolved to a single value", p)
	case pathMove:
		if move == nil {
			return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: is a move.* path but no move was provided to resolve against", p)
		}
		return resolveProp(p, parsed.prop, move.Reader())
	}

	return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: unknown path kind", p)
}

// resolveProp fetches propName's value and PropertyType from reader,
// returning an error naming both the full path p and the missing property
// if propName doesn't exist, or if reading it fails.
func resolveProp(p LegalPropPath, propName string, reader PropertyReader) (interface{}, PropertyType, error) {
	if reader == nil {
		return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: reader was nil", p)
	}

	propType, ok := reader.Props()[propName]
	if !ok {
		return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: property %q does not exist", p, propName)
	}

	val, err := reader.Prop(propName)
	if err != nil {
		return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: error reading property %q: %w", p, propName, err)
	}

	return val, propType, nil
}

// facetSurvives reports whether reading facet from a property sanitized
// with policy yields trustworthy data, per the design spec §6 table. This is
// what makes client evaluability precise under sanitization: a predicate
// that only needs FacetCount from a stack stays evaluable under PolicyLen,
// even though PolicyLen hides the stack's actual values.
//
// Truth table (spec §6):
//   - LegalFacetValues survives PolicyVisible only.
//   - LegalFacetCount survives PolicyVisible, PolicyOrder, and PolicyLen.
//   - LegalFacetOccupancy survives PolicyVisible and PolicyOrder.
//   - LegalFacetOrder survives PolicyVisible and PolicyOrder.
//   - Nothing survives PolicyHidden or PolicyNonEmpty.
//
// PolicyNonEmpty is a deliberate conservative call, not a spec oversight:
// PolicyNonEmpty only reveals whether a group has zero or more-than-zero
// items (sanitization.go's "0 components or a single generic component"),
// which is strictly less information than a count. Treating FacetCount as
// surviving PolicyNonEmpty would let a predicate report e.g. "3 cards left"
// to a viewer who is only supposed to be able to tell "empty or not," so
// FacetCount (like every other facet) does not survive PolicyNonEmpty here.
func facetSurvives(policy Policy, facet LegalFacet) bool {
	switch facet {
	case LegalFacetValues:
		return policy == PolicyVisible
	case LegalFacetCount:
		return policy == PolicyVisible || policy == PolicyOrder || policy == PolicyLen
	case LegalFacetOccupancy:
		return policy == PolicyVisible || policy == PolicyOrder
	case LegalFacetOrder:
		return policy == PolicyVisible || policy == PolicyOrder
	default:
		return false
	}
}
