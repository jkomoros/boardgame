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
  it('keeps spatial, property, transform, and opacity changes orthogonal', () => {
    const draft = createStructuralMotionDraft({
      subjectId: 'card-7',
      presence: 'retained',
      provenance: { kind: 'identity' },
      visualSubject: { kind: 'silhouette', shape: 'rounded-rectangle' },
      from,
      to,
      viewportFrom,
      viewportTo,
      inversion: solveFlipGeometry(from, to),
      beforeTransform: 'rotate(1deg)',
      afterTransform: 'rotate(2deg)',
      beforeProperties: { faceUp: false, rotated: false },
      afterProperties: { faceUp: true, rotated: false },
      animatingProperties: ['faceUp', 'rotated'],
      beforeOpacity: '0.4',
      afterOpacity: '1',
    });
    assert.equal(draft.spatial?.offsetFrom, from);
    assert.deepEqual(draft.visualSubject, {
      kind: 'silhouette', shape: 'rounded-rectangle',
    });
    assert.equal(draft.viewport?.from, viewportFrom);
    assert.equal(draft.viewport?.to, viewportTo);
    assert.deepEqual(draft.transform, { before: 'rotate(1deg)', after: 'rotate(2deg)' });
    assert.deepEqual(draft.properties, [{ name: 'faceUp', before: false, after: true }]);
    assert.deepEqual(draft.opacity, { before: 0.4, after: 1 });
    assert.equal(Object.isFrozen(draft), true);
    assert.equal(Object.isFrozen(draft.properties), true);
  });

  it('publishes an immutable generation-bound plan with requested timing', () => {
    const draft = createStructuralMotionDraft({
      subjectId: 'card-8',
      presence: 'appearing',
      provenance: {
        kind: 'stack-history', endpoint: 'source', stackId: 'deck', evidence: 'runner-up',
      },
      from,
      to,
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
      from,
      to,
      viewportFrom,
      viewportTo,
      inversion: solveFlipGeometry(from, to),
    });
    const planned = publishStructuralMotionPlan(13, [{
      draft,
      timingRequest: { policy: 'version', delayMs: 0, durationMs: 250 },
    }]);
    const executing = updateStructuralMotionExecutions(planned, new Map([[
      'card-9',
      {
        status: 'started' as const,
        animations: Object.freeze([Object.freeze({
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
    assert.equal(executing.segments[0].execution.status, 'started');

    const settled = updateStructuralMotionExecutions(executing, new Map([[
      'card-9', {
        status: 'cancelled' as const,
        animations: executing.segments[0].execution.status === 'started'
          ? executing.segments[0].execution.animations
          : [],
      },
    ]]));
    assert.equal(settled.phase, 'settled');
    assert.equal(settled.segments[0].execution.status, 'cancelled');
  });

  it('replaces mutable property values with opaque immutable snapshots', () => {
    const draft = createStructuralMotionDraft({
      subjectId: 'card-private',
      presence: 'retained',
      provenance: { kind: 'identity' },
      from,
      to,
      viewportFrom,
      viewportTo,
      inversion: solveFlipGeometry(from, to),
      beforeProperties: { custom: { secret: 'before' } },
      afterProperties: { custom: { secret: 'after' } },
      animatingProperties: ['custom'],
    });
    assert.deepEqual(draft.properties, [{
      name: 'custom', before: { kind: 'opaque' }, after: { kind: 'opaque' },
    }]);
    assert.equal(Object.isFrozen(draft.properties[0].before), true);
  });

  it('retains viewport endpoints for a stationary morph without claiming travel', () => {
    const draft = createStructuralMotionDraft({
      subjectId: 'card-flip',
      presence: 'retained',
      provenance: { kind: 'identity' },
      from,
      to: from,
      viewportFrom,
      viewportTo: viewportFrom,
      inversion: solveFlipGeometry(from, from),
      beforeProperties: { faceUp: false },
      afterProperties: { faceUp: true },
      animatingProperties: ['faceUp'],
    });
    assert.equal(draft.spatial, undefined);
    assert.deepEqual(draft.viewport, { from: viewportFrom, to: viewportFrom });
    assert.deepEqual(draft.properties, [{ name: 'faceUp', before: false, after: true }]);
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
