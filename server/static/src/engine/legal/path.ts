// The legal path grammar and its resolution: the TypeScript mirror of Go's
// legal_path.go (grammar, boot validation, evaluation) and legal/
// path_resolve.go (the typed accessors every catalog predicate reads through).
//
// Grammar:
//   game.X                         the game state
//   player.X                       the CURRENT player's state
//   proposer.X                     the proposing player's state
//   players[*].X                   quantifier-only; never resolves to a value
//   move.X                         the move being evaluated
//   players[move.<Field>].X        the player a move's PlayerIndex field names
// and, on game and player kinds, an indexed board space:
//   game.Board[move.<Field>]       the stack at that space of a board
//
// Failures are values, not exceptions: every resolver returns Go's error text
// so a predicate can turn it into an Unknown reason unchanged.

import type {
  LegalBoard,
  LegalContext,
  LegalEnumValue,
  LegalMove,
  LegalPropertyReader,
  LegalStack,
  LegalState,
} from './access.ts';
import { goQuote } from './goformat.ts';
import type { PropertyType } from './types.ts';

/** Go: legalPathKind. */
export type LegalPathKind = 'game' | 'player' | 'proposer' | 'playersAll' | 'move' | 'playersMoveField';

/** Go: parsedLegalPath. */
export interface ParsedLegalPath {
  kind: LegalPathKind;
  prop: string;
  /** For playersMoveField: the move field naming which player to index. */
  moveField: string;
  /** For an indexed board path: the move field selecting the space. Otherwise ''. */
  boardIndexField: string;
}

/** A value or Go's error text. */
export type Result<T> = { ok: true; value: T } | { ok: false; error: string };

function ok<T>(value: T): Result<T> {
  return { ok: true, value };
}

function fail<T>(error: string): Result<T> {
  return { ok: false, error };
}

function count(haystack: string, needle: string): number {
  return haystack.split(needle).length - 1;
}

/** Go: parseLegalPropertyExpression. */
function parsePropertyExpression(path: string, expression: string): Result<{ prop: string; boardIndexField: string }> {
  const open = expression.indexOf('[');
  if (open === -1) {
    if (expression.includes(']')) {
      return fail(`boardgame: invalid legal path ${goQuote(path)}: unexpected closing bracket`);
    }
    return ok({ prop: expression, boardIndexField: '' });
  }
  if (open === 0 || !expression.endsWith(']') || count(expression, '[') !== 1 || count(expression, ']') !== 1) {
    return fail(`boardgame: invalid legal path ${goQuote(path)}: expected Property[move.Field]`);
  }
  const selector = expression.slice(open + 1, expression.length - 1);
  const field = selector.startsWith('move.') ? selector.slice('move.'.length) : null;
  if (field === null || field === '' || /[.[\]]/.test(field)) {
    return fail(`boardgame: invalid legal path ${goQuote(path)}: board selector must be move.Field`);
  }
  return ok({ prop: expression.slice(0, open), boardIndexField: field });
}

const KINDS: Readonly<Record<string, LegalPathKind>> = {
  game: 'game',
  player: 'player',
  proposer: 'proposer',
  'players[*]': 'playersAll',
  move: 'move',
};

/**
 * Go: parseLegalPath. Checks the path's shape only; whether the property
 * exists is validateLegalPath's job (boot) or resolution's (evaluation).
 */
export function parseLegalPath(path: string): Result<ParsedLegalPath> {
  if (path.startsWith('players[move.')) {
    return parsePlayersMoveFieldPath(path);
  }
  const dot = path.indexOf('.');
  if (dot === -1) {
    return fail(
      `boardgame: invalid legal path ${goQuote(path)}: expected KIND.Property (KIND is one of game, player, proposer, players[*], move, players[move.Field])`,
    );
  }
  const kindStr = path.slice(0, dot);
  const prop = path.slice(dot + 1);
  if (prop === '') {
    return fail(`boardgame: invalid legal path ${goQuote(path)}: missing property name after ${goQuote(kindStr)}`);
  }
  const kind = Object.prototype.hasOwnProperty.call(KINDS, kindStr) ? KINDS[kindStr] : undefined;
  if (kind === undefined) {
    return fail(
      `boardgame: invalid legal path ${goQuote(path)}: unknown path kind ${goQuote(kindStr)} (expected game, player, proposer, players[*], move, or players[move.Field])`,
    );
  }
  const expression = parsePropertyExpression(path, prop);
  if (!expression.ok) {
    return expression;
  }
  if (kind === 'move' && expression.value.boardIndexField !== '') {
    return fail(`boardgame: invalid legal path ${goQuote(path)}: indexed boards must be game or player state properties`);
  }
  return ok({ kind, prop: expression.value.prop, moveField: '', boardIndexField: expression.value.boardIndexField });
}

