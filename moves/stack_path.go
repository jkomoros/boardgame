package moves

import (
	"fmt"
	"strings"

	"github.com/jkomoros/boardgame"
)

/*
Stack property specs -- the strings passed to [WithSourceProperty],
[WithDestinationProperty], [WithGameProperty] and [WithPlayerProperty] -- may be
qualified with a path kind, using the same vocabulary as the [legal] package's
path grammar:

	"DrawStack"                            gameState (the historical spelling)
	"game.DrawStack"                       gameState
	"player.Hand"                          the CURRENT player's state
	"players[move.TargetPlayerIndex].Hand" the player named by a move field

An unqualified name still means gameState, so every configuration written before
the grammar existed keeps its exact meaning.

The grammar is what lets the {Move,Deal,Collect}CountComponents family reach
player state at either end. Before it, MoveCountComponents read both of its
stacks off gameState only, which is why "play a card" and "draw a card" -- the
two most common verbs in the medium -- had no reusable move and were hand-rolled
in four of seven example games.

Every spec is parsed and resolved against the example state at
NewGameManager time (see [MoveCountComponents.ValidConfiguration]), so an
unknown path kind or a misspelled property is a boot error naming the move and
the path, never a nil stack discovered mid-game.
*/

type stackPathKind int

const (
	//stackPathGame is "game.X", or a bare "X".
	stackPathGame stackPathKind = iota
	//stackPathCurrentPlayer is "player.X".
	stackPathCurrentPlayer
	//stackPathMoveField is "players[move.<Field>].X".
	stackPathMoveField
)

type parsedStackPath struct {
	kind            stackPathKind
	prop            string
	moveField       string
	boardIndexField string
	//raw is the spec exactly as configured, for error messages.
	raw string
}

// qualified reports whether the spec named a path kind explicitly. An
// unqualified spec is a plain gameState property name, which is the only shape
// the framework's other readers of these configuration keys (for example
// boardgame.LegalStackConstraintsCheck) understand.
func (p parsedStackPath) qualified() bool {
	return p.raw != p.prop || p.boardIndexField != ""
}

func parseStackPropertyExpression(spec, expression string) (string, string, error) {
	open := strings.Index(expression, "[")
	if open == -1 {
		if strings.Contains(expression, "]") {
			return "", "", fmt.Errorf("invalid stack property spec %q: unexpected closing bracket", spec)
		}
		return expression, "", nil
	}
	if open == 0 || !strings.HasSuffix(expression, "]") || strings.Count(expression, "[") != 1 || strings.Count(expression, "]") != 1 {
		return "", "", fmt.Errorf("invalid stack property spec %q: expected Property[move.Field]", spec)
	}
	selector := expression[open+1 : len(expression)-1]
	field, ok := strings.CutPrefix(selector, "move.")
	if !ok || field == "" || strings.ContainsAny(field, ".[]") {
		return "", "", fmt.Errorf("invalid stack property spec %q: board selector must be move.Field", spec)
	}
	return expression[:open], field, nil
}

// parseStackPath parses spec per the grammar documented at the top of this
// file. It checks shape only; whether the named property actually exists is
// checked when the path is resolved against a real state.
func parseStackPath(spec string) (parsedStackPath, error) {

	if spec == "" {
		return parsedStackPath{}, fmt.Errorf("stack property spec was empty")
	}

	if strings.HasPrefix(spec, "players[move.") {
		rest := strings.TrimPrefix(spec, "players[move.")
		closeIndex := strings.Index(rest, "]")
		if closeIndex == -1 {
			return parsedStackPath{}, fmt.Errorf("invalid stack property spec %q: missing closing \"]\"", spec)
		}
		field := rest[:closeIndex]
		if field == "" {
			return parsedStackPath{}, fmt.Errorf("invalid stack property spec %q: missing move field name inside players[move.<Field>]", spec)
		}
		propExpression, ok := strings.CutPrefix(rest[closeIndex+1:], ".")
		if !ok || propExpression == "" {
			return parsedStackPath{}, fmt.Errorf("invalid stack property spec %q: missing property name after players[move.%s]", spec, field)
		}
		prop, boardIndexField, err := parseStackPropertyExpression(spec, propExpression)
		if err != nil {
			return parsedStackPath{}, err
		}
		return parsedStackPath{kind: stackPathMoveField, prop: prop, moveField: field, boardIndexField: boardIndexField, raw: spec}, nil
	}

	kindStr, prop, ok := strings.Cut(spec, ".")

	if !ok {
		//An unqualified name is a gameState property, the historical spelling.
		prop, boardIndexField, err := parseStackPropertyExpression(spec, spec)
		if err != nil {
			return parsedStackPath{}, err
		}
		return parsedStackPath{kind: stackPathGame, prop: prop, boardIndexField: boardIndexField, raw: spec}, nil
	}

	if prop == "" {
		return parsedStackPath{}, fmt.Errorf("invalid stack property spec %q: missing property name after %q", spec, kindStr)
	}

	switch kindStr {
	case "game":
		kind := stackPathGame
		base, boardIndexField, err := parseStackPropertyExpression(spec, prop)
		if err != nil {
			return parsedStackPath{}, err
		}
		return parsedStackPath{kind: kind, prop: base, boardIndexField: boardIndexField, raw: spec}, nil
	case "player":
		base, boardIndexField, err := parseStackPropertyExpression(spec, prop)
		if err != nil {
			return parsedStackPath{}, err
		}
		return parsedStackPath{kind: stackPathCurrentPlayer, prop: base, boardIndexField: boardIndexField, raw: spec}, nil
	}

	return parsedStackPath{}, fmt.Errorf("invalid stack property spec %q: unknown path kind %q (expected game, player, or players[move.Field])", spec, kindStr)
}

