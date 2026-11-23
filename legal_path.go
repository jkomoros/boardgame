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
	// pathPlayer denotes a "player.X" path, resolved against the move's
	// player: the player the move is made for. See legalMovePlayerIndex.
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
//
// A component path (legal_component_path.go) additionally carries a
// component selector and the property read from the selected component:
// "game.Stack[selector].Prop" or "game.Board[move.Space][selector].Prop".
// prop is still the owning stack (or board) property; component and
// componentProp describe the one component hop.
type parsedLegalPath struct {
	kind            legalPathKind
	prop            string
	moveField       string
	boardIndexField string
	component       legalComponentSelector
	componentProp   string
}

// readsMove reports whether resolving this path reads the move: a move.*
// path, a players[move.F] owner, a [move.F] board index, or a [move.F]
// component selector. Any such path is field-dependent.
func (p parsedLegalPath) readsMove() bool {
	return p.kind == pathMove || p.kind == pathPlayersMoveField || p.boardIndexField != "" || p.component.kind == legalSelectMoveField
}

// hasComponent reports whether this is a component path.
func (p parsedLegalPath) hasComponent() bool {
	return p.component.kind != legalSelectNone
}

// legalPropertyExpression is the parsed form of the part of a path after its
// KIND: a property name, optionally followed by a board index, a component
// selector and a component property.
type legalPropertyExpression struct {
	prop            string
	boardIndexField string
	component       legalComponentSelector
	componentProp   string
}

// parseLegalPropertyExpression parses "Prop", "Board[move.F]",
// "Stack[selector].Prop" or "Board[move.F][selector].Prop". The shape alone
// decides what each bracket means: a single bracket with no suffix is a
// board index (which yields a stack); a bracket followed by ".Prop" is a
// component selector. A component is not a property type, so a component
// selector must be followed by a property.
func parseLegalPropertyExpression(path, expression string) (legalPropertyExpression, error) {
	open := strings.Index(expression, "[")
	if open == -1 {
		if strings.Contains(expression, "]") {
			return legalPropertyExpression{}, fmt.Errorf("boardgame: invalid legal path %q: unexpected closing bracket", path)
		}
		return legalPropertyExpression{prop: expression}, nil
	}
	if open == 0 {
		return legalPropertyExpression{}, fmt.Errorf("boardgame: invalid legal path %q: expected Property[selector]", path)
	}
	result := legalPropertyExpression{prop: expression[:open]}
	rest := expression[open:]
	var brackets []string
	for strings.HasPrefix(rest, "[") {
		closeIdx := strings.Index(rest, "]")
		if closeIdx == -1 {
			return legalPropertyExpression{}, fmt.Errorf("boardgame: invalid legal path %q: missing closing bracket", path)
		}
		content := rest[1:closeIdx]
		if strings.Contains(content, "[") {
			return legalPropertyExpression{}, fmt.Errorf("boardgame: invalid legal path %q: nested brackets", path)
		}
		brackets = append(brackets, content)
		rest = rest[closeIdx+1:]
	}
	if rest != "" {
		suffix, ok := strings.CutPrefix(rest, ".")
		if !ok || suffix == "" || strings.ContainsAny(suffix, ".[]") {
			return legalPropertyExpression{}, fmt.Errorf("boardgame: invalid legal path %q: expected .Property after a component selector (one component hop only)", path)
		}
		result.componentProp = suffix
	}
	boardField := func(content string) (string, error) {
		field, ok := strings.CutPrefix(content, "move.")
		if !ok || field == "" || strings.ContainsAny(field, ".[]") {
			return "", fmt.Errorf("boardgame: invalid legal path %q: board selector must be move.Field", path)
		}
		return field, nil
	}
	var err error
	switch {
	case len(brackets) == 1 && result.componentProp == "":
		result.boardIndexField, err = boardField(brackets[0])
	case len(brackets) == 1:
		result.component, err = parseLegalComponentSelector(path, brackets[0])
	case len(brackets) == 2 && result.componentProp != "":
		if result.boardIndexField, err = boardField(brackets[0]); err == nil {
			result.component, err = parseLegalComponentSelector(path, brackets[1])
		}
	default:
		err = fmt.Errorf("boardgame: invalid legal path %q: expected Board[move.Field], Stack[selector].Property, or Board[move.Field][selector].Property", path)
	}
	if err != nil {
		return legalPropertyExpression{}, err
	}
	return result, nil
}

