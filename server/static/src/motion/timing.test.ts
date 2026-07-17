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
      timing: { duration: 0, easing: 'ease-in-out', fill: 'none', endDelay: 40 },
      activeContext: null,
      expectedSettleMs: 40,
    });

    // Preserve the established contract: an explicit duration wins over the
    // reduced-motion default supplied by the item.
    const explicit = resolveMotionTiming(
      { duration: 125 },
      { policy: 'immediate', defaults: { duration: 250 }, reducedMotion: true },
    );
    assert.equal(explicit.kind === 'play' && explicit.timing.duration, 125);
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
});
