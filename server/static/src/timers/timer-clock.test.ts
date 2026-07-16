import assert from 'node:assert/strict';
import test from 'node:test';
import { remainingTimerMs } from './timer-clock.ts';

test('countdown always subtracts from the installed baseline', () => {
  assert.equal(remainingTimerMs(10_000, 1_000, 1_100), 9_900);
  assert.equal(remainingTimerMs(10_000, 1_000, 1_200), 9_800);
  assert.equal(remainingTimerMs(10_000, 1_000, 20_000), 0);
});

test('clock rollback cannot add time and malformed clock data fails loudly', () => {
  assert.equal(remainingTimerMs(10_000, 1_000, 900), 10_000);
  assert.throws(() => remainingTimerMs(-1, 0, 0), /initialTimeLeft/);
  assert.throws(() => remainingTimerMs(1, Number.NaN, 0), /wall-clock/);
});
