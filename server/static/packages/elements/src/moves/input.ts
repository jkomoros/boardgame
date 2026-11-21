export type MoveInputDisposition =
  | 'required'
  | 'server-defaulted'
  | 'context-owned'
  | 'unsupported';

export type MoveInputCodec =
  | 'integer'
  | 'boolean'
  | 'enum'
  | 'player-index'
  | 'string';

export interface MoveInputSchemaField {
  readonly name: string;
  readonly wireType: string;
  readonly disposition: MoveInputDisposition;
  readonly codec?: MoveInputCodec;
  readonly enumName?: string;
  readonly enumValues?: readonly string[];
}

export interface MoveInputSchemaMove {
  readonly name: string;
  readonly fields: readonly MoveInputSchemaField[] | null;
}

export type MoveInputSchema = readonly MoveInputSchemaMove[];

const RESERVED_MOVE_INPUT_FIELDS = new Set(['MoveType', 'admin', 'player', 'ExpectedVersion']);

export type MoveInputErrorCode =
  | 'unknown-move'
  | 'missing-field'
  | 'unknown-field'
  | 'context-owned-field'
  | 'unsupported-field'
  | 'invalid-type'
  | 'invalid-value'
  | 'invalid-input';

export interface MoveInputIssue {
  readonly code: MoveInputErrorCode;
  readonly moveName: string;
  readonly field?: string;
  readonly message: string;
}

export interface MoveInputValidationResult {
  readonly valid: boolean;
  readonly errors: readonly MoveInputIssue[];
}

/** Validates the exact native creator input before any form serialization. */
export function validateCreatorMoveInput(
  schema: MoveInputSchema,
  moveName: string,
  input: unknown,
): MoveInputValidationResult {
  const move = schema.find(candidate => candidate.name === moveName);
  if (!move) {
    return invalid('unknown-move', moveName, undefined, `Unknown move ${JSON.stringify(moveName)}`);
  }

  if (!isPlainRecord(input)) {
    return invalid('invalid-input', moveName, undefined, 'Move input must be a plain object');
  }

  const fields = move.fields ?? [];
  const byName = new Map(fields.map(field => [field.name, field]));
  const errors: MoveInputIssue[] = [];

  for (const field of fields) {
    if (field.disposition === 'required' && !Object.prototype.hasOwnProperty.call(input, field.name)) {
      errors.push(error('missing-field', moveName, field.name, `Required field ${JSON.stringify(field.name)} is missing`));
    }
  }

  for (const [name, value] of Object.entries(input)) {
    if (RESERVED_MOVE_INPUT_FIELDS.has(name)) {
      errors.push(error(
        'context-owned-field', moveName, name,
        `Field ${JSON.stringify(name)} is reserved by the move proposal protocol`,
      ));
      continue;
    }
    const field = byName.get(name);
    if (!field) {
      errors.push(error('unknown-field', moveName, name, `Field ${JSON.stringify(name)} is not defined`));
      continue;
    }
    if (field.disposition === 'context-owned') {
      errors.push(error('context-owned-field', moveName, name, `Field ${JSON.stringify(name)} is supplied by proposal context`));
      continue;
    }
    if (field.disposition === 'unsupported') {
      errors.push(error('unsupported-field', moveName, name, `Field ${JSON.stringify(name)} is not supported by the safe creator API`));
      continue;
    }
    const fieldError = validateValue(moveName, field, value);
    if (fieldError) errors.push(fieldError);
  }

  return { valid: errors.length === 0, errors };
}

/** Validates and converts native creator values to the legacy form wire. */
export function serializeCreatorMoveInput(
  schema: MoveInputSchema,
  moveName: string,
  input: unknown,
): Readonly<Record<string, string>> {
  const validation = validateCreatorMoveInput(schema, moveName, input);
  if (!validation.valid) {
    throw new MoveInputValidationError(validation.errors);
  }
  return Object.fromEntries(Object.entries(input as Record<string, unknown>).map(([name, value]) => [
    name,
    typeof value === 'boolean' ? (value ? '1' : '0') : String(value),
  ]));
}

export class MoveInputValidationError extends Error {
  readonly code = 'BOARDGAME_INVALID_MOVE_INPUT';
	readonly errors: readonly MoveInputIssue[];

