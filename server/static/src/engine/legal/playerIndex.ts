// Go's PlayerIndex arithmetic (state.go), as much of it as the proposer
// predicates use: Valid, EnsureValid (and the Next it falls back to), and
// Equivalent. It reads players only through LegalState.

import type { LegalState } from './access.ts';
import { ADMIN_PLAYER_INDEX, ANY_PLAYER_INDEX, OBSERVER_PLAYER_INDEX } from './types.ts';

/**
 * A Go runtime panic the TypeScript reproduces on purpose, so that
 * evalLegalPredicate turns it into the same fail-closed Unknown Go's recover
 * does, with the same text.
 */
export interface LegalPanic extends Error {
  readonly name: 'LegalPanic';
}

/**
 * Creates a `LegalPanic`. An ordinary `Error` named `LegalPanic`, not a subclass:
 * a Common Fabric pattern module may not declare a class (see README.md,
 * "No classes").
 */
export function legalPanic(message: string): LegalPanic {
  return Object.assign(new Error(message), { name: 'LegalPanic' as const });
}

/** Whether `thrown` is a `LegalPanic`. */
export function isLegalPanic(thrown: unknown): thrown is LegalPanic {
  return thrown instanceof Error && thrown.name === 'LegalPanic';
}

function isSpecial(p: number): boolean {
  return p === ADMIN_PLAYER_INDEX || p === OBSERVER_PLAYER_INDEX || p === ANY_PLAYER_INDEX;
}

/**
 * Go's `state.ImmutablePlayerStates()[p]` bounds check: indexing outside the
 * slice panics, and legality inherits that panic (as an Unknown).
 */
function mayBeActiveAt(state: LegalState, p: number): boolean {
  const count = state.players().length;
  if (p < 0) {
    throw legalPanic(`runtime error: index out of range [${p}]`);
  }
  if (p >= count) {
    throw legalPanic(`runtime error: index out of range [${p}] with length ${count}`);
  }
  return state.playerMayBeActive(p);
}

/** Go: PlayerIndex.Valid(state). */
export function playerIndexValid(p: number, state: LegalState): boolean {
  if (isSpecial(p)) {
    return true;
  }
  if (p < 0 || p >= state.players().length) {
    return false;
  }
  return state.playerMayBeActive(p);
}

/** Go: PlayerIndex.Next(state). */
export function playerIndexNext(p: number, state: LegalState): number {
  if (isSpecial(p)) {
    return p;
  }
  const order = state.customPlayerOrder();
  if (order !== null) {
    return nextInOrder(p, state, order);
  }
  const count = state.players().length;
  const original = p;
  p++;
  if (p >= count) {
    p = 0;
  }
  let tried = 0;
  while (!mayBeActiveAt(state, p)) {
    p++;
    if (p >= count) {
      p = 0;
    }
    tried++;
    if (tried >= count) {
      return original;
    }
  }
  return p;
}

/** Go: PlayerIndex.nextInOrder. */
function nextInOrder(p: number, state: LegalState, order: readonly number[]): number {
  const pos = order.indexOf(p);
  if (pos === -1) {
    return p;
  }
  const n = order.length;
  const count = state.players().length;
  for (let i = 1; i <= n; i++) {
    const candidate = order[(pos + i) % n];
    if (candidate < 0 || candidate >= count) {
      return p;
    }
    if (state.playerMayBeActive(candidate)) {
      return candidate;
    }
  }
  return p;
}

/** Go: PlayerIndex.EnsureValid(state). */
export function playerIndexEnsureValid(p: number, state: LegalState): number {
  return playerIndexValid(p, state) ? p : playerIndexNext(p, state);
}

/**
 * Go: PlayerIndex.Equivalent(other). Observer matches nothing; Admin and Any
 * match everything else; anything below AnyPlayerIndex matches nothing.
 */
export function playerIndexEquivalent(p: number, other: number): boolean {
  if (p < ANY_PLAYER_INDEX || other < ANY_PLAYER_INDEX) {
    return false;
  }
  if (p === OBSERVER_PLAYER_INDEX || other === OBSERVER_PLAYER_INDEX) {
    return false;
  }
  if (p === ADMIN_PLAYER_INDEX || other === ADMIN_PLAYER_INDEX) {
    return true;
  }
  if (p === ANY_PLAYER_INDEX || other === ANY_PLAYER_INDEX) {
    return true;
  }
  return p === other;
}
