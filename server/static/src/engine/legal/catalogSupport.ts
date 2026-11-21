// Shared helpers for the catalog files: argument and template plumbing, and
// the typed property reads Go does through a PropertyReader's typed methods.

import type { LegalEnumValue, LegalPropertyReader } from './access.ts';
import type { Result } from './path.ts';
import { LegalSpecError } from './predicate.ts';
import { type LegalSpec, type LegalVerdict, unknownVerdict } from './types.ts';

/** A spec's args, never undefined. */
export function specArgs(spec: LegalSpec): readonly string[] {
  return spec.args ?? [];
}

/** Go: `template := spec.Message; if template == "" { template = default }`. */
export function specTemplate(spec: LegalSpec, defaultTemplate: string): string {
  return spec.message ? spec.message : defaultTemplate;
}

/** Throws the construction error Go's constructor returns. */
export function specError(message: string): never {
  throw new LegalSpecError(message);
}

/** Go: `return UnknownVerdict(err.Error())`. */
export function unknownFrom(result: { ok: false; error: string }): LegalVerdict {
  return unknownVerdict(result.error);
}

function hasProp(reader: LegalPropertyReader, name: string): boolean {
  return Object.prototype.hasOwnProperty.call(reader.props(), name);
}

/**
 * Go: reader.BoolProp(name) / reader.IntProp(name) as allActivePlayers' inner
 * leaves use them: only success or failure matters there, so the result is
 * the value or undefined.
 */
export function readTyped(reader: LegalPropertyReader, name: string, type: 'TypeBool' | 'TypeInt'): boolean | number | undefined {
  if (!hasProp(reader, name) || reader.props()[name] !== type) {
    return undefined;
  }
  let value: unknown;
  try {
    value = reader.prop(name);
  } catch {
    return undefined;
  }
  if (type === 'TypeBool') {
    return typeof value === 'boolean' ? value : undefined;
  }
  return typeof value === 'number' ? value : undefined;
}

/**
 * Go: reader.ImmutableEnumProp(name), with the error text Go's generated
 * readers produce for a property that is not an enum. A null value is Go's
 * nil enum value.
 */
export function readEnumProp(reader: LegalPropertyReader, name: string): Result<LegalEnumValue | null> {
  if (!hasProp(reader, name) || reader.props()[name] !== 'TypeEnum') {
    return { ok: false, error: 'No such Enum prop: ' + name };
  }
  try {
    return { ok: true, value: (reader.prop(name) as LegalEnumValue | null) ?? null };
  } catch (err) {
    return { ok: false, error: err instanceof Error ? err.message : String(err) };
  }
}

/** Go: enum.ImmutableVal.Equals. Same enum and same key. */
export function enumValuesEqual(a: LegalEnumValue, b: LegalEnumValue | null): boolean {
  return b !== null && a.enum.name === b.enum.name && a.value === b.value;
}