/** Go: parsePlayersMoveFieldPath. */
function parsePlayersMoveFieldPath(path: string): Result<ParsedLegalPath> {
  const rest = path.slice('players[move.'.length);
  const closeIdx = rest.indexOf(']');
  if (closeIdx === -1) {
    return fail(`boardgame: invalid legal path ${goQuote(path)}: missing closing ${goQuote(']')} in players[move.<Field>] path`);
  }
  const field = rest.slice(0, closeIdx);
  if (field === '') {
    return fail(`boardgame: invalid legal path ${goQuote(path)}: missing move field name inside players[move.<Field>]`);
  }
  const afterBracket = rest.slice(closeIdx + 1);
  const propExpression = afterBracket.startsWith('.') ? afterBracket.slice(1) : '';
  if (propExpression === '') {
    return fail(`boardgame: invalid legal path ${goQuote(path)}: missing property name after players[move.${field}]`);
  }
  const expression = parsePropertyExpression(path, propExpression);
  if (!expression.ok) {
    return expression;
  }
  return ok({
    kind: 'playersMoveField',
    moveField: field,
    prop: expression.value.prop,
    boardIndexField: expression.value.boardIndexField,
  });
}

/**
 * Go: legalReadsIncludeMovePath's test for one path. A read of move.X or
 * players[move.F].X makes a predicate field-dependent. A malformed path is
 * not a move path.
 */
export function legalPathReadsMove(path: string): boolean {
  const parsed = parseLegalPath(path);
  return parsed.ok && (parsed.value.kind === 'move' || parsed.value.kind === 'playersMoveField');
}

/** A resolved property: its raw value and declared type. */
export interface ResolvedLegalProp {
  value: unknown;
  type: PropertyType;
}

function hasProp(reader: LegalPropertyReader, name: string): boolean {
  return Object.prototype.hasOwnProperty.call(reader.props(), name);
}

