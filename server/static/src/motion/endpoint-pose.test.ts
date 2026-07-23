import assert from 'node:assert/strict';
import { test } from 'node:test';
import { motionAxesDiffer } from './endpoint-pose.ts';

test('endpoint orientation reports only finite quarter-turn axis changes', () => {
  assert.equal(motionAxesDiffer('natural', 'natural'), false);
  assert.equal(motionAxesDiffer('quarter-turned', 'quarter-turned'), false);
  assert.equal(motionAxesDiffer('natural', 'quarter-turned'), true);
  assert.equal(motionAxesDiffer('quarter-turned', 'natural'), true);
  assert.throws(() => motionAxesDiffer('diagonal' as never, 'natural'), /orientation/);
});
