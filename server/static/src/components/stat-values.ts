import type { ExpandedStack, RawStack } from '../types/boardgame-types.js';

/**
 * How a stat reads a count and a capacity off a stack, kept separate from the
 * element so it can be tested without a DOM.
 *
 * THE WHOLE POINT: `boardgame-stat` takes the STACK, never two numbers.
 *
 * Every renderer that wrote `Food ${n}/${m}` by hand had to know two things
 * that are not obvious and are not the same for every stack:
 *
 *   1. **Which field is the capacity.** A sized stack serializes its fixed slot
 *      count as `Size` and does NOT emit `MaxSize` (stack.go's
 *      `sizedStack.MarshalJSON`). A growable stack with a cap does the
 *      opposite: it emits `MaxSize` and no `Size`. An author who reaches for
 *      one of them gets a stat that silently renders no capacity for half the
 *      stacks in the game. Both are `json:",omitempty"`, so a zero -- a
 *      growable stack with no cap -- correctly arrives as absent.
 *   2. **Which slots count as full.** A sized stack pads `Indexes` with a -1
 *      sentinel and `Components` with `null` at every empty slot, so
 *      `Indexes.length` is the CAPACITY, not the count. `Components.length` is
 *      the same trap. Only the non-empty slots are the count -- including
 *      sanitization's opaque `{}` components, which are occupied slots whose
 *      contents this viewer may not see.
 *
 * Making the author restate the capacity is how the two get out of sync. The
 * framework already knows it; a stat asks the stack.
 */

/** Anything with the wire shape of a stack, expanded or not. */
export type StatStack = ExpandedStack | RawStack;

/** True when `value` has the wire shape of a stack. Used to fail loudly. */
export function isStatStack(value: unknown): value is StatStack {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return false;
  const candidate = value as Partial<ExpandedStack>;
  return typeof candidate.Deck === 'string'
    && Array.isArray(candidate.Indexes)
    && Array.isArray(candidate.IDs);
}

/**
 * The stack's own capacity, or undefined when it has none.
 *
 * `MaxSize` first: a merged stack over sized stacks reports `Size`, and a
 * growable stack with a cap reports `MaxSize`, but nothing reports both, so the
 * order only matters for defensiveness against a future stack kind that does.
 */
export function statCapacity(stack: StatStack | null | undefined): number | undefined {
  if (!stack) return undefined;
  const max = stack.MaxSize;
  if (typeof max === 'number' && Number.isFinite(max) && max > 0) return max;
  const size = stack.Size;
  if (typeof size === 'number' && Number.isFinite(size) && size > 0) return size;
  return undefined;
}

/**
 * How many slots of the stack are occupied.
 *
 * Prefers `Components` when the stack has been expanded, because an opaque
 * component -- sanitization's `{}` -- is a filled slot even though its index
 * may have been erased. Falls back to `Indexes` for a raw stack nested inside a
 * board, which the selector does not expand.
 */
export function statCount(stack: StatStack | null | undefined): number {
  if (!stack) return 0;
  const components = (stack as ExpandedStack).Components;
  if (Array.isArray(components)) {
    return components.reduce<number>((total, slot) => total + (slot === null || slot === undefined ? 0 : 1), 0);
  }
  return stack.Indexes.reduce<number>((total, index) => total + (index >= 0 ? 1 : 0), 0);
}