function thrownText(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

/** Go: resolveProp. */
function resolveProp(path: string, propName: string, reader: LegalPropertyReader): Result<ResolvedLegalProp> {
  if (!hasProp(reader, propName)) {
    return fail(`boardgame: legal path ${goQuote(path)}: property ${goQuote(propName)} does not exist`);
  }
  const type = reader.props()[propName];
  let value: unknown;
  try {
    value = reader.prop(propName);
  } catch (err) {
    return fail(`boardgame: legal path ${goQuote(path)}: error reading property ${goQuote(propName)}: ${thrownText(err)}`);
  }
  return ok({ value, type });
}

/**
 * Go: reader.ImmutableBoardProp(name), including the generated readers' own
 * error text for a property that is not a board.
 */
function readBoard(reader: LegalPropertyReader, name: string): { board: LegalBoard | null; error: string } {
  if (!hasProp(reader, name) || reader.props()[name] !== 'TypeBoard') {
    return { board: null, error: 'No such Board prop: ' + name };
  }
  try {
    return { board: (reader.prop(name) as LegalBoard | null) ?? null, error: '<nil>' };
  } catch (err) {
    return { board: null, error: thrownText(err) };
  }
}

/** Go: resolveParsedLegalProp. */
function resolveParsedProp(
  path: string,
  parsed: ParsedLegalPath,
  reader: LegalPropertyReader,
  move: LegalMove | null,
): Result<ResolvedLegalProp> {
  if (parsed.boardIndexField === '') {
    return resolveProp(path, parsed.prop, reader);
  }
  if (move === null) {
    return fail(`boardgame: legal path ${goQuote(path)}: indexed board requires a move`);
  }
  const { board, error } = readBoard(reader, parsed.prop);
  if (board === null) {
    return fail(`boardgame: legal path ${goQuote(path)}: board property ${goQuote(parsed.prop)} was unavailable: ${error}`);
  }
  const moveReader = move.reader();
  const field = parsed.boardIndexField;
  if (!hasProp(moveReader, field)) {
    return fail(`boardgame: legal path ${goQuote(path)}: move field ${goQuote(field)} does not exist`);
  }
  const selectorType = moveReader.props()[field];
  let stack: LegalStack | null;
  if (selectorType === 'TypeInt') {
    let index: unknown;
    try {
      index = moveReader.prop(field);
    } catch (err) {
      return fail(`boardgame: legal path ${goQuote(path)}: could not read move field ${goQuote(field)}: ${thrownText(err)}`);
    }
    stack = board.spaceAt(index as number);
  } else if (selectorType === 'TypeEnum') {
    let value: LegalEnumValue | null;
    try {
      value = (moveReader.prop(field) as LegalEnumValue | null) ?? null;
    } catch (err) {
      return fail(`boardgame: legal path ${goQuote(path)}: could not read enum move field ${goQuote(field)}: ${thrownText(err)}`);
    }
    if (value === null) {
      return fail(`boardgame: legal path ${goQuote(path)}: could not read enum move field ${goQuote(field)}: <nil>`);
    }
    const boardEnum = board.enum();
    if (boardEnum === null || boardEnum.name !== value.enum.name) {
      return fail(`boardgame: legal path ${goQuote(path)}: enum move field ${goQuote(field)} does not match board enum`);
    }
    stack = board.spaceAtKey(value.value);
  } else {
    return fail(
      `boardgame: legal path ${goQuote(path)}: move field ${goQuote(field)} has PropertyType ${selectorType}, expected TypeInt or TypeEnum`,
    );
  }
  if (stack === null) {
    return fail(`boardgame: legal path ${goQuote(path)}: board selector ${goQuote(field)} was out of range`);
  }
  return ok({ value: stack, type: 'TypeStack' });
}

/** Go: resolveLegalPlayerReader. */
function resolvePlayerReader(
  path: string,
  source: 'current' | 'proposer' | 'moveField',
  moveField: string,
  state: LegalState,
  move: LegalMove | null,
  proposer: number,
): Result<LegalPropertyReader> {
  let index: number;
  if (source === 'current') {
    index = state.currentPlayerIndex();
  } else if (source === 'proposer') {
    index = proposer;
  } else {
    if (move === null) {
      return fail(`boardgame: legal path ${goQuote(path)}: is a players[move.*] path but no move was provided to resolve against`);
    }
    const field = resolveProp(path, moveField, move.reader());
    if (!field.ok) {
      return field;
    }
    if (field.value.type !== 'TypePlayerIndex') {
      return fail(
        `boardgame: legal path ${goQuote(path)}: move field ${goQuote(moveField)} has PropertyType ${field.value.type}, expected TypePlayerIndex`,
      );
    }
    index = field.value.value as number;
  }
  const players = state.players();
  if (index < 0 || index >= players.length) {
    return fail(
      `boardgame: legal path ${goQuote(path)}: player index source resolved to ${index}, which is not a valid concrete player (Observer, Admin, Any, or out of bounds)`,
    );
  }
  return ok(players[index]);
}

/**
 * Go: LegalContext.ResolvePath (resolveLegalPathForProposer). THE path
 * resolution: every predicate turns a path into a value through this.
 */
export function resolveLegalPath(path: string, ctx: LegalContext): Result<ResolvedLegalProp> {
  const parsedResult = parseLegalPath(path);
  if (!parsedResult.ok) {
    return parsedResult;
  }
  const parsed = parsedResult.value;
  const state = ctx.state;
  if (state === null) {
    return fail(`boardgame: legal path ${goQuote(path)}: state was nil`);
  }
  switch (parsed.kind) {
    case 'game':
      return resolveParsedProp(path, parsed, state.game(), ctx.move);
    case 'player':
    case 'proposer':
    case 'playersMoveField': {
      const source = parsed.kind === 'player' ? 'current' : parsed.kind === 'proposer' ? 'proposer' : 'moveField';
      const reader = resolvePlayerReader(path, source, parsed.moveField, state, ctx.move, ctx.proposer);
      if (!reader.ok) {
        return reader;
      }
      return resolveParsedProp(path, parsed, reader.value, ctx.move);
    }
    case 'playersAll':
      return fail(
        `boardgame: legal path ${goQuote(path)}: players[*] paths are quantifier-only and cannot be resolved to a single value`,
      );
    case 'move':
      if (ctx.move === null) {
        return fail(`boardgame: legal path ${goQuote(path)}: is a move.* path but no move was provided to resolve against`);
      }
      return resolveParsedProp(path, parsed, ctx.move.reader(), ctx.move);
  }
}

/** Go: legal.resolveIntPath. */
export function resolveIntPath(path: string, ctx: LegalContext): Result<number> {
  const resolved = resolveLegalPath(path, ctx);
  if (!resolved.ok) {
    return resolved;
  }
  if (resolved.value.type !== 'TypeInt') {
    return fail(`legal: path ${goQuote(path)} is not an int property`);
  }
  if (typeof resolved.value.value !== 'number') {
    return fail(`legal: path ${goQuote(path)} resolved to an int-typed property but its value was not an int (${typeof resolved.value.value})`);
  }
  return ok(resolved.value.value);
}

/** Go: legal.resolveBoolPath. */
export function resolveBoolPath(path: string, ctx: LegalContext): Result<boolean> {
  const resolved = resolveLegalPath(path, ctx);
  if (!resolved.ok) {
    return resolved;
  }
  if (resolved.value.type !== 'TypeBool') {
    return fail(`legal: path ${goQuote(path)} is not a bool property`);
  }
  if (typeof resolved.value.value !== 'boolean') {
    return fail(`legal: path ${goQuote(path)} resolved to a bool-typed property but its value was not a bool (${typeof resolved.value.value})`);
  }
  return ok(resolved.value.value);
}

/** Go: legal.resolveStackPath. A null value is Go's nil stack. */
export function resolveStackPath(path: string, ctx: LegalContext): Result<LegalStack | null> {
  const resolved = resolveLegalPath(path, ctx);
  if (!resolved.ok) {
    return resolved;
  }
  if (resolved.value.type !== 'TypeStack') {
    return fail(`legal: path ${goQuote(path)} is not a stack property`);
  }
  return ok((resolved.value.value as LegalStack | null) ?? null);
}

/** Go: legal.resolvePlayerIndexPath. */
export function resolvePlayerIndexPath(path: string, ctx: LegalContext): Result<number> {
  const resolved = resolveLegalPath(path, ctx);
  if (!resolved.ok) {
    return resolved;
  }
  if (resolved.value.type !== 'TypePlayerIndex') {
    return fail(`legal: path ${goQuote(path)} is not a PlayerIndex property`);
  }
  if (typeof resolved.value.value !== 'number') {
    return fail(
      `legal: path ${goQuote(path)} resolved to a PlayerIndex-typed property but its value was not a PlayerIndex (${typeof resolved.value.value})`,
    );
  }
  return ok(resolved.value.value);
}

/** Go: legal.resolveEnumPath. A null value is Go's nil enum value. */
export function resolveEnumPath(path: string, ctx: LegalContext): Result<LegalEnumValue | null> {
  const resolved = resolveLegalPath(path, ctx);
  if (!resolved.ok) {
    return resolved;
  }
  if (resolved.value.type !== 'TypeEnum') {
    return fail(`legal: path ${goQuote(path)} is not an enum property`);
  }
  return ok((resolved.value.value as LegalEnumValue | null) ?? null);
}

/**
 * Go: validateLegalPath. Checks, against an example state and an example
 * move's reader, that the path's property (and any move field it indexes by)
 * exists. Returns Go's error text, or null when the path is valid.
 */
export function validateLegalPath(
  path: string,
  exampleState: LegalState | null,
  moveReader: LegalPropertyReader | null,
): string | null {
  const parsedResult = parseLegalPath(path);
  if (!parsedResult.ok) {
    return parsedResult.error;
  }
  const parsed = parsedResult.value;
  const q = goQuote(path);
  let reader: LegalPropertyReader;
  switch (parsed.kind) {
    case 'game':
      if (exampleState === null) {
        return `boardgame: legal path ${q}: no example state provided to validate against`;
      }
      reader = exampleState.game();
      break;
    case 'player':
    case 'proposer':
    case 'playersAll': {
      if (exampleState === null) {
        return `boardgame: legal path ${q}: no example state provided to validate against`;
      }
      const players = exampleState.players();
      if (players.length === 0) {
        return `boardgame: legal path ${q}: example state has no player states to validate against`;
      }
      reader = players[0];
      break;
    }
    case 'move':
      if (moveReader === null) {
        return `boardgame: legal path ${q}: is a move.* path but no move reader was provided to validate against`;
      }
      reader = moveReader;
      break;
    case 'playersMoveField': {
      if (moveReader === null) {
        return `boardgame: legal path ${q}: is a players[move.*] path but no move reader was provided to validate against`;
      }
      if (!hasProp(moveReader, parsed.moveField)) {
        return `boardgame: legal path ${q}: move field ${goQuote(parsed.moveField)} does not exist`;
      }
      const fieldType = moveReader.props()[parsed.moveField];
      if (fieldType !== 'TypePlayerIndex') {
        return `boardgame: legal path ${q}: move field ${goQuote(parsed.moveField)} has PropertyType ${fieldType}, expected TypePlayerIndex — declare the field as boardgame.PlayerIndex (an int-typed field is not accepted: nothing marks a plain int as a player index, so a wrong-but-in-range value would silently read another player's state)`;
      }
      if (exampleState === null) {
        return `boardgame: legal path ${q}: no example state provided to validate against`;
      }
      const players = exampleState.players();
      if (players.length === 0) {
        return `boardgame: legal path ${q}: example state has no player states to validate against`;
      }
      reader = players[0];
      break;
    }
  }
  if (!hasProp(reader, parsed.prop)) {
    return `boardgame: legal path ${q}: property ${goQuote(parsed.prop)} does not exist`;
  }
  if (parsed.boardIndexField === '') {
    return null;
  }
  const propType = reader.props()[parsed.prop];
  if (propType !== 'TypeBoard') {
    return `boardgame: legal path ${q}: property ${goQuote(parsed.prop)} has PropertyType ${propType}, expected TypeBoard before [move.${parsed.boardIndexField}]`;
  }
  if (moveReader === null) {
    return `boardgame: legal path ${q}: indexed board requires a move reader`;
  }
  if (!hasProp(moveReader, parsed.boardIndexField)) {
    return `boardgame: legal path ${q}: move field ${goQuote(parsed.boardIndexField)} does not exist`;
  }
  const selectorType = moveReader.props()[parsed.boardIndexField];
  if (selectorType !== 'TypeInt' && selectorType !== 'TypeEnum') {
    return `boardgame: legal path ${q}: move field ${goQuote(parsed.boardIndexField)} has PropertyType ${selectorType}, expected TypeInt or TypeEnum`;
  }
  const { board, error } = readBoard(reader, parsed.prop);
  if (board === null) {
    return `boardgame: legal path ${q}: board property ${goQuote(parsed.prop)} was unavailable: ${error}`;
  }
  if (selectorType === 'TypeEnum') {
    let selector: LegalEnumValue | null = null;
    let selectorError = '<nil>';
    try {
      selector = (moveReader.prop(parsed.boardIndexField) as LegalEnumValue | null) ?? null;
    } catch (err) {
      selectorError = thrownText(err);
    }
    if (selector === null) {
      return `boardgame: legal path ${q}: enum move field ${goQuote(parsed.boardIndexField)} was unavailable: ${selectorError}`;
    }
    const boardEnum = board.enum();
    if (boardEnum === null || boardEnum.name !== selector.enum.name) {
      return `boardgame: legal path ${q}: enum move field ${goQuote(parsed.boardIndexField)} does not match board enum`;
    }
  }
  return null;
}

/**
 * Go: validateLegalPathTypes. Checks the property at `path` has one of the
 * allowed types. Skipped (null) when the validation target it needs is
 * absent, exactly as Go skips it.
 */
export function validateLegalPathTypes(
  path: string,
  allowed: readonly PropertyType[],
  exampleState: LegalState | null,
  moveReader: LegalPropertyReader | null,
): string | null {
  const parsedResult = parseLegalPath(path);
  if (!parsedResult.ok) {
    return parsedResult.error;
  }
  const parsed = parsedResult.value;
  if (parsed.kind === 'move' && moveReader === null) {
    return null;
  }
  if (parsed.kind === 'playersMoveField' && (moveReader === null || exampleState === null)) {
    return null;
  }
  if (parsed.kind !== 'move' && parsed.kind !== 'playersMoveField' && exampleState === null) {
    return null;
  }
  const invalid = validateLegalPath(path, exampleState, moveReader);
  if (invalid !== null) {
    return invalid;
  }
  const reader =
    parsed.kind === 'game'
      ? (exampleState as LegalState).game()
      : parsed.kind === 'move'
        ? (moveReader as LegalPropertyReader)
        : (exampleState as LegalState).players()[0];
  let actual: PropertyType = reader.props()[parsed.prop];
  if (parsed.boardIndexField !== '') {
    actual = 'TypeStack';
  }
  if (allowed.includes(actual)) {
    return null;
  }
  if (allowed.length === 1) {
    return `boardgame: legal path ${goQuote(path)}: property ${goQuote(parsed.prop)} has PropertyType ${actual}, expected ${allowed[0]}`;
  }
  return `boardgame: legal path ${goQuote(path)}: property ${goQuote(parsed.prop)} has PropertyType ${actual}, expected one of [${allowed.join(' ')}]`;
}