// legalPlayerIndexSource is the single internal algebra behind the public
// legal.Player(), legal.Proposer(), and legal.PlayerFromMove helpers.
// The path grammar retains its stable wire spellings, but all three resolve
// through resolveLegalPlayerReader below.
type legalPlayerIndexSource int

const (
	legalPlayerFromCurrent legalPlayerIndexSource = iota
	legalPlayerFromProposer
	legalPlayerFromMoveField
)

// parseLegalPath parses p according to the path grammar in spec §1 and §3:
// "game.X", "player.X" (the move's player), "proposer.X", "players[*].X"
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

	expr, err := parseLegalPropertyExpression(s, prop)
	if err != nil {
		return parsedLegalPath{}, err
	}
	if kind == pathMove && (expr.boardIndexField != "" || expr.component.kind != legalSelectNone) {
		return parsedLegalPath{}, fmt.Errorf("boardgame: invalid legal path %q: indexed boards and component selectors must be game or player state properties", s)
	}
	if kind == pathPlayersAll && expr.component.kind != legalSelectNone {
		return parsedLegalPath{}, fmt.Errorf("boardgame: invalid legal path %q: component selectors are not supported on quantifier-only players[*] paths", s)
	}

	return parsedLegalPath{kind: kind, prop: expr.prop, boardIndexField: expr.boardIndexField, component: expr.component, componentProp: expr.componentProp}, nil
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
	propExpression, ok := strings.CutPrefix(afterBracket, ".")
	if !ok || propExpression == "" {
		return parsedLegalPath{}, fmt.Errorf("boardgame: invalid legal path %q: missing property name after players[move.%s]", s, field)
	}
	expr, err := parseLegalPropertyExpression(s, propExpression)
	if err != nil {
		return parsedLegalPath{}, err
	}

	return parsedLegalPath{kind: pathPlayersMoveField, moveField: field, prop: expr.prop, boardIndexField: expr.boardIndexField, component: expr.component, componentProp: expr.componentProp}, nil
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

	var reader PropertyReader
	switch parsed.kind {
	case pathGame:
		if exampleState == nil {
			return fmt.Errorf("boardgame: legal path %q: no example state provided to validate against", p)
		}
		reader = exampleState.ImmutableGameState().Reader()
	case pathPlayer, pathProposer, pathPlayersAll:
		if exampleState == nil {
			return fmt.Errorf("boardgame: legal path %q: no example state provided to validate against", p)
		}
		players := exampleState.ImmutablePlayerStates()
		if len(players) == 0 {
			return fmt.Errorf("boardgame: legal path %q: example state has no player states to validate against", p)
		}
		reader = players[0].Reader()
	case pathMove:
		if moveReader == nil {
			return fmt.Errorf("boardgame: legal path %q: is a move.* path but no move reader was provided to validate against", p)
		}
		reader = moveReader
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
		reader = players[0].Reader()
	}
	if reader == nil {
		return fmt.Errorf("boardgame: legal path %q: unknown path kind", p)
	}
	if err := validatePropOnReader(p, parsed.prop, reader); err != nil {
		return err
	}
	if parsed.boardIndexField != "" {
		if err := validateLegalBoardIndex(p, parsed, reader, moveReader); err != nil {
			return err
		}
	}
	if parsed.hasComponent() {
		_, err := legalComponentPathType(p, parsed, exampleState, moveReader)
		return err
	}
	return nil
}