	constructor(errors: readonly MoveInputIssue[]) {
    super(errors.map(item => item.message).join('; '));
    this.name = 'MoveInputValidationError';
		this.errors = errors;
  }
}

export class StaleMoveInputSchemaError extends Error {
  readonly code = 'BOARDGAME_STALE_MOVE_INPUT_SCHEMA';
	readonly expected: string;
	readonly actual: string | null | undefined;

	constructor(expected: string, actual: string | null | undefined) {
    super(actual
      ? `Generated move inputs are stale (client ${expected}, server ${actual})`
      : `The server did not provide a move-input schema fingerprint (client ${expected})`);
    this.name = 'StaleMoveInputSchemaError';
		this.expected = expected;
		this.actual = actual;
  }
}

/** Fails closed before proposal when generated client and server disagree. */
export function assertMoveInputSchemaFingerprint(expected: string, actual: string | null | undefined): void {
  if (!actual || actual !== expected) {
    throw new StaleMoveInputSchemaError(expected, actual);
  }
}

/** The single safe proposal boundary used by generated/bound renderers. */
export function serializeCreatorMoveInputForServer(
  schema: MoveInputSchema,
  expectedFingerprint: string,
  serverFingerprint: string | null | undefined,
  moveName: string,
  input: unknown,
): Readonly<Record<string, string>> {
  assertMoveInputSchemaFingerprint(expectedFingerprint, serverFingerprint);
  return serializeCreatorMoveInput(schema, moveName, input);
}

/**
 * Compatibility adapter for legacy `data-arg-*` controls. DOM datasets only
 * contain strings, so translate them to the generated native creator model
 * before passing them through the same exact validator/fingerprint boundary.
 */
export function creatorMoveInputFromLegacyStrings(
  schema: MoveInputSchema,
  moveName: string,
  input: Readonly<Record<string, string>>,
): Readonly<Record<string, unknown>> {
  const move = schema.find(candidate => candidate.name === moveName);
  if (!move) return input;
  const fields = new Map((move.fields ?? []).map(field => [field.name, field]));
  return Object.fromEntries(Object.entries(input).map(([name, value]) => {
    switch (fields.get(name)?.codec) {
      case 'integer':
      case 'player-index': {
        if (!/^-?\d+$/.test(value)) return [name, value];
        const number = Number(value);
        return [name, Number.isFinite(number) && Number.isInteger(number) ? number : value];
      }
      case 'boolean':
        return [name, value === '1' ? true : value === '0' ? false : value];
      default:
        return [name, value];
    }
  }));
}

function validateValue(moveName: string, field: MoveInputSchemaField, value: unknown): MoveInputIssue | null {
  switch (field.codec) {
    case 'integer':
    case 'player-index':
      return typeof value === 'number' && Number.isFinite(value) && Number.isInteger(value)
        ? null
        : error('invalid-type', moveName, field.name, `Field ${JSON.stringify(field.name)} must be a finite integer`);
    case 'boolean':
      return typeof value === 'boolean'
        ? null
        : error('invalid-type', moveName, field.name, `Field ${JSON.stringify(field.name)} must be a boolean`);
    case 'string':
      return typeof value === 'string'
        ? null
        : error('invalid-type', moveName, field.name, `Field ${JSON.stringify(field.name)} must be a string`);
    case 'enum':
      if (typeof value !== 'string') {
        return error('invalid-type', moveName, field.name, `Field ${JSON.stringify(field.name)} must be an enum string`);
      }
      return field.enumValues?.includes(value)
        ? null
        : error('invalid-value', moveName, field.name, `Field ${JSON.stringify(field.name)} is not a valid ${field.enumName ?? 'enum'} value`);
    default:
      return error('unsupported-field', moveName, field.name, `Field ${JSON.stringify(field.name)} has no supported codec`);
  }
}

function isPlainRecord(value: unknown): value is Readonly<Record<string, unknown>> {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) return false;
  const prototype = Object.getPrototypeOf(value);
  return prototype === Object.prototype || prototype === null;
}

function error(code: MoveInputErrorCode, moveName: string, field: string | undefined, message: string): MoveInputIssue {
  return field === undefined
    ? { code, moveName, message }
    : { code, moveName, field, message };
}

function invalid(code: MoveInputErrorCode, moveName: string, field: string | undefined, message: string): MoveInputValidationResult {
  return { valid: false, errors: [error(code, moveName, field, message)] };
}
