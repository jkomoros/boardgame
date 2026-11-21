import assert from 'node:assert/strict';
import test from 'node:test';
import { isStatStack, statCapacity, statCount, type StatStack } from './stat-values.ts';

/**
 * These are the exact wire shapes `stack.go` produces, transcribed from its
 * MarshalJSON implementations, because the whole reason `boardgame-stat` takes
 * a stack rather than two numbers is that reading them by hand is a trap:
 *
 *   growableStack.MarshalJSON -> MaxSize, never Size
 *   sizedStack.MarshalJSON   -> Size, never MaxSize
 *   mergedStack.MarshalJSON  -> Size when it is over sized stacks, else MaxSize
 *
 * and both fields are `json:",omitempty"`, so a zero arrives as absent.
 */

const growableUncapped = {
  Deck: 'cards', Indexes: [3, 7], IDs: ['a', 'b'], IDsLastSeen: {}, ShuffleCount: 0,
  GameName: 'g', Components: [{}, {}],
} as unknown as StatStack;

const growableCapped = {
  Deck: 'cards', Indexes: [3, 7], IDs: ['a', 'b'], IDsLastSeen: {}, ShuffleCount: 0,
  MaxSize: 5, GameName: 'g', Components: [{}, {}],
} as unknown as StatStack;

/** Six slots, three of them filled. `Food 3/6`. */
const sized = {
  Deck: 'food', Indexes: [4, -1, 9, -1, 2, -1], IDs: ['a', '', 'b', '', 'c', ''],
  IDsLastSeen: {}, ShuffleCount: 0, Size: 6, GameName: 'g',
  Components: [{ Index: 4 }, null, { Index: 9 }, null, { Index: 2 }, null],
} as unknown as StatStack;

/** A raw stack nested in a board: the selector never expands it. */
const raw = {
  Deck: 'food', Indexes: [4, -1, 9], IDs: ['a', '', 'b'], IDsLastSeen: {},
  ShuffleCount: 0, Size: 3,
} as unknown as StatStack;

test('statCapacity', async (t) => {
  await t.test('reads a sized stack\'s capacity from Size, which is the only field it emits', () => {
    assert.equal(statCapacity(sized), 6);
  });

  await t.test('reads a capped growable stack\'s capacity from MaxSize, which is the only field IT emits', () => {
    assert.equal(statCapacity(growableCapped), 5);
  });

  await t.test('reports no capacity for an uncapped growable stack', () => {
    // MaxSize 0 is omitempty'd away on the wire, so "no cap" arrives as absent
    // rather than as a zero -- and a stat must render a bare value, not "2/0".
    assert.equal(statCapacity(growableUncapped), undefined);
  });

  await t.test('treats an explicit zero or a non-finite value as no capacity', () => {
    assert.equal(statCapacity({ ...growableCapped, MaxSize: 0 } as StatStack), undefined);
    assert.equal(statCapacity({ ...growableCapped, MaxSize: Number.NaN } as StatStack), undefined);
  });

  await t.test('has no capacity for a null or undefined stack', () => {
    assert.equal(statCapacity(null), undefined);
    assert.equal(statCapacity(undefined), undefined);
  });
});

test('statCount', async (t) => {
  await t.test('counts the FILLED slots of a sized stack, not its length', () => {
    // The trap: Indexes.length and Components.length are both 6 here, which is
    // the capacity. Three renderers wrote `.Indexes.length` for a count.
    assert.equal(statCount(sized), 3);
    assert.equal((sized as { Indexes: readonly number[] }).Indexes.length, 6);
  });

  await t.test('counts every component of a growable stack', () => {
    assert.equal(statCount(growableUncapped), 2);
  });

  await t.test('falls back to Indexes for a raw stack that was never expanded', () => {
    assert.equal(statCount(raw), 2);
  });

  await t.test('counts an opaque sanitized component as a filled slot', () => {
    // `{}` is sanitization's "a component is here and you may not see it".
    const opaque = {
      Deck: 'cards', Indexes: [-1, -1], IDs: ['x', 'y'], IDsLastSeen: {}, ShuffleCount: 0,
      Size: 2, GameName: 'g', Components: [{}, null],
    } as unknown as StatStack;
    assert.equal(statCount(opaque), 1);
  });

  await t.test('is zero for a null or undefined stack', () => {
    assert.equal(statCount(null), 0);
    assert.equal(statCount(undefined), 0);
  });
});

test('isStatStack', async (t) => {
  await t.test('accepts a real stack', () => {
    assert.equal(isStatStack(sized), true);
    assert.equal(isStatStack(raw), true);
  });

  await t.test('rejects a bare number, which is the mistake it exists to catch', () => {
    assert.equal(isStatStack(6), false);
  });

  await t.test('rejects null, arrays, and objects that only look stack-ish', () => {
    assert.equal(isStatStack(null), false);
    assert.equal(isStatStack(undefined), false);
    assert.equal(isStatStack([1, 2, 3]), false);
    assert.equal(isStatStack({ Deck: 'cards' }), false);
    assert.equal(isStatStack({ Indexes: [], IDs: [] }), false);
  });
});
