package boardgame

import (
	"fmt"
	"strings"
)

// legalPathKind identifies which path grammar a parsed
// LegalPropPath uses. See the design spec §1 "Path grammar" and (for
// pathPlayersMoveField) spec §3.
type legalPathKind int

const (
	// pathGame denotes a "game.X" path, resolved against the game state.
	pathGame legalPathKind = iota
	// pathPlayer denotes a "player.X" path, resolved against the current
	// player, i.e. the player state at state.CurrentPlayerIndex(). See
	// resolveLegalPath for how an invalid/Observer/Admin/AnyPlayerIndex
	// current player is handled.
	pathPlayer
	// pathProposer denotes a "proposer.X" path, resolved against the concrete
	// player proposing the move. Unlike pathPlayer it does not depend on the
	// game's current-player state, so it also works during simultaneous play.
	pathProposer
	// pathPlayersAll denotes a "players[*].X" path. This kind is
	// quantifier-only (spec §1): it names a property to be read from every
	// player by a quantified predicate, but it does not itself resolve to a
	// single value. resolveLegalPath returns an error if asked to resolve
	// one directly.
	pathPlayersAll
	// pathMove denotes a "move.X" path, resolved against the move being
	// evaluated.
	pathMove
	// pathPlayersMoveField denotes a "players[move.<Field>].<Prop>" path
	// (spec §3): the playerState of the player whose index is the value of
	// the move's <Field> property. moveField names <Field>; prop names
	// <Prop>. See resolveLegalPath for how an invalid/Observer/Admin/Any
	// field value is handled — the same guard pathPlayer uses for an
	// invalid current player, applied to the field-derived index instead.
	// Predicates reading this path kind are field-dependent by construction
	// (legalReadsIncludeMovePath, legal_predicate.go, treats it as such)
	// since the value they resolve depends on the move's own field.
	pathPlayersMoveField
)

// parsedLegalPath is the result of successfully parsing a LegalPropPath: the
// path's kind, the property name to look up within that kind's reader, and
// (pathPlayersMoveField only) the move field naming which player to index.
type parsedLegalPath struct {
	kind      legalPathKind
	prop      string
	moveField string
}

// legalPlayerIndexSource is the single internal algebra behind the public
// legal.CurrentPlayer(), legal.Proposer(), and legal.PlayerFromMove helpers.
// The path grammar retains its stable wire spellings, but all three resolve
// through resolveLegalPlayerReader below.
type legalPlayerIndexSource int

const (
	legalPlayerFromCurrent legalPlayerIndexSource = iota
	legalPlayerFromProposer
	legalPlayerFromMoveField
)

// parseLegalPath parses p according to the path grammar in spec §1 and §3:
// "game.X", "player.X" (current player), "proposer.X", "players[*].X"
// (quantifier-only), "move.X", or "players[move.<Field>].<Prop>". The KIND segment is matched
// literally and case-sensitively against exactly those spellings
// ("players[0]" or any other concrete index is not valid grammar — only the
// "*" quantifier and a "move."-prefixed index expression are).
//
// parseLegalPath only checks the path's shape; it does not check that the
// named property (or, for pathPlayersMoveField, the named move field)
// actually exists on any reader. That is validateLegalPath's job, run once
// at boot (NewGameManager) against real example readers so a typo fails at
// boot naming the move and path, never mid-game.
func parseLegalPath(p LegalPropPath) (parsedLegalPath, error) {
	s := string(p)

	// "players[move.<Field>].<Prop>" is handled separately from the other
	// four kinds: cutting on the first "." would split inside "move.<Field>"
	// rather than at the KIND/Property boundary, since the KIND segment
	// itself contains a ".".
	if strings.HasPrefix(s, "players[move.") {
		return parsePlayersMoveFieldPath(p, s)
	}

	kindStr, prop, ok := strings.Cut(s, ".")
	if !ok {
		return parsedLegalPath{}, fmt.Errorf("boardgame: invalid legal path %q: expected KIND.Property (KIND is one of game, player, proposer, players[*], move, players[move.Field])", s)
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
	case "proposer":
		kind = pathProposer
	case "players[*]":
		kind = pathPlayersAll
	case "move":
		kind = pathMove
	default:
		return parsedLegalPath{}, fmt.Errorf("boardgame: invalid legal path %q: unknown path kind %q (expected game, player, proposer, players[*], move, or players[move.Field])", s, kindStr)
	}

	return parsedLegalPath{kind: kind, prop: prop}, nil
}

