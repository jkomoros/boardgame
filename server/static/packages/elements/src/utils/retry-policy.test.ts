import assert from 'node:assert/strict';
import test from 'node:test';
import { retryDelayMs } from './retry-policy.ts';

const policy = { baseDelayMs: 100, maxDelayMs: 1_000, jitterRatio: 0.2 };

test('retry delay grows exponentially, caps, and has bounded jitter', () => {
  assert.equal(retryDelayMs(0, () => 0.5, policy), 100);
  assert.equal(retryDelayMs(3, () => 0.5, policy), 800);
  assert.equal(retryDelayMs(4, () => 0.5, policy), 1_000);
  assert.equal(retryDelayMs(20, () => 0, policy), 800);
  assert.equal(retryDelayMs(20, () => 1, policy), 1_200);
});

test('retry policy rejects invalid inputs loudly', () => {
  assert.throws(() => retryDelayMs(-1), /non-negative integer/);
  assert.throws(() => retryDelayMs(0, () => 2), /between 0 and 1/);
  assert.throws(() => retryDelayMs(0, () => 0.5, { ...policy, jitterRatio: 2 }), /jitterRatio/);
});
