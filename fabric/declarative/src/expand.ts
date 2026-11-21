/**
 * Pure helpers that turn the wire view a viewer is released into what the
 * layer-1 stack element takes. Nothing here touches the DOM.
 *
 * `expandStack` restates `expandStack` in `server/static/src/selectors.ts`
 * (seam S3: that function lives beside `reselect` and Redux selectors, so it
 * cannot be imported without them). It differs in one deliberate way, for seam
 * S10: an opaque slot is emitted as `{ ID }` instead of `{}`, so the stack
 * never has to write the id onto its input, and the result can be frozen.
 */

export type Json = null | boolean | number | string | readonly Json[] | { readonly [key: string]: Json };

export interface WireStack {
  readonly Deck: string;
  readonly Indexes: readonly number[];
  readonly IDs?: readonly string[];
  readonly IDsLastSeen?: Readonly<Record<string, number>>;
  readonly [key: string]: unknown;
}

export interface Chest {
  readonly Decks?: Readonly<Record<string, readonly { readonly Index: number; readonly Values: Readonly<Record<string, unknown>> }[]>>;
  readonly [key: string]: unknown;
}

export interface WireState {
  readonly Game?: Readonly<Record<string, unknown>>;
  readonly Players?: readonly Readonly<Record<string, unknown>>[];
  readonly Components?: Readonly<Record<string, readonly unknown[]>>;
  readonly Computed?: unknown;
  readonly [key: string]: unknown;
}

/** One viewer's released view: the `game` object of the JSON API. */
export interface WireView {
  readonly CurrentState: WireState;
  readonly CurrentPlayerIndex?: number;
  readonly Finished?: boolean;
  readonly Winners?: readonly number[] | null;
  readonly Version?: number;
  readonly Name?: string;
  readonly [key: string]: unknown;
}

export function isWireStack(value: unknown): value is WireStack {
  return !!value && typeof value === 'object'
    && typeof (value as WireStack).Deck === 'string'
    && Array.isArray((value as WireStack).Indexes);
}

function deepFreeze<T>(value: T): T {
  if (value && typeof value === 'object' && !Object.isFrozen(value)) {
    Object.freeze(value);
    for (const child of Object.values(value as object)) deepFreeze(child);
  }
  return value;
}

/** Expand one wire stack against the chest. The result is deeply frozen. */
export function expandStack(
  stack: WireStack,
  chest: Chest | null | undefined,
  gameName: string,
  state?: WireState | null,
): unknown {
  const deck = chest?.Decks?.[stack.Deck];
  const components = stack.Indexes.map((index, slot) => {
    if (index === -1) return null;
    const id = stack.IDs?.[slot] ?? '';
    // -2 is sanitization's "occupied, contents withheld".
    if (index < 0 || !deck || !deck[index]) return id ? { ID: id } : {};
    const dynamic = state?.Components?.[stack.Deck]?.[index];
    return {
      Index: deck[index]!.Index,
      Values: deck[index]!.Values,
      ...(dynamic && typeof dynamic === 'object' ? { DynamicValues: dynamic } : {}),
      ID: id,
      Deck: stack.Deck,
      GameName: gameName,
    };
  });
  return deepFreeze({ ...stack, GameName: gameName, Components: components });
}

/**
 * Resolve a dotted path such as `Game.Cards` or `Players.0.WonCards`.
 * Two virtual trailing segments exist so that a count never has to be
 * computed by the caller and passed as a changing property:
 *   `.$count`  number of occupied slots of a stack
 *   `.$length` length of an array or of a stack's slot list
 */
export function resolvePath(root: unknown, path: string): unknown {
  if (!path) return root;
  let value: unknown = root;
  for (const segment of path.split('.')) {
    if (value === null || value === undefined) return undefined;
    if (segment === '$count') {
      return isWireStack(value) ? value.Indexes.filter(index => index !== -1).length : undefined;
    }
    if (segment === '$length') {
      if (isWireStack(value)) return value.Indexes.length;
      return Array.isArray(value) ? value.length : undefined;
    }
    if (typeof value !== 'object') return undefined;
    // Own properties only: a path must not be able to reach a prototype.
    if (!Object.prototype.hasOwnProperty.call(value, segment)) return undefined;
    value = (value as Record<string, unknown>)[segment];
  }
  return value;
}