// plainGameStackName reports whether spec names a plain gameState stack, and
// returns the bare property name if so. Callers that can only read gameState --
// boardgame.LegalStackConstraintsCheck and the legal package's
// "stackConstraints" predicate both do -- use it to tell "I can express this"
// from "this names a player-scoped stack I have no vocabulary for".
func plainGameStackName(spec string) (string, bool) {
	path, err := parseStackPath(spec)
	if err != nil || path.kind != stackPathGame || path.boardIndexField != "" {
		return "", false
	}
	return path.prop, true
}

// resolveStackPath resolves a parsed spec to a live stack. move is consulted
// only for the players[move.<Field>] kind.
func (p parsedStackPath) resolve(move boardgame.Move, state boardgame.State) (boardgame.Stack, error) {

	var reader boardgame.PropertyReadSetter

	switch p.kind {
	case stackPathGame:
		reader = state.GameState().ReadSetter()
	case stackPathCurrentPlayer:
		playerState := state.CurrentPlayer()
		if playerState == nil {
			return nil, fmt.Errorf("stack property %q: the game has no valid current player", p.raw)
		}
		reader = playerState.ReadSetter()
	case stackPathMoveField:
		playerState, err := p.moveFieldPlayerState(move, state)
		if err != nil {
			return nil, err
		}
		reader = playerState.ReadSetter()
	}

	if p.boardIndexField != "" {
		board, err := reader.BoardProp(p.prop)
		if err != nil {
			return nil, fmt.Errorf("stack property %q: %w", p.raw, err)
		}
		if board == nil {
			return nil, fmt.Errorf("stack property %q resolved to a nil board", p.raw)
		}
		return p.resolveBoardSpace(move, board)
	}

	stack, err := reader.StackProp(p.prop)

	if err != nil {
		return nil, fmt.Errorf("stack property %q: %w", p.raw, err)
	}

	if stack == nil {
		return nil, fmt.Errorf("stack property %q resolved to a nil stack", p.raw)
	}

	return stack, nil
}

func (p parsedStackPath) resolveBoardSpace(move boardgame.Move, board boardgame.Board) (boardgame.Stack, error) {
	if move == nil {
		return nil, fmt.Errorf("stack property %q: no move to read board selector %q from", p.raw, p.boardIndexField)
	}
	moveReader := move.ReadSetter()
	propType, ok := moveReader.Props()[p.boardIndexField]
	if !ok {
		return nil, fmt.Errorf("stack property %q: move field %q does not exist", p.raw, p.boardIndexField)
	}
	var stack boardgame.Stack
	switch propType {
	case boardgame.TypeInt:
		index, err := moveReader.IntProp(p.boardIndexField)
		if err != nil {
			return nil, fmt.Errorf("stack property %q: could not read move field %q: %w", p.raw, p.boardIndexField, err)
		}
		stack = board.SpaceAt(index)
	case boardgame.TypeEnum:
		value, err := moveReader.ImmutableEnumProp(p.boardIndexField)
		if err != nil || value == nil {
			return nil, fmt.Errorf("stack property %q: could not read enum move field %q: %v", p.raw, p.boardIndexField, err)
		}
		if board.Enum() == nil || board.Enum() != value.Enum() {
			return nil, fmt.Errorf("stack property %q: enum move field %q does not match board enum", p.raw, p.boardIndexField)
		}
		stack = board.SpaceAtKey(value.Value())
	default:
		return nil, fmt.Errorf("stack property %q: move field %q has PropertyType %v, expected TypeInt or TypeEnum", p.raw, p.boardIndexField, propType)
	}
	if stack == nil {
		return nil, fmt.Errorf("stack property %q: board selector %q was out of range", p.raw, p.boardIndexField)
	}
	return stack, nil
}

