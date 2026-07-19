import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { solveFlipGeometry } from './geometry.ts';
import {
  createStructuralMotionDraft,
  publishStructuralMotionPlan,
  updateStructuralMotionExecutions,
} from './structural-plan.ts';

const from = Object.freeze({
  space: 'offset' as const, top: 10, left: 20, width: 30, height: 40,
});
const to = Object.freeze({
  space: 'offset' as const, top: 40, left: 50, width: 30, height: 40,
});
const viewportFrom = Object.freeze({
  space: 'viewport' as const, top: 110, left: 120, width: 30, height: 40,
});
const viewportTo = Object.freeze({
  space: 'viewport' as const, top: 140, left: 150, width: 30, height: 40,
});

describe('structural motion plans', () => {
  it('publishes only a safe path, subject capability, provenance, and named channels', () => {
    const draft = createStructuralMotionDraft({
      subjectId: 'card-7',
      presence: 'retained',
      provenance: { kind: 'identity' },
      visualSubject: { kind: 'silhouette', shape: 'rounded-rectangle' },
      viewportFrom,
      viewportTo,
      inversion: solveFlipGeometry(from, to),
      channels: [
        { target: 'host', property: 'transform' },
        { target: 'host', property: 'opacity' },
        { target: 'visual', property: 'transform' },
      ],
    });
    assert.deepEqual(draft.visualSubject, {
      kind: 'silhouette', shape: 'rounded-rectangle',
    });
    assert.deepEqual(draft.path, {
      kind: 'travel', from: viewportFrom, to: viewportTo,
    });
    assert.deepEqual(draft.channels, [
      'host:transform',
      'host:opacity',
      'visual:transform',
    ]);
    assert.equal(Object.isFrozen(draft), true);
    assert.equal(Object.isFrozen(draft.path), true);
    assert.equal(Object.isFrozen(draft.channels), true);
  });

  it('publishes an immutable generation-bound plan with requested timing', () => {
    const draft = createStructuralMotionDraft({
      subjectId: 'card-8',
      presence: 'appearing',
      provenance: {
        kind: 'stack-history', endpoint: 'source', stackId: 'deck', evidence: 'runner-up',
      },
      viewportFrom,
      viewportTo,
      inversion: solveFlipGeometry(from, to),
    });
    const plan = publishStructuralMotionPlan(12, [{
      draft,
      timingRequest: { policy: 'version', delayMs: 75, durationMs: 250 },
    }]);
    assert.equal(plan.generation, 12);
    assert.equal(plan.source, 'flip');
    assert.equal(plan.phase, 'planned');
    assert.equal(plan.segments[0].presence, 'appearing');
    assert.deepEqual(plan.segments[0].timingRequest, {
      policy: 'version', delayMs: 75, durationMs: 250,
    });
    assert.equal(Object.isFrozen(plan), true);
    assert.equal(Object.isFrozen(plan.segments), true);
    assert.equal(Object.isFrozen(plan.segments[0]), true);
  });

  it('tracks actual start and terminal outcomes without mutating prior plans', () => {
    const draft = createStructuralMotionDraft({
      subjectId: 'card-9',
      presence: 'retained',
      provenance: { kind: 'identity' },
      viewportFrom,
      viewportTo,
      inversion: solveFlipGeometry(from, to),
    });
    const planned = publishStructuralMotionPlan(13, [{
      draft,
      timingRequest: { policy: 'version', delayMs: 0, durationMs: 250 },
    }]);
    const executing = updateStructuralMotionExecutions(planned, new Map([[
      0,
      {
        status: 'armed' as const,
        animations: Object.freeze([Object.freeze({
          channel: 'host:transform' as const,
          delayMs: 50,
          durationMs: 200,
          endDelayMs: 0,
          iterations: 1,
          easing: 'ease-in-out',
          fill: 'backwards' as const,
        })]),
      },
    ]]));
    assert.equal(planned.phase, 'planned');
    assert.equal(planned.segments[0].execution.status, 'planned');
    assert.equal(executing.phase, 'executing');
    assert.equal(executing.segments[0].execution.status, 'armed');
    assert.deepEqual(executing.segments[0].ref, {
      source: 'flip', generation: 13, segmentIndex: 0,
    });

    const settled = updateStructuralMotionExecutions(executing, new Map([[
      0, {
        status: 'cancelled' as const,
        animations: executing.segments[0].execution.status === 'armed'
          ? executing.segments[0].execution.animations
          : [],
      },
    ]]));
    assert.equal(settled.phase, 'settled');
    assert.equal(settled.segments[0].execution.status, 'cancelled');
  });

  it('updates exact indexes and ignores racing lifecycle regressions', () => {
    const duplicate = createStructuralMotionDraft({
      subjectId: 'duplicate',
      presence: 'retained',
      provenance: { kind: 'identity' },
      viewportFrom,
      viewportTo,
      inversion: solveFlipGeometry(from, to),
    });
    const plan = publishStructuralMotionPlan(20, [0, 1].map(() => ({
      draft: duplicate,
      timingRequest: { policy: 'version' as const, delayMs: 0, durationMs: 100 },
    })));
    const animations = Object.freeze([Object.freeze({
      channel: 'host:transform' as const,
      delayMs: 0,
      durationMs: 100,
      endDelayMs: 0,
      iterations: 1,
      easing: 'linear',
      fill: 'both' as const,
    })]);
    const armed = updateStructuralMotionExecutions(plan, new Map([[
      1, { status: 'armed' as const, animations },
    ]]));
    assert.equal(armed.segments[0].execution.status, 'planned');
    assert.equal(armed.segments[1].execution.status, 'armed');

    const invalidJump = updateStructuralMotionExecutions(armed, new Map([[
      1, { status: 'finished' as const, animations },
    ]]));
    assert.equal(invalidJump.segments[1].execution.status, 'armed');
    const active = updateStructuralMotionExecutions(armed, new Map([[
      1, { status: 'active-observed' as const, animations },
    ]]));
    const stale = updateStructuralMotionExecutions(active, new Map([[
      1, { status: 'armed' as const, animations },
    ]]));
    assert.equal(stale.segments[1].execution.status, 'active-observed');
  });

  it('does not retain primitive or object-valued component history', () => {
    const draft = createStructuralMotionDraft({
      subjectId: 'card-private',
      presence: 'retained',
      provenance: { kind: 'identity' },
      viewportFrom,
      viewportTo,
      inversion: solveFlipGeometry(from, to),
      beforeProperties: { custom: { secret: 'before' } },
      afterProperties: { custom: { secret: 'after' } },
      animatingProperties: ['custom'],
      beforeTransform: 'rotate(1deg)',
      afterTransform: 'rotate(2deg)',
      beforeOpacity: '0.4',
      afterOpacity: '1',
    } as Parameters<typeof createStructuralMotionDraft>[0] & Record<string, unknown>);
    assert.equal('properties' in draft, false);
    assert.equal('transform' in draft, false);
    assert.equal('opacity' in draft, false);
  });

  it('retains viewport endpoints for a stationary morph without claiming travel', () => {
    const draft = createStructuralMotionDraft({
      subjectId: 'card-flip',
      presence: 'retained',
      provenance: { kind: 'identity' },
      viewportFrom,
      viewportTo: viewportFrom,
      inversion: solveFlipGeometry(from, from),
    });
    assert.deepEqual(draft.path, {
      kind: 'stationary', from: viewportFrom, to: viewportFrom,
    });
  });

  it('drops malformed or content-bearing visual subject snapshots', () => {
    const draft = createStructuralMotionDraft({
      subjectId: 'private-card',
      presence: 'retained',
      provenance: { kind: 'identity' },
      visualSubject: {
        kind: 'silhouette', shape: 'rounded-rectangle', face: 'Ace of Spades',
      },
    });
    assert.equal(draft.visualSubject, undefined);
  });
});
