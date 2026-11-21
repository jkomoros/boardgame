import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { primaryStructuralAnimationIndex } from './activation.ts';
import type { StructuralExecutedTiming } from './structural-plan.ts';

const timing = (
  channel: StructuralExecutedTiming['channel'],
  delayMs: number,
): StructuralExecutedTiming => ({
  channel,
  delayMs,
  durationMs: 100,
  endDelayMs: 0,
  iterations: 1,
  easing: 'linear',
  fill: 'both',
});

describe('structural activation selection', () => {
  it('uses the spatial host channel for travel even when another channel starts first', () => {
    assert.equal(primaryStructuralAnimationIndex('travel', [
      timing('visual:transform', 0),
      timing('host:transform', 40),
    ]), 1);
  });

  it('uses the earliest participating channel for a stationary morph', () => {
    assert.equal(primaryStructuralAnimationIndex('stationary', [
      timing('visual:transform', 30),
      timing('host:opacity', 10),
    ]), 1);
    assert.equal(primaryStructuralAnimationIndex(undefined, []), null);
  });
});
