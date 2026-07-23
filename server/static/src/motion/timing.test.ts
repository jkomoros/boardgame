import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import {
  resolveMotionTiming,
  usableAnimationContext,
} from './timing.ts';

const context = {
  version: 7,
  startAtMs: 1_200,
  slotDurationMs: 1_000,
  maxAnimationDurationMs: 800,
};

describe('motion timing', () => {
  it('leaves immediate timing local and reports its settlement budget', () => {
    assert.deepEqual(resolveMotionTiming(
      { delay: 20, duration: 100, endDelay: 30, easing: 'linear' },
      { policy: 'immediate', context, nowMs: 1_000 },
    ), {
      kind: 'play',
      timing: { delay: 20, duration: 100, endDelay: 30, easing: 'linear' },
      activeContext: null,
      expectedSettleMs: 150,
    });
  });

  it('applies defaults, reduced motion, and post-delay before scheduling', () => {
    assert.deepEqual(resolveMotionTiming(
      {},
      {
        policy: 'immediate',
        defaults: { duration: 250, easing: 'ease-in-out', fill: 'none' },
        reducedMotion: true,
        postAnimationDelayMs: 40,
      },
    ), {
      kind: 'play',
      timing: {
        delay: 0,
        duration: 0,
        easing: 'ease-in-out',
        fill: 'none',
        endDelay: 40,
      },
      activeContext: null,
      expectedSettleMs: 40,
    });

    // Explicit motion cannot override the accessibility policy or wait for a
    // synchronized version slot, but its semantic hold remains.
    const explicit = resolveMotionTiming(
      { delay: 40, duration: 125, endDelay: 30 },
      { policy: 'version', context, defaults: { duration: 250 }, reducedMotion: true },
    );
    assert.deepEqual(explicit, {
      kind: 'play',
      timing: { delay: 0, duration: 0, endDelay: 30 },
      activeContext: null,
      expectedSettleMs: 30,
    });
  });

  it('compiles version wait, stagger, clipping, and backwards fill', () => {
    assert.deepEqual(resolveMotionTiming(
      { delay: 100, duration: 900, endDelay: 100, fill: 'none' },
      { context, nowMs: 1_000 },
    ), {
      kind: 'play',
      timing: { delay: 300, duration: 600, endDelay: 100, fill: 'backwards' },
      activeContext: context,
      expectedSettleMs: 1_000,
    });
  });

  it('preserves forward fill and clips repeated active duration to the slot', () => {
    const result = resolveMotionTiming(
      { delay: 100, duration: 400, iterations: 3, fill: 'forwards' },
      { context, nowMs: 1_000 },
    );
    assert.equal(result.kind, 'play');
    if (result.kind !== 'play') return;
    assert.equal(result.timing.delay, 300);
    assert.equal(result.timing.duration, 700 / 3);
    assert.equal(result.timing.iterations, 3);
    assert.equal(result.timing.fill, 'both');
    assert.equal(result.expectedSettleMs, 1_000);
  });

  it('joins a late version only for its remaining visible budget', () => {
    const result = resolveMotionTiming(
      { duration: 900 },
      { context, nowMs: 1_500 },
    );
    assert.equal(result.kind, 'play');
    if (result.kind !== 'play') return;
    assert.equal(result.timing.duration, 500);
    assert.equal(result.activeContext?.maxAnimationDurationMs, 500);
  });

  it('skips when stagger consumes the complete version window', () => {
    assert.deepEqual(resolveMotionTiming(
      { delay: 800, duration: 100 },
      { context, nowMs: 1_000 },
    ), { kind: 'skip', reason: 'timing' });
  });

  it('compiles an explicit local start without a version context', () => {
    const result = resolveMotionTiming(
      { delay: 25, duration: 100 },
      { policy: { localStartAtMs: 1_100 }, context, nowMs: 1_000 },
    );
    assert.equal(result.kind, 'play');
    if (result.kind !== 'play') return;
    assert.deepEqual(result.timing, { delay: 125, duration: 100, fill: 'backwards' });
    assert.equal(result.activeContext, null);
  });

  it('rejects unusable windows and degrades their playback to immediate', () => {
    assert.equal(usableAnimationContext(context, 12_000), null);
    assert.deepEqual(resolveMotionTiming(
      { duration: 100 },
      { context, nowMs: 12_000 },
    ), {
      kind: 'play',
      timing: { duration: 100 },
      activeContext: null,
      expectedSettleMs: 100,
    });
  });

  it('never emits non-finite WAAPI timing from malformed runtime input', () => {
    const result = resolveMotionTiming(
      { delay: Number.NaN, duration: Number.POSITIVE_INFINITY },
      {
        policy: { localStartAtMs: Number.NaN },
        context: { ...context, maxAnimationDurationMs: Number.NaN },
        nowMs: 1_000,
      },
    );
    assert.deepEqual(result, {
      kind: 'play',
      timing: { delay: 0, duration: 0 },
      activeContext: null,
      expectedSettleMs: 0,
    });
    assert.equal(usableAnimationContext(
      { ...context, startAtMs: Number.NaN },
      1_000,
    ), null);
  });

  it('reports the nonnegative WAAPI end time for repeats and negative delays', () => {
    const repeated = resolveMotionTiming(
      { delay: -50, duration: 100, iterations: 3, endDelay: -25 },
      { policy: 'immediate' },
    );
    assert.equal(repeated.kind === 'play' && repeated.expectedSettleMs, 225);

    const alreadyElapsed = resolveMotionTiming(
      { delay: -500, duration: 100, endDelay: -20 },
      { policy: 'immediate' },
    );
    assert.equal(alreadyElapsed.kind === 'play' && alreadyElapsed.expectedSettleMs, 0);
  });
});
