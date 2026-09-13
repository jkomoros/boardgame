import { isVisibleComponent } from '../types/boardgame-types.ts';

export type VisibleComponentDiff =
  | Readonly<{
    status: 'exact';
    added: readonly string[];
    removed: readonly string[];
    retained: readonly string[];
  }>
  | Readonly<{
    status: 'ambiguous';
    reason: 'duplicate-before' | 'duplicate-after';
    added: readonly [];
    removed: readonly [];
    retained: readonly [];
  }>;

const EMPTY = Object.freeze([]) as readonly [];

/**
 * Compare visible membership in two sanitized component collections.
 *
 * Opaque entries are unavailable and therefore ignored. A duplicate public ID
 * makes the result ambiguous, so callers cannot guess through malformed or
 * private state. Added and removed describe membership only; they do not prove
 * a transfer, capture, or any other game meaning.
 */
export function diffVisibleComponents(
  before: readonly unknown[],
  after: readonly unknown[],
): VisibleComponentDiff {
  const beforeIds = visibleIds(before);
  if (beforeIds === null) return ambiguous('duplicate-before');
  const afterIds = visibleIds(after);
  if (afterIds === null) return ambiguous('duplicate-after');

  const beforeSet = new Set(beforeIds);
  const afterSet = new Set(afterIds);
  return Object.freeze({
    status: 'exact',
    added: Object.freeze(afterIds.filter(id => !beforeSet.has(id))),
    removed: Object.freeze(beforeIds.filter(id => !afterSet.has(id))),
    retained: Object.freeze(afterIds.filter(id => beforeSet.has(id))),
  });
}

function visibleIds(components: readonly unknown[]): readonly string[] | null {
  const ids: string[] = [];
  const seen = new Set<string>();
  for (const component of components) {
    if (!isVisibleComponent(component)) continue;
    if (seen.has(component.ID)) return null;
    seen.add(component.ID);
    ids.push(component.ID);
  }
  return ids;
}

function ambiguous(reason: 'duplicate-before' | 'duplicate-after'): VisibleComponentDiff {
  return Object.freeze({
    status: 'ambiguous',
    reason,
    added: EMPTY,
    removed: EMPTY,
    retained: EMPTY,
  });
}