// parsePlayersMoveFieldPath parses the "players[move.<Field>].<Prop>" shape
// of s (already known to start with "players[move."), spec §3. Both <Field>
// (the move property naming which player to index) and <Prop> (the property
// to read from that player) must be non-empty, and the bracket must be
// closed before the "." that introduces <Prop>.
func parsePlayersMoveFieldPath(p LegalPropPath, s string) (parsedLegalPath, error) {
	rest := strings.TrimPrefix(s, "players[move.")

	closeIdx := strings.Index(rest, "]")
	if closeIdx == -1 {
		return parsedLegalPath{}, fmt.Errorf("boardgame: invalid legal path %q: missing closing %q in players[move.<Field>] path", s, "]")
	}

	field := rest[:closeIdx]
	if field == "" {
		return parsedLegalPath{}, fmt.Errorf("boardgame: invalid legal path %q: missing move field name inside players[move.<Field>]", s)
	}

	afterBracket := rest[closeIdx+1:]
	prop, ok := strings.CutPrefix(afterBracket, ".")
	if !ok || prop == "" {
		return parsedLegalPath{}, fmt.Errorf("boardgame: invalid legal path %q: missing property name after players[move.%s]", s, field)
	}

	return parsedLegalPath{kind: pathPlayersMoveField, moveField: field, prop: prop}, nil
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
	case pathPlayer, pathProposer, pathPlayersAll:
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
	case pathPlayersMoveField:
		if moveReader == nil {
			return fmt.Errorf("boardgame: legal path %q: is a players[move.*] path but no move reader was provided to validate against", p)
		}
		fieldType, ok := moveReader.Props()[parsed.moveField]
		if !ok {
			return fmt.Errorf("boardgame: legal path %q: move field %q does not exist", p, parsed.moveField)
		}
		// Footgun-batch F9: only TypePlayerIndex is accepted. An int-typed
		// field is grammatically index-shaped, but nothing marks it as a
		// PLAYER index — a plain int (a score, a count, a slot number) that
		// happens to land in range silently indexes a wrong-but-valid player,
		// with no error to catch it. Every real usage surveyed (moves.
		// CurrentPlayer's TargetPlayerIndex, darwin, valentine, tictactoe)
		// already uses a boardgame.PlayerIndex-typed field, which both
		// documents intent and gets PlayerIndex's own validity semantics.
		if fieldType != TypePlayerIndex {
			return fmt.Errorf("boardgame: legal path %q: move field %q has PropertyType %v, expected TypePlayerIndex — declare the field as boardgame.PlayerIndex (an int-typed field is not accepted: nothing marks a plain int as a player index, so a wrong-but-in-range value would silently read another player's state)", p, parsed.moveField, fieldType)
		}
		if exampleState == nil {
			return fmt.Errorf("boardgame: legal path %q: no example state provided to validate against", p)
		}
		players := exampleState.ImmutablePlayerStates()
		if len(players) == 0 {
			return fmt.Errorf("boardgame: legal path %q: example state has no player states to validate against", p)
		}
		return validatePropOnReader(p, parsed.prop, players[0].Reader())
	}

	return fmt.Errorf("boardgame: legal path %q: unknown path kind", p)
}

func validateLegalPathType(p LegalPropPath, expected PropertyType, exampleState ImmutableState, moveReader PropertyReader) error {
	return validateLegalPathTypes(p, []PropertyType{expected}, exampleState, moveReader)
}

func validateLegalPathTypes(p LegalPropPath, allowed []PropertyType, exampleState ImmutableState, moveReader PropertyReader) error {
	parsed, err := parseLegalPath(p)
	if err != nil {
		return err
	}
	if parsed.kind == pathMove && moveReader == nil {
		return nil
	}
	if parsed.kind == pathPlayersMoveField && (moveReader == nil || exampleState == nil) {
		return nil
	}
	if parsed.kind != pathMove && parsed.kind != pathPlayersMoveField && exampleState == nil {
		return nil
	}
	if err := validateLegalPath(p, exampleState, moveReader); err != nil {
		return err
	}
	var reader PropertyReader
	switch parsed.kind {
	case pathGame:
		reader = exampleState.ImmutableGameState().Reader()
	case pathPlayer, pathProposer, pathPlayersAll, pathPlayersMoveField:
		reader = exampleState.ImmutablePlayerStates()[0].Reader()
	case pathMove:
		reader = moveReader
	}
	actual := reader.Props()[parsed.prop]
	for _, expected := range allowed {
		if actual == expected {
			return nil
		}
	}
	if len(allowed) == 1 {
		return fmt.Errorf("boardgame: legal path %q: property %q has PropertyType %v, expected %v", p, parsed.prop, actual, allowed[0])
	}
	return fmt.Errorf("boardgame: legal path %q: property %q has PropertyType %v, expected one of %v", p, parsed.prop, actual, allowed)
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
// This is no longer dead outside tests: LegalContext.ResolvePath
// (legal_predicate.go) delegates directly here, which is what makes it THE
// path resolution every predicate — core's own future predicates and every
// predicate in package legal's catalog alike — must use. Boot validation
// (validateLegalPath, above) and evaluation (this function, reached via
// LegalContext.ResolvePath) share one grammar by construction: both are
// built on parseLegalPath, and a "kind.Property" spelling that boot
// accepted can never surprise evaluation with a different parse.
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
	return resolveLegalPathForProposer(p, state, move, ObserverPlayerIndex)
}