// moveFieldPlayerState resolves the players[move.<Field>] kind: the player
// state the move's own PlayerIndex field names.
//
// THIS MUST AGREE WITH CORE, index for index. The same path is resolved twice
// per move by two different resolvers -- at Legal() time by core's path grammar
// (resolveLegalPlayerReader in legal_path.go, which the contributed
// legal.MayMoveToSlot / legal.MayMoveFirstToSlot atoms go through) and at
// Apply() time by this one -- so any disagreement means Legal() judged one
// player's stacks while Apply() moved another player's components. Core's rule
// is "a concrete, in-bounds player, or an error", and this is that rule.
//
// Deliberately no EnsureValid, for the same reason SeatPlayer, ActivateEmptySeat
// and CloseEmptySeat say so in their own comments. EnsureValid advances past an
// index that is not Valid() -- out of bounds, or a player the delegate says may
// not be active -- onto the NEXT player who may be, so a move naming a player
// who is not there silently became a move against whoever happened to be next.
// Core rejects that index outright; before this, moves retargeted it, and on a
// four-player game a TargetPlayerIndex of 9 resolved to player 0's hand.
func (p parsedStackPath) moveFieldPlayerState(move boardgame.Move, state boardgame.State) (boardgame.SubState, error) {

	if move == nil {
		return nil, fmt.Errorf("stack property %q: no move to read %q from", p.raw, p.moveField)
	}

	index, err := move.ReadSetter().PlayerIndexProp(p.moveField)

	if err != nil {
		return nil, fmt.Errorf("stack property %q: move has no PlayerIndex property %q: %w", p.raw, p.moveField, err)
	}

	players := state.PlayerStates()

	if index < 0 || int(index) >= len(players) {
		return nil, fmt.Errorf("stack property %q: move field %q was %d, which is not a concrete player (Observer, Admin, Any, or out of bounds)", p.raw, p.moveField, index)
	}

	return players[index], nil
}

// configuredStackPath reads configPropName off the move's configuration and
// parses it. configured is false when the option was never passed.
func configuredStackPath(move moveInfoer, configPropName string) (path parsedStackPath, configured bool, err error) {

	val, ok := move.CustomConfiguration()[configPropName]

	if !ok {
		return parsedStackPath{}, false, nil
	}

	strVal, ok := val.(string)

	if !ok {
		return parsedStackPath{}, true, fmt.Errorf("stack property configuration was not a string")
	}

	parsed, err := parseStackPath(strVal)

	if err != nil {
		return parsedStackPath{}, true, err
	}

	return parsed, true, nil
}

// resolveConfiguredStack is the one-call form used by the stack getters on the
// component-moving moves: parse the configured spec and resolve it, returning
// nil for any failure (the getters' documented contract). Boot-time validation
// calls the parts separately so it can report why.
func resolveConfiguredStack(move boardgame.Move, configPropName string, state boardgame.State) boardgame.Stack {

	infoer, ok := move.(moveInfoer)

	if !ok {
		return nil
	}

	path, configured, err := configuredStackPath(infoer, configPropName)

	if !configured || err != nil {
		return nil
	}

	stack, err := path.resolve(move, state)

	if err != nil {
		return nil
	}

	return stack
}

// validateConfiguredStack is the boot-time counterpart of
// resolveConfiguredStack: it reports WHY a configured spec cannot be resolved,
// so a typo fails NewGameManager naming the move and the path rather than
// surfacing as a bare "returned nil" much later.
func validateConfiguredStack(move boardgame.Move, configPropName, optionName string, state boardgame.State) error {

	infoer, ok := move.(moveInfoer)

	if !ok {
		return fmt.Errorf("move did not expose its custom configuration")
	}

	path, configured, err := configuredStackPath(infoer, configPropName)

	if err != nil {
		return fmt.Errorf("%s: %w", optionName, err)
	}

	if !configured {
		//Not being configured is legitimate: the embedding move may override
		//the stack getter instead. The caller checks the getter separately.
		return nil
	}

	if _, err := path.resolve(move, state); err != nil {
		return fmt.Errorf("%s: %w", optionName, err)
	}

	return nil
}
