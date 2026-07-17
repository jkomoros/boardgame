import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { solveFlipGeometry } from './geometry.ts';
import {
  createStructuralMotionDraft,
  publishStructuralMotionPlan,
} from './structural-plan.ts';

const from = Object.freeze({
  space: 'offset' as const, top: 10, left: 20, width: 30, height: 40,
});
const to = Object.freeze({
  space: 'offset' as const, top: 40, left: 50, width: 30, height: 40,
});

describe('structural motion plans', () => {
  it('keeps spatial, property, transform, and opacity changes orthogonal', () => {
    const draft = createStructuralMotionDraft({
      subjectId: 'card-7',
      presence: 'retained',
      provenance: { kind: 'identity' },
      from,
      to,
      inversion: solveFlipGeometry(from, to),
      beforeTransform: 'rotate(1deg)',
      afterTransform: 'rotate(2deg)',
      beforeProperties: { faceUp: false, rotated: false },
      afterProperties: { faceUp: true, rotated: false },
      animatingProperties: ['faceUp', 'rotated'],
      beforeOpacity: '0.4',
      afterOpacity: '1',
    });
    assert.equal(draft.spatial?.from, from);
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
      inversion: solveFlipGeometry(from, to),
    });
    const plan = publishStructuralMotionPlan(12, [{
      draft,
      timingRequest: { policy: 'version', delayMs: 75, durationMs: 250 },
    }]);
    assert.equal(plan.generation, 12);
    assert.equal(plan.phase, 'ready-to-play');
    assert.equal(plan.segments[0].presence, 'appearing');
    assert.deepEqual(plan.segments[0].timingRequest, {
      policy: 'version', delayMs: 75, durationMs: 250,
    });
    assert.equal(Object.isFrozen(plan), true);
    assert.equal(Object.isFrozen(plan.segments), true);
    assert.equal(Object.isFrozen(plan.segments[0]), true);
  });
});