func resolveLegalPathForProposer(p LegalPropPath, state ImmutableState, move Move, proposer PlayerIndex) (interface{}, PropertyType, error) {
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
		reader, err := resolveLegalPlayerReader(p, legalPlayerFromCurrent, "", state, move, proposer)
		if err != nil {
			return nil, TypeIllegal, err
		}
		return resolveProp(p, parsed.prop, reader)
	case pathProposer:
		reader, err := resolveLegalPlayerReader(p, legalPlayerFromProposer, "", state, move, proposer)
		if err != nil {
			return nil, TypeIllegal, err
		}
		return resolveProp(p, parsed.prop, reader)
	case pathPlayersAll:
		return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: players[*] paths are quantifier-only and cannot be resolved to a single value", p)
	case pathMove:
		if move == nil {
			return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: is a move.* path but no move was provided to resolve against", p)
		}
		return resolveProp(p, parsed.prop, move.Reader())
	case pathPlayersMoveField:
		reader, err := resolveLegalPlayerReader(p, legalPlayerFromMoveField, parsed.moveField, state, move, proposer)
		if err != nil {
			return nil, TypeIllegal, err
		}
		return resolveProp(p, parsed.prop, reader)
	}

	return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: unknown path kind", p)
}

func resolveLegalPlayerReader(path LegalPropPath, source legalPlayerIndexSource, moveField string, state ImmutableState, move Move, proposer PlayerIndex) (PropertyReader, error) {
	var index PlayerIndex
	switch source {
	case legalPlayerFromCurrent:
		index = state.CurrentPlayerIndex()
	case legalPlayerFromProposer:
		index = proposer
	case legalPlayerFromMoveField:
		if move == nil {
			return nil, fmt.Errorf("boardgame: legal path %q: is a players[move.*] path but no move was provided to resolve against", path)
		}
		fieldVal, fieldType, err := resolveProp(path, moveField, move.Reader())
		if err != nil {
			return nil, err
		}
		if fieldType != TypePlayerIndex {
			return nil, fmt.Errorf("boardgame: legal path %q: move field %q has PropertyType %v, expected TypePlayerIndex", path, moveField, fieldType)
		}
		index = fieldVal.(PlayerIndex)
	default:
		return nil, fmt.Errorf("boardgame: legal path %q: unknown player index source", path)
	}

	players := state.ImmutablePlayerStates()
	if index < 0 || int(index) >= len(players) {
		return nil, fmt.Errorf("boardgame: legal path %q: player index source resolved to %d, which is not a valid concrete player (Observer, Admin, Any, or out of bounds)", path, index)
	}
	return players[index].Reader(), nil
}

// ResolvePath resolves p at evaluation time against c.State and, for
// "move.X" paths, c.Move. It returns the raw property value, its
// PropertyType, and an error if resolution failed.
//
// This is THE path resolution a LegalPredicate's Evaluate func must use to
// turn a LegalPropPath into a value: it delegates directly to
// resolveLegalPath, the same grammar validateLegalPath checks paths against
// at boot (NewGameManager). Sharing one implementation between boot
// validation and evaluation is deliberate — see spec §1 "Context is the
// entire vocabulary a predicate may reference" — so a path that validates
// at boot can never be parsed differently at evaluation time, and vice
// versa. Any predicate constructor, whether built into core or registered
// by package legal's catalog (or a game's own LegalPredicateConstructor),
// should call c.ResolvePath rather than hand-rolling its own "kind.Property"
// parsing.
func (c LegalContext) ResolvePath(p LegalPropPath) (interface{}, PropertyType, error) {
	return resolveLegalPathForProposer(p, c.State, c.Move, c.ProposerPlayerIndex)
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
// with policy yields trustworthy data, per the design spec §1 and §6 tables.
// This is what makes client evaluability precise under sanitization: a
// predicate that only needs FacetCount from a stack stays evaluable under
// PolicyLen, even though PolicyLen hides the stack's actual values.
//
// Truth table (spec §1, §6):
//   - LegalFacetValues survives PolicyVisible only.
//   - LegalFacetCount survives PolicyVisible, PolicyOrder, and PolicyLen.
//   - LegalFacetOccupancy survives PolicyVisible and PolicyOrder.
//   - LegalFacetOrder survives PolicyVisible and PolicyOrder.
//   - LegalFacetNonEmpty survives PolicyVisible, PolicyOrder, PolicyLen, and
//     PolicyNonEmpty (everything except PolicyHidden). Rationale: PolicyNonEmpty
//     reveals exactly emptiness; an emptiness predicate must stay
//     client-evaluable under it.
//   - Nothing survives PolicyHidden.
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
	case LegalFacetNonEmpty:
		return policy == PolicyVisible || policy == PolicyOrder || policy == PolicyLen || policy == PolicyNonEmpty
	default:
		return false
	}
}
