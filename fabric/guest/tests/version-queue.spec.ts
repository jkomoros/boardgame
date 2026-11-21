import { expect, test } from '@playwright/test';
import { VersionQueue } from '../src/version-queue.ts';

const entries = (...versions: number[]) => versions.map(version => ({ version }));
const drain = (queue: VersionQueue<{ version: number }>) => {
  const out: number[] = [];
  for (let next = queue.shift(); next; next = queue.shift()) out.push(next.version);
  return out;
};

test('overlapping windows yield every version once, in order', () => {
  const queue = new VersionQueue<{ version: number }>();
  expect(queue.offer('0:0', entries(0)).accepted).toBe(1);
  expect(queue.offer('0:0', entries(0, 1, 2)).accepted).toBe(2);
  expect(queue.offer('0:0', entries(1, 2, 3, 4)).accepted).toBe(2);
  expect(queue.offer('0:0', entries(2, 3, 4)).accepted).toBe(0);
  expect(drain(queue)).toEqual([0, 1, 2, 3, 4]);
});

test('a version already shown is ignored even after the queue drains', () => {
  const queue = new VersionQueue<{ version: number }>();
  queue.offer('0:0', entries(0));
  queue.offer('0:0', entries(0, 1));
  drain(queue);
  expect(queue.offer('0:0', entries(0, 1)).accepted).toBe(0);
  expect(queue.length).toBe(0);
});

test('an unsorted window is played in version order', () => {
  const queue = new VersionQueue<{ version: number }>();
  queue.offer('0:0', entries(0));
  queue.offer('0:0', entries(3, 1, 2));
  expect(drain(queue)).toEqual([0, 1, 2, 3]);
});

test('the first window under a key cuts to its newest version', () => {
  const queue = new VersionQueue<{ version: number }>();
  expect(queue.offer('0:0', entries(4, 5, 6)).accepted).toBe(1);
  expect(drain(queue)).toEqual([6]);
  // The versions it skipped count as seen.
  expect(queue.offer('0:0', entries(4, 5, 6, 7)).accepted).toBe(1);
  expect(drain(queue)).toEqual([7]);
});

test('a window that starts late reports the gap and still plays', () => {
  const queue = new VersionQueue<{ version: number }>();
  queue.offer('0:0', entries(1));
  const result = queue.offer('0:0', entries(5, 6));
  expect(result.gap).toEqual({ after: 1, next: 5 });
  expect(drain(queue)).toEqual([1, 5, 6]);
});

test('a new key discards the queue and accepts lower versions again', () => {
  const queue = new VersionQueue<{ version: number }>();
  queue.offer('0:0', entries(5));
  drain(queue);
  const result = queue.offer('0:1', entries(5));
  expect(result.reset).toBe(true);
  expect(drain(queue)).toEqual([5]);
  expect(queue.offer('1:1', entries(0)).accepted).toBe(1);
});

test('an invalid version is refused', () => {
  const queue = new VersionQueue<{ version: number }>();
  expect(() => queue.offer('0:0', entries(-1))).toThrow(/invalid version/);
  expect(() => queue.offer('0:0', entries(1.5))).toThrow(/invalid version/);
});
