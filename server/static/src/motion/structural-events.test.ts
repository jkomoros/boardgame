import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { solveFlipGeometry } from './geometry.ts';
import { compileStructuralMotionEvents } from './structural-events.ts';
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

function planned(generation = 7) {
  const draft = createStructuralMotionDraft({
    subjectId: 'card-1',
    presence: 'retained',
    provenance: { kind: 'identity' },
    from,
    to,
    viewportFrom,
    viewportTo,
    inversion: solveFlipGeometry(from, to),
  });
  return publishStructuralMotionPlan(generation, [{
    draft,
    timingRequest: { policy: 'version', delayMs: 0, durationMs: 250 },
  }]);
}

describe('structural motion event compilation', () => {
  it('emits only newly observed execution transitions', () => {
    const intention = planned();
    const [plannedEvent] = compileStructuralMotionEvents(null, intention);
    assert.equal(plannedEvent.kind, 'planned');
    assert.equal(plannedEvent.id, 'flip:7:0:planned');
    assert.equal(plannedEvent.kind !== 'generation-settled' && plannedEvent.segmentId, 'flip:7:0');
    assert.equal(plannedEvent.segment, intention.segments[0]);
    assert.deepEqual(compileStructuralMotionEvents(intention, intention), []);

    const executing = updateStructuralMotionExecutions(intention, new Map([[
      0, {
        status: 'armed' as const,
        animations: [{
          channel: 'host:transform' as const,
          delayMs: 10,
          durationMs: 200,
          endDelayMs: 0,
          iterations: 1,
          easing: 'ease-out',
          fill: 'backwards' as const,
        }],
      },
    ]]));
    assert.deepEqual(
      compileStructuralMotionEvents(intention, executing).map(event => event.kind),
      ['armed'],
    );
    assert.deepEqual(compileStructuralMotionEvents(executing, executing), []);

    const active = updateStructuralMotionExecutions(executing, new Map([[
      0, {
        status: 'active-observed' as const,
        animations: executing.segments[0].execution.status === 'armed'
          ? executing.segments[0].execution.animations
          : [],
      },
    ]]));
    const finished = updateStructuralMotionExecutions(active, new Map([[
      0, {
        status: 'finished' as const,
        animations: active.segments[0].execution.status === 'active-observed'
          ? active.segments[0].execution.animations
          : [],
      },
    ]]));
    assert.deepEqual(
      compileStructuralMotionEvents(active, finished).map(event => event.kind),
      ['finished', 'generation-settled'],
    );
  });

  it('does not fabricate intermediate states after a missed revision', () => {
    const intention = planned();
    const skipped = updateStructuralMotionExecutions(intention, new Map([[
      0, { status: 'skipped' as const, reason: 'timing' as const },
    ]]));
    const events = compileStructuralMotionEvents(null, skipped);
    assert.deepEqual(events.map(event => event.kind), ['skipped', 'generation-settled']);
    assert.equal(events[0].kind === 'skipped' && events[0].segment.execution.status, 'skipped');
  });

  it('treats a new source or generation as a new event identity', () => {
    const first = planned(1);
    const second = publishStructuralMotionPlan(
      1,
      [{ draft: first.segments[0], timingRequest: first.segments[0].timingRequest }],
      'explicit',
    );
    const [event] = compileStructuralMotionEvents(first, second);
    assert.equal(event.id, 'explicit:1:0:planned');
    assert.equal(Object.isFrozen(event), true);
    assert.equal(Object.isFrozen(compileStructuralMotionEvents(first, second)), true);
  });
});