// validateLegalBoardIndex validates the [move.Field] board index of parsed
// against the owning reader and the move reader.
func validateLegalBoardIndex(p LegalPropPath, parsed parsedLegalPath, reader PropertyReader, moveReader PropertyReader) error {
	if reader.Props()[parsed.prop] != TypeBoard {
		return fmt.Errorf("boardgame: legal path %q: property %q has PropertyType %v, expected TypeBoard before [move.%s]", p, parsed.prop, reader.Props()[parsed.prop], parsed.boardIndexField)
	}
	if moveReader == nil {
		return fmt.Errorf("boardgame: legal path %q: indexed board requires a move reader", p)
	}
	selectorType, ok := moveReader.Props()[parsed.boardIndexField]
	if !ok {
		return fmt.Errorf("boardgame: legal path %q: move field %q does not exist", p, parsed.boardIndexField)
	}
	if selectorType != TypeInt && selectorType != TypeEnum {
		return fmt.Errorf("boardgame: legal path %q: move field %q has PropertyType %v, expected TypeInt or TypeEnum", p, parsed.boardIndexField, selectorType)
	}
	board, err := reader.ImmutableBoardProp(parsed.prop)
	if err != nil || board == nil {
		return fmt.Errorf("boardgame: legal path %q: board property %q was unavailable: %v", p, parsed.prop, err)
	}
	if selectorType == TypeEnum {
		selector, err := moveReader.ImmutableEnumProp(parsed.boardIndexField)
		if err != nil || selector == nil {
			return fmt.Errorf("boardgame: legal path %q: enum move field %q was unavailable: %v", p, parsed.boardIndexField, err)
		}
		if board.Enum() == nil || board.Enum() != selector.Enum() {
			return fmt.Errorf("boardgame: legal path %q: enum move field %q does not match board enum", p, parsed.boardIndexField)
		}
	}
	return nil
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
	if parsed.boardIndexField != "" {
		actual = TypeStack
	}
	if parsed.hasComponent() {
		actual, err = legalComponentPathType(p, parsed, exampleState, moveReader)
		if err != nil {
			return err
		}
	}
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
// "player.X" paths resolve against the move's player (legalMovePlayerIndex):
// the player the move is made for. An index that is not a concrete player
// (ObserverPlayerIndex, out of bounds) is an error rather than a panic. move may be
// nil for non-move paths; resolving a "move.X" path with a nil move is
// itself an error. "players[*].X" paths are quantifier-only (spec §1) and
// cannot be resolved to a single value here; resolving one is an error.
func resolveLegalPath(p LegalPropPath, state ImmutableState, move Move) (interface{}, PropertyType, error) {
	return resolveLegalPathForProposer(p, state, move, ObserverPlayerIndex)
}

func resolveLegalPathForProposer(p LegalPropPath, state ImmutableState, move Move, proposer PlayerIndex) (interface{}, PropertyType, error) {
	movePlayer := ObserverPlayerIndex
	// Only "player.X" names the move's player; resolving it for any other
	// path would call the delegate's CurrentPlayerIndex for nothing.
	if parsed, err := parseLegalPath(p); err == nil && parsed.kind == pathPlayer && state != nil {
		movePlayer, _ = legalMovePlayerIndex(state, move, proposer)
	}
	return resolveLegalPathForMovePlayer(p, state, move, proposer, movePlayer)
}

// resolveLegalPathForMovePlayer is resolveLegalPath with the move's player
// (what "player.X" names) already resolved by the caller.
func resolveLegalPathForMovePlayer(p LegalPropPath, state ImmutableState, move Move, proposer PlayerIndex, movePlayer PlayerIndex) (interface{}, PropertyType, error) {
	parsed, err := parseLegalPath(p)
	if err != nil {
		return nil, TypeIllegal, err
	}

	if state == nil {
		return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: state was nil", p)
	}

	switch parsed.kind {
	case pathGame:
		return resolveParsedLegalProp(p, parsed, state.ImmutableGameState().Reader(), move)
	case pathPlayer:
		reader, err := resolveLegalPlayerReader(p, legalPlayerFromCurrent, "", state, move, movePlayer)
		if err != nil {
			return nil, TypeIllegal, err
		}
		return resolveParsedLegalProp(p, parsed, reader, move)
	case pathProposer:
		reader, err := resolveLegalPlayerReader(p, legalPlayerFromProposer, "", state, move, proposer)
		if err != nil {
			return nil, TypeIllegal, err
		}
		return resolveParsedLegalProp(p, parsed, reader, move)
	case pathPlayersAll:
		return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: players[*] paths are quantifier-only and cannot be resolved to a single value", p)
	case pathMove:
		if move == nil {
			return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: is a move.* path but no move was provided to resolve against", p)
		}
		return resolveParsedLegalProp(p, parsed, move.Reader(), move)
	case pathPlayersMoveField:
		reader, err := resolveLegalPlayerReader(p, legalPlayerFromMoveField, parsed.moveField, state, move, proposer)
		if err != nil {
			return nil, TypeIllegal, err
		}
		return resolveParsedLegalProp(p, parsed, reader, move)
	}

	return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: unknown path kind", p)
}

func resolveParsedLegalProp(path LegalPropPath, parsed parsedLegalPath, reader PropertyReader, move Move) (interface{}, PropertyType, error) {
	if parsed.boardIndexField == "" && !parsed.hasComponent() {
		return resolveProp(path, parsed.prop, reader)
	}
	var stack ImmutableStack
	if parsed.boardIndexField != "" {
		space, err := resolveLegalBoardSpace(path, parsed, reader, move)
		if err != nil {
			return nil, TypeIllegal, err
		}
		stack = space
	} else {
		val, propType, err := resolveProp(path, parsed.prop, reader)
		if err != nil {
			return nil, TypeIllegal, err
		}
		if propType != TypeStack {
			return nil, TypeIllegal, fmt.Errorf("boardgame: legal path %q: property %q has PropertyType %v, expected TypeStack before a component selector", path, parsed.prop, propType)
		}
		stack, _ = val.(ImmutableStack)
	}
	if !parsed.hasComponent() {
		return stack, TypeStack, nil
	}
	return resolveLegalComponentProp(path, parsed, stack, move)
}

// resolveLegalBoardSpace resolves the board space named by parsed's
// [move.Field] board index.
func resolveLegalBoardSpace(path LegalPropPath, parsed parsedLegalPath, reader PropertyReader, move Move) (ImmutableStack, error) {
	if move == nil {
		return nil, fmt.Errorf("boardgame: legal path %q: indexed board requires a move", path)
	}
	board, err := reader.ImmutableBoardProp(parsed.prop)
	if err != nil || board == nil {
		return nil, fmt.Errorf("boardgame: legal path %q: board property %q was unavailable: %v", path, parsed.prop, err)
	}
	moveReader := move.Reader()
	selectorType, ok := moveReader.Props()[parsed.boardIndexField]
	if !ok {
		return nil, fmt.Errorf("boardgame: legal path %q: move field %q does not exist", path, parsed.boardIndexField)
	}
	var stack ImmutableStack
	switch selectorType {
	case TypeInt:
		index, err := moveReader.IntProp(parsed.boardIndexField)
		if err != nil {
			return nil, fmt.Errorf("boardgame: legal path %q: could not read move field %q: %w", path, parsed.boardIndexField, err)
		}
		stack = board.ImmutableSpaceAt(index)
	case TypeEnum:
		value, err := moveReader.ImmutableEnumProp(parsed.boardIndexField)
		if err != nil || value == nil {
			return nil, fmt.Errorf("boardgame: legal path %q: could not read enum move field %q: %v", path, parsed.boardIndexField, err)
		}
		if board.Enum() == nil || board.Enum() != value.Enum() {
			return nil, fmt.Errorf("boardgame: legal path %q: enum move field %q does not match board enum", path, parsed.boardIndexField)
		}
		stack = board.ImmutableSpaceAtKey(value.Value())
	default:
		return nil, fmt.Errorf("boardgame: legal path %q: move field %q has PropertyType %v, expected TypeInt or TypeEnum", path, parsed.boardIndexField, selectorType)
	}
	if stack == nil {
		return nil, fmt.Errorf("boardgame: legal path %q: board selector %q was out of range", path, parsed.boardIndexField)
	}
	return stack, nil
}

// LegalOnBehalfOfer is implemented by framework move types that name the
// player they are made for (moves.CurrentPlayer and moves.AnyPlayer, via
// their TargetPlayerIndex) and whose contributed preconditions require that
// player to match the proposer. Game authors do not implement it: embedding
// one of those types is enough.
//
// For such a move, "player.X" names that player rather than the game's
// current player: the proposer, or, when an admin proposes on the player's
// behalf, LegalOnBehalfOf(). The returned index may be a non-concrete
// sentinel; resolution then fails closed.
type LegalOnBehalfOfer interface {
	LegalOnBehalfOf() PlayerIndex
}

// legalPlayerIndexIsWildcard reports whether index is one of the
// PlayerIndex sentinels PlayerIndex.Equivalent treats as matching every
// concrete player.
func legalPlayerIndexIsWildcard(index PlayerIndex) bool {
	return index == AnyPlayerIndex || index == AdminPlayerIndex
}

// legalMovePlayerIndex resolves "the move's player": the player the move is
// made for, which is what "player.X" paths and legal.Player() name.
//
//   - A move that names its own player (LegalOnBehalfOfer: moves.CurrentPlayer,
//     moves.AnyPlayer) is that player's move in every phase: the proposer
//     when it is a concrete player, and for an AdminPlayerIndex proposer the
//     player the admin acts on behalf of. The current player is never
//     consulted, so an AnyPlayer move made out of turn, or during an
//     ObserverPlayerIndex gathering phase, still reads its own player. For a
//     moves.CurrentPlayer move the contributed proposerIsCurrentPlayer check
//     has already required that player to be the current player, so the
//     answer is unchanged wherever that check passes.
//   - Any other move (moves.Default, fix-ups) is the current player's: the
//     game's current player when concrete, or, when the current player is a
//     wildcard (AnyPlayerIndex during simultaneous play, or
//     AdminPlayerIndex), the proposer.
//
// Whatever remains may be a non-concrete index (an Observer proposer, an
// ObserverPlayerIndex current player for a move with no player of its own,
// an admin with nobody to act for), which the caller turns into an error and
// the predicate into a fail-closed Unknown. fromMove reports that the index
// came from the move's on-behalf field, i.e. that the result depends on a
// move field.
func legalMovePlayerIndex(state ImmutableState, move Move, proposer PlayerIndex) (index PlayerIndex, fromMove bool) {
	if onBehalf, ok := move.(LegalOnBehalfOfer); ok {
		if proposer == AdminPlayerIndex {
			return onBehalf.LegalOnBehalfOf(), true
		}
		return proposer, false
	}
	index = state.CurrentPlayerIndex()
	if legalPlayerIndexIsWildcard(index) {
		return proposer, false
	}
	return index, false
}

// resolveLegalPlayerReader returns the player state reader source names.
// selected is the already-resolved index for the two non-field sources: the
// move's player (legalMovePlayerIndex) for legalPlayerFromCurrent, and the
// proposer for legalPlayerFromProposer.
func resolveLegalPlayerReader(path LegalPropPath, source legalPlayerIndexSource, moveField string, state ImmutableState, move Move, selected PlayerIndex) (PropertyReader, error) {
	var index PlayerIndex
	switch source {
	case legalPlayerFromCurrent, legalPlayerFromProposer:
		index = selected
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
//
// While the engine evaluates a predicate (evalLegalPredicate), ResolvePath
// is scoped to that predicate's declared Reads: resolving any other path
// returns an error and makes the predicate's verdict a fail-closed
// LegalUnknown, whatever Evaluate returns afterwards (legal_reads.go).
func (c LegalContext) ResolvePath(p LegalPropPath) (interface{}, PropertyType, error) {
	if err := c.reads.check(p); err != nil {
		return nil, TypeIllegal, err
	}
	if c.movePlayerResolved {
		return resolveLegalPathForMovePlayer(p, c.State, c.Move, c.ProposerPlayerIndex, c.movePlayer)
	}
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
